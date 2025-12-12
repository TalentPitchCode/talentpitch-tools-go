package groq

import (
	_ "embed"
	"log"
	"strings"
)

// defaultBlockedTermsFile is automatically loaded at compile time from blocked_terms.txt
// using the //go:embed directive. The file is embedded into the binary, so it works
// even when the package is installed from GitHub.
//
//go:embed blocked_terms.txt
var defaultBlockedTermsFile string

// defaultBlockedTerms returns a list of default offensive terms loaded from blocked_terms.txt
// The file is embedded at compile time, so no file I/O is needed at runtime.
// This is a basic list - projects can override with their own terms via Config
func defaultBlockedTerms() []string {
	// Load from embedded file (loaded at compile time via //go:embed)
	if defaultBlockedTermsFile == "" {
		log.Printf("blocked_terms.txt is empty or not found")
		return []string{}
	}

	// Split by newlines and filter empty lines
	lines := strings.Split(defaultBlockedTermsFile, "\n")
	terms := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines and comments (lines starting with #)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			terms = append(terms, trimmed)
		}
	}

	log.Printf("Loaded %d blocked terms from blocked_terms.txt", len(terms))
	return terms
}

// containsBlockedTerm checks if the message contains any of the blocked terms
// Performs case-insensitive matching
func containsBlockedTerm(messageText string, blockedTerms []string) (bool, string) {
	if len(blockedTerms) == 0 {
		return false, ""
	}

	messageLower := strings.ToLower(messageText)

	// Normalize message: replace common separators with spaces for better matching
	normalizedMessage := strings.ReplaceAll(messageLower, "_", " ")
	normalizedMessage = strings.ReplaceAll(normalizedMessage, "-", " ")

	// Check each blocked term
	for _, term := range blockedTerms {
		termLower := strings.ToLower(strings.TrimSpace(term))
		if termLower == "" {
			continue
		}

		// Simple contains check - blocked terms should be exact matches
		// Check in both original and normalized message
		if strings.Contains(messageLower, termLower) || strings.Contains(normalizedMessage, termLower) {
			// Additional validation: check if it's a word boundary
			// This helps avoid false positives (e.g., "class" in "classroom")
			if isWholeWord(messageLower, termLower) || isWholeWord(normalizedMessage, termLower) {
				return true, termLower
			}
		}
	}

	return false, ""
}

// isWholeWord checks if the term appears as a whole word in the message
func isWholeWord(message, term string) bool {
	// Find all occurrences
	index := 0
	for {
		pos := strings.Index(message[index:], term)
		if pos == -1 {
			break
		}
		actualPos := index + pos

		// Check character before
		beforeOK := actualPos == 0
		if !beforeOK && actualPos > 0 {
			beforeChar := message[actualPos-1]
			beforeOK = !isAlphanumeric(beforeChar)
		}

		// Check character after
		afterPos := actualPos + len(term)
		afterOK := afterPos >= len(message)
		if !afterOK {
			afterChar := message[afterPos]
			afterOK = !isAlphanumeric(afterChar)
		}

		// If both boundaries are OK, it's a whole word
		if beforeOK && afterOK {
			return true
		}

		// Move to next position
		index = actualPos + 1
	}

	return false
}

// isAlphanumeric checks if a byte is alphanumeric
func isAlphanumeric(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}
