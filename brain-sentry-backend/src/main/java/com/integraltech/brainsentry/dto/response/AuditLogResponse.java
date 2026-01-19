package com.integraltech.brainsentry.dto.response;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;
import java.util.List;
import java.util.Map;

/**
 * Response containing audit log entry.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class AuditLogResponse {

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
}
