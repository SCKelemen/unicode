package uax9

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func TestAnalyzeRemainingFailures(t *testing.T) {
	file, err := os.Open("BidiTest.txt")
	if err != nil {
		t.Skipf("BidiTest.txt not found: %v", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var expectedLevels []int
	var expectedReorder []int

	failuresByLength := make(map[int]int)
	failuresByFirstClass := make(map[BidiClass]int)
	reorderFailures := 0
	levelFailures := 0
	totalTests := 0
	totalFailed := 0
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

			totalTests++

			// Compute levels and reorder
			// Make copies since computeLevels modifies the classes array
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
				totalFailed++
				failuresByLength[len(classes)]++
				if len(classes) > 0 {
					failuresByFirstClass[classes[0]]++
				}

				if !levelsMatch {
					levelFailures++
				}
				if !reorderMatch {
					reorderFailures++
				}
			}
		}
	}

	passRate := float64(totalTests-totalFailed) / float64(totalTests) * 100
	t.Logf("\n=== REMAINING FAILURE ANALYSIS ===")
	t.Logf("Total tests: %d", totalTests)
	t.Logf("Total failed: %d", totalFailed)
	t.Logf("Pass rate: %.1f%%", passRate)
	t.Logf("\nFailure breakdown:")
	t.Logf("  Level failures: %d (%.1f%%)", levelFailures, float64(levelFailures)/float64(totalFailed)*100)
	t.Logf("  Reorder failures: %d (%.1f%%)", reorderFailures, float64(totalFailed)*100)
	t.Logf("  (Note: some failures have both level and reorder issues)")

	t.Logf("\nTop failures by input length:")
	type lengthKV struct {
		length int
		count  int
	}
	var lengthPairs []lengthKV
	for k, v := range failuresByLength {
		lengthPairs = append(lengthPairs, lengthKV{k, v})
	}
	// Sort
	for i := 0; i < len(lengthPairs); i++ {
		for j := i + 1; j < len(lengthPairs); j++ {
			if lengthPairs[j].count > lengthPairs[i].count {
				lengthPairs[i], lengthPairs[j] = lengthPairs[j], lengthPairs[i]
			}
		}
	}
	for i := 0; i < 10 && i < len(lengthPairs); i++ {
		t.Logf("  Length %d: %d failures (%.1f%%)",
			lengthPairs[i].length, lengthPairs[i].count,
			float64(lengthPairs[i].count)/float64(totalFailed)*100)
	}

	t.Logf("\nTop failures by first character class:")
	type classKV struct {
		class BidiClass
		count int
	}
	var classPairs []classKV
	for k, v := range failuresByFirstClass {
		classPairs = append(classPairs, classKV{k, v})
	}
	for i := 0; i < len(classPairs); i++ {
		for j := i + 1; j < len(classPairs); j++ {
			if classPairs[j].count > classPairs[i].count {
				classPairs[i], classPairs[j] = classPairs[j], classPairs[i]
			}
		}
	}
	for i := 0; i < 15 && i < len(classPairs); i++ {
		t.Logf("  %s: %d failures (%.1f%%)",
			classPairs[i].class.String(), classPairs[i].count,
			float64(classPairs[i].count)/float64(totalFailed)*100)
	}
}
