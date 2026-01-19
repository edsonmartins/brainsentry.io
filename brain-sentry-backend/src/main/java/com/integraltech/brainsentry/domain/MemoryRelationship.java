package com.integraltech.brainsentry.domain;

import com.integraltech.brainsentry.domain.enums.RelationshipType;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;

/**
 * Represents a relationship between two memories.
 *
 * Relationships are first-class citizens in Brain Sentry's graph-based
 * memory system, enabling context expansion through related concepts.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class MemoryRelationship {

    /**
     * Unique identifier for this relationship.
     */
    private String id;

    /**
     * ID of the source memory (from-node).
     */
    private String fromMemoryId;

    /**
     * ID of the target memory (to-node).
     */
    private String toMemoryId;

    /**
     * Type of relationship (USED_WITH, CONFLICTS_WITH, etc.)
     */
    private RelationshipType type;

    // ==================== Metadata ====================

    /**
     * How frequently these memories appear together.
     * Higher values indicate stronger association.
     */
    @Builder.Default
    private Integer frequency = 1;

    /**
     * Severity level for conflicts or important relationships.
     * Values: "high", "medium", "low", null
     */
    private String severity;

    /**
     * Strength of this relationship (0.0 to 1.0).
     * Computed from frequency, recency, and user feedback.
     */
    private Double strength;

    /**
     * Optional description explaining this relationship.
     */
    private String description;

    // ==================== Timestamps ====================

    /**
     * When this relationship was created.
     */
    private Instant createdAt;

    /**
     * When this relationship was last used/observed.
     */
    private Instant lastUsedAt;

    // ==================== Tenant Support ====================

    /**
     * Tenant ID for multi-tenancy support.
     */
    private String tenantId;
}
