package service

import (
	"fmt"
	"regexp"
	"strings"
)

// PIIType represents a type of personally identifiable information.
type PIIType string

const (
	PIIEmail      PIIType = "EMAIL"
	PIIPhone      PIIType = "PHONE"
	PIISSN        PIIType = "SSN"
	PIICreditCard PIIType = "CREDIT_CARD"
	PIIAPIKey     PIIType = "API_KEY"
	PIIJWTToken   PIIType = "JWT_TOKEN"
	PIIIPAddress  PIIType = "IP_ADDRESS"
	PIIPrivateKey PIIType = "PRIVATE_KEY"
)

// PIIMatch represents a detected PII occurrence.
type PIIMatch struct {
	Type  PIIType `json:"type"`
	Start int     `json:"start"`
	End   int     `json:"end"`
	Value string  `json:"-"` // never serialize the actual value
}

var piiPatterns = []struct {
	Type    PIIType
	Pattern *regexp.Regexp
	Mask    string
}{
	{PIIEmail, regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`), "[EMAIL]"},
	{PIIPhone, regexp.MustCompile(`(?:\+?1[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}`), "[PHONE]"},
	{PIISSN, regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`), "[SSN]"},
	{PIICreditCard, regexp.MustCompile(`\b(?:\d{4}[-\s]?){3}\d{4}\b`), "[CREDIT_CARD]"},
	{PIIAPIKey, regexp.MustCompile(`(?i)(?:api[_-]?key|apikey|api_secret|secret_key)[\s]*[=:]\s*["']?([a-zA-Z0-9_\-]{20,})["']?`), "[API_KEY]"},
	{PIIJWTToken, regexp.MustCompile(`eyJ[a-zA-Z0-9_-]+\.eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+`), "[JWT_TOKEN]"},
	{PIIIPAddress, regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`), "[IP_ADDRESS]"},
	{PIIPrivateKey, regexp.MustCompile(`-----BEGIN (?:RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`), "[PRIVATE_KEY_BLOCK]"},
}

// PIIService handles detection and masking of personally identifiable information.
type PIIService struct{}

// NewPIIService creates a new PIIService.
func NewPIIService() *PIIService {
	return &PIIService{}
}

// Detect finds all PII occurrences in text.
func (s *PIIService) Detect(text string) []PIIMatch {
	var matches []PIIMatch
	for _, p := range piiPatterns {
		locs := p.Pattern.FindAllStringIndex(text, -1)
		for _, loc := range locs {
			matches = append(matches, PIIMatch{
				Type:  p.Type,
				Start: loc[0],
				End:   loc[1],
				Value: text[loc[0]:loc[1]],
			})
		}
	}
	return matches
}

// ContainsPII checks if text contains any PII.
func (s *PIIService) ContainsPII(text string) bool {
	for _, p := range piiPatterns {
		if p.Pattern.MatchString(text) {
			return true
		}
	}
	return false
}

// Mask replaces all PII in text with type-specific placeholders.
func (s *PIIService) Mask(text string) string {
	result := text
	for _, p := range piiPatterns {
		result = p.Pattern.ReplaceAllString(result, p.Mask)
	}
	return result
}

// MaskForLLM masks PII before sending to external LLM APIs.
// Returns masked text and a summary of what was masked.
func (s *PIIService) MaskForLLM(text string) (string, string) {
	matches := s.Detect(text)
	if len(matches) == 0 {
		return text, ""
	}

	masked := s.Mask(text)

	// Build summary of masked items
	counts := make(map[PIIType]int)
	for _, m := range matches {
		counts[m.Type]++
	}
	var parts []string
	for t, c := range counts {
		parts = append(parts, fmt.Sprintf("%d %s", c, t))
	}
	summary := fmt.Sprintf("Masked PII: %s", strings.Join(parts, ", "))

	return masked, summary
}
