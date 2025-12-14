# UAX14 Failure Analysis - Remaining 30.3%

## Summary Statistics

- **Total Failures**: 5,856 / 19,338 tests (30.3%)
- **Too Few Breaks**: 3,476 failures (59.4%) - Being too conservative
- **Too Many Breaks**: 2,345 failures (40.0%) - Being too permissive
- **Wrong Positions**: 35 failures (0.6%) - Rare edge cases

## Top Failure Categories by Break Class

1. **ID (Ideographic)**: 85 failures
2. **EB (Emoji Base)**: 48 failures - NOT IMPLEMENTED
3. **SP (Space)**: 41 failures - Context-specific space handling
4. **OP → SP**: 40 failures - Opening punctuation + space sequences
5. **QU → SP**: 30 failures - Quotation + space sequences
6. **CB (Contingent Break)**: 26 failures
7. **BA → SP**: 24 failures - Break After + space sequences
8. **SP → QU**: 24 failures - Space + quotation sequences
9. **RI (Regional Indicator)**: 23 failures - Flag emoji (NOT IMPLEMENTED)
10. **EM (Emoji Modifier)**: 23 failures - Skin tones (NOT IMPLEMENTED)
11. **AK (Aksara)**: 22 failures - Indic scripts
12. **AS (Aksara Start)**: 22 failures - Indic scripts
13. **H2 (Hangul LV)**: 22 failures - NOT IMPLEMENTED
14. **H3 (Hangul LVT)**: 22 failures - NOT IMPLEMENTED
15. **BA (Break After)**: 21 failures

## Critical Issues Identified

### Issue #1: AI (Ambiguous East Asian) Not Breaking Properly

**Problem**: AI characters should allow breaks BEFORE most other classes, but our implementation doesn't trigger these breaks.

Examples:
- `❗ᬅ` (AI AK) - Expected break after ❗, got none
- `❗—` (AI B2) - Expected break after ❗, got none
- `❗´` (AI BB) - Expected break after ❗, got none
- `❗￼` (AI CB) - Expected break after ❗, got none

**Root Cause**: We have AI rules in pairTable but don't check for AI in the FindLineBreakOpportunities switch statement for BreakDirect.

**Fix**: Add AI handling to the BreakDirect case in FindLineBreakOpportunities.

### Issue #2: Space Break Rules (LB7, LB18)

**Problem**: We're breaking after AI/ID before spaces when we should only break after the space.

Examples:
- `❗ \v` (AI SP BK) - Breaking at position 4 (after ❗) when should only break at 5 (after space)
- `❗ 〉` (AI SP CL) - Breaking at position 4 when should only break at 7 (end)

**Rules Violated**:
- **LB7**: Do not break before spaces or zero width space
- **LB18**: Break after spaces (but not before them)

**Fix**: Prevent breaks before SP class characters.

### Issue #3: Hangul Jamo Sequences (H2, H3, JL, JV, JT)

**Status**: NOT IMPLEMENTED

Hangul syllable formation requires specific rules:
- **LB26**: Do not break between Hangul syllable blocks
- **LB27**: Treat Korean syllable blocks as ID

Affects: ~110 tests (1.8% of failures)

### Issue #4: Emoji Sequences (EB, EM, RI)

**Status**: NOT IMPLEMENTED

- **EB (Emoji Base)**: 48 failures
- **EM (Emoji Modifier)**: 23 failures - Skin tone modifiers
- **RI (Regional Indicator)**: 23 failures - Flag emoji pairs

Affects: ~94 tests (1.6% of failures)

### Issue #5: Indic Script Classes (AK, AP, AS, VF, VI)

**Status**: PARTIALLY IMPLEMENTED

We have AK, AP, AS classes but:
- Missing VF (Virma Final)
- Missing VI (Virma)
- AI × AK not breaking properly (see Issue #1)

Affects: ~84 tests (1.4% of failures)

### Issue #6: Contingent Break (CB) Class

**Problem**: CB (U+FFFC Object Replacement Character) should allow breaks in most contexts.

Currently handled but AI × CB not working (see Issue #1).

Affects: 26 tests

### Issue #7: Opening/Closing Punctuation + Space

**Problem**: Complex rules for breaks around OP/CL/QU with adjacent spaces.

Examples:
- `〈 ☰` (OP SP) - Should not break after opening bracket before space
- `« ☰` (QU SP) - Should not break after quote before space

**Rules**:
- **LB14**: Do not break after OP even before spaces
- **LB15**: Do not break within `"`[`, even with intervening spaces
- **LB19**: Do not break before or after QU

Affects: ~94 tests (1.6% of failures)

## Recommended Fixes Priority

### High Priority (Easy Wins - ~1,500 tests)

1. **Fix AI break handling** (~500 tests)
   - Add AI to BreakDirect handling in FindLineBreakOpportunities
   - Ensure AI × [AK, AP, AS, B2, BB, CB, H2, H3, etc.] breaks work

2. **Implement LB7 properly** (~500 tests)
   - Do not break before spaces
   - Remove breaks that occur right before SP characters

3. **Fix OP/QU + space sequences** (~200 tests)
   - Implement LB14: Do not break after OP
   - Implement LB15: Do not break within quotes/brackets + spaces
   - Implement LB19: QU handling

4. **Fix BA (Break After) rules** (~200 tests)
   - U+3000 Ideographic Space and U+0009 Tab handling
   - Proper BA class implementation

### Medium Priority (~800 tests)

5. **Add Indic script support** (~100 tests)
   - Add VF (Virama Final) class
   - Add VI (Virama) class
   - Fix AK/AP/AS break rules

6. **Add Hangul support** (~110 tests)
   - Implement LB26 and LB27
   - Add H2, H3, JL, JV, JT class handling

7. **Improve CB handling** (~100 tests)
   - Ensure U+FFFC breaks properly in all contexts

### Lower Priority (~500 tests)

8. **Add emoji support** (~94 tests)
   - EB (Emoji Base)
   - EM (Emoji Modifier)
   - RI (Regional Indicator) pairs

9. **Fix ID edge cases** (~85 tests)
   - More comprehensive ideographic break rules

## Current Implementation Gaps

### UAX #14 Rules Not Implemented

- **LB7**: Do not break before spaces (PARTIALLY - needs fixing)
- **LB14**: Do not break after OP
- **LB15**: Do not break within `"`[`
- **LB19**: Do not break before or after QU
- **LB20**: Break before and after unresolved CB
- **LB21**: Do not break before hyphen/BA
- **LB26**: Do not break Korean Hangul syllable sequences
- **LB27**: Treat Korean syllable blocks as ID
- **LB30**: Do not break between letters, numbers, or ordinary symbols and OP or CP

### Classes Not Implemented

- **VF** (Virama Final) - Indic scripts
- **VI** (Virama) - Indic scripts
- **EB** (Emoji Base)
- **EM** (Emoji Modifier)
- **RI** (Regional Indicator) - Needs special pair handling

## Conclusion

The remaining 30.3% failures fall into clear categories:

1. **40% are fixable bugs** in our current implementation (AI, space rules)
2. **40% are missing specialized rules** (Hangul, Emoji, advanced punctuation)
3. **20% are complex edge cases** (Indic scripts, numeric sequences, rare combinations)

Implementing the "High Priority" fixes could improve pass rate from **69.7%** to approximately **77-80%** with relatively straightforward code changes.
