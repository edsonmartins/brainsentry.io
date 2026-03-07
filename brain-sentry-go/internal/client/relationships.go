package client

import (
	"fmt"

	"github.com/integraltech/brainsentry/internal/domain"
)

// GetMemoryRelationships returns relationships for a memory.
func (c *Client) GetMemoryRelationships(memoryID string) ([]domain.MemoryRelationship, error) {
	var resp []domain.MemoryRelationship
	if err := c.Get("/v1/relationships/"+memoryID+"/related", &resp); err != nil {
		return nil, fmt.Errorf("get relationships: %w", err)
	}
	return resp, nil
}
