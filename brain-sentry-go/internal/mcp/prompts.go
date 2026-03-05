package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/integraltech/brainsentry/internal/dto"
)

func (s *Server) registerPrompts() { //nolint:funlen
	// Spec-required prompts
	s.prompts["capture_pattern"] = Prompt{
		Name:        "capture_pattern",
		Description: "Guide the user through capturing a coding pattern or best practice as a memory",
		Arguments: []PromptArg{
			{Name: "pattern", Description: "The pattern or practice to capture", Required: true},
			{Name: "language", Description: "Programming language (optional)"},
		},
		Handler: s.promptCapturePattern,
	}

	s.prompts["extract_learning"] = Prompt{
		Name:        "extract_learning",
		Description: "Extract learnings and insights from a coding session or conversation",
		Arguments: []PromptArg{
			{Name: "session", Description: "Description of the session or conversation", Required: true},
		},
		Handler: s.promptExtractLearning,
	}

	s.prompts["summarize_discussion"] = Prompt{
		Name:        "summarize_discussion",
		Description: "Summarize a technical discussion and extract key decisions and action items",
		Arguments: []PromptArg{
			{Name: "discussion", Description: "The discussion content to summarize", Required: true},
		},
		Handler: s.promptSummarizeDiscussion,
	}

	s.prompts["context_builder"] = Prompt{
		Name:        "context_builder",
		Description: "Build comprehensive context from memories for a specific task or question",
		Arguments: []PromptArg{
			{Name: "task", Description: "The task or question to build context for", Required: true},
			{Name: "maxMemories", Description: "Maximum memories to include (default: 10)"},
		},
		Handler: s.promptAgentContext, // Reuse agent_context implementation
	}

	s.prompts["agent_context"] = Prompt{
		Name:        "agent_context",
		Description: "Build a context prompt with relevant memories for an AI agent",
		Arguments: []PromptArg{
			{Name: "task", Description: "The task or question the agent needs context for", Required: true},
			{Name: "maxMemories", Description: "Maximum number of memories to include (default: 5)"},
		},
		Handler: s.promptAgentContext,
	}

	s.prompts["memory_summary"] = Prompt{
		Name:        "memory_summary",
		Description: "Generate a summary of all stored memories",
		Handler:     s.promptMemorySummary,
	}

	s.prompts["hindsight_review"] = Prompt{
		Name:        "hindsight_review",
		Description: "Review past errors and their resolutions to avoid repeating mistakes",
		Arguments: []PromptArg{
			{Name: "topic", Description: "Optional topic to filter hindsight notes"},
		},
		Handler: s.promptHindsightReview,
	}
}

