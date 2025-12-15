package uax50_test

import (
	"fmt"

	"github.com/SCKelemen/unicode/uax50"
)

// ExampleLookupOrientation demonstrates basic orientation lookup for various characters.
func ExampleLookupOrientation() {
	chars := []rune{'A', '中', '。', '〜'}

	for _, r := range chars {
		orientation := uax50.LookupOrientation(r)
		fmt.Printf("%c (%U): %v\n", r, r, orientation)
	}

	// Output:
	// A (U+0041): R
	// 中 (U+4E2D): U
	// 。 (U+3002): Tu
	// 〜 (U+301C): Tr
}

// ExampleIsUpright demonstrates checking if characters should be displayed upright.
func ExampleIsUpright() {
	// CJK characters are typically upright
	fmt.Printf("中 is upright: %v\n", uax50.IsUpright('中'))

	// Latin characters are typically rotated
	fmt.Printf("A is upright: %v\n", uax50.IsUpright('A'))

	// Ideographic punctuation is upright (with transformation)
	fmt.Printf("。 is upright: %v\n", uax50.IsUpright('。'))

	// Output:
	// 中 is upright: true
	// A is upright: false
	// 。 is upright: true
}

// ExampleIsRotated demonstrates checking if characters should be rotated.
func ExampleIsRotated() {
	// Latin characters should be rotated 90 degrees clockwise
	fmt.Printf("A is rotated: %v\n", uax50.IsRotated('A'))

	// Digits should be rotated
	fmt.Printf("5 is rotated: %v\n", uax50.IsRotated('5'))

	// CJK ideographs should not be rotated
	fmt.Printf("日 is rotated: %v\n", uax50.IsRotated('日'))

	// Output:
	// A is rotated: true
	// 5 is rotated: true
	// 日 is rotated: false
}

// ExampleRequiresTransformation demonstrates checking for glyph transformation needs.
func ExampleRequiresTransformation() {
	// Ideographic comma requires transformation for vertical text
	fmt.Printf("、 requires transformation: %v\n", uax50.RequiresTransformation('、'))

	// Regular characters don't require transformation
	fmt.Printf("中 requires transformation: %v\n", uax50.RequiresTransformation('中'))
	fmt.Printf("A requires transformation: %v\n", uax50.RequiresTransformation('A'))

	// Output:
	// 、 requires transformation: true
	// 中 requires transformation: false
	// A requires transformation: false
}

// ExampleGetBaseOrientation demonstrates getting the fallback orientation.
func ExampleGetBaseOrientation() {
	// TransformedUpright falls back to Upright
	ideographicComma := '、'
	fmt.Printf("、 base orientation: %v\n", uax50.GetBaseOrientation(ideographicComma))

	// TransformedRotated falls back to Rotated
	waveDash := '〜'
	fmt.Printf("〜 base orientation: %v\n", uax50.GetBaseOrientation(waveDash))

	// Regular orientations stay the same
	fmt.Printf("中 base orientation: %v\n", uax50.GetBaseOrientation('中'))

	// Output:
	// 、 base orientation: U
	// 〜 base orientation: R
	// 中 base orientation: U
}

// Example_verticalTextLayout demonstrates a simple vertical text layout algorithm
// using the vertical orientation property.
func Example_verticalTextLayout() {
	text := "Hello世界"

	fmt.Println("Vertical text layout:")
	for _, r := range text {
		orientation := uax50.LookupOrientation(r)

		switch orientation {
		case uax50.Upright, uax50.TransformedUpright:
			fmt.Printf("%c - Display upright\n", r)
		case uax50.Rotated, uax50.TransformedRotated:
			fmt.Printf("%c - Rotate 90° clockwise\n", r)
		}
	}

	// Output:
	// Vertical text layout:
	// H - Rotate 90° clockwise
	// e - Rotate 90° clockwise
	// l - Rotate 90° clockwise
	// l - Rotate 90° clockwise
	// o - Rotate 90° clockwise
	// 世 - Display upright
	// 界 - Display upright
}

// Example_glyphTransformation demonstrates handling transformed glyphs for vertical text.
func Example_glyphTransformation() {
	// Ideographic punctuation that needs transformation
	punctuation := []rune{'、', '。'}

	for _, r := range punctuation {
		if uax50.RequiresTransformation(r) {
			baseOrientation := uax50.GetBaseOrientation(r)
			fmt.Printf("%c: Use vertical glyph variant, fallback to %v\n", r, baseOrientation)
		}
	}

	// Output:
	// 、: Use vertical glyph variant, fallback to U
	// 。: Use vertical glyph variant, fallback to U
}
