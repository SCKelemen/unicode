// Package uts39 implements Unicode Security Mechanisms (UTS #39).
//
// This package provides mechanisms to detect and prevent security issues
// arising from Unicode's large character repertoire, including:
//   - Confusable character detection (spoofing prevention)
//   - Mixed-script detection
//   - Identifier restriction levels
//   - Security profiles
//
// Based on: https://www.unicode.org/reports/tr39/
//
// # Confusable Detection
//
// The skeleton algorithm identifies visually confusable strings by
// normalizing them to a canonical form:
//
//	skeleton(X) = applyConfusables( toNFD( toCaseFold( toNFKD(X) ) ) )
//
// (The base transform follows the legacy UTS #39 ordering documented in this
// package; the confusables map is applied to the final NFD form and the
// result is iterated to a fixed point.)
//
// Two strings are confusable if their skeletons are identical.
//
// # Mixed-Script Detection
//
// Detects suspicious mixing of scripts in identifiers. Provides
// restriction levels from ASCII-only to unrestricted.
//
// # Usage
//
//	import "github.com/SCKelemen/unicode/v6/uts39"
//
//	// Check if two strings are visually confusable
//	if uts39.AreConfusable("paypal", "pаypal") {  // Second uses Cyrillic 'а'
//	    // Strings look the same but are different
//	}
//
//	// Get the skeleton for comparison
//	skel := uts39.Skeleton("Hello")
//
//	// Check restriction level
//	level := uts39.GetRestrictionLevel("user_name")
//	if level >= uts39.HighlyRestrictive {
//	    // Identifier is safe
//	}
//
// # Conformance
//
// This implementation follows UTS #39 Security Mechanisms:
//   - https://www.unicode.org/reports/tr39/
//
// The implementation uses data from Unicode 17.0.0.
//
// # References
//
//   - UTS #39: https://www.unicode.org/reports/tr39/
//   - Confusables data: https://www.unicode.org/Public/security/latest/confusables.txt
//   - Identifier data: https://www.unicode.org/Public/security/latest/IdentifierStatus.txt
package uts39

import (
	"fmt"

	"github.com/SCKelemen/unicode/v6/uax24"
	"github.com/SCKelemen/unicode/v6/uax31"
	"github.com/SCKelemen/unicode/v6/uts15"
)

// RestrictionLevel represents the restriction level of an identifier.
// Higher levels are more restrictive and generally more secure.
type RestrictionLevel int

const (
	// Unrestricted allows any characters
	Unrestricted RestrictionLevel = iota

	// MinimallyRestrictive allows Latin + one other script
	MinimallyRestrictive

	// ModeratelyRestrictive allows multiple scripts with specific rules
	ModeratelyRestrictive

	// HighlyRestrictive allows single script + Common + Inherited
	HighlyRestrictive

	// SingleScript requires all characters from a single script
	// (excluding Common and Inherited)
	SingleScript

	// ASCIIOnly restricts to ASCII characters only
	ASCIIOnly
)

// String returns the string representation of a RestrictionLevel.
func (l RestrictionLevel) String() string {
	switch l {
	case ASCIIOnly:
		return "ASCII-Only"
	case SingleScript:
		return "Single-Script"
	case HighlyRestrictive:
		return "Highly-Restrictive"
	case ModeratelyRestrictive:
		return "Moderately-Restrictive"
	case MinimallyRestrictive:
		return "Minimally-Restrictive"
	case Unrestricted:
		return "Unrestricted"
	default:
		return "Unknown"
	}
}

