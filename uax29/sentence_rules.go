package uax29

// SentenceBreakRule checks if a rule applies and returns the action.
type SentenceBreakRule func(ctx *SentenceBreakContext) (matched bool, action BreakAction)

// ruleSB3 implements: CR × LF
// Do not break within CRLF.
// https://www.unicode.org/reports/tr29/#SB3
func ruleSB3(ctx *SentenceBreakContext) (bool, BreakAction) {
	i := ctx.Pos()
	if i > 0 && ctx.ClassAt(i-1) == SBCR && ctx.Curr() == SBLF {
		return true, BreakNo
	}
	return false, BreakYes
}

// ruleSB4 implements: (Sep | CR | LF) ÷
// Break after paragraph separators.
// https://www.unicode.org/reports/tr29/#SB4
func ruleSB4(ctx *SentenceBreakContext) (bool, BreakAction) {
	i := ctx.Pos()
	if i > 0 {
		prevClass := ctx.ClassAt(i - 1)
		if prevClass == SBCR || prevClass == SBLF || prevClass == SBSep {
			return true, BreakYes
		}
	}
	return false, BreakNo
}

// ruleSB5 implements: × (Format | Extend)
// Ignore format and extend characters.
// https://www.unicode.org/reports/tr29/#SB5
func ruleSB5(ctx *SentenceBreakContext) (bool, BreakAction) {
	curr := ctx.Curr()
	if curr == SBFormat || curr == SBExtend {
		return true, BreakNo
	}
	return false, BreakYes
}

// ruleSB6 implements: ATerm × Numeric
// https://www.unicode.org/reports/tr29/#SB6
func ruleSB6(ctx *SentenceBreakContext) (bool, BreakAction) {
	prev, _ := ctx.PrevNonIgnorable()
	curr := ctx.Curr()
	if prev == SBATerm && curr == SBNumeric {
		return true, BreakNo
	}
	return false, BreakYes
}

// ruleSB7 implements: (Upper | Lower) ATerm × Upper
// https://www.unicode.org/reports/tr29/#SB7
func ruleSB7(ctx *SentenceBreakContext) (bool, BreakAction) {
	prev, prevPos := ctx.PrevNonIgnorable()
	curr := ctx.Curr()

	if prev == SBATerm && curr == SBUpper {
		// Look back past ATerm for Upper or Lower
		if prevPos > 0 {
			prevPrevPos := prevPos - 1
			for prevPrevPos > 0 && (ctx.ClassAt(prevPrevPos) == SBFormat || ctx.ClassAt(prevPrevPos) == SBExtend) {
				prevPrevPos--
			}
			if prevPrevPos >= 0 {
				prevPrev := ctx.ClassAt(prevPrevPos)
				if prevPrev == SBUpper || prevPrev == SBLower {
					return true, BreakNo
				}
			}
		}
	}
	return false, BreakYes
}

// ruleSB8 implements: ATerm Close* Sp* × Lower
// Do not break after ATerm followed by lowercase.
// https://www.unicode.org/reports/tr29/#SB8
func ruleSB8(ctx *SentenceBreakContext) (bool, BreakAction) {
	curr := ctx.Curr()
	if curr != SBLower {
		return false, BreakYes
	}

	// Check if there's ATerm or STerm before (with possible Close/Sp in between)
	hasATerm, _ := ctx.HasATermBefore()
	return hasATerm, BreakNo
}

// ruleSB8a implements: (STerm | ATerm) Close* Sp* × (SContinue | STerm | ATerm)
// https://www.unicode.org/reports/tr29/#SB8a
func ruleSB8a(ctx *SentenceBreakContext) (bool, BreakAction) {
	curr := ctx.Curr()
	if curr != SBSContinue && curr != SBSTerm && curr != SBATerm {
		return false, BreakYes
	}

	// Check if there's ATerm or STerm before (with possible Close/Sp in between)
	hasATerm, _ := ctx.HasATermBefore()
	return hasATerm, BreakNo
}

// ruleSB9 implements: (STerm | ATerm) Close* × (Close | Sp | Sep | CR | LF)
// https://www.unicode.org/reports/tr29/#SB9
func ruleSB9(ctx *SentenceBreakContext) (bool, BreakAction) {
	curr := ctx.Curr()
	if curr != SBClose && curr != SBSp && curr != SBSep && curr != SBCR && curr != SBLF {
		return false, BreakYes
	}

	prev, prevPos := ctx.PrevNonIgnorable()
	if prev == SBATerm || prev == SBSTerm {
		return true, BreakNo
	}

	// Check if there's ATerm/STerm followed by Close* before current position
	if prev == SBClose {
		// Look back through Close* for ATerm/STerm
		pos := prevPos - 1
		for pos >= 0 && (ctx.ClassAt(pos) == SBClose || ctx.ClassAt(pos) == SBFormat || ctx.ClassAt(pos) == SBExtend) {
			pos--
		}
		if pos >= 0 && (ctx.ClassAt(pos) == SBATerm || ctx.ClassAt(pos) == SBSTerm) {
			return true, BreakNo
		}
	}

	return false, BreakYes
}

