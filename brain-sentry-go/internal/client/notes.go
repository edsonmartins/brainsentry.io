package client

import (
	"fmt"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/dto"
)

// ListNotes returns all notes.
func (c *Client) ListNotes() ([]domain.Note, error) {
	var resp []domain.Note
	if err := c.Get("/v1/notes/", &resp); err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	return resp, nil
}

// ListHindsightNotes returns all hindsight notes.
func (c *Client) ListHindsightNotes() ([]domain.HindsightNote, error) {
	var resp []domain.HindsightNote
	if err := c.Get("/v1/notes/hindsight", &resp); err != nil {
		return nil, fmt.Errorf("list hindsight notes: %w", err)
	}
	return resp, nil
}

// CreateHindsightNote creates a new hindsight note.
func (c *Client) CreateHindsightNote(req *dto.CreateHindsightNoteRequest) (*domain.HindsightNote, error) {
	var resp domain.HindsightNote
	if err := c.Post("/v1/notes/hindsight", req, &resp); err != nil {
		return nil, fmt.Errorf("create hindsight note: %w", err)
	}
	return &resp, nil
}

// GetSessionNotes returns notes for a specific session.
func (c *Client) GetSessionNotes(sessionID string) ([]domain.Note, error) {
	var resp []domain.Note
	if err := c.Get("/v1/notes/session/"+sessionID, &resp); err != nil {
		return nil, fmt.Errorf("get session notes: %w", err)
	}
	return resp, nil
}
