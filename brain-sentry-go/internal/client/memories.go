package client

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/integraltech/brainsentry/internal/dto"
)

// ListMemories returns a paginated list of memories.
func (c *Client) ListMemories(page, size int) (*dto.MemoryListResponse, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	params.Set("size", strconv.Itoa(size))

	var resp dto.MemoryListResponse
	if err := c.Get("/v1/memories?"+params.Encode(), &resp); err != nil {
		return nil, fmt.Errorf("list memories: %w", err)
	}
	return &resp, nil
}

// GetMemory returns a single memory by ID.
func (c *Client) GetMemory(id string) (*dto.MemoryResponse, error) {
	var resp dto.MemoryResponse
	if err := c.Get("/v1/memories/"+id, &resp); err != nil {
		return nil, fmt.Errorf("get memory: %w", err)
	}
	return &resp, nil
}

// CreateMemory creates a new memory.
func (c *Client) CreateMemory(req *dto.CreateMemoryRequest) (*dto.MemoryResponse, error) {
	var resp dto.MemoryResponse
	if err := c.Post("/v1/memories", req, &resp); err != nil {
		return nil, fmt.Errorf("create memory: %w", err)
	}
	return &resp, nil
}

// UpdateMemory updates an existing memory.
func (c *Client) UpdateMemory(id string, req *dto.UpdateMemoryRequest) (*dto.MemoryResponse, error) {
	var resp dto.MemoryResponse
	if err := c.Put("/v1/memories/"+id, req, &resp); err != nil {
		return nil, fmt.Errorf("update memory: %w", err)
	}
	return &resp, nil
}

// DeleteMemory deletes a memory by ID.
func (c *Client) DeleteMemory(id string) error {
	if err := c.Delete("/v1/memories/" + id); err != nil {
		return fmt.Errorf("delete memory: %w", err)
	}
	return nil
}

// SearchMemories performs a semantic search.
func (c *Client) SearchMemories(req *dto.SearchRequest) (*dto.SearchResponse, error) {
	var resp dto.SearchResponse
	if err := c.Post("/v1/memories/search", req, &resp); err != nil {
		return nil, fmt.Errorf("search memories: %w", err)
	}
	return &resp, nil
}

// FlagMemory flags a memory as incorrect.
func (c *Client) FlagMemory(id string, req *dto.FlagMemoryRequest) error {
	if err := c.Post("/v1/memories/"+id+"/flag", req, nil); err != nil {
		return fmt.Errorf("flag memory: %w", err)
	}
	return nil
}

// ReviewCorrection reviews a flagged correction.
func (c *Client) ReviewCorrection(memoryID string, req *dto.ReviewCorrectionRequest) error {
	path := fmt.Sprintf("/v1/memories/%s/review", memoryID)
	if err := c.Post(path, req, nil); err != nil {
		return fmt.Errorf("review correction: %w", err)
	}
	return nil
}

// ExportMemories exports all memories as JSON.
func (c *Client) ExportMemories() ([]dto.MemoryResponse, error) {
	var resp []dto.MemoryResponse
	if err := c.Get("/v1/batch/export", &resp); err != nil {
		return nil, fmt.Errorf("export memories: %w", err)
	}
	return resp, nil
}
