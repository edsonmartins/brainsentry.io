package com.integraltech.brainsentry.config;

import lombok.extern.slf4j.Slf4j;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.client.RestTemplate;

/**
 * OpenRouter / Grok API configuration.
 *
 * Configures the connection to OpenRouter for accessing
 * the x-ai/grok-4-fast model.
 */
@Slf4j
@Configuration
@ConfigurationProperties(prefix = "brain-sentry.llm")
public class OpenRouterConfig {

    /**
     * API provider (openrouter)
     */
    private String provider = "openrouter";

    /**
     * Model identifier (x-ai/grok-4-fast)
     */
    private String model = "x-ai/grok-4-fast";

    /**
     * OpenRouter API key
     */
    private String apiKey;

    /**
     * API base URL
     */
    private String baseUrl = "https://openrouter.ai/api/v1/chat/completions";

    /**
     * Temperature for LLM responses (0.0 - 1.0)
     */
    private Double temperature = 0.3;

    /**
     * Maximum tokens in response
     */
    private Integer maxTokens = 500;

    /**
     * Request timeout in milliseconds
     */
    private Integer timeout = 30000;

    // Getters
    public String getProvider() {
        return provider;
    }

    public String getModel() {
        return model;
    }

    public String getApiKey() {
        return apiKey;
    }

    public String getBaseUrl() {
        return baseUrl;
    }

    public Double getTemperature() {
        return temperature;
    }

    public Integer getMaxTokens() {
        return maxTokens;
    }

    public Integer getTimeout() {
        return timeout;
    }

    public void setProvider(String provider) {
        this.provider = provider;
    }

    public void setModel(String model) {
        this.model = model;
    }

    public void setApiKey(String apiKey) {
        this.apiKey = apiKey;
    }

    public void setBaseUrl(String baseUrl) {
        this.baseUrl = baseUrl;
    }

    public void setTemperature(Double temperature) {
        this.temperature = temperature;
    }

    public void setMaxTokens(Integer maxTokens) {
        this.maxTokens = maxTokens;
    }

    public void setTimeout(Integer timeout) {
        this.timeout = timeout;
    }

    @Bean
    public RestTemplate restTemplate() {
        RestTemplate restTemplate = new RestTemplate();
        restTemplate.getInterceptors().add((request, body, execution) -> {
            request.getHeaders().add("HTTP-Referer", "https://brainsentry.io");
            request.getHeaders().add("X-Title", "Brain Sentry");
            request.getHeaders().add("User-Agent", "Brain-Sentry/1.0");
            return execution.execute(request, body);
        });
        return restTemplate;
    }
}
