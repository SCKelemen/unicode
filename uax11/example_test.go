package uax11_test

import (
	"fmt"

	"github.com/SCKelemen/unicode/uax11"
)

// ExampleLookupWidth demonstrates basic width lookup for various characters.
func ExampleLookupWidth() {
	chars := []rune{'A', '中', 'Ω', '\uFF21'}

	for _, r := range chars {
		width := uax11.LookupWidth(r)
		fmt.Printf("%c (%U): %v\n", r, r, width)
	}

	// Output:
	// A (U+0041): Na
	// 中 (U+4E2D): W
	// Ω (U+03A9): A
	// Ａ (U+FF21): F
}

// ExampleIsWide demonstrates checking if characters are wide.
func ExampleIsWide() {
	// CJK characters are wide
	fmt.Printf("中 is wide: %v\n", uax11.IsWide('中'))

	// ASCII characters are not wide
	fmt.Printf("A is wide: %v\n", uax11.IsWide('A'))

	// Fullwidth forms are wide
	fmt.Printf("Ａ is wide: %v\n", uax11.IsWide('\uFF21'))

	// Output:
	// 中 is wide: true
	// A is wide: false
	// Ａ is wide: true
}

// ExampleIsNarrow demonstrates checking if characters are narrow.
func ExampleIsNarrow() {
	// ASCII characters are narrow
	fmt.Printf("A is narrow: %v\n", uax11.IsNarrow('A'))

	// CJK ideographs are not narrow
	fmt.Printf("中 is narrow: %v\n", uax11.IsNarrow('中'))

	// Halfwidth Katakana is narrow
	fmt.Printf("ｦ is narrow: %v\n", uax11.IsNarrow('\uFF66'))

	// Output:
	// A is narrow: true
	// 中 is narrow: false
	// ｦ is narrow: true
}

// ExampleIsAmbiguous demonstrates checking for ambiguous width characters.
func ExampleIsAmbiguous() {
	// Greek letters are ambiguous
	fmt.Printf("Ω is ambiguous: %v\n", uax11.IsAmbiguous('Ω'))
	fmt.Printf("α is ambiguous: %v\n", uax11.IsAmbiguous('α'))

	// ASCII is not ambiguous
	fmt.Printf("A is ambiguous: %v\n", uax11.IsAmbiguous('A'))

	// Output:
	// Ω is ambiguous: true
	// α is ambiguous: true
	// A is ambiguous: false
}

// ExampleResolveWidth demonstrates resolving ambiguous widths with context.
func ExampleResolveWidth() {
	omega := 'Ω'

	// In East Asian context, ambiguous becomes wide
	eaWidth := uax11.ResolveWidth(omega, uax11.ContextEastAsian)
	fmt.Printf("Ω in East Asian context: %v\n", eaWidth)

	// In narrow context, ambiguous becomes narrow
	narrowWidth := uax11.ResolveWidth(omega, uax11.ContextNarrow)
	fmt.Printf("Ω in narrow context: %v\n", narrowWidth)

	// Output:
	// Ω in East Asian context: W
	// Ω in narrow context: Na
}

// ExampleCharWidth demonstrates calculating character display widths.
func ExampleCharWidth() {
	// ASCII is 1 unit wide
	fmt.Printf("Width of 'A': %d\n", uax11.CharWidth('A', uax11.ContextNarrow))

	// CJK is 2 units wide
	fmt.Printf("Width of '中': %d\n", uax11.CharWidth('中', uax11.ContextNarrow))

	// Ambiguous depends on context
	fmt.Printf("Width of 'Ω' (narrow): %d\n", uax11.CharWidth('Ω', uax11.ContextNarrow))
	fmt.Printf("Width of 'Ω' (East Asian): %d\n", uax11.CharWidth('Ω', uax11.ContextEastAsian))

	// Output:
	// Width of 'A': 1
	// Width of '中': 2
	// Width of 'Ω' (narrow): 1
	// Width of 'Ω' (East Asian): 2
}

// ExampleStringWidth demonstrates calculating string display widths.
func ExampleStringWidth() {
	// Pure ASCII
	fmt.Printf("Width of 'Hello': %d\n", uax11.StringWidth("Hello", uax11.ContextNarrow))

	// Pure CJK
	fmt.Printf("Width of '中国': %d\n", uax11.StringWidth("中国", uax11.ContextNarrow))

	// Mixed content
	fmt.Printf("Width of 'Hello世界': %d\n", uax11.StringWidth("Hello世界", uax11.ContextNarrow))

	// Context matters for ambiguous characters
	greek := "ΩΩΩ"
	fmt.Printf("Width of '%s' (narrow): %d\n", greek, uax11.StringWidth(greek, uax11.ContextNarrow))
	fmt.Printf("Width of '%s' (East Asian): %d\n", greek, uax11.StringWidth(greek, uax11.ContextEastAsian))

	// Output:
	// Width of 'Hello': 5
	// Width of '中国': 4
	// Width of 'Hello世界': 9
	// Width of 'ΩΩΩ' (narrow): 3
	// Width of 'ΩΩΩ' (East Asian): 6
}

// Example_terminalWidth demonstrates calculating terminal display width for text.
func Example_terminalWidth() {
	// Terminal emulators typically need to know display widths
	lines := []string{
		"Hello",
		"Hello世界",
		"中国日本",
		"αβγδε",
	}

	fmt.Println("Terminal widths:")
	for _, line := range lines {
		width := uax11.StringWidth(line, uax11.ContextNarrow)
		fmt.Printf("%s width: %d\n", line, width)
	}

	// Output:
	// Terminal widths:
	// Hello width: 5
	// Hello世界 width: 9
	// 中国日本 width: 8
	// αβγδε width: 5
}

// Example_padding demonstrates using width information for text padding.
func Example_padding() {
	// Simple width calculation example
	texts := []string{"Alice", "田中", "Bob"}

	for _, text := range texts {
		width := uax11.StringWidth(text, uax11.ContextNarrow)
		fmt.Printf("'%s' width: %d\n", text, width)
	}

	// Output:
	// 'Alice' width: 5
	// '田中' width: 4
	// 'Bob' width: 3
}
