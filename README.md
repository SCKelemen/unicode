# unicode

[![CI](https://github.com/SCKelemen/unicode/workflows/CI/badge.svg)](https://github.com/SCKelemen/unicode/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/SCKelemen/unicode)](https://goreportcard.com/report/github.com/SCKelemen/unicode)

Implementations of various Unicode® Standard Annexes in Go.

This repository provides Go packages for Unicode text processing algorithms, organized by UAX (Unicode Standard Annex) specification.

## Packages

### [uax9](./uax9) - Bidirectional Algorithm

Implementation of UAX #9 (Unicode Bidirectional Algorithm) for handling bidirectional text with mixing LTR and RTL scripts.

**Status:** Complete with 100% conformance (513,494/513,494 tests passing)

Supports:
- **Full bidirectional text reordering** - Proper display of mixed LTR/RTL content
- **Isolating run sequences (BD13)** - Advanced context isolation for complex layouts
- **Explicit formatting characters** - LRE, RLE, LRO, RLO, PDF, LRI, RLI, FSI, PDI
- **Deep embedding nesting** - Up to 125 levels of explicit embedding
- **Bracket pair handling (N0)** - Proper neutral character resolution
- **Automatic direction detection** - Smart paragraph base direction

```go
import "github.com/SCKelemen/unicode/uax9"

// Reorder mixed LTR/RTL text
text := "Hello שלום world"
result := uax9.Reorder(text, uax9.DirectionLTR)

// Auto-detect paragraph direction
dir := uax9.GetParagraphDirection("שלום עולם")  // Returns DirectionRTL

// Get bidi class of a character
class := uax9.GetBidiClass('א')  // Returns R (Right-to-Left)
```

### [uax11](./uax11) - East Asian Width

Implementation of UAX #11 (East Asian Width) for determining character display width in East Asian typography contexts.

**Status:** Complete with comprehensive test coverage

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

### [uax14](./uax14) - Line Breaking Algorithm

Implementation of UAX #14 (Unicode Line Breaking Algorithm) for finding valid line break opportunities in text.

**Status:** Complete with 100% conformance (19,338/19,338 tests passing)

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

### [uax24](./uax24) - Script Property

Implementation of UAX #24 (Unicode Script Property) for identifying the writing system (script) to which a character belongs.

**Status:** Complete with 100% conformance (159,866/159,866 tests passing)

Supports:
- Script property lookup for all Unicode 17.0.0 characters
- 174 scripts including Latin, Greek, Cyrillic, Han, Arabic, Hebrew, and many others
- Mixed-script detection for security validation
- Common and Inherited script handling
- Single-script string validation

```go
import "github.com/SCKelemen/unicode/uax24"

// Get the script of a character
script := uax24.LookupScript('A')      // Returns ScriptLatin
script = uax24.LookupScript('中')      // Returns ScriptHan
script = uax24.LookupScript('5')       // Returns ScriptCommon

// Check if character belongs to a specific script
if uax24.IsLatin('A') {
    // Character is Latin
}

// Analyze a string for script composition
info := uax24.AnalyzeScripts("Hello мир")
fmt.Printf("Scripts: %v\n", info.Scripts)        // [Latin Cyrillic]
fmt.Printf("Mixed: %v\n", info.IsMixedScript)    // true

// Security: Detect homograph attacks
if !uax24.IsSingleScript("myVariаble") {  // 'а' is Cyrillic
    // Warning: Mixed scripts detected
}
```

### [uax29](./uax29) - Text Segmentation

Implementation of UAX #29 (Unicode Text Segmentation) for breaking text into grapheme clusters, words, and sentences.

**Status:** Complete with 100% conformance on all official Unicode tests

Supports:
- **Grapheme cluster boundaries** (100.0% - 766/766 tests)
  - User-perceived characters, emoji sequences, combining marks
  - Hangul syllable composition
  - Regional indicator pairs (flag emojis)
  - Indic conjunct sequences for 10+ scripts

- **Word boundaries** (100.0% - 1944/1944 tests)
  - Alphabetic and numeric sequences
  - Contractions, punctuation, hyphenated words
  - Hebrew letter handling, Katakana sequences
  - Emoji modifiers and ZWJ sequences

- **Sentence boundaries** (100.0% - 512/512 tests)
  - Period, question mark, exclamation handling
  - Abbreviation detection, quote and parenthesis handling
  - Multi-script sentence terminators

```go
import "github.com/SCKelemen/unicode/uax29"

// Grapheme clusters
graphemes := uax29.Graphemes("👨‍👩‍👧‍👦")  // Returns ["👨‍👩‍👧‍👦"]

// Words
words := uax29.Words("Hello, world!")  // Returns ["Hello", ",", " ", "world", "!"]

// Sentences
sentences := uax29.Sentences("Hello. World!")  // Returns ["Hello. ", "World!"]

// Single-pass API - get all three break types at once
breaks := uax29.FindAllBreaks("Hello, world!")
for _, pos := range breaks.Graphemes {
    // Process grapheme boundaries
}
for _, pos := range breaks.Words {
    // Process word boundaries
}
for _, pos := range breaks.Sentences {
    // Process sentence boundaries
}
```

### [uax31](./uax31) - Identifier and Pattern Syntax

Implementation of UAX #31 (Unicode Identifier and Pattern Syntax) for determining valid identifier characters in programming languages and pattern-based systems.

**Status:** Complete with 100% conformance (297,981/297,981 tests passing)

Supports:
- **XID_Start property** - Characters valid at the start of an identifier
  - Letters, ideographs, letter numbers across all scripts
  - Binary search for O(log n) lookups
- **XID_Continue property** - Characters valid after the first character
  - XID_Start plus marks, digits, connector punctuation
  - Includes zero-width joiner and combining marks
- **Pattern_Syntax property** - Reserved characters for pattern languages
  - ASCII punctuation and mathematical symbols
  - Used to identify syntactic elements
- **Pattern_White_Space property** - Whitespace in patterns
  - Spaces, tabs, line breaks for pattern tokenization
- **Default Identifier Syntax** - Complete identifier validation
  - Pattern: `<XID_Start> <XID_Continue>*`
  - Stable across Unicode versions

```go
import "github.com/SCKelemen/unicode/uax31"

// Check if character can start an identifier
if uax31.IsXIDStart('A') {
    // Valid identifier start (letters, ideographs)
}

// Check if character can continue an identifier
if uax31.IsXIDContinue('5') {
    // Valid after first character (includes digits, marks)
}

// Validate complete identifier
if uax31.IsValidIdentifier("myVar123") {
    // Valid: starts with letter, continues with letters/digits
}

// Pattern syntax detection
if uax31.IsPatternSyntax('*') {
    // Reserved for pattern languages (regex, etc.)
}

// Programming language tokenization example
func isIdentifierChar(r rune, isFirst bool) bool {
    if isFirst {
        return uax31.IsXIDStart(r)
    }
    return uax31.IsXIDContinue(r)
}

// Security: Validate identifiers for safety
identifier := "user_name"
if uax31.IsValidIdentifier(identifier) {
    // Identifier follows Unicode standard
}
```

### [uax50](./uax50) - Vertical Text Layout

Implementation of UAX #50 (Unicode Vertical Text Layout) for determining character orientation in vertical text.

**Status:** Complete with comprehensive test coverage

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

### [uts51](./uts51) - Unicode Emoji

Implementation of UTS #51 (Unicode Emoji) for emoji property detection, sequence validation, and terminal rendering support.

**Status:** Complete with 100% conformance (5,223/5,223 tests passing)

Supports:
- **Emoji properties** - All 6 core emoji properties
  - Emoji, Emoji_Presentation, Emoji_Modifier
  - Emoji_Modifier_Base, Emoji_Component, Extended_Pictographic
- **Sequence validation** - All emoji sequence types
  - ZWJ sequences (family emoji, etc.)
  - Modifier sequences (skin tones)
  - Flag sequences (regional indicators)
  - Keycap sequences (#️⃣, *️⃣, 0️⃣-9️⃣)
  - Tag sequences (subdivision flags)
- **Terminal rendering** - Width calculation for emoji display
- **Integration** with UAX #11, #14, #29, #50

```go
import "github.com/SCKelemen/unicode/uts51"

// Check if character is emoji
if uts51.IsEmoji('😀') {
    // Handle emoji
}

// Calculate width for terminal rendering
width := uts51.EmojiWidth('😀')  // Returns 2 (like CJK characters)

// Validate emoji sequences
sequence := []rune{0x1F468, 0x200D, 0x1F469, 0x200D, 0x1F467}  // Family
if uts51.IsValidEmojiSequence(sequence) {
    // Valid ZWJ sequence
}
```

### [uts15](./uts15) - Unicode Normalization Forms

Implementation of UTS #15 (Unicode Normalization Forms) for text normalization, comparison, and canonicalization.

**Status:** Complete with 100% conformance (20,034/20,034 tests passing)

Supports:
- **NFC (Canonical Composition)** - Recommended form for most uses
- **NFD (Canonical Decomposition)** - Fully decomposed form
- **NFKC (Compatibility Composition)** - Aggressive normalization for identifiers
- **NFKD (Compatibility Decomposition)** - Fully compatibility decomposed
- **Hangul composition/decomposition** - Algorithmic Hangul syllable handling
- **Canonical ordering** - Proper combining mark ordering
- **Normalization stability** - Idempotent operations
- Complete Unicode 17.0.0 normalization data

```go
import "github.com/SCKelemen/unicode/uts15"

// Normalize to NFC (recommended for most uses)
text := "café"  // May be composed or decomposed
normalized := uts15.NFC(text)

// Compare strings reliably
s1 := "café"  // Composed form
s2 := "cafe\u0301"  // Decomposed form (e + combining accent)
if uts15.NFC(s1) == uts15.NFC(s2) {
    // Strings are equivalent
}

// Normalize for searching (NFKC removes formatting distinctions)
query := "\uFB01le"  // Contains ﬁ ligature
normalized := uts15.NFKC(query)  // "file"

// Check if already normalized
if uts15.IsNFC("café") {
    // No normalization needed
}
```

### [uts39](./uts39) - Unicode Security Mechanisms

Implementation of UTS #39 (Unicode Security Mechanisms) for detecting and preventing security issues from confusable characters and mixed scripts.

**Status:** Complete with comprehensive test coverage

Supports:
- **Confusable detection** - Skeleton algorithm for visual similarity
  - Identifies lookalike characters (e.g., Cyrillic 'а' vs Latin 'a')
  - Case-insensitive confusable matching
  - 6,565 confusable mappings from Unicode 17.0.0
- **Mixed-script detection** - Identifies suspicious script mixing
  - Single-script, mixed-script, and cross-script analysis
  - Script-specific security policies
- **Restriction levels** - Security profiles for identifiers
  - ASCII-Only: Strictest, ASCII characters only
  - Single-Script: One script (excluding Common/Inherited)
  - Highly-Restrictive: Single script + Common + Inherited
  - Moderately-Restrictive: Multiple allowed script combinations
  - Minimally-Restrictive: Latin + one other script
  - Unrestricted: Any character combination
- **Safe identifier validation** - Checks for security issues
  - Invalid invisible characters
  - Proper identifier structure (UAX #31)
  - Minimum restriction level enforcement

```go
import "github.com/SCKelemen/unicode/uts39"

// Detect confusable strings (homograph attacks)
if uts39.AreConfusable("paypal", "pаypal") {  // Second uses Cyrillic 'а'
    // Warning: visually similar but different strings
}

// Get skeleton for comparison
skel := uts39.Skeleton("Hello")

// Check restriction level
level := uts39.GetRestrictionLevel("user_name")
if level >= uts39.HighlyRestrictive {
    // Identifier meets security requirements
}

// Detect mixed scripts
if uts39.IsMixedScript("hello мир") {  // Latin + Cyrillic
    // Warning: mixed script identifier
}

// Validate identifier safety
if uts39.IsSafeIdentifier("user_name") {
    // Safe: valid identifier, highly restrictive, no invisible chars
}

// Security validation example
func validateUsername(username string) error {
    if !uts39.IsValidIdentifier(username) {
        return errors.New("invalid identifier format")
    }

    level := uts39.GetRestrictionLevel(username)
    if level < uts39.HighlyRestrictive {
        return errors.New("username uses suspicious character mixing")
    }

    // Check against known legitimate usernames
    for _, legitimate := range knownUsernames {
        if uts39.AreConfusable(username, legitimate) {
            return errors.New("username too similar to existing user")
        }
    }

    return nil
}
```

### [uts51](./uts51) - Unicode Emoji

Implementation of UTS #51 (Unicode Emoji) for emoji property detection, sequence validation, and terminal rendering support.

**Status:** Complete with 100% conformance (5,223/5,223 tests passing)

Supports:
- **Emoji properties** - All 6 core emoji properties
  - Emoji, Emoji_Presentation, Emoji_Modifier
  - Emoji_Modifier_Base, Emoji_Component, Extended_Pictographic
- **Sequence validation** - All emoji sequence types
  - ZWJ sequences (family emoji, etc.)
  - Modifier sequences (skin tones)
  - Flag sequences (regional indicators)
  - Keycap sequences (#️⃣, *️⃣, 0️⃣-9️⃣)
  - Tag sequences (subdivision flags)
- **Terminal rendering** - Width calculation for emoji display
- **Integration** with UAX #11, #14, #29, #50

```go
import "github.com/SCKelemen/unicode/uts51"

// Check if character is emoji
if uts51.IsEmoji('😀') {
    // Handle emoji
}

// Calculate width for terminal rendering
width := uts51.EmojiWidth('😀')  // Returns 2 (like CJK characters)

// Validate emoji sequences
sequence := []rune{0x1F468, 0x200D, 0x1F469, 0x200D, 0x1F467}  // Family
if uts51.IsValidEmojiSequence(sequence) {
    // Valid ZWJ sequence
}
```

## Installation

```bash
go get github.com/SCKelemen/unicode/uax9
go get github.com/SCKelemen/unicode/uax11
go get github.com/SCKelemen/unicode/uax14
go get github.com/SCKelemen/unicode/uax24
go get github.com/SCKelemen/unicode/uax29
go get github.com/SCKelemen/unicode/uax31
go get github.com/SCKelemen/unicode/uax50
go get github.com/SCKelemen/unicode/uts15
go get github.com/SCKelemen/unicode/uts39
go get github.com/SCKelemen/unicode/uts51
```

## Design Philosophy

These implementations focus on practical text layout and rendering needs:
- Simple, focused APIs
- Minimal dependencies (standard library only)
- Performance-conscious
- Well-tested
- Layout-engine agnostic
- Full conformance with Unicode standards

## Version 2.0.0 Performance Improvements

Version 2.0.0 focuses on performance optimization while maintaining 100% conformance with Unicode standards.

### Table-Driven Binary Search

All packages now use **table-driven O(log n) binary search** for character classification, replacing sequential O(n) checks:

- **UAX #9**: Bidi class lookup optimized with 3,060 precomputed ranges from `DerivedBidiClass.txt`
- **UAX #29**: Unified packed data structure with 4,673 ranges encoding all three break types (grapheme, word, sentence) in 16-bit format

**Performance**: Character classification now runs at ~60-100 ns/op with 0 allocations on Apple M4 Pro.

### Generated Unicode Data

All Unicode property data is now generated directly from official Unicode 17.0.0 data files:
- Download from unicode.org during build
- Parse property files (`DerivedBidiClass.txt`, `GraphemeBreakProperty.txt`, etc.)
- Generate optimized Go code with binary search tables
- Ensures correctness and synchronization with Unicode standard

### Single-Pass API

UAX #29 provides a new `FindAllBreaks()` API that computes grapheme, word, and sentence boundaries in a single traversal:

```go
// Before: Three separate passes
graphemes := uax29.FindGraphemeBreaks(text)
words := uax29.FindWordBreaks(text)
sentences := uax29.FindSentenceBreaks(text)

// After: Single pass with shared classification
breaks := uax29.FindAllBreaks(text)
// Use breaks.Graphemes, breaks.Words, breaks.Sentences
```

This provides a convenient API for applications that need multiple break types, with framework in place for future hierarchical optimization.

## Version 3.0.0 Performance Improvements

Version 3.0.0 focuses on hierarchical optimization of the single-pass API introduced in v2.0.0.

### Hierarchical Break Detection

The `FindAllBreaks()` API now implements true hierarchical checking, leveraging the natural subset relationships between break types:

- **Words ⊆ Graphemes**: Word breaks only checked at grapheme cluster boundaries
- **Sentences ⊆ Words**: Sentence breaks only checked at word boundaries

This eliminates redundant checks and significantly improves performance for applications needing multiple break types.

### Performance Improvements

Benchmark results on Apple M4 Pro comparing v3.0.0 single-pass vs three separate function calls:

| Text Length | v2.0.0 Three Passes | v3.0.0 Single Pass | Speedup |
|-------------|--------------------|--------------------|---------|
| Short (33 chars) | 3,457 ns/op | 2,197 ns/op | **1.57x** |
| Medium (86 chars) | 16,191 ns/op | 9,636 ns/op | **1.68x** |
| Long (467 chars) | 423,491 ns/op | 188,982 ns/op | **2.24x** |

**Key benefits:**
- Speedup increases with text length (hierarchical pruning more effective on longer text)
- Single UTF-8 decode and classification pass
- Pre-classified data reused across all three break types
- No additional allocations compared to v2.0.0

### Maintained Conformance

100% conformance maintained on all official Unicode test suites:
- Grapheme: 766/766 tests passing
- Word: 1,944/1,944 tests passing
- Sentence: 512/512 tests passing

## Version 4.0.0 Performance Improvements

Version 4.0.0 focuses on code quality and maintainability through rule-based state machine architecture.

### Rule-Based State Machine Architecture

All break detection algorithms now use clean, rule-based implementations that directly map to the Unicode Standard specifications:

- **BreakContext abstractions**: `GraphemeBreakContext`, `WordBreakContext`, `SentenceBreakContext` provide clean navigation APIs
- **Named rule functions**: Each Unicode rule (GB3, WB5, SB8, etc.) becomes a named function with clear semantics
- **Declarative rule chains**: Rules checked in order with first-match-wins strategy
- **Maintained hierarchical optimization**: Words checked only at grapheme boundaries, sentences only at word boundaries

This architecture dramatically improves:
1. **Readability**: Rules directly match Unicode Standard specification
2. **Maintainability**: Easy to understand, modify, and extend
3. **Debuggability**: Each rule can be tested and traced independently

### Code Organization

New files implementing the rule-based architecture:
- `context.go` - Break context abstractions with navigation methods
- `grapheme_rules.go` - Grapheme breaking rules (ruleGB3 through ruleGB12_13)
- `word_rules.go` - Word breaking rules (ruleWB3 through ruleWB15_16)
- `sentence_rules.go` - Sentence breaking rules (ruleSB3 through ruleSB11)

### Performance Analysis

Benchmark results on Apple M4 Pro comparing v4.0.0 rule-based vs v3.0.0 inline:

**Single-Pass API:**
| Text Length | v3.0.0 Inline | v4.0.0 Rule-Based | Change |
|-------------|---------------|-------------------|---------|
| Short (33 chars) | 2,197 ns/op | 2,717 ns/op | 1.24x slower |
| Medium (86 chars) | 9,636 ns/op | 6,647 ns/op | **1.45x faster** |
| Long (467 chars) | 188,982 ns/op | 32,200 ns/op | **5.87x faster** |

**Rule-based grapheme breaking alone** (standalone function):
| Text Length | v3.0.0 Inline | v4.0.0 Rule-Based | Speedup |
|-------------|---------------|-------------------|---------|
| Short (33 chars) | 1,882 ns/op | 1,183 ns/op | **1.59x** |
| Medium (86 chars) | 8,759 ns/op | 3,041 ns/op | **2.88x** |
| Long (467 chars) | 168,060 ns/op | 15,170 ns/op | **11.08x** |

**Single-Pass vs Three Separate Passes (v4.0.0):**
| Text Length | Single Pass | Three Separate | Speedup |
|-------------|-------------|----------------|---------|
| Short (33 chars) | 2,717 ns/op | 3,380 ns/op | **1.24x** |
| Medium (86 chars) | 6,647 ns/op | 14,312 ns/op | **2.15x** |
| Long (467 chars) | 32,200 ns/op | 239,624 ns/op | **7.44x** |

**Key findings:**
- Rule-based grapheme breaking provides 1.6-11x speedup over inline implementation
- Performance improvements increase dramatically with text length
- Single-pass API maintains significant advantage over three separate calls
- Medium and long texts benefit most from rule-based architecture

### Maintained Conformance

100% conformance maintained on all official Unicode test suites:
- Grapheme: 766/766 tests passing
- Word: 1,944/1,944 tests passing
- Sentence: 512/512 tests passing

## Version 5.0.0 Improvements

Version 5.0.0 extends the rule-based state machine architecture from UAX #29 to UAX #14 (Line Breaking Algorithm), achieving 100% conformance and dramatically improved maintainability.

### Rule-Based Line Breaking Architecture

UAX #14 now uses a clean, rule-based implementation that directly maps to the Unicode Standard specification:

- **LineBreakContext abstraction**: Clean navigation API with helper methods (`SkipBackward`, `FindForward`, etc.)
- **Named rule functions**: Each Unicode rule (LB4, LB5, LB8, LB21, etc.) becomes a named function
- **Declarative rule chains**: Rules checked in order with first-match-wins strategy
- **Pair table fallback**: Common cases handled by efficient 2,064-entry lookup table

### Code Organization

New architecture improves code organization:
- **Original**: 1,112-line monolithic function with complex inline conditionals
- **Rule-based**: Isolated, independently testable rule functions with clear documentation
- **Complex rules decomposed**: LB21 (hyphen handling) and LB19 (quotation marks) broken into 7+ focused sub-rules

Key files:
- `context.go` - LineBreakContext abstraction with navigation methods
- `linebreak_rules.go` - Rule-based implementation (59 rule functions, 1,786 lines)
- Original monolithic implementation retained for comparison and fallback

### 100% Conformance Achievement

The rule-based implementation passes all official Unicode conformance tests:

**UAX #14 (Line Breaking)**: 19,338/19,338 tests passing (100.0%)

Key fixes for 100% conformance:
- **French guillemet separators**: `»word« pattern` (U+00AB/U+00BB) requiring special break handling
- **German quotes**: `„..."` and `‚...'` patterns where ClassQU_Pi acts as closing quote
- **Hebrew MAQAF**: HL × HH ÷ HL pattern for U+05BE hyphen
- **Regional indicators with combining marks**: RI × CM × RI sequences
- **Extended pictographic × emoji modifier**: Reserved emoji ranges (U+1F000-U+1FFFD)
- **Rule ordering**: Guillemet and German patterns must process before default quotation rules

### Performance Analysis

Benchmark results on Apple M4 Pro comparing rule-based vs original:

| Text Length | Original | Rule-Based | Change |
|-------------|----------|------------|---------|
| Short (10 chars) | 494 ns/op | 1,360 ns/op | 2.75x slower |
| Medium (64 chars) | 3,934 ns/op | 9,374 ns/op | 2.38x slower |
| Long (45 chars) | 2,138 ns/op | 5,209 ns/op | 2.44x slower |

**Trade-off analysis:**
- Rule-based implementation is 2-3x slower due to abstraction overhead
- Maintainability benefits are significant:
  - Isolated, testable rules directly mapping to spec
  - Clear documentation with spec links for each rule
  - Easy to add new rules without understanding entire state machine
  - Complex rules (LB21, LB19) broken into manageable sub-functions
- Performance acceptable for text layout applications (thousands of characters per millisecond)

### Benefits for Unicode Maintainability

The rule-based architecture provides critical benefits:

1. **Direct spec mapping**: Rule functions named after Unicode spec rules (ruleLB4, ruleLB21, etc.)
2. **Independent testing**: Each rule can be tested and traced independently
3. **Clear debugging**: Rule execution can be logged to understand break decisions
4. **Easy updates**: New Unicode versions can add rules without refactoring
5. **Reduced complexity**: No massive conditional chains or inline state tracking

This matches the successful pattern from UAX #29 v4.0.0, providing consistency across the codebase.

### Maintained Conformance

100% conformance maintained on all official Unicode test suites:
- Line Breaking: 19,338/19,338 tests passing

## Unicode Version

This repository implements **Unicode 17.0.0** (September 2024).

### Why Not Use Go's Standard Library?

Go's `unicode` package (as of Go 1.23) provides Unicode 15.0.0 data. While it includes some properties we need (e.g., `Regional_Indicator`, `Ideographic`, `Sentence_Terminal`), it is missing:

- **Emoji properties**: `Extended_Pictographic`, `Emoji`, `Emoji_Presentation`, `Emoji_Modifier`, `Emoji_Modifier_Base`, `Emoji_Component`
- **Text segmentation properties**: `Grapheme_Cluster_Break`, `Word_Break`, `Sentence_Break`
- **Layout properties**: `East_Asian_Width`, `Line_Break`, `Vertical_Orientation`

**Design Decision**: We implement all related properties within each specification package (e.g., all emoji properties in `uts51`) rather than mixing standard library and custom implementations. This ensures:

1. **Consistency**: All properties from a specification come from one authoritative source
2. **Completeness**: Unicode 17.0.0 support with the latest emoji and text handling
3. **Maintainability**: Single source of truth for each Unicode specification
4. **Testability**: 100% conformance against official Unicode 17.0.0 test files

When Go's `unicode` package updates to Unicode 17.0.0, we will continue maintaining our implementations to provide the specialized properties not available in the standard library.

## Conformance

All implementations follow the Unicode Standard and are tested against official Unicode conformance test suites where available:

### Test Coverage
- **UAX #9 (Bidirectional Algorithm)**: 100% conformance (513,494/513,494 tests)
  - All explicit embeddings and isolates
  - Multi-isolate sequences and deep nesting (up to 125 levels)
  - Empty isolate handling and overflow isolation
  - Bracket pair matching and neutral resolution
- **UAX #11 (East Asian Width)**: Comprehensive test coverage
  - Character width property lookup for all Unicode code points
  - Context-based ambiguous character resolution
  - Display width calculation for strings
  - Terminal emulator compatibility
- **UAX #14 (Line Breaking)**: 100% conformance (19,338/19,338 tests)
  - All line break classes and combining rules
  - Tailorable break opportunities
  - Complex script handling (CJK, Thai, etc.)
  - Hyphenation support (soft hyphens U+00AD)
- **UAX #24 (Script Property)**: 100% conformance (159,866/159,866 tests)
  - Script property lookup verified for all Unicode code points
  - All 174 scripts in Unicode 17.0.0
  - Mixed-script detection and analysis
  - Security validation for single-script strings
- **UAX #29 (Text Segmentation)**: 100% conformance (3,222/3,222 tests)
  - Grapheme cluster breaking: 766/766 tests
  - Word breaking: 1,944/1,944 tests
  - Sentence breaking: 512/512 tests
- **UAX #31 (Identifier and Pattern Syntax)**: 100% conformance (297,981/297,981 tests)
  - XID_Start (145,925 tests), XID_Continue (149,253 tests) properties verified
  - Pattern_Syntax (2,792 tests) and Pattern_White_Space (11 tests) properties verified
  - Default Identifier Syntax validation
  - Programming language identifier validation
- **UAX #50 (Vertical Text Layout)**: Comprehensive test coverage
  - Vertical orientation property for all Unicode code points
  - Glyph transformation detection
  - Base orientation determination
  - Mixed-script vertical layout support
- **UTS #15 (Unicode Normalization Forms)**: 100% conformance (20,034/20,034 tests)
  - All four normalization forms (NFC, NFD, NFKC, NFKD) verified
  - Hangul composition/decomposition, canonical ordering
  - Complex Indic script compositions
- **UTS #39 (Unicode Security Mechanisms)**: Comprehensive test coverage
  - Confusable detection with 6,565 mappings from Unicode 17.0.0
  - Skeleton algorithm for visual similarity detection
  - Mixed-script detection and restriction levels
  - Safe identifier validation with security policies
- **UTS #51 (Unicode Emoji)**: 100% conformance (5,223/5,223 tests)
  - All 6 emoji properties correctly implemented
  - Complete sequence validation (ZWJ, modifier, flag, keycap, tag sequences)

### Conformance Testing
Implementations are validated using the official Unicode Character Database (UCD) test files:
- [UAX #9 Test Files](https://www.unicode.org/Public/17.0.0/ucd/) - `BidiTest.txt` (513,494 tests), `BidiCharacterTest.txt`
- [UAX #11 Data Files](https://www.unicode.org/Public/17.0.0/ucd/) - `EastAsianWidth.txt` property data
- [UAX #14 Test Files](https://www.unicode.org/Public/17.0.0/ucd/auxiliary/) - `LineBreakTest.txt` (19,338 tests)
- [UAX #24 Data Files](https://www.unicode.org/Public/17.0.0/ucd/) - `Scripts.txt` (159,866 code point tests)
- [UAX #29 Test Files](https://www.unicode.org/Public/17.0.0/ucd/auxiliary/) - `GraphemeBreakTest.txt`, `WordBreakTest.txt`, `SentenceBreakTest.txt`
- [UAX #31 Data Files](https://www.unicode.org/Public/17.0.0/ucd/) - `DerivedCoreProperties.txt`, `PropList.txt` (297,981 property tests)
- [UAX #50 Data Files](https://www.unicode.org/Public/17.0.0/ucd/) - `VerticalOrientation.txt` property data
- [UTS #39 Data Files](https://www.unicode.org/Public/security/latest/) - `confusables.txt` (6,565 mappings), `IdentifierStatus.txt`
- [UTS #51 Test Files](https://www.unicode.org/Public/emoji/17.0/) - `emoji-test.txt` with 5,223 test cases
- [Unicode Character Database](https://www.unicode.org/Public/17.0.0/ucd/) - Character property data files

The implementations follow the conformance model described in [UTR #33: Unicode Conformance Model](https://www.unicode.org/reports/tr33/), which defines what it means to conform to Unicode Standard specifications.

## Related Projects

- [github.com/SCKelemen/layout](https://github.com/SCKelemen/layout) - Text layout engine using these UAX implementations

## References

### Metastandards
- [UTR #33: Unicode Conformance Model](https://www.unicode.org/reports/tr33/) - Defines conformance requirements for Unicode Standard implementations
- [UAX #41: Common References for Unicode Standard Annexes](https://www.unicode.org/reports/tr41/) - Common definitions and references used across Unicode Standard Annexes

### Implemented Standards
- [Unicode Standard Annexes](https://www.unicode.org/reports/)
- [UAX #9: Bidirectional Algorithm](https://www.unicode.org/reports/tr9/)
- [UAX #11: East Asian Width](https://www.unicode.org/reports/tr11/)
- [UAX #14: Line Breaking](https://www.unicode.org/reports/tr14/)
- [UAX #24: Script Property](https://www.unicode.org/reports/tr24/)
- [UAX #29: Text Segmentation](https://www.unicode.org/reports/tr29/)
- [UAX #31: Identifier and Pattern Syntax](https://www.unicode.org/reports/tr31/)
- [UAX #50: Vertical Text Layout](https://www.unicode.org/reports/tr50/)
- [UTS #15: Unicode Normalization Forms](https://www.unicode.org/reports/tr15/)
- [UTS #39: Unicode Security Mechanisms](https://www.unicode.org/reports/tr39/)
- [UTS #51: Unicode Emoji](https://www.unicode.org/reports/tr51/)

## License

🐻 BearWare 1.0 🐻

MIT License with bear emojis. See [LICENSE](LICENSE) file for details.
