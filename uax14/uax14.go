// Package uax14 implements the Unicode Line Breaking Algorithm (UAX #14).
//
// This package provides line break opportunity detection for text layout systems.
// It analyzes text and identifies positions where lines can be broken according
// to the Unicode Standard Annex #14 specification.
//
// This code was originally implemented in github.com/SCKelemen/layout and has been
// extracted to a standalone package for reusability across multiple projects.
//
// Based on: https://www.unicode.org/reports/tr14/
//
// Usage:
//
//	import "github.com/SCKelemen/unicode/uax14"
//
//	text := "Hello world! This is a test."
//	breakPoints := uax14.FindLineBreakOpportunities(text, uax14.HyphensManual)
//	// breakPoints contains byte positions where line breaks are allowed
//
// The implementation focuses on practical line breaking for word boundaries
// and common text layout scenarios, with support for:
//   - Mandatory breaks (newlines, paragraph separators)
//   - Word boundaries (spaces)
//   - Hyphenation (soft hyphens with configurable modes)
//   - Ideographic text (CJK characters)
//   - Punctuation and numeric sequences
//
// Reference implementation: https://pkg.go.dev/github.com/gorilla/i18n/linebreak
package uax14

import (
	"unicode"
)

// Hyphens controls automatic hyphenation behavior.
// Based on CSS Text Module Level 3 §4.3: https://www.w3.org/TR/css-text-3/#hyphenation
type Hyphens int

const (
	// HyphensNone disables all hyphenation (no breaks at hyphens)
	HyphensNone Hyphens = iota
	// HyphensManual only allows breaks at U+00AD soft hyphens
	HyphensManual
	// HyphensAuto allows automatic hyphenation with dictionaries (not yet fully implemented)
	HyphensAuto
)

// BreakClass represents a Unicode line breaking class.
type BreakClass int

const (
	// Mandatory breaks
	ClassBK BreakClass = iota // Mandatory Break
	ClassCR                   // Carriage Return
	ClassLF                   // Line Feed
	ClassNL                   // Next Line
	ClassSP                   // Space

	// Prohibited breaks
	ClassWJ BreakClass = iota + 5 // Word Joiner
	ClassZW                       // Zero Width Space

	// Break opportunities
	ClassBA BreakClass = iota + 10 // Break After
	ClassBB                        // Break Before
	ClassB2                        // Break Opportunity Before and After
	ClassHY                        // Hyphen
	ClassCB                        // Contingent Break Opportunity

	// Characters
	ClassAL BreakClass = iota + 20 // Alphabetic
	ClassHL                        // Hebrew Letter
	ClassID                        // Ideographic
	ClassIN                        // Inseparable
	ClassNU                        // Numeric
	ClassPR                        // Prefix Numeric
	ClassPO                        // Postfix Numeric
	ClassIS                        // Infix Numeric Separator
	ClassSY                        // Symbols Allowing Break After
	ClassAI                        // Ambiguous (Alphabetic or Ideographic) - East Asian Width
	ClassCJ                        // Conditional Japanese Starter
	ClassSA                        // Complex Context Dependent (South East Asian)
	ClassAK                        // Aksara (Indic scripts)
	ClassAP                        // Aksara Prebase (Indic scripts)
	ClassAS                        // Aksara Start (Indic scripts)

	// Punctuation
	ClassOP BreakClass = iota + 40 // Open Punctuation
	ClassCL                        // Close Punctuation
	ClassCP                        // Close Parenthesis
	ClassQU                        // Quotation
	ClassGL                        // Non-breaking ("Glue")
	ClassNS                        // Nonstarter
	ClassEX                        // Exclamation/Interrogation

	// Combining marks
	ClassCM BreakClass = iota + 60 // Combining Mark

	// Hangul
	ClassJL BreakClass = iota + 70 // Hangul L Jamo
	ClassJV                        // Hangul V Jamo
	ClassJT                        // Hangul T Jamo
	ClassH2                        // Hangul LV Syllable
	ClassH3                        // Hangul LVT Syllable

	// Regional indicators
	ClassRI BreakClass = iota + 80 // Regional Indicator

	// Surrogates
	ClassSG BreakClass = iota + 90 // Surrogate

	// Unknown
	ClassXX BreakClass = iota + 100 // Unknown
)

// BreakAction represents the action to take at a line break opportunity.
type BreakAction int

const (
	// BreakProhibited means no line break is allowed
	BreakProhibited BreakAction = iota
	// BreakDirect means a line break is allowed
	BreakDirect
	// BreakIndirect means a line break is allowed only if preceded by space
	BreakIndirect
	// BreakMandatory means a line break is required
	BreakMandatory
)

