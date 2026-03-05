package graph

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

// GraphRAGRepository provides advanced graph-based retrieval augmented generation.
type GraphRAGRepository struct {
	client *Client
}

// NewGraphRAGRepository creates a new GraphRAGRepository.
func NewGraphRAGRepository(client *Client) *GraphRAGRepository {
	return &GraphRAGRepository{client: client}
}

// EnsureVectorIndex creates a vector index on Memory nodes if it doesn't exist.
func (r *GraphRAGRepository) EnsureVectorIndex(ctx context.Context, dimensions int) error {
	cypher := fmt.Sprintf(
		`CREATE VECTOR INDEX FOR (m:Memory) ON (m.embedding) OPTIONS {dimension: %d, similarityFunction: 'cosine'}`,
		dimensions,
	)

	_, err := r.client.Query(ctx, cypher)
	if err != nil {
		// Index might already exist — not fatal
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "ERR") {
			slog.Info("vector index already exists or not supported", "error", err)
			return nil
		}
		return fmt.Errorf("creating vector index: %w", err)
	}

	slog.Info("vector index created", "dimensions", dimensions)
	return nil
}

// MultiHopResult represents a multi-hop reasoning result.
type MultiHopResult struct {
	MemoryID    string   `json:"memoryId"`
	Summary     string   `json:"summary"`
	Category    string   `json:"category"`
	Importance  string   `json:"importance"`
	HopDistance int      `json:"hopDistance"`
	Path        []string `json:"path"`
	Score       float64  `json:"score"`
}

// MultiHopSearch performs multi-hop graph traversal from seed memories.
func (r *GraphRAGRepository) MultiHopSearch(ctx context.Context, seedIDs []string, maxHops, limit int, tenantID string) ([]MultiHopResult, error) {
	if len(seedIDs) == 0 {
		return nil, nil
	}
	if maxHops <= 0 {
		maxHops = 3
	}
	if limit <= 0 {
		limit = 20
	}

	// Build seed ID list for Cypher
	seedList := make([]string, len(seedIDs))
	for i, id := range seedIDs {
		seedList[i] = "'" + EscapeCypher(id) + "'"
	}

	cypher := fmt.Sprintf(`MATCH path = (seed:Memory)-[r:RELATED_TO*1..%d]-(target:Memory)
WHERE seed.id IN [%s] AND target.tenantId = '%s' AND NOT target.id IN [%s]
WITH target, min(length(path)) AS hopDistance,
     [n IN nodes(path) | n.id] AS pathNodes,
     sum(reduce(s = 0.0, rel IN relationships(path) | s + rel.strength)) AS totalStrength
RETURN target.id AS memoryId, target.summary AS summary,
       target.category AS category, target.importance AS importance,
       hopDistance, pathNodes AS path, totalStrength AS score
ORDER BY hopDistance ASC, score DESC
LIMIT %d`,
		maxHops,
		strings.Join(seedList, ", "),
		EscapeCypher(tenantID),
		strings.Join(seedList, ", "),
		limit,
	)

	result, err := r.client.Query(ctx, cypher)
	if err != nil {
		return nil, fmt.Errorf("multi-hop search: %w", err)
	}

	results := make([]MultiHopResult, 0, len(result.Records))
	for _, rec := range result.Records {
		r := MultiHopResult{
			MemoryID:    GetString(rec.Values, "memoryId"),
			Summary:     GetString(rec.Values, "summary"),
			Category:    GetString(rec.Values, "category"),
			Importance:  GetString(rec.Values, "importance"),
			HopDistance: int(GetInt64(rec.Values, "hopDistance")),
			Score:       GetFloat64(rec.Values, "score"),
		}

		// Parse path
		if pathVal, ok := rec.Values["path"]; ok {
			if pathArr, ok := pathVal.([]any); ok {
				for _, p := range pathArr {
					if ps, ok := p.(string); ok {
						r.Path = append(r.Path, ps)
					}
				}
			}
		}

		results = append(results, r)
	}

	return results, nil
}

// EnrichContext builds a rich context by combining vector search results with multi-hop graph traversal.
func (r *GraphRAGRepository) EnrichContext(ctx context.Context, seedIDs []string, tenantID string) ([]MultiHopResult, error) {
	// First hop: direct relationships
	directResults, err := r.MultiHopSearch(ctx, seedIDs, 1, 10, tenantID)
	if err != nil {
		slog.Warn("direct hop search failed", "error", err)
		directResults = nil
	}

	// Second hop: extended network
	extendedResults, err := r.MultiHopSearch(ctx, seedIDs, 3, 10, tenantID)
	if err != nil {
		slog.Warn("extended hop search failed", "error", err)
		extendedResults = nil
	}

	// Merge and deduplicate
	seen := make(map[string]bool)
	for _, id := range seedIDs {
		seen[id] = true
	}

	var combined []MultiHopResult
	for _, r := range directResults {
		if !seen[r.MemoryID] {
			seen[r.MemoryID] = true
			combined = append(combined, r)
		}
	}
	for _, r := range extendedResults {
		if !seen[r.MemoryID] {
			seen[r.MemoryID] = true
			combined = append(combined, r)
		}
	}

	return combined, nil
}

// GetCluster returns a cluster of closely related memories around a seed.
func (r *GraphRAGRepository) GetCluster(ctx context.Context, memoryID string, tenantID string, maxSize int) ([]string, error) {
	if maxSize <= 0 {
		maxSize = 10
	}

	cypher := fmt.Sprintf(`MATCH (seed:Memory {id: '%s'})-[r:RELATED_TO*1..2]-(related:Memory)
WHERE related.tenantId = '%s'
WITH related, count(r) AS connections
ORDER BY connections DESC
LIMIT %d
RETURN related.id AS id`,
		EscapeCypher(memoryID),
		EscapeCypher(tenantID),
		maxSize,
	)

	result, err := r.client.Query(ctx, cypher)
	if err != nil {
		return nil, fmt.Errorf("getting cluster: %w", err)
	}

	ids := make([]string, 0, len(result.Records))
	for _, rec := range result.Records {
		ids = append(ids, GetString(rec.Values, "id"))
	}

	return ids, nil
}
