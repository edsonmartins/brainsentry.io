package com.integraltech.brainsentry.domain;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.integraltech.brainsentry.domain.enums.ImportanceLevel;
import com.integraltech.brainsentry.domain.enums.MemoryCategory;
import com.integraltech.brainsentry.config.FloatArrayConverter;
import com.integraltech.brainsentry.domain.enums.ValidationStatus;
import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.TenantId;

import java.time.Instant;
import java.util.List;
import java.util.Map;

/**
 * Core memory entity in the Brain Sentry system.
 *
 * A Memory represents a unit of knowledge that can be retrieved
 * and injected into AI agent context. Memories are stored in FalkorDB
 * with vector embeddings for semantic search.
 *
 * Multi-tenancy is handled automatically by Hibernate 6 using @TenantId.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "memories", indexes = {
    @Index(name = "idx_memory_tenant", columnList = "tenantId"),
    @Index(name = "idx_memory_category", columnList = "category"),
    @Index(name = "idx_memory_importance", columnList = "importance"),
    @Index(name = "idx_memory_created_at", columnList = "createdAt")
})
public class Memory {

    @Id
    @Column(length = 100)
    private String id;

    @Column(nullable = false, columnDefinition = "TEXT")
    private String content;

    @Column(length = 500)
    private String summary;

    // ==================== Classification ====================

    @Enumerated(EnumType.STRING)
    @Column(length = 50)
    private MemoryCategory category;

    @Enumerated(EnumType.STRING)
    @Column(length = 50)
    private ImportanceLevel importance;

    @Enumerated(EnumType.STRING)
    @Column(length = 50)
    private ValidationStatus validationStatus;

    // ==================== Vector Embedding ====================

    /**
     * Vector embedding for semantic search.
     * 384 dimensions for all-MiniLM-L6-v2 model.
     * Stored as array of floats in FalkorDB, serialized in PostgreSQL.
     */
    @Convert(converter = FloatArrayConverter.class)
    @Column(columnDefinition = "bytea")
    private float[] embedding;

    // ==================== Metadata ====================

    @Lob
    @Column(columnDefinition = "jsonb")
    private Map<String, Object> metadata;

    // @Transient  // Temporarily disabled - using DB storage
    // private Map<String, Object> metadata;

    @ElementCollection
    @CollectionTable(name = "memory_tags", joinColumns = @JoinColumn(name = "memory_id"))
    @Column(name = "tag")
    private List<String> tags;

    // ==================== Provenance ====================

    @Column(length = 50)
    private String sourceType;

    @Column(length = 500)
    private String sourceReference;

    @Column(length = 100)
    private String createdBy;

    /**
     * Tenant ID for multi-tenancy support.
     * Hibernate 6 @TenantId enables automatic filtering by this field
     * in all queries without manual WHERE clauses.
     */
    @TenantId
    @Column(length = 100, nullable = false)
    private String tenantId;

    // ==================== Timestamps ====================

    @Column(nullable = false, updatable = false)
    private Instant createdAt;

    private Instant updatedAt;

    private Instant lastAccessedAt;

    // ==================== Version Control ====================

    private Integer version;

    // ==================== Usage Tracking ====================

    @Builder.Default
    private Integer accessCount = 0;

    @Builder.Default
    private Integer injectionCount = 0;

    @Builder.Default
    private Integer helpfulCount = 0;

    @Builder.Default
    private Integer notHelpfulCount = 0;

    // ==================== Code Example ====================

    @Column(columnDefinition = "TEXT")
    private String codeExample;

    @Column(length = 50)
    private String programmingLanguage;

    // ==================== Computed Fields ====================

    /**
     * Calculated helpfulness ratio (helpful / total feedback).
     */
    @Transient
    @JsonIgnore
    public Double getHelpfulnessRate() {
        int total = helpfulCount + notHelpfulCount;
        return total > 0 ? (double) helpfulCount / total : 0.0;
    }

    /**
     * Calculated relevance score based on usage and feedback.
     * Combines access frequency, injection rate, and helpfulness.
     */
    @Transient
    @JsonIgnore
    public Double getRelevanceScore() {
        double accessScore = Math.log(1 + accessCount) / 10.0;
        double injectionScore = Math.log(1 + injectionCount) / 10.0;
        double helpfulnessScore = getHelpfulnessRate();
        return (accessScore + injectionScore) * 0.3 + helpfulnessScore * 0.4;
    }

    @PrePersist
    protected void onCreate() {
        createdAt = Instant.now();
        updatedAt = Instant.now();
        if (version == null) {
            version = 1;
        }
    }

    @PreUpdate
    protected void onUpdate() {
        updatedAt = Instant.now();
    }

    // Manual getters in case Lombok doesn't generate them
    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getContent() {
        return content;
    }

    public void setContent(String content) {
        this.content = content;
    }

    public String getSummary() {
        return summary;
    }

    public void setSummary(String summary) {
        this.summary = summary;
    }

    public MemoryCategory getCategory() {
        return category;
    }

    public void setCategory(MemoryCategory category) {
        this.category = category;
    }

    public ImportanceLevel getImportance() {
        return importance;
    }

    public void setImportance(ImportanceLevel importance) {
        this.importance = importance;
    }

    public ValidationStatus getValidationStatus() {
        return validationStatus;
    }

    public void setValidationStatus(ValidationStatus validationStatus) {
        this.validationStatus = validationStatus;
    }

    public float[] getEmbedding() {
        return embedding;
    }

    public void setEmbedding(float[] embedding) {
        this.embedding = embedding;
    }

    public Map<String, Object> getMetadata() {
        return metadata;
    }

    public void setMetadata(Map<String, Object> metadata) {
        this.metadata = metadata;
    }

    public List<String> getTags() {
        return tags;
    }

    public void setTags(List<String> tags) {
        this.tags = tags;
    }

    public String getSourceType() {
        return sourceType;
    }

    public void setSourceType(String sourceType) {
        this.sourceType = sourceType;
    }

    public String getSourceReference() {
        return sourceReference;
    }

    public void setSourceReference(String sourceReference) {
        this.sourceReference = sourceReference;
    }

    public String getCreatedBy() {
        return createdBy;
    }

    public void setCreatedBy(String createdBy) {
        this.createdBy = createdBy;
    }

    public String getTenantId() {
        return tenantId;
    }

    public void setTenantId(String tenantId) {
        this.tenantId = tenantId;
    }

    public Instant getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(Instant createdAt) {
        this.createdAt = createdAt;
    }

    public Instant getUpdatedAt() {
        return updatedAt;
    }

    public void setUpdatedAt(Instant updatedAt) {
        this.updatedAt = updatedAt;
    }

    public Instant getLastAccessedAt() {
        return lastAccessedAt;
    }

    public void setLastAccessedAt(Instant lastAccessedAt) {
        this.lastAccessedAt = lastAccessedAt;
    }

    public Integer getVersion() {
        return version;
    }

    public void setVersion(Integer version) {
        this.version = version;
    }

    public Integer getAccessCount() {
        return accessCount;
    }

    public void setAccessCount(Integer accessCount) {
        this.accessCount = accessCount;
    }

    public Integer getInjectionCount() {
        return injectionCount;
    }

    public void setInjectionCount(Integer injectionCount) {
        this.injectionCount = injectionCount;
    }

    public Integer getHelpfulCount() {
        return helpfulCount;
    }

    public void setHelpfulCount(Integer helpfulCount) {
        this.helpfulCount = helpfulCount;
    }

    public Integer getNotHelpfulCount() {
        return notHelpfulCount;
    }

    public void setNotHelpfulCount(Integer notHelpfulCount) {
        this.notHelpfulCount = notHelpfulCount;
    }

    public String getCodeExample() {
        return codeExample;
    }

    public void setCodeExample(String codeExample) {
        this.codeExample = codeExample;
    }

    public String getProgrammingLanguage() {
        return programmingLanguage;
    }

    public void setProgrammingLanguage(String programmingLanguage) {
        this.programmingLanguage = programmingLanguage;
    }
}
