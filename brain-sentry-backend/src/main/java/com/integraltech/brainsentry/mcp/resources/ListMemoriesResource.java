package com.integraltech.brainsentry.mcp.resources;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.integraltech.brainsentry.dto.response.MemoryListResponse;
import com.integraltech.brainsentry.service.MemoryService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.util.Map;

/**
 * MCP Resource for listing memories in the Brain Sentry system.
 *
 * This resource allows AI agents to list all available memories.
 */
@Slf4j
@Component
@RequiredArgsConstructor
public class ListMemoriesResource {

    private final MemoryService memoryService;

    /**
     * Resource definition for the MCP server.
     */
    public static final String RESOURCE_DEFINITION = """
        {
            "name": "list_memories",
            "description": "List all memories in the Brain Sentry system",
            "uri": "memory://brainsentry/memories"
        }
        """;

    /**
     * List memories for a tenant.
     *
     * @param tenantId the tenant ID
     * @return list of memories as JSON
     */
    public String list(String tenantId) {
        try {
            MemoryListResponse response = memoryService.listMemories(0, 100);

            ObjectMapper mapper = new ObjectMapper();
            return mapper.writeValueAsString(response);

        } catch (Exception e) {
            log.error("Error executing list_memories resource", e);
            ObjectMapper mapper = new ObjectMapper();
            try {
                return mapper.writeValueAsString(Map.of("error", e.getMessage()));
            } catch (Exception ex) {
                return "{\"error\": \"Internal error\"}";
            }
        }
    }

    /**
     * Get the resource definition for MCP discovery.
     */
    public static String getResourceDefinition() {
        return RESOURCE_DEFINITION;
    }
}
