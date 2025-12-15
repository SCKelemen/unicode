package uax9

import (
	"fmt"
	"testing"
)

// TestENEdgeCases tests European Number edge cases
func TestENEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		classes        []BidiClass
		paraLevel      int
		expectedLevels []int
	}{
		{
			name:           "R LRE EN - EN in LRE embedding",
			classes:        []BidiClass{ClassR, ClassLRE, ClassEN},
			paraLevel:      0,
			expectedLevels: []int{1, -1, 2},
		},
		{
			name:           "R EN ET - EN followed by ET",
			classes:        []BidiClass{ClassR, ClassEN, ClassET},
			paraLevel:      0,
			expectedLevels: []int{1, 2, 2},
		},
		{
			name:           "R ET EN - ET followed by EN",
			classes:        []BidiClass{ClassR, ClassET, ClassEN},
			paraLevel:      0,
			expectedLevels: []int{1, 2, 2},
		},
		{
			name:           "AL LRE EN - EN in LRE after AL",
			classes:        []BidiClass{ClassAL, ClassLRE, ClassEN},
			paraLevel:      0,
			expectedLevels: []int{1, -1, 2},
		},
		{
			name:           "R LRI EN - EN in LRI isolate",
			classes:        []BidiClass{ClassR, ClassLRI, ClassEN},
			paraLevel:      0,
			expectedLevels: []int{1, 0, 2},
		},
		{
			name:           "AL FSI EN - EN in FSI isolate",
			classes:        []BidiClass{ClassAL, ClassFSI, ClassEN},
			paraLevel:      0,
			expectedLevels: []int{1, 0, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualLevels := computeLevels(tt.classes, tt.paraLevel)

			match := true
			if len(actualLevels) != len(tt.expectedLevels) {
				match = false
			} else {
				for i := range tt.expectedLevels {
					if tt.expectedLevels[i] != -1 && tt.expectedLevels[i] != actualLevels[i] {
						match = false
						break
					}
				}
			}

			if !match {
				t.Errorf("\n  Classes:  %v\n  Expected: %v\n  Got:      %v",
					classesToString(tt.classes), tt.expectedLevels, actualLevels)
			}
		})
	}
}

// TestETEdgeCases tests European Terminator edge cases
func TestETEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		classes        []BidiClass
		paraLevel      int
		expectedLevels []int
	}{
		{
			name:           "RLE EN ET - ET after EN in RLE",
			classes:        []BidiClass{ClassRLE, ClassEN, ClassET},
			paraLevel:      0,
			expectedLevels: []int{-1, 2, 2},
		},
		{
			name:           "RLE ET EN - ET before EN in RLE",
			classes:        []BidiClass{ClassRLE, ClassET, ClassEN},
			paraLevel:      0,
			expectedLevels: []int{-1, 2, 2},
		},
		{
			name:           "RLI EN ET - ET after EN in RLI",
			classes:        []BidiClass{ClassRLI, ClassEN, ClassET},
			paraLevel:      0,
			expectedLevels: []int{0, 2, 2},
		},
		{
			name:           "RLI ET EN - ET before EN in RLI",
			classes:        []BidiClass{ClassRLI, ClassET, ClassEN},
			paraLevel:      0,
			expectedLevels: []int{0, 2, 2},
		},
		{
			name:           "EN ET L - ET between EN and L",
			classes:        []BidiClass{ClassEN, ClassET, ClassL},
			paraLevel:      1,
			expectedLevels: []int{2, 2, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualLevels := computeLevels(tt.classes, tt.paraLevel)

			match := true
			if len(actualLevels) != len(tt.expectedLevels) {
				match = false
			} else {
				for i := range tt.expectedLevels {
					if tt.expectedLevels[i] != -1 && tt.expectedLevels[i] != actualLevels[i] {
						match = false
						break
					}
				}
			}

			if !match {
				t.Errorf("\n  Classes:  %v\n  Expected: %v\n  Got:      %v",
					classesToString(tt.classes), tt.expectedLevels, actualLevels)
			}
		})
	}
}

// TestSeparatorEdgeCases tests CS/ES separator edge cases
func TestSeparatorEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		classes        []BidiClass
		paraLevel      int
		expectedLevels []int
	}{
		{
			name:           "AN ES AN - ES between two AN",
			classes:        []BidiClass{ClassAN, ClassES, ClassAN},
			paraLevel:      0,
			expectedLevels: []int{2, 1, 2}, // ES becomes ON per W6, then resolves to level 1
		},
		{
			name:           "AN CS R - CS between AN and R",
			classes:        []BidiClass{ClassAN, ClassCS, ClassR},
			paraLevel:      0,
			expectedLevels: []int{2, 1, 1},
		},
		{
			name:           "AL CS EN - CS between AL and EN",
			classes:        []BidiClass{ClassAL, ClassCS, ClassEN},
			paraLevel:      0,
			expectedLevels: []int{1, 1, 2},
		},
		{
			name:           "ET EN ES - ES after EN",
			classes:        []BidiClass{ClassET, ClassEN, ClassES},
			paraLevel:      1,
			expectedLevels: []int{2, 2, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualLevels := computeLevels(tt.classes, tt.paraLevel)

			match := true
			if len(actualLevels) != len(tt.expectedLevels) {
				match = false
			} else {
				for i := range tt.expectedLevels {
					if tt.expectedLevels[i] != -1 && tt.expectedLevels[i] != actualLevels[i] {
						match = false
						break
					}
				}
			}

			if !match {
				t.Errorf("\n  Classes:  %v\n  Expected: %v\n  Got:      %v",
					classesToString(tt.classes), tt.expectedLevels, actualLevels)
			}
		})
	}
}

