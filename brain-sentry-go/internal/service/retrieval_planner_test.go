package service

import (
	"strings"
	"testing"
)

func TestNewRetrievalPlannerService(t *testing.T) {
	svc := NewRetrievalPlannerService(nil, nil, nil, nil)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.maxRounds != 3 {
		t.Errorf("expected 3 max rounds, got %d", svc.maxRounds)
	}
	if svc.coverageTarget != 0.8 {
		t.Errorf("expected 0.8 coverage target, got %f", svc.coverageTarget)
	}
}

func TestInfoNeed_Structure(t *testing.T) {
	need := InfoNeed{
		Type:        "factual",
		Description: "what database is used",
		Satisfied:   false,
	}
	if need.Type != "factual" {
		t.Error("expected factual")
	}
}

func TestSubQuery_Structure(t *testing.T) {
	sq := SubQuery{
		Query:    "PostgreSQL database",
		Purpose:  "factual",
		ViewType: "semantic",
	}
	if sq.ViewType != "semantic" {
		t.Error("expected semantic")
	}
}

func TestPlanResult_Structure(t *testing.T) {
	r := PlanResult{
		MemoryID: "mem-1",
		Content:  "test content",
		Score:    0.9,
		Source:   "semantic:factual",
		Round:    1,
	}
	if r.Score != 0.9 {
		t.Error("expected 0.9 score")
	}
}

func TestSortPlanResults(t *testing.T) {
	results := []PlanResult{
		{MemoryID: "a", Score: 0.3},
		{MemoryID: "b", Score: 0.9},
		{MemoryID: "c", Score: 0.6},
	}
	sortPlanResults(results)
	if results[0].MemoryID != "b" {
		t.Errorf("expected 'b' first, got %s", results[0].MemoryID)
	}
	if results[2].MemoryID != "a" {
		t.Errorf("expected 'a' last, got %s", results[2].MemoryID)
	}
}

func TestContainsSubstr_Planner(t *testing.T) {
	if !strings.Contains("hello world", "world") {
		t.Error("expected true")
	}
	if strings.Contains("hello", "world") {
		t.Error("expected false")
	}
}

func TestRetrievalPlan_Defaults(t *testing.T) {
	plan := &RetrievalPlan{
		OriginalQuery: "test query",
		Results:       make([]PlanResult, 0),
	}
	if plan.Rounds != 0 {
		t.Error("expected 0 rounds initially")
	}
	if plan.Coverage != 0 {
		t.Error("expected 0 coverage initially")
	}
}

func TestEvaluateCoverage_EmptyNeeds(t *testing.T) {
	svc := NewRetrievalPlannerService(nil, nil, nil, nil)
	plan := &RetrievalPlan{InfoNeeds: nil}
	coverage, _ := svc.evaluateCoverage(nil, plan)
	if coverage != 1.0 {
		t.Errorf("expected 1.0 coverage for empty needs, got %f", coverage)
	}
}

func TestGenerateGapQueries_AllSatisfied(t *testing.T) {
	svc := NewRetrievalPlannerService(nil, nil, nil, nil)
	plan := &RetrievalPlan{
		InfoNeeds: []InfoNeed{
			{Type: "factual", Satisfied: true},
			{Type: "procedural", Satisfied: true},
		},
	}
	gaps := svc.generateGapQueries(nil, plan)
	if len(gaps) != 0 {
		t.Errorf("expected 0 gaps, got %d", len(gaps))
	}
}

func TestGenerateGapQueries_UnsatisfiedNeeds(t *testing.T) {
	svc := NewRetrievalPlannerService(nil, nil, nil, nil)
	plan := &RetrievalPlan{
		InfoNeeds: []InfoNeed{
			{Type: "factual", Description: "database info", Satisfied: true},
			{Type: "procedural", Description: "how to deploy", Satisfied: false},
		},
	}
	gaps := svc.generateGapQueries(nil, plan)
	if len(gaps) != 1 {
		t.Errorf("expected 1 gap, got %d", len(gaps))
	}
	if gaps[0].Query != "how to deploy" {
		t.Errorf("expected gap query 'how to deploy', got %s", gaps[0].Query)
	}
}

