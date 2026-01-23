package com.integraltech.brainsentry.repository.impl;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.falkordb.Driver;
import com.falkordb.FalkorDB;
import com.falkordb.Graph;
import com.falkordb.Record;
import com.falkordb.ResultSet;
import com.falkordb.graph_entities.Node;
import com.falkordb.graph_entities.Property;
import com.integraltech.brainsentry.domain.Memory;
import com.integraltech.brainsentry.repository.MemoryRepository;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.autoconfigure.condition.ConditionalOnBean;
import org.springframework.stereotype.Repository;
import redis.clients.jedis.JedisPool;

import java.time.Instant;
import java.util.*;
import java.util.stream.Collectors;

/**
 * FalkorDB implementation of MemoryRepository.
 *
 * Uses the official JFalkorDB Java client for:
 * - Storing memories as nodes with properties
 * - Creating relationships between related memories
 * - Vector similarity search for semantic queries
 *
 * Requires FalkorDB server running with graph module enabled.
 */
@Repository
@ConditionalOnBean(JedisPool.class)
@Slf4j
public class MemoryRepositoryImpl implements MemoryRepository {

    private final JedisPool jedisPool;
    private final ObjectMapper objectMapper;
    private final Graph graph;
    private final String graphName;

    private static final String MEMORY_PREFIX = "memory:";
    private static final String TAG_INDEX = "tag_idx:";
    private static final String TENANT_MEMORIES = "tenant_memories:";

    public MemoryRepositoryImpl(JedisPool jedisPool, ObjectMapper objectMapper,
                                @Value("${brain-sentry.graph.name:brainsentry}") String graphName,
                                @Value("${brain-sentry.redis.host:localhost}") String host,
                                @Value("${brain-sentry.redis.port:6379}") int port,
                                @Value("${brain-sentry.redis.password:}") String password) {
        this.jedisPool = jedisPool;
        this.objectMapper = objectMapper;
        this.graphName = graphName;

        // Create FalkorDB connection
        // API: FalkorDB.driver(host, port) or FalkorDB.driver(host, port, password, username)
        Driver driver;
        if (password != null && !password.isEmpty()) {
            // With password: driver(host, port, password, username)
            // Note: For FalkorDB with ACL enabled, try null username first
            try {
                driver = FalkorDB.driver(host, port, password, null);
            } catch (Exception e) {
                log.warn("Failed to connect with null username, trying 'default': {}", e.getMessage());
                driver = FalkorDB.driver(host, port, password, "default");
            }
        } else {
            // Without password
            driver = FalkorDB.driver(host, port);
        }
        this.graph = driver.graph(graphName);

        log.info("FalkorDB graph '{}' initialized with connection to {}:{}", graphName, host, port);
    }

    @Override
    public Memory save(Memory memory) {
        try {
            // Generate ID if not present
            if (memory.getId() == null) {
                memory.setId(generateId());
            }

            // Set timestamps
            if (memory.getCreatedAt() == null) {
                memory.setCreatedAt(Instant.now());
            }
            memory.setUpdatedAt(Instant.now());

            // Store in Redis KV for fast access
            try (var jedis = jedisPool.getResource()) {
                String key = MEMORY_PREFIX + memory.getId();
                String json = objectMapper.writeValueAsString(memory);
                jedis.set(key, json);

                // Add to tenant index
                jedis.sadd(TENANT_MEMORIES + memory.getTenantId(), memory.getId());

                // Index tags
                if (memory.getTags() != null) {
                    for (String tag : memory.getTags()) {
                        jedis.sadd(TAG_INDEX + tag, memory.getId());
                    }
                }
            }

            // Store in FalkorDB graph
            saveToGraph(memory);

            // Create relationships
            createRelationships(memory);

            log.debug("Saved memory: {}", memory.getId());
            return memory;
        } catch (Exception e) {
            log.error("Error saving memory: {}", memory.getId(), e);
            throw new RuntimeException("Failed to save memory", e);
        }
    }

