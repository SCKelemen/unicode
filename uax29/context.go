package uax29

// GraphemeBreakContext manages state for grapheme cluster boundary detection.
// It provides a clean abstraction over the text and classification data,
// making rule implementation straightforward and maintainable.
type GraphemeBreakContext struct {
	// Input data (immutable)
	text   string
	runes  []rune
	classes []GraphemeBreakClass

	// Position tracking
	pos int

	// Cached lookups (updated on Slide())
	prevClass GraphemeBreakClass
	currClass GraphemeBreakClass
	nextClass GraphemeBreakClass

	// Byte position tracking for output
	bytePositions []int
}

// NewGraphemeBreakContext creates a context for grapheme cluster boundary detection.
// It pre-classifies all runes using the packed data structure for efficiency.
func NewGraphemeBreakContext(text string) *GraphemeBreakContext {
	if text == "" {
		return &GraphemeBreakContext{
			text: "",
			runes: []rune{},
			classes: []GraphemeBreakClass{},
			pos: -1,
		}
	}

	runes := []rune(text)
	n := len(runes)

	// Pre-classify all runes using packed data
	classes := make([]GraphemeBreakClass, n)
	for i, r := range runes {
		classes[i] = classifyRune(r).Grapheme()
	}

	// Pre-compute byte positions for all rune boundaries
	bytePositions := make([]int, n+1)
	bytePositions[0] = 0
	bytePos := 0
	for i, r := range runes {
		bytePos += len(string(r))
		bytePositions[i+1] = bytePos
	}

	ctx := &GraphemeBreakContext{
		text: text,
		runes: runes,
		classes: classes,
		pos: 0,
		bytePositions: bytePositions,
	}

	// Initialize cached classes
	ctx.updateCache()

	return ctx
}

// NewGraphemeBreakContextFromClasses creates a context using pre-classified data.
// This is used by the single-pass API to avoid redundant classification.
func NewGraphemeBreakContextFromClasses(text string, runes []rune, packedClasses []PackedBreakClass) *GraphemeBreakContext {
	if len(runes) == 0 {
		return &GraphemeBreakContext{
			text: text,
			runes: []rune{},
			classes: []GraphemeBreakClass{},
			pos: -1,
		}
	}

	n := len(runes)

	// Extract grapheme classes from packed data
	classes := make([]GraphemeBreakClass, n)
	for i := range runes {
		classes[i] = packedClasses[i].Grapheme()
	}

	// Pre-compute byte positions for all rune boundaries
	bytePositions := make([]int, n+1)
	bytePositions[0] = 0
	bytePos := 0
	for i, r := range runes {
		bytePos += len(string(r))
		bytePositions[i+1] = bytePos
	}

	ctx := &GraphemeBreakContext{
		text: text,
		runes: runes,
		classes: classes,
		pos: 0,
		bytePositions: bytePositions,
	}

	// Initialize cached classes
	ctx.updateCache()

	return ctx
}

// Slide advances to the next position and returns true if there are more positions to check.
// Returns false when the end of the text is reached.
func (c *GraphemeBreakContext) Slide() bool {
	c.pos++
	if c.pos >= len(c.runes) {
		return false
	}
	c.updateCache()
	return true
}

// updateCache updates the cached prev/curr/next class values for the current position.
func (c *GraphemeBreakContext) updateCache() {
	if c.pos > 0 {
		c.prevClass = c.classes[c.pos-1]
	} else {
		c.prevClass = GBOther
	}

	if c.pos < len(c.classes) {
		c.currClass = c.classes[c.pos]
	} else {
		c.currClass = GBOther
	}

	if c.pos+1 < len(c.classes) {
		c.nextClass = c.classes[c.pos+1]
	} else {
		c.nextClass = GBOther
	}
}

