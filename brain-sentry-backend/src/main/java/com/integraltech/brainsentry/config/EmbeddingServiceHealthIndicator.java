package com.integraltech.brainsentry.config;

import com.integraltech.brainsentry.service.EmbeddingService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.actuate.health.Health;
import org.springframework.boot.actuate.health.HealthIndicator;
import org.springframework.boot.autoconfigure.condition.ConditionalOnBean;
import org.springframework.stereotype.Component;

import java.time.Instant;
import java.time.temporal.ChronoUnit;

/**
 * Health indicator for Embedding service.
 *
 * Verifies the embedding model is loaded and operational.
 */
@Component
@ConditionalOnBean(EmbeddingService.class)
public class EmbeddingServiceHealthIndicator implements HealthIndicator {

    private static final Logger log = LoggerFactory.getLogger(EmbeddingServiceHealthIndicator.class);

    private final EmbeddingService embeddingService;

    public EmbeddingServiceHealthIndicator(EmbeddingService embeddingService) {
        this.embeddingService = embeddingService;
    }

    @Override
    public Health health() {
        try {
            // Check if embedding service is initialized
            boolean isReady = embeddingService.isReady();

            if (!isReady) {
                return Health.down()
                        .withDetail("service", "EmbeddingService")
                        .withDetail("status", "initializing")
                        .withDetail("timestamp", Instant.now().truncatedTo(ChronoUnit.MILLIS))
                        .build();
            }

            // Test embedding generation
            long startTime = System.currentTimeMillis();
            float[] embedding = embeddingService.embed("health check test");
            long duration = System.currentTimeMillis() - startTime;

            if (embedding == null || embedding.length == 0) {
                return Health.down()
                        .withDetail("service", "EmbeddingService")
                        .withDetail("error", "Failed to generate embedding")
                        .withDetail("timestamp", Instant.now().truncatedTo(ChronoUnit.MILLIS))
                        .build();
            }

            return Health.up()
                    .withDetail("service", "EmbeddingService")
                    .withDetail("status", "ready")
                    .withDetail("dimension", embedding.length)
                    .withDetail("latency", duration + "ms")
                    .withDetail("timestamp", Instant.now().truncatedTo(ChronoUnit.MILLIS))
                    .build();

        } catch (Exception e) {
            log.error("Embedding service health check failed", e);
            return Health.down()
                    .withDetail("service", "EmbeddingService")
                    .withDetail("error", e.getMessage())
                    .withDetail("timestamp", Instant.now().truncatedTo(ChronoUnit.MILLIS))
                    .build();
        }
    }
}
