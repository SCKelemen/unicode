package uts39

import (
	"strings"
	"testing"
)

// TestSkeletonCaseFold covers Bug B: the Skeleton algorithm must apply
// Unicode "full" case folding (CaseFolding.txt status C+F) rather than the
// previously used strings.ToLower, which cannot handle expansions like
// "ß" → "ss" or "ﬃ" → "ffi".
func TestSkeletonCaseFold(t *testing.T) {
	// "ß" (U+00DF LATIN SMALL LETTER SHARP S) full-case-folds to "ss"; once
	// folded it should produce the same skeleton as the literal ASCII "ss".
	t.Run("sharp s folds to ss", func(t *testing.T) {
		skelSharp := Skeleton("ß")
		skelSS := Skeleton("ss")
		if skelSharp != "ss" {
			t.Errorf("Skeleton(%q) = %q, want %q", "ß", skelSharp, "ss")
		}
		if skelSharp != skelSS {
			t.Errorf("Skeleton(%q) (%q) != Skeleton(%q) (%q)",
				"ß", skelSharp, "ss", skelSS)
		}
		if !AreConfusable("ß", "ss") {
			t.Errorf("AreConfusable(%q, %q) = false, want true", "ß", "ss")
		}
		// And uppercase "SS" should also match after case folding.
		if !AreConfusable("ß", "SS") {
			t.Errorf("AreConfusable(%q, %q) = false, want true", "ß", "SS")
		}
	})

	// "ﬃ" (U+FB03 LATIN SMALL LIGATURE FFI) full-case-folds via NFKD to "ffi".
	// This also exercises the NFKD step: the ligature decomposes to "ffi"
	// under compatibility decomposition. With proper full folding the result
	// should equal Skeleton("ffi").
	t.Run("ffi ligature folds to ffi", func(t *testing.T) {
		skelLig := Skeleton("\uFB03") // ﬃ
		skelFFI := Skeleton("ffi")
		if skelLig != skelFFI {
			t.Errorf("Skeleton(%q) (%q) != Skeleton(%q) (%q)",
				"\uFB03", skelLig, "ffi", skelFFI)
		}
	})

	// "İ" (U+0130 LATIN CAPITAL LETTER I WITH DOT ABOVE) full-case-folds
	// (non-Turkic) to "i\u0307" (LATIN SMALL LETTER I + COMBINING DOT ABOVE),
	// not to plain "i". This test confirms the folding actually runs (versus
	// the old strings.ToLower path which produced just "i").
	t.Run("Turkish-I folds via full case folding", func(t *testing.T) {
		skel := Skeleton("İ")
		// Skeleton must contain the COMBINING DOT ABOVE that full case
		// folding introduces; if it does not, we are still using ToLower.
		if !strings.ContainsRune(skel, '\u0307') {
			t.Errorf("Skeleton(%q) = %q; expected to contain U+0307 "+
				"COMBINING DOT ABOVE from full case folding", "İ", skel)
		}
		// Per UTS #39 §4 with full (non-Turkic) folding, "İ" and "I" do not
		// produce identical skeletons. This is the documented spec behavior;
		// we assert it explicitly so a future change does not silently flip it.
		if AreConfusable("I", "İ") {
			t.Errorf("AreConfusable(%q, %q) = true; expected false under "+
				"full (non-Turkic) case folding per UTS #39 §4",
				"I", "İ")
		}
	})
}

// TestCaseFoldHelper exercises the unexported caseFold helper directly to
// verify its behavior independently of the surrounding Skeleton pipeline.
func TestCaseFoldHelper(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"abc", "abc"},     // ASCII fast-path no-op
		{"ABC", "abc"},     // ASCII fast-path fold
		{"Hello", "hello"}, // mixed-case ASCII
		{"ß", "ss"},        // expansion via F mapping
		{"İ", "i\u0307"},   // F mapping with combining mark
		{"\u00DF\u00DF", "ssss"},
		{"\u0130A", "i\u0307a"}, // mixed
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := caseFold(tt.in)
			if got != tt.want {
				t.Errorf("caseFold(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestSkeletonIdempotent verifies that Skeleton is idempotent on its output,
// which is part of how Bug C (the fixed iteration cap) is justified: if the
// transform is idempotent, looping to a fixed point is well-defined.
func TestSkeletonIdempotent(t *testing.T) {
	inputs := []string{
		"",
		"hello",
		"HELLO",
		"ß",
		"İ",
		"paypal",
		"p\u0430ypal",                    // Cyrillic а
		"\u0455\u0441\u043E\u0440\u0435", // Cyrillic ѕсоре
		"café",
		"cafe\u0301",
		"\uFB03ce", // ligature ﬃ + "ce"
		"hello世界",
		"hello漢字ㅎ",
	}
	for _, in := range inputs {
		t.Run(in, func(t *testing.T) {
			once := Skeleton(in)
			twice := Skeleton(once)
			if once != twice {
				t.Errorf("Skeleton not idempotent for %q: first=%q, second=%q",
					in, once, twice)
			}
		})
	}
}

// TestGetRestrictionLevelTable1 covers each restriction level in UAX #39
// §5.2 Table 1 with at least one representative input. This is intentionally
// duplicative with TestGetRestrictionLevel (in uts39_test.go) so the levels
// remain pinned even if the existing tests drift.
func TestGetRestrictionLevelTable1(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  RestrictionLevel
	}{
		// ASCIIOnly
		{"ASCII only", "hello", ASCIIOnly},
		{"ASCII with digits", "user_123", ASCIIOnly},

		// SingleScript
		{"Single Latin", "café", SingleScript},
		{"Single Cyrillic", "привет", SingleScript},
		{"Single Han", "你好", SingleScript},
		{"Single Greek", "γεια", SingleScript},
		{"Single Arabic", "مرحبا", SingleScript},

		// HighlyRestrictive: Latn + Jpan / Hanb / Kore
		{"Latin + Han (Latn+Jpan subset)", "hello世界", HighlyRestrictive},
		{"Latin + Han + Hiragana", "hello漢字ひらがな", HighlyRestrictive},
		{"Latin + Han + Katakana", "helloカタカナ漢", HighlyRestrictive},
		{"Latin + Han + Hangul", "hello漢字ㅎ", HighlyRestrictive},
		{"Latin + Han + Bopomofo", "hello漢ㄅ", HighlyRestrictive},
		{"Latin + Han + Hiragana + Katakana", "hi漢ひカ", HighlyRestrictive},

		// ModeratelyRestrictive: Latin + exactly one other non-{Cyrillic,Greek}
		// recommended script.
		{"Latin + Arabic", "helloمرحبا", ModeratelyRestrictive},
		{"Latin + Hebrew", "helloשלום", ModeratelyRestrictive},
		{"Latin + Devanagari", "helloनमस्ते", ModeratelyRestrictive},
		{"Latin + Thai", "helloสวัสดี", ModeratelyRestrictive},

		// MinimallyRestrictive: catch-all for other multi-script combinations.
		{"Latin + Cyrillic", "hello мир", MinimallyRestrictive},
		{"Latin + Greek", "helloαβ", MinimallyRestrictive},
		{"Latin + Cyrillic + Greek", "helloмαβ", MinimallyRestrictive},
		{"Latin + Arabic + Hebrew",
			"helloمرحباשלום", MinimallyRestrictive},

		// Unrestricted: empty fallback.
		{"Empty string", "", Unrestricted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRestrictionLevel(tt.input)
			if got != tt.want {
				t.Errorf("GetRestrictionLevel(%q) = %v, want %v",
					tt.input, got, tt.want)
			}
		})
	}
}
