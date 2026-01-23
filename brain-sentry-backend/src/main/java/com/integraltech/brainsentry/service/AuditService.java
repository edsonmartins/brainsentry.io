package com.integraltech.brainsentry.service;

import com.integraltech.brainsentry.domain.AuditLog;
import com.integraltech.brainsentry.domain.Memory;
import com.integraltech.brainsentry.dto.request.InterceptRequest;
import com.integraltech.brainsentry.repository.AuditLogJpaRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;

/**
 * Service for audit logging.
 *
 * Tracks all operations in the Brain Sentry system for
 * production requirements, debugging, and compliance.
 */
@Service
public class AuditService {

    private static final Logger log = LoggerFactory.getLogger(AuditService.class);

    private final AuditLogJpaRepository auditLogRepo;

    public AuditService(AuditLogJpaRepository auditLogRepo) {
        this.auditLogRepo = auditLogRepo;
    }

    /**
     * Log an interception event asynchronously.
     *
     * @param request the original request
     * @param memories memories that were used
     * @param latencyMs operation latency
     */
    @Async
    @Transactional(propagation = Propagation.REQUIRES_NEW)
    public void logInterception(InterceptRequest request, List<Memory> memories, long latencyMs) {
        try {
            AuditLog auditLog = AuditLog.builder()
                    .id(UUID.randomUUID().toString())
                    .eventType("context_injection")
                    .timestamp(Instant.now())
                    .userId(request.getUserId())
                    .sessionId(request.getSessionId())
                    .userRequest(request.getPrompt())
                    .decision(buildDecision(memories, latencyMs))
                    .latencyMs((int) latencyMs)
                    .llmCalls(1)
                    .outcome("success")
                    .tenantId(request.getTenantId())
                    .memoriesAccessed(memories.stream().map(Memory::getId).toList())
                    .inputData(buildInputData(request))
                    .outputData(buildOutputData(memories))
                    .build();

            auditLogRepo.save(auditLog);
            log.debug("Audit log created: {}", auditLog.getId());
        } catch (Exception e) {
            log.error("Error creating audit log", e);
        }
    }

    /**
     * Log a memory creation event.
     *
     * @param memoryId the memory ID
     * @param userId user who created it
     * @param tenantId tenant ID
     */
    @Async
    @Transactional(propagation = Propagation.REQUIRES_NEW)
    public void logMemoryCreated(String memoryId, String userId, String tenantId) {
        AuditLog auditLog = AuditLog.builder()
                .id(UUID.randomUUID().toString())
                .eventType("memory_created")
                .timestamp(Instant.now())
                .userId(userId)
                .tenantId(tenantId)
                .outcome("success")
                .memoriesCreated(List.of(memoryId))
                .build();

        auditLogRepo.save(auditLog);
        log.debug("Memory creation logged: {}", memoryId);
    }

    /**
     * Log a memory update event.
     *
     * @param memoryId the memory ID
     * @param userId user who updated it
     * @param tenantId tenant ID
     */
    @Async
    @Transactional(propagation = Propagation.REQUIRES_NEW)
    public void logMemoryUpdated(String memoryId, String userId, String tenantId) {
        AuditLog auditLog = AuditLog.builder()
                .id(UUID.randomUUID().toString())
                .eventType("memory_updated")
                .timestamp(Instant.now())
                .userId(userId)
                .tenantId(tenantId)
                .outcome("success")
                .memoriesModified(List.of(memoryId))
                .build();

        auditLogRepo.save(auditLog);
        log.debug("Memory update logged: {}", memoryId);
    }

    /**
     * Log a memory deletion event.
     *
     * @param memoryId the memory ID
     * @param userId user who deleted it
     * @param tenantId tenant ID
     */
    @Async
    @Transactional(propagation = Propagation.REQUIRES_NEW)
    public void logMemoryDeleted(String memoryId, String userId, String tenantId) {
        AuditLog auditLog = AuditLog.builder()
                .id(UUID.randomUUID().toString())
                .eventType("memory_deleted")
                .timestamp(Instant.now())
                .userId(userId)
                .tenantId(tenantId)
                .outcome("success")
                .memoriesModified(List.of(memoryId))
                .build();

        auditLogRepo.save(auditLog);
        log.debug("Memory deletion logged: {}", memoryId);
    }

    /**
     * Log an error event.
     *
     * @param eventType the event type
     * @param errorMessage the error message
     * @param userId the user ID
     * @param tenantId the tenant ID
     */
    @Async
    @Transactional(propagation = Propagation.REQUIRES_NEW)
    public void logError(String eventType, String errorMessage, String userId, String tenantId) {
        AuditLog auditLog = AuditLog.builder()
                .id(UUID.randomUUID().toString())
                .eventType(eventType)
                .timestamp(Instant.now())
                .userId(userId)
                .tenantId(tenantId)
                .outcome("failed")
                .errorMessage(errorMessage)
                .build();

        auditLogRepo.save(auditLog);
        log.warn("Error logged: {} - {}", eventType, errorMessage);
    }

