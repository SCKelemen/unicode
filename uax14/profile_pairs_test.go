package uax14

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"os"
	"sort"
	"strings"
	"testing"
)

// lineBreakTestPaths are the locations searched for the official
// LineBreakTest.txt conformance file.
var lineBreakTestPaths = []string{
	"LineBreakTest.txt",
	"/tmp/LineBreakTest.txt",
	"testdata/LineBreakTest.txt",
	"../testdata/LineBreakTest.txt",
}

// findLineBreakTestFile opens the official LineBreakTest.txt from any of the
// known locations, returning the open file or false if it cannot be found.
func findLineBreakTestFile() (*os.File, bool) {
	for _, path := range lineBreakTestPaths {
		if f, err := os.Open(path); err == nil {
			return f, true
		}
	}
	return nil, false
}

// pairProfile records, for every (prev, curr) break-class pair encountered in
// the conformance suite, how many positions were resolved by an explicit rule
// versus by the pair table.
type pairProfile struct {
	ruleHits   map[[2]BreakClass]int64 // pairs that matched a rule
	tableHits  map[[2]BreakClass]int64 // pairs that used the pair table
	totalPairs int64
	testCount  int
}

// newPairProfile returns an empty profile with initialized maps.
func newPairProfile() *pairProfile {
	return &pairProfile{
		ruleHits:  make(map[[2]BreakClass]int64),
		tableHits: make(map[[2]BreakClass]int64),
	}
}

// ruleOnlyPairs returns the set of pairs that ever required rule checking,
// keyed by pair with the observed rule-hit count.
func (p *pairProfile) ruleOnlyPairs() map[[2]BreakClass]int64 {
	out := make(map[[2]BreakClass]int64, len(p.ruleHits))
	for pair, count := range p.ruleHits {
		out[pair] = count
	}
	return out
}

// profileLineBreaks runs the line-break engine over every position in the given
// LineBreakTest.txt reader and records which (prev, curr) class pairs require
// rule checking. It is read-only and never writes any file.
//
// This mirrors the runtime decision path so the resulting set is exactly the
// set of pairs that must consult the rules (the basis for rule_exception_pairs.go).
func profileLineBreaks(file *os.File) (*pairProfile, error) {
	p := newPairProfile()
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments and empty lines.
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		// Parse test line (data is before the '#' comment).
		parts := strings.Split(line, "#")
		if len(parts) < 2 {
			continue
		}

		tokens := strings.Fields(strings.TrimSpace(parts[0]))
		if len(tokens) == 0 {
			continue
		}

		// Build the test string from the hex code points, ignoring the
		// '÷' / '×' break markers.
		var runes []rune
		for _, token := range tokens {
			if token == "÷" || token == "×" {
				continue
			}
			var codepoint rune
			if _, err := fmt.Sscanf(token, "%X", &codepoint); err == nil {
				runes = append(runes, codepoint)
			}
		}

		if len(runes) == 0 {
			continue
		}

		p.testCount++
		p.profile(runes)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return p, nil
}

