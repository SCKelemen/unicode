// Package uax11 implements the Unicode East Asian Width algorithm (UAX #11).
//
// This package provides East Asian Width property lookup for characters to
// determine their display width in the context of East Asian typography.
// This is essential for:
//   - Terminal emulators and monospace fonts
//   - Text editors with East Asian language support
//   - Legacy East Asian character encoding conversions
//   - Line breaking and text justification
//
// Based on: https://www.unicode.org/reports/tr11/
//
// The East_Asian_Width property has six values that resolve to two practical widths:
//   - Narrow: Characters that occupy half-width (0.5 em) in fixed-pitch fonts
//   - Wide: Characters that occupy full-width (1.0 em) in fixed-pitch fonts
//
// # Six Property Values
//
//   - Fullwidth (F): Characters with fullwidth compatibility decomposition
//   - Halfwidth (H): Characters with halfwidth compatibility decomposition
//   - Wide (W): Characters that are wide in East Asian context (ideographs)
//   - Narrow (Na): Characters that are narrow with fullwidth counterparts
//   - Ambiguous (A): Can be narrow or wide depending on context
//   - Neutral (N): Characters not in legacy East Asian encodings
//
// # Conformance
//
// As specified in UAX #11 §6 (https://www.unicode.org/reports/tr11/#Recommendations),
// the East_Asian_Width property is informative. It provides default classifications
// that can be overridden based on context, language, or application requirements.
//
// # Usage
//
//	import "github.com/SCKelemen/unicode/uax11"
//
//	// Get width classification for a single character
//	width := uax11.LookupWidth('A')       // Returns Narrow
//	width = uax11.LookupWidth('中')       // Returns Wide
//	width = uax11.LookupWidth('Ω')       // Returns Ambiguous
//
//	// Check if character is wide
//	if uax11.IsWide('中') {
//	    // Character occupies full width (1.0 em)
//	}
//
//	// Calculate string width (with context)
//	width := uax11.StringWidth("Hello世界", uax11.ContextEastAsian)
//
// # Ambiguous Width Handling
//
// Ambiguous (A) characters require contextual resolution per UAX #11 §5
// (https://www.unicode.org/reports/tr11/#Ambiguous). Common contexts include:
//   - East Asian context: Ambiguous characters are treated as Wide
//   - Non-East Asian context: Ambiguous characters are treated as Narrow
//   - Default: Narrow when context cannot be determined
//
// # References
//
//   - UAX #11: https://www.unicode.org/reports/tr11/
//   - §5 Ambiguous Characters: https://www.unicode.org/reports/tr11/#Ambiguous
//   - §6 Recommendations: https://www.unicode.org/reports/tr11/#Recommendations
package uax11

// Width represents the East Asian Width property of a Unicode character.
// See UAX #11 §3: https://www.unicode.org/reports/tr11/#ED1
type Width int

const (
	// Neutral (N) indicates characters not in legacy East Asian character sets.
	// These characters are neutral with respect to East Asian typography.
	// Default for most non–East Asian scripts (Devanagari, Arabic, etc.)
	// and modern symbols (emoji variants).
	// See UAX #11 §3: https://www.unicode.org/reports/tr11/#ED1
	Neutral Width = iota

	// Ambiguous (A) indicates characters that can be narrow or wide depending
	// on context. Examples: Greek, Cyrillic letters in East Asian character sets.
	// Applications must resolve these based on context per UAX #11 §5.
	// See: https://www.unicode.org/reports/tr11/#Ambiguous
	Ambiguous

	// Fullwidth (F) indicates characters with <wide> compatibility decomposition.
	// These are fullwidth forms of characters that also have narrow forms.
	// Examples: Fullwidth Latin letters (U+FF01..U+FF5E)
	// See UAX #11 §3: https://www.unicode.org/reports/tr11/#ED2
	Fullwidth

	// Halfwidth (H) indicates characters with <narrow> compatibility decomposition.
	// These are halfwidth forms, typically for katakana and Hangul.
	// Examples: Halfwidth Katakana (U+FF65..U+FFDC)
	// See UAX #11 §3: https://www.unicode.org/reports/tr11/#ED3
	Halfwidth

	// Narrow (Na) indicates characters that are always narrow and have
	// explicit fullwidth or wide counterparts.
	// Examples: ASCII characters (U+0020..U+007E)
	// See UAX #11 §3: https://www.unicode.org/reports/tr11/#ED4
	Narrow

	// Wide (W) indicates characters that are always wide in East Asian typography.
	// These behave like ideographs: allowing line breaks after each character
	// and remaining upright in vertical text.
	// Examples: CJK ideographs, emoji, fullwidth symbols
	// See UAX #11 §3: https://www.unicode.org/reports/tr11/#ED5
	Wide
)

