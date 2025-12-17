package uts15_test

import (
	"fmt"

	"github.com/SCKelemen/unicode/uts15"
)

func ExampleNFC() {
	// Normalize to NFC (Canonical Composition)
	// This is the recommended form for most uses
	decomposed := "e\u0301" // e + combining acute accent
	normalized := uts15.NFC(decomposed)
	fmt.Printf("%s\n", normalized) // é (composed)

	// Already composed text stays the same
	composed := "\u00E9"
	fmt.Printf("%s\n", uts15.NFC(composed))

	// Output:
	// é
	// é
}

func ExampleNFD() {
	// Normalize to NFD (Canonical Decomposition)
	composed := "\u00E9" // é (composed)
	normalized := uts15.NFD(composed)
	runes := []rune(normalized)
	for i, r := range runes {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Printf("%X", r)
	}
	fmt.Println()

	// Output:
	// 65 301
}

func ExampleNFKC() {
	// Normalize to NFKC (Compatibility Composition)
	// Decomposes ligatures and other compatibility characters
	ligature := "\uFB01" // ﬁ ligature
	normalized := uts15.NFKC(ligature)
	fmt.Printf("%s\n", normalized)

	// Full-width characters
	fullwidth := "\uFF21\uFF10" // Full-width "A0"
	fmt.Printf("%s\n", uts15.NFKC(fullwidth))

	// Output:
	// fi
	// A0
}

func ExampleNFKD() {
	// Normalize to NFKD (Compatibility Decomposition)
	ligature := "\uFB01" // ﬁ ligature
	normalized := uts15.NFKD(ligature)
	fmt.Printf("%s\n", normalized)

	// Circled numbers
	circled := "\u2460" // ①
	fmt.Printf("%s\n", uts15.NFKD(circled))

	// Output:
	// fi
	// 1
}

func ExampleNFC_stringComparison() {
	// Comparing strings that may be in different normalization forms
	s1 := "café" // Composed form (might be)
	s2 := "cafe\u0301" // Decomposed form (e + acute)

	// Direct comparison might fail
	fmt.Printf("Direct comparison: %v\n", s1 == s2)

	// Normalize both to NFC before comparing
	fmt.Printf("After NFC: %v\n", uts15.NFC(s1) == uts15.NFC(s2))

	// Output:
	// Direct comparison: false
	// After NFC: true
}

func ExampleIsNFC() {
	// Check if a string is already in NFC form
	composed := "\u00E9" // é (composed)
	decomposed := "e\u0301" // e + combining acute

	fmt.Printf("Composed is NFC: %v\n", uts15.IsNFC(composed))
	fmt.Printf("Decomposed is NFC: %v\n", uts15.IsNFC(decomposed))

	// Output:
	// Composed is NFC: true
	// Decomposed is NFC: false
}

func ExampleNFC_hangul() {
	// Hangul composition
	// Hangul jamos (L+V) compose into syllables
	jamos := "\u1100\u1161" // ᄀ + ᅡ
	syllable := uts15.NFC(jamos)
	fmt.Printf("%s\n", syllable) // 가

	// With trailing consonant (L+V+T)
	jamosWithT := "\u1100\u1161\u11A8" // ᄀ + ᅡ + ᆨ
	syllableWithT := uts15.NFC(jamosWithT)
	fmt.Printf("%s\n", syllableWithT) // 각

	// Output:
	// 가
	// 각
}

func ExampleNFD_hangul() {
	// Hangul decomposition
	syllable := "\uAC00" // 가
	jamos := uts15.NFD(syllable)
	printRunes([]rune(jamos))

	// With trailing consonant
	syllableWithT := "\uAC01" // 각
	jamosWithT := uts15.NFD(syllableWithT)
	printRunes([]rune(jamosWithT))

	// Output:
	// 1100 1161
	// 1100 1161 11A8
}

func printRunes(runes []rune) {
	for i, r := range runes {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Printf("%X", r)
	}
	fmt.Println()
}

func ExampleNFKC_security() {
	// Security: Normalize identifiers to prevent homograph attacks
	// and ensure consistent comparison

	identifiers := []string{
		"file",      // Normal ASCII
		"\uFB01le",  // Using ﬁ ligature
		"\uFF46ile", // Using full-width f
	}

	fmt.Println("Original identifiers:")
	for _, id := range identifiers {
		fmt.Printf("  %q\n", id)
	}

	fmt.Println("\nAfter NFKC normalization:")
	for _, id := range identifiers {
		normalized := uts15.NFKC(id)
		fmt.Printf("  %q\n", normalized)
	}

	// All normalize to the same form
	fmt.Printf("\nAll equal after NFKC: %v\n",
		uts15.NFKC(identifiers[0]) == uts15.NFKC(identifiers[1]) &&
			uts15.NFKC(identifiers[1]) == uts15.NFKC(identifiers[2]))

	// Output:
	// Original identifiers:
	//   "file"
	//   "ﬁle"
	//   "ｆile"
	//
	// After NFKC normalization:
	//   "file"
	//   "file"
	//   "file"
	//
	// All equal after NFKC: true
}

func ExampleNFC_database() {
	// Example: Normalizing text before storing in a database
	// to ensure consistent searching and indexing

	composed := "Jos\u00E9"   // Composed form
	decomposed := "Jose\u0301" // Decomposed form (e + acute)

	normalized1 := uts15.NFC(composed)
	normalized2 := uts15.NFC(decomposed)

	fmt.Printf("Same after normalization: %v\n", normalized1 == normalized2)

	// Output:
	// Same after normalization: true
}
