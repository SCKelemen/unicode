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

	// Control characters
	if unicode.Is(unicode.Cc, r) || unicode.Is(unicode.Cf, r) ||
		unicode.Is(unicode.Cs, r) || unicode.Is(unicode.Co, r) ||
		unicode.Is(unicode.Cn, r) {
		// Exclude CR, LF, ZWJ, and ZWNJ
		if r != 0x000D && r != 0x000A && r != 0x200C && r != 0x200D {
			return GBControl
		}
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

	// Prepend
	if isPrepend(r) {
		return GBPrepend
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
	// Prepend characters (simplified list)
	prepends := []rune{
		0x0600, 0x0601, 0x0602, 0x0603, 0x0604, 0x0605,
		0x06DD, 0x070F, 0x0890, 0x0891, 0x08E2,
		0x110BD, 0x110CD,
	}
	for _, p := range prepends {
		if r == p {
			return true
		}
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
	riCount := 0       // Count of consecutive Regional Indicators

	for i := 1; i < len(runes); i++ {
		prev := getGraphemeBreakClass(runes[i-1])
		curr := getGraphemeBreakClass(runes[i])

		shouldBreak := true

		// GB3: Don't break between CR and LF
		if prev == GBCR && curr == GBLF {
			shouldBreak = false
		} else if prev == GBCR || prev == GBLF || prev == GBControl {
			// GB4: Break after Control
			shouldBreak = true
		} else if curr == GBCR || curr == GBLF || curr == GBControl {
			// GB5: Break before Control
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
			// Count backwards to see if we have an even or odd number of RIs before this position
			riCount++
			if riCount%2 == 1 {
				// Odd number of RIs - don't break (pair them)
				shouldBreak = false
			}
		}

		// Reset RI count if not continuing RI sequence
		if prev != GBRegionalIndicator {
			riCount = 0
		}
		if curr == GBRegionalIndicator && prev == GBRegionalIndicator {
			// Continue counting
		} else if curr == GBRegionalIndicator {
			riCount = 1
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
