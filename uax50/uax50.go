// Package uax50 implements the Unicode Vertical Text Layout algorithm (UAX #50).
//
// This package provides vertical orientation property lookup for characters to
// determine how they should be displayed in vertical text layouts. This is
// particularly important for East Asian typography and mixed-script text.
//
// Based on: https://www.unicode.org/reports/tr50/
//
// The Vertical_Orientation property has four values:
//   - Rotated (R): Characters that should be rotated 90 degrees clockwise
//   - Upright (U): Characters that maintain the same orientation as in Unicode charts
//   - TransformedUpright (Tu): Characters requiring different glyphs, fallback to Upright
//   - TransformedRotated (Tr): Characters requiring different glyphs, fallback to Rotated
//
// # Conformance
//
// As specified in UAX #50, the Vertical_Orientation property is informative and
// provides default values for reliable document interchange. Higher-level protocols,
// markup, or application preferences may override these defaults.
//
// # Usage
//
//	import "github.com/SCKelemen/unicode/uax50"
//
//	// Get orientation for a single character
//	orientation := uax50.LookupOrientation('A')  // Returns Rotated
//	orientation = uax50.LookupOrientation('中')  // Returns Upright
//
//	// Check if a character is upright
//	if uax50.IsUpright('中') {
//	    // Display character upright in vertical layout
//	}
//
//	// Get orientation for a string (grapheme cluster aware)
//	text := "Hello世界"
//	for _, r := range text {
//	    orientation := uax50.LookupOrientation(r)
//	    // Use orientation to layout character
//	}
//
// # Grapheme Clusters
//
// UAX #50 §3.1 states that "the interesting unit of text is not the character,
// but a grapheme cluster." This ensures combining marks are oriented consistently
// with their base characters. For complete grapheme cluster handling, combine this
// package with UAX #29 (Text Segmentation).
//
// # References
//
//   - UAX #50: https://www.unicode.org/reports/tr50/
//   - §3: Determining Vertical Orientation: https://www.unicode.org/reports/tr50/#Determining_Vertical_Orientation
//   - §4: Default Orientation: https://www.unicode.org/reports/tr50/#Default_Orientation
package uax50

// Orientation represents the vertical orientation property of a Unicode character.
// See UAX #50 §3: https://www.unicode.org/reports/tr50/#Determining_Vertical_Orientation
type Orientation int

const (
	// Rotated (R) indicates the character should be rotated 90 degrees clockwise
	// compared to its appearance in the Unicode code charts.
	// This is the default for most Latin, Greek, Cyrillic, and other scripts.
	// See UAX #50 §4.1: https://www.unicode.org/reports/tr50/#Default_Orientation
	Rotated Orientation = iota

	// Upright (U) indicates the character should maintain the same orientation
	// as shown in the Unicode code charts when displayed in vertical text.
	// This is the default for CJK ideographs and related punctuation.
	// See UAX #50 §4.1: https://www.unicode.org/reports/tr50/#Default_Orientation
	Upright

	// TransformedUpright (Tu) indicates the character should use a different
	// glyph when displayed in vertical text, with fallback to Upright orientation.
	// Examples include certain punctuation marks that have vertical-specific forms.
	// See UAX #50 §3.2: https://www.unicode.org/reports/tr50/#Transformed_Glyphs
	TransformedUpright

	// TransformedRotated (Tr) indicates the character should use a different
	// glyph when displayed in vertical text, with fallback to Rotated orientation.
	// This applies to some combining marks and special characters.
	// See UAX #50 §3.2: https://www.unicode.org/reports/tr50/#Transformed_Glyphs
	TransformedRotated
)

// String returns the standard abbreviated form of the orientation value.
func (o Orientation) String() string {
	switch o {
	case Rotated:
		return "R"
	case Upright:
		return "U"
	case TransformedUpright:
		return "Tu"
	case TransformedRotated:
		return "Tr"
	default:
		return "Unknown"
	}
}

