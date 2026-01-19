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

        if (password != null && !password.isEmpty()) {
            return new JedisPool(poolConfig, host, port, timeout, password, database);
        } else {
            return new JedisPool(poolConfig, host, port, timeout, null, database);
        }
    }
}
