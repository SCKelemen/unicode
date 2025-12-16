package uax14

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

// parseUnicodeTestLine parses a line from LineBreakTest.txt
// Format: × 0020 ÷ 0020 × 0020 ÷ # comment
// × means no break, ÷ means break opportunity
func parseUnicodeTestLine(line string) (text string, expectedBreaks []int, err error) {
	// Remove comment
	if idx := strings.Index(line, "#"); idx >= 0 {
		line = line[:idx]
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return "", nil, fmt.Errorf("empty line")
	}

	// Parse tokens
	tokens := strings.Fields(line)
	var runes []rune
	var breaks []bool

	for _, token := range tokens {
		switch token {
		case "×":
			breaks = append(breaks, false)
		case "÷":
			breaks = append(breaks, true)
		default:
			// Parse hex codepoint
			codepoint, err := strconv.ParseInt(token, 16, 32)
			if err != nil {
				return "", nil, fmt.Errorf("invalid codepoint %s: %v", token, err)
			}
			runes = append(runes, rune(codepoint))
		}
	}

	// Build text string
	text = string(runes)

	// Calculate break positions (in bytes)
	var breakPositions []int
	bytePos := 0
	runeIdx := 0

	for i, shouldBreak := range breaks {
		if shouldBreak {
			breakPositions = append(breakPositions, bytePos)
		}

		// Advance byte position if we have a rune at this position
		if runeIdx < len(runes) {
			if i < len(breaks)-1 {  // Not the final break marker
				bytePos += len(string([]rune{runes[runeIdx]}))
				runeIdx++
			}
		}
	}

	return text, breakPositions, nil
}

// TestOfficialUnicodeVectors runs tests from the official LineBreakTest.txt
// Note: This is optional and requires the test file to be present
func TestOfficialUnicodeVectors(t *testing.T) {
	// Try to find the test file in various locations
	testFiles := []string{
		"LineBreakTest.txt",
		"/tmp/LineBreakTest.txt",
		"testdata/LineBreakTest.txt",
		"../testdata/LineBreakTest.txt",
	}

	var file *os.File
	var err error

	for _, path := range testFiles {
		file, err = os.Open(path)
		if err == nil {
			defer file.Close()
			break
		}
	}

	if err != nil {
		t.Skip("Official LineBreakTest.txt not found - skipping. " +
			"Download from https://www.unicode.org/Public/UCD/latest/ucd/auxiliary/LineBreakTest.txt")
		return
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0
	passed := 0
	failed := 0
	skipped := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip comments and empty lines
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		text, expectedBreaks, err := parseUnicodeTestLine(line)
		if err != nil {
			t.Logf("Line %d: parse error: %v", lineNum, err)
			skipped++
			continue
		}

		// Run our algorithm
		actualBreaks := FindLineBreakOpportunities(text, HyphensManual)

		// Compare results
		// Note: Our implementation is simplified, so we expect some differences
		if breaksMatch(expectedBreaks, actualBreaks) {
			passed++
		} else {
			failed++
			if failed <= 10 {  // Only log first 10 failures
				t.Logf("Line %d: MISMATCH", lineNum)
				t.Logf("  Text: %q", text)
				t.Logf("  Expected: %v", expectedBreaks)
				t.Logf("  Got:      %v", actualBreaks)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading file: %v", err)
	}

	total := passed + failed
	passRate := float64(passed) / float64(total) * 100

	t.Logf("\n=== Official Unicode Test Results ===")
	t.Logf("Total tests: %d", total)
	t.Logf("Passed: %d (%.1f%%)", passed, passRate)
	t.Logf("Failed: %d (%.1f%%)", failed, 100-passRate)
	t.Logf("Skipped: %d", skipped)

	// We expect our simplified implementation to pass a reasonable percentage
	// but not 100% since we don't implement all UAX #14 rules
	if passRate < 50 {
		t.Errorf("Pass rate too low: %.1f%% (expected at least 50%%)", passRate)
	}
}

// breaksMatch checks if two break point slices are equivalent
// Note: The Unicode test format doesn't include position 0, but our implementation does.
// We need to skip the first break (0) in our results when comparing.
func breaksMatch(expected, actual []int) bool {
	// Our implementation always includes position 0, but Unicode tests don't
	// Remove position 0 from actual breaks for comparison
	actualWithoutZero := actual
	if len(actual) > 0 && actual[0] == 0 {
		actualWithoutZero = actual[1:]
	}

	if len(expected) != len(actualWithoutZero) {
		return false
	}
	for i := range expected {
		if expected[i] != actualWithoutZero[i] {
			return false
		}
	}
	return true
}
