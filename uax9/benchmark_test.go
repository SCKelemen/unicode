package uax9

import (
	"testing"
)

// Test runes covering different bidi classes
var benchRunes = []rune{
	'a',     // ClassL (Latin)
	'א',     // ClassR (Hebrew)
	'ع',     // ClassAL (Arabic)
	'0',     // ClassEN (European Number)
	'+',     // ClassES (European Separator)
	'$',     // ClassET (European Terminator)
	'٠',     // ClassAN (Arabic Number) U+0660
	',',     // ClassCS (Common Separator)
	'\u0300', // ClassNSM (Nonspacing Mark)
	'\n',    // ClassB (Paragraph Separator)
	'\t',    // ClassS (Segment Separator)
	' ',     // ClassWS (Whitespace)
	'!',     // ClassON (Other Neutral)
	'\u202A', // ClassLRE (Left-to-Right Embedding)
	'\u2066', // ClassLRI (Left-to-Right Isolate)
}

func BenchmarkGetBidiClass(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, r := range benchRunes {
			_ = GetBidiClass(r)
		}
	}
}

func BenchmarkGetBidiClassASCII(b *testing.B) {
	ascii := []rune("Hello, World! 123")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, r := range ascii {
			_ = GetBidiClass(r)
		}
	}
}

func BenchmarkGetBidiClassHebrew(b *testing.B) {
	hebrew := []rune("שלום עולם")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, r := range hebrew {
			_ = GetBidiClass(r)
		}
	}
}

func BenchmarkGetBidiClassArabic(b *testing.B) {
	arabic := []rune("مرحبا بالعالم")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, r := range arabic {
			_ = GetBidiClass(r)
		}
	}
}

func BenchmarkGetBidiClassMixed(b *testing.B) {
	mixed := []rune("Hello שלום مرحبا 123")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, r := range mixed {
			_ = GetBidiClass(r)
		}
	}
}

func BenchmarkReorderShort(b *testing.B) {
	text := "Hello שלום world"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Reorder(text, DirectionLTR)
	}
}

func BenchmarkReorderMedium(b *testing.B) {
	text := "The quick brown fox jumps over שלום the lazy dog"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Reorder(text, DirectionLTR)
	}
}

func BenchmarkReorderLong(b *testing.B) {
	text := `Lorem ipsum dolor sit amet, consectetur adipiscing elit. שלום עולם
Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. مرحبا بالعالم
Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Reorder(text, DirectionLTR)
	}
}

func BenchmarkComputeLevels(b *testing.B) {
	text := "Hello שלום مرحبا world"
	runes := []rune(text)
	classes := make([]BidiClass, len(runes))
	for i, r := range runes {
		classes[i] = GetBidiClass(r)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ComputeLevels(classes, 0)
	}
}
