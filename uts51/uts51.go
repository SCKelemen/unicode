// Package uts51 implements Unicode Emoji (UTS #51).
//
// This package provides emoji property detection, sequence validation, and
// integration with other Unicode standards for complete emoji support in
// terminal emulators, text editors, and layout engines.
//
// Based on: https://www.unicode.org/reports/tr51/
//
// UTS #51 defines emoji properties, sequences, and presentation requirements.
// This implementation provides:
//   - Emoji property detection (Emoji, Emoji_Presentation, etc.)
//   - Emoji sequence validation (ZWJ, modifier, flag, keycap, tag sequences)
//   - Grapheme cluster segmentation for emoji
//   - Width calculation for terminal rendering (integrates with UAX #11)
//   - 100% conformance testing against emoji-test.txt
//
// # Emoji Properties
//
// UTS #51 defines six core properties (see §1.4):
// https://www.unicode.org/reports/tr51/#Emoji_Properties
//
//   - Emoji: Characters recommended for emoji use
//   - Emoji_Presentation: Characters that display as emoji by default
//   - Emoji_Modifier: Skin tone modifiers (U+1F3FB..U+1F3FF)
//   - Emoji_Modifier_Base: Characters that accept modifiers
//   - Emoji_Component: Characters used in emoji sequences
//   - Extended_Pictographic: Pictographic characters for segmentation
//
// # Emoji Sequences
//
// UTS #51 defines several sequence types (see §2):
// https://www.unicode.org/reports/tr51/#Emoji_Sequences
//
//   - Emoji presentation sequence: character + U+FE0F (emoji selector)
//   - Text presentation sequence: character + U+FE0E (text selector)
//   - Emoji modifier sequence: base + skin tone modifier
//   - Emoji flag sequence: two regional indicator characters
//   - Emoji keycap sequence: [0-9#*] + U+FE0F + U+20E3
//   - Emoji ZWJ sequence: characters joined by U+200D (zero-width joiner)
//   - Emoji tag sequence: base + tag characters + U+E007F terminator
//
// # Integration with Other Standards
//
// This package integrates with:
//   - UAX #11 (East Asian Width): Emoji width calculation for terminals
//   - UAX #14 (Line Breaking): Break opportunities around emoji
//   - UAX #29 (Text Segmentation): Grapheme cluster boundaries
//   - UAX #50 (Vertical Text Layout): Emoji orientation in vertical text
//
// # Terminal Rendering
//
// Emoji typically occupy 2 columns in terminal emulators (like CJK ideographs).
// This package provides width calculation that integrates with UAX #11.
//
// Per UTS #51 §4 (Display):
// https://www.unicode.org/reports/tr51/#Display
//
// "Emoji typically have the same vertical placement and advance width as
// CJK ideographs."
//
// # Usage
//
//	import "github.com/SCKelemen/unicode/uts51"
//
//	// Check if a character is emoji
//	if uts51.IsEmoji('😀') {
//	    // Handle emoji
//	}
//
//	// Check default presentation
//	if uts51.HasEmojiPresentation('😀') {
//	    // Displays as colored emoji by default
//	}
//
//	// Validate emoji sequences
//	if uts51.IsValidEmojiSequence([]rune{0x1F468, 0x200D, 0x1F469, 0x200D, 0x1F467}) {
//	    // Valid ZWJ sequence: family emoji
//	}
//
//	// Calculate width for terminal rendering
//	width := uts51.EmojiWidth('😀')  // Returns 2 (like CJK characters)
//
// # Conformance
//
// This implementation conforms to UTS #51 Version 17.0.
//
// Conformance requirements (see §5):
// https://www.unicode.org/reports/tr51/#Conformance
//
//   - C1: Version identification
//   - C2: Capability declaration for emoji sets
//   - C3: Rejection of invalid sequences
//
// References:
//   - UTS #51: https://www.unicode.org/reports/tr51/
//   - §1.4 Emoji Properties: https://www.unicode.org/reports/tr51/#Emoji_Properties
//   - §2 Emoji Sequences: https://www.unicode.org/reports/tr51/#Emoji_Sequences
//   - §4 Display: https://www.unicode.org/reports/tr51/#Display
//   - §5 Conformance: https://www.unicode.org/reports/tr51/#Conformance
package uts51

