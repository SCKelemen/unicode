# uax31 - Unicode Identifier and Pattern Syntax

[![Go Reference](https://pkg.go.dev/badge/github.com/SCKelemen/unicode/uax31.svg)](https://pkg.go.dev/github.com/SCKelemen/unicode/uax31)

Implementation of UAX #31 (Unicode Identifier and Pattern Syntax) for determining which Unicode characters can appear in identifiers and pattern syntax.

**Status:** Complete with comprehensive test coverage

## Overview

UAX #31 defines Unicode properties for identifier characters and pattern syntax characters. These properties are fundamental for:

- **Programming language identifiers** (variable names, function names, etc.)
- **Security identifiers** (usernames, domain names, etc.)
- **Pattern-based syntax** (regular expressions, query languages)
- **Text processing and parsing**

Based on: [UAX #31](https://www.unicode.org/reports/tr31/)

## Properties

### XID_Start
Characters valid at the start of an identifier:
- Unicode letters, ideographs, letter numbers
- Excludes Pattern_Syntax and Pattern_White_Space
- **Examples:** `A`, `z`, `中`, `α`, `א`

### XID_Continue
Characters valid after the first character in an identifier:
- All XID_Start characters
- Plus: nonspacing marks, spacing marks, decimal numbers
- Connector punctuation (like `_`) and a few other categories
- **Examples:** `A`, `5`, `_`, `́` (combining acute)

### Pattern_Syntax
Characters reserved for use in patterns and syntax:
- ASCII punctuation and mathematical symbols
- Used to identify syntactic elements in pattern languages
- **Examples:** `!`, `*`, `+`, `(`, `)`, `{`, `}`

### Pattern_White_Space
Characters treated as whitespace in patterns:
- Spaces, tabs, line breaks, form feeds
- Used for pattern tokenization
- **Examples:** ` ` (space), `\t`, `\n`, `\r`

## Installation

```bash
go get github.com/SCKelemen/unicode/uax31
```

## Usage

### Property Checks

```go
import "github.com/SCKelemen/unicode/uax31"

// Check if character can start an identifier
if uax31.IsXIDStart('A') {
    // Valid identifier start
}

// Check if character can continue an identifier
if uax31.IsXIDContinue('5') {
    // Valid in identifier (after first character)
}

// Check if character is pattern syntax
if uax31.IsPatternSyntax('*') {
    // Character is syntax, not identifier content
}

// Check if character is pattern whitespace
if uax31.IsPatternWhiteSpace(' ') {
    // Character is whitespace in patterns
}
```

### Identifier Validation

```go
// Validate complete identifier
if uax31.IsValidIdentifier("myVar123") {
    // Valid identifier
}

// Unicode identifiers
uax31.IsValidIdentifier("变量")        // true (Chinese)
uax31.IsValidIdentifier("переменная") // true (Russian)
uax31.IsValidIdentifier("μετβλητή")   // true (Greek)

// Invalid identifiers
uax31.IsValidIdentifier("123var")   // false (starts with digit)
uax31.IsValidIdentifier("_private") // false (underscore not XID_Start)
uax31.IsValidIdentifier("my-var")   // false (hyphen not in XID_Continue)
```

## Default Identifier Syntax

The XID_Start and XID_Continue properties define **Default Identifiers**, which are stable across Unicode versions:

```
<Identifier> := <XID_Start> <XID_Continue>*
```

This means:
- First character must have the XID_Start property
- Subsequent characters must have the XID_Continue property
- Empty strings are not valid identifiers

## Important Notes

### Underscore is not XID_Start

According to Unicode, underscore (`_`, U+005F) is **XID_Continue but not XID_Start**. This means:

```go
uax31.IsValidIdentifier("_private") // false - cannot start with underscore
uax31.IsValidIdentifier("my_var")   // true  - underscore in middle is OK
```

Many programming languages (Python, C, JavaScript, etc.) extend the default identifier syntax to allow leading underscores. If you need this behavior, you can modify the validation:

```go
func IsValidProgrammingIdentifier(s string) bool {
    if len(s) == 0 {
        return false
    }
    runes := []rune(s)
    // Allow underscore or XID_Start at the beginning
    if !(runes[0] == '_' || uax31.IsXIDStart(runes[0])) {
        return false
    }
    // Rest must be XID_Continue
    for _, r := range runes[1:] {
        if !uax31.IsXIDContinue(r) {
            return false
        }
    }
    return true
}
```

## Performance

Optimized with binary search over sorted ranges (O(log n) lookups):

```
BenchmarkIsXIDStart-14                   207337023    5.731 ns/op    0 B/op    0 allocs/op
BenchmarkIsXIDContinue-14                176693479    6.582 ns/op    0 B/op    0 allocs/op
BenchmarkIsPatternSyntax-14              307636860    4.040 ns/op    0 B/op    0 allocs/op
BenchmarkIsValidIdentifier-14             21146242   55.83  ns/op    0 B/op    0 allocs/op
BenchmarkIsValidIdentifier_Unicode-14     12188190   98.25  ns/op    0 B/op    0 allocs/op
```

Tested on Apple M4 Pro.

## Data Generation

The package uses generated data from official Unicode 17.0.0 files:

- [DerivedCoreProperties.txt](https://www.unicode.org/Public/17.0.0/ucd/DerivedCoreProperties.txt) - XID_Start, XID_Continue
- [PropList.txt](https://www.unicode.org/Public/17.0.0/ucd/PropList.txt) - Pattern_Syntax, Pattern_White_Space

To regenerate data (after Unicode updates):

```bash
cd uax31
go run generate_identifier_data.go
```

## Implementation Details

- **XID_Start:** 779 ranges
- **XID_Continue:** 1,422 ranges
- **Pattern_Syntax:** 255 ranges
- **Pattern_White_Space:** 6 ranges

Binary search performs approximately 10-11 comparisons for XID lookups.

## References

- [UAX #31: Unicode Identifier and Pattern Syntax](https://www.unicode.org/reports/tr31/)
- [§2.3 Default Identifier Syntax](https://www.unicode.org/reports/tr31/#Default_Identifier_Syntax)
- [§3 Pattern Syntax](https://www.unicode.org/reports/tr31/#Pattern_Syntax)
- [DerivedCoreProperties.txt](https://www.unicode.org/Public/17.0.0/ucd/DerivedCoreProperties.txt)
- [PropList.txt](https://www.unicode.org/Public/17.0.0/ucd/PropList.txt)

## License

MIT
