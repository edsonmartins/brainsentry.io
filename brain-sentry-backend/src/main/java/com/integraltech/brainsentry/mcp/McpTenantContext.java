package com.integraltech.brainsentry.mcp;

import com.integraltech.brainsentry.config.TenantContext;
import lombok.extern.slf4j.Slf4j;

import java.util.concurrent.Callable;

/**
 * Tenant context management for MCP Server operations.
 *
 * This class provides utilities for ensuring tenant isolation
 * during MCP tool and resource execution, especially in virtual
 * thread environments.
 */
@Slf4j
public class McpTenantContext {

    private static final String DEFAULT_TENANT = "default";

    /**
     * Execute an operation with a specific tenant context.
     * The tenant context is automatically cleared after execution.
     *
     * @param tenantId the tenant ID
     * @param operation the operation to execute
     * @param <T> the return type
     * @return the result of the operation
     * @throws Exception if the operation fails
     */
    public static <T> T withTenant(String tenantId, Callable<T> operation) throws Exception {
        String previousTenant = TenantContext.getTenantId();
        try {
            setTenant(tenantId);
            return operation.call();
        } finally {
            // Restore previous tenant
            if (previousTenant != null) {
                TenantContext.setTenantId(previousTenant);
            } else {
                TenantContext.clear();
            }
        }
    }

    /**
     * Execute a runnable operation with a specific tenant context.
     *
     * @param tenantId the tenant ID
     * @param operation the operation to execute
     */
    public static void withTenant(String tenantId, Runnable operation) {
        String previousTenant = TenantContext.getTenantId();
        try {
            setTenant(tenantId);
            operation.run();
        } finally {
            // Restore previous tenant
            if (previousTenant != null) {
                TenantContext.setTenantId(previousTenant);
            } else {
                TenantContext.clear();
            }
        }
    }

    /**
     * Set the tenant ID with validation.
     *
     * @param tenantId the tenant ID to set
     * @return the normalized tenant ID
     * @throws IllegalArgumentException if tenantId is invalid
     */
    public static String setTenant(String tenantId) {
        String normalized = normalizeTenantId(tenantId);
        TenantContext.setTenantId(normalized);
        log.trace("MCP tenant context set to: {}", normalized);
        return normalized;
    }

    /**
     * Normalize and validate tenant ID.
     *
     * @param tenantId the tenant ID to normalize
     * @return the normalized tenant ID
     * @throws IllegalArgumentException if tenantId is invalid
     */
    public static String normalizeTenantId(String tenantId) {
        if (tenantId == null || tenantId.isBlank()) {
            return DEFAULT_TENANT;
        }

        // Remove leading/trailing whitespace
        String normalized = tenantId.trim();

        // Validate format (alphanumeric, dash, underscore)
        if (!normalized.matches("^[a-zA-Z0-9_-]+$")) {
            throw new IllegalArgumentException(
                "Invalid tenant ID format: " + tenantId + ". " +
                "Tenant ID must contain only alphanumeric characters, dashes, and underscores."
            );
        }

        // Limit length
        if (normalized.length() > 64) {
            throw new IllegalArgumentException(
                "Tenant ID too long: " + normalized.length() + " characters. " +
                "Maximum allowed is 64 characters."
            );
        }

        return normalized;
    }

    /**
     * Get the current tenant ID from context.
     *
     * @return the current tenant ID
     */
    public static String getCurrentTenant() {
        return TenantContext.getTenantId();
    }

    /**
     * Validate that a tenant ID matches the current context.
     *
     * @param tenantId the tenant ID to validate
     * @throws IllegalStateException if tenant doesn't match
     */
    public static void validateTenant(String tenantId) {
        String current = getCurrentTenant();
        String normalized = normalizeTenantId(tenantId);

        if (!current.equals(normalized)) {
            throw new IllegalStateException(
                "Tenant mismatch: requested=" + normalized + ", current=" + current
            );
        }
    }

    /**
     * Check if access to a resource is allowed for the given tenant.
     *
     * @param resourceTenantId the tenant ID of the resource
     * @param requestingTenantId the tenant ID making the request
     * @return true if access is allowed
     */
    public static boolean isAccessAllowed(String resourceTenantId, String requestingTenantId) {
        if (resourceTenantId == null) {
            return true; // No tenant restriction
        }

        String normalizedResource = normalizeTenantId(resourceTenantId);
        String normalizedRequesting = normalizeTenantId(requestingTenantId);

        // For admin tenant, allow access to all resources (future feature)
        // For now, strict isolation
        return normalizedResource.equals(normalizedRequesting);
    }
}
