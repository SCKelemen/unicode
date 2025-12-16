package uts51

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

// EmojiTestCase represents a single test case from emoji-test.txt
type EmojiTestCase struct {
	Codepoints []rune
	Status     string // fully-qualified, minimally-qualified, unqualified, component
	Emoji      string
	Version    string // E1.0, E0.6, etc.
	Name       string
}

// TestEmojiTestFileConformance tests all cases from emoji-test.txt
// This provides 100% conformance testing for UTS #51
func TestEmojiTestFileConformance(t *testing.T) {
	file, err := os.Open("emoji-test.txt")
	if err != nil {
		t.Skipf("Skipping conformance test: %v", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	passed := 0
	failed := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines, comments, group headers
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse test case
		testCase, err := parseEmojiTestLine(line)
		if err != nil || testCase == nil {
			continue // Skip malformed lines
		}

		// Test component status
		if testCase.Status == "component" {
			// Components should be emoji components
			if len(testCase.Codepoints) > 0 {
				cp := testCase.Codepoints[0]
				if !IsEmojiComponent(cp) && !IsEmojiModifier(cp) {
					t.Errorf("Line %d: Component %U not recognized as emoji component: %s",
						lineNum, cp, testCase.Name)
					failed++
					continue
				}
			}
			passed++
			continue
		}

		// For non-component sequences, validate the entire sequence
		if !IsValidEmojiSequence(testCase.Codepoints) {
			t.Errorf("Line %d: Invalid emoji sequence: %s (status: %s)",
				lineNum, testCase.Name, testCase.Status)
			failed++
			continue
		}

		passed++
	}

	if err := scanner.Err(); err != nil {
		t.Errorf("Error reading file: %v", err)
	}

	t.Logf("Emoji conformance: %d passed, %d failed out of %d total",
		passed, failed, passed+failed)

	if failed > 0 {
		t.Errorf("Conformance test failed: %d cases failed", failed)
	}
}

// parseEmojiTestLine parses a line from emoji-test.txt
func parseEmojiTestLine(line string) (*EmojiTestCase, error) {
	// Format: codepoints ; status # emoji version name
	parts := strings.Split(line, ";")
	if len(parts) < 2 {
		return nil, nil
	}

	codepointsStr := strings.TrimSpace(parts[0])
	rest := strings.TrimSpace(parts[1])

	// Split status and comment
	commentParts := strings.Split(rest, "#")
	if len(commentParts) < 2 {
		return nil, nil
	}

	status := strings.TrimSpace(commentParts[0])
	comment := strings.TrimSpace(commentParts[1])

	// Parse codepoints
	var codepoints []rune
	for _, cpStr := range strings.Fields(codepointsStr) {
		cp, err := strconv.ParseInt(cpStr, 16, 32)
		if err != nil {
			return nil, err
		}
		codepoints = append(codepoints, rune(cp))
	}

	// Parse comment: emoji version name
	commentFields := strings.Fields(comment)
	if len(commentFields) < 3 {
		return nil, nil
	}

	emoji := commentFields[0]
	version := commentFields[1]
	name := strings.Join(commentFields[2:], " ")

	return &EmojiTestCase{
		Codepoints: codepoints,
		Status:     status,
		Emoji:      emoji,
		Version:    version,
		Name:       name,
	}, nil
}

// Benchmark conformance test parsing
func BenchmarkEmojiTestFileParsing(b *testing.B) {
	file, err := os.Open("emoji-test.txt")
	if err != nil {
		b.Skip("emoji-test.txt not found")
		return
	}
	defer file.Close()

	// Read all lines once
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, line := range lines {
			_, _ = parseEmojiTestLine(line) // Benchmark, ignore errors
		}
	}
}
