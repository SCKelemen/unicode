package uax11

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

// Test basic width lookups for well-known characters
func TestBasicWidths(t *testing.T) {
	tests := []struct {
		r        rune
		expected Width
		name     string
	}{
		// ASCII should be Narrow
		{'A', Narrow, "Latin capital A"},
		{'a', Narrow, "Latin lowercase a"},
		{'0', Narrow, "Digit 0"},
		{' ', Narrow, "Space"},
		{'!', Narrow, "Exclamation mark"},

		// CJK ideographs should be Wide
		{'中', Wide, "CJK ideograph (中)"},
		{'国', Wide, "CJK ideograph (国)"},
		{'日', Wide, "CJK ideograph (日)"},
		{'本', Wide, "CJK ideograph (本)"},

		// Hiragana and Katakana should be Wide
		{'あ', Wide, "Hiragana letter A"},
		{'ア', Wide, "Katakana letter A"},

		// Fullwidth forms should be Fullwidth
		{'\uFF01', Fullwidth, "Fullwidth exclamation mark"},
		{'\uFF21', Fullwidth, "Fullwidth Latin capital A"},
		{'\uFF41', Fullwidth, "Fullwidth Latin lowercase a"},

		// Halfwidth forms should be Halfwidth
		{'\uFF65', Halfwidth, "Halfwidth Katakana middle dot"},
		{'\uFF66', Halfwidth, "Halfwidth Katakana letter WO"},

		// Ambiguous characters
		{'\u00A1', Ambiguous, "Inverted exclamation mark"},
		{'\u00A7', Ambiguous, "Section sign"},
		{'\u00B1', Ambiguous, "Plus-minus sign"},
		{'Ω', Ambiguous, "Greek capital letter Omega"},
		{'α', Ambiguous, "Greek small letter alpha"},

		// Neutral characters
		{'\u0080', Neutral, "Control character"},
		{'\u00A9', Neutral, "Copyright sign"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LookupWidth(tt.r)
			if got != tt.expected {
				t.Errorf("LookupWidth(%U %q) = %v, want %v", tt.r, tt.r, got, tt.expected)
			}
		})
	}
}

// Test default Wide ranges for unassigned CJK ideographs
func TestDefaultWideRanges(t *testing.T) {
	tests := []struct {
		r    rune
		name string
	}{
		{0x3400, "CJK Extension A start"},
		{0x4DBF, "CJK Extension A end"},
		{0x4E00, "CJK Unified Ideographs start"},
		{0x9FFF, "CJK Unified Ideographs end"},
		{0xF900, "CJK Compatibility Ideographs start"},
		{0xFAFF, "CJK Compatibility Ideographs end"},
		{0x20000, "Plane 2 start"},
		{0x2FFFD, "Plane 2 end"},
		{0x30000, "Plane 3 start"},
		{0x3FFFD, "Plane 3 end"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LookupWidth(tt.r)
			if got != Wide {
				t.Errorf("LookupWidth(%U) = %v, want Wide", tt.r, got)
			}
		})
	}
}

// Test convenience functions
func TestIsWide(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
		name     string
	}{
		{'中', true, "CJK ideograph should be wide"},
		{'A', false, "Latin letter should not be wide"},
		{'\uFF21', true, "Fullwidth Latin should be wide"},
		{'\uFF65', false, "Halfwidth Katakana should not be wide"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsWide(tt.r)
			if got != tt.expected {
				t.Errorf("IsWide(%U %q) = %v, want %v", tt.r, tt.r, got, tt.expected)
			}
		})
	}
}

func TestIsNarrow(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
		name     string
	}{
		{'A', true, "Latin letter should be narrow"},
		{'中', false, "CJK ideograph should not be narrow"},
		{'\uFF65', true, "Halfwidth Katakana should be narrow"},
		{'\uFF21', false, "Fullwidth Latin should not be narrow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNarrow(tt.r)
			if got != tt.expected {
				t.Errorf("IsNarrow(%U %q) = %v, want %v", tt.r, tt.r, got, tt.expected)
			}
		})
	}
}

func TestIsAmbiguous(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
		name     string
	}{
		{'Ω', true, "Greek Omega should be ambiguous"},
		{'\u00A7', true, "Section sign should be ambiguous"},
		{'A', false, "Latin A should not be ambiguous"},
		{'中', false, "CJK ideograph should not be ambiguous"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAmbiguous(tt.r)
			if got != tt.expected {
				t.Errorf("IsAmbiguous(%U %q) = %v, want %v", tt.r, tt.r, got, tt.expected)
			}
		})
	}
}

// Test context-based width resolution
func TestResolveWidth(t *testing.T) {
	tests := []struct {
		r               rune
		ctx             Context
		expected        Width
		name            string
	}{
		// Ambiguous in East Asian context becomes Wide
		{'Ω', ContextEastAsian, Wide, "Greek Omega in East Asian context"},
		{'\u00A7', ContextEastAsian, Wide, "Section sign in East Asian context"},

		// Ambiguous in Narrow context becomes Narrow
		{'Ω', ContextNarrow, Narrow, "Greek Omega in narrow context"},
		{'\u00A7', ContextNarrow, Narrow, "Section sign in narrow context"},

		// Non-ambiguous characters stay the same
		{'A', ContextEastAsian, Narrow, "Latin A stays narrow in East Asian context"},
		{'中', ContextNarrow, Wide, "CJK ideograph stays wide in narrow context"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveWidth(tt.r, tt.ctx)
			if got != tt.expected {
				t.Errorf("ResolveWidth(%U %q, %v) = %v, want %v", tt.r, tt.r, tt.ctx, got, tt.expected)
			}
		})
	}
}

