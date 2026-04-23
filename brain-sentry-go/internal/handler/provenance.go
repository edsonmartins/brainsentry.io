package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/integraltech/brainsentry/internal/service"
)

// ProvenanceHandler exposes PROV-O exports.
type ProvenanceHandler struct {
	exporter *service.ProvenanceExporter
}

// NewProvenanceHandler builds the handler.
func NewProvenanceHandler(exp *service.ProvenanceExporter) *ProvenanceHandler {
	return &ProvenanceHandler{exporter: exp}
}

// Export handles GET /v1/export/provenance?format=turtle|jsonld&since=...&until=...&limit=500
func (h *ProvenanceHandler) Export(w http.ResponseWriter, r *http.Request) {
	if h.exporter == nil {
		writeError(w, http.StatusServiceUnavailable, "provenance exporter not available")
		return
	}
	q := r.URL.Query()
	opts := service.ExportOptions{Format: q.Get("format")}
	if v := q.Get("since"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			opts.Since = &t
		}
	}
	if v := q.Get("until"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			opts.Until = &t
		}
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			opts.Limit = n
		}
	}

	switch opts.Format {
	case "jsonld":
		w.Header().Set("Content-Type", "application/ld+json")
	default:
		w.Header().Set("Content-Type", "text/turtle")
	}
	if err := h.exporter.ExportPROV(r.Context(), w, opts); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
}
