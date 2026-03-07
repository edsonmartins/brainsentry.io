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

type mockCorrector struct {
	flagFn  func(ctx context.Context, memoryID string, req dto.FlagMemoryRequest) (*domain.MemoryCorrection, error)
	reviewFn func(ctx context.Context, memoryID string, req dto.ReviewCorrectionRequest) (*domain.Memory, error)
}

func (m *mockCorrector) FlagMemory(ctx context.Context, memoryID string, req dto.FlagMemoryRequest) (*domain.MemoryCorrection, error) {
	return m.flagFn(ctx, memoryID, req)
}

func (m *mockCorrector) ReviewCorrection(ctx context.Context, memoryID string, req dto.ReviewCorrectionRequest) (*domain.Memory, error) {
	return m.reviewFn(ctx, memoryID, req)
}

func TestCorrectCmd_Success(t *testing.T) {
	var capturedID string
	var capturedReq dto.FlagMemoryRequest
	a := &App{
		Corrector: &mockCorrector{
			flagFn: func(ctx context.Context, memoryID string, req dto.FlagMemoryRequest) (*domain.MemoryCorrection, error) {
				capturedID = memoryID
				capturedReq = req
				return &domain.MemoryCorrection{ID: "corr-1", MemoryID: memoryID}, nil
			},
		},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "table",
	}

	cmd := newCorrectCmd(a)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"mem-123", "--reason", "outdated information"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "mem-123" {
		t.Errorf("expected mem-123, got %s", capturedID)
	}
	if capturedReq.Reason != "outdated information" {
		t.Errorf("expected reason, got %q", capturedReq.Reason)
	}
	if !strings.Contains(buf.String(), "corr-1") {
		t.Error("expected correction ID in output")
	}
}

func TestCorrectCmd_MissingReason(t *testing.T) {
	a := &App{
		Corrector: &mockCorrector{},
		TenantID:  "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:    "table",
	}

	cmd := newCorrectCmd(a)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"mem-123"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing reason")
	}
	if !strings.Contains(err.Error(), "--reason") {
		t.Errorf("expected error about --reason, got %v", err)
	}
}

func TestCorrectCmd_MissingArgs(t *testing.T) {
	a := &App{Output: "table"}
	cmd := newCorrectCmd(a)
	cmd.SetArgs([]string{})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	if err := cmd.Execute(); err == nil {
		t.Error("expected error for missing args")
	}
}

func TestCorrectCmd_ServiceError(t *testing.T) {
	a := &App{
		Corrector: &mockCorrector{
			flagFn: func(ctx context.Context, memoryID string, req dto.FlagMemoryRequest) (*domain.MemoryCorrection, error) {
				return nil, fmt.Errorf("memory not found")
			},
		},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "table",
	}

	cmd := newCorrectCmd(a)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"mem-999", "--reason", "test"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestCorrectCmd_JSONOutput(t *testing.T) {
	a := &App{
		Corrector: &mockCorrector{
			flagFn: func(ctx context.Context, memoryID string, req dto.FlagMemoryRequest) (*domain.MemoryCorrection, error) {
				return &domain.MemoryCorrection{ID: "corr-2", MemoryID: memoryID}, nil
			},
		},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "json",
	}

	cmd := newCorrectCmd(a)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"mem-1", "--reason", "test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), `"id"`) {
		t.Error("expected JSON output")
	}
}