func (s *Server) promptAgentContext(ctx context.Context, args map[string]string) ([]PromptMessage, error) {
	task := args["task"]
	if task == "" {
		return nil, fmt.Errorf("task argument is required")
	}

	// Search for relevant memories
	searchResp, err := s.memoryService.SearchMemories(ctx, dto.SearchRequest{Query: task, Limit: 5})
	if err != nil {
		return nil, fmt.Errorf("failed to search memories: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("You have access to the following relevant context from Brain Sentry:\n\n")

	if len(searchResp.Results) == 0 {
		sb.WriteString("No relevant memories found for this task.\n")
	} else {
		for i, m := range searchResp.Results {
			sb.WriteString(fmt.Sprintf("### Memory %d [%s / %s]\n", i+1, m.Category, m.Importance))
			if m.Summary != "" {
				sb.WriteString(m.Summary + "\n")
			} else {
				content := m.Content
				if len(content) > 300 {
					content = content[:300] + "..."
				}
				sb.WriteString(content + "\n")
			}
			if m.CodeExample != "" {
				sb.WriteString(fmt.Sprintf("```%s\n%s\n```\n", m.ProgrammingLanguage, m.CodeExample))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\nUse this context to inform your response to the user's task.")

	return []PromptMessage{
		{
			Role: "user",
			Content: PromptContent{
				Type: "text",
				Text: sb.String(),
			},
		},
	}, nil
}

func (s *Server) promptMemorySummary(ctx context.Context, args map[string]string) ([]PromptMessage, error) {
	resp, err := s.memoryService.ListMemories(ctx, 0, 50)
	if err != nil {
		return nil, fmt.Errorf("failed to list memories: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("Here is a summary of all stored memories in Brain Sentry:\n\n")

	if resp.TotalElements == 0 {
		sb.WriteString("No memories stored yet.\n")
	} else {
		sb.WriteString(fmt.Sprintf("Total: %d memories\n\n", resp.TotalElements))
		for i, m := range resp.Memories {
			sb.WriteString(fmt.Sprintf("%d. [%s/%s] %s\n", i+1, m.Category, m.Importance, m.Summary))
			if len(m.Tags) > 0 {
				sb.WriteString(fmt.Sprintf("   Tags: %s\n", strings.Join(m.Tags, ", ")))
			}
		}
	}

	sb.WriteString("\nPlease review these memories and let me know if you'd like to search for specific topics or manage any memories.")

	return []PromptMessage{
		{
			Role: "user",
			Content: PromptContent{
				Type: "text",
				Text: sb.String(),
			},
		},
	}, nil
}

func (s *Server) promptHindsightReview(ctx context.Context, args map[string]string) ([]PromptMessage, error) {
	if s.noteService == nil {
		return []PromptMessage{
			{Role: "user", Content: PromptContent{Type: "text", Text: "Hindsight notes feature is not available."}},
		}, nil
	}

	notes, err := s.noteService.ListHindsightNotes(ctx, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to list hindsight notes: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("# Hindsight Review - Past Errors & Resolutions\n\n")

	if len(notes) == 0 {
		sb.WriteString("No hindsight notes recorded yet.\n")
	} else {
		for i, n := range notes {
			sb.WriteString(fmt.Sprintf("## %d. %s [%s]\n", i+1, n.Title, n.Severity))
			sb.WriteString(fmt.Sprintf("**Error**: %s\n", n.ErrorMessage))
			if n.Resolution != "" {
				sb.WriteString(fmt.Sprintf("**Resolution**: %s\n", n.Resolution))
			}
			if n.PreventionStrategy != "" {
				sb.WriteString(fmt.Sprintf("**Prevention**: %s\n", n.PreventionStrategy))
			}
			if n.LessonsLearned != "" {
				sb.WriteString(fmt.Sprintf("**Lessons**: %s\n", n.LessonsLearned))
			}
			sb.WriteString(fmt.Sprintf("Occurrences: %d | Prevention effectiveness: %.0f%%\n\n",
				n.OccurrenceCount, n.PreventionEffectiveness()*100))
		}
	}

	sb.WriteString("Review these past issues to avoid repeating similar mistakes.")

	return []PromptMessage{
		{
			Role: "user",
			Content: PromptContent{
				Type: "text",
				Text: sb.String(),
			},
		},
	}, nil
}

func (s *Server) promptCapturePattern(_ context.Context, args map[string]string) ([]PromptMessage, error) {
	pattern := args["pattern"]
	if pattern == "" {
		return nil, fmt.Errorf("pattern argument is required")
	}
	language := args["language"]

	var sb strings.Builder
	sb.WriteString("Please help me capture the following coding pattern as a Brain Sentry memory.\n\n")
	sb.WriteString(fmt.Sprintf("**Pattern:** %s\n", pattern))
	if language != "" {
		sb.WriteString(fmt.Sprintf("**Language:** %s\n", language))
	}
	sb.WriteString("\nPlease analyze this pattern and create a memory with:\n")
	sb.WriteString("1. A clear, concise summary\n")
	sb.WriteString("2. Category: PATTERN (or ANTIPATTERN if it's something to avoid)\n")
	sb.WriteString("3. Importance level (CRITICAL, IMPORTANT, or MINOR)\n")
	sb.WriteString("4. A code example if applicable\n")
	sb.WriteString("5. Relevant tags for easy discovery\n")
	sb.WriteString("\nUse the `create_memory` tool to store this pattern.")

	return []PromptMessage{
		{Role: "user", Content: PromptContent{Type: "text", Text: sb.String()}},
	}, nil
}

func (s *Server) promptExtractLearning(_ context.Context, args map[string]string) ([]PromptMessage, error) {
	session := args["session"]
	if session == "" {
		return nil, fmt.Errorf("session argument is required")
	}

	var sb strings.Builder
	sb.WriteString("Please extract learnings from the following session/conversation:\n\n")
	sb.WriteString(session)
	sb.WriteString("\n\nFor each learning, identify:\n")
	sb.WriteString("1. **Key decisions** made and their rationale\n")
	sb.WriteString("2. **Patterns** discovered (good practices to repeat)\n")
	sb.WriteString("3. **Anti-patterns** identified (mistakes to avoid)\n")
	sb.WriteString("4. **Domain knowledge** gained\n")
	sb.WriteString("5. **Action items** for follow-up\n")
	sb.WriteString("\nFor each significant learning, use `create_memory` to store it with appropriate category and importance.")

	return []PromptMessage{
		{Role: "user", Content: PromptContent{Type: "text", Text: sb.String()}},
	}, nil
}

func (s *Server) promptSummarizeDiscussion(_ context.Context, args map[string]string) ([]PromptMessage, error) {
	discussion := args["discussion"]
	if discussion == "" {
		return nil, fmt.Errorf("discussion argument is required")
	}

	var sb strings.Builder
	sb.WriteString("Please summarize the following technical discussion:\n\n")
	sb.WriteString(discussion)
	sb.WriteString("\n\nProvide a structured summary with:\n")
	sb.WriteString("1. **Topic**: What was discussed\n")
	sb.WriteString("2. **Key Decisions**: Decisions made with rationale\n")
	sb.WriteString("3. **Action Items**: Tasks assigned or identified\n")
	sb.WriteString("4. **Open Questions**: Unresolved items\n")
	sb.WriteString("5. **Technical Details**: Important technical points\n")
	sb.WriteString("\nStore key decisions and insights as memories using `create_memory`.")

	return []PromptMessage{
		{Role: "user", Content: PromptContent{Type: "text", Text: sb.String()}},
	}, nil
}
