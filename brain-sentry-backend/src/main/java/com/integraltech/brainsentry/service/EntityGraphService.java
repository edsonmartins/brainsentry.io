package com.integraltech.brainsentry.service;

import com.falkordb.Driver;
import com.falkordb.FalkorDB;
import com.falkordb.Graph;
import com.falkordb.ResultSet;
import com.integraltech.brainsentry.domain.Memory;
import com.integraltech.brainsentry.service.OpenRouterService.EntityExtractionResult;
import com.integraltech.brainsentry.service.OpenRouterService.ExtractedEntity;
import com.integraltech.brainsentry.service.OpenRouterService.ExtractedRelationship;
import jakarta.annotation.PostConstruct;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.HashMap;
import java.util.Map;
import java.util.UUID;

/**
 * Service for storing extracted entities and relationships in FalkorDB.
 *
 * This service takes entities and relationships extracted from text by the LLM
 * and stores them as nodes and edges in the knowledge graph.
 *
 * Example flow:
 * 1. User creates a Memory with content about a sales order
 * 2. LLM extracts entities: CLIENTE:Marcos, VENDEDOR:Ana, PEDIDO:#12345
 * 3. LLM extracts relationships: Marcos-[REALIZOU]->>#12345, Ana-[ATENDEU]->Marcos
 * 4. This service stores entities as nodes and relationships as edges
 * 5. Links everything back to the source Memory
 */
@Slf4j
@Service
@ConditionalOnProperty(name = "features.entity-graph.enabled", havingValue = "true", matchIfMissing = false)
public class EntityGraphService {

    private final OpenRouterService openRouterService;
    private final AuditService auditService;
    private Graph graph;

    @Value("${brain-sentry.graph.name:brainsentry}")
    private String graphName;

    @Value("${brain-sentry.redis.host:localhost}")
    private String redisHost;

    @Value("${brain-sentry.redis.port:6379}")
    private int redisPort;

    @Value("${brain-sentry.redis.password:}")
    private String redisPassword;

    public EntityGraphService(
            @Autowired(required = false) OpenRouterService openRouterService,
            AuditService auditService) {
        this.openRouterService = openRouterService;
        this.auditService = auditService;
    }

    /**
     * Initialize FalkorDB connection after Spring properties are injected.
     */
    @PostConstruct
    public void init() {
        try {
            Driver driver;
            if (redisPassword != null && !redisPassword.isEmpty()) {
                try {
                    driver = FalkorDB.driver(redisHost, redisPort, redisPassword, null);
                } catch (Exception e) {
                    log.warn("Failed to connect with null username, trying 'default': {}", e.getMessage());
                    driver = FalkorDB.driver(redisHost, redisPort, redisPassword, "default");
                }
            } else {
                driver = FalkorDB.driver(redisHost, redisPort);
            }
            this.graph = driver.graph(graphName);
            log.info("EntityGraphService: FalkorDB graph '{}' initialized with connection to {}:{}",
                    graphName, redisHost, redisPort);
        } catch (Exception e) {
            log.error("EntityGraphService: Failed to initialize FalkorDB connection: {}", e.getMessage());
        }
    }

    /**
     * Extract entities and relationships from a Memory's content
     * and store them in the knowledge graph.
     *
     * This method is called asynchronously after a Memory is created.
     *
     * @param memory the memory to extract entities from
     * @param tenantId the tenant ID
     */
    @Async
    public void extractAndStoreEntities(Memory memory, String tenantId) {
        extractAndStoreEntitiesSync(memory, tenantId);
    }

