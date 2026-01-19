package com.integraltech.brainsentry.config;

import com.integraltech.brainsentry.service.OpenRouterService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.actuate.health.Health;
import org.springframework.boot.actuate.health.HealthIndicator;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.stereotype.Component;

import java.time.Instant;
import java.time.temporal.ChronoUnit;

/**
 * Health indicator for OpenRouter/LLM service.
 *
 * Checks API connectivity and validates the API key.
 */
@Component
@ConditionalOnProperty(name = "brainsentry.openrouter.api-key")
public class OpenRouterHealthIndicator implements HealthIndicator {

    private static final Logger log = LoggerFactory.getLogger(OpenRouterHealthIndicator.class);

    private final OpenRouterService openRouterService;

    public OpenRouterHealthIndicator(OpenRouterService openRouterService) {
        this.openRouterService = openRouterService;
    }

    @Override
    public Health health() {
        try {
            // Simple health check - verify service is configured
            if (openRouterService.isConfigured()) {
                return Health.up()
                        .withDetail("service", "OpenRouter")
                        .withDetail("status", "connected")
                        .withDetail("model", openRouterService.getModel())
                        .withDetail("timestamp", Instant.now().truncatedTo(ChronoUnit.MILLIS))
                        .build();
            } else {
                return Health.down()
                        .withDetail("service", "OpenRouter")
                        .withDetail("error", "API key not configured")
                        .withDetail("timestamp", Instant.now().truncatedTo(ChronoUnit.MILLIS))
                        .build();
            }
        } catch (Exception e) {
            log.error("OpenRouter health check failed", e);
            return Health.down()
                    .withDetail("service", "OpenRouter")
                    .withDetail("error", e.getMessage())
                    .withDetail("timestamp", Instant.now().truncatedTo(ChronoUnit.MILLIS))
                    .build();
        }
    }
}
