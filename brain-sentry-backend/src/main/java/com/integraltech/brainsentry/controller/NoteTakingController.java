package com.integraltech.brainsentry.controller;

import com.integraltech.brainsentry.config.TenantContext;
import com.integraltech.brainsentry.domain.Memory;
import com.integraltech.brainsentry.dto.request.CreateHindsightNoteRequest;
import com.integraltech.brainsentry.dto.request.SessionAnalysisRequest;
import com.integraltech.brainsentry.dto.response.HindsightNoteResponse;
import com.integraltech.brainsentry.dto.response.SessionAnalysisResponse;
import com.integraltech.brainsentry.service.NoteTakingService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

/**
 * REST controller for note-taking operations.
 *
 * Inspired by Confucius Code Agent's note-taking system.
 * Provides endpoints for analyzing sessions, generating insights,
 * and creating hindsight notes for failures.
 */
@Slf4j
@RestController
@RequestMapping("/v1/notes")
@RequiredArgsConstructor
public class NoteTakingController {

    private final NoteTakingService noteTakingService;

    /**
     * Analyze a session and extract insights, decisions, and failures.
     * POST /api/v1/notes/analyze
     */
    @PostMapping("/analyze")
    public ResponseEntity<SessionAnalysisResponse> analyzeSession(
        @Valid @RequestBody SessionAnalysisRequest request
    ) {
        log.info("POST /v1/notes/analyze - session: {}", request.getSessionId());

        // Set tenant context
        TenantContext.setTenantId(request.getTenantId());

        SessionAnalysisResponse response = noteTakingService.analyzeSession(request);
        return ResponseEntity.ok(response);
    }

    /**
     * Get notes for a specific session.
     * GET /api/v1/notes/session/{sessionId}
     */
    @GetMapping("/session/{sessionId}")
    public ResponseEntity<SessionAnalysisResponse> getSessionNotes(
        @PathVariable String sessionId,
        @RequestParam String tenantId
    ) {
        log.info("GET /v1/notes/session/{}", sessionId);

        TenantContext.setTenantId(tenantId);

        SessionAnalysisRequest request = SessionAnalysisRequest.builder()
            .sessionId(sessionId)
            .tenantId(tenantId)
            .includeDecisions(true)
            .includeInsights(true)
            .includeFailures(true)
            .build();

        SessionAnalysisResponse response = noteTakingService.analyzeSession(request);
        return ResponseEntity.ok(response);
    }

    /**
     * Export session notes as Markdown.
     * GET /api/v1/notes/session/{sessionId}/md
     */
    @GetMapping(value = "/session/{sessionId}/md", produces = MediaType.TEXT_MARKDOWN_VALUE)
    public ResponseEntity<String> exportSessionMarkdown(
        @PathVariable String sessionId,
        @RequestParam String tenantId
    ) {
        log.info("GET /v1/notes/session/{}/md", sessionId);

        TenantContext.setTenantId(tenantId);

        String markdown = noteTakingService.generateMarkdownSummary(sessionId, tenantId);
        return ResponseEntity.ok()
            .header("Content-Disposition", "attachment; filename=\"session-" + sessionId + ".md\"")
            .body(markdown);
    }

    /**
     * Create a hindsight note manually.
     * POST /api/v1/notes/hindsight
     */
    @PostMapping("/hindsight")
    public ResponseEntity<HindsightNoteResponse> createHindsightNote(
        @Valid @RequestBody CreateHindsightNoteRequest request
    ) {
        log.info("POST /v1/notes/hindsight - error type: {}", request.getErrorType());

        HindsightNoteResponse response = noteTakingService.createHindsightNote(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }

    /**
     * Get all hindsight notes for a tenant.
     * GET /api/v1/notes/hindsight
     */
    @GetMapping("/hindsight")
    public ResponseEntity<List<HindsightNoteResponse>> getHindsightNotes(
        @RequestParam String tenantId
    ) {
        log.info("GET /v1/notes/hindsight - tenant: {}", tenantId);

        TenantContext.setTenantId(tenantId);

        List<HindsightNoteResponse> notes = noteTakingService.getHindsightNotes(tenantId);
        return ResponseEntity.ok(notes);
    }

    /**
     * Get frequent errors for a tenant.
     * GET /api/v1/notes/hindsight/frequent
     */
    @GetMapping("/hindsight/frequent")
    public ResponseEntity<List<HindsightNoteResponse>> getFrequentErrors(
        @RequestParam String tenantId
    ) {
        log.info("GET /v1/notes/hindsight/frequent - tenant: {}", tenantId);

        TenantContext.setTenantId(tenantId);

        List<HindsightNoteResponse> notes = noteTakingService.getFrequentErrors(tenantId);
        return ResponseEntity.ok(notes);
    }

    /**
     * Get insight notes for a tenant.
     * GET /api/v1/notes/insights
     */
    @GetMapping("/insights")
    public ResponseEntity<List<Memory>> getInsights(
        @RequestParam String tenantId
    ) {
        log.info("GET /v1/notes/insights - tenant: {}", tenantId);

        TenantContext.setTenantId(tenantId);

        List<Memory> insights = noteTakingService.getInsights(tenantId);
        return ResponseEntity.ok(insights);
    }

    /**
     * Distill a session into persistent memories.
     * POST /api/v1/notes/distill
     */
    @PostMapping("/distill")
    public ResponseEntity<List<Memory>> distillSession(
        @RequestParam String sessionId,
        @RequestParam String tenantId
    ) {
        log.info("POST /v1/notes/distill - session: {}", sessionId);

        TenantContext.setTenantId(tenantId);

        List<Memory> memories = noteTakingService.distillSession(sessionId, tenantId);
        return ResponseEntity.ok(memories);
    }
}
