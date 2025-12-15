// Package uax9 implements the Unicode Bidirectional Algorithm (UAX #9).
//
// This package provides bidirectional text reordering for proper display of text
// containing both left-to-right (LTR) and right-to-left (RTL) scripts, such as
// mixing Latin with Arabic or Hebrew text.
//
// Based on: https://www.unicode.org/reports/tr9/
//
// Usage:
//
//	import "github.com/SCKelemen/unicode/uax9"
//
//	text := "Hello مرحبا world"
//	result := uax9.Reorder(text, uax9.DirectionLTR)
//	// Returns properly reordered text for display
package uax9

import (
	"unicode"
)

// Direction represents the base text direction.
type Direction int

const (
	// DirectionLTR indicates left-to-right base direction (e.g., English)
	DirectionLTR Direction = iota
	// DirectionRTL indicates right-to-left base direction (e.g., Arabic, Hebrew)
	DirectionRTL
	// DirectionAuto automatically determines base direction from content
	DirectionAuto
)

// BidiClass represents the bidirectional character type as defined in UAX #9.
type BidiClass int

const (
	// Strong types
	ClassL   BidiClass = iota // Left-to-Right
	ClassR                    // Right-to-Left
	ClassAL                   // Right-to-Left Arabic

	// Weak types
	ClassEN                   // European Number
	ClassES                   // European Number Separator
	ClassET                   // European Number Terminator
	ClassAN                   // Arabic Number
	ClassCS                   // Common Number Separator
	ClassNSM                  // Nonspacing Mark
	ClassBN                   // Boundary Neutral

	// Neutral types
	ClassB                    // Paragraph Separator
	ClassS                    // Segment Separator
	ClassWS                   // Whitespace
	ClassON                   // Other Neutrals

	// Explicit formatting types
	ClassLRE                  // Left-to-Right Embedding
	ClassLRO                  // Left-to-Right Override
	ClassRLE                  // Right-to-Left Embedding
	ClassRLO                  // Right-to-Left Override
	ClassPDF                  // Pop Directional Format
	ClassLRI                  // Left-to-Right Isolate
	ClassRLI                  // Right-to-Left Isolate
	ClassFSI                  // First Strong Isolate
	ClassPDI                  // Pop Directional Isolate
)

// String returns the string representation of the BidiClass.
func (bc BidiClass) String() string {
	names := []string{
		"L", "R", "AL", "EN", "ES", "ET", "AN", "CS", "NSM", "BN",
		"B", "S", "WS", "ON", "LRE", "LRO", "RLE", "RLO", "PDF",
		"LRI", "RLI", "FSI", "PDI",
	}
	if int(bc) < len(names) {
		return names[bc]
	}
	return "Unknown"
}

// GetBidiClass returns the bidirectional character type for a given rune.
func GetBidiClass(r rune) BidiClass {
	// Explicit formatting characters
	switch r {
	case 0x202A:
		return ClassLRE
	case 0x202B:
		return ClassRLE
	case 0x202C:
		return ClassPDF
	case 0x202D:
		return ClassLRO
	case 0x202E:
		return ClassRLO
	case 0x2066:
		return ClassLRI
	case 0x2067:
		return ClassRLI
	case 0x2068:
		return ClassFSI
	case 0x2069:
		return ClassPDI
	}

	// Paragraph separators
	switch r {
	case 0x000A, 0x000D, 0x001C, 0x001D, 0x001E, 0x0085, 0x2029:
		return ClassB
	}

	// Segment separator
	if r == 0x001F {
		return ClassS
	}

	// Whitespace
	if r == ' ' || r == '\t' || r == 0x000B || r == 0x000C ||
		r == 0x1680 || (r >= 0x2000 && r <= 0x200A) ||
		r == 0x2028 || r == 0x205F || r == 0x3000 {
		return ClassWS
	}

	// Nonspacing marks
	if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) {
		return ClassNSM
	}

	// Format characters (BN)
	if unicode.Is(unicode.Cf, r) && r != 0x200C && r != 0x200D {
		return ClassBN
	}
	if r == 0x200B || r == 0xFEFF {
		return ClassBN
	}

	// Arabic letters
	if isArabicLetter(r) {
		return ClassAL
	}

	// Right-to-left characters
	if unicode.Is(unicode.Hebrew, r) {
		return ClassR
	}

	// European numbers
	if r >= '0' && r <= '9' {
		return ClassEN
	}

	// European number separators
	if r == '+' || r == '-' {
		return ClassES
	}

	// European number terminators
	if r == '#' || r == '$' || r == '%' || r == 0x00A2 || r == 0x00A3 ||
		r == 0x00A5 || r == 0x00B0 || r == 0x00B1 {
		return ClassET
	}

	// Arabic numbers (Extended Arabic-Indic)
	if (r >= 0x0660 && r <= 0x0669) || (r >= 0x066B && r <= 0x066C) {
		return ClassAN
	}

	// Common number separators
	if r == ',' || r == '.' || r == ':' || r == 0x00A0 {
		return ClassCS
	}

	// Left-to-right (most Latin, etc.)
	if unicode.Is(unicode.Latin, r) || unicode.Is(unicode.Greek, r) ||
		unicode.Is(unicode.Cyrillic, r) {
		return ClassL
	}

	// Default to Other Neutral
	return ClassON
}

// isArabicLetter checks if a rune is an Arabic letter.
func isArabicLetter(r rune) bool {
	return (r >= 0x0600 && r <= 0x06FF) || // Arabic
		(r >= 0x0750 && r <= 0x077F) || // Arabic Supplement
		(r >= 0x08A0 && r <= 0x08FF) || // Arabic Extended-A
		(r >= 0xFB50 && r <= 0xFDFF) || // Arabic Presentation Forms-A
		(r >= 0xFE70 && r <= 0xFEFF)    // Arabic Presentation Forms-B
}

// levelRun represents a run of characters with the same embedding level.
type levelRun struct {
	start int
	end   int
	level int
}

