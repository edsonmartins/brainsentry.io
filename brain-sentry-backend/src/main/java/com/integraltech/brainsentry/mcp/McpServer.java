package com.integraltech.brainsentry.mcp;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.integraltech.brainsentry.mcp.prompts.AgentPrompts;
import com.integraltech.brainsentry.mcp.resources.ListMemoriesResource;
import com.integraltech.brainsentry.mcp.tools.CreateMemoryTool;
import com.integraltech.brainsentry.mcp.tools.GetMemoryTool;
import com.integraltech.brainsentry.mcp.tools.InterceptPromptTool;
import com.integraltech.brainsentry.mcp.tools.SearchMemoryTool;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.util.HashMap;
import java.util.Map;

/**
 * MCP Server implementation for Brain Sentry.
 *
 * Exposes tools, resources, and prompts for AI agent interaction
 * through the Model Context Protocol (MCP).
 */
@Service
public class McpServer {

    private static final Logger log = LoggerFactory.getLogger(McpServer.class);

    private final CreateMemoryTool createMemoryTool;
    private final SearchMemoryTool searchMemoryTool;
    private final GetMemoryTool getMemoryTool;
    private final InterceptPromptTool interceptPromptTool;
    private final ListMemoriesResource listMemoriesResource;
    private final AgentPrompts agentPrompts;

    private final ObjectMapper objectMapper = new ObjectMapper();

    public McpServer(CreateMemoryTool createMemoryTool,
                      SearchMemoryTool searchMemoryTool,
                      GetMemoryTool getMemoryTool,
                      InterceptPromptTool interceptPromptTool,
                      ListMemoriesResource listMemoriesResource,
                      AgentPrompts agentPrompts) {
        this.createMemoryTool = createMemoryTool;
        this.searchMemoryTool = searchMemoryTool;
        this.getMemoryTool = getMemoryTool;
        this.interceptPromptTool = interceptPromptTool;
        this.listMemoriesResource = listMemoriesResource;
        this.agentPrompts = agentPrompts;
    }

    /**
     * Get all available tools.
     *
     * @return map of tool definitions
     */
    public Map<String, String> getAllTools() {
        try {
            Map<String, String> tools = new HashMap<>();

            tools.put("create_memory", CreateMemoryTool.getToolDefinition());
            tools.put("search_memories", SearchMemoryTool.getToolDefinition());
            tools.put("get_memory", GetMemoryTool.getToolDefinition());
            tools.put("intercept_prompt", InterceptPromptTool.getToolDefinition());

            return tools;
        } catch (Exception e) {
            log.error("Error getting tools", e);
            return Map.of();
        }
    }

    /**
     * Get all available resources.
     *
     * @return map of resource definitions
     */
    public Map<String, String> getAllResources() {
        try {
            Map<String, String> resources = new HashMap<>();

            resources.put("list_memories", ListMemoriesResource.getResourceDefinition());

            return resources;
        } catch (Exception e) {
            log.error("Error getting resources", e);
            return Map.of();
        }
    }

    /**
     * Get all available prompts.
     *
     * @return prompts definition
     */
    public String getAllPrompts() {
        return agentPrompts.getAllPrompts();
    }

    /**
     * Execute a tool by name with the given arguments.
     *
     * @param toolName the tool name
     * @param arguments the input arguments
     * @return tool execution result
     */
    public String executeTool(String toolName, Map<String, Object> arguments) {
        try {
            Map<String, Object> argsMap = arguments;

            return switch (toolName) {
                case "create_memory" -> createMemoryTool.execute(objectMapper.valueToTree(argsMap));
                case "search_memories" -> searchMemoryTool.execute(objectMapper.valueToTree(argsMap));
                case "get_memory" -> getMemoryTool.execute(objectMapper.valueToTree(argsMap));
                case "intercept_prompt" -> interceptPromptTool.execute(objectMapper.valueToTree(argsMap));
                default -> String.format("{\"error\": \"Unknown tool: %s\", \"success\": false}", toolName);
            };
        } catch (Exception e) {
            log.error("Error executing tool: {}", toolName, e);
            return String.format("{\"error\": \"%s\", \"errorType\": \"%s\", \"success\": false}",
                e.getMessage(), e.getClass().getSimpleName());
        }
    }

    /**
     * Get a resource by name.
     *
     * @param resourceName the resource name
     * @param parameters additional parameters
     * @return resource data
     */
    public String getResource(String resourceName, Map<String, Object> parameters) {
        try {
            if ("list_memories".equals(resourceName)) {
                String tenantId = parameters != null ? (String) parameters.get("tenantId") : null;
                return listMemoriesResource.list(tenantId);
            }

            return String.format("{\"error\": \"Unknown resource: %s\"}", resourceName);
        } catch (Exception e) {
            log.error("Error getting resource: {}", resourceName, e);
            return String.format("{\"error\": \"%s\"}", e.getMessage());
        }
    }

    /**
     * Get a prompt by name.
     *
     * @param promptName the prompt name
     * @return prompt template
     */
    public String getPrompt(String promptName) {
        return agentPrompts.getPrompt(promptName);
    }
}