// getBreakClass returns the line breaking class for a rune.
// This is a simplified implementation focusing on common cases.
// Reference: http://www.unicode.org/reports/tr14/#Table1
func getBreakClass(r rune) BreakClass {
	// Mandatory breaks
	switch r {
	case '\n':
		return ClassLF
	case '\r':
		return ClassCR
	case '\u0085': // NEL (Next Line)
		return ClassNL
	case '\u2028': // Line Separator
		return ClassBK
	case '\u2029': // Paragraph Separator
		return ClassBK
	}

	// Space characters
	if r == ' ' || r == '\t' {
		return ClassSP
	}

	// Non-breaking space (treated as regular character for our purposes)
	if r == '\u00A0' {
		return ClassGL // Non-breaking, similar to Word Joiner
	}

	// Zero Width Space (allows break)
	if r == '\u200B' {
		return ClassZW
	}

	// Word Joiner (prohibits break)
	if r == '\u2060' {
		return ClassWJ
	}

	// Soft Hyphen (allows break)
	if r == '\u00AD' {
		return ClassCB
	}

	// Break Before characters
	if r == '\u00B4' { // Acute accent
		return ClassBB
	}

	// Punctuation
	switch r {
	case '(', '[', '{', '⟨', '｟':
		return ClassOP
	case ')', ']', '}', '⟩', '｠':
		return ClassCP
	case '"', '\'', '«', '»', '„', '‚', '‹', '›':
		return ClassQU
	case '!', '?':
		return ClassEX
	case '-', '–', '—':
		return ClassHY
	case '/':
		return ClassSY
	case ',':
		return ClassIS
	case '.':
		return ClassIS
	case ':':
		return ClassIS
	case ';':
		return ClassIS
	}

	// CJK brackets and punctuation (U+3000-303F)
	// Even codepoints are opening, odd are closing
	if (r >= 0x3008 && r <= 0x3011) || (r >= 0x3014 && r <= 0x301B) {
		if r%2 == 0 {
			return ClassOP
		}
		return ClassCL
	}

	// Numeric
	if unicode.Is(unicode.N, r) {
		return ClassNU
	}

	// Combining marks
	if unicode.Is(unicode.M, r) {
		return ClassCM
	}

	// Ideographic (CJK)
	if unicode.Is(unicode.Ideographic, r) {
		return ClassID
	}

	// Hebrew letters
	if unicode.Is(unicode.Hebrew, r) {
		return ClassHL
	}

	// Alphabetic (default for letters)
	if unicode.Is(unicode.L, r) {
		return ClassAL
	}

	// Ambiguous East Asian Width characters (AI)
	// These should be treated as ideographic in East Asian contexts
	// Common AI ranges per UAX #14 and East Asian Width property
	if isAmbiguousEastAsian(r) {
		return ClassAI
	}

	// Symbols
	if unicode.Is(unicode.S, r) {
		return ClassSY
	}

	// Default: alphabetic
	return ClassAL
}

// isAmbiguousEastAsian checks if a rune is in the Ambiguous (A) East Asian Width category
// Per UAX #11 (East Asian Width) and UAX #14, these characters have ambiguous width
// and should allow line breaks in East Asian contexts like ideographs
func isAmbiguousEastAsian(r rune) bool {
	// Common ambiguous ranges - not exhaustive but covers most cases
	switch {
	// Miscellaneous Symbols (includes ❗ U+2757)
	case r >= 0x2600 && r <= 0x26FF:
		return true
	// Dingbats
	case r >= 0x2700 && r <= 0x27BF:
		return true
	// Common ambiguous punctuation and symbols
	case r == 0x00A7 || r == 0x00A8: // § ¨ (AI)
		return true
	case r == 0x00B0: // ° DEGREE SIGN (AI)
		return true
	case r == 0x00B2 || r == 0x00B3: // ² ³ SUPERSCRIPTS (AI)
		return true
	case r == 0x00B6 || r == 0x00B7: // ¶ · (AI)
		return true
	case r >= 0x2010 && r <= 0x2027: // Various dashes and punctuation
		return true
	case r >= 0x2030 && r <= 0x205E: // Various punctuation and symbols
		return true
	}
	return false
}

