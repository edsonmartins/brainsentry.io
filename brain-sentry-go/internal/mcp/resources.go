package mcp

import (
	"context"
)

func (s *Server) registerResources() {
	s.resources["brainsentry://memories"] = Resource{
		URI:         "brainsentry://memories",
		Name:        "All Memories",
		Description: "List all memories stored in Brain Sentry",
		MimeType:    "application/json",
		Handler:     s.resourceListMemories,
	}

	s.resources["brainsentry://notes"] = Resource{
		URI:         "brainsentry://notes",
		Name:        "All Notes",
		Description: "List all notes and insights",
		MimeType:    "application/json",
		Handler:     s.resourceListNotes,
	}

	s.resources["brainsentry://hindsight"] = Resource{
		URI:         "brainsentry://hindsight",
		Name:        "Hindsight Notes",
		Description: "List all hindsight notes (past errors and resolutions)",
		MimeType:    "application/json",
		Handler:     s.resourceListHindsight,
	}
}

func (s *Server) resourceListMemories(ctx context.Context) (any, error) {
	resp, err := s.memoryService.ListMemories(ctx, 0, 100)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *Server) resourceListNotes(ctx context.Context) (any, error) {
	if s.noteService == nil {
		return map[string]any{"notes": []any{}, "total": 0}, nil
	}
	notes, err := s.noteService.ListNotes(ctx, 100)
	if err != nil {
		return nil, err
	}
	return map[string]any{"notes": notes, "total": len(notes)}, nil
}

func (s *Server) resourceListHindsight(ctx context.Context) (any, error) {
	if s.noteService == nil {
		return map[string]any{"hindsightNotes": []any{}, "total": 0}, nil
	}
	notes, err := s.noteService.ListHindsightNotes(ctx, 100)
	if err != nil {
		return nil, err
	}
	return map[string]any{"hindsightNotes": notes, "total": len(notes)}, nil
}
