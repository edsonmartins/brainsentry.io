package com.integraltech.brainsentry.mcp;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.integraltech.brainsentry.mcp.tools.CreateMemoryTool;
import com.integraltech.brainsentry.mcp.tools.GetMemoryTool;
import com.integraltech.brainsentry.mcp.tools.InterceptPromptTool;
import com.integraltech.brainsentry.mcp.tools.SearchMemoryTool;
import com.integraltech.brainsentry.mcp.resources.ListMemoriesResource;
import com.integraltech.brainsentry.mcp.prompts.AgentPrompts;
import com.integraltech.brainsentry.service.InterceptionService;
import com.integraltech.brainsentry.service.MemoryService;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.HashMap;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyInt;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.when;

/**
 * Unit tests for MCP Server.
 * These tests don't require full Spring context.
 */
@ExtendWith(MockitoExtension.class)
@DisplayName("MCP Server Unit Tests")
class McpServerTest {

    @Mock
    private MemoryService memoryService;

    @Mock
    private InterceptionService interceptionService;

    private ObjectMapper objectMapper = new ObjectMapper();

    @Nested
    @DisplayName("getAllTools()")
    class GetAllToolsTests {

        @Test
        @DisplayName("Should return all tool definitions")
        void shouldReturnAllToolDefinitions() {
            CreateMemoryTool createMemoryTool = new CreateMemoryTool(memoryService);
            SearchMemoryTool searchMemoryTool = new SearchMemoryTool(memoryService);
            GetMemoryTool getMemoryTool = new GetMemoryTool(memoryService);
            InterceptPromptTool interceptPromptTool = new InterceptPromptTool(interceptionService);

            McpServer mcpServer = new McpServer(
                    createMemoryTool,
                    searchMemoryTool,
                    getMemoryTool,
                    interceptPromptTool,
                    new ListMemoriesResource(memoryService),
                    new AgentPrompts()
            );

            var tools = mcpServer.getAllTools();

            assertThat(tools).isNotNull();
            assertThat(tools).hasSize(4);
            assertThat(tools).containsKey("create_memory");
            assertThat(tools).containsKey("search_memories");
            assertThat(tools).containsKey("get_memory");
            assertThat(tools).containsKey("intercept_prompt");
        }
    }

    @Nested
    @DisplayName("getAllResources()")
    class GetAllResourcesTests {

        @Test
        @DisplayName("Should return all resource definitions")
        void shouldReturnAllResourceDefinitions() {
            ListMemoriesResource listMemoriesResource = new ListMemoriesResource(memoryService);

            McpServer mcpServer = new McpServer(
                    new CreateMemoryTool(memoryService),
                    new SearchMemoryTool(memoryService),
                    new GetMemoryTool(memoryService),
                    new InterceptPromptTool(interceptionService),
                    listMemoriesResource,
                    new AgentPrompts()
            );

            var resources = mcpServer.getAllResources();

            assertThat(resources).isNotNull();
            assertThat(resources).hasSize(1);
            assertThat(resources).containsKey("list_memories");
        }
    }

    @Nested
    @DisplayName("getAllPrompts()")
    class GetAllPromptsTests {

        @Test
        @DisplayName("Should return all prompts")
        void shouldReturnAllPrompts() {
            McpServer mcpServer = new McpServer(
                    new CreateMemoryTool(memoryService),
                    new SearchMemoryTool(memoryService),
                    new GetMemoryTool(memoryService),
                    new InterceptPromptTool(interceptionService),
                    new ListMemoriesResource(memoryService),
                    new AgentPrompts()
            );

            String prompts = mcpServer.getAllPrompts();

            assertThat(prompts).isNotNull();
            assertThat(prompts).contains("\"name\": \"capture_pattern\"");
            assertThat(prompts).contains("\"name\": \"extract_learning\"");
            assertThat(prompts).contains("\"name\": \"summarize_discussion\"");
        }
    }

    @Nested
    @DisplayName("executeTool()")
    class ExecuteToolTests {

