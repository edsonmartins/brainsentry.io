package com.integraltech.brainsentry.dto.request;

import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotNull;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

/**
 * Request to compress conversation context.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CompressionRequest {

    @NotNull
    private List<Message> messages;

    /**
     * Token threshold to trigger compression.
     */
    @Builder.Default
    @Min(1000)
    private Integer tokenThreshold = 100000;

    /**
     * Number of recent messages to always preserve.
     */
    @Builder.Default
    @Min(1)
    private Integer preserveRecent = 10;

    /**
     * Target compression ratio (0.0-1.0).
     * e.g., 0.5 means compress to 50% of original size.
     */
    @Builder.Default
    private Double targetRatio = 0.5;

    /**
     * Context hint for compression (e.g., "debugging", "implementation").
     */
    private String contextHint;

    /**
     * Specific items to always preserve.
     */
    private List<String> preserveKeywords;

    /**
     * Message representation for compression.
     */
    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Message {
        private String role; // user, assistant, system, tool
        private String content;
        private Long timestamp;
        private String toolName; // if role is "tool"
    }
}
