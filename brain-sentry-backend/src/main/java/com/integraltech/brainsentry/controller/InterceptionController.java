package com.integraltech.brainsentry.controller;

import com.integraltech.brainsentry.dto.request.InterceptRequest;
import com.integraltech.brainsentry.dto.response.InterceptResponse;
import com.integraltech.brainsentry.service.InterceptionService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

/**
 * REST controller for prompt interception.
 *
 * This is the main entry point for the Brain Sentry system.
 * Client applications send prompts here to get them enhanced
 * with relevant memory context.
 */
@Slf4j
@RestController
@RequestMapping("/v1/intercept")
@RequiredArgsConstructor
public class InterceptionController {

    private final InterceptionService interceptionService;

    /**
     * Intercept and enhance a prompt.
     * POST /api/v1/intercept
     *
     * This endpoint analyzes the prompt, searches for relevant memories,
     * and returns an enhanced prompt with context injected.
     */
    @PostMapping
    public ResponseEntity<InterceptResponse> intercept(
        @Valid @RequestBody InterceptRequest request
    ) {
        log.info("POST /v1/intercept - sessionId: {}, promptLength: {}",
            request.getSessionId(),
            request.getPrompt() != null ? request.getPrompt().length() : 0);

        InterceptResponse response = interceptionService.interceptAndEnhance(request);

        log.info("Response: enhanced={}, memoriesUsed={}, latencyMs={}",
            response.getEnhanced(),
            response.getMemoriesUsed() != null ? response.getMemoriesUsed().size() : 0,
            response.getLatencyMs());

        return ResponseEntity.ok(response);
    }
}
