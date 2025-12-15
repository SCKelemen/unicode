# unicode

Implementations of various Unicode® Standard Annexes in Go.

This repository provides Go packages for Unicode text processing algorithms, organized by UAX (Unicode Standard Annex) specification.

## Packages

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

### uax9 - Bidirectional Algorithm

*(Coming soon)* Implementation of UAX #9 for handling bidirectional text (e.g., mixing Latin and Arabic/Hebrew scripts).

### uax29 - Text Segmentation

*(Coming soon)* Implementation of UAX #29 for grapheme cluster, word, and sentence boundary detection.

## Installation

```bash
go get github.com/SCKelemen/unicode/uax11
go get github.com/SCKelemen/unicode/uax14
```

## Design Philosophy

These implementations focus on practical text layout and rendering needs:
- Simple, focused APIs
- Minimal dependencies (standard library only)
- Performance-conscious
- Well-tested
- Layout-engine agnostic

## Related Projects

- [github.com/SCKelemen/layout](https://github.com/SCKelemen/layout) - Text layout engine using these UAX implementations

## References

- [Unicode Standard Annexes](https://www.unicode.org/reports/)
- [UAX #11: East Asian Width](https://www.unicode.org/reports/tr11/)
- [UAX #14: Line Breaking](https://www.unicode.org/reports/tr14/)
- [UAX #9: Bidirectional Algorithm](https://www.unicode.org/reports/tr9/)
- [UAX #29: Text Segmentation](https://www.unicode.org/reports/tr29/)

## License

MIT
