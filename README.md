# unicode

[![CI](https://github.com/SCKelemen/unicode/workflows/CI/badge.svg)](https://github.com/SCKelemen/unicode/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/SCKelemen/unicode)](https://goreportcard.com/report/github.com/SCKelemen/unicode)

Implementations of various Unicode® Standard Annexes in Go.

This repository provides Go packages for Unicode text processing algorithms, organized by UAX (Unicode Standard Annex) specification.

## Packages

### [uax9](./uax9) - Bidirectional Algorithm

Implementation of UAX #9 (Unicode Bidirectional Algorithm) for handling bidirectional text with mixing LTR and RTL scripts.

**Status:** Complete with 100% conformance (513,494/513,494 tests passing)

Supports:
- **Full bidirectional text reordering** - Proper display of mixed LTR/RTL content
- **Isolating run sequences (BD13)** - Advanced context isolation for complex layouts
- **Explicit formatting characters** - LRE, RLE, LRO, RLO, PDF, LRI, RLI, FSI, PDI
- **Deep embedding nesting** - Up to 125 levels of explicit embedding
- **Bracket pair handling (N0)** - Proper neutral character resolution
- **Automatic direction detection** - Smart paragraph base direction

```go
import "github.com/SCKelemen/unicode/uax9"

// Reorder mixed LTR/RTL text
text := "Hello שלום world"
result := uax9.Reorder(text, uax9.DirectionLTR)

// Auto-detect paragraph direction
dir := uax9.GetParagraphDirection("שלום עולם")  // Returns DirectionRTL

// Get bidi class of a character
class := uax9.GetBidiClass('א')  // Returns R (Right-to-Left)
```

### [uax11](./uax11) - East Asian Width

Implementation of UAX #11 (East Asian Width) for determining character display width in East Asian typography contexts.

**Status:** Complete with comprehensive test coverage

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

**Status:** Complete with 100% conformance (19,338/19,338 tests passing)

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

// Single-pass API - get all three break types at once
breaks := uax29.FindAllBreaks("Hello, world!")
for _, pos := range breaks.Graphemes {
    // Process grapheme boundaries
}
for _, pos := range breaks.Words {
    // Process word boundaries
}
for _, pos := range breaks.Sentences {
    // Process sentence boundaries
}
```

### [uax50](./uax50) - Vertical Text Layout

Implementation of UAX #50 (Unicode Vertical Text Layout) for determining character orientation in vertical text.

**Status:** Complete with comprehensive test coverage

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

### [uts51](./uts51) - Unicode Emoji

Implementation of UTS #51 (Unicode Emoji) for emoji property detection, sequence validation, and terminal rendering support.

**Status:** Complete with 100% conformance (5,223/5,223 tests passing)

Supports:
- **Emoji properties** - All 6 core emoji properties
  - Emoji, Emoji_Presentation, Emoji_Modifier
  - Emoji_Modifier_Base, Emoji_Component, Extended_Pictographic
- **Sequence validation** - All emoji sequence types
  - ZWJ sequences (family emoji, etc.)
  - Modifier sequences (skin tones)
  - Flag sequences (regional indicators)
  - Keycap sequences (#️⃣, *️⃣, 0️⃣-9️⃣)
  - Tag sequences (subdivision flags)
- **Terminal rendering** - Width calculation for emoji display
- **Integration** with UAX #11, #14, #29, #50

```go
import "github.com/SCKelemen/unicode/uts51"

// Check if character is emoji
if uts51.IsEmoji('😀') {
    // Handle emoji
}

// Calculate width for terminal rendering
width := uts51.EmojiWidth('😀')  // Returns 2 (like CJK characters)

// Validate emoji sequences
sequence := []rune{0x1F468, 0x200D, 0x1F469, 0x200D, 0x1F467}  // Family
if uts51.IsValidEmojiSequence(sequence) {
    // Valid ZWJ sequence
}
```

## Installation

```bash
go get github.com/SCKelemen/unicode/uax9
go get github.com/SCKelemen/unicode/uax11
go get github.com/SCKelemen/unicode/uax14
go get github.com/SCKelemen/unicode/uax29
go get github.com/SCKelemen/unicode/uax50
go get github.com/SCKelemen/unicode/uts51
```

## Design Philosophy

These implementations focus on practical text layout and rendering needs:
- Simple, focused APIs
- Minimal dependencies (standard library only)
- Performance-conscious
- Well-tested
- Layout-engine agnostic
- Full conformance with Unicode standards

## Version 2.0.0 Performance Improvements

Version 2.0.0 focuses on performance optimization while maintaining 100% conformance with Unicode standards.

### Table-Driven Binary Search

All packages now use **table-driven O(log n) binary search** for character classification, replacing sequential O(n) checks:

- **UAX #9**: Bidi class lookup optimized with 3,060 precomputed ranges from `DerivedBidiClass.txt`
- **UAX #29**: Unified packed data structure with 4,673 ranges encoding all three break types (grapheme, word, sentence) in 16-bit format

**Performance**: Character classification now runs at ~60-100 ns/op with 0 allocations on Apple M4 Pro.

### Generated Unicode Data

All Unicode property data is now generated directly from official Unicode 17.0.0 data files:
- Download from unicode.org during build
- Parse property files (`DerivedBidiClass.txt`, `GraphemeBreakProperty.txt`, etc.)
- Generate optimized Go code with binary search tables
- Ensures correctness and synchronization with Unicode standard

### Single-Pass API

UAX #29 provides a new `FindAllBreaks()` API that computes grapheme, word, and sentence boundaries in a single traversal:

```go
// Before: Three separate passes
graphemes := uax29.FindGraphemeBreaks(text)
words := uax29.FindWordBreaks(text)
sentences := uax29.FindSentenceBreaks(text)

// After: Single pass with shared classification
breaks := uax29.FindAllBreaks(text)
// Use breaks.Graphemes, breaks.Words, breaks.Sentences
```

This provides a convenient API for applications that need multiple break types, with framework in place for future hierarchical optimization.

## Unicode Version

This repository implements **Unicode 17.0.0** (September 2024).

### Why Not Use Go's Standard Library?

Go's `unicode` package (as of Go 1.23) provides Unicode 15.0.0 data. While it includes some properties we need (e.g., `Regional_Indicator`, `Ideographic`, `Sentence_Terminal`), it is missing:

- **Emoji properties**: `Extended_Pictographic`, `Emoji`, `Emoji_Presentation`, `Emoji_Modifier`, `Emoji_Modifier_Base`, `Emoji_Component`
- **Text segmentation properties**: `Grapheme_Cluster_Break`, `Word_Break`, `Sentence_Break`
- **Layout properties**: `East_Asian_Width`, `Line_Break`, `Vertical_Orientation`

**Design Decision**: We implement all related properties within each specification package (e.g., all emoji properties in `uts51`) rather than mixing standard library and custom implementations. This ensures:

1. **Consistency**: All properties from a specification come from one authoritative source
2. **Completeness**: Unicode 17.0.0 support with the latest emoji and text handling
3. **Maintainability**: Single source of truth for each Unicode specification
4. **Testability**: 100% conformance against official Unicode 17.0.0 test files

When Go's `unicode` package updates to Unicode 17.0.0, we will continue maintaining our implementations to provide the specialized properties not available in the standard library.

## Conformance

All implementations follow the Unicode Standard and are tested against official Unicode conformance test suites where available:

### Test Coverage
- **UAX #9 (Bidirectional Algorithm)**: 100% conformance (513,494/513,494 tests)
  - All explicit embeddings and isolates
  - Multi-isolate sequences and deep nesting (up to 125 levels)
  - Empty isolate handling and overflow isolation
  - Bracket pair matching and neutral resolution
- **UAX #11 (East Asian Width)**: Comprehensive test coverage
  - Character width property lookup for all Unicode code points
  - Context-based ambiguous character resolution
  - Display width calculation for strings
  - Terminal emulator compatibility
- **UAX #14 (Line Breaking)**: 100% conformance (19,338/19,338 tests)
  - All line break classes and combining rules
  - Tailorable break opportunities
  - Complex script handling (CJK, Thai, etc.)
  - Hyphenation support (soft hyphens U+00AD)
- **UAX #29 (Text Segmentation)**: 100% conformance (3,222/3,222 tests)
  - Grapheme cluster breaking: 766/766 tests
  - Word breaking: 1,944/1,944 tests
  - Sentence breaking: 512/512 tests
- **UAX #50 (Vertical Text Layout)**: Comprehensive test coverage
  - Vertical orientation property for all Unicode code points
  - Glyph transformation detection
  - Base orientation determination
  - Mixed-script vertical layout support
- **UTS #51 (Unicode Emoji)**: 100% conformance (5,223/5,223 tests)
  - All 6 emoji properties correctly implemented
  - Complete sequence validation (ZWJ, modifier, flag, keycap, tag sequences)

### Conformance Testing
Implementations are validated using the official Unicode Character Database (UCD) test files:
- [UAX #9 Test Files](https://www.unicode.org/Public/17.0.0/ucd/) - `BidiTest.txt` (513,494 tests), `BidiCharacterTest.txt`
- [UAX #11 Data Files](https://www.unicode.org/Public/17.0.0/ucd/) - `EastAsianWidth.txt` property data
- [UAX #14 Test Files](https://www.unicode.org/Public/17.0.0/ucd/auxiliary/) - `LineBreakTest.txt` (19,338 tests)
- [UAX #29 Test Files](https://www.unicode.org/Public/17.0.0/ucd/auxiliary/) - `GraphemeBreakTest.txt`, `WordBreakTest.txt`, `SentenceBreakTest.txt`
- [UAX #50 Data Files](https://www.unicode.org/Public/17.0.0/ucd/) - `VerticalOrientation.txt` property data
- [UTS #51 Test Files](https://www.unicode.org/Public/emoji/17.0/) - `emoji-test.txt` with 5,223 test cases
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
- [UTS #51: Unicode Emoji](https://www.unicode.org/reports/tr51/)

## License

MIT
