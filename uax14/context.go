package uax14

import "unicode"

// isCombiningMark checks if a rune is a combining mark (Mn or Mc).
// Used for LB9 rule: SA characters that are combining marks should be treated as CM.
func isCombiningMark(r rune) bool {
	return unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r)
}

// LineBreakContext manages state for line break opportunity detection.
// It provides a clean abstraction over the text and classification data,
// making rule implementation straightforward and maintainable.
type LineBreakContext struct {
	// Input data (immutable)
	text   string
	runes  []rune
	classes []BreakClass
	hyphens Hyphens

	// Position tracking
	pos int

	// Cached lookups (updated on Slide())
	prevClass BreakClass
	currClass BreakClass
	nextClass BreakClass

	// State tracking for special rules
	lastNonSpaceClass BreakClass // For LB14 and other space-sensitive rules

	// Byte position tracking for output
	bytePositions []int
}

// NewLineBreakContext creates a context for line break opportunity detection.
// It pre-classifies all runes for efficiency.
func NewLineBreakContext(text string, hyphens Hyphens) *LineBreakContext {
	if text == "" {
		return &LineBreakContext{
			text: text,
			runes: []rune{},
			classes: []BreakClass{},
			hyphens: hyphens,
			pos: -1,
		}
	}

	runes := []rune(text)
	n := len(runes)

	// Pre-classify all runes
	classes := make([]BreakClass, n)
	for i, r := range runes {
		classes[i] = getBreakClass(r)
	}

	// Pre-compute byte positions for all rune boundaries
	bytePositions := make([]int, n+1)
	bytePositions[0] = 0
	bytePos := 0
	for i, r := range runes {
		bytePos += len(string(r))
		bytePositions[i+1] = bytePos
	}

	ctx := &LineBreakContext{
		text: text,
		runes: runes,
		classes: classes,
		hyphens: hyphens,
		pos: 0,
		bytePositions: bytePositions,
	}

	// Initialize cached classes and apply LB10 to first character
	ctx.updateCache()

	// LB10: Treat CM or ZWJ at start of text as AL
	if isClassOrVariant(ctx.prevClass, ClassCM) || ctx.prevClass == ClassZWJ {
		ctx.prevClass = ClassAL
	}
	ctx.lastNonSpaceClass = ctx.prevClass

	return ctx
}

// Slide advances to the next position and returns true if there are more positions to check.
// Returns false when the end of the text is reached.
func (c *LineBreakContext) Slide() bool {
	c.pos++
	if c.pos >= len(c.runes) {
		return false
	}
	c.updateCache()
	return true
}

// updateCache updates the cached prev/curr/next class values for the current position.
func (c *LineBreakContext) updateCache() {
	if c.pos > 0 {
		c.prevClass = c.classes[c.pos-1]
	} else {
		c.prevClass = ClassAL // Default for position 0
	}

	if c.pos < len(c.classes) {
		c.currClass = c.classes[c.pos]

		// LB9: SA characters that are combining marks (Mn, Mc) should be treated as CM
		if c.currClass == ClassSA {
			// Check if it's a combining mark
			if isCombiningMark(c.runes[c.pos]) {
				c.currClass = ClassCM
			}
		}
	} else {
		c.currClass = ClassAL
	}

	if c.pos+1 < len(c.classes) {
		c.nextClass = c.classes[c.pos+1]
	} else {
		c.nextClass = ClassAL
	}
}

// BytePos returns the byte offset for the current position.
func (c *LineBreakContext) BytePos() int {
	if c.pos < 0 || c.pos >= len(c.bytePositions) {
		return len(c.text)
	}
	return c.bytePositions[c.pos]
}

// Pos returns the current rune position.
func (c *LineBreakContext) Pos() int {
	return c.pos
}

// Prev returns the line break class of the previous character.
func (c *LineBreakContext) Prev() BreakClass {
	return c.prevClass
}

// Curr returns the line break class of the current character.
func (c *LineBreakContext) Curr() BreakClass {
	return c.currClass
}

// Next returns the line break class of the next character.
func (c *LineBreakContext) Next() BreakClass {
	return c.nextClass
}

// Rune returns the rune at the current position.
func (c *LineBreakContext) Rune() rune {
	if c.pos < 0 || c.pos >= len(c.runes) {
		return 0
	}
	return c.runes[c.pos]
}

// RuneAt returns the rune at the specified position.
func (c *LineBreakContext) RuneAt(pos int) rune {
	if pos < 0 || pos >= len(c.runes) {
		return 0
	}
	return c.runes[pos]
}

// ClassAt returns the line break class at the specified position.
func (c *LineBreakContext) ClassAt(pos int) BreakClass {
	if pos < 0 || pos >= len(c.classes) {
		return ClassAL // Default
	}
	return c.classes[pos]
}

