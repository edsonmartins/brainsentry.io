package com.integraltech.brainsentry.mcp.tools;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.integraltech.brainsentry.dto.request.InterceptRequest;
import com.integraltech.brainsentry.dto.response.InterceptResponse;
import com.integraltech.brainsentry.mcp.McpErrorHandler;
import com.integraltech.brainsentry.mcp.McpTenantContext;
import com.integraltech.brainsentry.service.InterceptionService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

/**
 * MCP Tool for intercepting and enhancing prompts.
 *
 * This tool allows AI agents to analyze prompts and inject
 * relevant memory context from the Brain Sentry system.
 */
@Slf4j
@Component
@RequiredArgsConstructor
public class InterceptPromptTool {

    private final InterceptionService interceptionService;

    /**
     * Tool definition for the MCP server.
     */
    public static final JsonNode TOOL_DEFINITION = new ObjectMapper().createObjectNode()
            .put("name", "intercept_prompt")
            .put("description", "Intercept and enhance a prompt with relevant memory context from Brain Sentry")
            .put("inputSchema", """
                {
                    "type": "object",
                    "properties": {
                        "prompt": {
                            "type": "string",
                            "description": "The original user prompt to analyze and enhance"
                        },
                        "sessionId": {
                            "type": "string",
                            "description": "Session ID for tracking the interception"
                        },
                        "userId": {
                            "type": "string",
                            "description": "User ID making the request"
                        },
                        "tenantId": {
                            "type": "string",
                            "description": "The tenant ID for multi-tenancy isolation"
                        },
                        "maxTokens": {
                            "type": "number",
                            "description": "Maximum tokens to inject (default: 500)",
                            "default": 500
                        },
                        "forceDeepAnalysis": {
                            "type": "boolean",
                            "description": "Force deep analysis skipping quick check (default: false)",
                            "default": false
                        },
                        "context": {
                            "type": "object",
                            "description": "Additional context about the request (project, file path, etc.)"
                        }
                    },
                    "required": ["prompt", "tenantId"]
                }
                """);

    /**
     * Execute the tool with the given arguments.
     *
     * @param arguments the input arguments as JSON
     * @return the result as JSON
     */
    public String execute(JsonNode arguments) {
        try {
            // Validate required parameters
            McpErrorHandler.requireParameter(arguments, "arguments");

            // Extract and validate tenantId first
            String tenantId = arguments.has("tenantId") ? arguments.get("tenantId").asText() : null;
            McpErrorHandler.validateTenantId(tenantId);
            tenantId = McpTenantContext.normalizeTenantId(tenantId);

            // Set tenant context for this operation
            McpTenantContext.setTenant(tenantId);

            // Extract other parameters
            String prompt = arguments.has("prompt") ? arguments.get("prompt").asText() : null;
            McpErrorHandler.requireParameter(prompt, "prompt");

            String sessionId = arguments.has("sessionId") ? arguments.get("sessionId").asText() : null;
            String userId = arguments.has("userId") ? arguments.get("userId").asText() : null;
            Integer maxTokens = arguments.has("maxTokens") ? arguments.get("maxTokens").asInt() : 500;
            Boolean forceDeepAnalysis = arguments.has("forceDeepAnalysis") ? arguments.get("forceDeepAnalysis").asBoolean() : false;

            // Extract context object if present
            ObjectMapper mapper = new ObjectMapper();
            java.util.Map<String, Object> context = null;
            if (arguments.has("context") && arguments.get("context").isObject()) {
                context = mapper.convertValue(
                    arguments.get("context"),
                    new com.fasterxml.jackson.core.type.TypeReference<java.util.Map<String, Object>>() {}
                );
            }

            // Build the request
            InterceptRequest request = InterceptRequest.builder()
                    .prompt(prompt)
                    .sessionId(sessionId)
                    .userId(userId)
                    .tenantId(tenantId)
                    .maxTokens(maxTokens)
                    .forceDeepAnalysis(forceDeepAnalysis)
                    .context(context)
                    .build();

            // Execute interception
            InterceptResponse response = interceptionService.interceptAndEnhance(request);

            // Format response for MCP
            com.fasterxml.jackson.databind.node.ObjectNode result = mapper.createObjectNode();
            result.put("success", true);
            result.put("enhanced", response.getEnhanced());
            result.put("originalPrompt", response.getOriginalPrompt());
            result.put("enhancedPrompt", response.getEnhancedPrompt());
            result.put("contextInjected", response.getContextInjected());
            result.set("memoriesUsed", mapper.valueToTree(response.getMemoriesUsed()));
            result.put("latencyMs", response.getLatencyMs() != null ? String.valueOf(response.getLatencyMs()) : "0");
            result.put("reasoning", response.getReasoning());
            result.put("confidence", response.getConfidence() != null ? String.valueOf(response.getConfidence()) : "0.0");
            result.put("tokensInjected", response.getTokensInjected() != null ? String.valueOf(response.getTokensInjected()) : "0");
            result.put("llmCalls", response.getLlmCalls() != null ? String.valueOf(response.getLlmCalls()) : "0");
            result.put("tenantId", tenantId);

            return result.toPrettyString();

        } catch (Exception e) {
            log.error("Error executing intercept_prompt tool", e);
            return McpErrorHandler.handleException(e, "intercept_prompt");
        }
    }

    /**
     * Get the tool definition for MCP discovery.
     */
    public static String getToolDefinition() {
        return TOOL_DEFINITION.toPrettyString();
    }
}
