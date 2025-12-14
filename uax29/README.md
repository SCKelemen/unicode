# uax29 - Unicode Text Segmentation

Implementation of [UAX #29: Unicode Text Segmentation](https://www.unicode.org/reports/tr29/) in Go.

**Status:** Not yet implemented

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

## Implementation Plan

1. **Grapheme cluster boundaries** (Priority: High)
   - Essential for cursor movement and text selection
   - Emoji support increasingly important

2. **Word boundaries** (Priority: High)
   - Needed for text selection and layout
   - Works with UAX #14 line breaking

3. **Sentence boundaries** (Priority: Medium)
   - Useful for text processing
   - Less critical for layout/rendering

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

## Relationship to Other UAX

- **UAX #14 (Line Breaking)**: UAX #29 word boundaries inform line break decisions
- **UAX #9 (Bidirectional)**: Both needed for proper text layout

## References

- [UAX #29: Unicode Text Segmentation](https://www.unicode.org/reports/tr29/)
- [Grapheme Cluster Boundaries](https://www.unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries)
- [Word Boundaries](https://www.unicode.org/reports/tr29/#Word_Boundaries)
- [Sentence Boundaries](https://www.unicode.org/reports/tr29/#Sentence_Boundaries)

## License

MIT
