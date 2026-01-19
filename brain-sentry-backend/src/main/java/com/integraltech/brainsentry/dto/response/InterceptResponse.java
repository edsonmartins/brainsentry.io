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
}