// String returns the standard abbreviated form of the width value.
func (w Width) String() string {
	switch w {
	case Neutral:
		return "N"
	case Ambiguous:
		return "A"
	case Fullwidth:
		return "F"
	case Halfwidth:
		return "H"
	case Narrow:
		return "Na"
	case Wide:
		return "W"
	default:
		return "Unknown"
	}
}

// Context represents the contextual environment for resolving Ambiguous widths.
// See UAX #11 §5: https://www.unicode.org/reports/tr11/#Ambiguous
type Context int

const (
	// ContextNarrow treats Ambiguous characters as narrow (default).
	// Use for non-East Asian environments.
	ContextNarrow Context = iota

	// ContextEastAsian treats Ambiguous characters as wide.
	// Use for East Asian typography contexts (Chinese, Japanese, Korean).
	ContextEastAsian
)

// LookupWidth returns the East_Asian_Width property value for a given rune.
//
// This function performs a binary search on the Unicode character database to find
// the width property. Characters not explicitly listed default to Neutral, except:
//   - Unassigned CJK ideograph ranges default to Wide
//   - Unassigned code points in Planes 2 and 3 default to Wide
//
// Per UAX #11 data file specification:
// https://www.unicode.org/Public/UCD/latest/ucd/EastAsianWidth.txt
//
// Example:
//
//	width := uax11.LookupWidth('A')      // Returns Narrow
//	width = uax11.LookupWidth('中')      // Returns Wide
//	width = uax11.LookupWidth('Ω')      // Returns Ambiguous
func LookupWidth(r rune) Width {
	// Binary search through the ranges
	left, right := 0, len(eastAsianWidthData)-1

	for left <= right {
		mid := (left + right) / 2
		entry := eastAsianWidthData[mid]

		if r < entry.start {
			right = mid - 1
		} else if r > entry.end {
			left = mid + 1
		} else {
			// Found the range containing r
			return entry.width
		}
	}

	// Check for default Wide ranges per UAX #11 data file
	// CJK Unified Ideographs Extension A: U+3400..U+4DBF
	// CJK Unified Ideographs: U+4E00..U+9FFF
	// CJK Compatibility Ideographs: U+F900..U+FAFF
	if (r >= 0x3400 && r <= 0x4DBF) ||
		(r >= 0x4E00 && r <= 0x9FFF) ||
		(r >= 0xF900 && r <= 0xFAFF) {
		return Wide
	}

	// Plane 2: U+20000..U+2FFFD
	// Plane 3: U+30000..U+3FFFD
	if (r >= 0x20000 && r <= 0x2FFFD) ||
		(r >= 0x30000 && r <= 0x3FFFD) {
		return Wide
	}

	// Default to Neutral if not found
	// Per UAX #11 data file: "@missing: 0000..10FFFF; N"
	return Neutral
}

// IsWide returns true if the character has Wide or Fullwidth classification.
//
// These characters occupy full width (1.0 em) in fixed-pitch fonts and behave
// like ideographs in East Asian typography.
// See UAX #11 §3: https://www.unicode.org/reports/tr11/#ED5
//
// Example:
//
//	if uax11.IsWide('中') {
//	    // Display character in full width
//	}
func IsWide(r rune) bool {
	w := LookupWidth(r)
	return w == Wide || w == Fullwidth
}