    /**
     * Synchronous version of entity extraction.
     * Used for manual triggering via API endpoints.
     *
     * @param memory the memory to extract entities from
     * @param tenantId the tenant ID
     */
    public void extractAndStoreEntitiesSync(Memory memory, String tenantId) {
        if (openRouterService == null || !openRouterService.isConfigured()) {
            log.warn("OpenRouterService not configured, skipping entity extraction");
            return;
        }

        if (graph == null) {
            log.warn("FalkorDB graph not available, skipping entity extraction");
            return;
        }

        if (memory.getContent() == null || memory.getContent().isBlank()) {
            log.debug("Memory {} has no content, skipping entity extraction", memory.getId());
            return;
        }

        try {
            log.info("Extracting entities from memory {}", memory.getId());

            // 1. Extract entities and relationships using LLM
            EntityExtractionResult extraction = openRouterService.extractEntitiesAndRelationships(memory.getContent());

            if (!extraction.hasEntities()) {
                log.debug("No entities found in memory {}", memory.getId());
                return;
            }

            log.info("Found {} entities and {} relationships in memory {}",
                    extraction.getEntities().size(),
                    extraction.getRelationships().size(),
                    memory.getId());

            // 2. Create a mapping from extraction IDs to graph node IDs
            Map<String, String> idMapping = new HashMap<>();

            // 3. Store each entity as a node in FalkorDB
            for (ExtractedEntity entity : extraction.getEntities()) {
                String nodeId = storeEntityNode(entity, tenantId, memory.getId());
                idMapping.put(entity.getId(), nodeId);
            }

            // 4. Store each relationship as an edge in FalkorDB
            for (ExtractedRelationship relationship : extraction.getRelationships()) {
                String sourceNodeId = idMapping.get(relationship.getSourceId());
                String targetNodeId = idMapping.get(relationship.getTargetId());

                if (sourceNodeId != null && targetNodeId != null) {
                    storeRelationshipEdge(sourceNodeId, targetNodeId, relationship, tenantId);
                } else {
                    log.warn("Could not find node IDs for relationship: {}", relationship);
                }
            }

            // 5. Log the extraction
            auditService.logEntityExtraction(memory.getId(), extraction.getEntities().size(),
                    extraction.getRelationships().size(), tenantId);

            log.info("Successfully stored {} entities and {} relationships for memory {}",
                    extraction.getEntities().size(),
                    extraction.getRelationships().size(),
                    memory.getId());

        } catch (Exception e) {
            log.error("Error extracting entities from memory {}", memory.getId(), e);
        }
    }

    /**
     * Store an entity as a node in FalkorDB.
     *
     * @param entity the entity to store
     * @param tenantId the tenant ID
     * @param sourceMemoryId the ID of the memory this entity was extracted from
     * @return the generated node ID
     */
    private String storeEntityNode(ExtractedEntity entity, String tenantId, String sourceMemoryId) {
        String nodeId = "ent_" + UUID.randomUUID().toString().replace("-", "").substring(0, 12);

        // Escape strings for Cypher
        String name = escapeCypherString(entity.getName());
        String type = escapeCypherString(entity.getType());
        String escapedTenantId = escapeCypherString(tenantId);
        String escapedSourceMemoryId = escapeCypherString(sourceMemoryId);

        // Build properties string
        StringBuilder propsBuilder = new StringBuilder();
        if (entity.getProperties() != null && !entity.getProperties().isEmpty()) {
            for (Map.Entry<String, String> prop : entity.getProperties().entrySet()) {
                propsBuilder.append(", ").append(prop.getKey()).append(": '")
                        .append(escapeCypherString(prop.getValue())).append("'");
            }
        }

        // Create node with Entity label and the specific type as an additional label
        String query = String.format(
            "CREATE (e:Entity:%s {" +
            "id: '%s', " +
            "name: '%s', " +
            "type: '%s', " +
            "tenantId: '%s', " +
            "sourceMemoryId: '%s', " +
            "createdAt: %d%s" +
            "}) RETURN e.id",
            type, // Additional label based on entity type
            nodeId,
            name,
            type,
            escapedTenantId,
            escapedSourceMemoryId,
            Instant.now().toEpochMilli(),
            propsBuilder.toString()
        );

        try {
            graph.query(query);
            log.debug("Created entity node: {} ({}:{})", nodeId, type, name);

            // Create relationship from Memory to Entity
            String linkQuery = String.format(
                "MATCH (m:Memory {id: '%s'}), (e:Entity {id: '%s'}) " +
                "CREATE (m)-[r:MENTIONS {createdAt: %d}]->(e)",
                escapedSourceMemoryId,
                nodeId,
                Instant.now().toEpochMilli()
            );
            graph.query(linkQuery);

        } catch (Exception e) {
            log.error("Error creating entity node: {}", entity, e);
        }

        return nodeId;
    }