// Reorder reorders text according to the Unicode Bidirectional Algorithm.
// It takes the input text and a base direction, and returns the reordered text
// for proper visual display.
func Reorder(text string, baseDir Direction) string {
	if len(text) == 0 {
		return text
	}

	// Convert to runes for proper Unicode handling
	runes := []rune(text)
	n := len(runes)

	// Get bidi classes for each character
	classes := make([]BidiClass, n)
	for i, r := range runes {
		classes[i] = GetBidiClass(r)
	}

	// Keep original classes for L1 rule
	originalClasses := make([]BidiClass, n)
	copy(originalClasses, classes)

	// Determine paragraph level
	paraLevel := 0
	if baseDir == DirectionRTL {
		paraLevel = 1
	} else if baseDir == DirectionAuto {
		// P2: Find first strong character
		for _, class := range classes {
			if class == ClassL {
				paraLevel = 0
				break
			} else if class == ClassR || class == ClassAL {
				paraLevel = 1
				break
			}
		}
	}

	// Initialize levels
	levels := make([]int, n)
	for i := range levels {
		levels[i] = paraLevel
	}

	// Process explicit embeddings and isolates (X1-X8)
	processExplicitLevels(classes, levels, paraLevel)

	// Save explicit levels for sos/eos computation
	explicitLevels := make([]int, n)
	copy(explicitLevels, levels)

	// BD9: Determine matching isolate initiators and PDIs
	matchingPDI, matchingInitiator := determineMatchingIsolates(classes)

	// BD13: Determine isolating run sequences
	sequences := determineIsolatingRunSequences(classes, levels, matchingPDI, matchingInitiator)

	// Process each isolating run sequence
	for _, seqIndexes := range sequences {
		// Use explicit levels for sos/eos computation
		seq := newIsolatingRunSequence(seqIndexes, classes, originalClasses, explicitLevels, paraLevel)

		// Resolve weak types (W1-W7) within this sequence
		seq.resolveWeakTypes()

		// Resolve neutral types (N0-N2) within this sequence
		seq.resolveNeutralTypes()

		// Resolve implicit levels (I1-I2) within this sequence
		seq.resolveImplicitLevels()

		// Apply resolved types and levels back to original arrays
		for i, origIdx := range seq.indexes {
			classes[origIdx] = seq.types[i]
			levels[origIdx] = seq.levels[i]
		}
	}

	// Adjust empty isolate formatting character levels to match surrounding context
	adjustEmptyIsolateFormattingLevels(classes, originalClasses, levels, matchingPDI, paraLevel)

	// Adjust ALL isolate formatting character levels to match surrounding context
	adjustAllIsolateFormattingLevels(classes, levels, matchingPDI, matchingInitiator, paraLevel)

	// Apply L1: Reset levels for segment/paragraph separators and trailing whitespace
	applyL1(originalClasses, levels, paraLevel)

	// Reorder based on levels (L1-L4)
	return reorderByLevels(runes, levels, paraLevel)
}

// adjustEmptyIsolateFormattingLevels adjusts empty isolate formatting characters to match
// their surrounding resolved context. This applies after all W/N/I resolution is complete.
func adjustEmptyIsolateFormattingLevels(classes []BidiClass, originalClasses []BidiClass, levels []int, matchingPDI []int, paraLevel int) {
	n := len(classes)

	for i := 0; i < n; i++ {
		class := classes[i]
		// Check if this is an empty isolate initiator at paragraph level
		if (class == ClassLRI || class == ClassRLI || class == ClassFSI) && matchingPDI[i] != -1 && levels[i] == paraLevel {
			pdiIdx := matchingPDI[i]
			// Check if it's an empty isolate (no non-removed characters between initiator and PDI)
			isEmpty := true
			for j := i + 1; j < pdiIdx; j++ {
				if levels[j] >= 0 {
					isEmpty = false
					break
				}
			}

			// Only adjust empty isolates at paragraph level
			if isEmpty && levels[pdiIdx] == paraLevel {
				// Empty isolates take the level based on their surrounding resolved context

				// Find the left context (level, index, and resolved class)
				// Skip over isolate formatting characters to find actual content
				leftLevel := paraLevel
				leftIdx := -1
				for j := i - 1; j >= 0; j-- {
					if levels[j] >= 0 {
						c := classes[j]
						// Skip isolate formatting characters
						if c == ClassLRI || c == ClassRLI || c == ClassFSI || c == ClassPDI {
							continue
						}
						leftLevel = levels[j]
						leftIdx = j
						break
					}
				}

				// Find the right context (level, index, and resolved class)
				// Skip over isolate formatting characters to find actual content
				rightLevel := paraLevel
				rightIdx := -1
				for j := pdiIdx + 1; j < n; j++ {
					if levels[j] >= 0 {
						c := classes[j]
						// Skip isolate formatting characters
						if c == ClassLRI || c == ClassRLI || c == ClassFSI || c == ClassPDI {
							continue
						}
						rightLevel = levels[j]
						rightIdx = j
						break
					}
				}

				// Helper: check if both characters are strong and have the same directionality
				// Uses originalClasses to check types before resolution
				// Strong types: L (LTR), R and AL (RTL)
				bothStrongSameDir := func(leftIdx, rightIdx int) bool {
					if leftIdx < 0 || rightIdx < 0 {
						return false
					}
					leftClass := originalClasses[leftIdx]
					rightClass := originalClasses[rightIdx]

					// Check if both are strong types
					leftIsStrongLTR := leftClass == ClassL
					leftIsStrongRTL := leftClass == ClassR || leftClass == ClassAL
					rightIsStrongLTR := rightClass == ClassL
					rightIsStrongRTL := rightClass == ClassR || rightClass == ClassAL

					if !((leftIsStrongLTR || leftIsStrongRTL) && (rightIsStrongLTR || rightIsStrongRTL)) {
						return false
					}

					// Check if same directionality
					return (leftIsStrongLTR && rightIsStrongLTR) || (leftIsStrongRTL && rightIsStrongRTL)
				}

				// Helper: check if LEFT is strong with compatible directionality to RIGHT
				leftStrongCompat := func(leftIdx, rightIdx int) bool {
					if leftIdx < 0 || rightIdx < 0 {
						return false
					}
					leftClass := originalClasses[leftIdx]
					rightClass := originalClasses[rightIdx]

					// LEFT must be strong
					leftIsStrongLTR := leftClass == ClassL
					leftIsStrongRTL := leftClass == ClassR || leftClass == ClassAL
					if !leftIsStrongLTR && !leftIsStrongRTL {
						return false
					}

					// Check compatible directionality
					leftIsLTR := leftClass == ClassL || leftClass == ClassEN
					rightIsLTR := rightClass == ClassL || rightClass == ClassEN
					leftIsRTL := leftClass == ClassR || leftClass == ClassAL || leftClass == ClassAN
					rightIsRTL := rightClass == ClassR || rightClass == ClassAL || rightClass == ClassAN

					return (leftIsLTR && rightIsLTR) || (leftIsRTL && rightIsRTL)
				}

				// Rule: Empty isolates level assignment based on surrounding context
				if leftLevel != paraLevel && rightLevel != paraLevel {
					// Both sides at non-paragraph levels
					if leftLevel == rightLevel {
						// Both at same level
						if bothStrongSameDir(leftIdx, rightIdx) {
							// Both strong types with same directionality → match that level
							// Example: L(2) LRI PDI L(2) at para=1 → empty at 2
							// Example: R(1) LRI PDI R(1) at para=0 → empty at 1
							levels[i] = leftLevel
							levels[pdiIdx] = leftLevel
						} else if leftStrongCompat(leftIdx, rightIdx) {
							// LEFT is strong with compatible directionality → match surrounding level
							// Example: L(2) LRI PDI EN(2) at para=1 → empty at 2
							levels[i] = leftLevel
							levels[pdiIdx] = leftLevel
						} else if leftIdx >= 0 && rightIdx >= 0 && originalClasses[leftIdx] == originalClasses[rightIdx] {
							// Both same class (e.g., AN...AN, EN...EN) → use surrounding level - 1
							// Example: AN(2) LRI PDI AN(2) at para=0 → empty at 1 (2-1)
							// Example: AN(2) LRI PDI AN(2) at para=1 → empty at 1 (2-1)
							levels[i] = leftLevel - 1
							levels[pdiIdx] = leftLevel - 1
						}
						// else: incompatible directionality or LEFT is weak → stay at paragraph level
						// Example: EN(2) LRI PDI L(2) at para=1 → empty stays at 1
						// Example: L(2) LRI PDI AN(2) at para=1 → empty stays at 1
					} else {
						// Different levels → use minimum (closer to paragraph)
						// Example: R(1) LRI PDI EN(2) at para=0 → empty at 1
						minLevel := leftLevel
						if rightLevel < leftLevel {
							minLevel = rightLevel
						}
						levels[i] = minLevel
						levels[pdiIdx] = minLevel
					}
				}
				// else: one or both sides at paragraph level → stay at paragraph level
			}
		}
	}
}

