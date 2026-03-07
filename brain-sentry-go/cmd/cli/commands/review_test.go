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

func TestReviewCmd_Approve(t *testing.T) {
	var capturedReq dto.ReviewCorrectionRequest
	a := &App{
		Corrector: &mockCorrector{
			reviewFn: func(ctx context.Context, memoryID string, req dto.ReviewCorrectionRequest) (*domain.Memory, error) {
				capturedReq = req
				return &domain.Memory{ID: memoryID}, nil
			},
		},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "table",
	}

	cmd := newReviewCmd(a)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"mem-123", "--action", "approve"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedReq.Action != "approve" {
		t.Errorf("expected approve, got %s", capturedReq.Action)
	}
	if !strings.Contains(buf.String(), "approve") {
		t.Error("expected approve in output")
	}
}

func TestReviewCmd_Reject(t *testing.T) {
	a := &App{
		Corrector: &mockCorrector{
			reviewFn: func(ctx context.Context, memoryID string, req dto.ReviewCorrectionRequest) (*domain.Memory, error) {
				return &domain.Memory{ID: memoryID}, nil
			},
		},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "table",
	}

	cmd := newReviewCmd(a)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"mem-123", "--action", "reject", "--notes", "still valid"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "reject") {
		t.Error("expected reject in output")
	}
}

func TestReviewCmd_InvalidAction(t *testing.T) {
	a := &App{
		Corrector: &mockCorrector{},
		TenantID:  "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:    "table",
	}

	cmd := newReviewCmd(a)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"mem-123", "--action", "maybe"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid action")
	}
	if !strings.Contains(err.Error(), "approve") {
		t.Errorf("expected error about valid actions, got %v", err)
	}
}

func TestReviewCmd_MissingAction(t *testing.T) {
	a := &App{
		Corrector: &mockCorrector{},
		TenantID:  "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:    "table",
	}

	cmd := newReviewCmd(a)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"mem-123"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error for missing action")
	}
}

func TestReviewCmd_ServiceError(t *testing.T) {
	a := &App{
		Corrector: &mockCorrector{
			reviewFn: func(ctx context.Context, memoryID string, req dto.ReviewCorrectionRequest) (*domain.Memory, error) {
				return nil, fmt.Errorf("review failed")
			},
		},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "table",
	}

	cmd := newReviewCmd(a)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"mem-123", "--action", "approve"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error")
	}
}

func TestReviewCmd_JSONOutput(t *testing.T) {
	a := &App{
		Corrector: &mockCorrector{
			reviewFn: func(ctx context.Context, memoryID string, req dto.ReviewCorrectionRequest) (*domain.Memory, error) {
				return &domain.Memory{ID: memoryID}, nil
			},
		},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "json",
	}

	cmd := newReviewCmd(a)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"mem-1", "--action", "approve"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), `"id"`) {
		t.Error("expected JSON output")
	}
}