// profile classifies each adjacent (prev, curr) pair in runes as rule-matched
// or pair-table-resolved, accumulating the counts into p.
func (p *pairProfile) profile(runes []rune) {
	text := string(runes)

	ctx := NewLineBreakContext(text, HyphensNone)
	prevClass := getBreakClass(runes[0])
	if isClassOrVariant(prevClass, ClassCM) || prevClass == ClassZWJ {
		prevClass = ClassAL
	}

	for i := 1; i < len(runes); i++ {
		ctx.Slide()
		currClass := getBreakClass(runes[i])

		// LB9: SA combining marks → CM
		if currClass == ClassSA {
			if isCombiningMark(runes[i]) {
				currClass = ClassCM
			}
		}

		ctx.UpdatePrevClass(prevClass)
		p.totalPairs++

		// Check if any inlined rule matches
		ruleMatched := false

		// LB4: BK ÷
		if prevClass == ClassBK {
			ruleMatched = true
		}

		// LB5a: CR × LF
		if !ruleMatched && prevClass == ClassCR && currClass == ClassLF {
			ruleMatched = true
		}

		// LB5b: CR ÷, LF ÷, NL ÷
		if !ruleMatched && (prevClass == ClassCR || prevClass == ClassLF || prevClass == ClassNL) {
			ruleMatched = true
		}

		// LB7: × ZW
		if !ruleMatched && currClass == ClassZW {
			ruleMatched = true
		}

		// LB8a: ZWJ ×
		if !ruleMatched && i > 0 && runes[i-1] == '\u200D' {
			ruleMatched = true
		}

		// LB8, LB11, and remaining rules
		if !ruleMatched {
			// Check LB8 (index 5)
			if matched, _ := lineBreakRules[5](ctx); matched {
				ruleMatched = true
			}
		}

		if !ruleMatched && (prevClass == ClassWJ || currClass == ClassWJ) {
			ruleMatched = true
		}

		// Check remaining rules (7-43)
		if !ruleMatched {
			for idx := 7; idx < len(lineBreakRules); idx++ {
				if matched, _ := lineBreakRules[idx](ctx); matched {
					ruleMatched = true
					break
				}
			}
		}

		// Record hit
		pair := [2]BreakClass{prevClass, currClass}
		if ruleMatched {
			p.ruleHits[pair]++
		} else {
			p.tableHits[pair]++
		}

		// Update prevClass for next iteration (simplified)
		if isClassOrVariant(currClass, ClassCM) || currClass == ClassZWJ {
			prevClass = ClassAL
		} else {
			prevClass = currClass
		}
	}
}

// TestProfilePairTableUsage profiles which (prev, curr) class pairs use rules
// versus the pair table. It is observational only: it reads LineBreakTest.txt
// and logs statistics, and never writes any file. Regeneration of
// rule_exception_pairs.go is handled separately by
// TestRegenerateRuleExceptionPairs (env-gated).
func TestProfilePairTableUsage(t *testing.T) {
	file, ok := findLineBreakTestFile()
	if !ok {
		t.Skip("Official LineBreakTest.txt not found - skipping. " +
			"Download from https://www.unicode.org/Public/UCD/latest/ucd/auxiliary/LineBreakTest.txt")
		return
	}
	defer file.Close()

	p, err := profileLineBreaks(file)
	if err != nil {
		t.Fatalf("Error reading test file: %v", err)
	}

	t.Logf("Profiled %d test cases, %d total position pairs", p.testCount, p.totalPairs)
	p.report(t)
}

