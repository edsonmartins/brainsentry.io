package com.integraltech.brainsentry.domain.enums;

import com.fasterxml.jackson.annotation.JsonValue;

/**
 * Types of relationships between memories.
 *
 * Relationships enable the graph-based memory retrieval system,
 * allowing the system to find related concepts and patterns.
 */
public enum RelationshipType {
    /**
     * Two memories are frequently used together.
     * Example: UserService USED_WITH UserRepository
     */
    USED_WITH("Used With"),

    /**
     * One memory contradicts another.
     * Example: "Use JPA" CONFLICTS_WITH "Use MyBatis"
     */
    CONFLICTS_WITH("Conflicts With"),

    /**
     * One memory supersedes or replaces another.
     * Example: "Use virtual threads" SUPERSEDES "Use platform threads"
     */
    SUPERSEDES("Supersedes"),

    /**
     * General relatedness without specific type.
     * Example: "Spring Boot" RELATED_TO "Spring Framework"
     */
    RELATED_TO("Related To"),

    /**
     * One memory requires another to function.
     * Example: "Transactional caching" REQUIRES "Transaction management"
     */
    REQUIRES("Requires"),

    /**
     * One memory is a component or part of another.
     * Example: "ExceptionHandler" PART_OF "Global error handling"
     */
    PART_OF("Part Of");

    private final String displayName;

    RelationshipType(String displayName) {
        this.displayName = displayName;
    }

    @JsonValue
    public String getDisplayName() {
        return displayName;
    }
}
