package uax14_test

import (
	"fmt"
	"github.com/SCKelemen/unicode/uax14"
)

func ExampleFindLineBreakOpportunities() {
	text := "Hello world! How are you?"
	breaks := uax14.FindLineBreakOpportunities(text, uax14.HyphensManual)

	fmt.Println("Break positions:", breaks)
	// Output: Break positions: [0 6 13 17 21 25]
}

func ExampleFindLineBreakOpportunities_withSoftHyphens() {
	// Soft hyphen (U+00AD) between "super" and "cali"
	text := "super\u00ADcalifragilistic"

	// With HyphensManual, breaks are allowed at soft hyphens
	breaks := uax14.FindLineBreakOpportunities(text, uax14.HyphensManual)
	fmt.Println("Manual hyphens:", breaks)

	// With HyphensNone, no hyphenation breaks
	breaksNone := uax14.FindLineBreakOpportunities(text, uax14.HyphensNone)
	fmt.Println("No hyphens:", breaksNone)

	// Output:
	// Manual hyphens: [0 7 22]
	// No hyphens: [0 22]
}

func ExampleFindLineBreakOpportunities_wrapping() {
	text := "The quick brown fox jumps"
	breaks := uax14.FindLineBreakOpportunities(text, uax14.HyphensManual)

	// Simple demonstration of using break positions
	fmt.Println("Text:", text)
	fmt.Println("Break opportunities:", breaks)
	fmt.Println("Can break after: 'The ', 'quick ', 'brown ', 'fox '")

	// Output:
	// Text: The quick brown fox jumps
	// Break opportunities: [0 4 10 16 20 25]
	// Can break after: 'The ', 'quick ', 'brown ', 'fox '
}

func ExampleFindLineBreakOpportunities_cjk() {
	// Chinese/Japanese text can break between characters
	// Hiragana (こんにちは) and Kanji (世界) are both classified as ID (Ideographic)
	// allowing breaks between each character per UAX #14
	text := "こんにちは世界"
	breaks := uax14.FindLineBreakOpportunities(text, uax14.HyphensManual)

	fmt.Printf("Text has %d break opportunities\n", len(breaks))
	// Output: Text has 8 break opportunities
}
