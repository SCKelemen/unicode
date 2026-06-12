package uax29

import (
	"strings"
	"testing"
)

// invalidUTF8Inputs is the shared corpus of ill-formed UTF-8 inputs used to
// exercise the Find*Breaks regression tests. None of these may cause a panic,
// and every produced break must be a valid byte offset into the input.
var invalidUTF8Inputs = []struct {
	name string
	in   string
}{
	{"two_lone_bytes_between_ascii", "a\xff\xffb"},
	{"single_lone_byte", "\xff"},
	{"trailing_lone_byte", "abc\xff"},
	{"leading_lone_byte", "\xffabc"},
	{"truncated_two_byte_sequence", "a\xc2b"},
	{"truncated_three_byte_sequence", "a\xe0\xa0b"},
	{"only_continuation_bytes", "\x80\x80\x80"},
	{"overlong_then_ascii", "\xc0\xafabc"},
}

// assertBreaksWellFormed checks the universal invariants required of every
// Find*Breaks result for a non-empty input:
//   - the slice starts with 0 and ends with len(text);
//   - every offset is within [0, len(text)];
//   - offsets are strictly monotonically increasing;
//   - concatenating text[breaks[i]:breaks[i+1]] reproduces text exactly.
func assertBreaksWellFormed(t *testing.T, label, text string, breaks []int) {
	t.Helper()
	if len(text) == 0 {
		// Documented contract: empty input -> empty slice.
		if len(breaks) != 0 {
			t.Errorf("%s(%q): empty input must return empty slice, got %v", label, text, breaks)
		}
		return
	}
	if len(breaks) < 2 {
		t.Fatalf("%s(%q): expected at least [0, len(text)], got %v", label, text, breaks)
	}
	if breaks[0] != 0 {
		t.Errorf("%s(%q): first break = %d, want 0", label, text, breaks[0])
	}
	if breaks[len(breaks)-1] != len(text) {
		t.Errorf("%s(%q): last break = %d, want %d", label, text, breaks[len(breaks)-1], len(text))
	}
	var b strings.Builder
	for i, off := range breaks {
		if off < 0 || off > len(text) {
			t.Errorf("%s(%q): breaks[%d]=%d is out of range [0,%d]", label, text, i, off, len(text))
		}
		if i > 0 && off <= breaks[i-1] {
			t.Errorf("%s(%q): non-monotonic at index %d: %d <= %d", label, text, i, off, breaks[i-1])
		}
		if i > 0 {
			b.WriteString(text[breaks[i-1]:off])
		}
	}
	if b.String() != text {
		t.Errorf("%s(%q): reassembled %q != original", label, text, b.String())
	}
}

func TestFindGraphemeBreaks_InvalidUTF8(t *testing.T) {
	for _, tc := range invalidUTF8Inputs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("FindGraphemeBreaks(%q) panicked: %v", tc.in, r)
				}
			}()
			breaks := FindGraphemeBreaks(tc.in)
			assertBreaksWellFormed(t, "FindGraphemeBreaks", tc.in, breaks)
		})
	}
}

func TestGraphemes_InvalidUTF8(t *testing.T) {
	for _, tc := range invalidUTF8Inputs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Graphemes(%q) panicked: %v", tc.in, r)
				}
			}()
			got := Graphemes(tc.in)
			// Reassembling the clusters must reproduce the input exactly,
			// byte-for-byte, even for invalid UTF-8.
			if joined := strings.Join(got, ""); joined != tc.in {
				t.Errorf("Graphemes(%q): reassembled %q != original", tc.in, joined)
			}
		})
	}
}

func TestFindGraphemeBreaks_EmptyAndASCII(t *testing.T) {
	// Documented contract for empty input.
	if got := FindGraphemeBreaks(""); len(got) != 0 {
		t.Errorf("FindGraphemeBreaks(\"\") = %v, want []", got)
	}
	// Non-empty input must satisfy the universal invariants.
	assertBreaksWellFormed(t, "FindGraphemeBreaks", "a", FindGraphemeBreaks("a"))
	assertBreaksWellFormed(t, "FindGraphemeBreaks", "hello", FindGraphemeBreaks("hello"))
	// And the byte positions must match the documented examples.
	if got := FindGraphemeBreaks("a"); len(got) != 2 || got[0] != 0 || got[1] != 1 {
		t.Errorf("FindGraphemeBreaks(\"a\") = %v, want [0 1]", got)
	}
}
