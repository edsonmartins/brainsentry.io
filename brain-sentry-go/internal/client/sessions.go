package client

import (
	"fmt"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/dto"
)

// ListSessions returns all sessions.
func (c *Client) ListSessions() ([]domain.Session, error) {
	var resp []domain.Session
	if err := c.Get("/v1/sessions/active", &resp); err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	return resp, nil
}

// GetSession returns a single session by ID.
func (c *Client) GetSession(id string) (*domain.Session, error) {
	var resp domain.Session
	if err := c.Get("/v1/sessions/"+id, &resp); err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	return &resp, nil
}

// AnalyzeSession runs analysis on a session.
func (c *Client) AnalyzeSession(req *dto.SessionAnalysisRequest) (*dto.SessionAnalysisResponse, error) {
	var resp dto.SessionAnalysisResponse
	if err := c.Post("/v1/notes/analyze", req, &resp); err != nil {
		return nil, fmt.Errorf("analyze session: %w", err)
	}
	return &resp, nil
}

// GetSessionObservations returns typed observations for a session.
func (c *Client) GetSessionObservations(sessionID string) ([]dto.SessionObservationResponse, error) {
	var resp []dto.SessionObservationResponse
	if err := c.Get("/v1/sessions/"+sessionID+"/events", &resp); err != nil {
		return nil, fmt.Errorf("get session observations: %w", err)
	}
	return resp, nil
}
