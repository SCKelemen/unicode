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
	if unicode.Is(unicode.Mc, r) {
		return GBSpacingMark
	}

	return GBOther
}

// isExtendedPictographic checks if a rune is an extended pictographic character
// This is a simplified version - a complete implementation would use Unicode data files
func isExtendedPictographic(r rune) bool {
	// Common emoji ranges
	if r >= 0x1F300 && r <= 0x1F9FF {
		return true
	}
	if r >= 0x2600 && r <= 0x27BF {
		return true
	}
	if r >= 0x1F000 && r <= 0x1F02F {
		return true
	}
	if r >= 0x1FA00 && r <= 0x1FAFF {
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
		} else if prev == GBExtendedPictographic {
			// GB11: Check for emoji sequences
			// Look back for ExtendedPictographic followed by Extend* ZWJ
			lookback := i - 1
			foundZWJ := false
			for lookback > 0 && getGraphemeBreakClass(runes[lookback]) == GBExtend {
				lookback--
			}
			if lookback > 0 && getGraphemeBreakClass(runes[lookback]) == GBZWJ {
				lookback--
				for lookback >= 0 && getGraphemeBreakClass(runes[lookback]) == GBExtend {
					lookback--
				}
				if lookback >= 0 && getGraphemeBreakClass(runes[lookback]) == GBExtendedPictographic {
					foundZWJ = true
				}
			}
			if foundZWJ && curr == GBExtendedPictographic {
				shouldBreak = false
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
