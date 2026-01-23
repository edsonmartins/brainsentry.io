package com.integraltech.brainsentry.dto.response;

import lombok.AllArgsConstructor;
import lombok.Data;

/**
 * Response representing a relationship between memories in the graph.
 */
@Data
@AllArgsConstructor
public class GraphRelationshipResponse {
    private String id;
    private String fromMemoryId;
    private String toMemoryId;
    private String fromMemorySummary;
    private String toMemorySummary;
    private String type;
    private Double strength;
    private String tag; // The shared tag that created this relationship
}
