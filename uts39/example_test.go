package uts39_test

import (
	"fmt"

	"github.com/SCKelemen/unicode/uts39"
)

func ExampleSkeleton() {
	// The skeleton algorithm identifies confusable strings
	s1 := "paypal"
	s2 := "pаypal" // Contains Cyrillic 'а' (U+0430)

	skel1 := uts39.Skeleton(s1)
	skel2 := uts39.Skeleton(s2)

	fmt.Printf("'%s' skeleton: %q\n", s1, skel1)
	fmt.Printf("'%s' skeleton: %q\n", s2, skel2)
	fmt.Printf("Same skeleton: %v\n", skel1 == skel2)

	// Output:
	// 'paypal' skeleton: "paypal"
	// 'pаypal' skeleton: "paypal"
	// Same skeleton: true
}

func ExampleAreConfusable() {
	// Check if two strings are visually confusable
	fmt.Println(uts39.AreConfusable("scope", "ѕсоре")) // Cyrillic lookalikes
	fmt.Println(uts39.AreConfusable("hello", "world"))
	fmt.Println(uts39.AreConfusable("Test", "test")) // Case-insensitive

	// Output:
	// true
	// false
	// true
}

func ExampleGetRestrictionLevel() {
	// Check the restriction level of identifiers
	examples := []string{
		"hello_world",  // ASCII-Only
		"café",         // Single-Script (Latin)
		"hello世界",    // Minimally-Restrictive (Latin + Han)
		"hello мир",    // Minimally-Restrictive (Latin + Cyrillic)
	}

	for _, s := range examples {
		level := uts39.GetRestrictionLevel(s)
		fmt.Printf("%q: %s\n", s, level)
	}

	// Output:
	// "hello_world": ASCII-Only
	// "café": Single-Script
	// "hello世界": Minimally-Restrictive
	// "hello мир": Minimally-Restrictive
}

func ExampleIsMixedScript() {
	// Detect if an identifier mixes multiple scripts
	fmt.Println(uts39.IsMixedScript("hello"))     // Single script
	fmt.Println(uts39.IsMixedScript("hello123"))  // Numbers are Common
	fmt.Println(uts39.IsMixedScript("hello世界"))  // Latin + Han
	fmt.Println(uts39.IsMixedScript("café"))      // Single script

	// Output:
	// false
	// false
	// true
	// false
}

func ExampleIsSafeIdentifier() {
	// Check if an identifier is safe from security issues
	fmt.Println(uts39.IsSafeIdentifier("user_name"))
	fmt.Println(uts39.IsSafeIdentifier("userName"))
	fmt.Println(uts39.IsSafeIdentifier("user\u200Bname")) // Zero-width space
	fmt.Println(uts39.IsSafeIdentifier("123user"))        // Invalid start

	// Output:
	// true
	// true
	// false
	// false
}

func ExampleAreConfusable_homographAttack() {
	// Detecting homograph attacks in domain names or usernames
	legitimate := "paypal"
	suspicious := "pаypal" // Cyrillic 'а'

	if uts39.AreConfusable(legitimate, suspicious) {
		fmt.Println("Warning: Potential homograph attack detected!")
		fmt.Printf("'%s' and '%s' are visually similar\n", legitimate, suspicious)
	}

	// Output:
	// Warning: Potential homograph attack detected!
	// 'paypal' and 'pаypal' are visually similar
}

func ExampleGetRestrictionLevel_validation() {
	// Validating usernames with different security policies

	type Policy struct {
		name     string
		minLevel uts39.RestrictionLevel
	}

	policies := []Policy{
		{"strict", uts39.ASCIIOnly},
		{"moderate", uts39.SingleScript},
		{"relaxed", uts39.MinimallyRestrictive},
	}

	username := "user_name"
	level := uts39.GetRestrictionLevel(username)

	fmt.Printf("Username: %q\n", username)
	fmt.Printf("Restriction level: %s\n\n", level)

	for _, policy := range policies {
		if level >= policy.minLevel {
			fmt.Printf("✓ Allowed under %s policy\n", policy.name)
		} else {
			fmt.Printf("✗ Rejected by %s policy\n", policy.name)
		}
	}

	// Output:
	// Username: "user_name"
	// Restriction level: ASCII-Only
	//
	// ✓ Allowed under strict policy
	// ✓ Allowed under moderate policy
	// ✓ Allowed under relaxed policy
}

func ExampleGetIdentifierScripts() {
	// Analyze which scripts are used in an identifier
	identifier := "hello"
	scripts := uts39.GetIdentifierScripts(identifier)

	fmt.Printf("Identifier: %q\n", identifier)
	fmt.Printf("Number of scripts: %d\n", len(scripts))
	fmt.Printf("Is mixed-script: %v\n", uts39.IsMixedScript(identifier))

	// Output:
	// Identifier: "hello"
	// Number of scripts: 1
	// Is mixed-script: false
}

func ExampleSkeleton_normalization() {
	// Different Unicode representations normalize to the same skeleton
	composed := "café"    // Precomposed é (U+00E9)
	decomposed := "cafe\u0301" // e + combining acute accent

	// Both normalize to the same skeleton
	if uts39.Skeleton(composed) == uts39.Skeleton(decomposed) {
		fmt.Println("Different representations have the same skeleton")
	}

	// Output:
	// Different representations have the same skeleton
}
