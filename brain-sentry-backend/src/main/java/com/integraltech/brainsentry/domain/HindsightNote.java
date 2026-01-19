package com.integraltech.brainsentry.domain;

import com.integraltech.brainsentry.domain.enums.NoteSeverity;
import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.TenantId;

import java.time.Instant;
import java.util.UUID;

/**
 * Hindsight Note entity for tracking failures and learnings.
 *
 * Inspired by Confucius Code Agent's hindsight notes system.
 * These notes capture what went wrong, how it was fixed, and what was learned
 * to prevent similar errors in the future.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "hindsight_notes", indexes = {
        @Index(name = "idx_hindsight_tenant", columnList = "tenantId"),
        @Index(name = "idx_hindsight_session", columnList = "sessionId"),
        @Index(name = "idx_hindsight_error_type", columnList = "errorType"),
        @Index(name = "idx_hindsight_created_at", columnList = "createdAt"),
        @Index(name = "idx_hindsight_severity", columnList = "severity")
})
public class HindsightNote {

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
     * Session ID this note is associated with.
     */
    @Column(length = 100)
    private String sessionId;

    // ==================== Note Metadata (from Confucius spec) ====================

    /**
     * Title of the hindsight note.
     * Used for quick identification and display.
     */
    @Column(length = 500)
    private String title;

    /**
     * Regex pattern extracted from error message.
     * Used for pattern matching to find similar errors.
     * KEY FEATURE from Confucius - enables proactive error detection.
     */
    @Lob
    @Column(columnDefinition = "TEXT")
    private String errorPattern;

    /**
     * Severity level of this note.
     * Determines how strongly it should be followed.
     */
    @Enumerated(EnumType.STRING)
    @Column(length = 20, nullable = false)
    @Builder.Default
    private NoteSeverity severity = NoteSeverity.MEDIUM;

    /**
     * When this note was last accessed or suggested.
     */
    private Instant lastAccessedAt;

    /**
     * How many times this note has been accessed.
     */
    @Builder.Default
    private Integer accessCount = 0;

    // ==================== Error Information ====================

    /**
     * Type/category of the error (e.g., "NullPointerException", "API_TIMEOUT", "VALIDATION_ERROR").
     */
    @Column(length = 100)
    private String errorType;

    /**
     * The actual error message.
     */
    @Lob
    @Column(columnDefinition = "TEXT")
    private String errorMessage;

    /**
     * Context in which the error occurred (stack trace, relevant code snippets, etc.).
     */
    @Lob
    @Column(columnDefinition = "TEXT")
    private String errorContext;

    // ==================== Resolution ====================

    /**
     * Description of how the error was resolved.
     */
    @Lob
    @Column(columnDefinition = "TEXT")
    private String resolution;

    /**
     * Step-by-step resolution instructions.
     */
    @Lob
    @Column(columnDefinition = "TEXT")
    private String resolutionSteps;

    /**
     * Reference to a commit, PR, or code change that fixed the issue.
     */
    @Column(length = 500)
    private String resolutionReference;

    // ==================== Learning ====================

    /**
     * Key lessons learned from this failure.
     */
    @Lob
    @Column(columnDefinition = "TEXT")
    private String lessonsLearned;

    /**
     * Strategy to prevent this error in the future.
     */
    @Lob
    @Column(columnDefinition = "TEXT")
    private String preventionStrategy;

    // ==================== Tags & Metadata ====================

    /**
     * Tags for categorizing and searching hindsight notes.
     */
    @ElementCollection
    @CollectionTable(name = "hindsight_tags", joinColumns = @JoinColumn(name = "hindsight_id"))
    @Column(name = "tag")
    private java.util.List<String> tags;

    /**
     * Related memory IDs (memories that would have prevented this error).
     */
    @ElementCollection
    @CollectionTable(name = "hindsight_related_memories", joinColumns = @JoinColumn(name = "hindsight_id"))
    @Column(name = "memory_id")
    private java.util.List<String> relatedMemoryIds;

    // ==================== Usage Tracking ====================

    /**
     * How many times this type of error has occurred.
     * Incremented when similar errors are detected.
     */
    @Builder.Default
    private Integer occurrenceCount = 1;

    /**
     * How many times this hindsight note has been referenced/suggested.
     */
    @Builder.Default
    private Integer referenceCount = 0;

    /**
     * How many times the prevention strategy successfully prevented the error.
     */
    @Builder.Default
    private Integer preventionSuccessCount = 0;

    // ==================== Provenance ====================

    /**
     * User who created this hindsight note.
     */
    @Column(length = 100)
    private String createdBy;

    /**
     * Whether this note was created automatically by the system or manually.
     */
    @Builder.Default
    @Column(nullable = false)
    private Boolean autoGenerated = false;

    // ==================== Timestamps ====================

    @Column(nullable = false, updatable = false)
    private Instant createdAt;

    private Instant updatedAt;

    /**
     * When this error was last encountered.
     */
    private Instant lastOccurrenceAt;

    // ==================== Status ====================

    /**
     * Whether the prevention strategy has been verified/implemented.
     */
    @Builder.Default
    private Boolean preventionVerified = false;

    /**
     * Priority of addressing this error type (HIGH, MEDIUM, LOW).
     */
    @Column(length = 20)
    private String priority;

    // ==================== Lifecycle Callbacks ====================

    @PrePersist
    protected void onCreate() {
        createdAt = Instant.now();
        updatedAt = Instant.now();
        lastOccurrenceAt = Instant.now();
        if (priority == null) {
            priority = "MEDIUM";
        }
    }

    @PreUpdate
    protected void onUpdate() {
        updatedAt = Instant.now();
    }

    // ==================== Business Logic ====================

    /**
     * Increment occurrence count and update last occurrence timestamp.
     */
    public void recordOccurrence() {
        this.occurrenceCount++;
        this.lastOccurrenceAt = Instant.now();
    }

    /**
     * Increment reference count when this note is suggested.
     * Also updates lastAccessedAt and accessCount.
     */
    public void recordReference() {
        this.referenceCount++;
        this.lastAccessedAt = Instant.now();
        this.accessCount++;
    }

    /**
     * Record successful prevention.
     */
    public void recordPreventionSuccess() {
        this.preventionSuccessCount++;
    }

    /**
     * Calculate prevention effectiveness rate.
     */
    public double getPreventionEffectiveness() {
        if (occurrenceCount == 0) {
            return 0.0;
        }
        return (double) preventionSuccessCount / occurrenceCount;
    }

    /**
     * Check if this error is considered frequent (occurred more than 3 times).
     */
    @Transient
    public boolean isFrequent() {
        return occurrenceCount > 3;
    }

    /**
     * Check if prevention is considered effective (>50% success rate).
     */
    @Transient
    public boolean isPreventionEffective() {
        return getPreventionEffectiveness() > 0.5;
    }

    /**
     * Check if a given error message matches this note's error pattern.
     * KEY FEATURE from Confucius - enables proactive error detection.
     *
     * @param errorMessage the error message to check
     * @return true if the error message matches the pattern, false otherwise
     */
    @Transient
    public boolean matchesError(String errorMessage) {
        if (errorPattern == null || errorPattern.isBlank()) {
            return false;
        }
        if (errorMessage == null) {
            return false;
        }
        try {
            return errorMessage.matches(errorPattern);
        } catch (Exception e) {
            // Invalid regex pattern
            return false;
        }
    }

    /**
     * Check if this note should be suggested based on error type.
     * Simple matching by error type when pattern matching is not available.
     *
     * @param errorType the error type to check
     * @return true if error types match
     */
    @Transient
    public boolean matchesErrorType(String errorType) {
        if (this.errorType == null || errorType == null) {
            return false;
        }
        return this.errorType.equalsIgnoreCase(errorType);
    }
}
