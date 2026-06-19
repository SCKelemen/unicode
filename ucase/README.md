# ucase: Unicode Default Case Mapping & Folding

Implementation of the Unicode **default (language-insensitive) case algorithms** — case folding, lowercasing, uppercasing, and titlecasing — plus the `Cased`/`Case_Ignorable` properties and default caseless matching.

## Why `ucase` and not `uaxNN`/`utsNN`?

The rest of this repository follows the `uaxNN`/`utsNN` naming because those packages implement *numbered* Unicode annexes. **Case mapping is not a numbered annex.** It is specified in the core Unicode Standard, **Section 3.13 "Default Case Algorithms"** (with the caseless-matching definitions D144–D147), backed by the UCD data files. There is no "UAX #21." The package is therefore named for its function: `ucase` ("Unicode case").

## Overview

Section 3.13 defines four operations and two derived properties:

| Operation | Function | Notes |
|-----------|----------|-------|
| `toCasefold` | `ToFold` | Full folding (`CaseFolding.txt` status C + F) |
| `toLowercase` | `ToLower` | Full (1:many); applies the `Final_Sigma` context |
| `toUppercase` | `ToUpper` | Full (1:many); e.g. `ß` → `SS` |
| `toTitlecase` | `ToTitle` | Titlecase the first cased char of each word, lowercase the rest |

Properties: `IsCased` (D135) and `IsCaseIgnorable` (D136). Default caseless match (D144): `CaselessMatch`.

## Installation

```bash
go get github.com/SCKelemen/unicode/v6/ucase
```

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/SCKelemen/unicode/v6/ucase"
)

func main() {
	fmt.Println(ucase.ToLower("ΟΔΟΣ"))                 // οδος   (final sigma)
	fmt.Println(ucase.ToUpper("straße"))                // STRASSE
	fmt.Println(ucase.ToTitle("hello, world"))          // Hello, World
	fmt.Println(ucase.ToFold("eﬀort"))                  // effort
	fmt.Println(ucase.CaselessMatch("Straße", "STRASSE")) // true
}
```

## API Reference

#### String operations (full, 1:many)
- `ToFold(s string) string` — default full case folding (toCasefold).
- `ToLower(s string) string` — default full lowercasing, including `Final_Sigma`.
- `ToUpper(s string) string` — default full uppercasing.
- `ToTitle(s string) string` — default titlecasing (see word boundaries below).

#### Rune operations (simple, 1:1)
- `SimpleFold(r rune) rune` — simple case fold (status C + S). *Distinct from `unicode.SimpleFold`.*
- `SimpleLower(r rune) rune`, `SimpleUpper(r rune) rune`, `SimpleTitle(r rune) rune`.

#### Properties & matching
- `IsCased(r rune) bool` — Cased property (D135).
- `IsCaseIgnorable(r rune) bool` — Case_Ignorable property (D136).
- `CaselessMatch(a, b string) bool` — default caseless match (D144): `ToFold(a) == ToFold(b)`.

#### Metadata
- `UnicodeVersion` — the Unicode version of the generated tables (`"17.0.0"`).

## Semantics & Scope

### Titlecasing word boundaries
`ToTitle` titlecases the first cased character of each word and lowercases all
other cased characters. A word begins at the first cased character following the
start of the string or a character that is neither cased nor case-ignorable
(whitespace, punctuation, symbols). Case-ignorable characters (apostrophes,
combining marks) do not break words, so `"don't stop"` → `"Don't Stop"`. This is
a documented, self-contained boundary definition; it is not the full
[UAX #29](https://www.unicode.org/reports/tr29/) word segmentation.

### Locale tailoring — out of scope (future work)
Only the **default, language-insensitive** algorithm is implemented. The
Turkish/Azeri (`tr`/`az`) and Lithuanian (`lt`) tailorings in
`SpecialCasing.txt` are **not** applied. Consequently the only casing context
condition the default algorithm exercises is `Final_Sigma`; the others
(`After_Soft_Dotted`, `More_Above`, `Before_Dot`, `After_I`) appear exclusively
on language-tagged lines and will be added with locale tailoring.

### Caseless matching
`CaselessMatch` is the **default** caseless match (D144). It does **not** apply
canonical normalization, so it is not the **canonical** caseless match (D145),
which requires NFD. To compare strings that may differ in canonical composition,
normalize both (e.g. with
[`uts15`](https://github.com/SCKelemen/unicode/tree/main/uts15)) before calling.

## Conformance

Built from Unicode 17.0.0 UCD data:

- [`UnicodeData.txt`](https://www.unicode.org/Public/17.0.0/ucd/UnicodeData.txt) — simple mappings (fields 12/13/14)
- [`SpecialCasing.txt`](https://www.unicode.org/Public/17.0.0/ucd/SpecialCasing.txt) — full + conditional mappings
- [`CaseFolding.txt`](https://www.unicode.org/Public/17.0.0/ucd/CaseFolding.txt) — case folding (C/F/S/T)
- [`DerivedCoreProperties.txt`](https://www.unicode.org/Public/17.0.0/ucd/DerivedCoreProperties.txt) — `Cased`, `Case_Ignorable`

- `TestDataFileConsistency` checks the generated tables against the committed UCD files (every default mapping).
- `TestOfficialConformance` downloads the official files and re-verifies (skipped when offline or in `-short`).

## Code Generation

```bash
cd ucase
go run generate_ucase.go
```

The generator is version-pinned to Unicode 17.0.0, produces deterministic
(code-point-sorted) output, formats with `go/format`, and skips
language-sensitive `SpecialCasing.txt` lines. Zero external dependencies.

## References

- The Unicode Standard, **Section 3.13 "Default Case Algorithms"** — <https://www.unicode.org/versions/latest/>
- [UAX #44: Unicode Character Database](https://www.unicode.org/reports/tr44/)
- [UCD 17.0.0](https://www.unicode.org/Public/17.0.0/ucd/)
