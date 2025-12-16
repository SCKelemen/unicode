package uax29

// BreakAction represents the decision about whether to break at the current position.
type BreakAction int

const (
	BreakNo  BreakAction = iota // × - Don't break
	BreakYes                     // ÷ - Break allowed
)

// GraphemeBreakRule checks if a rule applies and returns the action.
type GraphemeBreakRule func(ctx *GraphemeBreakContext) (matched bool, action BreakAction)

// ruleGB3 implements: CR × LF
// Do not break between a CR and LF. Otherwise, break before and after controls.
// https://www.unicode.org/reports/tr29/#GB3
func ruleGB3(ctx *GraphemeBreakContext) (bool, BreakAction) {
	if ctx.Prev() == GBCR && ctx.Curr() == GBLF {
		return true, BreakNo
	}
	return false, BreakYes
}

// ruleGB4 implements: (Control | CR | LF) ÷
// Break after controls.
// https://www.unicode.org/reports/tr29/#GB4
func ruleGB4(ctx *GraphemeBreakContext) (bool, BreakAction) {
	prev := ctx.Prev()
	if prev == GBControl || prev == GBCR || prev == GBLF {
		return true, BreakYes
	}
	return false, BreakNo
}

// ruleGB5 implements: ÷ (Control | CR | LF)
// Break before controls.
// https://www.unicode.org/reports/tr29/#GB5
func ruleGB5(ctx *GraphemeBreakContext) (bool, BreakAction) {
	curr := ctx.Curr()
	if curr == GBControl || curr == GBCR || curr == GBLF {
		return true, BreakYes
	}
	return false, BreakNo
}

// ruleGB6 implements: L × (L | V | LV | LVT)
// Do not break Hangul syllable sequences.
// https://www.unicode.org/reports/tr29/#GB6
func ruleGB6(ctx *GraphemeBreakContext) (bool, BreakAction) {
	if ctx.Prev() == GBL {
		curr := ctx.Curr()
		if curr == GBL || curr == GBV || curr == GBLV || curr == GBLVT {
			return true, BreakNo
		}
	}
	return false, BreakYes
}

// ruleGB7 implements: (LV | V) × (V | T)
// Do not break Hangul syllable sequences.
// https://www.unicode.org/reports/tr29/#GB7
func ruleGB7(ctx *GraphemeBreakContext) (bool, BreakAction) {
	prev := ctx.Prev()
	if prev == GBLV || prev == GBV {
		curr := ctx.Curr()
		if curr == GBV || curr == GBT {
			return true, BreakNo
		}
	}
	return false, BreakYes
}

// ruleGB8 implements: (LVT | T) × T
// Do not break Hangul syllable sequences.
// https://www.unicode.org/reports/tr29/#GB8
func ruleGB8(ctx *GraphemeBreakContext) (bool, BreakAction) {
	prev := ctx.Prev()
	if prev == GBLVT || prev == GBT {
		if ctx.Curr() == GBT {
			return true, BreakNo
		}
	}
	return false, BreakYes
}

// ruleGB9 implements: × (Extend | ZWJ)
// Do not break before extending characters or ZWJ.
// https://www.unicode.org/reports/tr29/#GB9
func ruleGB9(ctx *GraphemeBreakContext) (bool, BreakAction) {
	curr := ctx.Curr()
	if curr == GBExtend || curr == GBZWJ {
		return true, BreakNo
	}
	return false, BreakYes
}

// ruleGB9a implements: × SpacingMark
// Do not break before SpacingMarks.
// https://www.unicode.org/reports/tr29/#GB9a
func ruleGB9a(ctx *GraphemeBreakContext) (bool, BreakAction) {
	if ctx.Curr() == GBSpacingMark {
		return true, BreakNo
	}
	return false, BreakYes
}

// ruleGB9b implements: Prepend ×
// Do not break after Prepend characters.
// https://www.unicode.org/reports/tr29/#GB9b
func ruleGB9b(ctx *GraphemeBreakContext) (bool, BreakAction) {
	if ctx.Prev() == GBPrepend {
		return true, BreakNo
	}
	return false, BreakYes
}

// ruleGB9c implements: InCB=Consonant [InCB=Extend InCB=Linker]* InCB=Linker [InCB=Extend InCB=Linker]* × InCB=Consonant
// Do not break within certain combinations with Indic_Conjunct_Break (InCB)=Linker.
// https://www.unicode.org/reports/tr29/#GB9c
func ruleGB9c(ctx *GraphemeBreakContext) (bool, BreakAction) {
	// Check if current is Indic Conjunct Consonant
	if !isIndicConjunctConsonant(ctx.Rune()) {
		return false, BreakYes
	}

	// Look back through Extend/ZWJ/Linker for a Linker
	j := ctx.Pos() - 1
	foundLinker := false
	for j >= 0 {
		if isIndicConjunctLinker(ctx.RuneAt(j)) {
			foundLinker = true
			j--
			break
		}
		rClass := ctx.ClassAt(j)
		if rClass == GBExtend || rClass == GBZWJ {
			j--
			continue
		}
		break
	}

	if !foundLinker {
		return false, BreakYes
	}

	// Continue looking back through Extend/Linker for a Consonant
	for j >= 0 {
		if isIndicConjunctConsonant(ctx.RuneAt(j)) {
			// Found the pattern: Consonant ... Linker ... Consonant
			return true, BreakNo
		}
		if isIndicConjunctLinker(ctx.RuneAt(j)) {
			j--
			continue
		}
		rClass := ctx.ClassAt(j)
		if rClass == GBExtend || rClass == GBZWJ {
			j--
			continue
		}
		break
	}

	return false, BreakYes
}

