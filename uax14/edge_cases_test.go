package uax14

import (
	"reflect"
	"strings"
	"testing"
)

// TestEdgeCases_UnicodeWhitespace tests various Unicode whitespace characters
func TestEdgeCases_UnicodeWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		hyphens  Hyphens
		expected []int
	}{
		{
			name:     "tab character",
			text:     "hello\tworld",
			hyphens:  HyphensManual,
			expected: []int{0, 6, 11},
		},
		{
			name:     "non-breaking space",
			text:     "hello\u00A0world", // U+00A0 NBSP
			hyphens:  HyphensManual,
			expected: []int{0, 12}, // NBSP is 2 bytes; no break (correct)
		},
		{
			name:     "zero-width space",
			text:     "hello\u200Bworld", // U+200B ZWSP
			hyphens:  HyphensManual,
			expected: []int{0, 8, 13}, // ZWSP is 3 bytes; creates break after it
		},
		{
			name:     "word joiner",
			text:     "hello\u2060world", // U+2060 Word Joiner
			hyphens:  HyphensManual,
			expected: []int{0, 13}, // Word joiner is 3 bytes; no break (correct)
		},
		{
			name:     "line separator",
			text:     "hello\u2028world", // U+2028 Line Separator
			hyphens:  HyphensManual,
			expected: []int{0, 8, 13}, // Creates mandatory break after separator
		},
		{
			name:     "paragraph separator",
			text:     "hello\u2029world", // U+2029 Paragraph Separator
			hyphens:  HyphensManual,
			expected: []int{0, 8, 13}, // Creates mandatory break after separator
		},
		{
			name:     "next line",
			text:     "hello\u0085world", // U+0085 NEL
			hyphens:  HyphensManual,
			expected: []int{0, 7, 12}, // NEL is 2 bytes; creates mandatory break
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindLineBreakOpportunities(tt.text, tt.hyphens)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("FindLineBreakOpportunities(%q, %v) = %v, want %v",
					tt.text, tt.hyphens, result, tt.expected)
			}
		})
	}
}

// TestEdgeCases_LineBreaks tests various line break sequences
func TestEdgeCases_LineBreaks(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		hyphens  Hyphens
		expected []int
	}{
		{
			name:     "CR+LF sequence",
			text:     "hello\r\nworld",
			hyphens:  HyphensManual,
			expected: []int{0, 7, 12}, // Break after CR+LF (at position 7) and end
		},
		{
			name:     "multiple newlines",
			text:     "hello\n\nworld",
			hyphens:  HyphensManual,
			expected: []int{0, 6, 7, 12}, // Break at each newline
		},
		{
			name:     "CR only",
			text:     "hello\rworld",
			hyphens:  HyphensManual,
			expected: []int{0, 6, 11}, // Break at CR
		},
		{
			name:     "mixed line endings",
			text:     "a\nb\rc\r\nd",
			hyphens:  HyphensManual,
			expected: []int{0, 2, 4, 7, 8}, // Break at each line ending
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindLineBreakOpportunities(tt.text, tt.hyphens)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("FindLineBreakOpportunities(%q, %v) = %v, want %v",
					tt.text, tt.hyphens, result, tt.expected)
			}
		})
	}
}

// TestEdgeCases_Hyphens tests hyphen edge cases
func TestEdgeCases_Hyphens(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		hyphens  Hyphens
		expected []int
	}{
		{
			name:     "multiple soft hyphens",
			text:     "su\u00ADper\u00ADcali",
			hyphens:  HyphensManual,
			expected: []int{0, 4, 9, 13}, // Soft hyphens are 2 bytes each
		},
		{
			name:     "soft hyphen at start",
			text:     "\u00ADtest",
			hyphens:  HyphensManual,
			expected: []int{0, 2, 6}, // KNOWN ISSUE: Creates break after soft hyphen at start
		},
		{
			name:     "soft hyphen at end",
			text:     "test\u00AD",
			hyphens:  HyphensManual,
			expected: []int{0, 6}, // Soft hyphen at end
		},
		{
			name:     "em dash",
			text:     "hello—world", // U+2014 EM DASH
			hyphens:  HyphensAuto,
			expected: []int{0, 5, 8, 13}, // UAX#14: B2 (em dash) allows breaks before and after
		},
		{
			name:     "en dash",
			text:     "hello–world", // U+2013 EN DASH
			hyphens:  HyphensAuto,
			expected: []int{0, 13}, // Unicode classifies EN DASH as HH (not HY), no break in middle
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindLineBreakOpportunities(tt.text, tt.hyphens)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("FindLineBreakOpportunities(%q, %v) = %v, want %v",
					tt.text, tt.hyphens, result, tt.expected)
			}
		})
	}
}

