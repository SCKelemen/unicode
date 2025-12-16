package uax29

// This file defines the break class types used in UAX #29 Text Segmentation.
// These classes are used both separately and in packed form for efficient
// single-pass break detection.

// GraphemeBreakClass represents the Grapheme_Cluster_Break property values
// defined in UAX #29 Table 2.
//
// See: https://www.unicode.org/reports/tr29/#Table_Grapheme_Cluster_Break_Property_Values
type GraphemeBreakClass uint8

const (
	GBOther GraphemeBreakClass = iota
	GBCR
	GBLF
	GBControl
	GBExtend
	GBZWJ
	GBRegionalIndicator
	GBPrepend
	GBSpacingMark
	GBL
	GBV
	GBT
	GBLV
	GBLVT
)

// Aliases for use in generator (with underscores)
const (
	GB_Other              = GBOther
	GB_CR                 = GBCR
	GB_LF                 = GBLF
	GB_Control            = GBControl
	GB_Extend             = GBExtend
	GB_ZWJ                = GBZWJ
	GB_Regional_Indicator = GBRegionalIndicator
	GB_Prepend            = GBPrepend
	GB_SpacingMark        = GBSpacingMark
	GB_L                  = GBL
	GB_V                  = GBV
	GB_T                  = GBT
	GB_LV                 = GBLV
	GB_LVT                = GBLVT
)

// WordBreakClass represents the Word_Break property values
// defined in UAX #29 Table 3.
//
// See: https://www.unicode.org/reports/tr29/#Table_Word_Break_Property_Values
type WordBreakClass uint8

const (
	WBOther WordBreakClass = iota
	WBCR
	WBLF
	WBNewline
	WBExtend
	WBZWJ
	WBRegionalIndicator
	WBFormat
	WBKatakana
	WBHebrewLetter
	WBALetter
	WBSingleQuote
	WBDoubleQuote
	WBMidNumLet
	WBMidLetter
	WBMidNum
	WBNumeric
	WBExtendNumLet
	WBWSegSpace
	WBExtendedPictographic
)

// Aliases for use in generator (with underscores)
const (
	WB_Other              = WBOther
	WB_CR                 = WBCR
	WB_LF                 = WBLF
	WB_Newline            = WBNewline
	WB_Extend             = WBExtend
	WB_ZWJ                = WBZWJ
	WB_Regional_Indicator = WBRegionalIndicator
	WB_Format             = WBFormat
	WB_Katakana           = WBKatakana
	WB_Hebrew_Letter      = WBHebrewLetter
	WB_ALetter            = WBALetter
	WB_Single_Quote       = WBSingleQuote
	WB_Double_Quote       = WBDoubleQuote
	WB_MidNumLet          = WBMidNumLet
	WB_MidLetter          = WBMidLetter
	WB_MidNum             = WBMidNum
	WB_Numeric             = WBNumeric
	WB_ExtendNumLet        = WBExtendNumLet
	WB_WSegSpace           = WBWSegSpace
	WB_ExtendedPictographic = WBExtendedPictographic
)

// SentenceBreakClass represents the Sentence_Break property values
// defined in UAX #29 Table 4.
//
// See: https://www.unicode.org/reports/tr29/#Table_Sentence_Break_Property_Values
type SentenceBreakClass uint8

const (
	SBOther SentenceBreakClass = iota
	SBCR
	SBLF
	SBExtend
	SBSep
	SBFormat
	SBSp
	SBLower
	SBUpper
	SBOLetter
	SBNumeric
	SBATerm
	SBSTerm
	SBClose
	SBSContinue
)

// Aliases for use in generator (with underscores)
const (
	SB_Other     = SBOther
	SB_CR        = SBCR
	SB_LF        = SBLF
	SB_Extend    = SBExtend
	SB_Sep       = SBSep
	SB_Format    = SBFormat
	SB_Sp        = SBSp
	SB_Lower     = SBLower
	SB_Upper     = SBUpper
	SB_OLetter   = SBOLetter
	SB_Numeric   = SBNumeric
	SB_ATerm     = SBATerm
	SB_STerm     = SBSTerm
	SB_Close     = SBClose
	SB_SContinue = SBSContinue
)
