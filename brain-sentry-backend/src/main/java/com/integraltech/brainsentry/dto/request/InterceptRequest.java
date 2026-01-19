package com.integraltech.brainsentry.dto.request;

import jakarta.validation.constraints.NotBlank;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.Map;

/**
 * Request to intercept and enhance a prompt.
 *
 * This is the main entry point for the Brain Sentry system.
 * The system analyzes the prompt and injects relevant memory context.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class InterceptRequest {

    /**
     * The original user prompt.
     */
    @NotBlank(message = "Prompt is required")
    private String prompt;

    /**
     * User ID making the request.
     */
    private String userId;

    /**
     * Session ID for tracking.
     */
    private String sessionId;

    /**
     * Additional context about the request.
     * May include project name, file path, etc.
     */
    private Map<String, Object> context;

    /**
     * Tenant ID.
     */
    private String tenantId;

    /**
     * Maximum tokens to inject.
     */
    @Builder.Default
    private Integer maxTokens = 500;

    /**
     * Whether to force deep analysis (skip quick check).
     */
    private Boolean forceDeepAnalysis;

    // Manual getters in case Lombok doesn't generate them
    public String getPrompt() {
        return prompt;
    }

    public void setPrompt(String prompt) {
        this.prompt = prompt;
    }

    public String getUserId() {
        return userId;
    }

    public void setUserId(String userId) {
        this.userId = userId;
    }

    public String getSessionId() {
        return sessionId;
    }

    public void setSessionId(String sessionId) {
        this.sessionId = sessionId;
    }

    public Map<String, Object> getContext() {
        return context;
    }

    public void setContext(Map<String, Object> context) {
        this.context = context;
    }

    public String getTenantId() {
        return tenantId;
    }

    public void setTenantId(String tenantId) {
        this.tenantId = tenantId;
    }

    public Integer getMaxTokens() {
        return maxTokens;
    }

    public void setMaxTokens(Integer maxTokens) {
        this.maxTokens = maxTokens;
    }

    public Boolean getForceDeepAnalysis() {
        return forceDeepAnalysis;
    }

    public void setForceDeepAnalysis(Boolean forceDeepAnalysis) {
        this.forceDeepAnalysis = forceDeepAnalysis;
    }
}
