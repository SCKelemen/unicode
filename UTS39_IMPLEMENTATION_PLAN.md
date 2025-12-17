# UTS #39 Implementation Plan

## Overview

This document outlines the implementation strategy for **UTS #39 (Unicode Security Mechanisms)** and its dependencies in the unicode repository.

## Goal

Achieve **100% conformance** with UTS #39 and all required dependencies, following the established patterns in the codebase.

## Standards Hierarchy

UTS #39 depends on several other Unicode standards:

```
UTS #39 (Unicode Security Mechanisms)
├── UAX #31 (Identifier and Pattern Syntax) - XID_Start/XID_Continue properties
├── UAX #24 (Script Property) - Script and Script_Extensions properties
├── UAX #15 (Unicode Normalization Forms) - NFD, NFC, NFKD, NFKC
└── Unicode Character Database (UCD) - Core character properties
```

### Package Structure
Each dependency will be implemented in its own package:
- `uax31/` - Identifier and Pattern Syntax
- `uax24/` - Script Property
- `uax15/` - Unicode Normalization Forms
- `uts39/` - Unicode Security Mechanisms (main package)

## Implementation Strategy

### Phase 1: Dependencies (UAX #31, UAX #24, UAX #15)

We'll implement the core dependencies first, each in its own package following the established patterns.

### Phase 2: UTS #39 Security Mechanisms

Once dependencies are in place, implement the security mechanisms.

### Phase 3: Integration and Testing

Comprehensive testing against official Unicode conformance test suites.

---

## Package 1: UAX #31 (Identifier and Pattern Syntax)

**Package:** `uax31/`
**Branch:** `uax31`

### Purpose
Defines which Unicode characters can appear in identifiers (programming variables, usernames, etc.).

### Key Properties
- **XID_Start**: Characters valid at the start of an identifier
- **XID_Continue**: Characters valid in the rest of an identifier
- **Pattern_Syntax**: Characters reserved for syntax
- **Pattern_White_Space**: Characters treated as whitespace

### Data Files Required
From https://www.unicode.org/Public/17.0.0/ucd/:
- `DerivedCoreProperties.txt` - Contains XID_Start and XID_Continue
- `PropList.txt` - Contains Pattern_Syntax and Pattern_White_Space

### API Design
```go
package uax31

// Property checks
func IsXIDStart(r rune) bool
func IsXIDContinue(r rune) bool
func IsPatternSyntax(r rune) bool
func IsPatternWhiteSpace(r rune) bool

// Identifier validation
func IsValidIdentifier(s string) bool
func IsValidIdentifierStart(r rune) bool
```

### Implementation Steps
1. Create `uax31/` directory
2. Download data files (DerivedCoreProperties.txt, PropList.txt)
3. Create generator (`generate_identifier_data.go`)
4. Generate compiled data (`identifier_data.go`)
5. Implement API (`uax31.go`)
6. Add tests (examples, unit tests)
7. Add README with conformance status

### Estimated Complexity
**Low-Medium** - Similar to uax11 or uax50, primarily property lookups.

---

## Package 2: UAX #24 (Script Property)

**Package:** `uax24/`
**Branch:** `uax24`

### Purpose
Identifies which script(s) a character belongs to (Latin, Cyrillic, Han, etc.) and provides script extension information for characters used across multiple scripts.

### Key Properties
- **Script**: Single script value per character
- **Script_Extensions**: Set of scripts a character can be used with

### Data Files Required
From https://www.unicode.org/Public/17.0.0/ucd/:
- `Scripts.txt` - Maps code points to Script values
- `ScriptExtensions.txt` - Lists script extension sets
- `PropertyValueAliases.txt` - Script names and aliases

### API Design
```go
package uax24

type Script string

const (
    ScriptLatin    Script = "Latn"
    ScriptCyrillic Script = "Cyrl"
    ScriptHan      Script = "Hani"
    ScriptCommon   Script = "Zyyy"
    ScriptInherited Script = "Zinh"
    // ... all ~150 scripts
)

// Property lookups
func GetScript(r rune) Script
func GetScriptExtensions(r rune) []Script
func HasScript(r rune, script Script) bool

// Script resolution for mixed text
func ResolveScript(r rune, context []Script) Script
func GetTextScripts(s string) []Script
```