    /**
     * Save memory as a node in FalkorDB graph with embeddings.
     */
    private void saveToGraph(Memory memory) {
        try {
            // Convert embedding to array format
            String embeddingStr = "[]";
            if (memory.getEmbedding() != null && memory.getEmbedding().length > 0) {
                StringBuilder sb = new StringBuilder("[");
                for (int i = 0; i < memory.getEmbedding().length; i++) {
                    if (i > 0) sb.append(",");
                    sb.append(memory.getEmbedding()[i]);
                }
                sb.append("]");
                embeddingStr = sb.toString();
            }

            // Escape strings for Cypher
            String id = escapeCypherString(memory.getId());
            String content = escapeCypherString(memory.getContent());
            String summary = memory.getSummary() != null ? escapeCypherString(memory.getSummary()) : "";
            String category = memory.getCategory() != null ? memory.getCategory().name() : "PATTERN";
            String importance = memory.getImportance() != null ? memory.getImportance().name() : "MINOR";
            String tenantId = memory.getTenantId() != null ? memory.getTenantId() : "default";

            // Build tag list as Cypher array
            String tagList = "[]";
            if (memory.getTags() != null && !memory.getTags().isEmpty()) {
                tagList = memory.getTags().stream()
                    .map(t -> "'" + escapeCypherString(t) + "'")
                    .collect(Collectors.joining(",", "[", "]"));
            }

            String query = String.format(
                "MERGE (m:Memory {id: '%s'}) " +
                "SET m.content = '%s', " +
                "m.summary = '%s', " +
                "m.category = '%s', " +
                "m.importance = '%s', " +
                "m.tenantId = '%s', " +
                "m.tags = %s, " +
                "m.embedding = %s, " +
                "m.createdAt = %d, " +
                "m.updatedAt = %d, " +
                "m.accessCount = %d, " +
                "m.version = %d",
                id, content, summary, category, importance, tenantId, tagList, embeddingStr,
                memory.getCreatedAt().toEpochMilli(),
                memory.getUpdatedAt() != null ? memory.getUpdatedAt().toEpochMilli() : Instant.now().toEpochMilli(),
                memory.getAccessCount() != null ? memory.getAccessCount() : 0,
                memory.getVersion() != null ? memory.getVersion() : 1
            );

            graph.query(query);
            log.debug("Saved to graph: {}", memory.getId());
        } catch (Exception e) {
            log.warn("Could not save to graph: {}", e.getMessage());
        }
    }

    /**
     * Create relationships with similar memories based on tags.
     */
    private void createRelationships(Memory memory) {
        try {
            String memoryId = escapeCypherString(memory.getId());
            String tenantId = memory.getTenantId() != null ? memory.getTenantId() : "default";

            // Create relationships with memories that share tags
            if (memory.getTags() != null && !memory.getTags().isEmpty()) {
                for (String tag : memory.getTags()) {
                    String escapedTag = escapeCypherString(tag);
                    String query = String.format(
                        "MATCH (m1:Memory {id: '%s'}), (m2:Memory) " +
                        "WHERE m2.tenantId = '%s' AND '%s' IN m2.tags AND m1.id <> m2.id " +
                        "MERGE (m1)-[r:RELATED_TO]->(m2) " +
                        "SET r.strength = coalesce(r.strength, 0) + 1, " +
                        "r.type = 'shared_tag', " +
                        "r.tag = '%s', " +
                        "r.updatedAt = %d",
                        memoryId, tenantId, escapedTag, escapedTag, Instant.now().toEpochMilli()
                    );
                    graph.query(query);
                }
            }

            // Create category relationship
            String category = memory.getCategory() != null ? memory.getCategory().name() : "PATTERN";
            String escapedCategory = escapeCypherString(category);
            String categoryQuery = String.format(
                "MERGE (c:Category {name: '%s'}) " +
                "WITH c " +
                "MATCH (m:Memory {id: '%s'}) " +
                "MERGE (m)-[r:BELONGS_TO]->(c) " +
                "ON CREATE SET r.createdAt = %d",
                escapedCategory, memoryId, Instant.now().toEpochMilli()
            );
            graph.query(categoryQuery);

            log.debug("Created relationships for memory: {}", memory.getId());
        } catch (Exception e) {
            log.warn("Could not create relationships: {}", e.getMessage());
        }
    }

    @Override
    public Optional<Memory> findById(String id) {
        try (var jedis = jedisPool.getResource()) {
            String key = MEMORY_PREFIX + id;
            String json = jedis.get(key);

            if (json != null) {
                Memory memory = objectMapper.readValue(json, Memory.class);
                return Optional.of(memory);
            }
        } catch (Exception e) {
            log.trace("KV lookup failed for {}: {}", id, e.getMessage());
        }

        // Try to fetch from graph
        return findByIdFromGraph(id);
    }