// adjustAllIsolateFormattingLevels adjusts ALL isolate formatting characters to match
// their surrounding resolved context. This applies after empty isolate adjustment.
func adjustAllIsolateFormattingLevels(classes []BidiClass, levels []int, matchingPDI, matchingInitiator []int, paraLevel int) {
	n := len(classes)

	for i := 0; i < n; i++ {
		class := classes[i]
		// Check if this is a MATCHED isolate initiator currently at paragraph level
		if (class == ClassLRI || class == ClassRLI || class == ClassFSI) && matchingPDI[i] != -1 && levels[i] == paraLevel {
			pdiIdx := matchingPDI[i]

			// Check if it's an empty isolate (already handled by adjustEmptyIsolateFormattingLevels)
			isEmpty := true
			for j := i + 1; j < pdiIdx; j++ {
				if levels[j] >= 0 {
					isEmpty = false
					break
				}
			}

			// Skip empty isolates - they're handled by the previous function
			if isEmpty {
				continue
			}

			// Also check if the corresponding PDI is at paragraph level
			if levels[pdiIdx] == paraLevel {
				// Find the left context (skip ENTIRE isolate sequences, not just formatting characters)
				leftLevel := paraLevel
				for j := i - 1; j >= 0; j-- {
					if levels[j] >= 0 {
						c := classes[j]
						// If we hit a PDI, skip backwards to before its matching initiator
						if c == ClassPDI && matchingInitiator[j] != -1 {
							initiatorPos := matchingInitiator[j]
							j = initiatorPos // Will be decremented by loop, so continue from before initiator
							continue
						}
						// Skip other isolate formatting characters
						if c == ClassLRI || c == ClassRLI || c == ClassFSI {
							continue
						}
						leftLevel = levels[j]
						break
					}
				}

				// Find the right context (skip ENTIRE isolate sequences, not just formatting characters)
				rightLevel := paraLevel
				for j := pdiIdx + 1; j < n; j++ {
					if levels[j] >= 0 {
						c := classes[j]
						// If we hit an isolate initiator, skip forward to after its matching PDI
						if (c == ClassLRI || c == ClassRLI || c == ClassFSI) && matchingPDI[j] != -1 {
							pdiPos := matchingPDI[j]
							j = pdiPos // Will be incremented by loop, so continue from after PDI
							continue
						}
						// Skip PDI formatting characters
						if c == ClassPDI {
							continue
						}
						rightLevel = levels[j]
						break
					}
				}

				// Adjust level based on surrounding context
				// Only adjust when BOTH sides have non-paragraph levels
				if leftLevel != paraLevel && rightLevel != paraLevel {
					// Both sides at non-paragraph levels
					if leftLevel == rightLevel {
						// Both at same level → match that level for both initiator and PDI
						levels[i] = leftLevel
						levels[pdiIdx] = leftLevel
					} else {
						// Different levels → use minimum (closer to paragraph)
						minLevel := leftLevel
						if rightLevel < minLevel {
							minLevel = rightLevel
						}
						levels[i] = minLevel
						levels[pdiIdx] = minLevel
					}
				}
				// else: one or both sides at paragraph level → keep current level (don't adjust)
			}
		}
	}
}

// embeddingLevel represents a level on the directional embedding stack
type embeddingLevel struct {
	level    int
	override BidiClass // ClassL for LRO, ClassR for RLO, or -1 for no override
	isolate  bool
}

