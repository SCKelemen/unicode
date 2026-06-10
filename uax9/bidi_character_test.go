package uax9

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

// TestOfficialBidiCharacterTest runs the official Unicode BidiCharacterTest.txt
// conformance vectors against the implementation. It exercises the algorithm
// on actual code-point sequences (whereas BidiTest.txt operates on synthetic
// bidi class sequences).
//
// File format (per the header of BidiCharacterTest-17.0.0.txt):
//
//	Field 0: hex code points, space separated
//	Field 1: paragraph direction (0=LTR, 1=RTL, 2=auto-LTR via P2/P3)
//	Field 2: resolved paragraph embedding level
//	Field 3: resolved levels per character ('x' = removed by X9)
//	Field 4: visual ordering (indices into the input, left-to-right)
//
// Reference: https://www.unicode.org/Public/17.0.0/ucd/BidiCharacterTest.txt
func TestOfficialBidiCharacterTest(t *testing.T) {
	file, err := os.Open("BidiCharacterTest.txt")
	if err != nil {
		t.Skipf("BidiCharacterTest.txt not found: %v "+
			"(download from https://www.unicode.org/Public/17.0.0/ucd/BidiCharacterTest.txt)", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Some lines in BidiCharacterTest.txt are long; bump the buffer.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var (
		lineNum       int
		testCount     int
		paraLevelPass int
		paraLevelFail int
		levelsPass    int
		levelsFail    int
		firstFails    int
	)

	for scanner.Scan() {
		lineNum++
		raw := scanner.Text()
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Split(line, ";")
		if len(fields) < 4 {
			continue
		}

		// Field 0: code points.
		cpTokens := strings.Fields(fields[0])
		runes := make([]rune, 0, len(cpTokens))
		for _, tok := range cpTokens {
			cp, err := strconv.ParseInt(tok, 16, 32)
			if err != nil {
				continue
			}
			runes = append(runes, rune(cp))
		}
		if len(runes) == 0 {
			continue
		}

		// Field 1: requested paragraph direction.
		paraDir, err := strconv.Atoi(strings.TrimSpace(fields[1]))
		if err != nil {
			continue
		}

		// Field 2: expected resolved paragraph embedding level.
		expectedParaLevel, err := strconv.Atoi(strings.TrimSpace(fields[2]))
		if err != nil {
			continue
		}

		// Field 3: expected resolved levels.
		levelToks := strings.Fields(fields[3])
		if len(levelToks) != len(runes) {
			continue
		}
		expectedLevels := make([]int, len(levelToks))
		for i, lt := range levelToks {
			if lt == "x" {
				expectedLevels[i] = -1
				continue
			}
			lv, err := strconv.Atoi(lt)
			if err != nil {
				expectedLevels[i] = -1
				continue
			}
			expectedLevels[i] = lv
		}

		// Map runes to bidi classes.
		classes := make([]BidiClass, len(runes))
		for i, r := range runes {
			classes[i] = GetBidiClass(r)
		}

		// Resolve paragraph level.
		var paraLevel int
		switch paraDir {
		case 0:
			paraLevel = 0
		case 1:
			paraLevel = 1
		case 2:
			// Auto: drive through the public API so we exercise the
			// fixed P2 path in GetParagraphDirection.
			if GetParagraphDirection(string(runes)) == DirectionRTL {
				paraLevel = 1
			} else {
				paraLevel = 0
			}
		default:
			continue
		}

		testCount++

		// Compare the resolved paragraph embedding level.
		if paraLevel == expectedParaLevel {
			paraLevelPass++
		} else {
			paraLevelFail++
			if firstFails < 5 {
				firstFails++
				t.Logf("line %d: para level mismatch (paraDir=%d): expected=%d got=%d input=%s",
					lineNum, paraDir, expectedParaLevel, paraLevel, fields[0])
			}
		}

		// Compute resolved levels via the rune-aware entry point so
		// that N0 paired-bracket resolution exercises the runes.
		classesCopy := make([]BidiClass, len(classes))
		copy(classesCopy, classes)
		actualLevels := ComputeLevelsFromRunes(runes, classesCopy, paraLevel)

		if len(actualLevels) != len(expectedLevels) {
			levelsFail++
			continue
		}
		match := true
		for i := range expectedLevels {
			if expectedLevels[i] == -1 {
				// 'x' means the position was removed by X9; the
				// implementation is free to report any level here.
				continue
			}
			if actualLevels[i] != expectedLevels[i] {
				match = false
				break
			}
		}
		if match {
			levelsPass++
		} else {
			levelsFail++
			if firstFails < 5 {
				firstFails++
				t.Logf("line %d: levels mismatch: expected=%v got=%v input=%s",
					lineNum, expectedLevels, actualLevels, fields[0])
			}
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner: %v", err)
	}

	if testCount == 0 {
		t.Fatalf("no BidiCharacterTest cases were parsed; file may be malformed")
	}

	t.Logf("\nOfficial BidiCharacterTest Results:")
	t.Logf("  Total tests: %d", testCount)
	t.Logf("  Paragraph level: %d passed, %d failed (%.2f%%)",
		paraLevelPass, paraLevelFail, 100.0*float64(paraLevelPass)/float64(testCount))
	t.Logf("  Resolved levels: %d passed, %d failed (%.2f%%)",
		levelsPass, levelsFail, 100.0*float64(levelsPass)/float64(testCount))

	// Paragraph-level resolution must be exact: P2/P3 is a simple,
	// well-understood part of the algorithm.
	if paraLevelFail != 0 {
		t.Errorf("paragraph-level mismatches: %d (must be 0)", paraLevelFail)
	}

	// Keep a tight regression floor against the current official
	// Unicode 17.0.0 baseline. At the time of this update the runner
	// passes 91,707 / 91,707 resolved-level vectors (100.00%), so the
	// requested floor max(0.999, actualPassRate-0.001) is 0.999.
	//
	// Historically the last stubborn vectors involved isolate formatting
	// character level adjustment in nested explicit-formatting scopes;
	// if this regresses, start by looking there before widening the
	// threshold.
	const resolvedLevelsFloor = 0.999
	resolvedLevelsPassRate := float64(levelsPass) / float64(testCount)
	if resolvedLevelsPassRate < resolvedLevelsFloor {
		t.Errorf("resolved-levels pass rate %.2f%% regressed below %.1f%% floor",
			100.0*resolvedLevelsPassRate, resolvedLevelsFloor*100)
	}
}
