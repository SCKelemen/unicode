package uax50

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

// Test basic orientation lookups for well-known characters
func TestBasicOrientations(t *testing.T) {
	tests := []struct {
		r        rune
		expected Orientation
		name     string
	}{
		// Latin letters should be Rotated
		{'A', Rotated, "Latin capital A"},
		{'a', Rotated, "Latin lowercase a"},
		{'Z', Rotated, "Latin capital Z"},

		// Digits should be Rotated
		{'0', Rotated, "Digit 0"},
		{'5', Rotated, "Digit 5"},
		{'9', Rotated, "Digit 9"},

		// Common punctuation should be Rotated
		{'.', Rotated, "Full stop"},
		{',', Rotated, "Comma"},
		{'!', Rotated, "Exclamation mark"},
		{'?', Rotated, "Question mark"},

		// CJK ideographs should be Upright (U+4E00-U+9FFF)
		{'中', Upright, "CJK ideograph (中)"},
		{'国', Upright, "CJK ideograph (国)"},
		{'日', Upright, "CJK ideograph (日)"},
		{'本', Upright, "CJK ideograph (本)"},

		// Hiragana should be Upright
		{'あ', Upright, "Hiragana letter A"},
		{'ん', Upright, "Hiragana letter N"},

		// Katakana should be Upright
		{'ア', Upright, "Katakana letter A"},
		{'ン', Upright, "Katakana letter N"},

		// Mathematical operators (some are Upright)
		{'\u00B1', Upright, "Plus-minus sign"},
		{'\u00D7', Upright, "Multiplication sign"},
		{'\u00F7', Upright, "Division sign"},

		// Some special punctuation that's Upright
		{'\u00A7', Upright, "Section sign"},
		{'\u00A9', Upright, "Copyright sign"},
		{'\u00AE', Upright, "Registered sign"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LookupOrientation(tt.r)
			if got != tt.expected {
				t.Errorf("LookupOrientation(%U %q) = %v, want %v", tt.r, tt.r, got, tt.expected)
			}
		})
	}
}

// Test transformed orientations (Tu and Tr)
func TestTransformedOrientations(t *testing.T) {
	tests := []struct {
		r        rune
		expected Orientation
		name     string
	}{
		// Wave dash should be TransformedRotated
		{'\u301C', TransformedRotated, "Wave dash"},

		// Ideographic punctuation is TransformedUpright
		{'\u3001', TransformedUpright, "Ideographic comma"},
		{'\u3002', TransformedUpright, "Ideographic full stop"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LookupOrientation(tt.r)
			if got != tt.expected {
				t.Errorf("LookupOrientation(%U %q) = %v, want %v", tt.r, tt.r, got, tt.expected)
			}
		})
	}
}

// Test convenience functions
func TestIsUpright(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
		name     string
	}{
		{'中', true, "CJK ideograph should be upright"},
		{'A', false, "Latin letter should not be upright"},
		{'\u00B1', true, "Plus-minus should be upright"},
		{'\u3001', true, "Ideographic comma (Tu) should be upright"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsUpright(tt.r)
			if got != tt.expected {
				t.Errorf("IsUpright(%U %q) = %v, want %v", tt.r, tt.r, got, tt.expected)
			}
		})
	}
}

func TestIsRotated(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
		name     string
	}{
		{'A', true, "Latin letter should be rotated"},
		{'中', false, "CJK ideograph should not be rotated"},
		{'0', true, "Digit should be rotated"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRotated(tt.r)
			if got != tt.expected {
				t.Errorf("IsRotated(%U %q) = %v, want %v", tt.r, tt.r, got, tt.expected)
			}
		})
	}
}

func TestRequiresTransformation(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
		name     string
	}{
		{'\u301C', true, "Wave dash (Tr) should require transformation"},
		{'\u3001', true, "Ideographic comma (Tu) should require transformation"},
		{'A', false, "Latin letter should not require transformation"},
		{'中', false, "CJK ideograph should not require transformation"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RequiresTransformation(tt.r)
			if got != tt.expected {
				t.Errorf("RequiresTransformation(%U %q) = %v, want %v", tt.r, tt.r, got, tt.expected)
			}
		})
	}
}

func TestGetBaseOrientation(t *testing.T) {
	tests := []struct {
		r        rune
		expected Orientation
		name     string
	}{
		{'A', Rotated, "Rotated stays Rotated"},
		{'中', Upright, "Upright stays Upright"},
		{'\u3001', Upright, "TransformedUpright becomes Upright"},
		{'\u301C', Rotated, "TransformedRotated becomes Rotated"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetBaseOrientation(tt.r)
			if got != tt.expected {
				t.Errorf("GetBaseOrientation(%U %q) = %v, want %v", tt.r, tt.r, got, tt.expected)
			}
		})
	}
}

// Test Orientation string representation
func TestOrientationString(t *testing.T) {
	tests := []struct {
		o        Orientation
		expected string
	}{
		{Rotated, "R"},
		{Upright, "U"},
		{TransformedUpright, "Tu"},
		{TransformedRotated, "Tr"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.o.String()
			if got != tt.expected {
				t.Errorf("Orientation.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Test that unassigned code points return default (Rotated)
func TestUnassignedCodePoints(t *testing.T) {
	// Test some unassigned code points in various ranges
	// These should return Rotated as the default
	unassigned := []rune{
		0x0378, // Unassigned in Greek
		0x0530, // Unassigned in Armenian
	}

	for _, r := range unassigned {
		got := LookupOrientation(r)
		// The default for most unassigned is Rotated
		// But some ranges default to Upright per the spec
		// We're just testing that we get a valid orientation
		if got != Rotated && got != Upright {
			t.Errorf("LookupOrientation(%U) = %v, expected Rotated or Upright", r, got)
		}
	}
}

// TestDataFileConsistency tests that the data file can be loaded and
// that every entry in the VerticalOrientation.txt file is correctly
// represented in our generated data.
func TestDataFileConsistency(t *testing.T) {
	file, err := os.Open("VerticalOrientation.txt")
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

		// Extract just the property value (R, U, Tu, Tr)
		valueStr = strings.Fields(valueStr)[0]

		var expectedOrientation Orientation
		switch valueStr {
		case "R":
			expectedOrientation = Rotated
		case "U":
			expectedOrientation = Upright
		case "Tu":
			expectedOrientation = TransformedUpright
		case "Tr":
			expectedOrientation = TransformedRotated
		default:
			t.Errorf("Line %d: Unknown orientation value: %s", lineNum, valueStr)
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
				got := LookupOrientation(r)
				if got != expectedOrientation {
					t.Errorf("Line %d: LookupOrientation(%U) = %v, want %v",
						lineNum, r, got, expectedOrientation)
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
			got := LookupOrientation(r)
			if got != expectedOrientation {
				t.Errorf("Line %d: LookupOrientation(%U) = %v, want %v",
					lineNum, r, got, expectedOrientation)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		t.Errorf("Error reading file: %v", err)
	}
}

// Benchmark the lookup function
func BenchmarkLookupOrientation(b *testing.B) {
	testRunes := []rune{'A', '中', '\u301C', '0', 'あ'}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, r := range testRunes {
			LookupOrientation(r)
		}
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
			_ = LookupOrientation(tt.r)
		})
	}
}
