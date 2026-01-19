package com.integraltech.brainsentry.mcp.tools;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.integraltech.brainsentry.dto.response.MemoryResponse;
import com.integraltech.brainsentry.mcp.McpErrorHandler;
import com.integraltech.brainsentry.mcp.McpTenantContext;
import com.integraltech.brainsentry.service.MemoryService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

/**
 * MCP Tool for retrieving a specific memory by ID.
 *
 * This tool allows AI agents to get details about a specific memory.
 */
@Slf4j
@Component
@RequiredArgsConstructor
public class GetMemoryTool {

    private final MemoryService memoryService;

    /**
     * Tool definition for the MCP server.
     */
    public static final JsonNode TOOL_DEFINITION = new ObjectMapper().createObjectNode()
            .put("name", "get_memory")
            .put("description", "Get a specific memory by ID from the Brain Sentry system")
            .put("inputSchema", """
                {
                    "type": "object",
                    "properties": {
                        "memoryId": {
                            "type": "string",
                            "description": "The ID of the memory to retrieve"
                        },
                        "tenantId": {
                            "type": "string",
                            "description": "The tenant ID"
                        }
                    },
                    "required": ["memoryId", "tenantId"]
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

            // Extract memoryId
            String memoryId = arguments.has("memoryId") ? arguments.get("memoryId").asText() : null;
            McpErrorHandler.requireParameter(memoryId, "memoryId");

            MemoryResponse memory = memoryService.getMemory(memoryId);

            // Verify tenant access (if memory has tenant info)
            if (memory.getTenantId() != null && !memory.getTenantId().equals(tenantId)) {
                throw new IllegalStateException(
                    "Memory belongs to a different tenant: " + memory.getTenantId()
                );
            }

            ObjectMapper mapper = new ObjectMapper();
            return mapper.createObjectNode()
                    .put("success", true)
                    .put("tenantId", tenantId)
                    .put("memory", new ObjectMapper().valueToTree(memory))
                    .toPrettyString();

        } catch (Exception e) {
            log.error("Error executing get_memory tool", e);
            return McpErrorHandler.handleException(e, "get_memory");
        }
    }

    /**
     * Get the tool definition for MCP discovery.
     */
    public static String getToolDefinition() {
        return TOOL_DEFINITION.toPrettyString();
    }
}
