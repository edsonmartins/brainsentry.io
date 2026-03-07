package client

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/integraltech/brainsentry/internal/dto"
)

func setupTestServer(handler http.Handler) (*httptest.Server, *Client) {
	srv := httptest.NewServer(handler)
	c := New(srv.URL, "test-tenant")
	return srv, c
}

func TestLogin(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("X-Tenant-ID") != "test-tenant" {
			t.Errorf("expected tenant test-tenant, got %s", r.Header.Get("X-Tenant-ID"))
		}

		var req dto.LoginRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Email != "test@example.com" {
			t.Errorf("expected email test@example.com, got %s", req.Email)
		}

		json.NewEncoder(w).Encode(dto.LoginResponse{
			Token:    "test-token-123",
			TenantID: "test-tenant",
			User:     dto.UserResponse{ID: "u1", Email: "test@example.com"},
		})
	})

	srv, c := setupTestServer(mux)
	defer srv.Close()

	resp, err := c.Login("test@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token != "test-token-123" {
		t.Errorf("expected token test-token-123, got %s", resp.Token)
	}
	if c.Token() != "test-token-123" {
		t.Errorf("expected client token to be set")
	}
	if !c.IsAuthenticated() {
		t.Error("expected client to be authenticated")
	}
}

func TestLoginDemo(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/auth/demo", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(dto.LoginResponse{
			Token:    "demo-token",
			TenantID: "default",
			User:     dto.UserResponse{ID: "demo", Email: "demo@example.com"},
		})
	})

	srv, c := setupTestServer(mux)
	defer srv.Close()

	resp, err := c.LoginDemo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token != "demo-token" {
		t.Errorf("expected demo-token, got %s", resp.Token)
	}
}

func TestListMemories(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/memories", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer my-token" {
			t.Errorf("expected auth header, got %s", r.Header.Get("Authorization"))
		}
		page := r.URL.Query().Get("page")
		size := r.URL.Query().Get("size")
		if page != "1" || size != "10" {
			t.Errorf("expected page=1&size=10, got page=%s&size=%s", page, size)
		}

		json.NewEncoder(w).Encode(dto.MemoryListResponse{
			Memories:      []dto.MemoryResponse{{ID: "m1", Content: "test memory"}},
			Page:          1,
			Size:          10,
			TotalElements: 1,
			TotalPages:    1,
		})
	})

	srv, c := setupTestServer(mux)
	defer srv.Close()
	c.SetToken("my-token")

	resp, err := c.ListMemories(1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Memories) != 1 {
		t.Errorf("expected 1 memory, got %d", len(resp.Memories))
	}
	if resp.Memories[0].ID != "m1" {
		t.Errorf("expected memory ID m1, got %s", resp.Memories[0].ID)
	}
}

func TestCreateMemory(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/memories", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req dto.CreateMemoryRequest
		json.NewDecoder(r.Body).Decode(&req)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(dto.MemoryResponse{
			ID:      "new-id",
			Content: req.Content,
		})
	})

	srv, c := setupTestServer(mux)
	defer srv.Close()
	c.SetToken("token")

	resp, err := c.CreateMemory(&dto.CreateMemoryRequest{Content: "new memory"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "new-id" {
		t.Errorf("expected ID new-id, got %s", resp.ID)
	}
}

func TestDeleteMemory(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/memories/del-id", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	srv, c := setupTestServer(mux)
	defer srv.Close()
	c.SetToken("token")

	err := c.DeleteMemory("del-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSearchMemories(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/memories/search", func(w http.ResponseWriter, r *http.Request) {
		var req dto.SearchRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Query != "test query" {
			t.Errorf("expected query 'test query', got '%s'", req.Query)
		}

		json.NewEncoder(w).Encode(dto.SearchResponse{
			Results:      []dto.MemoryResponse{{ID: "s1", Content: "search result"}},
			Total:        1,
			SearchTimeMs: 42,
		})
	})

	srv, c := setupTestServer(mux)
	defer srv.Close()
	c.SetToken("token")

	resp, err := c.SearchMemories(&dto.SearchRequest{Query: "test query", Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("expected 1 result, got %d", resp.Total)
	}
}

func TestGetStats(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/stats/overview", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(dto.StatsResponse{
			TotalMemories:     42,
			ActiveMemories24h: 10,
		})
	})

	srv, c := setupTestServer(mux)
	defer srv.Close()
	c.SetToken("token")

	resp, err := c.GetStats()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalMemories != 42 {
		t.Errorf("expected 42 memories, got %d", resp.TotalMemories)
	}
}

func TestAPIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/memories/bad-id", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "not found",
			"message": "memory not found",
		})
	})

	srv, c := setupTestServer(mux)
	defer srv.Close()
	c.SetToken("token")

	_, err := c.GetMemory("bad-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
}

func TestRetryOn429(t *testing.T) {
	attempts := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/stats/overview", func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		json.NewEncoder(w).Encode(dto.StatsResponse{TotalMemories: 1})
	})

	srv, c := setupTestServer(mux)
	defer srv.Close()
	c.SetToken("token")

	resp, err := c.GetStats()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalMemories != 1 {
		t.Errorf("expected 1, got %d", resp.TotalMemories)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}
