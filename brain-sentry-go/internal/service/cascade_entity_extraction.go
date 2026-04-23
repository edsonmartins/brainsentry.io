package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/integraltech/brainsentry/pkg/pipeline"
)

// CascadeEntityExtractionService performs entity extraction in three sequential
// LLM passes for better precision vs the single-pass ExtractEntities:
//
//   1. Extract nodes (entities) — focused prompt, only names + types
//   2. Extract edge triplets — candidate (source, target) pairs from node list
//   3. Extract relationship names — label each edge with a canonical verb
//
// Smaller, focused prompts reduce hallucination at the cost of ~3x LLM calls.
type CascadeEntityExtractionService struct {
	llm       LLMProvider
	coref     *CoreferenceService // optional; runs before Pass 1 to normalise aliases
}

// NewCascadeEntityExtractionService creates a new CascadeEntityExtractionService.
func NewCascadeEntityExtractionService(llm LLMProvider) *CascadeEntityExtractionService {
	return &CascadeEntityExtractionService{llm: llm}
}

// WithCoreference attaches a CoreferenceService so the cascade rewrites the
// text (resolving pronouns/aliases) before Pass 1. Nil disables it.
func (s *CascadeEntityExtractionService) WithCoreference(c *CoreferenceService) *CascadeEntityExtractionService {
	s.coref = c
	return s
}

// CascadeExtractionResult holds the output of the 3-pass pipeline.
type CascadeExtractionResult struct {
	Entities      []ExtractedEntity       `json:"entities"`
	Relationships []ExtractedRelationship `json:"relationships"`
	PassCount     int                     `json:"passCount"`
}

// ---------- Pass 1: nodes ----------

const cascadeNodesPrompt = `You extract named entities from text. An entity is a concrete noun representing a distinct real-world or conceptual thing.

Rules:
- Only output entities explicitly named in the text. Do not infer or add.
- Use canonical form. "PostgreSQL" not "postgres". "New York City" not "NYC" (unless NYC is the only form used).
- Type must be one of: TECHNOLOGY, PERSON, PROJECT, CONCEPT, LIBRARY, LANGUAGE, TOOL, SERVICE, FILE, ORGANIZATION, LOCATION.
- Skip pronouns, articles, and generic nouns like "thing", "system", "tool" when unnamed.

Respond with valid JSON only:
{
  "entities": [
    {"name": "full canonical name", "type": "TECHNOLOGY"}
  ]
}`

