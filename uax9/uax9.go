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

	// Resolve weak types (W1-W7)
	resolveWeakTypes(classes, levels)

	// Resolve neutral types (N0-N2)
	resolveNeutralTypes(classes, levels, paraLevel)

	// Resolve implicit levels (I1-I2)
	resolveImplicitLevels(classes, levels)

	// Apply L1: Reset levels for segment/paragraph separators and trailing whitespace
	applyL1(originalClasses, levels, paraLevel)

	// Reorder based on levels (L1-L4)
	return reorderByLevels(runes, levels, paraLevel)
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

			if newLevel <= maxDepth && len(stack) < maxDepth {
				stack = append(stack, embeddingLevel{level: newLevel, override: override, isolate: false})
				levels[i] = newLevel
			} else {
				overflowEmbeddingCount++
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

			// FSI: determine direction from content
			if class == ClassFSI {
				// Simplified: treat as LRI for now
				class = ClassLRI
			}

			if class == ClassRLI {
				newLevel = currentLevel + 1 + (currentLevel % 2)
			} else {
				newLevel = currentLevel + 2 - (currentLevel % 2)
			}

			if newLevel <= maxDepth && len(stack) < maxDepth {
				validIsolateCount++
				stack = append(stack, embeddingLevel{level: newLevel, override: -1, isolate: true})
				levels[i] = currentLevel // Isolate takes current level, not new level
			} else {
				overflowIsolateCount++
				levels[i] = currentLevel
			}

		case ClassPDI:
			// X6a: Pop directional isolate
			if overflowIsolateCount > 0 {
				overflowIsolateCount--
			} else if validIsolateCount > 0 {
				overflowEmbeddingCount = 0
				for len(stack) > 1 && !stack[len(stack)-1].isolate {
					stack = stack[:len(stack)-1]
				}
				if len(stack) > 1 {
					stack = stack[:len(stack)-1]
				}
				validIsolateCount--
			}
			levels[i] = stack[len(stack)-1].level

		case ClassBN:
			// BN characters are removed from reordering
			levels[i] = -1

		default:
			// X6: Set level for regular characters
			currentLevel := stack[len(stack)-1].level
			override := stack[len(stack)-1].override

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
	for i := 0; i < n; i++ {
		if classes[i] == ClassNSM {
			if i > 0 && levels[i-1] >= 0 {
				classes[i] = classes[i-1]
			} else {
				// Use paragraph level
				if levels[i]%2 == 0 {
					classes[i] = ClassL
				} else {
					classes[i] = ClassR
				}
			}
		}
	}

	// W2: EN after AL -> AN
	for i := 0; i < n; i++ {
		if classes[i] == ClassEN {
			// Look back for AL
			for j := i - 1; j >= 0; j-- {
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
	for i := 1; i < n-1; i++ {
		if classes[i] == ClassES || classes[i] == ClassCS {
			prevNum := classes[i-1] == ClassEN || classes[i-1] == ClassAN
			nextNum := classes[i+1] == ClassEN || classes[i+1] == ClassAN
			if prevNum && nextNum {
				if classes[i-1] == ClassEN && classes[i+1] == ClassEN {
					classes[i] = ClassEN
				} else if classes[i-1] == ClassAN && classes[i+1] == ClassAN {
					classes[i] = ClassAN
				}
			}
		}
	}

	// W5: Sequence of ET adjacent to EN -> EN
	for i := 0; i < n; i++ {
		if classes[i] == ClassET {
			// Check if adjacent to EN
			hasEN := false
			// Check before
			for j := i - 1; j >= 0 && classes[j] == ClassET; j-- {
				if j > 0 && classes[j-1] == ClassEN {
					hasEN = true
					break
				}
			}
			// Check after
			if !hasEN {
				for j := i + 1; j < n && classes[j] == ClassET; j++ {
					if j < n-1 && classes[j+1] == ClassEN {
						hasEN = true
						break
					}
				}
			}
			if hasEN {
				classes[i] = ClassEN
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
	for i := 0; i < n; i++ {
		if classes[i] == ClassEN {
			// Look back for strong type (L or R)
			foundStrong := false
			for j := i - 1; j >= 0; j-- {
				if levels[j] < 0 {
					continue // Skip removed characters
				}
				if classes[j] == ClassL {
					classes[i] = ClassL
					foundStrong = true
					break
				} else if classes[j] == ClassR {
					foundStrong = true
					break
				}
			}
			// If no strong type found, use sos (start of sequence)
			// sos is L for even levels, R for odd levels
			if !foundStrong {
				if levels[i]%2 == 0 {
					classes[i] = ClassL
				}
				// else stay EN for odd levels
			}
		}
	}
}

// resolveNeutralTypes implements rules N0-N2 of the algorithm.
func resolveNeutralTypes(classes []BidiClass, levels []int, paraLevel int) {
	n := len(classes)

	// N1 and N2: Neutrals take direction from surrounding strong types
	for i := 0; i < n; i++ {
		if classes[i] == ClassWS || classes[i] == ClassON ||
			classes[i] == ClassB || classes[i] == ClassS {

			// Find preceding strong type
			prevStrong := BidiClass(-1)
			for j := i - 1; j >= 0; j-- {
				if classes[j] == ClassL || classes[j] == ClassR {
					prevStrong = classes[j]
					break
				}
			}

			// Find following strong type
			nextStrong := BidiClass(-1)
			for j := i + 1; j < n; j++ {
				if classes[j] == ClassL || classes[j] == ClassR {
					nextStrong = classes[j]
					break
				}
			}

			// N1: If between same strong types, take that type
			if prevStrong == nextStrong && prevStrong != -1 {
				classes[i] = prevStrong
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

	// L1: Reset segment separators, paragraph separators, trailing whitespace, and isolates
	for i := 0; i < n; i++ {
		if levels[i] < 0 {
			continue
		}

		// Segment and paragraph separators get paragraph level
		if classes[i] == ClassS || classes[i] == ClassB {
			levels[i] = paraLevel
			// Also reset any following whitespace, separators, and isolates
			for j := i + 1; j < n; j++ {
				if levels[j] >= 0 && (classes[j] == ClassWS || classes[j] == ClassS ||
					classes[j] == ClassB || classes[j] == ClassLRI || classes[j] == ClassRLI ||
					classes[j] == ClassFSI || classes[j] == ClassPDI) {
					levels[j] = paraLevel
				} else if levels[j] >= 0 {
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

	// Reverse runs from highest level to lowest
	for level := maxLevel; level > paraLevel; level-- {
		i := 0
		for i < n {
			if levels[i] >= level {
				// Find end of run
				start := i
				for i < n && levels[i] >= level {
					i++
				}
				// Reverse this run
				end := i - 1
				for start < end {
					indices[start], indices[end] = indices[end], indices[start]
					start++
					end--
				}
			} else {
				i++
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