// processExplicitLevels handles explicit embedding and isolate formatting characters.
func processExplicitLevels(classes []BidiClass, levels []int, paraLevel int) {
	n := len(classes)
	const maxDepth = 125

	// Stack for tracking embeddings
	stack := []embeddingLevel{{level: paraLevel, override: -1, isolate: false}}
	overflowIsolateCount := 0
	overflowEmbeddingCount := 0
	validIsolateCount := 0
	// Stack to save overflowEmbeddingCount when overflow isolates are opened
	overflowEmbeddingStack := []int{}

	for i := 0; i < n; i++ {
		class := classes[i]

		// Handle explicit formatting characters
		switch class {
		case ClassRLE, ClassLRE, ClassRLO, ClassLRO:
			// X2-X5: Explicit embeddings and overrides
			currentLevel := stack[len(stack)-1].level
			newLevel := currentLevel
			override := BidiClass(-1)

			if class == ClassRLE || class == ClassRLO {
				// Right-to-left: next odd level
				newLevel = currentLevel + 1 + (currentLevel % 2)
				if class == ClassRLO {
					override = ClassR
				}
			} else {
				// Left-to-right: next even level
				newLevel = currentLevel + 2 - (currentLevel % 2)
				if class == ClassLRO {
					override = ClassL
				}
			}

			// Check for overflow: if already overflowing, continue overflowing
			// Otherwise check if new level would exceed limits
			if overflowEmbeddingCount > 0 || newLevel > maxDepth || len(stack) >= maxDepth {
				overflowEmbeddingCount++
			} else {
				stack = append(stack, embeddingLevel{level: newLevel, override: override, isolate: false})
				levels[i] = newLevel
			}
			// Mark for removal from reordering
			levels[i] = -1

		case ClassPDF:
			// X7: Pop directional formatting
			if overflowIsolateCount > 0 {
				// Do nothing
			} else if overflowEmbeddingCount > 0 {
				overflowEmbeddingCount--
			} else if len(stack) > 1 && !stack[len(stack)-1].isolate {
				stack = stack[:len(stack)-1]
			}
			levels[i] = -1

		case ClassLRI, ClassRLI, ClassFSI:
			// X5a-X5c: Isolate initiators
			currentLevel := stack[len(stack)-1].level
			newLevel := currentLevel
			isolateClass := class

			// FSI: determine direction from following strong character
			if class == ClassFSI {
				// Look ahead for first strong character to determine direction
				foundStrong := false
				isolateDepth := 0
				for j := i + 1; j < n; j++ {
					c := classes[j]
					// Track isolate depth
					if c == ClassLRI || c == ClassRLI || c == ClassFSI {
						isolateDepth++
					} else if c == ClassPDI {
						if isolateDepth > 0 {
							isolateDepth--
						} else {
							break // End of our isolate
						}
					}

					// Look for strong types at same isolate level
					// Note: EN and AN are weak types, not strong for FSI determination
					if isolateDepth == 0 {
						if c == ClassL {
							isolateClass = ClassLRI
							foundStrong = true
							break
						} else if c == ClassR || c == ClassAL {
							isolateClass = ClassRLI
							foundStrong = true
							break
						}
					}
				}
				// If no strong character found, use LTR
				if !foundStrong {
					isolateClass = ClassLRI
				}
			}

			if isolateClass == ClassRLI {
				newLevel = currentLevel + 1 + (currentLevel % 2)
			} else {
				newLevel = currentLevel + 2 - (currentLevel % 2)
			}

			if newLevel <= maxDepth && len(stack) < maxDepth {
				validIsolateCount++
				stack = append(stack, embeddingLevel{level: newLevel, override: -1, isolate: true})
				levels[i] = currentLevel // Isolate takes current level, not new level
			} else {
				// Save current overflowEmbeddingCount when opening overflow isolate
				overflowEmbeddingStack = append(overflowEmbeddingStack, overflowEmbeddingCount)
				overflowIsolateCount++
				levels[i] = currentLevel
			}

		case ClassPDI:
			// X6a: Pop directional isolate
			matched := false
			if overflowIsolateCount > 0 {
				overflowIsolateCount--
				// Restore overflowEmbeddingCount to the value before the overflow isolate
				// This discards overflow embeddings that happened inside the isolate
				if len(overflowEmbeddingStack) > 0 {
					overflowEmbeddingCount = overflowEmbeddingStack[len(overflowEmbeddingStack)-1]
					overflowEmbeddingStack = overflowEmbeddingStack[:len(overflowEmbeddingStack)-1]
				}
				matched = true
			} else if validIsolateCount > 0 {
				overflowEmbeddingCount = 0
				for len(stack) > 1 && !stack[len(stack)-1].isolate {
					stack = stack[:len(stack)-1]
				}
				if len(stack) > 1 {
					stack = stack[:len(stack)-1]
				}
				validIsolateCount--
				matched = true
			}
			levels[i] = stack[len(stack)-1].level

			// Unmatched PDI is treated as ON for neutral resolution
			if !matched {
				classes[i] = ClassON
			}

		case ClassBN:
			// BN characters are removed from reordering
			levels[i] = -1

		case ClassB:
			// X8: Paragraph separators always get paragraph level
			// They are not affected by embeddings or isolates
			levels[i] = paraLevel

		default:
			// X6: Set level for regular characters
			currentLevel := stack[len(stack)-1].level
			override := stack[len(stack)-1].override

			// NSM gets the current embedding level like any other character
			// NSM type resolution happens later in W1
			levels[i] = currentLevel

			// Apply override if present
			if override == ClassL {
				classes[i] = ClassL
			} else if override == ClassR {
				classes[i] = ClassR
			}
		}
	}
}

