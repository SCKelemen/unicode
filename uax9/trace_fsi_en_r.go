package uax9

import "fmt"

func TraceFSIENR() {
	classes := []BidiClass{ClassFSI, ClassEN, ClassR}
	paraLevel := 0
	
	fmt.Printf("=== FSI EN R with para=0 ===\n")
	fmt.Printf("Input: %v\n\n", classesToString(classes))
	
	levels := make([]int, len(classes))
	for i := range levels {
		levels[i] = paraLevel
	}
	
	originalClasses := make([]BidiClass, len(classes))
	copy(originalClasses, classes)
	
	processExplicitLevels(classes, levels, paraLevel)
	fmt.Printf("After explicit levels:\n")
	fmt.Printf("  Classes: %v\n", classesToString(classes))
	fmt.Printf("  Levels:  %v\n\n", levels)
	
	classesCopy := make([]BidiClass, len(classes))
	copy(classesCopy, classes)
	
	resolveWeakTypes(classesCopy, levels)
	fmt.Printf("After weak types:\n")
	fmt.Printf("  Classes: %v\n", classesToString(classesCopy))
	fmt.Printf("  Levels:  %v\n\n", levels)
	
	resolveNeutralTypes(classesCopy, levels, paraLevel)
	fmt.Printf("After neutral types:\n")
	fmt.Printf("  Classes: %v\n", classesToString(classesCopy))
	fmt.Printf("  Levels:  %v\n\n", levels)
	
	resolveImplicitLevels(classesCopy, levels)
	fmt.Printf("After implicit levels:\n")
	fmt.Printf("  Classes: %v\n", classesToString(classesCopy))
	fmt.Printf("  Levels:  %v\n\n", levels)
	
	applyL1(originalClasses, levels, paraLevel)
	fmt.Printf("After L1:\n")
	fmt.Printf("  Levels:  %v\n\n", levels)
	
	fmt.Printf("Expected: [0, 2, 1]\n")
	fmt.Printf("Actual:   %v\n", levels)
}