// pairTable defines line breaking actions for adjacent character classes.
// Simplified version focusing on common cases.
// Reference: http://www.unicode.org/reports/tr14/#Table2
var pairTable = map[[2]BreakClass]BreakAction{
	// Mandatory breaks
	{ClassLF, ClassXX}: BreakMandatory,
	{ClassCR, ClassXX}: BreakMandatory,
	{ClassCR, ClassLF}: BreakProhibited, // CR+LF is one break
	{ClassBK, ClassXX}: BreakMandatory,
	{ClassNL, ClassXX}: BreakMandatory,

	// Space
	{ClassSP, ClassXX}: BreakIndirect,
	{ClassXX, ClassSP}: BreakProhibited,

	// Prohibited breaks
	{ClassWJ, ClassXX}: BreakProhibited,
	{ClassXX, ClassWJ}: BreakProhibited,
	{ClassZW, ClassXX}: BreakDirect, // Zero Width Space allows break

	// Break after
	{ClassBA, ClassXX}: BreakDirect,
	{ClassHY, ClassXX}: BreakDirect,
	{ClassSY, ClassXX}: BreakDirect,

	// Break before
	{ClassXX, ClassBB}: BreakDirect,

	// Break before and after
	{ClassB2, ClassXX}: BreakDirect,
	{ClassXX, ClassB2}: BreakDirect,

	// Contingent break
	{ClassCB, ClassXX}: BreakDirect,

	// Punctuation
	{ClassOP, ClassXX}: BreakProhibited,
	{ClassQU, ClassXX}: BreakProhibited,
	{ClassGL, ClassXX}: BreakProhibited,
	{ClassXX, ClassOP}: BreakProhibited,
	{ClassXX, ClassQU}: BreakProhibited,
	{ClassXX, ClassGL}: BreakProhibited,

	// Close punctuation
	{ClassCL, ClassXX}: BreakProhibited,
	{ClassCP, ClassXX}: BreakProhibited,
	{ClassXX, ClassCL}: BreakProhibited,
	{ClassXX, ClassCP}: BreakProhibited,

	// Numeric
	{ClassNU, ClassNU}: BreakProhibited,
	{ClassNU, ClassAL}: BreakProhibited,
	{ClassAL, ClassNU}: BreakProhibited,
	{ClassPR, ClassNU}: BreakProhibited,
	{ClassNU, ClassPO}: BreakProhibited,
	{ClassIS, ClassNU}: BreakProhibited,
	{ClassNU, ClassIS}: BreakProhibited,

	// Ideographic
	{ClassID, ClassID}: BreakDirect,
	{ClassID, ClassAL}: BreakDirect,
	{ClassAL, ClassID}: BreakDirect,
	{ClassID, ClassNU}: BreakDirect,
	{ClassNU, ClassID}: BreakDirect,

	// Ambiguous East Asian (AI) - LB28: Do not break between alphabetics
	// AI behaves like AL when adjacent to alphabetics (no break)
	{ClassAI, ClassAI}: BreakProhibited,
	{ClassAI, ClassAL}: BreakProhibited,
	{ClassAL, ClassAI}: BreakProhibited,
	{ClassAI, ClassHL}: BreakProhibited,
	{ClassHL, ClassAI}: BreakProhibited,
	// AI allows breaks with special Indic classes (AK, AP, AS)
	{ClassAI, ClassAK}: BreakDirect,
	{ClassAK, ClassAI}: BreakDirect,
	{ClassAI, ClassAP}: BreakDirect,
	{ClassAP, ClassAI}: BreakDirect,
	{ClassAI, ClassAS}: BreakDirect,
	{ClassAS, ClassAI}: BreakDirect,

	// Combining marks (prohibited break before)
	{ClassXX, ClassCM}: BreakProhibited,
	{ClassCM, ClassCM}: BreakProhibited,
}

// getBreakAction returns the break action between two character classes.
func getBreakAction(before, after BreakClass) BreakAction {
	// Try exact match first
	if action, ok := pairTable[[2]BreakClass{before, after}]; ok {
		return action
	}

	// Try wildcard patterns: {before, XX} and {XX, after}
	if action, ok := pairTable[[2]BreakClass{before, ClassXX}]; ok {
		return action
	}
	if action, ok := pairTable[[2]BreakClass{ClassXX, after}]; ok {
		return action
	}

	// Default rules
	if before == ClassSP {
		return BreakIndirect
	}
	if after == ClassSP {
		return BreakProhibited
	}

	// Default: allow break (for word boundaries)
	return BreakDirect
}

