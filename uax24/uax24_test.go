package uax24

import (
	"testing"
)

func TestLookupScript(t *testing.T) {
	tests := []struct {
		r        rune
		expected Script
		name     string
	}{
		// Latin script
		{'A', ScriptLatin, "Latin uppercase"},
		{'z', ScriptLatin, "Latin lowercase"},
		{'\u00C0', ScriptLatin, "Latin with diacritic (À)"},
		{'\u1E00', ScriptLatin, "Latin Extended Additional"},

		// Greek script
		{'\u0391', ScriptGreek, "Greek uppercase Alpha"},
		{'\u03B1', ScriptGreek, "Greek lowercase alpha"},
		{'\u03C0', ScriptGreek, "Greek lowercase pi"},

		// Cyrillic script
		{'\u0410', ScriptCyrillic, "Cyrillic uppercase A"},
		{'\u0430', ScriptCyrillic, "Cyrillic lowercase a"},
		{'\u0416', ScriptCyrillic, "Cyrillic Zhe"},

		// Han (CJK) script
		{'\u4E00', ScriptHan, "CJK Unified Ideograph"},
		{'\u9FA5', ScriptHan, "CJK Unified Ideograph"},
		{'中', ScriptHan, "Chinese character"},

		// Hiragana
		{'\u3041', ScriptHiragana, "Hiragana small a"},
		{'\u3093', ScriptHiragana, "Hiragana n"},

		// Katakana
		{'\u30A1', ScriptKatakana, "Katakana small a"},
		{'\u30F6', ScriptKatakana, "Katakana ke"},

		// Hangul (Korean)
		{'\uAC00', ScriptHangul, "Hangul syllable GA"},
		{'\uD7A3', ScriptHangul, "Hangul syllable HIH"},

		// Arabic
		{'\u0627', ScriptArabic, "Arabic letter Alef"},
		{'\u0628', ScriptArabic, "Arabic letter Beh"},

		// Hebrew
		{'\u05D0', ScriptHebrew, "Hebrew letter Alef"},
		{'\u05D1', ScriptHebrew, "Hebrew letter Bet"},

		// Devanagari
		{'\u0905', ScriptDevanagari, "Devanagari letter A"},
		{'\u0915', ScriptDevanagari, "Devanagari letter Ka"},

		// Bengali
		{'\u0985', ScriptBengali, "Bengali letter A"},
		{'\u0995', ScriptBengali, "Bengali letter Ka"},

		// Thai
		{'\u0E01', ScriptThai, "Thai letter Ko Kai"},
		{'\u0E2F', ScriptThai, "Thai letter Paiyannoi"},

		// Common script (shared across scripts)
		{'0', ScriptCommon, "ASCII digit"},
		{'9', ScriptCommon, "ASCII digit"},
		{' ', ScriptCommon, "Space"},
		{',', ScriptCommon, "Comma"},
		{'.', ScriptCommon, "Period"},
		{'!', ScriptCommon, "Exclamation"},
		{'?', ScriptCommon, "Question mark"},

		// Inherited script (combining marks)
		{'\u0300', ScriptInherited, "Combining grave accent"},
		{'\u0301', ScriptInherited, "Combining acute accent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LookupScript(tt.r)
			if result != tt.expected {
				t.Errorf("LookupScript(%U) = %v, want %v", tt.r, result, tt.expected)
			}
		})
	}
}

func TestIsCommon(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
	}{
		{'0', true},
		{'9', true},
		{' ', true},
		{',', true},
		{'A', false},
		{'中', false},
		{'\u0410', false},
	}

	for _, tt := range tests {
		result := IsCommon(tt.r)
		if result != tt.expected {
			t.Errorf("IsCommon(%U) = %v, want %v", tt.r, result, tt.expected)
		}
	}
}

func TestIsInherited(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
	}{
		{'\u0300', true}, // Combining grave accent
		{'\u0301', true}, // Combining acute accent
		{'A', false},
		{'0', false},
	}

	for _, tt := range tests {
		result := IsInherited(tt.r)
		if result != tt.expected {
			t.Errorf("IsInherited(%U) = %v, want %v", tt.r, result, tt.expected)
		}
	}
}

