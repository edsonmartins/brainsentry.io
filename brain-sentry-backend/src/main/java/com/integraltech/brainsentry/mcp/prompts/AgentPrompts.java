package com.integraltech.brainsentry.mcp.prompts;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

/**
 * MCP Prompts for common AI agent interactions.
 *
 * These are pre-defined prompts that AI agents can use to interact with
 * the Brain Sentry system more effectively.
 */
@Slf4j
@Component
@RequiredArgsConstructor
public class AgentPrompts {

    /**
     * Prompt for capturing code patterns from development context.
     */
    public static final String CAPTURE_PATTERN_PROMPT = """
        You are analyzing code context and have identified a potential pattern or practice.

        Answer the following questions:
        1. What is the pattern or practice you identified?
        2. Why is this significant? (architectural decision, performance optimization, bug fix, etc.)
        3. What category does this belong to? (DECISION, PATTERN, ANTIPATTERN, DOMAIN, BUG, OPTIMIZATION, INTEGRATION)
        4. What importance level is this? (CRITICAL, IMPORTANT, MINOR)

        Provide your answer as JSON:
        {
            "content": "The pattern or practice you identified",
            "reasoning": "Why this is significant",
            "category": "One of: DECISION, PATTERN, ANTIPATTERN, DOMAIN, BUG, OHAPTERIMIZATION, INTEGRATION",
            "importance": "One of: CRITICAL, IMPORTANT, MINOR"
        }
        """;

    /**
     * Prompt for extracting key learnings from development sessions.
     */
    public static final String EXTRACT_LEARNING_PROMPT = """
        You are analyzing a development session and have identified a learning opportunity.

        Answer the following questions:
        1. What was learned? (concept, technique, workaround, issue)
        2. How was this resolved or discovered?
        3. What are the key takeaways?

        Provide your answer as JSON:
        {
            "content": "What was learned",
            "how": "How it was discovered",
            "takeaways": ["key takeaway 1", "key takeaway 2"]
        }
        """;

    /**
     * Prompt for summarizing technical discussions.
     */
    public static final String SUMMARIZE_DISCUSSION_PROMPT = """
        You are summarizing a technical discussion.

        Extract:
        1. The main topic discussed
        2. Key points raised
        3. Any decisions made
        4. Action items

        Provide your answer as JSON:
        {
            "topic": "Main topic",
            "keyPoints": ["point 1", "point 2", "point 3"],
            "decisions": ["decision 1", "decision 2"],
            "actionItems": ["action 1", "action 2"]
        }
        """;

    /**
     * Get all available prompts for MCP discovery.
     */
    public static String getAllPrompts() {
        return """
            {
                "prompts": [
                    {
                        "name": "capture_pattern",
                        "description": "Capture code patterns from context",
                        "prompt": "\"" + CAPTURE_PATTERN_PROMPT.replace("\n", "\\n") + "\""
                    },
                    {
                        "name": "extract_learning",
                        "description": "Extract learnings from development sessions",
                        "prompt": "\"" + EXTRACT_LEARNING_PROMPT.replace("\n", "\\n") + "\""
                    },
                    {
                        "name": "summarize_discussion",
                        "description": "Summarize technical discussions",
                        "prompt": "\"" + SUMMARIZE_DISCUSSION_PROMPT.replace("\n", "\\n") + "\""
                    },
                    {
                        "name": "context_builder",
                        "description": "Build comprehensive context from stored memories for prompt enhancement",
                        "prompt": "Context builder template for structuring memory context injection"
                    }
                ]
            }
            """;
    }

    /**
     * Get a specific prompt by name.
     *
     * @param name the prompt name
     * @return the prompt template
     */
    public String getPrompt(String name) {
        return switch (name) {
            case "capture_pattern" -> CAPTURE_PATTERN_PROMPT;
            case "extract_learning" -> EXTRACT_LEARNING_PROMPT;
            case "summarize_discussion" -> SUMMARIZE_DISCUSSION_PROMPT;
            case "context_builder" -> """
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
                The following relevant patterns and decisions from the system were found:

                [1] {IMPORTANCE} - {CATEGORY}
                    {summary}
                    {full_content}
                    {code_example_if_available}

                </system_context>

                Guidelines:
                1. Prioritize CRITICAL and IMPORTANT memories first
                2. Group related memories by category
                3. Include code examples when available
                4. Keep summaries concise but informative
                5. Omit MINOR importance memories if context is too long
                6. Target maximum of 500 tokens for the entire context block
                """;
            default -> throw new IllegalArgumentException("Unknown prompt: " + name);
        };
    }
}
