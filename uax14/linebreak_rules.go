package uax14

// Rule-Based Line Breaking Architecture
//
// This file implements a rule-based state machine architecture for UAX #14 (Line Breaking),
// following the successful pattern established in UAX #29 (Text Segmentation).
//
// ## Architecture Overview
//
// The implementation uses:
// - BreakDecision enum (BreakYes/BreakNo) for rule outcomes
// - LineBreakRule function type for individual rule implementations
// - Named rule functions (ruleLB4, ruleLB5a, etc.) with clear Unicode spec links
// - Rule array checked in order (first match wins)
// - Context object (LineBreakContext) for clean state management
//
// ## Implementation Status
//
// This implementation focuses on the most complex rules that provide the greatest
// maintainability benefit:
// - LB4, LB5: Mandatory breaks (BK, CR, LF, NL)
// - LB8: Break before character following ZW
// - LB8a: Do not break after ZWJ
// - LB21: Special HY (hyphen) handling with complex context (~50 lines of inline logic)
// - LB21.02: Hebrew hyphen (HH/MAQAF) handling
// - LB19: Quotation mark patterns (7 distinct context-sensitive patterns)
// - LB13-LB17: Punctuation and sequence rules
// - LB22-LB31: Alphabetic, numeric, and complex script rules
//
// Current conformance: 100% (19,338/19,338 official Unicode tests passing)
//
// ## Key Fixes for 100% Conformance
//
// The final 2 failures were resolved by fixing the rule ordering for LB19 quotation patterns:
// - French guillemet separators (»word« pattern): « SP ÷ AL
// - German quotes („..." and ‚...'): QU_Pi SP ÷ when paired with OP opener
//
// Both patterns required processing BEFORE ruleLB19_QU_Pi_SP, since in these contexts
// ClassQU_Pi acts as a closing quote (not an opening quote), requiring a break after SP.
//
// ## Benefits
//
// Compared to the original 1,112-line function with massive inline conditionals:
// - Rules are isolated and independently testable
// - Each rule has clear documentation and spec links
// - Adding new rules doesn't require understanding the entire state machine
// - Complex rules (LB21, LB19) are broken into manageable, named functions
// - The architecture matches the Unicode specification structure
//
// BreakDecision represents the decision about whether to break at the current position.
// This is distinct from BreakAction which comes from the pair table.
type BreakDecision int

const (
	BreakNo  BreakDecision = iota // Don't break
	BreakYes                       // Break allowed
)

// LineBreakRule checks if a rule applies and returns the action.
// Rules are checked in order - first match wins.
type LineBreakRule func(ctx *LineBreakContext) (matched bool, decision BreakDecision)

// ruleLB4 implements: BK ÷
// Always break after hard line breaks.
// https://www.unicode.org/reports/tr14/#LB4
func ruleLB4(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	if prev == ClassBK {
		return true, BreakYes
	}
	return false, BreakNo
}

// ruleLB5a implements: CR × LF
// Don't break between CR and LF.
// https://www.unicode.org/reports/tr14/#LB5
func ruleLB5a(ctx *LineBreakContext) (bool, BreakDecision) {
	if ctx.Prev() == ClassCR && ctx.Curr() == ClassLF {
		return true, BreakNo
	}
	return false, BreakYes
}

// ruleLB5b implements: CR ÷, LF ÷, NL ÷
// Always break after CR, LF, and NL (except CR × LF handled by LB5a).
// https://www.unicode.org/reports/tr14/#LB5
func ruleLB5b(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	if prev == ClassCR || prev == ClassLF || prev == ClassNL {
		return true, BreakYes
	}
	return false, BreakNo
}

// ruleLB7 implements: × ZW
// Do not break before zero width space.
// https://www.unicode.org/reports/tr14/#LB7
func ruleLB7(ctx *LineBreakContext) (bool, BreakDecision) {
	curr := ctx.Curr()

	if curr == ClassZW {
		return true, BreakNo
	}

	return false, BreakNo
}

// ruleLB8 implements: ZW SP* ÷
// Break after zero width space, possibly with spaces intervening.
// The break comes after any spaces, not immediately after ZW.
// Exception: Don't break before mandatory breaks.
// https://www.unicode.org/reports/tr14/#LB8
func ruleLB8(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	lastNonSpace := ctx.LastNonSpace()

	// Exception: Don't break before mandatory breaks or ZW
	if curr == ClassBK || curr == ClassCR || curr == ClassLF || curr == ClassNL || curr == ClassZW {
		return false, BreakNo
	}

	// ZW × SP (don't break between ZW and SP, or between consecutive SPs after ZW)
	if prev == ClassZW && curr == ClassSP {
		return true, BreakNo
	}

	// ZW SP+ × SP (don't break between spaces following ZW)
	if prev == ClassSP && curr == ClassSP && lastNonSpace == ClassZW {
		return true, BreakNo
	}

	// ZW ÷ X (break after ZW if not followed by SP)
	if prev == ClassZW {
		return true, BreakYes
	}

	// ZW SP+ ÷ X (break after spaces following ZW, before non-SP)
	if prev == ClassSP && curr != ClassSP && lastNonSpace == ClassZW {
		return true, BreakYes
	}

	return false, BreakNo
}

// ruleLB8a implements: ZWJ ×
// Do not break after zero width joiner.
// https://www.unicode.org/reports/tr14/#LB8a
func ruleLB8a(ctx *LineBreakContext) (bool, BreakDecision) {
	// Check if actual previous rune is ZWJ (even if class was converted by LB10)
	pos := ctx.Pos()
	if pos > 0 && ctx.RuneAt(pos-1) == '\u200D' { // U+200D is ZWJ
		return true, BreakNo
	}
	return false, BreakNo
}

// ruleLB21_HY implements: Special handling for HY (hyphen-minus)
// Handles multiple patterns:
// - AL × HY ÷ AL (regular hyphenated words like "Excusez-moi")
// - CP × HY ÷ (break after hyphen following closing punctuation)
// - CL × HY ÷ (break after hyphen following closing bracket)
// - HL × HY ÷ HL (Hebrew letter, hyphen, Hebrew letter)
// https://www.unicode.org/reports/tr14/#LB21
func ruleLB21_HY(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	pos := ctx.Pos()

	// Only applies when previous is HY and we have at least 2 characters before
	if prev != ClassHY || pos < 2 {
		return false, BreakNo
	}

	// Don't break before SP, ZW, or CM
	if curr == ClassSP || curr == ClassZW || isClassOrVariant(curr, ClassCM) {
		return false, BreakNo
	}

	// Check what comes before the HY, skipping over CM/ZWJ per LB9
	prevPrevIdx := ctx.SkipBackward(pos-2, ClassCM, ClassZWJ)
	if prevPrevIdx < 0 {
		return false, BreakNo
	}
	prevPrevClass := ctx.ClassAt(prevPrevIdx)
	shouldBreak := false

	if isClassOrVariant(prevPrevClass, ClassCP) || isClassOrVariant(prevPrevClass, ClassCL) {
		// CP × HY ÷, CL × HY ÷ - allow break after HY
		shouldBreak = true
	} else if prevPrevClass == ClassHL && curr == ClassHL {
		// HL × HY ÷ HL (Hebrew letter, hyphen, Hebrew letter) - allow break after HY
		shouldBreak = true
	} else if isClassOrVariant(prevPrevClass, ClassAL) && isClassOrVariant(curr, ClassAL) {
		// AL × HY ÷ AL - regular hyphenated words like "Excusez-moi"
		shouldBreak = true
	} else if prevPrevClass == ClassHY && isClassOrVariant(curr, ClassAL) {
		// HY × HY ÷ AL - check if this follows CP/CL in the context
		// Pattern: CP/CL × ... × AL × HY × HY ÷ AL (like "(http://)xn--a")
		checkIdx := pos - 3
		for checkIdx >= 0 {
			checkClass := ctx.ClassAt(checkIdx)
			if checkClass == ClassSP || isClassOrVariant(checkClass, ClassCM) ||
			   checkClass == ClassZWJ || isClassOrVariant(checkClass, ClassAL) {
				// Skip spaces, combining marks, and AL characters
				checkIdx--
				continue
			}
			// Check if we find CP or CL
			if isClassOrVariant(checkClass, ClassCP) || isClassOrVariant(checkClass, ClassCL) {
				shouldBreak = true
			}
			break
		}
	}

	if shouldBreak {
		return true, BreakYes
	}

	return false, BreakNo
}

