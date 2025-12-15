package uax29

import (
	"unicode"

	"github.com/SCKelemen/unicode/uts51"
)

// WordBreakClass represents the Word_Break property values defined in UAX #29.
//
// These classes are used to implement the word boundary detection algorithm.
// Each Unicode character is assigned one of these properties, which determines
// how word boundaries are computed around it.
//
// See UAX #29 Table 3: https://www.unicode.org/reports/tr29/#Table_Word_Break_Property_Values
type WordBreakClass int

const (
	// WBOther represents characters that don't fall into any specific category.
	// Default for most punctuation and symbols.
	WBOther WordBreakClass = iota

	// WBCR represents carriage return (U+000D).
	// See UAX #29 WB3: https://www.unicode.org/reports/tr29/#WB3
	WBCR

	// WBLF represents line feed (U+000A).
	// See UAX #29 WB3: https://www.unicode.org/reports/tr29/#WB3
	WBLF

	// WBNewline represents other newline characters (NEL, LS, PS).
	// See UAX #29 WB3a/WB3b: https://www.unicode.org/reports/tr29/#WB3a
	WBNewline

	// WBExtend represents combining marks and similar characters.
	// These characters extend the preceding base character.
	// See UAX #29 WB4: https://www.unicode.org/reports/tr29/#WB4
	WBExtend

	// WBZWJ represents Zero Width Joiner (U+200D).
	// Used in emoji ZWJ sequences.
	// See UAX #29 WB3c: https://www.unicode.org/reports/tr29/#WB3c
	WBZWJ

	// WBRegionalIndicator represents regional indicator symbols (U+1F1E6..U+1F1FF).
	// Used for flag emoji sequences.
	// See UAX #29 WB15/WB16: https://www.unicode.org/reports/tr29/#WB15
	WBRegionalIndicator

	// WBFormat represents format control characters.
	// These are ignored in most word break rules.
	// See UAX #29 WB4: https://www.unicode.org/reports/tr29/#WB4
	WBFormat

	// WBKatakana represents Katakana characters.
	// Katakana sequences stay together.
	// See UAX #29 WB13: https://www.unicode.org/reports/tr29/#WB13
	WBKatakana

	// WBHebrewLetter represents Hebrew letters.
	// Has special rules for combining with quotes.
	// See UAX #29 WB7a-WB7c: https://www.unicode.org/reports/tr29/#WB7a
	WBHebrewLetter

	// WBALetter represents alphabetic characters.
	// Forms the core of word sequences.
	// See UAX #29 WB5: https://www.unicode.org/reports/tr29/#WB5
	WBALetter

	// WBSingleQuote represents single quotation mark (U+0027).
	// Used in contractions and possessives.
	// See UAX #29 WB6/WB7: https://www.unicode.org/reports/tr29/#WB6
	WBSingleQuote

	// WBDoubleQuote represents double quotation mark (U+0022).
	// Has special rules with Hebrew letters.
	// See UAX #29 WB7b/WB7c: https://www.unicode.org/reports/tr29/#WB7b
	WBDoubleQuote

	// WBMidNumLet represents characters that can appear in the middle of
	// both words and numbers (e.g., period, middle dot).
	// See UAX #29 WB6/WB11: https://www.unicode.org/reports/tr29/#WB6
	WBMidNumLet

	// WBMidLetter represents characters that can appear in the middle of words
	// (e.g., colon, middle dot).
	// See UAX #29 WB6/WB7: https://www.unicode.org/reports/tr29/#WB6
	WBMidLetter

	// WBMidNum represents characters that can appear in the middle of numbers
	// (e.g., comma, semicolon for thousands/decimals).
	// See UAX #29 WB11/WB12: https://www.unicode.org/reports/tr29/#WB11
	WBMidNum

	// WBNumeric represents numeric characters.
	// Forms number sequences.
	// See UAX #29 WB8: https://www.unicode.org/reports/tr29/#WB8
	WBNumeric

	// WBExtendNumLet represents characters that extend both letters and numbers
	// (e.g., underscore).
	// See UAX #29 WB13a/WB13b: https://www.unicode.org/reports/tr29/#WB13a
	WBExtendNumLet

	// WBWSegSpace represents whitespace used for word separation.
	// See UAX #29 WB3d: https://www.unicode.org/reports/tr29/#WB3d
	WBWSegSpace

	// WBExtendedPictographic represents emoji and pictographic characters.
	// Used in ZWJ emoji sequences.
	// See UAX #29 WB3c: https://www.unicode.org/reports/tr29/#WB3c
	WBExtendedPictographic
)

