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
     * Analyze the relationship between two pieces of content.
     * Uses LLM to determine if and how they are related.
     *
     * @param content1 first content to compare
     * @param content2 second content to compare
     * @return RelationshipAnalysis with type, confidence, and reasoning
     */
    public RelationshipAnalysis analyzeRelationship(String content1, String content2) {
        String prompt = buildRelationshipPrompt(content1, content2);

        try {
            String response = callGrok(prompt, 300);
            return parseRelationshipAnalysis(response);
        } catch (Exception e) {
            log.error("Error analyzing relationship", e);
            return RelationshipAnalysis.noRelationship();
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

    /**
     * Extract entities and relationships from a single piece of content.
     * This method analyzes text and identifies:
     * - Named entities (people, organizations, products, concepts, etc.)
     * - Relationships between those entities
     *
     * Example input: "Cliente Marcos Silva fez pedido #12345 de Laptop ProMax com vendedor Ana"
     * Output entities: CLIENTE:Marcos Silva, PEDIDO:#12345, PRODUTO:Laptop ProMax, VENDEDOR:Ana
     * Output relationships: (Marcos Silva)-[REALIZOU]->(#12345), (Ana)-[ATENDEU]->(Marcos Silva)
     *
     * @param content the content to analyze
     * @return EntityExtractionResult with entities and relationships
     */
    public EntityExtractionResult extractEntitiesAndRelationships(String content) {
        if (content == null || content.isBlank()) {
            return EntityExtractionResult.empty();
        }

        String prompt = buildEntityExtractionPrompt(content);

        try {
            String response = callGrok(prompt, 1000);
            return parseEntityExtractionResult(response);
        } catch (Exception e) {
            log.error("Error extracting entities and relationships", e);
            return EntityExtractionResult.empty();
        }
    }

    /**
     * Build prompt for entity and relationship extraction.
     */
    private String buildEntityExtractionPrompt(String content) {
        // Truncate if too long
        String c = content.length() > 2000 ? content.substring(0, 2000) + "..." : content;

        return String.format("""
            Analyze this text and extract all named entities and their relationships.

            Text: %s

            INSTRUCTIONS:
            1. Identify all named entities (people, organizations, products, locations, concepts, events, etc.)
            2. Assign each entity a type (PESSOA, ORGANIZACAO, PRODUTO, LOCAL, CONCEITO, EVENTO, PEDIDO, DATA, VALOR, etc.)
            3. Identify relationships between entities found in the text
            4. Use descriptive relationship types in Portuguese (REALIZOU, ATENDEU, CONTEM, PERTENCE_A, TRABALHA_EM, LOCALIZADO_EM, etc.)

            Respond ONLY with valid JSON in this exact format:
            {
              "entities": [
                {"id": "e1", "name": "Entity Name", "type": "ENTITY_TYPE", "properties": {"key": "value"}},
                {"id": "e2", "name": "Another Entity", "type": "ENTITY_TYPE", "properties": {}}
              ],
              "relationships": [
                {"sourceId": "e1", "targetId": "e2", "type": "RELATIONSHIP_TYPE", "properties": {"key": "value"}}
              ]
            }

            Rules:
            - Each entity must have a unique id (e1, e2, e3...)
            - Entity types should be UPPERCASE
            - Relationship types should be UPPERCASE and descriptive
            - If no entities or relationships found, return empty arrays
            - Properties are optional but useful for additional context
            """, c);
    }

    /**
     * Parse entity extraction result from LLM response.
     */
    private EntityExtractionResult parseEntityExtractionResult(String response) {
        try {
            String json = extractJson(response);
            JsonNode root = objectMapper.readTree(json);

            List<ExtractedEntity> entities = new ArrayList<>();
            List<ExtractedRelationship> relationships = new ArrayList<>();

            // Parse entities
            JsonNode entitiesNode = root.path("entities");
            if (entitiesNode.isArray()) {
                for (JsonNode entityNode : entitiesNode) {
                    ExtractedEntity entity = new ExtractedEntity();
                    entity.setId(entityNode.path("id").asText());
                    entity.setName(entityNode.path("name").asText());
                    entity.setType(entityNode.path("type").asText());

                    // Parse properties
                    JsonNode propsNode = entityNode.path("properties");
                    if (propsNode.isObject()) {
                        Map<String, String> props = new java.util.HashMap<>();
                        propsNode.fields().forEachRemaining(field ->
                            props.put(field.getKey(), field.getValue().asText())
                        );
                        entity.setProperties(props);
                    }

                    entities.add(entity);
                }
            }

            // Parse relationships
            JsonNode relationshipsNode = root.path("relationships");
            if (relationshipsNode.isArray()) {
                for (JsonNode relNode : relationshipsNode) {
                    ExtractedRelationship rel = new ExtractedRelationship();
                    rel.setSourceId(relNode.path("sourceId").asText());
                    rel.setTargetId(relNode.path("targetId").asText());
                    rel.setType(relNode.path("type").asText());

                    // Parse properties
                    JsonNode propsNode = relNode.path("properties");
                    if (propsNode.isObject()) {
                        Map<String, String> props = new java.util.HashMap<>();
                        propsNode.fields().forEachRemaining(field ->
                            props.put(field.getKey(), field.getValue().asText())
                        );
                        rel.setProperties(props);
                    }

                    relationships.add(rel);
                }
            }

            return new EntityExtractionResult(entities, relationships);
        } catch (Exception e) {
            log.warn("Failed to parse entity extraction result: {}", response, e);
            return EntityExtractionResult.empty();
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

    private String buildRelationshipPrompt(String content1, String content2) {
        // Truncate content to avoid token limits
        String c1 = content1.length() > 500 ? content1.substring(0, 500) + "..." : content1;
        String c2 = content2.length() > 500 ? content2.substring(0, 500) + "..." : content2;

        return String.format("""
            Analyze the relationship between these two pieces of knowledge/content.

            Content 1: %s

            Content 2: %s

            Determine if they are meaningfully related and what type of relationship exists.

            Relationship types:
            - REQUIRES: Content 1 depends on or requires Content 2
            - CONFLICTS_WITH: Content 1 contradicts or conflicts with Content 2
            - SUPERSEDES: Content 1 replaces or updates Content 2
            - RELATED_TO: General semantic relationship (same topic/domain)
            - PART_OF: Content 1 is a component or subset of Content 2
            - USED_WITH: Content 1 and 2 are frequently used together

            Respond in JSON format:
            {
              "hasRelationship": true/false,
              "type": "REQUIRES|CONFLICTS_WITH|SUPERSEDES|RELATED_TO|PART_OF|USED_WITH",
              "confidence": 0.0-1.0,
              "reasoning": "brief explanation of why they are related"
            }
            """, c1, c2);
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

    private RelationshipAnalysis parseRelationshipAnalysis(String response) {
        try {
            String json = extractJson(response);
            JsonNode root = objectMapper.readTree(json);

            return new RelationshipAnalysis(
                root.path("hasRelationship").asBoolean(false),
                root.path("type").asText("RELATED_TO"),
                root.path("confidence").asDouble(0.0),
                root.path("reasoning").asText("")
            );
        } catch (Exception e) {
            log.warn("Failed to parse relationship analysis: {}", response, e);
            return RelationshipAnalysis.noRelationship();
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

    public static class RelationshipAnalysis {
        private boolean hasRelationship;
        private String type;  // REQUIRES, CONFLICTS_WITH, SUPERSEDES, RELATED_TO, PART_OF, USED_WITH
        private Double confidence;
        private String reasoning;

        public RelationshipAnalysis() {}

        public RelationshipAnalysis(boolean hasRelationship, String type, Double confidence, String reasoning) {
            this.hasRelationship = hasRelationship;
            this.type = type;
            this.confidence = confidence;
            this.reasoning = reasoning;
        }

        public static RelationshipAnalysis noRelationship() {
            return new RelationshipAnalysis(false, null, 0.0, "No relationship detected");
        }

        public boolean isHasRelationship() { return hasRelationship; }
        public void setHasRelationship(boolean hasRelationship) { this.hasRelationship = hasRelationship; }

        public String getType() { return type; }
        public void setType(String type) { this.type = type; }

        public Double getConfidence() { return confidence; }
        public void setConfidence(Double confidence) { this.confidence = confidence; }

        public String getReasoning() { return reasoning; }
        public void setReasoning(String reasoning) { this.reasoning = reasoning; }
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

    // ==================== Entity Extraction DTOs ====================

    /**
     * Result of entity and relationship extraction from text.
     */
    public static class EntityExtractionResult {
        private List<ExtractedEntity> entities;
        private List<ExtractedRelationship> relationships;

        public EntityExtractionResult() {
            this.entities = new ArrayList<>();
            this.relationships = new ArrayList<>();
        }

        public EntityExtractionResult(List<ExtractedEntity> entities, List<ExtractedRelationship> relationships) {
            this.entities = entities != null ? entities : new ArrayList<>();
            this.relationships = relationships != null ? relationships : new ArrayList<>();
        }

        public static EntityExtractionResult empty() {
            return new EntityExtractionResult(new ArrayList<>(), new ArrayList<>());
        }

        public boolean hasEntities() {
            return entities != null && !entities.isEmpty();
        }

        public boolean hasRelationships() {
            return relationships != null && !relationships.isEmpty();
        }

        public List<ExtractedEntity> getEntities() { return entities; }
        public void setEntities(List<ExtractedEntity> entities) { this.entities = entities; }

        public List<ExtractedRelationship> getRelationships() { return relationships; }
        public void setRelationships(List<ExtractedRelationship> relationships) { this.relationships = relationships; }

        @Override
        public String toString() {
            return "EntityExtractionResult{entities=" + entities.size() + ", relationships=" + relationships.size() + "}";
        }
    }

    /**
     * An extracted entity from text.
     * Example: {id: "e1", name: "Marcos Silva", type: "CLIENTE", properties: {telefone: "11999999999"}}
     */
    public static class ExtractedEntity {
        private String id;         // Unique identifier within extraction (e.g., "e1", "e2")
        private String name;       // Entity name (e.g., "Marcos Silva")
        private String type;       // Entity type (e.g., "CLIENTE", "PRODUTO", "PEDIDO")
        private Map<String, String> properties; // Additional properties

        public ExtractedEntity() {
            this.properties = new java.util.HashMap<>();
        }

        public ExtractedEntity(String id, String name, String type) {
            this.id = id;
            this.name = name;
            this.type = type;
            this.properties = new java.util.HashMap<>();
        }

        public String getId() { return id; }
        public void setId(String id) { this.id = id; }

        public String getName() { return name; }
        public void setName(String name) { this.name = name; }

        public String getType() { return type; }
        public void setType(String type) { this.type = type; }

        public Map<String, String> getProperties() { return properties; }
        public void setProperties(Map<String, String> properties) { this.properties = properties; }

        @Override
        public String toString() {
            return type + ":" + name;
        }
    }

    /**
     * An extracted relationship between two entities.
     * Example: {sourceId: "e1", targetId: "e2", type: "REALIZOU", properties: {data: "2025-01-20"}}
     */
    public static class ExtractedRelationship {
        private String sourceId;   // Source entity ID
        private String targetId;   // Target entity ID
        private String type;       // Relationship type (e.g., "REALIZOU", "ATENDEU", "CONTEM")
        private Map<String, String> properties; // Additional properties

        public ExtractedRelationship() {
            this.properties = new java.util.HashMap<>();
        }

        public ExtractedRelationship(String sourceId, String targetId, String type) {
            this.sourceId = sourceId;
            this.targetId = targetId;
            this.type = type;
            this.properties = new java.util.HashMap<>();
        }

        public String getSourceId() { return sourceId; }
        public void setSourceId(String sourceId) { this.sourceId = sourceId; }

        public String getTargetId() { return targetId; }
        public void setTargetId(String targetId) { this.targetId = targetId; }

        public String getType() { return type; }
        public void setType(String type) { this.type = type; }

        public Map<String, String> getProperties() { return properties; }
        public void setProperties(Map<String, String> properties) { this.properties = properties; }

        @Override
        public String toString() {
            return "(" + sourceId + ")-[" + type + "]->(" + targetId + ")";
        }
    }
}
