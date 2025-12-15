package uax9

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func TestInvestigateFSI(t *testing.T) {
	file, err := os.Open("BidiTest.txt")
	if err != nil {
		t.Skipf("BidiTest.txt not found: %v", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var expectedLevels []int

	fsiCases := make(map[string]struct {
		expected []int
		actual   []int
		para     int
	})

	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "@Levels:") {
			levels, err := parseExpectedLevels(line)
			if err != nil {
				continue
			}
			expectedLevels = levels
			continue
		}

		if strings.HasPrefix(line, "@") {
			continue
		}

		// Parse test case
		classes, bitset, err := parseBidiTestLine(line)
		if err != nil || classes == nil {
			continue
		}

		// Only look at FSI cases of length 3
		if len(classes) != 3 {
			continue
		}

		hasFSI := false
		hasPDI := false
		for _, c := range classes {
			if c == ClassFSI {
				hasFSI = true
			}
			if c == ClassPDI {
				hasPDI = true
			}
		}

		if !hasFSI {
			continue
		}

		// Test each paragraph level
		for paraLevel := 0; paraLevel <= 1; paraLevel++ {
			shouldTest := false
			if paraLevel == 0 && (bitset&2) != 0 {
				shouldTest = true
			} else if paraLevel == 1 && (bitset&4) != 0 {
				shouldTest = true
			}

			if !shouldTest {
				continue
			}

			classesCopy := make([]BidiClass, len(classes))
			copy(classesCopy, classes)
			actualLevels := computeLevels(classesCopy, paraLevel)

			// Check if levels match
			levelsMatch := len(expectedLevels) == len(actualLevels)
			if levelsMatch {
				for i := range expectedLevels {
					if expectedLevels[i] != -1 && expectedLevels[i] != actualLevels[i] {
						levelsMatch = false
						break
					}
				}
			}

			if !levelsMatch {
				key := classesToString(classes)
				if !hasPDI {
					key += " (no PDI)"
				}
				fsiCases[key] = struct {
					expected []int
					actual   []int
					para     int
				}{
					expected: expectedLevels,
					actual:   actualLevels,
					para:     paraLevel,
				}
			}
		}
	}

	t.Logf("\n=== FSI FAILURES (length 3) ===")
	t.Logf("Total unique patterns: %d\n", len(fsiCases))

	count := 0
	for pattern, data := range fsiCases {
		count++
		if count <= 20 {
			t.Logf("%s (para=%d)", pattern, data.para)
			t.Logf("  Expected: %v", data.expected)
			t.Logf("  Actual:   %v", data.actual)
		}
	}
}
