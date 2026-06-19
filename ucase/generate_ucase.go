//go:build ignore
// +build ignore

// This program generates ucase_data.go from the Unicode Character Database.
// Run with: go run generate_ucase.go
//
// Case mapping is specified in the Unicode core standard, Section 3.13
// "Default Case Algorithms" (NOT in any numbered annex). The data is built
// from four UCD files for Unicode 17.0.0:
//
//   - UnicodeData.txt          simple (1:1) mappings, fields 12/13/14
//     (Simple_Uppercase/Lowercase/Titlecase_Mapping)
//   - SpecialCasing.txt        full (1:many) and conditional mappings
//   - CaseFolding.txt          case folding (status C/F/S/T)
//   - DerivedCoreProperties.txt  Cased and Case_Ignorable properties
//
// Only the language-insensitive (default) data is emitted: SpecialCasing.txt
// lines carrying a language tag (lt, tr, az) are skipped, leaving the
// unconditional mappings plus the single non-language conditional mapping,
// Final_Sigma (U+03A3).
//
// References:
//   - The Unicode Standard, Section 3.13 "Default Case Algorithms"
//     https://www.unicode.org/versions/Unicode17.0.0/
//   - UAX #44 (Unicode Character Database): https://www.unicode.org/reports/tr44/
//   - https://www.unicode.org/Public/17.0.0/ucd/
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

const unicodeVersion = "17.0.0"

const (
	unicodeDataURL      = "https://www.unicode.org/Public/17.0.0/ucd/UnicodeData.txt"
	specialCasingURL    = "https://www.unicode.org/Public/17.0.0/ucd/SpecialCasing.txt"
	caseFoldingURL      = "https://www.unicode.org/Public/17.0.0/ucd/CaseFolding.txt"
	derivedCorePropsURL = "https://www.unicode.org/Public/17.0.0/ucd/DerivedCoreProperties.txt"
)

// knownContexts lists the casing-context tokens defined in Section 3.13. Any
// other token in a SpecialCasing condition list is a BCP 47 language tag, which
// marks a language-sensitive (out-of-scope) mapping.
var knownContexts = map[string]bool{
	"Final_Sigma":       true,
	"Not_Final_Sigma":   true,
	"After_Soft_Dotted": true,
	"More_Above":        true,
	"Before_Dot":        true,
	"Not_Before_Dot":    true,
	"After_I":           true,
}

type simpleCaseEntry struct {
	cp, up, lo, ti rune
}

type specialCaseEntry struct {
	cp                  rune
	lower, title, upper []rune
	cond                string // "", "Final_Sigma"
}

type foldEntry struct {
	cp     rune
	target string
}

type runePair struct {
	cp, to rune
}

type runeRange struct {
	lo, hi rune
}

func main() {
	simple := parseUnicodeData(fetch(unicodeDataURL))
	fmt.Printf("Parsed %d simple case mappings\n", len(simple))

	special := parseSpecialCasing(fetch(specialCasingURL))
	fmt.Printf("Parsed %d default special-casing mappings\n", len(special))

	foldFull, foldSimple := parseCaseFolding(fetch(caseFoldingURL))
	fmt.Printf("Parsed %d full (C+F) and %d simple (C+S) fold mappings\n", len(foldFull), len(foldSimple))

	cased := parseProperty(fetch(derivedCorePropsURL), "Cased")
	caseIgnorable := parseProperty(fetch(derivedCorePropsURL), "Case_Ignorable")
	fmt.Printf("Parsed %d Cased ranges and %d Case_Ignorable ranges\n", len(cased), len(caseIgnorable))

	if err := generate(simple, special, foldFull, foldSimple, cased, caseIgnorable); err != nil {
		log.Fatalf("generate: %v", err)
	}
	fmt.Println("Successfully generated ucase_data.go")
}

func fetch(url string) io.ReadCloser {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("download %s: %v", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("download %s: status %d", url, resp.StatusCode)
	}
	return resp.Body
}

