# UAX #50: Unicode Vertical Text Layout

Implementation of [UAX #50 (Unicode Vertical Text Layout)](https://www.unicode.org/reports/tr50/) for determining character orientation in vertical text.

## Overview

This package provides the `Vertical_Orientation` property for Unicode characters, which determines how characters should be displayed in vertical text layouts. This is essential for:

- East Asian typography (Chinese, Japanese, Korean)
- Mixed-script vertical text
- Text layout engines and rendering systems
- Font shaping for vertical text

## Installation

```bash
go get github.com/SCKelemen/unicode/uax50
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/SCKelemen/unicode/uax50"
)

func main() {
    // Check character orientation
    if uax50.IsUpright('中') {
        fmt.Println("Display CJK ideograph upright")
    }

    if uax50.IsRotated('A') {
        fmt.Println("Rotate Latin letter 90° clockwise")
    }

    // Get detailed orientation
    orientation := uax50.LookupOrientation('。')
    fmt.Printf("Ideographic full stop: %v\n", orientation)
}
```

## Orientation Values

The `Vertical_Orientation` property has four values:

### R - Rotated
Characters that should be rotated 90 degrees clockwise from their appearance in Unicode code charts. This is the default for most Latin, Greek, Cyrillic, and other scripts.

**Examples:** A-Z, 0-9, most punctuation

### U - Upright
Characters that maintain the same orientation as shown in Unicode code charts. This is the default for CJK ideographs and related characters.

**Examples:** 中, 日, 本, ひらがな, カタカナ

### Tu - Transformed Upright
Characters that require different glyphs in vertical text, with fallback to Upright orientation if the vertical glyph is not available.

**Examples:** 、 (ideographic comma), 。 (ideographic full stop)

### Tr - Transformed Rotated
Characters that require different glyphs in vertical text, with fallback to Rotated orientation if the vertical glyph is not available.

**Examples:** 〜 (wave dash)

## API Reference

### Core Functions

#### `LookupOrientation(r rune) Orientation`
Returns the `Vertical_Orientation` property value for a character. Performs efficient binary search on Unicode data.

```go
orientation := uax50.LookupOrientation('中')  // Returns Upright
```

#### `IsUpright(r rune) bool`
Returns true if the character should be displayed upright (U or Tu).

```go
if uax50.IsUpright('中') {
    // Display upright
}
```

#### `IsRotated(r rune) bool`
Returns true if the character should be rotated 90° clockwise (R or Tr).

```go
if uax50.IsRotated('A') {
    // Rotate 90° clockwise
}
```

#### `RequiresTransformation(r rune) bool`
Returns true if the character may need a different glyph in vertical text (Tu or Tr).

```go
if uax50.RequiresTransformation('。') {
    // Try to use vertical glyph variant
}
```

#### `GetBaseOrientation(r rune) Orientation`
Returns the base orientation (U or R) for use when transformed glyphs are not available.

```go
base := uax50.GetBaseOrientation('。')  // Returns Upright
```

## Usage Examples

### Simple Vertical Text Layout

```go
func layoutVertical(text string) {
    for _, r := range text {
        if uax50.IsUpright(r) {
            // Display character upright
            displayUpright(r)
        } else {
            // Rotate character 90° clockwise
            displayRotated(r)
        }
    }
}
```

### With Glyph Transformation

```go
func layoutVerticalWithTransform(text string) {
    for _, r := range text {
        if uax50.RequiresTransformation(r) {
            // Try to get vertical-specific glyph
            if glyph := getVerticalGlyph(r); glyph != nil {
                display(glyph)
                continue
            }
            // Fall back to base orientation
            r = getTransformedGlyph(r)
        }

        if uax50.IsUpright(r) {
            displayUpright(r)
        } else {
            displayRotated(r)
        }
    }
}
```

## Conformance

This implementation follows UAX #50 specifications:

- Based on Unicode 17.0.0 VerticalOrientation.txt data
- Provides default orientation values for reliable document interchange
- Supports all four orientation values (R, U, Tu, Tr)
- The property is informative and may be overridden by higher-level protocols

As specified in [UAX #50 §2](https://www.unicode.org/reports/tr50/#Scope), the Vertical_Orientation property is informative. Document formats should allow authors to specify character orientation preferences, with this property serving as the default.

## Implementation Details

### Data Source
- Unicode Character Database: VerticalOrientation-17.0.0.txt
- Downloaded from: https://unicode.org/Public/UCD/latest/ucd/VerticalOrientation.txt

### Performance
- Binary search for O(log n) lookup time
- Compact range-based data structure
- No external dependencies beyond Go standard library

### Code Generation
The `vertical_orientation_data.go` file is automatically generated from Unicode data:

```bash
cd uax50
go run generate_vo_data.go
```

## Grapheme Clusters

UAX #50 [§3.1](https://www.unicode.org/reports/tr50/#Grapheme_Clusters) notes that "the interesting unit of text is not the character, but a grapheme cluster." While this package provides character-level orientation, text layout engines should consider using UAX #29 (Text Segmentation) for proper grapheme cluster handling to ensure combining marks are oriented consistently with their base characters.

## References

- [UAX #50: Unicode Vertical Text Layout](https://www.unicode.org/reports/tr50/)
- [§3: Determining Vertical Orientation](https://www.unicode.org/reports/tr50/#Determining_Vertical_Orientation)
- [§4: Default Orientation](https://www.unicode.org/reports/tr50/#Default_Orientation)
- [Unicode Character Database](https://www.unicode.org/reports/tr44/)
- [TR #41: Common References](https://www.unicode.org/reports/tr41/)

## License

MIT