// IsNarrow returns true if the character has Narrow or Halfwidth classification.
//
// These characters occupy half width (0.5 em) in fixed-pitch fonts.
// See UAX #11 §3: https://www.unicode.org/reports/tr11/#ED4
//
// Example:
//
//	if uax11.IsNarrow('A') {
//	    // Display character in half width
//	}
func IsNarrow(r rune) bool {
	w := LookupWidth(r)
	return w == Narrow || w == Halfwidth
}

// IsAmbiguous returns true if the character has Ambiguous classification.
//
// Ambiguous characters require contextual resolution per UAX #11 §5.
// See: https://www.unicode.org/reports/tr11/#Ambiguous
//
// Example:
//
//	if uax11.IsAmbiguous('Ω') {
//	    // Resolve width based on context
//	}
func IsAmbiguous(r rune) bool {
	return LookupWidth(r) == Ambiguous
}

// ResolveWidth returns the practical display width for a character in a given context.
//
// For Ambiguous characters, the width is resolved based on context per UAX #11 §5:
//   - ContextEastAsian: Ambiguous → Wide
//   - ContextNarrow: Ambiguous → Narrow (default)
//
// For all other characters, returns their intrinsic width classification.
// See: https://www.unicode.org/reports/tr11/#Ambiguous
//
// Example:
//
//	// Greek Omega in East Asian context
//	width := uax11.ResolveWidth('Ω', uax11.ContextEastAsian)  // Returns Wide
//
//	// Greek Omega in non-East Asian context
//	width = uax11.ResolveWidth('Ω', uax11.ContextNarrow)      // Returns Narrow
func ResolveWidth(r rune, ctx Context) Width {
	w := LookupWidth(r)

	if w == Ambiguous {
		if ctx == ContextEastAsian {
			return Wide
		}
		return Narrow
	}

	return w
}

// CharWidth returns the display width of a character (1 or 2) in the given context.
//
// This is useful for calculating string display widths in monospace fonts:
//   - Wide and Fullwidth characters: 2 units
//   - All other characters: 1 unit
//   - Ambiguous characters: resolved based on context
//
// Note: This is a simplified width calculation. UAX #11 §6 states:
// "The East_Asian_Width property is not intended for use by modern terminal
// emulators without appropriate tailoring on a case-by-case basis."
// See: https://www.unicode.org/reports/tr11/#Recommendations
//
// Example:
//
//	width := uax11.CharWidth('A', uax11.ContextNarrow)       // Returns 1
//	width = uax11.CharWidth('中', uax11.ContextNarrow)       // Returns 2
//	width = uax11.CharWidth('Ω', uax11.ContextEastAsian)    // Returns 2
func CharWidth(r rune, ctx Context) int {
	w := ResolveWidth(r, ctx)
	if w == Wide || w == Fullwidth {
		return 2
	}
	return 1
}

// StringWidth calculates the total display width of a string in the given context.
//
// This sums the display widths of all characters in the string. Note that this
// is a simplified calculation that doesn't account for combining marks, control
// characters, or other special cases.
//
// For proper text width calculation, consider combining with:
//   - UAX #29 for grapheme cluster segmentation
//   - UAX #14 for line breaking
//
// Example:
//
//	width := uax11.StringWidth("Hello", uax11.ContextNarrow)         // Returns 5
//	width = uax11.StringWidth("Hello世界", uax11.ContextNarrow)      // Returns 9 (5 + 4)
//	width = uax11.StringWidth("ΩΩΩ", uax11.ContextEastAsian)        // Returns 6
func StringWidth(s string, ctx Context) int {
	width := 0
	for _, r := range s {
		width += CharWidth(r, ctx)
	}
	return width
}
