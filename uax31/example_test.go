package uax31_test

import (
	"fmt"

	"github.com/SCKelemen/unicode/uax31"
)

func ExampleIsXIDStart() {
	fmt.Println(uax31.IsXIDStart('A'))     // true
	fmt.Println(uax31.IsXIDStart('z'))     // true
	fmt.Println(uax31.IsXIDStart('5'))     // false (digit)
	fmt.Println(uax31.IsXIDStart('_'))     // false (not XID_Start)
	fmt.Println(uax31.IsXIDStart('中'))    // true (Han ideograph)
	// Output:
	// true
	// true
	// false
	// false
	// true
}

func ExampleIsXIDContinue() {
	fmt.Println(uax31.IsXIDContinue('A'))  // true
	fmt.Println(uax31.IsXIDContinue('5'))  // true (digit)
	fmt.Println(uax31.IsXIDContinue('_'))  // true (connector)
	fmt.Println(uax31.IsXIDContinue('+'))  // false (operator)
	fmt.Println(uax31.IsXIDContinue(' '))  // false (space)
	// Output:
	// true
	// true
	// true
	// false
	// false
}

func ExampleIsPatternSyntax() {
	fmt.Println(uax31.IsPatternSyntax('*'))  // true
	fmt.Println(uax31.IsPatternSyntax('+'))  // true
	fmt.Println(uax31.IsPatternSyntax('('))  // true
	fmt.Println(uax31.IsPatternSyntax('A'))  // false
	fmt.Println(uax31.IsPatternSyntax('5'))  // false
	// Output:
	// true
	// true
	// true
	// false
	// false
}

func ExampleIsPatternWhiteSpace() {
	fmt.Println(uax31.IsPatternWhiteSpace(' '))   // true
	fmt.Println(uax31.IsPatternWhiteSpace('\t'))  // true
	fmt.Println(uax31.IsPatternWhiteSpace('\n'))  // true
	fmt.Println(uax31.IsPatternWhiteSpace('A'))   // false
	fmt.Println(uax31.IsPatternWhiteSpace('_'))   // false
	// Output:
	// true
	// true
	// true
	// false
	// false
}

func ExampleIsValidIdentifier() {
	fmt.Println(uax31.IsValidIdentifier("myVar"))       // true
	fmt.Println(uax31.IsValidIdentifier("my_var"))      // true (underscore in middle)
	fmt.Println(uax31.IsValidIdentifier("myVar123"))    // true (digits after start)
	fmt.Println(uax31.IsValidIdentifier("_private"))    // false (underscore not XID_Start)
	fmt.Println(uax31.IsValidIdentifier("123var"))      // false (starts with digit)
	fmt.Println(uax31.IsValidIdentifier("my-var"))      // false (hyphen not in XID_Continue)
	fmt.Println(uax31.IsValidIdentifier(""))            // false (empty)
	// Output:
	// true
	// true
	// true
	// false
	// false
	// false
	// false
}

func ExampleIsValidIdentifier_unicode() {
	// Unicode identifiers
	fmt.Println(uax31.IsValidIdentifier("变量"))        // true (Chinese)
	fmt.Println(uax31.IsValidIdentifier("переменная")) // true (Russian)
	fmt.Println(uax31.IsValidIdentifier("μετβλητή"))   // true (Greek)
	fmt.Println(uax31.IsValidIdentifier("変数"))       // true (Japanese)
	// Output:
	// true
	// true
	// true
	// true
}