// LookupOrientation returns the Vertical_Orientation property value for a given rune.
//
// This function performs a binary search on the Unicode character database to find
// the orientation property. If the character is not explicitly listed in the data,
// it returns Rotated as the default value per UAX #50 §4.1.
//
// The property is defined at the character level, but UAX #50 §3.1 notes that
// applications should consider grapheme clusters as the unit of text for orientation
// purposes. See https://www.unicode.org/reports/tr50/#Grapheme_Clusters
//
// Example:
//
//	orientation := uax50.LookupOrientation('A')      // Returns Rotated
//	orientation = uax50.LookupOrientation('中')      // Returns Upright
//	orientation = uax50.LookupOrientation('\u301C')  // Returns TransformedUpright (wave dash)
func LookupOrientation(r rune) Orientation {
	// Binary search through the ranges
	left, right := 0, len(verticalOrientationData)-1

	for left <= right {
		mid := (left + right) / 2
		entry := verticalOrientationData[mid]

		if r < entry.start {
			right = mid - 1
		} else if r > entry.end {
			left = mid + 1
		} else {
			// Found the range containing r
			return entry.value
		}
	}

	// Default to Rotated if not found
	// Per UAX #50 §4.1: "All other code points...are given the value R"
	// https://www.unicode.org/reports/tr50/#Default_Orientation
	return Rotated
}

// IsUpright returns true if the character has Upright or TransformedUpright orientation.
//
// This is a convenience function for determining if a character should be displayed
// upright in vertical text layouts, regardless of whether glyph transformation is needed.
//
// Example:
//
//	if uax50.IsUpright('中') {
//	    // Display character upright in vertical layout
//	}
func IsUpright(r rune) bool {
	o := LookupOrientation(r)
	return o == Upright || o == TransformedUpright
}

// IsRotated returns true if the character has Rotated or TransformedRotated orientation.
//
// This is a convenience function for determining if a character should be rotated
// 90 degrees clockwise in vertical text layouts, regardless of whether glyph
// transformation is needed.
//
// Example:
//
//	if uax50.IsRotated('A') {
//	    // Rotate character 90 degrees clockwise in vertical layout
//	}
func IsRotated(r rune) bool {
	o := LookupOrientation(r)
	return o == Rotated || o == TransformedRotated
}

// RequiresTransformation returns true if the character requires a different glyph
// in vertical text (Tu or Tr values).
//
// Characters with Tu or Tr orientations may have vertical-specific glyph variants.
// Applications should check for these variants when available. If no variant exists,
// fall back to the base orientation (Upright for Tu, Rotated for Tr).
//
// See UAX #50 §3.2: https://www.unicode.org/reports/tr50/#Transformed_Glyphs
//
// Example:
//
//	if uax50.RequiresTransformation(r) {
//	    // Try to use vertical-specific glyph variant
//	    // Fall back to base orientation if not available
//	}
func RequiresTransformation(r rune) bool {
	o := LookupOrientation(r)
	return o == TransformedUpright || o == TransformedRotated
}

// GetBaseOrientation returns the base orientation (Upright or Rotated) for a character.
//
// For characters with Transformed values (Tu/Tr), this returns the fallback orientation
// to use when a transformed glyph is not available:
//   - TransformedUpright (Tu) → Upright (U)
//   - TransformedRotated (Tr) → Rotated (R)
//
// For characters already having Upright or Rotated values, returns the value unchanged.
//
// Example:
//
//	base := uax50.GetBaseOrientation(r)
//	if base == uax50.Upright {
//	    // Display upright
//	} else {
//	    // Rotate 90 degrees clockwise
//	}
func GetBaseOrientation(r rune) Orientation {
	o := LookupOrientation(r)
	switch o {
	case TransformedUpright:
		return Upright
	case TransformedRotated:
		return Rotated
	default:
		return o
	}
}
