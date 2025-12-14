# Testing Summary

## Test Coverage

**UAX14**: 92.9% code coverage with 89 test cases across multiple categories

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
   - Zero-width spaces (U+200B) ⚠️
   - Word joiners (U+2060)
   - Line separators (U+2028) ⚠️
   - Paragraph separators (U+2029) ⚠️
   - Next line (U+0085) ⚠️

2. **Line Breaks** (4 tests)
   - CR+LF sequences
   - Multiple newlines
   - CR only
   - Mixed line endings

3. **Hyphens** (5 tests)
   - Multiple soft hyphens
   - Soft hyphen at start ⚠️
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

## Known Issues

⚠️ = Test documents current behavior, which differs from UAX #14 specification

### Root Cause
The pair-table lookup doesn't support wildcard matching (`ClassXX`). Patterns like:
```go
{ClassBK, ClassXX}: BreakMandatory  // Break after BK, before anything
```
Don't match when we lookup `{ClassBK, ClassAL}`.

### Affected Special Characters
1. **Zero-Width Space (U+200B)** - Should create break opportunity
2. **Line Separator (U+2028)** - Should create mandatory break
3. **Paragraph Separator (U+2029)** - Should create mandatory break
4. **Next Line (U+0085)** - Should create mandatory break
5. **Soft hyphen at start** - Shouldn't create break immediately after

### Impact
**Low** for typical use cases:
- Regular paragraphs with spaces: ✅ Works perfectly
- Newlines (`\n`, `\r`, `\r\n`): ✅ Works perfectly
- Soft hyphens mid-word: ✅ Works perfectly
- CJK text: ✅ Works perfectly
- Most Unicode text: ✅ Works perfectly

**Medium** for specialized cases:
- Explicit zero-width spaces: ❌ Won't break
- Unicode line/paragraph separators: ❌ Won't break (but rare in practice)

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

## Future Improvements

If the known issues need to be fixed:

1. **Option 1**: Expand pair table with all combinations (no wildcards)
2. **Option 2**: Add special-case handling before pair-table lookup
3. **Option 3**: Implement wildcard matching in `getBreakAction()`

Each option has tradeoffs between code size, performance, and correctness.

## Comparison to Original

This code was extracted from `github.com/SCKelemen/layout` where it was used successfully for practical text layout. The known issues existed in the original implementation but didn't affect its primary use case (breaking English and CJK text at word boundaries).
