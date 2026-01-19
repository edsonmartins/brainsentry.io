package com.integraltech.brainsentry.mcp.tools;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.integraltech.brainsentry.dto.request.SearchRequest;
import com.integraltech.brainsentry.mcp.McpErrorHandler;
import com.integraltech.brainsentry.mcp.McpTenantContext;
import com.integraltech.brainsentry.service.MemoryService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

/**
 * MCP Tool for searching memories in the Brain Sentry system.
 *
 * This tool allows AI agents to perform semantic search across memories.
 */
@Slf4j
@Component
@RequiredArgsConstructor
public class SearchMemoryTool {

    private final MemoryService memoryService;

    /**
     * Tool definition for the MCP server.
     */
    public static final JsonNode TOOL_DEFINITION = new ObjectMapper().createObjectNode()
            .put("name", "search_memories")
            .put("description", "Search memories in the Brain Sentry system using semantic search")
            .put("inputSchema", """
                {
                    "type": "object",
                    "properties": {
                        "query": {
                            "type": "string",
                            "description": "The search query"
                        },
                        "limit": {
                            "type": "number",
                            "description": "Maximum number of results to return (default: 10)",
                            "default": 10
                        },
                        "tenantId": {
                            "type": "string",
                            "description": "The tenant ID"
                        }
                    },
                    "required": ["query", "tenantId"]
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
            String query = arguments.has("query") ? arguments.get("query").asText() : null;
            McpErrorHandler.requireParameter(query, "query");

            int limit = arguments.has("limit") ? arguments.get("limit").asInt() : 10;

            SearchRequest request = SearchRequest.builder()
                    .query(query)
                    .limit(limit)
                    .tenantId(tenantId)
                    .build();

            var results = memoryService.search(request);

            ObjectMapper mapper = new ObjectMapper();
            return mapper.createObjectNode()
                    .put("success", true)
                    .put("count", results.size())
                    .put("tenantId", tenantId)
                    .put("memories", new ObjectMapper().valueToTree(results))
                    .toPrettyString();

        } catch (Exception e) {
            log.error("Error executing search_memories tool", e);
            return McpErrorHandler.handleException(e, "search_memories");
        }
    }

    /**
     * Get the tool definition for MCP discovery.
     */
    public static String getToolDefinition() {
        return TOOL_DEFINITION.toPrettyString();
    }
}
