package commands

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/dto"
)

func TestImportCmd_Success(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "memories.json")
	content := `{"memories": [
		{"content": "Go is great"},
		{"content": "Python is versatile"}
	]}`
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var count int
	a := &App{
		Creator: &mockCreator{fn: func(ctx context.Context, req dto.CreateMemoryRequest) (*domain.Memory, error) {
			count++
			return &domain.Memory{ID: fmt.Sprintf("m%d", count), Content: req.Content}, nil
		}},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "table",
	}

	cmd := newImportCmd(a)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{file})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 imports, got %d", count)
	}
	if !strings.Contains(buf.String(), "Imported: 2") {
		t.Errorf("expected import count, got %q", buf.String())
	}
}

func TestImportCmd_DryRun(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "memories.json")
	content := `{"memories": [{"content": "test"}]}`
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var count int
	a := &App{
		Creator: &mockCreator{fn: func(ctx context.Context, req dto.CreateMemoryRequest) (*domain.Memory, error) {
			count++
			return &domain.Memory{ID: "m1"}, nil
		}},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "table",
	}

	cmd := newImportCmd(a)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{file, "--dry-run"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Error("dry run should not create memories")
	}
	if !strings.Contains(buf.String(), "Dry run") {
		t.Error("expected dry run message")
	}
}

func TestImportCmd_SkipDuplicates(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "memories.json")
	content := `{"memories": [
		{"content": "first"},
		{"content": "duplicate"},
		{"content": "third"}
	]}`
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	call := 0
	a := &App{
		Creator: &mockCreator{fn: func(ctx context.Context, req dto.CreateMemoryRequest) (*domain.Memory, error) {
			call++
			if call == 2 {
				return nil, fmt.Errorf("duplicate detected")
			}
			return &domain.Memory{ID: fmt.Sprintf("m%d", call)}, nil
		}},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "table",
	}

	cmd := newImportCmd(a)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{file, "--skip-duplicates"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "Imported: 2") {
		t.Errorf("expected 2 imported, got %q", buf.String())
	}
	if !strings.Contains(buf.String(), "Skipped: 1") {
		t.Errorf("expected 1 skipped, got %q", buf.String())
	}
}

func TestImportCmd_FileNotFound(t *testing.T) {
	a := &App{
		Creator:  &mockCreator{},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "table",
	}

	cmd := newImportCmd(a)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"/nonexistent/file.json"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestImportCmd_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(file, []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	a := &App{
		Creator:  &mockCreator{},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "table",
	}

	cmd := newImportCmd(a)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{file})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestImportCmd_EmptyMemories(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(file, []byte(`{"memories": []}`), 0644); err != nil {
		t.Fatal(err)
	}

	a := &App{
		Creator:  &mockCreator{},
		TenantID: "a9f814d2-4dae-41f3-851b-8aa3d4706561",
		Output:   "table",
	}

	cmd := newImportCmd(a)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{file})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for empty memories")
	}
}

func TestImportCmd_MissingArgs(t *testing.T) {
	a := &App{Output: "table"}
	cmd := newImportCmd(a)
	cmd.SetArgs([]string{})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	if err := cmd.Execute(); err == nil {
		t.Error("expected error for missing args")
	}
}

func TestImportCmd_InvalidTenantID(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "memories.json")
	content := `{"memories": [{"content": "test"}]}`
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	a := &App{
		Creator:  &mockCreator{fn: func(ctx context.Context, req dto.CreateMemoryRequest) (*domain.Memory, error) {
			return &domain.Memory{ID: "m1"}, nil
		}},
		TenantID: "bad",
		Output:   "table",
	}

	cmd := newImportCmd(a)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{file})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid tenant ID")
	}
	if !strings.Contains(err.Error(), "invalid tenant ID") {
		t.Errorf("expected tenant validation error, got %v", err)
	}
}
