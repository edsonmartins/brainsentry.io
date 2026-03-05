package service

import (
	"math"
	"testing"
)

func TestEmbeddingService_Embed(t *testing.T) {
	svc := NewEmbeddingService(384, "", "", "")

	emb := svc.Embed("hello world")
	if len(emb) != 384 {
		t.Errorf("expected 384 dimensions, got %d", len(emb))
	}

	// Same input should produce same output
	emb2 := svc.Embed("hello world")
	for i := range emb {
		if emb[i] != emb2[i] {
			t.Errorf("expected deterministic embedding at index %d", i)
			break
		}
	}

	// Different input should produce different output
	emb3 := svc.Embed("different text")
	same := true
	for i := range emb {
		if emb[i] != emb3[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("expected different embeddings for different inputs")
	}
}

func TestEmbeddingService_EmbedBatch(t *testing.T) {
	svc := NewEmbeddingService(128, "", "", "")

	embeddings := svc.EmbedBatch([]string{"hello", "world", "test"})
	if len(embeddings) != 3 {
		t.Errorf("expected 3 embeddings, got %d", len(embeddings))
	}

	for i, emb := range embeddings {
		if len(emb) != 128 {
			t.Errorf("embedding %d: expected 128 dimensions, got %d", i, len(emb))
		}
	}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a, b     []float32
		expected float64
	}{
		{"identical", []float32{1, 0, 0}, []float32{1, 0, 0}, 1.0},
		{"orthogonal", []float32{1, 0, 0}, []float32{0, 1, 0}, 0.0},
		{"opposite", []float32{1, 0, 0}, []float32{-1, 0, 0}, -1.0},
		{"similar", []float32{1, 1, 0}, []float32{1, 0, 0}, 1 / math.Sqrt(2)},
		{"empty", []float32{}, []float32{}, 0.0},
		{"different_length", []float32{1, 0}, []float32{1, 0, 0}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CosineSimilarity(tt.a, tt.b)
			if math.Abs(result-tt.expected) > 0.001 {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestEmbeddingService_EmptyInput(t *testing.T) {
	svc := NewEmbeddingService(64, "", "", "")

	emb := svc.Embed("")
	if len(emb) != 64 {
		t.Errorf("expected 64 dimensions for empty input, got %d", len(emb))
	}
}
