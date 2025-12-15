# Testing & Conformance

This document describes the conformance testing strategy for all UAX implementations in this repository.

## Official Unicode Conformance Tests

All UAX implementations are tested against official Unicode Consortium test files to ensure conformance with the Unicode Standard.

### Test Files & Conformance Status

#### UAX #29 - Text Segmentation

**Location:** `uax29/`
- `GraphemeBreakTest.txt` - Grapheme cluster boundary tests
- `WordBreakTest.txt` - Word boundary tests
- `SentenceBreakTest.txt` - Sentence boundary tests

**Source:** https://www.unicode.org/Public/17.0.0/ucd/auxiliary/

**Version:** Unicode 17.0.0

**Conformance Status:**
- ✅ Grapheme cluster boundaries: 766/766 tests (100.0%)
- ✅ Word boundaries: 1,944/1,944 tests (100.0%)
- ✅ Sentence boundaries: 512/512 tests (100.0%)
- **Total: 3,222/3,222 tests passing (100%)**

#### UAX #14 - Line Breaking

**Location:** `uax14/`
- `LineBreakTest.txt` - Line break opportunity tests (optional, auto-downloaded in CI)

**Source:** https://www.unicode.org/Public/17.0.0/ucd/auxiliary/LineBreakTest.txt

**Conformance Status:**
- ⚠️  Official conformance: 13,973/19,338 tests (72.3%)
- Note: This is expected for our simplified implementation focused on practical text layout

### CI Conformance Testing

The CI pipeline includes a dedicated `fetch-conformance-tests` job that:

1. **Downloads fresh test files** directly from unicode.org for Unicode 17.0.0
2. **Verifies** the downloaded files are valid and have correct version markers
3. **Runs all conformance tests** with the freshly downloaded files
4. **Ensures** our implementations work correctly with the latest official test data

This runs on every push and pull request, guaranteeing conformance claims are always accurate.

See: `.github/workflows/ci.yml` → `fetch-conformance-tests` job

### Updating Test Files

To update to a newer version of Unicode:

```bash
# UAX #29 test files
curl -o uax29/GraphemeBreakTest.txt \
  https://www.unicode.org/Public/17.0.0/ucd/auxiliary/GraphemeBreakTest.txt
curl -o uax29/WordBreakTest.txt \
  https://www.unicode.org/Public/17.0.0/ucd/auxiliary/WordBreakTest.txt
curl -o uax29/SentenceBreakTest.txt \
  https://www.unicode.org/Public/17.0.0/ucd/auxiliary/SentenceBreakTest.txt

# UAX #14 test file (optional, for local testing)
curl -o uax14/LineBreakTest.txt \
  https://www.unicode.org/Public/17.0.0/ucd/auxiliary/LineBreakTest.txt

# Run tests to verify
go test ./...
```

After updating, also update version references in:
- This TESTING.md file
- Package documentation (`uax29/uax29.go`, etc.)
- README.md conformance section
- CI workflow `.github/workflows/ci.yml` URLs

## Test Coverage

**UAX14**: 92.2% code coverage with comprehensive test suite:
- 89 manual test cases across multiple categories
- 167 comprehensive Unicode tests (control characters, Asian scripts, Middle Eastern, emoji, etc.)
- 19,338 official Unicode Consortium test vectors (72.9% pass rate)

## Test Categories

### Basic Functionality (13 tests)
- Empty strings
- Simple words
- Multiple words
- Spaces and whitespace
- Newlines
- Hyphenation modes
- CJK text
- Mixed scripts
- Punctuation
- Numbers

### Edge Cases (76 tests)
1. **Unicode Whitespace** (7 tests)
   - Tab characters
   - Non-breaking spaces (U+00A0)
   - Zero-width spaces (U+200B) ✅
   - Word joiners (U+2060)
   - Line separators (U+2028) ✅
   - Paragraph separators (U+2029) ✅
   - Next line (U+0085) ✅

2. **Line Breaks** (4 tests)
   - CR+LF sequences
   - Multiple newlines
   - CR only
   - Mixed line endings

3. **Hyphens** (5 tests)
   - Multiple soft hyphens
   - Soft hyphen at start
   - Soft hyphen at end
   - Em dash (U+2014)
   - En dash (U+2013)

4. **Empty and Spaces** (4 tests)
   - Only spaces
   - Single space
   - Leading spaces
   - Trailing spaces

5. **Punctuation** (9 tests)
   - Quoted text
   - Nested quotes
   - Apostrophes (contractions)
   - Ellipsis
   - Multiple exclamation marks
   - Question marks
   - Mixed punctuation
   - Brackets
   - Nested brackets

6. **Numbers** (8 tests)
   - Dates with slashes
   - Dates with dashes
   - Times (HH:MM:SS)
   - Decimals
   - Thousands separators
   - Phone numbers
   - Version numbers
   - Currency

7. **Combining Marks** (6 tests)
   - Precomposed characters (é)
   - Combining acute accent (e + U+0301)
   - Multiple combining marks
   - Combined marks in words
   - Emoji with skin tone modifiers
   - Emoji with ZWJ sequences

8. **URLs and Email** (5 tests)
   - Simple URLs
   - URLs with paths
   - URLs with query strings
   - Email addresses
   - Email with subdomains

