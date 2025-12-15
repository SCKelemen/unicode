package uax29

import "unicode"

// GraphemeBreakClass represents the grapheme cluster break property values
type GraphemeBreakClass int

const (
	GBOther GraphemeBreakClass = iota
	GBCR
	GBLF
	GBControl
	GBExtend
	GBZWJ
	GBRegionalIndicator
	GBPrepend
	GBSpacingMark
	GBL // Hangul L
	GBV // Hangul V
	GBT // Hangul T
	GBLV
	GBLVT
	GBExtendedPictographic
)

// getGraphemeBreakClass returns the grapheme cluster break class for a rune
func getGraphemeBreakClass(r rune) GraphemeBreakClass {
	// CR and LF
	if r == 0x000D {
		return GBCR
	}
	if r == 0x000A {
		return GBLF
	}

	// ZWJ
	if r == 0x200D {
		return GBZWJ
	}

	// Regional Indicators (U+1F1E6..U+1F1FF)
	if r >= 0x1F1E6 && r <= 0x1F1FF {
		return GBRegionalIndicator
	}

	// ZWNJ (U+200C) is treated as Extend
	if r == 0x200C {
		return GBExtend
	}

	// Prepend (must check before Control, as many Prepend chars are in Cf category)
	if isPrepend(r) {
		return GBPrepend
	}

	// Control characters
	// Note: Most Cf characters are Control, but Prepend chars (checked above) are excluded
	// Note: Only specific Cn (unassigned) chars are Control, not all of them
	// We use a simplified check here - a full implementation would need GraphemeBreakProperty.txt
	if unicode.Is(unicode.Cc, r) {
		// Exclude CR, LF (which have their own classes)
		if r != 0x000D && r != 0x000A {
			return GBControl
		}
	}
	// Common Cf control characters (excluding Prepend which was checked earlier)
	if (r >= 0x200B && r <= 0x200F) || (r >= 0x202A && r <= 0x202E) || (r >= 0x2060 && r <= 0x2064) || (r >= 0x2066 && r <= 0x206F) {
		return GBControl
	}
	if r == 0x00AD || r == 0x061C || r == 0x180E || r == 0xFEFF || (r >= 0xFFF9 && r <= 0xFFFB) {
		return GBControl
	}
	// Line and paragraph separators
	if r == 0x2028 || r == 0x2029 {
		return GBControl
	}

	// Hangul Syllables
	if r >= 0xAC00 && r <= 0xD7A3 {
		syllableIndex := r - 0xAC00
		if syllableIndex%28 == 0 {
			return GBLV
		}
		return GBLVT
	}

	// Hangul Jamo
	if r >= 0x1100 && r <= 0x115F { // L Jamo
		return GBL
	}
	if r >= 0xA960 && r <= 0xA97C { // Extended L Jamo
		return GBL
	}
	if (r >= 0x1160 && r <= 0x11A7) || (r >= 0xD7B0 && r <= 0xD7C6) { // V Jamo
		return GBV
	}
	if (r >= 0x11A8 && r <= 0x11FF) || (r >= 0xD7CB && r <= 0xD7FB) { // T Jamo
		return GBT
	}

	// Emoji modifiers (skin tones) are Extend (must check before ExtendedPictographic)
	if r >= 0x1F3FB && r <= 0x1F3FF {
		return GBExtend
	}

	// Extended Pictographic
	// This is a simplified check - full implementation would need the full Unicode data
	if isExtendedPictographic(r) {
		return GBExtendedPictographic
	}

	// Extend (combining marks, modifiers)
	if unicode.Is(unicode.Me, r) || unicode.Is(unicode.Mn, r) {
		return GBExtend
	}

	// Spacing marks
	// Note: Only SOME Mc characters are SpacingMark for grapheme breaking
	// Myanmar U+102C is Mc but NOT SpacingMark
	if unicode.Is(unicode.Mc, r) {
		// Exclude Myanmar vowel signs that are not SpacingMark
		if (r >= 0x1023 && r <= 0x1030 && r != 0x1031) || (r >= 0x1040 && r <= 0x104F) {
			return GBOther
		}
		return GBSpacingMark
	}

	return GBOther
}

