// Package uax31 implements Unicode Identifier and Pattern Syntax (UAX #31).
//
// This package provides Unicode properties for identifier characters and
// pattern syntax characters. These properties are fundamental for:
//   - Programming language identifiers (variable names, function names, etc.)
//   - Security identifiers (usernames, domain names, etc.)
//   - Pattern-based syntax (regular expressions, query languages)
//   - Text processing and parsing
//
// Based on: https://www.unicode.org/reports/tr31/
//
// # Properties
//
// XID_Start: Characters valid at the start of an identifier
//   - Unicode letters, ideographs, letter numbers
//   - Excludes Pattern_Syntax and Pattern_White_Space
//
// XID_Continue: Characters valid after the first character in an identifier
//   - XID_Start plus nonspacing marks, spacing marks, decimal numbers
//   - Connector punctuation and a few other categories
//
// Pattern_Syntax: Characters reserved for use in patterns and syntax
//   - ASCII punctuation and mathematical symbols
//   - Used to identify syntactic elements in pattern languages
//
// Pattern_White_Space: Characters treated as whitespace in patterns
//   - Spaces, tabs, line breaks, form feeds
//   - Used for pattern tokenization
//
// # Identifier Validation
//
// The XID_Start and XID_Continue properties define Default Identifiers,
// which are stable across Unicode versions and suitable for most uses:
//
//	<Identifier> := <XID_Start> <XID_Continue>*
//
// # Conformance
//
// This implementation follows UAX #31 Default Identifier Syntax (§2.3):
//   - https://www.unicode.org/reports/tr31/#Default_Identifier_Syntax
//
// The implementation uses the stable XID_Start and XID_Continue properties
// derived from Unicode 17.0.0 data files.
//
// # Usage
//
//	import "github.com/SCKelemen/unicode/uax31"
//
//	// Check if character can start an identifier
//	if uax31.IsXIDStart('A') {
//	    // Valid identifier start
//	}
//
//	// Check if character can continue an identifier
//	if uax31.IsXIDContinue('5') {
//	    // Valid in identifier (after first character)
//	}
//
//	// Validate complete identifier
//	if uax31.IsValidIdentifier("myVar123") {
//	    // Valid identifier
//	}
//
//	// Check if character is pattern syntax
//	if uax31.IsPatternSyntax('*') {
//	    // Character is syntax, not identifier content
//	}
//
// # References
//
//   - UAX #31: https://www.unicode.org/reports/tr31/
//   - §2.3 Default Identifier Syntax: https://www.unicode.org/reports/tr31/#Default_Identifier_Syntax
//   - DerivedCoreProperties.txt: https://www.unicode.org/Public/17.0.0/ucd/DerivedCoreProperties.txt
//   - PropList.txt: https://www.unicode.org/Public/17.0.0/ucd/PropList.txt
package uax31

// IsXIDStart reports whether the rune has the XID_Start property.
//
// XID_Start characters are valid at the start of an identifier.
// This includes letters, ideographs, and letter numbers, but excludes
// pattern syntax and pattern whitespace characters.
//
// Examples:
//   - Returns true for: 'A', 'z', '中', 'α', 'א'
//   - Returns false for: '5', '_', '+', ' '
//
// See: https://www.unicode.org/reports/tr31/#Table_Lexical_Classes_for_Identifiers
func IsXIDStart(r rune) bool {
	return inRanges(r, xidStartData)
}

// IsXIDContinue reports whether the rune has the XID_Continue property.
//
// XID_Continue characters are valid after the first character in an identifier.
// This includes all XID_Start characters plus marks, digits, connector
// punctuation, and a few other categories.
//
// Examples:
//   - Returns true for: 'A', '5', '_', '́' (combining acute)
//   - Returns false for: '+', '*', ' ', '('
//
// See: https://www.unicode.org/reports/tr31/#Table_Lexical_Classes_for_Identifiers
func IsXIDContinue(r rune) bool {
	return inRanges(r, xidContinueData)
}

