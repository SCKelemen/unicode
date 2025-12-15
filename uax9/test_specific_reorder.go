package uax9

import "fmt"

// TestReorderLevels tests reordering for specific level patterns
func TestReorderLevels() {
	// Test case: levels [1, -1, 0]
	levels := []int{1, -1, 0}
	paraLevel := 0

	n := len(levels)
	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}

	// Find max level
	maxLevel := paraLevel
	for _, level := range levels {
		if level > maxLevel {
			maxLevel = level
		}
	}

	lowestOddLevel := 1
	if paraLevel > lowestOddLevel {
		lowestOddLevel = paraLevel
	}

	fmt.Printf("Initial: levels=%v, indices=%v\n", levels, indices)
	fmt.Printf("maxLevel=%d, lowestOddLevel=%d\n", maxLevel, lowestOddLevel)

	for level := maxLevel; level >= lowestOddLevel; level-- {
		fmt.Printf("\nProcessing level %d:\n", level)
		i := 0
		for i < n {
			fmt.Printf("  i=%d, levels[%d]=%d\n", i, i, levels[i])

			// Skip removed characters when not in a run
			if levels[i] == -1 {
				fmt.Printf("    Skipping removed character\n")
				i++
				continue
			}

			// Skip characters below this level
			if levels[i] < level {
				fmt.Printf("    Skipping character below level\n")
				i++
				continue
			}

			// Found start of a run at this level
			start := i
			fmt.Printf("    Found run start at %d\n", start)
			i++

			// Extend run
			for i < n {
				fmt.Printf("      Checking i=%d: levels[%d]=%d\n", i, i, levels[i])
				if levels[i] == -1 {
					fmt.Printf("        Including removed character\n")
					i++
				} else if levels[i] >= level {
					fmt.Printf("        Including character at level >= %d\n", level)
					i++
				} else {
					fmt.Printf("        Stopping (character below level)\n")
					break
				}
			}
			end := i - 1
			fmt.Printf("    Run: [%d..%d], reversing\n", start, end)
			fmt.Printf("    Before reverse: indices=%v\n", indices)

			// Reverse
			for start < end {
				indices[start], indices[end] = indices[end], indices[start]
				start++
				end--
			}
			fmt.Printf("    After reverse: indices=%v\n", indices)
		}
	}

	// Filter removed
	result := make([]int, 0, n)
	for _, idx := range indices {
		if levels[idx] >= 0 {
			result = append(result, idx)
		}
	}

	fmt.Printf("\nFinal result: %v\n", result)
	fmt.Printf("Expected: [0 2]\n")
}
