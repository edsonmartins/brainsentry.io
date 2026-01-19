package com.integraltech.brainsentry.dto.response;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

/**
 * Response containing compressed context.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CompressedContextResponse {

    // Statistics
    private Integer originalMessageCount;
    private Integer compressedMessageCount;
    private Integer originalTokenCount;
    private Integer compressedTokenCount;
    private Double compressionRatio;
    private Boolean compressed;

    // Compressed content
    private StructuredSummary summary;
    private List<Message> preservedMessages;

    /**
     * Structured summary of the conversation.
     */
    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class StructuredSummary {
        private String taskGoal;
        private List<String> keyDecisions;
        private List<String> openTodos;
        private List<CriticalError> criticalErrors;
        private List<String> importantFileChanges;
        private String additionalContext;
    }

    /**
     * A critical error that must be preserved.
     */
    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class CriticalError {
        private String errorType;
        private String description;
        private String resolution;
    }

    /**
     * Preserved message from original history.
     */
    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Message {
        private String role;
        private String content;
        private Long timestamp;
        private Boolean isSummary;
    }
}
