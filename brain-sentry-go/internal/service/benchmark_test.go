package service

import (
	"context"
	"testing"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// Benchmark cosine similarity - core operation for vector search
func BenchmarkCosineSimilarity(b *testing.B) {
	// 384-dimension vectors (all-MiniLM-L6-v2 size)
	v1 := make([]float32, 384)
	v2 := make([]float32, 384)
	for i := range v1 {
		v1[i] = float32(i) * 0.001
		v2[i] = float32(i) * 0.002
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cosineSimilarity(v1, v2)
	}
}

func BenchmarkCosineSimilarity_SmallVectors(b *testing.B) {
	v1 := []float32{1, 2, 3, 4, 5}
	v2 := []float32{5, 4, 3, 2, 1}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cosineSimilarity(v1, v2)
	}
}

// Benchmark embedding generation (hash-based fallback)
func BenchmarkEmbeddingService_HashFallback(b *testing.B) {
	svc := NewEmbeddingService(384, "", "", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.Embed("This is a test sentence for embedding generation benchmark")
	}
}

func BenchmarkEmbeddingService_HashFallback_LongText(b *testing.B) {
	svc := NewEmbeddingService(384, "", "", "")
	longText := make([]byte, 10000)
	for i := range longText {
		longText[i] = byte('a' + i%26)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.Embed(string(longText))
	}
}

// Benchmark importance ranking
func BenchmarkImportanceRank(b *testing.B) {
	levels := []domain.ImportanceLevel{
		domain.ImportanceCritical,
		domain.ImportanceImportant,
		domain.ImportanceMinor,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		importanceRank(levels[i%3])
	}
}

// Benchmark promote/demote operations
func BenchmarkPromoteImportance(b *testing.B) {
	levels := []domain.ImportanceLevel{
		domain.ImportanceMinor,
		domain.ImportanceImportant,
		domain.ImportanceCritical,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		promoteImportance(levels[i%3])
	}
}

func BenchmarkDemoteImportance(b *testing.B) {
	levels := []domain.ImportanceLevel{
		domain.ImportanceCritical,
		domain.ImportanceImportant,
		domain.ImportanceMinor,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		demoteImportance(levels[i%3])
	}
}

// Benchmark memory helpfulness rate calculation
func BenchmarkMemory_HelpfulnessRate(b *testing.B) {
	m := &domain.Memory{
		HelpfulCount:    150,
		NotHelpfulCount: 50,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.HelpfulnessRate()
	}
}

// Benchmark memory relevance score calculation
func BenchmarkMemory_RelevanceScore(b *testing.B) {
	m := &domain.Memory{
		AccessCount:     100,
		InjectionCount:  50,
		HelpfulCount:    75,
		NotHelpfulCount: 25,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.RelevanceScore()
	}
}

// Benchmark session operations
func BenchmarkSessionService_CreateSession(b *testing.B) {
	svc := NewSessionService(DefaultSessionConfig(), nil)
	ctx := tenant.WithTenant(context.Background(), "bench-tenant")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.CreateSession(ctx, "user-123")
	}
}

func BenchmarkSessionService_TouchSession(b *testing.B) {
	svc := NewSessionService(DefaultSessionConfig(), nil)
	ctx := tenant.WithTenant(context.Background(), "bench-tenant")
	session := svc.CreateSession(ctx, "user-123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.TouchSession(ctx, session.ID)
	}
}

// Benchmark webhook HMAC computation
func BenchmarkComputeHMAC(b *testing.B) {
	payload := []byte(`{"event":"memory.created","data":{"id":"abc-123","content":"test"}}`)
	secret := "whsec_test_secret_key_12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		computeHMAC(payload, secret)
	}
}

func BenchmarkComputeHMAC_LargePayload(b *testing.B) {
	payload := make([]byte, 10000)
	for i := range payload {
		payload[i] = byte('a' + i%26)
	}
	secret := "whsec_test_secret_key_12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		computeHMAC(payload, secret)
	}
}

// Benchmark learning service decisions
func BenchmarkLearningService_ShouldPromote(b *testing.B) {
	svc := NewLearningService(nil, nil, nil, DefaultLearningConfig())
	m := &domain.Memory{
		HelpfulCount:    10,
		NotHelpfulCount: 2,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.shouldPromote(m)
	}
}

func BenchmarkLearningService_IsObsolete(b *testing.B) {
	svc := NewLearningService(nil, nil, nil, DefaultLearningConfig())
	m := &domain.Memory{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.isObsolete(m)
	}
}
