package uax9

import "testing"

func TestTraceExplicit(t *testing.T) {
	classes := []BidiClass{ClassFSI, ClassEN, ClassR}
	para := 0

	levels := make([]int, len(classes))
	for i := range levels {
		levels[i] = para
	}

	t.Logf("Before explicit processing:")
	t.Logf("  Classes: %v", classesToString(classes))
	t.Logf("  Levels:  %v", levels)

	processExplicitLevels(classes, levels, para)

	t.Logf("\nAfter explicit processing:")
	t.Logf("  Classes: %v", classesToString(classes))
	t.Logf("  Levels:  %v", levels)
	t.Logf("\nExpected levels: [0, 2, 1]")
	t.Logf("This means:")
	t.Logf("  - FSI should be at level 0")
	t.Logf("  - EN should be at level 2 (inside LTR isolate)")
	t.Logf("  - R should be at level 1 (outside isolate, at RTL level)")
}
