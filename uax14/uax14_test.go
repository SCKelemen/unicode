package uax14

import (
	"reflect"
	"testing"
)

func TestFindLineBreakOpportunities(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		hyphens  Hyphens
		expected []int
	}{
		{
			name:     "empty string",
			text:     "",
			hyphens:  HyphensManual,
			expected: []int{0},
		},
		{
			name:     "simple word",
			text:     "hello",
			hyphens:  HyphensManual,
			expected: []int{0, 5},
		},
		{
			name:     "two words",
			text:     "hello world",
			hyphens:  HyphensManual,
			expected: []int{0, 6, 11},
		},
		{
			name:     "multiple spaces",
			text:     "hello  world",
			hyphens:  HyphensManual,
			expected: []int{0, 7, 12}, // LB7: × SP (don't break between spaces)
		},
		{
			name:     "with newline",
			text:     "hello\nworld",
			hyphens:  HyphensManual,
			expected: []int{0, 6, 11}, // Break at newline position and end
		},
		{
			name:     "with soft hyphen - manual mode",
			text:     "super\u00ADcalifragilistic",
			hyphens:  HyphensManual,
			expected: []int{0, 7, 22}, // Soft hyphen (2 bytes at 5-6), break after at 7
		},
		{
			name:     "with soft hyphen - none mode",
			text:     "super\u00ADcalifragilistic",
			hyphens:  HyphensNone,
			expected: []int{0, 22}, // No hyphenation
		},
		{
			name:     "with hard hyphen - manual mode",
			text:     "twenty-one",
			hyphens:  HyphensManual,
			expected: []int{0, 7, 10}, // UAX#14: hyphen-minus (U+002D) is class BA, allows breaks
		},
		{
			name:     "with hard hyphen - auto mode",
			text:     "twenty-one",
			hyphens:  HyphensAuto,
			expected: []int{0, 7, 10}, // UAX#14: hyphen-minus (U+002D) is class BA, allows breaks
		},
		{
			name:     "CJK text",
			text:     "こんにちは世界",
			hyphens:  HyphensManual,
			expected: []int{0, 3, 6, 9, 12, 15, 18, 21}, // Hiragana & Kanji are ID: break between each char
		},
		{
			name:     "mixed text",
			text:     "Hello 世界 world",
			hyphens:  HyphensManual,
			expected: []int{0, 6, 9, 13, 18}, // Breaks at spaces and ideographic transitions
		},
		{
			name:     "punctuation",
			text:     "Hello, world!",
			hyphens:  HyphensManual,
			expected: []int{0, 7, 13},
		},
		{
			name:     "numbers",
			text:     "The year 2024 is here",
			hyphens:  HyphensManual,
			expected: []int{0, 4, 9, 14, 17, 21},
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

func TestGetBreakClass(t *testing.T) {
	tests := []struct {
		name     string
		r        rune
		expected BreakClass
	}{
		{"newline", '\n', ClassLF},
		{"carriage return", '\r', ClassCR},
		{"space", ' ', ClassSP},
		{"tab", '\t', ClassBA},         // Official Unicode: BA (Break After)
		{"soft hyphen", '\u00AD', ClassBA}, // Official Unicode: BA (Break After)
		{"hard hyphen", '-', ClassHY},
		{"letter", 'a', ClassAL},
		{"digit", '5', ClassNU},
		{"open paren", '(', ClassOP},
		{"close paren", ')', ClassCP},
		{"exclamation", '!', ClassEX},
		{"question", '?', ClassEX},
		{"comma", ',', ClassIS},
		{"period", '.', ClassIS},
		{"ideographic", '世', ClassID_EA}, // '世' has East Asian Width, returns EA variant
		{"hebrew", 'א', ClassHL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBreakClass(tt.r)
			if result != tt.expected {
				t.Errorf("getBreakClass(%q) = %v, want %v", tt.r, result, tt.expected)
			}
		})
	}
}

func TestHyphensConstants(t *testing.T) {
	if HyphensNone != 0 {
		t.Errorf("HyphensNone should be 0, got %d", HyphensNone)
	}
	if HyphensManual != 1 {
		t.Errorf("HyphensManual should be 1, got %d", HyphensManual)
	}
	if HyphensAuto != 2 {
		t.Errorf("HyphensAuto should be 2, got %d", HyphensAuto)
	}
}

func BenchmarkFindLineBreakOpportunities(b *testing.B) {
	text := "The quick brown fox jumps over the lazy dog. " +
		"Pack my box with five dozen liquor jugs. " +
		"How vexingly quick daft zebras jump!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindLineBreakOpportunities(text, HyphensManual)
	}
}

func BenchmarkFindLineBreakOpportunitiesCJK(b *testing.B) {
	text := "世界你好，这是一个测试文本。日本語のテキストも含まれています。"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindLineBreakOpportunities(text, HyphensManual)
	}
}
