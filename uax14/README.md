# uax14 - Unicode Line Breaking Algorithm

Implementation of [UAX #14: Unicode Line Breaking Algorithm](https://www.unicode.org/reports/tr14/) in Go.

This code was originally implemented in [github.com/SCKelemen/layout](https://github.com/SCKelemen/layout) and has been extracted to a standalone package for reusability.

## Features

- Finds valid line break opportunities in text according to UAX #14
- Supports multiple hyphenation modes (none, manual, auto)
- Handles word boundaries and spaces
- Supports CJK ideographic text
- Handles mandatory breaks (newlines, paragraph separators)
- Respects punctuation and numeric sequence rules
- Returns byte positions for direct string slicing

## Installation

```bash
go get github.com/SCKelemen/unicode/uax14
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/SCKelemen/unicode/uax14"
)

func main() {
    text := "Hello world! This is a test."
    breaks := uax14.FindLineBreakOpportunities(text, uax14.HyphensManual)

    fmt.Println("Break positions:", breaks)
    // Output: Break positions: [0 6 13 17 21 28]

    // Use break positions to wrap text
    for i := 1; i < len(breaks); i++ {
        segment := text[breaks[i-1]:breaks[i]]
        fmt.Printf("Segment %d: %q\n", i, segment)
    }
}
```

## Hyphenation Modes

The package supports three hyphenation modes:

### `HyphensNone`
Disables all hyphenation. No breaks are allowed at hyphens (hard or soft).

```go
breaks := uax14.FindLineBreakOpportunities("super-cali", uax14.HyphensNone)
// No break at hyphen
```

### `HyphensManual` (Recommended default)
Only allows breaks at U+00AD soft hyphens. Hard hyphens (regular `-`) are treated as regular characters.

```go
// Soft hyphen: U+00AD
breaks := uax14.FindLineBreakOpportunities("super\u00ADcali", uax14.HyphensManual)
// Break allowed at soft hyphen
```

### `HyphensAuto`
Allows breaks at all hyphens. Dictionary-based automatic hyphenation is not yet implemented.

```go
breaks := uax14.FindLineBreakOpportunities("super-cali", uax14.HyphensAuto)
// Break allowed at hard hyphen
```

## Implementation Details

This is a simplified implementation focusing on practical line breaking for common text layout scenarios:

- **Simplified break classes**: Uses the core UAX #14 break classes but focuses on common cases
- **Word-boundary focused**: Optimized for breaking at word boundaries (spaces)
- **Ideographic support**: Properly handles CJK text with breaks between ideographic characters
- **Byte positions**: Returns byte offsets (not rune indices) for direct string slicing
- **Zero dependencies**: Uses only the Go standard library

## Limitations

This is a simplified implementation focusing on practical line breaking. Some UAX #14 features are not fully implemented:

### Not Implemented
- Dictionary-based automatic hyphenation (for HyphensAuto mode)

### What Works
- Regular newlines (`\n`, `\r`, `\r\n`)
- Unicode line/paragraph separators (U+2028, U+2029, U+0085)
- Zero-width spaces (U+200B)
- Word boundaries (spaces and tabs)
- Soft hyphens (U+00AD)
- Hard hyphens with appropriate mode settings
- CJK ideographic text
- Non-breaking spaces (U+00A0)
- Word joiners (U+2060)
- Combining marks and emoji
- Mixed scripts
- Punctuation and numeric sequences

## Examples

See [example_test.go](./example_test.go) for more usage examples, including:
- Basic text wrapping
- Soft hyphen handling
- CJK text processing
- Line wrapping algorithms

## Testing

```bash
go test -v
go test -bench=.
```

## References

- [UAX #14: Unicode Line Breaking Algorithm](https://www.unicode.org/reports/tr14/)
- [CSS Text Module Level 3 - Hyphenation](https://www.w3.org/TR/css-text-3/#hyphenation)
- [Unicode Line Breaking Properties](http://www.unicode.org/reports/tr14/#Table1)

## License

MIT