// BytePos returns the byte offset for the current position.
func (c *GraphemeBreakContext) BytePos() int {
	if c.pos < 0 || c.pos >= len(c.bytePositions) {
		return len(c.text)
	}
	return c.bytePositions[c.pos]
}

// Prev returns the grapheme break class of the previous character.
func (c *GraphemeBreakContext) Prev() GraphemeBreakClass {
	return c.prevClass
}

// Curr returns the grapheme break class of the current character.
func (c *GraphemeBreakContext) Curr() GraphemeBreakClass {
	return c.currClass
}

// Next returns the grapheme break class of the next character.
func (c *GraphemeBreakContext) Next() GraphemeBreakClass {
	return c.nextClass
}

// PeekBehind looks backwards through Extend/ZWJ characters to find the first
// non-ignorable class. Returns the class and its position, or GBOther if not found.
func (c *GraphemeBreakContext) PeekBehind(ignoreMask GraphemeBreakClass) (GraphemeBreakClass, int) {
	pos := c.pos - 1
	for pos >= 0 {
		class := c.classes[pos]
		if class != GBExtend && class != GBZWJ {
			return class, pos
		}
		pos--
	}
	return GBOther, -1
}

// PeekAhead looks forward through Extend/ZWJ characters to find the first
// non-ignorable class. Returns the class and its position, or GBOther if not found.
func (c *GraphemeBreakContext) PeekAhead(ignoreMask GraphemeBreakClass) (GraphemeBreakClass, int) {
	pos := c.pos + 1
	for pos < len(c.classes) {
		class := c.classes[pos]
		if class != GBExtend && class != GBZWJ {
			return class, pos
		}
		pos++
	}
	return GBOther, -1
}

// Rune returns the rune at the current position.
func (c *GraphemeBreakContext) Rune() rune {
	if c.pos < 0 || c.pos >= len(c.runes) {
		return 0
	}
	return c.runes[c.pos]
}

// RuneAt returns the rune at the specified position.
func (c *GraphemeBreakContext) RuneAt(pos int) rune {
	if pos < 0 || pos >= len(c.runes) {
		return 0
	}
	return c.runes[pos]
}

// ClassAt returns the grapheme break class at the specified position.
func (c *GraphemeBreakContext) ClassAt(pos int) GraphemeBreakClass {
	if pos < 0 || pos >= len(c.classes) {
		return GBOther
	}
	return c.classes[pos]
}

// Pos returns the current rune position.
func (c *GraphemeBreakContext) Pos() int {
	return c.pos
}

// Len returns the total number of runes.
func (c *GraphemeBreakContext) Len() int {
	return len(c.runes)
}

// WordBreakContext manages state for word boundary detection.
// It provides a clean abstraction over the text and classification data,
// making rule implementation straightforward and maintainable.
type WordBreakContext struct {
	// Input data (immutable)
	text   string
	runes  []rune
	classes []WordBreakClass

	// Grapheme boundary positions (rune indices)
	graphemeBoundaries []int

	// Position tracking (index into graphemeBoundaries)
	boundaryIdx int

	// Cached lookups (updated on Slide())
	prevClass WordBreakClass
	currClass WordBreakClass
	nextClass WordBreakClass

	// Byte position tracking for output
	bytePositions []int
}

