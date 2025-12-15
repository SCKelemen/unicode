package uts51

import "testing"

// Test basic emoji property detection
func TestIsEmoji(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
		name     string
	}{
		// Emoji characters
		{'\U0001F600', true, "Grinning face"},
		{'\U0001F603', true, "Grinning face with big eyes"},
		{'😀', true, "Grinning face (literal)"},
		{'🎉', true, "Party popper"},

		// Emoji that can be used in sequences
		{'#', true, "Hash sign (keycap sequence)"},
		{'*', true, "Asterisk (keycap sequence)"},
		{'0', true, "Digit zero (keycap sequence)"},

		// Non-emoji characters
		{'A', false, "Latin letter A"},
		{'中', false, "CJK ideograph"},
		{' ', false, "Space"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsEmoji(tt.r)
			if got != tt.expected {
				t.Errorf("IsEmoji(%U %q) = %v, want %v", tt.r, tt.r, got, tt.expected)
			}
		})
	}
}

func TestHasEmojiPresentation(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
		name     string
	}{
		// Emoji with emoji presentation
		{'😀', true, "Grinning face"},
		{'🎉', true, "Party popper"},
		{'\U0001F600', true, "Grinning face (codepoint)"},

		// Emoji without emoji presentation (text default)
		{'☺', false, "Smiling face (text default)"},
		{'☹', false, "Frowning face (text default)"},
		{'✌', false, "Victory hand (text default)"},

		// Non-emoji
		{'A', false, "Latin letter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasEmojiPresentation(tt.r)
			if got != tt.expected {
				t.Errorf("HasEmojiPresentation(%U %q) = %v, want %v", tt.r, tt.r, got, tt.expected)
			}
		})
	}
}

func TestIsEmojiModifier(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
		name     string
	}{
		// Skin tone modifiers
		{'\U0001F3FB', true, "Light skin tone"},
		{'\U0001F3FC', true, "Medium-light skin tone"},
		{'\U0001F3FD', true, "Medium skin tone"},
		{'\U0001F3FE', true, "Medium-dark skin tone"},
		{'\U0001F3FF', true, "Dark skin tone"},

		// Not modifiers
		{'😀', false, "Grinning face"},
		{'👋', false, "Waving hand (base)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsEmojiModifier(tt.r)
			if got != tt.expected {
				t.Errorf("IsEmojiModifier(%U) = %v, want %v", tt.r, got, tt.expected)
			}
		})
	}
}

func TestIsEmojiModifierBase(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
		name     string
	}{
		// Characters that accept modifiers
		{'👋', true, "Waving hand"},
		{'👍', true, "Thumbs up"},
		{'✌', true, "Victory hand"},
		{'🤚', true, "Raised back of hand"},

		// Characters that don't accept modifiers
		{'😀', false, "Grinning face"},
		{'🎉', false, "Party popper"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsEmojiModifierBase(tt.r)
			if got != tt.expected {
				t.Errorf("IsEmojiModifierBase(%U %q) = %v, want %v", tt.r, tt.r, got, tt.expected)
			}
		})
	}
}

func TestIsRegionalIndicator(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
		name     string
	}{
		// Regional indicators (for flags)
		{'\U0001F1FA', true, "Regional indicator U"},
		{'\U0001F1F8', true, "Regional indicator S"},
		{'\U0001F1E6', true, "Regional indicator A (start)"},
		{'\U0001F1FF', true, "Regional indicator Z (end)"},

		// Not regional indicators
		{'😀', false, "Grinning face"},
		{'A', false, "Latin letter A"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRegionalIndicator(tt.r)
			if got != tt.expected {
				t.Errorf("IsRegionalIndicator(%U) = %v, want %v", tt.r, got, tt.expected)
			}
		})
	}
}

func TestEmojiWidth(t *testing.T) {
	tests := []struct {
		r        rune
		expected int
		name     string
	}{
		// Emoji with emoji presentation (2 columns)
		{'😀', 2, "Grinning face"},
		{'🎉', 2, "Party popper"},
		{'🌍', 2, "Earth globe Europe-Africa"},

		// Emoji with text presentation (1 column)
		{'☺', 1, "Smiling face"},
		{'✌', 1, "Victory hand"},

		// Emoji components (0 columns)
		{'\U0001F3FB', 0, "Light skin tone modifier"},
		{VariationSelector15, 0, "Text presentation selector"},
		{VariationSelector16, 0, "Emoji presentation selector"},
		{ZeroWidthJoiner, 0, "Zero width joiner"},

		// Non-emoji (0 columns)
		{'A', 0, "Latin letter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EmojiWidth(tt.r)
			if got != tt.expected {
				t.Errorf("EmojiWidth(%U %q) = %v, want %v", tt.r, tt.r, got, tt.expected)
			}
		})
	}
}

func TestDefaultPresentation(t *testing.T) {
	tests := []struct {
		r        rune
		expected rune
		name     string
	}{
		// Emoji presentation by default
		{'😀', 'E', "Grinning face"},
		{'🎉', 'E', "Party popper"},

		// Text presentation by default
		{'☺', 'T', "Smiling face"},
		{'✌', 'T', "Victory hand"},
		{'#', 'T', "Hash sign"},

		// Non-emoji
		{'A', 'T', "Latin letter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultPresentation(tt.r)
			if got != tt.expected {
				t.Errorf("DefaultPresentation(%U %q) = %q, want %q", tt.r, tt.r, got, tt.expected)
			}
		})
	}
}

