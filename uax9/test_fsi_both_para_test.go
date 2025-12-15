package uax9

import "testing"

func TestFSIBothPara(t *testing.T) {
	classes := []BidiClass{ClassFSI, ClassEN, ClassR}

	// Para 0
	classesCopy0 := make([]BidiClass, len(classes))
	copy(classesCopy0, classes)
	levels0 := computeLevels(classesCopy0, 0)
	t.Logf("FSI EN R para=0:")
	t.Logf("  Actual:   %v", levels0)
	t.Logf("  Expected: [0, 2, 1]")

	// Para 1
	classesCopy1 := make([]BidiClass, len(classes))
	copy(classesCopy1, classes)
	levels1 := computeLevels(classesCopy1, 1)
	t.Logf("\nFSI EN R para=1:")
	t.Logf("  Actual:   %v", levels1)
	t.Logf("  Expected: [1, 4, 3]")

	// Compare with LRI
	classesLRI := []BidiClass{ClassLRI, ClassEN, ClassR}
	classesCopyLRI := make([]BidiClass, len(classesLRI))
	copy(classesCopyLRI, classesLRI)
	levelsLRI := computeLevels(classesCopyLRI, 0)
	t.Logf("\nLRI EN R para=0:")
	t.Logf("  Actual:   %v", levelsLRI)
}
