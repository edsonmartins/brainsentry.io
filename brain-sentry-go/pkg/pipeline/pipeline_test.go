package pipeline

import (
	"context"
	"errors"
	"testing"
)

func TestPipelineBuildAndRun(t *testing.T) {
	b := NewBuilder().
		Add("a", func(ctx context.Context, in map[string]any) (any, error) { return 1, nil }).
		Add("b", func(ctx context.Context, in map[string]any) (any, error) {
			return in["a"].(int) + 1, nil
		}, "a").
		Add("c", func(ctx context.Context, in map[string]any) (any, error) {
			return in["b"].(int) * 10, nil
		}, "b")

	p, err := b.Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	res, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if res.Outputs["c"].(int) != 20 {
		t.Fatalf("expected 20, got %v", res.Outputs["c"])
	}
}

func TestPipelineMissingDep(t *testing.T) {
	_, err := NewBuilder().
		Add("x", func(ctx context.Context, in map[string]any) (any, error) { return 1, nil }, "missing").
		Build()
	if err == nil {
		t.Fatal("expected error for missing dep")
	}
}

func TestPipelineFailurePropagates(t *testing.T) {
	boom := errors.New("boom")
	p, err := NewBuilder().
		Add("a", func(ctx context.Context, in map[string]any) (any, error) { return nil, boom }).
		Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	_, err = p.Run(context.Background())
	if err == nil {
		t.Fatal("expected error to surface")
	}
}