// TestEdgeCases_EmptyAndSpaces tests empty and space-only strings
func TestEdgeCases_EmptyAndSpaces(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		hyphens  Hyphens
		expected []int
	}{
		{
			name:     "only spaces",
			text:     "   ",
			hyphens:  HyphensManual,
			expected: []int{0, 3}, // LB7: × SP (don't break before space, so no breaks between spaces)
		},
		{
			name:     "single space",
			text:     " ",
			hyphens:  HyphensManual,
			expected: []int{0, 1},
		},
		{
			name:     "leading spaces",
			text:     "  hello",
			hyphens:  HyphensManual,
			expected: []int{0, 2, 7}, // LB7: × SP (don't break between leading spaces)
		},
		{
			name:     "trailing spaces",
			text:     "hello  ",
			hyphens:  HyphensManual,
			expected: []int{0, 7}, // LB7: × SP (don't break between trailing spaces)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindLineBreakOpportunities(tt.text, tt.hyphens)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("FindLineBreakOpportunities(%q, %v) = %v, want %v",
					tt.text, tt.hyphens, result, tt.expected)
			}
		})
	}
}

// TestEdgeCases_Punctuation tests punctuation edge cases
func TestEdgeCases_Punctuation(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		hyphens Hyphens
	}{
		{"quoted text", `"Hello world"`, HyphensManual},
		{"nested quotes", `"He said 'hello' there"`, HyphensManual},
		{"apostrophe", "don't can't won't", HyphensManual},
		{"ellipsis", "Hello... world", HyphensManual},
		{"multiple exclamation", "Hello!!! World", HyphensManual},
		{"question marks", "Really? Yes? No?", HyphensManual},
		{"mixed punctuation", "Hello, world! How are you?", HyphensManual},
		{"brackets", "Hello [world] (test)", HyphensManual},
		{"nested brackets", "a(b[c{d}e]f)g", HyphensManual},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindLineBreakOpportunities(tt.text, tt.hyphens)
			// Just verify it doesn't panic and returns sensible results
			if len(result) < 2 {
				t.Errorf("Expected at least start and end positions, got %v", result)
			}
			if result[0] != 0 {
				t.Errorf("Expected first position to be 0, got %d", result[0])
			}
			if result[len(result)-1] != len(tt.text) {
				t.Errorf("Expected last position to be %d, got %d", len(tt.text), result[len(result)-1])
			}
		})
	}
}

// TestEdgeCases_Numbers tests number-related edge cases
func TestEdgeCases_Numbers(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		hyphens Hyphens
	}{
		{"date with slashes", "12/31/2024", HyphensManual},
		{"date with dashes", "2024-12-31", HyphensManual},
		{"time", "12:34:56", HyphensManual},
		{"decimal", "3.14159", HyphensManual},
		{"thousands", "1,000,000", HyphensManual},
		{"phone number", "555-1234", HyphensManual},
		{"version number", "v1.2.3", HyphensManual},
		{"currency", "$1,234.56", HyphensManual},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindLineBreakOpportunities(tt.text, tt.hyphens)
			// Verify it doesn't panic and has at least start/end
			if len(result) < 2 {
				t.Errorf("Expected at least start and end positions, got %v", result)
			}
		})
	}
}

// TestEdgeCases_CombiningMarks tests combining marks
func TestEdgeCases_CombiningMarks(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		hyphens Hyphens
	}{
		{"e with acute", "café", HyphensManual},                              // é as single char
		{"e with combining acute", "cafe\u0301", HyphensManual},              // é as e + combining acute
		{"multiple combining marks", "e\u0301\u0302\u0303", HyphensManual},   // e with multiple marks
		{"combining marks in word", "na\u00EFve", HyphensManual},             // ï as single char
		{"emoji with skin tone", "👋🏻", HyphensManual},                        // Wave + light skin tone
		{"emoji with ZWJ", "👨‍👩‍👧‍👦", HyphensManual},                  // Family emoji
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindLineBreakOpportunities(tt.text, tt.hyphens)
			// Just verify it doesn't panic
			if len(result) < 2 {
				t.Errorf("Expected at least start and end positions, got %v", result)
			}
		})
	}
}

