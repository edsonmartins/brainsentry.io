package service

import (
	"context"
	"testing"
)

func TestMeshSync_RegisterPeer(t *testing.T) {
	svc := NewMeshSyncService("self-1", DefaultMeshSyncConfig())

	err := svc.RegisterPeer(context.Background(), MeshPeer{
		ID:           "peer-1",
		URL:          "https://example.com",
		SharedScopes: []string{"memories"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	peers := svc.ListPeers()
	if len(peers) != 1 {
		t.Errorf("expected 1 peer, got %d", len(peers))
	}
}

func TestMeshSync_RejectSelfRegistration(t *testing.T) {
	svc := NewMeshSyncService("self-1", DefaultMeshSyncConfig())

	err := svc.RegisterPeer(context.Background(), MeshPeer{
		ID:  "self-1",
		URL: "https://example.com",
	})
	if err == nil {
		t.Error("expected error when registering self")
	}
}

func TestMeshSync_RejectPrivateIPs(t *testing.T) {
	svc := NewMeshSyncService("self-1", DefaultMeshSyncConfig())

	tests := []struct {
		url  string
		name string
	}{
		{"http://localhost:8080", "localhost"},
		{"http://127.0.0.1:8080", "loopback"},
		{"http://192.168.1.1:8080", "private 192.168"},
		{"http://10.0.0.1:8080", "private 10.x"},
	}

	for _, tt := range tests {
		err := svc.RegisterPeer(context.Background(), MeshPeer{
			ID:  "peer-" + tt.name,
			URL: tt.url,
		})
		if err == nil {
			t.Errorf("%s: expected rejection of private URL %s", tt.name, tt.url)
		}
	}
}

func TestMeshSync_MaxPeers(t *testing.T) {
	config := DefaultMeshSyncConfig()
	config.MaxPeers = 2
	svc := NewMeshSyncService("self-1", config)

	svc.RegisterPeer(context.Background(), MeshPeer{ID: "p1", URL: "https://a.example.com"})
	svc.RegisterPeer(context.Background(), MeshPeer{ID: "p2", URL: "https://b.example.com"})

	err := svc.RegisterPeer(context.Background(), MeshPeer{ID: "p3", URL: "https://c.example.com"})
	if err == nil {
		t.Error("expected error when max peers reached")
	}
}

func TestMeshSync_UnregisterPeer(t *testing.T) {
	svc := NewMeshSyncService("self-1", DefaultMeshSyncConfig())
	svc.RegisterPeer(context.Background(), MeshPeer{ID: "p1", URL: "https://a.example.com"})

	svc.UnregisterPeer(context.Background(), "p1")

	peers := svc.ListPeers()
	if len(peers) != 0 {
		t.Errorf("expected 0 peers, got %d", len(peers))
	}
}

func TestMeshSync_ScopeValidation(t *testing.T) {
	config := DefaultMeshSyncConfig()
	config.AllowedScopes = []string{"memories"}
	svc := NewMeshSyncService("self-1", config)

	if svc.isScopeAllowed("memories") != true {
		t.Error("memories should be allowed")
	}
	if svc.isScopeAllowed("secrets") != false {
		t.Error("secrets should not be allowed")
	}
}

func TestDefaultMeshSyncConfig(t *testing.T) {
	cfg := DefaultMeshSyncConfig()
	if cfg.SyncIntervalSec != 300 {
		t.Errorf("expected 300s interval, got %d", cfg.SyncIntervalSec)
	}
	if cfg.MaxPeers != 10 {
		t.Errorf("expected 10 max peers, got %d", cfg.MaxPeers)
	}
	if len(cfg.AllowedScopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(cfg.AllowedScopes))
	}
}