// Skeleton returns the skeleton of a string for confusable detection.
//
// The skeleton algorithm normalizes strings to identify visual confusability.
// This implementation follows the legacy UTS #39 ordering documented for this
// package:
//
//	skeleton(X) = applyConfusables( toNFD( toCaseFold( toNFKD(X) ) ) )
//
// where:
//
//   - toNFKD applies Unicode Normalization Form KD (compatibility decomposition).
//   - toCaseFold applies "full" Unicode case folding (status C+F mappings from
//     CaseFolding.txt); this is what causes "ß" to fold to "ss" and what
//     distinguishes proper folding from a simple strings.ToLower call.
//   - toNFD applies Unicode Normalization Form D (canonical decomposition).
//   - applyConfusables applies the confusables.txt prototype mappings to the
//     final NFD form.
//
// After the initial transform, the (NFD → applyConfusables) tail is repeated
// until a fixed point is reached. The function is mathematically idempotent
// per UTS #39, so convergence is expected within a handful of iterations.
// A defensive safety cap (16 iterations) panics on non-convergence to surface
// any future data regressions rather than silently truncate.
//
// Two strings are confusable if their skeletons are equal.
//
// Example:
//
//	Skeleton("paypal") == Skeleton("pаypal")  // true (Cyrillic 'а')
//	Skeleton("ß")      == Skeleton("ss")      // true (full case fold)
//
// See: https://www.unicode.org/reports/tr39/#Confusable_Detection
func Skeleton(s string) string {
	// Base transform: NFKD → caseFold → NFD → applyConfusables.
	s = uts15.NFKD(s)
	s = caseFold(s)
	s = uts15.NFD(s)
	s = applyConfusables(s)

	// Iterate the (NFD → confusables) tail to a fixed point. Per UTS #39
	// the skeleton transform is idempotent, but a confusable replacement
	// may produce a sequence that is not in NFD or that itself maps further,
	// so a small number of iterations may be required.
	const maxIterations = 16
	for i := 0; i < maxIterations; i++ {
		prev := s
		s = uts15.NFD(s)
		s = applyConfusables(s)
		if s == prev {
			return s
		}
	}

	// Reaching this point means the skeleton transform did not converge
	// within the safety cap, which would indicate a regression in the
	// confusables data (e.g. a mapping cycle). Surface this loudly rather
	// than silently truncate the result.
	panic(fmt.Sprintf("uts39: Skeleton failed to converge after %d iterations (input=%q)",
		maxIterations, s))
}

// AreConfusable reports whether two strings are visually confusable.
//
// Two strings are confusable if they have the same skeleton, meaning
// they look similar enough to be confused by users.
//
// Example:
//
//	AreConfusable("scope", "ѕсоре")  // true (contains Cyrillic lookalikes)
//	AreConfusable("hello", "world")  // false
//
// See: https://www.unicode.org/reports/tr39/#Confusable_Detection
func AreConfusable(s1, s2 string) bool {
	return Skeleton(s1) == Skeleton(s2)
}

// applyConfusables applies confusable character mappings to a string
func applyConfusables(s string) string {
	runes := []rune(s)
	result := make([]rune, 0, len(runes))

	for _, r := range runes {
		// Binary search for confusable mapping
		target := getConfusableTarget(r)
		if target != "" {
			result = append(result, []rune(target)...)
		} else {
			result = append(result, r)
		}
	}

	return string(result)
}

// getConfusableTarget returns the confusable target for a rune, or empty string if none
func getConfusableTarget(r rune) string {
	// Binary search in confusablesData
	lo, hi := 0, len(confusablesData)
	for lo < hi {
		mid := lo + (hi-lo)/2
		if r < confusablesData[mid].source {
			hi = mid
		} else if r > confusablesData[mid].source {
			lo = mid + 1
		} else {
			return confusablesData[mid].target
		}
	}
	return ""
}

// caseFold applies Unicode "full" case folding to s, using the C+F mappings
// from CaseFolding.txt. This is what UTS #39 §4 calls toCaseFold.
//
// Unlike strings.ToLower, full case folding can produce a string longer than
// its input (e.g. "ß" → "ss", "ﬃ" → "ffi", "İ" → "i\u0307") and uses Unicode's
// language-neutral folding (no Turkic-specific 'T' mappings).
func caseFold(s string) string {
	// Fast path: pure ASCII. Only A-Z need folding; everything else is
	// already at its folded form.
	allASCII := true
	needsFold := false
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b >= 0x80 {
			allASCII = false
			break
		}
		if b >= 'A' && b <= 'Z' {
			needsFold = true
		}
	}
	if allASCII {
		if !needsFold {
			return s
		}
		// ASCII-only with at least one upper-case letter: fold in place.
		buf := make([]byte, len(s))
		for i := 0; i < len(s); i++ {
			b := s[i]
			if b >= 'A' && b <= 'Z' {
				b += 'a' - 'A'
			}
			buf[i] = b
		}
		return string(buf)
	}

	// General path: walk runes and substitute via the case fold table.
	// Pre-grow capacity slightly to absorb the common 1→1 case without
	// reallocation, while still allowing 1→N expansion (e.g. ß → ss).
	var b []byte
	if cap(b) < len(s) {
		b = make([]byte, 0, len(s)+8)
	}
	for _, r := range s {
		if folded := getCaseFoldTarget(r); folded != "" {
			b = append(b, folded...)
		} else {
			b = append(b, string(r)...)
		}
	}
	return string(b)
}