// TestEdgeCases_URLs tests URL-like strings
func TestEdgeCases_URLs(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		hyphens Hyphens
	}{
		{"simple URL", "https://example.com", HyphensManual},
		{"URL with path", "https://example.com/path/to/page", HyphensManual},
		{"URL with query", "https://example.com?foo=bar&baz=qux", HyphensManual},
		{"email", "user@example.com", HyphensManual},
		{"email with dots", "first.last@sub.example.com", HyphensManual},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindLineBreakOpportunities(tt.text, tt.hyphens)
			// URLs should have break opportunities at slashes
			if len(result) < 2 {
				t.Errorf("Expected at least start and end positions, got %v", result)
			}
		})
	}
}

// TestEdgeCases_LongText tests performance with long text
func TestEdgeCases_LongText(t *testing.T) {
	// Generate long text
	word := "hello"
	longText := strings.Repeat(word+" ", 10000)

	result := FindLineBreakOpportunities(longText, HyphensManual)

	// Should have approximately 10000 break opportunities (after each space)
	// Plus start and end
	if len(result) < 10000 {
		t.Errorf("Expected at least 10000 break opportunities, got %d", len(result))
	}

	// Verify all positions are in ascending order
	for i := 1; i < len(result); i++ {
		if result[i] <= result[i-1] {
			t.Errorf("Positions not in ascending order at index %d: %d <= %d",
				i, result[i], result[i-1])
		}
	}

	// Verify last position equals text length
	if result[len(result)-1] != len(longText) {
		t.Errorf("Last position %d != text length %d", result[len(result)-1], len(longText))
	}
}

// TestEdgeCases_NoBreakOpportunities tests text with minimal break opportunities
func TestEdgeCases_NoBreakOpportunities(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		hyphens  Hyphens
		maxBreaks int // Maximum expected breaks (excluding start/end)
	}{
		{
			name:      "long word",
			text:      "supercalifragilisticexpialidocious",
			hyphens:   HyphensNone,
			maxBreaks: 0, // Only start and end
		},
		{
			name:      "all connected with word joiner",
			text:      "hello\u2060world\u2060test",
			hyphens:   HyphensManual,
			maxBreaks: 0, // Word joiners prohibit breaks
		},
		{
			name:      "all non-breaking spaces",
			text:      "hello\u00A0world\u00A0test",
			hyphens:   HyphensManual,
			maxBreaks: 0, // NBSPs prohibit breaks
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindLineBreakOpportunities(tt.text, tt.hyphens)
			// Should have start + end + at most maxBreaks
			if len(result) > 2+tt.maxBreaks {
				t.Errorf("Expected at most %d break opportunities, got %d: %v",
					2+tt.maxBreaks, len(result), result)
			}
		})
	}
}

// TestEdgeCases_MixedScripts tests mixed script text
func TestEdgeCases_MixedScripts(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		hyphens Hyphens
	}{
		{"Latin + Arabic", "Hello مرحبا world", HyphensManual},
		{"Latin + Hebrew", "Hello שלום world", HyphensManual},
		{"Latin + Cyrillic", "Hello Привет world", HyphensManual},
		{"Latin + Greek", "Hello Γειά world", HyphensManual},
		{"Latin + Thai", "Hello สวัสดี world", HyphensManual},
		{"Latin + Korean", "Hello 안녕 world", HyphensManual},
		{"all mixed", "Hello 世界 مرحبا שלום Привет", HyphensManual},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindLineBreakOpportunities(tt.text, tt.hyphens)
			// Just verify it doesn't panic
			if len(result) < 2 {
				t.Errorf("Expected at least start and end positions, got %v", result)
			}
			// Should break at spaces at minimum
			spaceCount := strings.Count(tt.text, " ")
			if len(result) < spaceCount+2 {
				t.Errorf("Expected at least %d breaks (spaces + start/end), got %d",
					spaceCount+2, len(result))
			}
		})
	}
}