// TestNSMAfterRemovedChars tests NSM after removed characters
func TestNSMAfterRemovedChars(t *testing.T) {
	tests := []struct {
		name           string
		classes        []BidiClass
		paraLevel      int
		expectedLevels []int
	}{
		{
			name:           "L RLE NSM - NSM after RLE",
			classes:        []BidiClass{ClassL, ClassRLE, ClassNSM},
			paraLevel:      0,
			expectedLevels: []int{0, -1, 1},
		},
		{
			name:           "AN RLE NSM - NSM after RLE following AN",
			classes:        []BidiClass{ClassAN, ClassRLE, ClassNSM},
			paraLevel:      0,
			expectedLevels: []int{2, -1, 1},
		},
		{
			name:           "R LRE NSM - NSM after LRE following R",
			classes:        []BidiClass{ClassR, ClassLRE, ClassNSM},
			paraLevel:      0,
			expectedLevels: []int{1, -1, 2},
		},
		{
			name:           "NSM LRE NSM - NSM after LRE, preceded by NSM",
			classes:        []BidiClass{ClassNSM, ClassLRE, ClassNSM},
			paraLevel:      1,
			expectedLevels: []int{1, -1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualLevels := computeLevels(tt.classes, tt.paraLevel)

			match := true
			if len(actualLevels) != len(tt.expectedLevels) {
				match = false
			} else {
				for i := range tt.expectedLevels {
					if tt.expectedLevels[i] != -1 && tt.expectedLevels[i] != actualLevels[i] {
						match = false
						break
					}
				}
			}

			if !match {
				t.Errorf("\n  Classes:  %v\n  Expected: %v\n  Got:      %v",
					classesToString(tt.classes), tt.expectedLevels, actualLevels)
			}
		})
	}
}

// Helper function to convert classes to readable string
func classesToString(classes []BidiClass) string {
	result := ""
	for i, c := range classes {
		if i > 0 {
			result += " "
		}
		result += c.String()
	}
	return result
}

// TestDebugSpecificFailure helps debug a specific failure
func TestDebugSpecificFailure(t *testing.T) {
	// R LRE EN - line 13568
	classes := []BidiClass{ClassR, ClassLRE, ClassEN}
	paraLevel := 0
	expectedLevels := []int{1, -1, 2}

	t.Logf("=== Debugging R LRE EN ===")
	t.Logf("Input classes: %v", classesToString(classes))
	t.Logf("Para level: %d", paraLevel)

	// Step by step
	n := len(classes)
	levels := make([]int, n)
	for i := range levels {
		levels[i] = paraLevel
	}
	t.Logf("Initial levels: %v", levels)

	originalClasses := make([]BidiClass, n)
	copy(originalClasses, classes)

	// Process explicit levels
	processExplicitLevels(classes, levels, paraLevel)
	t.Logf("After explicit levels: %v", levels)
	t.Logf("Classes after explicit: %v", classesToString(classes))

	// Detect empty isolates
	isEmptyIsolate := detectEmptyIsolates(classes, levels)
	t.Logf("Empty isolates: %v", isEmptyIsolate)

	// Resolve weak types
	classesCopy := make([]BidiClass, n)
	copy(classesCopy, classes)
	resolveWeakTypes(classesCopy, levels)
	t.Logf("After weak types: %v", levels)
	t.Logf("Classes after weak: %v", classesToString(classesCopy))

	// Resolve neutral types
	resolveNeutralTypes(classesCopy, levels, paraLevel, isEmptyIsolate)
	t.Logf("After neutral types: %v", levels)
	t.Logf("Classes after neutral: %v", classesToString(classesCopy))

	// Resolve implicit levels
	resolveImplicitLevels(classesCopy, levels)
	t.Logf("After implicit levels: %v", levels)

	// Apply L1
	applyL1(originalClasses, levels, paraLevel)
	t.Logf("After L1: %v", levels)

	t.Logf("\nExpected: %v", expectedLevels)
	t.Logf("Got:      %v", levels)

	if fmt.Sprintf("%v", levels) != fmt.Sprintf("%v", expectedLevels) {
		t.Errorf("Mismatch!")
	}
}
