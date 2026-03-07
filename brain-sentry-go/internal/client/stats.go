package client

import (
	"fmt"

	"github.com/integraltech/brainsentry/internal/dto"
)

// GetStats returns system statistics.
func (c *Client) GetStats() (*dto.StatsResponse, error) {
	var resp dto.StatsResponse
	if err := c.Get("/v1/stats/overview", &resp); err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}
	return &resp, nil
}

// GetKnowledgeGraph returns the knowledge graph visualization data.
func (c *Client) GetKnowledgeGraph(limit int) (*dto.KnowledgeGraphResponse, error) {
	path := "/v1/entity-graph/knowledge-graph"
	if limit > 0 {
		path = fmt.Sprintf("%s?limit=%d", path, limit)
	}
	var resp dto.KnowledgeGraphResponse
	if err := c.Get(path, &resp); err != nil {
		return nil, fmt.Errorf("get knowledge graph: %w", err)
	}
	return &resp, nil
}
