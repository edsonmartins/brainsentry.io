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

type mockLister struct {
	fn func(ctx context.Context, page, size int) (*dto.MemoryListResponse, error)
}

func (m *mockLister) ListMemories(ctx context.Context, page, size int) (*dto.MemoryListResponse, error) {
	return m.fn(ctx, page, size)
}

func TestListCmd_Success(t *testing.T) {
	a := &App{
		Lister: &mockLister{fn: func(ctx context.Context, page, size int) (*dto.MemoryListResponse, error) {
			return &dto.MemoryListResponse{
				Memories: []dto.MemoryResponse{
					{ID: "m1", Summary: "First memory", Category: domain.CategoryKnowledge},
					{ID: "m2", Summary: "Second memory", Category: domain.CategoryInsight},
				},
				Page:          0,
				Size:          20,
				TotalElements: 2,
				TotalPages:    1,
			}, nil
		}},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "table",
	}

	cmd := newListCmd(a)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "m1") {
		t.Error("expected m1")
	}
	if !strings.Contains(out, "Page 1/1") {
		t.Error("expected pagination info")
	}
}

func TestListCmd_WithPagination(t *testing.T) {
	var capturedPage, capturedSize int
	a := &App{
		Lister: &mockLister{fn: func(ctx context.Context, page, size int) (*dto.MemoryListResponse, error) {
			capturedPage = page
			capturedSize = size
			return &dto.MemoryListResponse{}, nil
		}},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "table",
	}

	cmd := newListCmd(a)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--page", "2", "--size", "10"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedPage != 2 {
		t.Errorf("expected page 2, got %d", capturedPage)
	}
	if capturedSize != 10 {
		t.Errorf("expected size 10, got %d", capturedSize)
	}
}

func TestListCmd_JSONOutput(t *testing.T) {
	a := &App{
		Lister: &mockLister{fn: func(ctx context.Context, page, size int) (*dto.MemoryListResponse, error) {
			return &dto.MemoryListResponse{
				Memories:      []dto.MemoryResponse{{ID: "m1"}},
				TotalElements: 1,
			}, nil
		}},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "json",
	}

	cmd := newListCmd(a)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), `"totalElements"`) {
		t.Error("expected JSON output")
	}
}

func TestListCmd_ServiceError(t *testing.T) {
	a := &App{
		Lister: &mockLister{fn: func(ctx context.Context, page, size int) (*dto.MemoryListResponse, error) {
			return nil, fmt.Errorf("db error")
		}},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "table",
	}

	cmd := newListCmd(a)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestListCmd_NilLister(t *testing.T) {
	a := &App{Output: "table"}
	cmd := newListCmd(a)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error for nil lister")
	}
}
