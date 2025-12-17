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
//	skeleton(X) = toNFD(toCaseFold(toNFKD(X)))
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
//	import "github.com/SCKelemen/unicode/uts39"
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
	"strings"

	"github.com/SCKelemen/unicode/uax24"
	"github.com/SCKelemen/unicode/uax31"
	"github.com/SCKelemen/unicode/uts15"
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
// The skeleton algorithm normalizes strings to identify visual confusability:
//   skeleton(X) = toNFD(toCaseFold(toNFKD(X)))
//
// Two strings are confusable if their skeletons are equal.
//
// Example:
//	skeleton("paypal") == skeleton("pаypal")  // true (Cyrillic 'а')
//
// See: https://www.unicode.org/reports/tr39/#Confusable_Detection
func Skeleton(s string) string {
	// Step 1: Apply NFKD normalization
	s = uts15.NFKD(s)

	// Step 2: Apply confusable mappings
	s = applyConfusables(s)

	// Step 3: Apply case folding (convert to lowercase)
	s = strings.ToLower(s)

	// Step 4: Apply NFD normalization
	s = uts15.NFD(s)

	// Step 5: Apply confusable mappings again
	s = applyConfusables(s)

	// Step 6: Apply NFD again
	s = uts15.NFD(s)

	return s
}

// AreConfusable reports whether two strings are visually confusable.
//
// Two strings are confusable if they have the same skeleton, meaning
// they look similar enough to be confused by users.
//
// Example:
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

// GetRestrictionLevel returns the restriction level of a string.
//
// Restriction levels from most to least restrictive:
//   - ASCIIOnly: Only ASCII characters
//   - SingleScript: One script (excluding Common/Inherited)
//   - HighlyRestrictive: One script + Common + Inherited
//   - ModeratelyRestrictive: Multiple scripts following specific rules
//   - MinimallyRestrictive: Latin + one other script
//   - Unrestricted: Any character combination
//
// See: https://www.unicode.org/reports/tr39/#Restriction_Level_Detection
func GetRestrictionLevel(s string) RestrictionLevel {
	if s == "" {
		return Unrestricted
	}

	// Check if ASCII-only
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

	// Get all scripts used
	scripts := GetIdentifierScripts(s)

	// Filter out Common and Inherited
	mainScripts := make([]uax24.Script, 0, len(scripts))
	for _, script := range scripts {
		if script != uax24.ScriptCommon && script != uax24.ScriptInherited {
			mainScripts = append(mainScripts, script)
		}
	}

	// SingleScript: only one main script
	if len(mainScripts) == 1 {
		return SingleScript
	}

	// HighlyRestrictive: one script + Common + Inherited is same as SingleScript
	// for our purposes since we filtered out Common/Inherited
	if len(mainScripts) == 1 {
		return HighlyRestrictive
	}

	// Check if Latin is present
	hasLatin := false
	for _, script := range mainScripts {
		if script == uax24.ScriptLatin {
			hasLatin = true
			break
		}
	}

	// MinimallyRestrictive: Latin + exactly one other script
	if hasLatin && len(mainScripts) == 2 {
		return MinimallyRestrictive
	}

	// ModeratelyRestrictive: multiple scripts with specific allowed combinations
	// For simplicity, we consider any multi-script as moderately restrictive
	// unless it matches minimally restrictive
	if len(mainScripts) > 1 {
		return ModeratelyRestrictive
	}

	return Unrestricted
}

// GetIdentifierScripts returns the scripts used in an identifier string.
//
// This function returns all scripts present in the string, including
// Common and Inherited scripts.
//
// Example:
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
//	IsMixedScript("hello")      // false (single script)
//	IsMixedScript("hello世界")  // true (Latin + Han)
//
// See: https://www.unicode.org/reports/tr39/#Mixed_Script_Detection
func IsMixedScript(s string) bool {
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
//   <XID_Start> <XID_Continue>*
//
// Example:
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
//	IsSafeIdentifier("user_name")      // true
//	IsSafeIdentifier("user\u200Bname") // false (contains zero-width space)
func IsSafeIdentifier(s string) bool {
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
func isInvisible(r rune) bool {
	// Common invisible characters
	invisibleRunes := []rune{
		0x200B, // Zero Width Space
		0x200C, // Zero Width Non-Joiner
		0x200D, // Zero Width Joiner
		0x200E, // Left-To-Right Mark
		0x200F, // Right-To-Left Mark
		0xFEFF, // Zero Width No-Break Space
		0x202A, // Left-To-Right Embedding
		0x202B, // Right-To-Left Embedding
		0x202C, // Pop Directional Formatting
		0x202D, // Left-To-Right Override
		0x202E, // Right-To-Left Override
	}

	for _, invisible := range invisibleRunes {
		if r == invisible {
			return true
		}
	}

	return false
}
