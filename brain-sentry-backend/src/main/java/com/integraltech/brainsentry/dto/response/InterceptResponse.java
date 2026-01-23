package com.integraltech.brainsentry.dto.response;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

/**
 * Response from prompt interception.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class InterceptResponse {

    /**
     * Whether the prompt was enhanced.
     */
    private Boolean enhanced;

    /**
     * The original prompt.
     */
    private String originalPrompt;

    /**
     * The enhanced prompt with context injected.
     */
    private String enhancedPrompt;

    /**
     * Formatted context that was injected (for display).
     */
    private String contextInjected;

    /**
     * Memories that were used for context.
     */
    private List<MemoryReference> memoriesUsed;

    /**
     * Notes that were used for context (Confucius integration).
     */
    private List<NoteReference> notesUsed;

    /**
     * Operation latency in milliseconds.
     */
    private Integer latencyMs;

    /**
     * Reasoning for the enhancement decision.
     */
    private String reasoning;

    /**
     * Confidence score of the relevance analysis.
     */
    private Double confidence;

    /**
     * Number of tokens in the injected context.
     */
    private Integer tokensInjected;

    /**
     * Number of LLM API calls made.
     */
    private Integer llmCalls;

    /**
     * Reference to a memory used in context.
     */
    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class MemoryReference {
        private String id;
        private String summary;
        private String category;
        private String importance;
        private Double relevanceScore;
        private String excerpt;
    }

    /**
     * Reference to a note used in context (Confucius integration).
     */
    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class NoteReference {
        private String id;
        private String title;
        private String type;  // NoteType: HINDSIGHT, INSIGHT, PATTERN, etc.
        private String severity;  // For hindsight notes
        private String excerpt;
    }
}
