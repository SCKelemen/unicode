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
var pairTable = map[[2]BreakClass]BreakAction{
	// Mandatory breaks
	{ClassLF, ClassXX}: BreakMandatory,
	{ClassCR, ClassXX}: BreakMandatory,
	{ClassCR, ClassLF}: BreakProhibited, // CR+LF is one break
	{ClassBK, ClassXX}: BreakMandatory,
	{ClassNL, ClassXX}: BreakMandatory,

	// LB6: Do not break before hard line breaks
	{ClassXX, ClassBK}: BreakProhibited,
	{ClassXX, ClassCR}: BreakProhibited,
	{ClassXX, ClassLF}: BreakProhibited,
	{ClassXX, ClassNL}: BreakProhibited,

	// Space
	{ClassSP, ClassSP}: BreakProhibited, // Do not break between spaces (LB18 special case)
	{ClassSP, ClassXX}: BreakIndirect,
	{ClassXX, ClassSP}: BreakProhibited,

	// Prohibited breaks
	{ClassWJ, ClassXX}: BreakProhibited,
	{ClassXX, ClassWJ}: BreakProhibited,
	{ClassZW, ClassXX}: BreakDirect, // Zero Width Space allows break

	// Break after (LB21a)
	{ClassBA, ClassXX}: BreakDirect,
	{ClassHY, ClassXX}: BreakDirect,
	{ClassSY, ClassXX}: BreakDirect,

	// LB21: Do not break before hyphen-minus, other hyphens, or BA
	{ClassXX, ClassBA}: BreakProhibited,
	{ClassXX, ClassHY}: BreakProhibited,

	// Special cases: don't break between BA and certain classes
	{ClassBA, ClassSP}: BreakProhibited,  // Let space handle the break
	{ClassBA, ClassCM}: BreakProhibited,  // Don't break before combining marks
	{ClassBA, ClassGL}: BreakProhibited,  // Don't break before glue
	{ClassB2, ClassBA}: BreakProhibited,  // Don't break between B2 and BA
	{ClassB2, ClassSP}: BreakProhibited,  // Let space handle the break
	{ClassB2, ClassCM}: BreakProhibited,  // Don't break before combining marks
	{ClassB2, ClassGL}: BreakProhibited,  // Don't break before glue

	// Break before
	{ClassXX, ClassBB}: BreakDirect,

	// Break before and after
	{ClassB2, ClassXX}: BreakDirect,
	{ClassXX, ClassB2}: BreakDirect,

	// Contingent break
	{ClassCB, ClassXX}: BreakDirect,

	// Punctuation
	// LB14: Do not break after OP (implemented in code with SP handling)
	{ClassOP, ClassXX}: BreakProhibited,
	{ClassQU, ClassXX}: BreakProhibited,
	{ClassGL, ClassXX}: BreakProhibited,
	// Note: CAN break before OP (removed incorrect {ClassXX, ClassOP} rule)
	{ClassXX, ClassQU}: BreakProhibited,
	{ClassXX, ClassGL}: BreakProhibited,

	// Close punctuation (LB13)
	{ClassCL, ClassXX}: BreakProhibited,
	{ClassCP, ClassXX}: BreakProhibited,
	{ClassXX, ClassCL}: BreakProhibited,
	{ClassXX, ClassCP}: BreakProhibited,

	// LB13: Do not break before EX, IS, SY
	{ClassXX, ClassEX}: BreakProhibited,
	{ClassXX, ClassIS}: BreakProhibited,
	{ClassXX, ClassSY}: BreakProhibited,

	// LB16: Do not break before NS (nonstarters)
	{ClassXX, ClassNS}: BreakProhibited,

	// LB11: Do not break before or after Word Joiner and related characters
	// (Already have WJ rules above)

	// LB12: Do not break after GL (non-breaking "glue")
	{ClassGL, ClassXX}: BreakProhibited,

	// IN (Inseparable) - don't break before or after
	{ClassXX, ClassIN}: BreakProhibited,
	{ClassIN, ClassXX}: BreakProhibited,

	// Numeric (LB23, LB24, LB25)
	{ClassNU, ClassNU}: BreakProhibited,
	{ClassNU, ClassAL}: BreakProhibited, // LB23
	{ClassAL, ClassNU}: BreakProhibited, // LB23
	{ClassNU, ClassHL}: BreakProhibited, // LB23
	{ClassHL, ClassNU}: BreakProhibited, // LB23
	{ClassIS, ClassNU}: BreakProhibited,
	{ClassNU, ClassIS}: BreakProhibited,
	{ClassNU, ClassSY}: BreakProhibited,
	{ClassSY, ClassNU}: BreakProhibited,

	// LB24: Do not break between numeric prefix and letters/ideographs
	{ClassPR, ClassAL}: BreakProhibited,
	{ClassPR, ClassHL}: BreakProhibited,
	{ClassPR, ClassID}: BreakProhibited,

	// LB25: Do not break between numeric prefix/postfix and numbers
	{ClassPR, ClassNU}: BreakProhibited,
	{ClassPR, ClassOP}: BreakProhibited, // PR × (OP | HY)? NU
	{ClassPR, ClassHY}: BreakProhibited,
	{ClassPO, ClassNU}: BreakProhibited,
	{ClassNU, ClassPO}: BreakProhibited,
	{ClassOP, ClassNU}: BreakProhibited,
	{ClassHY, ClassNU}: BreakProhibited,

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
	{ClassAI, ClassHH}: BreakProhibited, // Hebrew letters
	{ClassHH, ClassAI}: BreakProhibited,
	// AI doesn't break from most punctuation/symbols (treats like AL)
	{ClassAI, ClassEX}: BreakProhibited,
	{ClassEX, ClassAI}: BreakProhibited,
	{ClassAI, ClassIS}: BreakProhibited,
	{ClassIS, ClassAI}: BreakProhibited,
	{ClassAI, ClassNU}: BreakProhibited,
	{ClassNU, ClassAI}: BreakProhibited,
	{ClassAI, ClassHY}: BreakProhibited,
	{ClassHY, ClassAI}: BreakProhibited,
	{ClassAI, ClassPO}: BreakProhibited, // Postfix
	{ClassPO, ClassAI}: BreakProhibited,
	{ClassAI, ClassPR}: BreakProhibited, // Prefix
	{ClassPR, ClassAI}: BreakProhibited,
	// Note: AI CAN break before OP (unlike AL) - removed {ClassAI, ClassOP}: BreakProhibited
	{ClassAI, ClassCL}: BreakProhibited, // Closing
	{ClassAI, ClassCP}: BreakProhibited, // Close paren
	// Note: AI CAN break before QU in certain contexts - removed {ClassAI, ClassQU}: BreakProhibited
	{ClassAI, ClassNS}: BreakProhibited, // Nonstarter (LB16)
	// Note: AI can break with BA, BB, B2 (they override the prohibition)
	{ClassAI, ClassBB}: BreakDirect, // Break before
	{ClassBB, ClassAI}: BreakDirect,
	{ClassAI, ClassB2}: BreakDirect, // Break before/after
	{ClassB2, ClassAI}: BreakDirect,
	{ClassAI, ClassSY}: BreakProhibited, // Symbols
	{ClassSY, ClassAI}: BreakProhibited,
	// AI allows breaks with ideographic (like AL × ID)
	{ClassAI, ClassID}: BreakDirect,
	{ClassID, ClassAI}: BreakDirect,
	// AI allows breaks with special Indic classes (AK, AP, AS, VF, VI)
	{ClassAI, ClassAK}: BreakDirect,
	{ClassAK, ClassAI}: BreakDirect,
	{ClassAI, ClassAP}: BreakDirect,
	{ClassAP, ClassAI}: BreakDirect,
	{ClassAI, ClassAS}: BreakDirect,
	{ClassAS, ClassAI}: BreakDirect,
	// AI allows breaks with Hangul classes (H2, H3, JL, JV, JT)
	{ClassAI, ClassH2}: BreakDirect,
	{ClassH2, ClassAI}: BreakDirect,
	{ClassAI, ClassH3}: BreakDirect,
	{ClassH3, ClassAI}: BreakDirect,
	{ClassAI, ClassJL}: BreakDirect,
	{ClassJL, ClassAI}: BreakDirect,
	{ClassAI, ClassJV}: BreakDirect,
	{ClassJV, ClassAI}: BreakDirect,
	{ClassAI, ClassJT}: BreakDirect,
	{ClassJT, ClassAI}: BreakDirect,
	// AI does not break with complex context classes (SA, CJ, ZWJ)
	{ClassAI, ClassSA}: BreakProhibited, // South East Asian
	{ClassSA, ClassAI}: BreakProhibited,
	{ClassAI, ClassCJ}: BreakProhibited, // Conditional Japanese Starter
	{ClassCJ, ClassAI}: BreakProhibited,
	{ClassAI, ClassZWJ}: BreakProhibited, // Zero Width Joiner
	{ClassZWJ, ClassAI}: BreakProhibited,

	// Combining marks (LB9: prohibited break before)
	{ClassXX, ClassCM}: BreakProhibited,
	{ClassCM, ClassCM}: BreakProhibited,

	// LB8: Break before any character following a zero-width space
	// (ZW already has {ClassZW, ClassXX}: BreakDirect above)

	// LB8a: Do not break after a zero width joiner
	{ClassZWJ, ClassXX}: BreakProhibited,

	// LB20: Break before and after CB (contingent break)
	// (Already have {ClassCB, ClassXX}: BreakDirect)
	{ClassXX, ClassCB}: BreakDirect,

	// LB21a: Don't break after Hebrew + Hyphen
	{ClassHL, ClassHY}: BreakProhibited,
	{ClassHL, ClassBA}: BreakProhibited,

	// LB21b: Don't break between Solidus and Hebrew
	{ClassSY, ClassHL}: BreakProhibited,

	// LB22: Do not break before ellipsis
	// (IN rules already cover this)

	// LB23a: Do not break between numeric prefixes and ideographs
	{ClassPR, ClassID}: BreakProhibited, // Already have this
	{ClassID, ClassPO}: BreakProhibited,
	{ClassPO, ClassAL}: BreakProhibited,
	{ClassPO, ClassHL}: BreakProhibited,

	// LB26: Do not break Korean syllable
	{ClassJL, ClassJL}: BreakProhibited,
	{ClassJL, ClassJV}: BreakProhibited,
	{ClassJL, ClassH2}: BreakProhibited,
	{ClassJL, ClassH3}: BreakProhibited,
	{ClassJV, ClassJV}: BreakProhibited,
	{ClassJV, ClassJT}: BreakProhibited,
	{ClassH2, ClassJV}: BreakProhibited,
	{ClassH2, ClassJT}: BreakProhibited,
	{ClassJT, ClassJT}: BreakProhibited,
	{ClassH3, ClassJT}: BreakProhibited,

	// LB27: Treat Korean Syllable Block as ID
	{ClassJL, ClassIN}: BreakProhibited,
	{ClassJL, ClassPO}: BreakProhibited,
	{ClassJV, ClassIN}: BreakProhibited,
	{ClassJV, ClassPO}: BreakProhibited,
	{ClassJT, ClassIN}: BreakProhibited,
	{ClassJT, ClassPO}: BreakProhibited,
	{ClassH2, ClassIN}: BreakProhibited,
	{ClassH2, ClassPO}: BreakProhibited,
	{ClassH3, ClassIN}: BreakProhibited,
	{ClassH3, ClassPO}: BreakProhibited,
	{ClassJL, ClassPR}: BreakProhibited,
	{ClassJV, ClassPR}: BreakProhibited,
	{ClassJT, ClassPR}: BreakProhibited,
	{ClassH2, ClassPR}: BreakProhibited,
	{ClassH3, ClassPR}: BreakProhibited,

	// LB28: Do not break between alphabetics
	// (Already have AL × AL rules through default BreakProhibited)
	{ClassAL, ClassAL}: BreakProhibited,
	{ClassHL, ClassHL}: BreakProhibited,
	{ClassAL, ClassHL}: BreakProhibited,
	{ClassHL, ClassAL}: BreakProhibited,

	// LB29: Do not break between numeric punctuation and alphabetics
	{ClassIS, ClassAL}: BreakProhibited,
	{ClassIS, ClassHL}: BreakProhibited,

	// LB30: Do not break between letters, numbers, or ordinary symbols and OP/CP
	{ClassAL, ClassOP}: BreakProhibited,
	{ClassHL, ClassOP}: BreakProhibited,
	{ClassNU, ClassOP}: BreakProhibited,
	{ClassCP, ClassAL}: BreakProhibited,
	{ClassCP, ClassHL}: BreakProhibited,
	{ClassCP, ClassNU}: BreakProhibited,

	// LB30a: Break between two regional indicator symbols if and only if
	// there are an even number of RI preceding the position (complex - needs state)
	// Simplified: don't break RI × RI pairs
	{ClassRI, ClassRI}: BreakProhibited,

	// LB30b: Do not break between an emoji base and an emoji modifier
	{ClassEB, ClassEM}: BreakProhibited,
	{ClassEM, ClassEM}: BreakProhibited,

	// Indic conjunct rules (VF, VI, AS, AK, AP)
	{ClassAS, ClassVI}: BreakProhibited,
	{ClassAS, ClassAK}: BreakProhibited,
	{ClassAK, ClassVI}: BreakProhibited,
	{ClassAK, ClassVF}: BreakProhibited,
	{ClassAK, ClassAK}: BreakProhibited,
	{ClassAP, ClassAK}: BreakProhibited,
	{ClassAP, ClassAS}: BreakProhibited,
	{ClassVI, ClassAK}: BreakProhibited,
	{ClassVF, ClassAK}: BreakProhibited,
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
	lastNonSpaceClass := prevClass // Track last non-SP class for LB14

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
					// BA/B2 characters: allow break, but not immediately before SP or CM
					// The break should happen after the following space/character
					if currClass != ClassSP && currClass != ClassCM && currClass != ClassGL {
						// BA/B2 × ! (SP|CM|GL) - break after BA/B2
						if !(isSoftHyphen && hyphens == HyphensNone) {
							bytePos := len(string(runes[:i]))
							breakPoints = append(breakPoints, bytePos)
						}
					}
					// BA/B2 × SP - don't break here; let space create the break
					// BA/B2 × CM - don't break before combining mark
					// BA/B2 × GL - don't break before glue
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