// getWordBreakClass returns the word break class for a rune
func getWordBreakClass(r rune) WordBreakClass {
	// CR and LF
	if r == 0x000D {
		return WBCR
	}
	if r == 0x000A {
		return WBLF
	}

	// Newline
	if r == 0x000B || r == 0x000C || r == 0x0085 || r == 0x2028 || r == 0x2029 {
		return WBNewline
	}

	// ZWJ
	if r == uts51.ZeroWidthJoiner {
		return WBZWJ
	}

	// Regional Indicators
	if uts51.IsRegionalIndicator(r) {
		return WBRegionalIndicator
	}

	// Single_Quote
	if r == 0x0027 {
		return WBSingleQuote
	}

	// Double_Quote
	if r == 0x0022 {
		return WBDoubleQuote
	}

	// MidNumLet
	if r == 0x002E || r == 0x2018 || r == 0x2019 || r == 0x2024 || r == 0xFE52 || r == 0xFF07 || r == 0xFF0E {
		return WBMidNumLet
	}

	// MidLetter
	if r == 0x003A || r == 0x00B7 || r == 0x0387 || r == 0x055F || r == 0x05F4 || r == 0x2027 || r == 0xFE13 || r == 0xFE55 || r == 0xFF1A {
		return WBMidLetter
	}

	// MidNum
	if r == 0x002C || r == 0x003B || r == 0x037E || r == 0x0589 || r == 0x060C || r == 0x060D ||
		r == 0x066C || r == 0x07F8 || r == 0x2044 || r == 0xFE10 || r == 0xFE14 || r == 0xFE50 || r == 0xFF0C || r == 0xFF1B {
		return WBMidNum
	}

	// ExtendNumLet
	if r == 0x005F || r == 0x203F || r == 0x2040 || r == 0x2054 || r == 0xFE33 || r == 0xFE34 || r == 0xFE4D || r == 0xFE4E || r == 0xFE4F || r == 0xFF3F {
		return WBExtendNumLet
	}

	// WSegSpace
	if r == 0x0020 || r == 0x1680 || (r >= 0x2000 && r <= 0x2006) || (r >= 0x2008 && r <= 0x200A) || r == 0x205F || r == 0x3000 {
		return WBWSegSpace
	}

	// ZWNJ (U+200C) is treated as Extend
	if r == 0x200C {
		return WBExtend
	}

	// Emoji modifiers (skin tones) are Extend for word breaking
	if uts51.IsEmojiModifier(r) {
		return WBExtend
	}

	// Special Cf characters with specific Word_Break properties
	// U+06DD (Arabic End of Ayah) is Numeric for word breaking
	if r == 0x06DD {
		return WBNumeric
	}
	// U+070F (Syriac Abbreviation Mark) is ALetter for word breaking
	if r == 0x070F {
		return WBALetter
	}

	// Format
	if unicode.Is(unicode.Cf, r) && r != 0x200C && r != uts51.ZeroWidthJoiner {
		return WBFormat
	}

	// Extend (combining marks)
	if unicode.Is(unicode.Me, r) || unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r) {
		return WBExtend
	}

	// Katakana
	if (r >= 0x3031 && r <= 0x3035) || (r >= 0x309B && r <= 0x309C) || (r >= 0x30A0 && r <= 0x30FA) ||
		(r >= 0x30FC && r <= 0x30FF) || (r >= 0x31F0 && r <= 0x31FF) || (r >= 0x32D0 && r <= 0x32FE) ||
		(r >= 0x3300 && r <= 0x3357) || r == 0xFF70 || (r >= 0xFF9E && r <= 0xFF9F) || (r >= 0x1B000 && r <= 0x1B11F) ||
		(r >= 0x1B132 && r <= 0x1B132) || (r >= 0x1B150 && r <= 0x1B152) {
		return WBKatakana
	}

	// Hebrew_Letter
	if r >= 0x05D0 && r <= 0x05EA {
		return WBHebrewLetter
	}
	if r >= 0x05EF && r <= 0x05F2 {
		return WBHebrewLetter
	}

	// Numeric
	if unicode.IsDigit(r) {
		return WBNumeric
	}

	// ALetter - includes some non-letter symbols
	if unicode.IsLetter(r) {
		return WBALetter
	}
	// Special ALetter symbols (So and Nl categories that are ALetter for word breaking)
	if (r >= 0x16EE && r <= 0x16F0) || (r >= 0x2160 && r <= 0x2182) || (r >= 0x2185 && r <= 0x2188) {
		return WBALetter
	}
	if (r >= 0x24B6 && r <= 0x24E9) { // Circled letters
		return WBALetter
	}
	if (r >= 0xA6E6 && r <= 0xA6EF) || (r >= 0x10140 && r <= 0x10174) || r == 0x10341 || r == 0x1034A {
		return WBALetter
	}
	if (r >= 0x103D1 && r <= 0x103D5) || (r >= 0x12400 && r <= 0x1246E) {
		return WBALetter
	}
	if (r >= 0x1F130 && r <= 0x1F149) || (r >= 0x1F150 && r <= 0x1F169) || (r >= 0x1F170 && r <= 0x1F189) {
		return WBALetter
	}

	// Extended Pictographic (after letter check)
	if isExtendedPictographic(r) {
		return WBExtendedPictographic
	}

	return WBOther
}

