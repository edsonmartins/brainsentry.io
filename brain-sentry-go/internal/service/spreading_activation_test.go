package service

import (
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
