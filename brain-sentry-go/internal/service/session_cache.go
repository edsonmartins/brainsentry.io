package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/dto"
)

func buildMemoryReqFromInteraction(sessionID, content string) dto.CreateMemoryRequest {
	return dto.CreateMemoryRequest{
		Content:    content,
		Category:   domain.CategoryContext,
		Importance: domain.ImportanceMinor,
		Tags:       []string{"from-session:" + sessionID},
	}
}

// SessionInteraction is a single Q&A unit stored in the session cache.
type SessionInteraction struct {
	ID        string    `json:"id"`
	Query     string    `json:"query"`
	Response  string    `json:"response"`
	MemoryIDs []string  `json:"memoryIds,omitempty"` // memories used as context
	CreatedAt time.Time `json:"createdAt"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// SessionCacheConfig configures cache behavior.
type SessionCacheConfig struct {
	TTLSeconds       int // how long interactions live in the cache (default 3600)
	MaxPerSession    int // cap per session (default 100)
	KeyPrefix        string // Redis key prefix (default "session_cache:")
}

// DefaultSessionCacheConfig returns sensible defaults.
func DefaultSessionCacheConfig() SessionCacheConfig {
	return SessionCacheConfig{
		TTLSeconds:    3600,
		MaxPerSession: 100,
		KeyPrefix:     "session_cache:",
	}
}

// SessionCacheBackend is the storage abstraction.
// Implementations: RedisSessionBackend, InMemorySessionBackend.
type SessionCacheBackend interface {
	Push(ctx context.Context, sessionID string, interaction SessionInteraction, ttl time.Duration) error
	List(ctx context.Context, sessionID string, limit int) ([]SessionInteraction, error)
	Clear(ctx context.Context, sessionID string) error
	Keys(ctx context.Context) ([]string, error)
}

// SessionMemoryCache stores recent Q&A pairs per session for fast recall.
// Bridges to permanent memory via Cognify().
type SessionMemoryCache struct {
	backend     SessionCacheBackend
	config      SessionCacheConfig
	memoryService *MemoryService // optional — used by Cognify
}

// NewSessionMemoryCache creates a new cache with the given backend.
// Set memoryService to nil to disable Cognify (read-only cache).
func NewSessionMemoryCache(backend SessionCacheBackend, config SessionCacheConfig, memoryService *MemoryService) *SessionMemoryCache {
	if config.TTLSeconds <= 0 {
		config.TTLSeconds = 3600
	}
	if config.MaxPerSession <= 0 {
		config.MaxPerSession = 100
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = "session_cache:"
	}
	return &SessionMemoryCache{
		backend:       backend,
		config:        config,
		memoryService: memoryService,
	}
}

// Push adds a new interaction to the session. Auto-generates ID if empty.
func (c *SessionMemoryCache) Push(ctx context.Context, sessionID string, interaction SessionInteraction) error {
	if sessionID == "" {
		return fmt.Errorf("sessionID is required")
	}
	if interaction.ID == "" {
		interaction.ID = fmt.Sprintf("i-%d", time.Now().UnixNano())
	}
	if interaction.CreatedAt.IsZero() {
		interaction.CreatedAt = time.Now()
	}

	ttl := time.Duration(c.config.TTLSeconds) * time.Second
	return c.backend.Push(ctx, sessionID, interaction, ttl)
}

// Recent returns the N most recent interactions for a session (newest first).
func (c *SessionMemoryCache) Recent(ctx context.Context, sessionID string, limit int) ([]SessionInteraction, error) {
	if limit <= 0 {
		limit = 10
	}
	return c.backend.List(ctx, sessionID, limit)
}

// Clear removes all interactions for a session.
func (c *SessionMemoryCache) Clear(ctx context.Context, sessionID string) error {
	return c.backend.Clear(ctx, sessionID)
}

// CognifyResult summarizes what was persisted.
type CognifyResult struct {
	SessionID   string   `json:"sessionId"`
	Interactions int     `json:"interactions"`
	MemoriesCreated []string `json:"memoriesCreated"`
}

// Cognify converts cached interactions into permanent memories.
// Each Q&A becomes a single memory tagged "from-session:<id>".
func (c *SessionMemoryCache) Cognify(ctx context.Context, sessionID string, clearAfter bool) (*CognifyResult, error) {
	if c.memoryService == nil {
		return nil, fmt.Errorf("memory service is not configured")
	}

	interactions, err := c.backend.List(ctx, sessionID, c.config.MaxPerSession)
	if err != nil {
		return nil, fmt.Errorf("list interactions: %w", err)
	}

	result := &CognifyResult{
		SessionID:    sessionID,
		Interactions: len(interactions),
	}

	for _, it := range interactions {
		content := fmt.Sprintf("Q: %s\nA: %s", it.Query, it.Response)
		memory, err := c.memoryService.CreateMemory(ctx, buildMemoryReqFromInteraction(sessionID, content))
		if err != nil {
			continue
		}
		result.MemoriesCreated = append(result.MemoriesCreated, memory.ID)
	}

	if clearAfter {
		_ = c.backend.Clear(ctx, sessionID)
	}

	return result, nil
}

// ListSessions returns all session IDs with cached data.
func (c *SessionMemoryCache) ListSessions(ctx context.Context) ([]string, error) {
	return c.backend.Keys(ctx)
}

// ------------------ Redis backend ------------------

// RedisSessionBackend implements SessionCacheBackend using Redis lists.
type RedisSessionBackend struct {
	client    *redis.Client
	keyPrefix string
	maxLen    int
}

// NewRedisSessionBackend creates a new Redis-backed session cache.
func NewRedisSessionBackend(client *redis.Client, keyPrefix string, maxLen int) *RedisSessionBackend {
	if keyPrefix == "" {
		keyPrefix = "session_cache:"
	}
	if maxLen <= 0 {
		maxLen = 100
	}
	return &RedisSessionBackend{
		client:    client,
		keyPrefix: keyPrefix,
		maxLen:    maxLen,
	}
}

func (r *RedisSessionBackend) key(sessionID string) string {
	return r.keyPrefix + sessionID
}

// Push prepends an interaction and trims to maxLen.
func (r *RedisSessionBackend) Push(ctx context.Context, sessionID string, interaction SessionInteraction, ttl time.Duration) error {
	data, err := json.Marshal(interaction)
	if err != nil {
		return err
	}

	key := r.key(sessionID)
	pipe := r.client.Pipeline()
	pipe.LPush(ctx, key, data)
	pipe.LTrim(ctx, key, 0, int64(r.maxLen-1))
	pipe.Expire(ctx, key, ttl)
	_, err = pipe.Exec(ctx)
	return err
}

// List returns up to `limit` most recent interactions (newest first).
func (r *RedisSessionBackend) List(ctx context.Context, sessionID string, limit int) ([]SessionInteraction, error) {
	items, err := r.client.LRange(ctx, r.key(sessionID), 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	out := make([]SessionInteraction, 0, len(items))
	for _, raw := range items {
		var it SessionInteraction
		if err := json.Unmarshal([]byte(raw), &it); err == nil {
			out = append(out, it)
		}
	}
	return out, nil
}

// Clear deletes all interactions for a session.
func (r *RedisSessionBackend) Clear(ctx context.Context, sessionID string) error {
	return r.client.Del(ctx, r.key(sessionID)).Err()
}

// Keys returns all session IDs with cached data.
func (r *RedisSessionBackend) Keys(ctx context.Context) ([]string, error) {
	keys, err := r.client.Keys(ctx, r.keyPrefix+"*").Result()
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		if len(k) > len(r.keyPrefix) {
			out = append(out, k[len(r.keyPrefix):])
		}
	}
	return out, nil
}

// ------------------ In-memory backend (fallback) ------------------

// InMemorySessionBackend stores sessions in local memory. Useful when Redis is unavailable.
type InMemorySessionBackend struct {
	mu       sync.RWMutex
	sessions map[string][]SessionInteraction
	maxLen   int
}

// NewInMemorySessionBackend creates a new in-memory backend.
func NewInMemorySessionBackend(maxLen int) *InMemorySessionBackend {
	if maxLen <= 0 {
		maxLen = 100
	}
	return &InMemorySessionBackend{
		sessions: make(map[string][]SessionInteraction),
		maxLen:   maxLen,
	}
}

// Push prepends and trims.
func (m *InMemorySessionBackend) Push(_ context.Context, sessionID string, interaction SessionInteraction, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	list := m.sessions[sessionID]
	list = append([]SessionInteraction{interaction}, list...)
	if len(list) > m.maxLen {
		list = list[:m.maxLen]
	}
	m.sessions[sessionID] = list
	return nil
}

// List returns interactions for a session.
func (m *InMemorySessionBackend) List(_ context.Context, sessionID string, limit int) ([]SessionInteraction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := m.sessions[sessionID]
	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}
	out := make([]SessionInteraction, len(list))
	copy(out, list)
	return out, nil
}

// Clear removes a session.
func (m *InMemorySessionBackend) Clear(_ context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
	return nil
}

// Keys returns all session IDs.
func (m *InMemorySessionBackend) Keys(_ context.Context) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.sessions))
	for k := range m.sessions {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}