// NewWordBreakContextFromClasses creates a context for word boundary detection
// using pre-classified data. Only checks at grapheme boundaries.
func NewWordBreakContextFromClasses(text string, runes []rune, packedClasses []PackedBreakClass, graphemeBreaks []int) *WordBreakContext {
	if len(runes) == 0 {
		return &WordBreakContext{
			text: text,
			runes: []rune{},
			classes: []WordBreakClass{},
			graphemeBoundaries: []int{},
			boundaryIdx: -1,
		}
	}

	n := len(runes)

	// Extract word classes from packed data
	classes := make([]WordBreakClass, n)
	for i := range runes {
		classes[i] = packedClasses[i].Word()
	}

	// Convert grapheme break byte positions to rune indices
	graphemeBoundaries := make([]int, len(graphemeBreaks))
	byteToRune := 0
	runeIdx := 0
	for i, bytePos := range graphemeBreaks {
		for byteToRune < bytePos && runeIdx < len(runes) {
			byteToRune += len(string(runes[runeIdx]))
			runeIdx++
		}
		graphemeBoundaries[i] = runeIdx
	}

	// Pre-compute byte positions for all rune boundaries
	bytePositions := make([]int, n+1)
	bytePositions[0] = 0
	bytePos := 0
	for i, r := range runes {
		bytePos += len(string(r))
		bytePositions[i+1] = bytePos
	}

	ctx := &WordBreakContext{
		text: text,
		runes: runes,
		classes: classes,
		graphemeBoundaries: graphemeBoundaries,
		boundaryIdx: 0,
		bytePositions: bytePositions,
	}

	// Initialize cached classes
	ctx.updateCache()

	return ctx
}

// Slide advances to the next grapheme boundary and returns true if there are more to check.
// Returns false when the end of the text is reached.
func (w *WordBreakContext) Slide() bool {
	w.boundaryIdx++
	if w.boundaryIdx >= len(w.graphemeBoundaries)-1 {
		return false
	}
	w.updateCache()
	return true
}

// updateCache updates the cached prev/curr/next class values for the current position.
func (w *WordBreakContext) updateCache() {
	pos := w.graphemeBoundaries[w.boundaryIdx]

	if pos > 0 {
		w.prevClass = w.classes[pos-1]
	} else {
		w.prevClass = WBOther
	}

	if pos < len(w.classes) {
		w.currClass = w.classes[pos]
	} else {
		w.currClass = WBOther
	}

	if pos+1 < len(w.classes) {
		w.nextClass = w.classes[pos+1]
	} else {
		w.nextClass = WBOther
	}
}

// BytePos returns the byte offset for the current grapheme boundary position.
func (w *WordBreakContext) BytePos() int {
	pos := w.graphemeBoundaries[w.boundaryIdx]
	if pos < 0 || pos >= len(w.bytePositions) {
		return len(w.text)
	}
	return w.bytePositions[pos]
}

// Pos returns the current rune position (at grapheme boundary).
func (w *WordBreakContext) Pos() int {
	if w.boundaryIdx < 0 || w.boundaryIdx >= len(w.graphemeBoundaries) {
		return len(w.runes)
	}
	return w.graphemeBoundaries[w.boundaryIdx]
}

// Prev returns the word break class of the previous character.
func (w *WordBreakContext) Prev() WordBreakClass {
	return w.prevClass
}

// Curr returns the word break class of the current character.
func (w *WordBreakContext) Curr() WordBreakClass {
	return w.currClass
}

// Next returns the word break class of the next character.
func (w *WordBreakContext) Next() WordBreakClass {
	return w.nextClass
}

// Rune returns the rune at the current position.
func (w *WordBreakContext) Rune() rune {
	pos := w.Pos()
	if pos < 0 || pos >= len(w.runes) {
		return 0
	}
	return w.runes[pos]
}

// RuneAt returns the rune at the specified position.
func (w *WordBreakContext) RuneAt(pos int) rune {
	if pos < 0 || pos >= len(w.runes) {
		return 0
	}
	return w.runes[pos]
}

// ClassAt returns the word break class at the specified position.
func (w *WordBreakContext) ClassAt(pos int) WordBreakClass {
	if pos < 0 || pos >= len(w.classes) {
		return WBOther
	}
	return w.classes[pos]
}

// PrevNonIgnorable returns the previous non-Format/Extend/ZWJ class and its position.
func (w *WordBreakContext) PrevNonIgnorable() (WordBreakClass, int) {
	pos := w.Pos() - 1
	for pos > 0 && (w.classes[pos] == WBFormat || w.classes[pos] == WBExtend || w.classes[pos] == WBZWJ) {
		pos--
	}
	if pos < 0 {
		return WBOther, -1
	}
	return w.classes[pos], pos
}

