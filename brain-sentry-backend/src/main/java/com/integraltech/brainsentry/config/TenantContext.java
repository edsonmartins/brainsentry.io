package com.integraltech.brainsentry.config;

/**
 * Thread-local context for storing the current tenant identifier.
 *
 * Uses ThreadLocal to ensure tenant isolation per request thread.
 * In virtual thread environments, this still works correctly as
 * each virtual thread has its own copy of thread-local values.
 */
public class TenantContext {

    private static final ThreadLocal<String> CURRENT_TENANT = new ThreadLocal<>();
    private static final String DEFAULT_TENANT = "a9f814d2-4dae-41f3-851b-8aa3d4706561";

    /**
     * Set the current tenant identifier.
     *
     * @param tenantId the tenant ID (use null for default)
     */
    public static void setTenantId(String tenantId) {
        if (tenantId == null || tenantId.isEmpty()) {
            CURRENT_TENANT.set(DEFAULT_TENANT);
        } else {
            CURRENT_TENANT.set(tenantId);
        }
    }

    /**
     * Get the current tenant identifier.
     *
     * @return the tenant ID, or "default" if not set
     */
    public static String getTenantId() {
        String tenantId = CURRENT_TENANT.get();
        return tenantId != null ? tenantId : DEFAULT_TENANT;
    }

    /**
     * Clear the current tenant identifier.
     * Should be called at the end of request processing.
     */
    public static void clear() {
        CURRENT_TENANT.remove();
    }

    /**
     * Check if a specific tenant is currently active.
     *
     * @param tenantId the tenant ID to check
     * @return true if the given tenant is active
     */
    public static boolean isTenant(String tenantId) {
        return getTenantId().equals(tenantId);
    }
}
