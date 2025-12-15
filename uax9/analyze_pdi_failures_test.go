package uax9

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func TestAnalyzePDIFailures(t *testing.T) {
	file, err := os.Open("BidiTest.txt")
	if err != nil {
		t.Skipf("BidiTest.txt not found: %v", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var expectedLevels []int
	var expectedReorder []int

	pdiFailures := 0
	totalFailures := 0
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

		if strings.HasPrefix(line, "@Reorder:") {
			reorder, err := parseExpectedReorder(line)
			if err != nil {
				continue
			}
			expectedReorder = reorder
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

			// Compute levels and reorder
			classesCopy1 := make([]BidiClass, len(classes))
			copy(classesCopy1, classes)
			actualLevels := computeLevels(classesCopy1, paraLevel)

			classesCopy2 := make([]BidiClass, len(classes))
			copy(classesCopy2, classes)
			actualReorder := computeReorder(classesCopy2, paraLevel)

			// Check levels
			levelsMatch := len(expectedLevels) == len(actualLevels)
			if levelsMatch {
				for i := range expectedLevels {
					if expectedLevels[i] != -1 && expectedLevels[i] != actualLevels[i] {
						levelsMatch = false
						break
					}
				}
			}

			// Check reorder
			reorderMatch := len(expectedReorder) == len(actualReorder)
			if reorderMatch {
				for i := range expectedReorder {
					if expectedReorder[i] != actualReorder[i] {
						reorderMatch = false
						break
					}
				}
			}

			if !levelsMatch || !reorderMatch {
				totalFailures++

				// Check if this involves PDI
				hasPDI := false
				pdiIndex := -1
				for i, c := range classes {
					if c == ClassPDI {
						hasPDI = true
						pdiIndex = i
						break
					}
				}

				if hasPDI && !levelsMatch {
					pdiFailures++
					if pdiFailures <= 10 {
						t.Logf("PDI level failure at line %d: %s (para=%d)", lineNum, classesToString(classes), paraLevel)
						t.Logf("  PDI at index %d", pdiIndex)
						t.Logf("  Expected levels: %v", expectedLevels)
						t.Logf("  Actual levels:   %v", actualLevels)
						t.Logf("")
					}
				}
			}
		}
	}

	t.Logf("\n=== PDI FAILURE ANALYSIS ===")
	t.Logf("Total failures: %d", totalFailures)
	t.Logf("PDI level failures: %d (%.1f%%)", pdiFailures, float64(pdiFailures)/float64(totalFailures)*100)
}