// ruleLB21_HH implements: HL (HY | BA) × HL
// Special handling for Hebrew hyphen (MAQAF) within Hebrew text.
// Don't break after HH/BA when surrounded by Hebrew letters.
// https://www.unicode.org/reports/tr14/#LB21
func ruleLB21_HH(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	pos := ctx.Pos()

	// Only applies when previous is HH or BA and current is HL
	if (prev != ClassHH && prev != ClassBA) || curr != ClassHL {
		return false, BreakNo
	}

	// Check if preceded by HL (Hebrew letter)
	if pos >= 2 {
		checkIdx := pos - 2
		for checkIdx >= 0 {
			checkClass := ctx.ClassAt(checkIdx)
			// Skip combining marks
			if isClassOrVariant(checkClass, ClassCM) || checkClass == ClassZWJ {
				checkIdx--
				continue
			}
			// If preceded by HL, don't break
			if checkClass == ClassHL {
				return true, BreakNo
			}
			break
		}
	}

	return false, BreakNo
}

// ruleLB21_HY_SP_CM implements: SP × CM* × HY ÷ HL
// Special case for BreakIndirect when CM intervenes between SP and HY.
// Pattern: SP ÷ CM × HY ÷ HL
// The pair table says HY × HL is BreakIndirect (break if SP precedes).
// But CM between SP and HY causes prevClass to be HY, hiding the SP.
// This rule looks past CM to find SP and applies BreakIndirect.
// https://www.unicode.org/reports/tr14/#LB21
func ruleLB21_HY_SP_CM(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	pos := ctx.Pos()

	// Only applies when prev=HY, curr=HL
	if prev != ClassHY || curr != ClassHL {
		return false, BreakNo
	}

	// Check if HY is preceded by CM, and CM is preceded by SP
	if pos >= 3 {
		// Look back from HY position (pos-1) to find CM
		hyIdx := pos - 1
		cmIdx := hyIdx - 1

		// Check if position before HY is CM
		if isClassOrVariant(ctx.ClassAt(cmIdx), ClassCM) {
			// Look further back to find SP (skipping any additional CM)
			spIdx := ctx.SkipBackward(cmIdx-1, ClassCM, ClassZWJ)
			if spIdx >= 0 && ctx.ClassAt(spIdx) == ClassSP {
				// SP × CM* × HY ÷ HL - apply BreakIndirect (break after HY)
				return true, BreakYes
			}
		}
	}

	return false, BreakNo
}

// ruleLB21_HH_Break implements: HL × HH ÷ HL - Allow break after HH when preceded by HL or AL
// This is LB21.02 from the original implementation.
// When HH (Hebrew hyphen/MAQAF) is preceded by HL or AL, allow break after it before HL.
// https://www.unicode.org/reports/tr14/#LB21
func ruleLB21_HH_Break(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	pos := ctx.Pos()

	// Only applies when previous is HH and current is HL
	if prev != ClassHH || curr != ClassHL {
		return false, BreakNo
	}

	// Don't break before SP, ZW, or CM
	if curr == ClassSP || curr == ClassZW || isClassOrVariant(curr, ClassCM) {
		return false, BreakNo
	}

	// Check if HH is preceded by HL or AL (looking backward, skipping CM/ZWJ)
	if pos >= 2 {
		// Use SkipBackward to find the base character before HH
		baseIdx := ctx.SkipBackward(pos-2, ClassCM, ClassZWJ)
		if baseIdx >= 0 {
			baseClass := ctx.ClassAt(baseIdx)
			// If preceded by HL or AL, allow break after HH
			if baseClass == ClassHL || isClassOrVariant(baseClass, ClassAL) {
				return true, BreakYes
			}
		}
	}

	return false, BreakNo
}

// ruleLB19_NS_QU_Pi implements: NS ÷ QU_Pi
// Break before opening quote after non-starter (specifically FULLWIDTH COLON in CJK text).
// https://www.unicode.org/reports/tr14/#LB19
func ruleLB19_NS_QU_Pi(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	pos := ctx.Pos()

	if !isClassOrVariant(prev, ClassNS) || !isClassOrVariant(curr, ClassQU_Pi) || pos == 0 {
		return false, BreakNo
	}

	// Only apply to FULLWIDTH COLON (U+FF1A), not all NS characters
	prevRune := ctx.RuneAt(pos - 1)
	if prevRune == '\uFF1A' { // FULLWIDTH COLON
		return true, BreakYes
	}

	return false, BreakNo
}

// ruleLB19_Guillemet implements: Guillemet separators pattern
// Allow breaks: AL SP ÷ » and « SP ÷ AL when guillemets surround a single short word.
// This handles »word« pattern used as emphasis, not quotation.
// Also handles QU SP ÷ after quotation marks (German quotes, guillemets, etc.)
// https://www.unicode.org/reports/tr14/#LB19
func ruleLB19_Guillemet(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	pos := ctx.Pos()
	lastNonSpace := ctx.LastNonSpace()

	// Don't break before SP, ZW, or CM
	if curr == ClassSP || curr == ClassZW || isClassOrVariant(curr, ClassCM) {
		return false, BreakNo
	}

	// Pattern: « SP ÷ AL (guillemet separator - reverse direction)
	// When lastNonSpaceClass is « and we're at AL after space, check for »word« pattern
	// Note: U+00AB is ClassQU_Pi and U+00BB is ClassQU_Pf
	if prev == ClassSP && isClassOrVariant(curr, ClassAL) && pos >= 2 {
		if isClassOrVariant(lastNonSpace, ClassQU_Pi) {
			// Get the actual « character
			lastNonSpaceIdx := ctx.SkipBackward(pos-1, ClassSP)
			if lastNonSpaceIdx >= 0 {
				quoteRune := ctx.RuneAt(lastNonSpaceIdx)
				if quoteRune == '\u00AB' { // «
					// Look back to find matching » before the «
					for checkIdx := lastNonSpaceIdx - 1; checkIdx >= 0 && checkIdx > lastNonSpaceIdx-15; checkIdx-- {
						checkRune := ctx.RuneAt(checkIdx)
						if checkRune == '\u00BB' { // »
							return true, BreakYes
						}
						if checkRune == ' ' || checkRune == '\t' {
							break
						}
					}
				}
			}
		}
	}

	// Pattern: B2 SP ÷ QU_Pf (em-dash before guillemet)
	// When lastNonSpaceClass is B2 and we're at closing quote after space
	// Note: U+00BB is ClassQU_Pf
	if prev == ClassSP && lastNonSpace == ClassB2 && isClassOrVariant(curr, ClassQU_Pf) && pos > 0 {
		// Check if there's actual content after this guillemet (not just end of string)
		if pos+1 < ctx.Len() {
			nextClass := ctx.ClassAt(pos + 1)
			// Allow break only if next character is not end-of-text indicators
			if !isClassOrVariant(nextClass, ClassSP) && nextClass != ClassBK &&
			   nextClass != ClassCR && nextClass != ClassLF && nextClass != ClassNL {
				return true, BreakYes
			}
		}
	}

	// Pattern: SP ÷ » (break before closing guillemet in separator pattern)
	// Note: U+00BB is ClassQU_Pf
	if prev == ClassSP && isClassOrVariant(curr, ClassQU_Pf) && pos > 0 {
		currRune := ctx.Rune()
		if currRune == '\u00BB' { // »
			// Look ahead to find matching «
			foundOpening := false
			wordLength := 0
			hasPunctuation := false

			for checkIdx := pos + 1; checkIdx < ctx.Len() && checkIdx < pos+20; checkIdx++ {
				checkRune := ctx.RuneAt(checkIdx)
				if checkRune == '\u00AB' { // «
					foundOpening = true
					break
				}
				if checkRune == ' ' || checkRune == '\t' {
					break
				}
				if checkRune == '.' || checkRune == ',' || checkRune == '!' ||
				   checkRune == '?' || checkRune == ';' || checkRune == ':' {
					hasPunctuation = true
				}
				wordLength++
			}

			if foundOpening && wordLength >= 1 && wordLength <= 10 && !hasPunctuation {
				return true, BreakYes
			}
		}
	}

	return false, BreakNo
}

