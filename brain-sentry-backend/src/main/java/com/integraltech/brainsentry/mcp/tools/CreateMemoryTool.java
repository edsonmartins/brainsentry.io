package com.integraltech.brainsentry.mcp.tools;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.integraltech.brainsentry.dto.request.CreateMemoryRequest;
import com.integraltech.brainsentry.mcp.McpErrorHandler;
import com.integraltech.brainsentry.mcp.McpTenantContext;
import com.integraltech.brainsentry.service.MemoryService;
import lombok.RequiredArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import java.util.List;

/**
 * MCP Tool for creating memories in the Brain Sentry system.
 *
 * This tool allows AI agents to create new memories based on their
 * interactions, enabling autonomous memory capture.
 */
@Component
@RequiredArgsConstructor
public class CreateMemoryTool {

    private static final Logger log = LoggerFactory.getLogger(CreateMemoryTool.class);

    private final MemoryService memoryService;

    /**
     * Tool definition for the MCP server.
     */
    public static final JsonNode TOOL_DEFINITION = new ObjectMapper().createObjectNode()
            .put("name", "create_memory")
            .put("description", "Create a new memory in the Brain Sentry system")
            .put("inputSchema", """
                {
                    "type": "object",
                    "properties": {
                        "content": {
                            "type": "string",
                            "description": "The content of the memory to be stored"
                        },
                        "summary": {
                            "type": "string",
                            "description": "A brief summary of the memory"
                        },
                        "category": {
                            "type": "string",
                            "enum": ["DECISION", "PATTERN", "ANTIPATTERN", "DOMAIN", "BUG", "OPTIMIZATION", "INTEGRATION"],
                            "description": "The category of the memory"
                        },
                        "importance": {
                            "type": "string",
                            "enum": ["CRITICAL", "IMPORTANT", "MINOR"],
                            "description": "The importance level of the memory"
                        },
                        "tags": {
                            "type": "array",
                            "items": { "type": "string" },
                            "description": "Tags to categorize the memory"
                        },
                        "tenantId": {
                            "type": "string",
                            "description": "The tenant ID"
                        }
                    },
                    "required": ["content", "tenantId"]
                }
                """);

    /**
     * Execute the tool with the given arguments.
     *
     * @param arguments the input arguments as JSON
     * @return the result as JSON
     */
    public String execute(JsonNode arguments) {
        try {
            // Validate required parameters
            McpErrorHandler.requireParameter(arguments, "arguments");

            // Extract and validate tenantId first
            String tenantId = arguments.has("tenantId") ? arguments.get("tenantId").asText() : null;
            McpErrorHandler.validateTenantId(tenantId);
            tenantId = McpTenantContext.normalizeTenantId(tenantId);

            // Set tenant context for this operation
            McpTenantContext.setTenant(tenantId);

            // Extract other parameters
            String content = arguments.has("content") ? arguments.get("content").asText() : null;
            McpErrorHandler.requireParameter(content, "content");

            String summary = arguments.has("summary") ? arguments.get("summary").asText() : null;
            String category = arguments.has("category") ? arguments.get("category").asText() : null;
            String importance = arguments.has("importance") ? arguments.get("importance").asText() : null;

            // Build the request
            CreateMemoryRequest request = CreateMemoryRequest.builder()
                    .content(content)
                    .summary(summary)
                    .category(category != null ? com.integraltech.brainsentry.domain.enums.MemoryCategory.valueOf(category) : null)
                    .importance(importance != null ? com.integraltech.brainsentry.domain.enums.ImportanceLevel.valueOf(importance) : null)
                    .tags(arguments.has("tags") ? new ObjectMapper().readValue(arguments.get("tags").traverse(),
                            new com.fasterxml.jackson.core.type.TypeReference<java.util.List<String>>() {}) : List.of())
                    .tenantId(tenantId)
                    .build();

            // Create the memory
            var response = memoryService.createMemory(request);

            // Return success result
            ObjectMapper mapper = new ObjectMapper();
            return mapper.createObjectNode()
                    .put("success", true)
                    .put("memoryId", response.getId())
                    .put("message", "Memory created successfully")
                    .put("tenantId", tenantId)
                    .toPrettyString();

        } catch (Exception e) {
            log.error("Error executing create_memory tool", e);
            return McpErrorHandler.handleException(e, "create_memory");
        }
    }

    /**
     * Get the tool definition for MCP discovery.
     */
    public static String getToolDefinition() {
        return TOOL_DEFINITION.toPrettyString();
    }
}
