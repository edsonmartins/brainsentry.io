package com.integraltech.brainsentry.service;

import com.integraltech.brainsentry.domain.HindsightNote;
import com.integraltech.brainsentry.domain.Memory;
import com.integraltech.brainsentry.dto.request.InterceptRequest;
import com.integraltech.brainsentry.dto.response.InterceptResponse;
import com.integraltech.brainsentry.repository.MemoryRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.regex.Pattern;

/**
 * Service for intercepting and enhancing prompts.
 *
 * This is the core functionality of Brain Sentry - analyzing
 * user prompts and injecting relevant memory context.
 *
 * Enhanced with NoteRetrievalService (Confucius spec integration).
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class InterceptionService {

    private final OpenRouterService openRouterService;
    private final EmbeddingService embeddingService;
    private final MemoryRepository memoryRepository;
    private final AuditService auditService;
    private final NoteRetrievalService noteRetrievalService;  // Confucius integration

    // Quick check patterns for fast-path filtering
    private static final List<Pattern> RELEVANCE_PATTERNS = List.of(
        Pattern.compile(Pattern.quote("agent"), Pattern.CASE_INSENSITIVE),
        Pattern.compile(Pattern.quote("service"), Pattern.CASE_INSENSITIVE),
        Pattern.compile(Pattern.quote("repository"), Pattern.CASE_INSENSITIVE),
        Pattern.compile(Pattern.quote("controller"), Pattern.CASE_INSENSITIVE),
        Pattern.compile(Pattern.quote("component"), Pattern.CASE_INSENSITIVE),
        Pattern.compile(Pattern.quote("class"), Pattern.CASE_INSENSITIVE),
        Pattern.compile(Pattern.quote("create"), Pattern.CASE_INSENSITIVE),
        Pattern.compile(Pattern.quote("implement"), Pattern.CASE_INSENSITIVE),
        Pattern.compile(Pattern.quote("add"), Pattern.CASE_INSENSITIVE),
        Pattern.compile(Pattern.quote("fix"), Pattern.CASE_INSENSITIVE),
        Pattern.compile(Pattern.quote("bug"), Pattern.CASE_INSENSITIVE),
        Pattern.compile(Pattern.quote("error"), Pattern.CASE_INSENSITIVE),
        Pattern.compile(Pattern.quote("pattern"), Pattern.CASE_INSENSITIVE),
        Pattern.compile(Pattern.quote("decision"), Pattern.CASE_INSENSITIVE),
        Pattern.compile(Pattern.quote("use"), Pattern.CASE_INSENSITIVE)
    );

    /**
     * Intercept and enhance a prompt with relevant memory context.
     *
     * @param request the intercept request
     * @return the enhanced response
     */
    public InterceptResponse interceptAndEnhance(InterceptRequest request) {
        long startTime = System.currentTimeMillis();

        log.debug("Intercepting prompt for session: {}", request.getSessionId());

        // Step 1: Quick check (fast path)
        if (!quickCheck(request.getPrompt()) && !Boolean.TRUE.equals(request.getForceDeepAnalysis())) {
            long latency = System.currentTimeMillis() - startTime;
            log.debug("Quick check failed - passing through ({}ms)", latency);
            return InterceptResponse.builder()
                .enhanced(false)
                .originalPrompt(request.getPrompt())
                .enhancedPrompt(request.getPrompt())
                .contextInjected("")
                .memoriesUsed(List.of())
                .notesUsed(List.of())
                .latencyMs((int) latency)
                .reasoning("Quick check: No relevant keywords detected")
                .confidence(0.0)
                .tokensInjected(0)
                .llmCalls(0)
                .build();
        }

        // Step 2: Deep analysis (LLM)
        var relevance = openRouterService.analyzeRelevance(
            request.getPrompt(),
            request.getContext()
        );

        if (!relevance.isNeedsContext()) {
            long latency = System.currentTimeMillis() - startTime;
            log.debug("Relevance analysis: no context needed ({}ms)", latency);
            return InterceptResponse.builder()
                .enhanced(false)
                .originalPrompt(request.getPrompt())
                .enhancedPrompt(request.getPrompt())
                .contextInjected("")
                .memoriesUsed(List.of())
                .notesUsed(List.of())
                .latencyMs((int) latency)
                .reasoning(relevance.getReasoning())
                .confidence(relevance.getConfidence())
                .tokensInjected(0)
                .llmCalls(1)
                .build();
        }

        // Step 3: Search relevant memories
        String tenantId = request.getTenantId() != null ? request.getTenantId() : "default";
        float[] embedding = embeddingService.embed(request.getPrompt());
        List<Memory> memories = memoryRepository.vectorSearch(
            embedding,
            5,  // top 5 memories
            tenantId
        );

        // Filter by importance (only CRITICAL and IMPORTANT)
        memories = memories.stream()
            .filter(m -> m.getImportance().name().equals("CRITICAL") ||
                         m.getImportance().name().equals("IMPORTANT"))
            .limit(3)
            .toList();

        // Step 3.5: Search relevant notes (Confucius spec integration)
        List<HindsightNote> hindsightNotes = new ArrayList<>();

        // If prompt contains error keywords, search for matching hindsight notes
        if (containsErrorKeywords(request.getPrompt())) {
            hindsightNotes = noteRetrievalService.searchHindsightNotes(
                request.getPrompt(),
                extractErrorType(request.getPrompt()),
                tenantId
            );
            // Record access for retrieved notes
            hindsightNotes.forEach(note -> noteRetrievalService.recordNoteAccess(note.getId()));
        }

        // Get relevant notes by query (fallback if no memories or no error notes)
        if (memories.isEmpty() || hindsightNotes.isEmpty()) {
            // Fall back to semantic note search
            List<HindsightNote> fallbackNotes = noteRetrievalService.getRelevantNotes(
                request.getPrompt(),
                tenantId,
                3
            );
            // Add unique notes to hindsightNotes
            for (HindsightNote note : fallbackNotes) {
                if (!hindsightNotes.contains(note)) {
                    hindsightNotes.add(note);
                }
            }
        }

        // Check if we have any relevant context (memories OR notes)
        boolean hasMemories = !memories.isEmpty();
        boolean hasNotes = !hindsightNotes.isEmpty();

        if (!hasMemories && !hasNotes) {
            long latency = System.currentTimeMillis() - startTime;
            log.debug("No relevant memories or notes found ({}ms)", latency);
            return InterceptResponse.builder()
                .enhanced(false)
                .originalPrompt(request.getPrompt())
                .enhancedPrompt(request.getPrompt())
                .contextInjected("")
                .memoriesUsed(List.of())
                .notesUsed(List.of())
                .latencyMs((int) latency)
                .reasoning("No relevant memories or notes found")
                .confidence(relevance.getConfidence())
                .tokensInjected(0)
                .llmCalls(1)
                .build();
        }

        // Step 4: Format and inject context (with notes support)
        String context = formatContextWithNotes(memories, hindsightNotes);
        String enhancedPrompt = injectContext(request.getPrompt(), context);

        long latency = System.currentTimeMillis() - startTime;
        int tokens = estimateTokens(context);

        // Build memory references
        List<InterceptResponse.MemoryReference> memoryRefs = memories.stream()
            .map(m -> InterceptResponse.MemoryReference.builder()
                .id(m.getId())
                .summary(m.getSummary())
                .category(m.getCategory().name())
                .importance(m.getImportance().name())
                .relevanceScore(m.getRelevanceScore())
                .excerpt(m.getContent().length() > 100 ?
                    m.getContent().substring(0, 100) + "..." : m.getContent())
                .build())
            .toList();

        // Build note references (Confucius integration)
        List<InterceptResponse.NoteReference> noteRefs = new ArrayList<>();
        for (HindsightNote note : hindsightNotes) {
            noteRefs.add(InterceptResponse.NoteReference.builder()
                .id(note.getId())
                .title(note.getTitle() != null ? note.getTitle() : note.getErrorType())
                .type("HINDSIGHT")
                .severity(note.getSeverity().name())
                .excerpt(note.getResolution() != null && note.getResolution().length() > 100 ?
                    note.getResolution().substring(0, 100) + "..." : note.getResolution())
                .build());
        }

        // Update injection counts for memories
        memories.forEach(m -> {
            m.setInjectionCount(m.getInjectionCount() + 1);
            m.setLastAccessedAt(Instant.now());
            memoryRepository.save(m);
        });

        log.info("Enhanced prompt with {} memories and {} notes ({}ms, {} tokens)",
            memories.size(), hindsightNotes.size(), latency, tokens);

        // Audit log
        auditService.logInterception(request, memories, latency);

        return InterceptResponse.builder()
            .enhanced(true)
            .originalPrompt(request.getPrompt())
            .enhancedPrompt(enhancedPrompt)
            .contextInjected(context)
            .memoriesUsed(memoryRefs)
            .notesUsed(noteRefs)
            .latencyMs((int) latency)
            .reasoning("Found " + memories.size() + " memories and " +
                hindsightNotes.size() + " notes")
            .confidence(relevance.getConfidence())
            .tokensInjected(tokens)
            .llmCalls(1)
            .build();
    }

    /**
     * Quick regex-based check for potential relevance.
     *
     * @param prompt the prompt to check
     * @return true if prompt might be relevant
     */
    private boolean quickCheck(String prompt) {
        if (prompt == null || prompt.length() < 10) {
            return false;
        }

        return RELEVANCE_PATTERNS.stream()
            .anyMatch(pattern -> pattern.matcher(prompt).find());
    }

    /**
     * Format memories and notes into a context block.
     * Confucius spec integration: includes notes alongside memories.
     *
     * @param memories list of memories
     * @param hindsightNotes list of hindsight notes
     * @return formatted context string
     */
    private String formatContextWithNotes(List<Memory> memories,
                                          List<HindsightNote> hindsightNotes) {
        StringBuilder sb = new StringBuilder();
        sb.append("<system_context>\n");
        sb.append("The following relevant patterns and decisions were found:\n\n");

        // Format memories
        for (int i = 0; i < memories.size(); i++) {
            Memory m = memories.get(i);
            sb.append(String.format("[%d] %s - %s\n",
                i + 1,
                m.getImportance().getDisplayName(),
                m.getCategory().getDisplayName()));
            sb.append(String.format("    %s\n", m.getSummary()));
            if (m.getCodeExample() != null && !m.getCodeExample().isEmpty()) {
                sb.append("    Code:\n");
                sb.append("    ```").append(m.getProgrammingLanguage() != null ?
                    m.getProgrammingLanguage() : "java").append("\n");
                sb.append("    ").append(m.getCodeExample().replace("\n", "\n    "));
                sb.append("\n    ```\n");
            }
            sb.append("\n");
        }

        // Format hindsight notes (Confucius feature)
        if (!hindsightNotes.isEmpty()) {
            sb.append("## Past Learnings (Hindsight Notes)\n\n");
            for (HindsightNote note : hindsightNotes) {
                sb.append(String.format("- [%s] %s: %s\n",
                    note.getSeverity().name(),
                    note.getTitle() != null ? note.getTitle() : note.getErrorType(),
                    note.getResolution() != null ? note.getResolution() : "No resolution available"));
                if (note.getLessonsLearned() != null) {
                    sb.append(String.format("  Lesson: %s\n",
                        note.getLessonsLearned().length() > 100 ?
                            note.getLessonsLearned().substring(0, 100) + "..." : note.getLessonsLearned()));
                }
                sb.append("\n");
            }
        }

        sb.append("</system_context>\n");
        return sb.toString();
    }

    /**
     * Check if prompt contains error-related keywords.
     * Used to trigger hindsight note retrieval.
     *
     * @param prompt the prompt to check
     * @return true if error keywords detected
     */
    private boolean containsErrorKeywords(String prompt) {
        if (prompt == null) {
            return false;
        }
        String lower = prompt.toLowerCase();
        return lower.contains("error") ||
               lower.contains("exception") ||
               lower.contains("failed") ||
               lower.contains("failure") ||
               lower.contains("bug") ||
               lower.contains("issue") ||
               lower.contains("nullpointer") ||
               lower.contains("runtime") ||
               lower.contains("timeout");
    }

    /**
     * Extract error type from prompt.
     * Simplified extraction - in production use NLP.
     *
     * @param prompt the prompt to analyze
     * @return extracted error type or "UNKNOWN"
     */
    private String extractErrorType(String prompt) {
        if (prompt == null) {
            return "UNKNOWN";
        }
        String lower = prompt.toLowerCase();
        if (lower.contains("nullpointer") || lower.contains("null pointer")) {
            return "NullPointerException";
        } else if (lower.contains("timeout")) {
            return "TimeoutException";
        } else if (lower.contains("sql") || lower.contains("database")) {
            return "SQLException";
        } else if (lower.contains("io")) {
            return "IOException";
        } else if (lower.contains("runtime")) {
            return "RuntimeException";
        }
        return "UNKNOWN";
    }

    /**
     * Inject context into the prompt.
     *
     * @param originalPrompt the original prompt
     * @param context the context to inject
     * @return enhanced prompt
     */
    private String injectContext(String originalPrompt, String context) {
        return context + "\n\n" + originalPrompt;
    }

    /**
     * Estimate token count (rough approximation).
     *
     * @param text the text to measure
     * @return estimated token count
     */
    private int estimateTokens(String text) {
        // Rough estimate: ~4 characters per token
        return (int) Math.ceil(text.length() / 4.0);
    }
}
