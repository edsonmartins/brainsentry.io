package service

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// AgentTraceConfig controls retention and limits.
type AgentTraceConfig struct {
	MaxPerTenant int // cap to prevent unbounded growth; 0 = unlimited
}

// DefaultAgentTraceConfig returns sensible defaults.
func DefaultAgentTraceConfig() AgentTraceConfig {
	return AgentTraceConfig{
		MaxPerTenant: 10000,
	}
}

// AgentTraceService stores and retrieves AgentTrace entries.
// In-memory for now — can be swapped with a Postgres-backed implementation later.
type AgentTraceService struct {
	mu     sync.RWMutex
	traces map[string][]*domain.AgentTrace // keyed by tenantID
	config AgentTraceConfig
}

// NewAgentTraceService creates a new AgentTraceService.
func NewAgentTraceService(config AgentTraceConfig) *AgentTraceService {
	return &AgentTraceService{
		traces: make(map[string][]*domain.AgentTrace),
		config: config,
	}
}

// RecordTraceRequest is the input for recording a new trace.
type RecordTraceRequest struct {
	SessionID      string
	AgentID        string
	OriginFunction string
	WithMemory     bool
	MemoryQuery    string
	MethodParams   map[string]any
	MethodReturn   any
	MemoryContext  string
	MemoryIDs      []string
	Status         domain.AgentTraceStatus
	ErrorMessage   string
	DurationMs     int64
	BelongsToSets  []string
}

// Record creates and stores a new AgentTrace.
func (s *AgentTraceService) Record(ctx context.Context, req RecordTraceRequest) (*domain.AgentTrace, error) {
	if req.OriginFunction == "" {
		return nil, fmt.Errorf("originFunction is required")
	}
	if req.Status == "" {
		req.Status = domain.AgentTraceSuccess
	}

	tenantID := tenant.FromContext(ctx)

	trace := &domain.AgentTrace{
		ID:             uuid.NewString(),
		TenantID:       tenantID,
		SessionID:      req.SessionID,
		AgentID:        req.AgentID,
		OriginFunction: req.OriginFunction,
		WithMemory:     req.WithMemory,
		MemoryQuery:    req.MemoryQuery,
		MethodParams:   req.MethodParams,
		MethodReturn:   req.MethodReturn,
		MemoryContext:  req.MemoryContext,
		Status:         req.Status,
		ErrorMessage:   req.ErrorMessage,
		DurationMs:     req.DurationMs,
		MemoryIDs:      req.MemoryIDs,
		BelongsToSets:  req.BelongsToSets,
		CreatedAt:      time.Now(),
	}
	trace.Text = s.buildEmbeddableText(trace)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.traces[tenantID] = append(s.traces[tenantID], trace)

	// Enforce cap: drop oldest
	if s.config.MaxPerTenant > 0 && len(s.traces[tenantID]) > s.config.MaxPerTenant {
		overflow := len(s.traces[tenantID]) - s.config.MaxPerTenant
		s.traces[tenantID] = s.traces[tenantID][overflow:]
	}

	return trace, nil
}

// buildEmbeddableText creates a concise summary for semantic search.
func (s *AgentTraceService) buildEmbeddableText(t *domain.AgentTrace) string {
	if t.Status == domain.AgentTraceError {
		return fmt.Sprintf("%s failed: %s (query=%q)", t.OriginFunction, t.ErrorMessage, t.MemoryQuery)
	}
	if t.WithMemory && t.MemoryQuery != "" {
		return fmt.Sprintf("%s used memory query %q (%d context chars)",
			t.OriginFunction, t.MemoryQuery, len(t.MemoryContext))
	}
	return fmt.Sprintf("%s completed successfully", t.OriginFunction)
}

// Get returns a single trace by ID within the tenant.
func (s *AgentTraceService) Get(ctx context.Context, id string) (*domain.AgentTrace, error) {
	tenantID := tenant.FromContext(ctx)
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, t := range s.traces[tenantID] {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf("trace not found: %s", id)
}

// ListFilter narrows which traces to return.
type ListFilter struct {
	SessionID string
	AgentID   string
	Status    domain.AgentTraceStatus
	Set       string // only include traces belonging to this set
	Limit     int    // max results (0 = all, default 100)
}

// List returns traces matching the filter, newest first.
func (s *AgentTraceService) List(ctx context.Context, filter ListFilter) []*domain.AgentTrace {
	tenantID := tenant.FromContext(ctx)
	s.mu.RLock()
	defer s.mu.RUnlock()

	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}

	all := s.traces[tenantID]
	filtered := make([]*domain.AgentTrace, 0, len(all))
	for _, t := range all {
		if filter.SessionID != "" && t.SessionID != filter.SessionID {
			continue
		}
		if filter.AgentID != "" && t.AgentID != filter.AgentID {
			continue
		}
		if filter.Status != "" && t.Status != filter.Status {
			continue
		}
		if filter.Set != "" && !containsString(t.BelongsToSets, filter.Set) {
			continue
		}
		filtered = append(filtered, t)
	}

	// Newest first
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	if len(filtered) > limit {
		filtered = filtered[:limit]
	}
	return filtered
}

// Delete removes a trace by ID.
func (s *AgentTraceService) Delete(ctx context.Context, id string) error {
	tenantID := tenant.FromContext(ctx)
	s.mu.Lock()
	defer s.mu.Unlock()

	traces := s.traces[tenantID]
	for i, t := range traces {
		if t.ID == id {
			s.traces[tenantID] = append(traces[:i], traces[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("trace not found: %s", id)
}

// Stats returns aggregate statistics for a tenant's traces.
type TraceStats struct {
	Total        int     `json:"total"`
	Success      int     `json:"success"`
	Errors       int     `json:"errors"`
	WithMemory   int     `json:"withMemory"`
	AvgDurationMs float64 `json:"avgDurationMs"`
	ErrorRate    float64 `json:"errorRate"`
}

// Stats computes aggregate statistics.
func (s *AgentTraceService) Stats(ctx context.Context) TraceStats {
	tenantID := tenant.FromContext(ctx)
	s.mu.RLock()
	defer s.mu.RUnlock()

	all := s.traces[tenantID]
	stats := TraceStats{Total: len(all)}
	if stats.Total == 0 {
		return stats
	}

	var totalDuration int64
	for _, t := range all {
		if t.Status == domain.AgentTraceSuccess {
			stats.Success++
		} else {
			stats.Errors++
		}
		if t.WithMemory {
			stats.WithMemory++
		}
		totalDuration += t.DurationMs
	}

	stats.AvgDurationMs = float64(totalDuration) / float64(stats.Total)
	stats.ErrorRate = float64(stats.Errors) / float64(stats.Total)
	return stats
}

func containsString(ss []string, target string) bool {
	for _, s := range ss {
		if s == target {
			return true
		}
	}
	return false
}
