# Code Duplication and Integration Analysis

## Summary

Multiple packages reimplement functionality that should be imported from other packages or the standard library. This creates maintenance burden and potential inconsistencies. Additionally, some packages claim integration with others but don't actually import them.

## Integration Requirements

Per Unicode Standard design, these packages must work together:

1. **UTS #51 (Emoji) ↔ UAX #29 (Graphemes)**: Emoji sequences must be treated as single grapheme clusters
2. **UTS #51 (Emoji) ↔ UAX #11 (Width)**: Emoji width calculation
3. **UAX #29 (Graphemes) ↔ UAX #11 (Width)**: Terminal cursor movement must respect grapheme boundaries
4. **UAX #14 (Line Breaking) ↔ UAX #9 (Bidi)**: Break opportunities in bidi text (future work)

## Critical Duplications Found

### 1. **Extended Pictographic Property**

**Problem**: Both `uax29` and `uts51` implement ExtendedPictographic detection differently.

**Locations**:
- `uax29/grapheme.go:196` - `isExtendedPictographic()` with hardcoded ranges
- `uax29/word.go:165` - Uses the hardcoded function
- `uts51/uts51.go:228` - `IsExtendedPictographic()` with generated data

**Impact**:
- uax29 has partial, hardcoded implementation
- uts51 has complete, data-driven implementation
- Risk of inconsistency between packages

**Solution**: uax29 should import and use `uts51.IsExtendedPictographic()`

### 2. **Regional Indicators (Flag Emoji)**

**Problem**: Hardcoded range checks `0x1F1E6..0x1F1FF` scattered across packages.

**Locations**:
- `uax29/grapheme.go:105` - Hardcoded check
- `uax29/word.go:132` - Hardcoded check
- `uts51/uts51.go:243` - `IsRegionalIndicator()` function with constants

**Impact**:
- Magic numbers duplicated in multiple files
- Harder to maintain
- uts51 has proper API

**Solution**: Use `uts51.IsRegionalIndicator()` everywhere

### 3. **Emoji Modifiers (Skin Tones)**

**Problem**: Hardcoded range checks `0x1F3FB..0x1F3FF` scattered across packages.

**Locations**:
- `uax29/grapheme.go:165` - Hardcoded check
- `uax29/grapheme.go:198` - Hardcoded check (duplicate!)
- `uax29/word.go:178` - Hardcoded check
- `uts51/emoji_data.go` - Proper data-driven implementation
- `uts51/uts51.go` - `IsEmojiModifier()` function

**Impact**:
- Same check duplicated 3 times in uax29 alone!
- uts51 has proper emoji modifier detection

**Solution**: Use `uts51.IsEmojiModifier()` everywhere

### 4. **Zero Width Joiner (ZWJ)**

**Problem**: Hardcoded check for `0x200D` in multiple places.

**Locations**:
- `uax29/grapheme.go:72`
- `uax29/word.go:46`
- `uts51/uts51.go:113` - Exports `ZeroWidthJoiner` constant

**Impact**: Magic number duplication

**Solution**: Import `uts51.ZeroWidthJoiner` constant

## Recommended Refactoring

### Phase 1: Make uax29 depend on uts51

```go
// uax29/grapheme.go
import "github.com/SCKelemen/unicode/uts51"

// Replace isExtendedPictographic with:
func isExtendedPictographic(r rune) bool {
    return uts51.IsExtendedPictographic(r)
}

// Replace hardcoded regional indicator checks:
if r >= 0x1F1E6 && r <= 0x1F1FF {
// With:
if uts51.IsRegionalIndicator(r) {

// Replace hardcoded emoji modifier checks:
if r >= 0x1F3FB && r <= 0x1F3FF {
// With:
if uts51.IsEmojiModifier(r) {

// Replace hardcoded ZWJ:
if r == 0x200D {
// With:
if r == uts51.ZeroWidthJoiner {
```

### Phase 2: Remove Duplicate Logic

**Files to modify**:
1. `uax29/grapheme.go` - 3 replacements
2. `uax29/word.go` - 3 replacements

