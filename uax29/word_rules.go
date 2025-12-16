package uax29

// WordBreakRule checks if a rule applies and returns the action.
type WordBreakRule func(ctx *WordBreakContext) (matched bool, action BreakAction)

// ruleWB3 implements: CR × LF
// Do not break within CRLF.
// https://www.unicode.org/reports/tr29/#WB3
func ruleWB3(ctx *WordBreakContext) (bool, BreakAction) {
	i := ctx.Pos()
	if i > 0 && ctx.ClassAt(i-1) == WBCR && ctx.Curr() == WBLF {
		return true, BreakNo
	}
	return false, BreakYes
}

// ruleWB3a implements: (Newline | CR | LF) ÷
// Break after newlines.
// https://www.unicode.org/reports/tr29/#WB3a
func ruleWB3a(ctx *WordBreakContext) (bool, BreakAction) {
	i := ctx.Pos()
	if i > 0 {
		prevClass := ctx.ClassAt(i - 1)
		if prevClass == WBCR || prevClass == WBLF || prevClass == WBNewline {
			return true, BreakYes
		}
	}
	return false, BreakNo
}

// ruleWB3b implements: ÷ (Newline | CR | LF)
// Break before newlines.
// https://www.unicode.org/reports/tr29/#WB3b
func ruleWB3b(ctx *WordBreakContext) (bool, BreakAction) {
	curr := ctx.Curr()
	if curr == WBCR || curr == WBLF || curr == WBNewline {
		return true, BreakYes
	}
	return false, BreakNo
}

// ruleWB3c implements: ZWJ × ExtendedPictographic
// Do not break within emoji ZWJ sequences.
// https://www.unicode.org/reports/tr29/#WB3c
func ruleWB3c(ctx *WordBreakContext) (bool, BreakAction) {
	i := ctx.Pos()
	if i > 0 && ctx.ClassAt(i-1) == WBZWJ && isExtendedPictographic(ctx.Rune()) {
		return true, BreakNo
	}
	return false, BreakYes
}

// ruleWB3d implements: WSegSpace × WSegSpace
// Keep horizontal whitespace together.
// https://www.unicode.org/reports/tr29/#WB3d
func ruleWB3d(ctx *WordBreakContext) (bool, BreakAction) {
	i := ctx.Pos()
	if i > 0 && ctx.ClassAt(i-1) == WBWSegSpace && ctx.Curr() == WBWSegSpace {
		return true, BreakNo
	}
	return false, BreakYes
}

// ruleWB4 implements: × (Format | Extend | ZWJ)
// Ignore Format and Extend characters.
// https://www.unicode.org/reports/tr29/#WB4
func ruleWB4(ctx *WordBreakContext) (bool, BreakAction) {
	curr := ctx.Curr()
	if curr == WBFormat || curr == WBExtend || curr == WBZWJ {
		return true, BreakNo
	}
	return false, BreakYes
}

// ruleWB5 implements: AHLetter × AHLetter
// Do not break between most letters.
// https://www.unicode.org/reports/tr29/#WB5
func ruleWB5(ctx *WordBreakContext) (bool, BreakAction) {
	prev, _ := ctx.PrevNonIgnorable()
	curr := ctx.Curr()
	if (prev == WBALetter || prev == WBHebrewLetter) && (curr == WBALetter || curr == WBHebrewLetter) {
		return true, BreakNo
	}
	return false, BreakYes
}

// ruleWB6 implements: AHLetter × (MidLetter | MidNumLet | Single_Quote) AHLetter
// https://www.unicode.org/reports/tr29/#WB6
func ruleWB6(ctx *WordBreakContext) (bool, BreakAction) {
	prev, _ := ctx.PrevNonIgnorable()
	curr := ctx.Curr()

	if (prev == WBALetter || prev == WBHebrewLetter) &&
	   (curr == WBMidLetter || curr == WBMidNumLet || curr == WBSingleQuote) {
		// Look ahead for AHLetter
		next, _ := ctx.NextNonIgnorable()
		if next == WBALetter || next == WBHebrewLetter {
			return true, BreakNo
		}
	}
	return false, BreakYes
}

