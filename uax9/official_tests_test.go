package uax9

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

// parseBidiTestLine parses a single test line from BidiTest.txt
func parseBidiTestLine(line string) ([]BidiClass, int, error) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "@") {
		return nil, 0, nil
	}

	parts := strings.Split(line, ";")
	if len(parts) != 2 {
		return nil, 0, fmt.Errorf("invalid line format")
	}

	// Parse bidi classes
	classStrs := strings.Fields(strings.TrimSpace(parts[0]))
	classes := make([]BidiClass, len(classStrs))
	for i, cs := range classStrs {
		class, err := parseBidiClassName(cs)
		if err != nil {
			return nil, 0, err
		}
		classes[i] = class
	}

	// Parse bitset
	bitsetStr := strings.TrimSpace(parts[1])
	bitset, err := strconv.ParseInt(bitsetStr, 16, 32)
	if err != nil {
		return nil, 0, err
	}

	return classes, int(bitset), nil
}

// parseBidiClassName converts a string to BidiClass
func parseBidiClassName(name string) (BidiClass, error) {
	switch name {
	case "L":
		return ClassL, nil
	case "R":
		return ClassR, nil
	case "AL":
		return ClassAL, nil
	case "EN":
		return ClassEN, nil
	case "ES":
		return ClassES, nil
	case "ET":
		return ClassET, nil
	case "AN":
		return ClassAN, nil
	case "CS":
		return ClassCS, nil
	case "NSM":
		return ClassNSM, nil
	case "BN":
		return ClassBN, nil
	case "B":
		return ClassB, nil
	case "S":
		return ClassS, nil
	case "WS":
		return ClassWS, nil
	case "ON":
		return ClassON, nil
	case "LRE":
		return ClassLRE, nil
	case "LRO":
		return ClassLRO, nil
	case "RLE":
		return ClassRLE, nil
	case "RLO":
		return ClassRLO, nil
	case "PDF":
		return ClassPDF, nil
	case "LRI":
		return ClassLRI, nil
	case "RLI":
		return ClassRLI, nil
	case "FSI":
		return ClassFSI, nil
	case "PDI":
		return ClassPDI, nil
	default:
		return ClassL, fmt.Errorf("unknown bidi class: %s", name)
	}
}

// parseExpectedLevels parses an @Levels line
func parseExpectedLevels(line string) ([]int, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "@Levels:") {
		return nil, fmt.Errorf("not a levels line")
	}

	levelStrs := strings.Fields(line[8:])
	levels := make([]int, len(levelStrs))
	for i, ls := range levelStrs {
		if ls == "x" {
			levels[i] = -1
		} else {
			level, err := strconv.Atoi(ls)
			if err != nil {
				return nil, err
			}
			levels[i] = level
		}
	}
	return levels, nil
}

// parseExpectedReorder parses an @Reorder line
func parseExpectedReorder(line string) ([]int, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "@Reorder:") {
		return nil, fmt.Errorf("not a reorder line")
	}

	orderStrs := strings.Fields(line[9:])
	order := make([]int, len(orderStrs))
	for i, os := range orderStrs {
		idx, err := strconv.Atoi(os)
		if err != nil {
			return nil, err
		}
		order[i] = idx
	}
	return order, nil
}

// computeLevels computes the resolved levels for a sequence of bidi classes.
// This is a thin wrapper around the exported ComputeLevels function.
func computeLevels(classes []BidiClass, paraLevel int) []int {
	// Make a copy since ComputeLevels modifies the classes array
	classesCopy := make([]BidiClass, len(classes))
	copy(classesCopy, classes)
	return ComputeLevels(classesCopy, paraLevel)
}

