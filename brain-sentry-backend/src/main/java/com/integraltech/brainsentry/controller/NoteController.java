package com.integraltech.brainsentry.controller;

import com.integraltech.brainsentry.config.TenantContext;
import com.integraltech.brainsentry.domain.HindsightNote;
import com.integraltech.brainsentry.domain.enums.NoteSeverity;
import com.integraltech.brainsentry.dto.response.HindsightNoteResponse;
import com.integraltech.brainsentry.repository.HindsightNoteJpaRepository;
import com.integraltech.brainsentry.service.NoteRetrievalService;
import com.integraltech.brainsentry.service.NoteTakingService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * REST controller for note retrieval operations.
 *
 * Provides endpoints for searching and retrieving notes
 * based on patterns, error types, and severity levels.
 *
 * KEY FEATURE from Confucius: Proactive error detection via pattern matching.
 */
@Slf4j
@RestController
@RequestMapping("/v1/notes")
@RequiredArgsConstructor
public class NoteController {

    private final NoteRetrievalService noteRetrievalService;
    private final NoteTakingService noteTakingService;
    private final HindsightNoteJpaRepository hindsightNoteRepo;

    /**
     * Search hindsight notes by error message and/or error type.
     * Uses pattern matching for fast results.
     * POST /api/v1/notes/search
     */
    @PostMapping("/search")
    public ResponseEntity<List<HindsightNoteResponse>> searchNotes(
        @RequestParam String tenantId,
        @RequestParam(required = false) String errorMessage,
        @RequestParam(required = false) String errorType
    ) {
        log.info("POST /v1/notes/search - tenant: {}, errorType: {}", tenantId, errorType);

        TenantContext.setTenantId(tenantId);

        List<HindsightNote> notes = noteRetrievalService.searchHindsightNotes(
            errorMessage, errorType, tenantId
        );

        // Record access for each note
        for (HindsightNote note : notes) {
            noteRetrievalService.recordNoteAccess(note.getId());
        }

        return ResponseEntity.ok(notes.stream()
            .map(this::toResponse)
            .toList());
    }

    /**
     * Get relevant notes for a query.
     * Used during autonomous interception.
     * GET /api/v1/notes/relevant
     */
    @GetMapping("/relevant")
    public ResponseEntity<List<HindsightNoteResponse>> getRelevantNotes(
        @RequestParam String tenantId,
        @RequestParam String query,
        @RequestParam(defaultValue = "10") int limit
    ) {
        log.info("GET /v1/notes/relevant - tenant: {}, query: {}", tenantId, query);

        TenantContext.setTenantId(tenantId);

        List<HindsightNote> notes = noteRetrievalService.getRelevantNotes(query, tenantId, limit);

        return ResponseEntity.ok(notes.stream()
            .map(this::toResponse)
            .toList());
    }

    /**
     * Get critical errors for a tenant.
     * CRITICAL and HIGH severity notes.
     * GET /api/v1/notes/critical
     */
    @GetMapping("/critical")
    public ResponseEntity<List<HindsightNoteResponse>> getCriticalErrors(
        @RequestParam String tenantId
    ) {
        log.info("GET /v1/notes/critical - tenant: {}", tenantId);

        TenantContext.setTenantId(tenantId);

        List<HindsightNote> notes = noteRetrievalService.getCriticalErrors(tenantId);

        return ResponseEntity.ok(notes.stream()
            .map(this::toResponse)
            .toList());
    }

    /**
     * Get notes by severity level.
     * GET /api/v1/notes/by-severity
     */
    @GetMapping("/by-severity")
    public ResponseEntity<List<HindsightNoteResponse>> getNotesBySeverity(
        @RequestParam String tenantId,
        @RequestParam NoteSeverity severity
    ) {
        log.info("GET /v1/notes/by-severity - tenant: {}, severity: {}", tenantId, severity);

        TenantContext.setTenantId(tenantId);

        List<HindsightNote> notes = hindsightNoteRepo.findByTenantIdAndSeverity(tenantId, severity);

        return ResponseEntity.ok(notes.stream()
            .map(this::toResponse)
            .toList());
    }

