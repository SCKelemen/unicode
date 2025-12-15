package uax29

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

// parseUnicodeTestLine parses a line from GraphemeBreakTest.txt, WordBreakTest.txt, or SentenceBreakTest.txt
// Format: ÷ 0020 × 0020 ÷ 0020 ÷ # comment
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
	// The breaks array has breaks[i] = true if there's a break before runes[i]
	bytePos := 0
	for i, r := range runes {
		if i < len(breaks) && breaks[i] {
			expectedBreaks = append(expectedBreaks, bytePos)
		}
		bytePos += len(string(r))
	}

	// Add final break
	if len(breaks) > len(runes) && breaks[len(runes)] {
		expectedBreaks = append(expectedBreaks, bytePos)
	}

	return text, expectedBreaks, nil
}

func TestGraphemeBreakOfficial(t *testing.T) {
	file, err := os.Open("GraphemeBreakTest.txt")
	if err != nil {
		t.Fatalf("Failed to open GraphemeBreakTest.txt: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	passed := 0
	failed := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		text, expectedBreaks, err := parseUnicodeTestLine(line)
		if err != nil {
			continue // Skip malformed lines
		}

		actualBreaks := FindGraphemeBreaks(text)

		// Compare breaks
		if !equalBreaks(expectedBreaks, actualBreaks) {
			failed++
			if failed <= 10 { // Only show first 10 failures
				t.Errorf("Line %d failed:\n  Text: %q\n  Expected breaks: %v\n  Actual breaks: %v\n  Line: %s",
					lineNum, text, expectedBreaks, actualBreaks, line)
			}
		} else {
			passed++
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading file: %v", err)
	}

	t.Logf("Grapheme Break Test Results: %d passed, %d failed (%.1f%% pass rate)",
		passed, failed, 100.0*float64(passed)/float64(passed+failed))

	if failed > 0 {
		t.Errorf("%d/%d tests failed", failed, passed+failed)
	}
}

func TestWordBreakOfficial(t *testing.T) {
	file, err := os.Open("WordBreakTest.txt")
	if err != nil {
		t.Fatalf("Failed to open WordBreakTest.txt: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	passed := 0
	failed := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		text, expectedBreaks, err := parseUnicodeTestLine(line)
		if err != nil {
			continue // Skip malformed lines
		}

		actualBreaks := FindWordBreaks(text)

		// Compare breaks
		if !equalBreaks(expectedBreaks, actualBreaks) {
			failed++
			if failed <= 10 { // Only show first 10 failures
				t.Errorf("Line %d failed:\n  Text: %q\n  Expected breaks: %v\n  Actual breaks: %v\n  Line: %s",
					lineNum, text, expectedBreaks, actualBreaks, line)
			}
		} else {
			passed++
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading file: %v", err)
	}

	t.Logf("Word Break Test Results: %d passed, %d failed (%.1f%% pass rate)",
		passed, failed, 100.0*float64(passed)/float64(passed+failed))

	if failed > 0 {
		t.Errorf("%d/%d tests failed", failed, passed+failed)
	}
}

func TestSentenceBreakOfficial(t *testing.T) {
	file, err := os.Open("SentenceBreakTest.txt")
	if err != nil {
		t.Fatalf("Failed to open SentenceBreakTest.txt: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	passed := 0
	failed := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		text, expectedBreaks, err := parseUnicodeTestLine(line)
		if err != nil {
			continue // Skip malformed lines
		}

		actualBreaks := FindSentenceBreaks(text)

		// Compare breaks
		if !equalBreaks(expectedBreaks, actualBreaks) {
			failed++
			if failed <= 10 { // Only show first 10 failures
				t.Errorf("Line %d failed:\n  Text: %q\n  Expected breaks: %v\n  Actual breaks: %v\n  Line: %s",
					lineNum, text, expectedBreaks, actualBreaks, line)
			}
		} else {
			passed++
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading file: %v", err)
	}

	t.Logf("Sentence Break Test Results: %d passed, %d failed (%.1f%% pass rate)",
		passed, failed, 100.0*float64(passed)/float64(passed+failed))

	if failed > 0 {
		t.Errorf("%d/%d tests failed", failed, passed+failed)
	}
}

func equalBreaks(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
