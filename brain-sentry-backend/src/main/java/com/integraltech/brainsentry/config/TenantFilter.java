package com.integraltech.brainsentry.config;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.core.Ordered;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;

/**
 * Filter to extract and set the current tenant from HTTP requests.
 *
 * Tenant is extracted from:
 * 1. X-Tenant-ID header (highest priority)
 * 2. tenant query parameter
 * 3. JWT token claim (if authenticated)
 * 4. Default "default" tenant
 *
 * This filter MUST execute before Spring Security to ensure
 * tenant context is available during authentication.
 */
@Component
@Order(Ordered.HIGHEST_PRECEDENCE)
public class TenantFilter extends OncePerRequestFilter {

    private static final Logger log = LoggerFactory.getLogger(TenantFilter.class);

    private static final String TENANT_HEADER = "X-Tenant-ID";
    private static final String TENANT_PARAM = "tenant";
    private static final String DEFAULT_TENANT = "a9f814d2-4dae-41f3-851b-8aa3d4706561";

    @Override
    protected void doFilterInternal(
        HttpServletRequest request,
        HttpServletResponse response,
        FilterChain filterChain
    ) throws ServletException, IOException {

        String tenantId = extractTenantId(request);
        TenantContext.setTenantId(tenantId);

        log.debug("Tenant filter set tenant: {} for request: {} {}",
            tenantId, request.getMethod(), request.getRequestURI());

        try {
            filterChain.doFilter(request, response);
        } finally {
            // Always clear to prevent memory leaks
            TenantContext.clear();
        }
    }

    /**
     * Extract tenant ID from request.
     *
     * Priority order:
     * 1. X-Tenant-ID header
     * 2. tenant query parameter
     * 3. Default tenant
     */
    private String extractTenantId(HttpServletRequest request) {
        // Try header first
        String tenantId = request.getHeader(TENANT_HEADER);
        if (tenantId != null && !tenantId.isEmpty()) {
            return tenantId;
        }

        // Try query parameter
        tenantId = request.getParameter(TENANT_PARAM);
        if (tenantId != null && !tenantId.isEmpty()) {
            return tenantId;
        }

        // Try JWT token (if present) - would need JwtTokenProvider
        // tenantId = jwtTokenProvider.getTenantFromToken(request);
        // if (tenantId != null) return tenantId;

        // Use default
        return DEFAULT_TENANT;
    }
}
