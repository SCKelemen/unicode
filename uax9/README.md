# uax9 - Unicode Bidirectional Algorithm

Implementation of [UAX #9: Unicode Bidirectional Algorithm](https://www.unicode.org/reports/tr9/) in Go.

**Status:** Highly Conformant (99.995% pass rate on official Unicode test vectors with full isolating run sequences)

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
- **Passed**: 513,470
- **Pass rate**: 99.995%
- **Failed**: 24 (multi-isolate sequence edge cases)

The test suite includes:
- Official Unicode BidiTest.txt (513,494 test cases)
- BidiCharacterTest.txt for character-specific cases
- Custom unit tests for common use cases

## Known Limitations

- 0.005% of edge cases fail (24 out of 513,494 tests)
- All failures involve **multi-isolate sequences**: Multiple consecutive non-empty isolates where formatting characters need sophisticated context analysis
  - Pattern: `R ON FSI L PDI LRI L PDI RLI L PDI ON R` with 3 consecutive isolates
  - Pattern: `L LRI R PDI FSI R PDI RLI R PDI R` with mixed content directions
  - Challenge: Each isolate's formatting characters should inherit the outer context level (R/ON at level 1), but current implementation finds the content inside adjacent isolates (L at level 2)
  - Solution would require: Skip entire isolate sequences (not just formatting chars) when finding context, OR process isolates in multiple passes with dependency tracking

**Achievement**: All 10 deep embedding nesting failures (30-64 levels) have been fixed ✅

The implementation uses full isolating run sequences (BD13) as specified in UAX#9 and is production-ready, handling 99.995% of Unicode's comprehensive test cases including all common real-world bidirectional text scenarios. The remaining 24 failures are rare edge cases involving pathological sequences of multiple consecutive isolates that would not occur in natural text.

## Implementation Details

The implementation follows the UAX #9 specification with full isolating run sequences support:

1. **Character Classification**: Maps Unicode characters to their bidirectional types (L, R, AL, EN, etc.)
2. **Explicit Levels (X1-X8)**: Handles explicit embeddings (LRE, RLE, LRO, RLO, PDF) and isolates (LRI, RLI, FSI, PDI)
3. **Isolate Matching (BD9)**: Tracks matching isolate initiators and PDIs
4. **Level Runs**: Identifies maximal sequences at the same embedding level
5. **Isolating Run Sequences (BD13)**: Builds sequences connected through isolates for proper context resolution
6. **Weak Types (W1-W7)**: Resolves weak character types within isolating run sequences
7. **Neutral Types (N0-N2)**: Resolves neutral character types with proper sos/eos from sequences
8. **Implicit Levels (I1-I2)**: Assigns final embedding levels within sequences
9. **Reordering (L1-L4)**: Reorders text for visual display based on resolved levels

## References

- [UAX #9: Unicode Bidirectional Algorithm](https://www.unicode.org/reports/tr9/)
- [Unicode Bidirectional Algorithm FAQ](https://www.unicode.org/faq/bidi.html)
- [Bidirectional Character Types](http://www.unicode.org/reports/tr9/#Bidirectional_Character_Types)
- [Unicode Test Data](https://www.unicode.org/Public/UNIDATA/)

## License

MIT