// NextNonIgnorable returns the next non-Format/Extend/ZWJ class and its position.
func (w *WordBreakContext) NextNonIgnorable() (WordBreakClass, int) {
	pos := w.Pos() + 1
	for pos < len(w.classes) && (w.classes[pos] == WBFormat || w.classes[pos] == WBExtend || w.classes[pos] == WBZWJ) {
		pos++
	}
	if pos >= len(w.classes) {
		return WBOther, -1
	}
	return w.classes[pos], pos
}

// Len returns the total number of runes.
func (w *WordBreakContext) Len() int {
	return len(w.runes)
}

// SentenceBreakContext manages state for sentence boundary detection.
// It provides a clean abstraction over the text and classification data,
// making rule implementation straightforward and maintainable.
type SentenceBreakContext struct {
	// Input data (immutable)
	text   string
	runes  []rune
	classes []SentenceBreakClass

	// Word boundary positions (rune indices)
	wordBoundaries []int

	// Position tracking (index into wordBoundaries)
	boundaryIdx int

	// Cached lookups (updated on Slide())
	prevClass SentenceBreakClass
	currClass SentenceBreakClass
	nextClass SentenceBreakClass

	// Byte position tracking for output
	bytePositions []int
}

// NewSentenceBreakContextFromClasses creates a context for sentence boundary detection
// using pre-classified data. Only checks at word boundaries.
func NewSentenceBreakContextFromClasses(text string, runes []rune, packedClasses []PackedBreakClass, wordBreaks []int) *SentenceBreakContext {
	if len(runes) == 0 {
		return &SentenceBreakContext{
			text: text,
			runes: []rune{},
			classes: []SentenceBreakClass{},
			wordBoundaries: []int{},
			boundaryIdx: -1,
		}
	}

	n := len(runes)

	// Extract sentence classes from packed data
	classes := make([]SentenceBreakClass, n)
	for i := range runes {
		classes[i] = packedClasses[i].Sentence()
	}

	// Convert word break byte positions to rune indices
	wordBoundaries := make([]int, len(wordBreaks))
	byteToRune := 0
	runeIdx := 0
	for i, bytePos := range wordBreaks {
		for byteToRune < bytePos && runeIdx < len(runes) {
			byteToRune += len(string(runes[runeIdx]))
			runeIdx++
		}
		wordBoundaries[i] = runeIdx
	}

	// Pre-compute byte positions for all rune boundaries
	bytePositions := make([]int, n+1)
	bytePositions[0] = 0
	bytePos := 0
	for i, r := range runes {
		bytePos += len(string(r))
		bytePositions[i+1] = bytePos
	}

	ctx := &SentenceBreakContext{
		text: text,
		runes: runes,
		classes: classes,
		wordBoundaries: wordBoundaries,
		boundaryIdx: 0,
		bytePositions: bytePositions,
	}

	// Initialize cached classes
	ctx.updateCache()

	return ctx
}

// Slide advances to the next word boundary and returns true if there are more to check.
// Returns false when the end of the text is reached.
func (s *SentenceBreakContext) Slide() bool {
	s.boundaryIdx++
	if s.boundaryIdx >= len(s.wordBoundaries)-1 {
		return false
	}
	s.updateCache()
	return true
}

// updateCache updates the cached prev/curr/next class values for the current position.
func (s *SentenceBreakContext) updateCache() {
	pos := s.wordBoundaries[s.boundaryIdx]

	if pos > 0 {
		s.prevClass = s.classes[pos-1]
	} else {
		s.prevClass = SBOther
	}

	if pos < len(s.classes) {
		s.currClass = s.classes[pos]
	} else {
		s.currClass = SBOther
	}

	if pos+1 < len(s.classes) {
		s.nextClass = s.classes[pos+1]
	} else {
		s.nextClass = SBOther
	}
}

