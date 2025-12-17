// Package uts15 implements Unicode Normalization Forms (UTS #15).
//
// This package provides Unicode text normalization, which is essential for:
//   - Text comparison and matching
//   - String equality testing across different representations
//   - Database indexing and searching
//   - Security identifier validation
//   - Data interchange and storage
//
// Based on: https://www.unicode.org/reports/tr15/
//
// # Normalization Forms
//
// Unicode provides four normalization forms:
//
// NFD (Canonical Decomposition):
//   - Decomposes characters into their canonical components
//   - Example: "é" (U+00E9) → "e" (U+0065) + "́" (U+0301)
//   - Use for comparing text where different representations should match
//
// NFC (Canonical Composition):
//   - Decomposes then recomposes using canonical equivalence
//   - Preferred form for most text processing and display
//   - Example: "e" + "́" → "é"
//   - Used by most modern applications and protocols
//
// NFKD (Compatibility Decomposition):
//   - Decomposes using compatibility equivalence
//   - More aggressive decomposition (loses formatting distinctions)
//   - Example: "ﬁ" (U+FB01) → "f" (U+0066) + "i" (U+0069)
//   - Use for searching and matching where formatting is irrelevant
//
// NFKC (Compatibility Composition):
//   - Compatibility decomposition followed by canonical composition
//   - Most aggressive normalization
//   - Use for identifier comparison and security-sensitive contexts
//
// # Conformance
//
// This implementation follows UTS #15 Unicode Normalization Forms:
//   - https://www.unicode.org/reports/tr15/
//
// The implementation uses normalization data from Unicode 17.0.0:
//   - Canonical decomposition mappings
//   - Compatibility decomposition mappings
//   - Canonical combining classes
//   - Composition exclusions
//
// # Usage
//
//	import "github.com/SCKelemen/unicode/uts15"
//
//	// Normalize to NFC (recommended for most uses)
//	text := "café"  // may be composed or decomposed
//	normalized := uts15.NFC(text)
//
//	// Compare strings reliably
//	s1 := "café"  // composed form
//	s2 := "café"  // decomposed form (e + combining accent)
//	if uts15.NFC(s1) == uts15.NFC(s2) {
//	    // Strings are equivalent
//	}
//
//	// Normalize for searching (NFKC)
//	query := "ﬁle"  // contains ligature
//	normalized := uts15.NFKC(query)  // "file"
//
//	// Check if already normalized
//	if uts15.IsNFC("café") {
//	    // No normalization needed
//	}
//
// # References
//
//   - UTS #15: https://www.unicode.org/reports/tr15/
//   - Unicode Normalization FAQ: https://unicode.org/faq/normalization.html
//   - UnicodeData.txt: https://www.unicode.org/Public/17.0.0/ucd/UnicodeData.txt
//   - DerivedNormalizationProps.txt: https://www.unicode.org/Public/17.0.0/ucd/DerivedNormalizationProps.txt
package uts15

// NFC returns the NFC (Canonical Composition) normalization of the string.
//
// NFC is the recommended normalization form for most uses. It produces
// composed characters when possible while maintaining canonical equivalence.
//
// Example:
//	NFC("e\u0301") // Returns "é" (U+00E9)
//
// See: https://www.unicode.org/reports/tr15/#Norm_Forms
func NFC(s string) string {
	// Decompose canonically
	decomposed := nfd(s)
	// Compose
	return compose(decomposed, false)
}

// NFD returns the NFD (Canonical Decomposition) normalization of the string.
//
// NFD decomposes all characters into their canonical components and
// orders combining marks by their canonical combining class.
//
// Example:
//	NFD("é") // Returns "e\u0301" (e + combining acute accent)
//
// See: https://www.unicode.org/reports/tr15/#Norm_Forms
func NFD(s string) string {
	return nfd(s)
}

// NFKC returns the NFKC (Compatibility Composition) normalization of the string.
//
// NFKC uses compatibility decomposition followed by canonical composition.
// It's more aggressive than NFC and removes formatting distinctions.
//
// Example:
//	NFKC("ﬁ") // Returns "fi" (U+0066 U+0069)
//
// Use NFKC for:
//   - Identifier comparison
//   - Security-sensitive validation
//   - Case-insensitive matching
//
// See: https://www.unicode.org/reports/tr15/#Norm_Forms
func NFKC(s string) string {
	// Decompose with compatibility
	decomposed := nfkd(s)
	// Compose
	return compose(decomposed, false)
}