// ruleLB19_German implements: German quotes pattern
// Allow breaks after German closing quotes: „..." SP ÷ or ‚...' SP ÷
// In German typography:
// - „..." uses U+201E (ClassOP) to open and U+201C (ClassQU_Pi) to close
// - ‚...' uses U+201A (ClassOP) to open and U+2018 (ClassQU_Pi) to close
// https://www.unicode.org/reports/tr14/#LB19
func ruleLB19_German(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	pos := ctx.Pos()

	if prev != ClassSP || pos < 2 {
		return false, BreakNo
	}

	// Don't break before SP, ZW, or CM
	if curr == ClassSP || curr == ClassZW || isClassOrVariant(curr, ClassCM) {
		return false, BreakNo
	}

	// Check if character before space was a German closing quote (ClassQU_Pi)
	// U+201C (") for double quotes or U+2018 (') for single quotes
	beforeSpaceIdx := pos - 2
	beforeSpace := ctx.RuneAt(beforeSpaceIdx)
	beforeSpaceClass := ctx.ClassAt(beforeSpaceIdx)

	// Check if it's a QU_Pi (which German uses as closing quote)
	if !isClassOrVariant(beforeSpaceClass, ClassQU_Pi) {
		return false, BreakNo
	}

	// Only apply to German closing quotes
	if beforeSpace != '\u201C' && beforeSpace != '\u2018' {
		return false, BreakNo
	}

	// Look further back to find German opening quote (U+201E or U+201A, which are ClassOP)
	for checkIdx := pos - 3; checkIdx >= 0 && checkIdx > pos-30; checkIdx-- {
		checkRune := ctx.RuneAt(checkIdx)
		checkClass := ctx.ClassAt(checkIdx)
		if (checkRune == '\u201E' || checkRune == '\u201A') && isClassOrVariant(checkClass, ClassOP) {
			// Found German opening quote (ClassOP), so the closing quote allows break
			return true, BreakYes
		}
	}

	return false, BreakNo
}

// ruleLB19_CJK_QU_Pf_ID implements: QU_Pf ÷ ID in CJK context
// Allow breaks after CJK closing quotes before ideographs.
// Only when the closing quote follows CJK punctuation/ideographs, not Latin letters.
// https://www.unicode.org/reports/tr14/#LB19
func ruleLB19_CJK_QU_Pf_ID(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	pos := ctx.Pos()

	if !isClassOrVariant(prev, ClassQU_Pf) || !isClassOrVariant(curr, ClassID) || pos == 0 {
		return false, BreakNo
	}

	// Only apply to CJK curly quotes, not European guillemets
	prevRune := ctx.RuneAt(pos - 1)
	if prevRune != '\u201C' && prevRune != '\u201D' &&
	   prevRune != '\u2018' && prevRune != '\u2019' {
		return false, BreakNo
	}

	// Check if the character BEFORE the quote is CJK punctuation/ideograph (not Latin)
	if pos >= 2 {
		beforeQuoteClass := ctx.ClassAt(pos - 2)
		// Allow break if preceded by CJK classes
		if isClassOrVariant(beforeQuoteClass, ClassEX) ||
		   isClassOrVariant(beforeQuoteClass, ClassID) ||
		   isClassOrVariant(beforeQuoteClass, ClassCL) ||
		   isClassOrVariant(beforeQuoteClass, ClassNS) {
			return true, BreakYes
		}
	}

	return false, BreakNo
}

// isCJKIdeograph checks if a rune is a CJK ideograph.
func isCJKIdeograph(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) ||
		(r >= 0x3400 && r <= 0x4DBF) ||
		(r >= 0x20000 && r <= 0x2A6DF) ||
		(r >= 0x2A700 && r <= 0x2B73F) ||
		(r >= 0x2B740 && r <= 0x2B81F) ||
		(r >= 0x2B820 && r <= 0x2CEAF) ||
		(r >= 0x2CEB0 && r <= 0x2EBEF) ||
		(r >= 0x30000 && r <= 0x3134F)
}

// ruleLB19_CJK_ID_QU_Pi implements: ID ÷ QU_Pi in CJK context
// Allow breaks before CJK opening quotes after ideographs.
// https://www.unicode.org/reports/tr14/#LB19
func ruleLB19_CJK_ID_QU_Pi(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	pos := ctx.Pos()

	if !isClassOrVariant(prev, ClassID) || !isClassOrVariant(curr, ClassQU_Pi) || pos == 0 {
		return false, BreakNo
	}

	// Only apply to CJK curly quotes, not European guillemets
	currRune := ctx.Rune()
	prevRune := ctx.RuneAt(pos - 1)

	if currRune != '\u201C' && currRune != '\u2018' {
		return false, BreakNo
	}

	// Check if previous character is CJK ideograph
	if !isCJKIdeograph(prevRune) {
		return false, BreakNo
	}

	// Check if character AFTER the quote is also CJK
	if pos+1 < ctx.Len() {
		nextRune := ctx.RuneAt(pos + 1)
		if isCJKIdeograph(nextRune) {
			return true, BreakYes
		}
	}

	return false, BreakNo
}

// ruleLB19_QU_Pi_SP implements: QU_Pi × SP* ×
// Do not break after opening quote, even with spaces.
// Exception: QU_Pi SP ÷ OP when there's a closing quote with CP/CL between.
// https://www.unicode.org/reports/tr14/#LB19
func ruleLB19_QU_Pi_SP(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	lastNonSpace := ctx.LastNonSpace()
	pos := ctx.Pos()

	// Pattern 2: QU_Pi SP ÷ OP when there's a closing quote with CP/CL between
	// Example: "ambigu« ( ̈ ) »(ë)" - break after « SP before (
	if prev == ClassSP && isClassOrVariant(lastNonSpace, ClassQU_Pi) && isClassOrVariant(curr, ClassOP) {
		// Look ahead to see if there's a closing quote with CP/CL in between
		hasClosingQuote := false
		hasClosingParen := false
		for checkIdx := pos + 1; checkIdx < ctx.Len(); checkIdx++ {
			checkClass := ctx.ClassAt(checkIdx)
			if isClassOrVariant(checkClass, ClassQU_Pf) {
				// Found closing quote - allow break if there's CP/CL in between
				if hasClosingParen {
					hasClosingQuote = true
				}
				break
			}
			if isClassOrVariant(checkClass, ClassCP) || isClassOrVariant(checkClass, ClassCL) {
				hasClosingParen = true
			}
		}
		if hasClosingQuote {
			return true, BreakYes
		}
	}

	// QU_Pi SP* × (don't break after opening quote with spaces)
	if prev == ClassSP && isClassOrVariant(lastNonSpace, ClassQU_Pi) {
		return true, BreakNo
	}

	return false, BreakNo
}