## Benefits

1. **Single Source of Truth**: Emoji properties defined once in uts51
2. **Consistency**: All packages use same emoji detection logic
3. **Maintainability**: Update emoji data in one place
4. **Correctness**: uts51 has complete, tested emoji data
5. **Code Reduction**: Remove ~100 lines of duplicate code

## Dependencies Impact

Currently:
- uax29 has no dependencies
- uts51 has no dependencies

After refactoring:
- uax29 → depends on uts51
- uts51 has no dependencies

This is the correct direction: text segmentation (uax29) naturally depends on emoji properties (uts51) for proper grapheme cluster and word boundary detection.

## Testing Impact

- All existing tests should pass
- Behavior should be identical or improved (uts51 has more complete data)
- Conformance tests ensure correctness

## Additional Integration Issues

### 5. **Emoji Width Calculation (UTS #51 ↔ UAX #11)**

**Problem**: `uts51` implements its own `EmojiWidth()` function instead of using `uax11`.

**Locations**:
- `uts51/uts51.go:303` - `EmojiWidth()` with hardcoded logic
- Comments mention "This integrates with UAX #11" but no actual import

**Current Implementation**:
```go
func EmojiWidth(r rune) int {
    if IsEmojiComponent(r) {
        return 0
    }
    if HasEmojiPresentation(r) {
        return 2  // Hardcoded!
    }
    if IsEmoji(r) {
        return 1
    }
    return 1
}
```

**Impact**:
- Parallel implementation of width calculation
- Comments claim integration but no actual code integration
- Could diverge from UAX #11 East Asian Width specifications

**Solution Options**:

**Option A (Recommended)**: Keep `EmojiWidth()` as a specialized function since:
- Emoji width logic has special cases (components = 0, presentation selector handling)
- Not all emoji widths map directly to East Asian Width property
- Performance: avoid circular dependencies

**Option B**: Make uts51 depend on uax11 and use it as a fallback:
```go
import "github.com/SCKelemen/unicode/uax11"

func EmojiWidth(r rune) int {
    if IsEmojiComponent(r) {
        return 0
    }
    // ... emoji-specific logic ...

    // Fallback to East Asian Width for non-emoji cases
    return uax11.CharWidth(r, uax11.ContextNarrow)
}
```

**Recommendation**: Keep current approach (Option A) since emoji width has special requirements distinct from general East Asian Width. However, add tests to verify alignment where they should match.

### 6. **Standard Library Usage**

**Good News**: Already using standard library extensively!

**Current Usage**:
- `unicode.Is(unicode.Cc, r)` - Control characters
- `unicode.Is(unicode.Me, r)`, `unicode.Is(unicode.Mn, r)`, `unicode.Is(unicode.Mc, r)` - Marks
- `unicode.IsDigit(r)` - Digits
- `unicode.IsLetter(r)` - Letters
- `unicode.Is(unicode.Ideographic, r)` - CJK ideographs
- `unicode.Is(unicode.Hebrew, r)` - Hebrew script
- And many more...

**Standard Library Limitations**:

Go's `unicode` package (Unicode 15.0.0) does provide `Regional_Indicator`, but is missing:
- `Extended_Pictographic` ❌
- `Emoji`, `Emoji_Presentation`, `Emoji_Modifier`, `Emoji_Modifier_Base`, `Emoji_Component` ❌
- All UAX #29 break properties ❌
- All UAX #11, #14, #50 properties ❌

**Design Decision**: We implement all emoji properties in `uts51` (including `Regional_Indicator`) rather than mixing stdlib and custom implementations because:

1. **Version consistency**: We implement Unicode 17.0.0, stdlib has 15.0.0
2. **Completeness**: All emoji-related properties in one authoritative package
3. **Single source of truth**: UTS #51 is the emoji specification
4. **Maintainability**: Updates to emoji data happen in one place

**No Action Needed**: We're appropriately using the standard library for basic Unicode properties and only implementing specialized properties that the standard library doesn't provide or that belong together as part of a Unicode specification.

## Priority Summary