// report logs the pair-table-versus-rule usage statistics for the profile.
func (p *pairProfile) report(t *testing.T) {
	t.Logf("\n=== Pair Table vs Rule Usage Profile ===")
	t.Logf("Total pairs checked: %d", p.totalPairs)
	t.Logf("")

	totalRuleHits := int64(0)
	totalTableHits := int64(0)
	for _, count := range p.ruleHits {
		totalRuleHits += count
	}
	for _, count := range p.tableHits {
		totalTableHits += count
	}

	t.Logf("Pairs matched by rules: %d (%.2f%%)", totalRuleHits, float64(totalRuleHits)/float64(p.totalPairs)*100)
	t.Logf("Pairs matched by pair table: %d (%.2f%%)", totalTableHits, float64(totalTableHits)/float64(p.totalPairs)*100)
	t.Logf("")

	uniqueTotal := make(map[[2]BreakClass]bool)
	for pair := range p.ruleHits {
		uniqueTotal[pair] = true
	}
	for pair := range p.tableHits {
		uniqueTotal[pair] = true
	}

	t.Logf("Unique (prev, curr) pairs:")
	t.Logf("  Rule-only pairs: %d", len(p.ruleHits))
	t.Logf("  Table-only pairs: %d", len(p.tableHits))
	t.Logf("  Total unique pairs: %d (out of %d possible)", len(uniqueTotal), 65*65)
	t.Logf("")

	ruleOnly := p.ruleOnlyPairs()
	t.Logf("Pairs that EVER need rule checking (may also use pair table sometimes):")
	t.Logf("  Found %d rule-only pairs", len(ruleOnly))
	t.Logf("  To regenerate rule_exception_pairs.go: UAX14_REGEN=1 go test ./uax14/ -run TestRegenerateRuleExceptionPairs")

	// Show a deterministic sample of the rule-only pairs.
	keys := sortedPairs(ruleOnly)
	limit := len(keys)
	if limit > 20 {
		limit = 20
		t.Logf("  (Showing first 20 pairs)")
	} else {
		t.Logf("  List:")
	}
	for _, pair := range keys[:limit] {
		t.Logf("    (%s, %s): %d hits", classNameShort(pair[0]), classNameShort(pair[1]), ruleOnly[pair])
	}

	t.Logf("")
	t.Logf("Strategy recommendation:")
	ruleOnlyPercent := float64(len(ruleOnly)) / float64(65*65) * 100
	t.Logf("  Mark %.1f%% of possible pairs (%d pairs) as BreakCheckRules", ruleOnlyPercent, len(ruleOnly))
	t.Logf("  This would give instant pair table results for %.1f%% of positions", float64(totalTableHits)/float64(p.totalPairs)*100)
}

// TestRegenerateRuleExceptionPairs regenerates rule_exception_pairs.go from the
// official conformance suite. It is skipped by default so that a normal
// "go test" run never mutates tracked source; set UAX14_REGEN=1 to run it.
//
// The generation logic lives here (rather than in a standalone //go:build
// ignore program like the other generate_*.go tools) because it must drive the
// package-internal line-break engine (getBreakClass, lineBreakRules, ...),
// which a separate package main cannot reach. Output is deterministic (pairs
// sorted by class) and gofmt-clean (formatted via go/format).
func TestRegenerateRuleExceptionPairs(t *testing.T) {
	if os.Getenv("UAX14_REGEN") == "" {
		t.Skip("set UAX14_REGEN=1 to regenerate rule_exception_pairs.go")
	}

	file, ok := findLineBreakTestFile()
	if !ok {
		t.Fatal("LineBreakTest.txt not found; cannot regenerate rule_exception_pairs.go")
	}
	defer file.Close()

	p, err := profileLineBreaks(file)
	if err != nil {
		t.Fatalf("profile LineBreakTest.txt: %v", err)
	}

	if err := writeRuleExceptionPairs("rule_exception_pairs.go", p.ruleOnlyPairs()); err != nil {
		t.Fatalf("write rule_exception_pairs.go: %v", err)
	}
	t.Logf("Regenerated rule_exception_pairs.go with %d rule-only pairs", len(p.ruleHits))
}

// sortedPairs returns the keys of pairs sorted ascending by (prev, curr) so
// that any emitted output is deterministic.
func sortedPairs(pairs map[[2]BreakClass]int64) [][2]BreakClass {
	keys := make([][2]BreakClass, 0, len(pairs))
	for pair := range pairs {
		keys = append(keys, pair)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i][0] != keys[j][0] {
			return keys[i][0] < keys[j][0]
		}
		return keys[i][1] < keys[j][1]
	})
	return keys
}

