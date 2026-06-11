package uax9

import (
	"reflect"
	"testing"
)

// These tests exercise BD16 (findBracketPairs) and N0 (resolveBracketPairs)
// in isolation, on hand-built isolatingRunSequence values. They are the
// regression suite for the paired-bracket implementation; the high-level
// conformance tests in bidi_character_test.go and official_tests_test.go
// give end-to-end coverage.

// makeSeq builds an isolatingRunSequence with the given runes and post-W
// bidi classes at the given embedding level. It is *not* a full setup —
// originalTypes is set to the same classes (no NSMs), and sos/eos are
// derived from the level so the BD16/N0 logic has sensible context.
func makeSeq(level int, runes []rune, types []BidiClass) *isolatingRunSequence {
	if len(runes) != len(types) {
		panic("makeSeq: runes/types length mismatch")
	}
	t := make([]BidiClass, len(types))
	copy(t, types)
	o := make([]BidiClass, len(types))
	copy(o, types)
	return &isolatingRunSequence{
		indexes:       nil, // unused by N0
		runes:         runes,
		types:         t,
		originalTypes: o,
		levels:        make([]int, len(types)),
		level:         level,
		sos:           typeForLevel(level),
		eos:           typeForLevel(level),
	}
}

// ----------------------------------------------------------------------
// Layer 1 — BD16 (paired-bracket identification)
// ----------------------------------------------------------------------

func TestBD16_FindBracketPairs(t *testing.T) {
	type tc struct {
		name string
		in   string
		want []bracketPair
	}
	// All inputs are pure ON characters so the BD16 ON-filter does not
	// affect anything. Position values are indexes into the rune slice
	// (which equals indexes into seq.types here).
	cases := []tc{
		{"outer paren contains inner square",
			"(a[b]c)",
			[]bracketPair{{0, 6}, {2, 4}},
		},
		{"outer square contains inner paren",
			"[a(b)c]",
			[]bracketPair{{0, 6}, {2, 4}},
		},
		{"nested same-shape parens", "((()))",
			[]bracketPair{{0, 5}, {1, 4}, {2, 3}},
		},
		{"unmatched open at position 0",
			"(()",
			[]bracketPair{{1, 2}},
		},
		{
			// '(' at 0 matches ')' at 1; then trailing ')' has
			// nothing to match; trailing '(' has no closer.
			name: "trailing junk drops",
			in:   "())(",
			want: []bracketPair{{0, 1}},
		},
		{
			// Per BD16: when a close matches an opener deeper in
			// the stack, the shallower (more recently pushed)
			// openers above it are popped and discarded. So when
			// ']' at position 2 matches '[' at position 0, '(' at
			// position 1 is dropped. Then ')' at position 3 has
			// nothing on the stack to match.
			name: "interleaved [(])",
			in:   "[(])",
			want: []bracketPair{{0, 2}},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			runes := []rune(c.in)
			types := make([]BidiClass, len(runes))
			for i := range types {
				types[i] = ClassON
			}
			seq := makeSeq(0, runes, types)
			got, ok := seq.findBracketPairs()
			if !ok {
				t.Fatalf("findBracketPairs returned !ok (stack overflow) for input %q", c.in)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("input %q\n  got  %v\n  want %v", c.in, got, c.want)
			}
		})
	}
}

// TestBD16_StackOverflowAbandons verifies that exceeding 63 openers
// causes BD16 to abandon (return ok=false).
func TestBD16_StackOverflowAbandons(t *testing.T) {
	runes := make([]rune, 64)
	types := make([]BidiClass, 64)
	for i := range runes {
		runes[i] = '('
		types[i] = ClassON
	}
	seq := makeSeq(0, runes, types)
	_, ok := seq.findBracketPairs()
	if ok {
		t.Fatalf("expected ok=false on >63 openers, got ok=true")
	}
}

// TestBD16_SkipsOverriddenBrackets verifies that brackets whose type has
// been changed by an X6 override (so they are no longer ON) are excluded
// from BD16 pair construction. This is the fix for BidiCharacterTest.txt
// vectors that mix LRE/LRO scopes with brackets.
func TestBD16_SkipsOverriddenBrackets(t *testing.T) {
	// `(` ON, `)` L (as if LRO had overridden it).
	runes := []rune{'(', 'X', ')'}
	types := []BidiClass{ClassON, ClassL, ClassL} // ')' overridden
	seq := makeSeq(0, runes, types)
	got, ok := seq.findBracketPairs()
	if !ok {
		t.Fatalf("unexpected !ok")
	}
	if len(got) != 0 {
		t.Fatalf("expected no pairs when close is overridden, got %v", got)
	}
}

// ----------------------------------------------------------------------
// Layer 2 — N0 (paired-bracket resolution)
// ----------------------------------------------------------------------

