package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"time"
)

// EmbeddingService handles text-to-vector embedding via external API.
type EmbeddingService struct {
	dimensions int
	apiKey     string
	baseURL    string
	model      string
	client     *http.Client
}

// NewEmbeddingService creates a new EmbeddingService.
// If apiKey is empty, falls back to hash-based placeholder embeddings.
func NewEmbeddingService(dimensions int, apiKey, baseURL, model string) *EmbeddingService {
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	if model == "" {
		model = "openai/text-embedding-3-small"
	}
	return &EmbeddingService{
		dimensions: dimensions,
		apiKey:     apiKey,
		baseURL:    baseURL,
		model:      model,
		client:     &http.Client{Timeout: 30 * time.Second},
	}
}

type embeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Embed converts text to a vector embedding.
func (s *EmbeddingService) Embed(text string) []float32 {
	if s.apiKey == "" {
		return s.hashEmbed(text)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	embeddings, err := s.callAPI(ctx, []string{text})
	if err != nil {
		slog.Warn("embedding API call failed, using hash fallback", "error", err)
		return s.hashEmbed(text)
	}

	if len(embeddings) == 0 {
		return s.hashEmbed(text)
	}

	return embeddings[0]
}

// EmbedBatch converts multiple texts to vector embeddings.
func (s *EmbeddingService) EmbedBatch(texts []string) [][]float32 {
	if s.apiKey == "" {
		results := make([][]float32, len(texts))
		for i, t := range texts {
			results[i] = s.hashEmbed(t)
		}
		return results
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	embeddings, err := s.callAPI(ctx, texts)
	if err != nil {
		slog.Warn("batch embedding API call failed, using hash fallback", "error", err)
		results := make([][]float32, len(texts))
		for i, t := range texts {
			results[i] = s.hashEmbed(t)
		}
		return results
	}

	return embeddings
}

func (s *EmbeddingService) callAPI(ctx context.Context, texts []string) ([][]float32, error) {
	reqBody := embeddingRequest{
		Model: s.model,
		Input: texts,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result embeddingResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("embedding API error: %s", result.Error.Message)
	}

	embeddings := make([][]float32, len(result.Data))
	for _, d := range result.Data {
		vec := make([]float32, len(d.Embedding))
		for j, v := range d.Embedding {
			vec[j] = float32(v)
		}
		// Truncate or pad to configured dimensions
		if len(vec) > s.dimensions {
			vec = vec[:s.dimensions]
		} else if len(vec) < s.dimensions {
			padded := make([]float32, s.dimensions)
			copy(padded, vec)
			vec = padded
		}
		embeddings[d.Index] = vec
	}

	return embeddings, nil
}

// hashEmbed is a deterministic hash-based fallback when no API key is configured.
func (s *EmbeddingService) hashEmbed(text string) []float32 {
	embedding := make([]float32, s.dimensions)

	// Simple hash-based pseudo-embedding for testing/development
	h := uint64(0)
	for i, c := range text {
		h = h*31 + uint64(c) + uint64(i)
	}

	for i := 0; i < s.dimensions; i++ {
		// Mix hash with dimension index
		v := h ^ uint64(i)*2654435761
		v = (v ^ (v >> 16)) * 0x85ebca6b
		v = (v ^ (v >> 13)) * 0xc2b2ae35
		v = v ^ (v >> 16)
		embedding[i] = float32(float64(v)/float64(math.MaxUint64)*2-1) * 0.1
	}

	// Normalize to unit vector
	var norm float64
	for _, v := range embedding {
		norm += float64(v) * float64(v)
	}
	norm = math.Sqrt(norm)
	if norm > 0 {
		for i := range embedding {
			embedding[i] = float32(float64(embedding[i]) / norm)
		}
	}

	return embedding
}

// CosineSimilarity calculates the cosine similarity between two vectors.
func CosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	denominator := math.Sqrt(normA) * math.Sqrt(normB)
	if denominator == 0 {
		return 0
	}
	return dotProduct / denominator
}

// Dimensions returns the embedding dimensions.
func (s *EmbeddingService) Dimensions() int {
	return s.dimensions
}

// HasAPI returns true if the embedding service is configured with an API key.
func (s *EmbeddingService) HasAPI() bool {
	return s.apiKey != ""
}
