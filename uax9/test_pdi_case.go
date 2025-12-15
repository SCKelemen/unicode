package uax9

import "fmt"

func TestPDICase() {
	// Test case: R PDI AL with para=0
	classes := []BidiClass{ClassR, ClassPDI, ClassAL}
	paraLevel := 0
	
	fmt.Printf("Original classes: %v\n", classes)
	
	levels := make([]int, len(classes))
	for i := range levels {
		levels[i] = paraLevel
	}
	
	// Process explicit levels
	processExplicitLevels(classes, levels, paraLevel)
	
	fmt.Printf("After explicit processing:\n")
	fmt.Printf("  Classes: %v\n", classes)
	fmt.Printf("  Levels: %v\n", levels)
	
	// Make a copy for further processing
	classesCopy := make([]BidiClass, len(classes))
	copy(classesCopy, classes)
	
	// Resolve weak types
	resolveWeakTypes(classesCopy, levels)
	fmt.Printf("After weak resolution:\n")
	fmt.Printf("  Classes: %v\n", classesCopy)
	fmt.Printf("  Levels: %v\n", levels)
	
	// Resolve neutral types
	resolveNeutralTypes(classesCopy, levels, paraLevel)
	fmt.Printf("After neutral resolution:\n")
	fmt.Printf("  Classes: %v\n", classesCopy)
	fmt.Printf("  Levels: %v\n", levels)
	
	// Resolve implicit levels
	resolveImplicitLevels(classesCopy, levels)
	fmt.Printf("After implicit resolution:\n")
	fmt.Printf("  Classes: %v\n", classesCopy)
	fmt.Printf("  Levels: %v\n", levels)
	
	fmt.Printf("\nExpected levels: [1, 1, 1]\n")
	fmt.Printf("Actual levels:   %v\n", levels)
}
