package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func TestDefaultTaskSchedulerConfig(t *testing.T) {
	config := DefaultTaskSchedulerConfig()
	if config.StreamName != "brainsentry:tasks" {
		t.Errorf("expected brainsentry:tasks, got %s", config.StreamName)
	}
	if config.WorkerCount != 3 {
		t.Errorf("expected 3 workers, got %d", config.WorkerCount)
	}
	if config.MaxRetries != 3 {
		t.Errorf("expected 3 max retries, got %d", config.MaxRetries)
	}
}

func TestNewTaskScheduler(t *testing.T) {
	config := DefaultTaskSchedulerConfig()
	sched := NewTaskScheduler(nil, config)
	if sched == nil {
		t.Fatal("expected non-nil scheduler")
	}
	if len(sched.handlers) != 0 {
		t.Error("expected no handlers initially")
	}
}

func TestRegisterHandler(t *testing.T) {
	sched := NewTaskScheduler(nil, DefaultTaskSchedulerConfig())
	called := false
	sched.RegisterHandler(TaskEntityExtraction, func(ctx context.Context, task *AsyncTask) error {
		called = true
		return nil
	})

	if len(sched.handlers) != 1 {
		t.Errorf("expected 1 handler, got %d", len(sched.handlers))
	}

	// Test inline processing
	task, err := sched.Submit(context.Background(), TaskEntityExtraction, "t1", "u1", PriorityNormal, map[string]string{"memoryId": "m1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("handler should have been called inline")
	}
	if task.Status != TaskStatusCompleted {
		t.Errorf("expected completed, got %s", task.Status)
	}
}

func TestSubmitWithoutHandler(t *testing.T) {
	sched := NewTaskScheduler(nil, DefaultTaskSchedulerConfig())
	task, err := sched.Submit(context.Background(), TaskSummarization, "t1", "u1", PriorityHigh, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Type != TaskSummarization {
		t.Errorf("expected summarization, got %s", task.Type)
	}
}

func TestTenantWeights(t *testing.T) {
	sched := NewTaskScheduler(nil, DefaultTaskSchedulerConfig())

	// Default weight
	w := sched.GetTenantWeight("unknown")
	if w != 1.0 {
		t.Errorf("expected default weight 1.0, got %f", w)
	}

	// Set custom weight
	sched.SetTenantWeight("premium", 2.0)
	w = sched.GetTenantWeight("premium")
	if w != 2.0 {
		t.Errorf("expected weight 2.0, got %f", w)
	}
}

func TestTaskPriorities(t *testing.T) {
	if PriorityLow >= PriorityNormal {
		t.Error("low should be less than normal")
	}
	if PriorityNormal >= PriorityHigh {
		t.Error("normal should be less than high")
	}
	if PriorityHigh >= PriorityCritical {
		t.Error("high should be less than critical")
	}
}

func TestAsyncTask_Structure(t *testing.T) {
	payload, _ := json.Marshal(map[string]string{"key": "value"})
	task := AsyncTask{
		ID:       "t1",
		Type:     TaskReflection,
		TenantID: "tenant1",
		Priority: PriorityHigh,
		Payload:  payload,
		Status:   TaskStatusPending,
	}

	if task.Type != TaskReflection {
		t.Error("expected reflection type")
	}
	if task.Priority != PriorityHigh {
		t.Error("expected high priority")
	}
	if task.Status != TaskStatusPending {
		t.Error("expected pending status")
	}

	// Verify payload is valid JSON
	var p map[string]string
	if err := json.Unmarshal(task.Payload, &p); err != nil {
		t.Errorf("invalid payload JSON: %v", err)
	}
	if p["key"] != "value" {
		t.Error("expected key=value in payload")
	}
}

func TestSubmitInlineFailed(t *testing.T) {
	sched := NewTaskScheduler(nil, DefaultTaskSchedulerConfig())
	sched.RegisterHandler(TaskEmbedding, func(ctx context.Context, task *AsyncTask) error {
		return fmt.Errorf("embedding service unavailable")
	})

	task, err := sched.Submit(context.Background(), TaskEmbedding, "t1", "u1", PriorityNormal, nil)
	if err == nil {
		t.Fatal("expected error from failed handler")
	}
	if task.Status != TaskStatusFailed {
		t.Errorf("expected failed status, got %s", task.Status)
	}
	if task.Error == "" {
		t.Error("expected error message")
	}
}

func TestMetrics(t *testing.T) {
	sched := NewTaskScheduler(nil, DefaultTaskSchedulerConfig())
	processed, failed, recovered := sched.Metrics()
	if processed != 0 || failed != 0 || recovered != 0 {
		t.Error("expected all metrics to be 0 initially")
	}
}

func TestPendingCount_NilClient(t *testing.T) {
	sched := NewTaskScheduler(nil, DefaultTaskSchedulerConfig())
	count, err := sched.PendingCount(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 pending, got %d", count)
	}
}

func TestTaskTypes(t *testing.T) {
	types := []TaskType{
		TaskEntityExtraction,
		TaskSummarization,
		TaskReflection,
		TaskReconciliation,
		TaskEmbedding,
		TaskGraphUpdate,
		TaskDecayCleanup,
		TaskProfileUpdate,
	}
	seen := make(map[TaskType]bool)
	for _, tt := range types {
		if seen[tt] {
			t.Errorf("duplicate task type: %s", tt)
		}
		seen[tt] = true
	}
	if len(types) != 8 {
		t.Errorf("expected 8 task types, got %d", len(types))
	}
}

func TestStartStop_NilClient(t *testing.T) {
	sched := NewTaskScheduler(nil, DefaultTaskSchedulerConfig())
	err := sched.Start(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sched.Stop()
}
