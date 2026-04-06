package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// ActionStatus represents the lifecycle state of an action.
type ActionStatus string

const (
	ActionPending    ActionStatus = "pending"
	ActionInProgress ActionStatus = "in_progress"
	ActionBlocked    ActionStatus = "blocked"
	ActionCompleted  ActionStatus = "completed"
	ActionCancelled  ActionStatus = "cancelled"
)

// Action represents a workflow item that can be coordinated across agents.
type Action struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      ActionStatus      `json:"status"`
	Priority    int               `json:"priority"` // 1-10
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
	CreatedBy   string            `json:"createdBy"`
	AssignedTo  string            `json:"assignedTo,omitempty"`
	ParentID    string            `json:"parentId,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	DependsOn   []string          `json:"dependsOn,omitempty"` // action IDs this depends on
}

// Lease represents a distributed lock on an action.
type Lease struct {
	ActionID  string    `json:"actionId"`
	HeldBy    string    `json:"heldBy"`
	AcquiredAt time.Time `json:"acquiredAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

// ActionService manages workflow actions and leases for multi-agent coordination.
type ActionService struct {
	mu      sync.RWMutex
	actions map[string]*Action
	leases  map[string]*Lease
	nextID  int
}

// NewActionService creates a new ActionService.
func NewActionService() *ActionService {
	return &ActionService{
		actions: make(map[string]*Action),
		leases:  make(map[string]*Lease),
	}
}

// CreateAction creates a new action.
func (s *ActionService) CreateAction(_ context.Context, title, description, createdBy string, priority int, tags []string, parentID string, dependsOn []string) (*Action, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	now := time.Now()
	action := &Action{
		ID:          fmt.Sprintf("act-%d", s.nextID),
		Title:       title,
		Description: description,
		Status:      ActionPending,
		Priority:    priority,
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   createdBy,
		Tags:        tags,
		ParentID:    parentID,
		DependsOn:   dependsOn,
	}

	s.actions[action.ID] = action
	return action, nil
}

// UpdateStatus updates the status of an action, propagating completion to dependents.
func (s *ActionService) UpdateStatus(_ context.Context, actionID string, status ActionStatus) (*Action, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	action, ok := s.actions[actionID]
	if !ok {
		return nil, fmt.Errorf("action not found: %s", actionID)
	}

	action.Status = status
	action.UpdatedAt = time.Now()

	// If completed, propagate — unblock dependents
	if status == ActionCompleted {
		s.propagateCompletion(actionID)
	}

	return action, nil
}

// propagateCompletion unblocks actions that depended on the completed action.
func (s *ActionService) propagateCompletion(completedID string) {
	for _, action := range s.actions {
		if action.Status != ActionBlocked && action.Status != ActionPending {
			continue
		}

		allDepsComplete := true
		for _, depID := range action.DependsOn {
			dep, ok := s.actions[depID]
			if !ok || dep.Status != ActionCompleted {
				allDepsComplete = false
				break
			}
		}

		if allDepsComplete && len(action.DependsOn) > 0 {
			// Check if this action was blocked by the completed action
			for _, depID := range action.DependsOn {
				if depID == completedID {
					action.Status = ActionPending
					action.UpdatedAt = time.Now()
					slog.Info("action unblocked by dependency completion",
						"actionId", action.ID,
						"completedDep", completedID,
					)
					break
				}
			}
		}
	}
}

// ListActions returns all actions, optionally filtered by status.
func (s *ActionService) ListActions(_ context.Context, statusFilter *ActionStatus) []*Action {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Action
	for _, a := range s.actions {
		if statusFilter != nil && a.Status != *statusFilter {
			continue
		}
		result = append(result, a)
	}
	return result
}

// GetAction returns a single action by ID.
func (s *ActionService) GetAction(_ context.Context, id string) (*Action, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	action, ok := s.actions[id]
	if !ok {
		return nil, fmt.Errorf("action not found: %s", id)
	}
	return action, nil
}

// AcquireLease acquires a distributed lock on an action.
func (s *ActionService) AcquireLease(_ context.Context, actionID, agentID string, ttl time.Duration) (*Lease, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.actions[actionID]; !ok {
		return nil, fmt.Errorf("action not found: %s", actionID)
	}

	if ttl < time.Minute {
		ttl = 10 * time.Minute
	}
	if ttl > time.Hour {
		ttl = time.Hour
	}

	// Check existing lease
	if existing, ok := s.leases[actionID]; ok {
		if time.Now().Before(existing.ExpiresAt) {
			return nil, fmt.Errorf("action %s already leased by %s until %s",
				actionID, existing.HeldBy, existing.ExpiresAt.Format(time.RFC3339))
		}
		// Expired — allow override
	}

	now := time.Now()
	lease := &Lease{
		ActionID:   actionID,
		HeldBy:     agentID,
		AcquiredAt: now,
		ExpiresAt:  now.Add(ttl),
	}
	s.leases[actionID] = lease

	// Update action status
	s.actions[actionID].Status = ActionInProgress
	s.actions[actionID].AssignedTo = agentID
	s.actions[actionID].UpdatedAt = now

	return lease, nil
}

// ReleaseLease releases a lease on an action.
func (s *ActionService) ReleaseLease(_ context.Context, actionID, agentID string, completed bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lease, ok := s.leases[actionID]
	if !ok {
		return fmt.Errorf("no lease found for action %s", actionID)
	}
	if lease.HeldBy != agentID {
		return fmt.Errorf("lease for action %s is held by %s, not %s", actionID, lease.HeldBy, agentID)
	}

	delete(s.leases, actionID)

	action := s.actions[actionID]
	if completed {
		action.Status = ActionCompleted
		s.propagateCompletion(actionID)
	} else {
		action.Status = ActionPending
	}
	action.AssignedTo = ""
	action.UpdatedAt = time.Now()

	return nil
}

// CleanupExpiredLeases releases all expired leases.
func (s *ActionService) CleanupExpiredLeases(_ context.Context) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	cleaned := 0

	for actionID, lease := range s.leases {
		if now.After(lease.ExpiresAt) {
			delete(s.leases, actionID)
			if action, ok := s.actions[actionID]; ok {
				action.Status = ActionPending
				action.AssignedTo = ""
				action.UpdatedAt = now
			}
			cleaned++
			slog.Info("expired lease cleaned up", "actionId", actionID, "heldBy", lease.HeldBy)
		}
	}

	return cleaned
}
