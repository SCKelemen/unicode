# Priority Fixes for Remaining 27.1% Failures

**Current Status**: 14,099 / 19,338 passing (72.9%)
**Remaining**: 5,239 failures (27.1%)

## Phase 1 Attempt Results

Attempted to implement Phase 1 (Quick Wins) with the following outcomes:

1. **BA (Break After)** - Attempted but reverted
   - Issue: Classifying U+3000 as BA made tests worse (-17 tests)
   - Root cause: BA requires precise Unicode property data, not just manual classification
   - Conclusion: Needs official LineBreak.txt data

2. **PR/PO (Numeric Prefix/Postfix)** - **SUCCESS: +3 tests** ✓
   - Added detection for common currency symbols ($, £, €, ¥, etc.)
   - Implemented LB23, LB24, LB25 rules
   - Result: 14,099 / 19,338 (72.9%)
   - Note: Limited improvement because comprehensive PR/PO detection requires full Unicode data

3. **Emoji (EB/EM/RI)** - Attempted but reverted
   - Issue: Manual emoji ranges were too broad and misclassified characters (-77 tests)
   - Root cause: Emoji property data is complex and context-dependent
   - Conclusion: Requires official Unicode emoji property data

**Key Learning**: Manual character range detection for complex classes (BA, EB/EM/RI) is insufficient. To achieve significant improvements beyond 73%, the implementation needs:
- Official Unicode LineBreak.txt property data
- Official Unicode emoji property data
- Context-aware break rules based on Unicode algorithms

The current implementation at 72.9% represents the practical limit for manual range-based character classification.

## Executive Summary (Original Analysis)

