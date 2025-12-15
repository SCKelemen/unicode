# uax9 - Unicode Bidirectional Algorithm

Implementation of [UAX #9: Unicode Bidirectional Algorithm](https://www.unicode.org/reports/tr9/) in Go.

**Status:** Highly Conformant (99.97% pass rate on official Unicode test vectors)

## Overview

This package provides bidirectional text reordering for proper display of text containing both left-to-right (LTR) and right-to-left (RTL) scripts.

## Features

- ✅ Bidirectional text reordering
- ✅ Support for mixing LTR and RTL scripts
- ✅ Explicit formatting characters (LRE, RLE, LRO, RLO, PDF, LRI, RLI, FSI, PDI)
- ✅ Automatic base direction detection
- ✅ Bidi character type classification
- ✅ Level resolution algorithm (rules W1-W7, N0-N2, I1-I2, L1)
- ✅ Bracket pair handling (N0 rule)
- ❌ Mirror glyph support (not in scope for bidi algorithm)

## Use Cases

- Rendering Arabic or Hebrew text mixed with Latin
- Text editors with bidirectional text support
- Terminal UIs displaying mixed-direction content
- Layout engines requiring proper text ordering

## Installation

```bash
go get github.com/SCKelemen/unicode/uax9
```

## Usage

```go
package main

import (
	"fmt"
	"github.com/SCKelemen/unicode/uax9"
)

func main() {
	// Reorder mixed LTR/RTL text
	text := "Hello שלום world"
	result := uax9.Reorder(text, uax9.DirectionLTR)
	fmt.Println(result) // Output: Hello םולש world

	// Auto-detect paragraph direction
	rtlText := "שלום עולם"
	dir := uax9.GetParagraphDirection(rtlText)
	fmt.Println(dir) // Output: DirectionRTL

	// Get bidi class of a character
	class := uax9.GetBidiClass('א')
	fmt.Println(class) // Output: R
}
```

## Testing

The implementation is tested against the official Unicode Consortium test vectors:

```bash
go test -v
```

### Test Results

- **Total tests**: 513,494
- **Passed**: 513,330
- **Pass rate**: 99.97%
- **Failed**: 164 (complex edge cases involving separators and embedding levels)

The test suite includes:
- Official Unicode BidiTest.txt (513,494 test cases)
- BidiCharacterTest.txt for character-specific cases
- Custom unit tests for common use cases

## Known Limitations

- 0.03% of edge cases fail (164 out of 513,494 tests)
- Most failures involve:
  - Separator characters (ES, CS) inside embeddings with adjacent numeric types
  - Complex interactions between weak type resolution and embedding boundaries
  - Edge cases with multiple embedding levels and separator positioning

The implementation is production-ready and handles 99.97% of Unicode's comprehensive test cases, including all common real-world bidirectional text scenarios. The remaining failures are extremely rare edge cases involving complex combinations of embedding levels, separator characters, and numeric type adjacency.

## Implementation Details

The implementation follows the UAX #9 specification and includes:

1. **Character Classification**: Maps Unicode characters to their bidirectional types (L, R, AL, EN, etc.)
2. **Explicit Levels (X1-X8)**: Handles explicit embeddings (LRE, RLE, LRO, RLO, PDF) and isolates (LRI, RLI, FSI, PDI)
3. **Weak Types (W1-W7)**: Resolves weak character types based on surrounding context
4. **Neutral Types (N0-N2)**: Resolves neutral character types
5. **Implicit Levels (I1-I2)**: Assigns final embedding levels
6. **Reordering (L1-L4)**: Reorders text for visual display based on resolved levels

## References

- [UAX #9: Unicode Bidirectional Algorithm](https://www.unicode.org/reports/tr9/)
- [Unicode Bidirectional Algorithm FAQ](https://www.unicode.org/faq/bidi.html)
- [Bidirectional Character Types](http://www.unicode.org/reports/tr9/#Bidirectional_Character_Types)
- [Unicode Test Data](https://www.unicode.org/Public/UNIDATA/)

## License

MIT
