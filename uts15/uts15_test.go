package uts15

import (
	"strings"
	"testing"
)

func TestNFC(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		name     string
	}{
		// Basic ASCII (no change)
		{"hello", "hello", "ASCII unchanged"},

		// Canonical composition
		{"e\u0301", "\u00E9", "e + acute accent -> é"},
		{"a\u0300", "\u00E0", "a + grave accent -> à"},
		{"o\u0308", "\u00F6", "o + diaeresis -> ö"},

		// Already composed (no change)
		{"\u00E9", "\u00E9", "é already composed"},

		// Hangul composition
		{"\u1100\u1161", "\uAC00", "Hangul L+V -> syllable GA"},
		{"\u1100\u1161\u11A8", "\uAC01", "Hangul L+V+T -> syllable GAG"},

		// Empty string
		{"", "", "empty string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NFC(tt.input)
			if result != tt.expected {
				t.Errorf("NFC(%q) = %q (% X), want %q (% X)",
					tt.input, result, []rune(result),
					tt.expected, []rune(tt.expected))
			}
		})
	}
}

func TestNFD(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		name     string
	}{
		// Basic ASCII (no change)
		{"hello", "hello", "ASCII unchanged"},

		// Canonical decomposition
		{"\u00E9", "e\u0301", "é -> e + acute accent"},
		{"\u00E0", "a\u0300", "à -> a + grave accent"},
		{"\u00F6", "o\u0308", "ö -> o + diaeresis"},

		// Already decomposed (no change)
		{"e\u0301", "e\u0301", "e + acute already decomposed"},

		// Hangul decomposition
		{"\uAC00", "\u1100\u1161", "Hangul GA -> L+V"},
		{"\uAC01", "\u1100\u1161\u11A8", "Hangul GAG -> L+V+T"},

		// Empty string
		{"", "", "empty string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NFD(tt.input)
			if result != tt.expected {
				t.Errorf("NFD(%q) = %q (% X), want %q (% X)",
					tt.input, result, []rune(result),
					tt.expected, []rune(tt.expected))
			}
		})
	}
}

func TestNFKC(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		name     string
	}{
		// Basic ASCII (no change)
		{"hello", "hello", "ASCII unchanged"},

		// Compatibility decomposition + composition
		{"\uFB01", "fi", "ligature ﬁ -> fi"},
		{"\uFB00", "ff", "ligature ﬀ -> ff"},
		{"\u2460", "1", "circled 1 -> 1"},
		{"\u00BC", "1\u20444", "¼ -> 1/4"},

		// Full-width to half-width
		{"\uFF21", "A", "full-width A -> A"},
		{"\uFF10", "0", "full-width 0 -> 0"},

		// Canonical composition also applied
		{"e\u0301", "\u00E9", "e + acute -> é"},

		// Empty string
		{"", "", "empty string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NFKC(tt.input)
			if result != tt.expected {
				t.Errorf("NFKC(%q) = %q (% X), want %q (% X)",
					tt.input, result, []rune(result),
					tt.expected, []rune(tt.expected))
			}
		})
	}
}

func TestNFKD(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		name     string
	}{
		// Basic ASCII (no change)
		{"hello", "hello", "ASCII unchanged"},

		// Compatibility decomposition
		{"\uFB01", "fi", "ligature ﬁ -> fi"},
		{"\uFB00", "ff", "ligature ﬀ -> ff"},
		{"\u2460", "1", "circled 1 -> 1"},

		// Full-width to half-width
		{"\uFF21", "A", "full-width A -> A"},
		{"\uFF10", "0", "full-width 0 -> 0"},

		// Canonical decomposition also applied
		{"\u00E9", "e\u0301", "é -> e + acute"},

		// Empty string
		{"", "", "empty string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NFKD(tt.input)
			if result != tt.expected {
				t.Errorf("NFKD(%q) = %q (% X), want %q (% X)",
					tt.input, result, []rune(result),
					tt.expected, []rune(tt.expected))
			}
		})
	}
}

