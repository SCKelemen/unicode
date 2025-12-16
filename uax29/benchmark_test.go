package uax29

import (
	"testing"
)

// Test text samples for benchmarking
const (
	benchTextShort  = "Hello, world! How are you today?"
	benchTextMedium = "The quick brown fox jumps over the lazy dog. This is a test sentence. How wonderful!"
	benchTextLong   = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.`
	benchTextUnicode = "Hello 世界! Привет мир! שלום עולם! 👨‍👩‍👧‍👦🇺🇸"
)

// Benchmarks for Grapheme breaks

func BenchmarkGraphemesShort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Graphemes(benchTextShort)
	}
}

func BenchmarkGraphemesMedium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Graphemes(benchTextMedium)
	}
}

func BenchmarkGraphemesLong(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Graphemes(benchTextLong)
	}
}

func BenchmarkGraphemesUnicode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Graphemes(benchTextUnicode)
	}
}

func BenchmarkFindGraphemeBreaksShort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FindGraphemeBreaks(benchTextShort)
	}
}

func BenchmarkFindGraphemeBreaksMedium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FindGraphemeBreaks(benchTextMedium)
	}
}

func BenchmarkFindGraphemeBreaksLong(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FindGraphemeBreaks(benchTextLong)
	}
}

// Benchmarks for Word breaks

func BenchmarkWordsShort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Words(benchTextShort)
	}
}

func BenchmarkWordsMedium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Words(benchTextMedium)
	}
}

func BenchmarkWordsLong(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Words(benchTextLong)
	}
}

func BenchmarkWordsUnicode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Words(benchTextUnicode)
	}
}

func BenchmarkFindWordBreaksShort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FindWordBreaks(benchTextShort)
	}
}

func BenchmarkFindWordBreaksMedium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FindWordBreaks(benchTextMedium)
	}
}

func BenchmarkFindWordBreaksLong(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FindWordBreaks(benchTextLong)
	}
}

// Benchmarks for Sentence breaks

func BenchmarkSentencesShort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Sentences(benchTextShort)
	}
}

func BenchmarkSentencesMedium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Sentences(benchTextMedium)
	}
}

func BenchmarkSentencesLong(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Sentences(benchTextLong)
	}
}

func BenchmarkSentencesUnicode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Sentences(benchTextUnicode)
	}
}

func BenchmarkFindSentenceBreaksShort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FindSentenceBreaks(benchTextShort)
	}
}

func BenchmarkFindSentenceBreaksMedium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FindSentenceBreaks(benchTextMedium)
	}
}

func BenchmarkFindSentenceBreaksLong(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FindSentenceBreaks(benchTextLong)
	}
}

// Benchmark classification functions directly

func BenchmarkClassifyRune(b *testing.B) {
	runes := []rune{'a', '世', 'א', '👨', '\n', '.', ' '}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, r := range runes {
			_ = classifyRune(r)
		}
	}
}

func BenchmarkGetGraphemeBreakClass(b *testing.B) {
	runes := []rune{'a', '世', 'א', '👨', '\n', '.', ' '}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, r := range runes {
			_ = getGraphemeBreakClass(r)
		}
	}
}

func BenchmarkGetWordBreakClass(b *testing.B) {
	runes := []rune{'a', '世', 'א', '👨', '\n', '.', ' '}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, r := range runes {
			_ = getWordBreakClass(r)
		}
	}
}

func BenchmarkGetSentenceBreakClass(b *testing.B) {
	runes := []rune{'a', '世', 'א', '👨', '\n', '.', ' '}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, r := range runes {
			_ = getSentenceBreakClass(r)
		}
	}
}

// Tests for single-pass FindAllBreaks

func TestFindAllBreaks(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Empty", ""},
		{"Simple", "Hello, world!"},
		{"Unicode", "Hello 世界! שלום"},
		{"Emoji", "Hello 👨‍👩‍👧‍👦 world"},
		{"Mixed", benchTextMedium},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get results from single-pass
			result := FindAllBreaks(tt.input)

			// Get results from separate passes
			expectedG := FindGraphemeBreaks(tt.input)
			expectedW := FindWordBreaks(tt.input)
			expectedS := FindSentenceBreaks(tt.input)

			// Verify they match
			if !equalSlices(result.Graphemes, expectedG) {
				t.Errorf("Grapheme breaks mismatch\n  Got:      %v\n  Expected: %v", result.Graphemes, expectedG)
			}
			if !equalSlices(result.Words, expectedW) {
				t.Errorf("Word breaks mismatch\n  Got:      %v\n  Expected: %v", result.Words, expectedW)
			}
			if !equalSlices(result.Sentences, expectedS) {
				t.Errorf("Sentence breaks mismatch\n  Got:      %v\n  Expected: %v", result.Sentences, expectedS)
			}

			// Verify hierarchical property: Words ⊆ Graphemes, Sentences ⊆ Words
			if !isSubset(result.Words, result.Graphemes) {
				t.Errorf("Word breaks are not a subset of grapheme breaks")
			}
			if !isSubset(result.Sentences, result.Words) {
				t.Errorf("Sentence breaks are not a subset of word breaks")
			}
		})
	}
}

func equalSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func isSubset(subset, superset []int) bool {
	j := 0
	for _, v := range subset {
		for j < len(superset) && superset[j] < v {
			j++
		}
		if j >= len(superset) || superset[j] != v {
			return false
		}
		j++
	}
	return true
}

// Benchmarks for single-pass FindAllBreaks

func BenchmarkFindAllBreaksShort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FindAllBreaks(benchTextShort)
	}
}

func BenchmarkFindAllBreaksMedium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FindAllBreaks(benchTextMedium)
	}
}

func BenchmarkFindAllBreaksLong(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FindAllBreaks(benchTextLong)
	}
}

func BenchmarkFindAllBreaksUnicode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FindAllBreaks(benchTextUnicode)
	}
}

// Benchmark comparison: single-pass vs three separate calls

func BenchmarkThreeSeparatePassesShort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FindGraphemeBreaks(benchTextShort)
		_ = FindWordBreaks(benchTextShort)
		_ = FindSentenceBreaks(benchTextShort)
	}
}

func BenchmarkThreeSeparatePassesMedium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FindGraphemeBreaks(benchTextMedium)
		_ = FindWordBreaks(benchTextMedium)
		_ = FindSentenceBreaks(benchTextMedium)
	}
}

func BenchmarkThreeSeparatePassesLong(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FindGraphemeBreaks(benchTextLong)
		_ = FindWordBreaks(benchTextLong)
		_ = FindSentenceBreaks(benchTextLong)
	}
}