        @Test
        @DisplayName("Should execute create_memory tool")
        void shouldExecuteCreateMemoryTool() {
            CreateMemoryTool createMemoryTool = new CreateMemoryTool(memoryService);
            SearchMemoryTool searchMemoryTool = new SearchMemoryTool(memoryService);
            GetMemoryTool getMemoryTool = new GetMemoryTool(memoryService);
            InterceptPromptTool interceptPromptTool = new InterceptPromptTool(interceptionService);

            McpServer mcpServer = new McpServer(
                    createMemoryTool,
                    searchMemoryTool,
                    getMemoryTool,
                    interceptPromptTool,
                    new ListMemoriesResource(memoryService),
                    new AgentPrompts()
            );

            Map<String, Object> arguments = new HashMap<>();
            arguments.put("content", "Test memory content");
            arguments.put("tenantId", "default");

            String result = mcpServer.executeTool("create_memory", arguments);

            // The tool should return a response (even with mocked service returning null)
            assertThat(result).isNotNull();
            assertThat(result).contains("success");
        }

        @Test
        @DisplayName("Should return error for unknown tool")
        void shouldReturnErrorForUnknownTool() {
            McpServer mcpServer = new McpServer(
                    new CreateMemoryTool(memoryService),
                    new SearchMemoryTool(memoryService),
                    new GetMemoryTool(memoryService),
                    new InterceptPromptTool(interceptionService),
                    new ListMemoriesResource(memoryService),
                    new AgentPrompts()
            );

            Map<String, Object> arguments = new HashMap<>();

            String result = mcpServer.executeTool("unknown_tool", arguments);

            assertThat(result).contains("\"error\":");
            assertThat(result).contains("Unknown tool");
        }
    }

    @Nested
    @DisplayName("getResource()")
    class GetResourceTests {

        @Test
        @DisplayName("Should return list of memories")
        void shouldReturnListOfMemories() {
            ListMemoriesResource listMemoriesResource = new ListMemoriesResource(memoryService);

            McpServer mcpServer = new McpServer(
                    new CreateMemoryTool(memoryService),
                    new SearchMemoryTool(memoryService),
                    new GetMemoryTool(memoryService),
                    new InterceptPromptTool(interceptionService),
                    listMemoriesResource,
                    new AgentPrompts()
            );

            Map<String, Object> parameters = new HashMap<>();
            parameters.put("tenantId", "default");

            String result = mcpServer.getResource("list_memories", parameters);

            assertThat(result).isNotNull();
        }

        @Test
        @DisplayName("Should return error for unknown resource")
        void shouldReturnErrorForUnknownResource() {
            McpServer mcpServer = new McpServer(
                    new CreateMemoryTool(memoryService),
                    new SearchMemoryTool(memoryService),
                    new GetMemoryTool(memoryService),
                    new InterceptPromptTool(interceptionService),
                    new ListMemoriesResource(memoryService),
                    new AgentPrompts()
            );

            Map<String, Object> parameters = new HashMap<>();

            String result = mcpServer.getResource("unknown_resource", parameters);

            assertThat(result).contains("\"error\":");
            assertThat(result).contains("Unknown resource");
        }
    }

    @Nested
    @DisplayName("getPrompt()")
    class GetPromptTests {

        @Test
        @DisplayName("Should return capture_pattern prompt")
        void shouldReturnCapturePatternPrompt() {
            McpServer mcpServer = new McpServer(
                    new CreateMemoryTool(memoryService),
                    new SearchMemoryTool(memoryService),
                    new GetMemoryTool(memoryService),
                    new InterceptPromptTool(interceptionService),
                    new ListMemoriesResource(memoryService),
                    new AgentPrompts()
            );

            String prompt = mcpServer.getPrompt("capture_pattern");

            assertThat(prompt).contains("What is the pattern");
            assertThat(prompt).contains("pattern or practice");
        }

        @Test
        @DisplayName("Should return error for unknown prompt")
        void shouldReturnErrorForUnknownPrompt() {
            McpServer mcpServer = new McpServer(
                    new CreateMemoryTool(memoryService),
                    new SearchMemoryTool(memoryService),
                    new GetMemoryTool(memoryService),
                    new InterceptPromptTool(interceptionService),
                    new ListMemoriesResource(memoryService),
                    new AgentPrompts()
            );

            assertThatThrownBy(() -> mcpServer.getPrompt("unknown_prompt"))
                    .isInstanceOf(IllegalArgumentException.class)
                    .hasMessage("Unknown prompt: unknown_prompt");
        }
    }
}
