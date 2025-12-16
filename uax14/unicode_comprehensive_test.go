package uax14

import (
	"testing"
)

// TestUnicodeControlCharacters tests various Unicode control and format characters
func TestUnicodeControlCharacters(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		// Zero-width characters
		{"Zero Width Space", "hello\u200Bworld"},
		{"Zero Width Non-Joiner", "hello\u200Cworld"},
		{"Zero Width Joiner", "hello\u200Dworld"},
		{"Word Joiner", "hello\u2060world"},

		// Directional marks
		{"Left-to-Right Mark", "hello\u200Eworld"},
		{"Right-to-Left Mark", "hello\u200Fworld"},
		{"Left-to-Right Embedding", "hello\u202Aworld"},
		{"Right-to-Left Embedding", "hello\u202Bworld"},
		{"Pop Directional Formatting", "hello\u202Cworld"},
		{"Left-to-Right Override", "hello\u202Dworld"},
		{"Right-to-Left Override", "hello\u202Eworld"},

		// Line and paragraph separators
		{"Line Separator", "hello\u2028world"},
		{"Paragraph Separator", "hello\u2029world"},
		{"Next Line", "hello\u0085world"},

		// Other format characters
		{"Soft Hyphen", "hello\u00ADworld"},
		{"Non-Breaking Space", "hello\u00A0world"},
		{"Narrow No-Break Space", "hello\u202Fworld"},
		{"Zero Width No-Break Space (BOM)", "\uFEFFhello world"},

		// Variation selectors
		{"Variation Selector", "hello\uFE0Fworld"},
		{"Variation Selector-16", "a\uFE0F"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			breaks := FindLineBreakOpportunities(tt.text, HyphensManual)

			// Basic validation
			if len(breaks) < 2 {
				t.Errorf("Expected at least 2 breaks (start and end), got %d", len(breaks))
			}
			if breaks[0] != 0 {
				t.Errorf("First break should be 0, got %d", breaks[0])
			}
			if breaks[len(breaks)-1] != len(tt.text) {
				t.Errorf("Last break should be %d, got %d", len(tt.text), breaks[len(breaks)-1])
			}

			// Check ascending order
			for i := 1; i < len(breaks); i++ {
				if breaks[i] <= breaks[i-1] {
					t.Errorf("Breaks not in ascending order at %d: %v", i, breaks)
				}
			}
		})
	}
}

// TestAsianScripts tests various Asian writing systems
func TestAsianScripts(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		// Chinese (Simplified and Traditional)
		{"Chinese Simplified", "你好世界这是一个测试"},
		{"Chinese Traditional", "繁體中文測試文本"},
		{"Chinese with punctuation", "你好，世界！这是测试。"},
		{"Chinese mixed with English", "Hello 你好 world 世界"},

		// Japanese
		{"Japanese Hiragana", "こんにちは"},
		{"Japanese Katakana", "カタカナ"},
		{"Japanese Kanji", "日本語"},
		{"Japanese mixed", "こんにちは、世界！"},
		{"Japanese with English", "Hello こんにちは world"},

		// Korean
		{"Korean Hangul", "안녕하세요"},
		{"Korean with spaces", "한국어 텍스트 테스트"},
		{"Korean mixed", "Hello 안녕 world"},

		// Thai
		{"Thai", "สวัสดีครับ"},
		{"Thai with spaces", "ภาษา ไทย"},

		// Myanmar
		{"Myanmar", "မြန်မာ"},

		// Tibetan
		{"Tibetan", "བོད་ཡིག"},

		// Mongolian
		{"Mongolian", "ᠮᠣᠩᠭᠣᠯ"},

		// Vietnamese
		{"Vietnamese", "Tiếng Việt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			breaks := FindLineBreakOpportunities(tt.text, HyphensManual)

			// Validation
			if len(breaks) < 2 {
				t.Errorf("Expected at least 2 breaks, got %d", len(breaks))
			}
			if breaks[0] != 0 || breaks[len(breaks)-1] != len(tt.text) {
				t.Errorf("Invalid start/end breaks: %v", breaks)
			}

			// For CJK ideographic text, expect more break opportunities
			// (CJK allows breaks between characters)
			hasCJK := false
			for _, r := range tt.text {
				if (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
					(r >= 0x3400 && r <= 0x4DBF) { // CJK Extension A
					hasCJK = true
					break
				}
			}

			if hasCJK && len(breaks) < 3 {
				// CJK text should have multiple break opportunities
				t.Logf("CJK text with limited breaks: %q -> %v", tt.text, breaks)
			}
		})
	}
}

// TestMiddleEasternScripts tests Arabic, Hebrew, and related scripts
func TestMiddleEasternScripts(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		// Arabic
		{"Arabic", "مرحبا بك"},
		{"Arabic with punctuation", "مرحبا، كيف حالك؟"},
		{"Arabic mixed with English", "Hello مرحبا world"},
		{"Arabic numerals", "العدد ١٢٣٤"},

		// Hebrew
		{"Hebrew", "שלום עולם"},
		{"Hebrew with punctuation", "שלום, מה שלומך?"},
		{"Hebrew mixed with English", "Hello שלום world"},

		// Persian/Farsi
		{"Persian", "سلام دنیا"},

		// Urdu
		{"Urdu", "ہیلو دنیا"},

		// Pashto
		{"Pashto", "سلام"},

		// Syriac
		{"Syriac", "ܫܠܡܐ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			breaks := FindLineBreakOpportunities(tt.text, HyphensManual)

			// Basic validation
			if len(breaks) < 2 {
				t.Errorf("Expected at least 2 breaks, got %d", len(breaks))
			}
			if breaks[0] != 0 || breaks[len(breaks)-1] != len(tt.text) {
				t.Errorf("Invalid start/end breaks: %v", breaks)
			}

			// RTL text with spaces should allow breaks at spaces
			// (Just like LTR text)
		})
	}
}

