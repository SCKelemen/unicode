package uax29

// BreakOpportunities contains all break positions (byte offsets) for
// grapheme clusters, words, and sentences in a single unified structure.
//
// This enables efficient single-pass processing where the text is decoded
// and classified once, then all three break types are computed in a single
// traversal.
//
// The break positions are hierarchical:
//   - All word breaks occur at grapheme boundaries (Words ⊆ Graphemes)
//   - All sentence breaks occur at word boundaries (Sentences ⊆ Words)
//
// This natural hierarchy allows optimization: words are only checked at
// grapheme boundaries, and sentences only at word boundaries.
type BreakOpportunities struct {
	// Graphemes contains byte positions of all grapheme cluster boundaries.
	// This is the most granular level of text segmentation.
	Graphemes []int

	// Words contains byte positions of all word boundaries.
	// Every word break is also a grapheme break.
	Words []int

	// Sentences contains byte positions of all sentence boundaries.
	// Every sentence break is also a word break (and grapheme break).
	Sentences []int
}

// FindAllBreaks computes grapheme, word, and sentence boundaries in a single
// pass over the text. This is significantly more efficient than calling
// FindGraphemeBreaks, FindWordBreaks, and FindSentenceBreaks separately.
//
// Performance benefits:
//   - UTF-8 decoded once (not three times)
//   - Runes classified once (not three times)
//   - Hierarchical optimization: words checked only at grapheme boundaries,
//     sentences checked only at word boundaries
//
// Expected speedup: 3-5× faster than separate calls when all three break
// types are needed.
//
// Example:
//
//	text := "Hello, world! How are you?"
//	breaks := uax29.FindAllBreaks(text)
//
//	// Use grapheme breaks for cursor movement
//	for _, pos := range breaks.Graphemes {
//	    // ...
//	}
//
//	// Use word breaks for text selection
//	for _, pos := range breaks.Words {
//	    // ...
//	}
//
//	// Use sentence breaks for NLP
//	for _, pos := range breaks.Sentences {
//	    // ...
//	}
func FindAllBreaks(text string) BreakOpportunities {
	if text == "" {
		return BreakOpportunities{
			Graphemes: []int{},
			Words:     []int{},
			Sentences: []int{},
		}
	}

	// Single UTF-8 decode and classification pass
	runes := []rune(text)
	n := len(runes)

	// Classify all runes once using the unified packed data structure
	classes := make([]PackedBreakClass, n)
	for i, r := range runes {
		classes[i] = classifyRune(r)
	}

	// Find grapheme breaks (most granular)
	graphemeBreaks := findGraphemeBreaksFromClasses(text, runes, classes)

	// Find word breaks (only at grapheme boundaries)
	wordBreaks := findWordBreaksFromClasses(text, runes, classes, graphemeBreaks)

	// Find sentence breaks (only at word boundaries)
	sentenceBreaks := findSentenceBreaksFromClasses(text, runes, classes, wordBreaks)

	return BreakOpportunities{
		Graphemes: graphemeBreaks,
		Words:     wordBreaks,
		Sentences: sentenceBreaks,
	}
}

