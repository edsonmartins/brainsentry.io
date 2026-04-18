package service

import (
	"regexp"
	"sort"
	"strings"
)

// SearchStrategy represents a high-level search strategy chosen by the router.
type SearchStrategy string

const (
	StrategyLexical    SearchStrategy = "LEXICAL"     // exact phrase / keyword
	StrategySemantic   SearchStrategy = "SEMANTIC"    // vector similarity
	StrategyGraph      SearchStrategy = "GRAPH"       // graph traversal
	StrategyTemporal   SearchStrategy = "TEMPORAL"    // time-based lookup
	StrategyEntity     SearchStrategy = "ENTITY"      // entity-centric lookup
	StrategyCoding     SearchStrategy = "CODING"      // code pattern / rule
	StrategyCypher     SearchStrategy = "CYPHER"      // raw graph query
	StrategyHybrid     SearchStrategy = "HYBRID"      // default mixed
)

// RouterDecision is the outcome of query routing.
type RouterDecision struct {
	Strategy   SearchStrategy `json:"strategy"`
	Confidence float64        `json:"confidence"` // 0-1
	Scores     map[SearchStrategy]float64 `json:"scores,omitempty"`
	Matched    []string       `json:"matchedPatterns,omitempty"`
	Fallback   bool           `json:"fallback"` // true if no strong match → default
}

// queryPattern associates a regex with a strategy and weight.
// Negation patterns, when matched within the negationWindow of the primary match,
// nullify the score (e.g. "not a bug" should not match CODING_BUG).
type queryPattern struct {
	strategy       SearchStrategy
	regex          *regexp.Regexp
	weight         float64
	name           string
	negationRegex  *regexp.Regexp // optional
	negationWindow int            // characters before/after
}

// QueryRouterConfig configures the router.
type QueryRouterConfig struct {
	MinConfidence    float64 // minimum confidence to return non-fallback (default 0.3)
	RunnerUpMultiple float64 // top score must exceed runner-up by this factor (default 1.5)
}

// DefaultQueryRouterConfig returns sensible defaults.
func DefaultQueryRouterConfig() QueryRouterConfig {
	return QueryRouterConfig{
		MinConfidence:    0.3,
		RunnerUpMultiple: 1.5,
	}
}

// QueryRouterService classifies queries into search strategies without calling the LLM.
// Uses regex patterns with weighted scoring and negation window detection.
type QueryRouterService struct {
	patterns []queryPattern
	config   QueryRouterConfig
}

// NewQueryRouterService creates a new router with built-in patterns.
func NewQueryRouterService(config QueryRouterConfig) *QueryRouterService {
	return &QueryRouterService{
		patterns: builtinPatterns(),
		config:   config,
	}
}

