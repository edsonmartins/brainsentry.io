package commands

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/dto"
)

type mockSearcher struct {
	fn func(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error)
}

func (m *mockSearcher) SearchMemories(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error) {
	return m.fn(ctx, req)
}

func TestSearchCmd_Success(t *testing.T) {
	a := &App{
		Searcher: &mockSearcher{fn: func(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error) {
			return &dto.SearchResponse{
				Results: []dto.MemoryResponse{
					{ID: "m1", Summary: "Go backend tips", Category: domain.CategoryKnowledge, Importance: domain.ImportanceImportant},
					{ID: "m2", Content: "Python ML patterns", Category: domain.CategoryPattern},
				},
				Total:        2,
				SearchTimeMs: 15,
			}, nil
		}},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "table",
	}

	cmd := newSearchCmd(a)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"Go backend"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "m1") {
		t.Error("expected m1 in output")
	}
	if !strings.Contains(out, "2 results") {
		t.Error("expected result count")
	}
}

func TestSearchCmd_WithLimit(t *testing.T) {
	var captured dto.SearchRequest
	a := &App{
		Searcher: &mockSearcher{fn: func(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error) {
			captured = req
			return &dto.SearchResponse{}, nil
		}},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "table",
	}

	cmd := newSearchCmd(a)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"test", "--limit", "5"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured.Limit != 5 {
		t.Errorf("expected limit 5, got %d", captured.Limit)
	}
}

func TestSearchCmd_JSONOutput(t *testing.T) {
	a := &App{
		Searcher: &mockSearcher{fn: func(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error) {
			return &dto.SearchResponse{
				Results: []dto.MemoryResponse{{ID: "m1"}},
				Total:   1,
			}, nil
		}},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "json",
	}

	cmd := newSearchCmd(a)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), `"total"`) {
		t.Error("expected JSON output")
	}
}

func TestSearchCmd_MissingArgs(t *testing.T) {
	a := &App{Output: "table"}
	cmd := newSearchCmd(a)
	cmd.SetArgs([]string{})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	if err := cmd.Execute(); err == nil {
		t.Error("expected error for missing args")
	}
}

func TestSearchCmd_ServiceError(t *testing.T) {
	a := &App{
		Searcher: &mockSearcher{fn: func(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error) {
			return nil, fmt.Errorf("search failed")
		}},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "table",
	}

	cmd := newSearchCmd(a)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"test"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}