    private Optional<Memory> findByIdFromGraph(String id) {
        try {
            String escapedId = escapeCypherString(id);
            String query = String.format(
                "MATCH (m:Memory {id: '%s'}) RETURN m",
                escapedId
            );

            ResultSet resultSet = graph.query(query);
            for (Record record : resultSet) {
                Node node = record.getValue("m");
                return Optional.of(nodeToMemory(node));
            }

            return Optional.empty();
        } catch (Exception e) {
            log.trace("Graph lookup failed for {}: {}", id, e.getMessage());
            return Optional.empty();
        }
    }

    @Override
    public List<Memory> findByTenantId(String tenantId) {
        try (var jedis = jedisPool.getResource()) {
            Set<String> memoryIds = jedis.smembers(TENANT_MEMORIES + tenantId);
            List<Memory> memories = new ArrayList<>();

            for (String id : memoryIds) {
                findById(id).ifPresent(memories::add);
            }

            return memories;
        } catch (Exception e) {
            log.error("Error finding memories by tenant: {}", tenantId, e);
            return List.of();
        }
    }

    @Override
    public List<Memory> findByCategory(String category, String tenantId) {
        try {
            String query = String.format(
                "MATCH (m:Memory) " +
                "WHERE m.tenantId = '%s' AND m.category = '%s' " +
                "RETURN m " +
                "ORDER BY m.createdAt DESC",
                tenantId, category
            );

            List<Memory> result = queryMemories(query);
            if (!result.isEmpty()) {
                return result;
            }
        } catch (Exception e) {
            log.debug("Graph query failed for category {}: {}", category, e.getMessage());
        }

        // Fallback to Redis-based search
        try {
            return findByTenantId(tenantId).stream()
                .filter(m -> m.getCategory() != null && m.getCategory().name().equals(category))
                .sorted((a, b) -> {
                    Instant aTime = a.getCreatedAt() != null ? a.getCreatedAt() : Instant.MIN;
                    Instant bTime = b.getCreatedAt() != null ? b.getCreatedAt() : Instant.MIN;
                    return bTime.compareTo(aTime);
                })
                .collect(Collectors.toList());
        } catch (Exception e) {
            log.error("Error finding memories by category: {}", category, e);
            return List.of();
        }
    }

    @Override
    public List<Memory> findByImportance(String importance, String tenantId) {
        try {
            String query = String.format(
                "MATCH (m:Memory) " +
                "WHERE m.tenantId = '%s' AND m.importance = '%s' " +
                "RETURN m " +
                "ORDER BY m.accessCount DESC",
                tenantId, importance
            );

            List<Memory> result = queryMemories(query);
            if (!result.isEmpty()) {
                return result;
            }
        } catch (Exception e) {
            log.debug("Graph query failed for importance {}: {}", importance, e.getMessage());
        }

        // Fallback to Redis-based search
        try {
            return findByTenantId(tenantId).stream()
                .filter(m -> m.getImportance() != null && m.getImportance().name().equals(importance))
                .sorted((a, b) -> {
                    int aCount = a.getAccessCount() != null ? a.getAccessCount() : 0;
                    int bCount = b.getAccessCount() != null ? b.getAccessCount() : 0;
                    return Integer.compare(bCount, aCount);
                })
                .collect(Collectors.toList());
        } catch (Exception e) {
            log.error("Error finding memories by importance: {}", importance, e);
            return List.of();
        }
    }

    @Override
    public List<Memory> findByTags(List<String> tags, String tenantId) {
        try (var jedis = jedisPool.getResource()) {
            Set<String> memoryIds = null;

            for (String tag : tags) {
                Set<String> tagMembers = jedis.smembers(TAG_INDEX + tag);
                if (memoryIds == null) {
                    memoryIds = new HashSet<>(tagMembers);
                } else {
                    memoryIds.retainAll(tagMembers);  // Intersection
                }
            }

            if (memoryIds == null || memoryIds.isEmpty()) {
                return List.of();
            }

            List<Memory> memories = new ArrayList<>();
            for (String id : memoryIds) {
                findById(id).ifPresent(mem -> {
                    if (tenantId.equals(mem.getTenantId())) {
                        memories.add(mem);
                    }
                });
            }

            return memories;
        } catch (Exception e) {
            log.error("Error finding memories by tags: {}", tags, e);
            return List.of();
        }
    }