func TestIsValidKeycapSequence(t *testing.T) {
	tests := []struct {
		name     string
		runes    []rune
		expected bool
	}{
		// Valid fully-qualified keycap sequences
		{"Keycap 9 (fully-qualified)", []rune{'9', VariationSelector16, CombiningEnclosingKeycap}, true},
		{"Keycap # (fully-qualified)", []rune{'#', VariationSelector16, CombiningEnclosingKeycap}, true},
		{"Keycap * (fully-qualified)", []rune{'*', VariationSelector16, CombiningEnclosingKeycap}, true},
		{"Keycap 0 (fully-qualified)", []rune{'0', VariationSelector16, CombiningEnclosingKeycap}, true},

		// Valid minimally-qualified keycap sequences (no FE0F)
		{"Keycap 9 (minimally-qualified)", []rune{'9', CombiningEnclosingKeycap}, true},
		{"Keycap # (minimally-qualified)", []rune{'#', CombiningEnclosingKeycap}, true},

		// Invalid sequences
		{"Not a keycap base", []rune{'A', VariationSelector16, CombiningEnclosingKeycap}, false},
		{"Missing keycap combining", []rune{'9', VariationSelector16}, false},
		{"Empty sequence", []rune{}, false},
		{"Single character", []rune{'9'}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidKeycapSequence(tt.runes)
			if got != tt.expected {
				t.Errorf("IsValidKeycapSequence(%v) = %v, want %v", tt.runes, got, tt.expected)
			}
		})
	}
}

func TestIsValidTagSequence(t *testing.T) {
	tests := []struct {
		name     string
		runes    []rune
		expected bool
	}{
		// Valid tag sequences (subdivision flags)
		{
			"England flag",
			[]rune{0x1F3F4, 0xE0067, 0xE0062, 0xE0065, 0xE006E, 0xE0067, TagTerminator},
			true,
		},
		{
			"Scotland flag",
			[]rune{0x1F3F4, 0xE0067, 0xE0062, 0xE0073, 0xE0063, 0xE0074, TagTerminator},
			true,
		},
		{
			"Wales flag",
			[]rune{0x1F3F4, 0xE0067, 0xE0062, 0xE0077, 0xE006C, 0xE0073, TagTerminator},
			true,
		},

		// Invalid sequences
		{"Missing terminator", []rune{0x1F3F4, 0xE0067, 0xE0062}, false},
		{"Non-emoji base", []rune{'A', 0xE0067, TagTerminator}, false},
		{"Empty sequence", []rune{}, false},
		{"Too short", []rune{0x1F3F4, TagTerminator}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTagSequence(tt.runes)
			if got != tt.expected {
				t.Errorf("IsValidTagSequence(%v) = %v, want %v", tt.runes, got, tt.expected)
			}
		})
	}
}

func TestIsValidEmojiSequence(t *testing.T) {
	tests := []struct {
		name     string
		runes    []rune
		expected bool
	}{
		// Single emoji
		{"Single emoji", []rune{'😀'}, true},

		// Keycap sequences
		{"Keycap 9", []rune{'9', VariationSelector16, CombiningEnclosingKeycap}, true},

		// Modifier sequences
		{"Waving hand + light skin", []rune{0x1F44B, 0x1F3FB}, true},
		{"Thumbs up + dark skin", []rune{0x1F44D, 0x1F3FF}, true},

		// Presentation sequences
		{"Emoji with text selector", []rune{'😀', VariationSelector15}, true},
		{"Emoji with emoji selector", []rune{'☺', VariationSelector16}, true},

		// Flag sequences
		{"US flag", []rune{0x1F1FA, 0x1F1F8}, true},
		{"UK flag", []rune{0x1F1EC, 0x1F1E7}, true},

		// ZWJ sequences
		{"Family ZWJ", []rune{0x1F468, ZeroWidthJoiner, 0x1F469, ZeroWidthJoiner, 0x1F467}, true},
		{"Kiss ZWJ", []rune{0x1F469, ZeroWidthJoiner, 0x2764, VariationSelector16, ZeroWidthJoiner, 0x1F48B, ZeroWidthJoiner, 0x1F468}, true},

		// Tag sequences
		{"England flag", []rune{0x1F3F4, 0xE0067, 0xE0062, 0xE0065, 0xE006E, 0xE0067, TagTerminator}, true},

		// Invalid sequences
		{"Empty", []rune{}, false},
		{"Non-emoji", []rune{'A'}, false},
		{"Invalid ZWJ", []rune{'A', ZeroWidthJoiner, 'B'}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidEmojiSequence(tt.runes)
			if got != tt.expected {
				t.Errorf("IsValidEmojiSequence(%v) = %v, want %v", tt.runes, got, tt.expected)
			}
		})
	}
}

// Benchmark property lookups
func BenchmarkIsEmoji(b *testing.B) {
	testRunes := []rune{'😀', '🎉', 'A', '中', '#'}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, r := range testRunes {
			IsEmoji(r)
		}
	}
}

func BenchmarkHasEmojiPresentation(b *testing.B) {
	testRunes := []rune{'😀', '☺', '✌', 'A'}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, r := range testRunes {
			HasEmojiPresentation(r)
		}
	}
}

func BenchmarkEmojiWidth(b *testing.B) {
	testRunes := []rune{'😀', '☺', '\U0001F3FB', 'A'}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, r := range testRunes {
			EmojiWidth(r)
		}
	}
}
