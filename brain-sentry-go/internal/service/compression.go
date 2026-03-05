package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/dto"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// CompressionService handles context compression and summarization.
type CompressionService struct {
	summaryRepo *postgres.ContextSummaryRepository
	openRouter  *OpenRouterService
}

// NewCompressionService creates a new CompressionService.
func NewCompressionService(summaryRepo *postgres.ContextSummaryRepository, openRouter *OpenRouterService) *CompressionService {
	return &CompressionService{
		summaryRepo: summaryRepo,
		openRouter:  openRouter,
	}
}

// CompressionResult represents the result of context compression.
type CompressionResult struct {
	Summary              string   `json:"summary"`
	Goals                []string `json:"goals"`
	Decisions            []string `json:"decisions"`
	Errors               []string `json:"errors"`
	Todos                []string `json:"todos"`
	OriginalTokenCount   int      `json:"originalTokenCount"`
	CompressedTokenCount int      `json:"compressedTokenCount"`
	CompressionRatio     float64  `json:"compressionRatio"`
	PreservedMessages    []dto.CompressionMessage `json:"preservedMessages,omitempty"`
}

// Compress compresses a conversation context.
func (s *CompressionService) Compress(ctx context.Context, req dto.CompressionRequest, sessionID string) (*CompressionResult, error) {
	if s.openRouter == nil {
		return nil, fmt.Errorf("LLM service not available")
	}

	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("no messages to compress")
	}

	// Calculate original token count
	originalTokens := 0
	for _, m := range req.Messages {
		originalTokens += estimateTokens(m.Content)
	}

	// Determine how many recent messages to preserve
	preserveRecent := req.PreserveRecent
	if preserveRecent <= 0 {
		preserveRecent = 3
	}
	if preserveRecent > len(req.Messages) {
		preserveRecent = len(req.Messages)
	}

	// Split messages: to compress vs to preserve
	toCompress := req.Messages[:len(req.Messages)-preserveRecent]
	preserved := req.Messages[len(req.Messages)-preserveRecent:]

	if len(toCompress) == 0 {
		return &CompressionResult{
			Summary:              "No messages to compress",
			OriginalTokenCount:   originalTokens,
			CompressedTokenCount: originalTokens,
			CompressionRatio:     1.0,
			PreservedMessages:    preserved,
		}, nil
	}

	// Build conversation text for compression
	conversationText := formatMessagesForCompression(toCompress)

	// Build keyword preservation instruction
	keywordInst := ""
	if len(req.PreserveKeywords) > 0 {
		keywordInst = fmt.Sprintf("\nIMPORTANT: Preserve mentions of these keywords: %s", strings.Join(req.PreserveKeywords, ", "))
	}

	contextHint := ""
	if req.ContextHint != "" {
		contextHint = fmt.Sprintf("\nContext hint: %s", req.ContextHint)
	}

	targetRatio := req.TargetRatio
	if targetRatio <= 0 || targetRatio >= 1 {
		targetRatio = 0.3
	}

	prompt := fmt.Sprintf(`Compress the following conversation into a structured summary. Target approximately %.0f%% of original size.
Extract and categorize:
1. A comprehensive summary of the conversation
2. Active goals or objectives
3. Key decisions made
4. Errors or issues encountered
5. Pending tasks or TODOs
%s%s

Respond in JSON format only:
{
  "summary": "comprehensive summary",
  "goals": ["goal1", "goal2"],
  "decisions": ["decision1"],
  "errors": ["error1"],
  "todos": ["todo1"]
}

Conversation:
%s`, targetRatio*100, keywordInst, contextHint, truncate(conversationText, 6000))

	response, err := s.openRouter.Chat(ctx, []ChatMessage{
		{Role: "system", Content: "You are a conversation compressor. Create structured summaries that preserve essential context. Respond with valid JSON only."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, fmt.Errorf("compression LLM call failed: %w", err)
	}

	var parsed struct {
		Summary   string   `json:"summary"`
		Goals     []string `json:"goals"`
		Decisions []string `json:"decisions"`
		Errors    []string `json:"errors"`
		Todos     []string `json:"todos"`
	}
	if err := json.Unmarshal([]byte(cleanJSON(response)), &parsed); err != nil {
		slog.Warn("failed to parse compression result, using raw response", "error", err)
		parsed.Summary = response
	}

	compressedTokens := estimateTokens(parsed.Summary)
	for _, p := range preserved {
		compressedTokens += estimateTokens(p.Content)
	}

	ratio := 0.0
	if originalTokens > 0 {
		ratio = float64(compressedTokens) / float64(originalTokens)
	}

	result := &CompressionResult{
		Summary:              parsed.Summary,
		Goals:                parsed.Goals,
		Decisions:            parsed.Decisions,
		Errors:               parsed.Errors,
		Todos:                parsed.Todos,
		OriginalTokenCount:   originalTokens,
		CompressedTokenCount: compressedTokens,
		CompressionRatio:     ratio,
		PreservedMessages:    preserved,
	}

	// Persist the summary
	if sessionID != "" {
		tenantID := tenant.FromContext(ctx)
		cs := &domain.ContextSummary{
			TenantID:             tenantID,
			SessionID:            sessionID,
			OriginalTokenCount:   originalTokens,
			CompressedTokenCount: compressedTokens,
			CompressionRatio:     ratio,
			Summary:              parsed.Summary,
			Goals:                parsed.Goals,
			Decisions:            parsed.Decisions,
			Errors:               parsed.Errors,
			Todos:                parsed.Todos,
			RecentWindowSize:     preserveRecent,
			ModelUsed:            "openrouter",
			CompressionMethod:    "llm_structured",
		}
		if err := s.summaryRepo.Create(ctx, cs); err != nil {
			slog.Warn("failed to persist context summary", "error", err)
		}
	}

	return result, nil
}

// GetSessionSummaries returns compression summaries for a session.
func (s *CompressionService) GetSessionSummaries(ctx context.Context, sessionID string) ([]domain.ContextSummary, error) {
	return s.summaryRepo.FindBySession(ctx, sessionID)
}

// GetLatestSummary returns the most recent compression summary for a session.
func (s *CompressionService) GetLatestSummary(ctx context.Context, sessionID string) (*domain.ContextSummary, error) {
	return s.summaryRepo.GetLatestBySession(ctx, sessionID)
}

func formatMessagesForCompression(messages []dto.CompressionMessage) string {
	var sb strings.Builder
	for _, m := range messages {
		role := m.Role
		if m.ToolName != "" {
			role = fmt.Sprintf("tool:%s", m.ToolName)
		}
		sb.WriteString(fmt.Sprintf("[%s]: %s\n\n", role, m.Content))
	}
	return sb.String()
}
