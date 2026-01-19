package com.integraltech.brainsentry.config;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.actuate.health.Health;
import org.springframework.boot.actuate.health.HealthIndicator;
import org.springframework.boot.autoconfigure.condition.ConditionalOnBean;
import org.springframework.stereotype.Component;
import redis.clients.jedis.Jedis;
import redis.clients.jedis.JedisPool;
import redis.clients.jedis.exceptions.JedisConnectionException;

import java.time.Instant;
import java.time.temporal.ChronoUnit;
import java.util.Map;

/**
 * Health indicator for FalkorDB/Redis.
 *
 * Checks connection pool status and database connectivity.
 */
@Component
@ConditionalOnBean(JedisPool.class)
public class FalkorDBHealthIndicator implements HealthIndicator {

    private static final Logger log = LoggerFactory.getLogger(FalkorDBHealthIndicator.class);

    private final JedisPool jedisPool;

    public FalkorDBHealthIndicator(JedisPool jedisPool) {
        this.jedisPool = jedisPool;
    }

    @Override
    public Health health() {
        try (Jedis jedis = jedisPool.getResource()) {
            // Test connection with PING
            String response = jedis.ping();

            if (!"PONG".equals(response)) {
                return Health.down()
                        .withDetail("database", "FalkorDB")
                        .withDetail("error", "Unexpected PING response: " + response)
                        .build();
            }

            // Get database info
            String info = jedis.info("server");
            String[] lines = info.split("\r?\n");

            Map<String, String> details = new java.util.HashMap<>();
            details.put("database", "FalkorDB");
            details.put("ping", response);
            details.put("timestamp", Instant.now().truncatedTo(ChronoUnit.MILLIS).toString());

            // Parse info for useful details
            for (String line : lines) {
                if (line.startsWith("redis_version:") ||
                    line.startsWith("falkordb_version:") ||
                    line.startsWith("uptime_in_days:") ||
                    line.startsWith("connected_clients:") ||
                    line.startsWith("used_memory_human:")) {
                    String[] parts = line.split(":", 2);
                    if (parts.length == 2) {
                        details.put(parts[0], parts[1]);
                    }
                }
            }

            // Test graph capabilities
            try {
                jedis.graphquery("health-check", "RETURN 1");
                details.put("graph_module", "enabled");
            } catch (Exception e) {
                details.put("graph_module", "disabled");
            }

            return Health.up().withDetails(details).build();

        } catch (JedisConnectionException e) {
            log.error("FalkorDB health check failed - connection error", e);
            return Health.down()
                    .withDetail("database", "FalkorDB")
                    .withDetail("error", "Connection refused")
                    .withDetail("timestamp", Instant.now().truncatedTo(ChronoUnit.MILLIS))
                    .build();
        } catch (Exception e) {
            log.error("FalkorDB health check failed", e);
            return Health.down()
                    .withDetail("database", "FalkorDB")
                    .withDetail("error", e.getMessage())
                    .withDetail("timestamp", Instant.now().truncatedTo(ChronoUnit.MILLIS))
                    .build();
        }
    }
}