// Test character width calculation
func TestCharWidth(t *testing.T) {
	tests := []struct {
		r        rune
		ctx      Context
		expected int
		name     string
	}{
		{'A', ContextNarrow, 1, "ASCII in narrow context"},
		{'中', ContextNarrow, 2, "CJK in narrow context"},
		{'Ω', ContextNarrow, 1, "Ambiguous in narrow context"},
		{'Ω', ContextEastAsian, 2, "Ambiguous in East Asian context"},
		{'\uFF21', ContextNarrow, 2, "Fullwidth in narrow context"},
		{'\uFF65', ContextNarrow, 1, "Halfwidth in narrow context"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CharWidth(tt.r, tt.ctx)
			if got != tt.expected {
				t.Errorf("CharWidth(%U %q, %v) = %v, want %v", tt.r, tt.r, tt.ctx, got, tt.expected)
			}
		})
	}
}

// Test string width calculation
func TestStringWidth(t *testing.T) {
	tests := []struct {
		s        string
		ctx      Context
		expected int
		name     string
	}{
		{"Hello", ContextNarrow, 5, "ASCII in narrow context"},
		{"中国", ContextNarrow, 4, "CJK in narrow context"},
		{"Hello世界", ContextNarrow, 9, "Mixed ASCII and CJK"},
		{"ΩΩΩ", ContextNarrow, 3, "Greek in narrow context"},
		{"ΩΩΩ", ContextEastAsian, 6, "Greek in East Asian context"},
		{"", ContextNarrow, 0, "Empty string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StringWidth(tt.s, tt.ctx)
			if got != tt.expected {
				t.Errorf("StringWidth(%q, %v) = %v, want %v", tt.s, tt.ctx, got, tt.expected)
			}
		})
	}
}

// Test Width string representation
func TestWidthString(t *testing.T) {
	tests := []struct {
		w        Width
		expected string
	}{
		{Neutral, "N"},
		{Ambiguous, "A"},
		{Fullwidth, "F"},
		{Halfwidth, "H"},
		{Narrow, "Na"},
		{Wide, "W"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.w.String()
			if got != tt.expected {
				t.Errorf("Width.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestDataFileConsistency tests that the data file can be loaded and
// that every entry in the EastAsianWidth.txt file is correctly
// represented in our generated data.
func TestDataFileConsistency(t *testing.T) {
	file, err := os.Open("EastAsianWidth.txt")
	if err != nil {
		t.Skipf("Skipping data file consistency test: %v", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on semicolon
		parts := strings.Split(line, ";")
		if len(parts) < 2 {
			continue
		}

		codePoints := strings.TrimSpace(parts[0])
		valueStr := strings.TrimSpace(parts[1])

		// Extract just the property value (A, F, H, N, Na, W)
		valueStr = strings.Fields(valueStr)[0]

		var expectedWidth Width
		switch valueStr {
		case "A":
			expectedWidth = Ambiguous
		case "F":
			expectedWidth = Fullwidth
		case "H":
			expectedWidth = Halfwidth
		case "N":
			expectedWidth = Neutral
		case "Na":
			expectedWidth = Narrow
		case "W":
			expectedWidth = Wide
		default:
			t.Errorf("Line %d: Unknown width value: %s", lineNum, valueStr)
			continue
		}

		// Parse code point or range
		if strings.Contains(codePoints, "..") {
			// Range
			rangeParts := strings.Split(codePoints, "..")
			startVal, err1 := strconv.ParseInt(rangeParts[0], 16, 32)
			endVal, err2 := strconv.ParseInt(rangeParts[1], 16, 32)

			if err1 != nil || err2 != nil {
				t.Errorf("Line %d: Failed to parse range %s", lineNum, codePoints)
				continue
			}

			// Test a few points in the range
			testPoints := []rune{rune(startVal), rune(endVal)}
			if endVal-startVal > 2 {
				testPoints = append(testPoints, rune((startVal+endVal)/2))
			}

			for _, r := range testPoints {
				got := LookupWidth(r)
				if got != expectedWidth {
					t.Errorf("Line %d: LookupWidth(%U) = %v, want %v",
						lineNum, r, got, expectedWidth)
				}
			}
		} else {
			// Single code point
			val, err := strconv.ParseInt(codePoints, 16, 32)
			if err != nil {
				t.Errorf("Line %d: Failed to parse code point %s", lineNum, codePoints)
				continue
			}

			r := rune(val)
			got := LookupWidth(r)
			if got != expectedWidth {
				t.Errorf("Line %d: LookupWidth(%U) = %v, want %v",
					lineNum, r, got, expectedWidth)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		t.Errorf("Error reading file: %v", err)
	}
}

// Test edge cases
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		r    rune
		name string
	}{
		{0x0000, "Null character"},
		{0x10FFFF, "Maximum Unicode code point"},
		{0xFFFF, "Noncharacter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just ensure it doesn't panic
			_ = LookupWidth(tt.r)
			_ = CharWidth(tt.r, ContextNarrow)
		})
	}
}

// Benchmark the lookup function
func BenchmarkLookupWidth(b *testing.B) {
	testRunes := []rune{'A', '中', 'Ω', '\uFF21', '\uFF65'}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, r := range testRunes {
			LookupWidth(r)
		}
	}
}

// Benchmark string width calculation
func BenchmarkStringWidth(b *testing.B) {
	testStrings := []string{
		"Hello",
		"中国日本",
		"Hello世界",
		"ΩαβγδΩ",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range testStrings {
			StringWidth(s, ContextNarrow)
		}
	}
}
