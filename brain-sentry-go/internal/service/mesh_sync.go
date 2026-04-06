package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// MeshPeer represents a known peer instance for P2P sync.
type MeshPeer struct {
	ID          string   `json:"id"`
	URL         string   `json:"url"`
	SharedScopes []string `json:"sharedScopes"` // e.g., ["memories", "actions"]
	LastSyncAt  *time.Time `json:"lastSyncAt,omitempty"`
	Status      string   `json:"status"` // "active", "unreachable"
}

// MeshSyncConfig holds configuration for mesh synchronization.
type MeshSyncConfig struct {
	SyncIntervalSec int      `json:"syncIntervalSec"`
	AllowedScopes   []string `json:"allowedScopes"` // scopes this instance shares
	MaxPeers        int      `json:"maxPeers"`
}

// DefaultMeshSyncConfig returns default mesh config.
func DefaultMeshSyncConfig() MeshSyncConfig {
	return MeshSyncConfig{
		SyncIntervalSec: 300, // 5 minutes
		AllowedScopes:   []string{"memories", "actions"},
		MaxPeers:        10,
	}
}

// SyncPayload is the data exchanged between peers.
type SyncPayload struct {
	PeerID    string          `json:"peerId"`
	Scope     string          `json:"scope"`
	Items     json.RawMessage `json:"items"`
	Timestamp time.Time       `json:"timestamp"`
}

// MeshSyncResult holds the outcome of a sync operation.
type MeshSyncResult struct {
	PeerID   string `json:"peerId"`
	Scope    string `json:"scope"`
	Sent     int    `json:"sent"`
	Received int    `json:"received"`
	Merged   int    `json:"merged"`
	Error    string `json:"error,omitempty"`
}

// MeshSyncService handles peer-to-peer synchronization between instances.
type MeshSyncService struct {
	mu      sync.RWMutex
	peers   map[string]*MeshPeer
	selfID  string
	config  MeshSyncConfig
	client  *http.Client
}

// NewMeshSyncService creates a new MeshSyncService.
func NewMeshSyncService(selfID string, config MeshSyncConfig) *MeshSyncService {
	return &MeshSyncService{
		peers:  make(map[string]*MeshPeer),
		selfID: selfID,
		config: config,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// RegisterPeer adds a peer for sync. Validates URL to prevent SSRF.
func (s *MeshSyncService) RegisterPeer(_ context.Context, peer MeshPeer) error {
	if err := s.validatePeerURL(peer.URL); err != nil {
		return fmt.Errorf("invalid peer URL: %w", err)
	}

	if peer.ID == s.selfID {
		return fmt.Errorf("cannot register self as peer")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.peers) >= s.config.MaxPeers {
		return fmt.Errorf("max peers reached (%d)", s.config.MaxPeers)
	}

	peer.Status = "active"
	s.peers[peer.ID] = &peer
	slog.Info("mesh peer registered", "peerId", peer.ID, "url", peer.URL)
	return nil
}

// UnregisterPeer removes a peer.
func (s *MeshSyncService) UnregisterPeer(_ context.Context, peerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.peers, peerID)
}

// ListPeers returns all registered peers.
func (s *MeshSyncService) ListPeers() []*MeshPeer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*MeshPeer, 0, len(s.peers))
	for _, p := range s.peers {
		result = append(result, p)
	}
	return result
}

// SyncWithPeer pushes data to a specific peer and receives their data.
func (s *MeshSyncService) SyncWithPeer(ctx context.Context, peerID string, payload SyncPayload) (*MeshSyncResult, error) {
	s.mu.RLock()
	peer, ok := s.peers[peerID]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("peer not found: %s", peerID)
	}

	// Check scope is allowed
	if !s.isScopeAllowed(payload.Scope) {
		return nil, fmt.Errorf("scope %s not allowed by config", payload.Scope)
	}

	payload.PeerID = s.selfID
	payload.Timestamp = time.Now()

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling sync payload: %w", err)
	}

	syncURL := fmt.Sprintf("%s/api/v1/mesh/sync", peer.URL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, syncURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		s.mu.Lock()
		peer.Status = "unreachable"
		s.mu.Unlock()
		return &MeshSyncResult{PeerID: peerID, Scope: payload.Scope, Error: err.Error()}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return &MeshSyncResult{PeerID: peerID, Scope: payload.Scope, Error: string(respBody)}, fmt.Errorf("sync failed: %s", string(respBody))
	}

	var result MeshSyncResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	// Update last sync time
	s.mu.Lock()
	now := time.Now()
	peer.LastSyncAt = &now
	peer.Status = "active"
	s.mu.Unlock()

	return &result, nil
}

// SyncWithAllPeers syncs a scope with all registered peers.
func (s *MeshSyncService) SyncWithAllPeers(ctx context.Context, scope string, items json.RawMessage) []MeshSyncResult {
	s.mu.RLock()
	peerIDs := make([]string, 0, len(s.peers))
	for id := range s.peers {
		peerIDs = append(peerIDs, id)
	}
	s.mu.RUnlock()

	var results []MeshSyncResult
	for _, peerID := range peerIDs {
		result, err := s.SyncWithPeer(ctx, peerID, SyncPayload{
			Scope: scope,
			Items: items,
		})
		if err != nil {
			results = append(results, MeshSyncResult{PeerID: peerID, Scope: scope, Error: err.Error()})
		} else {
			results = append(results, *result)
		}
	}
	return results
}

// validatePeerURL validates the peer URL to prevent SSRF attacks.
func (s *MeshSyncService) validatePeerURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("only http/https schemes allowed")
	}

	host := parsed.Hostname()

	// Block private/loopback IPs
	blockedPrefixes := []string{"127.", "10.", "192.168.", "0.", "169.254."}
	for _, prefix := range blockedPrefixes {
		if len(host) >= len(prefix) && host[:len(prefix)] == prefix {
			return fmt.Errorf("private IP addresses not allowed: %s", host)
		}
	}

	if host == "localhost" || host == "::1" {
		return fmt.Errorf("localhost not allowed")
	}

	// Check 172.16-31.x.x range
	if len(host) >= 4 && host[:4] == "172." {
		var b2 int
		fmt.Sscanf(host[4:], "%d", &b2)
		if b2 >= 16 && b2 <= 31 {
			return fmt.Errorf("private IP range not allowed: %s", host)
		}
	}

	// DNS resolve to check for private IP hiding behind hostname
	ips, err := net.LookupIP(host)
	if err == nil {
		for _, ip := range ips {
			if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
				return fmt.Errorf("hostname %s resolves to private IP %s", host, ip.String())
			}
		}
	}

	return nil
}

func (s *MeshSyncService) isScopeAllowed(scope string) bool {
	for _, allowed := range s.config.AllowedScopes {
		if allowed == scope {
			return true
		}
	}
	return false
}
