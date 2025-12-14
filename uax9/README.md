# uax9 - Unicode Bidirectional Algorithm

Implementation of [UAX #9: Unicode Bidirectional Algorithm](https://www.unicode.org/reports/tr9/) in Go.

**Status:** Not yet implemented

## Overview

This package will provide bidirectional text reordering for proper display of text containing both left-to-right (LTR) and right-to-left (RTL) scripts.

## Planned Features

- Bidirectional text reordering
- Support for mixing LTR and RTL scripts
- Explicit formatting characters (LRE, RLE, LRO, RLO, PDF, LRI, RLI, FSI, PDI)
- Automatic base direction detection
- Bracket pair handling
- Mirror glyph support

## Use Cases

- Rendering Arabic or Hebrew text mixed with Latin
- Text editors with bidirectional text support
- Terminal UIs displaying mixed-direction content
- Layout engines requiring proper text ordering

## Implementation Plan

1. Bidi character type classification
2. Level resolution algorithm
3. Explicit formatting character handling
4. Reordering for display
5. Bracket pair matching
6. Mirror glyph detection

## References

- [UAX #9: Unicode Bidirectional Algorithm](https://www.unicode.org/reports/tr9/)
- [Unicode Bidirectional Algorithm FAQ](https://www.unicode.org/faq/bidi.html)
- [Bidirectional Character Types](http://www.unicode.org/reports/tr9/#Bidirectional_Character_Types)

## License

MIT
