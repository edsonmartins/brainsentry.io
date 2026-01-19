package com.integraltech.brainsentry.domain.enums;

import com.fasterxml.jackson.annotation.JsonValue;

/**
 * Importance levels for memories.
 *
 * Indicates how strongly a memory should be considered when
 * building context for AI agent prompts.
 */
public enum ImportanceLevel {
    /**
     * Critical memories that should ALWAYS be followed.
     * Violating critical memories can lead to bugs or security issues.
     */
    CRITICAL("Critical"),

    /**
     * Important memories that should generally be followed.
     * These represent best practices and established patterns.
     */
    IMPORTANT("Important"),

    /**
     * Minor memories that are nice to know but not essential.
     * These provide helpful context but won't cause issues if ignored.
     */
    MINOR("Minor");

    private final String displayName;

    ImportanceLevel(String displayName) {
        this.displayName = displayName;
    }

    @JsonValue
    public String getDisplayName() {
        return displayName;
    }
}
