package uax31

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

const (
	derivedCorePropsURL = "https://www.unicode.org/Public/17.0.0/ucd/DerivedCoreProperties.txt"
	propListURL         = "https://www.unicode.org/Public/17.0.0/ucd/PropList.txt"
)

func TestOfficialIdentifierProperties(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping official conformance tests in short mode")
	}

	totalTests := 0
	totalFails := 0

	// Test XID_Start
	t.Run("XID_Start", func(t *testing.T) {
		tests, fails := testProperty(t, derivedCorePropsURL, "XID_Start", IsXIDStart)
		totalTests += tests
		totalFails += fails
	})

	// Test XID_Continue
	t.Run("XID_Continue", func(t *testing.T) {
		tests, fails := testProperty(t, derivedCorePropsURL, "XID_Continue", IsXIDContinue)
		totalTests += tests
		totalFails += fails
	})

	// Test Pattern_Syntax
	t.Run("Pattern_Syntax", func(t *testing.T) {
		tests, fails := testProperty(t, propListURL, "Pattern_Syntax", IsPatternSyntax)
		totalTests += tests
		totalFails += fails
	})

	// Test Pattern_White_Space
	t.Run("Pattern_White_Space", func(t *testing.T) {
		tests, fails := testProperty(t, propListURL, "Pattern_White_Space", IsPatternWhiteSpace)
		totalTests += tests
		totalFails += fails
	})

	// Overall summary
	if totalFails > 0 {
		t.Errorf("OVERALL: %d/%d tests failed", totalFails, totalTests)
	} else {
		t.Logf("OVERALL: %d/%d tests passed (100%% conformance)", totalTests, totalTests)
	}
}

func testProperty(t *testing.T, url, propertyName string, testFunc func(rune) bool) (int, int) {
	// Download the property file
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to download %s: %v", url, err)
		return 0, 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("HTTP error downloading %s: %d", url, resp.StatusCode)
		return 0, 0
	}

	scanner := bufio.NewScanner(resp.Body)
	lineNum := 0
	testCount := 0
	failCount := 0
	var firstFailures []string

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse line: CODE_POINT(S) ; PROPERTY_NAME # COMMENT
		parts := strings.Split(line, ";")
		if len(parts) < 2 {
			continue
		}

		codePointPart := strings.TrimSpace(parts[0])
		propName := strings.TrimSpace(strings.Split(parts[1], "#")[0])

		// Only process lines for our property
		if propName != propertyName {
			continue
		}

		// Parse code point or range
		var start, end rune
		if strings.Contains(codePointPart, "..") {
			rangeParts := strings.Split(codePointPart, "..")
			startVal, _ := strconv.ParseInt(strings.TrimSpace(rangeParts[0]), 16, 32)
			endVal, _ := strconv.ParseInt(strings.TrimSpace(rangeParts[1]), 16, 32)
			start = rune(startVal)
			end = rune(endVal)
		} else {
			val, _ := strconv.ParseInt(codePointPart, 16, 32)
			start = rune(val)
			end = rune(val)
		}

		// Test each code point in range
		for r := start; r <= end; r++ {
			testCount++
			if !testFunc(r) {
				failCount++
				if len(firstFailures) < 100 {
					firstFailures = append(firstFailures,
						fmt.Sprintf("Line %d: U+%04X should have %s property but doesn't",
							lineNum, r, propertyName))
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading property file: %v", err)
		return 0, 0
	}

	// Also test that code points NOT in the file return false
	// Sample some code points that should not have the property
	sampleTests := 0
	sampleFails := 0

	// Test ASCII control characters (should not have most properties except Pattern_White_Space for some)
	if propertyName != "Pattern_White_Space" {
		for r := rune(0x00); r <= rune(0x1F); r++ {
			sampleTests++
			if testFunc(r) && !shouldHaveProperty(r, propertyName) {
				sampleFails++
				if len(firstFailures) < 100 {
					firstFailures = append(firstFailures,
						fmt.Sprintf("U+%04X should NOT have %s property but does", r, propertyName))
				}
			}
		}
	}

	testCount += sampleTests
	failCount += sampleFails

	// Report results
	if failCount > 0 {
		t.Errorf("FAILED: %d/%d tests failed for %s", failCount, testCount, propertyName)
		t.Logf("First failures:")
		for i, fail := range firstFailures {
			if i >= 10 {
				t.Logf("  ... and %d more failures", len(firstFailures)-10)
				break
			}
			t.Logf("  %s", fail)
		}
	} else {
		t.Logf("PASSED: %d/%d tests for %s (100%% conformance)", testCount, testCount, propertyName)
	}

	return testCount, failCount
}

// shouldHaveProperty checks if a code point should have a property based on general rules
func shouldHaveProperty(r rune, propertyName string) bool {
	// For Pattern_White_Space, some ASCII controls are expected
	if propertyName == "Pattern_White_Space" {
		return r == '\t' || r == '\n' || r == '\r' || r == '\f' || r == '\v' || r == ' '
	}
	// For other properties, ASCII controls generally don't have them
	return false
}
