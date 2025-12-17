package uax14

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"testing"
)

// TestRulesBasicCases tests basic line break rule functionality
func TestRulesBasicCases(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		hyphens  Hyphens
		expected []int
	}{
		{
			name:     "LB4: Break after BK",
			text:     "abc\u2028def", // BK = U+2028 (Line Separator)
			hyphens:  HyphensNone,
			expected: []int{0, 6, 9}, // Break after BK (3 bytes "abc" + 3 bytes U+2028)
		},
		{
			name:     "LB5a: CR × LF",
			text:     "abc\r\ndef",
			hyphens:  HyphensNone,
			expected: []int{0, 5, 8}, // Break after CRLF pair, not between
		},
		{
			name:     "LB5b: Break after CR",
			text:     "abc\rdef",
			hyphens:  HyphensNone,
			expected: []int{0, 4, 7}, // Break after CR when not followed by LF
		},
		{
			name:     "LB8a: Do not break after ZWJ",
			text:     "a\u200Db", // U+200D = ZWJ
			hyphens:  HyphensNone,
			expected: []int{0, 5}, // No break after ZWJ
		},
		{
			name:     "LB21: AL × HY ÷ AL (hyphenated word)",
			text:     "Excusez-moi",
			hyphens:  HyphensNone,
			expected: []int{0, 8, 11}, // Break after hyphen
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindLineBreakOpportunitiesWithRules(tt.text, tt.hyphens)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("FindLineBreakOpportunitiesWithRules() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestRulesVsOriginal compares rule-based implementation against original
func TestRulesVsOriginal(t *testing.T) {
	tests := []string{
		"Hello, world!",
		"This is a test.",
		"Line 1\nLine 2\nLine 3",
		"Hyphen-ated words",
		"Hebrew: \u05D0\u05D1\u05D2",
		"CJK: 这是一个测试",
		"Mixed: Hello世界",
		"Quotes: \"hello\" 'world'",
		"ZWJ: 👨‍👩‍👧‍👦", // Family emoji with ZWJ
		"",
		"a",
		"ab",
	}

	for _, text := range tests {
		t.Run(fmt.Sprintf("text=%q", text), func(t *testing.T) {
			original := FindLineBreakOpportunities(text, HyphensNone)
			rules := FindLineBreakOpportunitiesWithRules(text, HyphensNone)

			if !reflect.DeepEqual(original, rules) {
				t.Errorf("Mismatch for text %q:\nOriginal: %v\nRules:    %v", text, original, rules)
			}
		})
	}
}

// TestRulesOfficialConformance runs official Unicode conformance tests using the rule-based implementation
func TestRulesOfficialConformance(t *testing.T) {
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
		if line == "" || line[0] == '#' {
			skipped++
			continue
		}

		text, expectedBreaks, err := parseUnicodeTestLine(line)
		if err != nil {
			// Skip unparseable lines
			skipped++
			continue
		}

		// Test with HyphensManual (to match official test expectations)
		actualBreaks := FindLineBreakOpportunitiesWithRules(text, HyphensManual)

		if breaksMatch(expectedBreaks, actualBreaks) {
			passed++
		} else {
			failed++
			if failed <= 10 { // Only show first 10 failures
				t.Errorf("Line %d failed:\nText: %q\nExpected: %v\nActual:   %v\nOriginal line: %s",
					lineNum, text, expectedBreaks, actualBreaks, line)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading test file: %v", err)
	}

	total := passed + failed
	passRate := float64(passed) / float64(total) * 100

	t.Logf("\n=== Rule-Based Implementation Test Results ===")
	t.Logf("Total tests: %d", total)
	t.Logf("Passed: %d (%.1f%%)", passed, passRate)
	t.Logf("Failed: %d (%.1f%%)", failed, float64(failed)/float64(total)*100)
	t.Logf("Skipped: %d", skipped)

	if failed > 0 {
		t.Errorf("Rule-based implementation failed %d/%d tests (%.1f%% pass rate)", failed, total, passRate)
	}
}

// BenchmarkRulesVsOriginal compares performance of rule-based vs original implementation
func BenchmarkRulesVsOriginal(b *testing.B) {
	testCases := []string{
		"Short text",
		"This is a longer text with multiple words and punctuation marks!",
		"Mixed languages: Hello 世界 שלום мир",
		"Line 1\nLine 2\nLine 3\nLine 4\nLine 5",
	}

	for _, text := range testCases {
		b.Run(fmt.Sprintf("Original/len=%d", len(text)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				FindLineBreakOpportunities(text, HyphensNone)
			}
		})

		b.Run(fmt.Sprintf("Rules/len=%d", len(text)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				FindLineBreakOpportunitiesWithRules(text, HyphensNone)
			}
		})
	}
}