// isExtendedPictographic checks if a rune is an extended pictographic character
// This is a simplified version - a complete implementation would use Unicode data files
func isExtendedPictographic(r rune) bool {
	// Emoji modifiers (skin tones) are NOT ExtendedPictographic, they are Extend
	if r >= 0x1F3FB && r <= 0x1F3FF {
		return false
	}
	// Copyright, registered, trade mark
	if r == 0x00A9 || r == 0x00AE {
		return true
	}
	// Misc text symbols
	if r >= 0x203C && r <= 0x3299 {
		// Specific ExtPict ranges within this broader range
		if (r >= 0x203C && r <= 0x2049) || // ‼️, ⁉️
			(r >= 0x2122 && r <= 0x2139) || // ™️, ℹ️
			(r >= 0x2194 && r <= 0x21AA) || // Arrows
			(r >= 0x231A && r <= 0x231B) || // Watches
			r == 0x2328 || // Keyboard
			r == 0x23CF || // Eject
			(r >= 0x23E9 && r <= 0x23F3) || // Media controls
			(r >= 0x23F8 && r <= 0x23FA) || // Pause, play
			r == 0x24C2 || // Ⓜ️
			(r >= 0x25AA && r <= 0x25FE) || // Squares
			(r >= 0x2600 && r <= 0x27BF && !(r >= 0x2700 && r <= 0x2704)) || // Misc symbols (except scissors)
			(r >= 0x2934 && r <= 0x2935) || // Arrows
			(r >= 0x2B00 && r <= 0x2BFF) || // Misc symbols and arrows
			r == 0x3030 || r == 0x303D || // Wavy dash, part alternation
			r == 0x3297 || r == 0x3299 { // Circled ideographs
			return true
		}
		return false
	}
	// Common emoji ranges (excluding modifiers)
	if r >= 0x1F300 && r <= 0x1F5FF {
		return true
	}
	if r >= 0x1F600 && r <= 0x1F64F {
		return true
	}
	if r >= 0x1F680 && r <= 0x1F6FF {
		return true
	}
	if r >= 0x1F900 && r <= 0x1F9FF {
		return true
	}
	// Misc Symbols (sparse - only some are ExtPict)
	if r >= 0x1F000 && r <= 0x1F02F {
		return true
	}
	if r >= 0x1FA00 && r <= 0x1FAFF {
		return true
	}
	return false
}

// isIndicConjunctLinker checks if a rune has InCB=Linker property (virama, etc.)
func isIndicConjunctLinker(r rune) bool {
	// Common virama/linker characters
	linkers := []rune{
		0x094D, 0x09CD, 0x0ACD, 0x0B4D, 0x0C4D, 0x0D4D, // Devanagari, Bengali, Gujarati, Oriya, Telugu, Malayalam
		0x1039, 0x17D2, 0x1A60, 0x1B44, 0x1BAB, 0xA9C0, 0xAAF6, // Myanmar, Khmer, Tai Tham, Balinese, Sundanese, Javanese, Meetei
		0x10A3F, 0x11133, 0x113D0, 0x1193E, 0x11A47, 0x11A99, // Kharoshthi, Chakma, Tulu-Tigalari, Dives Akuru, Zanabazar, Soyombo
		0x11C3F, 0x11D45, 0x11D97, // Bhaiksuki, Masaram Gondi, Gunjala Gondi
	}
	for _, linker := range linkers {
		if r == linker {
			return true
		}
	}
	return false
}

