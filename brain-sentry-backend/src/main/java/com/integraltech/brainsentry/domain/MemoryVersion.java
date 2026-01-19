package com.integraltech.brainsentry.domain;

import com.integraltech.brainsentry.domain.enums.ImportanceLevel;
import com.integraltech.brainsentry.domain.enums.MemoryCategory;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;
import java.util.List;
import java.util.Map;

/**
 * Historical version of a memory.
 *
 * When a memory is updated, the previous version is archived
 * to support rollback and audit trails.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class MemoryVersion {

    /**
     * Unique identifier for this version entry.
     */
    private String id;

    /**
     * ID of the memory this version belongs to.
     */
    private String memoryId;

    /**
     * Version number (starts at 1, increments on update).
     */
    private Integer version;

    // ==================== Memory State at Version ====================

    /**
     * Content of the memory at this version.
     */
    private String content;

    /**
     * Summary of the memory at this version.
     */
    private String summary;

    /**
     * Category at this version.
     */
    private MemoryCategory category;

    /**
     * Importance level at this version.
     */
    private ImportanceLevel importance;

    /**
     * Metadata at this version.
     */
    private Map<String, Object> metadata;

    /**
     * Tags at this version.
     */
    private List<String> tags;

    /**
     * Code example at this version.
     */
    private String codeExample;

    // ==================== Change Info ====================

    /**
     * User who made this change.
     */
    private String changedBy;

    /**
     * Reason for the change.
     */
    private String changeReason;

    /**
     * Type of change.
     * Values: "create", "update", "auto_learned", "import"
     */
    private String changeType;

    // ==================== Timestamps ====================

    /**
     * When this version was created.
     */
    private Instant createdAt;

    // ==================== Tenant Support ====================

    /**
     * Tenant ID for multi-tenancy support.
     */
    private String tenantId;
}