### Implementation Steps
1. Create `uax24/` directory
2. Download data files (Scripts.txt, ScriptExtensions.txt, PropertyValueAliases.txt)
3. Create generator for script data
4. Generate compiled data with binary search tables
5. Implement API with script resolution logic
6. Add tests and examples
7. Add README

### Estimated Complexity
**Medium** - More complex than simple property lookups due to Script_Extensions and resolution logic.

---

## Package 3: UAX #15 (Unicode Normalization Forms)

**Package:** `uax15/`
**Branch:** `uax15`

### Purpose
Provides Unicode normalization to convert between equivalent character representations. Essential for text comparison, searching, and security (confusable detection).

### The Four Normalization Forms
- **NFD** (Normalization Form D): Canonical Decomposition
- **NFC** (Normalization Form C): Canonical Decomposition + Canonical Composition
- **NFKD** (Normalization Form KD): Compatibility Decomposition
- **NFKC** (Normalization Form KC): Compatibility Decomposition + Canonical Composition

### Data Files Required
From https://www.unicode.org/Public/17.0.0/ucd/:
- `UnicodeData.txt` - Decomposition mappings and combining classes
- `CompositionExclusions.txt` - Characters excluded from composition
- `DerivedNormalizationProps.txt` - Quick_Check properties
- `NormalizationTest.txt` - Official conformance tests

### Core Algorithms

1. **Decomposition**: Recursively apply Decomposition_Mapping
2. **Canonical Ordering**: Sort combining marks by Canonical_Combining_Class
3. **Canonical Composition**: Pair and replace with precomposed forms
4. **Quick Check**: Fast detection if text is already normalized

### API Design
```go
package uax15

// Normalization forms
type Form int

const (
    NFD Form = iota
    NFC
    NFKD
    NFKC
)

// String normalization
func (f Form) String(s string) string
func (f Form) Bytes(b []byte) []byte

// Quick check
func (f Form) QuickCheck(s string) bool
func (f Form) IsNormalized(s string) bool

// Streaming API for large text
type Normalizer struct { /* ... */ }
func (f Form) NewNormalizer() *Normalizer
```

### Implementation Steps
1. Create `uax15/` directory
2. Download data files (UnicodeData.txt, CompositionExclusions.txt, etc.)
3. Create generator for decomposition tables
4. Create generator for composition tables
5. Implement decomposition algorithm
6. Implement canonical ordering algorithm
7. Implement composition algorithm
8. Implement quick check optimization
9. Test against NormalizationTest.txt (100% conformance)
10. Add README with examples

### Special Considerations
- **Hangul syllables**: Special algorithmic decomposition/composition
- **Composition exclusions**: Prevent certain compositions for stability
- **Combining marks**: Proper ordering by canonical combining class
- **Stream-safe**: Handle text in chunks without breaking sequences

### Estimated Complexity
**High** - Complex algorithms with multiple stages, but well-documented in spec.

### Alternative
We could use `golang.org/x/text/unicode/norm` as a dependency, but implementing `uax15/` maintains consistency with the repository pattern and ensures 100% control over Unicode 17.0.0 support.

---

## Package 4: UTS #39 (Unicode Security Mechanisms)

**Package:** `uts39/`
**Branch:** `uts39` (or `text-security`)

### Purpose
Provides mechanisms to detect and prevent security issues arising from Unicode's large character repertoire, including:
- Confusable character detection (spoofing prevention)
- Mixed-script detection
- Identifier restriction levels
- Security profiles

### Dependencies
- `uax31` - Identifier properties
- `uax24` - Script properties
- `uax15` - Unicode normalization

### Data Files Required
From https://www.unicode.org/Public/security/17.0.0/:
- `confusables.txt` - Visual similarity mappings
- `IdentifierStatus.txt` - Character restriction status
- `IdentifierType.txt` - Character type classifications

### Core Algorithms

#### 1. Confusable Detection
Implements the skeleton algorithm:
```
skeleton(X) = toNFD(toCaseFold(toNFKD(X)))
confusable(X, Y) = (skeleton(X) == skeleton(Y))
```

#### 2. Mixed-Script Detection
Determines if an identifier mixes scripts in suspicious ways:
- Highly Restrictive: Single script (+ Common + Inherited)
- Moderately Restrictive: Multiple scripts with specific rules
- Minimally Restrictive: Latin + one other script
- Unrestricted: Any combination