// getCaseFoldTarget returns the case-folded target for a rune, or empty
// string if r folds to itself. Uses binary search over caseFoldData.
func getCaseFoldTarget(r rune) string {
	// caseFoldData entries are sorted by source code point.
	lo, hi := 0, len(caseFoldData)
	for lo < hi {
		mid := lo + (hi-lo)/2
		if r < caseFoldData[mid].source {
			hi = mid
		} else if r > caseFoldData[mid].source {
			lo = mid + 1
		} else {
			return caseFoldData[mid].target
		}
	}
	return ""
}

// GetRestrictionLevel returns the restriction level of a string per
// UAX #39 §5.2 Table 1.
//
// Restriction levels are ordered from most to least restrictive:
//
//   - ASCIIOnly:              Only ASCII characters.
//   - SingleScript:           Exactly one script (Common/Inherited ignored).
//   - HighlyRestrictive:      Single script, OR Latin combined with one of the
//     JCB script groups: {Han, Hiragana, Katakana} (Latn+Jpan),
//     {Han, Bopomofo} (Latn+Hanb), or {Han, Hangul} (Latn+Kore).
//   - ModeratelyRestrictive:  Latin plus exactly one other script that is not
//     Cyrillic or Greek (per UAX #39 §5.2 Table 1).
//   - MinimallyRestrictive:   Any other multi-script combination (e.g. Latin +
//     Cyrillic, Latin + Greek).
//   - Unrestricted:           Empty string / fallback.
//
// See: https://www.unicode.org/reports/tr39/#Restriction_Level_Detection
func GetRestrictionLevel(s string) RestrictionLevel {
	if s == "" {
		return Unrestricted
	}

	// Check if ASCII-only.
	isASCII := true
	for _, r := range s {
		if r > 127 {
			isASCII = false
			break
		}
	}
	if isASCII {
		return ASCIIOnly
	}

	// Collect all scripts and filter out Common / Inherited.
	scripts := GetIdentifierScripts(s)
	mainScripts := make([]uax24.Script, 0, len(scripts))
	for _, script := range scripts {
		if script != uax24.ScriptCommon && script != uax24.ScriptInherited {
			mainScripts = append(mainScripts, script)
		}
	}

	// SingleScript: only one resolved script.
	if len(mainScripts) == 1 {
		return SingleScript
	}

	// HighlyRestrictive: per UAX #39 §5.2 Table 1, the set of scripts (after
	// removing Common and Inherited) must be a subset of one of the following
	// CJK augmentations of Latin:
	//
	//     a) {Latin, Han, Hiragana, Katakana}   (Latn + Jpan)
	//     b) {Latin, Han, Bopomofo}             (Latn + Hanb)
	//     c) {Latin, Han, Hangul}               (Latn + Kore)
	//
	// These are intentionally checked as subset relations so that, e.g.,
	// "Latin + Han" alone also qualifies (it is a subset of all three).
	if isScriptSubset(mainScripts, latnJpanSet) ||
		isScriptSubset(mainScripts, latnHanbSet) ||
		isScriptSubset(mainScripts, latnKoreSet) {
		return HighlyRestrictive
	}

	// ModeratelyRestrictive: Latin + exactly one other script, excluding the
	// historically high-confusability scripts Cyrillic and Greek per
	// UAX #39 §5.2 Table 1.
	hasLatin := containsScript(mainScripts, uax24.ScriptLatin)
	if hasLatin && len(mainScripts) == 2 {
		other := otherScript(mainScripts, uax24.ScriptLatin)
		if other != uax24.ScriptCyrillic && other != uax24.ScriptGreek {
			return ModeratelyRestrictive
		}
	}

	// MinimallyRestrictive: any remaining multi-script combination
	// (e.g. Latin + Cyrillic, Latin + Greek, or three or more scripts not
	// captured by HighlyRestrictive).
	return MinimallyRestrictive
}

// latnJpanSet is the script set for Latin + Japanese (Latn + Jpan).
var latnJpanSet = map[uax24.Script]struct{}{
	uax24.ScriptLatin:    {},
	uax24.ScriptHan:      {},
	uax24.ScriptHiragana: {},
	uax24.ScriptKatakana: {},
}

// latnHanbSet is the script set for Latin + Han-with-Bopomofo (Latn + Hanb).
var latnHanbSet = map[uax24.Script]struct{}{
	uax24.ScriptLatin:    {},
	uax24.ScriptHan:      {},
	uax24.ScriptBopomofo: {},
}

// latnKoreSet is the script set for Latin + Korean (Latn + Kore).
var latnKoreSet = map[uax24.Script]struct{}{
	uax24.ScriptLatin:  {},
	uax24.ScriptHan:    {},
	uax24.ScriptHangul: {},
}