// ruleGB11 implements: ExtendedPictographic Extend* ZWJ × ExtendedPictographic
// Do not break within emoji ZWJ sequences.
// https://www.unicode.org/reports/tr29/#GB11
func ruleGB11(ctx *GraphemeBreakContext) (bool, BreakAction) {
	// Current must be ExtendedPictographic
	if !isExtendedPictographic(ctx.Rune()) {
		return false, BreakYes
	}

	// Look back through Extend for ZWJ
	j := ctx.Pos() - 1
	for j >= 0 && ctx.ClassAt(j) == GBExtend {
		j--
	}

	// Check if we have ZWJ
	if j < 0 || ctx.ClassAt(j) != GBZWJ {
		return false, BreakYes
	}

	// Look back through Extend for ExtendedPictographic
	j--
	for j >= 0 && ctx.ClassAt(j) == GBExtend {
		j--
	}

	if j >= 0 && isExtendedPictographic(ctx.RuneAt(j)) {
		// Found the pattern: ExtPict Extend* ZWJ Extend* ExtPict
		return true, BreakNo
	}

	return false, BreakYes
}

// ruleGB12 implements: sot (RI RI)* RI × RI
// ruleGB13 implements: [^RI] (RI RI)* RI × RI
// Do not break within emoji flag sequences. That is, do not break between regional indicator (RI) symbols
// if there is an odd number of RI characters before the break point.
// https://www.unicode.org/reports/tr29/#GB12
// https://www.unicode.org/reports/tr29/#GB13
func ruleGB12_13(ctx *GraphemeBreakContext) (bool, BreakAction) {
	if ctx.Prev() != GBRegionalIndicator || ctx.Curr() != GBRegionalIndicator {
		return false, BreakYes
	}

	// Count consecutive RIs before current position
	riCountBefore := 0
	for j := ctx.Pos() - 1; j >= 0 && ctx.ClassAt(j) == GBRegionalIndicator; j-- {
		riCountBefore++
	}

	// If odd number of RIs before, don't break (pair with previous RI)
	if riCountBefore%2 == 1 {
		return true, BreakNo
	}

	return false, BreakYes
}

// graphemeBreakRules is the ordered list of all grapheme boundary rules.
// Rules are checked in order - first match wins.
var graphemeBreakRules = []GraphemeBreakRule{
	ruleGB3,    // CR × LF
	ruleGB4,    // (Control | CR | LF) ÷
	ruleGB5,    // ÷ (Control | CR | LF)
	ruleGB6,    // L × (L | V | LV | LVT)
	ruleGB7,    // (LV | V) × (V | T)
	ruleGB8,    // (LVT | T) × T
	ruleGB9,    // × (Extend | ZWJ)
	ruleGB9a,   // × SpacingMark
	ruleGB9b,   // Prepend ×
	ruleGB9c,   // Indic conjunct sequences
	ruleGB11,   // ExtPict Extend* ZWJ × ExtPict
	ruleGB12_13, // RI × RI (pairs)
	// GB999: Otherwise, break everywhere (default in main loop)
}

// FindGraphemeBreaksWithRules finds grapheme cluster boundaries using the rule-based approach.
// This is an alternative implementation to FindGraphemeBreaks for testing and benchmarking.
func FindGraphemeBreaksWithRules(text string) []int {
	if text == "" {
		return []int{}
	}

	ctx := NewGraphemeBreakContext(text)
	breaks := []int{0} // GB1: Break at start

	for ctx.Slide() {
		// Apply rules in order - first match wins
		action := BreakYes // GB999: Default is break

		for _, rule := range graphemeBreakRules {
			if matched, ruleAction := rule(ctx); matched {
				action = ruleAction
				break // Stop at first matching rule
			}
		}

		if action == BreakYes {
			breaks = append(breaks, ctx.BytePos())
		}
	}

	// GB2: Break at end
	breaks = append(breaks, len(text))

	return breaks
}

// findGraphemeBreaksFromClassesWithRules finds grapheme cluster boundaries using
// the rule-based approach with pre-classified data. This is used by the single-pass
// API to avoid redundant classification.
func findGraphemeBreaksFromClassesWithRules(text string, runes []rune, classes []PackedBreakClass) []int {
	if len(runes) == 0 {
		return []int{}
	}

	ctx := NewGraphemeBreakContextFromClasses(text, runes, classes)
	breaks := []int{0} // GB1: Break at start

	for ctx.Slide() {
		// Apply rules in order - first match wins
		action := BreakYes // GB999: Default is break

		for _, rule := range graphemeBreakRules {
			if matched, ruleAction := rule(ctx); matched {
				action = ruleAction
				break // Stop at first matching rule
			}
		}

		if action == BreakYes {
			breaks = append(breaks, ctx.BytePos())
		}
	}

	// GB2: Break at end
	breaks = append(breaks, len(text))

	return breaks
}