#### 3. Restriction Levels
Classifies identifiers by allowed character sets:
- ASCII-Only
- Single Script (Highly Restrictive)
- Multi-Script (Moderately Restrictive)
- Latin + Other (Minimally Restrictive)
- Unrestricted

### API Design
```go
package uts39

import (
    "github.com/SCKelemen/unicode/uax24"
    "github.com/SCKelemen/unicode/uax31"
)

// Restriction levels
type RestrictionLevel int

const (
    ASCIIOnly RestrictionLevel = iota
    SingleScript
    HighlyRestrictive
    ModeratelyRestrictive
    MinimallyRestrictive
    Unrestricted
)

// Confusable detection
func GetSkeleton(s string) string
func AreConfusable(s1, s2 string) bool
func GetConfusablePrototype(s string) string

// Mixed-script detection
func GetRestrictionLevel(s string) RestrictionLevel
func IsMixedScript(s string) bool
func GetIdentifierScripts(s string) []uax24.Script

// Security profiles
func IsAllowedIdentifier(s string, level RestrictionLevel) bool
func GetSecurityIssues(s string) []SecurityIssue

type SecurityIssue struct {
    Type        IssueType
    Position    int
    Character   rune
    Description string
}

type IssueType int

const (
    IssueConfusable IssueType = iota
    IssueMixedScript
    IssueRestricted
    IssueInvisible
)
```

### Implementation Steps
1. Create `uts39/` directory
2. Download security data files (confusables.txt, IdentifierStatus.txt, IdentifierType.txt)
3. Create confusables generator
4. Generate confusables data structures
5. Implement skeleton algorithm
6. Implement confusable detection
7. Implement mixed-script detection
8. Implement restriction level classification
9. Comprehensive testing against official test data
10. Add README with security guidelines

### Estimated Complexity
**High** - Most complex package due to multiple algorithms and security considerations.

---

## Branch Strategy

Following the user's requirements:

### Development Workflow

1. **Create branches off `develop`:**
   ```
   develop
   ├── uax31 (UAX #31 implementation)
   ├── uax24 (UAX #24 implementation)
   ├── uax15 (UAX #15 implementation)
   └── text-security (UTS #39 implementation, depends on uax31 + uax24 + uax15)
   ```

2. **Work on each standard independently:**
   - Each standard gets its own branch
   - Each branch is fully tested and conformant before merging
   - Merge to `develop` when complete

3. **Merge strategy to `develop`:**
   - Merge `uax31` → `develop`
   - Merge `uax24` → `develop`
   - Merge `uax15` → `develop`
   - Merge `text-security` → `develop`
   - All commits preserved in develop branch

4. **Clean history for `main`:**
   - Use rebase and cherry-pick to create clean, logical commits on main
   - Each standard gets a clear, concise commit (like existing v2.0.0, v3.0.0, v4.0.0 commits)
   - Maintain the clean history pattern visible in current main branch

### Example Main Branch History (after completion)
```
main
├── ac2f800 Add v4.0.0: Rule-based state machine architecture
├── ...existing commits...
├── [new] Add UAX #31: Identifier and Pattern Syntax
├── [new] Add UAX #24: Script Property
├── [new] Add UAX #15: Unicode Normalization Forms
├── [new] Add UTS #39: Unicode Security Mechanisms
```

---

## Testing Strategy

Each package must achieve **100% conformance** with official Unicode test suites:

### UAX #31
- Property lookups verified against DerivedCoreProperties.txt
- Identifier validation test cases
- Edge cases (emoji, combining marks, etc.)

### UAX #24
- Script property verification against Scripts.txt
- Script_Extensions verification against ScriptExtensions.txt
- Script resolution algorithm tests
- Mixed-script text handling

### UTS #39
- Official confusables test data (if available)
- Skeleton algorithm verification
- Mixed-script detection test cases
- Restriction level classification tests
- Security issue detection tests

### Integration Tests
- Test UTS #39 using UAX #31 and UAX #24
- End-to-end security checks
- Real-world identifier examples (domain names, usernames, etc.)

---

## Documentation Requirements

Each package must include:

1. **README.md** with:
   - Purpose and specification reference
   - Status (% conformance)
   - Supported features
   - Code examples
   - Installation instructions

