package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/integraltech/brainsentry/internal/dto"
	"github.com/integraltech/brainsentry/internal/service"
)

// NoteTakingHandler handles note-taking endpoints.
type NoteTakingHandler struct {
	noteTakingService *service.NoteTakingService
}

// NewNoteTakingHandler creates a new NoteTakingHandler.
func NewNoteTakingHandler(noteTakingService *service.NoteTakingService) *NoteTakingHandler {
	return &NoteTakingHandler{noteTakingService: noteTakingService}
}

// AnalyzeSession handles POST /v1/notes/analyze
func (h *NoteTakingHandler) AnalyzeSession(w http.ResponseWriter, r *http.Request) {
	var req dto.SessionAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SessionID == "" {
		writeError(w, http.StatusBadRequest, "sessionId is required")
		return
	}

	resp, err := h.noteTakingService.AnalyzeSession(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "session analysis failed")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// CreateHindsight handles POST /v1/notes/hindsight
func (h *NoteTakingHandler) CreateHindsight(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateHindsightNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SessionID == "" || req.ErrorType == "" || req.ErrorMessage == "" {
		writeError(w, http.StatusBadRequest, "sessionId, errorType, and errorMessage are required")
		return
	}

	note, err := h.noteTakingService.CreateHindsightNote(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create hindsight note")
		return
	}

	writeJSON(w, http.StatusCreated, note)
}

// ListNotes handles GET /v1/notes
func (h *NoteTakingHandler) ListNotes(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	notes, err := h.noteTakingService.ListNotes(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list notes")
		return
	}
	writeJSON(w, http.StatusOK, notes)
}

// ListHindsight handles GET /v1/notes/hindsight
func (h *NoteTakingHandler) ListHindsight(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	notes, err := h.noteTakingService.ListHindsightNotes(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list hindsight notes")
		return
	}
	writeJSON(w, http.StatusOK, notes)
}

// GetSessionNotes handles GET /v1/notes/session/{sessionId}
func (h *NoteTakingHandler) GetSessionNotes(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	notes, err := h.noteTakingService.GetSessionNotes(r.Context(), sessionID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get session notes")
		return
	}
	writeJSON(w, http.StatusOK, notes)
}

// GetSessionHindsight handles GET /v1/notes/session/{sessionId}/hindsight
func (h *NoteTakingHandler) GetSessionHindsight(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	notes, err := h.noteTakingService.GetSessionHindsight(r.Context(), sessionID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get session hindsight")
		return
	}
	writeJSON(w, http.StatusOK, notes)
}