    @Override
    public List<Memory> vectorSearch(float[] embedding, int limit, String tenantId) {
        try {
            // Convert embedding to string format
            StringBuilder sb = new StringBuilder("[");
            for (int i = 0; i < embedding.length; i++) {
                if (i > 0) sb.append(",");
                sb.append(embedding[i]);
            }
            sb.append("]");
            String embeddingStr = sb.toString();

            // Use FalkorDB's vector.similarity function
            String query = String.format(
                "CALL vector.similarity(" +
                "  $embeddings, " +
                "  '%s', " +
                "  3, " +
                "  {tenantId: '%s'}" +
                ") YIELD node, score " +
                "RETURN node.id as id, score " +
                "ORDER BY score DESC " +
                "LIMIT %d",
                embeddingStr, tenantId, limit
            );

            ResultSet resultSet = graph.query(query);
            List<Memory> memories = new ArrayList<>();

            for (Record record : resultSet) {
                String id = record.getValue("id");
                if (id != null) {
                    findById(id).ifPresent(memories::add);
                }
            }

            if (!memories.isEmpty()) {
                log.debug("Vector search returned {} results", memories.size());
                return memories;
            }

            // Fallback: return most accessed memories
            return findByTenantId(tenantId).stream()
                .sorted((a, b) -> Integer.compare(
                    b.getAccessCount() != null ? b.getAccessCount() : 0,
                    a.getAccessCount() != null ? a.getAccessCount() : 0
                ))
                .limit(limit)
                .toList();
        } catch (Exception e) {
            log.debug("Vector search failed: {}", e.getMessage());

            // Fallback
            return findByTenantId(tenantId).stream()
                .sorted((a, b) -> Integer.compare(
                    b.getAccessCount() != null ? b.getAccessCount() : 0,
                    a.getAccessCount() != null ? a.getAccessCount() : 0
                ))
                .limit(limit)
                .toList();
        }
    }

    @Override
    public List<Memory> findRelated(String memoryId, int depth, String tenantId) {
        try {
            String id = escapeCypherString(memoryId);

            // Query for related memories through graph relationships
            String query = String.format(
                "MATCH (m:Memory {id: '%s'})-[r:RELATED_TO*1..%d]-(related:Memory) " +
                "WHERE related.tenantId = '%s' " +
                "RETURN DISTINCT related.id as id, " +
                "count(r) as relationshipCount " +
                "ORDER BY relationshipCount DESC " +
                "LIMIT %d",
                id, depth, tenantId, depth * 5
            );

            ResultSet resultSet = graph.query(query);
            List<Memory> memories = new ArrayList<>();

            for (Record record : resultSet) {
                String relatedId = record.getValue("id");
                if (relatedId != null) {
                    findById(relatedId).ifPresent(memories::add);
                }
            }

            if (!memories.isEmpty()) {
                log.debug("Found {} related memories", memories.size());
                return memories;
            }

            // Fallback: use tag-based matching
            return findById(memoryId)
                .map(memory -> {
                    if (memory.getTags() == null || memory.getTags().isEmpty()) {
                        return findByTenantId(tenantId).stream()
                            .filter(m -> !m.getId().equals(memoryId))
                            .sorted((a, b) -> Integer.compare(
                                b.getAccessCount() != null ? b.getAccessCount() : 0,
                                a.getAccessCount() != null ? a.getAccessCount() : 0
                            ))
                            .limit(5)
                            .toList();
                    }
                    return findByTags(memory.getTags(), tenantId).stream()
                        .filter(m -> !m.getId().equals(memoryId))
                        .limit(depth * 5)
                        .toList();
                })
                .orElse(List.of());
        } catch (Exception e) {
            log.error("Error finding related memories for: {}", memoryId, e);
            return List.of();
        }
    }

    @Override
    public boolean deleteById(String id) {
        try {
            Optional<Memory> memoryOpt = findById(id);
            if (memoryOpt.isEmpty()) {
                return false;
            }

            Memory memory = memoryOpt.get();

            // Remove from graph
            try {
                String escapedId = escapeCypherString(id);
                String graphDelete = String.format(
                    "MATCH (m:Memory {id: '%s'}) DETACH DELETE m",
                    escapedId
                );
                graph.query(graphDelete);
            } catch (Exception e) {
                log.trace("Graph delete: {}", e.getMessage());
            }

            // Remove from KV store
            try (var jedis = jedisPool.getResource()) {
                String key = MEMORY_PREFIX + id;
                jedis.del(key);

                // Remove from tenant index
                String tenantId = memory.getTenantId();
                if (tenantId != null) {
                    jedis.srem(TENANT_MEMORIES + tenantId, id);
                }

                // Remove tag indexes
                if (memory.getTags() != null) {
                    for (String tag : memory.getTags()) {
                        jedis.srem(TAG_INDEX + tag, id);
                    }
                }
            }

            log.debug("Deleted memory: {}", id);
            return true;
        } catch (Exception e) {
            log.error("Error deleting memory: {}", id, e);
            return false;
        }
    }