func TestIsLatin(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
	}{
		{'A', true},
		{'z', true},
		{'\u00C0', true}, // À
		{'0', false},
		{'中', false},
	}

	for _, tt := range tests {
		result := IsLatin(tt.r)
		if result != tt.expected {
			t.Errorf("IsLatin(%U) = %v, want %v", tt.r, result, tt.expected)
		}
	}
}

func TestIsGreek(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
	}{
		{'\u0391', true}, // Α (Alpha)
		{'\u03B1', true}, // α (alpha)
		{'\u03C0', true}, // π (pi)
		{'A', false},
		{'中', false},
	}

	for _, tt := range tests {
		result := IsGreek(tt.r)
		if result != tt.expected {
			t.Errorf("IsGreek(%U) = %v, want %v", tt.r, result, tt.expected)
		}
	}
}

func TestIsCyrillic(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
	}{
		{'\u0410', true}, // А (Cyrillic A)
		{'\u0430', true}, // а (Cyrillic a)
		{'\u0416', true}, // Ж (Zhe)
		{'A', false},
		{'中', false},
	}

	for _, tt := range tests {
		result := IsCyrillic(tt.r)
		if result != tt.expected {
			t.Errorf("IsCyrillic(%U) = %v, want %v", tt.r, result, tt.expected)
		}
	}
}

func TestIsHan(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
	}{
		{'\u4E00', true}, // CJK ideograph
		{'中', true},
		{'日', true},
		{'本', true},
		{'A', false},
		{'\u3041', false}, // Hiragana
	}

	for _, tt := range tests {
		result := IsHan(tt.r)
		if result != tt.expected {
			t.Errorf("IsHan(%U) = %v, want %v", tt.r, result, tt.expected)
		}
	}
}

func TestIsHiragana(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
	}{
		{'\u3041', true}, // ぁ
		{'\u3093', true}, // ん
		{'\u30A1', false}, // Katakana
		{'A', false},
	}

	for _, tt := range tests {
		result := IsHiragana(tt.r)
		if result != tt.expected {
			t.Errorf("IsHiragana(%U) = %v, want %v", tt.r, result, tt.expected)
		}
	}
}

func TestIsKatakana(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
	}{
		{'\u30A1', true}, // ァ
		{'\u30F6', true}, // ヶ
		{'\u3041', false}, // Hiragana
		{'A', false},
	}

	for _, tt := range tests {
		result := IsKatakana(tt.r)
		if result != tt.expected {
			t.Errorf("IsKatakana(%U) = %v, want %v", tt.r, result, tt.expected)
		}
	}
}

func TestIsArabic(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
	}{
		{'\u0627', true}, // Arabic Alef
		{'\u0628', true}, // Arabic Beh
		{'A', false},
	}

	for _, tt := range tests {
		result := IsArabic(tt.r)
		if result != tt.expected {
			t.Errorf("IsArabic(%U) = %v, want %v", tt.r, result, tt.expected)
		}
	}
}

func TestIsHebrew(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
	}{
		{'\u05D0', true}, // Hebrew Alef
		{'\u05D1', true}, // Hebrew Bet
		{'A', false},
	}

	for _, tt := range tests {
		result := IsHebrew(tt.r)
		if result != tt.expected {
			t.Errorf("IsHebrew(%U) = %v, want %v", tt.r, result, tt.expected)
		}
	}
}

func TestIsDevanagari(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
	}{
		{'\u0905', true}, // Devanagari A
		{'\u0915', true}, // Devanagari Ka
		{'A', false},
	}

	for _, tt := range tests {
		result := IsDevanagari(tt.r)
		if result != tt.expected {
			t.Errorf("IsDevanagari(%U) = %v, want %v", tt.r, result, tt.expected)
		}
	}
}

func TestIsBengali(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
	}{
		{'\u0985', true}, // Bengali A
		{'\u0995', true}, // Bengali Ka
		{'A', false},
	}

	for _, tt := range tests {
		result := IsBengali(tt.r)
		if result != tt.expected {
			t.Errorf("IsBengali(%U) = %v, want %v", tt.r, result, tt.expected)
		}
	}
}