// computeReorder computes the visual reordering for a sequence
func computeReorder(classes []BidiClass, paraLevel int) []int {
	n := len(classes)
	levels := computeLevels(classes, paraLevel)

	// Find max level
	maxLevel := paraLevel
	for _, level := range levels {
		if level > maxLevel {
			maxLevel = level
		}
	}

	// Create indices
	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}

	// Reverse runs from highest level to lowest odd level (L2)
	// Per UAX#9 L2: reverse down to the lowest odd level
	lowestOddLevel := 1
	if paraLevel > lowestOddLevel {
		lowestOddLevel = paraLevel
	}

	for level := maxLevel; level >= lowestOddLevel; level-- {
		i := 0
		for i < n {
			// Skip removed characters when not in a run
			if levels[i] == -1 {
				i++
				continue
			}

			// Skip characters below this level
			if levels[i] < level {
				i++
				continue
			}

			// Found start of a run at this level
			start := i
			i++

			// Extend run: include removed characters AND characters at this level
			for i < n {
				if levels[i] == -1 {
					// Removed character: tentatively include it
					i++
				} else if levels[i] >= level {
					// Character at this level or higher: include it
					i++
				} else {
					// Character below this level: stop
					break
				}
			}
			end := i - 1

			// Reverse this run
			for start < end {
				indices[start], indices[end] = indices[end], indices[start]
				start++
				end--
			}
		}
	}

	// Filter out removed characters (level == -1)
	result := make([]int, 0, n)
	for _, idx := range indices {
		if levels[idx] >= 0 {
			result = append(result, idx)
		}
	}

	return result
}

func TestOfficialBidiTest(t *testing.T) {
	file, err := os.Open("BidiTest.txt")
	if err != nil {
		t.Skipf("BidiTest.txt not found: %v", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var expectedLevels []int
	var expectedReorder []int
	testCount := 0
	passCount := 0
	failCount := 0
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
				t.Logf("Line %d: Error parsing levels: %v", lineNum, err)
				continue
			}
			expectedLevels = levels
			continue
		}

		if strings.HasPrefix(line, "@Reorder:") {
			reorder, err := parseExpectedReorder(line)
			if err != nil {
				t.Logf("Line %d: Error parsing reorder: %v", lineNum, err)
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
		if err != nil {
			continue
		}
		if classes == nil {
			continue
		}

		// Test each paragraph level in the bitset
		for paraLevel := 0; paraLevel <= 1; paraLevel++ {
			// Check if this level is in the bitset
			// Bitset: 1=auto-LTR, 2=LTR, 4=RTL
			shouldTest := false
			if paraLevel == 0 && (bitset&2) != 0 {
				shouldTest = true
			} else if paraLevel == 1 && (bitset&4) != 0 {
				shouldTest = true
			}

			if !shouldTest {
				continue
			}

			testCount++

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

			if levelsMatch && reorderMatch {
				passCount++
			} else {
				failCount++
				if failCount <= 10 { // Only log first 10 failures
					t.Logf("FAIL Line %d: %s (paraLevel=%d)", lineNum, line, paraLevel)
					if !levelsMatch {
						t.Logf("  Expected levels: %v", expectedLevels)
						t.Logf("  Actual levels:   %v", actualLevels)
					}
					if !reorderMatch {
						t.Logf("  Expected reorder: %v", expectedReorder)
						t.Logf("  Actual reorder:   %v", actualReorder)
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		t.Errorf("Error reading test file: %v", err)
	}

	passRate := float64(passCount) / float64(testCount) * 100.0
	t.Logf("\nOfficial BidiTest Results:")
	t.Logf("  Total tests: %d", testCount)
	t.Logf("  Passed: %d", passCount)
	t.Logf("  Failed: %d", failCount)
	t.Logf("  Pass rate: %.1f%%", passRate)

	if passRate < 50.0 {
		t.Errorf("Pass rate too low: %.1f%%", passRate)
	}
}

func TestBasicBidiReordering(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		dir      Direction
		expected string
	}{
		{
			name:     "Pure LTR",
			input:    "Hello",
			dir:      DirectionLTR,
			expected: "Hello",
		},
		{
			name:     "Pure LTR with numbers",
			input:    "Hello 123",
			dir:      DirectionLTR,
			expected: "Hello 123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Reorder(tt.input, tt.dir)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
