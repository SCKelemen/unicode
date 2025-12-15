package uax29

import "unicode"

// SentenceBreakClass represents the sentence break property values
type SentenceBreakClass int

const (
	SBOther SentenceBreakClass = iota
	SBCR
	SBLF
	SBSep
	SBExtend
	SBFormat
	SBSp
	SBLower
	SBUpper
	SBOLetter
	SBATerm
	SBSTerm
	SBNumeric
	SBSContinue
	SBClose
)

// getSentenceBreakClass returns the sentence break class for a rune
func getSentenceBreakClass(r rune) SentenceBreakClass {
	// CR and LF
	if r == 0x000D {
		return SBCR
	}
	if r == 0x000A {
		return SBLF
	}

	// Sep (paragraph separators)
	if r == 0x0085 || r == 0x2028 || r == 0x2029 {
		return SBSep
	}

	// ATerm (sentence terminators that look like periods)
	if r == 0x002E || r == 0x2024 || r == 0xFE52 || r == 0xFF0E {
		return SBATerm
	}

	// STerm (other sentence terminators)
	if r == 0x0021 || r == 0x003F || r == 0x0589 || r == 0x061D || r == 0x061F || r == 0x06D4 ||
		r == 0x0700 || r == 0x0701 || r == 0x0702 || r == 0x07F9 || r == 0x0964 || r == 0x0965 ||
		r == 0x104A || r == 0x104B || r == 0x1362 || r == 0x1367 || r == 0x1368 || r == 0x166E ||
		r == 0x1803 || r == 0x1809 || r == 0x1944 || r == 0x1945 || r == 0x1AA8 || r == 0x1AA9 ||
		r == 0x1AAA || r == 0x1AAB || r == 0x1B5A || r == 0x1B5B || r == 0x1B5E || r == 0x1B5F ||
		r == 0x1C3B || r == 0x1C3C || r == 0x1C7E || r == 0x1C7F || r == 0x203C || r == 0x203D ||
		r == 0x2047 || r == 0x2048 || r == 0x2049 || r == 0x2E2E || r == 0x2E3C || r == 0x3002 ||
		r == 0xA4FF || r == 0xA60E || r == 0xA60F || r == 0xA6F3 || r == 0xA6F7 || r == 0xA876 ||
		r == 0xA877 || r == 0xA8CE || r == 0xA8CF || r == 0xA92F || r == 0xA9C8 || r == 0xA9C9 ||
		r == 0xAA5D || r == 0xAA5E || r == 0xAA5F || r == 0xAAF0 || r == 0xAAF1 || r == 0xABEB ||
		r == 0xFE15 || r == 0xFE16 || r == 0xFE56 || r == 0xFE57 || r == 0xFF01 || r == 0xFF1F ||
		r == 0xFF61 || r == 0x10A56 || r == 0x10A57 || r == 0x11047 || r == 0x11048 || r == 0x110BE ||
		r == 0x110BF || r == 0x110C0 || r == 0x110C1 || r == 0x11141 || r == 0x11142 || r == 0x11143 ||
		r == 0x111C5 || r == 0x111C6 || r == 0x111CD || r == 0x111DE || r == 0x111DF || r == 0x11238 ||
		r == 0x11239 || r == 0x1123B || r == 0x1123C || r == 0x112A9 || r == 0x1144B || r == 0x1144C ||
		r == 0x115C2 || r == 0x115C3 || r == 0x115C9 || r == 0x115CA || r == 0x115CB || r == 0x115CC ||
		r == 0x115CD || r == 0x115CE || r == 0x115CF || r == 0x115D0 || r == 0x115D1 || r == 0x115D2 ||
		r == 0x115D3 || r == 0x115D4 || r == 0x115D5 || r == 0x115D6 || r == 0x115D7 || r == 0x11641 ||
		r == 0x11642 || r == 0x1173C || r == 0x1173D || r == 0x1173E || r == 0x11944 || r == 0x11946 ||
		r == 0x11A42 || r == 0x11A43 || r == 0x11A9B || r == 0x11A9C || r == 0x11C41 || r == 0x11C42 ||
		r == 0x11EF7 || r == 0x11EF8 || r == 0x16A6E || r == 0x16A6F || r == 0x16AF5 || r == 0x16B37 ||
		r == 0x16B38 || r == 0x16B44 || r == 0x16E98 || r == 0x1BC9F || r == 0x1DA88 {
		return SBSTerm
	}

	// SContinue (comma, colon, etc.)
	if r == 0x002C || r == 0x002D || r == 0x003A || r == 0x055D || r == 0x060C || r == 0x060D ||
		r == 0x07F8 || r == 0x1802 || r == 0x1808 || r == 0x2013 || r == 0x2014 || r == 0xFE10 ||
		r == 0xFE11 || r == 0xFE13 || r == 0xFE31 || r == 0xFE32 || r == 0xFE50 || r == 0xFE51 ||
		r == 0xFE55 || r == 0xFE58 || r == 0xFE63 || r == 0xFF0C || r == 0xFF0D || r == 0xFF1A ||
		r == 0xFF64 {
		return SBSContinue
	}

	// Close (closing punctuation - includes Pe, Pf, Ps, Pi, and some Po)
	if unicode.Is(unicode.Pe, r) || unicode.Is(unicode.Pf, r) || unicode.Is(unicode.Ps, r) || unicode.Is(unicode.Pi, r) {
		return SBClose
	}
	if r == 0x0022 || r == 0x0027 || r == 0x00BB || r == 0x2019 || r == 0x201D || r == 0x203A || r == 0x2E03 || r == 0x2E05 ||
		r == 0x2E0A || r == 0x2E0D || r == 0x2E1D || r == 0x2E21 {
		return SBClose
	}

	// Sp (whitespace, but not separators)
	if unicode.Is(unicode.Zs, r) || r == 0x0009 {
		return SBSp
	}

	// Lower
	if unicode.IsLower(r) {
		return SBLower
	}

	// Upper
	if unicode.IsUpper(r) || unicode.IsTitle(r) {
		return SBUpper
	}

	// Numeric
	if unicode.IsDigit(r) {
		return SBNumeric
	}

	// Format
	if unicode.Is(unicode.Cf, r) && r != 0x200C && r != 0x200D {
		return SBFormat
	}

	// Extend
	if unicode.Is(unicode.Me, r) || unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r) || r == 0x200D {
		return SBExtend
	}

	// OLetter (other letters)
	if unicode.IsLetter(r) || r == 0x00A0 {
		return SBOLetter
	}

	return SBOther
}