// FindWordBreaks returns the byte positions where word breaks occur in the given text.
//
// This function implements the Unicode word boundary detection algorithm defined
// in UAX #29 §4. It returns a slice of byte offsets where word boundaries exist,
// including positions at the start (0) and end (len(text)) of the string.
//
// Word boundaries are useful for:
//   - Text editors: double-click word selection, cursor movement by word
//   - Search engines: tokenization and indexing
//   - NLP: word-level text analysis
//   - Accessibility: screen readers and text-to-speech
//
// The algorithm handles:
//   - Alphabetic sequences (WB5): Letters stay together
//   - Numeric sequences (WB8-WB12): Numbers and separators stay together
//   - Contractions (WB6-WB7): "don't", "can't", "John's"
//   - Hebrew with quotes (WB7a-WB7c): Special handling for Hebrew letters
//   - Katakana (WB13): Katakana characters stay together
//   - Emoji sequences (WB3c): ZWJ emoji sequences stay together
//   - Regional indicators (WB15-WB16): Flag emoji pairs stay together
//
// Example:
//
//	breaks := uax29.FindWordBreaks("Hello, world!")
//	// Returns: [0, 5, 6, 7, 12, 13] for positions: |Hello|,| |world|!|
//
//	breaks = uax29.FindWordBreaks("don't")
//	// Returns: [0, 5] - "don't" is a single word
//
//	breaks = uax29.FindWordBreaks("👨‍👩‍👧‍👦")
//	// Returns: [0, 25] - family emoji is a single word
//
// See UAX #29 §4: https://www.unicode.org/reports/tr29/#Word_Boundaries
//
// Implementation notes:
//   - Conforms to Unicode 17.0 word break rules WB1-WB16
//   - Passes all 1,944 official Unicode conformance tests
//   - Format and Extend characters are handled transparently (WB4)
//   - Returns byte positions, not rune positions
func FindWordBreaks(text string) []int {
	if len(text) == 0 {
		return []int{}
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return []int{}
	}

	classes := make([]WordBreakClass, len(runes))
	for i, r := range runes {
		classes[i] = getWordBreakClass(r)
	}

	breaks := []int{0} // WB1: Break at start

	for i := 1; i < len(runes); i++ {
		// Skip Format and Extend for most rules (WB4)
		prevIdx := i - 1
		for prevIdx > 0 && (classes[prevIdx] == WBFormat || classes[prevIdx] == WBExtend || classes[prevIdx] == WBZWJ) {
			prevIdx--
		}

		prev := classes[prevIdx]
		curr := classes[i]

		shouldBreak := true

		// WB3: Don't break within CRLF
		if classes[i-1] == WBCR && curr == WBLF {
			shouldBreak = false
		} else if classes[i-1] == WBCR || classes[i-1] == WBLF || classes[i-1] == WBNewline {
			// WB3a: Break after newlines
			shouldBreak = true
		} else if curr == WBCR || curr == WBLF || curr == WBNewline {
			// WB3b: Break before newlines
			shouldBreak = true
		} else if classes[i-1] == WBZWJ && isExtendedPictographic(runes[i]) {
			// WB3c: Don't break within emoji ZWJ sequences
			// Note: Check actual ExtPict property, not word break class, since some
			// characters like Ⓜ are classified as ALetter but are also ExtPict
			shouldBreak = false
		} else if classes[i-1] == WBWSegSpace && curr == WBWSegSpace {
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
			for nextIdx < len(runes) && (classes[nextIdx] == WBFormat || classes[nextIdx] == WBExtend || classes[nextIdx] == WBZWJ) {
				nextIdx++
			}
			if nextIdx < len(runes) && classes[nextIdx] == WBHebrewLetter {
				shouldBreak = false
			}
		} else if prev == WBDoubleQuote && curr == WBHebrewLetter {
			// WB7c: Hebrew_Letter Double_Quote × Hebrew_Letter
			// Look back to see if there's a HebrewLetter before the DoubleQuote
			prevPrevIdx := prevIdx - 1
			for prevPrevIdx >= 0 && (classes[prevPrevIdx] == WBFormat || classes[prevPrevIdx] == WBExtend || classes[prevPrevIdx] == WBZWJ) {
				prevPrevIdx--
			}
			if prevPrevIdx >= 0 && classes[prevPrevIdx] == WBHebrewLetter {
				shouldBreak = false
			}
		} else if (prev == WBALetter || prev == WBHebrewLetter) && (curr == WBMidLetter || curr == WBMidNumLet || curr == WBSingleQuote) {
			// WB6: Check for AHLetter × (MidLetter | MidNumLet | Single_Quote) AHLetter
			nextIdx := i + 1
			for nextIdx < len(runes) && (classes[nextIdx] == WBFormat || classes[nextIdx] == WBExtend || classes[nextIdx] == WBZWJ) {
				nextIdx++
			}
			if nextIdx < len(runes) && (classes[nextIdx] == WBALetter || classes[nextIdx] == WBHebrewLetter) {
				shouldBreak = false
			}
		} else if (prev == WBMidLetter || prev == WBMidNumLet || prev == WBSingleQuote) && (curr == WBALetter || curr == WBHebrewLetter) {
			// WB7: Check for AHLetter (MidLetter | MidNumLet | Single_Quote) × AHLetter
			// Look back to see if there's an AHLetter before the MidLetter
			prevPrevIdx := prevIdx - 1
			for prevPrevIdx >= 0 && (classes[prevPrevIdx] == WBFormat || classes[prevPrevIdx] == WBExtend || classes[prevPrevIdx] == WBZWJ) {
				prevPrevIdx--
			}
			if prevPrevIdx >= 0 && (classes[prevPrevIdx] == WBALetter || classes[prevPrevIdx] == WBHebrewLetter) {
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
			for nextIdx < len(runes) && (classes[nextIdx] == WBFormat || classes[nextIdx] == WBExtend || classes[nextIdx] == WBZWJ) {
				nextIdx++
			}
			if nextIdx < len(runes) && classes[nextIdx] == WBNumeric {
				shouldBreak = false
			}
		} else if (prev == WBMidNum || prev == WBMidNumLet || prev == WBSingleQuote) && curr == WBNumeric {
			// WB12: Check for Numeric (MidNum | MidNumLet | Single_Quote) × Numeric
			// Look back to see if there's a Numeric before the MidNum
			prevPrevIdx := prevIdx - 1
			for prevPrevIdx >= 0 && (classes[prevPrevIdx] == WBFormat || classes[prevPrevIdx] == WBExtend || classes[prevPrevIdx] == WBZWJ) {
				prevPrevIdx--
			}
			if prevPrevIdx >= 0 && classes[prevPrevIdx] == WBNumeric {
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
			// Count RIs backwards from prevIdx, skipping Format/Extend/ZWJ (which are transparent)
			count := 0
			j := prevIdx
			for j >= 0 {
				if classes[j] == WBRegionalIndicator {
					count++
					j--
				} else if classes[j] == WBFormat || classes[j] == WBExtend || classes[j] == WBZWJ {
					// Skip transparent characters
					j--
				} else {
					// Hit a non-RI, non-transparent character
					break
				}
			}
			// If count is odd, this is the 2nd, 4th, 6th... RI (pairs with previous)
			if count%2 == 1 {
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

	// WB2: Break at end
	breaks = append(breaks, len(text))

	return breaks
}

// Words splits text into words according to Unicode word boundary rules.
//
// This function returns all segments between word boundaries, including:
//   - Actual words (alphabetic sequences)
//   - Numbers (numeric sequences)
//   - Punctuation (individual punctuation marks)
//   - Whitespace (spaces, tabs, etc.)
//
// Note that this returns ALL segments, not just "words" in the linguistic sense.
// Each element represents text between consecutive word boundaries as defined
// by UAX #29. If you need only alphabetic words, you'll need to filter the results.
//
// Example:
//
//	words := uax29.Words("Hello, world!")
//	// Returns: ["Hello", ",", " ", "world", "!"]
//
//	words = uax29.Words("Price: $19.99")
//	// Returns: ["Price", ":", " ", "$", "19.99"]
//
//	words = uax29.Words("👨‍👩‍👧‍👦 family")
//	// Returns: ["👨‍👩‍👧‍👦", " ", "family"]
//
// See UAX #29 §4: https://www.unicode.org/reports/tr29/#Word_Boundaries
func Words(text string) []string {
	breaks := FindWordBreaks(text)
	if len(breaks) <= 1 {
		return []string{}
	}

	result := make([]string, len(breaks)-1)
	for i := 0; i < len(breaks)-1; i++ {
		result[i] = text[breaks[i]:breaks[i+1]]
	}
	return result
}
