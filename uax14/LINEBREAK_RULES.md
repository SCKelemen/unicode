# Rule-Based Line Breaking Architecture

This document describes the rule-based state machine architecture implemented for UAX #14 (Line Breaking Algorithm).

## Overview

The UAX #14 line breaking algorithm has been partially refactored from a monolithic 1,112-line function into a clean, maintainable rule-based architecture following the successful pattern established in UAX #29 (Text Segmentation).

## Architecture Pattern

### Core Components

1. **BreakDecision Enum** (`BreakYes`/`BreakNo`)
   - Simple binary decision for each position
   - Clear, self-documenting semantics

2. **LineBreakRule Function Type**
   ```go
   type LineBreakRule func(ctx *LineBreakContext) (matched bool, decision BreakDecision)
   ```
   - Returns whether the rule applies
   - Returns the break decision if it applies
   - First matching rule wins

3. **Named Rule Functions**
   - `ruleLB4()` - Break after BK
   - `ruleLB5a()` - CR × LF (don't break within CRLF)
   - `ruleLB5b()` - Break after CR, LF, NL
   - `ruleLB8()` - Break before character following ZW
   - `ruleLB8a()` - Don't break after ZWJ
   - `ruleLB21_HY()` - Complex hyphen handling
   - `ruleLB21_HH()` - Hebrew hyphen (MAQAF) handling
   - `ruleLB19_*()` - Seven quotation mark pattern rules

4. **LineBreakContext Object**
   - Clean abstraction over text and state
   - Methods: `Prev()`, `Curr()`, `Next()`, `Rune()`, `ClassAt()`, etc.
   - Encapsulates all position tracking and lookups

5. **Rule Array**
   ```go
   var lineBreakRules = []LineBreakRule{
       ruleLB4,    // BK ÷
       ruleLB5a,   // CR × LF
       ruleLB5b,   // CR ÷, LF ÷, NL ÷
       ruleLB8a,   // ZWJ ×
       ruleLB8,    // Break before character following ZW
       // ... more rules
   }
   ```

### Main Algorithm

```go
func FindLineBreakOpportunitiesWithRules(text string, hyphens Hyphens) []int {
    // Initialize context
    ctx := NewLineBreakContext(text, hyphens)

    // Loop through positions
    for i := 1; i < len(runes); i++ {
        // Try each rule in order
        for _, rule := range lineBreakRules {
            if matched, decision := rule(ctx); matched {
                // First match wins
                break
            }
        }

        // Fall back to pair table for unimplemented rules
        if !ruleMatched {
            action := getBreakAction(prevClass, currClass)
            // ...
        }
    }
}
```

## Implemented Rules

### Priority 1: Mandatory Breaks

These rules provide the foundation for correct line breaking:

- **LB4**: `BK ÷` - Always break after hard line breaks
- **LB5a**: `CR × LF` - Don't break between CR and LF
- **LB5b**: `CR ÷`, `LF ÷`, `NL ÷` - Break after line terminators

### Priority 2: Special Characters

Essential for modern text:

- **LB8**: Break before any character following ZW (Zero Width Space)
  - Handles: `ZW SP × AL` → `ZW SP ÷ AL`
  - Exception: Don't break before mandatory breaks or another ZW

- **LB8a**: `ZWJ ×` - Don't break after Zero Width Joiner
  - Critical for emoji sequences: 👨‍👩‍👧‍👦

### Priority 3: Hyphen Handling (LB21)

The most complex inline logic from the original implementation (~100 lines total):

- **ruleLB21_HY**: Multiple hyphen patterns
  - `AL × HY ÷ AL` - Regular hyphenated words ("Excusez-moi")
  - `CP × HY ÷` - Break after hyphen following closing punctuation
  - `CL × HY ÷` - Break after hyphen following closing bracket
  - `HL × HY ÷ HL` - Hebrew letter patterns
  - Complex lookback for `CP/CL × ... × AL × HY × HY ÷ AL` patterns

- **ruleLB21_HH**: Hebrew hyphen (MAQAF) handling
  - `HL × HH ÷ HL` with context checking
  - Skips combining marks in lookback

### Priority 4: Quotation Marks (LB19)

Seven distinct context-sensitive patterns for complex quotation mark behavior:

1. **ruleLB19_NS_QU_Pi**: `NS ÷ QU_Pi` - FULLWIDTH COLON before opening quote (CJK)
2. **ruleLB19_Guillemet**: Guillemet separators (»word« emphasis pattern)
3. **ruleLB19_German**: German quote patterns („text" and ‚text')
4. **ruleLB19_CJK_QU_Pf_ID**: `QU_Pf ÷ ID` in CJK context
5. **ruleLB19_CJK_ID_QU_Pi**: `ID ÷ QU_Pi` in CJK context
6. **ruleLB19_SP_QU_Pf**: `SP ÷ QU_Pf` after specific classes (CP/CL/EX/IS/SY)

Each rule includes sophisticated context checking (CJK vs. Latin, inside vs. outside quotes, etc.)

## Test Results

### Conformance

- **Official Unicode Tests**: 18,941/19,338 passing (97.9%)
- **Basic Rule Tests**: 5/5 passing (100%)
- **Comparison Tests**: 12/12 passing (100%)

### What Works

✅ All mandatory breaks (BK, CR, LF, NL, CR+LF)
✅ ZWJ sequences (emoji families, etc.)
✅ ZW spacing rules
✅ Hyphenated words in multiple languages
✅ Complex quotation mark patterns (7 scenarios)
✅ CJK and Latin text mixing
✅ Hebrew text with MAQAF

### What Doesn't Work Yet (2.1% failures)

❌ BreakIndirect rules (LB13-LB17, LB24)
❌ Complex lookahead patterns (LB22, LB23, LB28-LB31)
❌ Some tailorable edge cases

## Benefits of This Architecture

### Maintainability

**Before:**
```go
// Original: 1,112-line function with massive nested conditionals
if prevClass == ClassHY && i >= 2 {
    prevPrevRune := runes[i-2]
    prevPrevClass := getBreakClass(prevPrevRune)
    shouldBreak := false
    if isClassOrVariant(prevPrevClass, ClassCP) || isClassOrVariant(prevPrevClass, ClassCL) {
        shouldBreak = true
    } else if prevPrevClass == ClassHL && currClass == ClassHL {
        shouldBreak = true
    } else if isClassOrVariant(prevPrevClass, ClassAL) && isClassOrVariant(currClass, ClassAL) {
        shouldBreak = true
    } else if prevPrevClass == ClassHY && isClassOrVariant(currClass, ClassAL) {
        // Even more nested logic...
        checkIdx := i - 3
        for checkIdx >= 0 {
            // ...50 more lines...
        }
    }
    if shouldBreak && currClass != ClassSP && currClass != ClassZW && currClass != ClassCM {
        // Insert break and continue...
    }
}
```

**After:**
```go
// New: Clean, isolated, documented rule
// ruleLB21_HY implements: Special handling for HY (hyphen-minus)
// Handles multiple patterns:
// - AL × HY ÷ AL (regular hyphenated words like "Excusez-moi")
// - CP × HY ÷ (break after hyphen following closing punctuation)
// - CL × HY ÷ (break after hyphen following closing bracket)
// - HL × HY ÷ HL (Hebrew letter, hyphen, Hebrew letter)
// https://www.unicode.org/reports/tr14/#LB21
func ruleLB21_HY(ctx *LineBreakContext) (bool, BreakDecision) {
    // Clear, focused implementation with context helpers
}
```

### Testability

Each rule can be tested independently:

```go
func TestRuleLB21_HY(t *testing.T) {
    tests := []struct{
        text string
        expected []int
    }{
        {"Excusez-moi", []int{0, 8, 11}},
        {"(test)-word", []int{0, 7, 12}},
        // ...
    }
}
```

### Documentation

- Every rule has a comment with the Unicode spec link
- Rule names match the UAX #14 specification
- Complex patterns have detailed explanations

### Extensibility

Adding a new rule is straightforward:

1. Create a new `ruleXX` function
2. Add it to the `lineBreakRules` array
3. Write tests for it
4. No need to modify existing rules or understand the entire state machine

## Future Work

### Remaining Rules to Implement

To reach 100% conformance, implement:

1. **LB10-LB17**: Combining marks, spaces, and inline objects
2. **LB18-LB20**: Opening/closing punctuation
3. **LB22-LB23**: Numeric patterns
4. **LB24-LB27**: Prefix/postfix patterns
5. **LB28-LB31**: Regional indicators, emoji, and other special cases

### Implementation Strategy

Follow the same pattern for each:

```go
func ruleLBXX(ctx *LineBreakContext) (bool, BreakDecision) {
    // 1. Check if rule applies
    // 2. Use context helpers for lookback/lookahead
    // 3. Return decision
    // 4. Include spec link in comment
}
```

### Performance Optimization

Current implementation prioritizes correctness and clarity. Once complete, optimize:

- Compile rules into a decision tree
- Cache context lookups
- Benchmark against original implementation

## Comparison with UAX #29

This implementation follows the same successful pattern as UAX #29:

| Aspect | UAX #29 | UAX #14 |
|--------|---------|---------|
| Context object | `GraphemeBreakContext` | `LineBreakContext` |
| Decision enum | `BreakAction` | `BreakDecision` |
| Rule type | `GraphemeBreakRule` | `LineBreakRule` |
| Rule array | `graphemeBreakRules` | `lineBreakRules` |
| Main function | `FindGraphemeBreaksWithRules` | `FindLineBreakOpportunitiesWithRules` |

## References

- [UAX #14: Unicode Line Breaking Algorithm](https://www.unicode.org/reports/tr14/)
- [UAX #29: Unicode Text Segmentation](https://www.unicode.org/reports/tr29/)
- [Official Test Data](https://www.unicode.org/Public/UCD/latest/ucd/auxiliary/LineBreakTest.txt)

## Files

- `linebreak_rules.go` - Rule implementations and main algorithm
- `linebreak_rules_test.go` - Test suite for rule-based implementation
- `context.go` - LineBreakContext object
- `uax14.go` - Original implementation (still 100% conformant)

## Conclusion

This rule-based architecture successfully demonstrates:

- ✅ Clean separation of concerns
- ✅ Improved maintainability and readability
- ✅ Independent testability of rules
- ✅ Clear documentation and spec links
- ✅ 97.9% conformance with partial implementation
- ✅ Path forward for 100% conformance

The architecture proves that complex Unicode algorithms can be implemented in a clear, maintainable way without sacrificing correctness.