    /**
     * Log entity extraction event.
     *
     * @param memoryId the memory ID entities were extracted from
     * @param entityCount number of entities extracted
     * @param relationshipCount number of relationships extracted
     * @param tenantId tenant ID
     */
    @Async
    @Transactional(propagation = Propagation.REQUIRES_NEW)
    public void logEntityExtraction(String memoryId, int entityCount, int relationshipCount, String tenantId) {
        Map<String, Object> outputData = new HashMap<>();
        outputData.put("memoryId", memoryId);
        outputData.put("entityCount", entityCount);
        outputData.put("relationshipCount", relationshipCount);

        AuditLog auditLog = AuditLog.builder()
                .id(UUID.randomUUID().toString())
                .eventType("entity_extraction")
                .timestamp(Instant.now())
                .userId("system")
                .tenantId(tenantId)
                .outcome("success")
                .outputData(outputData)
                .memoriesAccessed(List.of(memoryId))
                .build();

        auditLogRepo.save(auditLog);
        log.debug("Entity extraction logged for memory {}: {} entities, {} relationships",
                memoryId, entityCount, relationshipCount);
    }

    /**
     * Log relationship creation event.
     *
     * @param fromMemoryId source memory ID
     * @param toMemoryId target memory ID
     * @param relationshipType type of relationship
     * @param userId user who created it
     * @param tenantId tenant ID
     */
    @Async
    @Transactional(propagation = Propagation.REQUIRES_NEW)
    public void logRelationshipCreated(String fromMemoryId, String toMemoryId, String relationshipType,
                                      String userId, String tenantId) {
        Map<String, Object> outputData = new HashMap<>();
        outputData.put("fromMemoryId", fromMemoryId);
        outputData.put("toMemoryId", toMemoryId);
        outputData.put("relationshipType", relationshipType);

        AuditLog auditLog = AuditLog.builder()
                .id(UUID.randomUUID().toString())
                .eventType("relationship_created")
                .timestamp(Instant.now())
                .userId(userId)
                .tenantId(tenantId)
                .outcome("success")
                .outputData(outputData)
                .build();

        auditLogRepo.save(auditLog);
        log.debug("Relationship creation logged: {} -> {} ({})", fromMemoryId, toMemoryId, relationshipType);
    }

    /**
     * Get audit logs by tenant ID.
     *
     * @param tenantId the tenant ID
     * @return list of audit logs
     */
    @Transactional(readOnly = true)
    public List<AuditLog> getAuditLogsByTenant(String tenantId) {
        return auditLogRepo.findByTenantId(tenantId);
    }

    /**
     * Get audit logs by tenant ID and event type.
     *
     * @param tenantId the tenant ID
     * @param eventType the event type
     * @return list of audit logs
     */
    @Transactional(readOnly = true)
    public List<AuditLog> getAuditLogsByTenantAndEventType(String tenantId, String eventType) {
        return auditLogRepo.findByTenantIdAndEventType(tenantId, eventType);
    }

    /**
     * Get audit logs by tenant ID and user ID.
     *
     * @param tenantId the tenant ID
     * @param userId the user ID
     * @return list of audit logs
     */
    @Transactional(readOnly = true)
    public List<AuditLog> getAuditLogsByTenantAndUser(String tenantId, String userId) {
        return auditLogRepo.findByTenantIdAndUserId(tenantId, userId);
    }

    /**
     * Get recent audit logs by tenant ID.
     *
     * @param tenantId the tenant ID
     * @param limit maximum number of logs to return
     * @return list of audit logs
     */
    @Transactional(readOnly = true)
    public List<AuditLog> getRecentAuditLogs(String tenantId, int limit) {
        List<AuditLog> logs = auditLogRepo.findRecentByTenantId(tenantId);
        return logs.stream().limit(limit).toList();
    }

    /**
     * Get audit logs by tenant ID within a date range.
     *
     * @param tenantId the tenant ID
     * @param startDate start date
     * @param endDate end date
     * @return list of audit logs
     */
    @Transactional(readOnly = true)
    public List<AuditLog> getAuditLogsByDateRange(String tenantId, Instant startDate, Instant endDate) {
        return auditLogRepo.findByTenantIdAndTimestampBetween(tenantId, startDate, endDate);
    }

    /**
     * Get audit logs by tenant ID and session ID.
     *
     * @param tenantId the tenant ID
     * @param sessionId the session ID
     * @return list of audit logs
     */
    @Transactional(readOnly = true)
    public List<AuditLog> getAuditLogsByTenantAndSession(String tenantId, String sessionId) {
        return auditLogRepo.findByTenantIdAndSessionId(tenantId, sessionId);
    }

    /**
     * Count audit logs by tenant ID and event type.
     *
     * @param tenantId the tenant ID
     * @param eventType the event type
     * @return count of audit logs
     */
    @Transactional(readOnly = true)
    public long countByTenantAndEventType(String tenantId, String eventType) {
        return auditLogRepo.countByTenantIdAndEventType(tenantId, eventType);
    }

    private Map<String, Object> buildDecision(List<Memory> memories, long latencyMs) {
        Map<String, Object> decision = new HashMap<>();
        decision.put("enhanced", true);
        decision.put("memoryCount", memories.size());
        decision.put("latencyMs", latencyMs);
        return decision;
    }

    private Map<String, Object> buildInputData(InterceptRequest request) {
        Map<String, Object> data = new HashMap<>();
        data.put("promptLength", request.getPrompt() != null ? request.getPrompt().length() : 0);
        data.put("hasContext", request.getContext() != null && !request.getContext().isEmpty());
        data.put("maxTokens", request.getMaxTokens());
        return data;
    }

    private Map<String, Object> buildOutputData(List<Memory> memories) {
        Map<String, Object> data = new HashMap<>();
        data.put("memoryCount", memories.size());
        data.put("categories", memories.stream()
                .map(m -> m.getCategory().getDisplayName())
                .toList());
        return data;
    }
}
