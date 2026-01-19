package com.integraltech.brainsentry.config;

import io.micrometer.core.instrument.MeterRegistry;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.actuate.autoconfigure.metrics.MetricsProperties;
import org.springframework.boot.actuate.metrics.export.MetricsExportProperties;
import org.springframework.context.annotation.Configuration;

import java.time.Duration;

/**
 * Configuration for Prometheus metrics.
 *
 * Enables Prometheus endpoint for metrics scraping and configures
 * custom metrics for the Brain Sentry application.
 */
@Configuration
public class PrometheusConfig {

    private static final Logger log = LoggerFactory.getLogger(PrometheusConfig.class);

    public PrometheusConfig(
            MeterRegistry registry,
            MetricsProperties metricsProperties,
            MetricsExportProperties exportProperties
    ) {
        log.info("Configuring Prometheus metrics");

        // Configure common tags for all metrics
        registry.config().commonTags(
            "application", "brain-sentry",
            "environment", getEnvironment()
        );

        log.info("Prometheus metrics configured at /actuator/prometheus");
    }

    private String getEnvironment() {
        String env = System.getProperty("spring.profiles.active", "development");
        if (env.contains("prod")) {
            return "production";
        } else if (env.contains("test")) {
            return "test";
        }
        return "development";
    }
}
