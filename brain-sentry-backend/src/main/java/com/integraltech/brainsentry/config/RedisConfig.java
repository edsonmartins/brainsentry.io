package com.integraltech.brainsentry.config;

import lombok.Data;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import redis.clients.jedis.JedisPool;
import redis.clients.jedis.JedisPoolConfig;

import java.time.Duration;

/**
 * Redis/FalkorDB configuration.
 *
 * Configures Jedis pool for connecting to FalkorDB which provides
 * graph database and vector search capabilities on top of Redis.
 */
@Data
@Configuration
@ConditionalOnProperty(name = "brainsentry.redis.host")
@ConfigurationProperties(prefix = "brainsentry.redis")
public class RedisConfig {

    private String host = "localhost";
    private int port = 6379;
    private String password;
    private String user = "default";  // FalkorDB/Redis ACL username
    private int database = 0;
    private int timeout = 2000;

    @Bean
    public JedisPool jedisPool() {
        JedisPoolConfig poolConfig = new JedisPoolConfig();
        poolConfig.setMaxTotal(20);
        poolConfig.setMaxIdle(10);
        poolConfig.setMinIdle(5);
        poolConfig.setTestOnBorrow(true);
        poolConfig.setTestWhileIdle(true);
        poolConfig.setMinEvictableIdleTimeMillis(Duration.ofSeconds(60).toMillis());
        poolConfig.setTimeBetweenEvictionRunsMillis(Duration.ofSeconds(30).toMillis());

        // For FalkorDB with ACL, use user parameter
        // If password is provided, use the constructor with user and password
        if (password != null && !password.isEmpty()) {
            // Try with user (for Redis 6+ ACL)
            try {
                return new JedisPool(poolConfig, host, port, timeout, user, password, database);
            } catch (NoSuchMethodError e) {
                // Fallback for older Jedis versions without user parameter
                return new JedisPool(poolConfig, host, port, timeout, password, database);
            }
        } else {
            return new JedisPool(poolConfig, host, port, timeout, null, database);
        }
    }
}