// BytePos returns the byte offset for the current word boundary position.
func (s *SentenceBreakContext) BytePos() int {
	pos := s.wordBoundaries[s.boundaryIdx]
	if pos < 0 || pos >= len(s.bytePositions) {
		return len(s.text)
	}
	return s.bytePositions[pos]
}

// Pos returns the current rune position (at word boundary).
func (s *SentenceBreakContext) Pos() int {
	if s.boundaryIdx < 0 || s.boundaryIdx >= len(s.wordBoundaries) {
		return len(s.runes)
	}
	return s.wordBoundaries[s.boundaryIdx]
}

// Prev returns the sentence break class of the previous character.
func (s *SentenceBreakContext) Prev() SentenceBreakClass {
	return s.prevClass
}

// Curr returns the sentence break class of the current character.
func (s *SentenceBreakContext) Curr() SentenceBreakClass {
	return s.currClass
}

// Next returns the sentence break class of the next character.
func (s *SentenceBreakContext) Next() SentenceBreakClass {
	return s.nextClass
}

// Rune returns the rune at the current position.
func (s *SentenceBreakContext) Rune() rune {
	pos := s.Pos()
	if pos < 0 || pos >= len(s.runes) {
		return 0
	}
	return s.runes[pos]
}

// RuneAt returns the rune at the specified position.
func (s *SentenceBreakContext) RuneAt(pos int) rune {
	if pos < 0 || pos >= len(s.runes) {
		return 0
	}
	return s.runes[pos]
}

// ClassAt returns the sentence break class at the specified position.
func (s *SentenceBreakContext) ClassAt(pos int) SentenceBreakClass {
	if pos < 0 || pos >= len(s.classes) {
		return SBOther
	}
	return s.classes[pos]
}

// PrevNonIgnorable returns the previous non-Format/Extend class and its position.
func (s *SentenceBreakContext) PrevNonIgnorable() (SentenceBreakClass, int) {
	pos := s.Pos() - 1
	for pos > 0 && (s.classes[pos] == SBFormat || s.classes[pos] == SBExtend) {
		pos--
	}
	if pos < 0 {
		return SBOther, -1
	}
	return s.classes[pos], pos
}

// NextNonIgnorable returns the next non-Format/Extend class and its position.
func (s *SentenceBreakContext) NextNonIgnorable() (SentenceBreakClass, int) {
	pos := s.Pos() + 1
	for pos < len(s.classes) && (s.classes[pos] == SBFormat || s.classes[pos] == SBExtend) {
		pos++
	}
	if pos >= len(s.classes) {
		return SBOther, -1
	}
	return s.classes[pos], pos
}

// LookAhead scans forward through Close/Sp/Format/Extend characters.
// Returns the first non-ignorable class found and its position.
func (s *SentenceBreakContext) LookAhead() (SentenceBreakClass, int) {
	pos := s.Pos()
	for pos < len(s.classes) {
		class := s.classes[pos]
		if class != SBClose && class != SBSp && class != SBFormat && class != SBExtend {
			return class, pos
		}
		pos++
	}
	return SBOther, -1
}

// HasATermBefore checks if there's an ATerm before the current position,
// possibly with Close/Sp in between. Returns true and the ATerm position if found.
func (s *SentenceBreakContext) HasATermBefore() (bool, int) {
	pos := s.Pos() - 1
	// Skip back through Close/Sp/Format/Extend
	for pos >= 0 {
		class := s.classes[pos]
		if class == SBATerm || class == SBSTerm {
			return true, pos
		}
		if class != SBClose && class != SBSp && class != SBFormat && class != SBExtend {
			return false, -1
		}
		pos--
	}
	return false, -1
}

// Len returns the total number of runes.
func (s *SentenceBreakContext) Len() int {
	return len(s.runes)
}
