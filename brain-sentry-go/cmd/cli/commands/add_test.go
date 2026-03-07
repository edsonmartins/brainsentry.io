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

type mockCreator struct {
	fn func(ctx context.Context, req dto.CreateMemoryRequest) (*domain.Memory, error)
}

func (m *mockCreator) CreateMemory(ctx context.Context, req dto.CreateMemoryRequest) (*domain.Memory, error) {
	return m.fn(ctx, req)
}

func TestAddCmd_Success(t *testing.T) {
	var captured dto.CreateMemoryRequest
	a := &App{
		Creator: &mockCreator{fn: func(ctx context.Context, req dto.CreateMemoryRequest) (*domain.Memory, error) {
			captured = req
			return &domain.Memory{
				ID:         "mem-123",
				Content:    req.Content,
				Category:   domain.CategoryKnowledge,
				Importance: domain.ImportanceMinor,
			}, nil
		}},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "table",
	}

	cmd := newAddCmd(a)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"Go is great for backend development"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if captured.Content != "Go is great for backend development" {
		t.Errorf("expected captured content, got %q", captured.Content)
	}
	if !strings.Contains(buf.String(), "mem-123") {
		t.Error("expected memory ID in output")
	}
}

func TestAddCmd_WithFlags(t *testing.T) {
	var captured dto.CreateMemoryRequest
	a := &App{
		Creator: &mockCreator{fn: func(ctx context.Context, req dto.CreateMemoryRequest) (*domain.Memory, error) {
			captured = req
			return &domain.Memory{ID: "mem-456", Content: req.Content, Category: req.Category, Importance: req.Importance}, nil
		}},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "table",
	}

	cmd := newAddCmd(a)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"test content", "--category", "KNOWLEDGE", "--importance", "CRITICAL", "--tags", "go,backend"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured.Category != domain.CategoryKnowledge {
		t.Errorf("expected KNOWLEDGE, got %s", captured.Category)
	}
	if captured.Importance != domain.ImportanceCritical {
		t.Errorf("expected CRITICAL, got %s", captured.Importance)
	}
	if len(captured.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(captured.Tags))
	}
}

func TestAddCmd_JSONOutput(t *testing.T) {
	a := &App{
		Creator: &mockCreator{fn: func(ctx context.Context, req dto.CreateMemoryRequest) (*domain.Memory, error) {
			return &domain.Memory{ID: "mem-789", Content: req.Content}, nil
		}},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "json",
	}

	cmd := newAddCmd(a)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), `"id"`) {
		t.Error("expected JSON output with id field")
	}
}

func TestAddCmd_MissingArgs(t *testing.T) {
	a := &App{Output: "table"}
	cmd := newAddCmd(a)
	cmd.SetArgs([]string{})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	if err := cmd.Execute(); err == nil {
		t.Error("expected error for missing args")
	}
}

func TestAddCmd_ServiceError(t *testing.T) {
	a := &App{
		Creator: &mockCreator{fn: func(ctx context.Context, req dto.CreateMemoryRequest) (*domain.Memory, error) {
			return nil, fmt.Errorf("database connection failed")
		}},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "table",
	}

	cmd := newAddCmd(a)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"test"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "database connection failed") {
		t.Errorf("expected service error message, got %v", err)
	}
}

func TestAddCmd_NilCreator(t *testing.T) {
	a := &App{Output: "table"}
	cmd := newAddCmd(a)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"test"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for nil creator")
	}
}

func TestAddCmd_InvalidTenantID(t *testing.T) {
	a := &App{
		Creator: &mockCreator{fn: func(ctx context.Context, req dto.CreateMemoryRequest) (*domain.Memory, error) {
			return &domain.Memory{ID: "m1"}, nil
		}},
		TenantID: "not-a-uuid",
		Output:   "table",
	}

	cmd := newAddCmd(a)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"test"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid tenant ID")
	}
	if !strings.Contains(err.Error(), "invalid tenant ID") {
		t.Errorf("expected tenant validation error, got %v", err)
	}
}