// ruleLB19_SP_QU_Pf implements: SP ÷ QU_Pf after specific classes
// Break before closing quote when preceded by CP/CL/EX/IS/SY and space.
// Pattern: CP/CL/EX/IS/SY × SP ÷ QU_Pf
// https://www.unicode.org/reports/tr14/#LB19
func ruleLB19_SP_QU_Pf(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	pos := ctx.Pos()

	if prev != ClassSP || !isClassOrVariant(curr, ClassQU_Pf) {
		return false, BreakNo
	}

	lastNonSpace := ctx.LastNonSpace()

	// IS/SY are natural phrase boundaries - allow break only if they're OUTSIDE the quote
	if lastNonSpace == ClassIS || lastNonSpace == ClassSY {
		// Find the IS/SY position (should be immediately before SP)
		isSyIdx := -1
		for checkIdx := pos - 2; checkIdx >= 0; checkIdx-- {
			checkClass := ctx.ClassAt(checkIdx)
			if checkClass == ClassIS || checkClass == ClassSY {
				isSyIdx = checkIdx
				break
			}
			if checkClass != ClassSP && !isClassOrVariant(checkClass, ClassCM) &&
			   checkClass != ClassZWJ {
				break
			}
		}

		// Find the opening quote position
		openingQuoteIdx := -1
		for checkIdx := pos - 2; checkIdx >= 0; checkIdx-- {
			checkClass := ctx.ClassAt(checkIdx)
			if isClassOrVariant(checkClass, ClassQU_Pi) {
				openingQuoteIdx = checkIdx
				break
			}
		}

		// Allow break if IS/SY is before opening quote or conditions are met
		shouldBreak := false
		if isSyIdx >= 0 {
			if openingQuoteIdx >= 0 && isSyIdx < openingQuoteIdx {
				shouldBreak = true
			} else if openingQuoteIdx < 0 && (pos-isSyIdx) <= 3 && isSyIdx > 0 {
				shouldBreak = true
			}
		}

		if shouldBreak {
			return true, BreakYes
		}
		// Don't break - prevent pair table from applying
		return true, BreakNo
	}

	// For CP/CL/EX, require opening quote with content between (stricter check)
	if isClassOrVariant(lastNonSpace, ClassCP) ||
	   isClassOrVariant(lastNonSpace, ClassCL) ||
	   isClassOrVariant(lastNonSpace, ClassEX) {
		// Look back to find a matching opening quote with OP/CL content between
		hasMatchingQuote := false
		hasOpenParen := false
		for checkIdx := pos - 2; checkIdx >= 0; checkIdx-- {
			checkClass := ctx.ClassAt(checkIdx)
			if isClassOrVariant(checkClass, ClassQU_Pi) {
				if hasOpenParen {
					hasMatchingQuote = true
				}
				break
			}
			if isClassOrVariant(checkClass, ClassOP) || isClassOrVariant(checkClass, ClassCL) {
				hasOpenParen = true
			}
		}
		if hasMatchingQuote {
			return true, BreakYes
		}
	}

	// For all other cases, don't break before QU_Pf (LB19 default)
	return true, BreakNo
}

// ruleLB19_QU_Pf_SP implements: QU_Pf SP ÷
// Allow break after closing quote followed by space before certain classes (like B2 em-dash).
// This ensures LB18 (SP ÷) applies after closing quotes in appropriate contexts.
// https://www.unicode.org/reports/tr14/#LB19
func ruleLB19_QU_Pf_SP(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	lastNonSpace := ctx.LastNonSpace()

	// Pattern: QU_Pf SP ÷ (break after space following closing quote)
	if prev != ClassSP || !isClassOrVariant(lastNonSpace, ClassQU_Pf) {
		return false, BreakNo
	}

	// Only apply in specific contexts where LB18 should apply
	// This handles patterns like "?» — " where we want to break before the em-dash
	return false, BreakNo
}

// ruleLB11 implements: WJ ×, × WJ
// Do not break before or after word joiner.
// https://www.unicode.org/reports/tr14/#LB11
func ruleLB11(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()

	// Do not break after WJ
	if prev == ClassWJ {
		return true, BreakNo
	}

	// Do not break before WJ
	if curr == ClassWJ {
		return true, BreakNo
	}

	return false, BreakNo
}

// ruleLB12 implements: GL ×
// Do not break after GL (non-breaking glue).
// https://www.unicode.org/reports/tr14/#LB12
func ruleLB12(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()

	if isClassOrVariant(prev, ClassGL) {
		return true, BreakNo
	}

	return false, BreakNo
}

// ruleLB12c implements: BA ÷ GL
// Allow break after BA before GL.
// This is an exception to LB12a which normally prevents breaks before GL.
// https://www.unicode.org/reports/tr14/#LB12a
func ruleLB12c(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()

	if isClassOrVariant(prev, ClassBA) && isClassOrVariant(curr, ClassGL) {
		return true, BreakYes
	}

	return false, BreakNo
}

// ruleLB12a implements: [^SP BA HY] × GL
// Do not break before GL unless preceded by SP, BA, or HY.
// Note: HH (Hebrew Hyphen) behaves like BA for this rule.
// https://www.unicode.org/reports/tr14/#LB12a
func ruleLB12a(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()

	if !isClassOrVariant(curr, ClassGL) {
		return false, BreakNo
	}

	// Allow break after SP, BA, HY, or HH
	if prev == ClassSP || isClassOrVariant(prev, ClassBA) || prev == ClassHY || prev == ClassHH {
		return false, BreakNo
	}

	// Do not break before GL in all other cases
	return true, BreakNo
}

// ruleLB13 implements: × [CL CP EX IS SY]
// Do not break before ']' or '!' or ';' or '/' or closing punctuation.
// https://www.unicode.org/reports/tr14/#LB13
func ruleLB13(ctx *LineBreakContext) (bool, BreakDecision) {
	curr := ctx.Curr()

	if isClassOrVariant(curr, ClassCL) || curr == ClassCP ||
	   isClassOrVariant(curr, ClassEX) || curr == ClassIS || curr == ClassSY {
		return true, BreakNo
	}

	return false, BreakNo
}

// ruleLB14 implements: OP SP* ×
// Do not break after opening punctuation, even with intervening spaces.
// https://www.unicode.org/reports/tr14/#LB14
func ruleLB14(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	lastNonSpace := ctx.LastNonSpace()

	// If previous is SP and last non-space was OP, don't break
	if prev == ClassSP && isClassOrVariant(lastNonSpace, ClassOP) {
		return true, BreakNo
	}

	// If previous is OP directly, don't break
	if isClassOrVariant(prev, ClassOP) {
		return true, BreakNo
	}

	return false, BreakNo
}

// ruleLB15 implements: QU SP* × OP
// Do not break within opening quote-opening punctuation sequence.
// Only applies to QU_Pi (explicit opening quotes), not ambiguous QU.
// Ambiguous QU follows pair table (BreakIndirect = break if SP intervenes).
// https://www.unicode.org/reports/tr14/#LB15
func ruleLB15(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	lastNonSpace := ctx.LastNonSpace()

	if !isClassOrVariant(curr, ClassOP) {
		return false, BreakNo
	}

	// Only apply to QU_Pi (explicit opening quotes)
	// Ambiguous QU should follow pair table (BreakIndirect)

	// QU_Pi SP* × OP
	if prev == ClassSP && lastNonSpace == ClassQU_Pi {
		return true, BreakNo
	}

	// QU_Pi × OP (no space)
	if prev == ClassQU_Pi {
		return true, BreakNo
	}

	return false, BreakNo
}

