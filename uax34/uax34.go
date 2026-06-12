// Package uax34 implements Unicode Named Character Sequences (UAX #34).
//
// A Named Character Sequence assigns a unique name to a sequence of two or
// more code points. Names use the same restricted character set as Unicode
// character names (A-Z, 0-9, space, and hyphen) and are guaranteed unique.
// For example, the name "KEYCAP NUMBER SIGN" denotes the three code point
// sequence U+0023 U+FE0F U+20E3.
//
// Implements UAX #34: Unicode Named Character Sequences
// https://www.unicode.org/reports/tr34/
//
// # Data Source
//
// The data is derived from the normative NamedSequences.txt file in the
// Unicode Character Database. The provisional NamedSequencesProv.txt file is
// intentionally excluded, as it is not normative.
//
// This implementation uses Unicode 17.0.0 data:
//   - https://www.unicode.org/Public/17.0.0/ucd/NamedSequences.txt
//
// # Semantics
//
// In NamedSequences.txt every name is unique and, for a given Unicode version,
// each code point sequence maps to exactly one name. Lookup and Name therefore
// form a bijection between names and sequences. The generator verifies this
// invariant when the table is built, failing loudly if a future data file
// ever violates it.
//
// # Usage
//
//	import "github.com/SCKelemen/unicode/v6/uax34"
//
//	// Resolve a name to its code point sequence.
//	seq, ok := uax34.Lookup("KEYCAP NUMBER SIGN") // []rune{0x0023, 0xFE0F, 0x20E3}, true
//
//	// Resolve a code point sequence back to its name.
//	name, ok := uax34.Name([]rune{0x0023, 0xFE0F, 0x20E3}) // "KEYCAP NUMBER SIGN", true
//
//	// Total number of named sequences.
//	n := uax34.Count()
//
//	// Iterate over every named sequence, sorted by name.
//	for _, s := range uax34.All() {
//	    fmt.Println(s.Name, s.Runes)
//	}
//
// # References
//
//   - UAX #34: https://www.unicode.org/reports/tr34/
//   - NamedSequences.txt: https://www.unicode.org/Public/17.0.0/ucd/NamedSequences.txt
//   - UAX #44 (Unicode Character Database): https://www.unicode.org/reports/tr44/
package uax34

import (
	"slices"
	"sync"
)

// Sequence is a single Named Character Sequence: a unique name paired with the
// sequence of two or more code points it denotes.
type Sequence struct {
	// Name is the unique name of the sequence (A-Z, 0-9, space, hyphen).
	Name string
	// Runes is the code point sequence the name denotes.
	Runes []rune
}

// nameIndex lazily maps each code point sequence (encoded as a string of its
// runes) to its unique name, supporting reverse lookups in Name.
var (
	nameIndexOnce sync.Once
	nameIndex     map[string]string
)

func buildNameIndex() {
	nameIndex = make(map[string]string, len(namedSequenceData))
	for _, e := range namedSequenceData {
		nameIndex[string(e.seq)] = e.name
	}
}

// Lookup returns the code point sequence for the named character sequence with
// the given name. The returned slice is a copy that the caller may modify.
// ok reports whether name is a defined Named Character Sequence.
func Lookup(name string) (seq []rune, ok bool) {
	i, found := slices.BinarySearchFunc(namedSequenceData, name,
		func(e namedSequence, target string) int {
			switch {
			case e.name < target:
				return -1
			case e.name > target:
				return 1
			default:
				return 0
			}
		})
	if !found {
		return nil, false
	}
	return slices.Clone(namedSequenceData[i].seq), true
}

// Name returns the unique name of the given code point sequence. ok reports
// whether seq is exactly a defined Named Character Sequence; partial matches
// and supersequences do not match.
func Name(seq []rune) (name string, ok bool) {
	nameIndexOnce.Do(buildNameIndex)
	name, ok = nameIndex[string(seq)]
	return name, ok
}

// Count returns the number of Named Character Sequences.
func Count() int {
	return len(namedSequenceData)
}

// All returns every Named Character Sequence, sorted ascending by name. The
// returned slices are copies that the caller may modify.
func All() []Sequence {
	out := make([]Sequence, len(namedSequenceData))
	for i, e := range namedSequenceData {
		out[i] = Sequence{Name: e.name, Runes: slices.Clone(e.seq)}
	}
	return out
}