// parseUnicodeData reads UnicodeData.txt and returns the simple case mappings.
// Field 12 is Simple_Uppercase_Mapping, field 13 Simple_Lowercase_Mapping, and
// field 14 Simple_Titlecase_Mapping. Per UAX #44, when the titlecase field is
// empty it defaults to the uppercase mapping.
func parseUnicodeData(r io.ReadCloser) []simpleCaseEntry {
	defer r.Close()
	var out []simpleCaseEntry
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Split(line, ";")
		if len(f) < 15 {
			continue
		}
		cp := mustHex(f[0])
		up := optHex(f[12])
		lo := optHex(f[13])
		ti := optHex(f[14])
		if ti == 0 && up != 0 {
			ti = up // UAX #44 default: Simple_Titlecase_Mapping <- Simple_Uppercase_Mapping
		}
		if up == 0 && lo == 0 && ti == 0 {
			continue
		}
		out = append(out, simpleCaseEntry{cp: cp, up: up, lo: lo, ti: ti})
	}
	if err := sc.Err(); err != nil {
		log.Fatalf("scan UnicodeData.txt: %v", err)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].cp < out[j].cp })
	return out
}

// parseSpecialCasing reads SpecialCasing.txt and returns the default
// (language-insensitive) full and conditional mappings. Lines whose condition
// list contains a language tag (e.g. lt, tr, az) are skipped.
func parseSpecialCasing(r io.ReadCloser) []specialCaseEntry {
	defer r.Close()
	var out []specialCaseEntry
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// <code>; <lower>; <title>; <upper>; (<condition_list>;)?
		f := strings.Split(line, ";")
		if len(f) < 4 {
			continue
		}
		cp := mustHex(strings.TrimSpace(f[0]))
		lower := codePoints(f[1])
		title := codePoints(f[2])
		upper := codePoints(f[3])

		cond := ""
		// f[4] (when present) is the optional condition list. It is empty for
		// unconditional mappings because every data line ends with a ';'.
		if len(f) >= 5 && strings.TrimSpace(f[4]) != "" {
			tokens := strings.Fields(f[4])
			isLanguageSensitive := false
			for _, t := range tokens {
				if !knownContexts[t] {
					isLanguageSensitive = true // BCP 47 language tag
					break
				}
			}
			if isLanguageSensitive {
				continue // out of scope: locale tailoring (tr/az/lt)
			}
			// The default algorithm only ever encounters Final_Sigma. Skip any
			// other non-language context defensively (none exist in 17.0.0).
			switch {
			case len(tokens) == 1 && tokens[0] == "Final_Sigma":
				cond = "Final_Sigma"
			default:
				log.Printf("skipping unsupported default context %q for U+%04X", f[4], cp)
				continue
			}
		}
		out = append(out, specialCaseEntry{cp: cp, lower: lower, title: title, upper: upper, cond: cond})
	}
	if err := sc.Err(); err != nil {
		log.Fatalf("scan SpecialCasing.txt: %v", err)
	}
	// Sort by code point; for a given code point place conditional entries
	// before the unconditional one so the runtime checks conditions first.
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].cp != out[j].cp {
			return out[i].cp < out[j].cp
		}
		return out[i].cond != "" && out[j].cond == ""
	})
	return out
}

// parseCaseFolding reads CaseFolding.txt. Full folding (toCasefold) uses status
// C (common) + F (full); simple folding uses C + S (simple). Turkic (T)
// mappings are excluded from the default algorithm.
func parseCaseFolding(r io.ReadCloser) ([]foldEntry, []runePair) {
	defer r.Close()
	var full []foldEntry
	var simple []runePair
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		f := strings.Split(line, ";")
		if len(f) < 3 {
			continue
		}
		cp := mustHex(strings.TrimSpace(f[0]))
		status := strings.TrimSpace(f[1])
		target := codePoints(f[2])
		if len(target) == 0 {
			continue
		}
		switch status {
		case "C", "F":
			full = append(full, foldEntry{cp: cp, target: string(target)})
		}
		switch status {
		case "C", "S":
			simple = append(simple, runePair{cp: cp, to: target[0]})
		}
	}
	if err := sc.Err(); err != nil {
		log.Fatalf("scan CaseFolding.txt: %v", err)
	}
	sort.Slice(full, func(i, j int) bool { return full[i].cp < full[j].cp })
	sort.Slice(simple, func(i, j int) bool { return simple[i].cp < simple[j].cp })
	return full, simple
}

