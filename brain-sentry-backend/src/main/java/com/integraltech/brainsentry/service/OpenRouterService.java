package com.integraltech.brainsentry.service;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.integraltech.brainsentry.config.OpenRouterConfig;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.*;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

/**
 * Service for interacting with OpenRouter API (Grok model).
 *
 * Provides methods for relevance analysis and importance scoring
 * using the x-ai/grok-4.1-fast model via OpenRouter.
 */
@Service
public class OpenRouterService {

    private static final Logger log = LoggerFactory.getLogger(OpenRouterService.class);

    private final OpenRouterConfig config;
    private final RestTemplate restTemplate;
    private final ObjectMapper objectMapper;

    public OpenRouterService(OpenRouterConfig config, RestTemplate restTemplate, ObjectMapper objectMapper) {
        this.config = config;
        this.restTemplate = restTemplate;
        this.objectMapper = objectMapper;
    }

    /**
     * Analyze relevance of a prompt to determine if context should be injected.
     *
     * @param prompt the user's prompt
     * @param context additional context (project, file, etc.)
     * @return RelevanceAnalysis with decision and confidence
     */
    public RelevanceAnalysis analyzeRelevance(String prompt, Map<String, Object> context) {
        String analysisPrompt = buildRelevancePrompt(prompt, context);

        try {
            String response = callGrok(analysisPrompt, 300);
            return parseRelevanceAnalysis(response);
        } catch (Exception e) {
            log.error("Error analyzing relevance", e);
            return new RelevanceAnalysis(false, "Error during analysis", 0.0);
        }
    }

    /**
     * Analyze importance of content to determine how it should be stored.
     *
     * @param content the content to analyze
     * @return ImportanceAnalysis with category and importance level
     */
    public ImportanceAnalysis analyzeImportance(String content) {
        String prompt = buildImportancePrompt(content);

        try {
            String response = callGrok(prompt, 500);
            return parseImportanceAnalysis(response);
        } catch (Exception e) {
            log.error("Error analyzing importance", e);
            return ImportanceAnalysis.defaultResult();
        }
    }

    /**
     * Generic chat method for sending prompts to the LLM.
     *
     * @param systemPrompt the system prompt
     * @param userPrompt the user prompt
     * @return the LLM response
     */
    public String chat(String systemPrompt, String userPrompt) {
        OpenRouterRequest request = new OpenRouterRequest(
            config.getModel(),
            List.of(
                new Message("system", systemPrompt),
                new Message("user", userPrompt)
            ),
            config.getTemperature(),
            4000
        );

        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);
        headers.setBearerAuth(config.getApiKey());
        headers.set("HTTP-Referer", "https://brainsentry.io");
        headers.set("X-Title", "Brain Sentry");

        HttpEntity<OpenRouterRequest> entity = new HttpEntity<>(request, headers);

