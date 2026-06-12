package uax34_test

import (
	"fmt"

	"github.com/SCKelemen/unicode/v6/uax34"
)

func ExampleLookup() {
	seq, ok := uax34.Lookup("KEYCAP NUMBER SIGN")
	fmt.Printf("%v %t\n", seq, ok)
	// Output: [35 65039 8419] true
}

func ExampleName() {
	name, ok := uax34.Name([]rune{0x0023, 0xFE0F, 0x20E3})
	fmt.Printf("%q %t\n", name, ok)
	// Output: "KEYCAP NUMBER SIGN" true
}

func ExampleCount() {
	fmt.Println(uax34.Count())
	// Output: 461
}
