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
// Uses official Unicode LineBreak.txt property data.
// Reference: http://www.unicode.org/reports/tr14/#Table1
func getBreakClass(r rune) BreakClass {
	// Use official Unicode data
	if class, ok := getBreakClassFromData(r); ok {
		return class
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

var pairTable = map[[2]BreakClass]BreakAction{
	{ClassAI, ClassAI}: BreakIndirect,
	{ClassAI, ClassAK}: BreakDirect,
	{ClassAI, ClassAL}: BreakIndirect,
	{ClassAI, ClassAP}: BreakDirect,
	{ClassAI, ClassAS}: BreakDirect,
	{ClassAI, ClassB2}: BreakDirect,
	{ClassAI, ClassBA}: BreakIndirect,
	{ClassAI, ClassBB}: BreakDirect,
	{ClassAI, ClassBK}: BreakProhibited,
	{ClassAI, ClassCB}: BreakDirect,
	{ClassAI, ClassCJ}: BreakIndirect,
	{ClassAI, ClassCL}: BreakProhibited,
	{ClassAI, ClassCM}: BreakIndirect,
	{ClassAI, ClassCP}: BreakProhibited,
	{ClassAI, ClassCR}: BreakProhibited,
	{ClassAI, ClassEB}: BreakDirect,
	{ClassAI, ClassEM}: BreakDirect,
	{ClassAI, ClassEX}: BreakProhibited,
	{ClassAI, ClassGL}: BreakIndirect,
	{ClassAI, ClassH2}: BreakDirect,
	{ClassAI, ClassH3}: BreakDirect,
	{ClassAI, ClassHH}: BreakIndirect,
	{ClassAI, ClassHL}: BreakIndirect,
	{ClassAI, ClassHY}: BreakIndirect,
	{ClassAI, ClassID}: BreakDirect,
	{ClassAI, ClassIN}: BreakIndirect,
	{ClassAI, ClassIS}: BreakProhibited,
	{ClassAI, ClassJL}: BreakDirect,
	{ClassAI, ClassJT}: BreakDirect,
	{ClassAI, ClassJV}: BreakDirect,
	{ClassAI, ClassLF}: BreakProhibited,
	{ClassAI, ClassNL}: BreakProhibited,
	{ClassAI, ClassNS}: BreakIndirect,
	{ClassAI, ClassNU}: BreakIndirect,
	{ClassAI, ClassOP}: BreakDirect,
	{ClassAI, ClassPO}: BreakIndirect,
	{ClassAI, ClassPR}: BreakIndirect,
	{ClassAI, ClassQU}: BreakProhibited,
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
	{ClassAK, ClassAI}: BreakDirect,
	{ClassAK, ClassAK}: BreakDirect,
	{ClassAK, ClassAL}: BreakDirect,
	{ClassAK, ClassAP}: BreakDirect,
	{ClassAK, ClassAS}: BreakDirect,
	{ClassAK, ClassB2}: BreakDirect,
	{ClassAK, ClassBA}: BreakIndirect,
	{ClassAK, ClassBB}: BreakDirect,
	{ClassAK, ClassBK}: BreakProhibited,
	{ClassAK, ClassCB}: BreakDirect,
	{ClassAK, ClassCJ}: BreakIndirect,
	{ClassAK, ClassCL}: BreakProhibited,
	{ClassAK, ClassCM}: BreakIndirect,
	{ClassAK, ClassCP}: BreakProhibited,
	{ClassAK, ClassCR}: BreakProhibited,
	{ClassAK, ClassEB}: BreakDirect,
	{ClassAK, ClassEM}: BreakDirect,
	{ClassAK, ClassEX}: BreakProhibited,
	{ClassAK, ClassGL}: BreakIndirect,
	{ClassAK, ClassH2}: BreakDirect,
	{ClassAK, ClassH3}: BreakDirect,
	{ClassAK, ClassHH}: BreakIndirect,
	{ClassAK, ClassHL}: BreakDirect,
	{ClassAK, ClassHY}: BreakIndirect,
	{ClassAK, ClassID}: BreakDirect,
	{ClassAK, ClassIN}: BreakIndirect,
	{ClassAK, ClassIS}: BreakProhibited,
	{ClassAK, ClassJL}: BreakDirect,
	{ClassAK, ClassJT}: BreakDirect,
	{ClassAK, ClassJV}: BreakDirect,
	{ClassAK, ClassLF}: BreakProhibited,
	{ClassAK, ClassNL}: BreakProhibited,
	{ClassAK, ClassNS}: BreakIndirect,
	{ClassAK, ClassNU}: BreakDirect,
	{ClassAK, ClassOP}: BreakDirect,
	{ClassAK, ClassPO}: BreakDirect,
	{ClassAK, ClassPR}: BreakDirect,
	{ClassAK, ClassQU}: BreakProhibited,
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
	{ClassAL, ClassAK}: BreakDirect,
	{ClassAL, ClassAL}: BreakIndirect,
	{ClassAL, ClassAP}: BreakDirect,
	{ClassAL, ClassAS}: BreakDirect,
	{ClassAL, ClassB2}: BreakDirect,
	{ClassAL, ClassBA}: BreakIndirect,
	{ClassAL, ClassBB}: BreakDirect,
	{ClassAL, ClassBK}: BreakProhibited,
	{ClassAL, ClassCB}: BreakDirect,
	{ClassAL, ClassCJ}: BreakIndirect,
	{ClassAL, ClassCL}: BreakProhibited,
	{ClassAL, ClassCM}: BreakIndirect,
	{ClassAL, ClassCP}: BreakProhibited,
	{ClassAL, ClassCR}: BreakProhibited,
	{ClassAL, ClassEB}: BreakDirect,
	{ClassAL, ClassEM}: BreakDirect,
	{ClassAL, ClassEX}: BreakProhibited,
	{ClassAL, ClassGL}: BreakIndirect,
	{ClassAL, ClassH2}: BreakDirect,
	{ClassAL, ClassH3}: BreakDirect,
	{ClassAL, ClassHH}: BreakIndirect,
	{ClassAL, ClassHL}: BreakIndirect,
	{ClassAL, ClassHY}: BreakIndirect,
	{ClassAL, ClassID}: BreakDirect,
	{ClassAL, ClassIN}: BreakIndirect,
	{ClassAL, ClassIS}: BreakProhibited,
	{ClassAL, ClassJL}: BreakDirect,
	{ClassAL, ClassJT}: BreakDirect,
	{ClassAL, ClassJV}: BreakDirect,
	{ClassAL, ClassLF}: BreakProhibited,
	{ClassAL, ClassNL}: BreakProhibited,
	{ClassAL, ClassNS}: BreakIndirect,
	{ClassAL, ClassNU}: BreakIndirect,
	{ClassAL, ClassOP}: BreakDirect,
	{ClassAL, ClassPO}: BreakIndirect,
	{ClassAL, ClassPR}: BreakIndirect,
	{ClassAL, ClassQU}: BreakProhibited,
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
	{ClassAP, ClassAI}: BreakDirect,
	{ClassAP, ClassAK}: BreakIndirect,
	{ClassAP, ClassAL}: BreakDirect,
	{ClassAP, ClassAP}: BreakDirect,
	{ClassAP, ClassAS}: BreakIndirect,
	{ClassAP, ClassB2}: BreakDirect,
	{ClassAP, ClassBA}: BreakIndirect,
	{ClassAP, ClassBB}: BreakDirect,
	{ClassAP, ClassBK}: BreakProhibited,
	{ClassAP, ClassCB}: BreakDirect,
	{ClassAP, ClassCJ}: BreakIndirect,
	{ClassAP, ClassCL}: BreakProhibited,
	{ClassAP, ClassCM}: BreakIndirect,
	{ClassAP, ClassCP}: BreakProhibited,
	{ClassAP, ClassCR}: BreakProhibited,
	{ClassAP, ClassEB}: BreakDirect,
	{ClassAP, ClassEM}: BreakDirect,
	{ClassAP, ClassEX}: BreakProhibited,
	{ClassAP, ClassGL}: BreakIndirect,
	{ClassAP, ClassH2}: BreakDirect,
	{ClassAP, ClassH3}: BreakDirect,
	{ClassAP, ClassHH}: BreakIndirect,
	{ClassAP, ClassHL}: BreakDirect,
	{ClassAP, ClassHY}: BreakIndirect,
	{ClassAP, ClassID}: BreakDirect,
	{ClassAP, ClassIN}: BreakIndirect,
	{ClassAP, ClassIS}: BreakProhibited,
	{ClassAP, ClassJL}: BreakDirect,
	{ClassAP, ClassJT}: BreakDirect,
	{ClassAP, ClassJV}: BreakDirect,
	{ClassAP, ClassLF}: BreakProhibited,
	{ClassAP, ClassNL}: BreakProhibited,
	{ClassAP, ClassNS}: BreakIndirect,
	{ClassAP, ClassNU}: BreakDirect,
	{ClassAP, ClassOP}: BreakDirect,
	{ClassAP, ClassPO}: BreakDirect,
	{ClassAP, ClassPR}: BreakDirect,
	{ClassAP, ClassQU}: BreakProhibited,
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
	{ClassAS, ClassAK}: BreakDirect,
	{ClassAS, ClassAL}: BreakDirect,
	{ClassAS, ClassAP}: BreakDirect,
	{ClassAS, ClassAS}: BreakDirect,
	{ClassAS, ClassB2}: BreakDirect,
	{ClassAS, ClassBA}: BreakIndirect,
	{ClassAS, ClassBB}: BreakDirect,
	{ClassAS, ClassBK}: BreakProhibited,
	{ClassAS, ClassCB}: BreakDirect,
	{ClassAS, ClassCJ}: BreakIndirect,
	{ClassAS, ClassCL}: BreakProhibited,
	{ClassAS, ClassCM}: BreakIndirect,
	{ClassAS, ClassCP}: BreakProhibited,
	{ClassAS, ClassCR}: BreakProhibited,
	{ClassAS, ClassEB}: BreakDirect,
	{ClassAS, ClassEM}: BreakDirect,
	{ClassAS, ClassEX}: BreakProhibited,
	{ClassAS, ClassGL}: BreakIndirect,
	{ClassAS, ClassH2}: BreakDirect,
	{ClassAS, ClassH3}: BreakDirect,
	{ClassAS, ClassHH}: BreakIndirect,
	{ClassAS, ClassHL}: BreakDirect,
	{ClassAS, ClassHY}: BreakIndirect,
	{ClassAS, ClassID}: BreakDirect,
	{ClassAS, ClassIN}: BreakIndirect,
	{ClassAS, ClassIS}: BreakProhibited,
	{ClassAS, ClassJL}: BreakDirect,
	{ClassAS, ClassJT}: BreakDirect,
	{ClassAS, ClassJV}: BreakDirect,
	{ClassAS, ClassLF}: BreakProhibited,
	{ClassAS, ClassNL}: BreakProhibited,
	{ClassAS, ClassNS}: BreakIndirect,
	{ClassAS, ClassNU}: BreakDirect,
	{ClassAS, ClassOP}: BreakDirect,
	{ClassAS, ClassPO}: BreakDirect,
	{ClassAS, ClassPR}: BreakDirect,
	{ClassAS, ClassQU}: BreakProhibited,
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
	{ClassB2, ClassAK}: BreakDirect,
	{ClassB2, ClassAL}: BreakDirect,
	{ClassB2, ClassAP}: BreakDirect,
	{ClassB2, ClassAS}: BreakDirect,
	{ClassB2, ClassB2}: BreakProhibited,
	{ClassB2, ClassBA}: BreakIndirect,
	{ClassB2, ClassBB}: BreakDirect,
	{ClassB2, ClassBK}: BreakProhibited,
	{ClassB2, ClassCB}: BreakDirect,
	{ClassB2, ClassCJ}: BreakIndirect,
	{ClassB2, ClassCL}: BreakProhibited,
	{ClassB2, ClassCM}: BreakIndirect,
	{ClassB2, ClassCP}: BreakProhibited,
	{ClassB2, ClassCR}: BreakProhibited,
	{ClassB2, ClassEB}: BreakDirect,
	{ClassB2, ClassEM}: BreakDirect,
	{ClassB2, ClassEX}: BreakProhibited,
	{ClassB2, ClassGL}: BreakIndirect,
	{ClassB2, ClassH2}: BreakDirect,
	{ClassB2, ClassH3}: BreakDirect,
	{ClassB2, ClassHH}: BreakIndirect,
	{ClassB2, ClassHL}: BreakDirect,
	{ClassB2, ClassHY}: BreakIndirect,
	{ClassB2, ClassID}: BreakDirect,
	{ClassB2, ClassIN}: BreakIndirect,
	{ClassB2, ClassIS}: BreakProhibited,
	{ClassB2, ClassJL}: BreakDirect,
	{ClassB2, ClassJT}: BreakDirect,
	{ClassB2, ClassJV}: BreakDirect,
	{ClassB2, ClassLF}: BreakProhibited,
	{ClassB2, ClassNL}: BreakProhibited,
	{ClassB2, ClassNS}: BreakIndirect,
	{ClassB2, ClassNU}: BreakDirect,
	{ClassB2, ClassOP}: BreakDirect,
	{ClassB2, ClassPO}: BreakDirect,
	{ClassB2, ClassPR}: BreakDirect,
	{ClassB2, ClassQU}: BreakProhibited,
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
	{ClassBA, ClassAI}: BreakDirect,
	{ClassBA, ClassAK}: BreakDirect,
	{ClassBA, ClassAL}: BreakDirect,
	{ClassBA, ClassAP}: BreakDirect,
	{ClassBA, ClassAS}: BreakDirect,
	{ClassBA, ClassB2}: BreakDirect,
	{ClassBA, ClassBA}: BreakIndirect,
	{ClassBA, ClassBB}: BreakDirect,
	{ClassBA, ClassBK}: BreakProhibited,
	{ClassBA, ClassCB}: BreakDirect,
	{ClassBA, ClassCJ}: BreakIndirect,
	{ClassBA, ClassCL}: BreakProhibited,
	{ClassBA, ClassCM}: BreakIndirect,
	{ClassBA, ClassCP}: BreakProhibited,
	{ClassBA, ClassCR}: BreakProhibited,
	{ClassBA, ClassEB}: BreakDirect,
	{ClassBA, ClassEM}: BreakDirect,
	{ClassBA, ClassEX}: BreakProhibited,
	{ClassBA, ClassGL}: BreakDirect,
	{ClassBA, ClassH2}: BreakDirect,
	{ClassBA, ClassH3}: BreakDirect,
	{ClassBA, ClassHH}: BreakIndirect,
	{ClassBA, ClassHL}: BreakDirect,
	{ClassBA, ClassHY}: BreakIndirect,
	{ClassBA, ClassID}: BreakDirect,
	{ClassBA, ClassIN}: BreakIndirect,
	{ClassBA, ClassIS}: BreakProhibited,
	{ClassBA, ClassJL}: BreakDirect,
	{ClassBA, ClassJT}: BreakDirect,
	{ClassBA, ClassJV}: BreakDirect,
	{ClassBA, ClassLF}: BreakProhibited,
	{ClassBA, ClassNL}: BreakProhibited,
	{ClassBA, ClassNS}: BreakIndirect,
	{ClassBA, ClassNU}: BreakDirect,
	{ClassBA, ClassOP}: BreakDirect,
	{ClassBA, ClassPO}: BreakDirect,
	{ClassBA, ClassPR}: BreakDirect,
	{ClassBA, ClassQU}: BreakProhibited,
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
	{ClassBB, ClassAK}: BreakIndirect,
	{ClassBB, ClassAL}: BreakIndirect,
	{ClassBB, ClassAP}: BreakIndirect,
	{ClassBB, ClassAS}: BreakIndirect,
	{ClassBB, ClassB2}: BreakIndirect,
	{ClassBB, ClassBA}: BreakIndirect,
	{ClassBB, ClassBB}: BreakIndirect,
	{ClassBB, ClassBK}: BreakProhibited,
	{ClassBB, ClassCB}: BreakDirect,
	{ClassBB, ClassCJ}: BreakIndirect,
	{ClassBB, ClassCL}: BreakProhibited,
	{ClassBB, ClassCM}: BreakIndirect,
	{ClassBB, ClassCP}: BreakProhibited,
	{ClassBB, ClassCR}: BreakProhibited,
	{ClassBB, ClassEB}: BreakIndirect,
	{ClassBB, ClassEM}: BreakIndirect,
	{ClassBB, ClassEX}: BreakProhibited,
	{ClassBB, ClassGL}: BreakIndirect,
	{ClassBB, ClassH2}: BreakIndirect,
	{ClassBB, ClassH3}: BreakIndirect,
	{ClassBB, ClassHH}: BreakIndirect,
	{ClassBB, ClassHL}: BreakIndirect,
	{ClassBB, ClassHY}: BreakIndirect,
	{ClassBB, ClassID}: BreakIndirect,
	{ClassBB, ClassIN}: BreakIndirect,
	{ClassBB, ClassIS}: BreakProhibited,
	{ClassBB, ClassJL}: BreakIndirect,
	{ClassBB, ClassJT}: BreakIndirect,
	{ClassBB, ClassJV}: BreakIndirect,
	{ClassBB, ClassLF}: BreakProhibited,
	{ClassBB, ClassNL}: BreakProhibited,
	{ClassBB, ClassNS}: BreakIndirect,
	{ClassBB, ClassNU}: BreakIndirect,
	{ClassBB, ClassOP}: BreakIndirect,
	{ClassBB, ClassPO}: BreakIndirect,
	{ClassBB, ClassPR}: BreakIndirect,
	{ClassBB, ClassQU}: BreakProhibited,
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
	{ClassCB, ClassAK}: BreakDirect,
	{ClassCB, ClassAL}: BreakDirect,
	{ClassCB, ClassAP}: BreakDirect,
	{ClassCB, ClassAS}: BreakDirect,
	{ClassCB, ClassB2}: BreakDirect,
	{ClassCB, ClassBA}: BreakDirect,
	{ClassCB, ClassBB}: BreakDirect,
	{ClassCB, ClassBK}: BreakProhibited,
	{ClassCB, ClassCB}: BreakDirect,
	{ClassCB, ClassCJ}: BreakDirect,
	{ClassCB, ClassCL}: BreakProhibited,
	{ClassCB, ClassCM}: BreakIndirect,
	{ClassCB, ClassCP}: BreakProhibited,
	{ClassCB, ClassCR}: BreakProhibited,
	{ClassCB, ClassEB}: BreakDirect,
	{ClassCB, ClassEM}: BreakDirect,
	{ClassCB, ClassEX}: BreakProhibited,
	{ClassCB, ClassGL}: BreakIndirect,
	{ClassCB, ClassH2}: BreakDirect,
	{ClassCB, ClassH3}: BreakDirect,
	{ClassCB, ClassHH}: BreakDirect,
	{ClassCB, ClassHL}: BreakDirect,
	{ClassCB, ClassHY}: BreakDirect,
	{ClassCB, ClassID}: BreakDirect,
	{ClassCB, ClassIN}: BreakDirect,
	{ClassCB, ClassIS}: BreakProhibited,
	{ClassCB, ClassJL}: BreakDirect,
	{ClassCB, ClassJT}: BreakDirect,
	{ClassCB, ClassJV}: BreakDirect,
	{ClassCB, ClassLF}: BreakProhibited,
	{ClassCB, ClassNL}: BreakProhibited,
	{ClassCB, ClassNS}: BreakDirect,
	{ClassCB, ClassNU}: BreakDirect,
	{ClassCB, ClassOP}: BreakDirect,
	{ClassCB, ClassPO}: BreakDirect,
	{ClassCB, ClassPR}: BreakDirect,
	{ClassCB, ClassQU}: BreakProhibited,
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
	{ClassCJ, ClassAK}: BreakDirect,
	{ClassCJ, ClassAL}: BreakDirect,
	{ClassCJ, ClassAP}: BreakDirect,
	{ClassCJ, ClassAS}: BreakDirect,
	{ClassCJ, ClassB2}: BreakDirect,
	{ClassCJ, ClassBA}: BreakIndirect,
	{ClassCJ, ClassBB}: BreakDirect,
	{ClassCJ, ClassBK}: BreakProhibited,
	{ClassCJ, ClassCB}: BreakDirect,
	{ClassCJ, ClassCJ}: BreakIndirect,
	{ClassCJ, ClassCL}: BreakProhibited,
	{ClassCJ, ClassCM}: BreakIndirect,
	{ClassCJ, ClassCP}: BreakProhibited,
	{ClassCJ, ClassCR}: BreakProhibited,
	{ClassCJ, ClassEB}: BreakDirect,
	{ClassCJ, ClassEM}: BreakDirect,
	{ClassCJ, ClassEX}: BreakProhibited,
	{ClassCJ, ClassGL}: BreakIndirect,
	{ClassCJ, ClassH2}: BreakDirect,
	{ClassCJ, ClassH3}: BreakDirect,
	{ClassCJ, ClassHH}: BreakIndirect,
	{ClassCJ, ClassHL}: BreakDirect,
	{ClassCJ, ClassHY}: BreakIndirect,
	{ClassCJ, ClassID}: BreakDirect,
	{ClassCJ, ClassIN}: BreakIndirect,
	{ClassCJ, ClassIS}: BreakProhibited,
	{ClassCJ, ClassJL}: BreakDirect,
	{ClassCJ, ClassJT}: BreakDirect,
	{ClassCJ, ClassJV}: BreakDirect,
	{ClassCJ, ClassLF}: BreakProhibited,
	{ClassCJ, ClassNL}: BreakProhibited,
	{ClassCJ, ClassNS}: BreakIndirect,
	{ClassCJ, ClassNU}: BreakDirect,
	{ClassCJ, ClassOP}: BreakDirect,
	{ClassCJ, ClassPO}: BreakDirect,
	{ClassCJ, ClassPR}: BreakDirect,
	{ClassCJ, ClassQU}: BreakProhibited,
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
	{ClassCL, ClassAK}: BreakDirect,
	{ClassCL, ClassAL}: BreakDirect,
	{ClassCL, ClassAP}: BreakDirect,
	{ClassCL, ClassAS}: BreakDirect,
	{ClassCL, ClassB2}: BreakDirect,
	{ClassCL, ClassBA}: BreakIndirect,
	{ClassCL, ClassBB}: BreakDirect,
	{ClassCL, ClassBK}: BreakProhibited,
	{ClassCL, ClassCB}: BreakDirect,
	{ClassCL, ClassCJ}: BreakProhibited,
	{ClassCL, ClassCL}: BreakProhibited,
	{ClassCL, ClassCM}: BreakIndirect,
	{ClassCL, ClassCP}: BreakProhibited,
	{ClassCL, ClassCR}: BreakProhibited,
	{ClassCL, ClassEB}: BreakDirect,
	{ClassCL, ClassEM}: BreakDirect,
	{ClassCL, ClassEX}: BreakProhibited,
	{ClassCL, ClassGL}: BreakIndirect,
	{ClassCL, ClassH2}: BreakDirect,
	{ClassCL, ClassH3}: BreakDirect,
	{ClassCL, ClassHH}: BreakIndirect,
	{ClassCL, ClassHL}: BreakDirect,
	{ClassCL, ClassHY}: BreakIndirect,
	{ClassCL, ClassID}: BreakDirect,
	{ClassCL, ClassIN}: BreakIndirect,
	{ClassCL, ClassIS}: BreakProhibited,
	{ClassCL, ClassJL}: BreakDirect,
	{ClassCL, ClassJT}: BreakDirect,
	{ClassCL, ClassJV}: BreakDirect,
	{ClassCL, ClassLF}: BreakProhibited,
	{ClassCL, ClassNL}: BreakProhibited,
	{ClassCL, ClassNS}: BreakProhibited,
	{ClassCL, ClassNU}: BreakDirect,
	{ClassCL, ClassOP}: BreakDirect,
	{ClassCL, ClassPO}: BreakDirect,
	{ClassCL, ClassPR}: BreakDirect,
	{ClassCL, ClassQU}: BreakProhibited,
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
	{ClassCM, ClassAI}: BreakIndirect,
	{ClassCM, ClassAK}: BreakDirect,
	{ClassCM, ClassAL}: BreakIndirect,
	{ClassCM, ClassAP}: BreakDirect,
	{ClassCM, ClassAS}: BreakDirect,
	{ClassCM, ClassB2}: BreakDirect,
	{ClassCM, ClassBA}: BreakIndirect,
	{ClassCM, ClassBB}: BreakDirect,
	{ClassCM, ClassBK}: BreakProhibited,
	{ClassCM, ClassCB}: BreakDirect,
	{ClassCM, ClassCJ}: BreakIndirect,
	{ClassCM, ClassCL}: BreakProhibited,
	{ClassCM, ClassCM}: BreakIndirect,
	{ClassCM, ClassCP}: BreakProhibited,
	{ClassCM, ClassCR}: BreakProhibited,
	{ClassCM, ClassEB}: BreakDirect,
	{ClassCM, ClassEM}: BreakDirect,
	{ClassCM, ClassEX}: BreakProhibited,
	{ClassCM, ClassGL}: BreakIndirect,
	{ClassCM, ClassH2}: BreakDirect,
	{ClassCM, ClassH3}: BreakDirect,
	{ClassCM, ClassHH}: BreakIndirect,
	{ClassCM, ClassHL}: BreakIndirect,
	{ClassCM, ClassHY}: BreakIndirect,
	{ClassCM, ClassID}: BreakDirect,
	{ClassCM, ClassIN}: BreakIndirect,
	{ClassCM, ClassIS}: BreakProhibited,
	{ClassCM, ClassJL}: BreakDirect,
	{ClassCM, ClassJT}: BreakDirect,
	{ClassCM, ClassJV}: BreakDirect,
	{ClassCM, ClassLF}: BreakProhibited,
	{ClassCM, ClassNL}: BreakProhibited,
	{ClassCM, ClassNS}: BreakIndirect,
	{ClassCM, ClassNU}: BreakIndirect,
	{ClassCM, ClassOP}: BreakDirect,
	{ClassCM, ClassPO}: BreakIndirect,
	{ClassCM, ClassPR}: BreakIndirect,
	{ClassCM, ClassQU}: BreakProhibited,
	{ClassCM, ClassRI}: BreakDirect,
	{ClassCM, ClassSA}: BreakIndirect,
	{ClassCM, ClassSP}: BreakProhibited,
	{ClassCM, ClassSY}: BreakProhibited,
	{ClassCM, ClassVF}: BreakDirect,
	{ClassCM, ClassVI}: BreakDirect,
	{ClassCM, ClassWJ}: BreakProhibited,
	{ClassCM, ClassXX}: BreakIndirect,
	{ClassCM, ClassZW}: BreakProhibited,
	{ClassCM, ClassZWJ}: BreakIndirect,
	{ClassCP, ClassAI}: BreakIndirect,
	{ClassCP, ClassAK}: BreakDirect,
	{ClassCP, ClassAL}: BreakIndirect,
	{ClassCP, ClassAP}: BreakDirect,
	{ClassCP, ClassAS}: BreakDirect,
	{ClassCP, ClassB2}: BreakDirect,
	{ClassCP, ClassBA}: BreakIndirect,
	{ClassCP, ClassBB}: BreakDirect,
	{ClassCP, ClassBK}: BreakProhibited,
	{ClassCP, ClassCB}: BreakDirect,
	{ClassCP, ClassCJ}: BreakProhibited,
	{ClassCP, ClassCL}: BreakProhibited,
	{ClassCP, ClassCM}: BreakIndirect,
	{ClassCP, ClassCP}: BreakProhibited,
	{ClassCP, ClassCR}: BreakProhibited,
	{ClassCP, ClassEB}: BreakDirect,
	{ClassCP, ClassEM}: BreakDirect,
	{ClassCP, ClassEX}: BreakProhibited,
	{ClassCP, ClassGL}: BreakIndirect,
	{ClassCP, ClassH2}: BreakDirect,
	{ClassCP, ClassH3}: BreakDirect,
	{ClassCP, ClassHH}: BreakIndirect,
	{ClassCP, ClassHL}: BreakIndirect,
	{ClassCP, ClassHY}: BreakIndirect,
	{ClassCP, ClassID}: BreakDirect,
	{ClassCP, ClassIN}: BreakIndirect,
	{ClassCP, ClassIS}: BreakProhibited,
	{ClassCP, ClassJL}: BreakDirect,
	{ClassCP, ClassJT}: BreakDirect,
	{ClassCP, ClassJV}: BreakDirect,
	{ClassCP, ClassLF}: BreakProhibited,
	{ClassCP, ClassNL}: BreakProhibited,
	{ClassCP, ClassNS}: BreakProhibited,
	{ClassCP, ClassNU}: BreakIndirect,
	{ClassCP, ClassOP}: BreakDirect,
	{ClassCP, ClassPO}: BreakDirect,
	{ClassCP, ClassPR}: BreakDirect,
	{ClassCP, ClassQU}: BreakProhibited,
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
	{ClassEB, ClassAK}: BreakDirect,
	{ClassEB, ClassAL}: BreakDirect,
	{ClassEB, ClassAP}: BreakDirect,
	{ClassEB, ClassAS}: BreakDirect,
	{ClassEB, ClassB2}: BreakDirect,
	{ClassEB, ClassBA}: BreakIndirect,
	{ClassEB, ClassBB}: BreakDirect,
	{ClassEB, ClassBK}: BreakProhibited,
	{ClassEB, ClassCB}: BreakDirect,
	{ClassEB, ClassCJ}: BreakIndirect,
	{ClassEB, ClassCL}: BreakProhibited,
	{ClassEB, ClassCM}: BreakIndirect,
	{ClassEB, ClassCP}: BreakProhibited,
	{ClassEB, ClassCR}: BreakProhibited,
	{ClassEB, ClassEB}: BreakDirect,
	{ClassEB, ClassEM}: BreakIndirect,
	{ClassEB, ClassEX}: BreakProhibited,
	{ClassEB, ClassGL}: BreakIndirect,
	{ClassEB, ClassH2}: BreakDirect,
	{ClassEB, ClassH3}: BreakDirect,
	{ClassEB, ClassHH}: BreakIndirect,
	{ClassEB, ClassHL}: BreakDirect,
	{ClassEB, ClassHY}: BreakIndirect,
	{ClassEB, ClassID}: BreakDirect,
	{ClassEB, ClassIN}: BreakIndirect,
	{ClassEB, ClassIS}: BreakProhibited,
	{ClassEB, ClassJL}: BreakDirect,
	{ClassEB, ClassJT}: BreakDirect,
	{ClassEB, ClassJV}: BreakDirect,
	{ClassEB, ClassLF}: BreakProhibited,
	{ClassEB, ClassNL}: BreakProhibited,
	{ClassEB, ClassNS}: BreakIndirect,
	{ClassEB, ClassNU}: BreakDirect,
	{ClassEB, ClassOP}: BreakDirect,
	{ClassEB, ClassPO}: BreakIndirect,
	{ClassEB, ClassPR}: BreakDirect,
	{ClassEB, ClassQU}: BreakProhibited,
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
	{ClassEM, ClassAI}: BreakDirect,
	{ClassEM, ClassAK}: BreakDirect,
	{ClassEM, ClassAL}: BreakDirect,
	{ClassEM, ClassAP}: BreakDirect,
	{ClassEM, ClassAS}: BreakDirect,
	{ClassEM, ClassB2}: BreakDirect,
	{ClassEM, ClassBA}: BreakIndirect,
	{ClassEM, ClassBB}: BreakDirect,
	{ClassEM, ClassBK}: BreakProhibited,
	{ClassEM, ClassCB}: BreakDirect,
	{ClassEM, ClassCJ}: BreakIndirect,
	{ClassEM, ClassCL}: BreakProhibited,
	{ClassEM, ClassCM}: BreakIndirect,
	{ClassEM, ClassCP}: BreakProhibited,
	{ClassEM, ClassCR}: BreakProhibited,
	{ClassEM, ClassEB}: BreakDirect,
	{ClassEM, ClassEM}: BreakDirect,
	{ClassEM, ClassEX}: BreakProhibited,
	{ClassEM, ClassGL}: BreakIndirect,
	{ClassEM, ClassH2}: BreakDirect,
	{ClassEM, ClassH3}: BreakDirect,
	{ClassEM, ClassHH}: BreakIndirect,
	{ClassEM, ClassHL}: BreakDirect,
	{ClassEM, ClassHY}: BreakIndirect,
	{ClassEM, ClassID}: BreakDirect,
	{ClassEM, ClassIN}: BreakIndirect,
	{ClassEM, ClassIS}: BreakProhibited,
	{ClassEM, ClassJL}: BreakDirect,
	{ClassEM, ClassJT}: BreakDirect,
	{ClassEM, ClassJV}: BreakDirect,
	{ClassEM, ClassLF}: BreakProhibited,
	{ClassEM, ClassNL}: BreakProhibited,
	{ClassEM, ClassNS}: BreakIndirect,
	{ClassEM, ClassNU}: BreakDirect,
	{ClassEM, ClassOP}: BreakDirect,
	{ClassEM, ClassPO}: BreakIndirect,
	{ClassEM, ClassPR}: BreakDirect,
	{ClassEM, ClassQU}: BreakProhibited,
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
	{ClassEX, ClassAK}: BreakDirect,
	{ClassEX, ClassAL}: BreakDirect,
	{ClassEX, ClassAP}: BreakDirect,
	{ClassEX, ClassAS}: BreakDirect,
	{ClassEX, ClassB2}: BreakDirect,
	{ClassEX, ClassBA}: BreakIndirect,
	{ClassEX, ClassBB}: BreakDirect,
	{ClassEX, ClassBK}: BreakProhibited,
	{ClassEX, ClassCB}: BreakDirect,
	{ClassEX, ClassCJ}: BreakIndirect,
	{ClassEX, ClassCL}: BreakProhibited,
	{ClassEX, ClassCM}: BreakIndirect,
	{ClassEX, ClassCP}: BreakProhibited,
	{ClassEX, ClassCR}: BreakProhibited,
	{ClassEX, ClassEB}: BreakDirect,
	{ClassEX, ClassEM}: BreakDirect,
	{ClassEX, ClassEX}: BreakProhibited,
	{ClassEX, ClassGL}: BreakIndirect,
	{ClassEX, ClassH2}: BreakDirect,
	{ClassEX, ClassH3}: BreakDirect,
	{ClassEX, ClassHH}: BreakIndirect,
	{ClassEX, ClassHL}: BreakDirect,
	{ClassEX, ClassHY}: BreakIndirect,
	{ClassEX, ClassID}: BreakDirect,
	{ClassEX, ClassIN}: BreakIndirect,
	{ClassEX, ClassIS}: BreakProhibited,
	{ClassEX, ClassJL}: BreakDirect,
	{ClassEX, ClassJT}: BreakDirect,
	{ClassEX, ClassJV}: BreakDirect,
	{ClassEX, ClassLF}: BreakProhibited,
	{ClassEX, ClassNL}: BreakProhibited,
	{ClassEX, ClassNS}: BreakIndirect,
	{ClassEX, ClassNU}: BreakDirect,
	{ClassEX, ClassOP}: BreakDirect,
	{ClassEX, ClassPO}: BreakDirect,
	{ClassEX, ClassPR}: BreakDirect,
	{ClassEX, ClassQU}: BreakProhibited,
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
	{ClassGL, ClassAI}: BreakIndirect,
	{ClassGL, ClassAK}: BreakIndirect,
	{ClassGL, ClassAL}: BreakIndirect,
	{ClassGL, ClassAP}: BreakIndirect,
	{ClassGL, ClassAS}: BreakIndirect,
	{ClassGL, ClassB2}: BreakIndirect,
	{ClassGL, ClassBA}: BreakIndirect,
	{ClassGL, ClassBB}: BreakIndirect,
	{ClassGL, ClassBK}: BreakProhibited,
	{ClassGL, ClassCB}: BreakIndirect,
	{ClassGL, ClassCJ}: BreakIndirect,
	{ClassGL, ClassCL}: BreakProhibited,
	{ClassGL, ClassCM}: BreakIndirect,
	{ClassGL, ClassCP}: BreakProhibited,
	{ClassGL, ClassCR}: BreakProhibited,
	{ClassGL, ClassEB}: BreakIndirect,
	{ClassGL, ClassEM}: BreakIndirect,
	{ClassGL, ClassEX}: BreakProhibited,
	{ClassGL, ClassGL}: BreakIndirect,
	{ClassGL, ClassH2}: BreakIndirect,
	{ClassGL, ClassH3}: BreakIndirect,
	{ClassGL, ClassHH}: BreakIndirect,
	{ClassGL, ClassHL}: BreakIndirect,
	{ClassGL, ClassHY}: BreakIndirect,
	{ClassGL, ClassID}: BreakIndirect,
	{ClassGL, ClassIN}: BreakIndirect,
	{ClassGL, ClassIS}: BreakProhibited,
	{ClassGL, ClassJL}: BreakIndirect,
	{ClassGL, ClassJT}: BreakIndirect,
	{ClassGL, ClassJV}: BreakIndirect,
	{ClassGL, ClassLF}: BreakProhibited,
	{ClassGL, ClassNL}: BreakProhibited,
	{ClassGL, ClassNS}: BreakIndirect,
	{ClassGL, ClassNU}: BreakIndirect,
	{ClassGL, ClassOP}: BreakIndirect,
	{ClassGL, ClassPO}: BreakIndirect,
	{ClassGL, ClassPR}: BreakIndirect,
	{ClassGL, ClassQU}: BreakProhibited,
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
	{ClassH2, ClassAI}: BreakDirect,
	{ClassH2, ClassAK}: BreakDirect,
	{ClassH2, ClassAL}: BreakDirect,
	{ClassH2, ClassAP}: BreakDirect,
	{ClassH2, ClassAS}: BreakDirect,
	{ClassH2, ClassB2}: BreakDirect,
	{ClassH2, ClassBA}: BreakIndirect,
	{ClassH2, ClassBB}: BreakDirect,
	{ClassH2, ClassBK}: BreakProhibited,
	{ClassH2, ClassCB}: BreakDirect,
	{ClassH2, ClassCJ}: BreakIndirect,
	{ClassH2, ClassCL}: BreakProhibited,
	{ClassH2, ClassCM}: BreakIndirect,
	{ClassH2, ClassCP}: BreakProhibited,
	{ClassH2, ClassCR}: BreakProhibited,
	{ClassH2, ClassEB}: BreakDirect,
	{ClassH2, ClassEM}: BreakDirect,
	{ClassH2, ClassEX}: BreakProhibited,
	{ClassH2, ClassGL}: BreakIndirect,
	{ClassH2, ClassH2}: BreakDirect,
	{ClassH2, ClassH3}: BreakDirect,
	{ClassH2, ClassHH}: BreakIndirect,
	{ClassH2, ClassHL}: BreakDirect,
	{ClassH2, ClassHY}: BreakIndirect,
	{ClassH2, ClassID}: BreakDirect,
	{ClassH2, ClassIN}: BreakIndirect,
	{ClassH2, ClassIS}: BreakProhibited,
	{ClassH2, ClassJL}: BreakDirect,
	{ClassH2, ClassJT}: BreakIndirect,
	{ClassH2, ClassJV}: BreakIndirect,
	{ClassH2, ClassLF}: BreakProhibited,
	{ClassH2, ClassNL}: BreakProhibited,
	{ClassH2, ClassNS}: BreakIndirect,
	{ClassH2, ClassNU}: BreakDirect,
	{ClassH2, ClassOP}: BreakDirect,
	{ClassH2, ClassPO}: BreakIndirect,
	{ClassH2, ClassPR}: BreakDirect,
	{ClassH2, ClassQU}: BreakProhibited,
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
	{ClassH3, ClassAK}: BreakDirect,
	{ClassH3, ClassAL}: BreakDirect,
	{ClassH3, ClassAP}: BreakDirect,
	{ClassH3, ClassAS}: BreakDirect,
	{ClassH3, ClassB2}: BreakDirect,
	{ClassH3, ClassBA}: BreakIndirect,
	{ClassH3, ClassBB}: BreakDirect,
	{ClassH3, ClassBK}: BreakProhibited,
	{ClassH3, ClassCB}: BreakDirect,
	{ClassH3, ClassCJ}: BreakIndirect,
	{ClassH3, ClassCL}: BreakProhibited,
	{ClassH3, ClassCM}: BreakIndirect,
	{ClassH3, ClassCP}: BreakProhibited,
	{ClassH3, ClassCR}: BreakProhibited,
	{ClassH3, ClassEB}: BreakDirect,
	{ClassH3, ClassEM}: BreakDirect,
	{ClassH3, ClassEX}: BreakProhibited,
	{ClassH3, ClassGL}: BreakIndirect,
	{ClassH3, ClassH2}: BreakDirect,
	{ClassH3, ClassH3}: BreakDirect,
	{ClassH3, ClassHH}: BreakIndirect,
	{ClassH3, ClassHL}: BreakDirect,
	{ClassH3, ClassHY}: BreakIndirect,
	{ClassH3, ClassID}: BreakDirect,
	{ClassH3, ClassIN}: BreakIndirect,
	{ClassH3, ClassIS}: BreakProhibited,
	{ClassH3, ClassJL}: BreakDirect,
	{ClassH3, ClassJT}: BreakIndirect,
	{ClassH3, ClassJV}: BreakDirect,
	{ClassH3, ClassLF}: BreakProhibited,
	{ClassH3, ClassNL}: BreakProhibited,
	{ClassH3, ClassNS}: BreakIndirect,
	{ClassH3, ClassNU}: BreakDirect,
	{ClassH3, ClassOP}: BreakDirect,
	{ClassH3, ClassPO}: BreakIndirect,
	{ClassH3, ClassPR}: BreakDirect,
	{ClassH3, ClassQU}: BreakProhibited,
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
	{ClassHH, ClassAK}: BreakDirect,
	{ClassHH, ClassAL}: BreakIndirect,
	{ClassHH, ClassAP}: BreakDirect,
	{ClassHH, ClassAS}: BreakDirect,
	{ClassHH, ClassB2}: BreakDirect,
	{ClassHH, ClassBA}: BreakIndirect,
	{ClassHH, ClassBB}: BreakDirect,
	{ClassHH, ClassBK}: BreakProhibited,
	{ClassHH, ClassCB}: BreakDirect,
	{ClassHH, ClassCJ}: BreakIndirect,
	{ClassHH, ClassCL}: BreakProhibited,
	{ClassHH, ClassCM}: BreakIndirect,
	{ClassHH, ClassCP}: BreakProhibited,
	{ClassHH, ClassCR}: BreakProhibited,
	{ClassHH, ClassEB}: BreakDirect,
	{ClassHH, ClassEM}: BreakDirect,
	{ClassHH, ClassEX}: BreakProhibited,
	{ClassHH, ClassGL}: BreakDirect,
	{ClassHH, ClassH2}: BreakDirect,
	{ClassHH, ClassH3}: BreakDirect,
	{ClassHH, ClassHH}: BreakIndirect,
	{ClassHH, ClassHL}: BreakIndirect,
	{ClassHH, ClassHY}: BreakIndirect,
	{ClassHH, ClassID}: BreakDirect,
	{ClassHH, ClassIN}: BreakIndirect,
	{ClassHH, ClassIS}: BreakProhibited,
	{ClassHH, ClassJL}: BreakDirect,
	{ClassHH, ClassJT}: BreakDirect,
	{ClassHH, ClassJV}: BreakDirect,
	{ClassHH, ClassLF}: BreakProhibited,
	{ClassHH, ClassNL}: BreakProhibited,
	{ClassHH, ClassNS}: BreakIndirect,
	{ClassHH, ClassNU}: BreakDirect,
	{ClassHH, ClassOP}: BreakDirect,
	{ClassHH, ClassPO}: BreakDirect,
	{ClassHH, ClassPR}: BreakDirect,
	{ClassHH, ClassQU}: BreakProhibited,
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
	{ClassHL, ClassAK}: BreakDirect,
	{ClassHL, ClassAL}: BreakIndirect,
	{ClassHL, ClassAP}: BreakDirect,
	{ClassHL, ClassAS}: BreakDirect,
	{ClassHL, ClassB2}: BreakDirect,
	{ClassHL, ClassBA}: BreakIndirect,
	{ClassHL, ClassBB}: BreakDirect,
	{ClassHL, ClassBK}: BreakProhibited,
	{ClassHL, ClassCB}: BreakDirect,
	{ClassHL, ClassCJ}: BreakIndirect,
	{ClassHL, ClassCL}: BreakProhibited,
	{ClassHL, ClassCM}: BreakIndirect,
	{ClassHL, ClassCP}: BreakProhibited,
	{ClassHL, ClassCR}: BreakProhibited,
	{ClassHL, ClassEB}: BreakDirect,
	{ClassHL, ClassEM}: BreakDirect,
	{ClassHL, ClassEX}: BreakProhibited,
	{ClassHL, ClassGL}: BreakIndirect,
	{ClassHL, ClassH2}: BreakDirect,
	{ClassHL, ClassH3}: BreakDirect,
	{ClassHL, ClassHH}: BreakIndirect,
	{ClassHL, ClassHL}: BreakIndirect,
	{ClassHL, ClassHY}: BreakIndirect,
	{ClassHL, ClassID}: BreakDirect,
	{ClassHL, ClassIN}: BreakIndirect,
	{ClassHL, ClassIS}: BreakProhibited,
	{ClassHL, ClassJL}: BreakDirect,
	{ClassHL, ClassJT}: BreakDirect,
	{ClassHL, ClassJV}: BreakDirect,
	{ClassHL, ClassLF}: BreakProhibited,
	{ClassHL, ClassNL}: BreakProhibited,
	{ClassHL, ClassNS}: BreakIndirect,
	{ClassHL, ClassNU}: BreakIndirect,
	{ClassHL, ClassOP}: BreakDirect,
	{ClassHL, ClassPO}: BreakIndirect,
	{ClassHL, ClassPR}: BreakIndirect,
	{ClassHL, ClassQU}: BreakProhibited,
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
	{ClassHY, ClassAK}: BreakDirect,
	{ClassHY, ClassAL}: BreakIndirect,
	{ClassHY, ClassAP}: BreakDirect,
	{ClassHY, ClassAS}: BreakDirect,
	{ClassHY, ClassB2}: BreakDirect,
	{ClassHY, ClassBA}: BreakIndirect,
	{ClassHY, ClassBB}: BreakDirect,
	{ClassHY, ClassBK}: BreakProhibited,
	{ClassHY, ClassCB}: BreakDirect,
	{ClassHY, ClassCJ}: BreakIndirect,
	{ClassHY, ClassCL}: BreakProhibited,
	{ClassHY, ClassCM}: BreakIndirect,
	{ClassHY, ClassCP}: BreakProhibited,
	{ClassHY, ClassCR}: BreakProhibited,
	{ClassHY, ClassEB}: BreakDirect,
	{ClassHY, ClassEM}: BreakDirect,
	{ClassHY, ClassEX}: BreakProhibited,
	{ClassHY, ClassGL}: BreakDirect,
	{ClassHY, ClassH2}: BreakDirect,
	{ClassHY, ClassH3}: BreakDirect,
	{ClassHY, ClassHH}: BreakIndirect,
	{ClassHY, ClassHL}: BreakIndirect,
	{ClassHY, ClassHY}: BreakIndirect,
	{ClassHY, ClassID}: BreakDirect,
	{ClassHY, ClassIN}: BreakIndirect,
	{ClassHY, ClassIS}: BreakProhibited,
	{ClassHY, ClassJL}: BreakDirect,
	{ClassHY, ClassJT}: BreakDirect,
	{ClassHY, ClassJV}: BreakDirect,
	{ClassHY, ClassLF}: BreakProhibited,
	{ClassHY, ClassNL}: BreakProhibited,
	{ClassHY, ClassNS}: BreakIndirect,
	{ClassHY, ClassNU}: BreakIndirect,
	{ClassHY, ClassOP}: BreakDirect,
	{ClassHY, ClassPO}: BreakDirect,
	{ClassHY, ClassPR}: BreakDirect,
	{ClassHY, ClassQU}: BreakProhibited,
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
	{ClassID, ClassAK}: BreakDirect,
	{ClassID, ClassAL}: BreakDirect,
	{ClassID, ClassAP}: BreakDirect,
	{ClassID, ClassAS}: BreakDirect,
	{ClassID, ClassB2}: BreakDirect,
	{ClassID, ClassBA}: BreakIndirect,
	{ClassID, ClassBB}: BreakDirect,
	{ClassID, ClassBK}: BreakProhibited,
	{ClassID, ClassCB}: BreakDirect,
	{ClassID, ClassCJ}: BreakIndirect,
	{ClassID, ClassCL}: BreakProhibited,
	{ClassID, ClassCM}: BreakIndirect,
	{ClassID, ClassCP}: BreakProhibited,
	{ClassID, ClassCR}: BreakProhibited,
	{ClassID, ClassEB}: BreakDirect,
	{ClassID, ClassEM}: BreakDirect,
	{ClassID, ClassEX}: BreakProhibited,
	{ClassID, ClassGL}: BreakIndirect,
	{ClassID, ClassH2}: BreakDirect,
	{ClassID, ClassH3}: BreakDirect,
	{ClassID, ClassHH}: BreakIndirect,
	{ClassID, ClassHL}: BreakDirect,
	{ClassID, ClassHY}: BreakIndirect,
	{ClassID, ClassID}: BreakDirect,
	{ClassID, ClassIN}: BreakIndirect,
	{ClassID, ClassIS}: BreakProhibited,
	{ClassID, ClassJL}: BreakDirect,
	{ClassID, ClassJT}: BreakDirect,
	{ClassID, ClassJV}: BreakDirect,
	{ClassID, ClassLF}: BreakProhibited,
	{ClassID, ClassNL}: BreakProhibited,
	{ClassID, ClassNS}: BreakIndirect,
	{ClassID, ClassNU}: BreakDirect,
	{ClassID, ClassOP}: BreakDirect,
	{ClassID, ClassPO}: BreakIndirect,
	{ClassID, ClassPR}: BreakDirect,
	{ClassID, ClassQU}: BreakProhibited,
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
	{ClassIN, ClassAI}: BreakDirect,
	{ClassIN, ClassAK}: BreakDirect,
	{ClassIN, ClassAL}: BreakDirect,
	{ClassIN, ClassAP}: BreakDirect,
	{ClassIN, ClassAS}: BreakDirect,
	{ClassIN, ClassB2}: BreakDirect,
	{ClassIN, ClassBA}: BreakIndirect,
	{ClassIN, ClassBB}: BreakDirect,
	{ClassIN, ClassBK}: BreakProhibited,
	{ClassIN, ClassCB}: BreakDirect,
	{ClassIN, ClassCJ}: BreakIndirect,
	{ClassIN, ClassCL}: BreakProhibited,
	{ClassIN, ClassCM}: BreakIndirect,
	{ClassIN, ClassCP}: BreakProhibited,
	{ClassIN, ClassCR}: BreakProhibited,
	{ClassIN, ClassEB}: BreakDirect,
	{ClassIN, ClassEM}: BreakDirect,
	{ClassIN, ClassEX}: BreakProhibited,
	{ClassIN, ClassGL}: BreakIndirect,
	{ClassIN, ClassH2}: BreakDirect,
	{ClassIN, ClassH3}: BreakDirect,
	{ClassIN, ClassHH}: BreakIndirect,
	{ClassIN, ClassHL}: BreakDirect,
	{ClassIN, ClassHY}: BreakIndirect,
	{ClassIN, ClassID}: BreakDirect,
	{ClassIN, ClassIN}: BreakIndirect,
	{ClassIN, ClassIS}: BreakProhibited,
	{ClassIN, ClassJL}: BreakDirect,
	{ClassIN, ClassJT}: BreakDirect,
	{ClassIN, ClassJV}: BreakDirect,
	{ClassIN, ClassLF}: BreakProhibited,
	{ClassIN, ClassNL}: BreakProhibited,
	{ClassIN, ClassNS}: BreakIndirect,
	{ClassIN, ClassNU}: BreakDirect,
	{ClassIN, ClassOP}: BreakDirect,
	{ClassIN, ClassPO}: BreakDirect,
	{ClassIN, ClassPR}: BreakDirect,
	{ClassIN, ClassQU}: BreakProhibited,
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
	{ClassIS, ClassAI}: BreakIndirect,
	{ClassIS, ClassAK}: BreakDirect,
	{ClassIS, ClassAL}: BreakIndirect,
	{ClassIS, ClassAP}: BreakDirect,
	{ClassIS, ClassAS}: BreakDirect,
	{ClassIS, ClassB2}: BreakDirect,
	{ClassIS, ClassBA}: BreakIndirect,
	{ClassIS, ClassBB}: BreakDirect,
	{ClassIS, ClassBK}: BreakProhibited,
	{ClassIS, ClassCB}: BreakDirect,
	{ClassIS, ClassCJ}: BreakIndirect,
	{ClassIS, ClassCL}: BreakProhibited,
	{ClassIS, ClassCM}: BreakIndirect,
	{ClassIS, ClassCP}: BreakProhibited,
	{ClassIS, ClassCR}: BreakProhibited,
	{ClassIS, ClassEB}: BreakDirect,
	{ClassIS, ClassEM}: BreakDirect,
	{ClassIS, ClassEX}: BreakProhibited,
	{ClassIS, ClassGL}: BreakIndirect,
	{ClassIS, ClassH2}: BreakDirect,
	{ClassIS, ClassH3}: BreakDirect,
	{ClassIS, ClassHH}: BreakIndirect,
	{ClassIS, ClassHL}: BreakIndirect,
	{ClassIS, ClassHY}: BreakIndirect,
	{ClassIS, ClassID}: BreakDirect,
	{ClassIS, ClassIN}: BreakIndirect,
	{ClassIS, ClassIS}: BreakProhibited,
	{ClassIS, ClassJL}: BreakDirect,
	{ClassIS, ClassJT}: BreakDirect,
	{ClassIS, ClassJV}: BreakDirect,
	{ClassIS, ClassLF}: BreakProhibited,
	{ClassIS, ClassNL}: BreakProhibited,
	{ClassIS, ClassNS}: BreakIndirect,
	{ClassIS, ClassNU}: BreakIndirect,
	{ClassIS, ClassOP}: BreakDirect,
	{ClassIS, ClassPO}: BreakDirect,
	{ClassIS, ClassPR}: BreakDirect,
	{ClassIS, ClassQU}: BreakProhibited,
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
	{ClassJL, ClassAK}: BreakDirect,
	{ClassJL, ClassAL}: BreakDirect,
	{ClassJL, ClassAP}: BreakDirect,
	{ClassJL, ClassAS}: BreakDirect,
	{ClassJL, ClassB2}: BreakDirect,
	{ClassJL, ClassBA}: BreakIndirect,
	{ClassJL, ClassBB}: BreakDirect,
	{ClassJL, ClassBK}: BreakProhibited,
	{ClassJL, ClassCB}: BreakDirect,
	{ClassJL, ClassCJ}: BreakIndirect,
	{ClassJL, ClassCL}: BreakProhibited,
	{ClassJL, ClassCM}: BreakIndirect,
	{ClassJL, ClassCP}: BreakProhibited,
	{ClassJL, ClassCR}: BreakProhibited,
	{ClassJL, ClassEB}: BreakDirect,
	{ClassJL, ClassEM}: BreakDirect,
	{ClassJL, ClassEX}: BreakProhibited,
	{ClassJL, ClassGL}: BreakIndirect,
	{ClassJL, ClassH2}: BreakIndirect,
	{ClassJL, ClassH3}: BreakIndirect,
	{ClassJL, ClassHH}: BreakIndirect,
	{ClassJL, ClassHL}: BreakDirect,
	{ClassJL, ClassHY}: BreakIndirect,
	{ClassJL, ClassID}: BreakDirect,
	{ClassJL, ClassIN}: BreakIndirect,
	{ClassJL, ClassIS}: BreakProhibited,
	{ClassJL, ClassJL}: BreakIndirect,
	{ClassJL, ClassJT}: BreakDirect,
	{ClassJL, ClassJV}: BreakIndirect,
	{ClassJL, ClassLF}: BreakProhibited,
	{ClassJL, ClassNL}: BreakProhibited,
	{ClassJL, ClassNS}: BreakIndirect,
	{ClassJL, ClassNU}: BreakDirect,
	{ClassJL, ClassOP}: BreakDirect,
	{ClassJL, ClassPO}: BreakIndirect,
	{ClassJL, ClassPR}: BreakDirect,
	{ClassJL, ClassQU}: BreakProhibited,
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
	{ClassJT, ClassAK}: BreakDirect,
	{ClassJT, ClassAL}: BreakDirect,
	{ClassJT, ClassAP}: BreakDirect,
	{ClassJT, ClassAS}: BreakDirect,
	{ClassJT, ClassB2}: BreakDirect,
	{ClassJT, ClassBA}: BreakIndirect,
	{ClassJT, ClassBB}: BreakDirect,
	{ClassJT, ClassBK}: BreakProhibited,
	{ClassJT, ClassCB}: BreakDirect,
	{ClassJT, ClassCJ}: BreakIndirect,
	{ClassJT, ClassCL}: BreakProhibited,
	{ClassJT, ClassCM}: BreakIndirect,
	{ClassJT, ClassCP}: BreakProhibited,
	{ClassJT, ClassCR}: BreakProhibited,
	{ClassJT, ClassEB}: BreakDirect,
	{ClassJT, ClassEM}: BreakDirect,
	{ClassJT, ClassEX}: BreakProhibited,
	{ClassJT, ClassGL}: BreakIndirect,
	{ClassJT, ClassH2}: BreakDirect,
	{ClassJT, ClassH3}: BreakDirect,
	{ClassJT, ClassHH}: BreakIndirect,
	{ClassJT, ClassHL}: BreakDirect,
	{ClassJT, ClassHY}: BreakIndirect,
	{ClassJT, ClassID}: BreakDirect,
	{ClassJT, ClassIN}: BreakIndirect,
	{ClassJT, ClassIS}: BreakProhibited,
	{ClassJT, ClassJL}: BreakDirect,
	{ClassJT, ClassJT}: BreakIndirect,
	{ClassJT, ClassJV}: BreakDirect,
	{ClassJT, ClassLF}: BreakProhibited,
	{ClassJT, ClassNL}: BreakProhibited,
	{ClassJT, ClassNS}: BreakIndirect,
	{ClassJT, ClassNU}: BreakDirect,
	{ClassJT, ClassOP}: BreakDirect,
	{ClassJT, ClassPO}: BreakIndirect,
	{ClassJT, ClassPR}: BreakDirect,
	{ClassJT, ClassQU}: BreakProhibited,
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
	{ClassJV, ClassAK}: BreakDirect,
	{ClassJV, ClassAL}: BreakDirect,
	{ClassJV, ClassAP}: BreakDirect,
	{ClassJV, ClassAS}: BreakDirect,
	{ClassJV, ClassB2}: BreakDirect,
	{ClassJV, ClassBA}: BreakIndirect,
	{ClassJV, ClassBB}: BreakDirect,
	{ClassJV, ClassBK}: BreakProhibited,
	{ClassJV, ClassCB}: BreakDirect,
	{ClassJV, ClassCJ}: BreakIndirect,
	{ClassJV, ClassCL}: BreakProhibited,
	{ClassJV, ClassCM}: BreakIndirect,
	{ClassJV, ClassCP}: BreakProhibited,
	{ClassJV, ClassCR}: BreakProhibited,
	{ClassJV, ClassEB}: BreakDirect,
	{ClassJV, ClassEM}: BreakDirect,
	{ClassJV, ClassEX}: BreakProhibited,
	{ClassJV, ClassGL}: BreakIndirect,
	{ClassJV, ClassH2}: BreakDirect,
	{ClassJV, ClassH3}: BreakDirect,
	{ClassJV, ClassHH}: BreakIndirect,
	{ClassJV, ClassHL}: BreakDirect,
	{ClassJV, ClassHY}: BreakIndirect,
	{ClassJV, ClassID}: BreakDirect,
	{ClassJV, ClassIN}: BreakIndirect,
	{ClassJV, ClassIS}: BreakProhibited,
	{ClassJV, ClassJL}: BreakDirect,
	{ClassJV, ClassJT}: BreakIndirect,
	{ClassJV, ClassJV}: BreakIndirect,
	{ClassJV, ClassLF}: BreakProhibited,
	{ClassJV, ClassNL}: BreakProhibited,
	{ClassJV, ClassNS}: BreakIndirect,
	{ClassJV, ClassNU}: BreakDirect,
	{ClassJV, ClassOP}: BreakDirect,
	{ClassJV, ClassPO}: BreakIndirect,
	{ClassJV, ClassPR}: BreakDirect,
	{ClassJV, ClassQU}: BreakProhibited,
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
	{ClassNS, ClassAK}: BreakDirect,
	{ClassNS, ClassAL}: BreakDirect,
	{ClassNS, ClassAP}: BreakDirect,
	{ClassNS, ClassAS}: BreakDirect,
	{ClassNS, ClassB2}: BreakDirect,
	{ClassNS, ClassBA}: BreakIndirect,
	{ClassNS, ClassBB}: BreakDirect,
	{ClassNS, ClassBK}: BreakProhibited,
	{ClassNS, ClassCB}: BreakDirect,
	{ClassNS, ClassCJ}: BreakIndirect,
	{ClassNS, ClassCL}: BreakProhibited,
	{ClassNS, ClassCM}: BreakIndirect,
	{ClassNS, ClassCP}: BreakProhibited,
	{ClassNS, ClassCR}: BreakProhibited,
	{ClassNS, ClassEB}: BreakDirect,
	{ClassNS, ClassEM}: BreakDirect,
	{ClassNS, ClassEX}: BreakProhibited,
	{ClassNS, ClassGL}: BreakIndirect,
	{ClassNS, ClassH2}: BreakDirect,
	{ClassNS, ClassH3}: BreakDirect,
	{ClassNS, ClassHH}: BreakIndirect,
	{ClassNS, ClassHL}: BreakDirect,
	{ClassNS, ClassHY}: BreakIndirect,
	{ClassNS, ClassID}: BreakDirect,
	{ClassNS, ClassIN}: BreakIndirect,
	{ClassNS, ClassIS}: BreakProhibited,
	{ClassNS, ClassJL}: BreakDirect,
	{ClassNS, ClassJT}: BreakDirect,
	{ClassNS, ClassJV}: BreakDirect,
	{ClassNS, ClassLF}: BreakProhibited,
	{ClassNS, ClassNL}: BreakProhibited,
	{ClassNS, ClassNS}: BreakIndirect,
	{ClassNS, ClassNU}: BreakDirect,
	{ClassNS, ClassOP}: BreakDirect,
	{ClassNS, ClassPO}: BreakDirect,
	{ClassNS, ClassPR}: BreakDirect,
	{ClassNS, ClassQU}: BreakProhibited,
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
	{ClassNU, ClassAI}: BreakIndirect,
	{ClassNU, ClassAK}: BreakDirect,
	{ClassNU, ClassAL}: BreakIndirect,
	{ClassNU, ClassAP}: BreakDirect,
	{ClassNU, ClassAS}: BreakDirect,
	{ClassNU, ClassB2}: BreakDirect,
	{ClassNU, ClassBA}: BreakIndirect,
	{ClassNU, ClassBB}: BreakDirect,
	{ClassNU, ClassBK}: BreakProhibited,
	{ClassNU, ClassCB}: BreakDirect,
	{ClassNU, ClassCJ}: BreakIndirect,
	{ClassNU, ClassCL}: BreakProhibited,
	{ClassNU, ClassCM}: BreakIndirect,
	{ClassNU, ClassCP}: BreakProhibited,
	{ClassNU, ClassCR}: BreakProhibited,
	{ClassNU, ClassEB}: BreakDirect,
	{ClassNU, ClassEM}: BreakDirect,
	{ClassNU, ClassEX}: BreakProhibited,
	{ClassNU, ClassGL}: BreakIndirect,
	{ClassNU, ClassH2}: BreakDirect,
	{ClassNU, ClassH3}: BreakDirect,
	{ClassNU, ClassHH}: BreakIndirect,
	{ClassNU, ClassHL}: BreakIndirect,
	{ClassNU, ClassHY}: BreakIndirect,
	{ClassNU, ClassID}: BreakDirect,
	{ClassNU, ClassIN}: BreakIndirect,
	{ClassNU, ClassIS}: BreakProhibited,
	{ClassNU, ClassJL}: BreakDirect,
	{ClassNU, ClassJT}: BreakDirect,
	{ClassNU, ClassJV}: BreakDirect,
	{ClassNU, ClassLF}: BreakProhibited,
	{ClassNU, ClassNL}: BreakProhibited,
	{ClassNU, ClassNS}: BreakIndirect,
	{ClassNU, ClassNU}: BreakIndirect,
	{ClassNU, ClassOP}: BreakDirect,
	{ClassNU, ClassPO}: BreakIndirect,
	{ClassNU, ClassPR}: BreakIndirect,
	{ClassNU, ClassQU}: BreakProhibited,
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
	{ClassOP, ClassAK}: BreakProhibited,
	{ClassOP, ClassAL}: BreakProhibited,
	{ClassOP, ClassAP}: BreakProhibited,
	{ClassOP, ClassAS}: BreakProhibited,
	{ClassOP, ClassB2}: BreakProhibited,
	{ClassOP, ClassBA}: BreakProhibited,
	{ClassOP, ClassBB}: BreakProhibited,
	{ClassOP, ClassBK}: BreakProhibited,
	{ClassOP, ClassCB}: BreakProhibited,
	{ClassOP, ClassCJ}: BreakProhibited,
	{ClassOP, ClassCL}: BreakProhibited,
	{ClassOP, ClassCM}: BreakProhibited,
	{ClassOP, ClassCP}: BreakProhibited,
	{ClassOP, ClassCR}: BreakProhibited,
	{ClassOP, ClassEB}: BreakProhibited,
	{ClassOP, ClassEM}: BreakProhibited,
	{ClassOP, ClassEX}: BreakProhibited,
	{ClassOP, ClassGL}: BreakProhibited,
	{ClassOP, ClassH2}: BreakProhibited,
	{ClassOP, ClassH3}: BreakProhibited,
	{ClassOP, ClassHH}: BreakProhibited,
	{ClassOP, ClassHL}: BreakProhibited,
	{ClassOP, ClassHY}: BreakProhibited,
	{ClassOP, ClassID}: BreakProhibited,
	{ClassOP, ClassIN}: BreakProhibited,
	{ClassOP, ClassIS}: BreakProhibited,
	{ClassOP, ClassJL}: BreakProhibited,
	{ClassOP, ClassJT}: BreakProhibited,
	{ClassOP, ClassJV}: BreakProhibited,
	{ClassOP, ClassLF}: BreakProhibited,
	{ClassOP, ClassNL}: BreakProhibited,
	{ClassOP, ClassNS}: BreakProhibited,
	{ClassOP, ClassNU}: BreakProhibited,
	{ClassOP, ClassOP}: BreakProhibited,
	{ClassOP, ClassPO}: BreakProhibited,
	{ClassOP, ClassPR}: BreakProhibited,
	{ClassOP, ClassQU}: BreakProhibited,
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
	{ClassPO, ClassAI}: BreakIndirect,
	{ClassPO, ClassAK}: BreakDirect,
	{ClassPO, ClassAL}: BreakIndirect,
	{ClassPO, ClassAP}: BreakDirect,
	{ClassPO, ClassAS}: BreakDirect,
	{ClassPO, ClassB2}: BreakDirect,
	{ClassPO, ClassBA}: BreakIndirect,
	{ClassPO, ClassBB}: BreakDirect,
	{ClassPO, ClassBK}: BreakProhibited,
	{ClassPO, ClassCB}: BreakDirect,
	{ClassPO, ClassCJ}: BreakIndirect,
	{ClassPO, ClassCL}: BreakProhibited,
	{ClassPO, ClassCM}: BreakIndirect,
	{ClassPO, ClassCP}: BreakProhibited,
	{ClassPO, ClassCR}: BreakProhibited,
	{ClassPO, ClassEB}: BreakDirect,
	{ClassPO, ClassEM}: BreakDirect,
	{ClassPO, ClassEX}: BreakProhibited,
	{ClassPO, ClassGL}: BreakIndirect,
	{ClassPO, ClassH2}: BreakDirect,
	{ClassPO, ClassH3}: BreakDirect,
	{ClassPO, ClassHH}: BreakIndirect,
	{ClassPO, ClassHL}: BreakIndirect,
	{ClassPO, ClassHY}: BreakIndirect,
	{ClassPO, ClassID}: BreakDirect,
	{ClassPO, ClassIN}: BreakIndirect,
	{ClassPO, ClassIS}: BreakProhibited,
	{ClassPO, ClassJL}: BreakDirect,
	{ClassPO, ClassJT}: BreakDirect,
	{ClassPO, ClassJV}: BreakDirect,
	{ClassPO, ClassLF}: BreakProhibited,
	{ClassPO, ClassNL}: BreakProhibited,
	{ClassPO, ClassNS}: BreakIndirect,
	{ClassPO, ClassNU}: BreakIndirect,
	{ClassPO, ClassOP}: BreakDirect,
	{ClassPO, ClassPO}: BreakDirect,
	{ClassPO, ClassPR}: BreakDirect,
	{ClassPO, ClassQU}: BreakProhibited,
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
	{ClassPR, ClassAI}: BreakIndirect,
	{ClassPR, ClassAK}: BreakDirect,
	{ClassPR, ClassAL}: BreakIndirect,
	{ClassPR, ClassAP}: BreakDirect,
	{ClassPR, ClassAS}: BreakDirect,
	{ClassPR, ClassB2}: BreakDirect,
	{ClassPR, ClassBA}: BreakIndirect,
	{ClassPR, ClassBB}: BreakDirect,
	{ClassPR, ClassBK}: BreakProhibited,
	{ClassPR, ClassCB}: BreakDirect,
	{ClassPR, ClassCJ}: BreakIndirect,
	{ClassPR, ClassCL}: BreakProhibited,
	{ClassPR, ClassCM}: BreakIndirect,
	{ClassPR, ClassCP}: BreakProhibited,
	{ClassPR, ClassCR}: BreakProhibited,
	{ClassPR, ClassEB}: BreakIndirect,
	{ClassPR, ClassEM}: BreakIndirect,
	{ClassPR, ClassEX}: BreakProhibited,
	{ClassPR, ClassGL}: BreakIndirect,
	{ClassPR, ClassH2}: BreakIndirect,
	{ClassPR, ClassH3}: BreakIndirect,
	{ClassPR, ClassHH}: BreakIndirect,
	{ClassPR, ClassHL}: BreakIndirect,
	{ClassPR, ClassHY}: BreakIndirect,
	{ClassPR, ClassID}: BreakIndirect,
	{ClassPR, ClassIN}: BreakIndirect,
	{ClassPR, ClassIS}: BreakProhibited,
	{ClassPR, ClassJL}: BreakIndirect,
	{ClassPR, ClassJT}: BreakIndirect,
	{ClassPR, ClassJV}: BreakIndirect,
	{ClassPR, ClassLF}: BreakProhibited,
	{ClassPR, ClassNL}: BreakProhibited,
	{ClassPR, ClassNS}: BreakIndirect,
	{ClassPR, ClassNU}: BreakIndirect,
	{ClassPR, ClassOP}: BreakDirect,
	{ClassPR, ClassPO}: BreakDirect,
	{ClassPR, ClassPR}: BreakDirect,
	{ClassPR, ClassQU}: BreakProhibited,
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
	{ClassQU, ClassAI}: BreakProhibited,
	{ClassQU, ClassAK}: BreakProhibited,
	{ClassQU, ClassAL}: BreakProhibited,
	{ClassQU, ClassAP}: BreakProhibited,
	{ClassQU, ClassAS}: BreakProhibited,
	{ClassQU, ClassB2}: BreakProhibited,
	{ClassQU, ClassBA}: BreakProhibited,
	{ClassQU, ClassBB}: BreakProhibited,
	{ClassQU, ClassBK}: BreakProhibited,
	{ClassQU, ClassCB}: BreakProhibited,
	{ClassQU, ClassCJ}: BreakProhibited,
	{ClassQU, ClassCL}: BreakProhibited,
	{ClassQU, ClassCM}: BreakProhibited,
	{ClassQU, ClassCP}: BreakProhibited,
	{ClassQU, ClassCR}: BreakProhibited,
	{ClassQU, ClassEB}: BreakProhibited,
	{ClassQU, ClassEM}: BreakProhibited,
	{ClassQU, ClassEX}: BreakProhibited,
	{ClassQU, ClassGL}: BreakProhibited,
	{ClassQU, ClassH2}: BreakProhibited,
	{ClassQU, ClassH3}: BreakProhibited,
	{ClassQU, ClassHH}: BreakProhibited,
	{ClassQU, ClassHL}: BreakProhibited,
	{ClassQU, ClassHY}: BreakProhibited,
	{ClassQU, ClassID}: BreakProhibited,
	{ClassQU, ClassIN}: BreakProhibited,
	{ClassQU, ClassIS}: BreakProhibited,
	{ClassQU, ClassJL}: BreakProhibited,
	{ClassQU, ClassJT}: BreakProhibited,
	{ClassQU, ClassJV}: BreakProhibited,
	{ClassQU, ClassLF}: BreakProhibited,
	{ClassQU, ClassNL}: BreakProhibited,
	{ClassQU, ClassNS}: BreakProhibited,
	{ClassQU, ClassNU}: BreakProhibited,
	{ClassQU, ClassOP}: BreakProhibited,
	{ClassQU, ClassPO}: BreakProhibited,
	{ClassQU, ClassPR}: BreakProhibited,
	{ClassQU, ClassQU}: BreakProhibited,
	{ClassQU, ClassRI}: BreakProhibited,
	{ClassQU, ClassSA}: BreakProhibited,
	{ClassQU, ClassSP}: BreakProhibited,
	{ClassQU, ClassSY}: BreakProhibited,
	{ClassQU, ClassVF}: BreakProhibited,
	{ClassQU, ClassVI}: BreakProhibited,
	{ClassQU, ClassWJ}: BreakProhibited,
	{ClassQU, ClassXX}: BreakProhibited,
	{ClassQU, ClassZW}: BreakProhibited,
	{ClassQU, ClassZWJ}: BreakProhibited,
	{ClassRI, ClassAI}: BreakDirect,
	{ClassRI, ClassAK}: BreakDirect,
	{ClassRI, ClassAL}: BreakDirect,
	{ClassRI, ClassAP}: BreakDirect,
	{ClassRI, ClassAS}: BreakDirect,
	{ClassRI, ClassB2}: BreakDirect,
	{ClassRI, ClassBA}: BreakIndirect,
	{ClassRI, ClassBB}: BreakDirect,
	{ClassRI, ClassBK}: BreakProhibited,
	{ClassRI, ClassCB}: BreakDirect,
	{ClassRI, ClassCJ}: BreakIndirect,
	{ClassRI, ClassCL}: BreakProhibited,
	{ClassRI, ClassCM}: BreakIndirect,
	{ClassRI, ClassCP}: BreakProhibited,
	{ClassRI, ClassCR}: BreakProhibited,
	{ClassRI, ClassEB}: BreakDirect,
	{ClassRI, ClassEM}: BreakDirect,
	{ClassRI, ClassEX}: BreakProhibited,
	{ClassRI, ClassGL}: BreakIndirect,
	{ClassRI, ClassH2}: BreakDirect,
	{ClassRI, ClassH3}: BreakDirect,
	{ClassRI, ClassHH}: BreakIndirect,
	{ClassRI, ClassHL}: BreakDirect,
	{ClassRI, ClassHY}: BreakIndirect,
	{ClassRI, ClassID}: BreakDirect,
	{ClassRI, ClassIN}: BreakIndirect,
	{ClassRI, ClassIS}: BreakProhibited,
	{ClassRI, ClassJL}: BreakDirect,
	{ClassRI, ClassJT}: BreakDirect,
	{ClassRI, ClassJV}: BreakDirect,
	{ClassRI, ClassLF}: BreakProhibited,
	{ClassRI, ClassNL}: BreakProhibited,
	{ClassRI, ClassNS}: BreakIndirect,
	{ClassRI, ClassNU}: BreakDirect,
	{ClassRI, ClassOP}: BreakDirect,
	{ClassRI, ClassPO}: BreakDirect,
	{ClassRI, ClassPR}: BreakDirect,
	{ClassRI, ClassQU}: BreakProhibited,
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
	{ClassSA, ClassAK}: BreakDirect,
	{ClassSA, ClassAL}: BreakIndirect,
	{ClassSA, ClassAP}: BreakDirect,
	{ClassSA, ClassAS}: BreakDirect,
	{ClassSA, ClassB2}: BreakDirect,
	{ClassSA, ClassBA}: BreakIndirect,
	{ClassSA, ClassBB}: BreakDirect,
	{ClassSA, ClassBK}: BreakProhibited,
	{ClassSA, ClassCB}: BreakDirect,
	{ClassSA, ClassCJ}: BreakIndirect,
	{ClassSA, ClassCL}: BreakProhibited,
	{ClassSA, ClassCM}: BreakIndirect,
	{ClassSA, ClassCP}: BreakProhibited,
	{ClassSA, ClassCR}: BreakProhibited,
	{ClassSA, ClassEB}: BreakDirect,
	{ClassSA, ClassEM}: BreakDirect,
	{ClassSA, ClassEX}: BreakProhibited,
	{ClassSA, ClassGL}: BreakIndirect,
	{ClassSA, ClassH2}: BreakDirect,
	{ClassSA, ClassH3}: BreakDirect,
	{ClassSA, ClassHH}: BreakIndirect,
	{ClassSA, ClassHL}: BreakIndirect,
	{ClassSA, ClassHY}: BreakIndirect,
	{ClassSA, ClassID}: BreakDirect,
	{ClassSA, ClassIN}: BreakIndirect,
	{ClassSA, ClassIS}: BreakProhibited,
	{ClassSA, ClassJL}: BreakDirect,
	{ClassSA, ClassJT}: BreakDirect,
	{ClassSA, ClassJV}: BreakDirect,
	{ClassSA, ClassLF}: BreakProhibited,
	{ClassSA, ClassNL}: BreakProhibited,
	{ClassSA, ClassNS}: BreakIndirect,
	{ClassSA, ClassNU}: BreakIndirect,
	{ClassSA, ClassOP}: BreakDirect,
	{ClassSA, ClassPO}: BreakIndirect,
	{ClassSA, ClassPR}: BreakIndirect,
	{ClassSA, ClassQU}: BreakProhibited,
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
	{ClassSP, ClassAK}: BreakDirect,
	{ClassSP, ClassAL}: BreakDirect,
	{ClassSP, ClassAP}: BreakDirect,
	{ClassSP, ClassAS}: BreakDirect,
	{ClassSP, ClassB2}: BreakDirect,
	{ClassSP, ClassBA}: BreakDirect,
	{ClassSP, ClassBB}: BreakDirect,
	{ClassSP, ClassBK}: BreakProhibited,
	{ClassSP, ClassCB}: BreakDirect,
	{ClassSP, ClassCJ}: BreakDirect,
	{ClassSP, ClassCL}: BreakProhibited,
	{ClassSP, ClassCM}: BreakDirect,
	{ClassSP, ClassCP}: BreakProhibited,
	{ClassSP, ClassCR}: BreakProhibited,
	{ClassSP, ClassEB}: BreakDirect,
	{ClassSP, ClassEM}: BreakDirect,
	{ClassSP, ClassEX}: BreakProhibited,
	{ClassSP, ClassGL}: BreakDirect,
	{ClassSP, ClassH2}: BreakDirect,
	{ClassSP, ClassH3}: BreakDirect,
	{ClassSP, ClassHH}: BreakDirect,
	{ClassSP, ClassHL}: BreakDirect,
	{ClassSP, ClassHY}: BreakDirect,
	{ClassSP, ClassID}: BreakDirect,
	{ClassSP, ClassIN}: BreakDirect,
	{ClassSP, ClassIS}: BreakProhibited,
	{ClassSP, ClassJL}: BreakDirect,
	{ClassSP, ClassJT}: BreakDirect,
	{ClassSP, ClassJV}: BreakDirect,
	{ClassSP, ClassLF}: BreakProhibited,
	{ClassSP, ClassNL}: BreakProhibited,
	{ClassSP, ClassNS}: BreakDirect,
	{ClassSP, ClassNU}: BreakDirect,
	{ClassSP, ClassOP}: BreakDirect,
	{ClassSP, ClassPO}: BreakDirect,
	{ClassSP, ClassPR}: BreakDirect,
	{ClassSP, ClassQU}: BreakProhibited,
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
	{ClassSY, ClassAK}: BreakDirect,
	{ClassSY, ClassAL}: BreakDirect,
	{ClassSY, ClassAP}: BreakDirect,
	{ClassSY, ClassAS}: BreakDirect,
	{ClassSY, ClassB2}: BreakDirect,
	{ClassSY, ClassBA}: BreakIndirect,
	{ClassSY, ClassBB}: BreakDirect,
	{ClassSY, ClassBK}: BreakProhibited,
	{ClassSY, ClassCB}: BreakDirect,
	{ClassSY, ClassCJ}: BreakIndirect,
	{ClassSY, ClassCL}: BreakProhibited,
	{ClassSY, ClassCM}: BreakIndirect,
	{ClassSY, ClassCP}: BreakProhibited,
	{ClassSY, ClassCR}: BreakProhibited,
	{ClassSY, ClassEB}: BreakDirect,
	{ClassSY, ClassEM}: BreakDirect,
	{ClassSY, ClassEX}: BreakProhibited,
	{ClassSY, ClassGL}: BreakIndirect,
	{ClassSY, ClassH2}: BreakDirect,
	{ClassSY, ClassH3}: BreakDirect,
	{ClassSY, ClassHH}: BreakIndirect,
	{ClassSY, ClassHL}: BreakIndirect,
	{ClassSY, ClassHY}: BreakIndirect,
	{ClassSY, ClassID}: BreakDirect,
	{ClassSY, ClassIN}: BreakIndirect,
	{ClassSY, ClassIS}: BreakProhibited,
	{ClassSY, ClassJL}: BreakDirect,
	{ClassSY, ClassJT}: BreakDirect,
	{ClassSY, ClassJV}: BreakDirect,
	{ClassSY, ClassLF}: BreakProhibited,
	{ClassSY, ClassNL}: BreakProhibited,
	{ClassSY, ClassNS}: BreakIndirect,
	{ClassSY, ClassNU}: BreakDirect,
	{ClassSY, ClassOP}: BreakDirect,
	{ClassSY, ClassPO}: BreakDirect,
	{ClassSY, ClassPR}: BreakDirect,
	{ClassSY, ClassQU}: BreakProhibited,
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
	{ClassVF, ClassAK}: BreakDirect,
	{ClassVF, ClassAL}: BreakDirect,
	{ClassVF, ClassAP}: BreakDirect,
	{ClassVF, ClassAS}: BreakDirect,
	{ClassVF, ClassB2}: BreakDirect,
	{ClassVF, ClassBA}: BreakIndirect,
	{ClassVF, ClassBB}: BreakDirect,
	{ClassVF, ClassBK}: BreakProhibited,
	{ClassVF, ClassCB}: BreakDirect,
	{ClassVF, ClassCJ}: BreakIndirect,
	{ClassVF, ClassCL}: BreakProhibited,
	{ClassVF, ClassCM}: BreakIndirect,
	{ClassVF, ClassCP}: BreakProhibited,
	{ClassVF, ClassCR}: BreakProhibited,
	{ClassVF, ClassEB}: BreakDirect,
	{ClassVF, ClassEM}: BreakDirect,
	{ClassVF, ClassEX}: BreakProhibited,
	{ClassVF, ClassGL}: BreakIndirect,
	{ClassVF, ClassH2}: BreakDirect,
	{ClassVF, ClassH3}: BreakDirect,
	{ClassVF, ClassHH}: BreakIndirect,
	{ClassVF, ClassHL}: BreakDirect,
	{ClassVF, ClassHY}: BreakIndirect,
	{ClassVF, ClassID}: BreakDirect,
	{ClassVF, ClassIN}: BreakIndirect,
	{ClassVF, ClassIS}: BreakProhibited,
	{ClassVF, ClassJL}: BreakDirect,
	{ClassVF, ClassJT}: BreakDirect,
	{ClassVF, ClassJV}: BreakDirect,
	{ClassVF, ClassLF}: BreakProhibited,
	{ClassVF, ClassNL}: BreakProhibited,
	{ClassVF, ClassNS}: BreakIndirect,
	{ClassVF, ClassNU}: BreakDirect,
	{ClassVF, ClassOP}: BreakDirect,
	{ClassVF, ClassPO}: BreakDirect,
	{ClassVF, ClassPR}: BreakDirect,
	{ClassVF, ClassQU}: BreakProhibited,
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
	{ClassVI, ClassAK}: BreakDirect,
	{ClassVI, ClassAL}: BreakDirect,
	{ClassVI, ClassAP}: BreakDirect,
	{ClassVI, ClassAS}: BreakDirect,
	{ClassVI, ClassB2}: BreakDirect,
	{ClassVI, ClassBA}: BreakIndirect,
	{ClassVI, ClassBB}: BreakDirect,
	{ClassVI, ClassBK}: BreakProhibited,
	{ClassVI, ClassCB}: BreakDirect,
	{ClassVI, ClassCJ}: BreakIndirect,
	{ClassVI, ClassCL}: BreakProhibited,
	{ClassVI, ClassCM}: BreakIndirect,
	{ClassVI, ClassCP}: BreakProhibited,
	{ClassVI, ClassCR}: BreakProhibited,
	{ClassVI, ClassEB}: BreakDirect,
	{ClassVI, ClassEM}: BreakDirect,
	{ClassVI, ClassEX}: BreakProhibited,
	{ClassVI, ClassGL}: BreakIndirect,
	{ClassVI, ClassH2}: BreakDirect,
	{ClassVI, ClassH3}: BreakDirect,
	{ClassVI, ClassHH}: BreakIndirect,
	{ClassVI, ClassHL}: BreakDirect,
	{ClassVI, ClassHY}: BreakIndirect,
	{ClassVI, ClassID}: BreakDirect,
	{ClassVI, ClassIN}: BreakIndirect,
	{ClassVI, ClassIS}: BreakProhibited,
	{ClassVI, ClassJL}: BreakDirect,
	{ClassVI, ClassJT}: BreakDirect,
	{ClassVI, ClassJV}: BreakDirect,
	{ClassVI, ClassLF}: BreakProhibited,
	{ClassVI, ClassNL}: BreakProhibited,
	{ClassVI, ClassNS}: BreakIndirect,
	{ClassVI, ClassNU}: BreakDirect,
	{ClassVI, ClassOP}: BreakDirect,
	{ClassVI, ClassPO}: BreakDirect,
	{ClassVI, ClassPR}: BreakDirect,
	{ClassVI, ClassQU}: BreakProhibited,
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
	{ClassWJ, ClassAK}: BreakIndirect,
	{ClassWJ, ClassAL}: BreakIndirect,
	{ClassWJ, ClassAP}: BreakIndirect,
	{ClassWJ, ClassAS}: BreakIndirect,
	{ClassWJ, ClassB2}: BreakIndirect,
	{ClassWJ, ClassBA}: BreakIndirect,
	{ClassWJ, ClassBB}: BreakIndirect,
	{ClassWJ, ClassBK}: BreakProhibited,
	{ClassWJ, ClassCB}: BreakIndirect,
	{ClassWJ, ClassCJ}: BreakIndirect,
	{ClassWJ, ClassCL}: BreakProhibited,
	{ClassWJ, ClassCM}: BreakIndirect,
	{ClassWJ, ClassCP}: BreakProhibited,
	{ClassWJ, ClassCR}: BreakProhibited,
	{ClassWJ, ClassEB}: BreakIndirect,
	{ClassWJ, ClassEM}: BreakIndirect,
	{ClassWJ, ClassEX}: BreakProhibited,
	{ClassWJ, ClassGL}: BreakIndirect,
	{ClassWJ, ClassH2}: BreakIndirect,
	{ClassWJ, ClassH3}: BreakIndirect,
	{ClassWJ, ClassHH}: BreakIndirect,
	{ClassWJ, ClassHL}: BreakIndirect,
	{ClassWJ, ClassHY}: BreakIndirect,
	{ClassWJ, ClassID}: BreakIndirect,
	{ClassWJ, ClassIN}: BreakIndirect,
	{ClassWJ, ClassIS}: BreakProhibited,
	{ClassWJ, ClassJL}: BreakIndirect,
	{ClassWJ, ClassJT}: BreakIndirect,
	{ClassWJ, ClassJV}: BreakIndirect,
	{ClassWJ, ClassLF}: BreakProhibited,
	{ClassWJ, ClassNL}: BreakProhibited,
	{ClassWJ, ClassNS}: BreakIndirect,
	{ClassWJ, ClassNU}: BreakIndirect,
	{ClassWJ, ClassOP}: BreakIndirect,
	{ClassWJ, ClassPO}: BreakIndirect,
	{ClassWJ, ClassPR}: BreakIndirect,
	{ClassWJ, ClassQU}: BreakProhibited,
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
	{ClassZW, ClassAK}: BreakDirect,
	{ClassZW, ClassAL}: BreakDirect,
	{ClassZW, ClassAP}: BreakDirect,
	{ClassZW, ClassAS}: BreakDirect,
	{ClassZW, ClassB2}: BreakDirect,
	{ClassZW, ClassBA}: BreakDirect,
	{ClassZW, ClassBB}: BreakDirect,
	{ClassZW, ClassBK}: BreakProhibited,
	{ClassZW, ClassCB}: BreakDirect,
	{ClassZW, ClassCJ}: BreakDirect,
	{ClassZW, ClassCL}: BreakDirect,
	{ClassZW, ClassCM}: BreakDirect,
	{ClassZW, ClassCP}: BreakDirect,
	{ClassZW, ClassCR}: BreakProhibited,
	{ClassZW, ClassEB}: BreakDirect,
	{ClassZW, ClassEM}: BreakDirect,
	{ClassZW, ClassEX}: BreakDirect,
	{ClassZW, ClassGL}: BreakDirect,
	{ClassZW, ClassH2}: BreakDirect,
	{ClassZW, ClassH3}: BreakDirect,
	{ClassZW, ClassHH}: BreakDirect,
	{ClassZW, ClassHL}: BreakDirect,
	{ClassZW, ClassHY}: BreakDirect,
	{ClassZW, ClassID}: BreakDirect,
	{ClassZW, ClassIN}: BreakDirect,
	{ClassZW, ClassIS}: BreakDirect,
	{ClassZW, ClassJL}: BreakDirect,
	{ClassZW, ClassJT}: BreakDirect,
	{ClassZW, ClassJV}: BreakDirect,
	{ClassZW, ClassLF}: BreakProhibited,
	{ClassZW, ClassNL}: BreakProhibited,
	{ClassZW, ClassNS}: BreakDirect,
	{ClassZW, ClassNU}: BreakDirect,
	{ClassZW, ClassOP}: BreakDirect,
	{ClassZW, ClassPO}: BreakDirect,
	{ClassZW, ClassPR}: BreakDirect,
	{ClassZW, ClassQU}: BreakDirect,
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
	{ClassZWJ, ClassAK}: BreakIndirect,
	{ClassZWJ, ClassAL}: BreakIndirect,
	{ClassZWJ, ClassAP}: BreakIndirect,
	{ClassZWJ, ClassAS}: BreakIndirect,
	{ClassZWJ, ClassB2}: BreakIndirect,
	{ClassZWJ, ClassBA}: BreakIndirect,
	{ClassZWJ, ClassBB}: BreakIndirect,
	{ClassZWJ, ClassBK}: BreakProhibited,
	{ClassZWJ, ClassCB}: BreakIndirect,
	{ClassZWJ, ClassCJ}: BreakIndirect,
	{ClassZWJ, ClassCL}: BreakProhibited,
	{ClassZWJ, ClassCM}: BreakIndirect,
	{ClassZWJ, ClassCP}: BreakProhibited,
	{ClassZWJ, ClassCR}: BreakProhibited,
	{ClassZWJ, ClassEB}: BreakIndirect,
	{ClassZWJ, ClassEM}: BreakIndirect,
	{ClassZWJ, ClassEX}: BreakProhibited,
	{ClassZWJ, ClassGL}: BreakIndirect,
	{ClassZWJ, ClassH2}: BreakIndirect,
	{ClassZWJ, ClassH3}: BreakIndirect,
	{ClassZWJ, ClassHH}: BreakIndirect,
	{ClassZWJ, ClassHL}: BreakIndirect,
	{ClassZWJ, ClassHY}: BreakIndirect,
	{ClassZWJ, ClassID}: BreakIndirect,
	{ClassZWJ, ClassIN}: BreakIndirect,
	{ClassZWJ, ClassIS}: BreakProhibited,
	{ClassZWJ, ClassJL}: BreakIndirect,
	{ClassZWJ, ClassJT}: BreakIndirect,
	{ClassZWJ, ClassJV}: BreakIndirect,
	{ClassZWJ, ClassLF}: BreakProhibited,
	{ClassZWJ, ClassNL}: BreakProhibited,
	{ClassZWJ, ClassNS}: BreakIndirect,
	{ClassZWJ, ClassNU}: BreakIndirect,
	{ClassZWJ, ClassOP}: BreakIndirect,
	{ClassZWJ, ClassPO}: BreakIndirect,
	{ClassZWJ, ClassPR}: BreakIndirect,
	{ClassZWJ, ClassQU}: BreakProhibited,
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
	lastNonSpaceClass := prevClass // Track last non-SP class for LB14

	for i := 1; i < len(runes); i++ {
		currClass := getBreakClass(runes[i])

		// LB4, LB5: Mandatory breaks - handle BEFORE consulting pair table
		// Always break after BK, CR (except before LF), LF, NL
		if prevClass == ClassBK || prevClass == ClassLF || prevClass == ClassNL {
			bytePos := len(string(runes[:i]))
			breakPoints = append(breakPoints, bytePos)
			prevClass = currClass
			if currClass != ClassSP {
				lastNonSpaceClass = currClass
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
				prevClass = currClass
				if currClass != ClassSP {
					lastNonSpaceClass = currClass
				}
				continue
			}
		}

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
			// But respect LB6, LB13, LB14
			if prevClass == ClassSP {
				// LB14: Do not break after OP, even if spaces intervene (OP SP* ×)
				// LB19: Do not break before or after QU (× SP* QU and QU SP* ×)
				if lastNonSpaceClass == ClassOP || lastNonSpaceClass == ClassQU {
					// Don't break - we're in "OP SP*" or "QU SP*" sequence
				} else if currClass == ClassQU && lastNonSpaceClass != ClassAI {
					// LB19: Do not break before QU (× SP* QU), except after AI
				} else if currClass == ClassGL && lastNonSpaceClass != ClassAI {
					// Do not break before GL (× SP* GL), except after AI
				} else if currClass == ClassBK || currClass == ClassCR || currClass == ClassLF ||
					currClass == ClassNL || currClass == ClassCL || currClass == ClassCP ||
					currClass == ClassEX || currClass == ClassIS || currClass == ClassSY ||
					currClass == ClassWJ || currClass == ClassZW {
					// LB6: Do not break before hard line breaks (BK, CR, LF, NL)
					// LB7: Do not break before ZW (× ZW)
					// LB11: Do not break before WJ (× WJ)
					// LB13: Do not break before CL, CP, EX, IS, SY (closing punct)
					// Note: NS removed - LB18 (break after space) overrides LB16 (× NS)
					// Note: GL removed - handled above with AI exception
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
				// Explicit break opportunities (hyphens, soft hyphens, BA, B2)
				// For BA/B2: Break after, but not immediately before SP, CM, or other special chars
				// For HY/CB: Respect the hyphens property
				isSoftHyphen := i > 0 && runes[i-1] == '\u00AD'

				if prevClass == ClassBA || prevClass == ClassB2 {
					// BA/B2 characters: allow break, but not immediately before SP, CM, GL, ZW
					// The break should happen after the following space/character
					if currClass != ClassSP && currClass != ClassCM && currClass != ClassGL && currClass != ClassZW {
						// BA/B2 × ! (SP|CM|GL|ZW) - break after BA/B2
						if !(isSoftHyphen && hyphens == HyphensNone) {
							bytePos := len(string(runes[:i]))
							breakPoints = append(breakPoints, bytePos)
						}
					}
					// BA/B2 × SP - don't break here; let space create the break
					// BA/B2 × CM - don't break before combining mark
					// BA/B2 × GL - don't break before glue
					// BA/B2 × ZW - don't break before zero-width space (LB8)
				} else {
					// HY/CB: Respect hyphens setting
					if hyphens == HyphensNone {
						// Don't break at any hyphens
					} else if hyphens == HyphensManual && isSoftHyphen {
						// Only break at soft hyphens in manual mode
						bytePos := len(string(runes[:i]))
						breakPoints = append(breakPoints, bytePos)
					} else if hyphens == HyphensManual && !isSoftHyphen {
						// Don't break at hard hyphens in manual mode
					} else if hyphens == HyphensAuto {
						// Break at all hyphens in auto mode
						bytePos := len(string(runes[:i]))
						breakPoints = append(breakPoints, bytePos)
					}
				}
			} else if prevClass == ClassSP {
				// LB18: Break after spaces (word boundaries)
				// But respect LB6, LB7, LB11, LB12, LB13, LB14, LB19
				// LB14: Do not break after OP, even if spaces intervene (OP SP* ×)
				// LB19: Do not break before or after QU (× SP* QU and QU SP* ×)
				if lastNonSpaceClass == ClassOP || lastNonSpaceClass == ClassQU {
					// Don't break - we're in "OP SP*" or "QU SP*" sequence
				} else if currClass == ClassQU && lastNonSpaceClass != ClassAI {
					// LB19: Do not break before QU (× SP* QU), except after AI
				} else if currClass == ClassGL && lastNonSpaceClass != ClassAI {
					// Do not break before GL (× SP* GL), except after AI
				} else if currClass == ClassBK || currClass == ClassCR || currClass == ClassLF ||
					currClass == ClassNL || currClass == ClassCL || currClass == ClassCP ||
					currClass == ClassEX || currClass == ClassIS || currClass == ClassSY ||
					currClass == ClassWJ || currClass == ClassZW {
					// LB6: Do not break before hard line breaks
					// LB7: Do not break before ZW (× ZW)
					// LB11: Do not break before WJ (× WJ)
					// LB13: Do not break before CL, CP, EX, IS, SY
					// Note: NS removed - LB18 (break after space) overrides LB16 (× NS)
					// Note: GL removed - handled above with AI exception
				} else {
					bytePos := len(string(runes[:i]))
					breakPoints = append(breakPoints, bytePos)
				}
			} else if prevClass == ClassID || currClass == ClassID ||
				prevClass == ClassAI || currClass == ClassAI {
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
				// Respect special rules: don't break before SP, ZW, CM, GL, WJ
				if currClass != ClassSP && currClass != ClassZW && currClass != ClassCM &&
					currClass != ClassGL && currClass != ClassWJ {
					bytePos := len(string(runes[:i]))
					breakPoints = append(breakPoints, bytePos)
				}
			}
		}

		// Update previous class (combining marks don't change it)
		if currClass != ClassCM {
			prevClass = currClass
			// Track last non-space class for LB14 (OP SP* ×)
			if currClass != ClassSP {
				lastNonSpaceClass = currClass
			}
		}
	}

	// End of text is always a break point
	breakPoints = append(breakPoints, len(text))

	return breakPoints
}
