package mcp

import (
	"context"
	"encoding/json"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/dto"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

func (s *Server) registerTools() { //nolint:funlen
	// Interception tool
	if s.interceptionService != nil {
		s.tools["intercept_prompt"] = Tool{
			Name:        "intercept_prompt",
			Description: "Intercept a prompt and enhance it with relevant context from memories and notes. Returns the enhanced prompt with injected context.",
			InputSchema: mustJSON(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"prompt":            map[string]string{"type": "string", "description": "The prompt to intercept and enhance"},
					"tenantId":          map[string]string{"type": "string", "description": "Tenant identifier"},
					"sessionId":         map[string]string{"type": "string", "description": "Session identifier"},
					"userId":            map[string]string{"type": "string", "description": "User identifier"},
					"maxTokens":         map[string]any{"type": "integer", "description": "Maximum tokens to inject", "default": 500},
					"forceDeepAnalysis": map[string]any{"type": "boolean", "description": "Force deep LLM analysis", "default": false},
					"context":           map[string]string{"type": "string", "description": "Additional context"},
				},
				"required": []string{"prompt"},
			}),
			Handler: s.toolInterceptPrompt,
		}
	}

	s.tools["create_memory"] = Tool{
		Name:        "create_memory",
		Description: "Create a new memory in Brain Sentry. Use this to store important insights, decisions, patterns, or knowledge.",
		InputSchema: mustJSON(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"content":    map[string]string{"type": "string", "description": "The content of the memory"},
				"summary":    map[string]string{"type": "string", "description": "Brief summary (optional)"},
				"category":   map[string]string{"type": "string", "description": "Category: DECISION, PATTERN, ANTIPATTERN, DOMAIN, BUG, OPTIMIZATION, INTEGRATION, INSIGHT, WARNING, KNOWLEDGE, ACTION, CONTEXT, REFERENCE"},
				"importance": map[string]string{"type": "string", "description": "Importance: CRITICAL, IMPORTANT, MINOR"},
				"tags":       map[string]any{"type": "array", "items": map[string]string{"type": "string"}, "description": "Tags for categorization"},
			},
			"required": []string{"content"},
		}),
		Handler: s.toolCreateMemory,
	}

	s.tools["search_memories"] = Tool{
		Name:        "search_memories",
		Description: "Search memories in Brain Sentry by query text. Returns relevant memories that match the query.",
		InputSchema: mustJSON(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]string{"type": "string", "description": "Search query text"},
				"limit": map[string]any{"type": "integer", "description": "Maximum results to return", "default": 10},
			},
			"required": []string{"query"},
		}),
		Handler: s.toolSearchMemories,
	}

	s.tools["get_memory"] = Tool{
		Name:        "get_memory",
		Description: "Get a specific memory by ID.",
		InputSchema: mustJSON(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]string{"type": "string", "description": "Memory ID"},
			},
			"required": []string{"id"},
		}),
		Handler: s.toolGetMemory,
	}

	s.tools["list_memories"] = Tool{
		Name:        "list_memories",
		Description: "List all memories, optionally filtered by category or importance.",
		InputSchema: mustJSON(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"category":   map[string]string{"type": "string", "description": "Filter by category"},
				"importance": map[string]string{"type": "string", "description": "Filter by importance"},
				"page":       map[string]any{"type": "integer", "description": "Page number", "default": 0},
				"size":       map[string]any{"type": "integer", "description": "Page size", "default": 20},
			},
		}),
		Handler: s.toolListMemories,
	}

	s.tools["update_memory"] = Tool{
		Name:        "update_memory",
		Description: "Update an existing memory's content, category, or importance.",
		InputSchema: mustJSON(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id":         map[string]string{"type": "string", "description": "Memory ID to update"},
				"content":    map[string]string{"type": "string", "description": "New content"},
				"summary":    map[string]string{"type": "string", "description": "New summary"},
				"category":   map[string]string{"type": "string", "description": "New category"},
				"importance": map[string]string{"type": "string", "description": "New importance"},
			},
			"required": []string{"id"},
		}),
		Handler: s.toolUpdateMemory,
	}

	s.tools["delete_memory"] = Tool{
		Name:        "delete_memory",
		Description: "Delete a memory by ID.",
		InputSchema: mustJSON(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]string{"type": "string", "description": "Memory ID to delete"},
			},
			"required": []string{"id"},
		}),
		Handler: s.toolDeleteMemory,
	}
}