type passNodesResponse struct {
	Entities []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"entities"`
}

// ---------- Pass 2: edges ----------

const cascadeEdgesPrompt = `You identify relationship candidates between entities based on a text.

Input: a text and a list of known entities.
Task: list pairs (source, target) of entities that are related to each other in the text. Do not invent new entities — only use the names from the provided list.

Respond with valid JSON only:
{
  "edges": [
    {"source": "Entity A name", "target": "Entity B name"}
  ]
}`

type passEdgesResponse struct {
	Edges []struct {
		Source string `json:"source"`
		Target string `json:"target"`
	} `json:"edges"`
}

// ---------- Pass 3: relationship names ----------

const cascadeRelationshipPrompt = `You name the relationship between two entities based on a text.

Input: a text and a pair (source, target) of related entities.
Task: provide a short canonical verb phrase describing the relationship. Use snake_case like "uses", "depends_on", "implements", "authored_by", "is_part_of", "caused_by", "related_to", "extends".

Respond with valid JSON only:
{
  "relationship": "verb_phrase"
}`

type passRelationshipResponse struct {
	Relationship string `json:"relationship"`
}

// Extract runs the 3-pass cascade as an explicit DAG using pkg/pipeline. Each
// pass is a named step; downstream steps depend on upstream outputs.
// Errors inside a single pass downgrade (return partial result) instead of
// aborting the whole extraction — we always prefer partial data to none.
func (s *CascadeEntityExtractionService) Extract(ctx context.Context, content string) (*CascadeExtractionResult, error) {
	if s.llm == nil {
		return &CascadeExtractionResult{}, nil
	}

	result := &CascadeExtractionResult{}

	// Pass 0 (optional, outside the DAG to keep content mutation explicit):
	// coreference resolution.
	if s.coref != nil {
		if cr, err := s.coref.Resolve(ctx, content); err == nil && cr != nil && cr.Resolved != "" {
			content = cr.Resolved
		} else if err != nil {
			slog.Warn("coreference resolution failed, using original text", "error", err)
		}
	}

	p, err := pipeline.NewBuilder().
		Add("nodes", func(ctx context.Context, in map[string]any) (any, error) {
			entities, err := s.extractNodes(ctx, content)
			if err != nil {
				return nil, fmt.Errorf("cascade pass 1 (nodes): %w", err)
			}
			return entities, nil
		}).
		Add("edges", func(ctx context.Context, in map[string]any) (any, error) {
			entities, _ := in["nodes"].([]ExtractedEntity)
			if len(entities) < 2 {
				return []edgeCandidate{}, nil
			}
			edges, err := s.extractEdges(ctx, content, entities)
			if err != nil {
				slog.Warn("cascade pass 2 failed, returning nodes only", "error", err)
				return []edgeCandidate{}, nil
			}
			return edges, nil
		}, "nodes").
		Add("relationships", func(ctx context.Context, in map[string]any) (any, error) {
			edges, _ := in["edges"].([]edgeCandidate)
			rels := make([]ExtractedRelationship, 0, len(edges))
			for _, e := range edges {
				relType, err := s.extractRelationshipName(ctx, content, e.Source, e.Target)
				if err != nil {
					slog.Warn("cascade pass 3 failed for edge",
						"source", e.Source, "target", e.Target, "error", err)
					continue
				}
				if relType == "" {
					continue
				}
				rels = append(rels, ExtractedRelationship{
					Source: e.Source,
					Target: e.Target,
					Type:   relType,
				})
			}
			return rels, nil
		}, "edges").
		Build()
	if err != nil {
		return nil, err
	}

	run, err := p.Run(ctx)
	if err != nil {
		return nil, err
	}

	if entities, ok := run.Outputs["nodes"].([]ExtractedEntity); ok {
		result.Entities = entities
		result.PassCount = 1
	}
	if edges, ok := run.Outputs["edges"].([]edgeCandidate); ok && len(edges) > 0 {
		result.PassCount = 2
	}
	if rels, ok := run.Outputs["relationships"].([]ExtractedRelationship); ok {
		result.Relationships = rels
		if len(rels) > 0 {
			result.PassCount = 3
		}
	}

	return result, nil
}

func (s *CascadeEntityExtractionService) extractNodes(ctx context.Context, content string) ([]ExtractedEntity, error) {
	userPrompt := fmt.Sprintf("Text:\n\n%s", truncateForLLM(content, 4000))

	response, err := s.llm.Chat(ctx, []ChatMessage{
		{Role: "system", Content: cascadeNodesPrompt},
		{Role: "user", Content: userPrompt},
	})
	if err != nil {
		return nil, err
	}

	var parsed passNodesResponse
	if err := json.Unmarshal([]byte(cleanJSON(response)), &parsed); err != nil {
		return nil, fmt.Errorf("parse nodes: %w", err)
	}

	entities := make([]ExtractedEntity, 0, len(parsed.Entities))
	seen := make(map[string]bool, len(parsed.Entities))
	for _, e := range parsed.Entities {
		name := strings.TrimSpace(e.Name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name) + "|" + strings.ToUpper(e.Type)
		if seen[key] {
			continue
		}
		seen[key] = true
		entities = append(entities, ExtractedEntity{
			Name: name,
			Type: strings.ToUpper(e.Type),
		})
	}
	return entities, nil
}

type edgeCandidate struct {
	Source string
	Target string
}

func (s *CascadeEntityExtractionService) extractEdges(ctx context.Context, content string, entities []ExtractedEntity) ([]edgeCandidate, error) {
	var entityList strings.Builder
	for i, e := range entities {
		if i > 0 {
			entityList.WriteString(", ")
		}
		entityList.WriteString(e.Name)
	}

	userPrompt := fmt.Sprintf("Entities: %s\n\nText:\n\n%s",
		entityList.String(),
		truncateForLLM(content, 3500),
	)

	response, err := s.llm.Chat(ctx, []ChatMessage{
		{Role: "system", Content: cascadeEdgesPrompt},
		{Role: "user", Content: userPrompt},
	})
	if err != nil {
		return nil, err
	}

	var parsed passEdgesResponse
	if err := json.Unmarshal([]byte(cleanJSON(response)), &parsed); err != nil {
		return nil, fmt.Errorf("parse edges: %w", err)
	}

	// Build lowercase set of valid entity names for filtering.
	valid := make(map[string]string, len(entities))
	for _, e := range entities {
		valid[strings.ToLower(e.Name)] = e.Name
	}

	out := make([]edgeCandidate, 0, len(parsed.Edges))
	seen := make(map[string]bool, len(parsed.Edges))
	for _, e := range parsed.Edges {
		src, okS := valid[strings.ToLower(strings.TrimSpace(e.Source))]
		tgt, okT := valid[strings.ToLower(strings.TrimSpace(e.Target))]
		if !okS || !okT || src == tgt {
			continue
		}
		key := src + "→" + tgt
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, edgeCandidate{Source: src, Target: tgt})
	}
	return out, nil
}

func (s *CascadeEntityExtractionService) extractRelationshipName(ctx context.Context, content, source, target string) (string, error) {
	userPrompt := fmt.Sprintf("Source: %s\nTarget: %s\n\nText:\n\n%s",
		source, target, truncateForLLM(content, 2000))

	response, err := s.llm.Chat(ctx, []ChatMessage{
		{Role: "system", Content: cascadeRelationshipPrompt},
		{Role: "user", Content: userPrompt},
	})
	if err != nil {
		return "", err
	}

	var parsed passRelationshipResponse
	if err := json.Unmarshal([]byte(cleanJSON(response)), &parsed); err != nil {
		return "", fmt.Errorf("parse relationship: %w", err)
	}

	name := strings.ToLower(strings.TrimSpace(parsed.Relationship))
	if name == "" {
		return "", nil
	}
	// Normalize: replace spaces with underscores, remove non-alphanumeric.
	name = strings.ReplaceAll(name, " ", "_")
	name = nonAlphanumeric.ReplaceAllString(name, "_")
	name = strings.Trim(name, "_")
	return name, nil
}
