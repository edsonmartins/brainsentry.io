package com.integraltech.brainsentry.config;

import org.springframework.context.annotation.Configuration;
import org.springframework.data.jpa.repository.config.EnableJpaAuditing;

/**
 * JPA Auditing configuration.
 * Moved to a separate class to allow exclusion in tests.
 */
@Configuration
@EnableJpaAuditing
public class JpaAuditingConfig {
}
