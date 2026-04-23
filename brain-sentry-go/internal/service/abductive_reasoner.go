package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
)

// AbductiveReasoner answers "given outcome Y, what likely causes explain it?".
// It walks a decision's ancestors, pulls memories cited as evidence, and asks
// the LLM to rank plausible causal hypotheses with confidence scores.
type AbductiveReasoner struct {
	llm          LLMProvider
	decisionRepo *postgres.DecisionRepository
	memoryRepo   *postgres.MemoryRepository
	decisionSvc  *DecisionService
}

// NewAbductiveReasoner constructs the reasoner.
func NewAbductiveReasoner(llm LLMProvider, decisionRepo *postgres.DecisionRepository, memoryRepo *postgres.MemoryRepository, decisionSvc *DecisionService) *AbductiveReasoner {
	return &AbductiveReasoner{
		llm:          llm,
		decisionRepo: decisionRepo,
		memoryRepo:   memoryRepo,
		decisionSvc:  decisionSvc,
	}
}

// Hypothesis is a single candidate explanation with its confidence.
type Hypothesis struct {
	Cause       string   `json:"cause"`
	Confidence  float64  `json:"confidence"`
	Evidence    []string `json:"evidence,omitempty"`
	EntityIDs   []string `json:"entityIds,omitempty"`
	MemoryIDs   []string `json:"memoryIds,omitempty"`
}

// AbduceRequest is the input for Abduce.
type AbduceRequest struct {
	DecisionID string `json:"decisionId"`
	Question   string `json:"question,omitempty"`
	MaxHypotheses int `json:"maxHypotheses,omitempty"`
}

// AbduceResult is the full reasoning report.
type AbduceResult struct {
	Decision     *domain.Decision `json:"decision"`
	Question     string           `json:"question"`
	Hypotheses   []Hypothesis     `json:"hypotheses"`
	EvidenceUsed int              `json:"evidenceUsed"`
	Model        string           `json:"model,omitempty"`
}

// Abduce runs the reasoner for a target decision.
func (r *AbductiveReasoner) Abduce(ctx context.Context, req AbduceRequest) (*AbduceResult, error) {
	if r == nil || r.llm == nil {
		return nil, fmt.Errorf("LLM provider not configured")
	}
	if req.DecisionID == "" {
		return nil, fmt.Errorf("decisionId is required")
	}
	maxH := req.MaxHypotheses
	if maxH <= 0 {
		maxH = 5
	}

	target, err := r.decisionRepo.FindByID(ctx, req.DecisionID)
	if err != nil {
		return nil, err
	}

	// Gather ancestors as context.
	chain, err := r.decisionSvc.CausalChain(ctx, target.ID, 5)
	if err != nil {
		chain = nil
	}

	// Gather cited memories.
	evidence := []string{}
	for _, mid := range target.MemoryIDs {
		m, err := r.memoryRepo.FindByID(ctx, mid)
		if err == nil {
			summary := m.Summary
			if summary == "" {
				content := m.Content
				if len(content) > 400 {
					content = content[:400] + "…"
				}
				summary = content
			}
			evidence = append(evidence, fmt.Sprintf("[%s] %s", m.ID, summary))
		}
	}

	q := req.Question
	if q == "" {
		q = fmt.Sprintf("Why was outcome %q reached in scenario %q?", target.Outcome, target.Scenario)
	}

	var ancestry strings.Builder
	for _, node := range chain {
		if node.Relation != "ancestor" {
			continue
		}
		ancestry.WriteString(fmt.Sprintf("- (%s) %s → outcome=%s, confidence=%.2f\n",
			node.Decision.Category, node.Decision.Scenario, node.Decision.Outcome, node.Decision.Confidence))
	}

	prompt := fmt.Sprintf(`You are performing ABDUCTIVE reasoning — given an observed outcome,
propose the most plausible causes and rank them.

Decision under analysis:
  id: %s
  category: %s
  scenario: %s
  reasoning: %s
  outcome: %s
  confidence: %.2f

Ancestor decisions (from oldest to most recent):
%s

Evidence (memories cited by the decision):
%s

Question: %s

Return ONLY a JSON array with at most %d objects, each with:
  - cause (short hypothesis)
  - confidence (0..1 float)
  - evidence (array of short phrases)
  - memoryIds (strings, subset of cited evidence)
  - entityIds (strings, from ancestry if relevant)

Rank from highest confidence to lowest. No commentary.`,
		target.ID, target.Category, target.Scenario, target.Reasoning, target.Outcome, target.Confidence,
		orPlaceholder(ancestry.String(), "(none)"),
		orPlaceholder(strings.Join(evidence, "\n"), "(none)"),
		q, maxH)

	raw, err := r.llm.Chat(ctx, []ChatMessage{
		{Role: "system", Content: "You perform abductive reasoning over structured traces. Output JSON only."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, err
	}

	raw = strings.TrimSpace(raw)
	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start < 0 || end <= start {
		return nil, fmt.Errorf("LLM did not return a JSON array")
	}
	var payload []Hypothesis
	if err := json.Unmarshal([]byte(raw[start:end+1]), &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON from LLM: %w", err)
	}
	if len(payload) > maxH {
		payload = payload[:maxH]
	}

	return &AbduceResult{
		Decision:     target,
		Question:     q,
		Hypotheses:   payload,
		EvidenceUsed: len(evidence),
		Model:        r.llm.Name(),
	}, nil
}

func orPlaceholder(s, placeholder string) string {
	if strings.TrimSpace(s) == "" {
		return placeholder
	}
	return s
}
