package com.integraltech.brainsentry.service;

import com.integraltech.brainsentry.config.TenantContext;
import com.integraltech.brainsentry.domain.ContextSummary;
import com.integraltech.brainsentry.dto.request.CompressionRequest;
import com.integraltech.brainsentry.dto.response.CompressedContextResponse;
import com.integraltech.brainsentry.repository.ContextSummaryJpaRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

/**
 * Service for intelligent context compression.
 *
 * Inspired by Confucius Code Agent's Architect agent.
 * This service compresses conversation history when it becomes too long,
 * preserving critical information while reducing token usage.
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class ArchitectService {

    private final OpenRouterService openRouterService;
    private final ContextSummaryJpaRepository contextSummaryRepo;

    /**
     * Default token threshold for compression.
     */
    private static final int DEFAULT_TOKEN_THRESHOLD = 100000;

    /**
     * Default recent window size (messages to keep in full).
     */
    private static final int DEFAULT_RECENT_WINDOW = 10;

    /**
     * Compress context history when it exceeds the threshold.
     *
     * @param messages the conversation history
     * @param threshold the token threshold for compression
     * @return compressed context
     */
    @Transactional
    public CompressedContextResponse compressContext(List<CompressionRequest.Message> messages, Integer threshold) {
        return compressContext(messages, threshold, null, null);
    }

    /**
     * Compress context history when it exceeds the threshold, with session tracking.
     * KEY FEATURE from Confucius spec - saves ContextSummary for audit trail.
     *
     * @param messages the conversation history
     * @param threshold the token threshold for compression
     * @param sessionId the session ID for tracking
     * @param tenantId the tenant ID
     * @return compressed context
     */
    @Transactional
    public CompressedContextResponse compressContext(
            List<CompressionRequest.Message> messages,
            Integer threshold,
            String sessionId,
            String tenantId) {

        if (threshold == null) {
            threshold = DEFAULT_TOKEN_THRESHOLD;
        }

        // Estimate token count (rough estimation: 4 chars ~ 1 token)
        int estimatedTokens = estimateTokens(messages);

        if (estimatedTokens < threshold) {
            log.info("Context size ({}) below threshold ({}), no compression needed",
                estimatedTokens, threshold);
            return noCompressionResponse(messages, estimatedTokens);
        }

        log.info("Compressing context: {} tokens -> target {}", estimatedTokens, threshold);

        // Compress using LLM
        CompressedContextResponse response = compressWithLLM(messages, threshold);

        // Save ContextSummary for audit trail (Confucius spec requirement)
        if (sessionId != null && tenantId != null && response.getCompressed()) {
            saveContextSummary(sessionId, tenantId, messages, response);
        }

        return response;
    }

    /**
     * Extract structured summary from messages.
     *
     * @param messages the messages to summarize
     * @return structured summary
     */
    public CompressedContextResponse.StructuredSummary extractSummary(
            List<CompressionRequest.Message> messages) {

        if (messages.isEmpty()) {
            return CompressedContextResponse.StructuredSummary.builder().build();
        }

        String prompt = buildSummaryPrompt(messages);

        String llmResponse = openRouterService.chat(prompt,
            "Return JSON only with structure: {taskGoal, keyDecisions[], openTodos[], criticalErrors[], importantFileChanges[], additionalContext}");

        return parseSummaryResponse(llmResponse);
    }

    /**
     * Identify critical messages that must be preserved.
     *
     * @param messages all messages
     * @param preserveKeywords keywords that indicate importance
     * @return list of critical messages
     */
    public List<CompressionRequest.Message> identifyCriticalMessages(
            List<CompressionRequest.Message> messages,
            List<String> preserveKeywords) {

        return messages.stream()
            .filter(msg -> isCritical(msg, preserveKeywords))
            .collect(Collectors.toList());
    }

    /**
     * Check if compression is needed based on token count.
     *
     * @param messages the messages to check
     * @param threshold the token threshold
     * @return true if compression is recommended
     */
    public boolean shouldCompress(List<CompressionRequest.Message> messages, Integer threshold) {
        if (threshold == null) {
            threshold = DEFAULT_TOKEN_THRESHOLD;
        }
        return estimateTokens(messages) >= threshold;
    }

    // ==================== Private Helper Methods ====================

    /**
     * Estimate token count for messages.
     * Rough estimation: 4 characters ~ 1 token.
     */
    private int estimateTokens(List<CompressionRequest.Message> messages) {
        int totalChars = messages.stream()
            .mapToInt(msg -> msg.getContent() != null ? msg.getContent().length() : 0)
            .sum();
        return totalChars / 4;  // Rough estimation
    }

    /**
     * Create response when no compression was needed.
     */
    private CompressedContextResponse noCompressionResponse(
            List<CompressionRequest.Message> messages,
            int tokenCount) {

        return CompressedContextResponse.builder()
            .originalMessageCount(messages.size())
            .compressedMessageCount(messages.size())
            .originalTokenCount(tokenCount)
            .compressedTokenCount(tokenCount)
            .compressionRatio(1.0)
            .compressed(false)
            .preservedMessages(toResponseMessages(messages))
            .build();
    }

    /**
     * Compress messages using LLM.
     */
    private CompressedContextResponse compressWithLLM(
            List<CompressionRequest.Message> messages,
            int threshold) {

        // Build compression prompt
        String prompt = buildCompressionPrompt(messages, threshold);

        // Call LLM
        String llmResponse = openRouterService.chat(prompt,
            "Return JSON with structure: {summary: {taskGoal, keyDecisions[], openTodos[], criticalErrors[], importantFileChanges[], additionalContext}, preservedMessages[]}");

        return parseCompressionResponse(llmResponse, messages);
    }

    /**
     * Build prompt for context compression.
     */
    private String buildCompressionPrompt(List<CompressionRequest.Message> messages, int threshold) {
        StringBuilder prompt = new StringBuilder();
        prompt.append("Compress this conversation history to fit within ");
        prompt.append(threshold);
        prompt.append(" tokens.\n\n");
        prompt.append("CRITICAL - PRESERVE:\n");
        prompt.append("- Task goals and requirements\n");
        prompt.append("- Key decisions made (with rationale)\n");
        prompt.append("- Critical errors encountered\n");
        prompt.append("- Open TODOs and pending items\n");
        prompt.append("- Important file changes (file names + nature of change)\n\n");
        prompt.append("OMIT:\n");
        prompt.append("- Redundant tool outputs\n");
        prompt.append("- Verbose logs\n");
        prompt.append("- Intermediate failed attempts (keep final solution only)\n\n");
        prompt.append("Return JSON preserving the last ");
        prompt.append(Math.min(10, messages.size() / 3));
        prompt.append(" recent messages separately.\n\n");

        prompt.append("Conversation:\n");
        for (var msg : messages) {
            prompt.append(String.format("[%s]: %s\n",
                msg.getRole(), truncate(msg.getContent(), 500)));
        }

        return prompt.toString();
    }

    /**
     * Build prompt for structured summary.
     */
    private String buildSummaryPrompt(List<CompressionRequest.Message> messages) {
        StringBuilder prompt = new StringBuilder();
        prompt.append("Analyze this conversation and create a structured summary:\n\n");

        prompt.append("Conversation:\n");
        for (var msg : messages) {
            prompt.append(String.format("[%s]: %s\n",
                msg.getRole(), truncate(msg.getContent(), 500)));
        }

        prompt.append("\nExtract:\n");
        prompt.append("1. taskGoal - What is being attempted\n");
        prompt.append("2. keyDecisions - Important decisions made\n");
        prompt.append("3. openTodos - Pending tasks\n");
        prompt.append("4. criticalErrors - Errors that occurred\n");
        prompt.append("5. importantFileChanges - Files that were changed\n");
        prompt.append("6. additionalContext - Any other important context\n");

        return prompt.toString();
    }

    /**
     * Parse LLM compression response.
     */
    private CompressedContextResponse parseCompressionResponse(
            String llmResponse,
            List<CompressionRequest.Message> originalMessages) {

        // Simplified parsing - in production use proper JSON deserialization
        List<CompressedContextResponse.Message> preserved = new ArrayList<>();

        // Extract preserved messages from response
        // (In production, parse JSON properly)
        for (int i = Math.max(0, originalMessages.size() - 10); i < originalMessages.size(); i++) {
            CompressionRequest.Message orig = originalMessages.get(i);
            preserved.add(CompressedContextResponse.Message.builder()
                .role(orig.getRole())
                .content(truncate(orig.getContent(), 1000))
                .timestamp(orig.getTimestamp())
                .isSummary(false)
                .build());
        }

        int originalTokens = estimateTokens(originalMessages);
        int compressedTokens = estimateTokens(originalMessages) / 2;  // Estimate

        return CompressedContextResponse.builder()
            .originalMessageCount(originalMessages.size())
            .compressedMessageCount(preserved.size())
            .originalTokenCount(originalTokens)
            .compressedTokenCount(compressedTokens)
            .compressionRatio((double) compressedTokens / originalTokens)
            .compressed(true)
            .preservedMessages(preserved)
            .summary(extractSummaryFromResponse(llmResponse))
            .build();
    }

    /**
     * Parse summary from LLM response.
     */
    private CompressedContextResponse.StructuredSummary extractSummaryFromResponse(String response) {
        // Simplified parsing - extract sections from text
        return CompressedContextResponse.StructuredSummary.builder()
            .taskGoal(extractSection(response, "task", "Complete the task"))
            .keyDecisions(extractList(response, "decision"))
            .openTodos(extractList(response, "todo"))
            .criticalErrors(extractErrorList(response))
            .importantFileChanges(extractList(response, "file"))
            .additionalContext(extractSection(response, "context", ""))
            .build();
    }

    /**
     * Parse summary response.
     */
    private CompressedContextResponse.StructuredSummary parseSummaryResponse(String llmResponse) {
        return extractSummaryFromResponse(llmResponse);
    }

    private String truncate(String str, int maxLength) {
        if (str == null) return null;
        return str.length() > maxLength ? str.substring(0, maxLength) + "..." : str;
    }

    private boolean isCritical(CompressionRequest.Message msg, List<String> keywords) {
        if (keywords == null || keywords.isEmpty()) {
            return "error".equalsIgnoreCase(msg.getRole()) ||
                   "system".equals(msg.getRole());
        }

        String content = msg.getContent() != null ? msg.getContent().toLowerCase() : "";
        return keywords.stream().anyMatch(keyword ->
            content.toLowerCase().contains(keyword.toLowerCase()));
    }

    private List<CompressedContextResponse.Message> toResponseMessages(
            List<CompressionRequest.Message> messages) {

        return messages.stream()
            .map(msg -> CompressedContextResponse.Message.builder()
                .role(msg.getRole())
                .content(msg.getContent())
                .timestamp(msg.getTimestamp())
                .isSummary(false)
                .build())
            .collect(Collectors.toList());
    }

    private String extractSection(String text, String key, String defaultValue) {
        String lower = text.toLowerCase();
        if (!lower.contains(key)) {
            return defaultValue;
        }

        int start = lower.indexOf(key);
        int end = lower.indexOf("\n", start);
        if (end == -1) end = lower.length();

        return text.substring(start, Math.min(start + 200, end)).trim();
    }

    private List<String> extractList(String text, String key) {
        List<String> items = new ArrayList<>();
        String[] lines = text.split("\n");

        for (String line : lines) {
            if (line.toLowerCase().contains(key)) {
                items.add(line.trim());
            }
        }

        return items;
    }

    private List<CompressedContextResponse.CriticalError> extractErrorList(String text) {
        List<CompressedContextResponse.CriticalError> errors = new ArrayList<>();
        String[] lines = text.split("\n");

        for (String line : lines) {
            if (line.toLowerCase().contains("error")) {
                errors.add(CompressedContextResponse.CriticalError.builder()
                    .errorType("ERROR")
                    .description(line.trim())
                    .resolution("See context")
                    .build());
            }
        }

        return errors;
    }

    /**
     * Save ContextSummary for audit trail.
     * KEY FEATURE from Confucius spec - enables analysis of compression effectiveness.
     *
     * @param sessionId the session ID
     * @param tenantId the tenant ID
     * @param originalMessages the original messages before compression
     * @param response the compression response
     */
    private void saveContextSummary(
            String sessionId,
            String tenantId,
            List<CompressionRequest.Message> originalMessages,
            CompressedContextResponse response) {

        try {
            CompressedContextResponse.StructuredSummary summary = response.getSummary();
            if (summary == null) {
                summary = CompressedContextResponse.StructuredSummary.builder().build();
            }

            ContextSummary contextSummary = ContextSummary.builder()
                .tenantId(tenantId)
                .sessionId(sessionId)
                .originalTokenCount(response.getOriginalTokenCount())
                .compressedTokenCount(response.getCompressedTokenCount())
                .compressionRatio(response.getCompressionRatio())
                .summary(formatSummaryAsMarkdown(summary))
                .goals(summary.getTaskGoal() != null ? List.of(summary.getTaskGoal()) : List.of())
                .decisions(summary.getKeyDecisions() != null ? summary.getKeyDecisions() : List.of())
                .errors(extractErrorMessages(summary.getCriticalErrors()))
                .todos(summary.getOpenTodos() != null ? summary.getOpenTodos() : List.of())
                .recentWindowSize(DEFAULT_RECENT_WINDOW)
                .modelUsed("llm-compression")
                .compressionMethod("LLM")
                .createdAt(Instant.now())
                .build();

            contextSummaryRepo.save(contextSummary);
            log.info("Saved ContextSummary for session: {}, ratio: {:.2f}",
                sessionId, response.getCompressionRatio());
        } catch (Exception e) {
            log.warn("Failed to save ContextSummary: {}", e.getMessage());
            // Don't fail compression if summary save fails
        }
    }

    /**
     * Format structured summary as Markdown.
     */
    private String formatSummaryAsMarkdown(CompressedContextResponse.StructuredSummary summary) {
        StringBuilder md = new StringBuilder();

        if (summary.getTaskGoal() != null) {
            md.append("## Goal\n\n").append(summary.getTaskGoal()).append("\n\n");
        }

        if (summary.getKeyDecisions() != null && !summary.getKeyDecisions().isEmpty()) {
            md.append("## Decisions\n\n");
            for (String decision : summary.getKeyDecisions()) {
                md.append("- ").append(decision).append("\n");
            }
            md.append("\n");
        }

        if (summary.getCriticalErrors() != null && !summary.getCriticalErrors().isEmpty()) {
            md.append("## Errors\n\n");
            for (CompressedContextResponse.CriticalError error : summary.getCriticalErrors()) {
                md.append("- **").append(error.getErrorType())
                  .append("**: ").append(error.getDescription()).append("\n");
            }
            md.append("\n");
        }

        if (summary.getOpenTodos() != null && !summary.getOpenTodos().isEmpty()) {
            md.append("## TODOs\n\n");
            for (String todo : summary.getOpenTodos()) {
                md.append("- ").append(todo).append("\n");
            }
        }

        return md.toString();
    }

    /**
     * Extract error messages from CriticalError list.
     */
    private List<String> extractErrorMessages(List<CompressedContextResponse.CriticalError> errors) {
        if (errors == null || errors.isEmpty()) {
            return List.of();
        }
        return errors.stream()
            .map(e -> e.getErrorType() + ": " + e.getDescription())
            .collect(Collectors.toList());
    }
}
