package com.integraltech.brainsentry.config;

import lombok.RequiredArgsConstructor;
import org.springframework.context.annotation.Configuration;
import org.springframework.format.FormatterRegistry;
import org.springframework.web.servlet.config.annotation.CorsRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

/**
 * Web MVC configuration.
 *
 * Configures CORS, formatters, and other web-related settings.
 */
@Configuration
@RequiredArgsConstructor
public class WebConfig implements WebMvcConfigurer {

    private final CorsProperties corsProperties;

    @Override
    public void addCorsMappings(CorsRegistry registry) {
        registry.addMapping("/**")
            .allowedOrigins(corsProperties.getAllowedOrigins())
            .allowedMethods(corsProperties.getAllowedMethods())
            .allowedHeaders(corsProperties.getAllowedHeaders())
            .allowCredentials(corsProperties.getAllowCredentials())
            .maxAge(3600);
    }

    @org.springframework.boot.context.properties.ConfigurationProperties(prefix = "security.cors")
    @Configuration
    @RequiredArgsConstructor
    public static class CorsProperties {
        private String[] allowedOrigins = new String[]{"http://localhost:5173"};
        private String[] allowedMethods = new String[]{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"};
        private String[] allowedHeaders = new String[]{"*"};
        private Boolean allowCredentials = true;

        public String[] getAllowedOrigins() { return allowedOrigins; }
        public void setAllowedOrigins(String[] allowedOrigins) { this.allowedOrigins = allowedOrigins; }
        public String[] getAllowedMethods() { return allowedMethods; }
        public void setAllowedMethods(String[] allowedMethods) { this.allowedMethods = allowedMethods; }
        public String[] getAllowedHeaders() { return allowedHeaders; }
        public void setAllowedHeaders(String[] allowedHeaders) { this.allowedHeaders = allowedHeaders; }
        public Boolean getAllowCredentials() { return allowCredentials; }
        public void setAllowCredentials(Boolean allowCredentials) { this.allowCredentials = allowCredentials; }
    }
}
