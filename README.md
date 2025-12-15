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

### uax9 - Bidirectional Algorithm

*(Coming soon)* Implementation of UAX #9 for handling bidirectional text (e.g., mixing Latin and Arabic/Hebrew scripts).

### uax29 - Text Segmentation

*(Coming soon)* Implementation of UAX #29 for grapheme cluster, word, and sentence boundary detection.

## Installation

```bash
go get github.com/SCKelemen/unicode/uax14
go get github.com/SCKelemen/unicode/uax50
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
- [UAX #14: Line Breaking](https://www.unicode.org/reports/tr14/)
- [UAX #50: Vertical Text Layout](https://www.unicode.org/reports/tr50/)
- [UAX #9: Bidirectional Algorithm](https://www.unicode.org/reports/tr9/)
- [UAX #29: Text Segmentation](https://www.unicode.org/reports/tr29/)

## License

MIT