    @Override
    public long countByTenantId(String tenantId) {
        try (var jedis = jedisPool.getResource()) {
            Long count = jedis.scard(TENANT_MEMORIES + tenantId);
            return count != null ? count : 0L;
        } catch (Exception e) {
            log.error("Error counting memories for tenant: {}", tenantId, e);
            return 0;
        }
    }

    @Override
    public void archiveVersion(Memory memory) {
        log.debug("Archive version for memory: {}", memory.getId());
        // Version archiving is handled by PostgreSQL
    }

    @Override
    public List<Integer> getVersionHistory(String memoryId, String tenantId) {
        return findById(memoryId)
            .map(m -> List.of(m.getVersion()))
            .orElse(List.of());
    }

    @Override
    public com.falkordb.ResultSet query(String query) {
        try {
            return graph.query(query);
        } catch (Exception e) {
            log.error("Error executing query: {}", e.getMessage());
            throw new RuntimeException("Query execution failed", e);
        }
    }

    @Override
    public void createAllRelationships(String tenantId) {
        try {
            String escapedTenantId = escapeCypherString(tenantId);

            // Clear existing RELATED_TO relationships for this tenant
            String deleteQuery = String.format(
                "MATCH (m1:Memory)-[r:RELATED_TO]->(m2:Memory) " +
                "WHERE m1.tenantId = '%s' " +
                "DELETE r",
                escapedTenantId
            );
            graph.query(deleteQuery);
            log.info("Cleared existing relationships for tenant: {}", tenantId);

            // First, let's test with a simpler query - find pairs with at least one overlapping tag
            // Using a simpler approach that should work with FalkorDB
            long timestamp = Instant.now().toEpochMilli();

            // Step 1: Create relationships in one direction
            String relationshipQuery = String.format(
                "MATCH (m1:Memory), (m2:Memory) " +
                "WHERE m1.tenantId = '%s' AND m2.tenantId = '%s' " +
                "AND m1.id < m2.id " +
                "AND EXISTS (" +
                "    SELECT tag IN unwind(m1.tags) AS tag " +
                "    WHERE tag IN m2.tags" +
                ") " +
                "WITH m1, m2 LIMIT 100 " +
                "RETURN m1.id as id1, m2.id as id2",
                escapedTenantId, escapedTenantId
            );

            ResultSet testResult = graph.query(relationshipQuery);
            log.info("Test query found {} memory pairs with shared tags", testResult.size());

            // Now use a simpler query that should work
            String createQuery = String.format(
                "MATCH (m1:Memory) " +
                "WHERE m1.tenantId = '%s' " +
                "UNWIND m1.tags AS tag1 " +
                "MATCH (m2:Memory) " +
                "WHERE m2.tenantId = '%s' " +
                "AND m1.id < m2.id " +
                "AND tag1 IN m2.tags " +
                "WITH m1, m2, tag1 " +
                "ORDER BY m1.id, m2.id, tag1 " +
                "WITH m1, m2, collect(DISTINCT tag1)[0] as sharedTag " +
                "CREATE (m1)-[r1:RELATED_TO]->(m2) " +
                "CREATE (m2)-[r2:RELATED_TO]->(m1) " +
                "SET r1.type = 'shared_tag', r1.tag = sharedTag, r1.strength = 1, r1.updatedAt = %d, " +
                "r2.type = 'shared_tag', r2.tag = sharedTag, r2.strength = 1, r2.updatedAt = %d",
                escapedTenantId, escapedTenantId, timestamp, timestamp
            );
            ResultSet relResult = graph.query(createQuery);
            log.info("Relationship creation result size: {}", relResult.size());

            log.info("Created all graph relationships for tenant: {}", tenantId);
        } catch (Exception e) {
            log.error("Error creating all relationships for tenant {}: {}", tenantId, e.getMessage(), e);
            throw new RuntimeException("Failed to create all relationships", e);
        }
    }

    // ==================== Private Methods ====================

