package uax9_test

import (
	"fmt"

	"github.com/SCKelemen/unicode/uax9"
)

// Example demonstrates basic bidirectional text reordering.
func Example() {
	// Mix English (LTR) with Hebrew (RTL)
	text := "Hello שלום world"
	result := uax9.Reorder(text, uax9.DirectionLTR)
	fmt.Println(result)
	// Output: Hello םולש world
}

// ExampleReorder demonstrates reordering text with explicit direction.
func ExampleReorder() {
	// English text with embedded Hebrew
	text := "Hello שלום"
	result := uax9.Reorder(text, uax9.DirectionLTR)
	fmt.Println(result)
	// Output: Hello םולש
}

// ExampleReorder_rtl demonstrates RTL paragraph with embedded LTR text.
func ExampleReorder_rtl() {
	// Hebrew text in RTL paragraph
	text := "שלום Hello עולם"
	result := uax9.Reorder(text, uax9.DirectionRTL)
	fmt.Println(result)
	// Output: םלוע Hello םולש
}

// ExampleReorder_auto demonstrates automatic direction detection.
func ExampleReorder_auto() {
	// Let the algorithm detect direction from first strong character
	textLTR := "Hello שלום"
	resultLTR := uax9.Reorder(textLTR, uax9.DirectionAuto)
	fmt.Println(resultLTR)

	textRTL := "שלום Hello"
	resultRTL := uax9.Reorder(textRTL, uax9.DirectionAuto)
	fmt.Println(resultRTL)
	// Output:
	// Hello םולש
	// Hello םולש
}

// ExampleReorder_numbers demonstrates proper handling of numbers in bidirectional text.
func ExampleReorder_numbers() {
	// Numbers maintain their order within RTL text
	text := "Price: 123 שקלים"
	result := uax9.Reorder(text, uax9.DirectionLTR)
	fmt.Println(result)
	// Output: Price: 123 םילקש
}

// ExampleReorder_arabic demonstrates Arabic text with English embedded.
func ExampleReorder_arabic() {
	// Arabic (RTL) with embedded English (LTR)
	text := "مرحبا Hello العالم"
	result := uax9.Reorder(text, uax9.DirectionRTL)
	fmt.Println(result)
	// Output: ملاعلا Hello ابحرم
}

// ExampleReorder_punctuation demonstrates punctuation handling in mixed text.
func ExampleReorder_punctuation() {
	// Punctuation takes directionality from context
	text := "Hello, שלום!"
	result := uax9.Reorder(text, uax9.DirectionLTR)
	fmt.Println(result)
	// Output: Hello, םולש!
}

// ExampleGetBidiClass demonstrates checking the bidirectional class of characters.
func ExampleGetBidiClass() {
	// Check various character types
	fmt.Println(uax9.GetBidiClass('a'))  // Latin letter
	fmt.Println(uax9.GetBidiClass('א'))  // Hebrew letter
	fmt.Println(uax9.GetBidiClass('ا'))  // Arabic letter
	fmt.Println(uax9.GetBidiClass('5'))  // European digit
	fmt.Println(uax9.GetBidiClass('+'))  // Number separator
	fmt.Println(uax9.GetBidiClass(' '))  // Space
	fmt.Println(uax9.GetBidiClass('!'))  // Punctuation
	// Output:
	// L
	// R
	// AL
	// EN
	// ES
	// WS
	// ON
}

// ExampleGetBidiClass_explicitFormatting demonstrates explicit formatting characters.
func ExampleGetBidiClass_explicitFormatting() {
	// Explicit formatting characters (normally invisible)
	fmt.Println(uax9.GetBidiClass('\u202A')) // LEFT-TO-RIGHT EMBEDDING
	fmt.Println(uax9.GetBidiClass('\u202B')) // RIGHT-TO-LEFT EMBEDDING
	fmt.Println(uax9.GetBidiClass('\u202C')) // POP DIRECTIONAL FORMATTING
	fmt.Println(uax9.GetBidiClass('\u2066')) // LEFT-TO-RIGHT ISOLATE
	fmt.Println(uax9.GetBidiClass('\u2067')) // RIGHT-TO-LEFT ISOLATE
	fmt.Println(uax9.GetBidiClass('\u2069')) // POP DIRECTIONAL ISOLATE
	// Output:
	// LRE
	// RLE
	// PDF
	// LRI
	// RLI
	// PDI
}

