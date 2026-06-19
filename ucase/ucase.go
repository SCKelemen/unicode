// Package ucase implements the Unicode default (language-insensitive) case
// algorithms: case folding, lowercasing, uppercasing, and titlecasing, together
// with the Cased and Case_Ignorable properties and default caseless matching.
//
// # Specification
//
// Case mapping is NOT a numbered Unicode annex. It is specified in the core
// Unicode Standard, Section 3.13 "Default Case Algorithms", which defines
// toCasefold, toLowercase, toUppercase, and toTitlecase, the Cased (D135) and
// Case_Ignorable (D136) derived properties, the casing context conditions
// (Final_Sigma, After_Soft_Dotted, More_Above, Before_Dot, After_I), and the
// caseless matching definitions D144-D147.
//
//   - The Unicode Standard, Section 3.13 "Default Case Algorithms":
//     https://www.unicode.org/versions/latest/ (see the core specification, chapter 3)
//   - UAX #44 (Unicode Character Database): https://www.unicode.org/reports/tr44/
//
// # Data sources
//
// The tables are generated for Unicode 17.0.0 from the UCD:
//
//   - https://www.unicode.org/Public/17.0.0/ucd/UnicodeData.txt
//     (simple 1:1 mappings, fields 12/13/14)
//   - https://www.unicode.org/Public/17.0.0/ucd/SpecialCasing.txt
//     (full 1:many and conditional mappings)
//   - https://www.unicode.org/Public/17.0.0/ucd/CaseFolding.txt
//     (case folding, status C/F/S/T)
//   - https://www.unicode.org/Public/17.0.0/ucd/DerivedCoreProperties.txt
//     (Cased, Case_Ignorable)
//
// # Scope
//
// This package implements only the DEFAULT, language-insensitive algorithms.
// The full (possibly 1:many) mappings are applied, including the single
// non-language context condition that the default algorithm requires,
// Final_Sigma (used to produce the Greek final sigma when lowercasing).
//
// Locale tailoring (the Turkish/Azeri "tr"/"az" and Lithuanian "lt" mappings in
// SpecialCasing.txt) is OUT OF SCOPE for this version and is documented as
// future work. Because the remaining casing contexts (After_Soft_Dotted,
// More_Above, Before_Dot, After_I) appear in SpecialCasing.txt exclusively on
// those language-tagged lines, they are not part of the default algorithm and
// are intentionally not evaluated here; they will be added alongside locale
// tailoring.
//
// CaselessMatch implements the default caseless match of D144 (toCasefold-based
// equality). The canonical caseless match of D145 additionally requires
// canonical decomposition (NFD); normalization is out of scope for this
// package, so D145/D146 (compatibility) are not provided. Callers that need a
// canonical caseless match can normalize with a normalization package (e.g.
// github.com/SCKelemen/unicode/v6/uts15) before calling CaselessMatch.
//
// # Usage
//
//	import "github.com/SCKelemen/unicode/v6/ucase"
//
//	ucase.ToLower("ΟΔΟΣ")            // "οδος"  (final sigma)
//	ucase.ToUpper("straße")          // "STRASSE"
//	ucase.ToTitle("hello world")     // "Hello World"
//	ucase.ToFold("ﬀ")                // "ff"
//	ucase.CaselessMatch("Σίσυφος", "ΣΊΣΥΦΟΣ") // true
package ucase

import (
	"slices"
	"strings"
)

// caseKind selects which case mapping a transformation applies.
type caseKind uint8

const (
	kindLower caseKind = iota
	kindUpper
	kindTitle
)

// ToFold returns s after applying the Unicode default full case folding
// (toCasefold, Section 3.13), using the C (common) and F (full) mappings of
// CaseFolding.txt. Case folding is context-free and locale-independent; it is
// the canonical operation for caseless string comparison.
func ToFold(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	changed := false
	for _, r := range s {
		if t, ok := lookupFold(r); ok {
			b.WriteString(t)
			changed = true
		} else {
			b.WriteRune(r)
		}
	}
	if !changed {
		return s
	}
	return b.String()
}

// ToLower returns s mapped to lowercase using the Unicode default full
// toLowercase algorithm (Section 3.13). Full (1:many) and the Final_Sigma
// conditional mapping are applied; locale tailoring is not.
func ToLower(s string) string { return mapString(s, kindLower) }

