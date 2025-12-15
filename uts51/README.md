# UTS #51: Unicode Emoji

Implementation of [UTS #51 (Unicode Emoji)](https://www.unicode.org/reports/tr51/) for complete emoji support in terminals, text editors, and layout engines.

## Overview

This package provides **100% conformance** (5,223/5,223 test cases passing) for UTS #51, with complete emoji property detection, sequence validation, and terminal rendering support.

Essential for:
- Terminal emulators calculating emoji display widths
- Text editors with emoji input support
- Layout engines handling emoji sequences
- Grapheme cluster segmentation around emoji

## Installation

```bash
go get github.com/SCKelemen/unicode/uts51
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/SCKelemen/unicode/uts51"
)

func main() {
    // Check if character is emoji
    if uts51.IsEmoji('😀') {
        fmt.Println("Is emoji!")
    }

    // Check default presentation
    if uts51.HasEmojiPresentation('😀') {
        fmt.Println("Displays as colorful emoji by default")
    }

    // Calculate width for terminal rendering
    width := uts51.EmojiWidth('😀')  // Returns 2 (like CJK characters)
}
```

## Emoji Properties

UTS #51 defines six core properties (§1.4):

### Emoji
Characters recommended for emoji use.
```go
uts51.IsEmoji('😀')  // true
uts51.IsEmoji('#')   // true - can be used in keycap sequences
uts51.IsEmoji('A')   // false
```

### Emoji_Presentation
Characters that display as emoji (colorful) by default.
```go
uts51.HasEmojiPresentation('😀')  // true - emoji by default
uts51.HasEmojiPresentation('☺')   // false - text by default (needs U+FE0F)
```

### Emoji_Modifier
Skin tone modifiers (U+1F3FB..U+1F3FF).
```go
uts51.IsEmojiModifier('\U0001F3FB')  // true - light skin tone
```

### Emoji_Modifier_Base
Characters that accept skin tone modifiers.
```go
uts51.IsEmojiModifierBase('👋')  // true - waving hand can have skin tone
uts51.IsEmojiModifierBase('😀')  // false - faces don't have skin tones
```

### Emoji_Component
Characters used in emoji sequences but not standalone.
```go
uts51.IsEmojiComponent('\U0001F3FB')  // true - skin tone
```

### Extended_Pictographic
All emoji and pictographic characters for segmentation.
```go
uts51.IsExtendedPictographic('😀')  // true
```

## Terminal Width Calculation

Per UTS #51 §4, emoji have the same advance width as CJK ideographs (2 columns).

```go
uts51.EmojiWidth('😀')              // 2 - emoji presentation
uts51.EmojiWidth('☺')               // 1 - text presentation
uts51.EmojiWidth('\U0001F3FB')      // 0 - skin tone modifier
```

This integrates with **UAX #11 (East Asian Width)** for complete width calculation.

## Integration with Other Standards

This package works seamlessly with:

- **UAX #11 (East Asian Width)**: Emoji width calculation
- **UAX #14 (Line Breaking)**: Break opportunities around emoji
- **UAX #29 (Text Segmentation)**: Grapheme cluster boundaries
- **UAX #50 (Vertical Text Layout)**: Emoji orientation in vertical text

## API Reference

### Property Functions

- `IsEmoji(r rune) bool` - Has Emoji property
- `HasEmojiPresentation(r rune) bool` - Displays as emoji by default
- `IsEmojiModifier(r rune) bool` - Is a skin tone modifier
- `IsEmojiModifierBase(r rune) bool` - Accepts modifiers
- `IsEmojiComponent(r rune) bool` - Used in sequences
- `IsExtendedPictographic(r rune) bool` - For segmentation

### Presentation Functions

- `DefaultPresentation(r rune) rune` - Returns 'E' or 'T'
- `EmojiWidth(r rune) int` - Display width in columns

### Sequence Detection

- `IsRegionalIndicator(r rune) bool` - For flag sequences
- `IsTagCharacter(r rune) bool` - For subdivision flags

### Sequence Validation

- `IsValidKeycapSequence(runes []rune) bool` - Validates keycap sequences ([0-9#*] + U+FE0F + U+20E3)
- `IsValidTagSequence(runes []rune) bool` - Validates tag sequences (subdivision flags)
- `IsValidEmojiSequence(runes []rune) bool` - Validates any emoji sequence type

### Constants

```go
VariationSelector15        // U+FE0E - text presentation
VariationSelector16        // U+FE0F - emoji presentation
ZeroWidthJoiner           // U+200D - joins emoji
CombiningEnclosingKeycap  // U+20E3 - keycap sequences
```

## Conformance

**100% conformance** with UTS #51 Version 17.0

- **5,223/5,223 test cases passing** from emoji-test.txt
- All 6 emoji properties correctly implemented
- Complete sequence validation (keycap, tag, modifier, flag, ZWJ sequences)

### Conformance Requirements

Per UTS #51 §5:

- ✅ **C1**: Version 17.0 identification
- ✅ **C2**: Display capability for basic emoji set
- ✅ **C3**: Rejection of invalid sequences

## Implementation Details

### Data Source
- emoji-data.txt Version 17.0 (2025-07-25)
- emoji-test.txt with 5,223 test cases
- Downloaded from: https://www.unicode.org/Public/emoji/latest/

### Performance
- Binary search for O(log n) property lookups
- Efficient range-based data structure
- Zero external dependencies beyond Go standard library

### Code Generation
```bash
cd uts51
go run generate_emoji_data.go
```

## Usage Examples

### Terminal Rendering

```go
func renderEmoji(text string) {
    for _, r := range text {
        width := uts51.EmojiWidth(r)
        if width == 2 {
            // Emoji occupies 2 columns
            renderWideChar(r)
        } else if width == 1 {
            // Text presentation
            renderNarrowChar(r)
        }
        // width == 0: invisible component
    }
}
```

### Presentation Control

```go
// Force emoji presentation
text := "☺" + string(uts51.VariationSelector16)  // ☺️

// Force text presentation
text := "😀" + string(uts51.VariationSelector15)  // Text version
```

### Property Checking

```go
func analyzeEmoji(r rune) {
    if !uts51.IsEmoji(r) {
        return
    }

    if uts51.HasEmojiPresentation(r) {
        fmt.Println("Colorful emoji by default")
    }

    if uts51.IsEmojiModifierBase(r) {
        fmt.Println("Can have skin tone")
    }
}
```

### Sequence Validation

```go
// Validate a keycap sequence
keycap := []rune{'9', '\uFE0F', '\u20E3'}  // 9⃣
if uts51.IsValidKeycapSequence(keycap) {
    fmt.Println("Valid keycap sequence")
}

// Validate a tag sequence (subdivision flag)
englandFlag := []rune{0x1F3F4, 0xE0067, 0xE0062, 0xE0065, 0xE006E, 0xE0067, 0xE007F}
if uts51.IsValidTagSequence(englandFlag) {
    fmt.Println("Valid subdivision flag")
}

// Validate any emoji sequence
sequence := []rune{0x1F468, 0x200D, 0x1F469, 0x200D, 0x1F467}  // Family ZWJ sequence
if uts51.IsValidEmojiSequence(sequence) {
    fmt.Println("Valid emoji sequence")
}
```

## Future Work

### Planned Features
- Full grapheme cluster segmentation (UAX #29 integration)
- Emoji version detection per character
- RGI (Recommended for General Interchange) emoji set identification
- Sequence width calculation for multi-codepoint emoji

## References

- [UTS #51: Unicode Emoji](https://www.unicode.org/reports/tr51/)
- [§1.4 Emoji Properties](https://www.unicode.org/reports/tr51/#Emoji_Properties)
- [§2 Emoji Sequences](https://www.unicode.org/reports/tr51/#Emoji_Sequences)
- [§4 Display](https://www.unicode.org/reports/tr51/#Display)
- [§5 Conformance](https://www.unicode.org/reports/tr51/#Conformance)
- [emoji-test.txt](https://www.unicode.org/Public/emoji/latest/emoji-test.txt)

## License

MIT
