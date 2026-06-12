package uax29

import (
	"strings"
	"testing"
)

func TestFindWordBreaks_InvalidUTF8(t *testing.T) {
	for _, tc := range invalidUTF8Inputs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("FindWordBreaks(%q) panicked: %v", tc.in, r)
				}
			}()
			breaks := FindWordBreaks(tc.in)
			assertBreaksWellFormed(t, "FindWordBreaks", tc.in, breaks)
		})
	}
}

func TestWords_InvalidUTF8(t *testing.T) {
	for _, tc := range invalidUTF8Inputs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Words(%q) panicked: %v", tc.in, r)
				}
			}()
			got := Words(tc.in)
			if joined := strings.Join(got, ""); joined != tc.in {
				t.Errorf("Words(%q): reassembled %q != original", tc.in, joined)
			}
		})
	}
}

func TestFindWordBreaks_EmptyAndASCII(t *testing.T) {
	if got := FindWordBreaks(""); len(got) != 0 {
		t.Errorf("FindWordBreaks(\"\") = %v, want []", got)
	}
	assertBreaksWellFormed(t, "FindWordBreaks", "a", FindWordBreaks("a"))
	assertBreaksWellFormed(t, "FindWordBreaks", "Hello, world!", FindWordBreaks("Hello, world!"))
}
