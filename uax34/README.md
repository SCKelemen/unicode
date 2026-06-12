# UAX #34: Unicode Named Character Sequences

Implementation of [UAX #34 (Unicode Named Character Sequences)](https://www.unicode.org/reports/tr34/) for resolving the unique names assigned to sequences of two or more code points.

## Overview

A *Named Character Sequence* assigns a unique name to a sequence of two or more code points. Names use the same restricted character set as Unicode character names (`A-Z`, `0-9`, space, and hyphen). For example, the name `KEYCAP NUMBER SIGN` denotes the three code point sequence `U+0023 U+FE0F U+20E3`.

This package provides forward lookup (name → sequence), reverse lookup (sequence → name), and iteration over the full set, backed by a generated table from the normative `NamedSequences.txt`.

## Installation

```bash
go get github.com/SCKelemen/unicode/v6/uax34
```

## Quick Start

```go
package main

import (
    "fmt"

    "github.com/SCKelemen/unicode/v6/uax34"
)

func main() {
    // Resolve a name to its code point sequence.
    seq, ok := uax34.Lookup("KEYCAP NUMBER SIGN")
    fmt.Println(seq, ok) // [35 65039 8419] true

    // Resolve a code point sequence back to its name.
    name, ok := uax34.Name([]rune{0x0023, 0xFE0F, 0x20E3})
    fmt.Println(name, ok) // KEYCAP NUMBER SIGN true

    // Total number of named sequences.
    fmt.Println(uax34.Count()) // 461

    // Iterate over every named sequence, sorted by name.
    for _, s := range uax34.All() {
        fmt.Println(s.Name, s.Runes)
    }
}
```

## API Reference

#### `Lookup(name string) (seq []rune, ok bool)`
Returns the code point sequence for the named character sequence with the given name. The returned slice is a copy. `ok` reports whether `name` is a defined Named Character Sequence.

#### `Name(seq []rune) (name string, ok bool)`
Returns the unique name of the given code point sequence. The match must be exact: proper prefixes and supersequences do not match.

#### `Count() int`
Returns the number of Named Character Sequences.

#### `All() []Sequence`
Returns every Named Character Sequence, sorted ascending by name. The returned slices are copies.

```go
type Sequence struct {
    Name  string // unique name (A-Z, 0-9, space, hyphen)
    Runes []rune // the code point sequence the name denotes
}
```

## Semantics

In `NamedSequences.txt` every name is unique and, for a given Unicode version, each code point sequence maps to exactly one name. `Lookup` and `Name` therefore form a bijection between names and sequences. The generator verifies this invariant when the table is built and fails loudly if a future data file ever violates it.

The provisional file `NamedSequencesProv.txt` is **not** normative and is intentionally excluded.

## Conformance

This implementation follows UAX #34 and is built from Unicode 17.0.0 data:

- Data source: [`NamedSequences.txt`](https://www.unicode.org/Public/17.0.0/ucd/NamedSequences.txt) (461 named sequences)
- `TestDataFileConsistency` validates the generated table against the committed `NamedSequences.txt`.
- `TestOfficialNamedSequences` downloads the official file and verifies every entry round-trips through `Lookup` and `Name`.

## Code Generation

The `named_sequences_data.go` file is generated from Unicode data:

```bash
cd uax34
go run generate_named_sequences.go
```

The generator is version-pinned to Unicode 17.0.0, produces deterministic (name-sorted) output, and formats the result with `go/format`. It has no external dependencies.

## References

- [UAX #34: Unicode Named Character Sequences](https://www.unicode.org/reports/tr34/)
- [NamedSequences.txt](https://www.unicode.org/Public/17.0.0/ucd/NamedSequences.txt)
- [UAX #44: Unicode Character Database](https://www.unicode.org/reports/tr44/)
