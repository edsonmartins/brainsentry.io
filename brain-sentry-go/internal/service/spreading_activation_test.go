package service

import (
	"context"
	"math"
	"testing"
)

func TestNewSpreadingActivationService(t *testing.T) {
	svc := NewSpreadingActivationService(nil, nil)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.maxHops != 3 {
		t.Errorf("expected 3 max hops, got %d", svc.maxHops)
	}
	if svc.decayFactor != 0.5 {
		t.Errorf("expected 0.5 decay, got %f", svc.decayFactor)
	}
	if svc.minThreshold != 0.05 {
		t.Errorf("expected 0.05 threshold, got %f", svc.minThreshold)
	}
}

func TestSpread_NilClient(t *testing.T) {
	svc := NewSpreadingActivationService(nil, nil)
	result, err := svc.Spread(nil, []string{"m1", "m2"}, []float64{1.0, 0.8}, "t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.SeedIDs) != 2 {
		t.Errorf("expected 2 seeds, got %d", len(result.SeedIDs))
	}
	if result.TotalActivated != 0 {
		t.Errorf("expected 0 activations without client, got %d", result.TotalActivated)
	}
}

func TestSpread_EmptySeeds(t *testing.T) {
	svc := NewSpreadingActivationService(nil, nil)
	result, err := svc.Spread(nil, nil, nil, "t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalActivated != 0 {
		t.Errorf("expected 0 activations for empty seeds, got %d", result.TotalActivated)
	}
}

func TestMemoryActivation_Structure(t *testing.T) {
	act := MemoryActivation{
		MemoryID:     "m1",
		Activation:   0.75,
		HopsFromSeed: 1,
		PathStrength: 5.0,
	}
	if act.MemoryID != "m1" {
		t.Error("expected m1")
	}
	if act.Activation != 0.75 {
		t.Error("expected 0.75")
	}
	if act.HopsFromSeed != 1 {
		t.Error("expected 1 hop")
	}
}

func TestSortActivations(t *testing.T) {
	activations := []MemoryActivation{
		{MemoryID: "a", Activation: 0.3},
		{MemoryID: "b", Activation: 0.9},
		{MemoryID: "c", Activation: 0.6},
	}
	sortActivations(activations)
	if activations[0].MemoryID != "b" {
		t.Errorf("expected b first, got %s", activations[0].MemoryID)
	}
	if activations[2].MemoryID != "a" {
		t.Errorf("expected a last, got %s", activations[2].MemoryID)
	}
}

func TestActivationResult_Structure(t *testing.T) {
	result := ActivationResult{
		SeedIDs:        []string{"s1", "s2"},
		TotalActivated: 5,
		MaxHops:        3,
		Activations: []MemoryActivation{
			{MemoryID: "n1", Activation: 0.5, HopsFromSeed: 1},
		},
	}
	if result.TotalActivated != 5 {
		t.Error("expected 5")
	}
	if len(result.Activations) != 1 {
		t.Error("expected 1 activation")
	}
}

// --- Extended spreading activation tests ---

func TestSpread_SeedActivationsLessThanSeedIDs(t *testing.T) {
	svc := NewSpreadingActivationService(nil, nil)
	// 3 seeds but only 2 activations provided
	result, err := svc.Spread(nil, []string{"m1", "m2", "m3"}, []float64{0.8, 0.6}, "t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.SeedIDs) != 3 {
		t.Errorf("expected 3 seeds, got %d", len(result.SeedIDs))
	}
	// Third seed should use default activation=1.0 (not panic)
}

func TestSpread_SingleSeed_NilClient(t *testing.T) {
	svc := NewSpreadingActivationService(nil, nil)
	result, err := svc.Spread(nil, []string{"m1"}, []float64{1.0}, "t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.SeedIDs) != 1 {
		t.Errorf("expected 1 seed, got %d", len(result.SeedIDs))
	}
	if result.TotalActivated != 0 {
		t.Error("expected 0 activated without graph client")
	}
}

func TestDecayFormula_PerHop(t *testing.T) {
	// Verify the math: each hop decays by factor 0.5
	// hop 0 (seed): 1.0
	// hop 1: 1.0 * 0.5 = 0.5
	// hop 2: 0.5 * 0.5 = 0.25
	// hop 3: 0.25 * 0.5 = 0.125 (> 0.05 threshold, would propagate if there were neighbors)
	svc := NewSpreadingActivationService(nil, nil)
	activation := 1.0
	for hop := 1; hop <= svc.maxHops; hop++ {
		activation *= svc.decayFactor
		expected := 1.0
		for i := 0; i < hop; i++ {
			expected *= 0.5
		}
		if activation != expected {
			t.Errorf("hop %d: got %f, want %f", hop, activation, expected)
		}
	}
	// After 3 hops: 0.125 > 0.05 (would still propagate)
	if activation < svc.minThreshold {
		t.Error("activation at max hops should still be above threshold")
	}
}

func TestMinThresholdFormula(t *testing.T) {
	svc := NewSpreadingActivationService(nil, nil)
	// Starting activation that decays below threshold in 1 hop
	startActivation := 0.09
	propagated := startActivation * svc.decayFactor // 0.09 * 0.5 = 0.045
	if propagated >= svc.minThreshold {
		t.Error("expected 0.045 < 0.05 threshold")
	}

	// Starting activation that stays above threshold
	startActivation = 0.2
	propagated = startActivation * svc.decayFactor // 0.2 * 0.5 = 0.1
	if propagated < svc.minThreshold {
		t.Error("expected 0.1 >= 0.05 threshold")
	}
}