// TestEmojiAndSymbols tests emoji, symbols, and pictographs
func TestEmojiAndSymbols(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		// Basic emoji
		{"Single emoji", "Hello 😀 world"},
		{"Multiple emoji", "😀😁😂"},

		// Emoji with modifiers
		{"Emoji with skin tone", "👋🏻"},
		{"Emoji with skin tone medium", "👋🏽"},

		// ZWJ sequences (complex emoji)
		{"Family emoji", "👨‍👩‍👧‍👦"},
		{"Kiss emoji", "👩‍❤️‍💋‍👨"},
		{"Couple emoji", "👨‍❤️‍👨"},

		// Regional indicators (flags)
		{"Flag USA", "🇺🇸"},
		{"Flag Japan", "🇯🇵"},
		{"Multiple flags", "🇺🇸 🇬🇧 🇫🇷"},

		// Emoji with variation selectors
		{"Heart plain", "❤"},
		{"Heart emoji style", "❤️"},
		{"Star plain", "⭐"},
		{"Star emoji style", "⭐️"},

		// Mathematical symbols
		{"Math symbols", "∑∫∂√∞"},
		{"Math mixed with text", "The sum ∑ of x"},

		// Currency symbols
		{"Currency symbols", "$€£¥₹"},

		// Arrows and shapes
		{"Arrows", "→←↑↓"},
		{"Shapes", "■□●○"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			breaks := FindLineBreakOpportunities(tt.text, HyphensManual)

			// Basic validation
			if len(breaks) < 2 {
				t.Errorf("Expected at least 2 breaks, got %d", len(breaks))
			}
			if breaks[0] != 0 || breaks[len(breaks)-1] != len(tt.text) {
				t.Errorf("Invalid start/end breaks: %v", breaks)
			}

			// Emoji sequences with ZWJ should not break within the sequence
			// (handled by combining mark rules)
		})
	}
}

// TestVerticalTextCharacters tests characters commonly used in vertical text
func TestVerticalTextCharacters(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		// Vertical forms
		{"Vertical punctuation", "︰︱︲︳"},
		{"Vertical brackets", "︵︶︷︸"},

		// Bopomofo (used in vertical traditional Chinese)
		{"Bopomofo", "ㄅㄆㄇㄈ"},

		// Vertical text compatibility
		{"CJK vertical", "縦書き"},

		// Mongolian (traditionally vertical)
		{"Mongolian vertical", "ᠮᠣᠩᠭᠣᠯ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			breaks := FindLineBreakOpportunities(tt.text, HyphensManual)

			// Basic validation
			if len(breaks) < 2 {
				t.Errorf("Expected at least 2 breaks, got %d", len(breaks))
			}
			if breaks[0] != 0 || breaks[len(breaks)-1] != len(tt.text) {
				t.Errorf("Invalid start/end breaks: %v", breaks)
			}
		})
	}
}

// TestComplexScripts tests scripts with complex shaping
func TestComplexScripts(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		// Devanagari
		{"Hindi", "नमस्ते दुनिया"},
		{"Sanskrit", "संस्कृतम्"},

		// Bengali
		{"Bengali", "হ্যালো"},

		// Tamil
		{"Tamil", "வணக்கம்"},

		// Telugu
		{"Telugu", "హలో"},

		// Malayalam
		{"Malayalam", "ഹലോ"},

		// Kannada
		{"Kannada", "ಹಲೋ"},

		// Gujarati
		{"Gujarati", "હલો"},

		// Gurmukhi (Punjabi)
		{"Gurmukhi", "ਸਤਿ ਸ੍ਰੀ ਅਕਾਲ"},

		// Sinhala
		{"Sinhala", "හෙලෝ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			breaks := FindLineBreakOpportunities(tt.text, HyphensManual)

			// Basic validation
			if len(breaks) < 2 {
				t.Errorf("Expected at least 2 breaks, got %d", len(breaks))
			}
			if breaks[0] != 0 || breaks[len(breaks)-1] != len(tt.text) {
				t.Errorf("Invalid start/end breaks: %v", breaks)
			}
		})
	}
}

// TestMixedDirectionality tests BiDi text mixing LTR and RTL
func TestMixedDirectionality(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{"English + Arabic", "Hello مرحبا world"},
		{"English + Hebrew", "Hello שלום world"},
		{"Arabic + English + Arabic", "مرحبا Hello مرحبا"},
		{"Hebrew + English + Hebrew", "שלום Hello שלום"},
		{"Mixed with numbers", "Hello 123 مرحبا 456"},
		{"URL in Arabic", "Check مرحبا https://example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			breaks := FindLineBreakOpportunities(tt.text, HyphensManual)

			// Should handle BiDi text without panicking
			if len(breaks) < 2 {
				t.Errorf("Expected at least 2 breaks, got %d", len(breaks))
			}
			if breaks[0] != 0 || breaks[len(breaks)-1] != len(tt.text) {
				t.Errorf("Invalid start/end breaks: %v", breaks)
			}
		})
	}
}
