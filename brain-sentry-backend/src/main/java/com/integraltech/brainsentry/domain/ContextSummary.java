package com.integraltech.brainsentry.domain;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.TenantId;

import java.time.Instant;
import java.util.UUID;

/**
 * Context Summary entity for tracking compression events.
 *
 * Stores compressed conversation history to provide audit trail
 * and enable analysis of compression effectiveness over time.
 *
 * Inspired by Confucius Code Agent's context compression system.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "context_summaries", indexes = {
        @Index(name = "idx_summary_tenant", columnList = "tenantId"),
        @Index(name = "idx_summary_session", columnList = "sessionId"),
        @Index(name = "idx_summary_created", columnList = "createdAt")
})
public class ContextSummary {

    @Id
    @Column(length = 100)
    @Builder.Default
    private String id = UUID.randomUUID().toString();

    /**
     * Tenant ID for multi-tenancy support.
     */
    @TenantId
    @Column(length = 100, nullable = false)
    private String tenantId;

    /**
     * Session ID this summary belongs to.
     */
    @Column(length = 100, nullable = false)
    private String sessionId;

    // ==================== Token Metrics ====================

    /**
     * Original token count before compression.
     */
    @Column(nullable = false)
    private Integer originalTokenCount;

    /**
     * Token count after compression.
     */
    @Column(nullable = false)
    private Integer compressedTokenCount;

    /**
     * Compression ratio (compressed / original).
     * Target: < 0.5 (50% reduction).
     */
    @Column(nullable = false)
    private Double compressionRatio;

    // ==================== Structured Summary ====================

    /**
     * Full summary text (markdown format).
     */
    @Lob
    @Column(columnDefinition = "TEXT")
    private String summary;

    /**
     * Task goals preserved from the original context.
     */
    @ElementCollection
    @CollectionTable(name = "context_summary_goals", joinColumns = @JoinColumn(name = "summary_id"))
    @Column(name = "goal")
    private java.util.List<String> goals;

    /**
     * Key decisions made during the session.
     */
    @ElementCollection
    @CollectionTable(name = "context_summary_decisions", joinColumns = @JoinColumn(name = "summary_id"))
    @Column(name = "decision")
    private java.util.List<String> decisions;

    /**
     * Critical errors encountered with resolutions.
     */
    @ElementCollection
    @CollectionTable(name = "context_summary_errors", joinColumns = @JoinColumn(name = "summary_id"))
    @Column(name = "error")
    private java.util.List<String> errors;

    /**
     * Open TODOs and pending actions.
     */
    @ElementCollection
    @CollectionTable(name = "context_summary_todos", joinColumns = @JoinColumn(name = "summary_id"))
    @Column(name = "todo")
    private java.util.List<String> todos;

    // ==================== Metadata ====================

    /**
     * Number of recent messages kept in full (not compressed).
     */
    @Column(nullable = false)
    private Integer recentWindowSize;

    /**
     * When this compression was created.
     */
    @Column(nullable = false, updatable = false)
    private Instant createdAt;

    /**
     * Model used for LLM-based compression (if applicable).
     */
    @Column(length = 100)
    private String modelUsed;

    /**
     * Compression method used (LLM, RULE_BASED, HYBRID).
     */
    @Column(length = 50)
    private String compressionMethod;

    // ==================== Lifecycle Callbacks ====================

    @PrePersist
    protected void onCreate() {
        createdAt = Instant.now();
        if (compressionMethod == null) {
            compressionMethod = "LLM";
        }
        if (recentWindowSize == null) {
            recentWindowSize = 10;  // Default from Confucius spec
        }
    }

    // ==================== Business Logic ====================

    /**
     * Check if compression achieved target ratio (< 0.5).
     */
    @Transient
    public boolean isTargetAchieved() {
        return compressionRatio < 0.5;
    }

    /**
     * Calculate token savings.
     * @return tokens saved
     */
    @Transient
    public Integer getTokenSavings() {
        return originalTokenCount - compressedTokenCount;
    }

    /**
     * Calculate percentage reduction.
     * @return percentage (0-100)
     */
    @Transient
    public Double getPercentageReduction() {
        if (originalTokenCount == 0) {
            return 0.0;
        }
        return ((double) getTokenSavings() / originalTokenCount) * 100;
    }

    /**
     * Check if this is an effective compression (>25% reduction).
     */
    @Transient
    public boolean isEffective() {
        return getPercentageReduction() > 25.0;
    }

    /**
     * Get information preservation score estimate.
     * Based on compression ratio and structure preservation.
     *
     * @return score 0-100
     */
    @Transient
    public int getInformationPreservationScore() {
        // Simple heuristic based on compression ratio
        // In production, this would use human evaluation
        if (compressionRatio <= 0.3) {
            return 95;  // Very good compression
        } else if (compressionRatio <= 0.5) {
            return 90;  // Target achieved
        } else if (compressionRatio <= 0.7) {
            return 75;  // Acceptable
        } else {
            return 50;  // Poor compression
        }
    }
}
