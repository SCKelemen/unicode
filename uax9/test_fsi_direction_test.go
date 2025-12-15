package uax9

import "testing"

func TestFSIDirection(t *testing.T) {
	// Test what direction FSI chooses for "FSI EN R"
	_ = []BidiClass{ClassFSI, ClassEN, ClassR}

	// Manually trace through the FSI logic
	t.Logf("Input: FSI EN R")
	t.Logf("FSI scans ahead for first strong:")
	t.Logf("  - Position 1: EN (European Number) - this is treated as L for FSI purposes")
	t.Logf("  - So FSI should become LRI")

	t.Logf("\nBut test expects FSI to behave like RLI:")
	t.Logf("  - RLI EN R para=0: [0, 2, 1]")
	t.Logf("  - FSI EN R para=0: [0, 2, 1] (same)")

	t.Logf("\nThis suggests EN should NOT be treated as strong L for FSI scanning")
	t.Logf("FSI should look for L, R, or AL - not EN")
	t.Logf("If no L/R/AL found before PDI or end, use embedding direction")

	t.Logf("\nIn 'FSI EN R':")
	t.Logf("  - EN is not L/R/AL, so skip it")
	t.Logf("  - R is R, so FSI becomes RLI")
	t.Logf("  - That matches the expected behavior!")
}
