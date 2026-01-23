package com.integraltech.brainsentry.config;

import lombok.RequiredArgsConstructor;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.method.configuration.EnableMethodSecurity;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.http.HttpMethod;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.config.annotation.web.configurers.AbstractHttpConfigurer;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;

/**
 * Security configuration.
 *
 * Configures JWT-based authentication and authorization.
 * NOTE: JPA/Hibernate configuration is auto-configured by Spring Boot.
 *
 * TODO: Re-enable manual JPA configuration when Spring Boot 4.x stabilizes:
 * - Multi-tenancy with CurrentTenantIdentifierResolver
 * - Custom Hibernate properties
 */
@Configuration
@EnableWebSecurity
@EnableMethodSecurity
@RequiredArgsConstructor
public class SecurityConfig {

    private final JwtAuthenticationFilter jwtAuthenticationFilter;

    @Bean
    public SecurityFilterChain securityFilterChain(HttpSecurity http) throws Exception {
        http
            .cors(cors -> {}) // Enable CORS with default configuration from WebConfig
            .csrf(AbstractHttpConfigurer::disable)
            .sessionManagement(session ->
                session.sessionCreationPolicy(SessionCreationPolicy.STATELESS)
            )
            .authorizeHttpRequests(auth -> auth
                // Public endpoints
                .requestMatchers(
                    "/actuator/**",
                    "/v3/api-docs/**",
                    "/swagger-ui/**",
                    "/swagger-ui.html"
                ).permitAll()
                // Auth endpoints
                .requestMatchers("/v1/auth/**").permitAll()
                // Stats/health endpoints (for testing)
                .requestMatchers("/v1/stats/**").permitAll()
                // Memories endpoints (for development testing)
                .requestMatchers("/v1/memories/**").permitAll()
                // All other public endpoints for development
                .requestMatchers("/v1/intercepts/**").permitAll()
                .requestMatchers("/v1/notes/**").permitAll()
                .requestMatchers("/v1/relationships/**").permitAll()
                .requestMatchers("/v1/users/**").permitAll()
                .requestMatchers("/v1/tenants/**").permitAll()
                // Allow OPTIONS for CORS preflight
                .requestMatchers(HttpMethod.OPTIONS).permitAll()
                // All other endpoints require authentication
                .anyRequest().authenticated()
            )
            .addFilterBefore(jwtAuthenticationFilter, UsernamePasswordAuthenticationFilter.class);

        return http.build();
    }

    @Bean
    public BCryptPasswordEncoder passwordEncoder() {
        return new BCryptPasswordEncoder(12);
    }
}