// isIndicConjunctConsonant checks if a rune has InCB=Consonant property
func isIndicConjunctConsonant(r rune) bool {
	// Devanagari consonants
	if (r >= 0x0915 && r <= 0x0939) || (r >= 0x0958 && r <= 0x095F) || (r >= 0x0978 && r <= 0x097F) {
		return true
	}
	// Bengali consonants
	if (r >= 0x0995 && r <= 0x09A8) || (r >= 0x09AA && r <= 0x09B0) || r == 0x09B2 ||
		(r >= 0x09B6 && r <= 0x09B9) || (r >= 0x09DC && r <= 0x09DD) || r == 0x09DF ||
		(r >= 0x09F0 && r <= 0x09F1) {
		return true
	}
	// Gujarati consonants
	if (r >= 0x0A95 && r <= 0x0AA8) || (r >= 0x0AAA && r <= 0x0AB0) ||
		(r >= 0x0AB2 && r <= 0x0AB3) || (r >= 0x0AB5 && r <= 0x0AB9) || r == 0x0AF9 {
		return true
	}
	// Oriya consonants
	if (r >= 0x0B15 && r <= 0x0B28) || (r >= 0x0B2A && r <= 0x0B30) ||
		(r >= 0x0B32 && r <= 0x0B33) || (r >= 0x0B35 && r <= 0x0B39) ||
		(r >= 0x0B5C && r <= 0x0B5D) || r == 0x0B5F || r == 0x0B71 {
		return true
	}
	// Telugu consonants
	if (r >= 0x0C15 && r <= 0x0C28) || (r >= 0x0C2A && r <= 0x0C39) ||
		(r >= 0x0C58 && r <= 0x0C5A) || (r >= 0x0C78 && r <= 0x0C7F) {
		return true
	}
	// Malayalam consonants
	if (r >= 0x0D15 && r <= 0x0D28) || (r >= 0x0D2A && r <= 0x0D39) ||
		(r >= 0x0D54 && r <= 0x0D56) || (r >= 0x0D5F && r <= 0x0D61) || (r >= 0x0D7A && r <= 0x0D7F) {
		return true
	}
	// Myanmar consonants
	if (r >= 0x1000 && r <= 0x102A) || r == 0x103F || (r >= 0x1050 && r <= 0x1055) ||
		(r >= 0x105A && r <= 0x105D) || r == 0x1061 || (r >= 0x1065 && r <= 0x1066) ||
		(r >= 0x106E && r <= 0x1070) || (r >= 0x1075 && r <= 0x1081) || r == 0x108E {
		return true
	}
	// Balinese consonants
	if (r >= 0x1B0B && r <= 0x1B0C) || (r >= 0x1B13 && r <= 0x1B33) || (r >= 0x1B45 && r <= 0x1B4C) {
		return true
	}
	// Sundanese consonants
	if (r >= 0x1B83 && r <= 0x1BA0) || (r >= 0x1BAE && r <= 0x1BAF) || (r >= 0x1BBB && r <= 0x1BBD) {
		return true
	}
	// Khmer consonants
	if (r >= 0x1780 && r <= 0x17A2) || (r >= 0x17A5 && r <= 0x17A7) || (r >= 0x17A9 && r <= 0x17B3) {
		return true
	}
	return false
}

// isPrepend checks if a rune has the Prepend property
func isPrepend(r rune) bool {
	// Prepend characters from Unicode GraphemeBreakProperty.txt
	if r >= 0x0600 && r <= 0x0605 {
		return true
	}
	if r == 0x06DD || r == 0x070F {
		return true
	}
	if r >= 0x0890 && r <= 0x0891 {
		return true
	}
	if r == 0x08E2 || r == 0x0D4E {
		return true
	}
	if r == 0x110BD || r == 0x110CD {
		return true
	}
	if r >= 0x111C2 && r <= 0x111C3 {
		return true
	}
	if r == 0x113D1 || r == 0x1193F || r == 0x11941 {
		return true
	}
	if r >= 0x11A84 && r <= 0x11A89 {
		return true
	}
	if r == 0x11D46 || r == 0x11F02 {
		return true
	}
	return false
}

