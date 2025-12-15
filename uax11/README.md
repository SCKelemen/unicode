# UAX #11: East Asian Width

Implementation of [UAX #11 (East Asian Width)](https://www.unicode.org/reports/tr11/) for determining character display width in East Asian typography contexts.

## Overview

This package provides the `East_Asian_Width` property for Unicode characters, which determines how characters should be displayed in terms of width in the context of East Asian typography. This is essential for:

- Terminal emulators and monospace fonts
- Text editors with East Asian language support
- Legacy East Asian character encoding conversions
- Line breaking and text justification
- Calculating string display widths

## Installation

```bash
go get github.com/SCKelemen/unicode/uax11
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/SCKelemen/unicode/uax11"
)

func main() {
    // Check character width classification
    fmt.Printf("Width of 'A': %v\n", uax11.LookupWidth('A'))      // Narrow
    fmt.Printf("Width of '中': %v\n", uax11.LookupWidth('中'))     // Wide

    // Calculate string display width
    width := uax11.StringWidth("Hello世界", uax11.ContextNarrow)
    fmt.Printf("Display width: %d\n", width)  // 9 (5 + 4)
}
```

## Width Classifications

The `East_Asian_Width` property has six values:

### F - Fullwidth
Characters with fullwidth `<wide>` compatibility decomposition. These are fullwidth forms of characters that also have narrow forms.

**Examples:** Fullwidth Latin letters (U+FF01..U+FF5E)
**Display width:** 2 units (1.0 em in fixed-pitch fonts)

### H - Halfwidth
Characters with halfwidth `<narrow>` compatibility decomposition. Typically for katakana and Hangul.

**Examples:** Halfwidth Katakana (U+FF65..U+FFDC)
**Display width:** 1 unit (0.5 em in fixed-pitch fonts)

### W - Wide
Characters that are always wide in East Asian typography. They behave like ideographs: allowing line breaks after each character and remaining upright in vertical text.

**Examples:** CJK ideographs, emoji, fullwidth symbols
**Display width:** 2 units (1.0 em in fixed-pitch fonts)

### Na - Narrow
Characters that are always narrow and have explicit fullwidth or wide counterparts.

**Examples:** ASCII characters (U+0020..U+007E)
**Display width:** 1 unit (0.5 em in fixed-pitch fonts)

### A - Ambiguous
Characters that can be narrow or wide depending on context. Requires contextual resolution.

**Examples:** Greek, Cyrillic letters in East Asian character sets
**Display width:** Context-dependent (1 or 2 units)

### N - Neutral
Characters not in legacy East Asian character sets. Neutral with respect to East Asian typography.

**Examples:** Most non-East Asian scripts (Devanagari, Arabic), modern emoji variants
**Display width:** 1 unit (0.5 em in fixed-pitch fonts)

## API Reference

### Core Functions

#### `LookupWidth(r rune) Width`
Returns the `East_Asian_Width` property value for a character.

```go
width := uax11.LookupWidth('中')  // Returns Wide
width = uax11.LookupWidth('A')   // Returns Narrow
width = uax11.LookupWidth('Ω')   // Returns Ambiguous
```

#### `IsWide(r rune) bool`
Returns true if the character has Wide or Fullwidth classification.

```go
if uax11.IsWide('中') {
    // Character occupies full width (2 units)
}
```

#### `IsNarrow(r rune) bool`
Returns true if the character has Narrow or Halfwidth classification.

```go
if uax11.IsNarrow('A') {
    // Character occupies half width (1 unit)
}
```

#### `IsAmbiguous(r rune) bool`
Returns true if the character requires contextual width resolution.

```go
if uax11.IsAmbiguous('Ω') {
    // Need to resolve based on context
}
```

### Context-Based Resolution

#### `ResolveWidth(r rune, ctx Context) Width`
Returns the practical display width for a character in a given context.

```go
// Greek Omega in East Asian context
width := uax11.ResolveWidth('Ω', uax11.ContextEastAsian)  // Returns Wide

// Greek Omega in non-East Asian context
width = uax11.ResolveWidth('Ω', uax11.ContextNarrow)      // Returns Narrow
```

#### `CharWidth(r rune, ctx Context) int`
Returns the display width of a character (1 or 2 units) in the given context.

```go
width := uax11.CharWidth('A', uax11.ContextNarrow)       // Returns 1
width = uax11.CharWidth('中', uax11.ContextNarrow)       // Returns 2
width = uax11.CharWidth('Ω', uax11.ContextEastAsian)    // Returns 2
```

#### `StringWidth(s string, ctx Context) int`
Calculates the total display width of a string in the given context.

```go
width := uax11.StringWidth("Hello", uax11.ContextNarrow)        // Returns 5
width = uax11.StringWidth("Hello世界", uax11.ContextNarrow)     // Returns 9
width = uax11.StringWidth("ΩΩΩ", uax11.ContextEastAsian)       // Returns 6
```

## Ambiguous Width Handling

Ambiguous (A) characters require contextual resolution per [UAX #11 §5](https://www.unicode.org/reports/tr11/#Ambiguous).

### Contexts

#### `ContextNarrow` (default)
- Treats Ambiguous characters as narrow (1 unit)
- Use for non-East Asian environments

#### `ContextEastAsian`
- Treats Ambiguous characters as wide (2 units)
- Use for East Asian typography contexts (Chinese, Japanese, Korean)

### Context Determination

Context can be determined by:
- Language tags
- Script identification
- Font association
- Data source
- Explicit markup

## Usage Examples

### Terminal Width Calculation

```go
func displayLine(text string) {
    width := uax11.StringWidth(text, uax11.ContextNarrow)
    fmt.Printf("Text: %s (width: %d)\n", text, width)
}

displayLine("Hello")      // width: 5
displayLine("Hello世界")  // width: 9
displayLine("中国日本")   // width: 8
```

### Text Alignment

```go
func padTo(text string, targetWidth int) string {
    width := uax11.StringWidth(text, uax11.ContextNarrow)
    padding := targetWidth - width
    if padding <= 0 {
        return text
    }
    return strings.Repeat(" ", padding) + text
}
```

### Width-Aware Truncation

```go
func truncate(text string, maxWidth int) string {
    currentWidth := 0
    for i, r := range text {
        charWidth := uax11.CharWidth(r, uax11.ContextNarrow)
        if currentWidth + charWidth > maxWidth {
            return text[:i]
        }
        currentWidth += charWidth
    }
    return text
}
```

## Conformance

This implementation follows UAX #11 specifications:

- Based on Unicode 17.0.0 EastAsianWidth.txt data
- Provides default width classifications for reliable rendering
- Supports all six width values (A, F, H, N, Na, W)
- Implements contextual resolution for Ambiguous characters
- The property is informative and may be overridden by higher-level protocols

As specified in [UAX #11 §6](https://www.unicode.org/reports/tr11/#Recommendations), the East_Asian_Width property is informative. Document formats should allow authors to specify character width preferences, with this property serving as the default.

## Important Notes

### Terminal Emulator Warning

Per [UAX #11 §6](https://www.unicode.org/reports/tr11/#Recommendations):

> "The East_Asian_Width property is not intended for use by modern terminal emulators without appropriate tailoring on a case-by-case basis."

Modern terminal emulators may need additional tailoring for:
- Emoji and symbol characters
- Combining marks
- Control characters
- Font-specific rendering

### Simplified Width Calculation

The `CharWidth` and `StringWidth` functions provide simplified width calculations. For complete text width analysis, consider combining with:
- **UAX #29** for grapheme cluster segmentation
- **UAX #14** for line breaking
- **UAX #50** for vertical text orientation

## Implementation Details

### Data Source
- Unicode Character Database: EastAsianWidth-17.0.0.txt
- Downloaded from: https://www.unicode.org/Public/UCD/latest/ucd/EastAsianWidth.txt

### Performance
- Binary search for O(log n) lookup time
- Compact range-based data structure
- No external dependencies beyond Go standard library

### Code Generation
The `east_asian_width_data.go` file is automatically generated from Unicode data:

```bash
cd uax11
go run generate_eaw_data.go
```

### Default Width Ranges

Per the UAX #11 data file, certain unassigned code point ranges default to Wide:
- CJK Unified Ideographs Extension A: U+3400..U+4DBF
- CJK Unified Ideographs: U+4E00..U+9FFF
- CJK Compatibility Ideographs: U+F900..U+FAFF
- Plane 2: U+20000..U+2FFFD
- Plane 3: U+30000..U+3FFFD

All other unassigned code points default to Neutral.

## References

- [UAX #11: East Asian Width](https://www.unicode.org/reports/tr11/)
- [§3 Definitions](https://www.unicode.org/reports/tr11/#ED1)
- [§5 Ambiguous Characters](https://www.unicode.org/reports/tr11/#Ambiguous)
- [§6 Recommendations](https://www.unicode.org/reports/tr11/#Recommendations)
- [Unicode Character Database](https://www.unicode.org/reports/tr44/)

## License

MIT
