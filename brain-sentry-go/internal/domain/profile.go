package domain

import "time"

// UserProfile represents an auto-generated user profile.
type UserProfile struct {
	ID            string         `json:"id" db:"id"`
	TenantID      string         `json:"tenantId" db:"tenant_id"`
	UserID        string         `json:"userId" db:"user_id"`
	StaticProfile StaticProfile  `json:"staticProfile"`
	DynamicProfile DynamicProfile `json:"dynamicProfile"`
	LastUpdatedAt time.Time      `json:"lastUpdatedAt" db:"last_updated_at"`
	Version       int            `json:"version" db:"version"`
}

// StaticProfile contains stable, long-term facts about the user.
type StaticProfile struct {
	Facts       []ProfileFact `json:"facts"`
	Preferences []ProfileFact `json:"preferences"`
	Expertise   []string      `json:"expertise"`
	Summary     string        `json:"summary"`
}

// DynamicProfile contains recent, context-dependent information.
type DynamicProfile struct {
	RecentTopics    []string      `json:"recentTopics"`
	ActiveTasks     []string      `json:"activeTasks"`
	CurrentContext  string        `json:"currentContext"`
	RecentInsights  []string      `json:"recentInsights"`
	GeneratedAt     time.Time     `json:"generatedAt"`
}

// ProfileFact represents a single fact in the user profile.
type ProfileFact struct {
	Key        string `json:"key"`
	Value      string `json:"value"`
	Confidence float64 `json:"confidence"`
	Source     string `json:"source,omitempty"` // memory ID that sourced this fact
}
