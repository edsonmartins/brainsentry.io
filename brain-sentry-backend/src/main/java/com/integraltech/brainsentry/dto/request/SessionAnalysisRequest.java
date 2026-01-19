package com.integraltech.brainsentry.dto.request;

import jakarta.validation.constraints.NotBlank;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * Request to analyze a session and extract notes.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class SessionAnalysisRequest {

    @NotBlank
    private String sessionId;

    @NotBlank
    private String tenantId;

    /**
     * Whether to generate hindsight notes for failures.
     */
    @Builder.Default
    private Boolean includeFailures = true;

    /**
     * Whether to extract decisions made.
     */
    @Builder.Default
    private Boolean includeDecisions = true;

    /**
     * Whether to extract general insights.
     */
    @Builder.Default
    private Boolean includeInsights = true;

    /**
     * Maximum number of insights to extract.
     */
    @Builder.Default
    private Integer maxInsights = 10;

    /**
     * Optional: Time range to analyze (from timestamp).
     */
    private Long fromTimestamp;

    /**
     * Optional: Time range to analyze (to timestamp).
     */
    private Long toTimestamp;
}