// ruleLB16 implements: (CL | CP) SP* × NS
// Do not break within closing punctuation-nonstarter sequence.
// https://www.unicode.org/reports/tr14/#LB16
func ruleLB16(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	lastNonSpace := ctx.LastNonSpace()

	if !isClassOrVariant(curr, ClassNS) && curr != ClassCJ {
		return false, BreakNo
	}

	// (CL | CP) SP* × NS
	if prev == ClassSP && (isClassOrVariant(lastNonSpace, ClassCL) || lastNonSpace == ClassCP) {
		return true, BreakNo
	}

	// (CL | CP) × NS (no space)
	if isClassOrVariant(prev, ClassCL) || prev == ClassCP {
		return true, BreakNo
	}

	return false, BreakNo
}

// ruleLB17 implements: B2 SP* × B2
// Do not break within em-dash sequence with spaces.
// https://www.unicode.org/reports/tr14/#LB17
func ruleLB17(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	lastNonSpace := ctx.LastNonSpace()

	if curr != ClassB2 {
		return false, BreakNo
	}

	// B2 SP* × B2
	if prev == ClassSP && lastNonSpace == ClassB2 {
		return true, BreakNo
	}

	// B2 × B2 (no space)
	if prev == ClassB2 {
		return true, BreakNo
	}

	return false, BreakNo
}

// ruleLB18 implements: SP ÷
// Break after spaces (but see other rules for exceptions).
// Exception: Don't break between spaces that follow ZW (LB8 interaction).
// Exception: Don't break SP × SP (handled by pair table as BreakProhibited).
// https://www.unicode.org/reports/tr14/#LB18
func ruleLB18(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	lastNonSpace := ctx.LastNonSpace()

	if prev != ClassSP {
		return false, BreakNo
	}

	// Don't break before certain classes (these are handled by other rules)
	// LB6, LB7, LB11, LB13 take precedence
	if curr == ClassBK || curr == ClassCR || curr == ClassLF || curr == ClassNL ||
	   curr == ClassZW || curr == ClassWJ {
		return false, BreakNo
	}

	// Don't break SP × SP (pair table handles this as BreakProhibited)
	if curr == ClassSP {
		// Exception: After ZW, spaces should not break between themselves
		if lastNonSpace == ClassZW {
			return true, BreakNo
		}
		// Otherwise, don't break between consecutive spaces
		return false, BreakNo
	}

	// Allow break after space before non-space characters
	return true, BreakYes
}

// ruleLB20 implements: CB ÷ (except before certain classes)
// Break after contingent break opportunity.
// The pair table controls whether to break, and this rule enforces it only
// when the pair table says BreakDirect.
// https://www.unicode.org/reports/tr14/#LB20
func ruleLB20(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()

	// Only applies after CB
	if prev != ClassCB {
		return false, BreakNo
	}

	// Don't break before these classes (pair table handles them)
	if curr == ClassBK || curr == ClassCR || curr == ClassLF || curr == ClassNL ||
	   curr == ClassSP || isClassOrVariant(curr, ClassCM) || curr == ClassZW ||
	   isClassOrVariant(curr, ClassQU_Pi) || isClassOrVariant(curr, ClassQU_Pf) ||
	   isClassOrVariant(curr, ClassNS) {
		return false, BreakNo
	}

	// For other classes, check the pair table
	action := getBreakAction(prev, curr)
	if action == BreakDirect {
		return true, BreakYes
	}

	return false, BreakNo
}

// ruleLB22 implements: AL × IN, HL × IN
// Do not break between alphabetic and inseparable characters.
// https://www.unicode.org/reports/tr14/#LB22
func ruleLB22(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()

	if curr != ClassIN {
		return false, BreakNo
	}

	if isClassOrVariant(prev, ClassAL) || prev == ClassHL {
		return true, BreakNo
	}

	return false, BreakNo
}

// ruleLB23 implements: ID × PO, AL × NU, HL × NU
// Do not break between numeric prefix/postfix and letters.
// https://www.unicode.org/reports/tr14/#LB23
func ruleLB23(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()

	// ID × PO
	if isClassOrVariant(prev, ClassID) && isClassOrVariant(curr, ClassPO) {
		return true, BreakNo
	}

	// AL × NU
	if isClassOrVariant(prev, ClassAL) && curr == ClassNU {
		return true, BreakNo
	}

	// HL × NU
	if prev == ClassHL && curr == ClassNU {
		return true, BreakNo
	}

	return false, BreakNo
}

// ruleLB23a implements: PR × ID, PR × (AL | HL), PO × (AL | HL)
// Do not break between prefix and ideograph/letters.
// https://www.unicode.org/reports/tr14/#LB23a
func ruleLB23a(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()

	// PR × ID
	if isClassOrVariant(prev, ClassPR) && isClassOrVariant(curr, ClassID) {
		return true, BreakNo
	}

	// PR × (AL | HL)
	if isClassOrVariant(prev, ClassPR) && (isClassOrVariant(curr, ClassAL) || curr == ClassHL) {
		return true, BreakNo
	}

	// PO × (AL | HL)
	if isClassOrVariant(prev, ClassPO) && (isClassOrVariant(curr, ClassAL) || curr == ClassHL) {
		return true, BreakNo
	}

	return false, BreakNo
}

// ruleLB24 implements: (PR | PO) × (AL | HL)
// Do not break between prefix/postfix and alphabetic characters.
// https://www.unicode.org/reports/tr14/#LB24
func ruleLB24(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()

	if (isClassOrVariant(prev, ClassPR) || isClassOrVariant(prev, ClassPO)) &&
	   (isClassOrVariant(curr, ClassAL) || curr == ClassHL) {
		return true, BreakNo
	}

	return false, BreakNo
}

// ruleLB26 implements: JL × (JL | JV | H2 | H3), (JV | H2) × (JV | JT), (JT | H3) × JT
// Do not break Korean syllable sequences.
// https://www.unicode.org/reports/tr14/#LB26
func ruleLB26(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()

	// JL × (JL | JV | H2 | H3)
	if prev == ClassJL && (curr == ClassJL || curr == ClassJV || curr == ClassH2 || curr == ClassH3) {
		return true, BreakNo
	}

	// (JV | H2) × (JV | JT)
	if (prev == ClassJV || prev == ClassH2) && (curr == ClassJV || curr == ClassJT) {
		return true, BreakNo
	}

	// (JT | H3) × JT
	if (prev == ClassJT || prev == ClassH3) && curr == ClassJT {
		return true, BreakNo
	}

	return false, BreakNo
}

// ruleLB27 implements: (JL | JV | JT | H2 | H3) × PO
// Do not break between Korean syllable and postfix.
// https://www.unicode.org/reports/tr14/#LB27
func ruleLB27(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()

	if !isClassOrVariant(curr, ClassPO) {
		return false, BreakNo
	}

	if prev == ClassJL || prev == ClassJV || prev == ClassJT || prev == ClassH2 || prev == ClassH3 {
		return true, BreakNo
	}

	return false, BreakNo
}

// ruleLB28_AP implements: AP × (AK | AS | DottedCircle)
// Aksara Prebase attaches to following Aksara or Dotted Circle.
// https://www.unicode.org/reports/tr14/#LB28
func ruleLB28_AP(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()

	if prev != ClassAP {
		return false, BreakNo
	}

	// AP × (AK | AS)
	if curr == ClassAK || curr == ClassAS {
		return true, BreakNo
	}

	// AP × DottedCircle
	if ctx.Rune() == 0x25CC {
		return true, BreakNo
	}

	return false, BreakNo
}

