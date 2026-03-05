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
