package com.integraltech.brainsentry.dto.request;

import jakarta.validation.constraints.NotBlank;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

/**
 * Request to create a hindsight note manually.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CreateHindsightNoteRequest {

    @NotBlank
    private String sessionId;

    @NotBlank
    private String errorType;

    @NotBlank
    private String errorMessage;

    private String errorContext;

    private String resolution;

    private String resolutionSteps;

    private String resolutionReference;

    private String lessonsLearned;

    private String preventionStrategy;

    private List<String> tags;

    private List<String> relatedMemoryIds;

    private String priority; // HIGH, MEDIUM, LOW
}