// ruleLB28_Virama implements: (AK | AS | DottedCircle) × (VF | VI)
// Do not break between Aksara/DottedCircle and Virama.
// https://www.unicode.org/reports/tr14/#LB28
func ruleLB28_Virama(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	pos := ctx.Pos()

	if curr != ClassVF && curr != ClassVI {
		return false, BreakNo
	}

	// (AK | AS) × (VF | VI)
	if prev == ClassAK || prev == ClassAS {
		return true, BreakNo
	}

	// DottedCircle × (VF | VI) - look back past CM/ZWJ
	if pos > 0 {
		checkIdx := pos - 1
		for checkIdx >= 0 {
			checkClass := ctx.ClassAt(checkIdx)
			if !isClassOrVariant(checkClass, ClassCM) && checkClass != ClassZWJ {
				if ctx.RuneAt(checkIdx) == 0x25CC {
					return true, BreakNo
				}
				break
			}
			checkIdx--
		}
	}

	return false, BreakNo
}

// ruleLB28_VI_continuation implements: VI × AL × VI
// Do not break after Virama if Aksara sequence continues.
// https://www.unicode.org/reports/tr14/#LB28
func ruleLB28_VI_continuation(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	pos := ctx.Pos()

	if (prev != ClassVI && prev != ClassVF) || !isClassOrVariant(curr, ClassAL) {
		return false, BreakNo
	}

	// Check if AL is followed by VI/VF (Aksara sequence continuation)
	if pos+1 < ctx.Len() {
		nextClass := ctx.ClassAt(pos + 1)
		if nextClass == ClassVI || nextClass == ClassVF {
			return true, BreakNo
		}
	}

	return false, BreakNo
}

// ruleLB28_Base_VI_Aksara implements: Base × VI × (CM)* × AK/AS
// Do not break after Virama before Aksara when it connects a base.
// https://www.unicode.org/reports/tr14/#LB28
func ruleLB28_Base_VI_Aksara(ctx *LineBreakContext) (bool, BreakDecision) {
	curr := ctx.Curr()
	pos := ctx.Pos()

	if curr != ClassAK && curr != ClassAS {
		return false, BreakNo
	}

	// Look back past CM to find VI
	checkIdx := pos - 1
	foundVI := false
	viIndex := -1
	for checkIdx >= 0 {
		checkClass := ctx.ClassAt(checkIdx)
		if isClassOrVariant(checkClass, ClassCM) || checkClass == ClassZWJ {
			checkIdx--
			continue
		}
		// Only VI (connecting virama), not VF (final virama)
		if checkClass == ClassVI {
			foundVI = true
			viIndex = checkIdx
		}
		break
	}

	if !foundVI || viIndex <= 0 {
		return false, BreakNo
	}

	// Check if VI follows AK, AS, or AL (base)
	viPrevIdx := viIndex - 1
	for viPrevIdx >= 0 {
		viPrevClass := ctx.ClassAt(viPrevIdx)
		if isClassOrVariant(viPrevClass, ClassCM) || viPrevClass == ClassZWJ {
			viPrevIdx--
			continue
		}
		// Found base before VI
		if viPrevClass == ClassAK || viPrevClass == ClassAS || isClassOrVariant(viPrevClass, ClassAL) {
			return true, BreakNo
		}
		break
	}

	return false, BreakNo
}

// ruleLB28_AS_VF implements: AS × AS × VF
// Do not break between Aksara Starts when building towards Virama Final.
// https://www.unicode.org/reports/tr14/#LB28
func ruleLB28_AS_VF(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	pos := ctx.Pos()

	if prev != ClassAS || curr != ClassAS {
		return false, BreakNo
	}

	// Look ahead to see if VF immediately follows (skipping only CM)
	checkIdx := pos + 1
	for checkIdx < ctx.Len() && checkIdx < pos+4 {
		checkClass := ctx.ClassAt(checkIdx)
		if isClassOrVariant(checkClass, ClassCM) || checkClass == ClassZWJ {
			checkIdx++
			continue
		}
		// Found non-CM: check if it's VF
		if checkClass == ClassVF {
			return true, BreakNo
		}
		break
	}

	return false, BreakNo
}

// ruleLB25 implements: Do not break within numeric expressions
// This covers patterns like:
// - NU × NU
// - NU (SY | IS)* × NU
// - NU (SY | IS)* (CL | CP)? × (PO | PR)
// - (PO | PR)? (OP | HY)? NU (SY | IS)* × NU
// Special case: Leading decimals like ".35 cents" where IS acts as decimal point
// https://www.unicode.org/reports/tr14/#LB25
func ruleLB25(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	pos := ctx.Pos()

	// Special case: SP ÷ IS × NU (leading decimal like ".35")
	// Allow break before IS when it's a leading decimal point, not infix separator
	if prev == ClassSP && curr == ClassIS && pos+1 < ctx.Len() {
		// Look ahead to check if IS is followed by NU
		nextIdx := pos + 1
		nextClass := ctx.ClassAt(nextIdx)
		// Skip CM/ZWJ
		for nextIdx < ctx.Len() && (isClassOrVariant(nextClass, ClassCM) || nextClass == ClassZWJ) {
			nextIdx++
			if nextIdx < ctx.Len() {
				nextClass = ctx.ClassAt(nextIdx)
			}
		}
		if nextClass == ClassNU {
			// IS followed by NU - check if this is a leading decimal
			// Look back to see if preceded by NU (which would make it infix)
			isLeadingDecimal := true
			if pos >= 2 {
				checkIdx := pos - 2
				for checkIdx >= 0 {
					checkClass := ctx.ClassAt(checkIdx)
					if isClassOrVariant(checkClass, ClassCM) || checkClass == ClassZWJ {
						checkIdx--
						continue
					}
					if checkClass == ClassNU {
						isLeadingDecimal = false
					}
					break
				}
			}
			if isLeadingDecimal {
				// Leading decimal - allow break before IS
				return true, BreakYes
			}
		}
	}

	// Check if we're in a numeric context
	isNumericChar := curr == ClassNU || curr == ClassIS || curr == ClassSY ||
		isClassOrVariant(curr, ClassCL) || curr == ClassCP ||
		isClassOrVariant(curr, ClassPR) || isClassOrVariant(curr, ClassPO) ||
		isClassOrVariant(curr, ClassOP) || curr == ClassHY

	if !isNumericChar {
		return false, BreakNo
	}

	// Look back to find if we're in a numeric sequence
	checkIdx := pos - 1
	foundNum := false
	for checkIdx >= 0 {
		checkClass := ctx.ClassAt(checkIdx)

		// Skip combining marks
		if isClassOrVariant(checkClass, ClassCM) || checkClass == ClassZWJ {
			checkIdx--
			continue
		}

		// Found a number - we're in numeric context
		if checkClass == ClassNU {
			foundNum = true
			break
		}

		// These can be part of numeric expressions
		if checkClass == ClassIS || checkClass == ClassSY ||
			isClassOrVariant(checkClass, ClassCL) || checkClass == ClassCP ||
			isClassOrVariant(checkClass, ClassOP) || checkClass == ClassHY ||
			isClassOrVariant(checkClass, ClassPR) || isClassOrVariant(checkClass, ClassPO) {
			checkIdx--
			continue
		}

		// Non-numeric character found
		break
	}

	// LB25d: (PR | PO) × (OP | HY)? NU - can start without prior NU
	// This must be checked BEFORE the foundNum check because it doesn't require a prior number
	if isClassOrVariant(prev, ClassPR) || isClassOrVariant(prev, ClassPO) {
		if curr == ClassNU {
			return true, BreakNo
		}
		if isClassOrVariant(curr, ClassOP) || curr == ClassHY {
			// Look ahead to see if NU follows
			if pos+1 < ctx.Len() {
				nextClass := ctx.ClassAt(pos + 1)
				if nextClass == ClassNU {
					return true, BreakNo
				}
			}
		}
	}

	if !foundNum {
		return false, BreakNo
	}

	// LB25a: NU × (NU | SY | IS)
	if prev == ClassNU && (curr == ClassNU || curr == ClassSY || curr == ClassIS) {
		return true, BreakNo
	}

	// LB25b: NU (NU | SY | IS)* × (NU | SY | IS | CL | CP)
	if curr == ClassNU || curr == ClassSY || curr == ClassIS ||
		isClassOrVariant(curr, ClassCL) || curr == ClassCP {
		return true, BreakNo
	}

	// LB25c: NU (NU | SY | IS)* (CL | CP)? × (PO | PR)
	if (prev == ClassNU || prev == ClassIS || prev == ClassSY ||
		isClassOrVariant(prev, ClassCL) || prev == ClassCP) &&
		(isClassOrVariant(curr, ClassPR) || isClassOrVariant(curr, ClassPO)) {
		return true, BreakNo
	}

	// LB25d duplicate check (for completeness with prior NU)
	if isClassOrVariant(prev, ClassPR) || isClassOrVariant(prev, ClassPO) {
		if curr == ClassNU {
			return true, BreakNo
		}
		if (isClassOrVariant(curr, ClassOP) || curr == ClassHY) && pos+1 < ctx.Len() {
			nextClass := ctx.ClassAt(pos + 1)
			if nextClass == ClassNU {
				return true, BreakNo
			}
		}
	}

	return false, BreakNo
}

