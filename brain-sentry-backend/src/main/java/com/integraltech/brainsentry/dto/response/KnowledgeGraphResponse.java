package com.integraltech.brainsentry.dto.response;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;
import java.util.Map;

/**
 * Response DTO for the knowledge graph visualization.
 * Contains entities (nodes) and relationships (edges) extracted from memories.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class KnowledgeGraphResponse {

    /**
     * List of entity nodes in the graph.
     */
    private List<EntityNode> nodes;

    /**
     * List of relationship edges in the graph.
     */
    private List<EntityEdge> edges;

    /**
     * Total count of entities.
     */
    private int totalNodes;

    /**
     * Total count of relationships.
     */
    private int totalEdges;

    /**
     * Represents an entity node in the knowledge graph.
     */
    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class EntityNode {
        /**
         * Unique identifier of the entity.
         */
        private String id;

        /**
         * Name/label of the entity.
         */
        private String name;

        /**
         * Type of the entity (CLIENTE, VENDEDOR, PRODUTO, etc.).
         */
        private String type;

        /**
         * ID of the memory this entity was extracted from.
         */
        private String sourceMemoryId;

        /**
         * Additional properties of the entity.
         */
        private Map<String, String> properties;
    }

    /**
     * Represents a relationship edge in the knowledge graph.
     */
    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class EntityEdge {
        /**
         * Unique identifier of the edge.
         */
        private String id;

        /**
         * ID of the source entity.
         */
        private String sourceId;

        /**
         * ID of the target entity.
         */
        private String targetId;

        /**
         * Name of the source entity.
         */
        private String sourceName;

        /**
         * Name of the target entity.
         */
        private String targetName;

        /**
         * Type of the relationship (REALIZOU, ATENDEU, CONTEM, etc.).
         */
        private String type;

        /**
         * Additional properties of the relationship.
         */
        private Map<String, String> properties;
    }
}
