package com.integraltech.brainsentry.dto.response;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.Map;

/**
 * Response containing system statistics.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class StatsResponse {

    /**
     * Total number of memories.
     */
    private Long totalMemories;

    /**
     * Memories grouped by category.
     */
    private Map<String, Long> memoriesByCategory;

    /**
     * Memories grouped by importance level.
     */
    private Map<String, Long> memoriesByImportance;

    /**
     * Number of requests processed today.
     */
    private Long requestsToday;

    /**
     * Rate of context injection (0.0 to 1.0).
     */
    private Double injectionRate;

    /**
     * Average latency in milliseconds.
     */
    private Double avgLatencyMs;

    /**
     * Average helpfulness rate.
     */
    private Double helpfulnessRate;

    /**
     * Total number of context injections.
     */
    private Long totalInjections;

    /**
     * Unique memories used in last 24 hours.
     */
    private Long activeMemories24h;
}