// writeRuleExceptionPairs regenerates the rule-exception lookup file at path
// from the given set of rule-only pairs. Output is deterministic (sorted) and
// is run through go/format so the result is always gofmt-clean (including the
// blank "//" separator required before the //go:inline directive).
func writeRuleExceptionPairs(path string, pairs map[[2]BreakClass]int64) error {
	var buf bytes.Buffer
	pct := float64(len(pairs)) / 4225 * 100
	fmt.Fprintf(&buf, `package uax14

// rule_exception_pairs.go - Code generated by TestRegenerateRuleExceptionPairs. DO NOT EDIT.
//
// This file contains the set of (prev, curr) class pairs that need
// rule checking (vs using the pair table directly).
//
// %d unique pairs marked as needing rules (%.1f%% of 4,225 possible pairs).
//
// Regenerate with:
//	UAX14_REGEN=1 go test ./uax14/ -run TestRegenerateRuleExceptionPairs

// isRuleExceptionPair returns true if (prev, curr) needs rule checking.
// Uses direct array indexing for O(1) lookup (no map overhead).
//
//go:inline
func isRuleExceptionPair(prev, curr BreakClass) bool {
	return ruleExceptionArray[prev][curr]
}

// ruleExceptionArray is a 2D array for O(1) pair lookups.
// Size: 128x128 = 16KB (same as pair table, fits in L1 cache).
var ruleExceptionArray [128][128]bool

func init() {
	// Mark pairs that need rule checking.
`, len(pairs), pct)

	for _, pair := range sortedPairs(pairs) {
		fmt.Fprintf(&buf, "\truleExceptionArray[%s][%s] = true\n",
			classNameFull(pair[0]), classNameFull(pair[1]))
	}
	buf.WriteString("}\n")

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("format generated source: %w", err)
	}
	return os.WriteFile(path, formatted, 0o644)
}

func classNameShort(c BreakClass) string {
	names := map[BreakClass]string{
		ClassBK: "BK", ClassCR: "CR", ClassLF: "LF", ClassNL: "NL", ClassSP: "SP",
		ClassZW: "ZW", ClassZWJ: "ZWJ", ClassWJ: "WJ", ClassGL: "GL", ClassBA: "BA",
		ClassHY: "HY", ClassCL: "CL", ClassCP: "CP", ClassEX: "EX", ClassIN: "IN",
		ClassNS: "NS", ClassOP: "OP", ClassQU: "QU", ClassIS: "IS", ClassNU: "NU",
		ClassPO: "PO", ClassPR: "PR", ClassSY: "SY", ClassAL: "AL", ClassHL: "HL",
		ClassH2: "H2", ClassH3: "H3", ClassID: "ID", ClassEB: "EB", ClassEM: "EM",
		ClassCM: "CM", ClassB2: "B2", ClassCB: "CB", ClassJL: "JL", ClassJV: "JV",
		ClassJT: "JT", ClassRI: "RI", ClassXX: "XX",
	}
	if name, ok := names[c]; ok {
		return name
	}
	return fmt.Sprintf("%d", c)
}

func classNameFull(c BreakClass) string {
	names := map[BreakClass]string{
		ClassBK: "ClassBK", ClassCR: "ClassCR", ClassLF: "ClassLF", ClassNL: "ClassNL",
		ClassSP: "ClassSP", ClassZW: "ClassZW", ClassZWJ: "ClassZWJ", ClassWJ: "ClassWJ",
		ClassGL: "ClassGL", ClassBA: "ClassBA", ClassHY: "ClassHY", ClassCL: "ClassCL",
		ClassCP: "ClassCP", ClassEX: "ClassEX", ClassIN: "ClassIN", ClassNS: "ClassNS",
		ClassOP: "ClassOP", ClassQU: "ClassQU", ClassIS: "ClassIS", ClassNU: "ClassNU",
		ClassPO: "ClassPO", ClassPR: "ClassPR", ClassSY: "ClassSY", ClassAL: "ClassAL",
		ClassHL: "ClassHL", ClassH2: "ClassH2", ClassH3: "ClassH3", ClassID: "ClassID",
		ClassEB: "ClassEB", ClassEM: "ClassEM", ClassCM: "ClassCM", ClassB2: "ClassB2",
		ClassCB: "ClassCB", ClassJL: "ClassJL", ClassJV: "ClassJV", ClassJT: "ClassJT",
		ClassRI: "ClassRI", ClassXX: "ClassXX",
	}
	if name, ok := names[c]; ok {
		return name
	}
	return fmt.Sprintf("BreakClass(%d)", c)
}
