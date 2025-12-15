// Package uax29 implements Unicode Text Segmentation (UAX #29).
//
// This package provides algorithms for breaking text into grapheme clusters,
// words, and sentences according to the Unicode Standard Annex #29 specification.
//
// Based on: https://www.unicode.org/reports/tr29/
//
// # Status
//
// Complete implementation with 100% conformance on all official Unicode tests:
//   - Grapheme cluster boundaries: 766/766 tests (100.0%)
//   - Word boundaries: 1,944/1,944 tests (100.0%)
//   - Sentence boundaries: 512/512 tests (100.0%)
//
// # Overview
//
// Text segmentation is fundamental to many text processing tasks. UAX #29 defines
// three types of boundaries:
//
// 1. Grapheme Cluster Boundaries (https://www.unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries)
//    - User-perceived characters (what users think of as "characters")
//    - Handles combining marks, emoji sequences, Hangul syllables
//    - Essential for cursor movement, text selection, and character counting
//
// 2. Word Boundaries (https://www.unicode.org/reports/tr29/#Word_Boundaries)
//    - Linguistic word boundaries for text selection and indexing
//    - Handles contractions, hyphenated words, numeric sequences
//    - Used in text editors, search engines, and NLP applications
//
// 3. Sentence Boundaries (https://www.unicode.org/reports/tr29/#Sentence_Boundaries)
//    - Sentence breaks for text processing and analysis
//    - Handles abbreviations, quotes, and various punctuation
//    - Used in NLP, summarization, and text-to-speech
//
// # Usage
//
//	import "github.com/SCKelemen/unicode/uax29"
//
//	text := "Hello, world! How are you?"
//
//	// Find grapheme cluster boundaries
//	graphemes := uax29.Graphemes("👨‍👩‍👧‍👦")  // Returns ["👨‍👩‍👧‍👦"]
//
//	// Find word boundaries
//	words := uax29.Words("Hello, world!")  // Returns ["Hello", ",", " ", "world", "!"]
//
//	// Find sentence boundaries
//	sentences := uax29.Sentences("Hello. World!")  // Returns ["Hello. ", "World!"]
//
//	// Get boundary positions (byte offsets)
//	breaks := uax29.FindGraphemeBreaks(text)  // Returns []int of byte positions
//	breaks = uax29.FindWordBreaks(text)
//	breaks = uax29.FindSentenceBreaks(text)
//
// # Grapheme Cluster Boundaries
//
// Grapheme clusters represent user-perceived characters. This is more complex
// than Unicode code points due to:
//   - Combining marks: base + combining diacriticals (e + ◌́ = é)
//   - Emoji sequences: ZWJ sequences (👨‍👩‍👧‍👦), modifiers (👋🏽)
//   - Hangul syllables: conjoining Jamo (ᄒ + ᅡ + ᆫ = 한)
//   - Regional indicators: flag emojis (🇺🇸 = U+1F1FA + U+1F1F8)
//   - Indic conjuncts: consonant clusters with virama
//
// See UAX #29 §3: https://www.unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries
//
// Rules implemented:
//   - GB3-GB5: CR, LF, and Control handling
//   - GB6-GB8: Hangul syllable composition
//   - GB9-GB9c: Extend, ZWJ, and SpacingMark handling
//   - GB11: Extended pictographic ZWJ sequences (emoji)
//   - GB12-GB13: Regional indicator pairs (flags)
//
// # Word Boundaries
//
// Word boundaries enable text selection, search indexing, and linguistic analysis.
// This implementation handles:
//   - Alphabetic and numeric sequences
//   - Contractions (don't, can't) and possessives (John's)
//   - Hebrew letters with quotes
//   - Katakana sequences
//   - Emoji sequences with modifiers
//   - Regional indicator pairs with transparency
//
// See UAX #29 §4: https://www.unicode.org/reports/tr29/#Word_Boundaries
//
// Rules implemented:
//   - WB3-WB4: Format and Extend transparency
//   - WB5-WB7c: Letter sequences and contractions
//   - WB8-WB12: Numeric sequences
//   - WB13-WB13b: ExtendNumLet and Katakana
//   - WB15-WB16: Regional indicator pairs
//
// # Sentence Boundaries
//
// Sentence boundaries enable proper text chunking for NLP and document processing.
// This implementation handles:
//   - Sentence terminators: period, question mark, exclamation
//   - Abbreviations: Dr., Mrs., etc.
//   - Quotes and parentheses
//   - Multiple punctuation: ..., ?!
//   - Script-specific terminators
//
// See UAX #29 §5: https://www.unicode.org/reports/tr29/#Sentence_Boundaries
//
// Rules implemented:
//   - SB3-SB5: CR, LF, and paragraph separator handling
//   - SB6-SB7: ATerm handling with abbreviations
//   - SB8-SB8a: Lowercase and SContinue after terminators
//   - SB9-SB10: Close and space sequences
//   - SB11: Breaking after terminal sequences
//
// # Dependencies
//
// This package depends on UTS #51 (Unicode Emoji) for authoritative emoji property data:
//   - Extended_Pictographic property for emoji detection (GB11)
//   - Regional_Indicator property for flag emoji sequences (GB12/GB13)
//   - Emoji_Modifier property for skin tone handling (GB9)
//   - ZeroWidthJoiner constant for ZWJ sequences (GB11)
//
// Using UTS #51 ensures consistency across all Unicode implementations in this module
// and provides complete, data-driven emoji support.
//
// # Conformance
//
// This implementation conforms to Unicode 17.0 and passes all official conformance
// tests from the Unicode Character Database:
//   - GraphemeBreakTest.txt: https://www.unicode.org/Public/17.0.0/ucd/auxiliary/GraphemeBreakTest.txt
//   - WordBreakTest.txt: https://www.unicode.org/Public/17.0.0/ucd/auxiliary/WordBreakTest.txt
//   - SentenceBreakTest.txt: https://www.unicode.org/Public/17.0.0/ucd/auxiliary/SentenceBreakTest.txt
//
// # References
//
//   - UAX #29: https://www.unicode.org/reports/tr29/
//   - §3 Grapheme Cluster Boundaries: https://www.unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries
//   - §4 Word Boundaries: https://www.unicode.org/reports/tr29/#Word_Boundaries
//   - §5 Sentence Boundaries: https://www.unicode.org/reports/tr29/#Sentence_Boundaries
//   - Unicode 17.0 Test Data: https://www.unicode.org/Public/17.0.0/ucd/auxiliary/
package uax29
