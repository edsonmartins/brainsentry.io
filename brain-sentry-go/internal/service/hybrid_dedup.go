package service

import (
	"math"
	"strings"
	"unicode"
)

// DedupStrategy selects which algorithm HybridDeduplicator uses.
type DedupStrategy string

const (
	// DedupJaroWinkler favours short-string name matches.
	DedupJaroWinkler DedupStrategy = "jaro_winkler"
	// DedupBlockingJW blocks by prefix then runs Jaro-Winkler within each block;
	// scales O(N log N) instead of O(N^2).
	DedupBlockingJW DedupStrategy = "blocking_jw"
	// DedupSemantic uses embedding cosine similarity.
	DedupSemantic DedupStrategy = "semantic"
)

// HybridDeduplicator exposes tunable deduplication strategies on top of the
// existing SimHash pipeline. All strategies return pairs { AID, BID, Score }
// where Score ∈ [0, 1]; the caller decides the threshold.
type HybridDeduplicator struct {
	// PrefixBlockSize controls how many leading characters define a block in
	// the blocking strategy. Larger blocks = tighter filter but fewer matches.
	PrefixBlockSize int
}

// NewHybridDeduplicator builds the deduplicator with sensible defaults.
func NewHybridDeduplicator() *HybridDeduplicator {
	return &HybridDeduplicator{PrefixBlockSize: 3}
}

// DedupItem is a minimal view of a record to compare. Embedding is optional.
type DedupItem struct {
	ID        string
	Text      string
	Embedding []float32
}

// DedupPair is a candidate duplicate pair with its similarity score.
type DedupPair struct {
	AID   string  `json:"a"`
	BID   string  `json:"b"`
	Score float64 `json:"score"`
}

// FindDuplicates returns candidate duplicate pairs from items using strategy.
// threshold filters out pairs with score below the cutoff.
func (d *HybridDeduplicator) FindDuplicates(items []DedupItem, strategy DedupStrategy, threshold float64) []DedupPair {
	switch strategy {
	case DedupSemantic:
		return d.semantic(items, threshold)
	case DedupBlockingJW:
		return d.blocking(items, threshold)
	case DedupJaroWinkler:
		fallthrough
	default:
		return d.jaroWinklerAll(items, threshold)
	}
}

func (d *HybridDeduplicator) jaroWinklerAll(items []DedupItem, threshold float64) []DedupPair {
	pairs := make([]DedupPair, 0)
	norm := make([]string, len(items))
	for i, it := range items {
		norm[i] = normaliseText(it.Text)
	}
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			s := jaroWinkler(norm[i], norm[j])
			if s >= threshold {
				pairs = append(pairs, DedupPair{AID: items[i].ID, BID: items[j].ID, Score: s})
			}
		}
	}
	return pairs
}

func (d *HybridDeduplicator) blocking(items []DedupItem, threshold float64) []DedupPair {
	blockSize := d.PrefixBlockSize
	if blockSize <= 0 {
		blockSize = 3
	}
	blocks := make(map[string][]int)
	norm := make([]string, len(items))
	for i, it := range items {
		norm[i] = normaliseText(it.Text)
		prefix := safePrefix(norm[i], blockSize)
		blocks[prefix] = append(blocks[prefix], i)
	}
	pairs := make([]DedupPair, 0)
	for _, idxs := range blocks {
		for i := 0; i < len(idxs); i++ {
			for j := i + 1; j < len(idxs); j++ {
				a, b := idxs[i], idxs[j]
				s := jaroWinkler(norm[a], norm[b])
				if s >= threshold {
					pairs = append(pairs, DedupPair{AID: items[a].ID, BID: items[b].ID, Score: s})
				}
			}
		}
	}
	return pairs
}

func (d *HybridDeduplicator) semantic(items []DedupItem, threshold float64) []DedupPair {
	pairs := make([]DedupPair, 0)
	for i := 0; i < len(items); i++ {
		if len(items[i].Embedding) == 0 {
			continue
		}
		for j := i + 1; j < len(items); j++ {
			if len(items[j].Embedding) == 0 {
				continue
			}
			s := cosine(items[i].Embedding, items[j].Embedding)
			if s >= threshold {
				pairs = append(pairs, DedupPair{AID: items[i].ID, BID: items[j].ID, Score: s})
			}
		}
	}
	return pairs
}

// normaliseText lowercases, strips punctuation, and collapses whitespace.
func normaliseText(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prevSpace := false
	for _, r := range strings.ToLower(s) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			prevSpace = false
		case unicode.IsSpace(r):
			if !prevSpace {
				b.WriteRune(' ')
				prevSpace = true
			}
		}
	}
	return strings.TrimSpace(b.String())
}

func safePrefix(s string, n int) string {
	runes := []rune(s)
	if len(runes) < n {
		return string(runes)
	}
	return string(runes[:n])
}

// jaroWinkler returns the Jaro-Winkler similarity between two strings.
func jaroWinkler(a, b string) float64 {
	j := jaro(a, b)
	if j < 0.7 {
		return j
	}
	prefix := 0
	max := 4
	for i := 0; i < len(a) && i < len(b) && i < max; i++ {
		if a[i] != b[i] {
			break
		}
		prefix++
	}
	return j + float64(prefix)*0.1*(1-j)
}

func jaro(a, b string) float64 {
	if a == b {
		return 1
	}
	la, lb := len(a), len(b)
	if la == 0 || lb == 0 {
		return 0
	}
	matchDist := max(la, lb)/2 - 1
	if matchDist < 0 {
		matchDist = 0
	}
	matchesA := make([]bool, la)
	matchesB := make([]bool, lb)
	matches := 0
	for i := 0; i < la; i++ {
		start := i - matchDist
		if start < 0 {
			start = 0
		}
		end := i + matchDist + 1
		if end > lb {
			end = lb
		}
		for j := start; j < end; j++ {
			if matchesB[j] {
				continue
			}
			if a[i] != b[j] {
				continue
			}
			matchesA[i] = true
			matchesB[j] = true
			matches++
			break
		}
	}
	if matches == 0 {
		return 0
	}
	var transpositions float64
	k := 0
	for i := 0; i < la; i++ {
		if !matchesA[i] {
			continue
		}
		for !matchesB[k] {
			k++
		}
		if a[i] != b[k] {
			transpositions++
		}
		k++
	}
	m := float64(matches)
	return (m/float64(la) + m/float64(lb) + (m-transpositions/2)/m) / 3
}

func cosine(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		af := float64(a[i])
		bf := float64(b[i])
		dot += af * bf
		na += af * af
		nb += bf * bf
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