- **41.9%** of failures are "too many breaks" (breaking where we shouldn't)
- **57.7%** of failures are "too few breaks" (not breaking where we should)
- **0.4%** have wrong positions

The top 6 categories account for **2,596 failures (49.5% of all failures, 13.4% of total tests)**

## Top Priority Fixes (Ranked by Impact)

### 1. Emoji Support (EB/EM/RI) - **644 failures (12.3%)**

**Problem**: No detection for emoji-specific break classes
- **EB (Emoji Base)**: Characters like ✊ (U+270A), ☝ (U+261D)
- **EM (Emoji Modifier)**: Skin tone modifiers like 🏻 (U+1F3FB)
- **RI (Regional Indicator)**: Flag emoji pairs 🇦 (U+1F1E6)

**Examples**:
- "❗✊" - Expected [3, 6], Got [6] - AI × EB should break
- "◌🏻" - Emoji modifier detection missing

**Fix Required**:
- Add character range detection for EB: U+261D, U+270A, etc.
- Add EM detection: U+1F3FB-U+1F3FF (skin tone modifiers)
- Add RI detection: U+1F1E6-U+1F1FF (regional indicators)
- Implement LB30b: Do not break emoji modifier sequences

**Estimated Effort**: Medium (need comprehensive emoji ranges)
**Estimated Gain**: +644 tests (+3.3%)

### 2. Numeric Prefix/Postfix (PR/PO) - **489 failures (9.3%)**

**Problem**: No detection for numeric prefix/postfix classes
- **PR (Prefix Numeric)**: Currency symbols $ ₩ £ € ¥
- **PO (Postfix Numeric)**: Percent %, degree °, etc.

**Examples**:
- "❗₩" - PR detection missing
- "❗$" - Should classify as PR

**Fix Required**:
- Add PR class detection for currency symbols
- Add PO class detection for percent, degree, etc.
- Implement LB23: Do not break between digits and letters
- Implement LB24: Do not break between prefix and letters/ideographs
- Implement LB25: Do not break between prefix/postfix and numbers

**Estimated Effort**: Low-Medium
**Estimated Gain**: +489 tests (+2.5%)

### 3. Ideographic Detection (ID) - **468 failures (8.9%)**

**Problem**: Characters that should be ClassID aren't detected
- ☀ (U+2600 SUN) - Detected as symbol, should be ID
- ⌚ (U+231A WATCH) - Detected as symbol, should be ID
- Unassigned codepoints in ideographic ranges

**Examples**:
- "❗☀" - Expected [3, 6], Got [6] - AI × ID should break
- "❗\U0001fffd" - Unassigned ideographic

**Fix Required**:
- Load official Unicode LineBreak.txt property data
- OR: Add comprehensive ID character ranges

**Estimated Effort**: High (requires data file or extensive ranges)
**Estimated Gain**: +468 tests (+2.4%)

### 4. Indic Virama (VF/VI) - **355 failures (6.8%)**

**Problem**: Indic script virama characters not detected
- **VF (Virama Final)**: ᯲ (U+1BF2) Balinese Musical Symbol Dang
- **VI (Virama)**: ᭄ (U+1B44) Balinese Sign Rerekan

**Examples**:
- "❗᯲" - Expected [3, 6], Got [6]
- "❗᭄" - VF/VI detection missing

**Fix Required**:
- Add VF class detection for U+1BF2, etc.
- Add VI class detection for U+1B44, etc.
- Indic script conjunct handling

**Estimated Effort**: Medium
**Estimated Gain**: +355 tests (+1.8%)

### 5. Break After (BA) - **324 failures (6.2%)**

**Problem**: BA class not properly detected or handled
- U+3000 (IDEOGRAPHIC SPACE) - Should be BA
- U+0009 (TAB) in some contexts

**Examples**:
- "\u3000❗" - Ideographic space before text
- "\u3000  " - BA followed by spaces

**Fix Required**:
- Add U+3000 as ClassBA (currently ClassSP?)
- Implement LB21: Do not break before hyphen-minus or BA
- Handle BA + SP sequences properly

**Estimated Effort**: Low
**Estimated Gain**: +324 tests (+1.7%)

### 6. Quotation + Space (QU SP) - **316 failures (6.0%)**

**Problem**: QU followed by space has complex rules
- "❗ »" - Space before closing quotation
- "» ❗" - After closing quotation + space

**Fix Required**:
- Implement LB19: Do not break before or after quotation marks
- Handle QU SP sequences (related to LB14 we just implemented)
- May need to distinguish opening vs closing QU

**Estimated Effort**: Medium
**Estimated Gain**: +316 tests (+1.6%)

## Medium Priority (7-12)

### 7. Zero-Width Space (ZW) - **227 failures (4.3%)**

**Examples**: "ᬅ \u200b" - ZW interactions with spaces

**Fix**: Better ZW handling in various contexts

### 8. Break Before/After (B2) - **150 failures (2.9%)**

**Examples**: "❗—" (EM DASH U+2014)

**Fix**: B2 class rules - can break before and after

### 9. Space + Closing Punct (SP CL) - **146 failures (2.8%)**

**Examples**: "❗ 〉" - Extra break at position 4

**Fix**: LB13 extensions for space before closing punct

### 10. Ambiguous AI Issues - **140 failures (2.7%)**

**Examples**: "❗◌" (DOTTED CIRCLE) - Extra breaks

**Fix**: Refine AI rules for edge cases

### 11. Opening Punct + Space (OP SP) - **111 failures (2.1%)**

**Examples**: "\v̈ 〈" - LB14 edge cases

**Fix**: Extend LB14 to more contexts

### 12. Word Joiner (WJ) - **96 failures (1.8%)**

**Examples**: "ᬅ \ufeff"

**Fix**: WJ handling in space contexts

## Low Priority (Under 1% each)

- Multiple Spaces (SP SP): 41 failures (0.8%)
- Various script-specific classes: AS, H2, H3, AK, AP, JL, JV, JT
- Each under 20 failures

## Implementation Roadmap

### Phase 1: Quick Wins (~1,500 tests, +7.8%)

1. **BA (Break After)** - 324 tests, Low effort
   - Add U+3000 as ClassBA
   - Implement LB21

2. **PR/PO (Numeric)** - 489 tests, Medium effort
   - Add currency symbol detection
   - Implement LB23, LB24, LB25

3. **Emoji (EB/EM/RI)** - 644 tests, Medium effort
   - Add emoji character ranges
   - Implement LB30b

4. **Zero-Width Space (ZW)** - 227 tests, Low effort
   - Fix ZW + SP interactions

**Estimated Impact**: +1,684 tests (+8.7%)
**New Pass Rate**: ~81.6%

### Phase 2: Character Detection (~800 tests, +4.1%)

5. **Ideographic (ID)** - 468 tests, High effort
   - Requires LineBreak.txt data or comprehensive ranges

6. **Indic Virama (VF/VI)** - 355 tests, Medium effort
   - Add VF/VI character detection

**Estimated Impact**: +823 tests (+4.3%)
**New Pass Rate**: ~85.9%

### Phase 3: Complex Rules (~450 tests, +2.3%)

7. **QU SP (Quotation)** - 316 tests
   - Implement LB19

8. **B2 (Break Before/After)** - 150 tests
   - B2 class rules

9. **Remaining AI issues** - 140 tests

**Estimated Impact**: +606 tests (+3.1%)
**New Pass Rate**: ~89.0%

### Phase 4: Edge Cases (~300 tests, +1.5%)

- SP CL refinements
- OP SP edge cases
- WJ handling
- Multiple space handling

**Estimated Impact**: +300 tests (+1.5%)
**Final Pass Rate**: ~90.5%

## Theoretical Maximum

With full UAX #14 implementation including:
- Official LineBreak property data
- All complex script rules
- Dictionary-based breaking
- Full LB1-LB31 implementation

**Estimated Maximum**: 85-90% pass rate

Remaining 10-15% would be extreme edge cases, rarely-used scripts, and context-dependent rules requiring language/dictionary knowledge.

## Recommendation

**Start with Phase 1** (Quick Wins):
1. BA class (easiest, 324 tests)
2. PR/PO classes (good ROI, 489 tests)
3. Emoji support (highest impact, 644 tests)

This would bring us from **72.9% to ~81.6%** with reasonable effort and no need for external data files.
