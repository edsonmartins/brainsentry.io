package com.integraltech.brainsentry.domain.enums;

import com.fasterxml.jackson.annotation.JsonValue;

/**
 * Categories of memories in the Brain Sentry system.
 *
 * These categories are designed to be universal, supporting both
 * software development contexts and business/sales contexts.
 */
public enum MemoryCategory {
    /**
     * Patterns, best practices, customer preferences, behaviors.
     * Dev: "Always use dependency injection for testability"
     * Sales: "Customer prefers email communication over calls"
     */
    INSIGHT("Insight"),

    /**
     * Architectural, technical, or business decisions.
     * Dev: "We decided to use Spring Events for cross-module communication"
     * Sales: "Agreed on 15% discount for annual contract"
     */
    DECISION("Decision"),

    /**
     * Anti-patterns, bugs, objections, points of attention.
     * Dev: "Never use @Transactional in controller methods"
     * Sales: "Customer concerned about implementation timeline"
     */
    WARNING("Warning"),

    /**
     * Domain knowledge, customer knowledge, product knowledge.
     * Dev: "VIP customers get 24h support SLA"
     * Sales: "Customer uses Salesforce as primary CRM"
     */
    KNOWLEDGE("Knowledge"),

    /**
     * Actions, optimizations, follow-ups, next steps.
     * Dev: "Added database index on user.email reduced query time by 80%"
     * Sales: "Schedule demo for next Tuesday at 2pm"
     */
    ACTION("Action"),

    /**
     * Context, integrations, environment, conversation history.
     * Dev: "OpenRouter API requires x-api-key header"
     * Sales: "Met at TechConf 2024, interested in enterprise plan"
     */
    CONTEXT("Context"),

    /**
     * References, documentation, proposals, contracts, materials.
     * Dev: "API documentation at docs.example.com/api"
     * Sales: "Proposal sent on Jan 15, contract template v2.1"
     */
    REFERENCE("Reference"),

    // Legacy categories for backward compatibility
    /**
     * @deprecated Use INSIGHT instead
     */
    @Deprecated
    PATTERN("Pattern"),

    /**
     * @deprecated Use WARNING instead
     */
    @Deprecated
    ANTIPATTERN("Anti-Pattern"),

    /**
     * @deprecated Use KNOWLEDGE instead
     */
    @Deprecated
    DOMAIN("Domain Knowledge"),

    /**
     * @deprecated Use WARNING instead
     */
    @Deprecated
    BUG("Bug Fix"),

    /**
     * @deprecated Use ACTION instead
     */
    @Deprecated
    OPTIMIZATION("Optimization"),

    /**
     * @deprecated Use CONTEXT instead
     */
    @Deprecated
    INTEGRATION("Integration Detail");

    private final String displayName;

    MemoryCategory(String displayName) {
        this.displayName = displayName;
    }

    @JsonValue
    public String getDisplayName() {
        return displayName;
    }

    /**
     * Maps legacy categories to new categories.
     */
    public MemoryCategory toModernCategory() {
        return switch (this) {
            case PATTERN -> INSIGHT;
            case ANTIPATTERN, BUG -> WARNING;
            case DOMAIN -> KNOWLEDGE;
            case OPTIMIZATION -> ACTION;
            case INTEGRATION -> CONTEXT;
            default -> this;
        };
    }
}