// FindGraphemeBreaks returns the byte positions where grapheme cluster breaks occur
func FindGraphemeBreaks(text string) []int {
	if len(text) == 0 {
		return []int{}
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return []int{}
	}

	breaks := []int{0} // GB1: Break at start

	for i := 1; i < len(runes); i++ {
		prev := getGraphemeBreakClass(runes[i-1])
		curr := getGraphemeBreakClass(runes[i])

		shouldBreak := true

		// GB3: Don't break between CR and LF
		if prev == GBCR && curr == GBLF {
			shouldBreak = false
		} else if prev == GBCR || prev == GBLF || prev == GBControl {
			// GB4: Break after Control/CR/LF
			shouldBreak = true
		} else if curr == GBCR || curr == GBLF || curr == GBControl {
			// GB5: Break before Control/CR/LF
			shouldBreak = true
		} else if prev == GBL && (curr == GBL || curr == GBV || curr == GBLV || curr == GBLVT) {
			// GB6: Don't break Hangul L with following
			shouldBreak = false
		} else if (prev == GBLV || prev == GBV) && (curr == GBV || curr == GBT) {
			// GB7: Don't break Hangul vowels/finals
			shouldBreak = false
		} else if (prev == GBLVT || prev == GBT) && curr == GBT {
			// GB8: Don't break Hangul finals
			shouldBreak = false
		} else if curr == GBExtend || curr == GBZWJ {
			// GB9: Don't break before Extend or ZWJ
			shouldBreak = false
		} else if curr == GBSpacingMark {
			// GB9a: Don't break before SpacingMark
			shouldBreak = false
		} else if prev == GBPrepend {
			// GB9b: Don't break after Prepend
			shouldBreak = false
		} else if isIndicConjunctConsonant(runes[i]) {
			// GB9c: InCB=Consonant [InCB=Extend InCB=Linker]* InCB=Linker [InCB=Extend InCB=Linker]* × InCB=Consonant
			// Check if there's a Linker before current Consonant (with optional Extend/ZWJ/Linker in between)
			j := i - 1
			foundLinker := false
			// Skip back through Extend, ZWJ, and Linker characters
			for j >= 0 {
				// Check for Linker first (before checking Extend, since Linkers are also Extend)
				if isIndicConjunctLinker(runes[j]) {
					foundLinker = true
					j--
					break
				}
				rClass := getGraphemeBreakClass(runes[j])
				if rClass == GBExtend || rClass == GBZWJ {
					j--
					continue
				}
				break
			}
			if foundLinker {
				// Continue looking back through Extend/Linker for a Consonant
				for j >= 0 {
					// Check for Consonant
					if isIndicConjunctConsonant(runes[j]) {
						// Found the pattern: Consonant ... Linker ... Consonant
						shouldBreak = false
						break
					}
					// Check for Linker
					if isIndicConjunctLinker(runes[j]) {
						j--
						continue
					}
					rClass := getGraphemeBreakClass(runes[j])
					if rClass == GBExtend || rClass == GBZWJ {
						j--
						continue
					}
					break
				}
			}
		} else if curr == GBExtendedPictographic {
			// GB11: ExtendedPictographic Extend* ZWJ × ExtendedPictographic
			// Check if there's a ZWJ before current position (with optional Extend in between)
			j := i - 1
			// Skip any Extend characters
			for j >= 0 && getGraphemeBreakClass(runes[j]) == GBExtend {
				j--
			}
			// Check if we have ZWJ
			if j >= 0 && getGraphemeBreakClass(runes[j]) == GBZWJ {
				// Now look back further for ExtendedPictographic (with optional Extend in between)
				j--
				for j >= 0 && getGraphemeBreakClass(runes[j]) == GBExtend {
					j--
				}
				if j >= 0 && getGraphemeBreakClass(runes[j]) == GBExtendedPictographic {
					// We found the pattern: ExtPict Extend* ZWJ Extend* ExtPict
					shouldBreak = false
				}
			}
		} else if prev == GBRegionalIndicator && curr == GBRegionalIndicator {
			// GB12/GB13: Regional Indicator pairs
			// Count how many consecutive RIs come before the current position
			riCountBefore := 0
			for j := i - 1; j >= 0 && getGraphemeBreakClass(runes[j]) == GBRegionalIndicator; j-- {
				riCountBefore++
			}
			// If odd number of RIs before, don't break (pair with previous RI)
			if riCountBefore%2 == 1 {
				shouldBreak = false
			}
		}

		if shouldBreak {
			// Calculate byte position
			bytePos := 0
			for j := 0; j < i; j++ {
				bytePos += len(string(runes[j]))
			}
			breaks = append(breaks, bytePos)
		}
	}

	// GB2: Break at end
	breaks = append(breaks, len(text))

	return breaks
}

// Graphemes splits text into grapheme clusters
func Graphemes(text string) []string {
	breaks := FindGraphemeBreaks(text)
	if len(breaks) <= 1 {
		return []string{}
	}

	result := make([]string, len(breaks)-1)
	for i := 0; i < len(breaks)-1; i++ {
		result[i] = text[breaks[i]:breaks[i+1]]
	}
	return result
}
