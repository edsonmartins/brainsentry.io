package com.integraltech.brainsentry.domain.enums;

import com.fasterxml.jackson.annotation.JsonValue;

/**
 * Validation status for memories.
 *
 * Memories can be validated by developers to ensure accuracy
 * and relevance before being promoted to high-importance status.
 */
public enum ValidationStatus {
    /**
     * Memory has been reviewed and approved as accurate.
     */
    APPROVED("Approved"),

    /**
     * Memory is awaiting validation.
     * This is the default status for newly created memories.
     */
    PENDING("Pending"),

    /**
     * Memory has been flagged for review.
     * May be outdated, conflicting, or potentially incorrect.
     */
    FLAGGED("Flagged"),

    /**
     * Memory has been rejected as incorrect or not useful.
     */
    REJECTED("Rejected");

    private final String displayName;

    ValidationStatus(String displayName) {
        this.displayName = displayName;
    }

    @JsonValue
    public String getDisplayName() {
        return displayName;
    }
}
