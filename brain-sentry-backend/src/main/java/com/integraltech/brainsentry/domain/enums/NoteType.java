package com.integraltech.brainsentry.domain.enums;

/**
 * Types of notes that can be created by the Note-Taking Agent.
 *
 * Inspired by Confucius Code Agent's note classification system.
 * Each type represents a different category of learning or documentation.
 */
public enum NoteType {
    /**
     * General learning from successful interactions.
     * Captures insights that don't fit other categories.
     */
    INSIGHT,

    /**
     * Failure + resolution documentation.
     * Critical for avoiding repeated mistakes.
     * KEY FEATURE from Confucius.
     */
    HINDSIGHT,

    /**
     * Discovered code pattern that works well.
     * Positive patterns to follow.
     */
    PATTERN,

    /**
     * Identified anti-pattern to avoid.
     * Negative patterns with warnings.
     */
    ANTIPATTERN,

    /**
     * Architectural decision made during session.
     * Records design choices with rationale.
     */
    ARCHITECTURE,

    /**
     * Integration knowledge between components.
     * How different parts of the system work together.
     */
    INTEGRATION
}
