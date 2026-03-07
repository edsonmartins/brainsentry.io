package handler

import (
	"net/http"

	"github.com/integraltech/brainsentry/internal/service"
)

// TasksHandler handles task scheduler endpoints.
type TasksHandler struct {
	scheduler *service.TaskScheduler
}

// NewTasksHandler creates a new TasksHandler.
func NewTasksHandler(scheduler *service.TaskScheduler) *TasksHandler {
	return &TasksHandler{scheduler: scheduler}
}

// Metrics handles GET /v1/tasks/metrics — returns task scheduler metrics.
func (h *TasksHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	processed, failed, recovered := h.scheduler.Metrics()
	writeJSON(w, http.StatusOK, map[string]any{
		"processed": processed,
		"failed":    failed,
		"recovered": recovered,
	})
}

// Pending handles GET /v1/tasks/pending — returns count of pending tasks.
func (h *TasksHandler) Pending(w http.ResponseWriter, r *http.Request) {
	count, err := h.scheduler.PendingCount(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get pending count: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"pending": count,
	})
}