// toolSuccessResponse wraps a tool result with success=true and tenantId per MCP spec.
func toolSuccessResponse(ctx context.Context, data map[string]any) map[string]any {
	data["success"] = true
	data["tenantId"] = tenant.FromContext(ctx)
	return data
}

// toolErrorMap returns a structured MCP error response per MCP_SERVER_API.md spec.
func toolErrorMap(category, errorCode, message string) map[string]any {
	return map[string]any{
		"success":       false,
		"error":         message,
		"errorCode":     errorCode,
		"errorCategory": category,
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
	}
}

func (s *Server) toolCreateMemory(ctx context.Context, params json.RawMessage) (any, error) {
	var args struct {
		Content    string   `json:"content"`
		Summary    string   `json:"summary"`
		Category   string   `json:"category"`
		Importance string   `json:"importance"`
		Tags       []string `json:"tags"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return toolErrorMap(ErrCategoryValidation, "validation", "invalid parameters: "+err.Error()), nil
	}

	if args.Content == "" {
		return toolErrorMap(ErrCategoryValidation, "validation", "content is required"), nil
	}

	req := dto.CreateMemoryRequest{
		Content:    args.Content,
		Summary:    args.Summary,
		Category:   domain.MemoryCategory(args.Category),
		Importance: domain.ImportanceLevel(args.Importance),
		Tags:       args.Tags,
		SourceType: "mcp",
	}

	memory, err := s.memoryService.CreateMemory(ctx, req)
	if err != nil {
		return toolErrorMap(ErrCategoryInternal, "internal", "failed to create memory: "+err.Error()), nil
	}

	return toolSuccessResponse(ctx, map[string]any{
		"memoryId":   memory.ID,
		"message":    "Memory created successfully",
		"id":         memory.ID,
		"content":    memory.Content,
		"summary":    memory.Summary,
		"category":   memory.Category,
		"importance": memory.Importance,
		"createdAt":  memory.CreatedAt,
	}), nil
}

func (s *Server) toolSearchMemories(ctx context.Context, params json.RawMessage) (any, error) {
	var args struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return toolErrorMap(ErrCategoryValidation, "validation", "invalid parameters: "+err.Error()), nil
	}

	if args.Query == "" {
		return toolErrorMap(ErrCategoryValidation, "validation", "query is required"), nil
	}

	if args.Limit <= 0 {
		args.Limit = 10
	}

	searchResp, err := s.memoryService.SearchMemories(ctx, dto.SearchRequest{Query: args.Query, Limit: args.Limit})
	if err != nil {
		return toolErrorMap(ErrCategoryInternal, "internal", "search failed: "+err.Error()), nil
	}

	results := make([]map[string]any, 0, len(searchResp.Results))
	for _, m := range searchResp.Results {
		results = append(results, map[string]any{
			"id":             m.ID,
			"content":        m.Content,
			"summary":        m.Summary,
			"category":       m.Category,
			"importance":     m.Importance,
			"tags":           m.Tags,
			"relevanceScore": m.RelevanceScore,
		})
	}

	return toolSuccessResponse(ctx, map[string]any{
		"count":        len(results),
		"memories":     results,
		"searchTimeMs": searchResp.SearchTimeMs,
	}), nil
}

func (s *Server) toolGetMemory(ctx context.Context, params json.RawMessage) (any, error) {
	var args struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return toolErrorMap(ErrCategoryValidation, "validation", "invalid parameters: "+err.Error()), nil
	}

	if args.ID == "" {
		return toolErrorMap(ErrCategoryValidation, "validation", "id is required"), nil
	}

	memory, err := s.memoryService.GetMemory(ctx, args.ID)
	if err != nil {
		return toolErrorMap(ErrCategoryNotFound, "not_found", "memory not found: "+err.Error()), nil
	}

	return toolSuccessResponse(ctx, map[string]any{
		"memory": memory,
	}), nil
}

func (s *Server) toolListMemories(ctx context.Context, params json.RawMessage) (any, error) {
	var args struct {
		Category   string `json:"category"`
		Importance string `json:"importance"`
		Page       int    `json:"page"`
		Size       int    `json:"size"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return toolErrorMap(ErrCategoryValidation, "validation", "invalid parameters: "+err.Error()), nil
	}

	if args.Size <= 0 {
		args.Size = 20
	}

	if args.Category != "" {
		memories, err := s.memoryService.GetByCategory(ctx, domain.MemoryCategory(args.Category))
		if err != nil {
			return toolErrorMap(ErrCategoryInternal, "internal", err.Error()), nil
		}
		return toolSuccessResponse(ctx, map[string]any{"memories": memories, "total": len(memories)}), nil
	}

	if args.Importance != "" {
		memories, err := s.memoryService.GetByImportance(ctx, domain.ImportanceLevel(args.Importance))
		if err != nil {
			return toolErrorMap(ErrCategoryInternal, "internal", err.Error()), nil
		}
		return toolSuccessResponse(ctx, map[string]any{"memories": memories, "total": len(memories)}), nil
	}

	resp, err := s.memoryService.ListMemories(ctx, args.Page, args.Size)
	if err != nil {
		return toolErrorMap(ErrCategoryInternal, "internal", err.Error()), nil
	}

	return toolSuccessResponse(ctx, map[string]any{
		"memories": resp.Memories,
		"total":    resp.TotalElements,
		"page":     resp.Page,
		"pageSize": resp.Size,
	}), nil
}

