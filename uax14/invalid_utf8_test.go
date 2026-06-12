package uax14

import (
	"testing"
)

// Regression tests for the invalid-UTF-8 byte-offset drift bug.
//
// Background: prior versions of FindLineBreakOpportunities and
// FindLineBreakOpportunitiesWithRules converted the input via []rune(text) and
// then derived byte offsets from the resulting runes (via utf8.RuneLen or
// len(string(runes[:i]))). For ill-formed UTF-8, []rune replaces each bad byte
// with U+FFFD, which re-encodes to 3 bytes, so byte offsets drifted away from
// the real input. The fix is to walk the input with utf8.DecodeRuneInString
// and build a parallel bytePositions slice that tracks the ORIGINAL offsets,
// matching the stdlib behavior where a lone invalid byte advances by 1.
//
// These tests assert, for each of the two public entry points and for several
// inputs containing invalid UTF-8 sequences, that:
//   - the call does not panic
//   - returned break offsets are monotonically non-decreasing
//   - every offset lies in [0, len(text)]
//   - the first offset is 0 and the last is len(text)
//   - reassembling chunks via the break offsets reproduces text byte-for-byte
//
// Inputs cover three flavors of malformation:
//   - "hello \xff world": invalid byte embedded in valid surrounding text
//   - "\xff":              a single invalid byte
//   - "a\xc2b":            truncated 2-byte sequence wedged between ASCII

var invalidUTF8Inputs = []struct {
	name string
	text string
}{
	{"invalid_byte_in_middle", "hello \xff world"},
	{"lone_invalid_byte", "\xff"},
	{"truncated_two_byte_sequence", "a\xc2b"},
	// A few extra shapes for breadth.
	{"two_invalid_bytes", "\xff\xfe"},
	{"invalid_then_multibyte", "\xffé"},
	{"multibyte_then_invalid", "é\xff"},
	{"invalid_at_end", "abc\xff"},
	{"invalid_at_start", "\xffabc"},
}

func assertValidBreakOffsets(t *testing.T, label, text string, breaks []int) {
	t.Helper()

	if len(breaks) == 0 {
		t.Fatalf("%s: %q -> empty break slice", label, text)
	}
	if breaks[0] != 0 {
		t.Errorf("%s: %q -> first break = %d, want 0; got %v", label, text, breaks[0], breaks)
	}
	if breaks[len(breaks)-1] != len(text) {
		t.Errorf("%s: %q -> last break = %d, want len(text)=%d; got %v",
			label, text, breaks[len(breaks)-1], len(text), breaks)
	}

	// Monotonic non-decreasing and within range.
	for i, off := range breaks {
		if off < 0 || off > len(text) {
			t.Errorf("%s: %q -> break[%d]=%d out of range [0, %d]; got %v",
				label, text, i, off, len(text), breaks)
		}
		if i > 0 && off < breaks[i-1] {
			t.Errorf("%s: %q -> non-monotonic at index %d (%d < %d); got %v",
				label, text, i, off, breaks[i-1], breaks)
		}
	}

	// Reassembly preserves bytes exactly.
	var reassembled []byte
	prev := 0
	for _, off := range breaks {
		reassembled = append(reassembled, text[prev:off]...)
		prev = off
	}
	if string(reassembled) != text {
		t.Errorf("%s: %q -> reassembled %q != original; breaks=%v",
			label, text, string(reassembled), breaks)
	}
}

func TestFindLineBreakOpportunities_InvalidUTF8(t *testing.T) {
	for _, tc := range invalidUTF8Inputs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic for %q: %v", tc.text, r)
				}
			}()
			breaks := FindLineBreakOpportunities(tc.text, HyphensManual)
			assertValidBreakOffsets(t, "FindLineBreakOpportunities", tc.text, breaks)
		})
	}
}

func TestFindLineBreakOpportunitiesWithRules_InvalidUTF8(t *testing.T) {
	for _, tc := range invalidUTF8Inputs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic for %q: %v", tc.text, r)
				}
			}()
			breaks := FindLineBreakOpportunitiesWithRules(tc.text, HyphensManual)
			assertValidBreakOffsets(t, "FindLineBreakOpportunitiesWithRules", tc.text, breaks)
		})
	}
}

// TestInvalidUTF8_OffsetsMatchRawBytes is the focused regression assertion:
// for the input "hello \xff world" the offsets returned MUST be valid indices
// into the raw byte string. Specifically, slicing text[breaks[i]:breaks[i+1]]
// for every i must succeed and the concatenation must equal text. Prior to the
// fix, the trailing "world" chunk could slip past len(text) because U+FFFD
// re-encoded as 3 bytes was being added to the running byte position.
func TestInvalidUTF8_OffsetsMatchRawBytes(t *testing.T) {
	const text = "hello \xff world"

	for _, fn := range []struct {
		name string
		call func(string, Hyphens) []int
	}{
		{"FindLineBreakOpportunities", FindLineBreakOpportunities},
		{"FindLineBreakOpportunitiesWithRules", FindLineBreakOpportunitiesWithRules},
	} {
		fn := fn
		t.Run(fn.name, func(t *testing.T) {
			breaks := fn.call(text, HyphensManual)

			// Every offset must be a valid index into text.
			for _, off := range breaks {
				if off < 0 || off > len(text) {
					t.Fatalf("%s: offset %d out of range [0, %d]; breaks=%v",
						fn.name, off, len(text), breaks)
				}
			}

			// Pairwise slicing reproduces the input verbatim.
			var got []byte
			for i := 0; i < len(breaks)-1; i++ {
				got = append(got, text[breaks[i]:breaks[i+1]]...)
			}
			if string(got) != text {
				t.Fatalf("%s: reassembled %q != original %q; breaks=%v",
					fn.name, string(got), text, breaks)
			}
		})
	}
}