2. **API documentation** (godoc):
   - Clear function descriptions
   - Parameter explanations
   - Return value semantics
   - Usage examples

3. **Top-level README update**:
   - Add sections for uax31, uax24, uts39
   - Update conformance status
   - Add security mechanisms overview

---

## Implementation Order

### Phase 1: Foundation (UAX #31, UAX #24, UAX #15)
1. **Implement UAX #31** - Identifier and Pattern Syntax
   - Setup package structure
   - Data generation
   - Core API
   - Testing
   - Documentation

2. **Implement UAX #24** - Script Property
   - Setup package structure
   - Parse script data files
   - Implement script resolution
   - Testing
   - Documentation

3. **Implement UAX #15** - Unicode Normalization Forms
   - Setup package structure
   - Parse normalization data files
   - Implement decomposition algorithm
   - Implement canonical ordering
   - Implement composition algorithm
   - Testing against NormalizationTest.txt
   - Documentation

### Phase 2: Security (UTS #39)
4. **Implement UTS #39** - Unicode Security Mechanisms
   - Setup package structure
   - Parse confusables data
   - Implement skeleton algorithm (using uax15)
   - Implement confusable detection
   - Implement mixed-script detection (using uax24)
   - Implement restriction levels (using uax31)
   - Comprehensive testing
   - Documentation

### Phase 3: Integration
5. **Integration and polish**
   - Cross-package integration tests
   - Performance optimization
   - Security audit
   - Documentation review
   - Prepare for merge to develop

---

## Success Criteria

### UAX #31
- [ ] All properties (XID_Start, XID_Continue, Pattern_Syntax, Pattern_White_Space) implemented
- [ ] Binary search O(log n) lookups
- [ ] Comprehensive test coverage
- [ ] README with examples

### UAX #24
- [ ] Script and Script_Extensions properties implemented
- [ ] Script resolution algorithm working
- [ ] All ~150 scripts supported
- [ ] Comprehensive test coverage
- [ ] README with examples

### UAX #15
- [ ] All four normalization forms implemented (NFD, NFC, NFKD, NFKC)
- [ ] Decomposition algorithm working
- [ ] Canonical ordering implemented
- [ ] Composition algorithm working
- [ ] Quick check optimization functional
- [ ] Passes NormalizationTest.txt (100% conformance)
- [ ] Handles Hangul syllables correctly
- [ ] README with examples

### UTS #39
- [ ] Confusable detection working (skeleton algorithm)
- [ ] Mixed-script detection implemented
- [ ] All restriction levels supported
- [ ] Security issue reporting functional
- [ ] Passes official test data (if available)
- [ ] Real-world test cases verified
- [ ] README with security guidelines

### Integration
- [ ] All packages follow established patterns
- [ ] No circular dependencies
- [ ] Clean git history on main branch
- [ ] Updated top-level README
- [ ] No "Generated with Claude Code" footers in commits

---

## Next Steps

1. Create `text-security` branch off `develop`
2. Create `uax31` branch off `text-security`
3. Begin UAX #31 implementation
4. Create `uax24` branch off `text-security`
5. Begin UAX #24 implementation
6. Create `uax15` branch off `text-security`
7. Begin UAX #15 implementation
8. Complete UTS #39 implementation
9. Merge all branches to `develop`
10. Clean up history for `main` branch

---

## References

### Standards
- [UTS #39: Unicode Security Mechanisms](https://www.unicode.org/reports/tr39/)
- [UAX #31: Unicode Identifier and Pattern Syntax](https://www.unicode.org/reports/tr31/)
- [UAX #24: Unicode Script Property](https://www.unicode.org/reports/tr24/)
- [UAX #15: Unicode Normalization Forms](https://www.unicode.org/reports/tr15/)
- [UTR #36: Unicode Security Considerations](https://www.unicode.org/reports/tr36/)

### Data Files
- [Unicode 17.0.0 Character Database](https://www.unicode.org/Public/17.0.0/ucd/)
- [Unicode Security Data](https://www.unicode.org/Public/security/17.0.0/)

### Related Documentation
- [ARCHITECTURE.md](./ARCHITECTURE.md) - Repository design patterns
- [TESTING.md](./TESTING.md) - Testing guidelines
- [README.md](./README.md) - Package overview