9. **Performance** (1 test)
   - Long text (10,000 words)
   - Ascending order verification
   - Position validation

10. **No Break Opportunities** (3 tests)
    - Long words without breaks
    - Text with word joiners
    - Text with non-breaking spaces

11. **Mixed Scripts** (7 tests)
    - Latin + Arabic
    - Latin + Hebrew
    - Latin + Cyrillic
    - Latin + Greek
    - Latin + Thai
    - Latin + Korean
    - Multiple scripts mixed

## Fixed Issues

The following issues were identified during comprehensive edge case testing and have been **fixed**:

### Fixes Applied
1. **Wildcard Pattern Matching**: Updated `getBreakAction()` to support wildcard lookups with `ClassXX`
2. **Zero-Width Space (U+200B)**: Now correctly creates break opportunities ✅
3. **Line Separator (U+2028)**: Now creates mandatory breaks ✅
4. **Paragraph Separator (U+2029)**: Now creates mandatory breaks ✅
5. **Next Line (U+0085)**: Now creates mandatory breaks ✅

### How It Was Fixed
- Added fallback wildcard pattern matching in `getBreakAction()` to check `{before, ClassXX}` and `{ClassXX, after}` patterns
- Added explicit handling for `ClassZW` (zero-width space) in the `BreakDirect` case

All special Unicode whitespace and line breaking characters now work correctly according to UAX #14 specification.

## Benchmarks

```
BenchmarkFindLineBreakOpportunities      - Basic English text
BenchmarkFindLineBreakOpportunitiesCJK   - Chinese/Japanese text
```

## Running Tests

```bash
# All tests
go test ./...

# With coverage
go test ./... -cover

# Verbose
go test -v ./...

# Specific category
go test -run TestEdgeCases_URLs

# Benchmarks
go test -bench=.
```

## Official Unicode Test Vectors

We test against the official [LineBreakTest.txt](https://www.unicode.org/Public/UCD/latest/ucd/auxiliary/LineBreakTest.txt) provided by the Unicode Consortium.

**Results**: 78.5% pass rate (15,186 / 19,338 tests)

### Character Classification

The implementation now uses **official Unicode LineBreak-17.0.0.txt property data** for comprehensive and authoritative character classification. The generated data file (`linebreak_data.go`) contains 3,685 lines covering all Unicode character line break properties with binary search lookup.

### Why Not 100%?

This is expected for our simplified implementation:

**What we implement well** (explains the 76% pass rate):
- Word boundaries (spaces, tabs)
- Mandatory breaks (newlines, line separators, including \v and \f)
- LB6: Do not break before hard line breaks
- LB7: Do not break before spaces, zero width space, or ZW
- LB11: Do not break before or after word joiner (WJ)
- LB12: Do not break after non-breaking glue (GL)
- LB13: Do not break before closing punctuation (CL, CP, EX, IS, SY)
- LB14: Do not break after opening punctuation (OP SP* ×), including QU
- LB16: Do not break before nonstarters (NS)
- LB18: Break after spaces (with proper exclusions for special characters)
- LB19: Do not break before or after quotation marks (partial implementation)
- LB21: Do not break before BA or HY
- LB23, LB24, LB25: Partial numeric expression breaks (PR/PO prefix/postfix for common currency symbols)
- LB26, LB27: Korean syllable rules (H2/H3) and Jamo (JL/JV/JT)
- LB28: Do not break between alphabetics (AL, HL)
- LB29: Do not break between numeric punctuation and alphabetics
- LB30: Do not break between letters/numbers and OP/CP
- Zero-width spaces and joiners
- CJK ideographic breaks
- Ambiguous East Asian (AI) character breaks with comprehensive rules
- Hangul syllables (H2/H3) and Jamo (JL/JV/JT)
- Indic Aksara scripts (AK class for Balinese, Brahmi)
- Exclamation/interrogation marks (including presentation forms and fullwidth)
- Non-breaking spaces
- Soft hyphens with hyphenation mode support

**What remains to be implemented** (explains the 24% fail rate):
- Character class detection (need official LineBreak.txt property data for complete accuracy)
- Complex East Asian width rules (EA class)
- Advanced Hangul rules (LB26/LB27 - treat all Hangul as H2 for simplicity)
- Tailored break rules for specific scripts (SA, SG, AP, AS, VF, VI classes)
- Comprehensive PR/PO detection (only common currency symbols, need full Unicode data for all prefix/postfix)
- Break After class (BA) - requires precise Unicode property data
- Complex quotation mark handling (LB15, LB19)
- Regional indicator sequences (RI - flag emoji pairs)
- Emoji modifiers (EB, EM classes) - requires comprehensive emoji property data
- Conditional Japanese starter (CJ) rules

For practical text layout in Western + CJK + Korean contexts, the 76% pass rate provides excellent real-world coverage. The failures are mostly in edge cases for less common scripts (SA, EB, EM, RI, Indic conjuncts) and complex typographic scenarios that require context-dependent state tracking.

## Comparison to Original

This code was extracted from `github.com/SCKelemen/layout` where it was used successfully for practical text layout. During extraction, comprehensive edge case testing was added, which revealed and fixed several issues with special Unicode characters that were not properly handled in the original implementation.