    /**
     * Store a relationship as an edge in FalkorDB.
     *
     * @param sourceNodeId the source entity node ID
     * @param targetNodeId the target entity node ID
     * @param relationship the relationship details
     * @param tenantId the tenant ID
     */
    private void storeRelationshipEdge(String sourceNodeId, String targetNodeId,
                                       ExtractedRelationship relationship, String tenantId) {
        String relType = escapeCypherString(relationship.getType());

        // Build properties string
        StringBuilder propsBuilder = new StringBuilder();
        propsBuilder.append("tenantId: '").append(escapeCypherString(tenantId)).append("'");
        propsBuilder.append(", createdAt: ").append(Instant.now().toEpochMilli());

        if (relationship.getProperties() != null && !relationship.getProperties().isEmpty()) {
            for (Map.Entry<String, String> prop : relationship.getProperties().entrySet()) {
                propsBuilder.append(", ").append(prop.getKey()).append(": '")
                        .append(escapeCypherString(prop.getValue())).append("'");
            }
        }

        // Create relationship edge
        String query = String.format(
            "MATCH (source:Entity {id: '%s'}), (target:Entity {id: '%s'}) " +
            "CREATE (source)-[r:%s {%s}]->(target)",
            sourceNodeId,
            targetNodeId,
            relType,
            propsBuilder.toString()
        );

        try {
            graph.query(query);
            log.debug("Created relationship: ({})−[{}]->({})", sourceNodeId, relType, targetNodeId);
        } catch (Exception e) {
            log.error("Error creating relationship edge: {}", relationship, e);
        }
    }

    /**
     * Find all entities extracted from a specific memory.
     *
     * @param memoryId the memory ID
     * @param tenantId the tenant ID
     * @return the entities found
     */
    public ResultSet findEntitiesByMemory(String memoryId, String tenantId) {
        if (graph == null) {
            log.warn("FalkorDB graph not available");
            return null;
        }

        String query = String.format(
            "MATCH (m:Memory {id: '%s'})-[:MENTIONS]->(e:Entity) " +
            "WHERE e.tenantId = '%s' " +
            "RETURN e.id as id, e.name as name, e.type as type, e.properties as properties",
            escapeCypherString(memoryId),
            escapeCypherString(tenantId)
        );

        try {
            return graph.query(query);
        } catch (Exception e) {
            log.error("Error finding entities for memory {}", memoryId, e);
            return null;
        }
    }

    /**
     * Find all relationships between entities from a specific memory.
     *
     * @param memoryId the memory ID
     * @param tenantId the tenant ID
     * @return the relationships found
     */
    public ResultSet findRelationshipsByMemory(String memoryId, String tenantId) {
        if (graph == null) {
            log.warn("FalkorDB graph not available");
            return null;
        }

        String query = String.format(
            "MATCH (m:Memory {id: '%s'})-[:MENTIONS]->(source:Entity)-[r]->(target:Entity) " +
            "WHERE source.tenantId = '%s' AND target.tenantId = '%s' " +
            "RETURN source.name as sourceName, source.type as sourceType, " +
            "type(r) as relationshipType, target.name as targetName, target.type as targetType",
            escapeCypherString(memoryId),
            escapeCypherString(tenantId),
            escapeCypherString(tenantId)
        );

        try {
            return graph.query(query);
        } catch (Exception e) {
            log.error("Error finding relationships for memory {}", memoryId, e);
            return null;
        }
    }

    /**
     * Search for entities by name or type.
     *
     * @param searchTerm the search term
     * @param tenantId the tenant ID
     * @param limit maximum results
     * @return matching entities
     */
    public ResultSet searchEntities(String searchTerm, String tenantId, int limit) {
        if (graph == null) {
            log.warn("FalkorDB graph not available");
            return null;
        }

        String escapedTerm = escapeCypherString(searchTerm);
        String query = String.format(
            "MATCH (e:Entity) " +
            "WHERE e.tenantId = '%s' AND (e.name CONTAINS '%s' OR e.type CONTAINS '%s') " +
            "RETURN e.id as id, e.name as name, e.type as type, e.sourceMemoryId as sourceMemoryId " +
            "ORDER BY e.name " +
            "LIMIT %d",
            escapeCypherString(tenantId),
            escapedTerm,
            escapedTerm,
            limit
        );

        try {
            return graph.query(query);
        } catch (Exception e) {
            log.error("Error searching entities: {}", searchTerm, e);
            return null;
        }
    }

