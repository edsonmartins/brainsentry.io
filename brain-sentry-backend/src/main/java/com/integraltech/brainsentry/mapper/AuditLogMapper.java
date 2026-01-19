package com.integraltech.brainsentry.mapper;

import com.integraltech.brainsentry.domain.AuditLog;
import com.integraltech.brainsentry.dto.response.AuditLogResponse;

import java.util.List;
import java.util.stream.Collectors;

/**
 * Mapper for AuditLog entity and DTOs.
 */
public class AuditLogMapper {

    /**
     * Convert AuditLog entity to AuditLogResponse DTO.
     */
    public static AuditLogResponse toResponse(AuditLog auditLog) {
        if (auditLog == null) {
            return null;
        }

        return AuditLogResponse.builder()
                .id(auditLog.getId())
                .eventType(auditLog.getEventType())
                .timestamp(auditLog.getTimestamp())
                .userId(auditLog.getUserId())
                .sessionId(auditLog.getSessionId())
                .userRequest(auditLog.getUserRequest())
                .decision(auditLog.getDecision())
                .reasoning(auditLog.getReasoning())
                .confidence(auditLog.getConfidence())
                .inputData(auditLog.getInputData())
                .outputData(auditLog.getOutputData())
                .memoriesAccessed(auditLog.getMemoriesAccessed())
                .memoriesCreated(auditLog.getMemoriesCreated())
                .memoriesModified(auditLog.getMemoriesModified())
                .latencyMs(auditLog.getLatencyMs())
                .llmCalls(auditLog.getLlmCalls())
                .tokensUsed(auditLog.getTokensUsed())
                .outcome(auditLog.getOutcome())
                .errorMessage(auditLog.getErrorMessage())
                .userFeedback(auditLog.getUserFeedback())
                .tenantId(auditLog.getTenantId())
                .build();
    }

    /**
     * Convert a list of AuditLog entities to AuditLogResponse DTOs.
     */
    public static List<AuditLogResponse> toResponseList(List<AuditLog> auditLogs) {
        if (auditLogs == null) {
            return List.of();
        }
        return auditLogs.stream()
                .map(AuditLogMapper::toResponse)
                .collect(Collectors.toList());
    }
}
