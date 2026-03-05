package service

import (
	"testing"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
)

func TestNewReflectionService(t *testing.T) {
	svc := NewReflectionService(nil, nil, nil)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.minClusterSize != 3 {
		t.Errorf("expected min cluster size 3, got %d", svc.minClusterSize)
	}
	if svc.similarityThreshold != 8 {
		t.Errorf("expected threshold 8, got %f", svc.similarityThreshold)
	}
}

func TestComputeClusterSaliency_Empty(t *testing.T) {
	svc := NewReflectionService(nil, nil, nil)
	cluster := MemoryCluster{Size: 0}
	saliency := svc.computeClusterSaliency(cluster)
	if saliency != 0 {
		t.Errorf("expected 0 saliency for empty cluster, got %f", saliency)
	}
}

func TestComputeClusterSaliency_WithMemories(t *testing.T) {
	svc := NewReflectionService(nil, nil, nil)
	now := time.Now()
	cluster := MemoryCluster{
		Size: 3,
		Memories: []domain.Memory{
			{CreatedAt: now, Importance: domain.ImportanceCritical, AccessCount: 10, MemoryType: domain.MemoryTypeSemantic},
			{CreatedAt: now, Importance: domain.ImportanceImportant, AccessCount: 5, MemoryType: domain.MemoryTypeSemantic},
			{CreatedAt: now, Importance: domain.ImportanceMinor, AccessCount: 2, MemoryType: domain.MemoryTypeSemantic},
		},
	}
	saliency := svc.computeClusterSaliency(cluster)
	if saliency <= 0 {
		t.Errorf("expected positive saliency, got %f", saliency)
	}
}

func TestComputeClusterSaliency_LargerClusterBoosted(t *testing.T) {
	svc := NewReflectionService(nil, nil, nil)
	now := time.Now()
	makeMemory := func() domain.Memory {
		return domain.Memory{CreatedAt: now, Importance: domain.ImportanceImportant, AccessCount: 5, MemoryType: domain.MemoryTypeSemantic}
	}

	small := MemoryCluster{Size: 3, Memories: []domain.Memory{makeMemory(), makeMemory(), makeMemory()}}
	large := MemoryCluster{Size: 6, Memories: []domain.Memory{makeMemory(), makeMemory(), makeMemory(), makeMemory(), makeMemory(), makeMemory()}}

	smallSaliency := svc.computeClusterSaliency(small)
	largeSaliency := svc.computeClusterSaliency(large)

	if largeSaliency <= smallSaliency {
		t.Errorf("larger cluster (%.4f) should have higher saliency than smaller (%.4f)", largeSaliency, smallSaliency)
	}
}

func TestSimpleSynthesis(t *testing.T) {
	svc := NewReflectionService(nil, nil, nil)
	cluster := MemoryCluster{
		Size: 3,
		Memories: []domain.Memory{
			{Summary: "Go patterns"},
			{Summary: "Go best practices"},
			{Content: "Go error handling is important"},
		},
	}
	result := svc.simpleSynthesis(cluster)
	if result == "" {
		t.Error("expected non-empty synthesis")
	}
	if !containsSubstr(result, "3 memories") {
		t.Errorf("expected '3 memories' in result, got: %s", result)
	}
}

func TestMemoryCluster_Structure(t *testing.T) {
	cluster := MemoryCluster{
		ID:       "cluster-0",
		Topic:    "Go development",
		Size:     3,
		Saliency: 0.75,
	}
	if cluster.ID != "cluster-0" {
		t.Error("expected cluster-0")
	}
	if cluster.Saliency != 0.75 {
		t.Error("expected 0.75 saliency")
	}
}

func TestReflectionResult_Structure(t *testing.T) {
	result := ReflectionResult{
		ClustersFound:      3,
		ReflectionsCreated: 2,
		MemoriesProcessed:  15,
		ConsolidatedIDs:    []string{"m1", "m2", "m3"},
	}
	if result.ClustersFound != 3 {
		t.Error("expected 3 clusters")
	}
	if len(result.ConsolidatedIDs) != 3 {
		t.Error("expected 3 consolidated IDs")
	}
}
