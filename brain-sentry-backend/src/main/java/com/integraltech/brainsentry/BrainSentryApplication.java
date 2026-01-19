package com.integraltech.brainsentry;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.context.properties.ConfigurationPropertiesScan;
import org.springframework.scheduling.annotation.EnableAsync;

/**
 * Brain Sentry - Agent Memory System for Developers
 *
 * Main application entry point for the Brain Sentry backend.
 * This system provides persistent, autonomous, and intelligent memory
 * for developers using AI agents.
 *
 * @author Brain Sentry Team
 * @version 1.0.0
 */
@SpringBootApplication
@EnableAsync
@ConfigurationPropertiesScan
public class BrainSentryApplication {

    private static final String SPRING_PROFILES_ACTIVE = "spring.profiles.active";

    public static void main(String[] args) {
        SpringApplication app = new SpringApplication(BrainSentryApplication.class);

        // Set default profile if not specified
        if (System.getProperty(SPRING_PROFILES_ACTIVE) == null) {
            System.setProperty(SPRING_PROFILES_ACTIVE, "dev");
        }

        app.run(args);
    }
}