// resolveWeakTypes implements rules W1-W7 of the algorithm.
func resolveWeakTypes(classes []BidiClass, levels []int) {
	n := len(classes)

	// W1: NSM -> preceding class (or embedding level direction)
	// Look for preceding character at same level
	for i := 0; i < n; i++ {
		if classes[i] == ClassNSM {
			currentLevel := levels[i]
			if currentLevel < 0 {
				continue
			}

			// Find preceding character at same level that wasn't removed
			foundPreceding := false
			precedingLevel := -1

			for j := i - 1; j >= 0; j-- {
				if levels[j] < 0 {
					continue // Skip removed characters
				}
				if levels[j] == currentLevel {
					classes[i] = classes[j]
					foundPreceding = true
					break
				}
				// Track the preceding level for sos calculation
				if precedingLevel == -1 {
					precedingLevel = levels[j]
				}
			}

			if !foundPreceding {
				// Use sos (start-of-sequence) type
				// sos is determined by the higher of: current level or preceding character's level
				// If there's no preceding character, use paragraph level
				if precedingLevel == -1 {
					precedingLevel = currentLevel // Use current level if no preceding char
				}

				sosLevel := currentLevel
				if precedingLevel > sosLevel {
					sosLevel = precedingLevel
				}

				// sos is R if higher level is odd, L if even
				if sosLevel%2 == 0 {
					classes[i] = ClassL
				} else {
					classes[i] = ClassR
				}
			}
		}
	}

	// W2: EN after AL -> AN
	// Only look at same embedding level
	for i := 0; i < n; i++ {
		if classes[i] == ClassEN {
			currentLevel := levels[i]
			if currentLevel < 0 {
				continue
			}

			// Look back for AL at same level
			for j := i - 1; j >= 0; j-- {
				if levels[j] < 0 {
					continue // Skip removed characters
				}
				if levels[j] != currentLevel {
					continue // Different level
				}
				if classes[j] == ClassAL {
					classes[i] = ClassAN
					break
				} else if classes[j] == ClassL || classes[j] == ClassR {
					break
				}
			}
		}
	}

	// W3: AL -> R
	for i := 0; i < n; i++ {
		if classes[i] == ClassAL {
			classes[i] = ClassR
		}
	}

	// W4: Single separator between numbers -> number
	// Skip removed characters and only look at same embedding level
	for i := 0; i < n; i++ {
		if classes[i] == ClassES || classes[i] == ClassCS {
			currentLevel := levels[i]
			if currentLevel < 0 {
				continue // Skip removed separators
			}

			// Look backward for number at same level (skip removed characters)
			var prevClass BidiClass = -1
			for j := i - 1; j >= 0; j-- {
				if levels[j] < 0 {
					continue // Skip removed
				}
				if levels[j] != currentLevel {
					break // Different level, stop searching
				}
				prevClass = classes[j]
				break
			}

			// Look forward for number at same level (skip removed characters)
			var nextClass BidiClass = -1
			for j := i + 1; j < n; j++ {
				if levels[j] < 0 {
					continue // Skip removed
				}
				if levels[j] != currentLevel {
					break // Different level, stop searching
				}
				nextClass = classes[j]
				break
			}

			// ES between ENs at same level -> EN
			if classes[i] == ClassES && prevClass == ClassEN && nextClass == ClassEN {
				classes[i] = ClassEN
			}
			// CS between same number types at same level -> that type
			if classes[i] == ClassCS {
				if prevClass == ClassEN && nextClass == ClassEN {
					classes[i] = ClassEN
				} else if prevClass == ClassAN && nextClass == ClassAN {
					classes[i] = ClassAN
				}
			}
		}
	}

	// W5: Sequence of ET adjacent to EN -> EN
	// A sequence of ET is adjacent to EN if there's an EN before or after the sequence
	// Note: Skip removed characters AND only look at same embedding level
	// ET sequences can have removed characters in between (e.g., "ET PDF ET")
	for i := 0; i < n; i++ {
		if classes[i] == ClassET {
			currentLevel := levels[i]
			if currentLevel < 0 {
				continue // Skip removed ET
			}

			hasEN := false

			// Find the start of the ET sequence at this level (skipping removed)
			start := i
			for j := i - 1; j >= 0; j-- {
				if levels[j] < 0 {
					continue // Skip removed
				}
				if levels[j] != currentLevel {
					break // Different level
				}
				if classes[j] == ClassET {
					start = j
				} else {
					break
				}
			}

			// Find the end of the ET sequence at this level (skipping removed)
			end := i
			for j := i + 1; j < n; j++ {
				if levels[j] < 0 {
					continue // Skip removed
				}
				if levels[j] != currentLevel {
					break // Different level
				}
				if classes[j] == ClassET {
					end = j
				} else {
					break
				}
			}

			// Check if there's an EN before the sequence at same level (skip removed)
			for j := start - 1; j >= 0; j-- {
				if levels[j] < 0 {
					continue // Skip removed characters
				}
				if levels[j] != currentLevel {
					break // Different level, stop searching
				}
				if classes[j] == ClassEN {
					hasEN = true
				}
				break
			}

			// Check if there's an EN after the sequence at same level (skip removed)
			if !hasEN {
				for j := end + 1; j < n; j++ {
					if levels[j] < 0 {
						continue // Skip removed characters
					}
					if levels[j] != currentLevel {
						break // Different level, stop searching
					}
					if classes[j] == ClassEN {
						hasEN = true
					}
					break
				}
			}

			// If adjacent to EN at same level, mark all ET in sequence as EN
			if hasEN {
				for j := start; j <= end; j++ {
					if classes[j] == ClassET {
						classes[j] = ClassEN
					}
				}
				// Skip to end of sequence
				i = end
			}
		}
	}

	// W6: Separators and terminators -> ON
	for i := 0; i < n; i++ {
		if classes[i] == ClassES || classes[i] == ClassET || classes[i] == ClassCS {
			classes[i] = ClassON
		}
	}

	// W7: EN after L -> L (or sos L)
	// Search for strong types at the same level
	for i := 0; i < n; i++ {
		if classes[i] == ClassEN {
			currentLevel := levels[i]
			if currentLevel < 0 {
				continue
			}

			// Look back for strong type (L or R) at same level
			foundStrong := false
			precedingLevel := -1

			for j := i - 1; j >= 0; j-- {
				if levels[j] < 0 {
					continue // Skip removed characters
				}
				// Only consider characters at same level
				if levels[j] == currentLevel {
					if classes[j] == ClassL {
						classes[i] = ClassL
						foundStrong = true
						break
					} else if classes[j] == ClassR {
						foundStrong = true
						break
					}
				}
				// Track the preceding level for sos calculation
				if precedingLevel == -1 {
					precedingLevel = levels[j]
				}
			}

			// If no strong type found at same level, use sos (start of sequence)
			// sos is determined by the higher of: current level or preceding character's level
			if !foundStrong {
				if precedingLevel == -1 {
					precedingLevel = currentLevel
				}

				sosLevel := currentLevel
				if precedingLevel > sosLevel {
					sosLevel = precedingLevel
				}

				// sos is R if higher level is odd, L if even
				// W7: EN after sos L -> L (else stay EN)
				if sosLevel%2 == 0 {
					classes[i] = ClassL
				}
				// else stay EN for odd levels
			}
		}
	}
}

// detectEmptyIsolates marks empty isolates (isolate initiator immediately followed by PDI)
// Empty isolate formatting characters should be treated as neutrals for proper level resolution
// determineLevelRuns finds all maximal sequences of characters at the same embedding level.
// Removed characters (level < 0) are skipped.
// Returns a slice of level runs, where each run is a slice of character indexes.
func determineLevelRuns(levels []int) [][]int {
	n := len(levels)
	var runs [][]int
	var currentRun []int

	for i := 0; i < n; i++ {
		if levels[i] < 0 {
			// Skip removed characters
			continue
		}

		if len(currentRun) == 0 {
			// Start a new run
			currentRun = []int{i}
		} else if levels[i] == levels[currentRun[0]] {
			// Continue current run
			currentRun = append(currentRun, i)
		} else {
			// Level changed, save current run and start new one
			runs = append(runs, currentRun)
			currentRun = []int{i}
		}
	}

	// Don't forget the last run
	if len(currentRun) > 0 {
		runs = append(runs, currentRun)
	}

	return runs
}

// determineMatchingIsolates implements BD9: determine matching PDI for each isolate initiator.
// Returns two slices:
// - matchingPDI[i] = index of matching PDI for isolate initiator at i, or -1 if not an initiator
// - matchingInitiator[i] = index of matching initiator for PDI at i, or -1 if not a PDI
func determineMatchingIsolates(classes []BidiClass) ([]int, []int) {
	n := len(classes)
	matchingPDI := make([]int, n)
	matchingInitiator := make([]int, n)

	// Initialize all to -1
	for i := 0; i < n; i++ {
		matchingPDI[i] = -1
		matchingInitiator[i] = -1
	}

	// For each isolate initiator, find its matching PDI
	for i := 0; i < n; i++ {
		class := classes[i]
		if class != ClassLRI && class != ClassRLI && class != ClassFSI {
			continue
		}

		// Found an isolate initiator, search for matching PDI
		depth := 1
		for j := i + 1; j < n; j++ {
			c := classes[j]
			if c == ClassLRI || c == ClassRLI || c == ClassFSI {
				depth++
			} else if c == ClassPDI {
				depth--
				if depth == 0 {
					// Found matching PDI
					matchingPDI[i] = j
					matchingInitiator[j] = i
					break
				}
			}
		}
		// If no matching PDI found, matchingPDI[i] remains -1
	}

	return matchingPDI, matchingInitiator
}

