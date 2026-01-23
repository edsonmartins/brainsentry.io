package com.integraltech.brainsentry.domain;

import com.integraltech.brainsentry.domain.enums.ImportanceLevel;
import com.integraltech.brainsentry.domain.enums.MemoryCategory;
import jakarta.persistence.*;
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
@Entity
@Table(name = "memory_versions")
public class MemoryVersion {

    @Id
    @Column(length = 100)
    private String id;

    /**
     * ID of the memory this version belongs to.
     */
    @Column(length = 100)
    private String memoryId;

    /**
     * Version number (starts at 1, increments on update).
     */
    private Integer version;

    // ==================== Memory State at Version ====================

    /**
     * Content of the memory at this version.
     */
    @Lob
    @Column(columnDefinition = "TEXT")
    private String content;

    /**
     * Summary of the memory at this version.
     */
    @Column(length = 500)
    private String summary;

    /**
     * Category at this version.
     */
    @Enumerated(EnumType.STRING)
    @Column(length = 50)
    private MemoryCategory category;

    /**
     * Importance level at this version.
     */
    @Enumerated(EnumType.STRING)
    @Column(length = 50)
    private ImportanceLevel importance;

    /**
     * Metadata at this version.
     */
    @Lob
    private Map<String, Object> metadata;

    /**
     * Tags at this version.
     */
    @ElementCollection
    @CollectionTable(name = "memory_version_tags", joinColumns = @JoinColumn(name = "memory_version_id"))
    @Column(name = "tag")
    private List<String> tags;

    /**
     * Code example at this version.
     */
    @Lob
    @Column(columnDefinition = "TEXT")
    private String codeExample;

    // ==================== Change Info ====================

    /**
     * User who made this change.
     */
    @Column(length = 100)
    private String changedBy;

    /**
     * Reason for the change.
     */
    @Lob
    @Column(columnDefinition = "TEXT")
    private String changeReason;

    /**
     * Type of change.
     * Values: "create", "update", "auto_learned", "import"
     */
    @Column(length = 50)
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
    @Column(length = 100)
    private String tenantId;
}