// ruleWB7 implements: AHLetter (MidLetter | MidNumLet | Single_Quote) × AHLetter
// https://www.unicode.org/reports/tr29/#WB7
func ruleWB7(ctx *WordBreakContext) (bool, BreakAction) {
	prev, prevPos := ctx.PrevNonIgnorable()
	curr := ctx.Curr()

	if (prev == WBMidLetter || prev == WBMidNumLet || prev == WBSingleQuote) &&
	   (curr == WBALetter || curr == WBHebrewLetter) {
		// Look back past the MidLetter for AHLetter
		if prevPos > 0 {
			prevPrevPos := prevPos - 1
			for prevPrevPos > 0 && (ctx.ClassAt(prevPrevPos) == WBFormat ||
			                        ctx.ClassAt(prevPrevPos) == WBExtend ||
			                        ctx.ClassAt(prevPrevPos) == WBZWJ) {
				prevPrevPos--
			}
			if prevPrevPos >= 0 {
				prevPrev := ctx.ClassAt(prevPrevPos)
				if prevPrev == WBALetter || prevPrev == WBHebrewLetter {
					return true, BreakNo
				}
			}
		}
	}
	return false, BreakYes
}

// ruleWB7a implements: Hebrew_Letter × Single_Quote
// https://www.unicode.org/reports/tr29/#WB7a
func ruleWB7a(ctx *WordBreakContext) (bool, BreakAction) {
	prev, _ := ctx.PrevNonIgnorable()
	curr := ctx.Curr()
	if prev == WBHebrewLetter && curr == WBSingleQuote {
		return true, BreakNo
	}
	return false, BreakYes
}

// ruleWB7b implements: Hebrew_Letter × Double_Quote Hebrew_Letter
// https://www.unicode.org/reports/tr29/#WB7b
func ruleWB7b(ctx *WordBreakContext) (bool, BreakAction) {
	prev, _ := ctx.PrevNonIgnorable()
	curr := ctx.Curr()

	if prev == WBHebrewLetter && curr == WBDoubleQuote {
		next, _ := ctx.NextNonIgnorable()
		if next == WBHebrewLetter {
			return true, BreakNo
		}
	}
	return false, BreakYes
}

// ruleWB7c implements: Hebrew_Letter Double_Quote × Hebrew_Letter
// https://www.unicode.org/reports/tr29/#WB7c
func ruleWB7c(ctx *WordBreakContext) (bool, BreakAction) {
	prev, prevPos := ctx.PrevNonIgnorable()
	curr := ctx.Curr()

	if prev == WBDoubleQuote && curr == WBHebrewLetter {
		// Look back past the DoubleQuote for HebrewLetter
		if prevPos > 0 {
			prevPrevPos := prevPos - 1
			for prevPrevPos > 0 && (ctx.ClassAt(prevPrevPos) == WBFormat ||
			                        ctx.ClassAt(prevPrevPos) == WBExtend ||
			                        ctx.ClassAt(prevPrevPos) == WBZWJ) {
				prevPrevPos--
			}
			if prevPrevPos >= 0 && ctx.ClassAt(prevPrevPos) == WBHebrewLetter {
				return true, BreakNo
			}
		}
	}
	return false, BreakYes
}

// ruleWB8 implements: Numeric × Numeric
// Do not break within sequences of digits.
// https://www.unicode.org/reports/tr29/#WB8
func ruleWB8(ctx *WordBreakContext) (bool, BreakAction) {
	prev, _ := ctx.PrevNonIgnorable()
	curr := ctx.Curr()
	if prev == WBNumeric && curr == WBNumeric {
		return true, BreakNo
	}
	return false, BreakYes
}