// ToUpper returns s mapped to uppercase using the Unicode default full
// toUppercase algorithm (Section 3.13). Full (1:many) mappings are applied
// (for example, U+00DF SHARP S maps to "SS"); locale tailoring is not.
func ToUpper(s string) string { return mapString(s, kindUpper) }

// ToTitle returns s mapped to titlecase using the Unicode default full
// toTitlecase algorithm (Section 3.13): the first cased character of each word
// is titlecased and all other cased characters are lowercased.
//
// Word boundaries use a self-contained, documented definition rather than the
// full UAX #29 word segmentation: a new word begins at the first cased
// character that follows the start of the string or a character that is neither
// cased nor case-ignorable (typically whitespace, punctuation, or a symbol).
// Case-ignorable characters (such as U+0027 APOSTROPHE and combining marks) do
// not start or end a word, so "don't" titlecases to "Don't" (not "Don'T") and
// "o'brien" to "O'brien". Conditional lowercasing (Final_Sigma) is honored
// within the remainder of each word.
func ToTitle(s string) string {
	runes := []rune(s)
	out := make([]rune, 0, len(runes)+4)
	inWord := false
	for i, r := range runes {
		switch {
		case IsCased(r):
			if inWord {
				out = appendMapping(out, runes, i, kindLower)
			} else {
				out = appendMapping(out, runes, i, kindTitle)
				inWord = true
			}
		case IsCaseIgnorable(r):
			// Case-ignorable characters do not change word state.
			out = appendMapping(out, runes, i, kindLower)
		default:
			inWord = false
			out = appendMapping(out, runes, i, kindLower)
		}
	}
	return string(out)
}

// mapString applies the full default case mapping of the given kind to every
// character of s, honoring the Final_Sigma context condition.
func mapString(s string, kind caseKind) string {
	runes := []rune(s)
	out := make([]rune, 0, len(runes)+4)
	for i := range runes {
		out = appendMapping(out, runes, i, kind)
	}
	return string(out)
}

// appendMapping appends the full case mapping of runes[i] (for the given kind)
// to dst and returns the extended slice. It applies, in order of precedence:
// a matching conditional special mapping, an unconditional special mapping, the
// simple mapping, and finally the identity mapping.
func appendMapping(dst []rune, runes []rune, i int, kind caseKind) []rune {
	r := runes[i]
	for _, e := range specialFor(r) {
		if e.cond == condNone || (e.cond == condFinalSigma && finalSigma(runes, i)) {
			switch kind {
			case kindLower:
				return append(dst, e.lower...)
			case kindUpper:
				return append(dst, e.upper...)
			default:
				return append(dst, e.title...)
			}
		}
	}
	if up, lo, ti, ok := lookupSimple(r); ok {
		var m rune
		switch kind {
		case kindLower:
			m = lo
		case kindUpper:
			m = up
		default:
			m = ti
		}
		if m != 0 {
			return append(dst, m)
		}
	}
	return append(dst, r)
}

// SimpleFold returns the simple (1:1) case fold of r using the C (common) and S
// (simple) mappings of CaseFolding.txt, or r itself if it has none.
//
// Note: this is the Unicode "simple case folding" of Section 3.13, which
// differs from the standard library's unicode.SimpleFold (which iterates over
// the members of a case-equivalence orbit).
func SimpleFold(r rune) rune {
	if to, ok := lookupSimpleFold(r); ok {
		return to
	}
	return r
}

// SimpleLower returns the simple (1:1) lowercase mapping of r, or r if none.
func SimpleLower(r rune) rune {
	if _, lo, _, ok := lookupSimple(r); ok && lo != 0 {
		return lo
	}
	return r
}

// SimpleUpper returns the simple (1:1) uppercase mapping of r, or r if none.
func SimpleUpper(r rune) rune {
	if up, _, _, ok := lookupSimple(r); ok && up != 0 {
		return up
	}
	return r
}

// SimpleTitle returns the simple (1:1) titlecase mapping of r, or r if none.
// Per UAX #44, the titlecase mapping defaults to the uppercase mapping when no
// distinct titlecase mapping exists; that default is baked into the table.
func SimpleTitle(r rune) rune {
	if _, _, ti, ok := lookupSimple(r); ok && ti != 0 {
		return ti
	}
	return r
}

