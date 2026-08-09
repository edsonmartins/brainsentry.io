package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOptionalHandlersFailClosedWhenServiceUnavailable(t *testing.T) {
	events := NewEventHandler(nil)
	policies := NewPolicyHandler(nil, nil)
	semantic := NewSemanticAPIHandler(nil)
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		handle http.HandlerFunc
	}{
		{"event record", http.MethodPost, "/events", `{}`, events.Record},
		{"event get", http.MethodGet, "/events/id", "", events.Get},
		{"event list", http.MethodGet, "/events", "", events.List},
		{"event delete", http.MethodDelete, "/events/id", "", events.Delete},
		{"event stats", http.MethodGet, "/events/stats", "", events.Stats},
		{"event extract", http.MethodPost, "/events/extract", `{}`, events.Extract},
		{"policy list", http.MethodGet, "/policies", "", policies.List},
		{"policy get", http.MethodGet, "/policies/id", "", policies.Get},
		{"policy create", http.MethodPost, "/policies", `{}`, policies.Create},
		{"policy update", http.MethodPut, "/policies/id", `{}`, policies.Update},
		{"policy delete", http.MethodDelete, "/policies/id", "", policies.Delete},
		{"policy enforce", http.MethodPost, "/policies/enforce", `{}`, policies.EnforceOnDecision},
		{"semantic remember", http.MethodPost, "/remember", `{}`, semantic.Remember},
		{"semantic recall", http.MethodPost, "/recall", `{}`, semantic.Recall},
		{"semantic improve", http.MethodPost, "/improve", `{}`, semantic.Improve},
		{"semantic forget", http.MethodPost, "/forget", `{}`, semantic.Forget},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			rr := httptest.NewRecorder()
			tc.handle(rr, req)
			if rr.Code != http.StatusServiceUnavailable {
				t.Fatalf("expected 503, got %d: %s", rr.Code, rr.Body.String())
			}
		})
	}
}

func TestActionsHandlerRejectsMalformedMutationBodies(t *testing.T) {
	h := NewActionsHandler(nil)
	tests := []struct {
		name   string
		handle http.HandlerFunc
	}{
		{"create", h.Create},
		{"update status", h.UpdateStatus},
		{"acquire lease", h.AcquireLease},
		{"release lease", h.ReleaseLease},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/actions/id", strings.NewReader(`{broken`))
			rr := httptest.NewRecorder()
			tc.handle(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", rr.Code)
			}
		})
	}
}

func TestAPIKeyHandlerRequiresTenantPathParameter(t *testing.T) {
	h := NewAPIKeyHandler(nil)
	for name, handle := range map[string]http.HandlerFunc{"create": h.Create, "list": h.List, "revoke": h.Revoke} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api-keys", strings.NewReader(`{}`))
			rr := httptest.NewRecorder()
			handle(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", rr.Code)
			}
		})
	}
}