// ExampleGetParagraphDirection demonstrates automatic paragraph direction detection.
func ExampleGetParagraphDirection() {
	// Detect direction from first strong character
	textLTR := "Hello world"
	dirLTR := uax9.GetParagraphDirection(textLTR)
	fmt.Printf("LTR text direction: %v\n", dirLTR == uax9.DirectionLTR)

	textRTL := "שלום עולם"
	dirRTL := uax9.GetParagraphDirection(textRTL)
	fmt.Printf("RTL text direction: %v\n", dirRTL == uax9.DirectionRTL)

	textMixed := "123 Hello שלום"
	dirMixed := uax9.GetParagraphDirection(textMixed)
	fmt.Printf("Mixed text (starts with number): %v\n", dirMixed == uax9.DirectionLTR)
	// Output:
	// LTR text direction: true
	// RTL text direction: true
	// Mixed text (starts with number): true
}

// ExampleComputeLevels demonstrates computing embedding levels for advanced use cases.
func ExampleComputeLevels() {
	// Text: "Hello שלום"
	text := "Hello שלום"
	runes := []rune(text)

	// Get bidi classes for each character
	classes := make([]uax9.BidiClass, len(runes))
	for i, r := range runes {
		classes[i] = uax9.GetBidiClass(r)
	}

	// Compute embedding levels (0 = LTR paragraph)
	levels := uax9.ComputeLevels(classes, 0)

	// Display results
	for i, r := range runes {
		if levels[i] >= 0 { // Skip removed characters (level -1)
			fmt.Printf("%c: level %d\n", r, levels[i])
		}
	}
	// Output:
	// H: level 0
	// e: level 0
	// l: level 0
	// l: level 0
	// o: level 0
	//  : level 0
	// ש: level 1
	// ל: level 1
	// ו: level 1
	// ם: level 1
}

// ExampleComputeLevels_isolates demonstrates isolating run sequences.
func ExampleComputeLevels_isolates() {
	// Using isolate formatting characters for nested contexts
	// Text with RLI...PDI isolate
	text := "Hello \u2067שלום\u2069 world"
	runes := []rune(text)

	classes := make([]uax9.BidiClass, len(runes))
	for i, r := range runes {
		classes[i] = uax9.GetBidiClass(r)
	}

	// Make a copy since ComputeLevels modifies the classes array
	originalClasses := make([]uax9.BidiClass, len(classes))
	copy(originalClasses, classes)

	levels := uax9.ComputeLevels(classes, 0)

	// The Hebrew text inside RLI...PDI is isolated
	fmt.Println("Text with isolate:")
	for i, r := range runes {
		if levels[i] >= 0 {
			fmt.Printf("%c: level %d, class %v\n", r, levels[i], originalClasses[i])
		}
	}
	// Output:
	// Text with isolate:
	// H: level 0, class L
	// e: level 0, class L
	// l: level 0, class L
	// l: level 0, class L
	// o: level 0, class L
	//  : level 0, class WS
	// ⁧: level 0, class RLI
	// ש: level 1, class R
	// ל: level 1, class R
	// ו: level 1, class R
	// ם: level 1, class R
	// ⁩: level 0, class PDI
	//  : level 0, class WS
	// w: level 0, class L
	// o: level 0, class L
	// r: level 0, class L
	// l: level 0, class L
	// d: level 0, class L
}

// Example_complexMixedText demonstrates handling of complex mixed-direction text.
func Example_complexMixedText() {
	// Complex real-world scenario: English sentence with Hebrew and Arabic
	text := "The Hebrew word שלום means peace, as does the Arabic word سلام."
	result := uax9.Reorder(text, uax9.DirectionLTR)
	fmt.Println(result)
	// Output: The Hebrew word םולש means peace, as does the Arabic word مالس.
}

// Example_nestedEmbeddings demonstrates handling of nested directional embeddings.
func Example_nestedEmbeddings() {
	// Simple case without explicit embeddings
	text := "English עברית English"
	result := uax9.Reorder(text, uax9.DirectionLTR)
	fmt.Println(result)
	// Output: English תירבע English
}

// Example_numbersAndPunctuation demonstrates proper handling of numbers and punctuation.
func Example_numbersAndPunctuation() {
	// Numbers and punctuation in mixed text
	text := "Price: $123.45 - מחיר"
	result := uax9.Reorder(text, uax9.DirectionLTR)
	fmt.Println(result)
	// Output: Price: $123.45 - ריחמ
}