// isolatingRunSequence represents a maximal sequence of level runs connected through isolates.
// It contains all information needed to resolve types within the sequence.
type isolatingRunSequence struct {
	indexes       []int       // Original character indexes in this sequence
	types         []BidiClass // Working copy of types for resolution
	levels        []int       // Resolved levels for each character
	level         int         // Embedding level of this sequence
	sos           BidiClass   // Start-of-sequence type (L or R)
	eos           BidiClass   // End-of-sequence type (L or R)
	originalTypes []BidiClass // Original types before resolution
}

// typeForLevel returns L or R based on level parity
func typeForLevel(level int) BidiClass {
	if level%2 == 0 {
		return ClassL
	}
	return ClassR
}

// newIsolatingRunSequence creates an isolating run sequence from character indexes.
// X10: Determine sos and eos for the sequence.
func newIsolatingRunSequence(indexes []int, classes, originalClasses []BidiClass, levels []int, paraLevel int) *isolatingRunSequence {
	seq := &isolatingRunSequence{
		indexes:       indexes,
		types:         make([]BidiClass, len(indexes)),
		levels:        make([]int, len(indexes)),
		originalTypes: make([]BidiClass, len(indexes)),
	}

	// Get the embedding level from the first character
	seq.level = levels[indexes[0]]

	// Copy types and initialize levels for this sequence
	for i, idx := range indexes {
		seq.types[i] = classes[idx]
		seq.originalTypes[i] = originalClasses[idx]
		// All characters in the sequence start at the sequence's base level
		seq.levels[i] = seq.level
	}

	// Determine sos (start-of-sequence type)
	// Look at the level before the first character in the sequence
	firstIdx := indexes[0]
	prevLevel := paraLevel
	for i := firstIdx - 1; i >= 0; i-- {
		if levels[i] >= 0 { // Skip removed characters
			prevLevel = levels[i]
			break
		}
	}
	seq.sos = typeForLevel(max(prevLevel, seq.level))

	// Determine eos (end-of-sequence type)
	lastIdx := indexes[len(indexes)-1]
	lastType := classes[lastIdx]
	var succLevel int

	// If sequence ends with an isolate initiator, eos is based on paragraph level
	if lastType == ClassLRI || lastType == ClassRLI || lastType == ClassFSI {
		succLevel = paraLevel
	} else {
		// Look at the level after the last character
		succLevel = paraLevel
		for i := lastIdx + 1; i < len(classes); i++ {
			if levels[i] >= 0 {
				succLevel = levels[i]
				break
			}
		}
	}
	seq.eos = typeForLevel(max(succLevel, seq.level))

	return seq
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// resolveWeakTypes resolves weak types (W1-W7) within this isolating run sequence.
func (seq *isolatingRunSequence) resolveWeakTypes() {
	n := len(seq.types)

	// W1: NSM takes the type of the preceding character
	precedingType := seq.sos
	for i := 0; i < n; i++ {
		if seq.types[i] == ClassNSM {
			seq.types[i] = precedingType
		} else {
			// Isolate formatting characters are treated as ON for preceding type
			if seq.types[i] == ClassLRI || seq.types[i] == ClassRLI ||
				seq.types[i] == ClassFSI || seq.types[i] == ClassPDI {
				precedingType = ClassON
			} else {
				precedingType = seq.types[i]
			}
		}
	}

	// W2: EN following AL becomes AN
	for i := 0; i < n; i++ {
		if seq.types[i] == ClassEN {
			// Look back for AL or strong L/R
			foundStrong := false
			for j := i - 1; j >= 0; j-- {
				t := seq.types[j]
				if t == ClassL || t == ClassR || t == ClassAL {
					if t == ClassAL {
						seq.types[i] = ClassAN
					}
					foundStrong = true
					break
				}
			}
			// If no strong type found, use sos (which is L or R, not AL, so no change)
			_ = foundStrong // sos cannot be AL, so this doesn't affect the result
		}
	}

	// W3: AL becomes R
	for i := 0; i < n; i++ {
		if seq.types[i] == ClassAL {
			seq.types[i] = ClassR
		}
	}

	// W4: Single separator between numbers takes number type
	for i := 1; i < n-1; i++ {
		if seq.types[i] == ClassES || seq.types[i] == ClassCS {
			prevType := seq.types[i-1]
			nextType := seq.types[i+1]

			if prevType == ClassEN && nextType == ClassEN && seq.types[i] == ClassES {
				seq.types[i] = ClassEN
			} else if seq.types[i] == ClassCS {
				if prevType == ClassEN && nextType == ClassEN {
					seq.types[i] = ClassEN
				} else if prevType == ClassAN && nextType == ClassAN {
					seq.types[i] = ClassAN
				}
			}
		}
	}

	// W5: Sequence of ET adjacent to EN becomes EN
	for i := 0; i < n; i++ {
		if seq.types[i] == ClassET {
			// Check if adjacent to EN (before or after)
			foundEN := false

			// Check before
			for j := i - 1; j >= 0; j-- {
				if seq.types[j] == ClassEN {
					foundEN = true
					break
				} else if seq.types[j] != ClassET {
					break
				}
			}

			// Check after
			if !foundEN {
				for j := i + 1; j < n; j++ {
					if seq.types[j] == ClassEN {
						foundEN = true
						break
					} else if seq.types[j] != ClassET {
						break
					}
				}
			}

			if foundEN {
				seq.types[i] = ClassEN
			}
		}
	}

	// W6: Separators and terminators become ON
	for i := 0; i < n; i++ {
		if seq.types[i] == ClassES || seq.types[i] == ClassET || seq.types[i] == ClassCS {
			seq.types[i] = ClassON
		}
	}

	// W7: EN following L becomes L
	for i := 0; i < n; i++ {
		if seq.types[i] == ClassEN {
			// Look back for L or strong R
			foundStrong := false
			for j := i - 1; j >= 0; j-- {
				t := seq.types[j]
				if t == ClassL {
					seq.types[i] = ClassL
					foundStrong = true
					break
				} else if t == ClassR {
					foundStrong = true
					break
				}
			}
			// If no strong type found, use sos
			if !foundStrong && seq.sos == ClassL {
				seq.types[i] = ClassL
			}
		}
	}
}

// resolveNeutralTypes resolves neutral types (N0-N2) within this isolating run sequence.
func (seq *isolatingRunSequence) resolveNeutralTypes() {
	n := len(seq.types)

	// N0: Paired bracket resolution would go here (not fully implemented)

	// N1 and N2: Neutrals take direction from surrounding strong types
	for i := 0; i < n; i++ {
		// Check if neutral or empty isolate
		isNeutral := seq.types[i] == ClassWS || seq.types[i] == ClassON ||
			seq.types[i] == ClassB || seq.types[i] == ClassS

		if !isNeutral {
			continue
		}

		// Find preceding strong type
		var prevType BidiClass = seq.sos
		for j := i - 1; j >= 0; j-- {
			if seq.types[j] == ClassL {
				prevType = ClassL
				break
			} else if seq.types[j] == ClassR || seq.types[j] == ClassEN || seq.types[j] == ClassAN {
				prevType = ClassR
				break
			}
		}

		// Find following strong type
		var nextType BidiClass = seq.eos
		for j := i + 1; j < n; j++ {
			if seq.types[j] == ClassL {
				nextType = ClassL
				break
			} else if seq.types[j] == ClassR || seq.types[j] == ClassEN || seq.types[j] == ClassAN {
				nextType = ClassR
				break
			}
		}

		// N1: If surrounded by same strong type, take that type
		if prevType == nextType {
			seq.types[i] = prevType
		} else {
			// N2: Take embedding direction
			seq.types[i] = typeForLevel(seq.level)
		}
	}
}

// resolveImplicitLevels resolves implicit levels (I1-I2) within this isolating run sequence.
func (seq *isolatingRunSequence) resolveImplicitLevels() {
	for i := 0; i < len(seq.types); i++ {
		level := seq.levels[i]
		class := seq.types[i]

		// I1: For even levels
		if level%2 == 0 {
			if class == ClassR {
				seq.levels[i] = level + 1
			} else if class == ClassAN || class == ClassEN {
				seq.levels[i] = level + 2
			}
		} else {
			// I2: For odd levels
			if class == ClassL || class == ClassAN || class == ClassEN {
				seq.levels[i] = level + 1
			}
		}
	}
}

// determineIsolatingRunSequences implements BD13: determine isolating run sequences.
// An isolating run sequence is a maximal sequence of level runs where each run after the
// first is the continuation of an isolate sequence started in a previous run.
// Returns a slice of isolating run sequences, where each sequence is a slice of character indexes.
func determineIsolatingRunSequences(classes []BidiClass, levels []int, matchingPDI, matchingInitiator []int) [][]int {
	levelRuns := determineLevelRuns(levels)
	n := len(classes)

	// Map each character to its run number
	runForChar := make([]int, n)
	for runNum, run := range levelRuns {
		for _, charIdx := range run {
			runForChar[charIdx] = runNum
		}
	}

	// Track which runs have been processed
	processed := make([]bool, len(levelRuns))
	var sequences [][]int

	// For each level run, if it hasn't been processed and doesn't start with a PDI
	// that has a matching initiator, build an isolating run sequence
	for runNum, run := range levelRuns {
		if processed[runNum] {
			continue
		}

		firstChar := run[0]
		// Skip runs that start with a PDI that has a matching initiator
		if classes[firstChar] == ClassPDI && matchingInitiator[firstChar] != -1 {
			continue
		}

		// Build an isolating run sequence starting from this run
		var sequence []int
		currentRun := runNum

		for {
			processed[currentRun] = true
			run := levelRuns[currentRun]

			// Add all characters from this run to the sequence
			sequence = append(sequence, run...)

			// Check if this run ends with an isolate initiator that has a matching PDI
			lastChar := run[len(run)-1]
			if (classes[lastChar] == ClassLRI || classes[lastChar] == ClassRLI || classes[lastChar] == ClassFSI) &&
				matchingPDI[lastChar] != -1 {
				// Continue with the run containing the matching PDI
				pdiIdx := matchingPDI[lastChar]
				currentRun = runForChar[pdiIdx]
			} else {
				// No continuation, end of sequence
				break
			}
		}

		sequences = append(sequences, sequence)
	}

	return sequences
}

func detectEmptyIsolates(classes []BidiClass, levels []int) []bool {
	n := len(classes)
	isEmptyIsolate := make([]bool, n)

	for i := 0; i < n; i++ {
		if classes[i] == ClassLRI || classes[i] == ClassRLI || classes[i] == ClassFSI {
			// Look ahead for matching PDI, skipping removed characters (level < 0)
			foundNonRemoved := false
			pdiIndex := -1

			for j := i + 1; j < n; j++ {
				if levels[j] < 0 {
					continue // Skip removed characters (BN, PDF, LRE, RLE, etc.)
				}
				if classes[j] == ClassPDI {
					pdiIndex = j
				} else {
					foundNonRemoved = true
				}
				break // Stop at first non-removed character
			}

			// If no non-removed characters between initiator and PDI, mark as empty
			if !foundNonRemoved && pdiIndex >= 0 {
				isEmptyIsolate[i] = true
				isEmptyIsolate[pdiIndex] = true
			}
		}
	}

	return isEmptyIsolate
}

// resolveNeutralTypes implements rules N0-N2 of the algorithm.
func resolveNeutralTypes(classes []BidiClass, levels []int, paraLevel int, isEmptyIsolate []bool) {
	n := len(classes)

	// Helper to check if a class is strong (L, R, or AN/EN which behave as R)
	isStrongL := func(c BidiClass) bool {
		return c == ClassL
	}
	isStrongR := func(c BidiClass) bool {
		return c == ClassR || c == ClassAN || c == ClassEN
	}

	// N1 and N2: Neutrals take direction from surrounding strong types
	// Note: Isolate formatting characters (RLI, LRI, FSI, PDI) keep their types
	// Exception: Empty isolate formatting characters are treated as neutrals
	// Search within same level run (same embedding level), use sos/eos at boundaries
	for i := 0; i < n; i++ {
		isNeutral := classes[i] == ClassWS || classes[i] == ClassON ||
			classes[i] == ClassB || classes[i] == ClassS

		// Empty isolate formatting characters should be treated as neutrals
		if !isNeutral && isEmptyIsolate[i] {
			if classes[i] == ClassLRI || classes[i] == ClassRLI ||
				classes[i] == ClassFSI || classes[i] == ClassPDI {
				isNeutral = true
			}
		}

		if isNeutral {

			currentLevel := levels[i]
			if currentLevel < 0 {
				continue
			}

			// Find preceding strong type at same level
			prevIsL := false
			prevIsR := false
			precedingLevel := -1

			for j := i - 1; j >= 0; j-- {
				if levels[j] < 0 {
					continue // Skip removed
				}
				if levels[j] == currentLevel {
					if isStrongL(classes[j]) {
						prevIsL = true
						break
					} else if isStrongR(classes[j]) {
						prevIsR = true
						break
					}
				}
				// Track preceding level for sos
				if precedingLevel == -1 {
					precedingLevel = levels[j]
				}
			}

			// If no strong type at same level, use sos
			if !prevIsL && !prevIsR {
				if precedingLevel == -1 {
					precedingLevel = currentLevel
				}
				sosLevel := currentLevel
				if precedingLevel > sosLevel {
					sosLevel = precedingLevel
				}
				// sos is R if higher level is odd, L if even
				if sosLevel%2 == 0 {
					prevIsL = true
				} else {
					prevIsR = true
				}
			}

			// Find following strong type at same level
			nextIsL := false
			nextIsR := false
			followingLevel := -1

			for j := i + 1; j < n; j++ {
				if levels[j] < 0 {
					continue // Skip removed
				}
				if levels[j] == currentLevel {
					if isStrongL(classes[j]) {
						nextIsL = true
						break
					} else if isStrongR(classes[j]) {
						nextIsR = true
						break
					}
				}
				// Track following level for eos
				if followingLevel == -1 {
					followingLevel = levels[j]
				}
			}

			// If no strong type at same level, use eos
			if !nextIsL && !nextIsR {
				if followingLevel == -1 {
					followingLevel = currentLevel
				}
				eosLevel := currentLevel
				if followingLevel > eosLevel {
					eosLevel = followingLevel
				}
				// eos is R if higher level is odd, L if even
				if eosLevel%2 == 0 {
					nextIsL = true
				} else {
					nextIsR = true
				}
			}

			// N1: If between same strong types, take that type
			if (prevIsL && nextIsL) || (prevIsR && nextIsR) {
				if prevIsL {
					classes[i] = ClassL
				} else {
					classes[i] = ClassR
				}
			} else {
				// N2: Take embedding level direction
				if levels[i]%2 == 0 {
					classes[i] = ClassL
				} else {
					classes[i] = ClassR
				}
			}
		}
	}
}

// applyL1 resets levels for segment/paragraph separators, trailing whitespace, and isolates.
func applyL1(classes []BidiClass, levels []int, paraLevel int) {
	n := len(classes)

	// L1: Reset segment separators, paragraph separators, and preceding/trailing whitespace and isolates
	// Per UAX#9 L1: reset separators and any WS/isolates PRECEDING them (not following)
	for i := 0; i < n; i++ {
		if levels[i] < 0 {
			continue
		}

		// Segment and paragraph separators get paragraph level
		if classes[i] == ClassS || classes[i] == ClassB {
			levels[i] = paraLevel
			// Also reset any PRECEDING whitespace, separators, and isolates
			for j := i - 1; j >= 0; j-- {
				if levels[j] < 0 {
					// Skip removed characters
					continue
				}
				if classes[j] == ClassWS || classes[j] == ClassS ||
					classes[j] == ClassB || classes[j] == ClassLRI || classes[j] == ClassRLI ||
					classes[j] == ClassFSI || classes[j] == ClassPDI {
					levels[j] = paraLevel
				} else {
					// Stop when we hit a non-whitespace/non-isolate
					break
				}
			}
		}
	}

	// Reset trailing whitespace and isolates
	for i := n - 1; i >= 0; i-- {
		if levels[i] < 0 {
			continue
		}
		if classes[i] == ClassWS || classes[i] == ClassS || classes[i] == ClassB ||
			classes[i] == ClassLRI || classes[i] == ClassRLI || classes[i] == ClassFSI ||
			classes[i] == ClassPDI {
			levels[i] = paraLevel
		} else {
			break
		}
	}
}

// resolveImplicitLevels implements rules I1-I2 of the algorithm.
func resolveImplicitLevels(classes []BidiClass, levels []int) {
	for i, class := range classes {
		if levels[i] < 0 {
			continue // Skip removed characters
		}

		level := levels[i]

		// I1: For even levels
		if level%2 == 0 {
			if class == ClassR {
				levels[i] = level + 1
			} else if class == ClassAN || class == ClassEN {
				levels[i] = level + 2
			}
		} else {
			// I2: For odd levels
			if class == ClassL || class == ClassAN || class == ClassEN {
				levels[i] = level + 1
			}
		}
	}
}

// reorderByLevels reorders the text based on resolved levels.
func reorderByLevels(runes []rune, levels []int, paraLevel int) string {
	n := len(runes)

	// Find the maximum level
	maxLevel := paraLevel
	for _, level := range levels {
		if level > maxLevel {
			maxLevel = level
		}
	}

	// Create a copy of indices
	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}

	// Reverse runs from highest level to lowest odd level (L2)
	// Per UAX#9 L2: reverse down to the lowest odd level
	lowestOddLevel := 1
	if paraLevel > lowestOddLevel {
		lowestOddLevel = paraLevel
	}

	for level := maxLevel; level >= lowestOddLevel; level-- {
		i := 0
		for i < n {
			// Skip removed characters when not in a run
			if levels[i] == -1 {
				i++
				continue
			}

			// Skip characters below this level
			if levels[i] < level {
				i++
				continue
			}

			// Found start of a run at this level
			start := i
			i++

			// Extend run: include removed characters AND characters at this level
			for i < n {
				if levels[i] == -1 {
					// Removed character: tentatively include it
					i++
				} else if levels[i] >= level {
					// Character at this level or higher: include it
					i++
				} else {
					// Character below this level: stop
					break
				}
			}
			end := i - 1

			// Reverse this run
			for start < end {
				indices[start], indices[end] = indices[end], indices[start]
				start++
				end--
			}
		}
	}

	// Build result
	result := make([]rune, 0, n)
	for _, idx := range indices {
		if levels[idx] >= 0 { // Skip removed characters
			result = append(result, runes[idx])
		}
	}

	return string(result)
}

// GetParagraphDirection automatically detects the paragraph direction.
func GetParagraphDirection(text string) Direction {
	for _, r := range text {
		class := GetBidiClass(r)
		if class == ClassL {
			return DirectionLTR
		} else if class == ClassR || class == ClassAL {
			return DirectionRTL
		}
	}
	return DirectionLTR // Default to LTR if no strong characters found
}