// ruleWB9 implements: AHLetter × Numeric
// https://www.unicode.org/reports/tr29/#WB9
func ruleWB9(ctx *WordBreakContext) (bool, BreakAction) {
	prev, _ := ctx.PrevNonIgnorable()
	curr := ctx.Curr()
	if (prev == WBALetter || prev == WBHebrewLetter) && curr == WBNumeric {
		return true, BreakNo
	}
	return false, BreakYes
}

// ruleWB10 implements: Numeric × AHLetter
// https://www.unicode.org/reports/tr29/#WB10
func ruleWB10(ctx *WordBreakContext) (bool, BreakAction) {
	prev, _ := ctx.PrevNonIgnorable()
	curr := ctx.Curr()
	if prev == WBNumeric && (curr == WBALetter || curr == WBHebrewLetter) {
		return true, BreakNo
	}
	return false, BreakYes
}

// ruleWB11 implements: Numeric (MidNum | MidNumLet | Single_Quote) × Numeric
// https://www.unicode.org/reports/tr29/#WB11
func ruleWB11(ctx *WordBreakContext) (bool, BreakAction) {
	prev, prevPos := ctx.PrevNonIgnorable()
	curr := ctx.Curr()

	if (prev == WBMidNum || prev == WBMidNumLet || prev == WBSingleQuote) && curr == WBNumeric {
		// Look back past the MidNum for Numeric
		if prevPos > 0 {
			prevPrevPos := prevPos - 1
			for prevPrevPos > 0 && (ctx.ClassAt(prevPrevPos) == WBFormat ||
			                        ctx.ClassAt(prevPrevPos) == WBExtend ||
			                        ctx.ClassAt(prevPrevPos) == WBZWJ) {
				prevPrevPos--
			}
			if prevPrevPos >= 0 && ctx.ClassAt(prevPrevPos) == WBNumeric {
				return true, BreakNo
			}
		}
	}
	return false, BreakYes
}

// ruleWB12 implements: Numeric × (MidNum | MidNumLet | Single_Quote) Numeric
// https://www.unicode.org/reports/tr29/#WB12
func ruleWB12(ctx *WordBreakContext) (bool, BreakAction) {
	prev, _ := ctx.PrevNonIgnorable()
	curr := ctx.Curr()

	if prev == WBNumeric && (curr == WBMidNum || curr == WBMidNumLet || curr == WBSingleQuote) {
		// Look ahead for Numeric
		next, _ := ctx.NextNonIgnorable()
		if next == WBNumeric {
			return true, BreakNo
		}
	}
	return false, BreakYes
}

// ruleWB13 implements: Katakana × Katakana
// Do not break between Katakana.
// https://www.unicode.org/reports/tr29/#WB13
func ruleWB13(ctx *WordBreakContext) (bool, BreakAction) {
	prev, _ := ctx.PrevNonIgnorable()
	curr := ctx.Curr()
	if prev == WBKatakana && curr == WBKatakana {
		return true, BreakNo
	}
	return false, BreakYes
}

// ruleWB13a implements: (AHLetter | Numeric | Katakana | ExtendNumLet) × ExtendNumLet
// https://www.unicode.org/reports/tr29/#WB13a
func ruleWB13a(ctx *WordBreakContext) (bool, BreakAction) {
	prev, _ := ctx.PrevNonIgnorable()
	curr := ctx.Curr()
	if (prev == WBALetter || prev == WBHebrewLetter || prev == WBNumeric ||
	    prev == WBKatakana || prev == WBExtendNumLet) && curr == WBExtendNumLet {
		return true, BreakNo
	}
	return false, BreakYes
}

// ruleWB13b implements: ExtendNumLet × (AHLetter | Numeric | Katakana)
// https://www.unicode.org/reports/tr29/#WB13b
func ruleWB13b(ctx *WordBreakContext) (bool, BreakAction) {
	prev, _ := ctx.PrevNonIgnorable()
	curr := ctx.Curr()
	if prev == WBExtendNumLet && (curr == WBALetter || curr == WBHebrewLetter ||
	                               curr == WBNumeric || curr == WBKatakana) {
		return true, BreakNo
	}
	return false, BreakYes
}

