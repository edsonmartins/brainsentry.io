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

type mockUpdater struct {
	fn func(ctx context.Context, id string, req dto.UpdateMemoryRequest) (*domain.Memory, error)
}

func (m *mockUpdater) UpdateMemory(ctx context.Context, id string, req dto.UpdateMemoryRequest) (*domain.Memory, error) {
	return m.fn(ctx, id, req)
}

func TestEditCmd_Success(t *testing.T) {
	var capturedID string
	var capturedReq dto.UpdateMemoryRequest
	a := &App{
		Updater: &mockUpdater{fn: func(ctx context.Context, id string, req dto.UpdateMemoryRequest) (*domain.Memory, error) {
			capturedID = id
			capturedReq = req
			return &domain.Memory{ID: id, Content: req.Content}, nil
		}},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "table",
	}

	cmd := newEditCmd(a)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"mem-123", "--content", "updated content", "--reason", "fixing typo"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "mem-123" {
		t.Errorf("expected mem-123, got %s", capturedID)
	}
	if capturedReq.Content != "updated content" {
		t.Errorf("expected updated content, got %q", capturedReq.Content)
	}
	if capturedReq.ChangeReason != "fixing typo" {
		t.Errorf("expected reason, got %q", capturedReq.ChangeReason)
	}
	if !strings.Contains(buf.String(), "Updated memory") {
		t.Error("expected success message")
	}
}

func TestEditCmd_JSONOutput(t *testing.T) {
	a := &App{
		Updater: &mockUpdater{fn: func(ctx context.Context, id string, req dto.UpdateMemoryRequest) (*domain.Memory, error) {
			return &domain.Memory{ID: id}, nil
		}},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "json",
	}

	cmd := newEditCmd(a)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"mem-1", "--content", "new"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), `"id"`) {
		t.Error("expected JSON output")
	}
}

func TestEditCmd_MissingArgs(t *testing.T) {
	a := &App{Output: "table"}
	cmd := newEditCmd(a)
	cmd.SetArgs([]string{})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	if err := cmd.Execute(); err == nil {
		t.Error("expected error for missing args")
	}
}

func TestEditCmd_ServiceError(t *testing.T) {
	a := &App{
		Updater: &mockUpdater{fn: func(ctx context.Context, id string, req dto.UpdateMemoryRequest) (*domain.Memory, error) {
			return nil, fmt.Errorf("not found")
		}},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "table",
	}

	cmd := newEditCmd(a)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"mem-1", "--content", "new"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestEditCmd_NilUpdater(t *testing.T) {
	a := &App{Output: "table"}
	cmd := newEditCmd(a)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"mem-1"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error for nil updater")
	}
}
