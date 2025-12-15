# Unicode Break Algorithm Architecture

**Status**: Design Document
**Version**: 1.0
**Date**: 2025-12-15

## Executive Summary

This document proposes a refactoring of all Unicode break algorithms (UAX #14, UAX #29) from ad-hoc if/else chains to a unified **state machine architecture** with the following goals:

- **Correctness**: Achieve 100% conformance on all official Unicode tests
- **Readability**: One function per Unicode rule, named and documented
- **Maintainability**: Easy to update for new Unicode versions
- **Performance**: 2-3x speedup (pure Go), 10-20x with SIMD (future)

## Current Problems

### 1. Unreadable Complex Logic

```go
// Current UAX #14 code (line ~4450)
if isClassOrVariant(lastNonSpaceClass, ClassOP) || lastNonSpaceClass == ClassQU_Pi {
    // Don't break - we're in "OP SP*" or "QU_Pi SP*" sequence
} else if (isClassOrVariant(lastNonSpaceClass, ClassCL) || lastNonSpaceClass == ClassCP) &&
    isClassOrVariant(currClass, ClassNS) {
    // LB16: Don't break in "(CL | CP) SP* × NS" sequence
```

**Problems**:
- Rules buried in nested if/else
- Hard to match against Unicode spec
- State tracking variables scattered throughout
- Manual index arithmetic for lookahead/lookback

### 2. Performance Issues

```go
// Current UAX #29 grapheme.go (line ~334)
j := i - 1
for j >= 0 && getGraphemeBreakClass(runes[j]) == GBExtend {
    j--
}
if j >= 0 && getGraphemeBreakClass(runes[j]) == GBZWJ {
    j--
    for j >= 0 && getGraphemeBreakClass(runes[j]) == GBExtend {
        j--
    }
    if j >= 0 && isExtendedPictographic(runes[j]) {
```

**Problems**:
- Sequential if/else: 20-30 branches per character
- Manual lookback loops: cache unfriendly
- Multiple property lookups per character
- Impossible to vectorize

### 3. Multi-Property Confusion

**LB30 Bug**: Extended_Pictographic is an emoji property (UTS #51), not a line break class (UAX #14). Current code tries to check it via line break class, causing misclassification.

Example: U+1FFFD is Extended_Pictographic but has line break class `ID`, not `EB`.

---

## Proposed Architecture

### Phase 1: Rule-Based State Machine (Pure Go)

**Goals**: 2-3x performance, 100% conformance, readable code

#### 1.1 Break Context Abstraction

Centralized state management for all break algorithms:

```go
// BreakContext manages state for break detection algorithms
type BreakContext struct {
    // Input data (immutable)
    runes []rune
    text  string

    // Primary property (e.g., LineBreakClass, GraphemeBreakClass)
    classes []uint64  // Bitflags for fast pattern matching

    // Secondary properties (for cross-spec rules)
    emojiData []EmojiDataClass  // Extended_Pictographic, etc.
    eaWidth   []EastAsianWidth  // For _EA variants

    // Position tracking
    pos int

    // Cached lookups (updated on Slide())
    prev, curr, next uint64

    // Algorithm-specific state
    state interface{}  // Polymorphic: LineBreakState, GraphemeBreakState, etc.
}

// Navigation methods
func (c *BreakContext) Slide() bool
func (c *BreakContext) BytePos() int

// Lookahead/lookback through ignorable classes
func (c *BreakContext) PeekBehind(ignoreMask uint64) uint64
func (c *BreakContext) PeekAhead(ignoreMask uint64) uint64
func (c *BreakContext) PeekBehindFrom(pos int, ignoreMask uint64) (uint64, int)

// State accessors
func (c *BreakContext) Prev() uint64
func (c *BreakContext) Curr() uint64
func (c *BreakContext) Next() uint64
```

**Example usage**:

```go
ctx := NewGraphemeBreakContext(text)
for ctx.Slide() {
    if ctx.Curr() & GraphemeExtend != 0 {
        // Current is Extend class
    }

    prev := ctx.PeekBehind(GraphemeExtend | GraphemeFormat)
    // prev is the last non-Extend, non-Format class
}
```

#### 1.2 Algorithm-Specific State

```go
// LineBreakState tracks UAX #14 specific state
type LineBreakState struct {
    lastNonSpace      uint64  // For "OP SP* ×" patterns
    lastNonIgnorable  uint64  // Skip Format/CM/ZWJ
    riPairCount       int     // Regional Indicator pair tracking
}

// GraphemeBreakState tracks UAX #29 grapheme state
type GraphemeBreakState struct {
    // Minimal state needed for GB rules
    riPairCount int  // For GB12/GB13
}

// WordBreakState tracks UAX #29 word state
type WordBreakState struct {
    riPairCount int  // For WB15/WB16
}

// SentenceBreakState tracks UAX #29 sentence state
type SentenceBreakState struct {
    aTermSequence bool  // For SB8-SB11
    sTermSequence bool
}
```

#### 1.3 Bitflag Architecture

Each Unicode property type gets its own bitflag namespace:

```go
// LineBreakClass - UAX #14 Line Break property
type LineBreakClass uint64

const (
    // Mandatory breaks
    LB_BK LineBreakClass = 1 << 0  // Mandatory Break
    LB_CR LineBreakClass = 1 << 1  // Carriage Return
    LB_LF LineBreakClass = 1 << 2  // Line Feed
    LB_NL LineBreakClass = 1 << 3  // Next Line
    LB_SP LineBreakClass = 1 << 4  // Space

    // Prohibited breaks
    LB_WJ  LineBreakClass = 1 << 5  // Word Joiner
    LB_ZW  LineBreakClass = 1 << 6  // Zero Width Space
    LB_ZWJ LineBreakClass = 1 << 7  // Zero Width Joiner
    LB_GL  LineBreakClass = 1 << 8  // Non-breaking Glue

    // Break opportunities
    LB_BA LineBreakClass = 1 << 9   // Break After
    LB_BB LineBreakClass = 1 << 10  // Break Before
    LB_B2 LineBreakClass = 1 << 11  // Break Before and After
    LB_HY LineBreakClass = 1 << 12  // Hyphen
    LB_CB LineBreakClass = 1 << 13  // Contingent Break

    // Letters and ideographs
    LB_AL LineBreakClass = 1 << 14  // Alphabetic
    LB_HL LineBreakClass = 1 << 15  // Hebrew Letter
    LB_ID LineBreakClass = 1 << 16  // Ideographic

    // Punctuation
    LB_OP    LineBreakClass = 1 << 17  // Open Punctuation
    LB_CL    LineBreakClass = 1 << 18  // Close Punctuation
    LB_CP    LineBreakClass = 1 << 19  // Close Parenthesis
    LB_QU_Pi LineBreakClass = 1 << 20  // Quotation - Initial
    LB_QU_Pf LineBreakClass = 1 << 21  // Quotation - Final
    LB_NS    LineBreakClass = 1 << 22  // Nonstarter
    LB_EX    LineBreakClass = 1 << 23  // Exclamation/Interrogation

    // Numeric
    LB_NU LineBreakClass = 1 << 24  // Numeric
    LB_PR LineBreakClass = 1 << 25  // Prefix
    LB_PO LineBreakClass = 1 << 26  // Postfix
    LB_IS LineBreakClass = 1 << 27  // Infix Separator
    LB_SY LineBreakClass = 1 << 28  // Symbols

    // Emoji
    LB_EM LineBreakClass = 1 << 29  // Emoji Modifier
    LB_EB LineBreakClass = 1 << 30  // Emoji Base

    // Complex scripts
    LB_CM LineBreakClass = 1 << 31  // Combining Mark
    LB_SA LineBreakClass = 1 << 32  // Complex Context (Southeast Asian)
    LB_AK LineBreakClass = 1 << 33  // Aksara (Indic)
    LB_AP LineBreakClass = 1 << 34  // Aksara Prebase
    LB_AS LineBreakClass = 1 << 35  // Aksara Start
    LB_VI LineBreakClass = 1 << 36  // Virama
    LB_VF LineBreakClass = 1 << 37  // Virama Final

    // Hangul
    LB_JL LineBreakClass = 1 << 38  // Jamo L
    LB_JV LineBreakClass = 1 << 39  // Jamo V
    LB_JT LineBreakClass = 1 << 40  // Jamo T
    LB_H2 LineBreakClass = 1 << 41  // Hangul LV
    LB_H3 LineBreakClass = 1 << 42  // Hangul LVT

    // Regional Indicators
    LB_RI LineBreakClass = 1 << 43  // Regional Indicator

    // East Asian Width variants (set bit + base class)
    LB_EA_FLAG LineBreakClass = 1 << 62  // Marker for EA width

    // Grouping masks for pattern matching
    LB_MANDATORY    = LB_BK | LB_CR | LB_LF | LB_NL
    LB_SPACES       = LB_SP | LB_ZW
    LB_OPENING      = LB_OP | LB_QU_Pi
    LB_CLOSING      = LB_CL | LB_CP | LB_QU_Pf
    LB_NUMERIC_SET  = LB_NU | LB_PR | LB_PO | LB_IS | LB_SY
    LB_LETTER_SET   = LB_AL | LB_HL | LB_ID
    LB_IGNORABLE    = LB_CM | LB_ZWJ
)
```

```go
// EmojiDataClass - UTS #51 Emoji properties (separate namespace!)
type EmojiDataClass uint64

const (
    Emoji_ExtendedPictographic EmojiDataClass = 1 << 0
    Emoji_Emoji                EmojiDataClass = 1 << 1
    Emoji_EmojiPresentation    EmojiDataClass = 1 << 2
    Emoji_EmojiModifier        EmojiDataClass = 1 << 3
    Emoji_EmojiModifierBase    EmojiDataClass = 1 << 4
    Emoji_EmojiComponent       EmojiDataClass = 1 << 5
)
```

```go
// GraphemeBreakClass - UAX #29 Grapheme_Cluster_Break property
type GraphemeBreakClass uint64

const (
    GB_CR      GraphemeBreakClass = 1 << 0
    GB_LF      GraphemeBreakClass = 1 << 1
    GB_Control GraphemeBreakClass = 1 << 2
    GB_Extend  GraphemeBreakClass = 1 << 3
    GB_ZWJ     GraphemeBreakClass = 1 << 4
    GB_RI      GraphemeBreakClass = 1 << 5
    GB_Prepend GraphemeBreakClass = 1 << 6
    GB_SpacingMark GraphemeBreakClass = 1 << 7

    // Hangul
    GB_L   GraphemeBreakClass = 1 << 8
    GB_V   GraphemeBreakClass = 1 << 9
    GB_T   GraphemeBreakClass = 1 << 10
    GB_LV  GraphemeBreakClass = 1 << 11
    GB_LVT GraphemeBreakClass = 1 << 12

    // Grouping masks
    GB_IGNORABLE = GB_Extend | GB_ZWJ
    GB_HANGUL_L_SET = GB_L | GB_V | GB_LV | GB_LVT
    GB_HANGUL_V_SET = GB_V | GB_T
    GB_HANGUL_T_SET = GB_T
)
```

**Usage example**:

```go
// Instead of:
if isClassOrVariant(class, ClassOP) || class == ClassQU_Pi {

// Write:
if ctx.Curr() & LB_OPENING != 0 {
```

#### 1.4 Rule Function Interface

Every Unicode rule becomes a named function:

```go
// BreakAction represents what to do at current position
type BreakAction int

const (
    BreakProhibited BreakAction = iota  // × - Don't break
    BreakAllowed                        // ÷ - Break allowed
    BreakMandatory                      // ! - Must break
)

// BreakRule checks if a rule applies and returns the action
type BreakRule func(ctx *BreakContext) (matched bool, action BreakAction)
```

**Example rules**:

```go
// ruleLB3: Don't break within CRLF (CR × LF)
// https://www.unicode.org/reports/tr14/#LB3
func ruleLB3(ctx *BreakContext) (bool, BreakAction) {
    if ctx.Prev() & LB_CR != 0 && ctx.Curr() & LB_LF != 0 {
        return true, BreakProhibited
    }
    return false, BreakAllowed
}

// ruleLB4: Always break after hard line breaks (BK !)
// https://www.unicode.org/reports/tr14/#LB4
func ruleLB4(ctx *BreakContext) (bool, BreakAction) {
    if ctx.Prev() & LB_BK != 0 {
        return true, BreakMandatory
    }
    return false, BreakAllowed
}

// ruleLB14: Don't break after opening punctuation, even with spaces (OP SP* ×)
// https://www.unicode.org/reports/tr14/#LB14
func ruleLB14(ctx *BreakContext) (bool, BreakAction) {
    state := ctx.state.(*LineBreakState)
    if state.lastNonSpace & LB_OPENING != 0 {
        return true, BreakProhibited
    }
    return false, BreakAllowed
}

// ruleLB30: Don't break between emoji base and modifier (ExtPict × EM)
// https://www.unicode.org/reports/tr14/#LB30
// NOTE: Uses emoji property, not line break class!
func ruleLB30(ctx *BreakContext) (bool, BreakAction) {
    // Check Extended_Pictographic from emoji data (UTS #51)
    prevPos := ctx.pos - 1
    if prevPos >= 0 && ctx.emojiData[prevPos] & Emoji_ExtendedPictographic != 0 {
        // Check if current is Emoji Modifier (line break class)
        if ctx.Curr() & LB_EM != 0 {
            return true, BreakProhibited
        }
    }
    return false, BreakAllowed
}

// ruleGB11: Don't break emoji ZWJ sequences (ExtPict Extend* ZWJ × ExtPict)
// https://www.unicode.org/reports/tr29/#GB11
func ruleGB11(ctx *BreakContext) (bool, BreakAction) {
    // Current is ExtendedPictographic?
    if ctx.emojiData[ctx.pos] & Emoji_ExtendedPictographic == 0 {
        return false, BreakAllowed
    }

    // Look back through Extend* for ZWJ
    prevClass, prevPos := ctx.PeekBehindFrom(ctx.pos-1, GB_Extend)
    if prevPos < 0 || prevClass & GB_ZWJ == 0 {
        return false, BreakAllowed
    }

    // Look back through Extend* for ExtendedPictographic
    _, prevPrevPos := ctx.PeekBehindFrom(prevPos-1, GB_Extend)
    if prevPrevPos >= 0 && ctx.emojiData[prevPrevPos] & Emoji_ExtendedPictographic != 0 {
        return true, BreakProhibited
    }

    return false, BreakAllowed
}
```

#### 1.5 Main Loop Structure

```go
// Rule table - ordered by Unicode spec
var lineBreakRules = []BreakRule{
    ruleLB1,   // Assign break classes
    ruleLB2,   // Never break at start
    ruleLB3,   // Don't break within CRLF
    ruleLB4,   // Always break after hard breaks
    ruleLB5,   // Treat CR, LF, NL as BK
    ruleLB6,   // Don't break before hard breaks or spaces
    ruleLB7,   // Don't break before spaces or ZW
    ruleLB8,   // Break after ZW
    ruleLB8a,  // Don't break after ZWJ
    ruleLB9,   // Don't break combining character sequences
    ruleLB10,  // Treat CM/ZWJ as AL
    ruleLB11,  // Don't break before/after WJ
    ruleLB12,  // Don't break after GL
    ruleLB12a, // Don't break before GL unless space
    ruleLB13,  // Don't break before closing punctuation
    ruleLB14,  // Don't break after opening punctuation
    ruleLB15,  // Don't break within quotation marks
    ruleLB15a, // Don't break after QU_Pi
    ruleLB16,  // Don't break between closing and nonstarter
    ruleLB17,  // Don't break within B2 sequences
    ruleLB18,  // Break after spaces
    ruleLB19,  // Don't break before/after quotation marks
    ruleLB20,  // Break before/after unresolved CB
    ruleLB21,  // Don't break before hyphen
    ruleLB21a, // Don't break after Hebrew Letter + hyphen
    ruleLB21b, // Don't break between SY and HL
    ruleLB22,  // Don't break before IN
    ruleLB23,  // Don't break between letters and numbers
    ruleLB23a, // Don't break between PR/PO and letters
    ruleLB24,  // Don't break between numeric prefix/postfix and letters
    ruleLB25,  // Don't break within numeric expressions
    ruleLB26,  // Don't break Hangul syllable sequences
    ruleLB27,  // Treat JL/JV/JT/H2/H3 as ID before PO
    ruleLB28,  // Don't break between alphabetics
    ruleLB29,  // Don't break between IS and alphabetics
    ruleLB30,  // Don't break between emoji base and modifier
    ruleLB30a, // Break between two RIs if odd number before
    ruleLB30b, // Don't break emoji sequences
    ruleLB31,  // Break everywhere else
}

func FindLineBreakOpportunities(text string, hyphens Hyphens) []int {
    ctx := NewLineBreakContext(text, hyphens)
    breaks := []int{0}  // LB2: Always break at start

    for ctx.Slide() {
        // Apply rules in order - first match wins
        action := BreakAllowed  // Default: LB31

        for _, rule := range lineBreakRules {
            if matched, ruleAction := rule(ctx); matched {
                action = ruleAction
                break  // Stop at first matching rule
            }
        }

        // Add break point if allowed or mandatory
        if action == BreakAllowed || action == BreakMandatory {
            breaks = append(breaks, ctx.BytePos())
        }

        // Update state for next iteration
        ctx.UpdateState()
    }

    // LB3: Always break at end
    breaks = append(breaks, len(text))
    return breaks
}
```

### Phase 2: FSM Compilation (Future Optimization)

**Goal**: 5-10x performance by compiling rules to state transition tables

#### 2.1 State Transition Table

```go
// Precompile rules into state machine at init time
type StateTransition struct {
    nextState uint8
    action    BreakAction
}

// State machine: current state × input class → next state + action
var lineBreakFSM [256][64]StateTransition

func init() {
    // Compile rules to FSM
    compileRulesToFSM(lineBreakRules, &lineBreakFSM)
}

func FindLineBreakOpportunitiesFSM(text string) []int {
    classes := classifyRunes(text)
    breaks := []int{0}
    state := uint8(0)

    for i, class := range classes {
        trans := lineBreakFSM[state][class]
        state = trans.nextState

        if trans.action == BreakAllowed || trans.action == BreakMandatory {
            breaks = append(breaks, bytePos(text, i))
        }
    }

    return breaks
}
```

#### 2.2 Optimization: Loop Unrolling

```go
// Process 4 characters per iteration
func FindLineBreakOpportunitiesFSM(text string) []int {
    classes := classifyRunes(text)
    breaks := []int{0}
    state := uint8(0)

    // Process 4 at a time
    for i := 0; i < len(classes)-3; i += 4 {
        trans0 := lineBreakFSM[state][classes[i]]
        trans1 := lineBreakFSM[trans0.nextState][classes[i+1]]
        trans2 := lineBreakFSM[trans1.nextState][classes[i+2]]
        trans3 := lineBreakFSM[trans2.nextState][classes[i+3]]

        if trans0.action != BreakProhibited { breaks = append(breaks, bytePos(text, i)) }
        if trans1.action != BreakProhibited { breaks = append(breaks, bytePos(text, i+1)) }
        if trans2.action != BreakProhibited { breaks = append(breaks, bytePos(text, i+2)) }
        if trans3.action != BreakProhibited { breaks = append(breaks, bytePos(text, i+3)) }

        state = trans3.nextState
    }

    // Handle remainder
    // ...
}
```

### Phase 3: SIMD Assembly (Maximum Performance)

**Goal**: 10-20x performance using vectorized operations

#### 3.1 Vectorized Class Lookup

```asm
// asm_amd64.s - AVX2 version
// func classifyRunesAVX2(runes []rune) []uint64

TEXT ·classifyRunesAVX2(SB),NOSPLIT,$0-48
    MOVQ runes+0(FP), SI    // Source pointer
    MOVQ len+8(FP), CX      // Length
    MOVQ classes+24(FP), DI // Destination pointer

    XORQ AX, AX             // Index counter

loop:
    CMPQ AX, CX
    JGE done

    // Load 8 runes (256 bits)
    VMOVDQU (SI)(AX*4), Y0

    // Parallel table lookups (simplified)
    // ... vectorized binary search in generated data

    // Store results
    VMOVDQU Y0, (DI)(AX*8)

    ADDQ $8, AX
    JMP loop

done:
    RET
```

#### 3.2 Runtime CPU Detection

```go
import "golang.org/x/sys/cpu"

func FindLineBreakOpportunities(text string) []int {
    // Use fastest available implementation
    if cpu.X86.HasAVX2 {
        return findLineBreakOpportunitiesAVX2(text)
    } else if cpu.ARM64.HasNEON {
        return findLineBreakOpportunitiesNEON(text)
    }

    // Fallback to FSM or rule-based
    return findLineBreakOpportunitiesFSM(text)
}

//go:noescape
func findLineBreakOpportunitiesAVX2(text string) []int

//go:noescape
func findLineBreakOpportunitiesNEON(text string) []int
```

---

## Migration Plan

### Step 1: Create Architecture Branch

```bash
git checkout -b refactor/state-machine
```

### Step 2: Implement Core Abstractions

1. Create `internal/breakcontext/` package with:
   - `BreakContext` interface and base implementation
   - Property-specific context types
   - Navigation and state management methods

2. Create bitflag definitions in each spec package:
   - `uax14/classes.go` - LineBreakClass bitflags
   - `uax29/grapheme_classes.go` - GraphemeBreakClass bitflags
   - `uax29/word_classes.go` - WordBreakClass bitflags
   - `uax29/sentence_classes.go` - SentenceBreakClass bitflags
   - `uts51/emoji_classes.go` - EmojiDataClass bitflags

### Step 3: Convert Data Generators

Update data generators to emit bitflag values instead of iota enums:

```go
// OLD: generate_grapheme_data.go
case "CR":
    class = "GBCR"  // iota enum

// NEW: generate_grapheme_data.go
case "CR":
    class = "GB_CR"  // bitflag constant
```

### Step 4: Implement Rules (One Spec at a Time)

#### Phase A: UAX #29 Grapheme

1. Create `uax29/grapheme_rules.go` with all GB rules as functions
2. Implement `GraphemeBreakContext` extending `BreakContext`
3. Update `FindGraphemeBreaks()` to use rule-based loop
4. **Verify**: Run conformance tests, ensure 100% pass rate
5. **Benchmark**: Compare performance vs old implementation

#### Phase B: UAX #29 Word

1. Create `uax29/word_rules.go` with all WB rules
2. Implement `WordBreakContext`
3. Update `FindWordBreaks()`
4. Verify conformance and benchmark

#### Phase C: UAX #29 Sentence

1. Create `uax29/sentence_rules.go` with all SB rules
2. Implement `SentenceBreakContext`
3. Update `FindSentenceBreaks()`
4. Verify conformance and benchmark

#### Phase D: UAX #14 Line Break

1. Create `uax14/linebreak_rules.go` with all LB rules (30+ rules)
2. Implement `LineBreakContext`
3. Fix LB30 Extended_Pictographic bug using emoji data
4. Update `FindLineBreakOpportunities()`
5. Verify conformance and benchmark

### Step 5: FSM Compilation (Future)

1. Create `internal/fsm/` package
2. Implement rule → FSM compiler
3. Add `FindXxxBreaksFSM()` variants
4. Benchmark and compare

### Step 6: SIMD Assembly (Future)

1. Create `asm_amd64.s` and `asm_arm64.s`
2. Implement vectorized class lookup
3. Implement parallel state transitions
4. Add CPU feature detection
5. Benchmark on different CPUs

---

## Testing Strategy

### Conformance Tests

All existing conformance tests must pass:

```bash
# UAX #29
go test ./uax29 -run TestGraphemeBreakOfficial  # 766 tests
go test ./uax29 -run TestWordBreakOfficial      # 1,944 tests
go test ./uax29 -run TestSentenceBreakOfficial  # 512 tests

# UAX #14
go test ./uax14 -run TestLineBreakOfficial      # TBD tests
```

**Success criteria**: 100% pass rate on all official Unicode tests.

### Performance Benchmarks

```bash
# Benchmark each implementation
go test ./uax29 -bench=BenchmarkGrapheme -benchmem
go test ./uax14 -bench=BenchmarkLineBreak -benchmem

# Compare implementations
go test ./uax29 -bench=. -benchmem > old.txt
# ... after refactor
go test ./uax29 -bench=. -benchmem > new.txt
benchstat old.txt new.txt
```

**Expected results**:
- Phase 1 (rules): 2-3x faster
- Phase 2 (FSM): 5-10x faster
- Phase 3 (SIMD): 10-20x faster

### Regression Tests

Ensure all existing tests still pass:

```bash
go test ./... -v
```

---

## Performance Expectations

### Current Baseline (Measured)

From actual benchmark on Apple M4 Pro:

```
BenchmarkIfElse-14    30426    38449 ns/op    128248 B/op    16 allocs/op
```

### Phase 1: Rule-Based (Projected)

Based on FSM benchmark results:

```
BenchmarkFSM-14       42678    26612 ns/op     87288 B/op    15 allocs/op
```

**Improvement**: 1.44x faster, 32% less memory

For real Unicode algorithms (20-30 rules, complex state):
- **Conservative estimate**: 2-3x faster
- **Memory**: 30-40% reduction
- **Code size**: Similar (trades if/else for functions)

### Phase 2: FSM Compiled (Projected)

Based on literature and similar projects:

```
BenchmarkCompiledFSM-14    200000    5000 ns/op    20000 B/op    1 allocs/op
```

**Improvement**: 5-10x faster than baseline
- Single state table lookup per character
- No function call overhead
- Better cache locality
- Predictable branching

### Phase 3: SIMD Assembly (Projected)

Based on other Go projects (minio/sha256-simd, klauspost/compress):

```
BenchmarkSIMD-14    1000000    2000 ns/op    10000 B/op    1 allocs/op
```

**Improvement**: 10-20x faster than baseline
- Process 8-16 characters per instruction
- Vectorized table lookups
- Cache-aligned data structures
- Minimal branching

---

## Risk Analysis

### Risks

1. **Complexity**: Rule-based approach adds function call overhead
   - **Mitigation**: Go's inliner is good, benchmark early

2. **FSM compilation**: Some rules may not be FSM-compatible (context-dependent)
   - **Mitigation**: Hybrid approach - FSM for simple rules, functions for complex ones

3. **SIMD maintenance**: Assembly is hard to maintain across Go versions
   - **Mitigation**: Keep pure Go fallback, add comprehensive tests

4. **Debugging**: FSM bugs are harder to trace than if/else
   - **Mitigation**: Keep rule functions as reference implementation

### Benefits vs Current Code

| Metric | Current | Phase 1 | Phase 2 | Phase 3 |
|--------|---------|---------|---------|---------|
| **Conformance** | 99.6% (UAX#29), <100% (UAX#14) | 100% | 100% | 100% |
| **Performance** | 1x | 2-3x | 5-10x | 10-20x |
| **Readability** | Poor | Excellent | Good | Good |
| **Maintainability** | Poor | Excellent | Good | Fair |
| **Debuggability** | Fair | Excellent | Good | Fair |

---

## Open Questions

1. **Rule ordering**: Should we strictly follow UAX spec order, or optimize for common cases first?
   - **Proposal**: Follow spec order for correctness, optimize later if needed

2. **State machine size**: Will compiled FSM fit in cache? (256 states × 64 classes = 16KB)
   - **Proposal**: Measure cache hit rates, consider compressed FSM if needed

3. **Cross-property rules**: How to handle rules that need multiple properties (e.g., LB30, GB11)?
   - **Proposal**: Store multiple property arrays in BreakContext, rules access both

4. **Unicode version updates**: How to make updates easy?
   - **Proposal**: Keep data generators, one git commit per Unicode version

5. **Backward compatibility**: Can we maintain the same public API?
   - **Proposal**: Yes - internal refactor only, external API unchanged

---

## Next Steps

1. **Review this document** - Gather feedback on architecture
2. **Create feature branch** - `refactor/state-machine`
3. **Implement core abstractions** - BreakContext, bitflags
4. **Prototype one spec** - UAX #29 grapheme as proof of concept
5. **Benchmark and validate** - Ensure performance and conformance goals met
6. **Iterate** - Apply learnings to other specs

---

## Prior Art: SIMD Unicode Research

### Industry Implementation: simdutf

**[simdutf](https://github.com/simdutf/simdutf)** is the leading SIMD-optimized Unicode library:

- **Performance**: Processes 1+ billion characters per second
- **Platforms**: SSE2, AVX2, AVX-512, NEON, RISC-V Vector Extension, LoongArch64, POWER
- **Adoption**: Used in Node.js, WebKit/Safari, Ladybird, Chromium, Cloudflare Workers, Bun
- **vs ICU**: 3-10x faster on non-ASCII strings, 20x faster on ASCII strings

[Academic research](https://onlinelibrary.wiley.com/doi/full/10.1002/spe.3261) using AVX-512 demonstrates:
- 4+ GB/s transcoding Chinese/Emoji text
- 8 GB/s on Arabic text (UTF-8)
- 20+ GB/s on Arabic text (UTF-16)

**Key Techniques**:
- Parallel table lookups (16 bytes at once)
- Bit manipulation for validation
- [Validation in 0.45 cycles/byte](https://lemire.me/blog/2023/09/13/transcoding-unicode-strings-at-crazy-speeds-with-avx-512/)

**Limitation**: Focuses on **UTF-8/UTF-16 transcoding**, not break algorithms

### Rust unicode-segmentation

**[unicode-segmentation](https://github.com/unicode-rs/unicode-segmentation)** implements UAX #29 in Rust:

- **No SIMD**: Uses table lookups + aggressive inlining
- **Optimizations**: 15-40% speedup from inlining, ASCII fast-paths
- **Approach**: Iterator-based, rule-driven (similar to our current code)

**Observation**: Even Rust's ecosystem hasn't SIMD-ified break algorithms yet.

### ICU Reference Implementation

**[ICU](https://icu.unicode.org/)** is the Unicode reference implementation:

- Uses **state machine approach** for break detection
- [Rule-based configuration](https://github.com/unicode-org/icu/blob/main/icu4c/source/data/brkitr/rules/line.txt)
- No public documentation of SIMD optimization for break algorithms
- Optimized for **correctness over speed**

**Validation**: ICU's use of state machines validates our Phase 2 FSM approach.

### Research Gap

**Nobody has published SIMD-optimized break algorithms**:
- Transcoding (UTF-8 ↔ UTF-16): **Extensively optimized** (simdutf, academic papers)
- Break detection (UAX #14, #29): **Unexplored territory**

**Why it's harder**:
1. **Context-dependent**: Rules like "OP SP* ×" need lookback through ignorable classes
2. **Multi-property**: LB30 needs emoji data (UTS #51) + line break class (UAX #14)
3. **State tracking**: Regional indicator pairs, ATerm sequences, etc.
4. **Complex patterns**: Not simple byte-to-byte mappings

**Our Opportunity**:
Our FSM architecture makes SIMD optimization **structurally possible**:

```
Transcoding (simdutf):          Break Detection (our approach):
───────────────────────         ─────────────────────────────────
Input: UTF-8 bytes              Input: UTF-8 text
  ↓ SIMD validation               ↓ UTF-8 → runes
  ↓ SIMD conversion               ↓ SIMD class lookup (Phase 3)
Output: UTF-16                    ↓ SIMD FSM transitions
                                  ↓ Filter break points
                                Output: Break positions
```

Both follow: **Input → Classify → Transform → Output**

### Inspiration for Our Phase 3

From simdutf's techniques, we can adapt:

1. **Vectorized lookups**: Process 8-16 classes per cycle
   ```nasm
   VPSHUFB table, classes, results  ; Parallel table lookup
   ```

2. **Parallel state transitions**: Maintain 8 FSM states simultaneously
   ```nasm
   VPCMPGTB curr_class, threshold, mask  ; Vectorized comparison
   VPBLENDVB state_a, state_b, mask, next_states  ; Branch-free select
   ```

3. **Bit manipulation for break detection**:
   ```nasm
   VPMOVMSKB break_flags, break_mask  ; Extract break bits to register
   TZCNT break_mask, first_break      ; Find first break position
   ```

4. **Cache-aligned data structures**:
   ```go
   type alignedTransitionTable struct {
       _ [64]byte  // Force cache alignment
       table [256][64]StateTransition
   }
   ```

**Expected Performance** (based on simdutf's gains):
- Current: ~1 char/ns (1 GHz = 1 billion chars/sec)
- Phase 3 SIMD: ~10-15 chars/ns (10-15 billion chars/sec)

This would match simdutf's performance class.

---

## References

### Unicode Specifications
- [UAX #14: Line Breaking](https://www.unicode.org/reports/tr14/)
- [UAX #29: Text Segmentation](https://www.unicode.org/reports/tr29/)
- [UTS #51: Emoji](https://www.unicode.org/reports/tr51/)

### SIMD Unicode Implementations
- [simdutf](https://simdutf.github.io/simdutf/) - SIMD UTF-8/UTF-16 transcoding
- [simdutf GitHub](https://github.com/simdutf/simdutf) - Source code and benchmarks
- [Daniel Lemire's blog](https://lemire.me/blog/2023/09/13/transcoding-unicode-strings-at-crazy-speeds-with-avx-512/) - AVX-512 Unicode optimization
- [Academic paper](https://onlinelibrary.wiley.com/doi/full/10.1002/spe.3261) - "Transcoding unicode characters with AVX‐512 instructions"

### Unicode Reference Implementations
- [ICU](https://icu.unicode.org/) - International Components for Unicode
- [Rust unicode-segmentation](https://github.com/unicode-rs/unicode-segmentation) - UAX #29 in Rust

### Go Assembly and SIMD
- [Go assembler guide](https://go.dev/doc/asm)
- [minio/sha256-simd](https://github.com/minio/sha256-simd) - SIMD hashing in Go
- [klauspost/compress](https://github.com/klauspost/compress) - SIMD compression in Go
- [rusticstuff/simdutf8](https://github.com/rusticstuff/simdutf8) - SIMD UTF-8 validation in Rust

---

**Document Status**: Ready for Review
**Authors**: Architecture discussion with @SCKelemen
**Last Updated**: 2025-12-15