        try {
            ResponseEntity<String> response = restTemplate.exchange(
                config.getBaseUrl(),
                HttpMethod.POST,
                entity,
                String.class
            );

            if (response.getStatusCode().is2xxSuccessful() && response.getBody() != null) {
                JsonNode root = objectMapper.readTree(response.getBody());
                return root.path("choices").get(0).path("message").path("content").asText();
            } else {
                log.warn("OpenRouter returned status: {}", response.getStatusCode());
                return "";
            }
        } catch (Exception e) {
            log.error("Error calling OpenRouter API", e);
            return "";
        }
    }

    /**
     * Extract key patterns from content for relationship detection.
     *
     * @param content the content to analyze
     * @return list of detected patterns
     */
    public List<String> extractPatterns(String content) {
        String prompt = String.format("""
            Extract the key technical patterns, technologies, and concepts from this content.
            Return as a JSON array of strings.

            Content: %s

            Response format: ["pattern1", "pattern2", ...]
            """, content);

        try {
            String response = callGrok(prompt, 200);
            return parsePatternList(response);
        } catch (Exception e) {
            log.error("Error extracting patterns", e);
            return List.of();
        }
    }

    // ==================== Private Methods ====================

    private String callGrok(String prompt, int maxTokens) {
        OpenRouterRequest request = new OpenRouterRequest(
            config.getModel(),
            List.of(
                new Message("system", "You are a technical analysis assistant for developers. Respond only with valid JSON."),
                new Message("user", prompt)
            ),
            config.getTemperature(),
            maxTokens
        );

        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);
        headers.setBearerAuth(config.getApiKey());
        headers.set("HTTP-Referer", "https://brainsentry.io");
        headers.set("X-Title", "Brain Sentry");

        HttpEntity<OpenRouterRequest> entity = new HttpEntity<>(request, headers);

        try {
            ResponseEntity<String> response = restTemplate.exchange(
                config.getBaseUrl(),
                HttpMethod.POST,
                entity,
                String.class
            );

            if (response.getStatusCode().is2xxSuccessful() && response.getBody() != null) {
                JsonNode root = objectMapper.readTree(response.getBody());
                return root.path("choices").get(0).path("message").path("content").asText();
            } else {
                log.warn("OpenRouter returned status: {}", response.getStatusCode());
                return "{}";
            }
        } catch (Exception e) {
            log.error("Error calling OpenRouter API", e);
            return "{}";
        }
    }

    private String buildRelevancePrompt(String prompt, Map<String, Object> context) {
        StringBuilder contextStr = new StringBuilder();
        if (context != null) {
            context.forEach((k, v) -> contextStr.append(k).append(": ").append(v).append("\n"));
        }

        return String.format("""
            Analyze if this developer prompt would benefit from additional context injection.

            Prompt: %s

            Context:
            %s

            Respond in JSON format:
            {
              "needsContext": true/false,
              "reasoning": "brief explanation",
              "confidence": 0.0-1.0,
              "categories": ["PATTERN", "DECISION", ...]
            }
            """, prompt, contextStr);
    }

    private String buildImportancePrompt(String content) {
        return String.format("""
            Analyze this technical content and classify it.

            Content: %s

            Respond in JSON format:
            {
              "shouldRemember": true/false,
              "importance": "CRITICAL|IMPORTANT|MINOR",
              "category": "DECISION|PATTERN|ANTIPATTERN|DOMAIN|BUG|OPTIMIZATION|INTEGRATION",
              "summary": "brief 1-sentence summary",
              "reasoning": "brief explanation"
            }
            """, content);
    }

    private RelevanceAnalysis parseRelevanceAnalysis(String response) {
        try {
            // Extract JSON from response (handle markdown code blocks)
            String json = extractJson(response);
            JsonNode root = objectMapper.readTree(json);

            return new RelevanceAnalysis(
                root.path("needsContext").asBoolean(false),
                root.path("reasoning").asText(""),
                root.path("confidence").asDouble(0.0)
            );
        } catch (Exception e) {
            log.warn("Failed to parse relevance analysis: {}", response, e);
            return new RelevanceAnalysis(false, "Parse error", 0.0);
        }
    }

    private ImportanceAnalysis parseImportanceAnalysis(String response) {
        try {
            String json = extractJson(response);
            JsonNode root = objectMapper.readTree(json);

            return new ImportanceAnalysis(
                root.path("shouldRemember").asBoolean(true),
                root.path("importance").asText("MINOR"),
                root.path("category").asText("PATTERN"),
                root.path("summary").asText(""),
                root.path("reasoning").asText("")
            );
        } catch (Exception e) {
            log.warn("Failed to parse importance analysis: {}", response, e);
            return ImportanceAnalysis.defaultResult();
        }
    }

    private List<String> parsePatternList(String response) {
        try {
            String json = extractJson(response);
            JsonNode root = objectMapper.readTree(json);

            List<String> patterns = new ArrayList<>();
            if (root.isArray()) {
                for (JsonNode node : root) {
                    patterns.add(node.asText());
                }
            }
            return patterns;
        } catch (Exception e) {
            log.warn("Failed to parse pattern list: {}", response, e);
            return List.of();
        }
    }

    private String extractJson(String response) {
        // Remove markdown code blocks if present
        response = response.trim();
        if (response.startsWith("```")) {
            int start = response.indexOf('\n');
            int end = response.lastIndexOf("```");
            if (start > 0 && end > start) {
                return response.substring(start + 1, end).trim();
            }
        }
        return response;
    }

    /**
     * Check if the OpenRouter service is configured.
     *
     * @return true if API key is set
     */
    public boolean isConfigured() {
        return config.getApiKey() != null && !config.getApiKey().isBlank();
    }

    /**
     * Get the model being used.
     *
     * @return the model identifier
     */
    public String getModel() {
        return config.getModel();
    }

    // ==================== Inner Classes ====================

    public static class RelevanceAnalysis {
        private boolean needsContext;
        private String reasoning;
        private Double confidence;

        public RelevanceAnalysis() {}

        public RelevanceAnalysis(boolean needsContext, String reasoning, Double confidence) {
            this.needsContext = needsContext;
            this.reasoning = reasoning;
            this.confidence = confidence;
        }

        public static RelevanceAnalysis needsContext(boolean needsContext, Double confidence, String reasoning) {
            return new RelevanceAnalysis(needsContext, reasoning, confidence);
        }

        public boolean isNeedsContext() { return needsContext; }
        public void setNeedsContext(boolean needsContext) { this.needsContext = needsContext; }

        public String getReasoning() { return reasoning; }
        public void setReasoning(String reasoning) { this.reasoning = reasoning; }

        public Double getConfidence() { return confidence; }
        public void setConfidence(Double confidence) { this.confidence = confidence; }
    }

    public static class ImportanceAnalysis {
        private boolean shouldRemember;
        private String importance;  // CRITICAL, IMPORTANT, MINOR
        private String category;    // DECISION, PATTERN, etc.
        private String summary;
        private String reasoning;

        public ImportanceAnalysis() {}

        public ImportanceAnalysis(boolean shouldRemember, String importance, String category, String summary, String reasoning) {
            this.shouldRemember = shouldRemember;
            this.importance = importance;
            this.category = category;
            this.summary = summary;
            this.reasoning = reasoning;
        }

        public boolean isShouldRemember() { return shouldRemember; }
        public void setShouldRemember(boolean shouldRemember) { this.shouldRemember = shouldRemember; }

        public String getImportance() { return importance; }
        public void setImportance(String importance) { this.importance = importance; }

        public String getCategory() { return category; }
        public void setCategory(String category) { this.category = category; }

        public String getSummary() { return summary; }
        public void setSummary(String summary) { this.summary = summary; }

        public String getReasoning() { return reasoning; }
        public void setReasoning(String reasoning) { this.reasoning = reasoning; }

        public static ImportanceAnalysis defaultResult() {
            return new ImportanceAnalysis(true, "MINOR", "PATTERN", "", "Default classification");
        }
    }

    private static class OpenRouterRequest {
        private final String model;
        private final List<Message> messages;
        private final Double temperature;
        private final Integer maxTokens;

        public OpenRouterRequest(String model, List<Message> messages, Double temperature, Integer maxTokens) {
            this.model = model;
            this.messages = messages;
            this.temperature = temperature;
            this.maxTokens = maxTokens;
        }

        public String getModel() { return model; }
        public List<Message> getMessages() { return messages; }
        public Double getTemperature() { return temperature; }
        public Integer getMaxTokens() { return maxTokens; }
    }

    private static class Message {
        private final String role;
        private final String content;

        public Message(String role, String content) {
            this.role = role;
            this.content = content;
        }

        public String getRole() { return role; }
        public String getContent() { return content; }
    }
}
