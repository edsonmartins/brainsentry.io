package handler

import (
	"encoding/json"
	"net/http"

	"github.com/integraltech/brainsentry/internal/service"
)

type AutoForgetHandler struct {
	svc *service.AutoForgetService
}

func NewAutoForgetHandler(svc *service.AutoForgetService) *AutoForgetHandler {
	return &AutoForgetHandler{svc: svc}
}

func (h *AutoForgetHandler) Run(w http.ResponseWriter, r *http.Request) {
	dryRun := r.URL.Query().Get("dryRun") == "true"

	result, err := h.svc.Run(r.Context(), dryRun)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
