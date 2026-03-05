package service

import (
	"regexp"
	"strings"

	"github.com/integraltech/brainsentry/internal/domain"
)

// classifierRule defines a pattern-based rule for memory type classification.
type classifierRule struct {
	memoryType domain.MemoryType
	keywords   []string
	patterns   []*regexp.Regexp
	weight     float64 // base weight for this rule
}

var classifierRules = []classifierRule{
	{
		memoryType: domain.MemoryTypePersonality,
		keywords: []string{
			"i am", "i'm", "my name is", "i always", "i never", "i usually",
			"i identify as", "my background", "i was born", "my role is",
			"i work as", "my job", "i live in", "i speak", "my nationality",
			"i believe", "my values", "my personality", "i tend to",
		},
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b(i am|i'm)\s+(a|an)\s+\w+`),
			regexp.MustCompile(`(?i)\bmy\s+(name|age|role|title|profession)\b`),
			regexp.MustCompile(`(?i)\bi\s+(always|never|usually|typically)\b`),
		},
		weight: 1.0,
	},
	{
		memoryType: domain.MemoryTypePreference,
		keywords: []string{
			"i prefer", "i like", "i dislike", "i hate", "i love",
			"i want", "i don't want", "favorite", "favourite",
			"i choose", "rather than", "instead of", "preferred",
			"my preference", "i enjoy", "i avoid", "dark mode", "light mode",
			"use bun", "use npm", "use yarn", "use pnpm",
		},
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\bi\s+(prefer|like|dislike|hate|love|enjoy|avoid)\b`),
			regexp.MustCompile(`(?i)\b(always|never)\s+use\b`),
			regexp.MustCompile(`(?i)\bfavou?rite\b`),
			regexp.MustCompile(`(?i)\bpreferred?\b`),
		},
		weight: 1.0,
	},
	{
		memoryType: domain.MemoryTypeProcedural,
		keywords: []string{
			"how to", "step by step", "procedure", "tutorial", "guide",
			"instructions", "recipe", "workflow", "pipeline", "process",
			"first,", "then,", "finally,", "next,", "steps:",
			"to do this", "run the command", "execute", "install",
			"configure", "setup", "set up",
		},
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\bhow\s+to\b`),
			regexp.MustCompile(`(?i)\bstep\s+\d+`),
			regexp.MustCompile(`(?i)\b(first|then|next|finally),?\s`),
			regexp.MustCompile(`(?i)\b(run|execute|install|configure)\s`),
		},
		weight: 1.0,
	},
	{
		memoryType: domain.MemoryTypeTask,
		keywords: []string{
			"todo", "to-do", "task", "action item", "deadline",
			"due date", "assigned to", "pending", "in progress",
			"blocked", "backlog", "sprint", "milestone", "deliverable",
			"need to", "must do", "should do", "will do", "plan to",
			"reminder", "follow up", "follow-up",
		},
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b(todo|to-do|task)\b`),
			regexp.MustCompile(`(?i)\b(need|must|should|will)\s+(to\s+)?(do|finish|complete|implement|fix)\b`),
			regexp.MustCompile(`(?i)\b(deadline|due\s+date|due\s+by)\b`),
			regexp.MustCompile(`(?i)\breminder\b`),
		},
		weight: 1.0,
	},
	{
		memoryType: domain.MemoryTypeEmotion,
		keywords: []string{
			"i feel", "i felt", "frustrated", "happy", "sad", "angry",
			"excited", "anxious", "worried", "disappointed", "grateful",
			"overwhelmed", "confused", "confident", "nervous", "proud",
			"stressed", "relieved", "annoyed", "thrilled",
		},
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\bi\s+(feel|felt|am feeling)\b`),
			regexp.MustCompile(`(?i)\b(frustrated|happy|sad|angry|excited|anxious|worried|disappointed)\b`),
			regexp.MustCompile(`(?i)\b(overwhelmed|confused|stressed|relieved|annoyed|thrilled)\b`),
		},
		weight: 1.0,
	},
	{
		memoryType: domain.MemoryTypeThread,
		keywords: []string{
			"in this conversation", "earlier you said", "as we discussed",
			"continuing from", "regarding our chat", "back to the topic",
			"as mentioned", "you asked about", "we were talking about",
			"let me continue", "going back to",
		},
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b(earlier|previously)\s+(you|we|i)\b`),
			regexp.MustCompile(`(?i)\b(in this|our)\s+(conversation|chat|discussion)\b`),
			regexp.MustCompile(`(?i)\bcontinuing\s+from\b`),
		},
		weight: 0.8, // lower weight — thread is the default for short-lived context
	},
	{
		memoryType: domain.MemoryTypeEpisodic,
		keywords: []string{
			"yesterday", "last week", "last month", "happened",
			"i experienced", "i encountered", "i discovered",
			"i found out", "i learned that", "i noticed",
			"during the meeting", "in the session", "when i was",
			"at the conference", "during deployment",
		},
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b(yesterday|last\s+(week|month|year))\b`),
			regexp.MustCompile(`(?i)\bi\s+(experienced|encountered|discovered|noticed)\b`),
			regexp.MustCompile(`(?i)\b(during|at|in)\s+the\s+(meeting|session|conference|deployment)\b`),
			regexp.MustCompile(`(?i)\b\d{4}-\d{2}-\d{2}\b`), // ISO date reference
		},
		weight: 1.0,
	},
	{
		memoryType: domain.MemoryTypeSemantic,
		keywords: []string{
			"is defined as", "means that", "refers to", "the concept of",
			"in general", "typically", "by definition", "according to",
			"a fact:", "note:", "important:", "the difference between",
			"is a type of", "is related to", "is part of",
		},
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b(is|are)\s+(defined\s+as|a\s+type\s+of)\b`),
			regexp.MustCompile(`(?i)\b(means|refers)\s+to\b`),
			regexp.MustCompile(`(?i)\baccording\s+to\b`),
			regexp.MustCompile(`(?i)\bthe\s+(concept|definition|difference)\s+of\b`),
		},
		weight: 0.7, // lower weight — semantic is the default fallback
	},
}

// ClassifyMemoryType classifies a memory's cognitive type using pattern matching.
// Returns the detected type and a confidence score (0-1).
func ClassifyMemoryType(content string, tags []string, category domain.MemoryCategory) (domain.MemoryType, float64) {
	if content == "" {
		return domain.MemoryTypeSemantic, 0
	}

	lower := strings.ToLower(content)
	scores := make(map[domain.MemoryType]float64)

	// Score based on keyword and pattern matching
	for _, rule := range classifierRules {
		score := 0.0

		// Keyword matches
		for _, kw := range rule.keywords {
			if strings.Contains(lower, kw) {
				score += 1.0
			}
		}

		// Regex pattern matches (stronger signal)
		for _, p := range rule.patterns {
			if p.MatchString(content) {
				score += 2.0
			}
		}

		scores[rule.memoryType] = score * rule.weight
	}

	// Boost from category hints
	switch category {
	case domain.CategoryAction:
		scores[domain.MemoryTypeTask] += 2.0
	case domain.CategoryPattern, domain.CategoryAntipattern:
		scores[domain.MemoryTypeProcedural] += 2.0
	case domain.CategoryInsight, domain.CategoryKnowledge, domain.CategoryDomain:
		scores[domain.MemoryTypeSemantic] += 1.5
	case domain.CategoryBug, domain.CategoryOptimization:
		scores[domain.MemoryTypeEpisodic] += 1.5
	case domain.CategoryDecision:
		scores[domain.MemoryTypeSemantic] += 1.0
	}

	// Boost from tags
	tagStr := strings.ToLower(strings.Join(tags, " "))
	if strings.Contains(tagStr, "preference") || strings.Contains(tagStr, "config") {
		scores[domain.MemoryTypePreference] += 2.0
	}
	if strings.Contains(tagStr, "todo") || strings.Contains(tagStr, "task") {
		scores[domain.MemoryTypeTask] += 2.0
	}
	if strings.Contains(tagStr, "personal") || strings.Contains(tagStr, "identity") {
		scores[domain.MemoryTypePersonality] += 2.0
	}

	// Find the winner
	bestType := domain.MemoryTypeSemantic
	bestScore := 0.0
	totalScore := 0.0

	for t, s := range scores {
		totalScore += s
		if s > bestScore {
			bestScore = s
			bestType = t
		}
	}

	// Calculate confidence as normalized score
	confidence := 0.0
	if totalScore > 0 {
		confidence = bestScore / totalScore
	}
	if bestScore == 0 {
		return domain.MemoryTypeSemantic, 0.1 // default fallback with low confidence
	}

	return bestType, confidence
}