// ruleWB15_16 implements: Regional_Indicator × Regional_Indicator (pairs)
// Do not break within emoji flag sequences (pairs of regional indicators).
// https://www.unicode.org/reports/tr29/#WB15
// https://www.unicode.org/reports/tr29/#WB16
func ruleWB15_16(ctx *WordBreakContext) (bool, BreakAction) {
	prev, prevPos := ctx.PrevNonIgnorable()
	curr := ctx.Curr()

	if prev == WBRegionalIndicator && curr == WBRegionalIndicator {
		// Count consecutive RIs before current position
		count := 0
		j := prevPos
		for j >= 0 {
			if ctx.ClassAt(j) == WBRegionalIndicator {
				count++
				j--
			} else if ctx.ClassAt(j) == WBFormat || ctx.ClassAt(j) == WBExtend || ctx.ClassAt(j) == WBZWJ {
				j--
			} else {
				break
			}
		}
		// If odd number, pair with previous RI (don't break)
		if count%2 == 1 {
			return true, BreakNo
		}
	}
	return false, BreakYes
}

// wordBreakRules is the ordered list of all word boundary rules.
// Rules are checked in order - first match wins.
var wordBreakRules = []WordBreakRule{
	ruleWB3,    // CR × LF
	ruleWB3a,   // (Newline | CR | LF) ÷
	ruleWB3b,   // ÷ (Newline | CR | LF)
	ruleWB3c,   // ZWJ × ExtPict
	ruleWB3d,   // WSegSpace × WSegSpace
	ruleWB4,    // × (Format | Extend | ZWJ)
	ruleWB5,    // AHLetter × AHLetter
	ruleWB6,    // AHLetter × (MidLetter | MidNumLet | Single_Quote) AHLetter
	ruleWB7,    // AHLetter (MidLetter | MidNumLet | Single_Quote) × AHLetter
	ruleWB7a,   // Hebrew_Letter × Single_Quote
	ruleWB7b,   // Hebrew_Letter × Double_Quote Hebrew_Letter
	ruleWB7c,   // Hebrew_Letter Double_Quote × Hebrew_Letter
	ruleWB8,    // Numeric × Numeric
	ruleWB9,    // AHLetter × Numeric
	ruleWB10,   // Numeric × AHLetter
	ruleWB11,   // Numeric (MidNum | MidNumLet | Single_Quote) × Numeric
	ruleWB12,   // Numeric × (MidNum | MidNumLet | Single_Quote) Numeric
	ruleWB13,   // Katakana × Katakana
	ruleWB13a,  // (AHLetter | Numeric | Katakana | ExtendNumLet) × ExtendNumLet
	ruleWB13b,  // ExtendNumLet × (AHLetter | Numeric | Katakana)
	ruleWB15_16, // RI × RI (pairs)
	// WB999: Otherwise, break everywhere (default in main loop)
}

// findWordBreaksFromClassesWithRules finds word boundaries using the rule-based
// approach with pre-classified data. This is used by the single-pass API.
func findWordBreaksFromClassesWithRules(text string, runes []rune, classes []PackedBreakClass, graphemeBreaks []int) []int {
	if len(runes) == 0 {
		return []int{}
	}

	ctx := NewWordBreakContextFromClasses(text, runes, classes, graphemeBreaks)
	breaks := []int{0} // WB1: Break at start

	for ctx.Slide() {
		// Apply rules in order - first match wins
		action := BreakYes // WB999: Default is break

		for _, rule := range wordBreakRules {
			if matched, ruleAction := rule(ctx); matched {
				action = ruleAction
				break // Stop at first matching rule
			}
		}

		if action == BreakYes {
			breaks = append(breaks, ctx.BytePos())
		}
	}

	// WB2: Break at end
	breaks = append(breaks, len(text))

	return breaks
}
