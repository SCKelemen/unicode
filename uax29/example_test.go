package uax29_test

import (
	"fmt"

	"github.com/SCKelemen/unicode/uax29"
)

func ExampleWords() {
	text := "Hello, world!"
	words := uax29.Words(text)
	for _, word := range words {
		fmt.Printf("[%s] ", word)
	}
	fmt.Println()

	// Output:
	// [Hello] [,] [ ] [world] [!]
}

func ExampleFindWordBreaks() {
	text := "Hello World"
	breaks := uax29.FindWordBreaks(text)
	fmt.Printf("Word break positions: %v\n", breaks)

	// Output:
	// Word break positions: [0 5 6 11]
}

func ExampleGraphemes_combining() {
	// Combining diacritical mark forms one grapheme cluster
	text := "e\u0301" // é as e + combining acute accent
	graphemes := uax29.Graphemes(text)
	fmt.Printf("Runes: %d, Graphemes: %d\n", len([]rune(text)), len(graphemes))

	// Output:
	// Runes: 2, Graphemes: 1
}

func ExampleFindGraphemeBreaks() {
	text := "abc"
	breaks := uax29.FindGraphemeBreaks(text)
	fmt.Printf("Breaks: %v\n", breaks)

	// Output:
	// Breaks: [0 1 2 3]
}
