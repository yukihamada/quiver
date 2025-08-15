package privacy

import (
	"regexp"
	"strings"
)

// SensitivePatterns defines patterns that should be redacted
var SensitivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`),                    // Credit card
	regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),          // Email
	regexp.MustCompile(`\b(?:password|passwd|pwd|token|api[_-]?key|secret)\s*[:=]\s*\S+`), // Passwords
	regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),                                         // SSN
	regexp.MustCompile(`\+?\d{1,3}[\s.-]?\(?\d{1,4}\)?[\s.-]?\d{1,4}[\s.-]?\d{1,4}`),   // Phone
}

// Redactor handles sensitive information redaction
type Redactor struct {
	enabled  bool
	patterns []*regexp.Regexp
}

// NewRedactor creates a new redactor
func NewRedactor(enabled bool) *Redactor {
	return &Redactor{
		enabled:  enabled,
		patterns: SensitivePatterns,
	}
}

// RedactPrompt removes sensitive information from prompts
func (r *Redactor) RedactPrompt(prompt string) string {
	if !r.enabled {
		return prompt
	}

	redacted := prompt
	for _, pattern := range r.patterns {
		redacted = pattern.ReplaceAllString(redacted, "[REDACTED]")
	}

	return redacted
}

// HasSensitiveContent checks if content contains sensitive information
func (r *Redactor) HasSensitiveContent(content string) bool {
	for _, pattern := range r.patterns {
		if pattern.MatchString(content) {
			return true
		}
	}
	return false
}

// MinimalRedaction performs minimal redaction for logging
func (r *Redactor) MinimalRedaction(content string) string {
	if len(content) > 100 {
		return content[:50] + "...[TRUNCATED]..." + content[len(content)-20:]
	}
	return strings.Repeat("*", len(content))
}