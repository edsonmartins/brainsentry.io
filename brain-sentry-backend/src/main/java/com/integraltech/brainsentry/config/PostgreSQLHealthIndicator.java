package com.integraltech.brainsentry.config;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.actuate.health.Health;
import org.springframework.boot.actuate.health.HealthIndicator;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.stereotype.Component;

import javax.sql.DataSource;
import java.sql.Connection;
import java.sql.SQLException;
import java.time.Instant;
import java.time.temporal.ChronoUnit;

/**
 * Health indicator for PostgreSQL database.
 *
 * Checks connection pool status and database connectivity.
 */
@Component
public class PostgreSQLHealthIndicator implements HealthIndicator {

    private static final Logger log = LoggerFactory.getLogger(PostgreSQLHealthIndicator.class);

    private final DataSource dataSource;

    public PostgreSQLHealthIndicator(DataSource dataSource) {
        this.dataSource = dataSource;
    }

    @Override
    public Health health() {
        try (Connection connection = dataSource.getConnection()) {
            boolean isValid = connection.isValid(2);

            if (!isValid) {
                return Health.down()
                        .withDetail("database", "PostgreSQL")
                        .withDetail("error", "Connection is not valid")
                        .build();
            }

            String dbUrl = connection.getMetaData().getURL();
            String dbUser = connection.getMetaData().getUserName();
            String dbProduct = connection.getMetaData().getDatabaseProductName();
            String dbVersion = connection.getMetaData().getDatabaseProductVersion();

            return Health.up()
                    .withDetail("database", "PostgreSQL")
                    .withDetail("url", dbUrl)
                    .withDetail("user", dbUser)
                    .withDetail("product", dbProduct)
                    .withDetail("version", dbVersion)
                    .withDetail("timestamp", Instant.now().truncatedTo(ChronoUnit.MILLIS))
                    .build();

        } catch (SQLException e) {
            log.error("PostgreSQL health check failed", e);
            return Health.down()
                    .withDetail("database", "PostgreSQL")
                    .withDetail("error", e.getMessage())
                    .withDetail("timestamp", Instant.now().truncatedTo(ChronoUnit.MILLIS))
                    .build();
        }
    }
}