// Special Unicode code points for emoji sequences
const (
	// VariationSelector15 (U+FE0E) requests text presentation
	// See UTS #51 §2.1: https://www.unicode.org/reports/tr51/#def_text_presentation_sequence
	VariationSelector15 = '\uFE0E'

	// VariationSelector16 (U+FE0F) requests emoji presentation
	// See UTS #51 §2.1: https://www.unicode.org/reports/tr51/#def_emoji_presentation_sequence
	VariationSelector16 = '\uFE0F'

	// ZeroWidthJoiner (U+200D) joins emoji to form ZWJ sequences
	// See UTS #51 §2.4: https://www.unicode.org/reports/tr51/#def_emoji_zwj_sequence
	ZeroWidthJoiner = '\u200D'

	// CombiningEnclosingKeycap (U+20E3) creates keycap sequences
	// See UTS #51 §2.3: https://www.unicode.org/reports/tr51/#def_emoji_keycap_sequence
	CombiningEnclosingKeycap = '\u20E3'

	// TagTerminator (U+E007F) terminates emoji tag sequences
	// See UTS #51 §2.6: https://www.unicode.org/reports/tr51/#def_emoji_tag_sequence
	TagTerminator = '\U000E007F'

	// RegionalIndicatorBase is the start of regional indicator symbols (U+1F1E6)
	// Regional indicators form flag sequences
	// See UTS #51 §2.5: https://www.unicode.org/reports/tr51/#def_emoji_flag_sequence
	RegionalIndicatorBase = '\U0001F1E6'

	// RegionalIndicatorEnd is the end of regional indicator symbols (U+1F1FF)
	RegionalIndicatorEnd = '\U0001F1FF'

	// TagBase is the start of tag characters (U+E0020)
	// See UTS #51 §2.6: https://www.unicode.org/reports/tr51/#def_emoji_tag_sequence
	TagBase = '\U000E0020'

	// TagEnd is the end of tag characters (U+E007E)
	TagEnd = '\U000E007E'
)

// IsEmoji returns true if the rune has the Emoji property.
//
// Per UTS #51 §1.4.1: https://www.unicode.org/reports/tr51/#Emoji_Properties
// "Characters with the Emoji property are recommended for emoji use."
//
// Note: Having Emoji=Yes does not mean the character displays as emoji by default.
// Use HasEmojiPresentation() to check default presentation.
//
// Example:
//
//	uts51.IsEmoji('😀')  // true - emoji character
//	uts51.IsEmoji('#')   // true - can be used in emoji keycap sequences
//	uts51.IsEmoji('A')   // false - not an emoji character
func IsEmoji(r rune) bool {
	return inRanges(r, emojiRanges)
}

// HasEmojiPresentation returns true if the rune has the Emoji_Presentation property.
//
// Per UTS #51 §1.4.2: https://www.unicode.org/reports/tr51/#Emoji_Properties
// "Characters with Emoji_Presentation=Yes display as emoji (colorful) by default."
//
// Characters with Emoji_Presentation=No display as text by default and require
// U+FE0F (emoji presentation selector) to display as emoji.
//
// Example:
//
//	uts51.HasEmojiPresentation('😀')  // true - displays as emoji by default
//	uts51.HasEmojiPresentation('☺')   // false - displays as text by default
func HasEmojiPresentation(r rune) bool {
	return inRanges(r, emojiPresentationRanges)
}

// IsEmojiModifier returns true if the rune is an emoji modifier (skin tone).
//
// Per UTS #51 §1.4.3: https://www.unicode.org/reports/tr51/#Emoji_Properties
// "Emoji modifiers are characters that modify the appearance of a preceding
// emoji modifier base character."
//
// The five skin tone modifiers are U+1F3FB..U+1F3FF.
//
// Example:
//
//	uts51.IsEmojiModifier('\U0001F3FB')  // true - light skin tone
//	uts51.IsEmojiModifier('😀')          // false - not a modifier
func IsEmojiModifier(r rune) bool {
	return inRanges(r, emojiModifierRanges)
}

// IsEmojiModifierBase returns true if the rune accepts emoji modifiers.
//
// Per UTS #51 §1.4.4: https://www.unicode.org/reports/tr51/#Emoji_Properties
// "Emoji modifier base characters can be modified by emoji modifiers."
//
// Example:
//
//	uts51.IsEmojiModifierBase('👋')  // true - can have skin tone
//	uts51.IsEmojiModifierBase('😀')  // false - faces don't have skin tones
func IsEmojiModifierBase(r rune) bool {
	return inRanges(r, emojiModifierBaseRanges)
}

// IsEmojiComponent returns true if the rune is an emoji component.
//
// Per UTS #51 §1.4.5: https://www.unicode.org/reports/tr51/#Emoji_Properties
// "Emoji components are characters used in emoji sequences but not intended
// for independent use."
//
// Example:
//
//	uts51.IsEmojiComponent('\U0001F3FB')  // true - skin tone modifier
//	uts51.IsEmojiComponent('😀')          // false - standalone emoji
func IsEmojiComponent(r rune) bool {
	return inRanges(r, emojiComponentRanges)
}

