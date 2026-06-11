# uax9 - Unicode Bidirectional Algorithm

Implementation of [UAX #9: Unicode Bidirectional Algorithm](https://www.unicode.org/reports/tr9/) in Go.

**Status:** Full official-test conformance on the bundled Unicode 17.0.0 data.
- BidiTest.txt (synthetic class sequences): **513,494 / 513,494 (100%)** ✅
- BidiCharacterTest.txt (real code-point sequences):
  - Paragraph direction (P2/P3): **91,707 / 91,707 (100%)** ✅
  - Resolved levels: **91,707 / 91,707 (100%)** ✅

Historically sensitive areas were paired-bracket resolution (N0) and isolate-formatting level adjustment inside nested explicit-formatting scopes; both are now covered by the official conformance suite and targeted regression tests.

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
go get github.com/SCKelemen/unicode/v6/uax9
```

## Usage

```go
package main

import (
	"fmt"
	"github.com/SCKelemen/unicode/v6/uax9"
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

- **BidiTest.txt** (Unicode 17.0.0): 513,494 / 513,494 passed (100.0%) ✅
- **BidiCharacterTest.txt** (Unicode 17.0.0): 91,707 test cases
  - Paragraph direction: 91,707 / 91,707 (100%) ✅
  - Resolved levels: 91,707 / 91,707 (100%) ✅

The test suite includes:
- Official Unicode BidiTest.txt (513,494 test cases) - **ALL PASSING** ✅
- Official Unicode BidiCharacterTest.txt (91,707 test cases) - **ALL PASSING** ✅
- Focused BD16/N0 regression tests for paired brackets and nested bracket cases
- Custom unit tests for common use cases

## Conformance Achievements

- ✅ **100% conformance** on BidiTest.txt (513,494 synthetic-class vectors)
- ✅ **100% conformance** on BidiCharacterTest.txt paragraph direction (91,707 real code-point vectors)
- ✅ **100% conformance** on BidiCharacterTest.txt resolved levels (91,707 real code-point vectors)
- ✅ **Paired-bracket resolution (N0)**: BD16 bracket matching + N0 resolution over isolating run sequences
- ✅ **Multi-isolate sequences**: Advanced context discovery that skips entire isolate sequences
- ✅ **Deep embedding nesting**: Correct handling of extreme nesting depths (up to 125 levels)
- ✅ **Overflow isolation**: Proper tracking of overflow embeddings inside/outside overflow isolates
- ✅ **Empty isolate adjustment**: Direction-sensitive level assignment for empty and matched isolate formatting characters in nested explicit-formatting scopes

The implementation uses full isolating run sequences (BD13) as specified in UAX#9 and handles all edge cases including pathological sequences that would rarely occur in natural text.

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