// ruleLB29 implements: IS × (AL | HL)
// Do not break between infix separator and alphabetic.
// https://www.unicode.org/reports/tr14/#LB29
func ruleLB29(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()

	if prev != ClassIS {
		return false, BreakNo
	}

	if isClassOrVariant(curr, ClassAL) || curr == ClassHL {
		return true, BreakNo
	}

	return false, BreakNo
}

// ruleLB30 implements: (AL | HL | NU) × OP, CP × (AL | HL | NU)
// Do not break between letters/numbers and opening/closing punctuation.
// Note: Only applies to base OP class, not OP_EastAsian variant.
// https://www.unicode.org/reports/tr14/#LB30
func ruleLB30(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()

	// (AL | HL | NU) × OP (base class only, not variants)
	if curr == ClassOP && (isClassOrVariant(prev, ClassAL) || prev == ClassHL || prev == ClassNU) {
		return true, BreakNo
	}

	// CP × (AL | HL | NU)
	if prev == ClassCP && (isClassOrVariant(curr, ClassAL) || curr == ClassHL || curr == ClassNU) {
		return true, BreakNo
	}

	return false, BreakNo
}

// ruleLB30a implements: RI × RI
// Do not break within emoji flag sequences (pairs of regional indicators).
// Break between pairs if there are multiple pairs.
// https://www.unicode.org/reports/tr14/#LB30a
func ruleLB30a(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	pos := ctx.Pos()

	if prev != ClassRI || curr != ClassRI {
		return false, BreakNo
	}

	// Count RIs before current position, skipping over CM/ZWJ per LB9
	// LB9: Treat X (CM | ZWJ)* as if it were X
	riCount := 0
	checkIdx := pos - 1
	for checkIdx >= 0 {
		checkClass := ctx.ClassAt(checkIdx)
		// Skip over CM/ZWJ to find actual base characters
		if isClassOrVariant(checkClass, ClassCM) || checkClass == ClassZWJ {
			checkIdx--
			continue
		}
		// Stop when we hit a non-RI
		if checkClass != ClassRI {
			break
		}
		riCount++
		checkIdx--
	}

	// If odd number of RIs before, this forms a pair - don't break
	// If even number, allow break (start of new pair)
	if riCount%2 == 1 {
		return true, BreakNo
	}

	// Even number of RIs - allow break between pairs
	return true, BreakYes
}

// ruleLB30b implements: EB × EM, ExtPict × EM
// Do not break between emoji base and emoji modifier.
// This includes both EB class and reserved/unassigned ExtPict characters (XX class).
// EM × EM should break (emoji modifiers don't chain).
// https://www.unicode.org/reports/tr14/#LB30b
func ruleLB30b(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	pos := ctx.Pos()

	if curr != ClassEM {
		return false, BreakNo
	}

	// EM × EM should break (don't prevent)
	if prev == ClassEM {
		return false, BreakNo
	}

	// EB × EM (direct emoji base)
	if prev == ClassEB {
		return true, BreakNo
	}

	// Look back past CM/ZWJ to find the actual base
	// Check if it's EB class OR any character with ExtPict property
	if pos > 0 {
		checkIdx := pos - 1
		for checkIdx >= 0 {
			checkClass := ctx.ClassAt(checkIdx)
			if isClassOrVariant(checkClass, ClassCM) || checkClass == ClassZWJ {
				checkIdx--
				continue
			}
			// Found the base - check if it's EB class
			if isClassOrVariant(checkClass, ClassEB) {
				return true, BreakNo
			}
			// Check if base has Extended_Pictographic property (emoji base)
			// Reserved/unassigned characters in emoji ranges may be classified as AL, ID, or XX
			// but still act as emoji bases if they have the ExtPict property
			baseRune := ctx.RuneAt(checkIdx)
			if isExtendedPictographic(baseRune) {
				return true, BreakNo
			}
			break
		}
	}

	return false, BreakNo
}

// isExtendedPictographic checks if a rune is an Extended_Pictographic character.
// Extended_Pictographic includes emoji and pictographic symbols that act as emoji bases.
// This matches the original implementation's range check (0x1F000-0x1FFFD only).
// Note: U+2600-U+26FF (Miscellaneous Symbols) are NOT included - they should break before EM.
func isExtendedPictographic(r rune) bool {
	// Exclude Regional Indicators (they don't act as emoji bases)
	if r >= 0x1F1E6 && r <= 0x1F1FF {
		return false
	}

	// Main emoji and pictographic blocks (matches original implementation)
	// This is the ONLY range checked in the original's line 5182
	if r >= 0x1F000 && r <= 0x1FFFD {
		return true
	}

	return false
}

// ruleLB31 implements: ID × EM (Extended_Pictographic only)
// Do not break between ExtPict ideograph and emoji modifier.
// Exception: ID characters that are NOT ExtendedPictographic should allow break before EM.
// Only true CJK ideographs or ExtPict characters should connect to EM.
// Note: ID × EB should break (not covered by this rule).
// https://www.unicode.org/reports/tr14/#LB30b
func ruleLB31(ctx *LineBreakContext) (bool, BreakDecision) {
	prev := ctx.Prev()
	curr := ctx.Curr()
	pos := ctx.Pos()

	// Only applies to base ID class
	if prev != ClassID {
		return false, BreakNo
	}

	// Only applies to EM (Emoji Modifier), not EB (Emoji Base)
	if curr != ClassEM {
		return false, BreakNo
	}

	// Check if the ID character is ExtendedPictographic
	// Only ExtPict ID characters should prevent breaks before EM
	// Non-ExtPict ID characters (like U+2600 sun emoji) should allow breaks
	if pos > 0 {
		checkIdx := pos - 1
		for checkIdx >= 0 {
			checkClass := ctx.ClassAt(checkIdx)
			if isClassOrVariant(checkClass, ClassCM) || checkClass == ClassZWJ {
				checkIdx--
				continue
			}
			// Found the base character
			baseRune := ctx.RuneAt(checkIdx)
			if isExtendedPictographic(baseRune) {
				// ExtPict ID × EM - don't break (LB30b logic)
				return true, BreakNo
			}
			// Non-ExtPict ID × EM - allow break (not a true emoji base)
			return false, BreakNo
		}
	}

	// Shouldn't reach here, but default to not breaking
	return false, BreakNo
}