// IsExtendedPictographic returns true if the rune has the Extended_Pictographic property.
//
// Per UTS #51 §1.4.6: https://www.unicode.org/reports/tr51/#Emoji_Properties
// "Extended_Pictographic characters include all emoji and pictographic characters
// for segmentation purposes."
//
// This property is used for grapheme cluster boundaries and line breaking.
//
// Example:
//
//	uts51.IsExtendedPictographic('😀')  // true
//	uts51.IsExtendedPictographic('🏳')   // true - flag
func IsExtendedPictographic(r rune) bool {
	return inRanges(r, extendedPictographicRanges)
}

// IsRegionalIndicator returns true if the rune is a regional indicator symbol.
//
// Regional indicators (U+1F1E6..U+1F1FF) represent letters A-Z and are used
// in pairs to form flag emoji.
//
// Per UTS #51 §2.5: https://www.unicode.org/reports/tr51/#def_emoji_flag_sequence
//
// Example:
//
//	uts51.IsRegionalIndicator('\U0001F1FA')  // true - 'U' (US flag part 1)
//	uts51.IsRegionalIndicator('\U0001F1F8')  // true - 'S' (US flag part 2)
func IsRegionalIndicator(r rune) bool {
	return r >= RegionalIndicatorBase && r <= RegionalIndicatorEnd
}

// IsTagCharacter returns true if the rune is an emoji tag character.
//
// Tag characters (U+E0020..U+E007E) are used in emoji tag sequences for
// subdivision flags.
//
// Per UTS #51 §2.6: https://www.unicode.org/reports/tr51/#def_emoji_tag_sequence
//
// Example:
//
//	uts51.IsTagCharacter('\U000E0067')  // true - tag letter 'g'
func IsTagCharacter(r rune) bool {
	return r >= TagBase && r <= TagEnd
}

// inRanges checks if a rune is within any of the given ranges
func inRanges(r rune, ranges []emojiRange) bool {
	// Binary search
	left, right := 0, len(ranges)-1

	for left <= right {
		mid := (left + right) / 2
		entry := ranges[mid]

		if r < entry.start {
			right = mid - 1
		} else if r > entry.end {
			left = mid + 1
		} else {
			return true
		}
	}

	return false
}

// EmojiWidth returns the display width of an emoji character in terminal columns.
//
// Per UTS #51 §4 (Display): https://www.unicode.org/reports/tr51/#Display
// "Emoji typically have the same vertical placement and advance width as
// CJK ideographs."
//
// This integrates with UAX #11 (East Asian Width) where emoji characters
// typically have Wide or Ambiguous width (2 columns in East Asian context).
//
// Returns:
//   - 2 for emoji with emoji presentation
//   - 1 for emoji without emoji presentation (text default)
//   - 0 for emoji components (modifiers, variation selectors, ZWJ)
//
// Note: For complete width calculation of emoji sequences, use EmojiSequenceWidth().
//
// Example:
//
//	uts51.EmojiWidth('😀')              // 2 - emoji presentation
//	uts51.EmojiWidth('☺')               // 1 - text presentation
//	uts51.EmojiWidth('\U0001F3FB')      // 0 - skin tone modifier
func EmojiWidth(r rune) int {
	// Emoji components are zero-width
	if IsEmojiComponent(r) {
		return 0
	}

	// Variation selectors are zero-width
	if r == VariationSelector15 || r == VariationSelector16 {
		return 0
	}

	// ZWJ is zero-width
	if r == ZeroWidthJoiner {
		return 0
	}

	// Emoji with emoji presentation are 2 columns wide (like CJK)
	if HasEmojiPresentation(r) {
		return 2
	}

	// Emoji without emoji presentation default to 1 column
	if IsEmoji(r) {
		return 1
	}

	// Not an emoji
	return 0
}

// DefaultPresentation returns the default presentation for an emoji character.
//
// Per UTS #51 §2.1: https://www.unicode.org/reports/tr51/#Emoji_Presentation
//
// Returns:
//   - 'E' for emoji presentation (colorful, 2 columns)
//   - 'T' for text presentation (monochrome, 1 column)
//
// Example:
//
//	uts51.DefaultPresentation('😀')  // 'E' - emoji by default
//	uts51.DefaultPresentation('☺')   // 'T' - text by default
func DefaultPresentation(r rune) rune {
	if HasEmojiPresentation(r) {
		return 'E'
	}
	if IsEmoji(r) {
		return 'T'
	}
	return 'T' // Non-emoji default to text
}