// NFKD returns the NFKD (Compatibility Decomposition) normalization of the string.
//
// NFKD uses compatibility decomposition, which is more aggressive than NFD.
// It decomposes compatibility characters like ligatures and circled letters.
//
// Example:
//	NFKD("ﬁ") // Returns "fi"
//	NFKD("①") // Returns "1"
//
// See: https://www.unicode.org/reports/tr15/#Norm_Forms
func NFKD(s string) string {
	return nfkd(s)
}

// IsNFC reports whether the string is in NFC (Canonical Composition) form.
//
// This is more efficient than normalizing and comparing if you only need
// to check normalization status.
func IsNFC(s string) bool {
	return s == NFC(s)
}

// IsNFD reports whether the string is in NFD (Canonical Decomposition) form.
func IsNFD(s string) bool {
	return s == NFD(s)
}

// IsNFKC reports whether the string is in NFKC (Compatibility Composition) form.
func IsNFKC(s string) bool {
	return s == NFKC(s)
}

// IsNFKD reports whether the string is in NFKD (Compatibility Decomposition) form.
func IsNFKD(s string) bool {
	return s == NFKD(s)
}

// nfd performs canonical decomposition
func nfd(s string) string {
	runes := []rune(s)
	result := make([]rune, 0, len(runes)*2)

	for _, r := range runes {
		result = append(result, decomposeCanonical(r)...)
	}

	// Canonical ordering
	return string(canonicalOrder(result))
}

// nfkd performs compatibility decomposition
func nfkd(s string) string {
	runes := []rune(s)
	result := make([]rune, 0, len(runes)*2)

	for _, r := range runes {
		result = append(result, decomposeCompatibility(r)...)
	}

	// Canonical ordering
	return string(canonicalOrder(result))
}

// decomposeCanonical recursively decomposes a rune using canonical decomposition
func decomposeCanonical(r rune) []rune {
	// Check for Hangul syllable
	if isHangulSyllable(r) {
		return decomposeHangul(r)
	}

	// Look up canonical decomposition
	if decomp, ok := canonicalDecompositionMap[r]; ok {
		result := make([]rune, 0, len(decomp)*2)
		for _, dr := range decomp {
			result = append(result, decomposeCanonical(dr)...)
		}
		return result
	}

	return []rune{r}
}

// decomposeCompatibility recursively decomposes a rune using compatibility decomposition
func decomposeCompatibility(r rune) []rune {
	// Check for Hangul syllable
	if isHangulSyllable(r) {
		return decomposeHangul(r)
	}

	// Look up compatibility decomposition (includes canonical)
	if decomp, ok := compatibilityDecompositionMap[r]; ok {
		result := make([]rune, 0, len(decomp)*2)
		for _, dr := range decomp {
			result = append(result, decomposeCompatibility(dr)...)
		}
		return result
	}

	// Fall back to canonical
	if decomp, ok := canonicalDecompositionMap[r]; ok {
		result := make([]rune, 0, len(decomp)*2)
		for _, dr := range decomp {
			result = append(result, decomposeCompatibility(dr)...)
		}
		return result
	}

	return []rune{r}
}

