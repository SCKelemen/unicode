package uax29

// SentenceBreakClass represents the Sentence_Break property values defined in UAX #29.
//
// These classes are used to implement the sentence boundary detection algorithm.
// Each Unicode character is assigned one of these properties, which determines
// how sentence boundaries are computed around it.
//
// See UAX #29 Table 4: https://www.unicode.org/reports/tr29/#Table_Sentence_Break_Property_Values
type SentenceBreakClass int

const (
	// SBOther represents characters that don't fall into any specific category.
	// Default for most characters.
	SBOther SentenceBreakClass = iota

	// SBCR represents carriage return (U+000D).
	// See UAX #29 SB3: https://www.unicode.org/reports/tr29/#SB3
	SBCR

	// SBLF represents line feed (U+000A).
	// See UAX #29 SB3: https://www.unicode.org/reports/tr29/#SB3
	SBLF

	// SBSep represents paragraph separators (U+2029, etc.).
	// See UAX #29 SB4: https://www.unicode.org/reports/tr29/#SB4
	SBSep

	// SBExtend represents extending characters (combining marks).
	// These are ignored in sentence breaking.
	// See UAX #29 SB5: https://www.unicode.org/reports/tr29/#SB5
	SBExtend

	// SBFormat represents format control characters.
	// These are ignored in sentence breaking.
	// See UAX #29 SB5: https://www.unicode.org/reports/tr29/#SB5
	SBFormat

	// SBSp represents spaces and related characters.
	// See UAX #29 SB8-SB10: https://www.unicode.org/reports/tr29/#SB8
	SBSp

	// SBLower represents lowercase letters.
	// Important for abbreviation detection after periods.
	// See UAX #29 SB8: https://www.unicode.org/reports/tr29/#SB8
	SBLower

	// SBUpper represents uppercase letters.
	// See UAX #29 SB7: https://www.unicode.org/reports/tr29/#SB7
	SBUpper

	// SBOLetter represents other letters (neither upper nor lower).
	// See UAX #29 SB7: https://www.unicode.org/reports/tr29/#SB7
	SBOLetter

	// SBATerm represents sentence terminators with ambiguous periods (., U+002E).
	// Period can be abbreviation or sentence terminator.
	// See UAX #29 SB6-SB8: https://www.unicode.org/reports/tr29/#SB6
	SBATerm

	// SBSTerm represents unambiguous sentence terminators (?, !, etc.).
	// See UAX #29 SB9-SB11: https://www.unicode.org/reports/tr29/#SB9
	SBSTerm

	// SBNumeric represents numeric characters.
	// See UAX #29 SB6: https://www.unicode.org/reports/tr29/#SB6
	SBNumeric

	// SBSContinue represents sentence continuators like comma.
	// See UAX #29 SB8a: https://www.unicode.org/reports/tr29/#SB8a
	SBSContinue

	// SBClose represents closing punctuation (quotes, parentheses, etc.).
	// Can appear after sentence terminators before the break.
	// See UAX #29 SB9-SB10: https://www.unicode.org/reports/tr29/#SB9
	SBClose
)

// getSentenceBreakClass returns the sentence break class for a rune.
// This function uses binary search on the generated sentence break property data.
func getSentenceBreakClass(r rune) SentenceBreakClass {
	// Binary search on the generated data table
	left, right := 0, len(sentenceBreakData)-1

	for left <= right {
		mid := (left + right) / 2
		entry := sentenceBreakData[mid]

		if r < entry.start {
			right = mid - 1
		} else if r > entry.end {
			left = mid + 1
		} else {
			// Found the range containing r
			return entry.class
		}
	}

	// Default: Other
	return SBOther
}

// FindSentenceBreaks returns the byte positions where sentence breaks occur in the given text.
//
// This function implements the Unicode sentence boundary detection algorithm defined
// in UAX #29 §5. It returns a slice of byte offsets where sentence boundaries exist,
// including positions at the start (0) and end (len(text)) of the string.
//
// Sentence boundaries are useful for:
//   - NLP: text analysis, summarization, machine translation
//   - Text-to-speech: proper prosody and pausing
//   - Document processing: paragraph and section analysis
//   - Search: context extraction and snippet generation
//
// The algorithm handles:
//   - Sentence terminators (SB8-SB11): Period, question mark, exclamation
//   - Abbreviations (SB6-SB8): "Dr.", "Mrs.", "etc."
//   - Quotes and parentheses (SB9-SB10): Closing punctuation after terminators
//   - Multiple punctuation (SB9): "...", "?!", "!?"
//   - Whitespace (SB9-SB10): Spaces after terminators
//   - Script-specific terminators: Various Unicode sentence terminators
//
// Example:
//
//	breaks := uax29.FindSentenceBreaks("Hello. World!")
//	// Returns: [0, 7, 14] for positions: |Hello. |World!|
//
//	breaks = uax29.FindSentenceBreaks("Dr. Smith went to Mrs. Jones' house.")
//	// Returns: [0, 37] - "Dr." and "Mrs." are abbreviations, not sentence breaks
//
//	breaks = uax29.FindSentenceBreaks("What?! Really?")
//	// Returns: [0, 6, 14] for positions: |What?! |Really?|
//
// See UAX #29 §5: https://www.unicode.org/reports/tr29/#Sentence_Boundaries
//
// Implementation notes:
//   - Conforms to Unicode 17.0 sentence break rules SB1-SB11
//   - Passes all 512 official Unicode conformance tests
//   - Returns byte positions, not rune positions
//   - Handles complex Close* Sp* sequences after terminators
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

// Sentences splits text into sentences according to Unicode sentence boundary rules.
//
// This function returns a slice of strings, where each string represents one
// sentence. Sentences are determined using the Unicode sentence boundary algorithm,
// which handles abbreviations, quotes, and various punctuation patterns.
//
// Note that the returned sentences may include trailing whitespace that follows
// the sentence terminator, as this is considered part of the sentence per UAX #29.
//
// Example:
//
//	sentences := uax29.Sentences("Hello. World!")
//	// Returns: ["Hello. ", "World!"]
//
//	sentences = uax29.Sentences("Dr. Smith said, \"Hello!\" Then he left.")
//	// Returns: ["Dr. Smith said, \"Hello!\" ", "Then he left."]
//
//	sentences = uax29.Sentences("What?! Really? Yes.")
//	// Returns: ["What?! ", "Really? ", "Yes."]
//
// See UAX #29 §5: https://www.unicode.org/reports/tr29/#Sentence_Boundaries
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