| Issue | Priority | Impact | Effort |
|-------|----------|--------|--------|
| 1. ExtendedPictographic duplication | **HIGH** | Inconsistency risk | Low |
| 2. Regional Indicators duplication | **HIGH** | Magic numbers | Low |
| 3. Emoji Modifiers duplication | **HIGH** | 3x duplication | Low |
| 4. ZWJ constant duplication | **MEDIUM** | Magic numbers | Trivial |
| 5. Emoji width integration | **LOW** | Documentation | Low |
| 6. Standard library usage | **NONE** | Already optimal | N/A |

## Recommended Implementation Order

### Phase 1: UAX #29 → UTS #51 Integration (HIGH PRIORITY)

Make uax29 depend on uts51 for emoji properties:

1. Add import: `import "github.com/SCKelemen/unicode/uts51"`
2. Replace `isExtendedPictographic()` implementation
3. Replace regional indicator checks (2 locations)
4. Replace emoji modifier checks (3 locations)
5. Replace ZWJ constant (2 locations)
6. Run all conformance tests to verify behavior unchanged
7. Update documentation to reflect dependency

**Expected Result**:
- Remove ~50 lines of duplicate code
- Single source of truth for emoji properties
- All 3,222 UAX #29 conformance tests still pass
- All 5,223 UTS #51 conformance tests still pass

### Phase 2: Documentation (MEDIUM PRIORITY)

Update documentation to clarify integration points:

1. Add "Integration" section to each README
2. Document which packages depend on which
3. Add examples showing cross-package usage
4. Update godoc comments to reference related packages

### Phase 3: Future Work (LOW PRIORITY)

Consider these for future enhancements:

1. **UAX #29 ↔ UAX #11**: Add examples showing how to use grapheme segmentation with width calculation for terminal cursor movement
2. **UAX #14 ↔ UAX #9**: Implement bidirectional text support (UAX #9 not yet implemented)
3. **Cross-package integration tests**: Add tests that verify packages work correctly together

## Status

### ✅ Phase 1: COMPLETED

**Date**: 2025-12-15

Successfully refactored uax29 to depend on uts51 for emoji properties:

**Changes Made**:
1. ✅ Added `import "github.com/SCKelemen/unicode/uts51"` to grapheme.go, word.go, and sentence.go
2. ✅ Replaced `isExtendedPictographic()` implementation with `uts51.IsExtendedPictographic()`
3. ✅ Replaced regional indicator checks (2 locations) with `uts51.IsRegionalIndicator()`
4. ✅ Replaced emoji modifier checks (3 locations) with `uts51.IsEmojiModifier()`
5. ✅ Replaced ZWJ constant (4 locations) with `uts51.ZeroWidthJoiner`
6. ✅ Updated documentation (README.md, uax29.go) to reflect dependency
7. ✅ Verified no circular dependencies
8. ✅ All conformance tests still pass at 100%

**Test Results**:
- ✅ Grapheme cluster boundaries: 766/766 tests (100.0%)
- ✅ Word boundaries: 1,944/1,944 tests (100.0%)
- ✅ Sentence boundaries: 512/512 tests (100.0%)
- ✅ UTS #51 emoji conformance: 5,223/5,223 tests (100.0%)

**Benefits Achieved**:
- Removed ~50 lines of duplicate hardcoded emoji detection logic
- Single source of truth for emoji properties (UTS #51)
- Consistent emoji handling across all packages
- Complete, data-driven emoji support (vs. partial hardcoded ranges)
- No performance regression
- No behavior changes (100% conformance maintained)

### 📋 Phase 2: Documentation (Recommended Next)

Update cross-package documentation:
1. Add "Integration" section to uts51/README.md
2. Document package dependency graph in main README.md
3. Add examples showing cross-package usage

### 🔮 Phase 3: Future Work

Consider for future enhancements:
1. UAX #29 ↔ UAX #11: Terminal cursor movement examples
2. UAX #14 ↔ UAX #9: Bidirectional text support (when UAX #9 is implemented)
3. Cross-package integration tests