// FindLineBreakOpportunities finds all valid line break opportunities in text.
// Returns a slice of byte positions where breaks are allowed.
//
// This implements UAX #14 but focuses on word boundaries for practical line breaking.
// The hyphens parameter controls hyphenation behavior:
//   - HyphensNone: No breaks at hyphens (hard or soft)
//   - HyphensManual: Only break at U+00AD soft hyphens
//   - HyphensAuto: Break at all hyphens (dictionary-based hyphenation not yet implemented)
//
// The returned slice always includes position 0 (start) and len(text) (end).
// All positions are byte offsets, not rune indices, for direct string slicing.
//
// Example:
//
//	text := "Hello world"
//	breaks := FindLineBreakOpportunities(text, HyphensManual)
//	// breaks = [0, 6, 11] - can break at start, after "Hello ", and at end
func FindLineBreakOpportunities(text string, hyphens Hyphens) []int {
	if text == "" {
		return []int{0}
	}

	var breakPoints []int
	breakPoints = append(breakPoints, 0) // Start is always a break point

	runes := []rune(text)
	if len(runes) == 0 {
		return breakPoints
	}

	prevClass := getBreakClass(runes[0])

	for i := 1; i < len(runes); i++ {
		currClass := getBreakClass(runes[i])
		action := getBreakAction(prevClass, currClass)

		// Only add break points for:
		// 1. Mandatory breaks (newlines, etc.)
		// 2. Spaces (word boundaries)
		// 3. Explicit break opportunities (hyphens, etc.) - respecting hyphens property
		switch action {
		case BreakMandatory:
			// Mandatory break - always add
			bytePos := len(string(runes[:i]))
			breakPoints = append(breakPoints, bytePos)
		case BreakIndirect:
			// Indirect break (usually spaces) - add for word boundaries
			// But respect LB6 and LB13
			if prevClass == ClassSP {
				// LB6: Do not break before hard line breaks (BK, CR, LF, NL)
				// LB13: Do not break before CL, CP, EX, IS (closing punct)
				// LB16: Do not break before NS (nonstarters)
				if currClass == ClassBK || currClass == ClassCR || currClass == ClassLF ||
					currClass == ClassNL || currClass == ClassCL || currClass == ClassCP ||
					currClass == ClassEX || currClass == ClassIS || currClass == ClassNS {
					// Don't break
				} else {
					bytePos := len(string(runes[:i]))
					breakPoints = append(breakPoints, bytePos)
				}
			}
		case BreakDirect:
			// Direct break - add for explicit break characters and ideographic text
			// Don't break between regular alphabetic characters (to keep words together)
			if prevClass == ClassZW {
				// Zero-width space always allows break
				bytePos := len(string(runes[:i]))
				breakPoints = append(breakPoints, bytePos)
			} else if prevClass == ClassHY || prevClass == ClassCB || prevClass == ClassBA || prevClass == ClassB2 {
				// Explicit break opportunities (hyphens, soft hyphens, etc.)
				// Respect the hyphens property
				isSoftHyphen := i > 0 && runes[i-1] == '\u00AD'

				if hyphens == HyphensNone {
					// Don't break at any hyphens (hard or soft)
					// Skip adding this break point
				} else if hyphens == HyphensManual && isSoftHyphen {
					// Only break at soft hyphens (U+00AD) in manual mode
					bytePos := len(string(runes[:i]))
					breakPoints = append(breakPoints, bytePos)
				} else if hyphens == HyphensManual && !isSoftHyphen {
					// Don't break at hard hyphens in manual mode
					// Skip adding this break point
				} else if hyphens == HyphensAuto {
					// Break at all hyphens (hard and soft) in auto mode
					// TODO: Add dictionary-based automatic hyphenation
					bytePos := len(string(runes[:i]))
					breakPoints = append(breakPoints, bytePos)
				}
			} else if prevClass == ClassSP {
				// LB18: Break after spaces (word boundaries)
				// But respect LB6 and LB13
				// LB6: Do not break before hard line breaks
				// LB13: Do not break before CL, CP, EX, IS
				// LB16: Do not break before NS (nonstarters)
				if currClass == ClassBK || currClass == ClassCR || currClass == ClassLF ||
					currClass == ClassNL || currClass == ClassCL || currClass == ClassCP ||
					currClass == ClassEX || currClass == ClassIS || currClass == ClassNS {
					// Don't break before these characters, even after space
				} else {
					bytePos := len(string(runes[:i]))
					breakPoints = append(breakPoints, bytePos)
				}
			} else if prevClass == ClassID || currClass == ClassID {
				// Allow breaks involving ideographic characters (CJK text)
				// Per UAX #14, ideographic characters can break between each other
				bytePos := len(string(runes[:i]))
				breakPoints = append(breakPoints, bytePos)
			}
		}

		// Update previous class (combining marks don't change it)
		if currClass != ClassCM {
			prevClass = currClass
		}
	}

	// End of text is always a break point
	breakPoints = append(breakPoints, len(text))

	return breakPoints
}