// parseProperty reads DerivedCoreProperties.txt and returns the ranges for the
// named binary property (e.g. "Cased" or "Case_Ignorable").
func parseProperty(r io.ReadCloser, name string) []runeRange {
	defer r.Close()
	var out []runeRange
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		f := strings.Split(line, ";")
		if len(f) < 2 {
			continue
		}
		if strings.TrimSpace(f[1]) != name {
			continue
		}
		lo, hi := parseRange(strings.TrimSpace(f[0]))
		out = append(out, runeRange{lo: lo, hi: hi})
	}
	if err := sc.Err(); err != nil {
		log.Fatalf("scan DerivedCoreProperties.txt: %v", err)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].lo < out[j].lo })
	return out
}

func parseRange(s string) (rune, rune) {
	if i := strings.Index(s, ".."); i >= 0 {
		return mustHex(s[:i]), mustHex(s[i+2:])
	}
	v := mustHex(s)
	return v, v
}

func codePoints(s string) []rune {
	fields := strings.Fields(s)
	out := make([]rune, 0, len(fields))
	for _, f := range fields {
		out = append(out, mustHex(f))
	}
	return out
}

func mustHex(s string) rune {
	v, err := strconv.ParseInt(strings.TrimSpace(s), 16, 32)
	if err != nil {
		log.Fatalf("bad hex %q: %v", s, err)
	}
	return rune(v)
}

func optHex(s string) rune {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	return mustHex(s)
}

