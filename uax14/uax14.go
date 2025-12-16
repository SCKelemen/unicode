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

	"github.com/SCKelemen/unicode/uax11"
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
	ClassWJ  BreakClass = iota + 5 // Word Joiner
	ClassZW                         // Zero Width Space
	ClassZWJ                        // Zero Width Joiner

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
	ClassVF                        // Virama Final (Indic scripts)
	ClassVI                        // Virama (Indic scripts)
	ClassHH                        // Hebrew Letter for Dictionary-based Breaking

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

	// Emoji
	ClassEB BreakClass = iota + 85 // Emoji Base
	ClassEM                        // Emoji Modifier

	// Surrogates
	ClassSG BreakClass = iota + 90 // Surrogate

	// Unknown
	ClassXX BreakClass = iota + 100 // Unknown

	// East Asian Width variants (for maximum conformance)
	// These distinguish characters with East Asian Width property
	ClassAI_EA BreakClass = iota + 110 // Ambiguous (East Asian context)
	ClassAL_EA                          // Alphabetic (East Asian context)
	ClassBA_EA                          // Break After (East Asian context)
	ClassCL_EA                          // Close Punctuation (East Asian context)
	ClassCM_EA                          // Combining Mark (East Asian context)
	ClassEB_EA                          // Emoji Base (East Asian context)
	ClassEX_EA                          // Exclamation (East Asian context)
	ClassGL_EA                          // Glue (East Asian context)
	ClassID_EA                          // Ideographic (East Asian context)
	ClassIN_EA                          // Inseparable (East Asian context)
	ClassNS_EA                          // Nonstarter (East Asian context)
	ClassOP_EA                          // Open Punctuation (East Asian context)
	ClassPO_EA                          // Postfix Numeric (East Asian context)
	ClassPR_EA                          // Prefix Numeric (East Asian context)

	// Quotation subclasses
	ClassQU_Pi BreakClass = iota + 125 // Quotation - Pi (initial punctuation)
	ClassQU_Pf                          // Quotation - Pf (final punctuation)
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

// applyEAWidthVariant returns the East Asian Width variant of a break class if applicable.
// Characters with Wide or Fullwidth EA width get the _EA variant.
// Note: Ambiguous characters use the base class (not _EA variant).
func applyEAWidthVariant(class BreakClass, r rune) BreakClass {
	// Check if character has East Asian Width (Wide or Fullwidth only)
	width := uax11.LookupWidth(r)
	hasEAWidth := width == uax11.Wide || width == uax11.Fullwidth

	if !hasEAWidth {
		// Check for quotation mark subtypes (Pi/Pf)
		if class == ClassQU {
			if unicode.Is(unicode.Pi, r) { // Initial punctuation
				return ClassQU_Pi
			}
			if unicode.Is(unicode.Pf, r) { // Final punctuation
				return ClassQU_Pf
			}
		}
		return class
	}

	// Return EA width variant if it exists
	switch class {
	case ClassAI:
		return ClassAI_EA
	case ClassAL:
		return ClassAL_EA
	case ClassBA:
		return ClassBA_EA
	case ClassCL:
		return ClassCL_EA
	case ClassCM:
		return ClassCM_EA
	case ClassEB:
		return ClassEB_EA
	case ClassEX:
		return ClassEX_EA
	case ClassGL:
		return ClassGL_EA
	case ClassID:
		return ClassID_EA
	case ClassIN:
		return ClassIN_EA
	case ClassNS:
		return ClassNS_EA
	case ClassOP:
		return ClassOP_EA
	case ClassPO:
		return ClassPO_EA
	case ClassPR:
		return ClassPR_EA
	default:
		return class
	}
}

// isClassOrVariant checks if a class matches a base class or its EA width variant.
// For example, isClassOrVariant(ClassAI_EA, ClassAI) returns true.
func isClassOrVariant(class, baseClass BreakClass) bool {
	if class == baseClass {
		return true
	}
	// Check EA width variants
	switch baseClass {
	case ClassAI:
		return class == ClassAI_EA
	case ClassAL:
		return class == ClassAL_EA
	case ClassBA:
		return class == ClassBA_EA
	case ClassCL:
		return class == ClassCL_EA
	case ClassCM:
		return class == ClassCM_EA
	case ClassEB:
		return class == ClassEB_EA
	case ClassEX:
		return class == ClassEX_EA
	case ClassGL:
		return class == ClassGL_EA
	case ClassID:
		return class == ClassID_EA
	case ClassIN:
		return class == ClassIN_EA
	case ClassNS:
		return class == ClassNS_EA
	case ClassOP:
		return class == ClassOP_EA
	case ClassPO:
		return class == ClassPO_EA
	case ClassPR:
		return class == ClassPR_EA
	case ClassQU:
		return class == ClassQU_Pi || class == ClassQU_Pf
	default:
		return false
	}
}

