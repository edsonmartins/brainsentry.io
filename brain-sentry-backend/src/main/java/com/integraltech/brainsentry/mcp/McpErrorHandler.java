package com.integraltech.brainsentry.mcp;

import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.extern.slf4j.Slf4j;

import java.time.Instant;
import java.util.Map;

/**
 * Centralized error handling for MCP Server operations.
 *
 * Provides consistent error responses and proper error logging
 * for all MCP tools, resources, and prompts.
 */
@Slf4j
public class McpErrorHandler {

    private static final ObjectMapper objectMapper = new ObjectMapper();

    /**
     * Error categories for better error handling.
     */
    public enum ErrorCategory {
        VALIDATION("validation", "Invalid input parameters"),
        AUTHORIZATION("authorization", "Access denied"),
        NOT_FOUND("not_found", "Resource not found"),
        INTERNAL("internal", "Internal server error"),
        TENANT("tenant", "Tenant-related error"),
        RATE_LIMIT("rate_limit", "Too many requests"),
        TIMEOUT("timeout", "Operation timed out");

        private final String code;
        private final String defaultMessage;

        ErrorCategory(String code, String defaultMessage) {
            this.code = code;
            this.defaultMessage = defaultMessage;
        }

        public String getCode() {
            return code;
        }

        public String getDefaultMessage() {
            return defaultMessage;
        }
    }

    /**
     * Create a standardized error response.
     *
     * @param category the error category
     * @param message the error message
     * @param details additional details
     * @return JSON error response
     */
    public static String error(ErrorCategory category, String message, Map<String, Object> details) {
        return error(category, message, null, details);
    }

    /**
     * Create a standardized error response with cause.
     *
     * @param category the error category
     * @param message the error message
     * @param cause the causing exception
     * @param details additional details
     * @return JSON error response
     */
    public static String error(ErrorCategory category, String message, Throwable cause, Map<String, Object> details) {
        try {
            var builder = objectMapper.createObjectNode();

            builder.put("success", false);
            builder.put("error", message);
            builder.put("errorCode", category.getCode());
            builder.put("errorCategory", category.name());
            builder.put("timestamp", Instant.now().toString());

            if (cause != null) {
                builder.put("errorType", cause.getClass().getSimpleName());
                if (log.isDebugEnabled()) {
                    builder.put("stackTrace", getStackTrace(cause));
                }
            }

            if (details != null && !details.isEmpty()) {
                builder.set("details", objectMapper.valueToTree(details));
            }

            String response = builder.toPrettyString();
            log.warn("MCP Error [{}]: {}", category.getCode(), message);

            return response;
        } catch (Exception e) {
            log.error("Failed to create error response", e);
            return "{\"success\":false,\"error\":\"Internal error creating error response\"}";
        }
    }

    /**
     * Handle an exception and return appropriate MCP error response.
     *
     * @param e the exception
     * @param context additional context about the error
     * @return JSON error response
     */
    public static String handleException(Exception e, String context) {
        ErrorCategory category = categorizeException(e);
        String message = e.getMessage();

        if (message == null || message.isBlank()) {
            message = category.getDefaultMessage();
        }

        return error(category, message, e, Map.of("context", context != null ? context : "unknown"));
    }

    /**
     * Categorize an exception for proper error handling.
     *
     * @param e the exception
     * @return the error category
     */
    public static ErrorCategory categorizeException(Exception e) {
        if (e instanceof IllegalArgumentException ||
            e instanceof jakarta.validation.ValidationException) {
            return ErrorCategory.VALIDATION;
        }

        if (e instanceof IllegalStateException &&
            e.getMessage() != null &&
            e.getMessage().contains("tenant")) {
            return ErrorCategory.TENANT;
        }

        if (e instanceof org.springframework.security.access.AccessDeniedException ||
            (e instanceof IllegalStateException && e.getMessage() != null && e.getMessage().contains("access"))) {
            return ErrorCategory.AUTHORIZATION;
        }

        if (e instanceof java.util.NoSuchElementException ||
            (e instanceof IllegalArgumentException && e.getMessage() != null && e.getMessage().contains("not found"))) {
            return ErrorCategory.NOT_FOUND;
        }

        if (e instanceof java.util.concurrent.TimeoutException ||
            e.getClass().getName().contains("Timeout")) {
            return ErrorCategory.TIMEOUT;
        }

        // Default to internal error
        return ErrorCategory.INTERNAL;
    }

    /**
     * Create a success response.
     *
     * @param data the response data
     * @return JSON success response
     */
    public static String success(Map<String, Object> data) {
        try {
            var builder = objectMapper.createObjectNode();
            builder.put("success", true);
            builder.put("timestamp", Instant.now().toString());

            if (data != null) {
                for (Map.Entry<String, Object> entry : data.entrySet()) {
                    builder.set(entry.getKey(), objectMapper.valueToTree(entry.getValue()));
                }
            }

            return builder.toPrettyString();
        } catch (Exception e) {
            log.error("Failed to create success response", e);
            return "{\"success\":true}";
        }
    }

    /**
     * Create a simple success response with just a message.
     *
     * @param message the success message
     * @return JSON success response
     */
    public static String success(String message) {
        return success(Map.of("message", message));
    }

    /**
     * Get abbreviated stack trace for debugging.
     *
     * @param e the exception
     * @return abbreviated stack trace
     */
    private static String getStackTrace(Throwable e) {
        StringBuilder sb = new StringBuilder();
        for (StackTraceElement element : e.getStackTrace()) {
            String className = element.getClassName();
            // Include only relevant stack frames
            if (className.contains("brainsentry") || className.contains("integraltech")) {
                sb.append("    at ").append(element).append("\n");
            }
        }
        return sb.toString();
    }

    /**
     * Validate required parameter and throw exception if missing.
     *
     * @param value the parameter value
     * @param paramName the parameter name
     * @throws IllegalArgumentException if parameter is missing
     */
    public static void requireParameter(Object value, String paramName) {
        if (value == null) {
            throw new IllegalArgumentException("Required parameter '" + paramName + "' is missing");
        }
        if (value instanceof String && ((String) value).isBlank()) {
            throw new IllegalArgumentException("Required parameter '" + paramName + "' cannot be empty");
        }
    }

    /**
     * Validate tenant ID and throw exception if invalid.
     *
     * @param tenantId the tenant ID
     * @throws IllegalArgumentException if tenant ID is invalid
     */
    public static void validateTenantId(String tenantId) {
        try {
            McpTenantContext.normalizeTenantId(tenantId);
        } catch (IllegalArgumentException e) {
            throw new IllegalArgumentException("Invalid tenant ID: " + e.getMessage(), e);
        }
    }
}
