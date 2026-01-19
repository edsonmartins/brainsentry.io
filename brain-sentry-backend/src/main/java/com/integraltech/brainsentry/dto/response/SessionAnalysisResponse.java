package com.integraltech.brainsentry.dto.response;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

/**
 * Response containing session analysis results.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class SessionAnalysisResponse {

    private String sessionId;
    private String tenantId;
    private String analyzedAt;

    // Summary statistics
    private Integer totalDecisions;
    private Integer totalInsights;
    private Integer totalFailures;

    // Extracted content
    private List<Decision> decisions;
    private List<Insight> insights;
    private List<FailureInsight> failures;

    // Generated notes
    private List<HindsightNoteResponse> hindsightNotes;

    /**
     * Represents a decision made during the session.
     */
    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Decision {
        private String title;
        private String description;
        private String rationale;
        private String timestamp;
        private String context; // e.g., file, component
    }

    /**
     * Represents an insight extracted from the session.
     */
    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Insight {
        private String category; // PATTERN, ANTIPATTERN, INTEGRATION, etc.
        private String content;
        private String importance; // HIGH, MEDIUM, LOW
        private String relatedTo; // e.g., "UserService.java"
    }

    /**
     * Represents a failure with learning opportunity.
     */
    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class FailureInsight {
        private String errorType;
        private String errorMessage;
        private String context;
        private String resolution;
        private String lessonsLearned;
        private String preventionHint;
    }
}