func builtinPatterns() []queryPattern {
	mustCompile := func(s string) *regexp.Regexp { return regexp.MustCompile(`(?i)` + s) }
	negation := mustCompile(`\b(not|no|without|except|avoid|don'?t|shouldn'?t)\b`)

	return []queryPattern{
		// CYPHER — explicit graph query language. Patterns require actual Cypher syntax:
		// - MATCH ( or MERGE ( with parenthesis
		// - WHERE with property access (node.prop)
		// - RETURN followed by a single letter variable or * (typical Cypher style)
		{
			name:     "cypher_keywords",
			strategy: StrategyCypher,
			regex:    regexp.MustCompile(`\bMATCH\s*\(|\bMERGE\s*\(|\bWHERE\s+\w+\.\w|\bRETURN\s+(?:\*|[a-z]\b)`),
			weight:   1.0,
		},

		// TEMPORAL — time-based queries
		{
			name:     "temporal_words",
			strategy: StrategyTemporal,
			regex:    mustCompile(`\b(yesterday|today|tomorrow|last\s+(week|month|year|day)|this\s+(week|month|year)|ago|since|before|after|between\s+\d|recent|latest|oldest)\b`),
			weight:   0.9,
		},
		{
			name:     "temporal_date",
			strategy: StrategyTemporal,
			regex:    mustCompile(`\b(january|february|march|april|may|june|july|august|september|october|november|december)\b|\d{4}[-/]\d{2}[-/]\d{2}|\bQ[1-4]\b`),
			weight:   0.7,
		},
		{
			name:     "temporal_when",
			strategy: StrategyTemporal,
			regex:    mustCompile(`\bwhen\s+(did|was|were|is|will)\b`),
			weight:   0.8,
		},

		// CODING — code/dev-related queries
		{
			name:     "coding_terms",
			strategy: StrategyCoding,
			regex:    mustCompile(`\b(function|method|class|module|api|endpoint|bug|error|stacktrace|exception|test|compile|deploy|refactor|commit|branch|merge|pull\s+request|pr\b|typescript|javascript|python|golang|rust|java\b|interface|struct|enum)\b`),
			weight:   0.6,
			negationRegex: negation,
			negationWindow: 20,
		},
		{
			name:     "coding_files",
			strategy: StrategyCoding,
			regex:    mustCompile(`\.(go|py|js|ts|tsx|jsx|java|rs|rb|cpp|c|h|hpp|cs|php|swift|kt)\b|\b(src/|tests?/|package\.json|go\.mod|cargo\.toml|requirements\.txt|pom\.xml)\b`),
			weight:   0.9,
		},

		// ENTITY — proper nouns, direct entity lookup
		{
			name:     "entity_quoted",
			strategy: StrategyEntity,
			regex:    regexp.MustCompile(`"[^"]{2,}"|'[^']{2,}'`),
			weight:   0.8,
		},
		{
			name:     "entity_proper_nouns",
			strategy: StrategyEntity,
			regex:    regexp.MustCompile(`\b(?:[A-Z][a-z]+){2,}\b|\b[A-Z][a-z]+\s+[A-Z][a-z]+\b`),
			weight:   0.5,
		},
		{
			name:     "entity_who_is",
			strategy: StrategyEntity,
			regex:    mustCompile(`\b(who\s+is|what\s+is|define|definition\s+of|meaning\s+of)\b`),
			weight:   0.6,
		},

		// GRAPH — multi-hop reasoning, relationships
		{
			name:     "graph_related",
			strategy: StrategyGraph,
			regex:    mustCompile(`\b(related\s+to|connected\s+to|linked\s+to|associated\s+with|depends?\s+on|uses|implements|extends|inherits|children\s+of|parents?\s+of|neighbors?|path\s+(from|to|between))\b`),
			weight:   0.8,
			negationRegex: negation,
			negationWindow: 15,
		},
		{
			// Require relationship verbs right after "how/why does" to avoid matching
			// generic "why does the code fail" (which is coding, not graph).
			name:     "graph_how_does",
			strategy: StrategyGraph,
			regex:    mustCompile(`\b(how\s+does|why\s+does|what\s+causes|what\s+triggers)\b\s+\S+\s+(depend|use|connect|relate|interact|affect|require|trigger|cause)|explain\s+the\s+relationship`),
			weight:   0.7,
		},

		// LEXICAL — exact phrase markers
		{
			name:     "lexical_exact_phrase",
			strategy: StrategyLexical,
			regex:    mustCompile(`\bexact(ly)?\s+(match|phrase|words)\b|\bliterally\b`),
			weight:   0.9,
		},

		// SEMANTIC — conceptual / similarity queries
		{
			name:     "semantic_similar",
			strategy: StrategySemantic,
			regex:    mustCompile(`\b(similar\s+to|like|resembles?|analogous|comparable|concept\s+of|idea\s+of|anything\s+about)\b`),
			weight:   0.8,
		},
	}
}

// Classify applies all patterns to a query and returns the best-matching strategy.
func (r *QueryRouterService) Classify(query string) *RouterDecision {
	if strings.TrimSpace(query) == "" {
		return &RouterDecision{Strategy: StrategyHybrid, Fallback: true}
	}

	scores := make(map[SearchStrategy]float64)
	var matched []string

	for _, p := range r.patterns {
		loc := p.regex.FindStringIndex(query)
		if loc == nil {
			continue
		}

		// Check negation within window
		if p.negationRegex != nil {
			if r.hasNegationNear(query, loc, p.negationRegex, p.negationWindow) {
				continue
			}
		}

		scores[p.strategy] += p.weight
		matched = append(matched, p.name)
	}

	if len(scores) == 0 {
		return &RouterDecision{
			Strategy: StrategyHybrid,
			Fallback: true,
			Scores:   scores,
		}
	}

	// Find top and runner-up
	type kv struct {
		s      SearchStrategy
		v      float64
	}
	ranked := make([]kv, 0, len(scores))
	for s, v := range scores {
		ranked = append(ranked, kv{s, v})
	}
	sort.Slice(ranked, func(i, j int) bool { return ranked[i].v > ranked[j].v })

	top := ranked[0]
	var runnerUp float64
	if len(ranked) > 1 {
		runnerUp = ranked[1].v
	}

	// Confidence: normalize top score to [0,1] with ceiling at 2.0 raw score.
	confidence := top.v / 2.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	// Runner-up multiple check: top must be clearly ahead
	clearWinner := runnerUp == 0 || top.v >= runnerUp*r.config.RunnerUpMultiple

	fallback := confidence < r.config.MinConfidence || !clearWinner
	strategy := top.s
	if fallback {
		strategy = StrategyHybrid
	}

	return &RouterDecision{
		Strategy:   strategy,
		Confidence: confidence,
		Scores:     scores,
		Matched:    matched,
		Fallback:   fallback,
	}
}

// hasNegationNear returns true if the negation regex matches within
// negationWindow characters before or after the primary match location.
func (r *QueryRouterService) hasNegationNear(query string, primaryLoc []int, negRegex *regexp.Regexp, window int) bool {
	start := primaryLoc[0] - window
	if start < 0 {
		start = 0
	}
	end := primaryLoc[1] + window
	if end > len(query) {
		end = len(query)
	}
	slice := query[start:end]
	return negRegex.MatchString(slice)
}