// IsCased reports whether r has the Cased property (Section 3.13, D135): r is
// lowercase, uppercase, or has General_Category Titlecase_Letter (Lt).
func IsCased(r rune) bool { return inRanges(casedRanges, r) }

// IsCaseIgnorable reports whether r has the Case_Ignorable property
// (Section 3.13, D136). Case-ignorable characters are skipped when evaluating
// the Final_Sigma context condition.
func IsCaseIgnorable(r rune) bool { return inRanges(caseIgnorableRanges, r) }

// CaselessMatch reports whether a and b are default caseless equivalents per
// the Unicode definition D144: toCasefold(a) == toCasefold(b).
//
// This is the "default caseless match". It does not perform canonical
// normalization, so it is not the canonical caseless match of D145. To compare
// strings that may differ in canonical composition, normalize both to NFD (or
// NFC) before calling.
func CaselessMatch(a, b string) bool {
	return ToFold(a) == ToFold(b)
}

// finalSigma reports whether the Final_Sigma context condition holds for the
// character at index i of runes (Section 3.13, Table "Context Specification for
// Casing"). It is true when i is preceded by a cased character followed by zero
// or more case-ignorable characters, and is NOT followed by zero or more
// case-ignorable characters and then a cased character.
func finalSigma(runes []rune, i int) bool {
	beforeCased := false
	for j := i - 1; j >= 0; j-- {
		if IsCaseIgnorable(runes[j]) {
			continue
		}
		beforeCased = IsCased(runes[j])
		break
	}
	if !beforeCased {
		return false
	}
	afterCased := false
	for j := i + 1; j < len(runes); j++ {
		if IsCaseIgnorable(runes[j]) {
			continue
		}
		afterCased = IsCased(runes[j])
		break
	}
	return !afterCased
}

// specialFor returns the full case mapping entries for r (conditional entries
// first), or nil if r has no special-casing mapping.
func specialFor(r rune) []specialCaseEntry {
	i, ok := slices.BinarySearchFunc(specialCaseData, r, func(e specialCaseEntry, t rune) int {
		switch {
		case e.cp < t:
			return -1
		case e.cp > t:
			return 1
		default:
			return 0
		}
	})
	if !ok {
		return nil
	}
	// Back up to the first entry for this code point.
	start := i
	for start > 0 && specialCaseData[start-1].cp == r {
		start--
	}
	end := i
	for end < len(specialCaseData) && specialCaseData[end].cp == r {
		end++
	}
	return specialCaseData[start:end]
}

// lookupSimple returns the simple case mappings of r.
func lookupSimple(r rune) (up, lo, ti rune, ok bool) {
	i, found := slices.BinarySearchFunc(simpleCaseData, r, func(e simpleCaseEntry, t rune) int {
		switch {
		case e.cp < t:
			return -1
		case e.cp > t:
			return 1
		default:
			return 0
		}
	})
	if !found {
		return 0, 0, 0, false
	}
	e := simpleCaseData[i]
	return e.up, e.lo, e.ti, true
}

// lookupSimpleFold returns the simple case fold of r.
func lookupSimpleFold(r rune) (rune, bool) {
	i, found := slices.BinarySearchFunc(simpleFoldData, r, func(e runePair, t rune) int {
		switch {
		case e.cp < t:
			return -1
		case e.cp > t:
			return 1
		default:
			return 0
		}
	})
	if !found {
		return 0, false
	}
	return simpleFoldData[i].to, true
}

// lookupFold returns the full case fold of r.
func lookupFold(r rune) (string, bool) {
	i, found := slices.BinarySearchFunc(foldFullData, r, func(e foldEntry, t rune) int {
		switch {
		case e.cp < t:
			return -1
		case e.cp > t:
			return 1
		default:
			return 0
		}
	})
	if !found {
		return "", false
	}
	return foldFullData[i].target, true
}

// inRanges reports whether r falls within any inclusive range in ranges, which
// must be sorted ascending by lo and be non-overlapping.
func inRanges(ranges []runeRange, r rune) bool {
	_, found := slices.BinarySearchFunc(ranges, r, func(rg runeRange, t rune) int {
		switch {
		case rg.hi < t:
			return -1
		case rg.lo > t:
			return 1
		default:
			return 0
		}
	})
	return found
}