func TestN0_ResolveBracketPairs(t *testing.T) {
	// Each case sets up runes + post-W types + level and asserts the
	// types after applyN0. We always start from a single isolating
	// run sequence with sos/eos derived from level.
	type tc struct {
		name    string
		level   int
		runes   []rune
		typesIn []BidiClass
		// expected types after applyN0 (length must match runes)
		want []BidiClass
	}

	// LTR (level 0) cases.
	ltr := []tc{
		{
			// The BidiCharacterTest.txt line 44 vector that
			// flushed out the original ClassL==0 sentinel bug.
			// Embedding direction is L. Outer pair (2,11) has L
			// at position 7 inside → resolved L. Inner pair (5,9)
			// also has L inside → resolved L.
			name:  "line44 outer with inner L",
			level: 0,
			runes: []rune{
				0x05D0, 0x05D1, '(', 0x05D2, 0x05D3, '[', '&',
				'e', 'f', ']', '.', ')', 'g', 'h',
			},
			typesIn: []BidiClass{
				ClassR, ClassR, ClassON, ClassR, ClassR, ClassON,
				ClassON, ClassL, ClassL, ClassON, ClassON,
				ClassON, ClassL, ClassL,
			},
			want: []BidiClass{
				ClassR, ClassR, ClassL, ClassR, ClassR, ClassL,
				ClassON, ClassL, ClassL, ClassL, ClassON,
				ClassL, ClassL, ClassL,
			},
		},
		{
			// N0.a/b: any embed strong inside → embed direction
			name:    "embed strong inside → embed dir",
			level:   0,
			runes:   []rune{'(', 'a', ')'},
			typesIn: []BidiClass{ClassON, ClassL, ClassON},
			want:    []BidiClass{ClassL, ClassL, ClassL},
		},
		{
			// N0.c.1: only opposite inside, preceding strong is
			// opposite → opposite dir.
			name:    "only-opposite inside, opposite precedes → opposite",
			level:   0,
			runes:   []rune{0x05D0, '(', 0x05D1, ')'},
			typesIn: []BidiClass{ClassR, ClassON, ClassR, ClassON},
			want:    []BidiClass{ClassR, ClassR, ClassR, ClassR},
		},
		{
			// N0.c.2: only opposite inside, preceding strong is
			// embed (or none) → embed dir.
			name:    "only-opposite inside, embed precedes → embed",
			level:   0,
			runes:   []rune{'a', '(', 0x05D0, ')'},
			typesIn: []BidiClass{ClassL, ClassON, ClassR, ClassON},
			want:    []BidiClass{ClassL, ClassL, ClassR, ClassL},
		},
		{
			// N0.d: no strong inside → leave alone.
			name:    "no strong inside → unchanged",
			level:   0,
			runes:   []rune{'(', ' ', ')'},
			typesIn: []BidiClass{ClassON, ClassWS, ClassON},
			want:    []BidiClass{ClassON, ClassWS, ClassON},
		},
		{
			// EN/AN are treated as R for the strong-type check.
			// At LTR level: pair with only EN inside → c.2
			// (preceding is sos=L) → embed → L.
			name:    "EN treated as R for inside-check",
			level:   0,
			runes:   []rune{'(', '1', ')'},
			typesIn: []BidiClass{ClassON, ClassEN, ClassON},
			want:    []BidiClass{ClassL, ClassEN, ClassL},
		},
	}

	// RTL (level 1) cases.
	rtl := []tc{
		{
			// embedDir=R. EN inside is treated as R → embed →
			// brackets resolve to R.
			name:    "rtl: EN inside → R",
			level:   1,
			runes:   []rune{'(', '1', ')'},
			typesIn: []BidiClass{ClassON, ClassEN, ClassON},
			want:    []BidiClass{ClassR, ClassEN, ClassR},
		},
		{
			// embedDir=R. Only L inside; the preceding strong
			// type is sos = R (level 1). The "opposite to
			// embedding direction" for embedDir=R is L, so
			// preceding=R fails the c.1 check and we fall through
			// to c.2 → resolved=embedDir=R.
			name:    "rtl: only L inside, R precedes → R",
			level:   1,
			runes:   []rune{'(', 'a', ')'},
			typesIn: []BidiClass{ClassON, ClassL, ClassON},
			want:    []BidiClass{ClassR, ClassL, ClassR},
		},
		{
			// embedDir=R. Only L inside; the preceding strong is
			// itself L (matches the opposite of embedDir) → c.1
			// fires → resolved=L.
			name:    "rtl: only L inside, L precedes → L",
			level:   1,
			runes:   []rune{'a', '(', 'b', ')'},
			typesIn: []BidiClass{ClassL, ClassON, ClassL, ClassON},
			want:    []BidiClass{ClassL, ClassL, ClassL, ClassL},
		},
	}

	for _, group := range [][]tc{ltr, rtl} {
		for _, c := range group {
			t.Run(c.name, func(t *testing.T) {
				seq := makeSeq(c.level, c.runes, c.typesIn)
				pairs, ok := seq.findBracketPairs()
				if !ok {
					t.Fatalf("BD16 abandoned unexpectedly")
				}
				seq.resolveBracketPairs(pairs)
				if !reflect.DeepEqual(seq.types, c.want) {
					t.Fatalf("input runes=%v level=%d\n  got  %v\n  want %v",
						c.runes, c.level, seq.types, c.want)
				}
			})
		}
	}
}
