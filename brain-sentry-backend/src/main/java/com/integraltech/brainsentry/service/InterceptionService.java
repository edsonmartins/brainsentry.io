package com.integraltech.brainsentry.service;

import com.integraltech.brainsentry.domain.Memory;
import com.integraltech.brainsentry.dto.request.InterceptRequest;
import com.integraltech.brainsentry.dto.response.InterceptResponse;
import com.integraltech.brainsentry.repository.MemoryRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.List;
import java.util.regex.Pattern;

/**
 * Service for intercepting and enhancing prompts.
 *
 * This is the core functionality of Brain Sentry - analyzing
 * user prompts and injecting relevant memory context.
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class InterceptionService {

    private final OpenRouterService openRouterService;
    private final EmbeddingService embeddingService;
    private final MemoryRepository memoryRepository;
    private final AuditService auditService;

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

        if (memories.isEmpty()) {
            long latency = System.currentTimeMillis() - startTime;
            log.debug("No relevant memories found ({}ms)", latency);
            return InterceptResponse.builder()
                .enhanced(false)
                .originalPrompt(request.getPrompt())
                .enhancedPrompt(request.getPrompt())
                .contextInjected("")
                .memoriesUsed(List.of())
                .latencyMs((int) latency)
                .reasoning("No relevant memories found")
                .confidence(relevance.getConfidence())
                .tokensInjected(0)
                .llmCalls(1)
                .build();
        }

        // Step 4: Format and inject context
        String context = formatContext(memories);
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

        // Update injection counts
        memories.forEach(m -> {
            m.setInjectionCount(m.getInjectionCount() + 1);
            m.setLastAccessedAt(Instant.now());
            memoryRepository.save(m);
        });

        log.info("Enhanced prompt with {} memories ({}ms, {} tokens)",
            memories.size(), latency, tokens);

        // Audit log
        auditService.logInterception(request, memories, latency);

        return InterceptResponse.builder()
            .enhanced(true)
            .originalPrompt(request.getPrompt())
            .enhancedPrompt(enhancedPrompt)
            .contextInjected(context)
            .memoriesUsed(memoryRefs)
            .latencyMs((int) latency)
            .reasoning("Found " + memories.size() + " relevant memories")
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
     * Format memories into a context block.
     *
     * @param memories list of memories
     * @return formatted context string
     */
    private String formatContext(List<Memory> memories) {
        StringBuilder sb = new StringBuilder();
        sb.append("<system_context>\n");
        sb.append("The following relevant patterns and decisions were found:\n\n");

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

        sb.append("</system_context>\n");
        return sb.toString();
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
