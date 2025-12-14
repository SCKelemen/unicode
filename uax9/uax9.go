// Package uax9 implements the Unicode Bidirectional Algorithm (UAX #9).
//
// This package provides bidirectional text reordering for proper display of text
// containing both left-to-right (LTR) and right-to-left (RTL) scripts, such as
// mixing Latin with Arabic or Hebrew text.
//
// Based on: https://www.unicode.org/reports/tr9/
//
// Status: Not yet implemented
//
// Usage (planned):
//
//	import "github.com/SCKelemen/unicode/uax9"
//
//	text := "Hello مرحبا world"
//	result := uax9.Reorder(text, uax9.DirectionLTR)
//	// Returns properly reordered text for display
package uax9

// Direction represents the base text direction.
type Direction int

const (
	// DirectionLTR indicates left-to-right base direction (e.g., English)
	DirectionLTR Direction = iota
	// DirectionRTL indicates right-to-left base direction (e.g., Arabic, Hebrew)
	DirectionRTL
	// DirectionAuto automatically determines base direction from content
	DirectionAuto
)

// TODO: Implement the Unicode Bidirectional Algorithm
// Key components needed:
// - Bidi character types (L, R, AL, EN, ES, etc.)
// - Bidi level resolution
// - Explicit formatting characters (LRE, RLE, PDF, etc.)
// - Reordering for display
// - Mirror glyph handling
