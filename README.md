# unicode

Implementations of various Unicode® Standard Annexes in Go.

This repository provides Go packages for Unicode text processing algorithms, organized by UAX (Unicode Standard Annex) specification.

## Packages

### uax9 - Bidirectional Algorithm

*(Coming soon)* Implementation of UAX #9 for handling bidirectional text (e.g., mixing Latin and Arabic/Hebrew scripts).

### [uax11](./uax11) - East Asian Width

Implementation of UAX #11 (East Asian Width) for determining character display width in East Asian typography contexts.

Supports:
- East Asian Width property lookup (Ambiguous, Fullwidth, Halfwidth, Narrow, Neutral, Wide)
- Context-based width resolution for ambiguous characters
- Character and string display width calculation
- Terminal emulator and monospace font support
- Complete Unicode 17.0.0 data

```go
import "github.com/SCKelemen/unicode/uax11"

// Determine character width
width := uax11.LookupWidth('中')  // Returns Wide
if uax11.IsWide('A') {
    // Character occupies 2 units
}

// Calculate string display width
width := uax11.StringWidth("Hello世界", uax11.ContextNarrow)  // Returns 9
```

### [uax14](./uax14) - Line Breaking Algorithm

Implementation of UAX #14 (Unicode Line Breaking Algorithm) for finding valid line break opportunities in text.

**Note:** This code was originally implemented in [github.com/SCKelemen/layout](https://github.com/SCKelemen/layout) and has been extracted to a standalone package for reusability.

Supports:
- Word boundaries and spaces
- Mandatory breaks (newlines)
- Configurable hyphenation (none, manual, auto)
- CJK ideographic text
- Punctuation and numeric sequences

```go
import "github.com/SCKelemen/unicode/uax14"

text := "Hello world! This is a test."
breaks := uax14.FindLineBreakOpportunities(text, uax14.HyphensManual)
```

### [uax29](./uax29) - Text Segmentation

Implementation of UAX #29 (Unicode Text Segmentation) for breaking text into grapheme clusters, words, and sentences.

**Status:** Complete with 100% conformance on all official Unicode tests

Supports:
- **Grapheme cluster boundaries** (100.0% - 766/766 tests)
  - User-perceived characters, emoji sequences, combining marks
  - Hangul syllable composition
  - Regional indicator pairs (flag emojis)
  - Indic conjunct sequences for 10+ scripts

- **Word boundaries** (100.0% - 1944/1944 tests)
  - Alphabetic and numeric sequences
  - Contractions, punctuation, hyphenated words
  - Hebrew letter handling, Katakana sequences
  - Emoji modifiers and ZWJ sequences

- **Sentence boundaries** (100.0% - 512/512 tests)
  - Period, question mark, exclamation handling
  - Abbreviation detection, quote and parenthesis handling
  - Multi-script sentence terminators

```go
import "github.com/SCKelemen/unicode/uax29"

// Grapheme clusters
graphemes := uax29.Graphemes("👨‍👩‍👧‍👦")  // Returns ["👨‍👩‍👧‍👦"]

// Words
words := uax29.Words("Hello, world!")  // Returns ["Hello", ",", " ", "world", "!"]

// Sentences
sentences := uax29.Sentences("Hello. World!")  // Returns ["Hello. ", "World!"]
```

### [uax50](./uax50) - Vertical Text Layout

Implementation of UAX #50 (Unicode Vertical Text Layout) for determining character orientation in vertical text.

Supports:
- Vertical orientation property lookup (Rotated, Upright, TransformedUpright, TransformedRotated)
- Character rotation determination for vertical text
- Glyph transformation detection for vertical-specific forms
- Complete Unicode 17.0.0 data
- East Asian typography and mixed-script vertical layouts

```go
import "github.com/SCKelemen/unicode/uax50"

// Determine how to display characters in vertical text
orientation := uax50.LookupOrientation('中')  // Returns Upright
if uax50.IsUpright('A') {
    // Display upright
} else {
    // Rotate 90 degrees clockwise
}
```

## Installation

```bash
go get github.com/SCKelemen/unicode/uax11
go get github.com/SCKelemen/unicode/uax14
go get github.com/SCKelemen/unicode/uax29
go get github.com/SCKelemen/unicode/uax50
```

## Design Philosophy

These implementations focus on practical text layout and rendering needs:
- Simple, focused APIs
- Minimal dependencies (standard library only)
- Performance-conscious
- Well-tested
- Layout-engine agnostic
- Full conformance with Unicode standards

## Conformance

All implementations follow the Unicode Standard and are tested against official Unicode conformance test suites where available:

### Test Coverage
- **UAX #29 (Text Segmentation)**: 100% conformance (3,222/3,222 tests)
  - Grapheme cluster breaking: 766/766 tests
  - Word breaking: 1,944/1,944 tests
  - Sentence breaking: 512/512 tests

### Conformance Testing
Implementations are validated using the official Unicode Character Database (UCD) test files:
- [UAX #29 Test Files](https://www.unicode.org/Public/17.0.0/ucd/auxiliary/) - `GraphemeBreakTest.txt`, `WordBreakTest.txt`, `SentenceBreakTest.txt`
- [Unicode Character Database](https://www.unicode.org/Public/17.0.0/ucd/) - Character property data files

The implementations follow the conformance model described in [UTR #33: Unicode Conformance Model](https://www.unicode.org/reports/tr33/), which defines what it means to conform to Unicode Standard specifications.

## Related Projects

- [github.com/SCKelemen/layout](https://github.com/SCKelemen/layout) - Text layout engine using these UAX implementations

## References

### Metastandards
- [UTR #33: Unicode Conformance Model](https://www.unicode.org/reports/tr33/) - Defines conformance requirements for Unicode Standard implementations
- [UAX #41: Common References for Unicode Standard Annexes](https://www.unicode.org/reports/tr41/) - Common definitions and references used across Unicode Standard Annexes

### Implemented Standards
- [Unicode Standard Annexes](https://www.unicode.org/reports/)
- [UAX #9: Bidirectional Algorithm](https://www.unicode.org/reports/tr9/)
- [UAX #11: East Asian Width](https://www.unicode.org/reports/tr11/)
- [UAX #14: Line Breaking](https://www.unicode.org/reports/tr14/)
- [UAX #29: Text Segmentation](https://www.unicode.org/reports/tr29/)
- [UAX #50: Vertical Text Layout](https://www.unicode.org/reports/tr50/)

## License

MIT
