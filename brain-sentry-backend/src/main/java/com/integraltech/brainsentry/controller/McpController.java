package com.integraltech.brainsentry.controller;

import com.integraltech.brainsentry.mcp.McpServer;
import lombok.RequiredArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

/**
 * REST controller for MCP (Model Context Protocol) Server.
 *
 * Exposes MCP tools, resources, and prompts as HTTP endpoints
 * for integration with AI agents and Claude Desktop.
 */
@RestController
@RequestMapping("/v1/mcp")
@RequiredArgsConstructor
public class McpController {

    private static final Logger log = LoggerFactory.getLogger(McpController.class);

    private final McpServer mcpServer;

    /**
     * Get all available MCP tools.
     * GET /v1/mcp/tools
     */
    @GetMapping(value = "/tools", produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<Map<String, String>> getTools() {
        log.info("GET /v1/mcp/tools");
        Map<String, String> tools = mcpServer.getAllTools();
        return ResponseEntity.ok(tools);
    }

    /**
     * Get all available MCP resources.
     * GET /v1/mcp/resources
     */
    @GetMapping(value = "/resources", produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<Map<String, String>> getResources() {
        log.info("GET /v1/mcp/resources");
        Map<String, String> resources = mcpServer.getAllResources();
        return ResponseEntity.ok(resources);
    }

    /**
     * Get all available MCP prompts.
     * GET /v1/mcp/prompts
     */
    @GetMapping(value = "/prompts", produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<String> getPrompts() {
        log.info("GET /v1/mcp/prompts");
        String prompts = mcpServer.getAllPrompts();
        return ResponseEntity.ok(prompts);
    }

    /**
     * Execute an MCP tool by name.
     * POST /v1/mcp/tools/execute
     */
    @PostMapping(value = "/tools/execute", produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<String> executeTool(
            @RequestParam String toolName,
            @RequestBody Map<String, Object> arguments
    ) {
        log.info("POST /v1/mcp/tools/execute - tool: {}", toolName);
        String result = mcpServer.executeTool(toolName, arguments);
        return ResponseEntity.ok(result);
    }

    /**
     * Get an MCP resource by name.
     * GET /v1/mcp/resources/{resourceName}
     */
    @GetMapping(value = "/resources/{resourceName}", produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<String> getResource(
            @PathVariable String resourceName,
            @RequestParam(required = false) Map<String, Object> parameters
    ) {
        log.info("GET /v1/mcp/resources/{}", resourceName);
        String result = mcpServer.getResource(resourceName, parameters);
        return ResponseEntity.ok(result);
    }

    /**
     * Get an MCP prompt by name.
     * GET /v1/mcp/prompts/{promptName}
     */
    @GetMapping(value = "/prompts/{promptName}", produces = MediaType.TEXT_PLAIN_VALUE)
    public ResponseEntity<String> getPrompt(@PathVariable String promptName) {
        log.info("GET /v1/mcp/prompts/{}", promptName);
        String prompt = mcpServer.getPrompt(promptName);
        return ResponseEntity.ok(prompt);
    }

    /**
     * Health check for MCP server.
     * GET /v1/mcp/health
     */
    @GetMapping(value = "/health", produces = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<Map<String, String>> health() {
        log.info("GET /v1/mcp/health");
        return ResponseEntity.ok(Map.of(
                "status", "healthy",
                "service", "Brain Sentry MCP Server"
        ));
    }
}