    /**
     * Get the knowledge graph for a tenant.
     * Returns all entities and relationships.
     *
     * @param tenantId the tenant ID
     * @param limit maximum entities to return
     * @return the graph data
     */
    public ResultSet getKnowledgeGraph(String tenantId, int limit) {
        if (graph == null) {
            log.warn("FalkorDB graph not available");
            return null;
        }

        String query = String.format(
            "MATCH (source:Entity)-[r]->(target:Entity) " +
            "WHERE source.tenantId = '%s' " +
            "RETURN source.id as sourceId, source.name as sourceName, source.type as sourceType, " +
            "type(r) as relationshipType, " +
            "target.id as targetId, target.name as targetName, target.type as targetType " +
            "LIMIT %d",
            escapeCypherString(tenantId),
            limit
        );

        try {
            return graph.query(query);
        } catch (Exception e) {
            log.error("Error getting knowledge graph for tenant {}", tenantId, e);
            return null;
        }
    }

    /**
     * Get the knowledge graph as a response DTO for API/frontend.
     *
     * @param tenantId the tenant ID
     * @param limit maximum entities to return
     * @return KnowledgeGraphResponse with nodes and edges
     */
    public com.integraltech.brainsentry.dto.response.KnowledgeGraphResponse getKnowledgeGraphResponse(String tenantId, int limit) {
        if (graph == null) {
            log.warn("FalkorDB graph not available");
            return com.integraltech.brainsentry.dto.response.KnowledgeGraphResponse.builder()
                    .nodes(java.util.Collections.emptyList())
                    .edges(java.util.Collections.emptyList())
                    .totalNodes(0)
                    .totalEdges(0)
                    .build();
        }

        java.util.List<com.integraltech.brainsentry.dto.response.KnowledgeGraphResponse.EntityNode> nodes = new java.util.ArrayList<>();
        java.util.List<com.integraltech.brainsentry.dto.response.KnowledgeGraphResponse.EntityEdge> edges = new java.util.ArrayList<>();

        try {
            // Get all entity nodes
            String nodeQuery = String.format(
                    "MATCH (e:Entity) WHERE e.tenantId = '%s' " +
                            "RETURN e.id as id, e.name as name, e.type as type, e.sourceMemoryId as sourceMemoryId " +
                            "LIMIT %d",
                    escapeCypherString(tenantId),
                    limit
            );

            ResultSet nodeResult = graph.query(nodeQuery);
            for (var record : nodeResult) {
                nodes.add(com.integraltech.brainsentry.dto.response.KnowledgeGraphResponse.EntityNode.builder()
                        .id((String) record.getValue("id"))
                        .name((String) record.getValue("name"))
                        .type((String) record.getValue("type"))
                        .sourceMemoryId((String) record.getValue("sourceMemoryId"))
                        .build());
            }

            // Get all relationship edges
            String edgeQuery = String.format(
                    "MATCH (source:Entity)-[r]->(target:Entity) " +
                            "WHERE source.tenantId = '%s' AND NOT type(r) = 'MENTIONS' " +
                            "RETURN source.id as sourceId, source.name as sourceName, " +
                            "target.id as targetId, target.name as targetName, type(r) as type " +
                            "LIMIT %d",
                    escapeCypherString(tenantId),
                    limit * 2
            );

            ResultSet edgeResult = graph.query(edgeQuery);
            int edgeCount = 0;
            for (var record : edgeResult) {
                String sourceId = (String) record.getValue("sourceId");
                String targetId = (String) record.getValue("targetId");
                edges.add(com.integraltech.brainsentry.dto.response.KnowledgeGraphResponse.EntityEdge.builder()
                        .id(sourceId + "-" + targetId + "-" + edgeCount++)
                        .sourceId(sourceId)
                        .targetId(targetId)
                        .sourceName((String) record.getValue("sourceName"))
                        .targetName((String) record.getValue("targetName"))
                        .type((String) record.getValue("type"))
                        .build());
            }

            log.info("Retrieved knowledge graph: {} nodes, {} edges", nodes.size(), edges.size());

        } catch (Exception e) {
            log.error("Error getting knowledge graph", e);
        }

        return com.integraltech.brainsentry.dto.response.KnowledgeGraphResponse.builder()
                .nodes(nodes)
                .edges(edges)
                .totalNodes(nodes.size())
                .totalEdges(edges.size())
                .build();
    }

    /**
     * Escape strings for Cypher queries.
     */
    private String escapeCypherString(String input) {
        if (input == null) {
            return "";
        }
        return input
            .replace("\\", "\\\\")
            .replace("'", "\\'")
            .replace("\n", "\\n")
            .replace("\r", "\\r")
            .replace("\t", "\\t")
            .replace("\u0000", "\\u0000");
    }
}