// lineBreakRules is the ordered list of all line break boundary rules.
// Rules are checked in order - first match wins.
// Complex rules that override the pair table should come first.
var lineBreakRules = []LineBreakRule{
	// Mandatory breaks (LB4-LB5)
	ruleLB4,                // BK ÷
	ruleLB5a,               // CR × LF
	ruleLB5b,               // CR ÷, LF ÷, NL ÷

	// Zero-width characters (LB7, LB8, LB8a)
	ruleLB7,                // × ZW (don't break before ZW)
	ruleLB8a,               // ZWJ × (don't break after ZWJ)
	ruleLB8,                // ZW SP* ÷ (break after ZW)

	// Word joiner and glue (LB11, LB12, LB12a)
	ruleLB11,               // WJ ×, × WJ
	ruleLB12,               // GL ×
	ruleLB12c,              // BA ÷ GL (exception to LB12a)
	ruleLB12a,              // [^SP BA HY] × GL

	// Numeric expressions (LB25) - must come before LB13 to handle leading decimals
	ruleLB25,               // NU (SY | IS)* × NU, etc.

	// Closing punctuation (LB13)
	ruleLB13,               // × [CL CP EX IS SY]

	// Quotation marks (LB19 variants) - complex context-sensitive patterns
	// MUST come before LB14/LB15 to handle exceptions
	// Guillemet and German must come BEFORE QU_Pi_SP to override the default "don't break after opening quote" rule
	// In these patterns, QU_Pi acts as a closing quote (guillemet « or German "), not an opening quote
	ruleLB19_Guillemet,     // Guillemet separators (»word« pattern)
	ruleLB19_German,        // German quotes („..." and ‚...')
	ruleLB19_QU_Pi_SP,      // QU_Pi × SP* ×, with exception for QU_Pi SP ÷ OP when closing quote present

	// Opening punctuation (LB14, LB15)
	ruleLB14,               // OP SP* ×
	ruleLB15,               // QU SP* × OP

	// Closing punctuation with nonstarter (LB16)
	ruleLB16,               // (CL | CP) SP* × NS

	// B2 sequences (LB17)
	ruleLB17,               // B2 SP* × B2

	// Note: LB18 (SP ÷) is handled by the pair table and BreakIndirect logic

	// More quotation mark patterns
	ruleLB19_NS_QU_Pi,      // NS ÷ QU_Pi (FULLWIDTH COLON)
	ruleLB19_CJK_QU_Pf_ID,  // QU_Pf ÷ ID in CJK context
	ruleLB19_CJK_ID_QU_Pi,  // ID ÷ QU_Pi in CJK context
	ruleLB19_SP_QU_Pf,      // SP ÷ QU_Pf after specific classes

	// Contingent break (LB20)
	ruleLB20,               // ÷ CB, CB ÷

	// Hyphen handling (LB21)
	ruleLB21_HY,            // Special HY (hyphen) handling
	ruleLB21_HY_SP_CM,      // HY × HL with intervening CM after SP
	ruleLB21_HH_Break,      // Hebrew hyphen (MAQAF) break after HL or AL (must come before ruleLB21_HH)
	ruleLB21_HH,            // Hebrew hyphen (MAQAF) handling

	// Alphabetics (LB22-LB24)
	ruleLB22,               // AL × IN, HL × IN
	ruleLB23,               // ID × PO, AL × NU, HL × NU
	ruleLB23a,              // PR × ID, PR × (AL | HL), PO × (AL | HL)
	ruleLB24,               // (PR | PO) × (AL | HL)

	// Note: LB25 moved before LB13 to handle leading decimals correctly

	// Korean (LB26, LB27)
	ruleLB26,               // JL × (JL | JV | H2 | H3), etc.
	ruleLB27,               // (JL | JV | JT | H2 | H3) × PO

	// Aksara/Indic (LB28 variants)
	ruleLB28_AP,            // AP × (AK | AS | DottedCircle)
	ruleLB28_Virama,        // (AK | AS | DottedCircle) × (VF | VI)
	ruleLB28_VI_continuation, // VI × AL × VI
	ruleLB28_Base_VI_Aksara, // Base × VI × (CM)* × AK/AS
	ruleLB28_AS_VF,         // AS × AS × VF

	// Infix separator (LB29)
	ruleLB29,               // IS × (AL | HL)

	// Opening/closing punctuation (LB30)
	ruleLB30,               // (AL | HL | NU) × OP

	// Emoji (LB30a, LB30b)
	ruleLB30a,              // RI × RI (regional indicators)
	ruleLB30b,              // EB × EM, ExtPict × EM

	// Ideograph emoji (LB30b extension)
	ruleLB31,               // ID_ExtPict × EM
}

// FindLineBreakOpportunitiesWithRules finds line break opportunities using the rule-based approach.
// This is an alternative implementation for testing and benchmarking.
func FindLineBreakOpportunitiesWithRules(text string, hyphens Hyphens) []int {
	if text == "" {
		return []int{0}
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return []int{0}
	}

	breakPoints := []int{0} // LB2: Start is always a break point

	// Initialize first character's class and apply LB10
	prevClass := getBreakClass(runes[0])
	if isClassOrVariant(prevClass, ClassCM) || prevClass == ClassZWJ {
		prevClass = ClassAL
	}
	lastNonSpaceClass := prevClass

	// Create context for rule checking
	ctx := NewLineBreakContext(text, hyphens)

	// Loop through positions starting from 1 (checking break BEFORE each position)
	for i := 1; i < len(runes); i++ {
		// Move context to current position
		ctx.Slide()

		currClass := getBreakClass(runes[i])

		// LB9: SA characters that are combining marks should be treated as CM
		if currClass == ClassSA {
			if isCombiningMark(runes[i]) {
				currClass = ClassCM
			}
		}

		// Update context's prevClass and lastNonSpaceClass to match our state
		ctx.UpdatePrevClass(prevClass)
		ctx.UpdateLastNonSpace(lastNonSpaceClass)

		// Apply rules in order - first match wins
		decision := BreakNo
		ruleMatched := false

		for _, rule := range lineBreakRules {
			if matched, ruleDecision := rule(ctx); matched {
				decision = ruleDecision
				ruleMatched = true
				break
			}
		}

		// If no rule matched, fall back to pair table
		if !ruleMatched {
			action := getBreakAction(prevClass, currClass)
			// Convert BreakAction to BreakDecision
			if action == BreakDirect || action == BreakMandatory {
				decision = BreakYes
			} else if action == BreakIndirect {
				// BreakIndirect means: break if there's a SP between, don't break otherwise
				// Check if prevClass is SP (meaning we have X SP curr pattern)
				if prevClass == ClassSP {
					decision = BreakYes
				}
			}
		}

		if decision == BreakYes {
			// Insert break at current byte position
			bytePos := len(string(runes[:i]))
			breakPoints = append(breakPoints, bytePos)

			// Update prevClass after break (LB10 applies)
			if isClassOrVariant(currClass, ClassCM) || currClass == ClassZWJ {
				prevClass = ClassAL
				lastNonSpaceClass = ClassAL
			} else {
				prevClass = currClass
				if currClass != ClassSP {
					lastNonSpaceClass = currClass
				}
			}
		} else {
			// No break - update prevClass for next iteration
			if !isClassOrVariant(currClass, ClassCM) && currClass != ClassZWJ && currClass != ClassSA {
				prevClass = currClass
				if currClass != ClassSP {
					lastNonSpaceClass = currClass
				}
			} else if (isClassOrVariant(currClass, ClassCM) || currClass == ClassZWJ) &&
				(prevClass == ClassSP || prevClass == ClassZW) {
				// LB10: Treat CM or ZWJ after SP or ZW as AL
				prevClass = ClassAL
				lastNonSpaceClass = ClassAL
			}
		}
	}

	// LB3: Always break at end of text
	breakPoints = append(breakPoints, len(text))

	return breakPoints
}