    /**
     * Query memories and convert to Memory objects.
     */
    private List<Memory> queryMemories(String query) {
        try {
            ResultSet resultSet = graph.query(query);
            List<Memory> memories = new ArrayList<>();

            for (Record record : resultSet) {
                Node node = record.getValue("m");
                if (node != null) {
                    memories.add(nodeToMemory(node));
                }
            }

            return memories;
        } catch (Exception e) {
            log.error("Error querying memories: {}", e.getMessage());
            return List.of();
        }
    }

    /**
     * Convert a FalkorDB Node to a Memory object using the correct API.
     */
    private Memory nodeToMemory(Node node) {
        Memory memory = new Memory();

        // Use getProperty(String) which returns a Property object
        memory.setId(getStringProperty(node, "id"));
        memory.setContent(getStringProperty(node, "content"));
        memory.setSummary(getStringProperty(node, "summary"));

        String categoryStr = getStringProperty(node, "category");
        if (categoryStr != null) {
            try {
                memory.setCategory(com.integraltech.brainsentry.domain.enums.MemoryCategory.valueOf(categoryStr));
            } catch (IllegalArgumentException e) {
                memory.setCategory(com.integraltech.brainsentry.domain.enums.MemoryCategory.PATTERN);
            }
        }

        String importanceStr = getStringProperty(node, "importance");
        if (importanceStr != null) {
            try {
                memory.setImportance(com.integraltech.brainsentry.domain.enums.ImportanceLevel.valueOf(importanceStr));
            } catch (IllegalArgumentException e) {
                memory.setImportance(com.integraltech.brainsentry.domain.enums.ImportanceLevel.MINOR);
            }
        }

        memory.setTenantId(getStringProperty(node, "tenantId"));
        memory.setCreatedBy(getStringProperty(node, "createdBy"));

        // Timestamps
        Long createdAt = getLongProperty(node, "createdAt");
        if (createdAt != null) {
            memory.setCreatedAt(Instant.ofEpochMilli(createdAt));
        }

        Long updatedAt = getLongProperty(node, "updatedAt");
        if (updatedAt != null) {
            memory.setUpdatedAt(Instant.ofEpochMilli(updatedAt));
        }

        // Counts
        memory.setAccessCount(getIntegerProperty(node, "accessCount"));
        memory.setInjectionCount(getIntegerProperty(node, "injectionCount"));
        memory.setHelpfulCount(getIntegerProperty(node, "helpfulCount"));
        memory.setNotHelpfulCount(getIntegerProperty(node, "notHelpfulCount"));
        memory.setVersion(getIntegerProperty(node, "version"));

        // Tags - stored as a list in the node
        Property tagsProp = node.getProperty("tags");
        if (tagsProp != null && tagsProp.getValue() instanceof List) {
            memory.setTags((List<String>) tagsProp.getValue());
        }

        // Embedding - stored as a list in the node
        Property embeddingProp = node.getProperty("embedding");
        if (embeddingProp != null && embeddingProp.getValue() instanceof List) {
            List<?> embeddingList = (List<?>) embeddingProp.getValue();
            if (!embeddingList.isEmpty()) {
                float[] embeddingArray = new float[embeddingList.size()];
                for (int i = 0; i < embeddingList.size(); i++) {
                    Object val = embeddingList.get(i);
                    if (val instanceof Number) {
                        embeddingArray[i] = ((Number) val).floatValue();
                    }
                }
                memory.setEmbedding(embeddingArray);
            }
        }

        return memory;
    }

    /**
     * Helper to get a String property from a Node.
     */
    private String getStringProperty(Node node, String name) {
        Property prop = node.getProperty(name);
        if (prop != null && prop.getValue() != null) {
            return prop.getValue().toString();
        }
        return null;
    }

    /**
     * Helper to get a Long property from a Node.
     */
    private Long getLongProperty(Node node, String name) {
        Property prop = node.getProperty(name);
        if (prop != null && prop.getValue() instanceof Number) {
            return ((Number) prop.getValue()).longValue();
        }
        return null;
    }

    /**
     * Helper to get an Integer property from a Node.
     */
    private Integer getIntegerProperty(Node node, String name) {
        Property prop = node.getProperty(name);
        if (prop != null && prop.getValue() instanceof Number) {
            return ((Number) prop.getValue()).intValue();
        }
        return null;
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

    private String generateId() {
        return "mem_" + UUID.randomUUID().toString().replace("-", "").substring(0, 12);
    }
}