    /**
     * Get most accessed notes for a tenant.
     * Notes with highest access count.
     * GET /api/v1/notes/most-accessed
     */
    @GetMapping("/most-accessed")
    public ResponseEntity<List<HindsightNoteResponse>> getMostAccessedNotes(
        @RequestParam String tenantId,
        @RequestParam(defaultValue = "10") int limit
    ) {
        log.info("GET /v1/notes/most-accessed - tenant: {}, limit: {}", tenantId, limit);

        TenantContext.setTenantId(tenantId);

        List<HindsightNote> notes = hindsightNoteRepo.findByTenantId(tenantId).stream()
            .sorted((a, b) -> {
                int accessCompare = b.getAccessCount().compareTo(a.getAccessCount());
                if (accessCompare != 0) return accessCompare;
                return (b.getLastAccessedAt() != null ? b.getLastAccessedAt() : a.getCreatedAt())
                    .compareTo(a.getLastAccessedAt() != null ? a.getLastAccessedAt() : a.getCreatedAt());
            })
            .limit(limit)
            .toList();

        return ResponseEntity.ok(notes.stream()
            .map(this::toResponse)
            .toList());
    }

    /**
     * Check if an error matches any existing hindsight note pattern.
     * KEY FEATURE from Confucius - proactive error detection.
     * POST /api/v1/notes/match
     */
    @PostMapping("/match")
    public ResponseEntity<List<HindsightNoteResponse>> matchError(
        @RequestParam String tenantId,
        @RequestParam String errorMessage,
        @RequestParam(required = false) String errorType
    ) {
        log.info("POST /v1/notes/match - tenant: {}, errorType: {}", tenantId, errorType);

        TenantContext.setTenantId(tenantId);

        List<HindsightNote> matchingNotes = noteRetrievalService.searchHindsightNotes(
            errorMessage, errorType, tenantId
        );

        // Record that these notes were suggested
        for (HindsightNote note : matchingNotes) {
            noteRetrievalService.recordNoteAccess(note.getId());
        }

        return ResponseEntity.ok(matchingNotes.stream()
            .map(this::toResponse)
            .toList());
    }

    /**
     * Get error type statistics for a tenant.
     * Shows which errors occur most frequently.
     * GET /api/v1/notes/stats
     */
    @GetMapping("/stats")
    public ResponseEntity<Map<String, Object>> getErrorStats(
        @RequestParam String tenantId
    ) {
        log.info("GET /v1/notes/stats - tenant: {}", tenantId);

        TenantContext.setTenantId(tenantId);

        List<Object[]> stats = hindsightNoteRepo.countByErrorType(tenantId);

        // Format stats for response
        Map<String, Object> response = new HashMap<>();
        response.put("tenantId", tenantId);

        Map<String, Long> errorCounts = new HashMap<>();
        for (Object[] row : stats) {
            errorCounts.put((String) row[0], (Long) row[1]);
        }
        response.put("errorTypeCounts", errorCounts);

        // Additional stats
        response.put("totalNotes", hindsightNoteRepo.findByTenantId(tenantId).size());
        response.put("frequentErrors", noteRetrievalService.getFrequentErrors(tenantId).size());

        return ResponseEntity.ok(response);
    }

    /**
     * Convert HindsightNote entity to response DTO.
     */
    private HindsightNoteResponse toResponse(HindsightNote note) {
        return HindsightNoteResponse.builder()
            .id(note.getId())
            .tenantId(note.getTenantId())
            .sessionId(note.getSessionId())
            .title(note.getTitle())
            .errorType(note.getErrorType())
            .errorMessage(note.getErrorMessage())
            .errorContext(note.getErrorContext())
            .resolution(note.getResolution())
            .resolutionSteps(note.getResolutionSteps())
            .resolutionReference(note.getResolutionReference())
            .lessonsLearned(note.getLessonsLearned())
            .preventionStrategy(note.getPreventionStrategy())
            .tags(note.getTags())
            .relatedMemoryIds(note.getRelatedMemoryIds())
            .occurrenceCount(note.getOccurrenceCount())
            .referenceCount(note.getReferenceCount())
            .preventionSuccessCount(note.getPreventionSuccessCount())
            .preventionEffectiveness(note.getPreventionEffectiveness())
            .createdBy(note.getCreatedBy())
            .autoGenerated(note.getAutoGenerated())
            .createdAt(note.getCreatedAt())
            .updatedAt(note.getUpdatedAt())
            .lastOccurrenceAt(note.getLastOccurrenceAt())
            .preventionVerified(note.getPreventionVerified())
            .priority(note.getPriority())
            .severity(note.getSeverity() != null ? note.getSeverity().getDisplayName() : null)
            .errorPattern(note.getErrorPattern())
            .lastAccessedAt(note.getLastAccessedAt() != null ? note.getLastAccessedAt().toString() : null)
            .accessCount(note.getAccessCount())
            .frequent(note.isFrequent())
            .preventionEffective(note.isPreventionEffective())
            .build();
    }
}
