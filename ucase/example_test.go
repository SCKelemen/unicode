package ucase_test

import (
	"fmt"

	"github.com/SCKelemen/unicode/v6/ucase"
)

func ExampleToLower() {
	// The Greek capital sigma lowercases to the final form at the end of a
	// word (Final_Sigma context) and the medial form elsewhere.
	fmt.Println(ucase.ToLower("ΟΔΟΣ"))
	// Output: οδος
}

func ExampleToUpper() {
	// Full (1:many) uppercasing: the German sharp s maps to "SS".
	fmt.Println(ucase.ToUpper("straße"))
	// Output: STRASSE
}

func ExampleToTitle() {
	fmt.Println(ucase.ToTitle("hello, world"))
	// Output: Hello, World
}

func ExampleToFold() {
	// Case folding normalizes for caseless comparison; the ligature ﬀ folds
	// to "ff".
	fmt.Println(ucase.ToFold("eﬀort"))
	// Output: effort
}

func ExampleCaselessMatch() {
	fmt.Println(ucase.CaselessMatch("Straße", "STRASSE"))
	fmt.Println(ucase.CaselessMatch("café", "cafe"))
	// Output:
	// true
	// false
}