func TestIsThai(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
	}{
		{'\u0E01', true}, // Thai Ko Kai
		{'\u0E2F', true}, // Thai Paiyannoi
		{'A', false},
	}

	for _, tt := range tests {
		result := IsThai(tt.r)
		if result != tt.expected {
			t.Errorf("IsThai(%U) = %v, want %v", tt.r, result, tt.expected)
		}
	}
}

func TestAnalyzeScripts(t *testing.T) {
	tests := []struct {
		input         string
		expectedCount int
		expectedMixed bool
		hasCommon     bool
		description   string
	}{
		{"Hello", 1, false, false, "Pure Latin"},
		{"Hello123", 1, false, true, "Latin with digits"},
		{"Hello мир", 2, true, true, "Mixed Latin and Cyrillic"},
		{"中文", 1, false, false, "Pure Han"},
		{"こんにちは", 1, false, false, "Pure Hiragana"},
		{"カタカナ", 1, false, false, "Pure Katakana"},
		{"Hello世界", 2, true, false, "Mixed Latin and Han"},
		{"Привет שלום", 2, true, true, "Mixed Cyrillic and Hebrew"},
		{"123456", 0, false, true, "Only digits (Common)"},
		{"", 0, false, false, "Empty string"},
		{"A\u0300", 1, false, false, "Latin with combining mark"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			info := AnalyzeScripts(tt.input)

			if len(info.Scripts) != tt.expectedCount {
				t.Errorf("AnalyzeScripts(%q): got %d scripts, want %d. Scripts: %v",
					tt.input, len(info.Scripts), tt.expectedCount, info.Scripts)
			}

			if info.IsMixedScript != tt.expectedMixed {
				t.Errorf("AnalyzeScripts(%q): IsMixedScript = %v, want %v",
					tt.input, info.IsMixedScript, tt.expectedMixed)
			}

			if info.HasCommon != tt.hasCommon {
				t.Errorf("AnalyzeScripts(%q): HasCommon = %v, want %v",
					tt.input, info.HasCommon, tt.hasCommon)
			}
		})
	}
}

func TestIsSingleScript(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"Hello", true},
		{"Hello123", true},       // Common doesn't count
		{"Hello мир", false},     // Latin + Cyrillic
		{"中文", true},
		{"こんにちは", true},
		{"Hello世界", false},      // Latin + Han
		{"", true},
		{"123", true},            // Only Common
	}

	for _, tt := range tests {
		result := IsSingleScript(tt.input)
		if result != tt.expected {
			t.Errorf("IsSingleScript(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestScriptString(t *testing.T) {
	tests := []struct {
		script   Script
		expected string
	}{
		{ScriptUnknown, "Unknown"},
		{ScriptCommon, "Common"},
		{ScriptInherited, "Inherited"},
		{ScriptLatin, "Latin"},
		{ScriptGreek, "Greek"},
		{ScriptCyrillic, "Cyrillic"},
		{ScriptHan, "Han"},
		{ScriptArabic, "Arabic"},
		{ScriptHebrew, "Hebrew"},
		{ScriptDevanagari, "Devanagari"},
	}

	for _, tt := range tests {
		result := tt.script.String()
		if result != tt.expected {
			t.Errorf("Script(%d).String() = %q, want %q", tt.script, result, tt.expected)
		}
	}
}

// Benchmark tests
func BenchmarkLookupScript(b *testing.B) {
	testRunes := []rune{
		'A',      // Latin
		'中',     // Han
		'\u0410', // Cyrillic
		'\u0391', // Greek
		'0',      // Common
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, r := range testRunes {
			LookupScript(r)
		}
	}
}

func BenchmarkAnalyzeScripts(b *testing.B) {
	text := "Hello мир 世界 שלום"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AnalyzeScripts(text)
	}
}

func BenchmarkIsSingleScript(b *testing.B) {
	text := "HelloWorld123"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsSingleScript(text)
	}
}
