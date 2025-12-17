package uts15

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

const normalizationTestURL = "https://www.unicode.org/Public/17.0.0/ucd/NormalizationTest.txt"

func TestOfficialNormalization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping official conformance tests in short mode")
	}

	// Download the test file
	resp, err := http.Get(normalizationTestURL)
	if err != nil {
		t.Fatalf("Failed to download test file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("HTTP error: %d", resp.StatusCode)
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
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "@") {
			continue
		}

		// Parse test line: source; NFC; NFD; NFKC; NFKD
		parts := strings.Split(line, ";")
		if len(parts) < 5 {
			continue
		}

		// Parse each column
		c1 := parseCodePoints(parts[0])
		c2 := parseCodePoints(parts[1])
		c3 := parseCodePoints(parts[2])
		c4 := parseCodePoints(parts[3])
		c5 := parseCodePoints(parts[4])

		testCount++

		// Test NFC invariants
		// c2 == toNFC(c1) == toNFC(c2) == toNFC(c3)
		if !testEqual(c2, NFC(c1), "NFC(c1)", lineNum, &failCount, &firstFailures) ||
			!testEqual(c2, NFC(c2), "NFC(c2)", lineNum, &failCount, &firstFailures) ||
			!testEqual(c2, NFC(c3), "NFC(c3)", lineNum, &failCount, &firstFailures) {
			// Failure recorded
		}
		// c4 == toNFC(c4) == toNFC(c5)
		if !testEqual(c4, NFC(c4), "NFC(c4)", lineNum, &failCount, &firstFailures) ||
			!testEqual(c4, NFC(c5), "NFC(c5)", lineNum, &failCount, &firstFailures) {
			// Failure recorded
		}

		// Test NFD invariants
		// c3 == toNFD(c1) == toNFD(c2) == toNFD(c3)
		if !testEqual(c3, NFD(c1), "NFD(c1)", lineNum, &failCount, &firstFailures) ||
			!testEqual(c3, NFD(c2), "NFD(c2)", lineNum, &failCount, &firstFailures) ||
			!testEqual(c3, NFD(c3), "NFD(c3)", lineNum, &failCount, &firstFailures) {
			// Failure recorded
		}
		// c5 == toNFD(c4) == toNFD(c5)
		if !testEqual(c5, NFD(c4), "NFD(c4)", lineNum, &failCount, &firstFailures) ||
			!testEqual(c5, NFD(c5), "NFD(c5)", lineNum, &failCount, &firstFailures) {
			// Failure recorded
		}

		// Test NFKC invariants
		// c4 == toNFKC(c1) == toNFKC(c2) == toNFKC(c3) == toNFKC(c4) == toNFKC(c5)
		if !testEqual(c4, NFKC(c1), "NFKC(c1)", lineNum, &failCount, &firstFailures) ||
			!testEqual(c4, NFKC(c2), "NFKC(c2)", lineNum, &failCount, &firstFailures) ||
			!testEqual(c4, NFKC(c3), "NFKC(c3)", lineNum, &failCount, &firstFailures) ||
			!testEqual(c4, NFKC(c4), "NFKC(c4)", lineNum, &failCount, &firstFailures) ||
			!testEqual(c4, NFKC(c5), "NFKC(c5)", lineNum, &failCount, &firstFailures) {
			// Failure recorded
		}

		// Test NFKD invariants
		// c5 == toNFKD(c1) == toNFKD(c2) == toNFKD(c3) == toNFKD(c4) == toNFKD(c5)
		if !testEqual(c5, NFKD(c1), "NFKD(c1)", lineNum, &failCount, &firstFailures) ||
			!testEqual(c5, NFKD(c2), "NFKD(c2)", lineNum, &failCount, &firstFailures) ||
			!testEqual(c5, NFKD(c3), "NFKD(c3)", lineNum, &failCount, &firstFailures) ||
			!testEqual(c5, NFKD(c4), "NFKD(c4)", lineNum, &failCount, &firstFailures) ||
			!testEqual(c5, NFKD(c5), "NFKD(c5)", lineNum, &failCount, &firstFailures) {
			// Failure recorded
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading test file: %v", err)
	}

	// Report results
	if failCount > 0 {
		t.Errorf("FAILED: %d/%d tests failed", failCount, testCount)
		t.Logf("First failures:")
		for i, fail := range firstFailures {
			if i >= 10 {
				t.Logf("  ... and %d more failures", len(firstFailures)-10)
				break
			}
			t.Logf("  %s", fail)
		}
	} else {
		t.Logf("PASSED: %d/%d tests (100%% conformance)", testCount, testCount)
	}
}

func parseCodePoints(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	parts := strings.Fields(s)
	runes := make([]rune, 0, len(parts))

	for _, part := range parts {
		val, err := strconv.ParseInt(part, 16, 32)
		if err != nil {
			continue
		}
		runes = append(runes, rune(val))
	}

	return string(runes)
}

func testEqual(expected, actual, label string, lineNum int, failCount *int, firstFailures *[]string) bool {
	if expected != actual {
		*failCount++
		if len(*firstFailures) < 100 {
			msg := fmt.Sprintf("Line %d: %s failed: expected % X, got % X",
				lineNum, label, []rune(expected), []rune(actual))
			*firstFailures = append(*firstFailures, msg)
		}
		return false
	}
	return true
}
