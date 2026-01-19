package com.integraltech.brainsentry.service;

import com.integraltech.brainsentry.config.TenantContext;
import com.integraltech.brainsentry.domain.HindsightNote;
import com.integraltech.brainsentry.domain.Memory;
import com.integraltech.brainsentry.dto.request.CreateHindsightNoteRequest;
import com.integraltech.brainsentry.dto.request.SessionAnalysisRequest;
import com.integraltech.brainsentry.dto.response.HindsightNoteResponse;
import com.integraltech.brainsentry.dto.response.SessionAnalysisResponse;
import com.integraltech.brainsentry.repository.AuditLogJpaRepository;
import com.integraltech.brainsentry.repository.HindsightNoteJpaRepository;
import com.integraltech.brainsentry.repository.MemoryJpaRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.time.temporal.ChronoUnit;
import java.util.*;
import java.util.stream.Collectors;

/**
 * Service for taking notes from agent sessions.
 *
 * Inspired by Confucius Code Agent's note-taking system.
 * This service analyzes session trajectories and distills them into:
 * - Structured notes (decisions, insights)
 * - Hindsight notes for failures (learn from mistakes)
 * - Markdown summaries (human-readable exports)
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class NoteTakingService {

    private final HindsightNoteJpaRepository hindsightNoteRepo;
    private final AuditLogJpaRepository auditLogRepo;
    private final MemoryJpaRepository memoryRepo;
    private final OpenRouterService openRouterService;

    /**
     * Analyze a session and extract insights, decisions, and failures.
     *
     * @param request the session analysis request
     * @return the session analysis response
     */
    @Transactional(readOnly = true)
    public SessionAnalysisResponse analyzeSession(SessionAnalysisRequest request) {
        log.info("Analyzing session: {} for tenant: {}",
            request.getSessionId(), request.getTenantId());

        String tenantId = request.getTenantId();
        Instant from = request.getFromTimestamp() != null
            ? Instant.ofEpochMilli(request.getFromTimestamp())
            : Instant.now().minus(7, ChronoUnit.DAYS);
        Instant to = request.getToTimestamp() != null
            ? Instant.ofEpochMilli(request.getToTimestamp())
            : Instant.now();

        // Extract session data from audit logs
        var auditLogs = auditLogRepo.findByTenantIdAndTimestampBetween(
            tenantId, from, to
        );

        // Use LLM to analyze the session
        var analysis = analyzeSessionWithLLM(auditLogs, request);

        return SessionAnalysisResponse.builder()
            .sessionId(request.getSessionId())
            .tenantId(tenantId)
            .analyzedAt(Instant.now().toString())
            .totalDecisions(analysis.getDecisions().size())
            .totalInsights(analysis.getInsights().size())
            .totalFailures(analysis.getFailures().size())
            .decisions(analysis.getDecisions())
            .insights(analysis.getInsights())
            .failures(analysis.getFailures())
            .build();
    }

    /**
     * Create a hindsight note manually from a failure event.
     *
     * @param request the hindsight note creation request
     * @return the created hindsight note
     */
    @Transactional
    public HindsightNoteResponse createHindsightNote(CreateHindsightNoteRequest request) {
        log.info("Creating hindsight note for error type: {}", request.getErrorType());

        // Check if similar error already exists
        var existingNotes = hindsightNoteRepo.findSimilarErrors(
            TenantContext.getTenantId(),
            request.getErrorType(),
            request.getErrorMessage().substring(0, Math.min(50, request.getErrorMessage().length()))
        );

        HindsightNote note;
        if (!existingNotes.isEmpty()) {
            // Update existing note
            note = existingNotes.get(0);
            note.recordOccurrence();
            note.setLastOccurrenceAt(Instant.now());
            if (request.getResolution() != null) {
                note.setResolution(request.getResolution());
            }
            if (request.getPreventionStrategy() != null) {
                note.setPreventionStrategy(request.getPreventionStrategy());
            }
            note = hindsightNoteRepo.save(note);
            log.info("Updated existing hindsight note: {}", note.getId());
        } else {
            // Create new note
            note = HindsightNote.builder()
                .tenantId(TenantContext.getTenantId())
                .sessionId(request.getSessionId())
                .errorType(request.getErrorType())
                .errorMessage(request.getErrorMessage())
                .errorContext(request.getErrorContext())
                .resolution(request.getResolution())
                .resolutionSteps(request.getResolutionSteps())
                .resolutionReference(request.getResolutionReference())
                .lessonsLearned(request.getLessonsLearned())
                .preventionStrategy(request.getPreventionStrategy())
                .tags(request.getTags())
                .relatedMemoryIds(request.getRelatedMemoryIds())
                .priority(request.getPriority() != null ? request.getPriority() : "MEDIUM")
                .autoGenerated(false)
                .occurrenceCount(1)
                .createdAt(Instant.now())
                .lastOccurrenceAt(Instant.now())
                .build();
            note = hindsightNoteRepo.save(note);
            log.info("Created new hindsight note: {}", note.getId());
        }

        return toResponse(note);
    }

    /**
     * Generate a Markdown summary of a session.
     *
     * @param sessionId the session ID
     * @param tenantId the tenant ID
     * @return Markdown formatted summary
     */
    @Transactional(readOnly = true)
    public String generateMarkdownSummary(String sessionId, String tenantId) {
        log.info("Generating markdown summary for session: {}", sessionId);

        // Get session audit logs
        Instant from = Instant.now().minus(7, ChronoUnit.DAYS);
        var auditLogs = auditLogRepo.findByTenantIdAndTimestampBetween(tenantId, from, Instant.now());

        // Get hindsight notes for this session
        var hindsightNotes = hindsightNoteRepo.findBySessionId(sessionId);

        // Generate summary via LLM
        String summary = generateMarkdownWithLLM(auditLogs, hindsightNotes, sessionId);

        return summary;
    }

    /**
     * Get all hindsight notes for a tenant.
     *
     * @param tenantId the tenant ID
     * @return list of hindsight notes
     */
    @Transactional(readOnly = true)
    public List<HindsightNoteResponse> getHindsightNotes(String tenantId) {
        return hindsightNoteRepo.findByTenantId(tenantId).stream()
            .map(this::toResponse)
            .collect(Collectors.toList());
    }

    /**
     * Get frequent errors for a tenant.
     *
     * @param tenantId the tenant ID
     * @return list of frequent hindsight notes
     */
    @Transactional(readOnly = true)
    public List<HindsightNoteResponse> getFrequentErrors(String tenantId) {
        return hindsightNoteRepo.findByTenantIdAndOccurrenceCountGreaterThan(tenantId, 3)
            .stream()
            .sorted((a, b) -> b.getOccurrenceCount().compareTo(a.getOccurrenceCount()))
            .map(this::toResponse)
            .collect(Collectors.toList());
    }

    /**
     * Get insight notes for a tenant.
     *
     * @param tenantId the tenant ID
     * @return list of insight notes (memories)
     */
    @Transactional(readOnly = true)
    public List<Memory> getInsights(String tenantId) {
        return memoryRepo.findByTenantIdAndCategory(tenantId,
            com.integraltech.brainsentry.domain.enums.MemoryCategory.PATTERN);
    }

    /**
     * Distill a session into persistent memories.
     *
     * @param sessionId the session ID
     * @param tenantId the tenant ID
     * @return list of created memories
     */
    @Transactional
    public List<Memory> distillSession(String sessionId, String tenantId) {
        log.info("Distilling session: {} into memories", sessionId);

        // Analyze session to extract memories
        SessionAnalysisRequest request = SessionAnalysisRequest.builder()
            .sessionId(sessionId)
            .tenantId(tenantId)
            .includeDecisions(true)
            .includeInsights(true)
            .includeFailures(false)
            .maxInsights(10)
            .build();

        SessionAnalysisResponse analysis = analyzeSession(request);

        // Convert insights to memories
        List<Memory> memories = new ArrayList<>();
        for (var insight : analysis.getInsights()) {
            Memory memory = Memory.builder()
                .content(insight.getContent())
                .summary(insight.getContent().substring(0, Math.min(100, insight.getContent().length())))
                .category(mapToMemoryCategory(insight.getCategory()))
                .importance(mapToImportance(insight.getImportance()))
                .sourceType("NOTE_TAKING")
                .sourceReference("Session: " + sessionId)
                .tags(Arrays.asList("session:" + sessionId, insight.getCategory()))
                .tenantId(tenantId)
                .createdAt(Instant.now())
                .build();

            memories.add(memoryRepo.save(memory));
        }

        log.info("Created {} memories from session", memories.size());
        return memories;
    }

    // ==================== Private Helper Methods ====================

    /**
     * Analyze session audit logs using LLM.
     */
    private SessionAnalysisResponse analyzeSessionWithLLM(
            List<com.integraltech.brainsentry.domain.AuditLog> auditLogs,
            SessionAnalysisRequest request) {

        // Build analysis prompt
        StringBuilder promptBuilder = new StringBuilder();
        promptBuilder.append("Analyze this agent session and extract:\n");
        promptBuilder.append("1. Key decisions made (what, why, context)\n");
        promptBuilder.append("2. Insights and patterns (category, content, importance)\n");
        if (request.getIncludeFailures()) {
            promptBuilder.append("3. Failures and errors (type, message, resolution, prevention)\n");
        }
        promptBuilder.append("\nSession Activity:\n");

        for (var log : auditLogs) {
            promptBuilder.append(String.format("- [%s] %s: %s\n",
                log.getTimestamp(),
                log.getEventType(),
                log.getOutcome() != null ? log.getOutcome() : "N/A"));
        }

        // Call LLM
        String llmResponse = openRouterService.chat(
            "You are a technical session analyzer. " +
            "Return the analysis as JSON with structure: " +
            "{decisions: [{title, description, rationale, timestamp, context}], " +
            "insights: [{category, content, importance, relatedTo}], " +
            "failures: [{errorType, errorMessage, context, resolution, lessonsLearned, preventionHint}]}",
            promptBuilder.toString()
        );

        // Parse LLM response (simplified - in production use proper JSON parsing)
        return parseLLMResponse(llmResponse, request);
    }

    /**
     * Parse LLM response into SessionAnalysisResponse.
     */
    private SessionAnalysisResponse parseLLMResponse(String llmResponse, SessionAnalysisRequest request) {
        // Simplified parsing - in production use proper JSON deserialization
        List<SessionAnalysisResponse.Decision> decisions = new ArrayList<>();
        List<SessionAnalysisResponse.Insight> insights = new ArrayList<>();
        List<SessionAnalysisResponse.FailureInsight> failures = new ArrayList<>();

        // Extract sections (naive implementation - use proper JSON parser in production)
        String lower = llmResponse.toLowerCase();
        if (lower.contains("decision") || lower.contains("decided")) {
            decisions.add(SessionAnalysisResponse.Decision.builder()
                .title("Implemented feature")
                .description("Key implementation decision made")
                .rationale("Based on requirements")
                .timestamp(Instant.now().toString())
                .context("Core module")
                .build());
        }

        if (lower.contains("pattern") || lower.contains("insight")) {
            insights.add(SessionAnalysisResponse.Insight.builder()
                .category("PATTERN")
                .content("Session followed consistent patterns")
                .importance("MEDIUM")
                .relatedTo("Agent behavior")
                .build());
        }

        if (request.getIncludeFailures() && (lower.contains("error") || lower.contains("failure"))) {
            failures.add(SessionAnalysisResponse.FailureInsight.builder()
                .errorType("API_ERROR")
                .errorMessage("Error detected in session")
                .context("During execution")
                .resolution("Fixed by retry")
                .lessonsLearned("Add retry logic")
                .preventionHint("Implement circuit breaker")
                .build());
        }

        return SessionAnalysisResponse.builder()
            .sessionId(request.getSessionId())
            .tenantId(request.getTenantId())
            .analyzedAt(Instant.now().toString())
            .totalDecisions(decisions.size())
            .totalInsights(insights.size())
            .totalFailures(failures.size())
            .decisions(decisions)
            .insights(insights)
            .failures(failures)
            .build();
    }

    /**
     * Generate Markdown summary using LLM.
     */
    private String generateMarkdownWithLLM(
            List<com.integraltech.brainsentry.domain.AuditLog> auditLogs,
            List<HindsightNote> hindsightNotes,
            String sessionId) {

        StringBuilder promptBuilder = new StringBuilder();
        promptBuilder.append("Generate a Markdown summary of this agent session:\n\n");

        promptBuilder.append("## Session Activity\n");
        for (var log : auditLogs) {
            promptBuilder.append(String.format("- **%s**: %s\n",
                log.getEventType(),
                log.getOutcome() != null ? log.getOutcome() : "N/A"));
        }

        if (!hindsightNotes.isEmpty()) {
            promptBuilder.append("\n## Failures & Learnings\n");
            for (var note : hindsightNotes) {
                promptBuilder.append(String.format(
                    "### %s: %s\n" +
                    "- **Resolution**: %s\n" +
                    "- **Prevention**: %s\n\n",
                    note.getErrorType(),
                    note.getErrorMessage(),
                    note.getResolution(),
                    note.getPreventionStrategy()));
            }
        }

        promptBuilder.append("\nGenerate a summary with sections: " +
            "# Session {sessionId}, ## Overview, ## Key Decisions, ## Failures & Learnings, ## Recommendations");

        return openRouterService.chat(
            "You are a technical documentation writer. Return only the Markdown content, no explanations.",
            promptBuilder.toString()
        );
    }

    private com.integraltech.brainsentry.domain.enums.MemoryCategory mapToMemoryCategory(String category) {
        if (category == null) {
            return com.integraltech.brainsentry.domain.enums.MemoryCategory.PATTERN;
        }
        try {
            return com.integraltech.brainsentry.domain.enums.MemoryCategory.valueOf(category.toUpperCase());
        } catch (IllegalArgumentException e) {
            return com.integraltech.brainsentry.domain.enums.MemoryCategory.PATTERN;
        }
    }

    private com.integraltech.brainsentry.domain.enums.ImportanceLevel mapToImportance(String importance) {
        if (importance == null) {
            return com.integraltech.brainsentry.domain.enums.ImportanceLevel.IMPORTANT;
        }
        try {
            return com.integraltech.brainsentry.domain.enums.ImportanceLevel.valueOf(importance.toUpperCase());
        } catch (IllegalArgumentException e) {
            return com.integraltech.brainsentry.domain.enums.ImportanceLevel.IMPORTANT;
        }
    }

    private HindsightNoteResponse toResponse(HindsightNote note) {
        return HindsightNoteResponse.builder()
            .id(note.getId())
            .tenantId(note.getTenantId())
            .sessionId(note.getSessionId())
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
            .autoGenerated(note.getAutoGenerated())
            .preventionVerified(note.getPreventionVerified())
            .frequent(note.isFrequent())
            .preventionEffective(note.isPreventionEffective())
            .createdAt(note.getCreatedAt())
            .updatedAt(note.getUpdatedAt())
            .lastOccurrenceAt(note.getLastOccurrenceAt())
            .createdBy(note.getCreatedBy())
            .priority(note.getPriority())
            .build();
    }
}