func generate(simple []simpleCaseEntry, special []specialCaseEntry, foldFull []foldEntry, foldSimple []runePair, cased, caseIgnorable []runeRange) error {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "// Code generated by generate_ucase.go. DO NOT EDIT.\n")
	fmt.Fprintf(&buf, "// Source: Unicode Character Database %s\n", unicodeVersion)
	fmt.Fprintf(&buf, "//   UnicodeData.txt, SpecialCasing.txt, CaseFolding.txt, DerivedCoreProperties.txt\n")
	fmt.Fprintf(&buf, "// Case mapping is specified in The Unicode Standard, Section 3.13\n")
	fmt.Fprintf(&buf, "// \"Default Case Algorithms\" (not a numbered annex).\n\n")
	fmt.Fprintf(&buf, "package ucase\n\n")

	fmt.Fprintf(&buf, "// UnicodeVersion is the Unicode version these tables were generated from.\n")
	fmt.Fprintf(&buf, "const UnicodeVersion = %q\n\n", unicodeVersion)

	// Condition identifiers.
	fmt.Fprintf(&buf, "// Casing-context conditions from SpecialCasing.txt. condNone is an\n")
	fmt.Fprintf(&buf, "// unconditional mapping; condFinalSigma is the sole non-language context\n")
	fmt.Fprintf(&buf, "// condition in the default algorithm (Section 3.13, Final_Sigma).\n")
	fmt.Fprintf(&buf, "const (\n")
	fmt.Fprintf(&buf, "\tcondNone uint8 = iota\n")
	fmt.Fprintf(&buf, "\tcondFinalSigma\n")
	fmt.Fprintf(&buf, ")\n\n")

	// Simple mappings.
	fmt.Fprintf(&buf, "// simpleCaseEntry holds the simple (1:1) case mappings for one code point\n")
	fmt.Fprintf(&buf, "// from UnicodeData.txt. A zero rune means \"no mapping\" (maps to itself).\n")
	fmt.Fprintf(&buf, "type simpleCaseEntry struct {\n\tcp, up, lo, ti rune\n}\n\n")
	fmt.Fprintf(&buf, "// simpleCaseData lists simple case mappings, sorted by code point.\n")
	fmt.Fprintf(&buf, "// Total entries: %d\n", len(simple))
	fmt.Fprintf(&buf, "var simpleCaseData = []simpleCaseEntry{\n")
	for _, e := range simple {
		fmt.Fprintf(&buf, "\t{0x%04X, 0x%04X, 0x%04X, 0x%04X},\n", e.cp, e.up, e.lo, e.ti)
	}
	fmt.Fprintf(&buf, "}\n\n")

	// Special (full + conditional) mappings.
	fmt.Fprintf(&buf, "// specialCaseEntry holds a full case mapping (possibly 1:many) from\n")
	fmt.Fprintf(&buf, "// SpecialCasing.txt for the default (language-insensitive) algorithm.\n")
	fmt.Fprintf(&buf, "type specialCaseEntry struct {\n")
	fmt.Fprintf(&buf, "\tcp                  rune\n")
	fmt.Fprintf(&buf, "\tlower, title, upper []rune\n")
	fmt.Fprintf(&buf, "\tcond                uint8\n")
	fmt.Fprintf(&buf, "}\n\n")
	fmt.Fprintf(&buf, "// specialCaseData lists the default full case mappings, sorted by code\n")
	fmt.Fprintf(&buf, "// point (conditional entries first per code point). Language-sensitive\n")
	fmt.Fprintf(&buf, "// (tr/az/lt) lines from SpecialCasing.txt are intentionally omitted.\n")
	fmt.Fprintf(&buf, "// Total entries: %d\n", len(special))
	fmt.Fprintf(&buf, "var specialCaseData = []specialCaseEntry{\n")
	for _, e := range special {
		cond := "condNone"
		if e.cond == "Final_Sigma" {
			cond = "condFinalSigma"
		}
		fmt.Fprintf(&buf, "\t{0x%04X, %s, %s, %s, %s},\n",
			e.cp, runeSlice(e.lower), runeSlice(e.title), runeSlice(e.upper), cond)
	}
	fmt.Fprintf(&buf, "}\n\n")

	// Full case folding.
	fmt.Fprintf(&buf, "// foldEntry maps a code point to its full case folding (C+F), which may\n")
	fmt.Fprintf(&buf, "// be more than one code point.\n")
	fmt.Fprintf(&buf, "type foldEntry struct {\n\tcp     rune\n\ttarget string\n}\n\n")
	fmt.Fprintf(&buf, "// foldFullData lists full case folding mappings (status C+F) from\n")
	fmt.Fprintf(&buf, "// CaseFolding.txt, sorted by code point. Total entries: %d\n", len(foldFull))
	fmt.Fprintf(&buf, "var foldFullData = []foldEntry{\n")
	for _, e := range foldFull {
		fmt.Fprintf(&buf, "\t{0x%04X, %+q},\n", e.cp, e.target)
	}
	fmt.Fprintf(&buf, "}\n\n")

	// Simple case folding.
	fmt.Fprintf(&buf, "// runePair maps one code point to one code point.\n")
	fmt.Fprintf(&buf, "type runePair struct{ cp, to rune }\n\n")
	fmt.Fprintf(&buf, "// simpleFoldData lists simple case folding mappings (status C+S) from\n")
	fmt.Fprintf(&buf, "// CaseFolding.txt, sorted by code point. Total entries: %d\n", len(foldSimple))
	fmt.Fprintf(&buf, "var simpleFoldData = []runePair{\n")
	for _, e := range foldSimple {
		fmt.Fprintf(&buf, "\t{0x%04X, 0x%04X},\n", e.cp, e.to)
	}
	fmt.Fprintf(&buf, "}\n\n")

	// Properties.
	fmt.Fprintf(&buf, "// runeRange is an inclusive code point range.\n")
	fmt.Fprintf(&buf, "type runeRange struct{ lo, hi rune }\n\n")
	writeRanges(&buf, "casedRanges", "Cased (DerivedCoreProperties.txt, Section 3.13 D135)", cased)
	writeRanges(&buf, "caseIgnorableRanges", "Case_Ignorable (DerivedCoreProperties.txt, Section 3.13 D136)", caseIgnorable)

	src, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("format generated source: %w", err)
	}
	return os.WriteFile("ucase_data.go", src, 0o644)
}

func writeRanges(buf *bytes.Buffer, name, doc string, ranges []runeRange) {
	fmt.Fprintf(buf, "// %s holds the %s property ranges, sorted by lo. Total ranges: %d\n", name, doc, len(ranges))
	fmt.Fprintf(buf, "var %s = []runeRange{\n", name)
	for _, r := range ranges {
		fmt.Fprintf(buf, "\t{0x%04X, 0x%04X},\n", r.lo, r.hi)
	}
	fmt.Fprintf(buf, "}\n\n")
}

func runeSlice(rs []rune) string {
	var b strings.Builder
	b.WriteString("[]rune{")
	for i, r := range rs {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "0x%04X", r)
	}
	b.WriteString("}")
	return b.String()
}
