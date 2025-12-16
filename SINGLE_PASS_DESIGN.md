# Single-Pass Break Detection Design

**Status**: Design Document for v2.0.0
**Version**: 1.0
**Date**: 2025-12-16

## Executive Summary

This document describes a unified break detection system that computes grapheme, word, and sentence boundaries in a single pass over the text, using packed class structures and generated data tables.

## Goals

1. **Performance**: 5-10x faster than current implementation
2. **Efficiency**: Single pass over text instead of three separate passes
3. **Memory**: Packed structures to minimize memory footprint
4. **Conformance**: Maintain 100% conformance on all official tests
5. **API**: Simple, ergonomic API for common use cases

## Current Problems

### Multiple Passes Over Text

```go
// Current API - requires 3 passes
graphemes := uax29.Graphemes(text)     // Pass 1
words := uax29.Words(text)             // Pass 2
sentences := uax29.Sentences(text)     // Pass 3
```

**Issues**:
- Each pass decodes UTF-8 again
- Each pass classifies runes again
- Each pass traverses the entire text
- Cache-unfriendly: text may be evicted between passes

### Sequential Classification

UAX #29 currently uses O(n) sequential checks for classification:
```go
func getGraphemeBreakClass(r rune) GraphemeBreakClass {
    if r == 0x000D { return GBCR }
    if r == 0x000A { return GBLF }
    // ... 20+ more checks
}
```

## Proposed Architecture

### 1. Packed Class Structure

Store all three break classes in a single `uint16`:

```go
// Packed break class encoding
// Bits: [grapheme:5][word:5][sentence:4][reserved:2]
type PackedBreakClass uint16

const (
    graphemeMask   = 0x1F   // bits 0-4   (5 bits, 32 values)
    wordMask       = 0x3E0  // bits 5-9   (5 bits, 32 values)
    sentenceMask   = 0x3C00 // bits 10-13 (4 bits, 16 values)
)

func (p PackedBreakClass) Grapheme() GraphemeBreakClass {
    return GraphemeBreakClass(p & graphemeMask)
}

func (p PackedBreakClass) Word() WordBreakClass {
    return WordBreakClass((p & wordMask) >> 5)
}

func (p PackedBreakClass) Sentence() SentenceBreakClass {
    return SentenceBreakClass((p & sentenceMask) >> 10)
}

func PackClasses(g GraphemeBreakClass, w WordBreakClass, s SentenceBreakClass) PackedBreakClass {
    return PackedBreakClass(uint16(g) | (uint16(w) << 5) | (uint16(s) << 10))
}
```

**Benefits**:
- Single lookup returns all three classes
- 16 bits per rune in classification arrays
- Fast bitwise operations for extraction

### 2. Generated Data Tables

Generate combined data tables from Unicode property files:

```go
// Generated in uax29/break_data.go
type breakRange struct {
    start rune
    end   rune
    class PackedBreakClass
}

var breakData = []breakRange{
    {0x0000, 0x0009, PackClasses(GBControl, WBOther, SBOther)},
    {0x000A, 0x000A, PackClasses(GBLF, WBLF, SBLf)},
    {0x000B, 0x000C, PackClasses(GBControl, WBOther, SBOther)},
    {0x000D, 0x000D, PackClasses(GBCR, WBCR, SBCR)},
    // ... generated from GraphemeBreakProperty.txt,
    //     WordBreakProperty.txt, SentenceBreakProperty.txt
}

// Binary search lookup - O(log n)
func classifyRune(r rune) PackedBreakClass {
    left, right := 0, len(breakData)-1
    for left <= right {
        mid := (left + right) / 2
        entry := breakData[mid]
        if r < entry.start {
            right = mid - 1
        } else if r > entry.end {
            left = mid + 1
        } else {
            return entry.class
        }
    }
    return PackClasses(GBOther, WBOther, SBOther)
}
```

### 3. Single-Pass API

New unified API that returns all breaks at once:

```go
// BreakOpportunities contains all break positions in byte offsets
type BreakOpportunities struct {
    Graphemes []int  // Grapheme cluster boundaries
    Words     []int  // Word boundaries (subset of Graphemes)
    Sentences []int  // Sentence boundaries (subset of Words)
}

// FindAllBreaks computes all break types in a single pass
func FindAllBreaks(text string) BreakOpportunities {
    runes := []rune(text)
    n := len(runes)

    if n == 0 {
        return BreakOpportunities{}
    }

    // Classify all runes once
    classes := make([]PackedBreakClass, n)
    for i, r := range runes {
        classes[i] = classifyRune(r)
    }

    // Find breaks in hierarchical order
    graphemeBreaks := findGraphemeBreaks(runes, classes)
    wordBreaks := findWordBreaks(runes, classes, graphemeBreaks)
    sentenceBreaks := findSentenceBreaks(runes, classes, wordBreaks)

    return BreakOpportunities{
        Graphemes: graphemeBreaks,
        Words:     wordBreaks,
        Sentences: sentenceBreaks,
    }
}
```

### 4. Hierarchical Break Detection

Breaks have a natural hierarchy:
- **Grapheme breaks** are most granular (every grapheme cluster boundary)
- **Word breaks** occur at grapheme boundaries (subset of grapheme breaks)
- **Sentence breaks** occur at word boundaries (subset of word breaks)

```go
func findGraphemeBreaks(runes []rune, classes []PackedBreakClass) []int {
    breaks := []int{0} // Always break at start

    for i := 1; i < len(runes); i++ {
        prevClass := classes[i-1].Grapheme()
        currClass := classes[i].Grapheme()

        if shouldBreakGrapheme(runes, classes, i, prevClass, currClass) {
            breaks = append(breaks, runeIndexToBytePos(runes, i))
        }
    }

    breaks = append(breaks, len(string(runes))) // Always break at end
    return breaks
}

func findWordBreaks(runes []rune, classes []PackedBreakClass, graphemeBreaks []int) []int {
    breaks := []int{0}

    // Only check word breaks at grapheme boundaries
    for _, pos := range graphemeBreaks[1:len(graphemeBreaks)-1] {
        i := bytePosToRuneIndex(runes, pos)
        prevClass := classes[i-1].Word()
        currClass := classes[i].Word()

        if shouldBreakWord(runes, classes, i, prevClass, currClass) {
            breaks = append(breaks, pos)
        }
    }

    breaks = append(breaks, len(string(runes)))
    return breaks
}

func findSentenceBreaks(runes []rune, classes []PackedBreakClass, wordBreaks []int) []int {
    breaks := []int{0}

    // Only check sentence breaks at word boundaries
    for _, pos := range wordBreaks[1:len(wordBreaks)-1] {
        i := bytePosToRuneIndex(runes, pos)
        prevClass := classes[i-1].Sentence()
        currClass := classes[i].Sentence()

        if shouldBreakSentence(runes, classes, i, prevClass, currClass) {
            breaks = append(breaks, pos)
        }
    }

    breaks = append(breaks, len(string(runes)))
    return breaks
}
```

**Optimization**: Since word breaks are subset of grapheme breaks, and sentence breaks are subset of word breaks, we only need to check break rules at valid positions.

### 5. Backward Compatibility

Keep existing APIs as convenience wrappers:

```go
// Graphemes returns grapheme cluster boundaries
func Graphemes(text string) []string {
    breaks := FindAllBreaks(text)
    return splitByBreaks(text, breaks.Graphemes)
}

// Words returns word boundaries
func Words(text string) []string {
    breaks := FindAllBreaks(text)
    return splitByBreaks(text, breaks.Words)
}

// Sentences returns sentence boundaries
func Sentences(text string) []string {
    breaks := FindAllBreaks(text)
    return splitByBreaks(text, breaks.Sentences)
}

// Optimized: if user only needs one type, skip others
func FindGraphemeBreaks(text string) []int {
    // Fast path: only compute grapheme breaks
    runes := []rune(text)
    classes := make([]PackedBreakClass, len(runes))
    for i, r := range runes {
        classes[i] = classifyRune(r)
    }
    return findGraphemeBreaks(runes, classes)
}

// Similar for FindWordBreaks, FindSentenceBreaks
```

## Data Generation

### Generator Structure