// IsValidKeycapSequence validates an emoji keycap sequence.
//
// Per UTS #51 §2.3: https://www.unicode.org/reports/tr51/#def_emoji_keycap_sequence
// "An emoji keycap sequence is a sequence of the following form:
// keycap_base + \uFE0F + \u20E3"
//
// Where keycap_base is one of: [0-9#*]
//
// Example:
//
//	uts51.IsValidKeycapSequence([]rune{'9', '\uFE0F', '\u20E3'})  // true - "9⃣"
//	uts51.IsValidKeycapSequence([]rune{'#', '\uFE0F', '\u20E3'})  // true - "#⃣"
func IsValidKeycapSequence(runes []rune) bool {
	if len(runes) < 2 {
		return false
	}

	// Check for keycap base: [0-9#*]
	base := runes[0]
	if !((base >= '0' && base <= '9') || base == '#' || base == '*') {
		return false
	}

	// For fully-qualified keycap sequences: base + FE0F + 20E3
	if len(runes) == 3 {
		return runes[1] == VariationSelector16 && runes[2] == CombiningEnclosingKeycap
	}

	// For minimally-qualified keycap sequences: base + 20E3 (no FE0F)
	if len(runes) == 2 {
		return runes[1] == CombiningEnclosingKeycap
	}

	return false
}

// IsValidTagSequence validates an emoji tag sequence.
//
// Per UTS #51 §2.6: https://www.unicode.org/reports/tr51/#def_emoji_tag_sequence
// "An emoji tag sequence is a sequence of the following form:
// tag_base + tag_spec + tag_term"
//
// Where:
//   - tag_base is a character with Emoji=Yes
//   - tag_spec is one or more tag characters (U+E0020..U+E007E)
//   - tag_term is U+E007F (cancel tag)
//
// Example:
//
//	// England flag: 🏴 + tag chars for "gbeng" + tag terminator
//	uts51.IsValidTagSequence([]rune{0x1F3F4, 0xE0067, 0xE0062, 0xE0065, 0xE006E, 0xE0067, 0xE007F})  // true
func IsValidTagSequence(runes []rune) bool {
	if len(runes) < 3 {
		return false
	}

	// First character must be emoji
	if !IsEmoji(runes[0]) {
		return false
	}

	// Last character must be tag terminator
	if runes[len(runes)-1] != TagTerminator {
		return false
	}

	// Middle characters must be tag characters
	for i := 1; i < len(runes)-1; i++ {
		if !IsTagCharacter(runes[i]) {
			return false
		}
	}

	return true
}

// IsValidEmojiSequence validates any type of emoji sequence.
//
// This is a convenience function that checks for all sequence types
// defined in UTS #51 §2: https://www.unicode.org/reports/tr51/#Emoji_Sequences
//
// Supported sequence types:
//   - Keycap sequences ([0-9#*] + U+FE0F + U+20E3)
//   - Tag sequences (base + tag chars + U+E007F)
//   - Modifier sequences (base + skin tone)
//   - ZWJ sequences (characters joined by U+200D)
//   - Flag sequences (two regional indicators)
//
// Example:
//
//	uts51.IsValidEmojiSequence([]rune{'9', '\uFE0F', '\u20E3'})  // true - keycap
//	uts51.IsValidEmojiSequence([]rune{0x1F44B, 0x1F3FB})         // true - waving hand + light skin
func IsValidEmojiSequence(runes []rune) bool {
	if len(runes) == 0 {
		return false
	}

	// Single character - just check if it's emoji
	if len(runes) == 1 {
		return IsEmoji(runes[0])
	}

	// Check for keycap sequence
	if IsValidKeycapSequence(runes) {
		return true
	}

	// Check for tag sequence
	if IsValidTagSequence(runes) {
		return true
	}

	// Check for modifier sequence (base + modifier)
	if len(runes) == 2 && IsEmojiModifierBase(runes[0]) && IsEmojiModifier(runes[1]) {
		return true
	}

	// Check for presentation sequence (emoji + variation selector)
	if len(runes) == 2 {
		if IsEmoji(runes[0]) && (runes[1] == VariationSelector15 || runes[1] == VariationSelector16) {
			return true
		}
	}

	// Check for flag sequence (two regional indicators)
	if len(runes) == 2 && IsRegionalIndicator(runes[0]) && IsRegionalIndicator(runes[1]) {
		return true
	}

	// Check for ZWJ sequence - contains at least one ZWJ
	hasZWJ := false
	for _, r := range runes {
		if r == ZeroWidthJoiner {
			hasZWJ = true
			break
		}
	}
	if hasZWJ {
		// Basic validation: all non-ZWJ, non-VS characters should be emoji-related
		for _, r := range runes {
			if r == ZeroWidthJoiner || r == VariationSelector15 || r == VariationSelector16 {
				continue
			}
			if !IsEmoji(r) && !IsEmojiModifier(r) {
				return false
			}
		}
		return true
	}

	return false
}
