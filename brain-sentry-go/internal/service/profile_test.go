package service

import (
	"testing"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
)

func TestNewProfileService(t *testing.T) {
	svc := NewProfileService(nil, nil)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestFormatProfileForInjection_Nil(t *testing.T) {
	result := FormatProfileForInjection(nil)
	if result != "" {
		t.Error("expected empty string for nil profile")
	}
}

func TestFormatProfileForInjection_Full(t *testing.T) {
	profile := &domain.UserProfile{
		StaticProfile: domain.StaticProfile{
			Summary: "Senior Go developer",
			Facts: []domain.ProfileFact{
				{Key: "role", Value: "backend engineer", Confidence: 0.9},
			},
			Preferences: []domain.ProfileFact{
				{Key: "language", Value: "Go", Confidence: 0.95},
			},
			Expertise: []string{"Go", "PostgreSQL", "Microservices"},
		},
		DynamicProfile: domain.DynamicProfile{
			RecentTopics:   []string{"memory systems", "LLM integration"},
			ActiveTasks:    []string{"implement user profiles"},
			CurrentContext: "working on Brain Sentry project",
		},
	}

	result := FormatProfileForInjection(profile)
	if result == "" {
		t.Fatal("expected non-empty profile")
	}
	if !containsSubstr(result, "<user_profile>") {
		t.Error("expected user_profile tags")
	}
	if !containsSubstr(result, "Senior Go developer") {
		t.Error("expected summary in output")
	}
	if !containsSubstr(result, "backend engineer") {
		t.Error("expected facts in output")
	}
	if !containsSubstr(result, "Go") {
		t.Error("expected preferences in output")
	}
	if !containsSubstr(result, "PostgreSQL") {
		t.Error("expected expertise in output")
	}
	if !containsSubstr(result, "memory systems") {
		t.Error("expected recent topics in output")
	}
	if !containsSubstr(result, "implement user profiles") {
		t.Error("expected active tasks in output")
	}
}

func TestBuildDynamicProfile(t *testing.T) {
	svc := NewProfileService(nil, nil)
	memories := []domain.Memory{
		{
			Content:    "Working on API integration",
			MemoryType: domain.MemoryTypeTask,
			Tags:       []string{"api", "integration"},
			CreatedAt:  time.Now(),
		},
		{
			Content:    "Debugging auth issue in session",
			Summary:    "Auth debugging session",
			MemoryType: domain.MemoryTypeEpisodic,
			Category:   domain.CategoryBug,
			CreatedAt:  time.Now(),
		},
	}

	profile := svc.buildDynamicProfile(memories)
	if len(profile.ActiveTasks) != 1 {
		t.Errorf("expected 1 active task, got %d", len(profile.ActiveTasks))
	}
	if profile.CurrentContext == "" {
		t.Error("expected non-empty current context from episodic memory")
	}
	if len(profile.RecentTopics) == 0 {
		t.Error("expected non-empty recent topics")
	}
}

func TestBuildStaticProfileFromTypes(t *testing.T) {
	svc := NewProfileService(nil, nil)
	memories := []domain.Memory{
		{
			ID:         "mem-1",
			Content:    "I am a senior backend developer",
			MemoryType: domain.MemoryTypePersonality,
		},
		{
			ID:         "mem-2",
			Content:    "I prefer dark mode and Go language",
			MemoryType: domain.MemoryTypePreference,
		},
		{
			ID:         "mem-3",
			Content:    "PostgreSQL is a relational database",
			MemoryType: domain.MemoryTypeSemantic,
		},
	}

	profile := svc.buildStaticProfileFromTypes(memories)
	if len(profile.Facts) != 1 {
		t.Errorf("expected 1 fact from personality, got %d", len(profile.Facts))
	}
	if len(profile.Preferences) != 1 {
		t.Errorf("expected 1 preference, got %d", len(profile.Preferences))
	}
	if profile.Facts[0].Source != "mem-1" {
		t.Error("expected source mem-1")
	}
}

func TestJoinStrings(t *testing.T) {
	result := joinStrings([]string{"a", "b", "c"}, ", ")
	if result != "a, b, c" {
		t.Errorf("expected 'a, b, c', got '%s'", result)
	}
}

func TestJoinStrings_Empty(t *testing.T) {
	result := joinStrings(nil, ", ")
	if result != "" {
		t.Errorf("expected empty, got '%s'", result)
	}
}

func TestProfileFact_Structure(t *testing.T) {
	fact := domain.ProfileFact{
		Key:        "name",
		Value:      "John",
		Confidence: 0.95,
		Source:     "mem-1",
	}
	if fact.Key != "name" || fact.Confidence != 0.95 {
		t.Error("unexpected fact values")
	}
}

func TestUserProfile_Structure(t *testing.T) {
	profile := domain.UserProfile{
		TenantID: "t1",
		UserID:   "u1",
		Version:  1,
	}
	if profile.TenantID != "t1" {
		t.Error("expected t1")
	}
}
