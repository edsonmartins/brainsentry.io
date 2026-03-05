package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// ProfileService generates and manages user profiles from accumulated memories.
type ProfileService struct {
	openRouter *OpenRouterService
	memoryRepo *postgres.MemoryRepository
}

// NewProfileService creates a new ProfileService.
func NewProfileService(
	openRouter *OpenRouterService,
	memoryRepo *postgres.MemoryRepository,
) *ProfileService {
	return &ProfileService{
		openRouter: openRouter,
		memoryRepo: memoryRepo,
	}
}

// GenerateProfile builds a user profile from the tenant's memories.
// Static profile from PERSONALITY/PREFERENCE/SEMANTIC memories.
// Dynamic profile from recent EPISODIC/THREAD/TASK memories.
func (s *ProfileService) GenerateProfile(ctx context.Context, userID string) (*domain.UserProfile, error) {
	tenantID := tenant.FromContext(ctx)

	profile := &domain.UserProfile{
		TenantID:      tenantID,
		UserID:        userID,
		LastUpdatedAt: time.Now(),
		Version:       1,
	}

	// Fetch all memories for the tenant
	allMemories, err := s.memoryRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching memories for profile: %w", err)
	}

	if len(allMemories) == 0 {
		return profile, nil
	}

	// Split memories by type for static vs dynamic
	var stableMemories, recentMemories []domain.Memory
	now := time.Now()
	recentThreshold := now.Add(-7 * 24 * time.Hour) // last 7 days

	for _, m := range allMemories {
		if IsExpired(&m, now) || m.SupersededBy != "" {
			continue
		}

		switch m.MemoryType {
		case domain.MemoryTypePersonality, domain.MemoryTypePreference, domain.MemoryTypeSemantic:
			stableMemories = append(stableMemories, m)
		case domain.MemoryTypeThread, domain.MemoryTypeTask, domain.MemoryTypeEpisodic:
			if m.CreatedAt.After(recentThreshold) {
				recentMemories = append(recentMemories, m)
			}
		default:
			if m.CreatedAt.After(recentThreshold) {
				recentMemories = append(recentMemories, m)
			} else {
				stableMemories = append(stableMemories, m)
			}
		}
	}

	// Build static profile
	staticProfile, err := s.buildStaticProfile(ctx, stableMemories)
	if err != nil {
		slog.Warn("failed to build static profile", "error", err)
		staticProfile = &domain.StaticProfile{}
	}
	profile.StaticProfile = *staticProfile

	// Build dynamic profile
	dynamicProfile := s.buildDynamicProfile(recentMemories)
	profile.DynamicProfile = *dynamicProfile

	return profile, nil
}

// FormatProfileForInjection formats a profile as a system prompt context block.
func FormatProfileForInjection(profile *domain.UserProfile) string {
	if profile == nil {
		return ""
	}

	var sb string
	sb += "<user_profile>\n"

	if profile.StaticProfile.Summary != "" {
		sb += "## User Summary\n" + profile.StaticProfile.Summary + "\n\n"
	}

	if len(profile.StaticProfile.Facts) > 0 {
		sb += "## Known Facts\n"
		for _, f := range profile.StaticProfile.Facts {
			sb += fmt.Sprintf("- %s: %s\n", f.Key, f.Value)
		}
		sb += "\n"
	}

	if len(profile.StaticProfile.Preferences) > 0 {
		sb += "## Preferences\n"
		for _, p := range profile.StaticProfile.Preferences {
			sb += fmt.Sprintf("- %s: %s\n", p.Key, p.Value)
		}
		sb += "\n"
	}

	if len(profile.StaticProfile.Expertise) > 0 {
		sb += "## Expertise\n"
		for _, e := range profile.StaticProfile.Expertise {
			sb += fmt.Sprintf("- %s\n", e)
		}
		sb += "\n"
	}

	if len(profile.DynamicProfile.RecentTopics) > 0 {
		sb += "## Recent Context\n"
		sb += fmt.Sprintf("Topics: %s\n", joinStrings(profile.DynamicProfile.RecentTopics, ", "))
	}

	if profile.DynamicProfile.CurrentContext != "" {
		sb += fmt.Sprintf("Current focus: %s\n", profile.DynamicProfile.CurrentContext)
	}

	if len(profile.DynamicProfile.ActiveTasks) > 0 {
		sb += "## Active Tasks\n"
		for _, t := range profile.DynamicProfile.ActiveTasks {
			sb += fmt.Sprintf("- %s\n", t)
		}
	}

	sb += "</user_profile>"
	return sb
}