// --- Extended retrieval planner tests ---

func TestEvaluateCoverage_SingleNeed_SourceMatch(t *testing.T) {
	svc := NewRetrievalPlannerService(nil, nil, nil, nil)
	plan := &RetrievalPlan{
		InfoNeeds: []InfoNeed{
			{Type: "factual", Description: "database info", Satisfied: false},
		},
		Results: []PlanResult{
			{MemoryID: "m1", Content: "PostgreSQL database", Source: "semantic:factual", Score: 0.9},
		},
	}
	coverage, satisfied := svc.evaluateCoverage(nil, plan)
	if coverage < 0 || coverage > 1 {
		t.Errorf("coverage should be 0-1, got %f", coverage)
	}
	_ = satisfied
}

func TestEvaluateCoverage_MultipleNeeds_PartialCoverage(t *testing.T) {
	svc := NewRetrievalPlannerService(nil, nil, nil, nil)
	plan := &RetrievalPlan{
		InfoNeeds: []InfoNeed{
			{Type: "factual", Description: "database info"},
			{Type: "procedural", Description: "deploy steps"},
			{Type: "conceptual", Description: "architecture overview"},
			{Type: "preference", Description: "coding style"},
		},
		Results: []PlanResult{
			{MemoryID: "m1", Content: "database configuration", Source: "semantic:factual", Score: 0.9},
			{MemoryID: "m2", Content: "architecture design patterns overview", Source: "semantic:conceptual", Score: 0.8},
		},
	}
	coverage, _ := svc.evaluateCoverage(nil, plan)
	if coverage != 0.5 {
		t.Errorf("expected coverage 0.5 (2/4 needs matched by results), got %f", coverage)
	}
}

func TestGenerateGapQueries_Empty(t *testing.T) {
	svc := NewRetrievalPlannerService(nil, nil, nil, nil)
	plan := &RetrievalPlan{InfoNeeds: nil}
	gaps := svc.generateGapQueries(nil, plan)
	if len(gaps) != 0 {
		t.Errorf("expected 0 gaps, got %d", len(gaps))
	}
}

func TestGenerateGapQueries_MultipleUnsatisfied(t *testing.T) {
	svc := NewRetrievalPlannerService(nil, nil, nil, nil)
	plan := &RetrievalPlan{
		InfoNeeds: []InfoNeed{
			{Type: "factual", Description: "database info", Satisfied: false},
			{Type: "procedural", Description: "how to deploy", Satisfied: false},
			{Type: "conceptual", Description: "architecture", Satisfied: true},
		},
	}
	gaps := svc.generateGapQueries(nil, plan)
	if len(gaps) != 2 {
		t.Errorf("expected 2 gaps, got %d", len(gaps))
	}
}

func TestSortPlanResults_Empty(t *testing.T) {
	sortPlanResults(nil) // should not panic
}

func TestSortPlanResults_AlreadySorted(t *testing.T) {
	results := []PlanResult{
		{MemoryID: "a", Score: 0.9},
		{MemoryID: "b", Score: 0.6},
		{MemoryID: "c", Score: 0.3},
	}
	sortPlanResults(results)
	if results[0].MemoryID != "a" {
		t.Error("expected 'a' first")
	}
}

func TestSortPlanResults_AllSameScore(t *testing.T) {
	results := []PlanResult{
		{MemoryID: "a", Score: 0.5},
		{MemoryID: "b", Score: 0.5},
		{MemoryID: "c", Score: 0.5},
	}
	sortPlanResults(results) // should not panic
	if len(results) != 3 {
		t.Error("expected 3 results")
	}
}

func TestRetrievalPlannerService_MaxRounds(t *testing.T) {
	svc := NewRetrievalPlannerService(nil, nil, nil, nil)
	if svc.maxRounds != 3 {
		t.Errorf("expected maxRounds=3, got %d", svc.maxRounds)
	}
}

func TestRetrievalPlannerService_CoverageTarget(t *testing.T) {
	svc := NewRetrievalPlannerService(nil, nil, nil, nil)
	if svc.coverageTarget != 0.8 {
		t.Errorf("expected coverageTarget=0.8, got %f", svc.coverageTarget)
	}
}
