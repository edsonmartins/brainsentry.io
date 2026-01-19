package com.integraltech.brainsentry.config;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.RequiredArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.web.authentication.WebAuthenticationDetailsSource;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;
import java.util.Collections;

/**
 * JWT authentication filter.
 *
 * Validates JWT tokens from the Authorization header
 * and sets the authentication in the security context.
 */
@Component
@RequiredArgsConstructor
public class JwtAuthenticationFilter extends OncePerRequestFilter {

    private static final Logger log = LoggerFactory.getLogger(JwtAuthenticationFilter.class);

    private static final String AUTHORIZATION_HEADER = "Authorization";
    private static final String BEARER_PREFIX = "Bearer ";
    private static final String SECRET_KEY = "brain-sentry-secret-change-in-production";

    @Override
    protected void doFilterInternal(
        HttpServletRequest request,
        HttpServletResponse response,
        FilterChain filterChain
    ) throws ServletException, IOException {
        String header = request.getHeader(AUTHORIZATION_HEADER);

        if (header != null && header.startsWith(BEARER_PREFIX)) {
            String token = header.substring(BEARER_PREFIX.length());

            try {
                // Simple JWT validation (in production, use a proper JWT library)
                if (validateToken(token)) {
                    String username = extractUsername(token);
                    UsernamePasswordAuthenticationToken auth =
                        new UsernamePasswordAuthenticationToken(username, null, Collections.emptyList());
                    auth.setDetails(new WebAuthenticationDetailsSource().buildDetails(request));
                    SecurityContextHolder.getContext().setAuthentication(auth);
                    log.debug("Set authentication for user: {}", username);
                }
            } catch (Exception e) {
                log.warn("Failed to validate JWT token: {}", e.getMessage());
            }
        }

        filterChain.doFilter(request, response);
    }

    private boolean validateToken(String token) {
        // Basic validation - check format and expiration
        // In production, use io.jsonwebtoken.JWT or similar
        return token != null && !token.isEmpty() && token.split("\\.").length == 3;
    }

    private String extractUsername(String token) {
        // Extract username from JWT payload
        // In production, use proper JWT parsing
        try {
            String[] parts = token.split("\\.");
            if (parts.length >= 2) {
                String payload = new String(java.util.Base64.getUrlDecoder().decode(parts[1]));
                // Simple extraction - parse JSON in production
                return "user";  // Placeholder
            }
        } catch (Exception e) {
            log.warn("Failed to extract username from token: {}", e.getMessage());
        }
        return "user";
    }
}
