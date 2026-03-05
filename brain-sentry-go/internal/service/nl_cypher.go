package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	graphrepo "github.com/integraltech/brainsentry/internal/repository/graph"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// NLCypherService translates natural language queries to Cypher and executes them.
type NLCypherService struct {
	openRouter      *OpenRouterService
	graphClient     *graphrepo.Client
	maxRetries      int
}

// NewNLCypherService creates a new NLCypherService.
func NewNLCypherService(
	openRouter *OpenRouterService,
	graphClient *graphrepo.Client,
) *NLCypherService {
	return &NLCypherService{
		openRouter:  openRouter,
		graphClient: graphClient,
		maxRetries:  3,
	}
}

// NLQueryResult represents the result of a natural language graph query.
type NLQueryResult struct {
	Query          string         `json:"query"`
	GeneratedCypher string        `json:"generatedCypher"`
	Results        []map[string]any `json:"results"`
	Attempts       int            `json:"attempts"`
	Success        bool           `json:"success"`
	ErrorMessage   string         `json:"errorMessage,omitempty"`
}

// GraphSchema describes the graph structure for the LLM.
const graphSchema = `Graph Schema:
- Node labels: Memory, Entity
- Memory properties: id (string), content (string), summary (string), category (string), importance (string), tenantId (string), tags (list), accessCount (int), version (int), createdAt (int, unix millis)
- Entity properties: name (string), type (string), tenantId (string)
- Relationship types: RELATED_TO (properties: type, tag, strength, mentions, updatedAt), HAS_ENTITY, RELATED_ENTITY
- RELATED_TO connects Memory→Memory (shared tags/concepts)
- HAS_ENTITY connects Memory→Entity
- RELATED_ENTITY connects Entity→Entity`

// QueryNaturalLanguage translates a natural language question to Cypher,
// executes it, and retries with feedback if results are empty.
func (s *NLCypherService) QueryNaturalLanguage(ctx context.Context, question string) (*NLQueryResult, error) {
	if s.openRouter == nil || s.graphClient == nil {
		return &NLQueryResult{Query: question, Success: false, ErrorMessage: "NL→Cypher requires LLM and graph client"}, nil
	}

	tenantID := tenant.FromContext(ctx)
	result := &NLQueryResult{Query: question}

	var lastCypher string
	var lastError string

	for attempt := 1; attempt <= s.maxRetries; attempt++ {
		result.Attempts = attempt

		// Generate Cypher
		cypher, err := s.generateCypher(ctx, question, tenantID, lastCypher, lastError)
		if err != nil {
			slog.Warn("failed to generate Cypher", "error", err, "attempt", attempt)
			lastError = err.Error()
			continue
		}

		result.GeneratedCypher = cypher
		lastCypher = cypher

		// Execute Cypher
		queryResult, err := s.graphClient.Query(ctx, cypher)
		if err != nil {
			slog.Warn("Cypher execution failed", "error", err, "attempt", attempt)
			lastError = fmt.Sprintf("execution error: %s", err.Error())
			continue
		}

		// Check if we got results
		if len(queryResult.Records) == 0 {
			lastError = "query returned 0 results"
			slog.Debug("NL→Cypher empty result, retrying", "attempt", attempt, "cypher", cypher)
			continue
		}

		// Convert records to map format
		result.Results = make([]map[string]any, 0, len(queryResult.Records))
		for _, rec := range queryResult.Records {
			result.Results = append(result.Results, rec.Values)
		}
		result.Success = true
		return result, nil
	}

	result.Success = false
	result.ErrorMessage = fmt.Sprintf("failed after %d attempts: %s", s.maxRetries, lastError)
	return result, nil
}

func (s *NLCypherService) generateCypher(ctx context.Context, question, tenantID, previousCypher, previousError string) (string, error) {
	var feedbackSection string
	if previousCypher != "" && previousError != "" {
		feedbackSection = fmt.Sprintf(`

Previous attempt that failed:
Cypher: %s
Error/Issue: %s

Generate a DIFFERENT and CORRECTED query.`, previousCypher, previousError)
	}

	prompt := fmt.Sprintf(`Convert the following natural language question into a Cypher query for FalkorDB.

%s

IMPORTANT RULES:
- Always filter by tenantId = '%s'
- Use MATCH patterns, not CALL procedures
- Return meaningful columns with aliases
- LIMIT results to 20
- Use case-insensitive matching with toLower() for text comparisons
- Do NOT use vector search procedures%s

Respond in JSON format only:
{"cypher": "the generated Cypher query"}

Question: %s`, graphSchema, graphrepo.EscapeCypher(tenantID), feedbackSection, question)

	response, err := s.openRouter.Chat(ctx, []ChatMessage{
		{Role: "system", Content: "You are a Cypher query generator for FalkorDB. Generate valid, efficient Cypher queries. Respond with valid JSON only."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return "", err
	}

	var result struct {
		Cypher string `json:"cypher"`
	}
	if err := json.Unmarshal([]byte(cleanJSON(response)), &result); err != nil {
		// Try to extract Cypher directly from response
		trimmed := strings.TrimSpace(response)
		if strings.HasPrefix(strings.ToUpper(trimmed), "MATCH") || strings.HasPrefix(strings.ToUpper(trimmed), "OPTIONAL") {
			return trimmed, nil
		}
		return "", fmt.Errorf("parsing Cypher response: %w", err)
	}

	return result.Cypher, nil
}
