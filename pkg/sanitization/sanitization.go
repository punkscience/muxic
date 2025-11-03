// Package sanitization provides functionality for sanitizing strings to be
// compatible with Windows filesystem requirements. It handles prohibited characters,
// Unicode transliteration, and trimming of leading/trailing periods and spaces.
package sanitization

import (
	"regexp"
	"strings"

	"github.com/fiam/gounidecode/unidecode"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Sanitizer defines the interface for string sanitization operations.
// This follows the Interface Segregation Principle by providing specific methods
// for different sanitization needs.
type Sanitizer interface {
	// SanitizeForFilesystem sanitizes a string for use in Windows filesystem paths
	SanitizeForFilesystem(input string) string
	
	// SanitizeFolderName sanitizes a string specifically for folder names
	SanitizeFolderName(input string) string
	
	// SanitizeFileName sanitizes a string specifically for file names
	SanitizeFileName(input string) string
}

// WindowsSanitizer implements filesystem sanitization for Windows compatibility.
// It follows the Single Responsibility Principle by focusing solely on sanitization.
type WindowsSanitizer struct {
	titleCaser        cases.Caser
	substitutions     map[string]string
	prohibitedPattern *regexp.Regexp
}

// NewWindowsSanitizer creates a new Windows filesystem sanitizer with default rules.
// This follows the Dependency Inversion Principle by allowing configuration
// of substitution rules.
func NewWindowsSanitizer() *WindowsSanitizer {
	// Default substitutions based on common music file conventions
	defaultSubstitutions := map[string]string{
		"feat.":     "ft",
		"Feat.":     "ft", 
		"Feat":      "ft",
		"featuring": "ft",
		"Featuring": "ft",
		"&":         "and",
		"@":         "at",
		"w/":        "with",
		"vs.":       "vs",
	}

	// Windows prohibited characters: < > : " | ? * \ /
	// Using regex for efficient replacement
	prohibitedPattern := regexp.MustCompile(`[<>:"|?*\\\/]`)

	return &WindowsSanitizer{
		titleCaser:        cases.Title(language.English),
		substitutions:     defaultSubstitutions,
		prohibitedPattern: prohibitedPattern,
	}
}

// NewWindowsSanitizerWithSubstitutions creates a sanitizer with custom substitution rules.
// This follows the Open/Closed Principle by allowing extension without modification.
func NewWindowsSanitizerWithSubstitutions(substitutions map[string]string) *WindowsSanitizer {
	sanitizer := NewWindowsSanitizer()
	sanitizer.substitutions = substitutions
	return sanitizer
}

// SanitizeForFilesystem performs comprehensive sanitization for filesystem compatibility.
// This is the main method that orchestrates all sanitization steps.
func (w *WindowsSanitizer) SanitizeForFilesystem(input string) string {
	if input == "" {
		return ""
	}

	// Step 1: Trim leading and trailing whitespace
	result := strings.TrimSpace(input)
	
	// Step 2: Convert Unicode/non-ASCII characters to ASCII equivalents
	result = unidecode.Unidecode(result)
	
	// Step 3: Apply specific substitutions BEFORE character replacement
	// This ensures patterns like "w/" are handled before "/" becomes "-"
	result = w.applySubstitutions(result)
	
	// Step 4: Replace prohibited characters with hyphens
	result = w.prohibitedPattern.ReplaceAllString(result, "-")
	
	// Step 5: Additional cleanup: remove spaces around hyphens if they result from substitutions
	result = regexp.MustCompile(`\s+-\s+`).ReplaceAllString(result, "-")
	
	// Step 6: Normalize multiple consecutive spaces to single spaces
	result = w.normalizeSpaces(result)
	
	// Step 7: Apply intelligent title casing (preserve existing uppercase)
	result = w.intelligentTitleCase(result)
	
	// Step 8: Trim leading and trailing periods and spaces
	result = w.trimPeriodsAndSpaces(result)
	
	return result
}

// SanitizeFolderName sanitizes a string for use as a folder name.
// Folders have the same restrictions as files in Windows.
func (w *WindowsSanitizer) SanitizeFolderName(input string) string {
	return w.SanitizeForFilesystem(input)
}

// SanitizeFileName sanitizes a string for use as a file name.
// Files have the same restrictions as folders in Windows.
func (w *WindowsSanitizer) SanitizeFileName(input string) string {
	return w.SanitizeForFilesystem(input)
}

// normalizeSpaces replaces multiple consecutive spaces with single spaces.
func (w *WindowsSanitizer) normalizeSpaces(input string) string {
	// Replace multiple spaces with single space
	spacePattern := regexp.MustCompile(`\s+`)
	return spacePattern.ReplaceAllString(input, " ")
}

// applySubstitutions applies the configured text substitutions.
// This method performs case-insensitive matching for better user experience.
func (w *WindowsSanitizer) applySubstitutions(input string) string {
	result := input
	
	// Apply substitutions with specific handling for different patterns
	for original, replacement := range w.substitutions {
		switch original {
		case "&":
			// Replace standalone ampersands
			pattern := regexp.MustCompile(`\s*&\s*`)
			result = pattern.ReplaceAllString(result, " "+replacement+" ")
		case "@":
			// Replace standalone @ symbols
			pattern := regexp.MustCompile(`\s*@\s*`)
			result = pattern.ReplaceAllString(result, " "+replacement+" ")
		case "w/":
			// Replace w/ pattern  
			pattern := regexp.MustCompile(`(?i)\bw/`)
			result = pattern.ReplaceAllString(result, replacement)
		default:
			// For feat. patterns, handle the period specially since it's followed by space
			if strings.HasSuffix(strings.ToLower(original), "feat.") {
				pattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(strings.TrimSuffix(original, ".")) + `\.`)
				result = pattern.ReplaceAllString(result, replacement)
			} else {
				// For all other substitutions  
				pattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(original) + `\b`)
				result = pattern.ReplaceAllString(result, replacement)
			}
		}
	}
	
	return result
}