func TestSortActivations_SingleElement(t *testing.T) {
	activations := []MemoryActivation{
		{MemoryID: "only", Activation: 0.5},
	}
	sortActivations(activations)
	if activations[0].MemoryID != "only" {
		t.Error("expected unchanged")
	}
}

func TestSortActivations_AllSameActivation(t *testing.T) {
	activations := []MemoryActivation{
		{MemoryID: "a", Activation: 0.5},
		{MemoryID: "b", Activation: 0.5},
		{MemoryID: "c", Activation: 0.5},
	}
	sortActivations(activations) // should not panic
	if len(activations) != 3 {
		t.Error("expected 3 elements preserved")
	}
}

func TestSortActivations_Empty(t *testing.T) {
	sortActivations(nil) // should not panic
}

func TestSpread_EmptySeedActivations_UsesDefault(t *testing.T) {
	svc := NewSpreadingActivationService(nil, nil)
	// nil activations should use default 1.0
	result, err := svc.Spread(nil, []string{"m1", "m2"}, nil, "t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.SeedIDs) != 2 {
		t.Errorf("expected 2 seeds, got %d", len(result.SeedIDs))
	}
}

type fakeSpreadingNeighborProvider struct {
	neighbors map[string][]graphNeighbor
	calls     []string
}

func (f *fakeSpreadingNeighborProvider) GetNeighbors(_ context.Context, memoryID, _ string) ([]graphNeighbor, error) {
	f.calls = append(f.calls, memoryID)
	return f.neighbors[memoryID], nil
}

func TestSpread_PropagatesThroughGraphWithDecayAndCycles(t *testing.T) {
	provider := &fakeSpreadingNeighborProvider{
		neighbors: map[string][]graphNeighbor{
			"a": {
				{ID: "b", Strength: 10},
			},
			"b": {
				{ID: "a", Strength: 10},
				{ID: "c", Strength: 10},
			},
			"c": {
				{ID: "a", Strength: 10},
			},
		},
	}
	svc := NewSpreadingActivationService(nil, nil)
	svc.neighborProvider = provider

	result, err := svc.Spread(context.Background(), []string{"a"}, []float64{1}, "tenant-graph")
	if err != nil {
		t.Fatalf("Spread() error = %v", err)
	}

	if result.TotalActivated != 2 {
		t.Fatalf("expected 2 activated memories, got %d: %#v", result.TotalActivated, result.Activations)
	}
	if result.Activations[0].MemoryID != "b" {
		t.Fatalf("expected b first, got %#v", result.Activations)
	}
	assertFloatClose(t, result.Activations[0].Activation, 0.5)
	if result.Activations[0].HopsFromSeed != 1 {
		t.Fatalf("b hops = %d, want 1", result.Activations[0].HopsFromSeed)
	}
	if result.Activations[1].MemoryID != "c" {
		t.Fatalf("expected c second, got %#v", result.Activations)
	}
	assertFloatClose(t, result.Activations[1].Activation, 0.25)
	if result.Activations[1].HopsFromSeed != 2 {
		t.Fatalf("c hops = %d, want 2", result.Activations[1].HopsFromSeed)
	}
}

func TestSpread_StrongerPathWinsForExistingActivation(t *testing.T) {
	provider := &fakeSpreadingNeighborProvider{
		neighbors: map[string][]graphNeighbor{
			"a": {
				{ID: "x", Strength: 2},
				{ID: "b", Strength: 10},
			},
			"b": {
				{ID: "x", Strength: 10},
			},
		},
	}
	svc := NewSpreadingActivationService(nil, nil)
	svc.neighborProvider = provider

	result, err := svc.Spread(context.Background(), []string{"a"}, []float64{1}, "tenant-graph")
	if err != nil {
		t.Fatalf("Spread() error = %v", err)
	}

	var x MemoryActivation
	found := false
	for _, activation := range result.Activations {
		if activation.MemoryID == "x" {
			x = activation
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected x activation in %#v", result.Activations)
	}
	assertFloatClose(t, x.Activation, 0.25)
	if x.HopsFromSeed != 2 {
		t.Fatalf("expected stronger two-hop path to win, got hops=%d activation=%f", x.HopsFromSeed, x.Activation)
	}
	if x.PathStrength != 10 {
		t.Fatalf("expected path strength 10, got %f", x.PathStrength)
	}
}

func TestSpread_DropsActivationBelowThreshold(t *testing.T) {
	provider := &fakeSpreadingNeighborProvider{
		neighbors: map[string][]graphNeighbor{
			"a": {
				{ID: "too-weak", Strength: 1},
			},
		},
	}
	svc := NewSpreadingActivationService(nil, nil)
	svc.neighborProvider = provider

	result, err := svc.Spread(context.Background(), []string{"a"}, []float64{0.9}, "tenant-graph")
	if err != nil {
		t.Fatalf("Spread() error = %v", err)
	}
	if result.TotalActivated != 0 {
		t.Fatalf("expected weak activation to be dropped, got %#v", result.Activations)
	}
}

func assertFloatClose(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.000001 {
		t.Fatalf("got %f, want %f", got, want)
	}
}