// FindSentenceBreaks returns the byte positions where sentence breaks occur
func FindSentenceBreaks(text string) []int {
	if len(text) == 0 {
		return []int{}
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return []int{}
	}

	classes := make([]SentenceBreakClass, len(runes))
	for i, r := range runes {
		classes[i] = getSentenceBreakClass(r)
	}

	breaks := []int{0} // SB1: Break at start

	for i := 1; i < len(runes); i++ {
		// Get previous non-Format/Extend character for most rules (SB5)
		prevIdx := i - 1
		for prevIdx > 0 && (classes[prevIdx] == SBFormat || classes[prevIdx] == SBExtend) {
			prevIdx--
		}

		prev := classes[prevIdx]
		curr := classes[i]

		shouldBreak := false // SB998: Default is no break

		// SB3: Don't break within CRLF
		if classes[i-1] == SBCR && curr == SBLF {
			shouldBreak = false
		} else if classes[i-1] == SBCR || classes[i-1] == SBLF || classes[i-1] == SBSep {
			// SB4: Break after paragraph separators (check immediately previous, not prev with Format/Extend skipped)
			// This rule takes precedence over SB5, so we break even if curr is Format/Extend
			shouldBreak = true
		} else if curr == SBFormat || curr == SBExtend {
			// SB5: Ignore Format and Extend (but not after paragraph separators)
			shouldBreak = false
		} else if prev == SBATerm && curr == SBNumeric {
			// SB6: ATerm × Numeric
			shouldBreak = false
		} else if (prev == SBUpper || prev == SBLower) && curr == SBATerm {
			// SB7: (Upper | Lower) ATerm × Upper
			// Need to check if followed by Upper
			nextIdx := i + 1
			for nextIdx < len(runes) && (classes[nextIdx] == SBFormat || classes[nextIdx] == SBExtend) {
				nextIdx++
			}
			if nextIdx < len(runes) && classes[nextIdx] == SBUpper {
				shouldBreak = false
			}
		} else if prev == SBATerm {
			// Check ATerm-related rules (SB7, SB8, SB8a, SB9, SB10, SB11)

			// SB9: ATerm Close* × (Close | Sp | Sep | CR | LF)
			if curr == SBClose || curr == SBSp || curr == SBSep || curr == SBCR || curr == SBLF {
				shouldBreak = false
			} else if curr == SBUpper {
				// SB7: (Upper | Lower) ATerm × Upper
				// Check if there's a Letter before ATerm
				prevPrevIdx := prevIdx - 1
				for prevPrevIdx > 0 && (classes[prevPrevIdx] == SBFormat || classes[prevPrevIdx] == SBExtend) {
					prevPrevIdx--
				}
				if prevPrevIdx >= 0 && (classes[prevPrevIdx] == SBUpper || classes[prevPrevIdx] == SBLower) {
					// Pattern matches: Letter ATerm Upper - don't break (SB7)
					shouldBreak = false
				} else {
					// No letter before ATerm - SB11 will break
					shouldBreak = true
				}
			} else {
				// For other characters after ATerm, check SB8, SB8a, SB11
				// Look forward through Close* Sp* for what follows
				j := i
				for j < len(runes) && (classes[j] == SBClose || classes[j] == SBSp || classes[j] == SBFormat || classes[j] == SBExtend) {
					j++
				}

				if j >= len(runes) {
					// SB11: ATerm Close* Sp* <end of text>
					shouldBreak = true
				} else {
					next := classes[j]
					// SB8a: (STerm | ATerm) Close* Sp* × (SContinue | STerm | ATerm)
					if next == SBSContinue || next == SBSTerm || next == SBATerm {
						shouldBreak = false
					} else if next == SBLower {
						// SB8: ATerm Close* Sp* × (¬(OLetter | Upper | Lower | Sep | CR | LF | STerm | ATerm))* Lower
						shouldBreak = false
					} else {
						// SB11: Break after ATerm Close* Sp* if followed by anything else
						shouldBreak = true
					}
				}
			}
		} else if prev == SBClose || prev == SBSp {
			// When prev is Close or Sp, check if there's ATerm/STerm before it
			// Look back through Close* Sp* to find ATerm or STerm
			hasSpBeforePrev := false
			j := prevIdx - 1
			// Check if there's any Sp between ATerm and prev (not including curr)
			for j >= 0 && (classes[j] == SBClose || classes[j] == SBSp || classes[j] == SBFormat || classes[j] == SBExtend) {
				if classes[j] == SBSp {
					hasSpBeforePrev = true
				}
				j--
			}
			if j >= 0 && (classes[j] == SBATerm || classes[j] == SBSTerm) {
				// Found ATerm/STerm before Close* Sp*
				// If prev is Close and there's Sp before it, we're past the break point (already broken at Sp)
				if prev == SBClose && hasSpBeforePrev {
					// Pattern: ATerm Close* Sp Close* × curr
					// Break point was after Sp, not here
					shouldBreak = false
				} else if prev == SBSp && curr == SBClose {
					// Pattern: ATerm Close* Sp × Close
					// Check if there's lowercase ahead (SB8)
					k := i + 1
					for k < len(runes) && (classes[k] == SBClose || classes[k] == SBFormat || classes[k] == SBExtend) {
						k++
					}
					if k < len(runes) && classes[k] == SBLower {
						// SB8: ATerm Close* Sp Close* × Lower - don't break
						shouldBreak = false
					} else {
						// SB10 doesn't cover Close after Sp, so SB11 breaks
						shouldBreak = true
					}
				} else if curr == SBLower {
					// SB8: ATerm Close* Sp* × Lower (don't break before lowercase)
					shouldBreak = false
				} else if curr == SBSContinue || curr == SBSTerm || curr == SBATerm {
					// SB8a: (ATerm|STerm) Close* Sp* × (SContinue | STerm | ATerm)
					shouldBreak = false
				} else if (curr == SBSp || curr == SBSep || curr == SBCR || curr == SBLF) {
					// SB10: ATerm Close* Sp* × (Sp | Sep | CR | LF) - don't break
					shouldBreak = false
				} else if curr == SBClose && !hasSpBeforePrev {
					// SB9: ATerm Close* × Close (no Sp in between) - don't break
					shouldBreak = false
				} else {
					// SB11: Break after (ATerm|STerm) Close* Sp* before other characters
					shouldBreak = true
				}
			}
			// If no ATerm/STerm before Close*/Sp*, default (no break) applies
		} else if prev == SBSTerm {
			// Check STerm-related rules (SB8a, SB9, SB10, SB11)
			// SB9: STerm Close* × (Close | Sp | Sep | CR | LF)
			if curr == SBClose || curr == SBSp || curr == SBSep || curr == SBCR || curr == SBLF {
				shouldBreak = false
			} else {
				// Look forward through Close* Sp* for what follows
				j := i
				for j < len(runes) && (classes[j] == SBClose || classes[j] == SBSp || classes[j] == SBFormat || classes[j] == SBExtend) {
					j++
				}

				if j >= len(runes) {
					// SB11: STerm Close* Sp* <end of text>
					shouldBreak = true
				} else {
					next := classes[j]
					// SB8a: (STerm | ATerm) Close* Sp* × (SContinue | STerm | ATerm)
					if next == SBSContinue || next == SBSTerm || next == SBATerm {
						shouldBreak = false
					} else {
						// SB11: Break after STerm Close* Sp*
						shouldBreak = true
					}
				}
			}
		} else {
			// SB998: Don't break by default
			shouldBreak = false
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

	// SB2: Break at end
	breaks = append(breaks, len(text))

	return breaks
}

// Sentences splits text into sentences
func Sentences(text string) []string {
	breaks := FindSentenceBreaks(text)
	if len(breaks) <= 1 {
		return []string{}
	}

	result := make([]string, len(breaks)-1)
	for i := 0; i < len(breaks)-1; i++ {
		result[i] = text[breaks[i]:breaks[i+1]]
	}
	return result
}
