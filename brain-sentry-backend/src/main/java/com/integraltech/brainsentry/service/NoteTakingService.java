package com.integraltech.brainsentry.service;

import com.integraltech.brainsentry.config.TenantContext;
import com.integraltech.brainsentry.domain.HindsightNote;
import com.integraltech.brainsentry.domain.Memory;
import com.integraltech.brainsentry.domain.Note;
import com.integraltech.brainsentry.domain.enums.NoteCategory;
import com.integraltech.brainsentry.domain.enums.NoteSeverity;
import com.integraltech.brainsentry.domain.enums.NoteType;
import com.integraltech.brainsentry.dto.request.CreateHindsightNoteRequest;
import com.integraltech.brainsentry.dto.request.SessionAnalysisRequest;
import com.integraltech.brainsentry.dto.response.HindsightNoteResponse;
import com.integraltech.brainsentry.dto.response.SessionAnalysisResponse;
import com.integraltech.brainsentry.repository.AuditLogJpaRepository;
import com.integraltech.brainsentry.repository.HindsightNoteJpaRepository;
import com.integraltech.brainsentry.repository.MemoryJpaRepository;
import com.integraltech.brainsentry.repository.NoteJpaRepository;
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
    private final NoteJpaRepository noteRepo;
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

        // Only add entries if the LLM response has actual content, not just empty arrays
        // Check for non-empty decisions (look for content after "decisions:" that's not just [])
        if (lower.contains("decision") && !lower.matches(".*decisions:\\s*\\[\\].*")) {
            // Look for actual decision content, not just the word "decisions"
            if (lower.contains("decided") || lower.matches(".*decisions:\\s*\\[.*\\{.*")) {
                decisions.add(SessionAnalysisResponse.Decision.builder()
                    .title("Implemented feature")
                    .description("Key implementation decision made")
                    .rationale("Based on requirements")
                    .timestamp(Instant.now().toString())
                    .context("Core module")
                    .build());
            }
        }

        // Similarly for insights - only add if there's actual content
        if (lower.contains("insight") && !lower.matches(".*insights:\\s*\\[\\].*")) {
            if (lower.contains("pattern") || lower.matches(".*insights:\\s*\\[.*\\{.*")) {
                insights.add(SessionAnalysisResponse.Insight.builder()
                    .category("PATTERN")
                    .content("Session followed consistent patterns")
                    .importance("MEDIUM")
                    .relatedTo("Agent behavior")
                    .build());
            }
        }

        if (request.getIncludeFailures() && lower.contains("error") && !lower.matches(".*failures:\\s*\\[\\].*")) {
            if (lower.matches(".*failures:\\s*\\[.*\\{.*")) {
                failures.add(SessionAnalysisResponse.FailureInsight.builder()
                    .errorType("API_ERROR")
                    .errorMessage("Error detected in session")
                    .context("During execution")
                    .resolution("Fixed by retry")
                    .lessonsLearned("Add retry logic")
                    .preventionHint("Implement circuit breaker")
                    .build());
            }
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
            .relatedNoteIds(note.getRelatedNoteIds())
            .title(note.getTitle())
            .errorPattern(note.getErrorPattern())
            .severity(note.getSeverity().toString())
            .accessCount(note.getAccessCount())
            .lastAccessedAt(note.getLastAccessedAt() != null ? note.getLastAccessedAt().toString() : null)
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

    // ==================== Methods from Confucius Spec ====================

    /**
     * Extract insights from a successful session.
     * KEY FEATURE from Confucius spec.
     *
     * @param sessionId the session ID
     * @param tenantId the tenant ID
     * @return list of insight notes
     */
    @Transactional
    public List<Note> extractInsights(String sessionId, String tenantId) {
        log.info("Extracting insights from session: {}", sessionId);

        // Get session audit logs
        Instant from = Instant.now().minus(7, ChronoUnit.DAYS);
        var auditLogs = auditLogRepo.findByTenantIdAndTimestampBetween(tenantId, from, Instant.now());

        // Use LLM to extract insights
        String prompt = buildInsightExtractionPrompt(auditLogs, sessionId);
        String llmResponse = openRouterService.chat(
            "Extract insights from this session. Return JSON: {insights: [{title, content, category, keywords}]}",
            prompt
        );

        return parseInsightsFromLLM(llmResponse, sessionId, tenantId);
    }

    /**
     * Extract hindsight notes from failures in a session.
     * KEY FEATURE from Confucius spec.
     *
     * @param sessionId the session ID
     * @param tenantId the tenant ID
     * @return list of hindsight notes (as generic Notes with HINDSIGHT type)
     */
    @Transactional
    public List<Note> extractHindsights(String sessionId, String tenantId) {
        log.info("Extracting hindsight notes from session: {}", sessionId);

        // Get session audit logs for errors
        Instant from = Instant.now().minus(7, ChronoUnit.DAYS);
        var auditLogs = auditLogRepo.findByTenantIdAndTimestampBetween(tenantId, from, Instant.now());

        // Filter for error events
        var errorLogs = auditLogs.stream()
            .filter(log -> log.getEventType() != null &&
                log.getEventType().toLowerCase().contains("error") ||
                (log.getOutcome() != null && log.getOutcome().toLowerCase().contains("error")))
            .toList();

        List<Note> hindsights = new ArrayList<>();
        for (var errorLog : errorLogs) {
            Note hindsight = Note.builder()
                .tenantId(tenantId)
                .sessionId(sessionId)
                .type(NoteType.HINDSIGHT)
                .title("Hindsight: " + errorLog.getEventType())
                .content(buildHindsightContent(errorLog))
                .category(NoteCategory.PROJECT_SPECIFIC)
                .severity(NoteSeverity.MEDIUM)
                .keywords(extractKeywordsFromError(errorLog))
                .createdAt(Instant.now())
                .autoGenerated(true)
                .build();

            hindsights.add(noteRepo.save(hindsight));
        }

        log.info("Extracted {} hindsight notes", hindsights.size());
        return hindsights;
    }

    /**
     * Identify patterns from a session.
     * KEY FEATURE from Confucius spec.
     *
     * @param sessionId the session ID
     * @param tenantId the tenant ID
     * @return list of pattern notes
     */
    @Transactional
    public List<Note> identifyPatterns(String sessionId, String tenantId) {
        log.info("Identifying patterns from session: {}", sessionId);

        // Get session data
        Instant from = Instant.now().minus(7, ChronoUnit.DAYS);
        var auditLogs = auditLogRepo.findByTenantIdAndTimestampBetween(tenantId, from, Instant.now());

        // Use LLM to identify patterns
        String prompt = buildPatternIdentificationPrompt(auditLogs, sessionId);
        String llmResponse = openRouterService.chat(
            "Identify recurring patterns in this session. Return JSON: {patterns: [{title, description, type, keywords}]}",
            prompt
        );

        return parsePatternsFromLLM(llmResponse, sessionId, tenantId);
    }

    /**
     * Extract architectural decisions from a session.
     * KEY FEATURE from Confucius spec.
     *
     * @param sessionId the session ID
     * @param tenantId the tenant ID
     * @return list of architecture notes
     */
    @Transactional
    public List<Note> extractArchitecturalDecisions(String sessionId, String tenantId) {
        log.info("Extracting architectural decisions from session: {}", sessionId);

        // Get session data
        Instant from = Instant.now().minus(7, ChronoUnit.DAYS);
        var auditLogs = auditLogRepo.findByTenantIdAndTimestampBetween(tenantId, from, Instant.now());

        // Use LLM to extract architectural decisions
        String prompt = buildArchitecturalDecisionPrompt(auditLogs, sessionId);
        String llmResponse = openRouterService.chat(
            "Extract architectural decisions from this session. Return JSON: {decisions: [{title, rationale, impact, alternatives}]}",
            prompt
        );

        return parseArchitecturalDecisionsFromLLM(llmResponse, sessionId, tenantId);
    }

    /**
     * Link a note to related memories in the graph.
     * Graph relationship: Note -[:DOCUMENTS]-> Memory
     * KEY FEATURE from Confucius spec.
     *
     * @param noteId the note ID
     * @param memoryIds the memory IDs to link
     */
    @Transactional
    public void linkToMemories(String noteId, List<String> memoryIds) {
        log.info("Linking note {} to {} memories", noteId, memoryIds.size());

        Note note = noteRepo.findById(noteId)
            .orElseThrow(() -> new IllegalArgumentException("Note not found: " + noteId));

        for (String memoryId : memoryIds) {
            note.addRelatedMemory(memoryId);
        }

        noteRepo.save(note);
        log.info("Successfully linked note to memories");
    }

    /**
     * Generate a structured session summary (like Confucius README.md).
     * KEY FEATURE from Confucius spec.
     *
     * @param sessionId the session ID
     * @param notes the notes generated from the session
     * @return the session summary note
     */
    @Transactional
    public Note generateSessionSummary(String sessionId, List<Note> notes) {
        log.info("Generating session summary for: {}", sessionId);

        StringBuilder summary = new StringBuilder();
        summary.append("# Session Summary\n\n");
        summary.append("**Session ID:** ").append(sessionId).append("\n");
        summary.append("**Generated:** ").append(Instant.now()).append("\n\n");

        // Group by type
        Map<NoteType, List<Note>> byType = notes.stream()
            .collect(Collectors.groupingBy(Note::getType));

        if (byType.containsKey(NoteType.INSIGHT)) {
            summary.append("## Insights Captured\n\n");
            for (Note note : byType.get(NoteType.INSIGHT)) {
                summary.append("- **").append(note.getTitle()).append("**\n");
                summary.append("  ").append(note.getPreview()).append("\n\n");
            }
        }

        if (byType.containsKey(NoteType.HINDSIGHT)) {
            summary.append("## Failures & Resolutions\n\n");
            for (Note note : byType.get(NoteType.HINDSIGHT)) {
                summary.append("- **").append(note.getTitle()).append("**\n");
                summary.append("  ").append(note.getPreview()).append("\n\n");
            }
        }

        if (byType.containsKey(NoteType.PATTERN)) {
            summary.append("## Patterns Identified\n\n");
            for (Note note : byType.get(NoteType.PATTERN)) {
                summary.append("- **").append(note.getTitle()).append("**\n");
                summary.append("  ").append(note.getPreview()).append("\n\n");
            }
        }

        if (byType.containsKey(NoteType.ARCHITECTURE)) {
            summary.append("## Architectural Decisions\n\n");
            for (Note note : byType.get(NoteType.ARCHITECTURE)) {
                summary.append("- **").append(note.getTitle()).append("**\n");
                summary.append("  ").append(note.getPreview()).append("\n\n");
            }
        }

        // Store summary as special note
        Note summaryNote = Note.builder()
            .tenantId(notes.isEmpty() ? TenantContext.getTenantId() : notes.get(0).getTenantId())
            .sessionId(sessionId)
            .type(NoteType.INSIGHT)
            .title("Session Summary: " + sessionId)
            .content(summary.toString())
            .category(NoteCategory.PROJECT_SPECIFIC)
            .createdAt(Instant.now())
            .autoGenerated(true)
            .build();

        return noteRepo.save(summaryNote);
    }

    // ==================== Private Helper Methods for Spec Implementation ====================

    private String buildInsightExtractionPrompt(List<com.integraltech.brainsentry.domain.AuditLog> auditLogs, String sessionId) {
        StringBuilder prompt = new StringBuilder();
        prompt.append("Extract key insights from this session:\n\n");
        for (var log : auditLogs) {
            prompt.append("- ").append(log.getEventType()).append(": ").append(log.getOutcome()).append("\n");
        }
        return prompt.toString();
    }

    private List<Note> parseInsightsFromLLM(String llmResponse, String sessionId, String tenantId) {
        List<Note> insights = new ArrayList<>();

        // Simplified parsing - in production use proper JSON deserialization
        if (llmResponse.toLowerCase().contains("insight") || llmResponse.toLowerCase().contains("pattern")) {
            Note insight = Note.builder()
                .tenantId(tenantId)
                .sessionId(sessionId)
                .type(NoteType.INSIGHT)
                .title("Session Insight")
                .content(llmResponse)
                .category(NoteCategory.PROJECT_SPECIFIC)
                .keywords(Arrays.asList("session", "insight"))
                .createdAt(Instant.now())
                .autoGenerated(true)
                .build();
            insights.add(noteRepo.save(insight));
        }

        return insights;
    }

    private String buildHindsightContent(com.integraltech.brainsentry.domain.AuditLog errorLog) {
        StringBuilder content = new StringBuilder();
        content.append("## Problem\n\n");
        content.append(errorLog.getEventType()).append("\n\n");
        if (errorLog.getOutcome() != null) {
            content.append("```\n").append(errorLog.getOutcome()).append("\n```\n\n");
        }
        content.append("## Context\n\n");
        content.append("Session ID: ").append(errorLog.getSessionId()).append("\n\n");
        content.append("## Resolution\n\n");
        content.append("See error logs for resolution steps.\n");
        return content.toString();
    }

    private List<String> extractKeywordsFromError(com.integraltech.brainsentry.domain.AuditLog errorLog) {
        List<String> keywords = new ArrayList<>();
        keywords.add("error");
        if (errorLog.getEventType() != null) {
            keywords.add(errorLog.getEventType().toLowerCase().replace(" ", "_"));
        }
        return keywords;
    }

    private String buildPatternIdentificationPrompt(List<com.integraltech.brainsentry.domain.AuditLog> auditLogs, String sessionId) {
        StringBuilder prompt = new StringBuilder();
        prompt.append("Identify recurring patterns in this session:\n\n");
        for (var log : auditLogs) {
            prompt.append("- ").append(log.getEventType()).append(": ").append(log.getOutcome()).append("\n");
        }
        return prompt.toString();
    }

    private List<Note> parsePatternsFromLLM(String llmResponse, String sessionId, String tenantId) {
        List<Note> patterns = new ArrayList<>();

        if (llmResponse.toLowerCase().contains("pattern")) {
            Note pattern = Note.builder()
                .tenantId(tenantId)
                .sessionId(sessionId)
                .type(NoteType.PATTERN)
                .title("Identified Pattern")
                .content(llmResponse)
                .category(NoteCategory.SHARED)
                .keywords(Arrays.asList("pattern", "reusable"))
                .createdAt(Instant.now())
                .autoGenerated(true)
                .build();
            patterns.add(noteRepo.save(pattern));
        }

        return patterns;
    }

    private String buildArchitecturalDecisionPrompt(List<com.integraltech.brainsentry.domain.AuditLog> auditLogs, String sessionId) {
        StringBuilder prompt = new StringBuilder();
        prompt.append("Extract architectural decisions from this session:\n\n");
        for (var log : auditLogs) {
            prompt.append("- ").append(log.getEventType()).append(": ").append(log.getOutcome()).append("\n");
        }
        return prompt.toString();
    }

    private List<Note> parseArchitecturalDecisionsFromLLM(String llmResponse, String sessionId, String tenantId) {
        List<Note> decisions = new ArrayList<>();

        if (llmResponse.toLowerCase().contains("decision") || llmResponse.toLowerCase().contains("architecture")) {
            Note decision = Note.builder()
                .tenantId(tenantId)
                .sessionId(sessionId)
                .type(NoteType.ARCHITECTURE)
                .title("Architectural Decision")
                .content(llmResponse)
                .category(NoteCategory.SHARED)
                .keywords(Arrays.asList("architecture", "decision"))
                .createdAt(Instant.now())
                .autoGenerated(true)
                .build();
            decisions.add(noteRepo.save(decision));
        }

        return decisions;
    }
}
