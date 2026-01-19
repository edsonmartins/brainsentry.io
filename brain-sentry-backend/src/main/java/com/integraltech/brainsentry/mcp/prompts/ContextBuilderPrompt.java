package com.integraltech.brainsentry.mcp.prompts;

import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

/**
 * MCP Prompt for building context from memories.
 *
 * This prompt helps AI agents construct comprehensive context
 * from stored memories for enhanced prompt generation.
 */
@Slf4j
@Component
public class ContextBuilderPrompt {

    /**
     * Template for building context from memories.
     * This prompt instructs the LLM on how to structure context
     * injected into user prompts.
     */
    public static final String CONTEXT_BUILDER_TEMPLATE = """
        You are the Brain Sentry Context Builder. Your task is to construct
        a comprehensive context block from the provided memories.

        Input: A list of memories with the following fields:
        - id: Unique identifier
        - summary: Brief description
        - category: Memory type (DECISION, PATTERN, ANTIPATTERN, DOMAIN, BUG, OPTIMIZATION, INTEGRATION)
        - importance: Importance level (CRITICAL, IMPORTANT, MINOR)
        - content: Full memory content
        - codeExample: Optional code snippet
        - programmingLanguage: Language for code example

        Output Format:
        <system_context>
        The following relevant patterns and decisions from {tenant_name} were found:

        [1] {IMPORTANCE} - {CATEGORY}
            {summary}
            {full_content}
            {code_example_if_available}

        [2] {IMPORTANCE} - {CATEGORY}
            ...

        </system_context>

        Guidelines:
        1. Prioritize CRITICAL and IMPORTANT memories first
        2. Group related memories by category
        3. Include code examples when available
        4. Keep summaries concise but informative
        5. Omit MINOR importance memories if context is too long
        6. Target maximum of 500 tokens for the entire context block

        Response should be the formatted context block only, without additional commentary.
        """;

    /**
     * Get the context builder prompt template.
     *
     * @return the prompt template
     */
    public String getPromptTemplate() {
        return CONTEXT_BUILDER_TEMPLATE;
    }

    /**
     * Get the context builder prompt for MCP discovery.
     *
     * @return JSON representation of the prompt
     */
    public static String getPromptDefinition() {
        return """
            {
                "name": "context_builder",
                "description": "Build comprehensive context from stored memories for prompt enhancement",
                "arguments": {
                    "memories": {
                        "type": "array",
                        "description": "List of memories to include in context",
                        "items": {
                            "type": "object"
                        }
                    },
                    "tenantName": {
                        "type": "string",
                        "description": "Name of the tenant/organization"
                    },
                    "maxTokens": {
                        "type": "number",
                        "description": "Maximum tokens for context (default: 500)",
                        "default": 500
                    }
                }
            }
            """;
    }

    /**
     * Get a formatted prompt with arguments.
     *
     * @param tenantName the tenant name
     * @param maxTokens maximum tokens
     * @return the formatted prompt
     */
    public String getFormattedPrompt(String tenantName, Integer maxTokens) {
        return CONTEXT_BUILDER_TEMPLATE
                .replace("{tenant_name}", tenantName != null ? tenantName : "the system")
                .replace("500", String.valueOf(maxTokens != null ? maxTokens : 500));
    }
}
