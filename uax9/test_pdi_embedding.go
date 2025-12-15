package uax9

import "fmt"

func TestPDIEmbedding() {
	// Test case: S RLE PDI ET with para=0
	classes := []BidiClass{ClassS, ClassRLE, ClassPDI, ClassET}
	paraLevel := 0
	
	fmt.Printf("Original classes: %v\n", classes)
	
	levels := make([]int, len(classes))
	for i := range levels {
		levels[i] = paraLevel
	}
	originalClasses := make([]BidiClass, len(classes))
	copy(originalClasses, classes)
	
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
	
	// Apply L1
	applyL1(originalClasses, levels, paraLevel)
	fmt.Printf("After L1:\n")
	fmt.Printf("  Levels: %v\n", levels)
	
	fmt.Printf("\nExpected levels: [0, -1, 1, 1]\n")
	fmt.Printf("Actual levels:   %v\n", levels)
}
