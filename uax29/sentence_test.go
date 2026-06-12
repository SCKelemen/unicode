package uax29

import (
	"strings"
	"testing"
)

func TestFindSentenceBreaks_InvalidUTF8(t *testing.T) {
	for _, tc := range invalidUTF8Inputs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("FindSentenceBreaks(%q) panicked: %v", tc.in, r)
				}
			}()
			breaks := FindSentenceBreaks(tc.in)
			assertBreaksWellFormed(t, "FindSentenceBreaks", tc.in, breaks)
		})
	}
}

func TestSentences_InvalidUTF8(t *testing.T) {
	for _, tc := range invalidUTF8Inputs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Sentences(%q) panicked: %v", tc.in, r)
				}
			}()
			got := Sentences(tc.in)
			if joined := strings.Join(got, ""); joined != tc.in {
				t.Errorf("Sentences(%q): reassembled %q != original", tc.in, joined)
			}
		})
	}
}

func TestFindSentenceBreaks_EmptyAndASCII(t *testing.T) {
	if got := FindSentenceBreaks(""); len(got) != 0 {
		t.Errorf("FindSentenceBreaks(\"\") = %v, want []", got)
	}
	assertBreaksWellFormed(t, "FindSentenceBreaks", "a", FindSentenceBreaks("a"))
	assertBreaksWellFormed(t, "FindSentenceBreaks", "Hello. World!", FindSentenceBreaks("Hello. World!"))
}

// TestFindAllBreaks_InvalidUTF8 ensures the single-pass API is equally robust
// to ill-formed UTF-8, since it shares the same decode path.
func TestFindAllBreaks_InvalidUTF8(t *testing.T) {
	for _, tc := range invalidUTF8Inputs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("FindAllBreaks(%q) panicked: %v", tc.in, r)
				}
			}()
			all := FindAllBreaks(tc.in)
			assertBreaksWellFormed(t, "FindAllBreaks.Graphemes", tc.in, all.Graphemes)
			assertBreaksWellFormed(t, "FindAllBreaks.Words", tc.in, all.Words)
			assertBreaksWellFormed(t, "FindAllBreaks.Sentences", tc.in, all.Sentences)
		})
	}
}
