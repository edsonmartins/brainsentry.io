package com.integraltech.brainsentry.domain.enums;

/**
 * Categorization of notes by their scope and applicability.
 *
 * Determines how widely a note should be shared and applied.
 */
public enum NoteCategory {
    /**
     * Notes specific to a single project.
     * Only relevant within that project's context.
     */
    PROJECT_SPECIFIC,

    /**
     * Notes shared across a team or organization.
     * Applicable to multiple related projects.
     */
    SHARED,

    /**
     * Universal notes that apply broadly.
     * Like Confucius "shared" directory - generally applicable patterns.
     */
    GENERIC
}