// isScriptSubset reports whether every script in scripts is also in allowed.
func isScriptSubset(scripts []uax24.Script, allowed map[uax24.Script]struct{}) bool {
	for _, s := range scripts {
		if _, ok := allowed[s]; !ok {
			return false
		}
	}
	return true
}

// containsScript reports whether scripts contains target.
func containsScript(scripts []uax24.Script, target uax24.Script) bool {
	for _, s := range scripts {
		if s == target {
			return true
		}
	}
	return false
}

// otherScript returns the first script in scripts that is not equal to skip.
// Caller must ensure such a script exists.
func otherScript(scripts []uax24.Script, skip uax24.Script) uax24.Script {
	for _, s := range scripts {
		if s != skip {
			return s
		}
	}
	return uax24.ScriptUnknown
}

// GetIdentifierScripts returns the scripts used in an identifier string.
//
// This function returns all scripts present in the string, including
// Common and Inherited scripts.
//
// Example:
//
//	scripts := GetIdentifierScripts("Hello мир")  // [Latin, Cyrillic, Common]
func GetIdentifierScripts(s string) []uax24.Script {
	scriptSet := make(map[uax24.Script]bool)

	for _, r := range s {
		script := uax24.LookupScript(r)
		scriptSet[script] = true
	}

	scripts := make([]uax24.Script, 0, len(scriptSet))
	for script := range scriptSet {
		scripts = append(scripts, script)
	}

	return scripts
}

// IsMixedScript reports whether an identifier uses multiple scripts.
//
// A string is considered mixed-script if it contains characters from
// more than one script, excluding Common and Inherited.
//
// Example:
//
//	IsMixedScript("hello")      // false (single script)
//	IsMixedScript("hello世界")  // true (Latin + Han)
//
// See: https://www.unicode.org/reports/tr39/#Mixed_Script_Detection
func IsMixedScript(s string) bool {
	// Fast path: ASCII is single-script (Latin)
	isASCII := true
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			isASCII = false
			break
		}
	}
	if isASCII {
		return false
	}

	scripts := GetIdentifierScripts(s)

	// Count non-Common, non-Inherited scripts
	count := 0
	for _, script := range scripts {
		if script != uax24.ScriptCommon && script != uax24.ScriptInherited {
			count++
		}
	}

	return count > 1
}

// IsValidIdentifier reports whether a string is a valid identifier
// according to UAX #31 Default Identifier Syntax.
//
// This checks that the string follows the pattern:
//
//	<XID_Start> <XID_Continue>*
//
// Example:
//
//	IsValidIdentifier("myVar")     // true
//	IsValidIdentifier("my-var")    // false (hyphen not allowed)
//	IsValidIdentifier("123var")    // false (starts with digit)
func IsValidIdentifier(s string) bool {
	return uax31.IsValidIdentifier(s)
}

// IsSafeIdentifier reports whether an identifier is safe from common
// security issues.
//
// An identifier is considered safe if it:
//   - Is a valid identifier (UAX #31)
//   - Has a restriction level of at least HighlyRestrictive
//   - Does not contain invisible or deprecated characters
//
// Example:
//
//	IsSafeIdentifier("user_name")      // true
//	IsSafeIdentifier("user\u200Bname") // false (contains zero-width space)
func IsSafeIdentifier(s string) bool {
	// Fast path: ASCII identifiers are safe if valid
	isASCII := true
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			isASCII = false
			break
		}
	}
	if isASCII {
		return IsValidIdentifier(s)
	}

	if !IsValidIdentifier(s) {
		return false
	}

	level := GetRestrictionLevel(s)
	if level < HighlyRestrictive {
		return false
	}

	// Check for invisible characters
	for _, r := range s {
		if isInvisible(r) {
			return false
		}
	}

	return true
}

// isInvisible reports whether a rune is invisible (zero-width, formatting, etc.)
// Optimized with switch statement for O(1) lookup instead of O(n) linear search
func isInvisible(r rune) bool {
	switch r {
	case 0x200B, // Zero Width Space
		0x200C, // Zero Width Non-Joiner
		0x200D, // Zero Width Joiner
		0x200E, // Left-To-Right Mark
		0x200F, // Right-To-Left Mark
		0xFEFF, // Zero Width No-Break Space
		0x202A, // Left-To-Right Embedding
		0x202B, // Right-To-Left Embedding
		0x202C, // Pop Directional Formatting
		0x202D, // Left-To-Right Override
		0x202E: // Right-To-Left Override
		return true
	default:
		return false
	}
}