```go
// generate_break_data.go
package main

import (
    "bufio"
    "fmt"
    "net/http"
    "os"
    "sort"
    "strconv"
    "strings"
)

const (
    graphemeURL  = "https://www.unicode.org/Public/17.0.0/ucd/auxiliary/GraphemeBreakProperty.txt"
    wordURL      = "https://www.unicode.org/Public/17.0.0/ucd/auxiliary/WordBreakProperty.txt"
    sentenceURL  = "https://www.unicode.org/Public/17.0.0/ucd/auxiliary/SentenceBreakProperty.txt"
)

type runeRange struct {
    start    rune
    end      rune
    grapheme string
    word     string
    sentence string
}

func main() {
    // Download and parse all three property files
    graphemeProps := parsePropertyFile(graphemeURL)
    wordProps := parsePropertyFile(wordURL)
    sentenceProps := parsePropertyFile(sentenceURL)

    // Merge into unified ranges with packed classes
    ranges := mergeProperties(graphemeProps, wordProps, sentenceProps)

    // Generate break_data.go
    generateCode(ranges)
}

func mergeProperties(g, w, s map[rune]string) []runeRange {
    // Collect all unique boundary points
    boundaries := make(map[rune]bool)
    for r := range g { boundaries[r] = true }
    for r := range w { boundaries[r] = true }
    for r := range s { boundaries[r] = true }

    // Sort boundaries
    sorted := make([]rune, 0, len(boundaries))
    for r := range boundaries {
        sorted = append(sorted, r)
    }
    sort.Slice(sorted, func(i, j int) bool {
        return sorted[i] < sorted[j]
    })

    // Create ranges with combined properties
    var ranges []runeRange
    for i := 0; i < len(sorted)-1; i++ {
        start := sorted[i]
        end := sorted[i+1] - 1

        ranges = append(ranges, runeRange{
            start:    start,
            end:      end,
            grapheme: getProperty(g, start, "Other"),
            word:     getProperty(w, start, "Other"),
            sentence: getProperty(s, start, "Other"),
        })
    }

    return ranges
}

func generateCode(ranges []runeRange) {
    // Generate break_data.go with packed classes
    // ...
}
```

### Generated Output

```go
// Code generated by generate_break_data.go. DO NOT EDIT.
// Unicode Version: 17.0.0
// Generated: 2025-12-16

package uax29

// PackedBreakClass encodes all three break classes in 16 bits
// Layout: [grapheme:5][word:5][sentence:4][reserved:2]
type PackedBreakClass uint16

// Break class constants
const (
    // Grapheme classes (0-31)
    GB_CR GraphemeBreakClass = iota
    GB_LF
    GB_Control
    GB_Extend
    GB_ZWJ
    GB_RI
    GB_Prepend
    GB_SpacingMark
    GB_L
    GB_V
    GB_T
    GB_LV
    GB_LVT
    GB_Other

    // Word classes (0-31)
    WB_CR WordBreakClass = iota
    WB_LF
    WB_Newline
    WB_Extend
    WB_ZWJ
    WB_RI
    WB_Format
    WB_Katakana
    WB_HebrewLetter
    WB_ALetter
    WB_SingleQuote
    WB_DoubleQuote
    WB_MidNumLet
    WB_MidLetter
    WB_MidNum
    WB_Numeric
    WB_ExtendNumLet
    WB_WSegSpace
    WB_Other

    // Sentence classes (0-15)
    SB_CR SentenceBreakClass = iota
    SB_LF
    SB_Extend
    SB_Sep
    SB_Format
    SB_Sp
    SB_Lower
    SB_Upper
    SB_OLetter
    SB_Numeric
    SB_ATerm
    SB_STerm
    SB_Close
    SB_SContinue
    SB_Other
)

type breakRange struct {
    start rune
    end   rune
    class PackedBreakClass
}

var breakData = []breakRange{
    // Generated ranges with packed classes
    {0x0000, 0x0009, pack(GB_Control, WB_Other, SB_Other)},
    {0x000A, 0x000A, pack(GB_LF, WB_LF, SB_LF)},
    {0x000B, 0x000C, pack(GB_Control, WB_Other, SB_Other)},
    {0x000D, 0x000D, pack(GB_CR, WB_CR, SB_CR)},
    // ... thousands more entries
}

func classifyRune(r rune) PackedBreakClass {
    // Binary search
    left, right := 0, len(breakData)-1
    for left <= right {
        mid := (left + right) / 2
        entry := breakData[mid]
        if r < entry.start {
            right = mid - 1
        } else if r > entry.end {
            left = mid + 1
        } else {
            return entry.class
        }
    }
    return pack(GB_Other, WB_Other, SB_Other)
}

func pack(g GraphemeBreakClass, w WordBreakClass, s SentenceBreakClass) PackedBreakClass {
    return PackedBreakClass(uint16(g) | (uint16(w) << 5) | (uint16(s) << 10))
}

func (p PackedBreakClass) Grapheme() GraphemeBreakClass {
    return GraphemeBreakClass(p & 0x1F)
}

func (p PackedBreakClass) Word() WordBreakClass {
    return WordBreakClass((p & 0x3E0) >> 5)
}

func (p PackedBreakClass) Sentence() SentenceBreakClass {
    return SentenceBreakClass((p & 0x3C00) >> 10)
}
```