func (s *Server) toolUpdateMemory(ctx context.Context, params json.RawMessage) (any, error) {
	var args struct {
		ID         string `json:"id"`
		Content    string `json:"content"`
		Summary    string `json:"summary"`
		Category   string `json:"category"`
		Importance string `json:"importance"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return toolErrorMap(ErrCategoryValidation, "validation", "invalid parameters: "+err.Error()), nil
	}

	if args.ID == "" {
		return toolErrorMap(ErrCategoryValidation, "validation", "id is required"), nil
	}

	req := dto.UpdateMemoryRequest{
		Content:    args.Content,
		Summary:    args.Summary,
		Category:   domain.MemoryCategory(args.Category),
		Importance: domain.ImportanceLevel(args.Importance),
	}

	memory, err := s.memoryService.UpdateMemory(ctx, args.ID, req)
	if err != nil {
		return toolErrorMap(ErrCategoryInternal, "internal", "update failed: "+err.Error()), nil
	}

	return toolSuccessResponse(ctx, map[string]any{
		"memory":  memory,
		"message": "Memory updated successfully",
	}), nil
}

func (s *Server) toolDeleteMemory(ctx context.Context, params json.RawMessage) (any, error) {
	var args struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return toolErrorMap(ErrCategoryValidation, "validation", "invalid parameters: "+err.Error()), nil
	}

	if args.ID == "" {
		return toolErrorMap(ErrCategoryValidation, "validation", "id is required"), nil
	}

	if err := s.memoryService.DeleteMemory(ctx, args.ID); err != nil {
		return toolErrorMap(ErrCategoryInternal, "internal", "delete failed: "+err.Error()), nil
	}

	return toolSuccessResponse(ctx, map[string]any{
		"message": "memory deleted",
		"id":      args.ID,
	}), nil
}

func (s *Server) toolInterceptPrompt(ctx context.Context, params json.RawMessage) (any, error) {
	var args struct {
		Prompt            string `json:"prompt"`
		TenantID          string `json:"tenantId"`
		SessionID         string `json:"sessionId"`
		UserID            string `json:"userId"`
		MaxTokens         int    `json:"maxTokens"`
		ForceDeepAnalysis bool   `json:"forceDeepAnalysis"`
		Context           string `json:"context"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return toolErrorMap(ErrCategoryValidation, "validation", "invalid parameters: "+err.Error()), nil
	}

	if args.Prompt == "" {
		return toolErrorMap(ErrCategoryValidation, "validation", "prompt is required"), nil
	}

	// If tenantId provided, set it in context
	if args.TenantID != "" {
		ctx = tenant.WithTenant(ctx, args.TenantID)
	}

	var ctxMap map[string]any
	if args.Context != "" {
		ctxMap = map[string]any{"additional": args.Context}
	}

	req := dto.InterceptRequest{
		Prompt:            args.Prompt,
		SessionID:         args.SessionID,
		UserID:            args.UserID,
		MaxTokens:         args.MaxTokens,
		ForceDeepAnalysis: args.ForceDeepAnalysis,
		Context:           ctxMap,
	}

	resp, err := s.interceptionService.Intercept(ctx, req)
	if err != nil {
		return toolErrorMap(ErrCategoryInternal, "internal", "interception failed: "+err.Error()), nil
	}

	result := map[string]any{
		"success":  true,
		"tenantId": tenant.FromContext(ctx),
	}
	// Merge intercept response fields
	respBytes, _ := json.Marshal(resp)
	var respMap map[string]any
	if json.Unmarshal(respBytes, &respMap) == nil {
		for k, v := range respMap {
			result[k] = v
		}
	}

	return result, nil
}

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