// canonicalOrder sorts combining marks by their canonical combining class
func canonicalOrder(runes []rune) []rune {
	if len(runes) <= 1 {
		return runes
	}

	// Bubble sort for canonical ordering (stable and simple)
	// We need to preserve the order of combining marks with the same class
	for {
		changed := false
		for i := 0; i < len(runes)-1; i++ {
			class1 := getCombiningClass(runes[i])
			class2 := getCombiningClass(runes[i+1])

			// Only reorder if both are combining marks and in wrong order
			if class1 > 0 && class2 > 0 && class1 > class2 {
				runes[i], runes[i+1] = runes[i+1], runes[i]
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	return runes
}

// compose performs canonical composition
func compose(s string, compat bool) string {
	runes := []rune(s)
	if len(runes) <= 1 {
		return s
	}

	result := make([]rune, 0, len(runes))
	i := 0

	for i < len(runes) {
		starter := runes[i]
		i++

		// Check for Hangul L+V composition first (both are starters)
		if i < len(runes) && isHangulL(starter) && isHangulV(runes[i]) {
			starter = composeHangul(starter, runes[i])
			i++
			// Check for Hangul LV+T composition (but NOT if there are combining marks)
			if i < len(runes) && isHangulT(runes[i]) {
				starter = composeHangul(starter, runes[i])
				i++
			}
		}

		result = append(result, starter)

		// Skip if this is not a starter (combining class != 0)
		if getCombiningClass(starter) != 0 {
			continue
		}

		starterIdx := len(result) - 1

		// Try starter+starter composition (but only if no combining marks follow)
		if i < len(runes) && getCombiningClass(runes[i]) == 0 {
			for i < len(runes) && getCombiningClass(runes[i]) == 0 {
				if composed, ok := compositionMap[[2]rune{result[starterIdx], runes[i]}]; ok {
					result[starterIdx] = composed
					i++
				} else {
					break
				}
			}
		}

		// Collect all following combining marks
		combiningStart := len(result)
		for i < len(runes) && getCombiningClass(runes[i]) != 0 {
			result = append(result, runes[i])
			i++
		}

		// Try to compose starter with combining marks
		// Keep trying until no more compositions are possible
		changed := true
		for changed {
			changed = false
			j := combiningStart
			for j < len(result) {
				ch := result[j]
				chClass := getCombiningClass(ch)

				// Check if composition is blocked
				blocked := false
				for k := combiningStart; k < j; k++ {
					kClass := getCombiningClass(result[k])
					if kClass >= chClass && kClass > 0 {
						blocked = true
						break
					}
				}

				// Try to compose if not blocked
				if !blocked {
					if composed, ok := compositionMap[[2]rune{result[starterIdx], ch}]; ok {
						result[starterIdx] = composed
						// Remove the combining mark that was composed
						result = append(result[:j], result[j+1:]...)
						changed = true
						break
					}
				}
				j++
			}
		}

		// Check for Hangul LV+T composition with next starter
		// But only if no combining marks were added
		if len(result) == starterIdx+1 && i < len(runes) && isHangulLV(result[starterIdx]) && isHangulT(runes[i]) {
			result[starterIdx] = composeHangul(result[starterIdx], runes[i])
			i++
		}
	}

	return string(result)
}

// getCombiningClass returns the canonical combining class for a rune
func getCombiningClass(r rune) int {
	if class, ok := combiningClassMap[r]; ok {
		return class
	}
	return 0
}

// Hangul constants
const (
	hangulSBase  = 0xAC00
	hangulLBase  = 0x1100
	hangulVBase  = 0x1161
	hangulTBase  = 0x11A7
	hangulLCount = 19
	hangulVCount = 21
	hangulTCount = 28
	hangulNCount = hangulVCount * hangulTCount // 588
	hangulSCount = hangulLCount * hangulNCount // 11172
)

// isHangulSyllable reports whether r is a precomposed Hangul syllable
func isHangulSyllable(r rune) bool {
	return r >= hangulSBase && r < hangulSBase+hangulSCount
}

// isHangulL reports whether r is a Hangul L (leading consonant)
func isHangulL(r rune) bool {
	return r >= hangulLBase && r < hangulLBase+hangulLCount
}

// isHangulV reports whether r is a Hangul V (vowel)
func isHangulV(r rune) bool {
	return r >= hangulVBase && r < hangulVBase+hangulVCount
}

// isHangulT reports whether r is a Hangul T (trailing consonant)
func isHangulT(r rune) bool {
	return r > hangulTBase && r < hangulTBase+hangulTCount
}

// isHangulLV reports whether r is a Hangul LV syllable (no trailing consonant)
func isHangulLV(r rune) bool {
	if !isHangulSyllable(r) {
		return false
	}
	return (r-hangulSBase)%hangulTCount == 0
}

// decomposeHangul decomposes a Hangul syllable into L, V, and optionally T
func decomposeHangul(r rune) []rune {
	sIndex := int(r - hangulSBase)
	if sIndex < 0 || sIndex >= hangulSCount {
		return []rune{r}
	}

	l := rune(hangulLBase + sIndex/hangulNCount)
	v := rune(hangulVBase + (sIndex%hangulNCount)/hangulTCount)
	t := rune(hangulTBase + sIndex%hangulTCount)

	if t == hangulTBase {
		return []rune{l, v}
	}
	return []rune{l, v, t}
}

// composeHangul composes Hangul L+V or LV+T
func composeHangul(a, b rune) rune {
	// L + V -> LV
	if isHangulL(a) && isHangulV(b) {
		lIndex := int(a - hangulLBase)
		vIndex := int(b - hangulVBase)
		return rune(hangulSBase + (lIndex*hangulVCount+vIndex)*hangulTCount)
	}

	// LV + T -> LVT
	if isHangulLV(a) && isHangulT(b) {
		tIndex := int(b - hangulTBase)
		return a + rune(tIndex)
	}

	return a
}
