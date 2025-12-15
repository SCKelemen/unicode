package uax9

import (
	"testing"
)

func TestDebugReorder(t *testing.T) {
	// Test case: R LRO L with para=0
	// Expected levels: [1, -1, 0]
	// Expected reorder: [0, 2]

	classes := []BidiClass{ClassR, ClassLRO, ClassL}
	paraLevel := 0

	actualLevels := computeLevels(classes, paraLevel)
	t.Logf("Levels: %v", actualLevels)

	actualReorder := computeReorder(classes, paraLevel)
	t.Logf("Reorder: %v", actualReorder)

	// Also test with [1, -1, 1]
	classes2 := []BidiClass{ClassR, ClassLRE, ClassB}
	actualLevels2 := computeLevels(classes2, paraLevel)
	t.Logf("Levels for R LRE B: %v", actualLevels2)

	actualReorder2 := computeReorder(classes2, paraLevel)
	t.Logf("Reorder for R LRE B: %v", actualReorder2)
}