// getBreakClass returns the line breaking class for a rune.
// Uses official Unicode LineBreak.txt property data and East Asian Width.
// Reference: http://www.unicode.org/reports/tr14/#Table1
func getBreakClass(r rune) BreakClass {
	// Use official Unicode data
	if class, ok := getBreakClassFromData(r); ok {
		// Check if this character has East Asian Width
		// and if so, return the EA variant of the class
		return applyEAWidthVariant(class, r)
	}

	// Fallback for unassigned characters (should rarely be hit with complete data)
	// Mandatory breaks
	switch r {
	case '\n':
		return ClassLF
	case '\r':
		return ClassCR
	case '\u000B': // LINE TABULATION (Vertical Tab \v)
		return ClassBK
	case '\u000C': // FORM FEED (\f)
		return ClassBK
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

	// Prefix Numeric (currency symbols and similar)
	switch r {
	case '$', '£', '€', '¥', '¢', '₩', '₪', '₹', '₽', '₺', '₴', '₱', '₦', '₡', '₵':
		return ClassPR
	case '฿', '៛', '₮', '₲', '₸', '₼', '₾', '＄', '￡', '￥', '￦':
		return ClassPR
	case '+', '\u2212': // + (plus), U+2212 (minus sign)
		return ClassPR
	case '#', '\uFF03': // # (hash), ＃ (fullwidth hash)
		return ClassAI // Actually varies, but commonly used as prefix
	}

	// Postfix Numeric (percent, degree, etc.)
	switch r {
	case '%', '‰', '‱': // %, ‰ (per mille), ‱ (per ten thousand)
		return ClassPO
	case '°', '℃', '℉': // degree, celsius, fahrenheit
		return ClassPO
	case '¢', '¤': // cent sign, currency sign
		return ClassPO
	}

	// Punctuation
	switch r {
	case '(', '[', '{', '⟨', '｟':
		return ClassOP
	case ')', ']', '}', '⟩', '｠':
		return ClassCP
	case '"', '\'', '«', '»', '„', '‚', '‹', '›':
		return ClassQU
	case '!', '?', '\uFE56', '\uFE57', '\uFF01', '\uFF1F':
		// ! ? (ASCII)
		// ﹖ ﹗ (Small question/exclamation marks)
		// ！ ？ (Fullwidth)
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

	// Hangul syllables - must check before generic letter check
	// Hangul Syllables block: U+AC00-U+D7AF
	if r >= 0xAC00 && r <= 0xD7AF {
		// Simplified: treat all Hangul syllables as H2
		// (Proper implementation would distinguish H2 vs H3 based on final jamo)
		return ClassH2
	}

	// Hangul Jamo - must check before generic letter check
	// Hangul Jamo: U+1100-U+11FF
	if r >= 0x1100 && r <= 0x11FF {
		if r >= 0x1100 && r <= 0x1159 {
			return ClassJL // Leading consonants
		} else if r >= 0x1160 && r <= 0x11A7 {
			return ClassJV // Vowels
		} else {
			return ClassJT // Trailing consonants
		}
	}

	// Indic scripts (Aksara-based) - must check before generic letter check
	// These scripts use virama-based conjunct formation
	// Balinese: U+1B00-U+1B7F
	// Brahmi: U+11000-U+1107F
	// Other Indic scripts would need more ranges
	if (r >= 0x1B00 && r <= 0x1B7F) || (r >= 0x11000 && r <= 0x1107F) {
		return ClassAK
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
// Generated from Unicode LineBreakTest.html
// Total pairs: 2064

// Generated from Unicode LineBreakTest.html (17.0.0)
// Keep EA Width variants separate for maximum conformance
// Total pairs: %!d(MISSING)\n\n", len(pairs))
var pairTable = map[[2]BreakClass]BreakAction{

	{ClassAI, ClassAI}: BreakIndirect,
	{ClassAI, ClassAI_EA}: BreakIndirect,
	{ClassAI, ClassAK}: BreakDirect,
	{ClassAI, ClassAL}: BreakIndirect,
	{ClassAI, ClassAL_EA}: BreakIndirect,
	{ClassAI, ClassAP}: BreakDirect,
	{ClassAI, ClassAS}: BreakDirect,
	{ClassAI, ClassB2}: BreakDirect,
	{ClassAI, ClassBA}: BreakIndirect,
	{ClassAI, ClassBA_EA}: BreakIndirect,
	{ClassAI, ClassBB}: BreakDirect,
	{ClassAI, ClassBK}: BreakProhibited,
	{ClassAI, ClassCB}: BreakDirect,
	{ClassAI, ClassCJ}: BreakIndirect,
	{ClassAI, ClassCL}: BreakProhibited,
	{ClassAI, ClassCL_EA}: BreakProhibited,
	{ClassAI, ClassCM}: BreakIndirect,
	{ClassAI, ClassCM_EA}: BreakIndirect,
	{ClassAI, ClassCP}: BreakProhibited,
	{ClassAI, ClassCR}: BreakProhibited,
	{ClassAI, ClassEB}: BreakDirect,
	{ClassAI, ClassEB_EA}: BreakDirect,
	{ClassAI, ClassEM}: BreakDirect,
	{ClassAI, ClassEX}: BreakProhibited,
	{ClassAI, ClassEX_EA}: BreakProhibited,
	{ClassAI, ClassGL}: BreakIndirect,
	{ClassAI, ClassGL_EA}: BreakIndirect,
	{ClassAI, ClassH2}: BreakDirect,
	{ClassAI, ClassH3}: BreakDirect,
	{ClassAI, ClassHH}: BreakIndirect,
	{ClassAI, ClassHL}: BreakIndirect,
	{ClassAI, ClassHY}: BreakIndirect,
	{ClassAI, ClassID}: BreakDirect,
	{ClassAI, ClassID_EA}: BreakDirect,
	{ClassAI, ClassIN}: BreakIndirect,
	{ClassAI, ClassIN_EA}: BreakIndirect,
	{ClassAI, ClassIS}: BreakProhibited,
	{ClassAI, ClassJL}: BreakDirect,
	{ClassAI, ClassJT}: BreakDirect,
	{ClassAI, ClassJV}: BreakDirect,
	{ClassAI, ClassLF}: BreakProhibited,
	{ClassAI, ClassNL}: BreakProhibited,
	{ClassAI, ClassNS}: BreakIndirect,
	{ClassAI, ClassNS_EA}: BreakIndirect,
	{ClassAI, ClassNU}: BreakIndirect,
	{ClassAI, ClassOP}: BreakIndirect,
	{ClassAI, ClassOP_EA}: BreakDirect,
	{ClassAI, ClassPO}: BreakIndirect,
	{ClassAI, ClassPO_EA}: BreakIndirect,
	{ClassAI, ClassPR}: BreakIndirect,
	{ClassAI, ClassPR_EA}: BreakIndirect,
	{ClassAI, ClassQU}: BreakIndirect,
	{ClassAI, ClassQU_Pf}: BreakProhibited,
	{ClassAI, ClassQU_Pi}: BreakIndirect,
	{ClassAI, ClassRI}: BreakDirect,
	{ClassAI, ClassSA}: BreakIndirect,
	{ClassAI, ClassSP}: BreakProhibited,
	{ClassAI, ClassSY}: BreakProhibited,
	{ClassAI, ClassVF}: BreakDirect,
	{ClassAI, ClassVI}: BreakDirect,
	{ClassAI, ClassWJ}: BreakProhibited,
	{ClassAI, ClassXX}: BreakIndirect,
	{ClassAI, ClassZW}: BreakProhibited,
	{ClassAI, ClassZWJ}: BreakIndirect,
	{ClassAI_EA, ClassAI}: BreakIndirect,
	{ClassAI_EA, ClassAI_EA}: BreakIndirect,
	{ClassAI_EA, ClassAK}: BreakDirect,
	{ClassAI_EA, ClassAL}: BreakIndirect,
	{ClassAI_EA, ClassAL_EA}: BreakIndirect,
	{ClassAI_EA, ClassAP}: BreakDirect,
	{ClassAI_EA, ClassAS}: BreakDirect,
	{ClassAI_EA, ClassB2}: BreakDirect,
	{ClassAI_EA, ClassBA}: BreakIndirect,
	{ClassAI_EA, ClassBA_EA}: BreakIndirect,
	{ClassAI_EA, ClassBB}: BreakDirect,
	{ClassAI_EA, ClassBK}: BreakProhibited,
	{ClassAI_EA, ClassCB}: BreakDirect,
	{ClassAI_EA, ClassCJ}: BreakIndirect,
	{ClassAI_EA, ClassCL}: BreakProhibited,
	{ClassAI_EA, ClassCL_EA}: BreakProhibited,
	{ClassAI_EA, ClassCM}: BreakIndirect,
	{ClassAI_EA, ClassCM_EA}: BreakIndirect,
	{ClassAI_EA, ClassCP}: BreakProhibited,
	{ClassAI_EA, ClassCR}: BreakProhibited,
	{ClassAI_EA, ClassEB}: BreakDirect,
	{ClassAI_EA, ClassEB_EA}: BreakDirect,
	{ClassAI_EA, ClassEM}: BreakDirect,
	{ClassAI_EA, ClassEX}: BreakProhibited,
	{ClassAI_EA, ClassEX_EA}: BreakProhibited,
	{ClassAI_EA, ClassGL}: BreakIndirect,
	{ClassAI_EA, ClassGL_EA}: BreakIndirect,
	{ClassAI_EA, ClassH2}: BreakDirect,
	{ClassAI_EA, ClassH3}: BreakDirect,
	{ClassAI_EA, ClassHH}: BreakIndirect,
	{ClassAI_EA, ClassHL}: BreakIndirect,
	{ClassAI_EA, ClassHY}: BreakIndirect,
	{ClassAI_EA, ClassID}: BreakDirect,
	{ClassAI_EA, ClassID_EA}: BreakDirect,
	{ClassAI_EA, ClassIN}: BreakIndirect,
	{ClassAI_EA, ClassIN_EA}: BreakIndirect,
	{ClassAI_EA, ClassIS}: BreakProhibited,
	{ClassAI_EA, ClassJL}: BreakDirect,
	{ClassAI_EA, ClassJT}: BreakDirect,
	{ClassAI_EA, ClassJV}: BreakDirect,
	{ClassAI_EA, ClassLF}: BreakProhibited,
	{ClassAI_EA, ClassNL}: BreakProhibited,
	{ClassAI_EA, ClassNS}: BreakIndirect,
	{ClassAI_EA, ClassNS_EA}: BreakIndirect,
	{ClassAI_EA, ClassNU}: BreakIndirect,
	{ClassAI_EA, ClassOP}: BreakIndirect,
	{ClassAI_EA, ClassOP_EA}: BreakDirect,
	{ClassAI_EA, ClassPO}: BreakIndirect,
	{ClassAI_EA, ClassPO_EA}: BreakIndirect,
	{ClassAI_EA, ClassPR}: BreakIndirect,
	{ClassAI_EA, ClassPR_EA}: BreakIndirect,
	{ClassAI_EA, ClassQU}: BreakIndirect,
	{ClassAI_EA, ClassQU_Pf}: BreakProhibited,
	{ClassAI_EA, ClassQU_Pi}: BreakIndirect,
	{ClassAI_EA, ClassRI}: BreakDirect,
	{ClassAI_EA, ClassSA}: BreakIndirect,
	{ClassAI_EA, ClassSP}: BreakProhibited,
	{ClassAI_EA, ClassSY}: BreakProhibited,
	{ClassAI_EA, ClassVF}: BreakDirect,
	{ClassAI_EA, ClassVI}: BreakDirect,
	{ClassAI_EA, ClassWJ}: BreakProhibited,
	{ClassAI_EA, ClassXX}: BreakIndirect,
	{ClassAI_EA, ClassZW}: BreakProhibited,
	{ClassAI_EA, ClassZWJ}: BreakIndirect,
	{ClassAK, ClassAI}: BreakDirect,
	{ClassAK, ClassAI_EA}: BreakDirect,
	{ClassAK, ClassAK}: BreakDirect,
	{ClassAK, ClassAL}: BreakDirect,
	{ClassAK, ClassAL_EA}: BreakDirect,
	{ClassAK, ClassAP}: BreakDirect,
	{ClassAK, ClassAS}: BreakDirect,
	{ClassAK, ClassB2}: BreakDirect,
	{ClassAK, ClassBA}: BreakIndirect,
	{ClassAK, ClassBA_EA}: BreakIndirect,
	{ClassAK, ClassBB}: BreakDirect,
	{ClassAK, ClassBK}: BreakProhibited,
	{ClassAK, ClassCB}: BreakDirect,
	{ClassAK, ClassCJ}: BreakIndirect,
	{ClassAK, ClassCL}: BreakProhibited,
	{ClassAK, ClassCL_EA}: BreakProhibited,
	{ClassAK, ClassCM}: BreakIndirect,
	{ClassAK, ClassCM_EA}: BreakIndirect,
	{ClassAK, ClassCP}: BreakProhibited,
	{ClassAK, ClassCR}: BreakProhibited,
	{ClassAK, ClassEB}: BreakDirect,
	{ClassAK, ClassEB_EA}: BreakDirect,
	{ClassAK, ClassEM}: BreakDirect,
	{ClassAK, ClassEX}: BreakProhibited,
	{ClassAK, ClassEX_EA}: BreakProhibited,
	{ClassAK, ClassGL}: BreakIndirect,
	{ClassAK, ClassGL_EA}: BreakIndirect,
	{ClassAK, ClassH2}: BreakDirect,
	{ClassAK, ClassH3}: BreakDirect,
	{ClassAK, ClassHH}: BreakIndirect,
	{ClassAK, ClassHL}: BreakDirect,
	{ClassAK, ClassHY}: BreakIndirect,
	{ClassAK, ClassID}: BreakDirect,
	{ClassAK, ClassID_EA}: BreakDirect,
	{ClassAK, ClassIN}: BreakIndirect,
	{ClassAK, ClassIN_EA}: BreakIndirect,
	{ClassAK, ClassIS}: BreakProhibited,
	{ClassAK, ClassJL}: BreakDirect,
	{ClassAK, ClassJT}: BreakDirect,
	{ClassAK, ClassJV}: BreakDirect,
	{ClassAK, ClassLF}: BreakProhibited,
	{ClassAK, ClassNL}: BreakProhibited,
	{ClassAK, ClassNS}: BreakIndirect,
	{ClassAK, ClassNS_EA}: BreakIndirect,
	{ClassAK, ClassNU}: BreakDirect,
	{ClassAK, ClassOP}: BreakDirect,
	{ClassAK, ClassOP_EA}: BreakDirect,
	{ClassAK, ClassPO}: BreakDirect,
	{ClassAK, ClassPO_EA}: BreakDirect,
	{ClassAK, ClassPR}: BreakDirect,
	{ClassAK, ClassPR_EA}: BreakDirect,
	{ClassAK, ClassQU}: BreakIndirect,
	{ClassAK, ClassQU_Pf}: BreakProhibited,
	{ClassAK, ClassQU_Pi}: BreakIndirect,
	{ClassAK, ClassRI}: BreakDirect,
	{ClassAK, ClassSA}: BreakDirect,
	{ClassAK, ClassSP}: BreakProhibited,
	{ClassAK, ClassSY}: BreakProhibited,
	{ClassAK, ClassVF}: BreakIndirect,
	{ClassAK, ClassVI}: BreakIndirect,
	{ClassAK, ClassWJ}: BreakProhibited,
	{ClassAK, ClassXX}: BreakDirect,
	{ClassAK, ClassZW}: BreakProhibited,
	{ClassAK, ClassZWJ}: BreakIndirect,
	{ClassAL, ClassAI}: BreakIndirect,
	{ClassAL, ClassAI_EA}: BreakIndirect,
	{ClassAL, ClassAK}: BreakDirect,
	{ClassAL, ClassAL}: BreakIndirect,
	{ClassAL, ClassAL_EA}: BreakIndirect,
	{ClassAL, ClassAP}: BreakDirect,
	{ClassAL, ClassAS}: BreakDirect,
	{ClassAL, ClassB2}: BreakDirect,
	{ClassAL, ClassBA}: BreakIndirect,
	{ClassAL, ClassBA_EA}: BreakIndirect,
	{ClassAL, ClassBB}: BreakDirect,
	{ClassAL, ClassBK}: BreakProhibited,
	{ClassAL, ClassCB}: BreakDirect,
	{ClassAL, ClassCJ}: BreakIndirect,
	{ClassAL, ClassCL}: BreakProhibited,
	{ClassAL, ClassCL_EA}: BreakProhibited,
	{ClassAL, ClassCM}: BreakIndirect,
	{ClassAL, ClassCM_EA}: BreakIndirect,
	{ClassAL, ClassCP}: BreakProhibited,
	{ClassAL, ClassCR}: BreakProhibited,
	{ClassAL, ClassEB}: BreakDirect,
	{ClassAL, ClassEB_EA}: BreakDirect,
	{ClassAL, ClassEM}: BreakDirect,
	{ClassAL, ClassEX}: BreakProhibited,
	{ClassAL, ClassEX_EA}: BreakProhibited,
	{ClassAL, ClassGL}: BreakIndirect,
	{ClassAL, ClassGL_EA}: BreakIndirect,
	{ClassAL, ClassH2}: BreakDirect,
	{ClassAL, ClassH3}: BreakDirect,
	{ClassAL, ClassHH}: BreakIndirect,
	{ClassAL, ClassHL}: BreakIndirect,
	{ClassAL, ClassHY}: BreakIndirect,
	{ClassAL, ClassID}: BreakDirect,
	{ClassAL, ClassID_EA}: BreakDirect,
	{ClassAL, ClassIN}: BreakIndirect,
	{ClassAL, ClassIN_EA}: BreakIndirect,
	{ClassAL, ClassIS}: BreakProhibited,
	{ClassAL, ClassJL}: BreakDirect,
	{ClassAL, ClassJT}: BreakDirect,
	{ClassAL, ClassJV}: BreakDirect,
	{ClassAL, ClassLF}: BreakProhibited,
	{ClassAL, ClassNL}: BreakProhibited,
	{ClassAL, ClassNS}: BreakIndirect,
	{ClassAL, ClassNS_EA}: BreakIndirect,
	{ClassAL, ClassNU}: BreakIndirect,
	{ClassAL, ClassOP}: BreakIndirect,
	{ClassAL, ClassOP_EA}: BreakDirect,
	{ClassAL, ClassPO}: BreakIndirect,
	{ClassAL, ClassPO_EA}: BreakIndirect,
	{ClassAL, ClassPR}: BreakIndirect,
	{ClassAL, ClassPR_EA}: BreakIndirect,
	{ClassAL, ClassQU}: BreakIndirect,
	{ClassAL, ClassQU_Pf}: BreakProhibited,
	{ClassAL, ClassQU_Pi}: BreakIndirect,
	{ClassAL, ClassRI}: BreakDirect,
	{ClassAL, ClassSA}: BreakIndirect,
	{ClassAL, ClassSP}: BreakProhibited,
	{ClassAL, ClassSY}: BreakProhibited,
	{ClassAL, ClassVF}: BreakDirect,
	{ClassAL, ClassVI}: BreakDirect,
	{ClassAL, ClassWJ}: BreakProhibited,
	{ClassAL, ClassXX}: BreakIndirect,
	{ClassAL, ClassZW}: BreakProhibited,
	{ClassAL, ClassZWJ}: BreakIndirect,
	{ClassAL_EA, ClassAI}: BreakIndirect,
	{ClassAL_EA, ClassAI_EA}: BreakIndirect,
	{ClassAL_EA, ClassAK}: BreakDirect,
	{ClassAL_EA, ClassAL}: BreakIndirect,
	{ClassAL_EA, ClassAL_EA}: BreakIndirect,
	{ClassAL_EA, ClassAP}: BreakDirect,
	{ClassAL_EA, ClassAS}: BreakDirect,
	{ClassAL_EA, ClassB2}: BreakDirect,
	{ClassAL_EA, ClassBA}: BreakIndirect,
	{ClassAL_EA, ClassBA_EA}: BreakIndirect,
	{ClassAL_EA, ClassBB}: BreakDirect,
	{ClassAL_EA, ClassBK}: BreakProhibited,
	{ClassAL_EA, ClassCB}: BreakDirect,
	{ClassAL_EA, ClassCJ}: BreakIndirect,
	{ClassAL_EA, ClassCL}: BreakProhibited,
	{ClassAL_EA, ClassCL_EA}: BreakProhibited,
	{ClassAL_EA, ClassCM}: BreakIndirect,
	{ClassAL_EA, ClassCM_EA}: BreakIndirect,
	{ClassAL_EA, ClassCP}: BreakProhibited,
	{ClassAL_EA, ClassCR}: BreakProhibited,
	{ClassAL_EA, ClassEB}: BreakDirect,
	{ClassAL_EA, ClassEB_EA}: BreakDirect,
	{ClassAL_EA, ClassEM}: BreakDirect,
	{ClassAL_EA, ClassEX}: BreakProhibited,
	{ClassAL_EA, ClassEX_EA}: BreakProhibited,
	{ClassAL_EA, ClassGL}: BreakIndirect,
	{ClassAL_EA, ClassGL_EA}: BreakIndirect,
	{ClassAL_EA, ClassH2}: BreakDirect,
	{ClassAL_EA, ClassH3}: BreakDirect,
	{ClassAL_EA, ClassHH}: BreakIndirect,
	{ClassAL_EA, ClassHL}: BreakIndirect,
	{ClassAL_EA, ClassHY}: BreakIndirect,
	{ClassAL_EA, ClassID}: BreakDirect,
	{ClassAL_EA, ClassID_EA}: BreakDirect,
	{ClassAL_EA, ClassIN}: BreakIndirect,
	{ClassAL_EA, ClassIN_EA}: BreakIndirect,
	{ClassAL_EA, ClassIS}: BreakProhibited,
	{ClassAL_EA, ClassJL}: BreakDirect,
	{ClassAL_EA, ClassJT}: BreakDirect,
	{ClassAL_EA, ClassJV}: BreakDirect,
	{ClassAL_EA, ClassLF}: BreakProhibited,
	{ClassAL_EA, ClassNL}: BreakProhibited,
	{ClassAL_EA, ClassNS}: BreakIndirect,
	{ClassAL_EA, ClassNS_EA}: BreakIndirect,
	{ClassAL_EA, ClassNU}: BreakIndirect,
	{ClassAL_EA, ClassOP}: BreakIndirect,
	{ClassAL_EA, ClassOP_EA}: BreakDirect,
	{ClassAL_EA, ClassPO}: BreakIndirect,
	{ClassAL_EA, ClassPO_EA}: BreakIndirect,
	{ClassAL_EA, ClassPR}: BreakIndirect,
	{ClassAL_EA, ClassPR_EA}: BreakIndirect,
	{ClassAL_EA, ClassQU}: BreakIndirect,
	{ClassAL_EA, ClassQU_Pf}: BreakProhibited,
	{ClassAL_EA, ClassQU_Pi}: BreakIndirect,
	{ClassAL_EA, ClassRI}: BreakDirect,
	{ClassAL_EA, ClassSA}: BreakIndirect,
	{ClassAL_EA, ClassSP}: BreakProhibited,
	{ClassAL_EA, ClassSY}: BreakProhibited,
	{ClassAL_EA, ClassVF}: BreakDirect,
	{ClassAL_EA, ClassVI}: BreakDirect,
	{ClassAL_EA, ClassWJ}: BreakProhibited,
	{ClassAL_EA, ClassXX}: BreakIndirect,
	{ClassAL_EA, ClassZW}: BreakProhibited,
	{ClassAL_EA, ClassZWJ}: BreakIndirect,
	{ClassAP, ClassAI}: BreakDirect,
	{ClassAP, ClassAI_EA}: BreakDirect,
	{ClassAP, ClassAK}: BreakIndirect,
	{ClassAP, ClassAL}: BreakDirect,
	{ClassAP, ClassAL_EA}: BreakDirect,
	{ClassAP, ClassAP}: BreakDirect,
	{ClassAP, ClassAS}: BreakIndirect,
	{ClassAP, ClassB2}: BreakDirect,
	{ClassAP, ClassBA}: BreakIndirect,
	{ClassAP, ClassBA_EA}: BreakIndirect,
	{ClassAP, ClassBB}: BreakDirect,
	{ClassAP, ClassBK}: BreakProhibited,
	{ClassAP, ClassCB}: BreakDirect,
	{ClassAP, ClassCJ}: BreakIndirect,
	{ClassAP, ClassCL}: BreakProhibited,
	{ClassAP, ClassCL_EA}: BreakProhibited,
	{ClassAP, ClassCM}: BreakIndirect,
	{ClassAP, ClassCM_EA}: BreakIndirect,
	{ClassAP, ClassCP}: BreakProhibited,
	{ClassAP, ClassCR}: BreakProhibited,
	{ClassAP, ClassEB}: BreakDirect,
	{ClassAP, ClassEB_EA}: BreakDirect,
	{ClassAP, ClassEM}: BreakDirect,
	{ClassAP, ClassEX}: BreakProhibited,
	{ClassAP, ClassEX_EA}: BreakProhibited,
	{ClassAP, ClassGL}: BreakIndirect,
	{ClassAP, ClassGL_EA}: BreakIndirect,
	{ClassAP, ClassH2}: BreakDirect,
	{ClassAP, ClassH3}: BreakDirect,
	{ClassAP, ClassHH}: BreakIndirect,
	{ClassAP, ClassHL}: BreakDirect,
	{ClassAP, ClassHY}: BreakIndirect,
	{ClassAP, ClassID}: BreakDirect,
	{ClassAP, ClassID_EA}: BreakDirect,
	{ClassAP, ClassIN}: BreakIndirect,
	{ClassAP, ClassIN_EA}: BreakIndirect,
	{ClassAP, ClassIS}: BreakProhibited,
	{ClassAP, ClassJL}: BreakDirect,
	{ClassAP, ClassJT}: BreakDirect,
	{ClassAP, ClassJV}: BreakDirect,
	{ClassAP, ClassLF}: BreakProhibited,
	{ClassAP, ClassNL}: BreakProhibited,
	{ClassAP, ClassNS}: BreakIndirect,
	{ClassAP, ClassNS_EA}: BreakIndirect,
	{ClassAP, ClassNU}: BreakDirect,
	{ClassAP, ClassOP}: BreakDirect,
	{ClassAP, ClassOP_EA}: BreakDirect,
	{ClassAP, ClassPO}: BreakDirect,
	{ClassAP, ClassPO_EA}: BreakDirect,
	{ClassAP, ClassPR}: BreakDirect,
	{ClassAP, ClassPR_EA}: BreakDirect,
	{ClassAP, ClassQU}: BreakIndirect,
	{ClassAP, ClassQU_Pf}: BreakProhibited,
	{ClassAP, ClassQU_Pi}: BreakIndirect,
	{ClassAP, ClassRI}: BreakDirect,
	{ClassAP, ClassSA}: BreakDirect,
	{ClassAP, ClassSP}: BreakProhibited,
	{ClassAP, ClassSY}: BreakProhibited,
	{ClassAP, ClassVF}: BreakDirect,
	{ClassAP, ClassVI}: BreakDirect,
	{ClassAP, ClassWJ}: BreakProhibited,
	{ClassAP, ClassXX}: BreakDirect,
	{ClassAP, ClassZW}: BreakProhibited,
	{ClassAP, ClassZWJ}: BreakIndirect,
	{ClassAS, ClassAI}: BreakDirect,
	{ClassAS, ClassAI_EA}: BreakDirect,
	{ClassAS, ClassAK}: BreakDirect,
	{ClassAS, ClassAL}: BreakDirect,
	{ClassAS, ClassAL_EA}: BreakDirect,
	{ClassAS, ClassAP}: BreakDirect,
	{ClassAS, ClassAS}: BreakDirect,
	{ClassAS, ClassB2}: BreakDirect,
	{ClassAS, ClassBA}: BreakIndirect,
	{ClassAS, ClassBA_EA}: BreakIndirect,
	{ClassAS, ClassBB}: BreakDirect,
	{ClassAS, ClassBK}: BreakProhibited,
	{ClassAS, ClassCB}: BreakDirect,
	{ClassAS, ClassCJ}: BreakIndirect,
	{ClassAS, ClassCL}: BreakProhibited,
	{ClassAS, ClassCL_EA}: BreakProhibited,
	{ClassAS, ClassCM}: BreakIndirect,
	{ClassAS, ClassCM_EA}: BreakIndirect,
	{ClassAS, ClassCP}: BreakProhibited,
	{ClassAS, ClassCR}: BreakProhibited,
	{ClassAS, ClassEB}: BreakDirect,
	{ClassAS, ClassEB_EA}: BreakDirect,
	{ClassAS, ClassEM}: BreakDirect,
	{ClassAS, ClassEX}: BreakProhibited,
	{ClassAS, ClassEX_EA}: BreakProhibited,
	{ClassAS, ClassGL}: BreakIndirect,
	{ClassAS, ClassGL_EA}: BreakIndirect,
	{ClassAS, ClassH2}: BreakDirect,
	{ClassAS, ClassH3}: BreakDirect,
	{ClassAS, ClassHH}: BreakIndirect,
	{ClassAS, ClassHL}: BreakDirect,
	{ClassAS, ClassHY}: BreakIndirect,
	{ClassAS, ClassID}: BreakDirect,
	{ClassAS, ClassID_EA}: BreakDirect,
	{ClassAS, ClassIN}: BreakIndirect,
	{ClassAS, ClassIN_EA}: BreakIndirect,
	{ClassAS, ClassIS}: BreakProhibited,
	{ClassAS, ClassJL}: BreakDirect,
	{ClassAS, ClassJT}: BreakDirect,
	{ClassAS, ClassJV}: BreakDirect,
	{ClassAS, ClassLF}: BreakProhibited,
	{ClassAS, ClassNL}: BreakProhibited,
	{ClassAS, ClassNS}: BreakIndirect,
	{ClassAS, ClassNS_EA}: BreakIndirect,
	{ClassAS, ClassNU}: BreakDirect,
	{ClassAS, ClassOP}: BreakDirect,
	{ClassAS, ClassOP_EA}: BreakDirect,
	{ClassAS, ClassPO}: BreakDirect,
	{ClassAS, ClassPO_EA}: BreakDirect,
	{ClassAS, ClassPR}: BreakDirect,
	{ClassAS, ClassPR_EA}: BreakDirect,
	{ClassAS, ClassQU}: BreakIndirect,
	{ClassAS, ClassQU_Pf}: BreakProhibited,
	{ClassAS, ClassQU_Pi}: BreakIndirect,
	{ClassAS, ClassRI}: BreakDirect,
	{ClassAS, ClassSA}: BreakDirect,
	{ClassAS, ClassSP}: BreakProhibited,
	{ClassAS, ClassSY}: BreakProhibited,
	{ClassAS, ClassVF}: BreakIndirect,
	{ClassAS, ClassVI}: BreakIndirect,
	{ClassAS, ClassWJ}: BreakProhibited,
	{ClassAS, ClassXX}: BreakDirect,
	{ClassAS, ClassZW}: BreakProhibited,
	{ClassAS, ClassZWJ}: BreakIndirect,
	{ClassB2, ClassAI}: BreakDirect,
	{ClassB2, ClassAI_EA}: BreakDirect,
	{ClassB2, ClassAK}: BreakDirect,
	{ClassB2, ClassAL}: BreakDirect,
	{ClassB2, ClassAL_EA}: BreakDirect,
	{ClassB2, ClassAP}: BreakDirect,
	{ClassB2, ClassAS}: BreakDirect,
	{ClassB2, ClassB2}: BreakProhibited,
	{ClassB2, ClassBA}: BreakIndirect,
	{ClassB2, ClassBA_EA}: BreakIndirect,
	{ClassB2, ClassBB}: BreakDirect,
	{ClassB2, ClassBK}: BreakProhibited,
	{ClassB2, ClassCB}: BreakDirect,
	{ClassB2, ClassCJ}: BreakIndirect,
	{ClassB2, ClassCL}: BreakProhibited,
	{ClassB2, ClassCL_EA}: BreakProhibited,
	{ClassB2, ClassCM}: BreakIndirect,
	{ClassB2, ClassCM_EA}: BreakIndirect,
	{ClassB2, ClassCP}: BreakProhibited,
	{ClassB2, ClassCR}: BreakProhibited,
	{ClassB2, ClassEB}: BreakDirect,
	{ClassB2, ClassEB_EA}: BreakDirect,
	{ClassB2, ClassEM}: BreakDirect,
	{ClassB2, ClassEX}: BreakProhibited,
	{ClassB2, ClassEX_EA}: BreakProhibited,
	{ClassB2, ClassGL}: BreakIndirect,
	{ClassB2, ClassGL_EA}: BreakIndirect,
	{ClassB2, ClassH2}: BreakDirect,
	{ClassB2, ClassH3}: BreakDirect,
	{ClassB2, ClassHH}: BreakIndirect,
	{ClassB2, ClassHL}: BreakDirect,
	{ClassB2, ClassHY}: BreakIndirect,
	{ClassB2, ClassID}: BreakDirect,
	{ClassB2, ClassID_EA}: BreakDirect,
	{ClassB2, ClassIN}: BreakIndirect,
	{ClassB2, ClassIN_EA}: BreakIndirect,
	{ClassB2, ClassIS}: BreakProhibited,
	{ClassB2, ClassJL}: BreakDirect,
	{ClassB2, ClassJT}: BreakDirect,
	{ClassB2, ClassJV}: BreakDirect,
	{ClassB2, ClassLF}: BreakProhibited,
	{ClassB2, ClassNL}: BreakProhibited,
	{ClassB2, ClassNS}: BreakIndirect,
	{ClassB2, ClassNS_EA}: BreakIndirect,
	{ClassB2, ClassNU}: BreakDirect,
	{ClassB2, ClassOP}: BreakDirect,
	{ClassB2, ClassOP_EA}: BreakDirect,
	{ClassB2, ClassPO}: BreakDirect,
	{ClassB2, ClassPO_EA}: BreakDirect,
	{ClassB2, ClassPR}: BreakDirect,
	{ClassB2, ClassPR_EA}: BreakDirect,
	{ClassB2, ClassQU}: BreakIndirect,
	{ClassB2, ClassQU_Pf}: BreakProhibited,
	{ClassB2, ClassQU_Pi}: BreakIndirect,
	{ClassB2, ClassRI}: BreakDirect,
	{ClassB2, ClassSA}: BreakDirect,
	{ClassB2, ClassSP}: BreakProhibited,
	{ClassB2, ClassSY}: BreakProhibited,
	{ClassB2, ClassVF}: BreakDirect,
	{ClassB2, ClassVI}: BreakDirect,
	{ClassB2, ClassWJ}: BreakProhibited,
	{ClassB2, ClassXX}: BreakDirect,
	{ClassB2, ClassZW}: BreakProhibited,
	{ClassB2, ClassZWJ}: BreakIndirect,
	{ClassBA_EA, ClassAI}: BreakDirect,
	{ClassBA_EA, ClassAI_EA}: BreakDirect,
	{ClassBA_EA, ClassAK}: BreakDirect,
	{ClassBA_EA, ClassAL}: BreakDirect,
	{ClassBA_EA, ClassAL_EA}: BreakDirect,
	{ClassBA_EA, ClassAP}: BreakDirect,
	{ClassBA_EA, ClassAS}: BreakDirect,
	{ClassBA_EA, ClassB2}: BreakDirect,
	{ClassBA_EA, ClassBA}: BreakIndirect,
	{ClassBA_EA, ClassBA_EA}: BreakIndirect,
	{ClassBA_EA, ClassBB}: BreakDirect,
	{ClassBA_EA, ClassBK}: BreakProhibited,
	{ClassBA_EA, ClassCB}: BreakDirect,
	{ClassBA_EA, ClassCJ}: BreakIndirect,
	{ClassBA_EA, ClassCL}: BreakProhibited,
	{ClassBA_EA, ClassCL_EA}: BreakProhibited,
	{ClassBA_EA, ClassCM}: BreakIndirect,
	{ClassBA_EA, ClassCM_EA}: BreakIndirect,
	{ClassBA_EA, ClassCP}: BreakProhibited,
	{ClassBA_EA, ClassCR}: BreakProhibited,
	{ClassBA_EA, ClassEB}: BreakDirect,
	{ClassBA_EA, ClassEB_EA}: BreakDirect,
	{ClassBA_EA, ClassEM}: BreakDirect,
	{ClassBA_EA, ClassEX}: BreakProhibited,
	{ClassBA_EA, ClassEX_EA}: BreakProhibited,
	{ClassBA_EA, ClassGL}: BreakDirect,
	{ClassBA_EA, ClassGL_EA}: BreakDirect,
	{ClassBA_EA, ClassH2}: BreakDirect,
	{ClassBA_EA, ClassH3}: BreakDirect,
	{ClassBA_EA, ClassHH}: BreakIndirect,
	{ClassBA_EA, ClassHL}: BreakDirect,
	{ClassBA_EA, ClassHY}: BreakIndirect,
	{ClassBA_EA, ClassID}: BreakDirect,
	{ClassBA_EA, ClassID_EA}: BreakDirect,
	{ClassBA_EA, ClassIN}: BreakIndirect,
	{ClassBA_EA, ClassIN_EA}: BreakIndirect,
	{ClassBA_EA, ClassIS}: BreakProhibited,
	{ClassBA_EA, ClassJL}: BreakDirect,
	{ClassBA_EA, ClassJT}: BreakDirect,
	{ClassBA_EA, ClassJV}: BreakDirect,
	{ClassBA_EA, ClassLF}: BreakProhibited,
	{ClassBA_EA, ClassNL}: BreakProhibited,
	{ClassBA_EA, ClassNS}: BreakIndirect,
	{ClassBA_EA, ClassNS_EA}: BreakIndirect,
	{ClassBA_EA, ClassNU}: BreakDirect,
	{ClassBA_EA, ClassOP}: BreakDirect,
	{ClassBA_EA, ClassOP_EA}: BreakDirect,
	{ClassBA_EA, ClassPO}: BreakDirect,
	{ClassBA_EA, ClassPO_EA}: BreakDirect,
	{ClassBA_EA, ClassPR}: BreakDirect,
	{ClassBA_EA, ClassPR_EA}: BreakDirect,
	{ClassBA_EA, ClassQU}: BreakIndirect,
	{ClassBA_EA, ClassQU_Pf}: BreakProhibited,
	{ClassBA_EA, ClassQU_Pi}: BreakIndirect,
	{ClassBA_EA, ClassRI}: BreakDirect,
	{ClassBA_EA, ClassSA}: BreakDirect,
	{ClassBA_EA, ClassSP}: BreakProhibited,
	{ClassBA_EA, ClassSY}: BreakProhibited,
	{ClassBA_EA, ClassVF}: BreakDirect,
	{ClassBA_EA, ClassVI}: BreakDirect,
	{ClassBA_EA, ClassWJ}: BreakProhibited,
	{ClassBA_EA, ClassXX}: BreakDirect,
	{ClassBA_EA, ClassZW}: BreakProhibited,
	{ClassBA_EA, ClassZWJ}: BreakIndirect,
	{ClassBA, ClassAI_EA}: BreakDirect,
	{ClassBA, ClassAI}: BreakDirect,
	{ClassBA, ClassAK}: BreakDirect,
	{ClassBA, ClassAL_EA}: BreakDirect,
	{ClassBA, ClassAL}: BreakDirect,
	{ClassBA, ClassAP}: BreakDirect,
	{ClassBA, ClassAS}: BreakDirect,
	{ClassBA, ClassB2}: BreakDirect,
	{ClassBA, ClassBA}: BreakIndirect,
	{ClassBA, ClassBA_EA}: BreakIndirect,
	{ClassBA, ClassBB}: BreakDirect,
	{ClassBA, ClassBK}: BreakProhibited,
	{ClassBA, ClassCB}: BreakDirect,
	{ClassBA, ClassCJ}: BreakIndirect,
	{ClassBA, ClassCL_EA}: BreakProhibited,
	{ClassBA, ClassCL}: BreakProhibited,
	{ClassBA, ClassCM_EA}: BreakIndirect,
	{ClassBA, ClassCM}: BreakIndirect,
	{ClassBA, ClassCP}: BreakProhibited,
	{ClassBA, ClassCR}: BreakProhibited,
	{ClassBA, ClassEB_EA}: BreakDirect,
	{ClassBA, ClassEB}: BreakDirect,
	{ClassBA, ClassEM}: BreakDirect,
	{ClassBA, ClassEX_EA}: BreakProhibited,
	{ClassBA, ClassEX}: BreakProhibited,
	{ClassBA, ClassGL_EA}: BreakDirect,
	{ClassBA, ClassGL}: BreakDirect,
	{ClassBA, ClassH2}: BreakDirect,
	{ClassBA, ClassH3}: BreakDirect,
	{ClassBA, ClassHH}: BreakIndirect,
	{ClassBA, ClassHL}: BreakDirect,
	{ClassBA, ClassHY}: BreakIndirect,
	{ClassBA, ClassID_EA}: BreakDirect,
	{ClassBA, ClassID}: BreakDirect,
	{ClassBA, ClassIN_EA}: BreakIndirect,
	{ClassBA, ClassIN}: BreakIndirect,
	{ClassBA, ClassIS}: BreakProhibited,
	{ClassBA, ClassJL}: BreakDirect,
	{ClassBA, ClassJT}: BreakDirect,
	{ClassBA, ClassJV}: BreakDirect,
	{ClassBA, ClassLF}: BreakProhibited,
	{ClassBA, ClassNL}: BreakProhibited,
	{ClassBA, ClassNS_EA}: BreakIndirect,
	{ClassBA, ClassNS}: BreakIndirect,
	{ClassBA, ClassNU}: BreakDirect,
	{ClassBA, ClassOP_EA}: BreakDirect,
	{ClassBA, ClassOP}: BreakDirect,
	{ClassBA, ClassPO_EA}: BreakDirect,
	{ClassBA, ClassPO}: BreakDirect,
	{ClassBA, ClassPR_EA}: BreakDirect,
	{ClassBA, ClassPR}: BreakDirect,
	{ClassBA, ClassQU_Pf}: BreakProhibited,
	{ClassBA, ClassQU_Pi}: BreakIndirect,
	{ClassBA, ClassQU}: BreakIndirect,
	{ClassBA, ClassRI}: BreakDirect,
	{ClassBA, ClassSA}: BreakDirect,
	{ClassBA, ClassSP}: BreakProhibited,
	{ClassBA, ClassSY}: BreakProhibited,
	{ClassBA, ClassVF}: BreakDirect,
	{ClassBA, ClassVI}: BreakDirect,
	{ClassBA, ClassWJ}: BreakProhibited,
	{ClassBA, ClassXX}: BreakDirect,
	{ClassBA, ClassZW}: BreakProhibited,
	{ClassBA, ClassZWJ}: BreakIndirect,
	{ClassBB, ClassAI}: BreakIndirect,
	{ClassBB, ClassAI_EA}: BreakIndirect,
	{ClassBB, ClassAK}: BreakIndirect,
	{ClassBB, ClassAL}: BreakIndirect,
	{ClassBB, ClassAL_EA}: BreakIndirect,
	{ClassBB, ClassAP}: BreakIndirect,
	{ClassBB, ClassAS}: BreakIndirect,
	{ClassBB, ClassB2}: BreakIndirect,
	{ClassBB, ClassBA}: BreakIndirect,
	{ClassBB, ClassBA_EA}: BreakIndirect,
	{ClassBB, ClassBB}: BreakIndirect,
	{ClassBB, ClassBK}: BreakProhibited,
	{ClassBB, ClassCB}: BreakDirect,
	{ClassBB, ClassCJ}: BreakIndirect,
	{ClassBB, ClassCL}: BreakProhibited,
	{ClassBB, ClassCL_EA}: BreakProhibited,
	{ClassBB, ClassCM}: BreakIndirect,
	{ClassBB, ClassCM_EA}: BreakIndirect,
	{ClassBB, ClassCP}: BreakProhibited,
	{ClassBB, ClassCR}: BreakProhibited,
	{ClassBB, ClassEB}: BreakIndirect,
	{ClassBB, ClassEB_EA}: BreakIndirect,
	{ClassBB, ClassEM}: BreakIndirect,
	{ClassBB, ClassEX}: BreakProhibited,
	{ClassBB, ClassEX_EA}: BreakProhibited,
	{ClassBB, ClassGL}: BreakIndirect,
	{ClassBB, ClassGL_EA}: BreakIndirect,
	{ClassBB, ClassH2}: BreakIndirect,
	{ClassBB, ClassH3}: BreakIndirect,
	{ClassBB, ClassHH}: BreakIndirect,
	{ClassBB, ClassHL}: BreakIndirect,
	{ClassBB, ClassHY}: BreakIndirect,
	{ClassBB, ClassID}: BreakIndirect,
	{ClassBB, ClassID_EA}: BreakIndirect,
	{ClassBB, ClassIN}: BreakIndirect,
	{ClassBB, ClassIN_EA}: BreakIndirect,
	{ClassBB, ClassIS}: BreakProhibited,
	{ClassBB, ClassJL}: BreakIndirect,
	{ClassBB, ClassJT}: BreakIndirect,
	{ClassBB, ClassJV}: BreakIndirect,
	{ClassBB, ClassLF}: BreakProhibited,
	{ClassBB, ClassNL}: BreakProhibited,
	{ClassBB, ClassNS}: BreakIndirect,
	{ClassBB, ClassNS_EA}: BreakIndirect,
	{ClassBB, ClassNU}: BreakIndirect,
	{ClassBB, ClassOP}: BreakIndirect,
	{ClassBB, ClassOP_EA}: BreakIndirect,
	{ClassBB, ClassPO}: BreakIndirect,
	{ClassBB, ClassPO_EA}: BreakIndirect,
	{ClassBB, ClassPR}: BreakIndirect,
	{ClassBB, ClassPR_EA}: BreakIndirect,
	{ClassBB, ClassQU}: BreakIndirect,
	{ClassBB, ClassQU_Pf}: BreakProhibited,
	{ClassBB, ClassQU_Pi}: BreakIndirect,
	{ClassBB, ClassRI}: BreakIndirect,
	{ClassBB, ClassSA}: BreakIndirect,
	{ClassBB, ClassSP}: BreakProhibited,
	{ClassBB, ClassSY}: BreakProhibited,
	{ClassBB, ClassVF}: BreakIndirect,
	{ClassBB, ClassVI}: BreakIndirect,
	{ClassBB, ClassWJ}: BreakProhibited,
	{ClassBB, ClassXX}: BreakIndirect,
	{ClassBB, ClassZW}: BreakProhibited,
	{ClassBB, ClassZWJ}: BreakIndirect,
	{ClassCB, ClassAI}: BreakDirect,
	{ClassCB, ClassAI_EA}: BreakDirect,
	{ClassCB, ClassAK}: BreakDirect,
	{ClassCB, ClassAL}: BreakDirect,
	{ClassCB, ClassAL_EA}: BreakDirect,
	{ClassCB, ClassAP}: BreakDirect,
	{ClassCB, ClassAS}: BreakDirect,
	{ClassCB, ClassB2}: BreakDirect,
	{ClassCB, ClassBA}: BreakDirect,
	{ClassCB, ClassBA_EA}: BreakDirect,
	{ClassCB, ClassBB}: BreakDirect,
	{ClassCB, ClassBK}: BreakProhibited,
	{ClassCB, ClassCB}: BreakDirect,
	{ClassCB, ClassCJ}: BreakDirect,
	{ClassCB, ClassCL}: BreakProhibited,
	{ClassCB, ClassCL_EA}: BreakProhibited,
	{ClassCB, ClassCM}: BreakIndirect,
	{ClassCB, ClassCM_EA}: BreakIndirect,
	{ClassCB, ClassCP}: BreakProhibited,
	{ClassCB, ClassCR}: BreakProhibited,
	{ClassCB, ClassEB}: BreakDirect,
	{ClassCB, ClassEB_EA}: BreakDirect,
	{ClassCB, ClassEM}: BreakDirect,
	{ClassCB, ClassEX}: BreakProhibited,
	{ClassCB, ClassEX_EA}: BreakProhibited,
	{ClassCB, ClassGL}: BreakIndirect,
	{ClassCB, ClassGL_EA}: BreakIndirect,
	{ClassCB, ClassH2}: BreakDirect,
	{ClassCB, ClassH3}: BreakDirect,
	{ClassCB, ClassHH}: BreakDirect,
	{ClassCB, ClassHL}: BreakDirect,
	{ClassCB, ClassHY}: BreakDirect,
	{ClassCB, ClassID}: BreakDirect,
	{ClassCB, ClassID_EA}: BreakDirect,
	{ClassCB, ClassIN}: BreakDirect,
	{ClassCB, ClassIN_EA}: BreakDirect,
	{ClassCB, ClassIS}: BreakProhibited,
	{ClassCB, ClassJL}: BreakDirect,
	{ClassCB, ClassJT}: BreakDirect,
	{ClassCB, ClassJV}: BreakDirect,
	{ClassCB, ClassLF}: BreakProhibited,
	{ClassCB, ClassNL}: BreakProhibited,
	{ClassCB, ClassNS}: BreakDirect,
	{ClassCB, ClassNS_EA}: BreakDirect,
	{ClassCB, ClassNU}: BreakDirect,
	{ClassCB, ClassOP}: BreakDirect,
	{ClassCB, ClassOP_EA}: BreakDirect,
	{ClassCB, ClassPO}: BreakDirect,
	{ClassCB, ClassPO_EA}: BreakDirect,
	{ClassCB, ClassPR}: BreakDirect,
	{ClassCB, ClassPR_EA}: BreakDirect,
	{ClassCB, ClassQU}: BreakIndirect,
	{ClassCB, ClassQU_Pf}: BreakProhibited,
	{ClassCB, ClassQU_Pi}: BreakIndirect,
	{ClassCB, ClassRI}: BreakDirect,
	{ClassCB, ClassSA}: BreakDirect,
	{ClassCB, ClassSP}: BreakProhibited,
	{ClassCB, ClassSY}: BreakProhibited,
	{ClassCB, ClassVF}: BreakDirect,
	{ClassCB, ClassVI}: BreakDirect,
	{ClassCB, ClassWJ}: BreakProhibited,
	{ClassCB, ClassXX}: BreakDirect,
	{ClassCB, ClassZW}: BreakProhibited,
	{ClassCB, ClassZWJ}: BreakIndirect,
	{ClassCJ, ClassAI}: BreakDirect,
	{ClassCJ, ClassAI_EA}: BreakDirect,
	{ClassCJ, ClassAK}: BreakDirect,
	{ClassCJ, ClassAL}: BreakDirect,
	{ClassCJ, ClassAL_EA}: BreakDirect,
	{ClassCJ, ClassAP}: BreakDirect,
	{ClassCJ, ClassAS}: BreakDirect,
	{ClassCJ, ClassB2}: BreakDirect,
	{ClassCJ, ClassBA}: BreakIndirect,
	{ClassCJ, ClassBA_EA}: BreakIndirect,
	{ClassCJ, ClassBB}: BreakDirect,
	{ClassCJ, ClassBK}: BreakProhibited,
	{ClassCJ, ClassCB}: BreakDirect,
	{ClassCJ, ClassCJ}: BreakIndirect,
	{ClassCJ, ClassCL}: BreakProhibited,
	{ClassCJ, ClassCL_EA}: BreakProhibited,
	{ClassCJ, ClassCM}: BreakIndirect,
	{ClassCJ, ClassCM_EA}: BreakIndirect,
	{ClassCJ, ClassCP}: BreakProhibited,
	{ClassCJ, ClassCR}: BreakProhibited,
	{ClassCJ, ClassEB}: BreakDirect,
	{ClassCJ, ClassEB_EA}: BreakDirect,
	{ClassCJ, ClassEM}: BreakDirect,
	{ClassCJ, ClassEX}: BreakProhibited,
	{ClassCJ, ClassEX_EA}: BreakProhibited,
	{ClassCJ, ClassGL}: BreakIndirect,
	{ClassCJ, ClassGL_EA}: BreakIndirect,
	{ClassCJ, ClassH2}: BreakDirect,
	{ClassCJ, ClassH3}: BreakDirect,
	{ClassCJ, ClassHH}: BreakIndirect,
	{ClassCJ, ClassHL}: BreakDirect,
	{ClassCJ, ClassHY}: BreakIndirect,
	{ClassCJ, ClassID}: BreakDirect,
	{ClassCJ, ClassID_EA}: BreakDirect,
	{ClassCJ, ClassIN}: BreakIndirect,
	{ClassCJ, ClassIN_EA}: BreakIndirect,
	{ClassCJ, ClassIS}: BreakProhibited,
	{ClassCJ, ClassJL}: BreakDirect,
	{ClassCJ, ClassJT}: BreakDirect,
	{ClassCJ, ClassJV}: BreakDirect,
	{ClassCJ, ClassLF}: BreakProhibited,
	{ClassCJ, ClassNL}: BreakProhibited,
	{ClassCJ, ClassNS}: BreakIndirect,
	{ClassCJ, ClassNS_EA}: BreakIndirect,
	{ClassCJ, ClassNU}: BreakDirect,
	{ClassCJ, ClassOP}: BreakDirect,
	{ClassCJ, ClassOP_EA}: BreakDirect,
	{ClassCJ, ClassPO}: BreakDirect,
	{ClassCJ, ClassPO_EA}: BreakDirect,
	{ClassCJ, ClassPR}: BreakDirect,
	{ClassCJ, ClassPR_EA}: BreakDirect,
	{ClassCJ, ClassQU}: BreakIndirect,
	{ClassCJ, ClassQU_Pf}: BreakProhibited,
	{ClassCJ, ClassQU_Pi}: BreakIndirect,
	{ClassCJ, ClassRI}: BreakDirect,
	{ClassCJ, ClassSA}: BreakDirect,
	{ClassCJ, ClassSP}: BreakProhibited,
	{ClassCJ, ClassSY}: BreakProhibited,
	{ClassCJ, ClassVF}: BreakDirect,
	{ClassCJ, ClassVI}: BreakDirect,
	{ClassCJ, ClassWJ}: BreakProhibited,
	{ClassCJ, ClassXX}: BreakDirect,
	{ClassCJ, ClassZW}: BreakProhibited,
	{ClassCJ, ClassZWJ}: BreakIndirect,
	{ClassCL, ClassAI}: BreakDirect,
	{ClassCL, ClassAI_EA}: BreakDirect,
	{ClassCL, ClassAK}: BreakDirect,
	{ClassCL, ClassAL}: BreakDirect,
	{ClassCL, ClassAL_EA}: BreakDirect,
	{ClassCL, ClassAP}: BreakDirect,
	{ClassCL, ClassAS}: BreakDirect,
	{ClassCL, ClassB2}: BreakDirect,
	{ClassCL, ClassBA}: BreakIndirect,
	{ClassCL, ClassBA_EA}: BreakIndirect,
	{ClassCL, ClassBB}: BreakDirect,
	{ClassCL, ClassBK}: BreakProhibited,
	{ClassCL, ClassCB}: BreakDirect,
	{ClassCL, ClassCJ}: BreakProhibited,
	{ClassCL, ClassCL}: BreakProhibited,
	{ClassCL, ClassCL_EA}: BreakProhibited,
	{ClassCL, ClassCM}: BreakIndirect,
	{ClassCL, ClassCM_EA}: BreakIndirect,
	{ClassCL, ClassCP}: BreakProhibited,
	{ClassCL, ClassCR}: BreakProhibited,
	{ClassCL, ClassEB}: BreakDirect,
	{ClassCL, ClassEB_EA}: BreakDirect,
	{ClassCL, ClassEM}: BreakDirect,
	{ClassCL, ClassEX}: BreakProhibited,
	{ClassCL, ClassEX_EA}: BreakProhibited,
	{ClassCL, ClassGL}: BreakIndirect,
	{ClassCL, ClassGL_EA}: BreakIndirect,
	{ClassCL, ClassH2}: BreakDirect,
	{ClassCL, ClassH3}: BreakDirect,
	{ClassCL, ClassHH}: BreakIndirect,
	{ClassCL, ClassHL}: BreakDirect,
	{ClassCL, ClassHY}: BreakIndirect,
	{ClassCL, ClassID}: BreakDirect,
	{ClassCL, ClassID_EA}: BreakDirect,
	{ClassCL, ClassIN}: BreakIndirect,
	{ClassCL, ClassIN_EA}: BreakIndirect,
	{ClassCL, ClassIS}: BreakProhibited,
	{ClassCL, ClassJL}: BreakDirect,
	{ClassCL, ClassJT}: BreakDirect,
	{ClassCL, ClassJV}: BreakDirect,
	{ClassCL, ClassLF}: BreakProhibited,
	{ClassCL, ClassNL}: BreakProhibited,
	{ClassCL, ClassNS}: BreakProhibited,
	{ClassCL, ClassNS_EA}: BreakProhibited,
	{ClassCL, ClassNU}: BreakDirect,
	{ClassCL, ClassOP}: BreakDirect,
	{ClassCL, ClassOP_EA}: BreakDirect,
	{ClassCL, ClassPO}: BreakDirect,
	{ClassCL, ClassPO_EA}: BreakDirect,
	{ClassCL, ClassPR}: BreakDirect,
	{ClassCL, ClassPR_EA}: BreakDirect,
	{ClassCL, ClassQU}: BreakIndirect,
	{ClassCL, ClassQU_Pf}: BreakProhibited,
	{ClassCL, ClassQU_Pi}: BreakIndirect,
	{ClassCL, ClassRI}: BreakDirect,
	{ClassCL, ClassSA}: BreakDirect,
	{ClassCL, ClassSP}: BreakProhibited,
	{ClassCL, ClassSY}: BreakProhibited,
	{ClassCL, ClassVF}: BreakDirect,
	{ClassCL, ClassVI}: BreakDirect,
	{ClassCL, ClassWJ}: BreakProhibited,
	{ClassCL, ClassXX}: BreakDirect,
	{ClassCL, ClassZW}: BreakProhibited,
	{ClassCL, ClassZWJ}: BreakIndirect,
	{ClassCL_EA, ClassAI}: BreakDirect,
	{ClassCL_EA, ClassAI_EA}: BreakDirect,
	{ClassCL_EA, ClassAK}: BreakDirect,
	{ClassCL_EA, ClassAL}: BreakDirect,
	{ClassCL_EA, ClassAL_EA}: BreakDirect,
	{ClassCL_EA, ClassAP}: BreakDirect,
	{ClassCL_EA, ClassAS}: BreakDirect,
	{ClassCL_EA, ClassB2}: BreakDirect,
	{ClassCL_EA, ClassBA}: BreakIndirect,
	{ClassCL_EA, ClassBA_EA}: BreakIndirect,
	{ClassCL_EA, ClassBB}: BreakDirect,
	{ClassCL_EA, ClassBK}: BreakProhibited,
	{ClassCL_EA, ClassCB}: BreakDirect,
	{ClassCL_EA, ClassCJ}: BreakProhibited,
	{ClassCL_EA, ClassCL}: BreakProhibited,
	{ClassCL_EA, ClassCL_EA}: BreakProhibited,
	{ClassCL_EA, ClassCM}: BreakIndirect,
	{ClassCL_EA, ClassCM_EA}: BreakIndirect,
	{ClassCL_EA, ClassCP}: BreakProhibited,
	{ClassCL_EA, ClassCR}: BreakProhibited,
	{ClassCL_EA, ClassEB}: BreakDirect,
	{ClassCL_EA, ClassEB_EA}: BreakDirect,
	{ClassCL_EA, ClassEM}: BreakDirect,
	{ClassCL_EA, ClassEX}: BreakProhibited,
	{ClassCL_EA, ClassEX_EA}: BreakProhibited,
	{ClassCL_EA, ClassGL}: BreakIndirect,
	{ClassCL_EA, ClassGL_EA}: BreakIndirect,
	{ClassCL_EA, ClassH2}: BreakDirect,
	{ClassCL_EA, ClassH3}: BreakDirect,
	{ClassCL_EA, ClassHH}: BreakIndirect,
	{ClassCL_EA, ClassHL}: BreakDirect,
	{ClassCL_EA, ClassHY}: BreakIndirect,
	{ClassCL_EA, ClassID}: BreakDirect,
	{ClassCL_EA, ClassID_EA}: BreakDirect,
	{ClassCL_EA, ClassIN}: BreakIndirect,
	{ClassCL_EA, ClassIN_EA}: BreakIndirect,
	{ClassCL_EA, ClassIS}: BreakProhibited,
	{ClassCL_EA, ClassJL}: BreakDirect,
	{ClassCL_EA, ClassJT}: BreakDirect,
	{ClassCL_EA, ClassJV}: BreakDirect,
	{ClassCL_EA, ClassLF}: BreakProhibited,
	{ClassCL_EA, ClassNL}: BreakProhibited,
	{ClassCL_EA, ClassNS}: BreakProhibited,
	{ClassCL_EA, ClassNS_EA}: BreakProhibited,
	{ClassCL_EA, ClassNU}: BreakDirect,
	{ClassCL_EA, ClassOP}: BreakDirect,
	{ClassCL_EA, ClassOP_EA}: BreakDirect,
	{ClassCL_EA, ClassPO}: BreakDirect,
	{ClassCL_EA, ClassPO_EA}: BreakDirect,
	{ClassCL_EA, ClassPR}: BreakDirect,
	{ClassCL_EA, ClassPR_EA}: BreakDirect,
	{ClassCL_EA, ClassQU}: BreakIndirect,
	{ClassCL_EA, ClassQU_Pf}: BreakProhibited,
	{ClassCL_EA, ClassQU_Pi}: BreakIndirect,
	{ClassCL_EA, ClassRI}: BreakDirect,
	{ClassCL_EA, ClassSA}: BreakDirect,
	{ClassCL_EA, ClassSP}: BreakProhibited,
	{ClassCL_EA, ClassSY}: BreakProhibited,
	{ClassCL_EA, ClassVF}: BreakDirect,
	{ClassCL_EA, ClassVI}: BreakDirect,
	{ClassCL_EA, ClassWJ}: BreakProhibited,
	{ClassCL_EA, ClassXX}: BreakDirect,
	{ClassCL_EA, ClassZW}: BreakProhibited,
	{ClassCL_EA, ClassZWJ}: BreakIndirect,
	{ClassCM_EA, ClassAI}: BreakIndirect,
	{ClassCM_EA, ClassAI_EA}: BreakIndirect,
	{ClassCM_EA, ClassAK}: BreakDirect,
	{ClassCM_EA, ClassAL}: BreakIndirect,
	{ClassCM_EA, ClassAL_EA}: BreakIndirect,
	{ClassCM_EA, ClassAP}: BreakDirect,
	{ClassCM_EA, ClassAS}: BreakDirect,
	{ClassCM_EA, ClassB2}: BreakDirect,
	{ClassCM_EA, ClassBA}: BreakIndirect,
	{ClassCM_EA, ClassBA_EA}: BreakIndirect,
	{ClassCM_EA, ClassBB}: BreakDirect,
	{ClassCM_EA, ClassBK}: BreakProhibited,
	{ClassCM_EA, ClassCB}: BreakDirect,
	{ClassCM_EA, ClassCJ}: BreakIndirect,
	{ClassCM_EA, ClassCL}: BreakProhibited,
	{ClassCM_EA, ClassCL_EA}: BreakProhibited,
	{ClassCM_EA, ClassCM}: BreakIndirect,
	{ClassCM_EA, ClassCM_EA}: BreakIndirect,
	{ClassCM_EA, ClassCP}: BreakProhibited,
	{ClassCM_EA, ClassCR}: BreakProhibited,
	{ClassCM_EA, ClassEB}: BreakDirect,
	{ClassCM_EA, ClassEB_EA}: BreakDirect,
	{ClassCM_EA, ClassEM}: BreakDirect,
	{ClassCM_EA, ClassEX}: BreakProhibited,
	{ClassCM_EA, ClassEX_EA}: BreakProhibited,
	{ClassCM_EA, ClassGL}: BreakIndirect,
	{ClassCM_EA, ClassGL_EA}: BreakIndirect,
	{ClassCM_EA, ClassH2}: BreakDirect,
	{ClassCM_EA, ClassH3}: BreakDirect,
	{ClassCM_EA, ClassHH}: BreakIndirect,
	{ClassCM_EA, ClassHL}: BreakIndirect,
	{ClassCM_EA, ClassHY}: BreakIndirect,
	{ClassCM_EA, ClassID}: BreakDirect,
	{ClassCM_EA, ClassID_EA}: BreakDirect,
	{ClassCM_EA, ClassIN}: BreakIndirect,
	{ClassCM_EA, ClassIN_EA}: BreakIndirect,
	{ClassCM_EA, ClassIS}: BreakProhibited,
	{ClassCM_EA, ClassJL}: BreakDirect,
	{ClassCM_EA, ClassJT}: BreakDirect,
	{ClassCM_EA, ClassJV}: BreakDirect,
	{ClassCM_EA, ClassLF}: BreakProhibited,
	{ClassCM_EA, ClassNL}: BreakProhibited,
	{ClassCM_EA, ClassNS}: BreakIndirect,
	{ClassCM_EA, ClassNS_EA}: BreakIndirect,
	{ClassCM_EA, ClassNU}: BreakIndirect,
	{ClassCM_EA, ClassOP}: BreakIndirect,
	{ClassCM_EA, ClassOP_EA}: BreakDirect,
	{ClassCM_EA, ClassPO}: BreakIndirect,
	{ClassCM_EA, ClassPO_EA}: BreakIndirect,
	{ClassCM_EA, ClassPR}: BreakIndirect,
	{ClassCM_EA, ClassPR_EA}: BreakIndirect,
	{ClassCM_EA, ClassQU}: BreakIndirect,
	{ClassCM_EA, ClassQU_Pf}: BreakProhibited,
	{ClassCM_EA, ClassQU_Pi}: BreakIndirect,
	{ClassCM_EA, ClassRI}: BreakDirect,
	{ClassCM_EA, ClassSA}: BreakIndirect,
	{ClassCM_EA, ClassSP}: BreakProhibited,
	{ClassCM_EA, ClassSY}: BreakProhibited,
	{ClassCM_EA, ClassVF}: BreakDirect,
	{ClassCM_EA, ClassVI}: BreakDirect,
	{ClassCM_EA, ClassWJ}: BreakProhibited,
	{ClassCM_EA, ClassXX}: BreakIndirect,
	{ClassCM_EA, ClassZW}: BreakProhibited,
	{ClassCM_EA, ClassZWJ}: BreakIndirect,
	{ClassCP, ClassAI}: BreakIndirect,
	{ClassCP, ClassAI_EA}: BreakIndirect,
	{ClassCP, ClassAK}: BreakDirect,
	{ClassCP, ClassAL}: BreakIndirect,
	{ClassCP, ClassAL_EA}: BreakIndirect,
	{ClassCP, ClassAP}: BreakDirect,
	{ClassCP, ClassAS}: BreakDirect,
	{ClassCP, ClassB2}: BreakDirect,
	{ClassCP, ClassBA}: BreakIndirect,
	{ClassCP, ClassBA_EA}: BreakIndirect,
	{ClassCP, ClassBB}: BreakDirect,
	{ClassCP, ClassBK}: BreakProhibited,
	{ClassCP, ClassCB}: BreakDirect,
	{ClassCP, ClassCJ}: BreakProhibited,
	{ClassCP, ClassCL}: BreakProhibited,
	{ClassCP, ClassCL_EA}: BreakProhibited,
	{ClassCP, ClassCM}: BreakIndirect,
	{ClassCP, ClassCM_EA}: BreakIndirect,
	{ClassCP, ClassCP}: BreakProhibited,
	{ClassCP, ClassCR}: BreakProhibited,
	{ClassCP, ClassEB}: BreakDirect,
	{ClassCP, ClassEB_EA}: BreakDirect,
	{ClassCP, ClassEM}: BreakDirect,
	{ClassCP, ClassEX}: BreakProhibited,
	{ClassCP, ClassEX_EA}: BreakProhibited,
	{ClassCP, ClassGL}: BreakIndirect,
	{ClassCP, ClassGL_EA}: BreakIndirect,
	{ClassCP, ClassH2}: BreakDirect,
	{ClassCP, ClassH3}: BreakDirect,
	{ClassCP, ClassHH}: BreakIndirect,
	{ClassCP, ClassHL}: BreakIndirect,
	{ClassCP, ClassHY}: BreakIndirect,
	{ClassCP, ClassID}: BreakDirect,
	{ClassCP, ClassID_EA}: BreakDirect,
	{ClassCP, ClassIN}: BreakIndirect,
	{ClassCP, ClassIN_EA}: BreakIndirect,
	{ClassCP, ClassIS}: BreakProhibited,
	{ClassCP, ClassJL}: BreakDirect,
	{ClassCP, ClassJT}: BreakDirect,
	{ClassCP, ClassJV}: BreakDirect,
	{ClassCP, ClassLF}: BreakProhibited,
	{ClassCP, ClassNL}: BreakProhibited,
	{ClassCP, ClassNS}: BreakProhibited,
	{ClassCP, ClassNS_EA}: BreakProhibited,
	{ClassCP, ClassNU}: BreakIndirect,
	{ClassCP, ClassOP}: BreakDirect,
	{ClassCP, ClassOP_EA}: BreakDirect,
	{ClassCP, ClassPO}: BreakDirect,
	{ClassCP, ClassPO_EA}: BreakDirect,
	{ClassCP, ClassPR}: BreakDirect,
	{ClassCP, ClassPR_EA}: BreakDirect,
	{ClassCP, ClassQU}: BreakIndirect,
	{ClassCP, ClassQU_Pf}: BreakProhibited,
	{ClassCP, ClassQU_Pi}: BreakIndirect,
	{ClassCP, ClassRI}: BreakDirect,
	{ClassCP, ClassSA}: BreakIndirect,
	{ClassCP, ClassSP}: BreakProhibited,
	{ClassCP, ClassSY}: BreakProhibited,
	{ClassCP, ClassVF}: BreakDirect,
	{ClassCP, ClassVI}: BreakDirect,
	{ClassCP, ClassWJ}: BreakProhibited,
	{ClassCP, ClassXX}: BreakIndirect,
	{ClassCP, ClassZW}: BreakProhibited,
	{ClassCP, ClassZWJ}: BreakIndirect,
	{ClassEB, ClassAI}: BreakDirect,
	{ClassEB, ClassAI_EA}: BreakDirect,
	{ClassEB, ClassAK}: BreakDirect,
	{ClassEB, ClassAL}: BreakDirect,
	{ClassEB, ClassAL_EA}: BreakDirect,
	{ClassEB, ClassAP}: BreakDirect,
	{ClassEB, ClassAS}: BreakDirect,
	{ClassEB, ClassB2}: BreakDirect,
	{ClassEB, ClassBA}: BreakIndirect,
	{ClassEB, ClassBA_EA}: BreakIndirect,
	{ClassEB, ClassBB}: BreakDirect,
	{ClassEB, ClassBK}: BreakProhibited,
	{ClassEB, ClassCB}: BreakDirect,
	{ClassEB, ClassCJ}: BreakIndirect,
	{ClassEB, ClassCL}: BreakProhibited,
	{ClassEB, ClassCL_EA}: BreakProhibited,
	{ClassEB, ClassCM}: BreakIndirect,
	{ClassEB, ClassCM_EA}: BreakIndirect,
	{ClassEB, ClassCP}: BreakProhibited,
	{ClassEB, ClassCR}: BreakProhibited,
	{ClassEB, ClassEB}: BreakDirect,
	{ClassEB, ClassEB_EA}: BreakDirect,
	{ClassEB, ClassEM}: BreakIndirect,
	{ClassEB, ClassEX}: BreakProhibited,
	{ClassEB, ClassEX_EA}: BreakProhibited,
	{ClassEB, ClassGL}: BreakIndirect,
	{ClassEB, ClassGL_EA}: BreakIndirect,
	{ClassEB, ClassH2}: BreakDirect,
	{ClassEB, ClassH3}: BreakDirect,
	{ClassEB, ClassHH}: BreakIndirect,
	{ClassEB, ClassHL}: BreakDirect,
	{ClassEB, ClassHY}: BreakIndirect,
	{ClassEB, ClassID}: BreakDirect,
	{ClassEB, ClassID_EA}: BreakDirect,
	{ClassEB, ClassIN}: BreakIndirect,
	{ClassEB, ClassIN_EA}: BreakIndirect,
	{ClassEB, ClassIS}: BreakProhibited,
	{ClassEB, ClassJL}: BreakDirect,
	{ClassEB, ClassJT}: BreakDirect,
	{ClassEB, ClassJV}: BreakDirect,
	{ClassEB, ClassLF}: BreakProhibited,
	{ClassEB, ClassNL}: BreakProhibited,
	{ClassEB, ClassNS}: BreakIndirect,
	{ClassEB, ClassNS_EA}: BreakIndirect,
	{ClassEB, ClassNU}: BreakDirect,
	{ClassEB, ClassOP}: BreakDirect,
	{ClassEB, ClassOP_EA}: BreakDirect,
	{ClassEB, ClassPO}: BreakIndirect,
	{ClassEB, ClassPO_EA}: BreakIndirect,
	{ClassEB, ClassPR}: BreakDirect,
	{ClassEB, ClassPR_EA}: BreakDirect,
	{ClassEB, ClassQU}: BreakIndirect,
	{ClassEB, ClassQU_Pf}: BreakProhibited,
	{ClassEB, ClassQU_Pi}: BreakIndirect,
	{ClassEB, ClassRI}: BreakDirect,
	{ClassEB, ClassSA}: BreakDirect,
	{ClassEB, ClassSP}: BreakProhibited,
	{ClassEB, ClassSY}: BreakProhibited,
	{ClassEB, ClassVF}: BreakDirect,
	{ClassEB, ClassVI}: BreakDirect,
	{ClassEB, ClassWJ}: BreakProhibited,
	{ClassEB, ClassXX}: BreakDirect,
	{ClassEB, ClassZW}: BreakProhibited,
	{ClassEB, ClassZWJ}: BreakIndirect,
	{ClassEB_EA, ClassAI}: BreakDirect,
	{ClassEB_EA, ClassAI_EA}: BreakDirect,
	{ClassEB_EA, ClassAK}: BreakDirect,
	{ClassEB_EA, ClassAL}: BreakDirect,
	{ClassEB_EA, ClassAL_EA}: BreakDirect,
	{ClassEB_EA, ClassAP}: BreakDirect,
	{ClassEB_EA, ClassAS}: BreakDirect,
	{ClassEB_EA, ClassB2}: BreakDirect,
	{ClassEB_EA, ClassBA}: BreakIndirect,
	{ClassEB_EA, ClassBA_EA}: BreakIndirect,
	{ClassEB_EA, ClassBB}: BreakDirect,
	{ClassEB_EA, ClassBK}: BreakProhibited,
	{ClassEB_EA, ClassCB}: BreakDirect,
	{ClassEB_EA, ClassCJ}: BreakIndirect,
	{ClassEB_EA, ClassCL}: BreakProhibited,
	{ClassEB_EA, ClassCL_EA}: BreakProhibited,
	{ClassEB_EA, ClassCM}: BreakIndirect,
	{ClassEB_EA, ClassCM_EA}: BreakIndirect,
	{ClassEB_EA, ClassCP}: BreakProhibited,
	{ClassEB_EA, ClassCR}: BreakProhibited,
	{ClassEB_EA, ClassEB}: BreakDirect,
	{ClassEB_EA, ClassEB_EA}: BreakDirect,
	{ClassEB_EA, ClassEM}: BreakIndirect,
	{ClassEB_EA, ClassEX}: BreakProhibited,
	{ClassEB_EA, ClassEX_EA}: BreakProhibited,
	{ClassEB_EA, ClassGL}: BreakIndirect,
	{ClassEB_EA, ClassGL_EA}: BreakIndirect,
	{ClassEB_EA, ClassH2}: BreakDirect,
	{ClassEB_EA, ClassH3}: BreakDirect,
	{ClassEB_EA, ClassHH}: BreakIndirect,
	{ClassEB_EA, ClassHL}: BreakDirect,
	{ClassEB_EA, ClassHY}: BreakIndirect,
	{ClassEB_EA, ClassID}: BreakDirect,
	{ClassEB_EA, ClassID_EA}: BreakDirect,
	{ClassEB_EA, ClassIN}: BreakIndirect,
	{ClassEB_EA, ClassIN_EA}: BreakIndirect,
	{ClassEB_EA, ClassIS}: BreakProhibited,
	{ClassEB_EA, ClassJL}: BreakDirect,
	{ClassEB_EA, ClassJT}: BreakDirect,
	{ClassEB_EA, ClassJV}: BreakDirect,
	{ClassEB_EA, ClassLF}: BreakProhibited,
	{ClassEB_EA, ClassNL}: BreakProhibited,
	{ClassEB_EA, ClassNS}: BreakIndirect,
	{ClassEB_EA, ClassNS_EA}: BreakIndirect,
	{ClassEB_EA, ClassNU}: BreakDirect,
	{ClassEB_EA, ClassOP}: BreakDirect,
	{ClassEB_EA, ClassOP_EA}: BreakDirect,
	{ClassEB_EA, ClassPO}: BreakIndirect,
	{ClassEB_EA, ClassPO_EA}: BreakIndirect,
	{ClassEB_EA, ClassPR}: BreakDirect,
	{ClassEB_EA, ClassPR_EA}: BreakDirect,
	{ClassEB_EA, ClassQU}: BreakIndirect,
	{ClassEB_EA, ClassQU_Pf}: BreakProhibited,
	{ClassEB_EA, ClassQU_Pi}: BreakIndirect,
	{ClassEB_EA, ClassRI}: BreakDirect,
	{ClassEB_EA, ClassSA}: BreakDirect,
	{ClassEB_EA, ClassSP}: BreakProhibited,
	{ClassEB_EA, ClassSY}: BreakProhibited,
	{ClassEB_EA, ClassVF}: BreakDirect,
	{ClassEB_EA, ClassVI}: BreakDirect,
	{ClassEB_EA, ClassWJ}: BreakProhibited,
	{ClassEB_EA, ClassXX}: BreakDirect,
	{ClassEB_EA, ClassZW}: BreakProhibited,
	{ClassEB_EA, ClassZWJ}: BreakIndirect,
	{ClassEM, ClassAI}: BreakDirect,
	{ClassEM, ClassAI_EA}: BreakDirect,
	{ClassEM, ClassAK}: BreakDirect,
	{ClassEM, ClassAL}: BreakDirect,
	{ClassEM, ClassAL_EA}: BreakDirect,
	{ClassEM, ClassAP}: BreakDirect,
	{ClassEM, ClassAS}: BreakDirect,
	{ClassEM, ClassB2}: BreakDirect,
	{ClassEM, ClassBA}: BreakIndirect,
	{ClassEM, ClassBA_EA}: BreakIndirect,
	{ClassEM, ClassBB}: BreakDirect,
	{ClassEM, ClassBK}: BreakProhibited,
	{ClassEM, ClassCB}: BreakDirect,
	{ClassEM, ClassCJ}: BreakIndirect,
	{ClassEM, ClassCL}: BreakProhibited,
	{ClassEM, ClassCL_EA}: BreakProhibited,
	{ClassEM, ClassCM}: BreakIndirect,
	{ClassEM, ClassCM_EA}: BreakIndirect,
	{ClassEM, ClassCP}: BreakProhibited,
	{ClassEM, ClassCR}: BreakProhibited,
	{ClassEM, ClassEB}: BreakDirect,
	{ClassEM, ClassEB_EA}: BreakDirect,
	{ClassEM, ClassEM}: BreakDirect,
	{ClassEM, ClassEX}: BreakProhibited,
	{ClassEM, ClassEX_EA}: BreakProhibited,
	{ClassEM, ClassGL}: BreakIndirect,
	{ClassEM, ClassGL_EA}: BreakIndirect,
	{ClassEM, ClassH2}: BreakDirect,
	{ClassEM, ClassH3}: BreakDirect,
	{ClassEM, ClassHH}: BreakIndirect,
	{ClassEM, ClassHL}: BreakDirect,
	{ClassEM, ClassHY}: BreakIndirect,
	{ClassEM, ClassID}: BreakDirect,
	{ClassEM, ClassID_EA}: BreakDirect,
	{ClassEM, ClassIN}: BreakIndirect,
	{ClassEM, ClassIN_EA}: BreakIndirect,
	{ClassEM, ClassIS}: BreakProhibited,
	{ClassEM, ClassJL}: BreakDirect,
	{ClassEM, ClassJT}: BreakDirect,
	{ClassEM, ClassJV}: BreakDirect,
	{ClassEM, ClassLF}: BreakProhibited,
	{ClassEM, ClassNL}: BreakProhibited,
	{ClassEM, ClassNS}: BreakIndirect,
	{ClassEM, ClassNS_EA}: BreakIndirect,
	{ClassEM, ClassNU}: BreakDirect,
	{ClassEM, ClassOP}: BreakDirect,
	{ClassEM, ClassOP_EA}: BreakDirect,
	{ClassEM, ClassPO}: BreakIndirect,
	{ClassEM, ClassPO_EA}: BreakIndirect,
	{ClassEM, ClassPR}: BreakDirect,
	{ClassEM, ClassPR_EA}: BreakDirect,
	{ClassEM, ClassQU}: BreakIndirect,
	{ClassEM, ClassQU_Pf}: BreakProhibited,
	{ClassEM, ClassQU_Pi}: BreakIndirect,
	{ClassEM, ClassRI}: BreakDirect,
	{ClassEM, ClassSA}: BreakDirect,
	{ClassEM, ClassSP}: BreakProhibited,
	{ClassEM, ClassSY}: BreakProhibited,
	{ClassEM, ClassVF}: BreakDirect,
	{ClassEM, ClassVI}: BreakDirect,
	{ClassEM, ClassWJ}: BreakProhibited,
	{ClassEM, ClassXX}: BreakDirect,
	{ClassEM, ClassZW}: BreakProhibited,
	{ClassEM, ClassZWJ}: BreakIndirect,
	{ClassEX, ClassAI}: BreakDirect,
	{ClassEX, ClassAI_EA}: BreakDirect,
	{ClassEX, ClassAK}: BreakDirect,
	{ClassEX, ClassAL}: BreakDirect,
	{ClassEX, ClassAL_EA}: BreakDirect,
	{ClassEX, ClassAP}: BreakDirect,
	{ClassEX, ClassAS}: BreakDirect,
	{ClassEX, ClassB2}: BreakDirect,
	{ClassEX, ClassBA}: BreakIndirect,
	{ClassEX, ClassBA_EA}: BreakIndirect,
	{ClassEX, ClassBB}: BreakDirect,
	{ClassEX, ClassBK}: BreakProhibited,
	{ClassEX, ClassCB}: BreakDirect,
	{ClassEX, ClassCJ}: BreakIndirect,
	{ClassEX, ClassCL}: BreakProhibited,
	{ClassEX, ClassCL_EA}: BreakProhibited,
	{ClassEX, ClassCM}: BreakIndirect,
	{ClassEX, ClassCM_EA}: BreakIndirect,
	{ClassEX, ClassCP}: BreakProhibited,
	{ClassEX, ClassCR}: BreakProhibited,
	{ClassEX, ClassEB}: BreakDirect,
	{ClassEX, ClassEB_EA}: BreakDirect,
	{ClassEX, ClassEM}: BreakDirect,
	{ClassEX, ClassEX}: BreakProhibited,
	{ClassEX, ClassEX_EA}: BreakProhibited,
	{ClassEX, ClassGL}: BreakIndirect,
	{ClassEX, ClassGL_EA}: BreakIndirect,
	{ClassEX, ClassH2}: BreakDirect,
	{ClassEX, ClassH3}: BreakDirect,
	{ClassEX, ClassHH}: BreakIndirect,
	{ClassEX, ClassHL}: BreakDirect,
	{ClassEX, ClassHY}: BreakIndirect,
	{ClassEX, ClassID}: BreakDirect,
	{ClassEX, ClassID_EA}: BreakDirect,
	{ClassEX, ClassIN}: BreakIndirect,
	{ClassEX, ClassIN_EA}: BreakIndirect,
	{ClassEX, ClassIS}: BreakProhibited,
	{ClassEX, ClassJL}: BreakDirect,
	{ClassEX, ClassJT}: BreakDirect,
	{ClassEX, ClassJV}: BreakDirect,
	{ClassEX, ClassLF}: BreakProhibited,
	{ClassEX, ClassNL}: BreakProhibited,
	{ClassEX, ClassNS}: BreakIndirect,
	{ClassEX, ClassNS_EA}: BreakIndirect,
	{ClassEX, ClassNU}: BreakDirect,
	{ClassEX, ClassOP}: BreakDirect,
	{ClassEX, ClassOP_EA}: BreakDirect,
	{ClassEX, ClassPO}: BreakDirect,
	{ClassEX, ClassPO_EA}: BreakDirect,
	{ClassEX, ClassPR}: BreakDirect,
	{ClassEX, ClassPR_EA}: BreakDirect,
	{ClassEX, ClassQU}: BreakIndirect,
	{ClassEX, ClassQU_Pf}: BreakProhibited,
	{ClassEX, ClassQU_Pi}: BreakIndirect,
	{ClassEX, ClassRI}: BreakDirect,
	{ClassEX, ClassSA}: BreakDirect,
	{ClassEX, ClassSP}: BreakProhibited,
	{ClassEX, ClassSY}: BreakProhibited,
	{ClassEX, ClassVF}: BreakDirect,
	{ClassEX, ClassVI}: BreakDirect,
	{ClassEX, ClassWJ}: BreakProhibited,
	{ClassEX, ClassXX}: BreakDirect,
	{ClassEX, ClassZW}: BreakProhibited,
	{ClassEX, ClassZWJ}: BreakIndirect,
	{ClassEX_EA, ClassAI}: BreakDirect,
	{ClassEX_EA, ClassAI_EA}: BreakDirect,
	{ClassEX_EA, ClassAK}: BreakDirect,
	{ClassEX_EA, ClassAL}: BreakDirect,
	{ClassEX_EA, ClassAL_EA}: BreakDirect,
	{ClassEX_EA, ClassAP}: BreakDirect,
	{ClassEX_EA, ClassAS}: BreakDirect,
	{ClassEX_EA, ClassB2}: BreakDirect,
	{ClassEX_EA, ClassBA}: BreakIndirect,
	{ClassEX_EA, ClassBA_EA}: BreakIndirect,
	{ClassEX_EA, ClassBB}: BreakDirect,
	{ClassEX_EA, ClassBK}: BreakProhibited,
	{ClassEX_EA, ClassCB}: BreakDirect,
	{ClassEX_EA, ClassCJ}: BreakIndirect,
	{ClassEX_EA, ClassCL}: BreakProhibited,
	{ClassEX_EA, ClassCL_EA}: BreakProhibited,
	{ClassEX_EA, ClassCM}: BreakIndirect,
	{ClassEX_EA, ClassCM_EA}: BreakIndirect,
	{ClassEX_EA, ClassCP}: BreakProhibited,
	{ClassEX_EA, ClassCR}: BreakProhibited,
	{ClassEX_EA, ClassEB}: BreakDirect,
	{ClassEX_EA, ClassEB_EA}: BreakDirect,
	{ClassEX_EA, ClassEM}: BreakDirect,
	{ClassEX_EA, ClassEX}: BreakProhibited,
	{ClassEX_EA, ClassEX_EA}: BreakProhibited,
	{ClassEX_EA, ClassGL}: BreakIndirect,
	{ClassEX_EA, ClassGL_EA}: BreakIndirect,
	{ClassEX_EA, ClassH2}: BreakDirect,
	{ClassEX_EA, ClassH3}: BreakDirect,
	{ClassEX_EA, ClassHH}: BreakIndirect,
	{ClassEX_EA, ClassHL}: BreakDirect,
	{ClassEX_EA, ClassHY}: BreakIndirect,
	{ClassEX_EA, ClassID}: BreakDirect,
	{ClassEX_EA, ClassID_EA}: BreakDirect,
	{ClassEX_EA, ClassIN}: BreakIndirect,
	{ClassEX_EA, ClassIN_EA}: BreakIndirect,
	{ClassEX_EA, ClassIS}: BreakProhibited,
	{ClassEX_EA, ClassJL}: BreakDirect,
	{ClassEX_EA, ClassJT}: BreakDirect,
	{ClassEX_EA, ClassJV}: BreakDirect,
	{ClassEX_EA, ClassLF}: BreakProhibited,
	{ClassEX_EA, ClassNL}: BreakProhibited,
	{ClassEX_EA, ClassNS}: BreakIndirect,
	{ClassEX_EA, ClassNS_EA}: BreakIndirect,
	{ClassEX_EA, ClassNU}: BreakDirect,
	{ClassEX_EA, ClassOP}: BreakDirect,
	{ClassEX_EA, ClassOP_EA}: BreakDirect,
	{ClassEX_EA, ClassPO}: BreakDirect,
	{ClassEX_EA, ClassPO_EA}: BreakDirect,
	{ClassEX_EA, ClassPR}: BreakDirect,
	{ClassEX_EA, ClassPR_EA}: BreakDirect,
	{ClassEX_EA, ClassQU}: BreakIndirect,
	{ClassEX_EA, ClassQU_Pf}: BreakProhibited,
	{ClassEX_EA, ClassQU_Pi}: BreakIndirect,
	{ClassEX_EA, ClassRI}: BreakDirect,
	{ClassEX_EA, ClassSA}: BreakDirect,
	{ClassEX_EA, ClassSP}: BreakProhibited,
	{ClassEX_EA, ClassSY}: BreakProhibited,
	{ClassEX_EA, ClassVF}: BreakDirect,
	{ClassEX_EA, ClassVI}: BreakDirect,
	{ClassEX_EA, ClassWJ}: BreakProhibited,
	{ClassEX_EA, ClassXX}: BreakDirect,
	{ClassEX_EA, ClassZW}: BreakProhibited,
	{ClassEX_EA, ClassZWJ}: BreakIndirect,
	{ClassGL, ClassAI}: BreakIndirect,
	{ClassGL, ClassAI_EA}: BreakIndirect,
	{ClassGL, ClassAK}: BreakIndirect,
	{ClassGL, ClassAL}: BreakIndirect,
	{ClassGL, ClassAL_EA}: BreakIndirect,
	{ClassGL, ClassAP}: BreakIndirect,
	{ClassGL, ClassAS}: BreakIndirect,
	{ClassGL, ClassB2}: BreakIndirect,
	{ClassGL, ClassBA}: BreakIndirect,
	{ClassGL, ClassBA_EA}: BreakIndirect,
	{ClassGL, ClassBB}: BreakIndirect,
	{ClassGL, ClassBK}: BreakProhibited,
	{ClassGL, ClassCB}: BreakIndirect,
	{ClassGL, ClassCJ}: BreakIndirect,
	{ClassGL, ClassCL}: BreakProhibited,
	{ClassGL, ClassCL_EA}: BreakProhibited,
	{ClassGL, ClassCM}: BreakIndirect,
	{ClassGL, ClassCM_EA}: BreakIndirect,
	{ClassGL, ClassCP}: BreakProhibited,
	{ClassGL, ClassCR}: BreakProhibited,
	{ClassGL, ClassEB}: BreakIndirect,
	{ClassGL, ClassEB_EA}: BreakIndirect,
	{ClassGL, ClassEM}: BreakIndirect,
	{ClassGL, ClassEX}: BreakProhibited,
	{ClassGL, ClassEX_EA}: BreakProhibited,
	{ClassGL, ClassGL}: BreakIndirect,
	{ClassGL, ClassGL_EA}: BreakIndirect,
	{ClassGL, ClassH2}: BreakIndirect,
	{ClassGL, ClassH3}: BreakIndirect,
	{ClassGL, ClassHH}: BreakIndirect,
	{ClassGL, ClassHL}: BreakIndirect,
	{ClassGL, ClassHY}: BreakIndirect,
	{ClassGL, ClassID}: BreakIndirect,
	{ClassGL, ClassID_EA}: BreakIndirect,
	{ClassGL, ClassIN}: BreakIndirect,
	{ClassGL, ClassIN_EA}: BreakIndirect,
	{ClassGL, ClassIS}: BreakProhibited,
	{ClassGL, ClassJL}: BreakIndirect,
	{ClassGL, ClassJT}: BreakIndirect,
	{ClassGL, ClassJV}: BreakIndirect,
	{ClassGL, ClassLF}: BreakProhibited,
	{ClassGL, ClassNL}: BreakProhibited,
	{ClassGL, ClassNS}: BreakIndirect,
	{ClassGL, ClassNS_EA}: BreakIndirect,
	{ClassGL, ClassNU}: BreakIndirect,
	{ClassGL, ClassOP}: BreakIndirect,
	{ClassGL, ClassOP_EA}: BreakIndirect,
	{ClassGL, ClassPO}: BreakIndirect,
	{ClassGL, ClassPO_EA}: BreakIndirect,
	{ClassGL, ClassPR}: BreakIndirect,
	{ClassGL, ClassPR_EA}: BreakIndirect,
	{ClassGL, ClassQU}: BreakIndirect,
	{ClassGL, ClassQU_Pf}: BreakProhibited,
	{ClassGL, ClassQU_Pi}: BreakIndirect,
	{ClassGL, ClassRI}: BreakIndirect,
	{ClassGL, ClassSA}: BreakIndirect,
	{ClassGL, ClassSP}: BreakProhibited,
	{ClassGL, ClassSY}: BreakProhibited,
	{ClassGL, ClassVF}: BreakIndirect,
	{ClassGL, ClassVI}: BreakIndirect,
	{ClassGL, ClassWJ}: BreakProhibited,
	{ClassGL, ClassXX}: BreakIndirect,
	{ClassGL, ClassZW}: BreakProhibited,
	{ClassGL, ClassZWJ}: BreakIndirect,
	{ClassGL_EA, ClassAI}: BreakIndirect,
	{ClassGL_EA, ClassAI_EA}: BreakIndirect,
	{ClassGL_EA, ClassAK}: BreakIndirect,
	{ClassGL_EA, ClassAL}: BreakIndirect,
	{ClassGL_EA, ClassAL_EA}: BreakIndirect,
	{ClassGL_EA, ClassAP}: BreakIndirect,
	{ClassGL_EA, ClassAS}: BreakIndirect,
	{ClassGL_EA, ClassB2}: BreakIndirect,
	{ClassGL_EA, ClassBA}: BreakIndirect,
	{ClassGL_EA, ClassBA_EA}: BreakIndirect,
	{ClassGL_EA, ClassBB}: BreakIndirect,
	{ClassGL_EA, ClassBK}: BreakProhibited,
	{ClassGL_EA, ClassCB}: BreakIndirect,
	{ClassGL_EA, ClassCJ}: BreakIndirect,
	{ClassGL_EA, ClassCL}: BreakProhibited,
	{ClassGL_EA, ClassCL_EA}: BreakProhibited,
	{ClassGL_EA, ClassCM}: BreakIndirect,
	{ClassGL_EA, ClassCM_EA}: BreakIndirect,
	{ClassGL_EA, ClassCP}: BreakProhibited,
	{ClassGL_EA, ClassCR}: BreakProhibited,
	{ClassGL_EA, ClassEB}: BreakIndirect,
	{ClassGL_EA, ClassEB_EA}: BreakIndirect,
	{ClassGL_EA, ClassEM}: BreakIndirect,
	{ClassGL_EA, ClassEX}: BreakProhibited,
	{ClassGL_EA, ClassEX_EA}: BreakProhibited,
	{ClassGL_EA, ClassGL}: BreakIndirect,
	{ClassGL_EA, ClassGL_EA}: BreakIndirect,
	{ClassGL_EA, ClassH2}: BreakIndirect,
	{ClassGL_EA, ClassH3}: BreakIndirect,
	{ClassGL_EA, ClassHH}: BreakIndirect,
	{ClassGL_EA, ClassHL}: BreakIndirect,
	{ClassGL_EA, ClassHY}: BreakIndirect,
	{ClassGL_EA, ClassID}: BreakIndirect,
	{ClassGL_EA, ClassID_EA}: BreakIndirect,
	{ClassGL_EA, ClassIN}: BreakIndirect,
	{ClassGL_EA, ClassIN_EA}: BreakIndirect,
	{ClassGL_EA, ClassIS}: BreakProhibited,
	{ClassGL_EA, ClassJL}: BreakIndirect,
	{ClassGL_EA, ClassJT}: BreakIndirect,
	{ClassGL_EA, ClassJV}: BreakIndirect,
	{ClassGL_EA, ClassLF}: BreakProhibited,
	{ClassGL_EA, ClassNL}: BreakProhibited,
	{ClassGL_EA, ClassNS}: BreakIndirect,
	{ClassGL_EA, ClassNS_EA}: BreakIndirect,
	{ClassGL_EA, ClassNU}: BreakIndirect,
	{ClassGL_EA, ClassOP}: BreakIndirect,
	{ClassGL_EA, ClassOP_EA}: BreakIndirect,
	{ClassGL_EA, ClassPO}: BreakIndirect,
	{ClassGL_EA, ClassPO_EA}: BreakIndirect,
	{ClassGL_EA, ClassPR}: BreakIndirect,
	{ClassGL_EA, ClassPR_EA}: BreakIndirect,
	{ClassGL_EA, ClassQU}: BreakIndirect,
	{ClassGL_EA, ClassQU_Pf}: BreakProhibited,
	{ClassGL_EA, ClassQU_Pi}: BreakIndirect,
	{ClassGL_EA, ClassRI}: BreakIndirect,
	{ClassGL_EA, ClassSA}: BreakIndirect,
	{ClassGL_EA, ClassSP}: BreakProhibited,
	{ClassGL_EA, ClassSY}: BreakProhibited,
	{ClassGL_EA, ClassVF}: BreakIndirect,
	{ClassGL_EA, ClassVI}: BreakIndirect,
	{ClassGL_EA, ClassWJ}: BreakProhibited,
	{ClassGL_EA, ClassXX}: BreakIndirect,
	{ClassGL_EA, ClassZW}: BreakProhibited,
	{ClassGL_EA, ClassZWJ}: BreakIndirect,
	{ClassH2, ClassAI}: BreakDirect,
	{ClassH2, ClassAI_EA}: BreakDirect,
	{ClassH2, ClassAK}: BreakDirect,
	{ClassH2, ClassAL}: BreakDirect,
	{ClassH2, ClassAL_EA}: BreakDirect,
	{ClassH2, ClassAP}: BreakDirect,
	{ClassH2, ClassAS}: BreakDirect,
	{ClassH2, ClassB2}: BreakDirect,
	{ClassH2, ClassBA}: BreakIndirect,
	{ClassH2, ClassBA_EA}: BreakIndirect,
	{ClassH2, ClassBB}: BreakDirect,
	{ClassH2, ClassBK}: BreakProhibited,
	{ClassH2, ClassCB}: BreakDirect,
	{ClassH2, ClassCJ}: BreakIndirect,
	{ClassH2, ClassCL}: BreakProhibited,
	{ClassH2, ClassCL_EA}: BreakProhibited,
	{ClassH2, ClassCM}: BreakIndirect,
	{ClassH2, ClassCM_EA}: BreakIndirect,
	{ClassH2, ClassCP}: BreakProhibited,
	{ClassH2, ClassCR}: BreakProhibited,
	{ClassH2, ClassEB}: BreakDirect,
	{ClassH2, ClassEB_EA}: BreakDirect,
	{ClassH2, ClassEM}: BreakDirect,
	{ClassH2, ClassEX}: BreakProhibited,
	{ClassH2, ClassEX_EA}: BreakProhibited,
	{ClassH2, ClassGL}: BreakIndirect,
	{ClassH2, ClassGL_EA}: BreakIndirect,
	{ClassH2, ClassH2}: BreakDirect,
	{ClassH2, ClassH3}: BreakDirect,
	{ClassH2, ClassHH}: BreakIndirect,
	{ClassH2, ClassHL}: BreakDirect,
	{ClassH2, ClassHY}: BreakIndirect,
	{ClassH2, ClassID}: BreakDirect,
	{ClassH2, ClassID_EA}: BreakDirect,
	{ClassH2, ClassIN}: BreakIndirect,
	{ClassH2, ClassIN_EA}: BreakIndirect,
	{ClassH2, ClassIS}: BreakProhibited,
	{ClassH2, ClassJL}: BreakDirect,
	{ClassH2, ClassJT}: BreakIndirect,
	{ClassH2, ClassJV}: BreakIndirect,
	{ClassH2, ClassLF}: BreakProhibited,
	{ClassH2, ClassNL}: BreakProhibited,
	{ClassH2, ClassNS}: BreakIndirect,
	{ClassH2, ClassNS_EA}: BreakIndirect,
	{ClassH2, ClassNU}: BreakDirect,
	{ClassH2, ClassOP}: BreakDirect,
	{ClassH2, ClassOP_EA}: BreakDirect,
	{ClassH2, ClassPO}: BreakIndirect,
	{ClassH2, ClassPO_EA}: BreakIndirect,
	{ClassH2, ClassPR}: BreakDirect,
	{ClassH2, ClassPR_EA}: BreakDirect,
	{ClassH2, ClassQU}: BreakIndirect,
	{ClassH2, ClassQU_Pf}: BreakProhibited,
	{ClassH2, ClassQU_Pi}: BreakIndirect,
	{ClassH2, ClassRI}: BreakDirect,
	{ClassH2, ClassSA}: BreakDirect,
	{ClassH2, ClassSP}: BreakProhibited,
	{ClassH2, ClassSY}: BreakProhibited,
	{ClassH2, ClassVF}: BreakDirect,
	{ClassH2, ClassVI}: BreakDirect,
	{ClassH2, ClassWJ}: BreakProhibited,
	{ClassH2, ClassXX}: BreakDirect,
	{ClassH2, ClassZW}: BreakProhibited,
	{ClassH2, ClassZWJ}: BreakIndirect,
	{ClassH3, ClassAI}: BreakDirect,
	{ClassH3, ClassAI_EA}: BreakDirect,
	{ClassH3, ClassAK}: BreakDirect,
	{ClassH3, ClassAL}: BreakDirect,
	{ClassH3, ClassAL_EA}: BreakDirect,
	{ClassH3, ClassAP}: BreakDirect,
	{ClassH3, ClassAS}: BreakDirect,
	{ClassH3, ClassB2}: BreakDirect,
	{ClassH3, ClassBA}: BreakIndirect,
	{ClassH3, ClassBA_EA}: BreakIndirect,
	{ClassH3, ClassBB}: BreakDirect,
	{ClassH3, ClassBK}: BreakProhibited,
	{ClassH3, ClassCB}: BreakDirect,
	{ClassH3, ClassCJ}: BreakIndirect,
	{ClassH3, ClassCL}: BreakProhibited,
	{ClassH3, ClassCL_EA}: BreakProhibited,
	{ClassH3, ClassCM}: BreakIndirect,
	{ClassH3, ClassCM_EA}: BreakIndirect,
	{ClassH3, ClassCP}: BreakProhibited,
	{ClassH3, ClassCR}: BreakProhibited,
	{ClassH3, ClassEB}: BreakDirect,
	{ClassH3, ClassEB_EA}: BreakDirect,
	{ClassH3, ClassEM}: BreakDirect,
	{ClassH3, ClassEX}: BreakProhibited,
	{ClassH3, ClassEX_EA}: BreakProhibited,
	{ClassH3, ClassGL}: BreakIndirect,
	{ClassH3, ClassGL_EA}: BreakIndirect,
	{ClassH3, ClassH2}: BreakDirect,
	{ClassH3, ClassH3}: BreakDirect,
	{ClassH3, ClassHH}: BreakIndirect,
	{ClassH3, ClassHL}: BreakDirect,
	{ClassH3, ClassHY}: BreakIndirect,
	{ClassH3, ClassID}: BreakDirect,
	{ClassH3, ClassID_EA}: BreakDirect,
	{ClassH3, ClassIN}: BreakIndirect,
	{ClassH3, ClassIN_EA}: BreakIndirect,
	{ClassH3, ClassIS}: BreakProhibited,
	{ClassH3, ClassJL}: BreakDirect,
	{ClassH3, ClassJT}: BreakIndirect,
	{ClassH3, ClassJV}: BreakDirect,
	{ClassH3, ClassLF}: BreakProhibited,
	{ClassH3, ClassNL}: BreakProhibited,
	{ClassH3, ClassNS}: BreakIndirect,
	{ClassH3, ClassNS_EA}: BreakIndirect,
	{ClassH3, ClassNU}: BreakDirect,
	{ClassH3, ClassOP}: BreakDirect,
	{ClassH3, ClassOP_EA}: BreakDirect,
	{ClassH3, ClassPO}: BreakIndirect,
	{ClassH3, ClassPO_EA}: BreakIndirect,
	{ClassH3, ClassPR}: BreakDirect,
	{ClassH3, ClassPR_EA}: BreakDirect,
	{ClassH3, ClassQU}: BreakIndirect,
	{ClassH3, ClassQU_Pf}: BreakProhibited,
	{ClassH3, ClassQU_Pi}: BreakIndirect,
	{ClassH3, ClassRI}: BreakDirect,
	{ClassH3, ClassSA}: BreakDirect,
	{ClassH3, ClassSP}: BreakProhibited,
	{ClassH3, ClassSY}: BreakProhibited,
	{ClassH3, ClassVF}: BreakDirect,
	{ClassH3, ClassVI}: BreakDirect,
	{ClassH3, ClassWJ}: BreakProhibited,
	{ClassH3, ClassXX}: BreakDirect,
	{ClassH3, ClassZW}: BreakProhibited,
	{ClassH3, ClassZWJ}: BreakIndirect,
	{ClassHH, ClassAI}: BreakIndirect,
	{ClassHH, ClassAI_EA}: BreakIndirect,
	{ClassHH, ClassAK}: BreakDirect,
	{ClassHH, ClassAL}: BreakIndirect,
	{ClassHH, ClassAL_EA}: BreakIndirect,
	{ClassHH, ClassAP}: BreakDirect,
	{ClassHH, ClassAS}: BreakDirect,
	{ClassHH, ClassB2}: BreakDirect,
	{ClassHH, ClassBA}: BreakIndirect,
	{ClassHH, ClassBA_EA}: BreakIndirect,
	{ClassHH, ClassBB}: BreakDirect,
	{ClassHH, ClassBK}: BreakProhibited,
	{ClassHH, ClassCB}: BreakDirect,
	{ClassHH, ClassCJ}: BreakIndirect,
	{ClassHH, ClassCL}: BreakProhibited,
	{ClassHH, ClassCL_EA}: BreakProhibited,
	{ClassHH, ClassCM}: BreakIndirect,
	{ClassHH, ClassCM_EA}: BreakIndirect,
	{ClassHH, ClassCP}: BreakProhibited,
	{ClassHH, ClassCR}: BreakProhibited,
	{ClassHH, ClassEB}: BreakDirect,
	{ClassHH, ClassEB_EA}: BreakDirect,
	{ClassHH, ClassEM}: BreakDirect,
	{ClassHH, ClassEX}: BreakProhibited,
	{ClassHH, ClassEX_EA}: BreakProhibited,
	{ClassHH, ClassGL}: BreakDirect,
	{ClassHH, ClassGL_EA}: BreakDirect,
	{ClassHH, ClassH2}: BreakDirect,
	{ClassHH, ClassH3}: BreakDirect,
	{ClassHH, ClassHH}: BreakIndirect,
	{ClassHH, ClassHL}: BreakIndirect,
	{ClassHH, ClassHY}: BreakIndirect,
	{ClassHH, ClassID}: BreakDirect,
	{ClassHH, ClassID_EA}: BreakDirect,
	{ClassHH, ClassIN}: BreakIndirect,
	{ClassHH, ClassIN_EA}: BreakIndirect,
	{ClassHH, ClassIS}: BreakProhibited,
	{ClassHH, ClassJL}: BreakDirect,
	{ClassHH, ClassJT}: BreakDirect,
	{ClassHH, ClassJV}: BreakDirect,
	{ClassHH, ClassLF}: BreakProhibited,
	{ClassHH, ClassNL}: BreakProhibited,
	{ClassHH, ClassNS}: BreakIndirect,
	{ClassHH, ClassNS_EA}: BreakIndirect,
	{ClassHH, ClassNU}: BreakDirect,
	{ClassHH, ClassOP}: BreakDirect,
	{ClassHH, ClassOP_EA}: BreakDirect,
	{ClassHH, ClassPO}: BreakDirect,
	{ClassHH, ClassPO_EA}: BreakDirect,
	{ClassHH, ClassPR}: BreakDirect,
	{ClassHH, ClassPR_EA}: BreakDirect,
	{ClassHH, ClassQU}: BreakIndirect,
	{ClassHH, ClassQU_Pf}: BreakProhibited,
	{ClassHH, ClassQU_Pi}: BreakIndirect,
	{ClassHH, ClassRI}: BreakDirect,
	{ClassHH, ClassSA}: BreakIndirect,
	{ClassHH, ClassSP}: BreakProhibited,
	{ClassHH, ClassSY}: BreakProhibited,
	{ClassHH, ClassVF}: BreakDirect,
	{ClassHH, ClassVI}: BreakDirect,
	{ClassHH, ClassWJ}: BreakProhibited,
	{ClassHH, ClassXX}: BreakIndirect,
	{ClassHH, ClassZW}: BreakProhibited,
	{ClassHH, ClassZWJ}: BreakIndirect,
	{ClassHL, ClassAI}: BreakIndirect,
	{ClassHL, ClassAI_EA}: BreakIndirect,
	{ClassHL, ClassAK}: BreakDirect,
	{ClassHL, ClassAL}: BreakIndirect,
	{ClassHL, ClassAL_EA}: BreakIndirect,
	{ClassHL, ClassAP}: BreakDirect,
	{ClassHL, ClassAS}: BreakDirect,
	{ClassHL, ClassB2}: BreakDirect,
	{ClassHL, ClassBA}: BreakIndirect,
	{ClassHL, ClassBA_EA}: BreakIndirect,
	{ClassHL, ClassBB}: BreakDirect,
	{ClassHL, ClassBK}: BreakProhibited,
	{ClassHL, ClassCB}: BreakDirect,
	{ClassHL, ClassCJ}: BreakIndirect,
	{ClassHL, ClassCL}: BreakProhibited,
	{ClassHL, ClassCL_EA}: BreakProhibited,
	{ClassHL, ClassCM}: BreakIndirect,
	{ClassHL, ClassCM_EA}: BreakIndirect,
	{ClassHL, ClassCP}: BreakProhibited,
	{ClassHL, ClassCR}: BreakProhibited,
	{ClassHL, ClassEB}: BreakDirect,
	{ClassHL, ClassEB_EA}: BreakDirect,
	{ClassHL, ClassEM}: BreakDirect,
	{ClassHL, ClassEX}: BreakProhibited,
	{ClassHL, ClassEX_EA}: BreakProhibited,
	{ClassHL, ClassGL}: BreakIndirect,
	{ClassHL, ClassGL_EA}: BreakIndirect,
	{ClassHL, ClassH2}: BreakDirect,
	{ClassHL, ClassH3}: BreakDirect,
	{ClassHL, ClassHH}: BreakIndirect,
	{ClassHL, ClassHL}: BreakIndirect,
	{ClassHL, ClassHY}: BreakIndirect,
	{ClassHL, ClassID}: BreakDirect,
	{ClassHL, ClassID_EA}: BreakDirect,
	{ClassHL, ClassIN}: BreakIndirect,
	{ClassHL, ClassIN_EA}: BreakIndirect,
	{ClassHL, ClassIS}: BreakProhibited,
	{ClassHL, ClassJL}: BreakDirect,
	{ClassHL, ClassJT}: BreakDirect,
	{ClassHL, ClassJV}: BreakDirect,
	{ClassHL, ClassLF}: BreakProhibited,
	{ClassHL, ClassNL}: BreakProhibited,
	{ClassHL, ClassNS}: BreakIndirect,
	{ClassHL, ClassNS_EA}: BreakIndirect,
	{ClassHL, ClassNU}: BreakIndirect,
	{ClassHL, ClassOP}: BreakIndirect,
	{ClassHL, ClassOP_EA}: BreakDirect,
	{ClassHL, ClassPO}: BreakIndirect,
	{ClassHL, ClassPO_EA}: BreakIndirect,
	{ClassHL, ClassPR}: BreakIndirect,
	{ClassHL, ClassPR_EA}: BreakIndirect,
	{ClassHL, ClassQU}: BreakIndirect,
	{ClassHL, ClassQU_Pf}: BreakProhibited,
	{ClassHL, ClassQU_Pi}: BreakIndirect,
	{ClassHL, ClassRI}: BreakDirect,
	{ClassHL, ClassSA}: BreakIndirect,
	{ClassHL, ClassSP}: BreakProhibited,
	{ClassHL, ClassSY}: BreakProhibited,
	{ClassHL, ClassVF}: BreakDirect,
	{ClassHL, ClassVI}: BreakDirect,
	{ClassHL, ClassWJ}: BreakProhibited,
	{ClassHL, ClassXX}: BreakIndirect,
	{ClassHL, ClassZW}: BreakProhibited,
	{ClassHL, ClassZWJ}: BreakIndirect,
	{ClassHY, ClassAI}: BreakIndirect,
	{ClassHY, ClassAI_EA}: BreakIndirect,
	{ClassHY, ClassAK}: BreakDirect,
	{ClassHY, ClassAL}: BreakIndirect,
	{ClassHY, ClassAL_EA}: BreakIndirect,
	{ClassHY, ClassAP}: BreakDirect,
	{ClassHY, ClassAS}: BreakDirect,
	{ClassHY, ClassB2}: BreakDirect,
	{ClassHY, ClassBA}: BreakIndirect,
	{ClassHY, ClassBA_EA}: BreakIndirect,
	{ClassHY, ClassBB}: BreakDirect,
	{ClassHY, ClassBK}: BreakProhibited,
	{ClassHY, ClassCB}: BreakDirect,
	{ClassHY, ClassCJ}: BreakIndirect,
	{ClassHY, ClassCL}: BreakProhibited,
	{ClassHY, ClassCL_EA}: BreakProhibited,
	{ClassHY, ClassCM}: BreakIndirect,
	{ClassHY, ClassCM_EA}: BreakIndirect,
	{ClassHY, ClassCP}: BreakProhibited,
	{ClassHY, ClassCR}: BreakProhibited,
	{ClassHY, ClassEB}: BreakDirect,
	{ClassHY, ClassEB_EA}: BreakDirect,
	{ClassHY, ClassEM}: BreakDirect,
	{ClassHY, ClassEX}: BreakProhibited,
	{ClassHY, ClassEX_EA}: BreakProhibited,
	{ClassHY, ClassGL}: BreakDirect,
	{ClassHY, ClassGL_EA}: BreakDirect,
	{ClassHY, ClassH2}: BreakDirect,
	{ClassHY, ClassH3}: BreakDirect,
	{ClassHY, ClassHH}: BreakIndirect,
	{ClassHY, ClassHL}: BreakIndirect,
	{ClassHY, ClassHY}: BreakIndirect,
	{ClassHY, ClassID}: BreakDirect,
	{ClassHY, ClassID_EA}: BreakDirect,
	{ClassHY, ClassIN}: BreakIndirect,
	{ClassHY, ClassIN_EA}: BreakIndirect,
	{ClassHY, ClassIS}: BreakProhibited,
	{ClassHY, ClassJL}: BreakDirect,
	{ClassHY, ClassJT}: BreakDirect,
	{ClassHY, ClassJV}: BreakDirect,
	{ClassHY, ClassLF}: BreakProhibited,
	{ClassHY, ClassNL}: BreakProhibited,
	{ClassHY, ClassNS}: BreakIndirect,
	{ClassHY, ClassNS_EA}: BreakIndirect,
	{ClassHY, ClassNU}: BreakIndirect,
	{ClassHY, ClassOP}: BreakDirect,
	{ClassHY, ClassOP_EA}: BreakDirect,
	{ClassHY, ClassPO}: BreakDirect,
	{ClassHY, ClassPO_EA}: BreakDirect,
	{ClassHY, ClassPR}: BreakDirect,
	{ClassHY, ClassPR_EA}: BreakDirect,
	{ClassHY, ClassQU}: BreakIndirect,
	{ClassHY, ClassQU_Pf}: BreakProhibited,
	{ClassHY, ClassQU_Pi}: BreakIndirect,
	{ClassHY, ClassRI}: BreakDirect,
	{ClassHY, ClassSA}: BreakIndirect,
	{ClassHY, ClassSP}: BreakProhibited,
	{ClassHY, ClassSY}: BreakProhibited,
	{ClassHY, ClassVF}: BreakDirect,
	{ClassHY, ClassVI}: BreakDirect,
	{ClassHY, ClassWJ}: BreakProhibited,
	{ClassHY, ClassXX}: BreakIndirect,
	{ClassHY, ClassZW}: BreakProhibited,
	{ClassHY, ClassZWJ}: BreakIndirect,
	{ClassID, ClassAI}: BreakDirect,
	{ClassID, ClassAI_EA}: BreakDirect,
	{ClassID, ClassAK}: BreakDirect,
	{ClassID, ClassAL}: BreakDirect,
	{ClassID, ClassAL_EA}: BreakDirect,
	{ClassID, ClassAP}: BreakDirect,
	{ClassID, ClassAS}: BreakDirect,
	{ClassID, ClassB2}: BreakDirect,
	{ClassID, ClassBA}: BreakIndirect,
	{ClassID, ClassBA_EA}: BreakIndirect,
	{ClassID, ClassBB}: BreakDirect,
	{ClassID, ClassBK}: BreakProhibited,
	{ClassID, ClassCB}: BreakDirect,
	{ClassID, ClassCJ}: BreakIndirect,
	{ClassID, ClassCL}: BreakProhibited,
	{ClassID, ClassCL_EA}: BreakProhibited,
	{ClassID, ClassCM}: BreakIndirect,
	{ClassID, ClassCM_EA}: BreakIndirect,
	{ClassID, ClassCP}: BreakProhibited,
	{ClassID, ClassCR}: BreakProhibited,
	{ClassID, ClassEB}: BreakDirect,
	{ClassID, ClassEB_EA}: BreakDirect,
	{ClassID, ClassEM}: BreakDirect,
	{ClassID, ClassEX}: BreakProhibited,
	{ClassID, ClassEX_EA}: BreakProhibited,
	{ClassID, ClassGL}: BreakIndirect,
	{ClassID, ClassGL_EA}: BreakIndirect,
	{ClassID, ClassH2}: BreakDirect,
	{ClassID, ClassH3}: BreakDirect,
	{ClassID, ClassHH}: BreakIndirect,
	{ClassID, ClassHL}: BreakDirect,
	{ClassID, ClassHY}: BreakIndirect,
	{ClassID, ClassID}: BreakDirect,
	{ClassID, ClassID_EA}: BreakDirect,
	{ClassID, ClassIN}: BreakIndirect,
	{ClassID, ClassIN_EA}: BreakIndirect,
	{ClassID, ClassIS}: BreakProhibited,
	{ClassID, ClassJL}: BreakDirect,
	{ClassID, ClassJT}: BreakDirect,
	{ClassID, ClassJV}: BreakDirect,
	{ClassID, ClassLF}: BreakProhibited,
	{ClassID, ClassNL}: BreakProhibited,
	{ClassID, ClassNS}: BreakIndirect,
	{ClassID, ClassNS_EA}: BreakIndirect,
	{ClassID, ClassNU}: BreakDirect,
	{ClassID, ClassOP}: BreakDirect,
	{ClassID, ClassOP_EA}: BreakDirect,
	{ClassID, ClassPO}: BreakIndirect,
	{ClassID, ClassPO_EA}: BreakIndirect,
	{ClassID, ClassPR}: BreakDirect,
	{ClassID, ClassPR_EA}: BreakDirect,
	{ClassID, ClassQU}: BreakIndirect,
	{ClassID, ClassQU_Pf}: BreakProhibited,
	{ClassID, ClassQU_Pi}: BreakIndirect,
	{ClassID, ClassRI}: BreakDirect,
	{ClassID, ClassSA}: BreakDirect,
	{ClassID, ClassSP}: BreakProhibited,
	{ClassID, ClassSY}: BreakProhibited,
	{ClassID, ClassVF}: BreakDirect,
	{ClassID, ClassVI}: BreakDirect,
	{ClassID, ClassWJ}: BreakProhibited,
	{ClassID, ClassXX}: BreakDirect,
	{ClassID, ClassZW}: BreakProhibited,
	{ClassID, ClassZWJ}: BreakIndirect,
	{ClassID_EA, ClassAI}: BreakDirect,
	{ClassID_EA, ClassAI_EA}: BreakDirect,
	{ClassID_EA, ClassAK}: BreakDirect,
	{ClassID_EA, ClassAL}: BreakDirect,
	{ClassID_EA, ClassAL_EA}: BreakDirect,
	{ClassID_EA, ClassAP}: BreakDirect,
	{ClassID_EA, ClassAS}: BreakDirect,
	{ClassID_EA, ClassB2}: BreakDirect,
	{ClassID_EA, ClassBA}: BreakIndirect,
	{ClassID_EA, ClassBA_EA}: BreakIndirect,
	{ClassID_EA, ClassBB}: BreakDirect,
	{ClassID_EA, ClassBK}: BreakProhibited,
	{ClassID_EA, ClassCB}: BreakDirect,
	{ClassID_EA, ClassCJ}: BreakIndirect,
	{ClassID_EA, ClassCL}: BreakProhibited,
	{ClassID_EA, ClassCL_EA}: BreakProhibited,
	{ClassID_EA, ClassCM}: BreakIndirect,
	{ClassID_EA, ClassCM_EA}: BreakIndirect,
	{ClassID_EA, ClassCP}: BreakProhibited,
	{ClassID_EA, ClassCR}: BreakProhibited,
	{ClassID_EA, ClassEB}: BreakDirect,
	{ClassID_EA, ClassEB_EA}: BreakDirect,
	{ClassID_EA, ClassEM}: BreakDirect,
	{ClassID_EA, ClassEX}: BreakProhibited,
	{ClassID_EA, ClassEX_EA}: BreakProhibited,
	{ClassID_EA, ClassGL}: BreakIndirect,
	{ClassID_EA, ClassGL_EA}: BreakIndirect,
	{ClassID_EA, ClassH2}: BreakDirect,
	{ClassID_EA, ClassH3}: BreakDirect,
	{ClassID_EA, ClassHH}: BreakIndirect,
	{ClassID_EA, ClassHL}: BreakDirect,
	{ClassID_EA, ClassHY}: BreakIndirect,
	{ClassID_EA, ClassID}: BreakDirect,
	{ClassID_EA, ClassID_EA}: BreakDirect,
	{ClassID_EA, ClassIN}: BreakIndirect,
	{ClassID_EA, ClassIN_EA}: BreakIndirect,
	{ClassID_EA, ClassIS}: BreakProhibited,
	{ClassID_EA, ClassJL}: BreakDirect,
	{ClassID_EA, ClassJT}: BreakDirect,
	{ClassID_EA, ClassJV}: BreakDirect,
	{ClassID_EA, ClassLF}: BreakProhibited,
	{ClassID_EA, ClassNL}: BreakProhibited,
	{ClassID_EA, ClassNS}: BreakIndirect,
	{ClassID_EA, ClassNS_EA}: BreakIndirect,
	{ClassID_EA, ClassNU}: BreakDirect,
	{ClassID_EA, ClassOP}: BreakDirect,
	{ClassID_EA, ClassOP_EA}: BreakDirect,
	{ClassID_EA, ClassPO}: BreakIndirect,
	{ClassID_EA, ClassPO_EA}: BreakIndirect,
	{ClassID_EA, ClassPR}: BreakDirect,
	{ClassID_EA, ClassPR_EA}: BreakDirect,
	{ClassID_EA, ClassQU}: BreakIndirect,
	{ClassID_EA, ClassQU_Pf}: BreakProhibited,
	{ClassID_EA, ClassQU_Pi}: BreakIndirect,
	{ClassID_EA, ClassRI}: BreakDirect,
	{ClassID_EA, ClassSA}: BreakDirect,
	{ClassID_EA, ClassSP}: BreakProhibited,
	{ClassID_EA, ClassSY}: BreakProhibited,
	{ClassID_EA, ClassVF}: BreakDirect,
	{ClassID_EA, ClassVI}: BreakDirect,
	{ClassID_EA, ClassWJ}: BreakProhibited,
	{ClassID_EA, ClassXX}: BreakDirect,
	{ClassID_EA, ClassZW}: BreakProhibited,
	{ClassID_EA, ClassZWJ}: BreakIndirect,
	{ClassIN, ClassAI}: BreakDirect,
	{ClassIN, ClassAI_EA}: BreakDirect,
	{ClassIN, ClassAK}: BreakDirect,
	{ClassIN, ClassAL}: BreakDirect,
	{ClassIN, ClassAL_EA}: BreakDirect,
	{ClassIN, ClassAP}: BreakDirect,
	{ClassIN, ClassAS}: BreakDirect,
	{ClassIN, ClassB2}: BreakDirect,
	{ClassIN, ClassBA}: BreakIndirect,
	{ClassIN, ClassBA_EA}: BreakIndirect,
	{ClassIN, ClassBB}: BreakDirect,
	{ClassIN, ClassBK}: BreakProhibited,
	{ClassIN, ClassCB}: BreakDirect,
	{ClassIN, ClassCJ}: BreakIndirect,
	{ClassIN, ClassCL}: BreakProhibited,
	{ClassIN, ClassCL_EA}: BreakProhibited,
	{ClassIN, ClassCM}: BreakIndirect,
	{ClassIN, ClassCM_EA}: BreakIndirect,
	{ClassIN, ClassCP}: BreakProhibited,
	{ClassIN, ClassCR}: BreakProhibited,
	{ClassIN, ClassEB}: BreakDirect,
	{ClassIN, ClassEB_EA}: BreakDirect,
	{ClassIN, ClassEM}: BreakDirect,
	{ClassIN, ClassEX}: BreakProhibited,
	{ClassIN, ClassEX_EA}: BreakProhibited,
	{ClassIN, ClassGL}: BreakIndirect,
	{ClassIN, ClassGL_EA}: BreakIndirect,
	{ClassIN, ClassH2}: BreakDirect,
	{ClassIN, ClassH3}: BreakDirect,
	{ClassIN, ClassHH}: BreakIndirect,
	{ClassIN, ClassHL}: BreakDirect,
	{ClassIN, ClassHY}: BreakIndirect,
	{ClassIN, ClassID}: BreakDirect,
	{ClassIN, ClassID_EA}: BreakDirect,
	{ClassIN, ClassIN}: BreakIndirect,
	{ClassIN, ClassIN_EA}: BreakIndirect,
	{ClassIN, ClassIS}: BreakProhibited,
	{ClassIN, ClassJL}: BreakDirect,
	{ClassIN, ClassJT}: BreakDirect,
	{ClassIN, ClassJV}: BreakDirect,
	{ClassIN, ClassLF}: BreakProhibited,
	{ClassIN, ClassNL}: BreakProhibited,
	{ClassIN, ClassNS}: BreakIndirect,
	{ClassIN, ClassNS_EA}: BreakIndirect,
	{ClassIN, ClassNU}: BreakDirect,
	{ClassIN, ClassOP}: BreakDirect,
	{ClassIN, ClassOP_EA}: BreakDirect,
	{ClassIN, ClassPO}: BreakDirect,
	{ClassIN, ClassPO_EA}: BreakDirect,
	{ClassIN, ClassPR}: BreakDirect,
	{ClassIN, ClassPR_EA}: BreakDirect,
	{ClassIN, ClassQU}: BreakIndirect,
	{ClassIN, ClassQU_Pf}: BreakProhibited,
	{ClassIN, ClassQU_Pi}: BreakIndirect,
	{ClassIN, ClassRI}: BreakDirect,
	{ClassIN, ClassSA}: BreakDirect,
	{ClassIN, ClassSP}: BreakProhibited,
	{ClassIN, ClassSY}: BreakProhibited,
	{ClassIN, ClassVF}: BreakDirect,
	{ClassIN, ClassVI}: BreakDirect,
	{ClassIN, ClassWJ}: BreakProhibited,
	{ClassIN, ClassXX}: BreakDirect,
	{ClassIN, ClassZW}: BreakProhibited,
	{ClassIN, ClassZWJ}: BreakIndirect,
	{ClassIN_EA, ClassAI}: BreakDirect,
	{ClassIN_EA, ClassAI_EA}: BreakDirect,
	{ClassIN_EA, ClassAK}: BreakDirect,
	{ClassIN_EA, ClassAL}: BreakDirect,
	{ClassIN_EA, ClassAL_EA}: BreakDirect,
	{ClassIN_EA, ClassAP}: BreakDirect,
	{ClassIN_EA, ClassAS}: BreakDirect,
	{ClassIN_EA, ClassB2}: BreakDirect,
	{ClassIN_EA, ClassBA}: BreakIndirect,
	{ClassIN_EA, ClassBA_EA}: BreakIndirect,
	{ClassIN_EA, ClassBB}: BreakDirect,
	{ClassIN_EA, ClassBK}: BreakProhibited,
	{ClassIN_EA, ClassCB}: BreakDirect,
	{ClassIN_EA, ClassCJ}: BreakIndirect,
	{ClassIN_EA, ClassCL}: BreakProhibited,
	{ClassIN_EA, ClassCL_EA}: BreakProhibited,
	{ClassIN_EA, ClassCM}: BreakIndirect,
	{ClassIN_EA, ClassCM_EA}: BreakIndirect,
	{ClassIN_EA, ClassCP}: BreakProhibited,
	{ClassIN_EA, ClassCR}: BreakProhibited,
	{ClassIN_EA, ClassEB}: BreakDirect,
	{ClassIN_EA, ClassEB_EA}: BreakDirect,
	{ClassIN_EA, ClassEM}: BreakDirect,
	{ClassIN_EA, ClassEX}: BreakProhibited,
	{ClassIN_EA, ClassEX_EA}: BreakProhibited,
	{ClassIN_EA, ClassGL}: BreakIndirect,
	{ClassIN_EA, ClassGL_EA}: BreakIndirect,
	{ClassIN_EA, ClassH2}: BreakDirect,
	{ClassIN_EA, ClassH3}: BreakDirect,
	{ClassIN_EA, ClassHH}: BreakIndirect,
	{ClassIN_EA, ClassHL}: BreakDirect,
	{ClassIN_EA, ClassHY}: BreakIndirect,
	{ClassIN_EA, ClassID}: BreakDirect,
	{ClassIN_EA, ClassID_EA}: BreakDirect,
	{ClassIN_EA, ClassIN}: BreakIndirect,
	{ClassIN_EA, ClassIN_EA}: BreakIndirect,
	{ClassIN_EA, ClassIS}: BreakProhibited,
	{ClassIN_EA, ClassJL}: BreakDirect,
	{ClassIN_EA, ClassJT}: BreakDirect,
	{ClassIN_EA, ClassJV}: BreakDirect,
	{ClassIN_EA, ClassLF}: BreakProhibited,
	{ClassIN_EA, ClassNL}: BreakProhibited,
	{ClassIN_EA, ClassNS}: BreakIndirect,
	{ClassIN_EA, ClassNS_EA}: BreakIndirect,
	{ClassIN_EA, ClassNU}: BreakDirect,
	{ClassIN_EA, ClassOP}: BreakDirect,
	{ClassIN_EA, ClassOP_EA}: BreakDirect,
	{ClassIN_EA, ClassPO}: BreakDirect,
	{ClassIN_EA, ClassPO_EA}: BreakDirect,
	{ClassIN_EA, ClassPR}: BreakDirect,
	{ClassIN_EA, ClassPR_EA}: BreakDirect,
	{ClassIN_EA, ClassQU}: BreakIndirect,
	{ClassIN_EA, ClassQU_Pf}: BreakProhibited,
	{ClassIN_EA, ClassQU_Pi}: BreakIndirect,
	{ClassIN_EA, ClassRI}: BreakDirect,
	{ClassIN_EA, ClassSA}: BreakDirect,
	{ClassIN_EA, ClassSP}: BreakProhibited,
	{ClassIN_EA, ClassSY}: BreakProhibited,
	{ClassIN_EA, ClassVF}: BreakDirect,
	{ClassIN_EA, ClassVI}: BreakDirect,
	{ClassIN_EA, ClassWJ}: BreakProhibited,
	{ClassIN_EA, ClassXX}: BreakDirect,
	{ClassIN_EA, ClassZW}: BreakProhibited,
	{ClassIN_EA, ClassZWJ}: BreakIndirect,
	{ClassIS, ClassAI}: BreakIndirect,
	{ClassIS, ClassAI_EA}: BreakIndirect,
	{ClassIS, ClassAK}: BreakDirect,
	{ClassIS, ClassAL}: BreakIndirect,
	{ClassIS, ClassAL_EA}: BreakIndirect,
	{ClassIS, ClassAP}: BreakDirect,
	{ClassIS, ClassAS}: BreakDirect,
	{ClassIS, ClassB2}: BreakDirect,
	{ClassIS, ClassBA}: BreakIndirect,
	{ClassIS, ClassBA_EA}: BreakIndirect,
	{ClassIS, ClassBB}: BreakDirect,
	{ClassIS, ClassBK}: BreakProhibited,
	{ClassIS, ClassCB}: BreakDirect,
	{ClassIS, ClassCJ}: BreakIndirect,
	{ClassIS, ClassCL}: BreakProhibited,
	{ClassIS, ClassCL_EA}: BreakProhibited,
	{ClassIS, ClassCM}: BreakIndirect,
	{ClassIS, ClassCM_EA}: BreakIndirect,
	{ClassIS, ClassCP}: BreakProhibited,
	{ClassIS, ClassCR}: BreakProhibited,
	{ClassIS, ClassEB}: BreakDirect,
	{ClassIS, ClassEB_EA}: BreakDirect,
	{ClassIS, ClassEM}: BreakDirect,
	{ClassIS, ClassEX}: BreakProhibited,
	{ClassIS, ClassEX_EA}: BreakProhibited,
	{ClassIS, ClassGL}: BreakIndirect,
	{ClassIS, ClassGL_EA}: BreakIndirect,
	{ClassIS, ClassH2}: BreakDirect,
	{ClassIS, ClassH3}: BreakDirect,
	{ClassIS, ClassHH}: BreakIndirect,
	{ClassIS, ClassHL}: BreakIndirect,
	{ClassIS, ClassHY}: BreakIndirect,
	{ClassIS, ClassID}: BreakDirect,
	{ClassIS, ClassID_EA}: BreakDirect,
	{ClassIS, ClassIN}: BreakIndirect,
	{ClassIS, ClassIN_EA}: BreakIndirect,
	{ClassIS, ClassIS}: BreakProhibited,
	{ClassIS, ClassJL}: BreakDirect,
	{ClassIS, ClassJT}: BreakDirect,
	{ClassIS, ClassJV}: BreakDirect,
	{ClassIS, ClassLF}: BreakProhibited,
	{ClassIS, ClassNL}: BreakProhibited,
	{ClassIS, ClassNS}: BreakIndirect,
	{ClassIS, ClassNS_EA}: BreakIndirect,
	{ClassIS, ClassNU}: BreakIndirect,
	{ClassIS, ClassOP}: BreakDirect,
	{ClassIS, ClassOP_EA}: BreakDirect,
	{ClassIS, ClassPO}: BreakDirect,
	{ClassIS, ClassPO_EA}: BreakDirect,
	{ClassIS, ClassPR}: BreakDirect,
	{ClassIS, ClassPR_EA}: BreakDirect,
	{ClassIS, ClassQU}: BreakIndirect,
	{ClassIS, ClassQU_Pf}: BreakProhibited,
	{ClassIS, ClassQU_Pi}: BreakIndirect,
	{ClassIS, ClassRI}: BreakDirect,
	{ClassIS, ClassSA}: BreakIndirect,
	{ClassIS, ClassSP}: BreakProhibited,
	{ClassIS, ClassSY}: BreakProhibited,
	{ClassIS, ClassVF}: BreakDirect,
	{ClassIS, ClassVI}: BreakDirect,
	{ClassIS, ClassWJ}: BreakProhibited,
	{ClassIS, ClassXX}: BreakIndirect,
	{ClassIS, ClassZW}: BreakProhibited,
	{ClassIS, ClassZWJ}: BreakIndirect,
	{ClassJL, ClassAI}: BreakDirect,
	{ClassJL, ClassAI_EA}: BreakDirect,
	{ClassJL, ClassAK}: BreakDirect,
	{ClassJL, ClassAL}: BreakDirect,
	{ClassJL, ClassAL_EA}: BreakDirect,
	{ClassJL, ClassAP}: BreakDirect,
	{ClassJL, ClassAS}: BreakDirect,
	{ClassJL, ClassB2}: BreakDirect,
	{ClassJL, ClassBA}: BreakIndirect,
	{ClassJL, ClassBA_EA}: BreakIndirect,
	{ClassJL, ClassBB}: BreakDirect,
	{ClassJL, ClassBK}: BreakProhibited,
	{ClassJL, ClassCB}: BreakDirect,
	{ClassJL, ClassCJ}: BreakIndirect,
	{ClassJL, ClassCL}: BreakProhibited,
	{ClassJL, ClassCL_EA}: BreakProhibited,
	{ClassJL, ClassCM}: BreakIndirect,
	{ClassJL, ClassCM_EA}: BreakIndirect,
	{ClassJL, ClassCP}: BreakProhibited,
	{ClassJL, ClassCR}: BreakProhibited,
	{ClassJL, ClassEB}: BreakDirect,
	{ClassJL, ClassEB_EA}: BreakDirect,
	{ClassJL, ClassEM}: BreakDirect,
	{ClassJL, ClassEX}: BreakProhibited,
	{ClassJL, ClassEX_EA}: BreakProhibited,
	{ClassJL, ClassGL}: BreakIndirect,
	{ClassJL, ClassGL_EA}: BreakIndirect,
	{ClassJL, ClassH2}: BreakIndirect,
	{ClassJL, ClassH3}: BreakIndirect,
	{ClassJL, ClassHH}: BreakIndirect,
	{ClassJL, ClassHL}: BreakDirect,
	{ClassJL, ClassHY}: BreakIndirect,
	{ClassJL, ClassID}: BreakDirect,
	{ClassJL, ClassID_EA}: BreakDirect,
	{ClassJL, ClassIN}: BreakIndirect,
	{ClassJL, ClassIN_EA}: BreakIndirect,
	{ClassJL, ClassIS}: BreakProhibited,
	{ClassJL, ClassJL}: BreakIndirect,
	{ClassJL, ClassJT}: BreakDirect,
	{ClassJL, ClassJV}: BreakIndirect,
	{ClassJL, ClassLF}: BreakProhibited,
	{ClassJL, ClassNL}: BreakProhibited,
	{ClassJL, ClassNS}: BreakIndirect,
	{ClassJL, ClassNS_EA}: BreakIndirect,
	{ClassJL, ClassNU}: BreakDirect,
	{ClassJL, ClassOP}: BreakDirect,
	{ClassJL, ClassOP_EA}: BreakDirect,
	{ClassJL, ClassPO}: BreakIndirect,
	{ClassJL, ClassPO_EA}: BreakIndirect,
	{ClassJL, ClassPR}: BreakDirect,
	{ClassJL, ClassPR_EA}: BreakDirect,
	{ClassJL, ClassQU}: BreakIndirect,
	{ClassJL, ClassQU_Pf}: BreakProhibited,
	{ClassJL, ClassQU_Pi}: BreakIndirect,
	{ClassJL, ClassRI}: BreakDirect,
	{ClassJL, ClassSA}: BreakDirect,
	{ClassJL, ClassSP}: BreakProhibited,
	{ClassJL, ClassSY}: BreakProhibited,
	{ClassJL, ClassVF}: BreakDirect,
	{ClassJL, ClassVI}: BreakDirect,
	{ClassJL, ClassWJ}: BreakProhibited,
	{ClassJL, ClassXX}: BreakDirect,
	{ClassJL, ClassZW}: BreakProhibited,
	{ClassJL, ClassZWJ}: BreakIndirect,
	{ClassJT, ClassAI}: BreakDirect,
	{ClassJT, ClassAI_EA}: BreakDirect,
	{ClassJT, ClassAK}: BreakDirect,
	{ClassJT, ClassAL}: BreakDirect,
	{ClassJT, ClassAL_EA}: BreakDirect,
	{ClassJT, ClassAP}: BreakDirect,
	{ClassJT, ClassAS}: BreakDirect,
	{ClassJT, ClassB2}: BreakDirect,
	{ClassJT, ClassBA}: BreakIndirect,
	{ClassJT, ClassBA_EA}: BreakIndirect,
	{ClassJT, ClassBB}: BreakDirect,
	{ClassJT, ClassBK}: BreakProhibited,
	{ClassJT, ClassCB}: BreakDirect,
	{ClassJT, ClassCJ}: BreakIndirect,
	{ClassJT, ClassCL}: BreakProhibited,
	{ClassJT, ClassCL_EA}: BreakProhibited,
	{ClassJT, ClassCM}: BreakIndirect,
	{ClassJT, ClassCM_EA}: BreakIndirect,
	{ClassJT, ClassCP}: BreakProhibited,
	{ClassJT, ClassCR}: BreakProhibited,
	{ClassJT, ClassEB}: BreakDirect,
	{ClassJT, ClassEB_EA}: BreakDirect,
	{ClassJT, ClassEM}: BreakDirect,
	{ClassJT, ClassEX}: BreakProhibited,
	{ClassJT, ClassEX_EA}: BreakProhibited,
	{ClassJT, ClassGL}: BreakIndirect,
	{ClassJT, ClassGL_EA}: BreakIndirect,
	{ClassJT, ClassH2}: BreakDirect,
	{ClassJT, ClassH3}: BreakDirect,
	{ClassJT, ClassHH}: BreakIndirect,
	{ClassJT, ClassHL}: BreakDirect,
	{ClassJT, ClassHY}: BreakIndirect,
	{ClassJT, ClassID}: BreakDirect,
	{ClassJT, ClassID_EA}: BreakDirect,
	{ClassJT, ClassIN}: BreakIndirect,
	{ClassJT, ClassIN_EA}: BreakIndirect,
	{ClassJT, ClassIS}: BreakProhibited,
	{ClassJT, ClassJL}: BreakDirect,
	{ClassJT, ClassJT}: BreakIndirect,
	{ClassJT, ClassJV}: BreakDirect,
	{ClassJT, ClassLF}: BreakProhibited,
	{ClassJT, ClassNL}: BreakProhibited,
	{ClassJT, ClassNS}: BreakIndirect,
	{ClassJT, ClassNS_EA}: BreakIndirect,
	{ClassJT, ClassNU}: BreakDirect,
	{ClassJT, ClassOP}: BreakDirect,
	{ClassJT, ClassOP_EA}: BreakDirect,
	{ClassJT, ClassPO}: BreakIndirect,
	{ClassJT, ClassPO_EA}: BreakIndirect,
	{ClassJT, ClassPR}: BreakDirect,
	{ClassJT, ClassPR_EA}: BreakDirect,
	{ClassJT, ClassQU}: BreakIndirect,
	{ClassJT, ClassQU_Pf}: BreakProhibited,
	{ClassJT, ClassQU_Pi}: BreakIndirect,
	{ClassJT, ClassRI}: BreakDirect,
	{ClassJT, ClassSA}: BreakDirect,
	{ClassJT, ClassSP}: BreakProhibited,
	{ClassJT, ClassSY}: BreakProhibited,
	{ClassJT, ClassVF}: BreakDirect,
	{ClassJT, ClassVI}: BreakDirect,
	{ClassJT, ClassWJ}: BreakProhibited,
	{ClassJT, ClassXX}: BreakDirect,
	{ClassJT, ClassZW}: BreakProhibited,
	{ClassJT, ClassZWJ}: BreakIndirect,
	{ClassJV, ClassAI}: BreakDirect,
	{ClassJV, ClassAI_EA}: BreakDirect,
	{ClassJV, ClassAK}: BreakDirect,
	{ClassJV, ClassAL}: BreakDirect,
	{ClassJV, ClassAL_EA}: BreakDirect,
	{ClassJV, ClassAP}: BreakDirect,
	{ClassJV, ClassAS}: BreakDirect,
	{ClassJV, ClassB2}: BreakDirect,
	{ClassJV, ClassBA}: BreakIndirect,
	{ClassJV, ClassBA_EA}: BreakIndirect,
	{ClassJV, ClassBB}: BreakDirect,
	{ClassJV, ClassBK}: BreakProhibited,
	{ClassJV, ClassCB}: BreakDirect,
	{ClassJV, ClassCJ}: BreakIndirect,
	{ClassJV, ClassCL}: BreakProhibited,
	{ClassJV, ClassCL_EA}: BreakProhibited,
	{ClassJV, ClassCM}: BreakIndirect,
	{ClassJV, ClassCM_EA}: BreakIndirect,
	{ClassJV, ClassCP}: BreakProhibited,
	{ClassJV, ClassCR}: BreakProhibited,
	{ClassJV, ClassEB}: BreakDirect,
	{ClassJV, ClassEB_EA}: BreakDirect,
	{ClassJV, ClassEM}: BreakDirect,
	{ClassJV, ClassEX}: BreakProhibited,
	{ClassJV, ClassEX_EA}: BreakProhibited,
	{ClassJV, ClassGL}: BreakIndirect,
	{ClassJV, ClassGL_EA}: BreakIndirect,
	{ClassJV, ClassH2}: BreakDirect,
	{ClassJV, ClassH3}: BreakDirect,
	{ClassJV, ClassHH}: BreakIndirect,
	{ClassJV, ClassHL}: BreakDirect,
	{ClassJV, ClassHY}: BreakIndirect,
	{ClassJV, ClassID}: BreakDirect,
	{ClassJV, ClassID_EA}: BreakDirect,
	{ClassJV, ClassIN}: BreakIndirect,
	{ClassJV, ClassIN_EA}: BreakIndirect,
	{ClassJV, ClassIS}: BreakProhibited,
	{ClassJV, ClassJL}: BreakDirect,
	{ClassJV, ClassJT}: BreakIndirect,
	{ClassJV, ClassJV}: BreakIndirect,
	{ClassJV, ClassLF}: BreakProhibited,
	{ClassJV, ClassNL}: BreakProhibited,
	{ClassJV, ClassNS}: BreakIndirect,
	{ClassJV, ClassNS_EA}: BreakIndirect,
	{ClassJV, ClassNU}: BreakDirect,
	{ClassJV, ClassOP}: BreakDirect,
	{ClassJV, ClassOP_EA}: BreakDirect,
	{ClassJV, ClassPO}: BreakIndirect,
	{ClassJV, ClassPO_EA}: BreakIndirect,
	{ClassJV, ClassPR}: BreakDirect,
	{ClassJV, ClassPR_EA}: BreakDirect,
	{ClassJV, ClassQU}: BreakIndirect,
	{ClassJV, ClassQU_Pf}: BreakProhibited,
	{ClassJV, ClassQU_Pi}: BreakIndirect,
	{ClassJV, ClassRI}: BreakDirect,
	{ClassJV, ClassSA}: BreakDirect,
	{ClassJV, ClassSP}: BreakProhibited,
	{ClassJV, ClassSY}: BreakProhibited,
	{ClassJV, ClassVF}: BreakDirect,
	{ClassJV, ClassVI}: BreakDirect,
	{ClassJV, ClassWJ}: BreakProhibited,
	{ClassJV, ClassXX}: BreakDirect,
	{ClassJV, ClassZW}: BreakProhibited,
	{ClassJV, ClassZWJ}: BreakIndirect,
	{ClassNS, ClassAI}: BreakDirect,
	{ClassNS, ClassAI_EA}: BreakDirect,
	{ClassNS, ClassAK}: BreakDirect,
	{ClassNS, ClassAL}: BreakDirect,
	{ClassNS, ClassAL_EA}: BreakDirect,
	{ClassNS, ClassAP}: BreakDirect,
	{ClassNS, ClassAS}: BreakDirect,
	{ClassNS, ClassB2}: BreakDirect,
	{ClassNS, ClassBA}: BreakIndirect,
	{ClassNS, ClassBA_EA}: BreakIndirect,
	{ClassNS, ClassBB}: BreakDirect,
	{ClassNS, ClassBK}: BreakProhibited,
	{ClassNS, ClassCB}: BreakDirect,
	{ClassNS, ClassCJ}: BreakIndirect,
	{ClassNS, ClassCL}: BreakProhibited,
	{ClassNS, ClassCL_EA}: BreakProhibited,
	{ClassNS, ClassCM}: BreakIndirect,
	{ClassNS, ClassCM_EA}: BreakIndirect,
	{ClassNS, ClassCP}: BreakProhibited,
	{ClassNS, ClassCR}: BreakProhibited,
	{ClassNS, ClassEB}: BreakDirect,
	{ClassNS, ClassEB_EA}: BreakDirect,
	{ClassNS, ClassEM}: BreakDirect,
	{ClassNS, ClassEX}: BreakProhibited,
	{ClassNS, ClassEX_EA}: BreakProhibited,
	{ClassNS, ClassGL}: BreakIndirect,
	{ClassNS, ClassGL_EA}: BreakIndirect,
	{ClassNS, ClassH2}: BreakDirect,
	{ClassNS, ClassH3}: BreakDirect,
	{ClassNS, ClassHH}: BreakIndirect,
	{ClassNS, ClassHL}: BreakDirect,
	{ClassNS, ClassHY}: BreakIndirect,
	{ClassNS, ClassID}: BreakDirect,
	{ClassNS, ClassID_EA}: BreakDirect,
	{ClassNS, ClassIN}: BreakIndirect,
	{ClassNS, ClassIN_EA}: BreakIndirect,
	{ClassNS, ClassIS}: BreakProhibited,
	{ClassNS, ClassJL}: BreakDirect,
	{ClassNS, ClassJT}: BreakDirect,
	{ClassNS, ClassJV}: BreakDirect,
	{ClassNS, ClassLF}: BreakProhibited,
	{ClassNS, ClassNL}: BreakProhibited,
	{ClassNS, ClassNS}: BreakIndirect,
	{ClassNS, ClassNS_EA}: BreakIndirect,
	{ClassNS, ClassNU}: BreakDirect,
	{ClassNS, ClassOP}: BreakDirect,
	{ClassNS, ClassOP_EA}: BreakDirect,
	{ClassNS, ClassPO}: BreakDirect,
	{ClassNS, ClassPO_EA}: BreakDirect,
	{ClassNS, ClassPR}: BreakDirect,
	{ClassNS, ClassPR_EA}: BreakDirect,
	{ClassNS, ClassQU}: BreakIndirect,
	{ClassNS, ClassQU_Pf}: BreakProhibited,
	{ClassNS, ClassQU_Pi}: BreakIndirect,
	{ClassNS, ClassRI}: BreakDirect,
	{ClassNS, ClassSA}: BreakDirect,
	{ClassNS, ClassSP}: BreakProhibited,
	{ClassNS, ClassSY}: BreakProhibited,
	{ClassNS, ClassVF}: BreakDirect,
	{ClassNS, ClassVI}: BreakDirect,
	{ClassNS, ClassWJ}: BreakProhibited,
	{ClassNS, ClassXX}: BreakDirect,
	{ClassNS, ClassZW}: BreakProhibited,
	{ClassNS, ClassZWJ}: BreakIndirect,
	{ClassNS_EA, ClassAI}: BreakDirect,
	{ClassNS_EA, ClassAI_EA}: BreakDirect,
	{ClassNS_EA, ClassAK}: BreakDirect,
	{ClassNS_EA, ClassAL}: BreakDirect,
	{ClassNS_EA, ClassAL_EA}: BreakDirect,
	{ClassNS_EA, ClassAP}: BreakDirect,
	{ClassNS_EA, ClassAS}: BreakDirect,
	{ClassNS_EA, ClassB2}: BreakDirect,
	{ClassNS_EA, ClassBA}: BreakIndirect,
	{ClassNS_EA, ClassBA_EA}: BreakIndirect,
	{ClassNS_EA, ClassBB}: BreakDirect,
	{ClassNS_EA, ClassBK}: BreakProhibited,
	{ClassNS_EA, ClassCB}: BreakDirect,
	{ClassNS_EA, ClassCJ}: BreakIndirect,
	{ClassNS_EA, ClassCL}: BreakProhibited,
	{ClassNS_EA, ClassCL_EA}: BreakProhibited,
	{ClassNS_EA, ClassCM}: BreakIndirect,
	{ClassNS_EA, ClassCM_EA}: BreakIndirect,
	{ClassNS_EA, ClassCP}: BreakProhibited,
	{ClassNS_EA, ClassCR}: BreakProhibited,
	{ClassNS_EA, ClassEB}: BreakDirect,
	{ClassNS_EA, ClassEB_EA}: BreakDirect,
	{ClassNS_EA, ClassEM}: BreakDirect,
	{ClassNS_EA, ClassEX}: BreakProhibited,
	{ClassNS_EA, ClassEX_EA}: BreakProhibited,
	{ClassNS_EA, ClassGL}: BreakIndirect,
	{ClassNS_EA, ClassGL_EA}: BreakIndirect,
	{ClassNS_EA, ClassH2}: BreakDirect,
	{ClassNS_EA, ClassH3}: BreakDirect,
	{ClassNS_EA, ClassHH}: BreakIndirect,
	{ClassNS_EA, ClassHL}: BreakDirect,
	{ClassNS_EA, ClassHY}: BreakIndirect,
	{ClassNS_EA, ClassID}: BreakDirect,
	{ClassNS_EA, ClassID_EA}: BreakDirect,
	{ClassNS_EA, ClassIN}: BreakIndirect,
	{ClassNS_EA, ClassIN_EA}: BreakIndirect,
	{ClassNS_EA, ClassIS}: BreakProhibited,
	{ClassNS_EA, ClassJL}: BreakDirect,
	{ClassNS_EA, ClassJT}: BreakDirect,
	{ClassNS_EA, ClassJV}: BreakDirect,
	{ClassNS_EA, ClassLF}: BreakProhibited,
	{ClassNS_EA, ClassNL}: BreakProhibited,
	{ClassNS_EA, ClassNS}: BreakIndirect,
	{ClassNS_EA, ClassNS_EA}: BreakIndirect,
	{ClassNS_EA, ClassNU}: BreakDirect,
	{ClassNS_EA, ClassOP}: BreakDirect,
	{ClassNS_EA, ClassOP_EA}: BreakDirect,
	{ClassNS_EA, ClassPO}: BreakDirect,
	{ClassNS_EA, ClassPO_EA}: BreakDirect,
	{ClassNS_EA, ClassPR}: BreakDirect,
	{ClassNS_EA, ClassPR_EA}: BreakDirect,
	{ClassNS_EA, ClassQU}: BreakIndirect,
	{ClassNS_EA, ClassQU_Pf}: BreakProhibited,
	{ClassNS_EA, ClassQU_Pi}: BreakIndirect,
	{ClassNS_EA, ClassRI}: BreakDirect,
	{ClassNS_EA, ClassSA}: BreakDirect,
	{ClassNS_EA, ClassSP}: BreakProhibited,
	{ClassNS_EA, ClassSY}: BreakProhibited,
	{ClassNS_EA, ClassVF}: BreakDirect,
	{ClassNS_EA, ClassVI}: BreakDirect,
	{ClassNS_EA, ClassWJ}: BreakProhibited,
	{ClassNS_EA, ClassXX}: BreakDirect,
	{ClassNS_EA, ClassZW}: BreakProhibited,
	{ClassNS_EA, ClassZWJ}: BreakIndirect,
	{ClassNU, ClassAI}: BreakIndirect,
	{ClassNU, ClassAI_EA}: BreakIndirect,
	{ClassNU, ClassAK}: BreakDirect,
	{ClassNU, ClassAL}: BreakIndirect,
	{ClassNU, ClassAL_EA}: BreakIndirect,
	{ClassNU, ClassAP}: BreakDirect,
	{ClassNU, ClassAS}: BreakDirect,
	{ClassNU, ClassB2}: BreakDirect,
	{ClassNU, ClassBA}: BreakIndirect,
	{ClassNU, ClassBA_EA}: BreakIndirect,
	{ClassNU, ClassBB}: BreakDirect,
	{ClassNU, ClassBK}: BreakProhibited,
	{ClassNU, ClassCB}: BreakDirect,
	{ClassNU, ClassCJ}: BreakIndirect,
	{ClassNU, ClassCL}: BreakProhibited,
	{ClassNU, ClassCL_EA}: BreakProhibited,
	{ClassNU, ClassCM}: BreakIndirect,
	{ClassNU, ClassCM_EA}: BreakIndirect,
	{ClassNU, ClassCP}: BreakProhibited,
	{ClassNU, ClassCR}: BreakProhibited,
	{ClassNU, ClassEB}: BreakDirect,
	{ClassNU, ClassEB_EA}: BreakDirect,
	{ClassNU, ClassEM}: BreakDirect,
	{ClassNU, ClassEX}: BreakProhibited,
	{ClassNU, ClassEX_EA}: BreakProhibited,
	{ClassNU, ClassGL}: BreakIndirect,
	{ClassNU, ClassGL_EA}: BreakIndirect,
	{ClassNU, ClassH2}: BreakDirect,
	{ClassNU, ClassH3}: BreakDirect,
	{ClassNU, ClassHH}: BreakIndirect,
	{ClassNU, ClassHL}: BreakIndirect,
	{ClassNU, ClassHY}: BreakIndirect,
	{ClassNU, ClassID}: BreakDirect,
	{ClassNU, ClassID_EA}: BreakDirect,
	{ClassNU, ClassIN}: BreakIndirect,
	{ClassNU, ClassIN_EA}: BreakIndirect,
	{ClassNU, ClassIS}: BreakProhibited,
	{ClassNU, ClassJL}: BreakDirect,
	{ClassNU, ClassJT}: BreakDirect,
	{ClassNU, ClassJV}: BreakDirect,
	{ClassNU, ClassLF}: BreakProhibited,
	{ClassNU, ClassNL}: BreakProhibited,
	{ClassNU, ClassNS}: BreakIndirect,
	{ClassNU, ClassNS_EA}: BreakIndirect,
	{ClassNU, ClassNU}: BreakIndirect,
	{ClassNU, ClassOP}: BreakIndirect,
	{ClassNU, ClassOP_EA}: BreakDirect,
	{ClassNU, ClassPO}: BreakIndirect,
	{ClassNU, ClassPO_EA}: BreakIndirect,
	{ClassNU, ClassPR}: BreakIndirect,
	{ClassNU, ClassPR_EA}: BreakIndirect,
	{ClassNU, ClassQU}: BreakIndirect,
	{ClassNU, ClassQU_Pf}: BreakProhibited,
	{ClassNU, ClassQU_Pi}: BreakIndirect,
	{ClassNU, ClassRI}: BreakDirect,
	{ClassNU, ClassSA}: BreakIndirect,
	{ClassNU, ClassSP}: BreakProhibited,
	{ClassNU, ClassSY}: BreakProhibited,
	{ClassNU, ClassVF}: BreakDirect,
	{ClassNU, ClassVI}: BreakDirect,
	{ClassNU, ClassWJ}: BreakProhibited,
	{ClassNU, ClassXX}: BreakIndirect,
	{ClassNU, ClassZW}: BreakProhibited,
	{ClassNU, ClassZWJ}: BreakIndirect,
	{ClassOP, ClassAI}: BreakProhibited,
	{ClassOP, ClassAI_EA}: BreakProhibited,
	{ClassOP, ClassAK}: BreakProhibited,
	{ClassOP, ClassAL}: BreakProhibited,
	{ClassOP, ClassAL_EA}: BreakProhibited,
	{ClassOP, ClassAP}: BreakProhibited,
	{ClassOP, ClassAS}: BreakProhibited,
	{ClassOP, ClassB2}: BreakProhibited,
	{ClassOP, ClassBA}: BreakProhibited,
	{ClassOP, ClassBA_EA}: BreakProhibited,
	{ClassOP, ClassBB}: BreakProhibited,
	{ClassOP, ClassBK}: BreakProhibited,
	{ClassOP, ClassCB}: BreakProhibited,
	{ClassOP, ClassCJ}: BreakProhibited,
	{ClassOP, ClassCL}: BreakProhibited,
	{ClassOP, ClassCL_EA}: BreakProhibited,
	{ClassOP, ClassCM}: BreakProhibited,
	{ClassOP, ClassCM_EA}: BreakProhibited,
	{ClassOP, ClassCP}: BreakProhibited,
	{ClassOP, ClassCR}: BreakProhibited,
	{ClassOP, ClassEB}: BreakProhibited,
	{ClassOP, ClassEB_EA}: BreakProhibited,
	{ClassOP, ClassEM}: BreakProhibited,
	{ClassOP, ClassEX}: BreakProhibited,
	{ClassOP, ClassEX_EA}: BreakProhibited,
	{ClassOP, ClassGL}: BreakProhibited,
	{ClassOP, ClassGL_EA}: BreakProhibited,
	{ClassOP, ClassH2}: BreakProhibited,
	{ClassOP, ClassH3}: BreakProhibited,
	{ClassOP, ClassHH}: BreakProhibited,
	{ClassOP, ClassHL}: BreakProhibited,
	{ClassOP, ClassHY}: BreakProhibited,
	{ClassOP, ClassID}: BreakProhibited,
	{ClassOP, ClassID_EA}: BreakProhibited,
	{ClassOP, ClassIN}: BreakProhibited,
	{ClassOP, ClassIN_EA}: BreakProhibited,
	{ClassOP, ClassIS}: BreakProhibited,
	{ClassOP, ClassJL}: BreakProhibited,
	{ClassOP, ClassJT}: BreakProhibited,
	{ClassOP, ClassJV}: BreakProhibited,
	{ClassOP, ClassLF}: BreakProhibited,
	{ClassOP, ClassNL}: BreakProhibited,
	{ClassOP, ClassNS}: BreakProhibited,
	{ClassOP, ClassNS_EA}: BreakProhibited,
	{ClassOP, ClassNU}: BreakProhibited,
	{ClassOP, ClassOP}: BreakProhibited,
	{ClassOP, ClassOP_EA}: BreakProhibited,
	{ClassOP, ClassPO}: BreakProhibited,
	{ClassOP, ClassPO_EA}: BreakProhibited,
	{ClassOP, ClassPR}: BreakProhibited,
	{ClassOP, ClassPR_EA}: BreakProhibited,
	{ClassOP, ClassQU}: BreakProhibited,
	{ClassOP, ClassQU_Pf}: BreakProhibited,
	{ClassOP, ClassQU_Pi}: BreakProhibited,
	{ClassOP, ClassRI}: BreakProhibited,
	{ClassOP, ClassSA}: BreakProhibited,
	{ClassOP, ClassSP}: BreakProhibited,
	{ClassOP, ClassSY}: BreakProhibited,
	{ClassOP, ClassVF}: BreakProhibited,
	{ClassOP, ClassVI}: BreakProhibited,
	{ClassOP, ClassWJ}: BreakProhibited,
	{ClassOP, ClassXX}: BreakProhibited,
	{ClassOP, ClassZW}: BreakProhibited,
	{ClassOP, ClassZWJ}: BreakProhibited,
	{ClassOP_EA, ClassAI}: BreakProhibited,
	{ClassOP_EA, ClassAI_EA}: BreakProhibited,
	{ClassOP_EA, ClassAK}: BreakProhibited,
	{ClassOP_EA, ClassAL}: BreakProhibited,
	{ClassOP_EA, ClassAL_EA}: BreakProhibited,
	{ClassOP_EA, ClassAP}: BreakProhibited,
	{ClassOP_EA, ClassAS}: BreakProhibited,
	{ClassOP_EA, ClassB2}: BreakProhibited,
	{ClassOP_EA, ClassBA}: BreakProhibited,
	{ClassOP_EA, ClassBA_EA}: BreakProhibited,
	{ClassOP_EA, ClassBB}: BreakProhibited,
	{ClassOP_EA, ClassBK}: BreakProhibited,
	{ClassOP_EA, ClassCB}: BreakProhibited,
	{ClassOP_EA, ClassCJ}: BreakProhibited,
	{ClassOP_EA, ClassCL}: BreakProhibited,
	{ClassOP_EA, ClassCL_EA}: BreakProhibited,
	{ClassOP_EA, ClassCM}: BreakProhibited,
	{ClassOP_EA, ClassCM_EA}: BreakProhibited,
	{ClassOP_EA, ClassCP}: BreakProhibited,
	{ClassOP_EA, ClassCR}: BreakProhibited,
	{ClassOP_EA, ClassEB}: BreakProhibited,
	{ClassOP_EA, ClassEB_EA}: BreakProhibited,
	{ClassOP_EA, ClassEM}: BreakProhibited,
	{ClassOP_EA, ClassEX}: BreakProhibited,
	{ClassOP_EA, ClassEX_EA}: BreakProhibited,
	{ClassOP_EA, ClassGL}: BreakProhibited,
	{ClassOP_EA, ClassGL_EA}: BreakProhibited,
	{ClassOP_EA, ClassH2}: BreakProhibited,
	{ClassOP_EA, ClassH3}: BreakProhibited,
	{ClassOP_EA, ClassHH}: BreakProhibited,
	{ClassOP_EA, ClassHL}: BreakProhibited,
	{ClassOP_EA, ClassHY}: BreakProhibited,
	{ClassOP_EA, ClassID}: BreakProhibited,
	{ClassOP_EA, ClassID_EA}: BreakProhibited,
	{ClassOP_EA, ClassIN}: BreakProhibited,
	{ClassOP_EA, ClassIN_EA}: BreakProhibited,
	{ClassOP_EA, ClassIS}: BreakProhibited,
	{ClassOP_EA, ClassJL}: BreakProhibited,
	{ClassOP_EA, ClassJT}: BreakProhibited,
	{ClassOP_EA, ClassJV}: BreakProhibited,
	{ClassOP_EA, ClassLF}: BreakProhibited,
	{ClassOP_EA, ClassNL}: BreakProhibited,
	{ClassOP_EA, ClassNS}: BreakProhibited,
	{ClassOP_EA, ClassNS_EA}: BreakProhibited,
	{ClassOP_EA, ClassNU}: BreakProhibited,
	{ClassOP_EA, ClassOP}: BreakProhibited,
	{ClassOP_EA, ClassOP_EA}: BreakProhibited,
	{ClassOP_EA, ClassPO}: BreakProhibited,
	{ClassOP_EA, ClassPO_EA}: BreakProhibited,
	{ClassOP_EA, ClassPR}: BreakProhibited,
	{ClassOP_EA, ClassPR_EA}: BreakProhibited,
	{ClassOP_EA, ClassQU}: BreakProhibited,
	{ClassOP_EA, ClassQU_Pf}: BreakProhibited,
	{ClassOP_EA, ClassQU_Pi}: BreakProhibited,
	{ClassOP_EA, ClassRI}: BreakProhibited,
	{ClassOP_EA, ClassSA}: BreakProhibited,
	{ClassOP_EA, ClassSP}: BreakProhibited,
	{ClassOP_EA, ClassSY}: BreakProhibited,
	{ClassOP_EA, ClassVF}: BreakProhibited,
	{ClassOP_EA, ClassVI}: BreakProhibited,
	{ClassOP_EA, ClassWJ}: BreakProhibited,
	{ClassOP_EA, ClassXX}: BreakProhibited,
	{ClassOP_EA, ClassZW}: BreakProhibited,
	{ClassOP_EA, ClassZWJ}: BreakProhibited,
	{ClassPO, ClassAI}: BreakIndirect,
	{ClassPO, ClassAI_EA}: BreakIndirect,
	{ClassPO, ClassAK}: BreakDirect,
	{ClassPO, ClassAL}: BreakIndirect,
	{ClassPO, ClassAL_EA}: BreakIndirect,
	{ClassPO, ClassAP}: BreakDirect,
	{ClassPO, ClassAS}: BreakDirect,
	{ClassPO, ClassB2}: BreakDirect,
	{ClassPO, ClassBA}: BreakIndirect,
	{ClassPO, ClassBA_EA}: BreakIndirect,
	{ClassPO, ClassBB}: BreakDirect,
	{ClassPO, ClassBK}: BreakProhibited,
	{ClassPO, ClassCB}: BreakDirect,
	{ClassPO, ClassCJ}: BreakIndirect,
	{ClassPO, ClassCL}: BreakProhibited,
	{ClassPO, ClassCL_EA}: BreakProhibited,
	{ClassPO, ClassCM}: BreakIndirect,
	{ClassPO, ClassCM_EA}: BreakIndirect,
	{ClassPO, ClassCP}: BreakProhibited,
	{ClassPO, ClassCR}: BreakProhibited,
	{ClassPO, ClassEB}: BreakDirect,
	{ClassPO, ClassEB_EA}: BreakDirect,
	{ClassPO, ClassEM}: BreakDirect,
	{ClassPO, ClassEX}: BreakProhibited,
	{ClassPO, ClassEX_EA}: BreakProhibited,
	{ClassPO, ClassGL}: BreakIndirect,
	{ClassPO, ClassGL_EA}: BreakIndirect,
	{ClassPO, ClassH2}: BreakDirect,
	{ClassPO, ClassH3}: BreakDirect,
	{ClassPO, ClassHH}: BreakIndirect,
	{ClassPO, ClassHL}: BreakIndirect,
	{ClassPO, ClassHY}: BreakIndirect,
	{ClassPO, ClassID}: BreakDirect,
	{ClassPO, ClassID_EA}: BreakDirect,
	{ClassPO, ClassIN}: BreakIndirect,
	{ClassPO, ClassIN_EA}: BreakIndirect,
	{ClassPO, ClassIS}: BreakProhibited,
	{ClassPO, ClassJL}: BreakDirect,
	{ClassPO, ClassJT}: BreakDirect,
	{ClassPO, ClassJV}: BreakDirect,
	{ClassPO, ClassLF}: BreakProhibited,
	{ClassPO, ClassNL}: BreakProhibited,
	{ClassPO, ClassNS}: BreakIndirect,
	{ClassPO, ClassNS_EA}: BreakIndirect,
	{ClassPO, ClassNU}: BreakIndirect,
	{ClassPO, ClassOP}: BreakDirect,
	{ClassPO, ClassOP_EA}: BreakDirect,
	{ClassPO, ClassPO}: BreakDirect,
	{ClassPO, ClassPO_EA}: BreakDirect,
	{ClassPO, ClassPR}: BreakDirect,
	{ClassPO, ClassPR_EA}: BreakDirect,
	{ClassPO, ClassQU}: BreakIndirect,
	{ClassPO, ClassQU_Pf}: BreakProhibited,
	{ClassPO, ClassQU_Pi}: BreakIndirect,
	{ClassPO, ClassRI}: BreakDirect,
	{ClassPO, ClassSA}: BreakIndirect,
	{ClassPO, ClassSP}: BreakProhibited,
	{ClassPO, ClassSY}: BreakProhibited,
	{ClassPO, ClassVF}: BreakDirect,
	{ClassPO, ClassVI}: BreakDirect,
	{ClassPO, ClassWJ}: BreakProhibited,
	{ClassPO, ClassXX}: BreakIndirect,
	{ClassPO, ClassZW}: BreakProhibited,
	{ClassPO, ClassZWJ}: BreakIndirect,
	{ClassPO_EA, ClassAI}: BreakIndirect,
	{ClassPO_EA, ClassAI_EA}: BreakIndirect,
	{ClassPO_EA, ClassAK}: BreakDirect,
	{ClassPO_EA, ClassAL}: BreakIndirect,
	{ClassPO_EA, ClassAL_EA}: BreakIndirect,
	{ClassPO_EA, ClassAP}: BreakDirect,
	{ClassPO_EA, ClassAS}: BreakDirect,
	{ClassPO_EA, ClassB2}: BreakDirect,
	{ClassPO_EA, ClassBA}: BreakIndirect,
	{ClassPO_EA, ClassBA_EA}: BreakIndirect,
	{ClassPO_EA, ClassBB}: BreakDirect,
	{ClassPO_EA, ClassBK}: BreakProhibited,
	{ClassPO_EA, ClassCB}: BreakDirect,
	{ClassPO_EA, ClassCJ}: BreakIndirect,
	{ClassPO_EA, ClassCL}: BreakProhibited,
	{ClassPO_EA, ClassCL_EA}: BreakProhibited,
	{ClassPO_EA, ClassCM}: BreakIndirect,
	{ClassPO_EA, ClassCM_EA}: BreakIndirect,
	{ClassPO_EA, ClassCP}: BreakProhibited,
	{ClassPO_EA, ClassCR}: BreakProhibited,
	{ClassPO_EA, ClassEB}: BreakDirect,
	{ClassPO_EA, ClassEB_EA}: BreakDirect,
	{ClassPO_EA, ClassEM}: BreakDirect,
	{ClassPO_EA, ClassEX}: BreakProhibited,
	{ClassPO_EA, ClassEX_EA}: BreakProhibited,
	{ClassPO_EA, ClassGL}: BreakIndirect,
	{ClassPO_EA, ClassGL_EA}: BreakIndirect,
	{ClassPO_EA, ClassH2}: BreakDirect,
	{ClassPO_EA, ClassH3}: BreakDirect,
	{ClassPO_EA, ClassHH}: BreakIndirect,
	{ClassPO_EA, ClassHL}: BreakIndirect,
	{ClassPO_EA, ClassHY}: BreakIndirect,
	{ClassPO_EA, ClassID}: BreakDirect,
	{ClassPO_EA, ClassID_EA}: BreakDirect,
	{ClassPO_EA, ClassIN}: BreakIndirect,
	{ClassPO_EA, ClassIN_EA}: BreakIndirect,
	{ClassPO_EA, ClassIS}: BreakProhibited,
	{ClassPO_EA, ClassJL}: BreakDirect,
	{ClassPO_EA, ClassJT}: BreakDirect,
	{ClassPO_EA, ClassJV}: BreakDirect,
	{ClassPO_EA, ClassLF}: BreakProhibited,
	{ClassPO_EA, ClassNL}: BreakProhibited,
	{ClassPO_EA, ClassNS}: BreakIndirect,
	{ClassPO_EA, ClassNS_EA}: BreakIndirect,
	{ClassPO_EA, ClassNU}: BreakIndirect,
	{ClassPO_EA, ClassOP}: BreakDirect,
	{ClassPO_EA, ClassOP_EA}: BreakDirect,
	{ClassPO_EA, ClassPO}: BreakDirect,
	{ClassPO_EA, ClassPO_EA}: BreakDirect,
	{ClassPO_EA, ClassPR}: BreakDirect,
	{ClassPO_EA, ClassPR_EA}: BreakDirect,
	{ClassPO_EA, ClassQU}: BreakIndirect,
	{ClassPO_EA, ClassQU_Pf}: BreakProhibited,
	{ClassPO_EA, ClassQU_Pi}: BreakIndirect,
	{ClassPO_EA, ClassRI}: BreakDirect,
	{ClassPO_EA, ClassSA}: BreakIndirect,
	{ClassPO_EA, ClassSP}: BreakProhibited,
	{ClassPO_EA, ClassSY}: BreakProhibited,
	{ClassPO_EA, ClassVF}: BreakDirect,
	{ClassPO_EA, ClassVI}: BreakDirect,
	{ClassPO_EA, ClassWJ}: BreakProhibited,
	{ClassPO_EA, ClassXX}: BreakIndirect,
	{ClassPO_EA, ClassZW}: BreakProhibited,
	{ClassPO_EA, ClassZWJ}: BreakIndirect,
	{ClassPR, ClassAI}: BreakIndirect,
	{ClassPR, ClassAI_EA}: BreakIndirect,
	{ClassPR, ClassAK}: BreakDirect,
	{ClassPR, ClassAL}: BreakIndirect,
	{ClassPR, ClassAL_EA}: BreakIndirect,
	{ClassPR, ClassAP}: BreakDirect,
	{ClassPR, ClassAS}: BreakDirect,
	{ClassPR, ClassB2}: BreakDirect,
	{ClassPR, ClassBA}: BreakIndirect,
	{ClassPR, ClassBA_EA}: BreakIndirect,
	{ClassPR, ClassBB}: BreakDirect,
	{ClassPR, ClassBK}: BreakProhibited,
	{ClassPR, ClassCB}: BreakDirect,
	{ClassPR, ClassCJ}: BreakIndirect,
	{ClassPR, ClassCL}: BreakProhibited,
	{ClassPR, ClassCL_EA}: BreakProhibited,
	{ClassPR, ClassCM}: BreakIndirect,
	{ClassPR, ClassCM_EA}: BreakIndirect,
	{ClassPR, ClassCP}: BreakProhibited,
	{ClassPR, ClassCR}: BreakProhibited,
	{ClassPR, ClassEB}: BreakIndirect,
	{ClassPR, ClassEB_EA}: BreakIndirect,
	{ClassPR, ClassEM}: BreakIndirect,
	{ClassPR, ClassEX}: BreakProhibited,
	{ClassPR, ClassEX_EA}: BreakProhibited,
	{ClassPR, ClassGL}: BreakIndirect,
	{ClassPR, ClassGL_EA}: BreakIndirect,
	{ClassPR, ClassH2}: BreakIndirect,
	{ClassPR, ClassH3}: BreakIndirect,
	{ClassPR, ClassHH}: BreakIndirect,
	{ClassPR, ClassHL}: BreakIndirect,
	{ClassPR, ClassHY}: BreakIndirect,
	{ClassPR, ClassID}: BreakIndirect,
	{ClassPR, ClassID_EA}: BreakIndirect,
	{ClassPR, ClassIN}: BreakIndirect,
	{ClassPR, ClassIN_EA}: BreakIndirect,
	{ClassPR, ClassIS}: BreakProhibited,
	{ClassPR, ClassJL}: BreakIndirect,
	{ClassPR, ClassJT}: BreakIndirect,
	{ClassPR, ClassJV}: BreakIndirect,
	{ClassPR, ClassLF}: BreakProhibited,
	{ClassPR, ClassNL}: BreakProhibited,
	{ClassPR, ClassNS}: BreakIndirect,
	{ClassPR, ClassNS_EA}: BreakIndirect,
	{ClassPR, ClassNU}: BreakIndirect,
	{ClassPR, ClassOP}: BreakDirect,
	{ClassPR, ClassOP_EA}: BreakDirect,
	{ClassPR, ClassPO}: BreakDirect,
	{ClassPR, ClassPO_EA}: BreakDirect,
	{ClassPR, ClassPR}: BreakDirect,
	{ClassPR, ClassPR_EA}: BreakDirect,
	{ClassPR, ClassQU}: BreakIndirect,
	{ClassPR, ClassQU_Pf}: BreakProhibited,
	{ClassPR, ClassQU_Pi}: BreakIndirect,
	{ClassPR, ClassRI}: BreakDirect,
	{ClassPR, ClassSA}: BreakIndirect,
	{ClassPR, ClassSP}: BreakProhibited,
	{ClassPR, ClassSY}: BreakProhibited,
	{ClassPR, ClassVF}: BreakDirect,
	{ClassPR, ClassVI}: BreakDirect,
	{ClassPR, ClassWJ}: BreakProhibited,
	{ClassPR, ClassXX}: BreakIndirect,
	{ClassPR, ClassZW}: BreakProhibited,
	{ClassPR, ClassZWJ}: BreakIndirect,
	{ClassPR_EA, ClassAI}: BreakIndirect,
	{ClassPR_EA, ClassAI_EA}: BreakIndirect,
	{ClassPR_EA, ClassAK}: BreakDirect,
	{ClassPR_EA, ClassAL}: BreakIndirect,
	{ClassPR_EA, ClassAL_EA}: BreakIndirect,
	{ClassPR_EA, ClassAP}: BreakDirect,
	{ClassPR_EA, ClassAS}: BreakDirect,
	{ClassPR_EA, ClassB2}: BreakDirect,
	{ClassPR_EA, ClassBA}: BreakIndirect,
	{ClassPR_EA, ClassBA_EA}: BreakIndirect,
	{ClassPR_EA, ClassBB}: BreakDirect,
	{ClassPR_EA, ClassBK}: BreakProhibited,
	{ClassPR_EA, ClassCB}: BreakDirect,
	{ClassPR_EA, ClassCJ}: BreakIndirect,
	{ClassPR_EA, ClassCL}: BreakProhibited,
	{ClassPR_EA, ClassCL_EA}: BreakProhibited,
	{ClassPR_EA, ClassCM}: BreakIndirect,
	{ClassPR_EA, ClassCM_EA}: BreakIndirect,
	{ClassPR_EA, ClassCP}: BreakProhibited,
	{ClassPR_EA, ClassCR}: BreakProhibited,
	{ClassPR_EA, ClassEB}: BreakIndirect,
	{ClassPR_EA, ClassEB_EA}: BreakIndirect,
	{ClassPR_EA, ClassEM}: BreakIndirect,
	{ClassPR_EA, ClassEX}: BreakProhibited,
	{ClassPR_EA, ClassEX_EA}: BreakProhibited,
	{ClassPR_EA, ClassGL}: BreakIndirect,
	{ClassPR_EA, ClassGL_EA}: BreakIndirect,
	{ClassPR_EA, ClassH2}: BreakIndirect,
	{ClassPR_EA, ClassH3}: BreakIndirect,
	{ClassPR_EA, ClassHH}: BreakIndirect,
	{ClassPR_EA, ClassHL}: BreakIndirect,
	{ClassPR_EA, ClassHY}: BreakIndirect,
	{ClassPR_EA, ClassID}: BreakIndirect,
	{ClassPR_EA, ClassID_EA}: BreakIndirect,
	{ClassPR_EA, ClassIN}: BreakIndirect,
	{ClassPR_EA, ClassIN_EA}: BreakIndirect,
	{ClassPR_EA, ClassIS}: BreakProhibited,
	{ClassPR_EA, ClassJL}: BreakIndirect,
	{ClassPR_EA, ClassJT}: BreakIndirect,
	{ClassPR_EA, ClassJV}: BreakIndirect,
	{ClassPR_EA, ClassLF}: BreakProhibited,
	{ClassPR_EA, ClassNL}: BreakProhibited,
	{ClassPR_EA, ClassNS}: BreakIndirect,
	{ClassPR_EA, ClassNS_EA}: BreakIndirect,
	{ClassPR_EA, ClassNU}: BreakIndirect,
	{ClassPR_EA, ClassOP}: BreakDirect,
	{ClassPR_EA, ClassOP_EA}: BreakDirect,
	{ClassPR_EA, ClassPO}: BreakDirect,
	{ClassPR_EA, ClassPO_EA}: BreakDirect,
	{ClassPR_EA, ClassPR}: BreakDirect,
	{ClassPR_EA, ClassPR_EA}: BreakDirect,
	{ClassPR_EA, ClassQU}: BreakIndirect,
	{ClassPR_EA, ClassQU_Pf}: BreakProhibited,
	{ClassPR_EA, ClassQU_Pi}: BreakIndirect,
	{ClassPR_EA, ClassRI}: BreakDirect,
	{ClassPR_EA, ClassSA}: BreakIndirect,
	{ClassPR_EA, ClassSP}: BreakProhibited,
	{ClassPR_EA, ClassSY}: BreakProhibited,
	{ClassPR_EA, ClassVF}: BreakDirect,
	{ClassPR_EA, ClassVI}: BreakDirect,
	{ClassPR_EA, ClassWJ}: BreakProhibited,
	{ClassPR_EA, ClassXX}: BreakIndirect,
	{ClassPR_EA, ClassZW}: BreakProhibited,
	{ClassPR_EA, ClassZWJ}: BreakIndirect,
	{ClassQU, ClassAI}: BreakIndirect,
	{ClassQU, ClassAI_EA}: BreakIndirect,
	{ClassQU, ClassAK}: BreakIndirect,
	{ClassQU, ClassAL}: BreakIndirect,
	{ClassQU, ClassAL_EA}: BreakIndirect,
	{ClassQU, ClassAP}: BreakIndirect,
	{ClassQU, ClassAS}: BreakIndirect,
	{ClassQU, ClassB2}: BreakIndirect,
	{ClassQU, ClassBA}: BreakIndirect,
	{ClassQU, ClassBA_EA}: BreakIndirect,
	{ClassQU, ClassBB}: BreakIndirect,
	{ClassQU, ClassBK}: BreakProhibited,
	{ClassQU, ClassCB}: BreakIndirect,
	{ClassQU, ClassCJ}: BreakIndirect,
	{ClassQU, ClassCL}: BreakProhibited,
	{ClassQU, ClassCL_EA}: BreakProhibited,
	{ClassQU, ClassCM}: BreakIndirect,
	{ClassQU, ClassCM_EA}: BreakIndirect,
	{ClassQU, ClassCP}: BreakProhibited,
	{ClassQU, ClassCR}: BreakProhibited,
	{ClassQU, ClassEB}: BreakIndirect,
	{ClassQU, ClassEB_EA}: BreakIndirect,
	{ClassQU, ClassEM}: BreakIndirect,
	{ClassQU, ClassEX}: BreakProhibited,
	{ClassQU, ClassEX_EA}: BreakProhibited,
	{ClassQU, ClassGL}: BreakIndirect,
	{ClassQU, ClassGL_EA}: BreakIndirect,
	{ClassQU, ClassH2}: BreakIndirect,
	{ClassQU, ClassH3}: BreakIndirect,
	{ClassQU, ClassHH}: BreakIndirect,
	{ClassQU, ClassHL}: BreakIndirect,
	{ClassQU, ClassHY}: BreakIndirect,
	{ClassQU, ClassID}: BreakIndirect,
	{ClassQU, ClassID_EA}: BreakIndirect,
	{ClassQU, ClassIN}: BreakIndirect,
	{ClassQU, ClassIN_EA}: BreakIndirect,
	{ClassQU, ClassIS}: BreakProhibited,
	{ClassQU, ClassJL}: BreakIndirect,
	{ClassQU, ClassJT}: BreakIndirect,
	{ClassQU, ClassJV}: BreakIndirect,
	{ClassQU, ClassLF}: BreakProhibited,
	{ClassQU, ClassNL}: BreakProhibited,
	{ClassQU, ClassNS}: BreakIndirect,
	{ClassQU, ClassNS_EA}: BreakIndirect,
	{ClassQU, ClassNU}: BreakIndirect,
	{ClassQU, ClassOP}: BreakIndirect,
	{ClassQU, ClassOP_EA}: BreakIndirect,
	{ClassQU, ClassPO}: BreakIndirect,
	{ClassQU, ClassPO_EA}: BreakIndirect,
	{ClassQU, ClassPR}: BreakIndirect,
	{ClassQU, ClassPR_EA}: BreakIndirect,
	{ClassQU, ClassQU}: BreakIndirect,
	{ClassQU, ClassQU_Pf}: BreakProhibited,
	{ClassQU, ClassQU_Pi}: BreakIndirect,
	{ClassQU, ClassRI}: BreakIndirect,
	{ClassQU, ClassSA}: BreakIndirect,
	{ClassQU, ClassSP}: BreakProhibited,
	{ClassQU, ClassSY}: BreakProhibited,
	{ClassQU, ClassVF}: BreakIndirect,
	{ClassQU, ClassVI}: BreakIndirect,
	{ClassQU, ClassWJ}: BreakProhibited,
	{ClassQU, ClassXX}: BreakIndirect,
	{ClassQU, ClassZW}: BreakProhibited,
	{ClassQU, ClassZWJ}: BreakIndirect,
	{ClassQU_Pf, ClassAI}: BreakIndirect,
	{ClassQU_Pf, ClassAI_EA}: BreakIndirect,
	{ClassQU_Pf, ClassAK}: BreakIndirect,
	{ClassQU_Pf, ClassAL}: BreakIndirect,
	{ClassQU_Pf, ClassAL_EA}: BreakIndirect,
	{ClassQU_Pf, ClassAP}: BreakIndirect,
	{ClassQU_Pf, ClassAS}: BreakIndirect,
	{ClassQU_Pf, ClassB2}: BreakIndirect,
	{ClassQU_Pf, ClassBA}: BreakIndirect,
	{ClassQU_Pf, ClassBA_EA}: BreakIndirect,
	{ClassQU_Pf, ClassBB}: BreakIndirect,
	{ClassQU_Pf, ClassBK}: BreakProhibited,
	{ClassQU_Pf, ClassCB}: BreakIndirect,
	{ClassQU_Pf, ClassCJ}: BreakIndirect,
	{ClassQU_Pf, ClassCL}: BreakProhibited,
	{ClassQU_Pf, ClassCL_EA}: BreakProhibited,
	{ClassQU_Pf, ClassCM}: BreakIndirect,
	{ClassQU_Pf, ClassCM_EA}: BreakIndirect,
	{ClassQU_Pf, ClassCP}: BreakProhibited,
	{ClassQU_Pf, ClassCR}: BreakProhibited,
	{ClassQU_Pf, ClassEB}: BreakIndirect,
	{ClassQU_Pf, ClassEB_EA}: BreakIndirect,
	{ClassQU_Pf, ClassEM}: BreakIndirect,
	{ClassQU_Pf, ClassEX}: BreakProhibited,
	{ClassQU_Pf, ClassEX_EA}: BreakProhibited,
	{ClassQU_Pf, ClassGL}: BreakIndirect,
	{ClassQU_Pf, ClassGL_EA}: BreakIndirect,
	{ClassQU_Pf, ClassH2}: BreakIndirect,
	{ClassQU_Pf, ClassH3}: BreakIndirect,
	{ClassQU_Pf, ClassHH}: BreakIndirect,
	{ClassQU_Pf, ClassHL}: BreakIndirect,
	{ClassQU_Pf, ClassHY}: BreakIndirect,
	{ClassQU_Pf, ClassID}: BreakIndirect,
	{ClassQU_Pf, ClassID_EA}: BreakIndirect,
	{ClassQU_Pf, ClassIN}: BreakIndirect,
	{ClassQU_Pf, ClassIN_EA}: BreakIndirect,
	{ClassQU_Pf, ClassIS}: BreakProhibited,
	{ClassQU_Pf, ClassJL}: BreakIndirect,
	{ClassQU_Pf, ClassJT}: BreakIndirect,
	{ClassQU_Pf, ClassJV}: BreakIndirect,
	{ClassQU_Pf, ClassLF}: BreakProhibited,
	{ClassQU_Pf, ClassNL}: BreakProhibited,
	{ClassQU_Pf, ClassNS}: BreakIndirect,
	{ClassQU_Pf, ClassNS_EA}: BreakIndirect,
	{ClassQU_Pf, ClassNU}: BreakIndirect,
	{ClassQU_Pf, ClassOP}: BreakIndirect,
	{ClassQU_Pf, ClassOP_EA}: BreakIndirect,
	{ClassQU_Pf, ClassPO}: BreakIndirect,
	{ClassQU_Pf, ClassPO_EA}: BreakIndirect,
	{ClassQU_Pf, ClassPR}: BreakIndirect,
	{ClassQU_Pf, ClassPR_EA}: BreakIndirect,
	{ClassQU_Pf, ClassQU}: BreakIndirect,
	{ClassQU_Pf, ClassQU_Pf}: BreakProhibited,
	{ClassQU_Pf, ClassQU_Pi}: BreakIndirect,
	{ClassQU_Pf, ClassRI}: BreakIndirect,
	{ClassQU_Pf, ClassSA}: BreakIndirect,
	{ClassQU_Pf, ClassSP}: BreakProhibited,
	{ClassQU_Pf, ClassSY}: BreakProhibited,
	{ClassQU_Pf, ClassVF}: BreakIndirect,
	{ClassQU_Pf, ClassVI}: BreakIndirect,
	{ClassQU_Pf, ClassWJ}: BreakProhibited,
	{ClassQU_Pf, ClassXX}: BreakIndirect,
	{ClassQU_Pf, ClassZW}: BreakProhibited,
	{ClassQU_Pf, ClassZWJ}: BreakIndirect,
	{ClassQU_Pi, ClassAI}: BreakProhibited,
	{ClassQU_Pi, ClassAI_EA}: BreakProhibited,
	{ClassQU_Pi, ClassAK}: BreakProhibited,
	{ClassQU_Pi, ClassAL}: BreakProhibited,
	{ClassQU_Pi, ClassAL_EA}: BreakProhibited,
	{ClassQU_Pi, ClassAP}: BreakProhibited,
	{ClassQU_Pi, ClassAS}: BreakProhibited,
	{ClassQU_Pi, ClassB2}: BreakProhibited,
	{ClassQU_Pi, ClassBA}: BreakProhibited,
	{ClassQU_Pi, ClassBA_EA}: BreakProhibited,
	{ClassQU_Pi, ClassBB}: BreakProhibited,
	{ClassQU_Pi, ClassBK}: BreakProhibited,
	{ClassQU_Pi, ClassCB}: BreakProhibited,
	{ClassQU_Pi, ClassCJ}: BreakProhibited,
	{ClassQU_Pi, ClassCL}: BreakProhibited,
	{ClassQU_Pi, ClassCL_EA}: BreakProhibited,
	{ClassQU_Pi, ClassCM}: BreakProhibited,
	{ClassQU_Pi, ClassCM_EA}: BreakProhibited,
	{ClassQU_Pi, ClassCP}: BreakProhibited,
	{ClassQU_Pi, ClassCR}: BreakProhibited,
	{ClassQU_Pi, ClassEB}: BreakProhibited,
	{ClassQU_Pi, ClassEB_EA}: BreakProhibited,
	{ClassQU_Pi, ClassEM}: BreakProhibited,
	{ClassQU_Pi, ClassEX}: BreakProhibited,
	{ClassQU_Pi, ClassEX_EA}: BreakProhibited,
	{ClassQU_Pi, ClassGL}: BreakProhibited,
	{ClassQU_Pi, ClassGL_EA}: BreakProhibited,
	{ClassQU_Pi, ClassH2}: BreakProhibited,
	{ClassQU_Pi, ClassH3}: BreakProhibited,
	{ClassQU_Pi, ClassHH}: BreakProhibited,
	{ClassQU_Pi, ClassHL}: BreakProhibited,
	{ClassQU_Pi, ClassHY}: BreakProhibited,
	{ClassQU_Pi, ClassID}: BreakProhibited,
	{ClassQU_Pi, ClassID_EA}: BreakProhibited,
	{ClassQU_Pi, ClassIN}: BreakProhibited,
	{ClassQU_Pi, ClassIN_EA}: BreakProhibited,
	{ClassQU_Pi, ClassIS}: BreakProhibited,
	{ClassQU_Pi, ClassJL}: BreakProhibited,
	{ClassQU_Pi, ClassJT}: BreakProhibited,
	{ClassQU_Pi, ClassJV}: BreakProhibited,
	{ClassQU_Pi, ClassLF}: BreakProhibited,
	{ClassQU_Pi, ClassNL}: BreakProhibited,
	{ClassQU_Pi, ClassNS}: BreakProhibited,
	{ClassQU_Pi, ClassNS_EA}: BreakProhibited,
	{ClassQU_Pi, ClassNU}: BreakProhibited,
	{ClassQU_Pi, ClassOP}: BreakProhibited,
	{ClassQU_Pi, ClassOP_EA}: BreakProhibited,
	{ClassQU_Pi, ClassPO}: BreakProhibited,
	{ClassQU_Pi, ClassPO_EA}: BreakProhibited,
	{ClassQU_Pi, ClassPR}: BreakProhibited,
	{ClassQU_Pi, ClassPR_EA}: BreakProhibited,
	{ClassQU_Pi, ClassQU}: BreakProhibited,
	{ClassQU_Pi, ClassQU_Pf}: BreakProhibited,
	{ClassQU_Pi, ClassQU_Pi}: BreakProhibited,
	{ClassQU_Pi, ClassRI}: BreakProhibited,
	{ClassQU_Pi, ClassSA}: BreakProhibited,
	{ClassQU_Pi, ClassSP}: BreakProhibited,
	{ClassQU_Pi, ClassSY}: BreakProhibited,
	{ClassQU_Pi, ClassVF}: BreakProhibited,
	{ClassQU_Pi, ClassVI}: BreakProhibited,
	{ClassQU_Pi, ClassWJ}: BreakProhibited,
	{ClassQU_Pi, ClassXX}: BreakProhibited,
	{ClassQU_Pi, ClassZW}: BreakProhibited,
	{ClassQU_Pi, ClassZWJ}: BreakProhibited,
	{ClassRI, ClassAI}: BreakDirect,
	{ClassRI, ClassAI_EA}: BreakDirect,
	{ClassRI, ClassAK}: BreakDirect,
	{ClassRI, ClassAL}: BreakDirect,
	{ClassRI, ClassAL_EA}: BreakDirect,
	{ClassRI, ClassAP}: BreakDirect,
	{ClassRI, ClassAS}: BreakDirect,
	{ClassRI, ClassB2}: BreakDirect,
	{ClassRI, ClassBA}: BreakIndirect,
	{ClassRI, ClassBA_EA}: BreakIndirect,
	{ClassRI, ClassBB}: BreakDirect,
	{ClassRI, ClassBK}: BreakProhibited,
	{ClassRI, ClassCB}: BreakDirect,
	{ClassRI, ClassCJ}: BreakIndirect,
	{ClassRI, ClassCL}: BreakProhibited,
	{ClassRI, ClassCL_EA}: BreakProhibited,
	{ClassRI, ClassCM}: BreakIndirect,
	{ClassRI, ClassCM_EA}: BreakIndirect,
	{ClassRI, ClassCP}: BreakProhibited,
	{ClassRI, ClassCR}: BreakProhibited,
	{ClassRI, ClassEB}: BreakDirect,
	{ClassRI, ClassEB_EA}: BreakDirect,
	{ClassRI, ClassEM}: BreakDirect,
	{ClassRI, ClassEX}: BreakProhibited,
	{ClassRI, ClassEX_EA}: BreakProhibited,
	{ClassRI, ClassGL}: BreakIndirect,
	{ClassRI, ClassGL_EA}: BreakIndirect,
	{ClassRI, ClassH2}: BreakDirect,
	{ClassRI, ClassH3}: BreakDirect,
	{ClassRI, ClassHH}: BreakIndirect,
	{ClassRI, ClassHL}: BreakDirect,
	{ClassRI, ClassHY}: BreakIndirect,
	{ClassRI, ClassID}: BreakDirect,
	{ClassRI, ClassID_EA}: BreakDirect,
	{ClassRI, ClassIN}: BreakIndirect,
	{ClassRI, ClassIN_EA}: BreakIndirect,
	{ClassRI, ClassIS}: BreakProhibited,
	{ClassRI, ClassJL}: BreakDirect,
	{ClassRI, ClassJT}: BreakDirect,
	{ClassRI, ClassJV}: BreakDirect,
	{ClassRI, ClassLF}: BreakProhibited,
	{ClassRI, ClassNL}: BreakProhibited,
	{ClassRI, ClassNS}: BreakIndirect,
	{ClassRI, ClassNS_EA}: BreakIndirect,
	{ClassRI, ClassNU}: BreakDirect,
	{ClassRI, ClassOP}: BreakDirect,
	{ClassRI, ClassOP_EA}: BreakDirect,
	{ClassRI, ClassPO}: BreakDirect,
	{ClassRI, ClassPO_EA}: BreakDirect,
	{ClassRI, ClassPR}: BreakDirect,
	{ClassRI, ClassPR_EA}: BreakDirect,
	{ClassRI, ClassQU}: BreakIndirect,
	{ClassRI, ClassQU_Pf}: BreakProhibited,
	{ClassRI, ClassQU_Pi}: BreakIndirect,
	{ClassRI, ClassRI}: BreakIndirect,
	{ClassRI, ClassSA}: BreakDirect,
	{ClassRI, ClassSP}: BreakProhibited,
	{ClassRI, ClassSY}: BreakProhibited,
	{ClassRI, ClassVF}: BreakDirect,
	{ClassRI, ClassVI}: BreakDirect,
	{ClassRI, ClassWJ}: BreakProhibited,
	{ClassRI, ClassXX}: BreakDirect,
	{ClassRI, ClassZW}: BreakProhibited,
	{ClassRI, ClassZWJ}: BreakIndirect,
	{ClassSA, ClassAI}: BreakIndirect,
	{ClassSA, ClassAI_EA}: BreakIndirect,
	{ClassSA, ClassAK}: BreakDirect,
	{ClassSA, ClassAL}: BreakIndirect,
	{ClassSA, ClassAL_EA}: BreakIndirect,
	{ClassSA, ClassAP}: BreakDirect,
	{ClassSA, ClassAS}: BreakDirect,
	{ClassSA, ClassB2}: BreakDirect,
	{ClassSA, ClassBA}: BreakIndirect,
	{ClassSA, ClassBA_EA}: BreakIndirect,
	{ClassSA, ClassBB}: BreakDirect,
	{ClassSA, ClassBK}: BreakProhibited,
	{ClassSA, ClassCB}: BreakDirect,
	{ClassSA, ClassCJ}: BreakIndirect,
	{ClassSA, ClassCL}: BreakProhibited,
	{ClassSA, ClassCL_EA}: BreakProhibited,
	{ClassSA, ClassCM}: BreakIndirect,
	{ClassSA, ClassCM_EA}: BreakIndirect,
	{ClassSA, ClassCP}: BreakProhibited,
	{ClassSA, ClassCR}: BreakProhibited,
	{ClassSA, ClassEB}: BreakDirect,
	{ClassSA, ClassEB_EA}: BreakDirect,
	{ClassSA, ClassEM}: BreakDirect,
	{ClassSA, ClassEX}: BreakProhibited,
	{ClassSA, ClassEX_EA}: BreakProhibited,
	{ClassSA, ClassGL}: BreakIndirect,
	{ClassSA, ClassGL_EA}: BreakIndirect,
	{ClassSA, ClassH2}: BreakDirect,
	{ClassSA, ClassH3}: BreakDirect,
	{ClassSA, ClassHH}: BreakIndirect,
	{ClassSA, ClassHL}: BreakIndirect,
	{ClassSA, ClassHY}: BreakIndirect,
	{ClassSA, ClassID}: BreakDirect,
	{ClassSA, ClassID_EA}: BreakDirect,
	{ClassSA, ClassIN}: BreakIndirect,
	{ClassSA, ClassIN_EA}: BreakIndirect,
	{ClassSA, ClassIS}: BreakProhibited,
	{ClassSA, ClassJL}: BreakDirect,
	{ClassSA, ClassJT}: BreakDirect,
	{ClassSA, ClassJV}: BreakDirect,
	{ClassSA, ClassLF}: BreakProhibited,
	{ClassSA, ClassNL}: BreakProhibited,
	{ClassSA, ClassNS}: BreakIndirect,
	{ClassSA, ClassNS_EA}: BreakIndirect,
	{ClassSA, ClassNU}: BreakIndirect,
	{ClassSA, ClassOP}: BreakIndirect,
	{ClassSA, ClassOP_EA}: BreakDirect,
	{ClassSA, ClassPO}: BreakIndirect,
	{ClassSA, ClassPO_EA}: BreakIndirect,
	{ClassSA, ClassPR}: BreakIndirect,
	{ClassSA, ClassPR_EA}: BreakIndirect,
	{ClassSA, ClassQU}: BreakIndirect,
	{ClassSA, ClassQU_Pf}: BreakProhibited,
	{ClassSA, ClassQU_Pi}: BreakIndirect,
	{ClassSA, ClassRI}: BreakDirect,
	{ClassSA, ClassSA}: BreakIndirect,
	{ClassSA, ClassSP}: BreakProhibited,
	{ClassSA, ClassSY}: BreakProhibited,
	{ClassSA, ClassVF}: BreakDirect,
	{ClassSA, ClassVI}: BreakDirect,
	{ClassSA, ClassWJ}: BreakProhibited,
	{ClassSA, ClassXX}: BreakIndirect,
	{ClassSA, ClassZW}: BreakProhibited,
	{ClassSA, ClassZWJ}: BreakIndirect,
	{ClassSP, ClassAI}: BreakDirect,
	{ClassSP, ClassAI_EA}: BreakDirect,
	{ClassSP, ClassAK}: BreakDirect,
	{ClassSP, ClassAL}: BreakDirect,
	{ClassSP, ClassAL_EA}: BreakDirect,
	{ClassSP, ClassAP}: BreakDirect,
	{ClassSP, ClassAS}: BreakDirect,
	{ClassSP, ClassB2}: BreakDirect,
	{ClassSP, ClassBA}: BreakDirect,
	{ClassSP, ClassBA_EA}: BreakDirect,
	{ClassSP, ClassBB}: BreakDirect,
	{ClassSP, ClassBK}: BreakProhibited,
	{ClassSP, ClassCB}: BreakDirect,
	{ClassSP, ClassCJ}: BreakDirect,
	{ClassSP, ClassCL}: BreakProhibited,
	{ClassSP, ClassCL_EA}: BreakProhibited,
	{ClassSP, ClassCM}: BreakDirect,
	{ClassSP, ClassCM_EA}: BreakDirect,
	{ClassSP, ClassCP}: BreakProhibited,
	{ClassSP, ClassCR}: BreakProhibited,
	{ClassSP, ClassEB}: BreakDirect,
	{ClassSP, ClassEB_EA}: BreakDirect,
	{ClassSP, ClassEM}: BreakDirect,
	{ClassSP, ClassEX}: BreakProhibited,
	{ClassSP, ClassEX_EA}: BreakProhibited,
	{ClassSP, ClassGL}: BreakDirect,
	{ClassSP, ClassGL_EA}: BreakDirect,
	{ClassSP, ClassH2}: BreakDirect,
	{ClassSP, ClassH3}: BreakDirect,
	{ClassSP, ClassHH}: BreakDirect,
	{ClassSP, ClassHL}: BreakDirect,
	{ClassSP, ClassHY}: BreakDirect,
	{ClassSP, ClassID}: BreakDirect,
	{ClassSP, ClassID_EA}: BreakDirect,
	{ClassSP, ClassIN}: BreakDirect,
	{ClassSP, ClassIN_EA}: BreakDirect,
	{ClassSP, ClassIS}: BreakProhibited,
	{ClassSP, ClassJL}: BreakDirect,
	{ClassSP, ClassJT}: BreakDirect,
	{ClassSP, ClassJV}: BreakDirect,
	{ClassSP, ClassLF}: BreakProhibited,
	{ClassSP, ClassNL}: BreakProhibited,
	{ClassSP, ClassNS}: BreakDirect,
	{ClassSP, ClassNS_EA}: BreakDirect,
	{ClassSP, ClassNU}: BreakDirect,
	{ClassSP, ClassOP}: BreakDirect,
	{ClassSP, ClassOP_EA}: BreakDirect,
	{ClassSP, ClassPO}: BreakDirect,
	{ClassSP, ClassPO_EA}: BreakDirect,
	{ClassSP, ClassPR}: BreakDirect,
	{ClassSP, ClassPR_EA}: BreakDirect,
	{ClassSP, ClassQU}: BreakDirect,
	{ClassSP, ClassQU_Pf}: BreakProhibited,
	{ClassSP, ClassQU_Pi}: BreakDirect,
	{ClassSP, ClassRI}: BreakDirect,
	{ClassSP, ClassSA}: BreakDirect,
	{ClassSP, ClassSP}: BreakProhibited,
	{ClassSP, ClassSY}: BreakProhibited,
	{ClassSP, ClassVF}: BreakDirect,
	{ClassSP, ClassVI}: BreakDirect,
	{ClassSP, ClassWJ}: BreakProhibited,
	{ClassSP, ClassXX}: BreakDirect,
	{ClassSP, ClassZW}: BreakProhibited,
	{ClassSP, ClassZWJ}: BreakDirect,
	{ClassSY, ClassAI}: BreakDirect,
	{ClassSY, ClassAI_EA}: BreakDirect,
	{ClassSY, ClassAK}: BreakDirect,
	{ClassSY, ClassAL}: BreakDirect,
	{ClassSY, ClassAL_EA}: BreakDirect,
	{ClassSY, ClassAP}: BreakDirect,
	{ClassSY, ClassAS}: BreakDirect,
	{ClassSY, ClassB2}: BreakDirect,
	{ClassSY, ClassBA}: BreakIndirect,
	{ClassSY, ClassBA_EA}: BreakIndirect,
	{ClassSY, ClassBB}: BreakDirect,
	{ClassSY, ClassBK}: BreakProhibited,
	{ClassSY, ClassCB}: BreakDirect,
	{ClassSY, ClassCJ}: BreakIndirect,
	{ClassSY, ClassCL}: BreakProhibited,
	{ClassSY, ClassCL_EA}: BreakProhibited,
	{ClassSY, ClassCM}: BreakIndirect,
	{ClassSY, ClassCM_EA}: BreakIndirect,
	{ClassSY, ClassCP}: BreakProhibited,
	{ClassSY, ClassCR}: BreakProhibited,
	{ClassSY, ClassEB}: BreakDirect,
	{ClassSY, ClassEB_EA}: BreakDirect,
	{ClassSY, ClassEM}: BreakDirect,
	{ClassSY, ClassEX}: BreakProhibited,
	{ClassSY, ClassEX_EA}: BreakProhibited,
	{ClassSY, ClassGL}: BreakIndirect,
	{ClassSY, ClassGL_EA}: BreakIndirect,
	{ClassSY, ClassH2}: BreakDirect,
	{ClassSY, ClassH3}: BreakDirect,
	{ClassSY, ClassHH}: BreakIndirect,
	{ClassSY, ClassHL}: BreakIndirect,
	{ClassSY, ClassHY}: BreakIndirect,
	{ClassSY, ClassID}: BreakDirect,
	{ClassSY, ClassID_EA}: BreakDirect,
	{ClassSY, ClassIN}: BreakIndirect,
	{ClassSY, ClassIN_EA}: BreakIndirect,
	{ClassSY, ClassIS}: BreakProhibited,
	{ClassSY, ClassJL}: BreakDirect,
	{ClassSY, ClassJT}: BreakDirect,
	{ClassSY, ClassJV}: BreakDirect,
	{ClassSY, ClassLF}: BreakProhibited,
	{ClassSY, ClassNL}: BreakProhibited,
	{ClassSY, ClassNS}: BreakIndirect,
	{ClassSY, ClassNS_EA}: BreakIndirect,
	{ClassSY, ClassNU}: BreakDirect,
	{ClassSY, ClassOP}: BreakDirect,
	{ClassSY, ClassOP_EA}: BreakDirect,
	{ClassSY, ClassPO}: BreakDirect,
	{ClassSY, ClassPO_EA}: BreakDirect,
	{ClassSY, ClassPR}: BreakDirect,
	{ClassSY, ClassPR_EA}: BreakDirect,
	{ClassSY, ClassQU}: BreakIndirect,
	{ClassSY, ClassQU_Pf}: BreakProhibited,
	{ClassSY, ClassQU_Pi}: BreakIndirect,
	{ClassSY, ClassRI}: BreakDirect,
	{ClassSY, ClassSA}: BreakDirect,
	{ClassSY, ClassSP}: BreakProhibited,
	{ClassSY, ClassSY}: BreakProhibited,
	{ClassSY, ClassVF}: BreakDirect,
	{ClassSY, ClassVI}: BreakDirect,
	{ClassSY, ClassWJ}: BreakProhibited,
	{ClassSY, ClassXX}: BreakDirect,
	{ClassSY, ClassZW}: BreakProhibited,
	{ClassSY, ClassZWJ}: BreakIndirect,
	{ClassVF, ClassAI}: BreakDirect,
	{ClassVF, ClassAI_EA}: BreakDirect,
	{ClassVF, ClassAK}: BreakDirect,
	{ClassVF, ClassAL}: BreakDirect,
	{ClassVF, ClassAL_EA}: BreakDirect,
	{ClassVF, ClassAP}: BreakDirect,
	{ClassVF, ClassAS}: BreakDirect,
	{ClassVF, ClassB2}: BreakDirect,
	{ClassVF, ClassBA}: BreakIndirect,
	{ClassVF, ClassBA_EA}: BreakIndirect,
	{ClassVF, ClassBB}: BreakDirect,
	{ClassVF, ClassBK}: BreakProhibited,
	{ClassVF, ClassCB}: BreakDirect,
	{ClassVF, ClassCJ}: BreakIndirect,
	{ClassVF, ClassCL}: BreakProhibited,
	{ClassVF, ClassCL_EA}: BreakProhibited,
	{ClassVF, ClassCM}: BreakIndirect,
	{ClassVF, ClassCM_EA}: BreakIndirect,
	{ClassVF, ClassCP}: BreakProhibited,
	{ClassVF, ClassCR}: BreakProhibited,
	{ClassVF, ClassEB}: BreakDirect,
	{ClassVF, ClassEB_EA}: BreakDirect,
	{ClassVF, ClassEM}: BreakDirect,
	{ClassVF, ClassEX}: BreakProhibited,
	{ClassVF, ClassEX_EA}: BreakProhibited,
	{ClassVF, ClassGL}: BreakIndirect,
	{ClassVF, ClassGL_EA}: BreakIndirect,
	{ClassVF, ClassH2}: BreakDirect,
	{ClassVF, ClassH3}: BreakDirect,
	{ClassVF, ClassHH}: BreakIndirect,
	{ClassVF, ClassHL}: BreakDirect,
	{ClassVF, ClassHY}: BreakIndirect,
	{ClassVF, ClassID}: BreakDirect,
	{ClassVF, ClassID_EA}: BreakDirect,
	{ClassVF, ClassIN}: BreakIndirect,
	{ClassVF, ClassIN_EA}: BreakIndirect,
	{ClassVF, ClassIS}: BreakProhibited,
	{ClassVF, ClassJL}: BreakDirect,
	{ClassVF, ClassJT}: BreakDirect,
	{ClassVF, ClassJV}: BreakDirect,
	{ClassVF, ClassLF}: BreakProhibited,
	{ClassVF, ClassNL}: BreakProhibited,
	{ClassVF, ClassNS}: BreakIndirect,
	{ClassVF, ClassNS_EA}: BreakIndirect,
	{ClassVF, ClassNU}: BreakDirect,
	{ClassVF, ClassOP}: BreakDirect,
	{ClassVF, ClassOP_EA}: BreakDirect,
	{ClassVF, ClassPO}: BreakDirect,
	{ClassVF, ClassPO_EA}: BreakDirect,
	{ClassVF, ClassPR}: BreakDirect,
	{ClassVF, ClassPR_EA}: BreakDirect,
	{ClassVF, ClassQU}: BreakIndirect,
	{ClassVF, ClassQU_Pf}: BreakProhibited,
	{ClassVF, ClassQU_Pi}: BreakIndirect,
	{ClassVF, ClassRI}: BreakDirect,
	{ClassVF, ClassSA}: BreakDirect,
	{ClassVF, ClassSP}: BreakProhibited,
	{ClassVF, ClassSY}: BreakProhibited,
	{ClassVF, ClassVF}: BreakDirect,
	{ClassVF, ClassVI}: BreakDirect,
	{ClassVF, ClassWJ}: BreakProhibited,
	{ClassVF, ClassXX}: BreakDirect,
	{ClassVF, ClassZW}: BreakProhibited,
	{ClassVF, ClassZWJ}: BreakIndirect,
	{ClassVI, ClassAI}: BreakDirect,
	{ClassVI, ClassAI_EA}: BreakDirect,
	{ClassVI, ClassAK}: BreakDirect,
	{ClassVI, ClassAL}: BreakDirect,
	{ClassVI, ClassAL_EA}: BreakDirect,
	{ClassVI, ClassAP}: BreakDirect,
	{ClassVI, ClassAS}: BreakDirect,
	{ClassVI, ClassB2}: BreakDirect,
	{ClassVI, ClassBA}: BreakIndirect,
	{ClassVI, ClassBA_EA}: BreakIndirect,
	{ClassVI, ClassBB}: BreakDirect,
	{ClassVI, ClassBK}: BreakProhibited,
	{ClassVI, ClassCB}: BreakDirect,
	{ClassVI, ClassCJ}: BreakIndirect,
	{ClassVI, ClassCL}: BreakProhibited,
	{ClassVI, ClassCL_EA}: BreakProhibited,
	{ClassVI, ClassCM}: BreakIndirect,
	{ClassVI, ClassCM_EA}: BreakIndirect,
	{ClassVI, ClassCP}: BreakProhibited,
	{ClassVI, ClassCR}: BreakProhibited,
	{ClassVI, ClassEB}: BreakDirect,
	{ClassVI, ClassEB_EA}: BreakDirect,
	{ClassVI, ClassEM}: BreakDirect,
	{ClassVI, ClassEX}: BreakProhibited,
	{ClassVI, ClassEX_EA}: BreakProhibited,
	{ClassVI, ClassGL}: BreakIndirect,
	{ClassVI, ClassGL_EA}: BreakIndirect,
	{ClassVI, ClassH2}: BreakDirect,
	{ClassVI, ClassH3}: BreakDirect,
	{ClassVI, ClassHH}: BreakIndirect,
	{ClassVI, ClassHL}: BreakDirect,
	{ClassVI, ClassHY}: BreakIndirect,
	{ClassVI, ClassID}: BreakDirect,
	{ClassVI, ClassID_EA}: BreakDirect,
	{ClassVI, ClassIN}: BreakIndirect,
	{ClassVI, ClassIN_EA}: BreakIndirect,
	{ClassVI, ClassIS}: BreakProhibited,
	{ClassVI, ClassJL}: BreakDirect,
	{ClassVI, ClassJT}: BreakDirect,
	{ClassVI, ClassJV}: BreakDirect,
	{ClassVI, ClassLF}: BreakProhibited,
	{ClassVI, ClassNL}: BreakProhibited,
	{ClassVI, ClassNS}: BreakIndirect,
	{ClassVI, ClassNS_EA}: BreakIndirect,
	{ClassVI, ClassNU}: BreakDirect,
	{ClassVI, ClassOP}: BreakDirect,
	{ClassVI, ClassOP_EA}: BreakDirect,
	{ClassVI, ClassPO}: BreakDirect,
	{ClassVI, ClassPO_EA}: BreakDirect,
	{ClassVI, ClassPR}: BreakDirect,
	{ClassVI, ClassPR_EA}: BreakDirect,
	{ClassVI, ClassQU}: BreakIndirect,
	{ClassVI, ClassQU_Pf}: BreakProhibited,
	{ClassVI, ClassQU_Pi}: BreakIndirect,
	{ClassVI, ClassRI}: BreakDirect,
	{ClassVI, ClassSA}: BreakDirect,
	{ClassVI, ClassSP}: BreakProhibited,
	{ClassVI, ClassSY}: BreakProhibited,
	{ClassVI, ClassVF}: BreakDirect,
	{ClassVI, ClassVI}: BreakDirect,
	{ClassVI, ClassWJ}: BreakProhibited,
	{ClassVI, ClassXX}: BreakDirect,
	{ClassVI, ClassZW}: BreakProhibited,
	{ClassVI, ClassZWJ}: BreakIndirect,
	{ClassWJ, ClassAI}: BreakIndirect,
	{ClassWJ, ClassAI_EA}: BreakIndirect,
	{ClassWJ, ClassAK}: BreakIndirect,
	{ClassWJ, ClassAL}: BreakIndirect,
	{ClassWJ, ClassAL_EA}: BreakIndirect,
	{ClassWJ, ClassAP}: BreakIndirect,
	{ClassWJ, ClassAS}: BreakIndirect,
	{ClassWJ, ClassB2}: BreakIndirect,
	{ClassWJ, ClassBA}: BreakIndirect,
	{ClassWJ, ClassBA_EA}: BreakIndirect,
	{ClassWJ, ClassBB}: BreakIndirect,
	{ClassWJ, ClassBK}: BreakProhibited,
	{ClassWJ, ClassCB}: BreakIndirect,
	{ClassWJ, ClassCJ}: BreakIndirect,
	{ClassWJ, ClassCL}: BreakProhibited,
	{ClassWJ, ClassCL_EA}: BreakProhibited,
	{ClassWJ, ClassCM}: BreakIndirect,
	{ClassWJ, ClassCM_EA}: BreakIndirect,
	{ClassWJ, ClassCP}: BreakProhibited,
	{ClassWJ, ClassCR}: BreakProhibited,
	{ClassWJ, ClassEB}: BreakIndirect,
	{ClassWJ, ClassEB_EA}: BreakIndirect,
	{ClassWJ, ClassEM}: BreakIndirect,
	{ClassWJ, ClassEX}: BreakProhibited,
	{ClassWJ, ClassEX_EA}: BreakProhibited,
	{ClassWJ, ClassGL}: BreakIndirect,
	{ClassWJ, ClassGL_EA}: BreakIndirect,
	{ClassWJ, ClassH2}: BreakIndirect,
	{ClassWJ, ClassH3}: BreakIndirect,
	{ClassWJ, ClassHH}: BreakIndirect,
	{ClassWJ, ClassHL}: BreakIndirect,
	{ClassWJ, ClassHY}: BreakIndirect,
	{ClassWJ, ClassID}: BreakIndirect,
	{ClassWJ, ClassID_EA}: BreakIndirect,
	{ClassWJ, ClassIN}: BreakIndirect,
	{ClassWJ, ClassIN_EA}: BreakIndirect,
	{ClassWJ, ClassIS}: BreakProhibited,
	{ClassWJ, ClassJL}: BreakIndirect,
	{ClassWJ, ClassJT}: BreakIndirect,
	{ClassWJ, ClassJV}: BreakIndirect,
	{ClassWJ, ClassLF}: BreakProhibited,
	{ClassWJ, ClassNL}: BreakProhibited,
	{ClassWJ, ClassNS}: BreakIndirect,
	{ClassWJ, ClassNS_EA}: BreakIndirect,
	{ClassWJ, ClassNU}: BreakIndirect,
	{ClassWJ, ClassOP}: BreakIndirect,
	{ClassWJ, ClassOP_EA}: BreakIndirect,
	{ClassWJ, ClassPO}: BreakIndirect,
	{ClassWJ, ClassPO_EA}: BreakIndirect,
	{ClassWJ, ClassPR}: BreakIndirect,
	{ClassWJ, ClassPR_EA}: BreakIndirect,
	{ClassWJ, ClassQU}: BreakIndirect,
	{ClassWJ, ClassQU_Pf}: BreakProhibited,
	{ClassWJ, ClassQU_Pi}: BreakIndirect,
	{ClassWJ, ClassRI}: BreakIndirect,
	{ClassWJ, ClassSA}: BreakIndirect,
	{ClassWJ, ClassSP}: BreakProhibited,
	{ClassWJ, ClassSY}: BreakProhibited,
	{ClassWJ, ClassVF}: BreakIndirect,
	{ClassWJ, ClassVI}: BreakIndirect,
	{ClassWJ, ClassWJ}: BreakProhibited,
	{ClassWJ, ClassXX}: BreakIndirect,
	{ClassWJ, ClassZW}: BreakProhibited,
	{ClassWJ, ClassZWJ}: BreakIndirect,
	{ClassZW, ClassAI}: BreakDirect,
	{ClassZW, ClassAI_EA}: BreakDirect,
	{ClassZW, ClassAK}: BreakDirect,
	{ClassZW, ClassAL}: BreakDirect,
	{ClassZW, ClassAL_EA}: BreakDirect,
	{ClassZW, ClassAP}: BreakDirect,
	{ClassZW, ClassAS}: BreakDirect,
	{ClassZW, ClassB2}: BreakDirect,
	{ClassZW, ClassBA}: BreakDirect,
	{ClassZW, ClassBA_EA}: BreakDirect,
	{ClassZW, ClassBB}: BreakDirect,
	{ClassZW, ClassBK}: BreakProhibited,
	{ClassZW, ClassCB}: BreakDirect,
	{ClassZW, ClassCJ}: BreakDirect,
	{ClassZW, ClassCL}: BreakDirect,
	{ClassZW, ClassCL_EA}: BreakDirect,
	{ClassZW, ClassCM}: BreakDirect,
	{ClassZW, ClassCM_EA}: BreakDirect,
	{ClassZW, ClassCP}: BreakDirect,
	{ClassZW, ClassCR}: BreakProhibited,
	{ClassZW, ClassEB}: BreakDirect,
	{ClassZW, ClassEB_EA}: BreakDirect,
	{ClassZW, ClassEM}: BreakDirect,
	{ClassZW, ClassEX}: BreakDirect,
	{ClassZW, ClassEX_EA}: BreakDirect,
	{ClassZW, ClassGL}: BreakDirect,
	{ClassZW, ClassGL_EA}: BreakDirect,
	{ClassZW, ClassH2}: BreakDirect,
	{ClassZW, ClassH3}: BreakDirect,
	{ClassZW, ClassHH}: BreakDirect,
	{ClassZW, ClassHL}: BreakDirect,
	{ClassZW, ClassHY}: BreakDirect,
	{ClassZW, ClassID}: BreakDirect,
	{ClassZW, ClassID_EA}: BreakDirect,
	{ClassZW, ClassIN}: BreakDirect,
	{ClassZW, ClassIN_EA}: BreakDirect,
	{ClassZW, ClassIS}: BreakDirect,
	{ClassZW, ClassJL}: BreakDirect,
	{ClassZW, ClassJT}: BreakDirect,
	{ClassZW, ClassJV}: BreakDirect,
	{ClassZW, ClassLF}: BreakProhibited,
	{ClassZW, ClassNL}: BreakProhibited,
	{ClassZW, ClassNS}: BreakDirect,
	{ClassZW, ClassNS_EA}: BreakDirect,
	{ClassZW, ClassNU}: BreakDirect,
	{ClassZW, ClassOP}: BreakDirect,
	{ClassZW, ClassOP_EA}: BreakDirect,
	{ClassZW, ClassPO}: BreakDirect,
	{ClassZW, ClassPO_EA}: BreakDirect,
	{ClassZW, ClassPR}: BreakDirect,
	{ClassZW, ClassPR_EA}: BreakDirect,
	{ClassZW, ClassQU}: BreakDirect,
	{ClassZW, ClassQU_Pf}: BreakDirect,
	{ClassZW, ClassQU_Pi}: BreakDirect,
	{ClassZW, ClassRI}: BreakDirect,
	{ClassZW, ClassSA}: BreakDirect,
	{ClassZW, ClassSP}: BreakProhibited,
	{ClassZW, ClassSY}: BreakDirect,
	{ClassZW, ClassVF}: BreakDirect,
	{ClassZW, ClassVI}: BreakDirect,
	{ClassZW, ClassWJ}: BreakDirect,
	{ClassZW, ClassXX}: BreakDirect,
	{ClassZW, ClassZW}: BreakProhibited,
	{ClassZW, ClassZWJ}: BreakDirect,
	{ClassZWJ, ClassAI}: BreakIndirect,
	{ClassZWJ, ClassAI_EA}: BreakIndirect,
	{ClassZWJ, ClassAK}: BreakIndirect,
	{ClassZWJ, ClassAL}: BreakIndirect,
	{ClassZWJ, ClassAL_EA}: BreakIndirect,
	{ClassZWJ, ClassAP}: BreakIndirect,
	{ClassZWJ, ClassAS}: BreakIndirect,
	{ClassZWJ, ClassB2}: BreakIndirect,
	{ClassZWJ, ClassBA}: BreakIndirect,
	{ClassZWJ, ClassBA_EA}: BreakIndirect,
	{ClassZWJ, ClassBB}: BreakIndirect,
	{ClassZWJ, ClassBK}: BreakProhibited,
	{ClassZWJ, ClassCB}: BreakIndirect,
	{ClassZWJ, ClassCJ}: BreakIndirect,
	{ClassZWJ, ClassCL}: BreakProhibited,
	{ClassZWJ, ClassCL_EA}: BreakProhibited,
	{ClassZWJ, ClassCM}: BreakIndirect,
	{ClassZWJ, ClassCM_EA}: BreakIndirect,
	{ClassZWJ, ClassCP}: BreakProhibited,
	{ClassZWJ, ClassCR}: BreakProhibited,
	{ClassZWJ, ClassEB}: BreakIndirect,
	{ClassZWJ, ClassEB_EA}: BreakIndirect,
	{ClassZWJ, ClassEM}: BreakIndirect,
	{ClassZWJ, ClassEX}: BreakProhibited,
	{ClassZWJ, ClassEX_EA}: BreakProhibited,
	{ClassZWJ, ClassGL}: BreakIndirect,
	{ClassZWJ, ClassGL_EA}: BreakIndirect,
	{ClassZWJ, ClassH2}: BreakIndirect,
	{ClassZWJ, ClassH3}: BreakIndirect,
	{ClassZWJ, ClassHH}: BreakIndirect,
	{ClassZWJ, ClassHL}: BreakIndirect,
	{ClassZWJ, ClassHY}: BreakIndirect,
	{ClassZWJ, ClassID}: BreakIndirect,
	{ClassZWJ, ClassID_EA}: BreakIndirect,
	{ClassZWJ, ClassIN}: BreakIndirect,
	{ClassZWJ, ClassIN_EA}: BreakIndirect,
	{ClassZWJ, ClassIS}: BreakProhibited,
	{ClassZWJ, ClassJL}: BreakIndirect,
	{ClassZWJ, ClassJT}: BreakIndirect,
	{ClassZWJ, ClassJV}: BreakIndirect,
	{ClassZWJ, ClassLF}: BreakProhibited,
	{ClassZWJ, ClassNL}: BreakProhibited,
	{ClassZWJ, ClassNS}: BreakIndirect,
	{ClassZWJ, ClassNS_EA}: BreakIndirect,
	{ClassZWJ, ClassNU}: BreakIndirect,
	{ClassZWJ, ClassOP}: BreakIndirect,
	{ClassZWJ, ClassOP_EA}: BreakIndirect,
	{ClassZWJ, ClassPO}: BreakIndirect,
	{ClassZWJ, ClassPO_EA}: BreakIndirect,
	{ClassZWJ, ClassPR}: BreakIndirect,
	{ClassZWJ, ClassPR_EA}: BreakIndirect,
	{ClassZWJ, ClassQU}: BreakIndirect,
	{ClassZWJ, ClassQU_Pf}: BreakProhibited,
	{ClassZWJ, ClassQU_Pi}: BreakIndirect,
	{ClassZWJ, ClassRI}: BreakIndirect,
	{ClassZWJ, ClassSA}: BreakIndirect,
	{ClassZWJ, ClassSP}: BreakProhibited,
	{ClassZWJ, ClassSY}: BreakProhibited,
	{ClassZWJ, ClassVF}: BreakIndirect,
	{ClassZWJ, ClassVI}: BreakIndirect,
	{ClassZWJ, ClassWJ}: BreakProhibited,
	{ClassZWJ, ClassXX}: BreakIndirect,
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
//
// This function implements line break detection based on UAX #14 (Unicode Line Breaking
// Algorithm). It returns a slice of byte positions where line breaks are allowed,
// enabling text layout systems to wrap text at appropriate boundaries.
//
// The algorithm handles:
//   - Mandatory breaks (LB4-LB6): Newlines, hard breaks, paragraph separators
//   - Word boundaries (LB18): Spaces between words
//   - Ideographic breaks (LB30a): Breaks between CJK characters
//   - Hyphenation (LB21-LB22): Configurable hyphenation behavior
//   - Punctuation: Proper handling of quotes, parentheses, and other marks
//   - Numeric sequences: Keeping numbers with units and separators together
//
// The hyphens parameter controls hyphenation behavior per CSS Text Module Level 3:
//   - HyphensNone: No breaks at hyphens (hard or soft)
//   - HyphensManual: Only break at U+00AD soft hyphens (https://www.w3.org/TR/css-text-3/#valdef-hyphens-manual)
//   - HyphensAuto: Break at all hyphens (dictionary-based hyphenation not yet fully implemented)
//
// The returned slice always includes position 0 (start) and len(text) (end).
// All positions are byte offsets, not rune indices, for direct string slicing.
//
// Example:
//
//	text := "Hello world! This is a test."
//	breaks := FindLineBreakOpportunities(text, HyphensManual)
//	// Returns: [0, 6, 13, 16, 19, 21, 26, 27] - breaks at spaces and end
//
//	text = "Hello­world"  // Contains soft hyphen (U+00AD)
//	breaks = FindLineBreakOpportunities(text, HyphensManual)
//	// Break allowed at soft hyphen position
//
//	text = "中文测试"  // Chinese text
//	breaks = FindLineBreakOpportunities(text, HyphensNone)
//	// Breaks allowed between each ideographic character
//
// See UAX #14: https://www.unicode.org/reports/tr14/
//
// Implementation notes:
//   - Based on UAX #14 with focus on practical word-boundary breaking
//   - Handles mandatory breaks per LB4-LB6: https://www.unicode.org/reports/tr14/#LB4
//   - Supports CJK ideographic breaking per LB30a: https://www.unicode.org/reports/tr14/#LB30a
//   - Hyphenation follows CSS Text Level 3 §4.3: https://www.w3.org/TR/css-text-3/#hyphenation
//   - Originally from github.com/SCKelemen/layout, extracted for reusability
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
	// LB10: Treat CM or ZWJ at start of text as AL
	if isClassOrVariant(prevClass, ClassCM) || prevClass == ClassZWJ {
		prevClass = ClassAL
	}
	lastNonSpaceClass := prevClass // Track last non-SP class for LB14

	for i := 1; i < len(runes); i++ {
		currClass := getBreakClass(runes[i])

		// LB9: SA characters that are combining marks (Mn, Mc) should be treated as CM
		// The official data distinguishes SA_Mn and SA_Mc, but our table consolidates them
		if currClass == ClassSA {
			r := runes[i]
			if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r) {
				currClass = ClassCM
			}
		}

		// LB4, LB5: Mandatory breaks - handle BEFORE consulting pair table
		// Always break after BK, CR (except before LF), LF, NL
		if prevClass == ClassBK || prevClass == ClassLF || prevClass == ClassNL {
			bytePos := len(string(runes[:i]))
			breakPoints = append(breakPoints, bytePos)
			// LB10: Treat CM or ZWJ following a mandatory break as AL
			if isClassOrVariant(currClass, ClassCM) || currClass == ClassZWJ {
				prevClass = ClassAL
				lastNonSpaceClass = ClassAL
			} else {
				prevClass = currClass
				if currClass != ClassSP {
					lastNonSpaceClass = currClass
				}
			}
			continue
		}

		// LB5: Treat CR × LF as unbreakable (don't break within CR LF sequence)
		if prevClass == ClassCR {
			if currClass == ClassLF {
				// Don't break within CR LF - treat as single unit
				prevClass = currClass
				if currClass != ClassSP {
					lastNonSpaceClass = currClass
				}
				continue
			} else {
				// CR followed by non-LF: mandatory break
				bytePos := len(string(runes[:i]))
				breakPoints = append(breakPoints, bytePos)
				// LB10: Treat CM or ZWJ following a mandatory break as AL
				if isClassOrVariant(currClass, ClassCM) || currClass == ClassZWJ {
					prevClass = ClassAL
					lastNonSpaceClass = ClassAL
				} else {
					prevClass = currClass
					if currClass != ClassSP {
						lastNonSpaceClass = currClass
					}
				}
				continue
			}
		}

		action := getBreakAction(prevClass, currClass)

		// LB8a: Do not break after ZWJ (Zero Width Joiner)
		// Check if the actual previous rune is ZWJ (even if prevClass was converted to AL by LB10)
		if i > 0 && runes[i-1] == '\u200D' { // U+200D is ZWJ
			// Do not break after ZWJ - skip to updating prevClass
			if !isClassOrVariant(currClass, ClassCM) && currClass != ClassZWJ && currClass != ClassSA {
				prevClass = currClass
				if currClass != ClassSP {
					lastNonSpaceClass = currClass
				}
			} else if (isClassOrVariant(currClass, ClassCM) || currClass == ClassZWJ) &&
				(prevClass == ClassSP || prevClass == ClassZW) {
				prevClass = ClassAL
				lastNonSpaceClass = ClassAL
			}
			continue
		}

		// LB8: Break before any character following ZW, even with intervening spaces
		// This overrides pair table (e.g., SP × CL is BreakProhibited, but ZW SP CL should break after SP)
		// Exception: Do not break before mandatory breaks (BK/CR/LF/NL) or ZW (LB6, LB7)
		if lastNonSpaceClass == ClassZW && prevClass == ClassSP && currClass != ClassSP &&
			currClass != ClassBK && currClass != ClassCR && currClass != ClassLF &&
			currClass != ClassNL && currClass != ClassZW {
			bytePos := len(string(runes[:i]))
			breakPoints = append(breakPoints, bytePos)
			// Skip the switch statement to avoid adding duplicate breaks
			// Update prevClass and continue
			if !isClassOrVariant(currClass, ClassCM) && currClass != ClassZWJ && currClass != ClassSA {
				prevClass = currClass
				if currClass != ClassSP {
					lastNonSpaceClass = currClass
				}
			} else if (isClassOrVariant(currClass, ClassCM) || currClass == ClassZWJ) &&
				(prevClass == ClassSP || prevClass == ClassZW) {
				prevClass = ClassAL
				lastNonSpaceClass = ClassAL
			}
			continue
		}

		// LB21: Special handling for HY (hyphen-minus)
		// HY generally allows breaks after it, with specific exceptions
		// Patterns: AL × HY ÷ AL, CP × HY ÷, CL × HY ÷, HL × HY ÷ HL
		if prevClass == ClassHY && i >= 2 {
			// Check what comes before the HY
			prevPrevRune := runes[i-2]
			prevPrevClass := getBreakClass(prevPrevRune)
			shouldBreak := false

			if isClassOrVariant(prevPrevClass, ClassCP) || isClassOrVariant(prevPrevClass, ClassCL) {
				// CP × HY ÷, CL × HY ÷ - allow break after HY
				shouldBreak = true
			} else if prevPrevClass == ClassHL && currClass == ClassHL {
				// HL × HY ÷ HL (Hebrew letter, hyphen, Hebrew letter) - allow break after HY
				shouldBreak = true
			} else if isClassOrVariant(prevPrevClass, ClassAL) && isClassOrVariant(currClass, ClassAL) {
				// AL × HY ÷ AL - regular hyphenated words like "Excusez-moi"
				shouldBreak = true
			} else if prevPrevClass == ClassHY && isClassOrVariant(currClass, ClassAL) {
				// HY × HY ÷ AL - check if this follows CP/CL in the context
				// Pattern: CP/CL × ... × AL × HY × HY ÷ AL (like "(http://)xn--a" or "{http://}xn--a")
				checkIdx := i - 3
				for checkIdx >= 0 {
					checkRune := runes[checkIdx]
					checkClass := getBreakClass(checkRune)
					if checkClass == ClassSP || isClassOrVariant(checkClass, ClassCM) || checkClass == ClassZWJ || isClassOrVariant(checkClass, ClassAL) {
						// Skip spaces, combining marks, and AL characters
						checkIdx--
						continue
					}
					// Check if we find CP or CL
					if isClassOrVariant(checkClass, ClassCP) || isClassOrVariant(checkClass, ClassCL) {
						shouldBreak = true
					}
					break
				}
			}

			if shouldBreak && currClass != ClassSP && currClass != ClassZW && currClass != ClassCM {
				bytePos := len(string(runes[:i]))
				breakPoints = append(breakPoints, bytePos)
				prevClass = currClass
				if currClass != ClassSP {
					lastNonSpaceClass = currClass
				}
				continue
			}
		}

		// LB21.02: Special handling for HH (Hebrew hyphen/MAQAF)
		// HL × HH ÷ HL - Allow break after HH when preceded by HL or AL
		if prevClass == ClassHH && currClass == ClassHL && i >= 2 {
			// Check if HH is preceded by HL or AL (not at start of text)
			prevPrevIdx := i - 2
			foundPreceding := false
			for prevPrevIdx >= 0 {
				prevPrevRune := runes[prevPrevIdx]
				prevPrevClass := getBreakClass(prevPrevRune)
				// Skip combining marks
				if isClassOrVariant(prevPrevClass, ClassCM) || prevPrevClass == ClassZWJ {
					prevPrevIdx--
					continue
				}
				// If HH is preceded by HL or AL, allow break
				if prevPrevClass == ClassHL || isClassOrVariant(prevPrevClass, ClassAL) {
					foundPreceding = true
				}
				break
			}

			if foundPreceding && currClass != ClassSP && currClass != ClassZW && currClass != ClassCM {
				bytePos := len(string(runes[:i]))
				breakPoints = append(breakPoints, bytePos)
				prevClass = currClass
				if currClass != ClassSP {
					lastNonSpaceClass = currClass
				}
				continue
			}
		}

		// LB19: Quotation marks with context-sensitive breaking
		// Handle multiple patterns:
		// 1. CP/CL/EX/IS/SY × SP ÷ QU_Pf - Break before closing quote after punctuation
		// 2. QU_Pi × SP ÷ OP - Break after opening quote before opening punctuation (when quote is closed)
		// 3. NS ÷ QU_Pi - Break before opening quote after non-starter (CJK text)

		// Pattern 3: FULLWIDTH COLON ÷ QU_Pi (CJK: break after FULLWIDTH COLON before opening quote)
		if isClassOrVariant(prevClass, ClassNS) && isClassOrVariant(currClass, ClassQU_Pi) && i > 0 {
			// Only apply to FULLWIDTH COLON (U+FF1A), not all NS characters
			prevRune := runes[i-1]
			if prevRune == '\uFF1A' { // FULLWIDTH COLON
				// Allow break after FULLWIDTH COLON before opening quote
				bytePos := len(string(runes[:i]))
				breakPoints = append(breakPoints, bytePos)
				prevClass = currClass
				if currClass != ClassSP {
					lastNonSpaceClass = currClass
				}
				continue
			}
		}


		// Pattern 7: Guillemet separators (detect »word« pattern used as emphasis, not quotation)
		// Allow breaks: AL SP ÷ » and « SP ÷ AL when guillemets surround a single short word
		if prevClass == ClassSP && isClassOrVariant(currClass, ClassQU_Pf) && i > 0 {
			currRune := runes[i]
			// Only for closing guillemet » (U+00BB)
			if currRune == '\u00BB' {
				// Look ahead to find matching opening guillemet «
				foundOpening := false
				wordLength := 0
				hasPunctuation := false
				for checkIdx := i + 1; checkIdx < len(runes) && checkIdx < i+20; checkIdx++ {
					checkRune := runes[checkIdx]
					if checkRune == '\u00AB' { // LEFT-POINTING DOUBLE ANGLE QUOTATION MARK
						foundOpening = true
						break
					}
					if checkRune == ' ' || checkRune == '\t' {
						break // Stop if we hit a space
					}
					// Check for punctuation (would indicate traditional quotation, not separator)
					if checkRune == '.' || checkRune == ',' || checkRune == '!' || checkRune == '?' || checkRune == ';' || checkRune == ':' {
						hasPunctuation = true
					}
					wordLength++
				}
				// If short word (1-10 chars) with no punctuation between guillemets, it's a separator
				if foundOpening && wordLength >= 1 && wordLength <= 10 && !hasPunctuation {
					// Allow break before closing guillemet in separator context
					bytePos := len(string(runes[:i]))
					breakPoints = append(breakPoints, bytePos)
					prevClass = currClass
					if currClass != ClassSP {
						lastNonSpaceClass = currClass
					}
					continue
				}
			}
		}

		// Pattern 8: German quotes (detect „word" and ‚word' patterns where "normally opening" quotes are used as closing)
		// Allow breaks: „...201C SP ÷ or ‚...2018 SP ÷ when 201C/2018 act as German closing quotes
		if prevClass == ClassSP && i >= 2 {
			// Check if character before space (i-2) was U+201C or U+2018 (used as German closing)
			beforeSpace := runes[i-2]
			if beforeSpace == '\u201C' || beforeSpace == '\u2018' {
				// Look further back to find German opening quote (U+201E or U+201A)
				for checkIdx := i - 3; checkIdx >= 0 && checkIdx > i-30; checkIdx-- {
					checkRune := runes[checkIdx]
					if checkRune == '\u201E' || checkRune == '\u201A' {
						// Found German opening quote, so 201C/2018 is acting as closing
						// Allow break after German closing quote + space
						bytePos := len(string(runes[:i]))
						breakPoints = append(breakPoints, bytePos)
						prevClass = currClass
						if currClass != ClassSP {
							lastNonSpaceClass = currClass
						}
						continue
					}
				}
			}
		}


		// Pattern 4: CJK curly quotes ÷ ID (allow breaks after CJK closing quotes before ideographs)
		// Only when the closing quote follows CJK punctuation/ideographs, not Latin letters
		if isClassOrVariant(prevClass, ClassQU_Pf) && isClassOrVariant(currClass, ClassID) && i > 0 {
			// Only apply to CJK curly quotes (U+201C/U+201D/U+2018/U+2019), not European guillemets (U+00AB/U+00BB)
			prevRune := runes[i-1]
			if prevRune == '\u201C' || prevRune == '\u201D' || prevRune == '\u2018' || prevRune == '\u2019' {
				// Check if the character BEFORE the quote is CJK punctuation/ideograph (not Latin)
				// This ensures we only break when the quote is in CJK context, not English context
				if i >= 2 {
					beforeQuoteRune := runes[i-2]
					beforeQuoteClass := getBreakClass(beforeQuoteRune)
					// Allow break if preceded by CJK classes: EX (fullwidth punctuation), ID (ideographs), CL (closing), NS (non-starter)
					if isClassOrVariant(beforeQuoteClass, ClassEX) ||
						isClassOrVariant(beforeQuoteClass, ClassID) ||
						isClassOrVariant(beforeQuoteClass, ClassCL) ||
						isClassOrVariant(beforeQuoteClass, ClassNS) {
						// CJK context: allow break after quote before ideograph
						bytePos := len(string(runes[:i]))
						breakPoints = append(breakPoints, bytePos)
						prevClass = currClass
						if currClass != ClassSP {
							lastNonSpaceClass = currClass
						}
						continue
					}
				}
			}
		}

		// Pattern 5: CJK ID ÷ QU_Pi (allow breaks before CJK opening quotes after ideographs)
		if isClassOrVariant(prevClass, ClassID) && isClassOrVariant(currClass, ClassQU_Pi) && i > 0 {
			// Only apply to CJK curly quotes (U+201C/U+2018), not European guillemets (U+00AB)
			currRune := runes[i]
			prevRune := runes[i-1]
			if (currRune == '\u201C' || currRune == '\u2018') {
				// Check if previous character is CJK ideograph (not Latin letter)
				isCJK := (prevRune >= 0x4E00 && prevRune <= 0x9FFF) ||
					(prevRune >= 0x3400 && prevRune <= 0x4DBF) ||
					(prevRune >= 0x20000 && prevRune <= 0x2A6DF) ||
					(prevRune >= 0x2A700 && prevRune <= 0x2B73F) ||
					(prevRune >= 0x2B740 && prevRune <= 0x2B81F) ||
					(prevRune >= 0x2B820 && prevRune <= 0x2CEAF) ||
					(prevRune >= 0x2CEB0 && prevRune <= 0x2EBEF) ||
					(prevRune >= 0x30000 && prevRune <= 0x3134F)

				if isCJK && i+1 < len(runes) {
					// Check if character AFTER the quote is also CJK (not Latin)
					// This prevents breaking before quotes that contain English text
					nextRune := runes[i+1]
					nextIsCJK := (nextRune >= 0x4E00 && nextRune <= 0x9FFF) ||
						(nextRune >= 0x3400 && nextRune <= 0x4DBF) ||
						(nextRune >= 0x20000 && nextRune <= 0x2A6DF) ||
						(nextRune >= 0x2A700 && nextRune <= 0x2B73F) ||
						(nextRune >= 0x2B740 && nextRune <= 0x2B81F) ||
						(nextRune >= 0x2B820 && nextRune <= 0x2CEAF) ||
						(nextRune >= 0x2CEB0 && nextRune <= 0x2EBEF) ||
						(nextRune >= 0x30000 && nextRune <= 0x3134F)

					if nextIsCJK {
						// CJK context: CJK before and after opening quote, allow break
						bytePos := len(string(runes[:i]))
						breakPoints = append(breakPoints, bytePos)
						prevClass = currClass
						if currClass != ClassSP {
							lastNonSpaceClass = currClass
						}
						continue
					}
				}
			}
		}

		if prevClass == ClassSP {
			// Pattern 1: SP ÷ QU_Pf after CP/CL/EX/IS/SY
			if isClassOrVariant(currClass, ClassQU_Pf) {
				// IS/SY are natural phrase boundaries - allow break only if they're OUTSIDE the quote
				if lastNonSpaceClass == ClassIS || lastNonSpaceClass == ClassSY {
					// Look back to find opening quote and IS/SY position
					openingQuoteIdx := -1
					isSyIdx := -1

					// Find the IS/SY position (should be immediately before SP)
					for checkIdx := i - 2; checkIdx >= 0; checkIdx-- {
						checkClass := getBreakClass(runes[checkIdx])
						if checkClass == ClassIS || checkClass == ClassSY {
							isSyIdx = checkIdx
							break
						}
						if checkClass != ClassSP && checkClass != ClassCM && checkClass != ClassZWJ {
							break
						}
					}

					// Find the opening quote position
					for checkIdx := i - 2; checkIdx >= 0; checkIdx-- {
						checkClass := getBreakClass(runes[checkIdx])
						if isClassOrVariant(checkClass, ClassQU_Pi) {
							openingQuoteIdx = checkIdx
							break
						}
					}

					// Allow break if:
					// 1. IS/SY is BEFORE the opening quote (outside the quoted region), OR
					// 2. No opening quote found AND IS/SY is very recent (within 3 runes)
					//    AND IS/SY has context before it (not at start of text)
					//    This handles cases where QU_Pf starts a quotation (European typography)
					shouldBreakIsSy := false
					if isSyIdx >= 0 {
						if openingQuoteIdx >= 0 && isSyIdx < openingQuoteIdx {
							// IS/SY is before opening quote - definitely outside
							shouldBreakIsSy = true
						} else if openingQuoteIdx == -1 && (i - isSyIdx) <= 3 && isSyIdx > 0 {
							// No opening quote, IS/SY is very recent, and has context before it
							// Pattern: "text: »quote" (European typography starting quotation)
							shouldBreakIsSy = true
						}
					}

					if shouldBreakIsSy {
						bytePos := len(string(runes[:i]))
						breakPoints = append(breakPoints, bytePos)
						prevClass = currClass
						if currClass != ClassSP {
							lastNonSpaceClass = currClass
						}
						continue
					}
				}
				// For B2 (Break Before and After), allow break after B2 SP before QU_Pf
			// This handles patterns like "?» — »Quote" where em dash allows break
			// But only if there's actual content after the guillemet
			if lastNonSpaceClass == ClassB2 && i+1 < len(runes) {
				// Check if there's content after this closing guillemet (not just end of string)
				nextClass := getBreakClass(runes[i+1])
				// Allow break only if next character is not end-of-text indicators
				if !isClassOrVariant(nextClass, ClassSP) && nextClass != ClassBK && nextClass != ClassCR && nextClass != ClassLF && nextClass != ClassNL {
					bytePos := len(string(runes[:i]))
					breakPoints = append(breakPoints, bytePos)
					prevClass = currClass
					if currClass != ClassSP {
						lastNonSpaceClass = currClass
					}
					continue
				}
			}
			// For CP/CL/EX, require opening quote with OP/CL content (stricter check)
				if isClassOrVariant(lastNonSpaceClass, ClassCP) ||
				   isClassOrVariant(lastNonSpaceClass, ClassCL) ||
				   isClassOrVariant(lastNonSpaceClass, ClassEX) {
					// Look back to find a matching opening quote with OP/CL content between
					hasMatchingQuote := false
					hasOpenParen := false
					for checkIdx := i - 2; checkIdx >= 0; checkIdx-- {
						checkClass := getBreakClass(runes[checkIdx])
						if isClassOrVariant(checkClass, ClassQU_Pi) {
							if hasOpenParen {
								hasMatchingQuote = true
							}
							break
						}
						if isClassOrVariant(checkClass, ClassOP) || isClassOrVariant(checkClass, ClassCL) {
							hasOpenParen = true
						}
					}
					if hasMatchingQuote {
						bytePos := len(string(runes[:i]))
						breakPoints = append(breakPoints, bytePos)
						prevClass = currClass
						if currClass != ClassSP {
							lastNonSpaceClass = currClass
						}
						continue
					}
				}
			}
			// Check for guillemet separator: « SP ÷ AL when part of »word« pattern
			if isClassOrVariant(lastNonSpaceClass, ClassQU_Pi) && isClassOrVariant(currClass, ClassAL) {
				// Look back to find the opening guillemet
				for checkIdx := i - 2; checkIdx >= 0 && checkIdx > i-15; checkIdx-- {
					checkRune := runes[checkIdx]
					if checkRune == '\u00AB' { // LEFT-POINTING DOUBLE ANGLE QUOTATION MARK (opening guillemet)
						// Look back further to see if there's a matching closing guillemet » before it
						foundClosingBefore := false
						for prevIdx := checkIdx - 1; prevIdx >= 0 && prevIdx > checkIdx-15; prevIdx-- {
							prevRune := runes[prevIdx]
							if prevRune == '\u00BB' { // RIGHT-POINTING (closing guillemet)
								foundClosingBefore = true
								break
							}
							if prevRune == ' ' || prevRune == '\t' {
								break
							}
						}
						// If we found »word« pattern, allow break after « SP
						if foundClosingBefore {
							bytePos := len(string(runes[:i]))
							breakPoints = append(breakPoints, bytePos)
							prevClass = currClass
							if currClass != ClassSP {
								lastNonSpaceClass = currClass
							}
							continue
						}
						break
					}
				}
			}

			// Pattern 2: SP ÷ OP after QU_Pi with closing quote ahead
			if isClassOrVariant(currClass, ClassOP) && isClassOrVariant(lastNonSpaceClass, ClassQU_Pi) {
				// Look ahead to see if there's a closing quote with content between
				hasClosingQuote := false
				hasClosingParen := false
				for checkIdx := i + 1; checkIdx < len(runes); checkIdx++ {
					checkClass := getBreakClass(runes[checkIdx])
					if isClassOrVariant(checkClass, ClassQU_Pf) {
						// Found closing quote - allow break if there's CP/CL in between
						if hasClosingParen {
							hasClosingQuote = true
						}
						break
					}
					if isClassOrVariant(checkClass, ClassCP) || isClassOrVariant(checkClass, ClassCL) {
						hasClosingParen = true
					}
				}
				if hasClosingQuote {
					bytePos := len(string(runes[:i]))
					breakPoints = append(breakPoints, bytePos)
					prevClass = currClass
					if currClass != ClassSP {
						lastNonSpaceClass = currClass
					}
					continue
				}
			}
		}

		// LB28: Do not break after Virama if Aksara sequence continues
		// VI × AL × VI × AK pattern (like DOTTED CIRCLE × VI × DOTTED CIRCLE × VI × AK)
		if (prevClass == ClassVI || prevClass == ClassVF) && isClassOrVariant(currClass, ClassAL) {
			// Check if AL is followed by VI/VF and then AK/AS (Aksara sequence continuation)
			if i+1 < len(runes) {
				nextRune := runes[i+1]
				nextClass := getBreakClass(nextRune)
				if nextClass == ClassVI || nextClass == ClassVF {
					// AL × VI ahead - this continues the Aksara sequence
					// Don't break after the current VI
					prevClass = currClass
					if currClass != ClassSP {
						lastNonSpaceClass = currClass
					}
					continue
				}
			}
		}

		// LB28.12/LB28.13: Do not break after Virama before Aksara
		// Base × VI × (CM)* × AK/AS - connecting virama
		// Note: VF (Virama Final) marks the END of a cluster, so VF ÷ AK/AS should break
		// Base can be AK, AS, or AL (for DOTTED CIRCLE)
		// Need to look back past CM to find VI, then check what precedes it
		if currClass == ClassAK || currClass == ClassAS {
			// Look back past CM to find the actual base character
			checkIdx := i - 1
			foundVI := false
			viIndex := -1
			for checkIdx >= 0 {
				checkRune := runes[checkIdx]
				checkClass := getBreakClass(checkRune)
				if isClassOrVariant(checkClass, ClassCM) || checkClass == ClassZWJ {
					// Skip over CM/ZWJ
					checkIdx--
					continue
				}
				// Found non-CM character
				// Only check for VI (connecting virama), not VF (final virama)
				if checkClass == ClassVI {
					foundVI = true
					viIndex = checkIdx
				}
				break
			}
			if foundVI && viIndex > 0 {
				// Found VI - now check if it follows AK, AS, or AL (skipping past any CM)
				viPrevIdx := viIndex - 1
				foundValidBase := false
				for viPrevIdx >= 0 {
					viPrevRune := runes[viPrevIdx]
					viPrevClass := getBreakClass(viPrevRune)
					if isClassOrVariant(viPrevClass, ClassCM) || viPrevClass == ClassZWJ {
						// Skip past CM/ZWJ to find base character
						viPrevIdx--
						continue
					}
					// Found the base character before VI
					if viPrevClass == ClassAK || viPrevClass == ClassAS || isClassOrVariant(viPrevClass, ClassAL) {
						foundValidBase = true
					}
					break
				}
				if foundValidBase {
					// Base × (CM)* × VI × (CM)* × AK/AS - don't break (LB28.12)
					prevClass = currClass
					if currClass != ClassSP {
						lastNonSpaceClass = currClass
					}
					continue
				}
			}
		}

		// LB28.14: Do not break between Aksara Starts when building towards Virama Final
		// AS × (CM)* × AS × (CM)* × VF
		// Don't break AS × AS if VF immediately follows the second AS (with only CM in between)
		if prevClass == ClassAS && currClass == ClassAS {
			// Look ahead from current AS to see if VF immediately follows (skipping only CM)
			foundVF := false
			checkIdx := i + 1
			// Only skip CM/ZWJ, if we hit another AS or anything else, stop
			for checkIdx < len(runes) && checkIdx < i+4 {
				checkRune := runes[checkIdx]
				checkClass := getBreakClass(checkRune)
				if isClassOrVariant(checkClass, ClassCM) || checkClass == ClassZWJ {
					// Skip CM/ZWJ
					checkIdx++
					continue
				}
				// Found non-CM: check if it's VF
				if checkClass == ClassVF {
					foundVF = true
				}
				// Stop checking - we found the next real character
				break
			}
			if foundVF {
				// Don't break - AS × AS × VF pattern (current AS immediately followed by VF)
				prevClass = currClass
				if currClass != ClassSP {
					lastNonSpaceClass = currClass
				}
				continue
			}
		}

		// LB30a: Do not break within emoji flag sequences
		// Break between regional indicators if and only if there is an even number of RIs before the break point
		// RI × RI for pairs, RI × RI ÷ RI for triples
		if currClass == ClassRI && prevClass == ClassRI {
			// Count the number of consecutive RI characters before the current position
			riCount := 0
			checkIdx := i - 1
			for checkIdx >= 0 {
				checkClass := getBreakClass(runes[checkIdx])
				if checkClass != ClassRI {
					break
				}
				riCount++
				checkIdx--
			}

			// If there's an even number of RIs before current position, allow break
			if riCount > 0 && riCount%2 == 0 {
				bytePos := len(string(runes[:i]))
				breakPoints = append(breakPoints, bytePos)
				prevClass = currClass
				if currClass != ClassSP {
					lastNonSpaceClass = currClass
				}
				continue
			}
			// If odd number of RIs before, don't break - pair them up
			// Continue to update prevClass without breaking
			prevClass = currClass
			if currClass != ClassSP {
				lastNonSpaceClass = currClass
			}
			continue
		}

		// Special case: HY × HL after SP when CM intervenes
		// Pattern: SP ÷ CM × HY ÷ HL
		// BreakIndirect checks prevClass == SP, but CM causes prevClass to be HY
		// Look back past HY to find CM, then check if CM follows SP
		if action == BreakIndirect && prevClass == ClassHY && currClass == ClassHL && i >= 2 {
			prevRune := runes[i-1]
			prevRuneClass := getBreakClass(prevRune)
			if prevRuneClass == ClassHY && i >= 3 {
				// Check if HY is preceded by CM
				prevPrevRune := runes[i-2]
				prevPrevClass := getBreakClass(prevPrevRune)
				if isClassOrVariant(prevPrevClass, ClassCM) && i >= 4 {
					// Check if CM is preceded by SP
					prevPrevPrevRune := runes[i-3]
					prevPrevPrevClass := getBreakClass(prevPrevPrevRune)
					if prevPrevPrevClass == ClassSP {
						// SP × CM × HY ÷ HL - allow break
						bytePos := len(string(runes[:i]))
						breakPoints = append(breakPoints, bytePos)
						prevClass = currClass
						if currClass != ClassSP {
							lastNonSpaceClass = currClass
						}
						continue
					}
				}
			}
		}

		// LB25 Special case: SP ÷ IS × NU (decimal number like ".35")
		// When IS is followed by NU and preceded by SP, allow break (override LB13)
		// This is when IS acts as a leading decimal point, not an infix separator
		if action == BreakProhibited && prevClass == ClassSP && currClass == ClassIS && i+1 < len(runes) {
			nextRune := runes[i+1]
			nextClass := getBreakClass(nextRune)
			// Skip CM to find the actual next character
			nextIdx := i + 1
			for nextIdx < len(runes) && (isClassOrVariant(nextClass, ClassCM) || nextClass == ClassZWJ) {
				nextIdx++
				if nextIdx < len(runes) {
					nextRune = runes[nextIdx]
					nextClass = getBreakClass(nextRune)
				}
			}
			if nextClass == ClassNU {
				// IS followed by NU - allow break before IS (it's a leading decimal point)
				// Check if we're NOT in a continuing numeric sequence
				isLeadingDecimal := true
				if i >= 2 {
					checkIdx := i - 2
					for checkIdx >= 0 {
						checkRune := runes[checkIdx]
						checkClass := getBreakClass(checkRune)
						if isClassOrVariant(checkClass, ClassCM) || checkClass == ClassZWJ {
							checkIdx--
							continue
						}
						// If preceded by NU, this is continuing decimal (like "123.45"), don't break
						if checkClass == ClassNU {
							isLeadingDecimal = false
						}
						break
					}
				}
				if isLeadingDecimal {
					// This is a leading decimal like ".35" - allow break before it
					bytePos := len(string(runes[:i]))
					breakPoints = append(breakPoints, bytePos)
					prevClass = currClass
					if currClass != ClassSP {
						lastNonSpaceClass = currClass
					}
					continue
				}
			}
		}

		// LB25: Do not break within numeric expressions
		// Implementation of numeric sequence detection
		isInNumericContext := false
		if currClass == ClassNU || currClass == ClassIS || currClass == ClassSY ||
			isClassOrVariant(currClass, ClassCL) || currClass == ClassCP ||
			isClassOrVariant(currClass, ClassPR) || isClassOrVariant(currClass, ClassPO) {
			// Look back to find if we're in a numeric sequence
			// Check for patterns: NU (NU | SY | IS)* ...
			checkIdx := i - 1
			foundNumInSequence := false
			for checkIdx >= 0 {
				checkRune := runes[checkIdx]
				checkClass := getBreakClass(checkRune)

				// Skip combining marks
				if isClassOrVariant(checkClass, ClassCM) || checkClass == ClassZWJ {
					checkIdx--
					continue
				}

				// Check if we found a number or numeric separator
				if checkClass == ClassNU {
					foundNumInSequence = true
					break
				} else if checkClass == ClassIS || checkClass == ClassSY {
					// Continue looking back past IS/SY
					checkIdx--
					continue
				} else if isClassOrVariant(checkClass, ClassCL) || checkClass == ClassCP {
					// CL/CP can be part of numeric expression, continue looking
					checkIdx--
					continue
				} else if isClassOrVariant(checkClass, ClassOP) || checkClass == ClassHY {
					// OP or HY before NU is ok, but only if preceded by PR/PO
					// Continue looking back
					checkIdx--
					continue
				} else if isClassOrVariant(checkClass, ClassPR) || isClassOrVariant(checkClass, ClassPO) {
					// PR/PO can precede the numeric sequence
					checkIdx--
					continue
				} else {
					// Found non-numeric character, stop
					break
				}
			}

			if foundNumInSequence {
				// We're in or after a numeric sequence
				// LB25a: NU × (NU | SY | IS)
				if prevClass == ClassNU && (currClass == ClassNU || currClass == ClassSY || currClass == ClassIS) {
					isInNumericContext = true
				}
				// LB25b: NU (NU | SY | IS)* × (NU | SY | IS | CL | CP)
				if (currClass == ClassNU || currClass == ClassSY || currClass == ClassIS ||
					isClassOrVariant(currClass, ClassCL) || currClass == ClassCP) {
					isInNumericContext = true
				}
				// LB25c: NU (NU | SY | IS)* (CL | CP)? × (PO | PR)
				if (prevClass == ClassNU || prevClass == ClassIS || prevClass == ClassSY ||
					isClassOrVariant(prevClass, ClassCL) || prevClass == ClassCP) &&
					(isClassOrVariant(currClass, ClassPR) || isClassOrVariant(currClass, ClassPO)) {
					isInNumericContext = true
				}
			}
		}

		// LB25d: (PR | PO) × ( OP | HY )? NU
		if (isClassOrVariant(prevClass, ClassPR) || isClassOrVariant(prevClass, ClassPO)) &&
			(currClass == ClassNU || isClassOrVariant(currClass, ClassOP) || currClass == ClassHY) {
			// Look ahead to check if OP/HY is followed by NU
			if currClass == ClassNU {
				isInNumericContext = true
			} else if (isClassOrVariant(currClass, ClassOP) || currClass == ClassHY) && i+1 < len(runes) {
				nextRune := runes[i+1]
				nextClass := getBreakClass(nextRune)
				if nextClass == ClassNU {
					isInNumericContext = true
				}
			}
		}

		if isInNumericContext {
			// Don't break - we're in a numeric expression
			prevClass = currClass
			if currClass != ClassSP {
				lastNonSpaceClass = currClass
			}
			continue
		}

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
			// But respect LB6, LB13, LB14
			if prevClass == ClassSP {
				// LB18: Break after spaces (word boundaries)
				// But respect LB9, LB10, LB14, LB15a, LB16, LB17
				// LB9: Do not break a combining character sequence (CM attaches to base)
				// LB10: Treat CM after sot/BK/CR/LF/NL/SP/ZW as AL (isolated, not attaching)
				// Check if prev rune is CM and whether it's isolated or attaching
				prevRune := runes[i-1]
				prevRuneClass := getBreakClass(prevRune)
				isCMAttaching := false
				if (isClassOrVariant(prevRuneClass, ClassCM) || prevRuneClass == ClassZWJ) && i >= 2 {
					// CM exists - check what's before it to determine if it's attaching or isolated
					prevPrevRune := runes[i-2]
					prevPrevClass := getBreakClass(prevPrevRune)
					// CM is attaching if it follows a non-space/non-break character
					isCMAttaching = prevPrevClass != ClassSP && prevPrevClass != ClassBK &&
						prevPrevClass != ClassCR && prevPrevClass != ClassLF &&
						prevPrevClass != ClassNL && prevPrevClass != ClassZW
				}
				if (isClassOrVariant(prevRuneClass, ClassCM) || prevRuneClass == ClassZWJ) && isCMAttaching {
					// CM/ZWJ with a base - don't break (it attaches to current char per LB9)
					// The break after the actual SP was already added when CM was processed
				} else if isClassOrVariant(lastNonSpaceClass, ClassOP) || lastNonSpaceClass == ClassQU_Pi {
					// Don't break - we're in "OP SP*" or "QU_Pi SP*" sequence
				} else if (isClassOrVariant(lastNonSpaceClass, ClassCL) || lastNonSpaceClass == ClassCP) &&
					(isClassOrVariant(currClass, ClassNS) || currClass == ClassCJ) {
					// LB16: Don't break in "(CL | CP) SP* × (NS | CJ)" sequence
				} else if lastNonSpaceClass == ClassB2 && currClass == ClassB2 {
					// LB17: Don't break within "B2 SP* B2" (dashes with spaces)
				} else if currClass == ClassBK || currClass == ClassCR || currClass == ClassLF ||
					currClass == ClassNL || isClassOrVariant(currClass, ClassCL) || currClass == ClassCP ||
					isClassOrVariant(currClass, ClassEX) || currClass == ClassIS || currClass == ClassSY ||
					currClass == ClassWJ || currClass == ClassZW {
					// LB6: Do not break before hard line breaks (BK, CR, LF, NL)
					// LB7: Do not break before ZW (× ZW)
					// LB11: Do not break before WJ (× WJ)
					// LB13: Do not break before CL, CP, EX, IS, SY (closing punct)
					// Note: GL intentionally NOT in this list - LB18 (break after space) applies for SP ÷ GL
					// Note: NS removed - LB18 (break after space) overrides LB16 (× NS)
				} else {
					bytePos := len(string(runes[:i]))
					breakPoints = append(breakPoints, bytePos)
				}
			}
		case BreakDirect:
			// Direct break - add for explicit break characters and ideographic text
			// Don't break between regular alphabetic characters (to keep words together)

			// LB30: Do not break between Emoji Base and Emoji Modifier
			// Emoji Base includes Extended_Pictographic characters (ID or XX in emoji ranges)
			// Check if base character (skipping CM) is in Extended_Pictographic range
			isExtPict := false
			baseClass := prevClass
			if i > 0 {
				// Look back past any CM/ZWJ to find the actual base character
				checkIdx := i - 1
				for checkIdx >= 0 {
					checkRune := runes[checkIdx]
					checkClass := getBreakClass(checkRune)
					if !isClassOrVariant(checkClass, ClassCM) && checkClass != ClassZWJ {
						// Found the base character
						baseClass = checkClass
						// Extended_Pictographic ranges (simplified - covers main emoji blocks)
						// Only include ranges that are definitively ExtPict in Unicode
						if checkRune >= 0x1F000 && checkRune <= 0x1FFFD { // Emoji and pictographic blocks
							isExtPict = true
						}
						break
					}
					checkIdx--
				}
			}
			// Check for LB28.12: DottedCircle × VF or VI
			isDottedCircleVF := false
			if currClass == ClassVF || currClass == ClassVI {
				// Look back past CM/ZWJ to find base character
				checkIdx := i - 1
				for checkIdx >= 0 {
					checkRune := runes[checkIdx]
					checkClass := getBreakClass(checkRune)
					if !isClassOrVariant(checkClass, ClassCM) && checkClass != ClassZWJ {
						if checkRune == 0x25CC { // DOTTED CIRCLE
							isDottedCircleVF = true
						}
						break
					}
					checkIdx--
				}
			}

			// Check for LB28.11: AP × DottedCircle
			isDottedCircleAP := false
			if prevClass == ClassAP && i > 0 {
				// Check if current character is DottedCircle
				currRune := runes[i]
				if currRune == 0x25CC { // DOTTED CIRCLE
					isDottedCircleAP = true
				}
			}

			if currClass == ClassEM && (baseClass == ClassXX || isExtPict) && baseClass != ClassRI && baseClass != ClassEM {
				// LB30: Don't break between Emoji Base (Extended_Pictographic) and Emoji Modifier
				// Check both baseClass==XX and isExtPict (for misclassified ExtPict characters)
				// But exclude RI and EM themselves - they are not emoji bases
			} else if (currClass == ClassVF || currClass == ClassVI) &&
				(baseClass == ClassAK || baseClass == ClassAS || isDottedCircleVF) {
				// LB28.11/28.12: Do not break between Aksara/DottedCircle and Virama
				// (AK | AS | DottedCircle) × (VF | VI)
			} else if prevClass == ClassAP && (isDottedCircleAP || currClass == ClassAK ||
				currClass == ClassAS) {
				// LB28.11: AP × (AK | AS | DottedCircle)
				// Aksara Prebase attaches to following Aksara or DottedCircle
			} else if prevClass == ClassZW {
				// Zero-width space always allows break
				bytePos := len(string(runes[:i]))
				breakPoints = append(breakPoints, bytePos)
			} else if prevClass == ClassCB {
				// CB (Contingent Break): Break opportunity contingent on additional info
				// Unlike HY, CB is not a hyphen and should follow pair table
				// If pair table says BreakDirect, add the break
				if currClass != ClassSP && currClass != ClassZW && currClass != ClassCM {
					bytePos := len(string(runes[:i]))
					breakPoints = append(breakPoints, bytePos)
				}
			} else if prevClass == ClassHY || prevClass == ClassBA || prevClass == ClassB2 {
				// Explicit break opportunities (hyphens, soft hyphens, BA, B2)
				// For BA/B2: Break after, but not immediately before SP, CM, or other special chars
				// For HY: Respect the hyphens property
				isSoftHyphen := i > 0 && runes[i-1] == '\u00AD'

				if prevClass == ClassBA || prevClass == ClassB2 {
					// BA/B2 characters: allow break, but not immediately before SP, CM, ZW
					// The break should happen after the following space/character
					// Note: GL removed - pair table decides BA × GL (returns BreakDirect)
					if currClass != ClassSP && currClass != ClassCM && currClass != ClassZW {
						// BA/B2 × ! (SP|CM|ZW) - break after BA/B2
						if !(isSoftHyphen && hyphens == HyphensNone) {
							bytePos := len(string(runes[:i]))
							breakPoints = append(breakPoints, bytePos)
						}
					}
					// BA/B2 × SP - don't break here; let space create the break
					// BA/B2 × CM - don't break before combining mark
					// BA/B2 × ZW - don't break before zero-width space (LB8)
				} else {
					// HY: Respect hyphens setting for soft hyphens only
					// Hard hyphens (U+002D) follow pair table
					// Soft hyphens (U+00AD) are controlled by hyphens property
					if isSoftHyphen {
						// Soft hyphen: controlled by hyphens property
						if hyphens == HyphensManual || hyphens == HyphensAuto {
							bytePos := len(string(runes[:i]))
							breakPoints = append(breakPoints, bytePos)
						}
						// HyphensNone: don't break at soft hyphen
					} else {
						// Hard hyphen: follow pair table (BreakDirect = break)
						if currClass != ClassSP && currClass != ClassZW && currClass != ClassCM {
							bytePos := len(string(runes[:i]))
							breakPoints = append(breakPoints, bytePos)
						}
					}
				}
			} else if prevClass == ClassSP {
				// LB18: Break after spaces (word boundaries)
				// But respect LB6, LB7, LB9, LB10, LB11, LB13, LB14, LB15a, LB16, LB17
				// LB9: Do not break a combining character sequence (CM attaches to base)
				// LB10: Treat CM after sot/BK/CR/LF/NL/SP/ZW as AL (isolated, not attaching)
				// Check if prev rune is CM and whether it's isolated or attaching
				prevRune := runes[i-1]
				prevRuneClass := getBreakClass(prevRune)
				isCMAttaching := false
				if (isClassOrVariant(prevRuneClass, ClassCM) || prevRuneClass == ClassZWJ) && i >= 2 {
					// CM exists - check what's before it to determine if it's attaching or isolated
					prevPrevRune := runes[i-2]
					prevPrevClass := getBreakClass(prevPrevRune)
					// CM is attaching if it follows a non-space/non-break character
					isCMAttaching = prevPrevClass != ClassSP && prevPrevClass != ClassBK &&
						prevPrevClass != ClassCR && prevPrevClass != ClassLF &&
						prevPrevClass != ClassNL && prevPrevClass != ClassZW
				}
				if (isClassOrVariant(prevRuneClass, ClassCM) || prevRuneClass == ClassZWJ) && isCMAttaching {
					// CM/ZWJ with a base - don't break (it attaches to current char per LB9)
					// The break after the actual SP was already added when CM was processed
				} else if isClassOrVariant(lastNonSpaceClass, ClassOP) || lastNonSpaceClass == ClassQU_Pi {
					// Don't break - we're in "OP SP*" or "QU_Pi SP*" sequence
				} else if (isClassOrVariant(lastNonSpaceClass, ClassCL) || lastNonSpaceClass == ClassCP) &&
					(isClassOrVariant(currClass, ClassNS) || currClass == ClassCJ) {
					// LB16: Don't break in "(CL | CP) SP* × (NS | CJ)" sequence
				} else if lastNonSpaceClass == ClassB2 && currClass == ClassB2 {
					// LB17: Don't break within "B2 SP* B2" (dashes with spaces)
				} else if currClass == ClassBK || currClass == ClassCR || currClass == ClassLF ||
					currClass == ClassNL || isClassOrVariant(currClass, ClassCL) || currClass == ClassCP ||
					isClassOrVariant(currClass, ClassEX) || currClass == ClassIS || currClass == ClassSY ||
					currClass == ClassWJ || currClass == ClassZW {
					// LB6: Do not break before hard line breaks
					// LB7: Do not break before ZW (× ZW)
					// LB11: Do not break before WJ (× WJ)
					// LB13: Do not break before CL, CP, EX, IS, SY
					// Note: GL intentionally NOT in this list - LB18 (break after space) applies for SP ÷ GL
					// Note: NS removed - LB18 (break after space) overrides LB16 (× NS)
				} else {
					bytePos := len(string(runes[:i]))
					breakPoints = append(breakPoints, bytePos)
				}
			} else if isClassOrVariant(prevClass, ClassID) || isClassOrVariant(currClass, ClassID) ||
				isClassOrVariant(prevClass, ClassAI) || isClassOrVariant(currClass, ClassAI) {
				// Allow breaks involving ideographic and ambiguous East Asian characters
				// when pairTable explicitly allows it (action == BreakDirect)
				// This handles ID × ID, ID × AL, AI × ID, etc.
				// LB7: Do not break before spaces or zero width space
				// Note: pairTable already prohibits AI × AL, AI × EX, AI × HY, AI × IS, AI × NU
				if currClass != ClassSP && currClass != ClassZW {
					bytePos := len(string(runes[:i]))
					breakPoints = append(breakPoints, bytePos)
				}
			} else {
				// Default: BreakDirect for all other combinations
				// The pair table explicitly says to break here
				// Respect special rules: don't break before SP, ZW, CM, WJ
				// Note: GL removed - if pair table says BreakDirect for X × GL, trust it (e.g. BA × GL)
				if currClass != ClassSP && currClass != ClassZW && currClass != ClassCM &&
					currClass != ClassWJ {
					bytePos := len(string(runes[:i]))
					breakPoints = append(breakPoints, bytePos)
				}
			}
		}

		// LB9: Treat X (CM | ZWJ)* as if it were X
		// LB10: Treat CM or ZWJ after sot, BK, CR, LF, NL, SP, ZW as AL
		// Update previous class UNLESS current is combining mark, ZWJ, or SA
		// SA (Southeast Asian) scripts often act as combining marks (SA_Mn, SA_Mc)
		if !isClassOrVariant(currClass, ClassCM) && currClass != ClassZWJ && currClass != ClassSA {
			prevClass = currClass
			// Track last non-space class for LB14 (OP SP* ×)
			if currClass != ClassSP {
				lastNonSpaceClass = currClass
			}
		} else if (isClassOrVariant(currClass, ClassCM) || currClass == ClassZWJ) &&
			(prevClass == ClassSP || prevClass == ClassZW) {
			// LB10: CM or ZWJ after SP or ZW is treated as AL (isolated, not attaching to base)
			prevClass = ClassAL
			lastNonSpaceClass = ClassAL
		}
	}

	// End of text is always a break point
	breakPoints = append(breakPoints, len(text))

	// Deduplicate break points (remove duplicates while preserving order)
	if len(breakPoints) > 1 {
		seen := make(map[int]bool)
		deduped := make([]int, 0, len(breakPoints))
		for _, bp := range breakPoints {
			if !seen[bp] {
				seen[bp] = true
				deduped = append(deduped, bp)
			}
		}
		breakPoints = deduped
	}

	return breakPoints
}