func TestIsNFC(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"hello", true},
		{"\u00E9", true},              // é (composed)
		{"e\u0301", false},             // e + acute (decomposed)
		{"\uAC00", true},               // Hangul syllable (composed)
		{"\u1100\u1161", false},        // Hangul L+V (decomposed)
		{"", true},
	}

	for _, tt := range tests {
		result := IsNFC(tt.input)
		if result != tt.expected {
			t.Errorf("IsNFC(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestIsNFD(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"hello", true},
		{"\u00E9", false},              // é (composed)
		{"e\u0301", true},              // e + acute (decomposed)
		{"\uAC00", false},              // Hangul syllable (composed)
		{"\u1100\u1161", true},         // Hangul L+V (decomposed)
		{"", true},
	}

	for _, tt := range tests {
		result := IsNFD(tt.input)
		if result != tt.expected {
			t.Errorf("IsNFD(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestCanonicalOrdering(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		name     string
	}{
		// Single combining mark (no change)
		{"a\u0301", "a\u0301", "a + acute"},

		// Multiple combining marks out of order (both class 230, so order preserved)
		{"a\u0301\u0300", "a\u0301\u0300", "a + acute + grave (same class, stable)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NFD(tt.input)
			if result != tt.expected {
				t.Errorf("Canonical ordering failed: NFD(%q) = %q (% X), want %q (% X)",
					tt.input, result, []rune(result),
					tt.expected, []rune(tt.expected))
			}
		})
	}
}

func TestHangulComposition(t *testing.T) {
	tests := []struct {
		l, v, tOpt rune
		expected   rune
		name       string
	}{
		{0x1100, 0x1161, 0, 0xAC00, "L+V -> GA"},
		{0x1100, 0x1161, 0x11A8, 0xAC01, "L+V+T -> GAG"},
		{0x1100, 0x1162, 0, 0xAC1C, "L+V -> GAE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input string
			if tt.tOpt == 0 {
				input = string([]rune{tt.l, tt.v})
			} else {
				input = string([]rune{tt.l, tt.v, tt.tOpt})
			}
			result := NFC(input)
			expected := string(tt.expected)
			if result != expected {
				t.Errorf("Hangul composition failed: NFC(%q) = %q (% X), want %q (% X)",
					input, result, []rune(result),
					expected, []rune(expected))
			}
		})
	}
}

func TestHangulDecomposition(t *testing.T) {
	tests := []struct {
		syllable rune
		l, v, tOpt rune
		name       string
	}{
		{0xAC00, 0x1100, 0x1161, 0, "GA -> L+V"},
		{0xAC01, 0x1100, 0x1161, 0x11A8, "GAG -> L+V+T"},
		{0xAC1C, 0x1100, 0x1162, 0, "GAE -> L+V"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := string(tt.syllable)
			result := NFD(input)
			var expected string
			if tt.tOpt == 0 {
				expected = string([]rune{tt.l, tt.v})
			} else {
				expected = string([]rune{tt.l, tt.v, tt.tOpt})
			}
			if result != expected {
				t.Errorf("Hangul decomposition failed: NFD(%q) = %q (% X), want %q (% X)",
					input, result, []rune(result),
					expected, []rune(expected))
			}
		})
	}
}

func TestNormalizationStability(t *testing.T) {
	// Test that normalizing twice gives the same result
	tests := []string{
		"hello",
		"café",
		"e\u0301",
		"\u00E9",
		"한글",
		"\uFB01",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			// NFC stability
			nfc1 := NFC(input)
			nfc2 := NFC(nfc1)
			if nfc1 != nfc2 {
				t.Errorf("NFC not stable: NFC(NFC(%q)) != NFC(%q)", input, input)
			}

			// NFD stability
			nfd1 := NFD(input)
			nfd2 := NFD(nfd1)
			if nfd1 != nfd2 {
				t.Errorf("NFD not stable: NFD(NFD(%q)) != NFD(%q)", input, input)
			}

			// NFKC stability
			nfkc1 := NFKC(input)
			nfkc2 := NFKC(nfkc1)
			if nfkc1 != nfkc2 {
				t.Errorf("NFKC not stable: NFKC(NFKC(%q)) != NFKC(%q)", input, input)
			}

			// NFKD stability
			nfkd1 := NFKD(input)
			nfkd2 := NFKD(nfkd1)
			if nfkd1 != nfkd2 {
				t.Errorf("NFKD not stable: NFKD(NFKD(%q)) != NFKD(%q)", input, input)
			}
		})
	}
}

func TestStringComparison(t *testing.T) {
	// Test that different representations compare equal after normalization
	tests := []struct {
		s1, s2 string
		name   string
	}{
		{"café", "cafe\u0301", "café composed vs decomposed"},
		{"\u00E9", "e\u0301", "é vs e+acute"},
		{"\uAC00", "\u1100\u1161", "Hangul GA composed vs decomposed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if NFC(tt.s1) != NFC(tt.s2) {
				t.Errorf("NFC comparison failed: %q != %q", tt.s1, tt.s2)
			}
			if NFD(tt.s1) != NFD(tt.s2) {
				t.Errorf("NFD comparison failed: %q != %q", tt.s1, tt.s2)
			}
		})
	}
}

// Benchmarks
func BenchmarkNFC(b *testing.B) {
	text := strings.Repeat("e\u0301", 100) // decomposed
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NFC(text)
	}
}

func BenchmarkNFD(b *testing.B) {
	text := strings.Repeat("\u00E9", 100) // composed
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NFD(text)
	}
}

func BenchmarkNFKC(b *testing.B) {
	text := strings.Repeat("\uFB01", 50) // ligatures
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NFKC(text)
	}
}

func BenchmarkNFKD(b *testing.B) {
	text := strings.Repeat("\uFB01", 50) // ligatures
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NFKD(text)
	}
}

func BenchmarkIsNFC(b *testing.B) {
	text := strings.Repeat("\u00E9", 100) // composed (already NFC)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsNFC(text)
	}
}