// Helper function to find grapheme breaks from pre-classified runes
func findGraphemeBreaksFromClasses(text string, runes []rune, classes []PackedBreakClass) []int {
	if len(runes) == 0 {
		return []int{}
	}

	breaks := []int{0} // GB1: Break at start

	for i := 1; i < len(runes); i++ {
		prev := classes[i-1].Grapheme()
		curr := classes[i].Grapheme()

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
			j := i - 1
			foundLinker := false
			// Skip back through Extend, ZWJ, and Linker characters
			for j >= 0 {
				if isIndicConjunctLinker(runes[j]) {
					foundLinker = true
					j--
					break
				}
				rClass := classes[j].Grapheme()
				if rClass == GBExtend || rClass == GBZWJ {
					j--
					continue
				}
				break
			}
			if foundLinker {
				// Continue looking back through Extend/Linker for a Consonant
				for j >= 0 {
					if isIndicConjunctConsonant(runes[j]) {
						shouldBreak = false
						break
					}
					if isIndicConjunctLinker(runes[j]) {
						j--
						continue
					}
					rClass := classes[j].Grapheme()
					if rClass == GBExtend || rClass == GBZWJ {
						j--
						continue
					}
					break
				}
			}
		} else if isExtendedPictographic(runes[i]) {
			// GB11: ExtendedPictographic Extend* ZWJ × ExtendedPictographic
			j := i - 1
			// Skip any Extend characters using pre-classified data
			for j >= 0 && classes[j].Grapheme() == GBExtend {
				j--
			}
			// Check if we have ZWJ
			if j >= 0 && classes[j].Grapheme() == GBZWJ {
				j--
				for j >= 0 && classes[j].Grapheme() == GBExtend {
					j--
				}
				if j >= 0 && isExtendedPictographic(runes[j]) {
					shouldBreak = false
				}
			}
		} else if prev == GBRegionalIndicator && curr == GBRegionalIndicator {
			// GB12/GB13: Regional Indicator pairs
			riCountBefore := 0
			for j := i - 1; j >= 0 && classes[j].Grapheme() == GBRegionalIndicator; j-- {
				riCountBefore++
			}
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

// Helper function to find word breaks from pre-classified runes at grapheme boundaries
func findWordBreaksFromClasses(text string, runes []rune, classes []PackedBreakClass, graphemeBreaks []int) []int {
	if len(runes) == 0 {
		return []int{}
	}

	// Convert grapheme break byte positions to rune indices
	graphemeRuneIndices := make([]int, len(graphemeBreaks))
	byteToRune := 0
	runeIdx := 0
	for i, bytePos := range graphemeBreaks {
		for byteToRune < bytePos && runeIdx < len(runes) {
			byteToRune += len(string(runes[runeIdx]))
			runeIdx++
		}
		graphemeRuneIndices[i] = runeIdx
	}

	// Extract word classes from packed data
	wordClasses := make([]WordBreakClass, len(runes))
	for i := range runes {
		wordClasses[i] = classes[i].Word()
	}

	breaks := []int{0} // WB1: Break at start

	// Only check word breaks at grapheme boundaries
	for gi := 1; gi < len(graphemeRuneIndices)-1; gi++ {
		i := graphemeRuneIndices[gi]
		if i >= len(runes) {
			continue
		}

		// Skip Format and Extend for most rules (WB4)
		prevIdx := i - 1
		for prevIdx > 0 && (wordClasses[prevIdx] == WBFormat || wordClasses[prevIdx] == WBExtend || wordClasses[prevIdx] == WBZWJ) {
			prevIdx--
		}

		prev := wordClasses[prevIdx]
		curr := wordClasses[i]

		shouldBreak := true

		// WB3: Don't break within CRLF
		if wordClasses[i-1] == WBCR && curr == WBLF {
			shouldBreak = false
		} else if wordClasses[i-1] == WBCR || wordClasses[i-1] == WBLF || wordClasses[i-1] == WBNewline {
			// WB3a: Break after newlines
			shouldBreak = true
		} else if curr == WBCR || curr == WBLF || curr == WBNewline {
			// WB3b: Break before newlines
			shouldBreak = true
		} else if wordClasses[i-1] == WBZWJ && isExtendedPictographic(runes[i]) {
			// WB3c: Don't break within emoji ZWJ sequences
			shouldBreak = false
		} else if wordClasses[i-1] == WBWSegSpace && curr == WBWSegSpace {
			// WB3d: Keep horizontal whitespace together
			shouldBreak = false
		} else if curr == WBFormat || curr == WBExtend || curr == WBZWJ {
			// WB4: Ignore Format and Extend
			shouldBreak = false
		} else if (prev == WBALetter || prev == WBHebrewLetter) && (curr == WBALetter || curr == WBHebrewLetter) {
			// WB5: Don't break between letters
			shouldBreak = false
		} else if prev == WBHebrewLetter && curr == WBSingleQuote {
			// WB7a: Hebrew_Letter × Single_Quote
			shouldBreak = false
		} else if prev == WBHebrewLetter && curr == WBDoubleQuote {
			// WB7b: Hebrew_Letter × Double_Quote Hebrew_Letter
			nextIdx := i + 1
			for nextIdx < len(runes) && (wordClasses[nextIdx] == WBFormat || wordClasses[nextIdx] == WBExtend || wordClasses[nextIdx] == WBZWJ) {
				nextIdx++
			}
			if nextIdx < len(runes) && wordClasses[nextIdx] == WBHebrewLetter {
				shouldBreak = false
			}
		} else if prev == WBDoubleQuote && curr == WBHebrewLetter {
			// WB7c: Hebrew_Letter Double_Quote × Hebrew_Letter
			prevPrevIdx := prevIdx - 1
			for prevPrevIdx >= 0 && (wordClasses[prevPrevIdx] == WBFormat || wordClasses[prevPrevIdx] == WBExtend || wordClasses[prevPrevIdx] == WBZWJ) {
				prevPrevIdx--
			}
			if prevPrevIdx >= 0 && wordClasses[prevPrevIdx] == WBHebrewLetter {
				shouldBreak = false
			}
		} else if (prev == WBALetter || prev == WBHebrewLetter) && (curr == WBMidLetter || curr == WBMidNumLet || curr == WBSingleQuote) {
			// WB6: Check for AHLetter × (MidLetter | MidNumLet | Single_Quote) AHLetter
			nextIdx := i + 1
			for nextIdx < len(runes) && (wordClasses[nextIdx] == WBFormat || wordClasses[nextIdx] == WBExtend || wordClasses[nextIdx] == WBZWJ) {
				nextIdx++
			}
			if nextIdx < len(runes) && (wordClasses[nextIdx] == WBALetter || wordClasses[nextIdx] == WBHebrewLetter) {
				shouldBreak = false
			}
		} else if (prev == WBMidLetter || prev == WBMidNumLet || prev == WBSingleQuote) && (curr == WBALetter || curr == WBHebrewLetter) {
			// WB7: Check for AHLetter (MidLetter | MidNumLet | Single_Quote) × AHLetter
			prevPrevIdx := prevIdx - 1
			for prevPrevIdx >= 0 && (wordClasses[prevPrevIdx] == WBFormat || wordClasses[prevPrevIdx] == WBExtend || wordClasses[prevPrevIdx] == WBZWJ) {
				prevPrevIdx--
			}
			if prevPrevIdx >= 0 && (wordClasses[prevPrevIdx] == WBALetter || wordClasses[prevPrevIdx] == WBHebrewLetter) {
				shouldBreak = false
			}
		} else if prev == WBNumeric && curr == WBNumeric {
			// WB8: Don't break within sequences of digits
			shouldBreak = false
		} else if (prev == WBALetter || prev == WBHebrewLetter) && curr == WBNumeric {
			// WB9: AHLetter × Numeric
			shouldBreak = false
		} else if prev == WBNumeric && (curr == WBALetter || curr == WBHebrewLetter) {
			// WB10: Numeric × AHLetter
			shouldBreak = false
		} else if prev == WBNumeric && (curr == WBMidNum || curr == WBMidNumLet || curr == WBSingleQuote) {
			// WB11: Check for Numeric × (MidNum | MidNumLet | Single_Quote) Numeric
			nextIdx := i + 1
			for nextIdx < len(runes) && (wordClasses[nextIdx] == WBFormat || wordClasses[nextIdx] == WBExtend || wordClasses[nextIdx] == WBZWJ) {
				nextIdx++
			}
			if nextIdx < len(runes) && wordClasses[nextIdx] == WBNumeric {
				shouldBreak = false
			}
		} else if (prev == WBMidNum || prev == WBMidNumLet || prev == WBSingleQuote) && curr == WBNumeric {
			// WB12: Check for Numeric (MidNum | MidNumLet | Single_Quote) × Numeric
			prevPrevIdx := prevIdx - 1
			for prevPrevIdx >= 0 && (wordClasses[prevPrevIdx] == WBFormat || wordClasses[prevPrevIdx] == WBExtend || wordClasses[prevPrevIdx] == WBZWJ) {
				prevPrevIdx--
			}
			if prevPrevIdx >= 0 && wordClasses[prevPrevIdx] == WBNumeric {
				shouldBreak = false
			}
		} else if prev == WBKatakana && curr == WBKatakana {
			// WB13: Don't break between Katakana
			shouldBreak = false
		} else if (prev == WBALetter || prev == WBHebrewLetter || prev == WBNumeric || prev == WBKatakana || prev == WBExtendNumLet) && curr == WBExtendNumLet {
			// WB13a: (AHLetter | Numeric | Katakana | ExtendNumLet) × ExtendNumLet
			shouldBreak = false
		} else if prev == WBExtendNumLet && (curr == WBALetter || curr == WBHebrewLetter || curr == WBNumeric || curr == WBKatakana) {
			// WB13b: ExtendNumLet × (AHLetter | Numeric | Katakana)
			shouldBreak = false
		} else if prev == WBRegionalIndicator && curr == WBRegionalIndicator {
			// WB15/16: Regional Indicator pairs
			count := 0
			j := prevIdx
			for j >= 0 {
				if wordClasses[j] == WBRegionalIndicator {
					count++
					j--
				} else if wordClasses[j] == WBFormat || wordClasses[j] == WBExtend || wordClasses[j] == WBZWJ {
					j--
				} else {
					break
				}
			}
			if count%2 == 1 {
				shouldBreak = false
			}
		}

		if shouldBreak {
			breaks = append(breaks, graphemeBreaks[gi])
		}
	}

	// WB2: Break at end
	breaks = append(breaks, len(text))

	return breaks
}

// Helper function to find sentence breaks from pre-classified runes at word boundaries
func findSentenceBreaksFromClasses(text string, runes []rune, classes []PackedBreakClass, wordBreaks []int) []int {
	if len(runes) == 0 {
		return []int{}
	}

	// Convert word break byte positions to rune indices
	wordRuneIndices := make([]int, len(wordBreaks))
	byteToRune := 0
	runeIdx := 0
	for i, bytePos := range wordBreaks {
		for byteToRune < bytePos && runeIdx < len(runes) {
			byteToRune += len(string(runes[runeIdx]))
			runeIdx++
		}
		wordRuneIndices[i] = runeIdx
	}

	// Extract sentence classes from packed data
	sentClasses := make([]SentenceBreakClass, len(runes))
	for i := range runes {
		sentClasses[i] = classes[i].Sentence()
	}

	breaks := []int{0} // SB1: Break at start

	// Only check sentence breaks at word boundaries
	for wi := 1; wi < len(wordRuneIndices)-1; wi++ {
		i := wordRuneIndices[wi]
		if i >= len(runes) {
			continue
		}

		// Get previous non-Format/Extend character for most rules (SB5)
		prevIdx := i - 1
		for prevIdx > 0 && (sentClasses[prevIdx] == SBFormat || sentClasses[prevIdx] == SBExtend) {
			prevIdx--
		}

		prev := sentClasses[prevIdx]
		curr := sentClasses[i]

		shouldBreak := false // SB998: Default is no break

		// SB3: Don't break within CRLF
		if sentClasses[i-1] == SBCR && curr == SBLF {
			shouldBreak = false
		} else if sentClasses[i-1] == SBCR || sentClasses[i-1] == SBLF || sentClasses[i-1] == SBSep {
			// SB4: Break after paragraph separators
			shouldBreak = true
		} else if curr == SBFormat || curr == SBExtend {
			// SB5: Ignore Format and Extend
			shouldBreak = false
		} else if prev == SBATerm && curr == SBNumeric {
			// SB6: ATerm × Numeric
			shouldBreak = false
		} else if (prev == SBUpper || prev == SBLower) && curr == SBATerm {
			// SB7: (Upper | Lower) ATerm × Upper
			nextIdx := i + 1
			for nextIdx < len(runes) && (sentClasses[nextIdx] == SBFormat || sentClasses[nextIdx] == SBExtend) {
				nextIdx++
			}
			if nextIdx < len(runes) && sentClasses[nextIdx] == SBUpper {
				shouldBreak = false
			}
		} else if prev == SBATerm {
			// Check ATerm-related rules (SB7, SB8, SB8a, SB9, SB10, SB11)

			// SB9: ATerm Close* × (Close | Sp | Sep | CR | LF)
			if curr == SBClose || curr == SBSp || curr == SBSep || curr == SBCR || curr == SBLF {
				shouldBreak = false
			} else if curr == SBUpper {
				// SB7: (Upper | Lower) ATerm × Upper
				prevPrevIdx := prevIdx - 1
				for prevPrevIdx > 0 && (sentClasses[prevPrevIdx] == SBFormat || sentClasses[prevPrevIdx] == SBExtend) {
					prevPrevIdx--
				}
				if prevPrevIdx >= 0 && (sentClasses[prevPrevIdx] == SBUpper || sentClasses[prevPrevIdx] == SBLower) {
					shouldBreak = false
				} else {
					shouldBreak = true
				}
			} else {
				// Look forward through Close* Sp* for what follows
				j := i
				for j < len(runes) && (sentClasses[j] == SBClose || sentClasses[j] == SBSp || sentClasses[j] == SBFormat || sentClasses[j] == SBExtend) {
					j++
				}

				if j >= len(runes) {
					// SB11: ATerm Close* Sp* <end of text>
					shouldBreak = true
				} else {
					next := sentClasses[j]
					// SB8a: (STerm | ATerm) Close* Sp* × (SContinue | STerm | ATerm)
					if next == SBSContinue || next == SBSTerm || next == SBATerm {
						shouldBreak = false
					} else if next == SBLower {
						// SB8: ATerm Close* Sp* × Lower
						shouldBreak = false
					} else {
						// SB11: Break after ATerm Close* Sp*
						shouldBreak = true
					}
				}
			}
		} else if prev == SBClose || prev == SBSp {
			// When prev is Close or Sp, check if there's ATerm/STerm before it
			hasSpBeforePrev := false
			j := prevIdx - 1
			for j >= 0 && (sentClasses[j] == SBClose || sentClasses[j] == SBSp || sentClasses[j] == SBFormat || sentClasses[j] == SBExtend) {
				if sentClasses[j] == SBSp {
					hasSpBeforePrev = true
				}
				j--
			}
			if j >= 0 && (sentClasses[j] == SBATerm || sentClasses[j] == SBSTerm) {
				if prev == SBClose && hasSpBeforePrev {
					shouldBreak = false
				} else if prev == SBSp && curr == SBClose {
					k := i + 1
					for k < len(runes) && (sentClasses[k] == SBClose || sentClasses[k] == SBFormat || sentClasses[k] == SBExtend) {
						k++
					}
					if k < len(runes) && sentClasses[k] == SBLower {
						shouldBreak = false
					} else {
						shouldBreak = true
					}
				} else if curr == SBLower {
					// SB8: ATerm Close* Sp* × Lower
					shouldBreak = false
				} else if curr == SBSContinue || curr == SBSTerm || curr == SBATerm {
					// SB8a
					shouldBreak = false
				} else if (curr == SBSp || curr == SBSep || curr == SBCR || curr == SBLF) {
					// SB10
					shouldBreak = false
				} else if curr == SBClose && !hasSpBeforePrev {
					// SB9
					shouldBreak = false
				} else {
					// SB11
					shouldBreak = true
				}
			}
		} else if prev == SBSTerm {
			// Check STerm-related rules (SB8a, SB9, SB10, SB11)
			if curr == SBClose || curr == SBSp || curr == SBSep || curr == SBCR || curr == SBLF {
				shouldBreak = false
			} else {
				j := i
				for j < len(runes) && (sentClasses[j] == SBClose || sentClasses[j] == SBSp || sentClasses[j] == SBFormat || sentClasses[j] == SBExtend) {
					j++
				}

				if j >= len(runes) {
					// SB11: STerm Close* Sp* <end of text>
					shouldBreak = true
				} else {
					next := sentClasses[j]
					// SB8a
					if next == SBSContinue || next == SBSTerm || next == SBATerm {
						shouldBreak = false
					} else {
						// SB11
						shouldBreak = true
					}
				}
			}
		} else {
			// SB998: Don't break by default
			shouldBreak = false
		}

		if shouldBreak {
			breaks = append(breaks, wordBreaks[wi])
		}
	}

	// SB2: Break at end
	breaks = append(breaks, len(text))

	return breaks
}
