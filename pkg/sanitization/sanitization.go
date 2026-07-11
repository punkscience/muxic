// Package sanitization provides functionality for sanitizing strings to be
// compatible with Windows filesystem requirements. It handles prohibited characters,
// Unicode transliteration, and trimming of leading/trailing periods and spaces.
package sanitization

import (
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode"

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
	titleCaserMu      sync.Mutex // Protects titleCaser which is not thread-safe for concurrent transformations
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

	// Step 5: Tidy hyphens. First collapse whitespace on BOTH sides of a hyphen
	// ("A - B" -> "A-B"), then drop whitespace that only precedes a hyphen
	// ("A -B" -> "A-B") — this happens when Unicode transliteration leaves a
	// trailing space right before a character that gets replaced by a hyphen
	// (e.g. "擁抱/Embrace" -> "Yong Bao /Embrace" -> "Yong Bao-Embrace"). A space
	// that only follows a hyphen is left intact so "AC-DC feat." spacing survives.
	result = regexp.MustCompile(`\s+-\s+`).ReplaceAllString(result, "-")
	result = regexp.MustCompile(`\s+-`).ReplaceAllString(result, "-")

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
// Files have the same restrictions as folders in Windows. When the input carries
// a recognizable file extension (e.g. "song.mp3"), the base name is sanitized
// normally while the extension is only lower-cased and stripped of prohibited
// characters — it is never title-cased ("song.mp3" -> "Song.mp3", not "Song.Mp3").
func (w *WindowsSanitizer) SanitizeFileName(input string) string {
	ext := filepath.Ext(input)
	if !isFileExtension(ext) {
		return w.SanitizeForFilesystem(input)
	}

	base := strings.TrimSuffix(input, ext)
	cleanExt := strings.ToLower(w.prohibitedPattern.ReplaceAllString(ext[1:], "-"))
	return w.SanitizeForFilesystem(base) + "." + cleanExt
}

// isFileExtension reports whether ext (as returned by filepath.Ext, including the
// leading dot) looks like a real file extension: a short, purely alphanumeric
// suffix. This keeps names that merely contain periods (e.g. "Album.Name") from
// being split.
func isFileExtension(ext string) bool {
	if len(ext) < 2 || len(ext) > 6 || ext[0] != '.' {
		return false
	}
	for _, r := range ext[1:] {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
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

	// Apply substitutions deterministically, longest key first, so overlapping
	// keys such as "Feat." (5) and "Feat" (4) don't depend on Go's randomized
	// map iteration order (which made results non-deterministic).
	originals := make([]string, 0, len(w.substitutions))
	for original := range w.substitutions {
		originals = append(originals, original)
	}
	sort.Slice(originals, func(i, j int) bool {
		if len(originals[i]) != len(originals[j]) {
			return len(originals[i]) > len(originals[j])
		}
		return originals[i] < originals[j]
	})

	for _, original := range originals {
		replacement := w.substitutions[original]
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
			// Keys ending in a period (e.g. "feat.", "vs.") need special handling:
			// a trailing \b won't match after the period (period→space is
			// non-word→non-word), so match the token plus the literal period.
			if strings.HasSuffix(original, ".") {
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

	// Transform each whole-word match in place to avoid partial replacement
	// inside other words (e.g., replacing "is" inside "This").
	wordPattern := regexp.MustCompile(`\b\w+\b`)
	return wordPattern.ReplaceAllStringFunc(input, func(word string) string {
		if w.shouldPreserveCase(word) {
			// Keep the word as-is if it should preserve case.
			return word
		}

		w.titleCaserMu.Lock()
		titleCased := w.titleCaser.String(word)
		w.titleCaserMu.Unlock()
		return titleCased
	})
}

// shouldPreserveCase determines if a word should preserve its current casing
// rather than applying standard title case rules.
func (w *WindowsSanitizer) shouldPreserveCase(word string) bool {
	// Preserve two-letter all-uppercase acronyms (like "AC", "DC", "UK").
	// Longer all-caps runs (e.g. "ALL CAPS TITLE") are treated as shouting and
	// normalized to title case rather than preserved.
	if len(word) == 2 && strings.ToUpper(word) == word {
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