// trimPeriodsAndSpaces removes leading and trailing periods and spaces.
// This is critical for Windows filesystem compatibility as files/folders
// cannot start or end with periods or spaces.
func (w *WindowsSanitizer) trimPeriodsAndSpaces(input string) string {
	// Trim leading and trailing periods and spaces
	return strings.Trim(input, ". ")
}

// intelligentTitleCase applies title casing while preserving existing uppercase letters
// where appropriate (e.g., "AC/DC" should become "AC-DC", not "Ac-Dc").
func (w *WindowsSanitizer) intelligentTitleCase(input string) string {
	if input == "" {
		return ""
	}
	
	// Use regex to handle title casing while preserving certain patterns
	words := regexp.MustCompile(`\b\w+\b`).FindAllString(input, -1)
	result := input
	
	for _, word := range words {
		if w.shouldPreserveCase(word) {
			// Keep the word as-is if it should preserve case
			continue
		}
		
		// Replace the word with its title-cased version
		titleCased := w.titleCaser.String(word)
		result = strings.Replace(result, word, titleCased, 1)
	}
	
	return result
}

// shouldPreserveCase determines if a word should preserve its current casing
// rather than applying standard title case rules.
func (w *WindowsSanitizer) shouldPreserveCase(word string) bool {
	// Only preserve short all-uppercase words (like "AC", "DC", "UK", etc.)
	// but NOT file extensions or very long uppercase strings
	if len(word) >= 2 && len(word) <= 4 && strings.ToUpper(word) == word && 
		!strings.Contains(word, ".") && !strings.Contains(word, "-") {
		return true
	}
	
	// Don't preserve case for most other patterns to ensure consistent title casing
	return false
}

// SanitizeTrackMetadata is a convenience function for sanitizing music track metadata.
// It takes artist, album, and title and returns sanitized versions suitable for
// building file paths.
func (w *WindowsSanitizer) SanitizeTrackMetadata(artist, album, title string) (string, string, string) {
	return w.SanitizeForFilesystem(artist),
		w.SanitizeForFilesystem(album),
		w.SanitizeForFilesystem(title)
}

// ValidateWindowsPath checks if a path is valid for Windows filesystem.
// Returns true if the path is valid, false otherwise.
func ValidateWindowsPath(path string) bool {
	if path == "" {
		return false
	}
	
	// Check for prohibited characters
	prohibitedChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range prohibitedChars {
		if strings.Contains(path, char) {
			return false
		}
	}
	
	// Check for paths ending with periods or spaces
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if part != "" && (strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ")) {
			return false
		}
		if part != "" && (strings.HasPrefix(part, ".") || strings.HasPrefix(part, " ")) {
			return false
		}
	}
	
	return true
}