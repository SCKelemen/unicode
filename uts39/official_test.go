package uts39

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

const confusablesTestURL = "https://www.unicode.org/Public/security/latest/confusables.txt"

// TestOfficialConfusables verifies our confusables data against the official Unicode file
func TestOfficialConfusables(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping official conformance tests in short mode")
	}

	// Download confusables.txt
	resp, err := http.Get(confusablesTestURL)
	if err != nil {
		t.Fatalf("Failed to download confusables.txt: %v", err)
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
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse line: SOURCE ; TARGET ; TYPE # COMMENT
		parts := strings.Split(line, ";")
		if len(parts) < 3 {
			continue
		}

		// Parse source code point
		sourceStr := strings.TrimSpace(parts[0])
		sourceVal, err := strconv.ParseInt(sourceStr, 16, 32)
		if err != nil {
			continue
		}
		source := rune(sourceVal)

		// Parse target code point(s)
		targetStr := strings.TrimSpace(parts[1])
		targetRunes := parseCodePointsForTest(targetStr)
		if len(targetRunes) == 0 {
			continue
		}
		expectedTarget := string(targetRunes)

		testCount++

		// Look up in our data
		actualTarget := getConfusableTarget(source)

		if actualTarget != expectedTarget {
			failCount++
			if len(firstFailures) < 100 {
				firstFailures = append(firstFailures,
					fmt.Sprintf("Line %d: U+%04X -> expected %q, got %q",
						lineNum, source, expectedTarget, actualTarget))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading confusables.txt: %v", err)
	}

	// Report results
	if failCount > 0 {
		t.Errorf("FAILED: %d/%d confusable mappings incorrect", failCount, testCount)
		t.Logf("First failures:")
		for i, fail := range firstFailures {
			if i >= 10 {
				t.Logf("  ... and %d more failures", len(firstFailures)-10)
				break
			}
			t.Logf("  %s", fail)
		}
	} else {
		t.Logf("PASSED: %d/%d confusable mappings verified (100%% conformance)", testCount, testCount)
	}
}

// TestConfusableExamples tests known confusable pairs from the Unicode documentation
func TestConfusableExamples(t *testing.T) {
	tests := []struct {
		name string
		s1   string
		s2   string
		want bool
	}{
		{
			name: "Cyrillic vs Latin (scope example)",
			s1:   "scope",
			s2:   "\u0455\u0441\u043E\u0440\u0435", // ѕсоре (Cyrillic)
			want: true,
		},
		{
			name: "Greek vs Latin",
			s1:   "apple",
			s2:   "\u0430pple", // аpple (Cyrillic а)
			want: true,
		},
		{
			name: "Different words",
			s1:   "hello",
			s2:   "world",
			want: false,
		},
		{
			name: "Mathematical vs ASCII",
			s1:   "x",
			s2:   "\U0001D465", // Mathematical italic x
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AreConfusable(tt.s1, tt.s2)
			if got != tt.want {
				skel1 := Skeleton(tt.s1)
				skel2 := Skeleton(tt.s2)
				t.Errorf("AreConfusable(%q, %q) = %v, want %v\n  skeleton1: %q\n  skeleton2: %q",
					tt.s1, tt.s2, got, tt.want, skel1, skel2)
			}
		})
	}
}

// TestSkeletonAlgorithm verifies the skeleton algorithm follows UTS #39 spec
func TestSkeletonAlgorithm(t *testing.T) {
	// Test that skeleton is idempotent
	tests := []string{
		"hello",
		"café",
		"世界",
		"test123",
		"pаypal", // Contains Cyrillic а
	}

	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			skel1 := Skeleton(s)
			skel2 := Skeleton(skel1)
			if skel1 != skel2 {
				t.Errorf("Skeleton not idempotent: Skeleton(%q) = %q, Skeleton(%q) = %q",
					s, skel1, skel1, skel2)
			}
		})
	}
}

// TestRestrictionLevelConformance tests restriction level detection
func TestRestrictionLevelConformance(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		minLevel RestrictionLevel
		maxLevel RestrictionLevel
	}{
		{"ASCII", "hello_world", ASCIIOnly, ASCIIOnly},
		{"Latin", "café", SingleScript, SingleScript},
		{"Cyrillic", "привет", SingleScript, SingleScript},
		{"Han", "你好", SingleScript, SingleScript},
		{"Latin+Han", "hello世界", MinimallyRestrictive, MinimallyRestrictive},
		{"Latin+Cyrillic", "hello мир", MinimallyRestrictive, MinimallyRestrictive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := GetRestrictionLevel(tt.input)
			if level < tt.minLevel || level > tt.maxLevel {
				t.Errorf("GetRestrictionLevel(%q) = %v, want between %v and %v",
					tt.input, level, tt.minLevel, tt.maxLevel)
			}
		})
	}
}

func parseCodePointsForTest(s string) []rune {
	parts := strings.Fields(s)
	runes := make([]rune, 0, len(parts))

	for _, part := range parts {
		val, err := strconv.ParseInt(part, 16, 32)
		if err != nil {
			continue
		}
		runes = append(runes, rune(val))
	}

	return runes
}
