# uax29 - Unicode Text Segmentation

Implementation of [UAX #29: Unicode Text Segmentation](https://www.unicode.org/reports/tr29/) in Go.

**Status:** Implemented with official Unicode test vectors (Unicode 17.0)

## Overview

This package will provide algorithms for breaking text into meaningful units:
- **Grapheme clusters**: User-perceived characters (what users think of as "characters")
- **Words**: Linguistic word boundaries for text selection and cursor movement
- **Sentences**: Sentence boundaries for text processing

## Planned Features

### Grapheme Cluster Boundaries
- Proper handling of combining marks (e.g., `e` + `́` = `é`)
- Hangul syllable composition
- Emoji sequences with Zero Width Joiner (ZWJ)
- Regional indicator sequences (flag emojis)
- Variation selectors

### Word Boundaries
- Alphabetic and numeric sequences
- Proper handling of contractions (don't, can't)
- Punctuation boundaries
- CJK word segmentation (requires dictionary)
- Hyphenated words

### Sentence Boundaries
- Period, question mark, exclamation handling
- Abbreviation detection (Dr., Mrs., etc.)
- Quote and parenthesis handling
- Whitespace rules
- Multiple punctuation handling (e.g., `...`, `?!`)

## Use Cases

- Text editors: cursor movement, selection, deletion
- Search: tokenization and indexing
- Natural language processing
- Text-to-speech: proper phrase boundaries
- Terminal UIs: text selection and wrapping

## Implementation Status

### Grapheme Cluster Boundaries ✅ (100.0% pass rate - 766/766) 🎉
- **COMPLETE** implementation with Unicode 17.0 test vectors
- Handles combining marks, Hangul syllables, all emoji sequences
- Regional indicator pairs (flag emojis) working correctly
- Prepend characters properly supported
- Emoji modifiers (skin tones) correctly classified as Extend
- GB11: Emoji ZWJ sequences fully implemented
- GB9c: Indic conjunct sequences for 10+ scripts (Devanagari, Bengali, Gujarati, Oriya, Telugu, Malayalam, Myanmar, Balinese, Sundanese, Khmer)

### Word Boundaries ✅ (100.0% pass rate - 1944/1944) 🎉
- **COMPLETE** implementation with Unicode 17.0 test vectors
- Handles all alphabetic/numeric sequences, contractions, punctuation
- Regional indicator pairs with ZWJ transparency
- Hebrew letter handling with single/double quotes
- Katakana sequences and ExtendNumLet
- Emoji sequences with modifiers and ZWJ
- Proper handling of Format character exceptions

### Sentence Boundaries ✅ (100.0% pass rate - 512/512) 🎉
- **COMPLETE** implementation with Unicode 17.0 test vectors
- Handles all sentence terminators (., ?, !, and many script-specific terminators)
- Proper handling of abbreviations with ATerm
- Complex Close* Sp* sequences correctly processed
- SB8: Lowercase handling after ATerm Close* Sp*
- SB8a: SContinue and sentence terminal sequences
- SB9/SB10: Close and space handling after terminators
- SB11: Breaking after sentence terminal sequences

## Examples (Planned)

```go
// Grapheme clusters
text := "👨‍👩‍👧‍👦"  // Family emoji (multiple codepoints)
graphemes := uax29.Graphemes(text)
// Returns 1 grapheme cluster

// Words
text := "Hello, world!"
words := uax29.Words(text)
// Returns: ["Hello", ",", " ", "world", "!"]

// Sentences
text := "Hello Dr. Smith. How are you?"
sentences := uax29.Sentences(text)
// Returns: ["Hello Dr. Smith. ", "How are you?"]
```

## Dependencies

This package depends on:
- **UTS #51 (Unicode Emoji)**: Provides authoritative emoji property data (Extended_Pictographic, Regional_Indicator, Emoji_Modifier, ZeroWidthJoiner constant)

## Integration with Other Standards

- **UTS #51 (Unicode Emoji)**: Emoji sequences are treated as single grapheme clusters per UAX #29 GB11
- **UAX #14 (Line Breaking)**: UAX #29 word boundaries inform line break decisions
- **UAX #9 (Bidirectional)**: Both needed for proper text layout
- **UAX #11 (East Asian Width)**: Terminal cursor movement should respect grapheme boundaries

## References

- [UAX #29: Unicode Text Segmentation](https://www.unicode.org/reports/tr29/)
- [Grapheme Cluster Boundaries](https://www.unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries)
- [Word Boundaries](https://www.unicode.org/reports/tr29/#Word_Boundaries)
- [Sentence Boundaries](https://www.unicode.org/reports/tr29/#Sentence_Boundaries)

## License

MIT