// IsPatternSyntax reports whether the rune has the Pattern_Syntax property.
//
// Pattern_Syntax characters are reserved for use in patterns and syntax.
// These characters should not appear in identifiers to avoid ambiguity
// in pattern-based languages.
//
// Examples:
//   - Returns true for: '!', '#', '$', '%', '(', ')', '*', '+', ',', '-', '.', '/', ':', ';', '<', '=', '>', '?', '@', '[', '\\', ']', '^', '`', '{', '|', '}', '~'
//   - Returns false for: 'A', '5', '_', ' '
//
// See: https://www.unicode.org/reports/tr31/#Pattern_Syntax
func IsPatternSyntax(r rune) bool {
	return inRanges(r, patternSyntaxData)
}

// IsPatternWhiteSpace reports whether the rune has the Pattern_White_Space property.
//
// Pattern_White_Space characters are treated as whitespace in patterns.
// This property is used for pattern tokenization and includes spaces,
// tabs, line breaks, and form feeds.
//
// Examples:
//   - Returns true for: ' ' (space), '\t' (tab), '\n' (newline), '\r' (carriage return), '\f' (form feed), '\v' (vertical tab)
//   - Returns false for: 'A', '5', '_', '+'
//
// See: https://www.unicode.org/reports/tr31/#Pattern_White_Space
func IsPatternWhiteSpace(r rune) bool {
	return inRanges(r, patternWhiteSpaceData)
}

// IsValidIdentifierStart reports whether the rune is valid as the first
// character of an identifier.
//
// This is equivalent to IsXIDStart(r) && !IsPatternSyntax(r) && !IsPatternWhiteSpace(r),
// though the last two checks are redundant as XID_Start already excludes these.
func IsValidIdentifierStart(r rune) bool {
	return IsXIDStart(r)
}

// IsValidIdentifierContinue reports whether the rune is valid as a non-first
// character in an identifier.
//
// This is equivalent to IsXIDContinue(r) && !IsPatternSyntax(r) && !IsPatternWhiteSpace(r),
// though the last two checks are redundant as XID_Continue already excludes these.
func IsValidIdentifierContinue(r rune) bool {
	return IsXIDContinue(r)
}

// IsValidIdentifier reports whether the string is a valid default identifier
// according to UAX #31 Default Identifier Syntax.
//
// A valid identifier has the pattern: <XID_Start> <XID_Continue>*
//
// This means:
//   - The first character must have the XID_Start property
//   - All subsequent characters must have the XID_Continue property
//   - The string must not be empty
//
// Example:
//	uax31.IsValidIdentifier("myVar")     // true
//	uax31.IsValidIdentifier("my_var")    // true
//	uax31.IsValidIdentifier("myVar123")  // true
//	uax31.IsValidIdentifier("123var")    // false (starts with digit)
//	uax31.IsValidIdentifier("my-var")    // false (hyphen not in XID_Continue)
//	uax31.IsValidIdentifier("")          // false (empty)
//
// See: https://www.unicode.org/reports/tr31/#Default_Identifier_Syntax
func IsValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}

	runes := []rune(s)

	// First character must be XID_Start
	if !IsXIDStart(runes[0]) {
		return false
	}

	// Remaining characters must be XID_Continue
	for _, r := range runes[1:] {
		if !IsXIDContinue(r) {
			return false
		}
	}

	return true
}

// inRanges performs a binary search to determine if r is in any of the ranges.
// The ranges slice must be sorted by start value.
func inRanges(r rune, ranges []idRange) bool {
	// Binary search
	lo, hi := 0, len(ranges)
	for lo < hi {
		mid := lo + (hi-lo)/2
		if r < ranges[mid].start {
			hi = mid
		} else if r > ranges[mid].end {
			lo = mid + 1
		} else {
			return true
		}
	}
	return false
}