// ruleSB10 implements: (STerm | ATerm) Close* Sp* × (Sp | Sep | CR | LF)
// https://www.unicode.org/reports/tr29/#SB10
func ruleSB10(ctx *SentenceBreakContext) (bool, BreakAction) {
	curr := ctx.Curr()
	if curr != SBSp && curr != SBSep && curr != SBCR && curr != SBLF {
		return false, BreakYes
	}

	// Check if there's ATerm or STerm before (with possible Close/Sp in between)
	hasATerm, _ := ctx.HasATermBefore()
	return hasATerm, BreakNo
}

// ruleSB11 implements: (STerm | ATerm) Close* Sp* ÷
// Break after sentence terminators, but include closing punctuation and trailing spaces.
// https://www.unicode.org/reports/tr29/#SB11
func ruleSB11(ctx *SentenceBreakContext) (bool, BreakAction) {
	prev, prevPos := ctx.PrevNonIgnorable()

	// Direct ATerm/STerm before
	if prev == SBATerm || prev == SBSTerm {
		// Look ahead - if we see certain classes, other rules apply
		curr := ctx.Curr()
		if curr == SBClose || curr == SBSp || curr == SBSep || curr == SBCR || curr == SBLF {
			// SB9 or SB10 handles this
			return false, BreakNo
		}
		if curr == SBSContinue || curr == SBSTerm || curr == SBATerm {
			// SB8a handles this
			return false, BreakNo
		}
		if curr == SBLower {
			// SB8 handles this
			return false, BreakNo
		}
		if curr == SBNumeric && prev == SBATerm {
			// SB6 handles this
			return false, BreakNo
		}
		// Otherwise, break
		return true, BreakYes
	}

	// Check for ATerm/STerm followed by Close* Sp* before current position
	if prev == SBSp || prev == SBClose {
		// Look back through Sp/Close for ATerm/STerm
		pos := prevPos

		for pos >= 0 {
			class := ctx.ClassAt(pos)
			if class == SBSp {
				pos--
			} else if class == SBClose {
				pos--
			} else if class == SBFormat || class == SBExtend {
				pos--
			} else if class == SBATerm || class == SBSTerm {
				// Found pattern: ATerm/STerm Close* Sp*
				// Check what comes next to decide if we should break
				curr := ctx.Curr()
				if curr == SBClose || curr == SBSp || curr == SBSep || curr == SBCR || curr == SBLF {
					// SB9 or SB10 handles this
					return false, BreakNo
				}
				if curr == SBSContinue || curr == SBSTerm || curr == SBATerm {
					// SB8a handles this
					return false, BreakNo
				}
				if curr == SBLower {
					// SB8 handles this
					return false, BreakNo
				}
				// Break here
				return true, BreakYes
			} else {
				// Found something else, not our pattern
				break
			}
		}
	}

	return false, BreakNo
}

// sentenceBreakRules is the ordered list of all sentence boundary rules.
// Rules are checked in order - first match wins.
var sentenceBreakRules = []SentenceBreakRule{
	ruleSB3,  // CR × LF
	ruleSB4,  // (Sep | CR | LF) ÷
	ruleSB5,  // × (Format | Extend)
	ruleSB6,  // ATerm × Numeric
	ruleSB7,  // (Upper | Lower) ATerm × Upper
	ruleSB8,  // ATerm Close* Sp* × Lower
	ruleSB8a, // (STerm | ATerm) Close* Sp* × (SContinue | STerm | ATerm)
	ruleSB9,  // (STerm | ATerm) Close* × (Close | Sp | Sep | CR | LF)
	ruleSB10, // (STerm | ATerm) Close* Sp* × (Sp | Sep | CR | LF)
	ruleSB11, // (STerm | ATerm) Close* Sp* ÷
	// SB998: Do not break by default
}

// findSentenceBreaksFromClassesWithRules finds sentence boundaries using the rule-based
// approach with pre-classified data. This is used by the single-pass API.
func findSentenceBreaksFromClassesWithRules(text string, runes []rune, classes []PackedBreakClass, wordBreaks []int) []int {
	if len(runes) == 0 {
		return []int{}
	}

	ctx := NewSentenceBreakContextFromClasses(text, runes, classes, wordBreaks)
	breaks := []int{0} // SB1: Break at start

	for ctx.Slide() {
		// Apply rules in order - first match wins
		action := BreakNo // SB998: Default is no break

		for _, rule := range sentenceBreakRules {
			if matched, ruleAction := rule(ctx); matched {
				action = ruleAction
				break // Stop at first matching rule
			}
		}

		if action == BreakYes {
			breaks = append(breaks, ctx.BytePos())
		}
	}

	// SB2: Break at end
	breaks = append(breaks, len(text))

	return breaks
}
