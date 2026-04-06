package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
)

// ExpandedQuery holds the original query and its LLM-generated reformulations.
type ExpandedQuery struct {
	Original      string   `json:"original"`
	Reformulations []string `json:"reformulations"`
	Entities      []string `json:"entities"`
}

// QueryExpansionService generates query reformulations using LLM for better search recall.
type QueryExpansionService struct {
	llm LLMProvider
}

// NewQueryExpansionService creates a new QueryExpansionService.
func NewQueryExpansionService(llm LLMProvider) *QueryExpansionService {
	return &QueryExpansionService{llm: llm}
}

const queryExpansionPrompt = `Given a search query, generate reformulations to improve search recall.
Respond with valid JSON only.

Rules:
- Generate 3-5 alternative query formulations
- Include: synonyms, related terms, temporal concretizations, broader/narrower terms
- Extract key entities (people, technologies, files, concepts)
- Keep reformulations concise (under 20 words each)

Output:
{
  "reformulations": ["query1", "query2", "query3"],
  "entities": ["entity1", "entity2"]
}`

// Expand generates reformulations of the query for better search coverage.
func (s *QueryExpansionService) Expand(ctx context.Context, query string) *ExpandedQuery {
	result := &ExpandedQuery{
		Original: query,
	}

	if s.llm == nil {
		result.Reformulations = s.simpleFallback(query)
		return result
	}

	response, err := s.llm.Chat(ctx, []ChatMessage{
		{Role: "system", Content: queryExpansionPrompt},
		{Role: "user", Content: "Search query: " + query},
	})
	if err != nil {
		slog.Warn("query expansion LLM failed, using fallback", "error", err)
		result.Reformulations = s.simpleFallback(query)
		return result
	}

	var parsed struct {
		Reformulations []string `json:"reformulations"`
		Entities       []string `json:"entities"`
	}
	if err := json.Unmarshal([]byte(cleanJSON(response)), &parsed); err != nil {
		slog.Warn("query expansion parse failed, using fallback", "error", err)
		result.Reformulations = s.simpleFallback(query)
		return result
	}

	// Cap at 5 reformulations
	if len(parsed.Reformulations) > 5 {
		parsed.Reformulations = parsed.Reformulations[:5]
	}

	result.Reformulations = parsed.Reformulations
	result.Entities = parsed.Entities
	return result
}

// simpleFallback generates basic reformulations without LLM.
func (s *QueryExpansionService) simpleFallback(query string) []string {
	words := strings.Fields(query)
	if len(words) <= 2 {
		return nil
	}

	var reformulations []string

	// Remove stop words variation
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "are": true,
		"was": true, "were": true, "how": true, "what": true, "when": true,
		"where": true, "which": true, "that": true, "this": true, "to": true,
		"in": true, "on": true, "at": true, "for": true, "with": true,
		"do": true, "does": true, "did": true, "can": true, "could": true,
	}

	var contentWords []string
	for _, w := range words {
		if !stopWords[strings.ToLower(w)] {
			contentWords = append(contentWords, w)
		}
	}
	if len(contentWords) > 0 && len(contentWords) != len(words) {
		reformulations = append(reformulations, strings.Join(contentWords, " "))
	}

	return reformulations
}
