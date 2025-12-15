package uax9

import "fmt"

func TestPDIPattern() {
	// Test case: ET RLE PDI L with para=0
	classes := []BidiClass{ClassET, ClassRLE, ClassPDI, ClassL}
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
	fmt.Printf("After explicit: Classes=%v, Levels=%v\n", classes, levels)
	
	classesCopy := make([]BidiClass, len(classes))
	copy(classesCopy, classes)
	
	resolveWeakTypes(classesCopy, levels)
	fmt.Printf("After weak: Classes=%v, Levels=%v\n", classesCopy, levels)
	
	resolveNeutralTypes(classesCopy, levels, paraLevel)
	fmt.Printf("After neutral: Classes=%v, Levels=%v\n", classesCopy, levels)
	
	resolveImplicitLevels(classesCopy, levels)
	fmt.Printf("After implicit: Classes=%v, Levels=%v\n", classesCopy, levels)
	
	applyL1(originalClasses, levels, paraLevel)
	fmt.Printf("After L1: Levels=%v\n", levels)
	
	fmt.Printf("\nExpected: [0, -1, 1, 2]\n")
	fmt.Printf("Actual:   %v\n", levels)
}