func (s *ProfileService) buildStaticProfile(ctx context.Context, memories []domain.Memory) (*domain.StaticProfile, error) {
	if len(memories) == 0 {
		return &domain.StaticProfile{}, nil
	}

	// If no LLM, build from memory types directly
	if s.openRouter == nil {
		return s.buildStaticProfileFromTypes(memories), nil
	}

	// Build memory summaries for LLM
	var memorySummaries string
	limit := 30 // max memories to analyze
	if len(memories) > limit {
		memories = memories[:limit]
	}
	for _, m := range memories {
		summary := m.Summary
		if summary == "" {
			summary = truncate(m.Content, 200)
		}
		memorySummaries += fmt.Sprintf("[%s/%s] %s\n", m.MemoryType, m.Category, summary)
	}

	prompt := fmt.Sprintf(`Based on the following memories, generate a user profile. Extract:
1. Stable facts about the user (name, role, background, etc.)
2. Preferences (tools, languages, workflows)
3. Areas of expertise
4. A brief summary paragraph

Respond in JSON format only:
{
  "facts": [{"key": "attribute name", "value": "attribute value", "confidence": 0.0-1.0}],
  "preferences": [{"key": "preference type", "value": "preferred option", "confidence": 0.0-1.0}],
  "expertise": ["area1", "area2"],
  "summary": "brief paragraph about the user"
}

Memories:
%s`, memorySummaries)

	response, err := s.openRouter.Chat(ctx, []ChatMessage{
		{Role: "system", Content: "You are a user profiling system. Extract stable facts and preferences from memory data. Respond with valid JSON only."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return s.buildStaticProfileFromTypes(memories), nil
	}

	var result struct {
		Facts       []domain.ProfileFact `json:"facts"`
		Preferences []domain.ProfileFact `json:"preferences"`
		Expertise   []string             `json:"expertise"`
		Summary     string               `json:"summary"`
	}
	if err := json.Unmarshal([]byte(cleanJSON(response)), &result); err != nil {
		return s.buildStaticProfileFromTypes(memories), nil
	}

	return &domain.StaticProfile{
		Facts:       result.Facts,
		Preferences: result.Preferences,
		Expertise:   result.Expertise,
		Summary:     result.Summary,
	}, nil
}

func (s *ProfileService) buildStaticProfileFromTypes(memories []domain.Memory) *domain.StaticProfile {
	profile := &domain.StaticProfile{}

	for _, m := range memories {
		switch m.MemoryType {
		case domain.MemoryTypePersonality:
			profile.Facts = append(profile.Facts, domain.ProfileFact{
				Key:        "identity",
				Value:      truncate(m.Content, 200),
				Confidence: 0.8,
				Source:     m.ID,
			})
		case domain.MemoryTypePreference:
			profile.Preferences = append(profile.Preferences, domain.ProfileFact{
				Key:        "preference",
				Value:      truncate(m.Content, 200),
				Confidence: 0.8,
				Source:     m.ID,
			})
		}
	}

	return profile
}

func (s *ProfileService) buildDynamicProfile(recentMemories []domain.Memory) *domain.DynamicProfile {
	profile := &domain.DynamicProfile{
		GeneratedAt: time.Now(),
	}

	topicSet := make(map[string]bool)
	var tasks []string

	for _, m := range recentMemories {
		// Extract topics from tags
		for _, tag := range m.Tags {
			topicSet[tag] = true
		}

		// Extract category as topic
		if m.Category != "" {
			topicSet[string(m.Category)] = true
		}

		// Collect active tasks
		if m.MemoryType == domain.MemoryTypeTask {
			tasks = append(tasks, truncate(m.Content, 100))
		}

		// Latest episodic memory becomes current context
		if m.MemoryType == domain.MemoryTypeEpisodic || m.MemoryType == domain.MemoryTypeThread {
			if profile.CurrentContext == "" {
				profile.CurrentContext = truncate(m.Summary, 200)
				if profile.CurrentContext == "" {
					profile.CurrentContext = truncate(m.Content, 200)
				}
			}
		}
	}

	for topic := range topicSet {
		profile.RecentTopics = append(profile.RecentTopics, topic)
	}
	profile.ActiveTasks = tasks

	return profile
}

func joinStrings(ss []string, sep string) string {
	if len(ss) == 0 {
		return ""
	}
	result := ss[0]
	for _, s := range ss[1:] {
		result += sep + s
	}
	return result
}