// LastNonSpace returns the last non-SP class seen.
func (c *LineBreakContext) LastNonSpace() BreakClass {
	return c.lastNonSpaceClass
}

// UpdatePrevClass updates the previous class (for state transitions after breaks).
func (c *LineBreakContext) UpdatePrevClass(class BreakClass) {
	c.prevClass = class
	if class != ClassSP {
		c.lastNonSpaceClass = class
	}
}

// UpdateLastNonSpace updates the last non-space class tracker.
func (c *LineBreakContext) UpdateLastNonSpace(class BreakClass) {
	c.lastNonSpaceClass = class
}

// Len returns the total number of runes.
func (c *LineBreakContext) Len() int {
	return len(c.runes)
}

// Hyphens returns the hyphenation mode.
func (c *LineBreakContext) Hyphens() Hyphens {
	return c.hyphens
}

// LookBack looks backward from the current position for a specific pattern.
// Returns the position if found, -1 otherwise.
func (c *LineBreakContext) LookBack(distance int) int {
	targetPos := c.pos - distance
	if targetPos < 0 {
		return -1
	}
	return targetPos
}

// GetPairTableAction looks up the break action from the pair table.
func (c *LineBreakContext) GetPairTableAction(prev, curr BreakClass) BreakAction {
	return getBreakAction(prev, curr)
}

// SkipBackward skips over specified classes backward from startIdx, returning the index
// of the first non-skipped class (or -1 if start of text reached).
// Commonly used to skip CM/ZWJ per LB9: treat X (CM | ZWJ)* as X
func (c *LineBreakContext) SkipBackward(startIdx int, skipClasses ...BreakClass) int {
	idx := startIdx
	for idx >= 0 {
		class := c.ClassAt(idx)
		shouldSkip := false
		for _, skipClass := range skipClasses {
			if class == skipClass || isClassOrVariant(class, skipClass) {
				shouldSkip = true
				break
			}
		}
		if !shouldSkip {
			return idx
		}
		idx--
	}
	return -1
}

// SkipForward skips over specified classes forward from startIdx, returning the index
// of the first non-skipped class (or c.Len() if end of text reached).
func (c *LineBreakContext) SkipForward(startIdx int, skipClasses ...BreakClass) int {
	idx := startIdx
	for idx < c.Len() {
		class := c.ClassAt(idx)
		shouldSkip := false
		for _, skipClass := range skipClasses {
			if class == skipClass || isClassOrVariant(class, skipClass) {
				shouldSkip = true
				break
			}
		}
		if !shouldSkip {
			return idx
		}
		idx++
	}
	return c.Len()
}

// FindForward searches forward from startIdx until it finds one of the target classes.
// Returns the index where found, or -1 if not found within maxDistance.
// Optionally skips over specified classes (like CM/ZWJ).
func (c *LineBreakContext) FindForward(startIdx int, maxDistance int, targetClasses []BreakClass, skipClasses ...BreakClass) int {
	idx := startIdx
	limit := c.Len()
	if maxDistance > 0 && startIdx+maxDistance < limit {
		limit = startIdx + maxDistance
	}

	for idx < limit {
		class := c.ClassAt(idx)

		// Check if this is a skip class
		shouldSkip := false
		for _, skipClass := range skipClasses {
			if class == skipClass || isClassOrVariant(class, skipClass) {
				shouldSkip = true
				break
			}
		}
		if shouldSkip {
			idx++
			continue
		}

		// Check if this is a target class
		for _, target := range targetClasses {
			if class == target || isClassOrVariant(class, target) {
				return idx
			}
		}

		// Not a target, stop searching
		return -1
	}
	return -1
}

// FindBackward searches backward from startIdx until it finds one of the target classes.
// Returns the index where found, or -1 if not found.
func (c *LineBreakContext) FindBackward(startIdx int, targetClasses []BreakClass, skipClasses ...BreakClass) int {
	idx := startIdx

	for idx >= 0 {
		class := c.ClassAt(idx)

		// Check if this is a skip class
		shouldSkip := false
		for _, skipClass := range skipClasses {
			if class == skipClass || isClassOrVariant(class, skipClass) {
				shouldSkip = true
				break
			}
		}
		if shouldSkip {
			idx--
			continue
		}

		// Check if this is a target class
		for _, target := range targetClasses {
			if class == target || isClassOrVariant(class, target) {
				return idx
			}
		}

		// Not a target, stop searching
		return -1
	}
	return -1
}

// MatchSequence checks if a sequence of classes matches forward from startIdx.
// Returns true if the entire sequence matches.
func (c *LineBreakContext) MatchSequence(startIdx int, sequence ...BreakClass) bool {
	for i, class := range sequence {
		idx := startIdx + i
		if idx >= c.Len() {
			return false
		}
		actualClass := c.ClassAt(idx)
		if actualClass != class && !isClassOrVariant(actualClass, class) {
			return false
		}
	}
	return true
}
