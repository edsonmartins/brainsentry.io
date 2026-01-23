package com.integraltech.brainsentry.config;

import com.integraltech.brainsentry.repository.MemoryRepository;
import com.integraltech.brainsentry.repository.impl.MemoryRepositoryImpl;
import org.mockito.Mockito;
import org.springframework.boot.test.context.TestConfiguration;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Primary;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;
import redis.clients.jedis.JedisPool;

/**
 * Test configuration providing mock beans for integration tests.
 */
@TestConfiguration
public class TestConfig {

    /**
     * Mock JedisPool for tests.
     */
    @Bean
    public JedisPool jedisPool() {
        return Mockito.mock(JedisPool.class);
    }

    /**
     * Mock MemoryRepository for tests.
     * Marked as @Primary to avoid conflict with MemoryRepositoryImpl.
     */
    @Bean
    @Primary
    public MemoryRepository memoryRepository() {
        return Mockito.mock(MemoryRepositoryImpl.class);
    }

    /**
     * PasswordEncoder/BCryptPasswordEncoder bean for tests.
     */
    @Bean
    public BCryptPasswordEncoder bCryptPasswordEncoder() {
        return new BCryptPasswordEncoder();
    }
}