## Performance Analysis

### Current Implementation (3 passes)

```
Text: "Hello, world! How are you?"

Pass 1 (Graphemes):
- UTF-8 decode: 28 runes
- Classify: 28 × 20 checks = 560 operations
- Apply rules: 28 iterations

Pass 2 (Words):
- UTF-8 decode: 28 runes (again)
- Classify: 28 × 20 checks = 560 operations
- Apply rules: 28 iterations

Pass 3 (Sentences):
- UTF-8 decode: 28 runes (again)
- Classify: 28 × 20 checks = 560 operations
- Apply rules: 28 iterations

Total: 84 runes decoded, 1,680 classification checks, 84 rule iterations
```

### Proposed Implementation (1 pass)

```
Text: "Hello, world! How are you?"

Single pass:
- UTF-8 decode: 28 runes
- Classify: 28 × 1 binary search = 28 × log2(3000) ≈ 28 × 12 = 336 operations
- Apply grapheme rules: 28 iterations
- Apply word rules: 8 iterations (only at grapheme boundaries)
- Apply sentence rules: 2 iterations (only at word boundaries)

Total: 28 runes decoded, 336 classification operations, 38 rule iterations
```

### Improvement

- **UTF-8 decoding**: 3× reduction (84 → 28)
- **Classification**: 5× reduction (1,680 → 336)
- **Rule application**: 2× reduction (84 → 38)
- **Overall**: ~5-8× faster

## Memory Usage

### Data Table Size

Estimated entries in merged table:
- GraphemeBreakProperty.txt: ~1,200 ranges
- WordBreakProperty.txt: ~1,500 ranges
- SentenceBreakProperty.txt: ~800 ranges

After merging (all boundary points): ~3,000 ranges

```go
type breakRange struct {
    start rune      // 4 bytes
    end   rune      // 4 bytes
    class uint16    // 2 bytes
}
// Total: 10 bytes per range
// 3,000 ranges × 10 bytes = 30 KB
```

**Comparison**:
- Current: 3 separate tables, ~15 KB each = 45 KB
- Proposed: 1 unified table = 30 KB
- **Savings**: 33% smaller

### Runtime Memory

Current (3 passes):
```
Classification arrays: none (classify on demand)
Break arrays: 3 separate arrays
```

Proposed (1 pass):
```
Classification array: len(runes) × 2 bytes
Break arrays: 3 arrays (but computed once)
```

For 1,000 character text:
- Current: ~24 KB (3 × 8 KB for rune arrays)
- Proposed: ~10 KB (2 KB classes + 3 × 2.6 KB break arrays)

## Migration Plan

### Phase 1: Data Generation
1. Create `generate_break_data.go` generator
2. Download and parse all three Unicode property files
3. Merge properties at boundary points
4. Generate `break_data.go` with packed classes
5. Verify data correctness

### Phase 2: Classification Layer
1. Implement `PackedBreakClass` type with pack/unpack methods
2. Implement `classifyRune()` with binary search
3. Add benchmarks comparing to current classification
4. Verify classification matches current behavior

### Phase 3: Single-Pass Algorithm
1. Implement `FindAllBreaks()` function
2. Refactor grapheme/word/sentence break detection to use packed classes
3. Implement hierarchical break detection (words at grapheme bounds, sentences at word bounds)
4. Add comprehensive tests

