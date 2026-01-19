package com.integraltech.brainsentry.domain.enums;

import com.fasterxml.jackson.annotation.JsonValue;

/**
 * Severity levels for notes, particularly hindsight notes.
 *
 * Indicates how critical the information is and how strongly
 * it should be followed or avoided.
 */
public enum NoteSeverity {
    /**
     * Must be followed or avoided.
     * Violating CRITICAL notes can lead to:
     * - Data loss
     * - Security vulnerabilities
     * - System failures
     */
    CRITICAL("Critical"),

    /**
     * Important to remember.
     * Following HIGH notes prevents significant issues.
     */
    HIGH("High"),

    /**
     * Good to know.
     * Useful optimizations and improvements.
     */
    MEDIUM("Medium"),

    /**
     * Nice to have.
     * Minor improvements or edge cases.
     */
    LOW("Low");

    private final String displayName;

    NoteSeverity(String displayName) {
        this.displayName = displayName;
    }

    @JsonValue
    public String getDisplayName() {
        return displayName;
    }

    /**
     * Converts string value to enum, case-insensitive.
     * Useful for parsing from external sources.
     */
    public static NoteSeverity fromString(String value) {
        if (value == null) {
            return MEDIUM;
        }
        try {
            return NoteSeverity.valueOf(value.toUpperCase());
        } catch (IllegalArgumentException e) {
            return MEDIUM;
        }
    }
}
