package service

import (
	"testing"
)

func TestNewLouvainService(t *testing.T) {
	svc := NewLouvainService(nil)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.maxIterations != 10 {
		t.Errorf("expected 10 max iterations, got %d", svc.maxIterations)
	}
	if svc.minModularity != 0.001 {
		t.Errorf("expected 0.001 min modularity, got %f", svc.minModularity)
	}
}

func TestDetectCommunities_NilClient(t *testing.T) {
	svc := NewLouvainService(nil)
	result, err := svc.DetectCommunities(nil, "t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Communities) != 0 {
		t.Errorf("expected 0 communities without client, got %d", len(result.Communities))
	}
}

func TestCommunity_Structure(t *testing.T) {
	c := Community{
		ID:        0,
		MemberIDs: []string{"m1", "m2", "m3"},
		Size:      3,
		Density:   0.67,
	}
	if c.ID != 0 {
		t.Error("expected ID 0")
	}
	if c.Size != 3 {
		t.Error("expected size 3")
	}
	if c.Density != 0.67 {
		t.Errorf("expected 0.67 density, got %f", c.Density)
	}
}

func TestCommunityResult_Structure(t *testing.T) {
	r := CommunityResult{
		Communities: []Community{
			{ID: 0, Size: 5},
			{ID: 1, Size: 3},
		},
		TotalNodes: 10,
		TotalEdges: 15,
		Modularity: 0.42,
		Iterations: 3,
	}
	if len(r.Communities) != 2 {
		t.Error("expected 2 communities")
	}
	if r.TotalNodes != 10 {
		t.Error("expected 10 nodes")
	}
	if r.Modularity != 0.42 {
		t.Errorf("expected 0.42, got %f", r.Modularity)
	}
}

func TestComputeModularity_Empty(t *testing.T) {
	q := computeModularity(nil, nil, nil, 0, 0)
	if q != 0 {
		t.Errorf("expected 0 modularity for empty graph, got %f", q)
	}
}

func TestComputeModularity_TwoCommunities(t *testing.T) {
	// Triangle A-B-C (community 0) and triangle D-E-F (community 1)
	// with one weak edge B-D
	adj := [][]weightedNeighbor{
		// A(0): B, C
		{{node: 1, weight: 1}, {node: 2, weight: 1}},
		// B(1): A, C, D
		{{node: 0, weight: 1}, {node: 2, weight: 1}, {node: 3, weight: 0.1}},
		// C(2): A, B
		{{node: 0, weight: 1}, {node: 1, weight: 1}},
		// D(3): B, E, F
		{{node: 1, weight: 0.1}, {node: 4, weight: 1}, {node: 5, weight: 1}},
		// E(4): D, F
		{{node: 3, weight: 1}, {node: 5, weight: 1}},
		// F(5): D, E
		{{node: 3, weight: 1}, {node: 4, weight: 1}},
	}

	strength := make([]float64, 6)
	for i := 0; i < 6; i++ {
		for _, nb := range adj[i] {
			strength[i] += nb.weight
		}
	}

	totalWeight := 0.0
	for _, s := range strength {
		totalWeight += s
	}
	m2 := totalWeight // each edge counted twice in undirected adj

	// All in one community
	allOne := []int{0, 0, 0, 0, 0, 0}
	qOne := computeModularity(allOne, adj, strength, m2, 6)

	// Two communities
	twoCommunities := []int{0, 0, 0, 1, 1, 1}
	qTwo := computeModularity(twoCommunities, adj, strength, m2, 6)

	if qTwo <= qOne {
		t.Errorf("two communities (%.4f) should have higher modularity than one (%.4f)", qTwo, qOne)
	}
}

func TestSortCommunities(t *testing.T) {
	communities := []Community{
		{ID: 0, Size: 2},
		{ID: 1, Size: 5},
		{ID: 2, Size: 3},
	}
	sortCommunities(communities)
	if communities[0].Size != 5 {
		t.Errorf("expected largest first, got size %d", communities[0].Size)
	}
	if communities[2].Size != 2 {
		t.Errorf("expected smallest last, got size %d", communities[2].Size)
	}
}

func TestComputeCommunityDensity_Complete(t *testing.T) {
	// Complete triangle: 3 nodes, 3 edges → density = 1.0
	nodeIndex := map[string]int{"a": 0, "b": 1, "c": 2}
	adj := [][]weightedNeighbor{
		{{node: 1, weight: 1}, {node: 2, weight: 1}},
		{{node: 0, weight: 1}, {node: 2, weight: 1}},
		{{node: 0, weight: 1}, {node: 1, weight: 1}},
	}

	density := computeCommunityDensity([]string{"a", "b", "c"}, nodeIndex, adj)
	if density != 1.0 {
		t.Errorf("expected density 1.0 for complete triangle, got %f", density)
	}
}

func TestComputeCommunityDensity_Sparse(t *testing.T) {
	// Path A-B-C: 3 nodes, 2 edges → density = 2/3 ≈ 0.667
	nodeIndex := map[string]int{"a": 0, "b": 1, "c": 2}
	adj := [][]weightedNeighbor{
		{{node: 1, weight: 1}},
		{{node: 0, weight: 1}, {node: 2, weight: 1}},
		{{node: 1, weight: 1}},
	}

	density := computeCommunityDensity([]string{"a", "b", "c"}, nodeIndex, adj)
	expected := 2.0 / 3.0
	if density < expected-0.01 || density > expected+0.01 {
		t.Errorf("expected density ~0.667, got %f", density)
	}
}

func TestComputeCommunityDensity_Single(t *testing.T) {
	nodeIndex := map[string]int{"a": 0}
	adj := [][]weightedNeighbor{{}}

	density := computeCommunityDensity([]string{"a"}, nodeIndex, adj)
	if density != 0 {
		t.Errorf("expected 0 density for single node, got %f", density)
	}
}
