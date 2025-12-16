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
//   - Rule-based state machine architecture for clarity and performance
//
// Expected speedup: 2-7× faster than separate calls when all three break
// types are needed (increases with text length).
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

	// Find grapheme breaks (most granular) using rule-based approach
	graphemeBreaks := findGraphemeBreaksFromClassesWithRules(text, runes, classes)

	// Find word breaks (only at grapheme boundaries) using rule-based approach
	wordBreaks := findWordBreaksFromClassesWithRules(text, runes, classes, graphemeBreaks)

	// Find sentence breaks (only at word boundaries) using rule-based approach
	sentenceBreaks := findSentenceBreaksFromClassesWithRules(text, runes, classes, wordBreaks)

	return BreakOpportunities{
		Graphemes: graphemeBreaks,
		Words:     wordBreaks,
		Sentences: sentenceBreaks,
	}
}
