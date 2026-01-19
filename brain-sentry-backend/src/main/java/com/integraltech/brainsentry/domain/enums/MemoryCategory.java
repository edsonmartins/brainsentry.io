package com.integraltech.brainsentry.domain.enums;

import com.fasterxml.jackson.annotation.JsonValue;

/**
 * Categories of memories in the Brain Sentry system.
 *
 * Each memory is classified into one of these categories based on
 * its content and purpose in the development context.
 */
public enum MemoryCategory {
    /**
     * Architectural or technical decisions made during development.
     * Example: "We decided to use Spring Events for cross-module communication"
     */
    DECISION("Decision"),

    /**
     * Code patterns and best practices.
     * Example: "Agents must validate input with BeanValidator before processing"
     */
    PATTERN("Pattern"),

    /**
     * Anti-patterns and practices to avoid.
     * Example: "Never use @Transactional in controller methods"
     */
    ANTIPATTERN("Anti-Pattern"),

    /**
     * Business domain knowledge.
     * Example: "VIP customers get 24h support SLA"
     */
    DOMAIN("Domain Knowledge"),

    /**
     * Bugs encountered and their fixes.
     * Example: "NPE in UserService when user ID is null - fixed by adding Optional"
     */
    BUG("Bug Fix"),

    /**
     * Performance optimizations.
     * Example: "Added database index on user.email reduced query time by 80%"
     */
    OPTIMIZATION("Optimization"),

    /**
     * External integration details.
     * Example: "OpenRouter API requires x-api-key header with Grok model"
     */
    INTEGRATION("Integration Detail");

    private final String displayName;

    MemoryCategory(String displayName) {
        this.displayName = displayName;
    }

    @JsonValue
    public String getDisplayName() {
        return displayName;
    }
}