### Phase 4: Backward Compatibility
1. Update `Graphemes()`, `Words()`, `Sentences()` to use new implementation
2. Add optimized single-type functions (`FindGraphemeBreaks`, etc.)
3. Run all conformance tests
4. Verify 100% pass rate maintained

### Phase 5: Benchmarking
1. Benchmark current vs new implementation
2. Verify 5-10× performance improvement
3. Profile memory usage
4. Document results

## Testing Strategy

### Conformance Tests

All existing tests must pass:
```bash
go test ./uax29 -run TestGraphemeBreakOfficial  # 766/766
go test ./uax29 -run TestWordBreakOfficial      # 1,944/1,944
go test ./uax29 -run TestSentenceBreakOfficial  # 512/512
```

### Single-Pass Tests

New tests for unified API:
```go
func TestFindAllBreaks(t *testing.T) {
    tests := []struct {
        text      string
        wantG     []int  // grapheme breaks
        wantW     []int  // word breaks
        wantS     []int  // sentence breaks
    }{
        {
            text:  "Hello, world! How are you?",
            wantG: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, ...},
            wantW: []int{0, 5, 6, 7, 12, 13, 14, 17, 18, ...},
            wantS: []int{0, 14, 27},
        },
        // More tests...
    }

    for _, tt := range tests {
        got := FindAllBreaks(tt.text)
        if !reflect.DeepEqual(got.Graphemes, tt.wantG) {
            t.Errorf("graphemes mismatch")
        }
        if !reflect.DeepEqual(got.Words, tt.wantW) {
            t.Errorf("words mismatch")
        }
        if !reflect.DeepEqual(got.Sentences, tt.wantS) {
            t.Errorf("sentences mismatch")
        }
    }
}
```

### Property Hierarchy Tests

Verify break hierarchy:
```go
func TestBreakHierarchy(t *testing.T) {
    text := "The quick brown fox jumps. Over the lazy dog."
    breaks := FindAllBreaks(text)

    // All word breaks must be grapheme breaks
    for _, wb := range breaks.Words {
        if !contains(breaks.Graphemes, wb) {
            t.Errorf("word break %d not in grapheme breaks", wb)
        }
    }

    // All sentence breaks must be word breaks
    for _, sb := range breaks.Sentences {
        if !contains(breaks.Words, sb) {
            t.Errorf("sentence break %d not in word breaks", sb)
        }
    }
}
```

## Success Criteria

- ✅ 100% conformance on all official Unicode tests (766 + 1,944 + 512 = 3,222 tests)
- ✅ 5-10× faster than current implementation
- ✅ 30-50% less memory usage
- ✅ Backward compatible API
- ✅ Single-pass API available for optimal performance
- ✅ Generated data tables (easy to update)
- ✅ Code is more maintainable (data-driven)

## Future Optimizations

### SIMD Classification

The packed class structure is SIMD-friendly:
```
Input:  [r0, r1, r2, r3, r4, r5, r6, r7]  (8 runes)
Output: [c0, c1, c2, c3, c4, c5, c6, c7]  (8 packed classes)
```

AVX2 can process 8 lookups in parallel using `vpgatherdd`.

### State Machine Compilation

After this refactoring, the rule-based break detection can be compiled to FSM (see ARCHITECTURE.md Phase 2) for an additional 2-3× speedup.

### String Interning for Splits

When returning `[]string`, we could intern common words/graphemes to reduce allocations.

## References

- [UAX #29: Text Segmentation](https://www.unicode.org/reports/tr29/)
- [GraphemeBreakProperty.txt](https://www.unicode.org/Public/17.0.0/ucd/auxiliary/GraphemeBreakProperty.txt)
- [WordBreakProperty.txt](https://www.unicode.org/Public/17.0.0/ucd/auxiliary/WordBreakProperty.txt)
- [SentenceBreakProperty.txt](https://www.unicode.org/Public/17.0.0/ucd/auxiliary/SentenceBreakProperty.txt)

---

**Document Status**: Ready for Implementation
**Authors**: Design discussion with @SCKelemen
**Last Updated**: 2025-12-16
