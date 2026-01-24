package com.integraltech.brainsentry.domain;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.JdbcTypeCode;
import org.hibernate.type.SqlTypes;

import java.time.Instant;
import java.util.List;
import java.util.Map;

/**
 * Audit log entry for tracking all system operations.
 *
 * Brain Sentry maintains comprehensive audit trails for production
 * requirements, enabling debugging, compliance, and system improvement.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "audit_logs", indexes = {
    @Index(name = "idx_audit_tenant", columnList = "tenantId"),
    @Index(name = "idx_audit_user", columnList = "userId"),
    @Index(name = "idx_audit_event_type", columnList = "eventType"),
    @Index(name = "idx_audit_timestamp", columnList = "timestamp"),
    @Index(name = "idx_audit_session", columnList = "sessionId")
})
public class AuditLog {

    /**
     * Unique identifier for this log entry.
     */
    @Id
    @Column(length = 100)
    private String id;

    // ==================== Event Info ====================

    /**
     * Type of event that occurred.
     * Values: "context_injection", "memory_created", "memory_updated",
     *         "memory_deleted", "relationship_created", etc.
     */
    @Column(length = 100)
    private String eventType;

    /**
     * When this event occurred.
     */
    @Column(nullable = false)
    private Instant timestamp;

    // ==================== Context ====================

    /**
     * User who triggered this event.
     */
    @Column(length = 100)
    private String userId;

    /**
     * Session identifier for grouping related events.
     */
    @Column(length = 100)
    private String sessionId;

    /**
     * The original user request/prompt that triggered this event.
     */
    @Lob
    @Column(columnDefinition = "TEXT")
    private String userRequest;

    // ==================== Decision (for AI operations) ====================

    /**
     * Decision made by the system (e.g., to inject context).
     */
    @JdbcTypeCode(SqlTypes.JSON)
    private Map<String, Object> decision;

    /**
     * Reasoning behind the decision.
     */
    @Lob
    @Column(columnDefinition = "TEXT")
    private String reasoning;

    /**
     * Confidence score of the decision (0.0 to 1.0).
     */
    private Double confidence;

    // ==================== Data ====================

    /**
     * Input data for this operation.
     */
    @JdbcTypeCode(SqlTypes.JSON)
    private Map<String, Object> inputData;

    /**
     * Output data from this operation.
     */
    @JdbcTypeCode(SqlTypes.JSON)
    private Map<String, Object> outputData;

    // ==================== Memory Tracking ====================

    /**
     * Memories that were accessed during this operation.
     */
    @ElementCollection
    @CollectionTable(name = "audit_memories_accessed", joinColumns = @JoinColumn(name = "audit_log_id"))
    @Column(name = "memory_id")
    private List<String> memoriesAccessed;

    /**
     * Memories that were created during this operation.
     */
    @ElementCollection
    @CollectionTable(name = "audit_memories_created", joinColumns = @JoinColumn(name = "audit_log_id"))
    @Column(name = "memory_id")
    private List<String> memoriesCreated;

    /**
     * Memories that were modified during this operation.
     */
    @ElementCollection
    @CollectionTable(name = "audit_memories_modified", joinColumns = @JoinColumn(name = "audit_log_id"))
    @Column(name = "memory_id")
    private List<String> memoriesModified;

    // ==================== Performance ====================

    /**
     * Operation latency in milliseconds.
     */
    private Integer latencyMs;

    /**
     * Number of LLM API calls made.
     */
    private Integer llmCalls;

    /**
     * Number of tokens used (input + output).
     */
    private Integer tokensUsed;

    // ==================== Outcome ====================

    /**
     * Operation outcome.
     * Values: "success", "failed", "rejected", "partial"
     */
    @Column(length = 50)
    private String outcome;

    /**
     * Error message if operation failed.
     */
    @Lob
    @Column(columnDefinition = "TEXT")
    private String errorMessage;

    /**
     * User feedback on the operation result.
     */
    @JdbcTypeCode(SqlTypes.JSON)
    private Map<String, Object> userFeedback;

    // ==================== Tenant Support ====================

    /**
     * Tenant ID for multi-tenancy support.
     */
    @org.hibernate.annotations.TenantId
    @Column(length = 100, nullable = false)
    private String tenantId;

    @PrePersist
    protected void onCreate() {
        if (timestamp == null) {
            timestamp = Instant.now();
        }
    }

    // Manual builder method in case Lombok doesn't generate it
    public static AuditLogBuilder builder() {
        return new AuditLogBuilder();
    }

    public static class AuditLogBuilder {
        private String id;
        private String eventType;
        private Instant timestamp;
        private String userId;
        private String sessionId;
        private String userRequest;
        private Map<String, Object> decision;
        private String reasoning;
        private Double confidence;
        private Map<String, Object> inputData;
        private Map<String, Object> outputData;
        private List<String> memoriesAccessed;
        private List<String> memoriesCreated;
        private List<String> memoriesModified;
        private Integer latencyMs;
        private Integer llmCalls;
        private Integer tokensUsed;
        private String outcome;
        private String errorMessage;
        private Map<String, Object> userFeedback;
        private String tenantId;

        public AuditLogBuilder id(String id) {
            this.id = id;
            return this;
        }

        public AuditLogBuilder eventType(String eventType) {
            this.eventType = eventType;
            return this;
        }

        public AuditLogBuilder timestamp(Instant timestamp) {
            this.timestamp = timestamp;
            return this;
        }

        public AuditLogBuilder userId(String userId) {
            this.userId = userId;
            return this;
        }

        public AuditLogBuilder sessionId(String sessionId) {
            this.sessionId = sessionId;
            return this;
        }

        public AuditLogBuilder userRequest(String userRequest) {
            this.userRequest = userRequest;
            return this;
        }

        public AuditLogBuilder decision(Map<String, Object> decision) {
            this.decision = decision;
            return this;
        }

        public AuditLogBuilder reasoning(String reasoning) {
            this.reasoning = reasoning;
            return this;
        }

        public AuditLogBuilder confidence(Double confidence) {
            this.confidence = confidence;
            return this;
        }

        public AuditLogBuilder inputData(Map<String, Object> inputData) {
            this.inputData = inputData;
            return this;
        }

        public AuditLogBuilder outputData(Map<String, Object> outputData) {
            this.outputData = outputData;
            return this;
        }

        public AuditLogBuilder memoriesAccessed(List<String> memoriesAccessed) {
            this.memoriesAccessed = memoriesAccessed;
            return this;
        }

        public AuditLogBuilder memoriesCreated(List<String> memoriesCreated) {
            this.memoriesCreated = memoriesCreated;
            return this;
        }

        public AuditLogBuilder memoriesModified(List<String> memoriesModified) {
            this.memoriesModified = memoriesModified;
            return this;
        }

        public AuditLogBuilder latencyMs(Integer latencyMs) {
            this.latencyMs = latencyMs;
            return this;
        }

        public AuditLogBuilder llmCalls(Integer llmCalls) {
            this.llmCalls = llmCalls;
            return this;
        }

        public AuditLogBuilder tokensUsed(Integer tokensUsed) {
            this.tokensUsed = tokensUsed;
            return this;
        }

        public AuditLogBuilder outcome(String outcome) {
            this.outcome = outcome;
            return this;
        }

        public AuditLogBuilder errorMessage(String errorMessage) {
            this.errorMessage = errorMessage;
            return this;
        }

        public AuditLogBuilder userFeedback(Map<String, Object> userFeedback) {
            this.userFeedback = userFeedback;
            return this;
        }

        public AuditLogBuilder tenantId(String tenantId) {
            this.tenantId = tenantId;
            return this;
        }

        public AuditLog build() {
            AuditLog auditLog = new AuditLog();
            auditLog.id = this.id;
            auditLog.eventType = this.eventType;
            auditLog.timestamp = this.timestamp;
            auditLog.userId = this.userId;
            auditLog.sessionId = this.sessionId;
            auditLog.userRequest = this.userRequest;
            auditLog.decision = this.decision;
            auditLog.reasoning = this.reasoning;
            auditLog.confidence = this.confidence;
            auditLog.inputData = this.inputData;
            auditLog.outputData = this.outputData;
            auditLog.memoriesAccessed = this.memoriesAccessed;
            auditLog.memoriesCreated = this.memoriesCreated;
            auditLog.memoriesModified = this.memoriesModified;
            auditLog.latencyMs = this.latencyMs;
            auditLog.llmCalls = this.llmCalls;
            auditLog.tokensUsed = this.tokensUsed;
            auditLog.outcome = this.outcome;
            auditLog.errorMessage = this.errorMessage;
            auditLog.userFeedback = this.userFeedback;
            auditLog.tenantId = this.tenantId;
            return auditLog;
        }
    }
}
