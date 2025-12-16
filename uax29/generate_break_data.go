// +build ignore

// This program generates break_data.go from Unicode property files
// Run with: go run generate_break_data.go

package main

import (
	"bufio"
	"fmt"
	"io"
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
	unicodeVersion = "17.0.0"
)

type runeRange struct {
	start rune
	end   rune
}

type properties struct {
	grapheme string
	word     string
	sentence string
}

// propertyMap stores property string by rune range
type propertyMap map[runeRange]string

func main() {
	fmt.Println("Downloading Unicode property files...")

	// Download and parse property files
	graphemeProps, err := downloadAndParse(graphemeURL, "GraphemeBreakProperty")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing grapheme properties: %v\n", err)
		os.Exit(1)
	}

	wordProps, err := downloadAndParse(wordURL, "WordBreakProperty")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing word properties: %v\n", err)
		os.Exit(1)
	}

	sentenceProps, err := downloadAndParse(sentenceURL, "SentenceBreakProperty")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing sentence properties: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Parsed %d grapheme ranges, %d word ranges, %d sentence ranges\n",
		len(graphemeProps), len(wordProps), len(sentenceProps))

	// Merge properties into unified ranges
	fmt.Println("Merging properties...")
	merged := mergeProperties(graphemeProps, wordProps, sentenceProps)
	fmt.Printf("Created %d unified ranges\n", len(merged))

	// Generate output file
	fmt.Println("Generating break_data.go...")
	if err := generateCode(merged); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully generated break_data.go")
}

func downloadAndParse(url, name string) (propertyMap, error) {
	// Try to download
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download %s: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to download %s: status %d", name, resp.StatusCode)
	}

	return parsePropertyFile(resp.Body)
}

func parsePropertyFile(r io.Reader) (propertyMap, error) {
	props := make(propertyMap)
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on semicolon: code_point(s) ; property # comment
		parts := strings.Split(line, ";")
		if len(parts) < 2 {
			continue
		}

		codePoints := strings.TrimSpace(parts[0])
		property := strings.TrimSpace(parts[1])

		// Extract just the property name (before any comment)
		property = strings.Fields(property)[0]

		// Parse code point or range
		var start, end rune
		if strings.Contains(codePoints, "..") {
			// Range: 0000..001F
			rangeParts := strings.Split(codePoints, "..")
			startVal, err := strconv.ParseInt(rangeParts[0], 16, 32)
			if err != nil {
				continue
			}
			endVal, err := strconv.ParseInt(rangeParts[1], 16, 32)
			if err != nil {
				continue
			}
			start = rune(startVal)
			end = rune(endVal)
		} else {
			// Single code point: 0020
			val, err := strconv.ParseInt(codePoints, 16, 32)
			if err != nil {
				continue
			}
			start = rune(val)
			end = rune(val)
		}

		props[runeRange{start, end}] = property
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return props, nil
}

func mergeProperties(g, w, s propertyMap) []mergedRange {
	// Collect all unique boundary points from all three property maps
	boundarySet := make(map[rune]bool)

	for rng := range g {
		boundarySet[rng.start] = true
		if rng.end < 0x10FFFF {
			boundarySet[rng.end+1] = true
		}
	}
	for rng := range w {
		boundarySet[rng.start] = true
		if rng.end < 0x10FFFF {
			boundarySet[rng.end+1] = true
		}
	}
	for rng := range s {
		boundarySet[rng.start] = true
		if rng.end < 0x10FFFF {
			boundarySet[rng.end+1] = true
		}
	}

	// Convert to sorted slice
	boundaries := make([]rune, 0, len(boundarySet))
	for b := range boundarySet {
		boundaries = append(boundaries, b)
	}
	sort.Slice(boundaries, func(i, j int) bool {
		return boundaries[i] < boundaries[j]
	})

	// Build merged ranges
	var merged []mergedRange
	for i := 0; i < len(boundaries)-1; i++ {
		start := boundaries[i]
		end := boundaries[i+1] - 1

		// Look up properties for this range
		gProp := lookupProperty(g, start, "Other")
		wProp := lookupProperty(w, start, "Other")
		sProp := lookupProperty(s, start, "Other")

		merged = append(merged, mergedRange{
			start: start,
			end:   end,
			props: properties{
				grapheme: gProp,
				word:     wProp,
				sentence: sProp,
			},
		})
	}

	return merged
}

func lookupProperty(pm propertyMap, r rune, defaultProp string) string {
	for rng, prop := range pm {
		if r >= rng.start && r <= rng.end {
			return prop
		}
	}
	return defaultProp
}

type mergedRange struct {
	start rune
	end   rune
	props properties
}

func generateCode(ranges []mergedRange) error {
	out, err := os.Create("break_data.go")
	if err != nil {
		return err
	}
	defer out.Close()

	// File header
	fmt.Fprintf(out, "// Code generated by generate_break_data.go. DO NOT EDIT.\n")
	fmt.Fprintf(out, "// Unicode Version: %s\n", unicodeVersion)
	fmt.Fprintf(out, "// Generated: 2025-12-16\n")
	fmt.Fprintf(out, "\npackage uax29\n\n")

	// Type definitions
	fmt.Fprintf(out, "// PackedBreakClass encodes all three break classes in 16 bits\n")
	fmt.Fprintf(out, "// Layout: [grapheme:5][word:5][sentence:4][reserved:2]\n")
	fmt.Fprintf(out, "type PackedBreakClass uint16\n\n")

	fmt.Fprintf(out, "type breakRange struct {\n")
	fmt.Fprintf(out, "\tstart rune\n")
	fmt.Fprintf(out, "\tend   rune\n")
	fmt.Fprintf(out, "\tclass PackedBreakClass\n")
	fmt.Fprintf(out, "}\n\n")

	// Data table
	fmt.Fprintf(out, "// breakData contains all Unicode break properties for UAX #29\n")
	fmt.Fprintf(out, "// Total ranges: %d\n", len(ranges))
	fmt.Fprintf(out, "var breakData = []breakRange{\n")

	for _, r := range ranges {
		graphemeClass := mapGraphemeClass(r.props.grapheme)
		wordClass := mapWordClass(r.props.word)
		sentenceClass := mapSentenceClass(r.props.sentence)

		comment := ""
		if r.start == r.end {
			comment = fmt.Sprintf("// U+%04X", r.start)
		} else {
			comment = fmt.Sprintf("// U+%04X..U+%04X", r.start, r.end)
		}
		if r.props.grapheme != "Other" || r.props.word != "Other" || r.props.sentence != "Other" {
			comment += fmt.Sprintf(" G=%s W=%s S=%s", r.props.grapheme, r.props.word, r.props.sentence)
		}

		fmt.Fprintf(out, "\t{0x%04X, 0x%04X, packClasses(GB_%s, WB_%s, SB_%s)}, %s\n",
			r.start, r.end, graphemeClass, wordClass, sentenceClass, comment)
	}

	fmt.Fprintf(out, "}\n\n")

	// Helper function
	fmt.Fprintf(out, "func packClasses(g GraphemeBreakClass, w WordBreakClass, s SentenceBreakClass) PackedBreakClass {\n")
	fmt.Fprintf(out, "\treturn PackedBreakClass(uint16(g) | (uint16(w) << 5) | (uint16(s) << 10))\n")
	fmt.Fprintf(out, "}\n\n")

	// Unpack methods
	fmt.Fprintf(out, "func (p PackedBreakClass) Grapheme() GraphemeBreakClass {\n")
	fmt.Fprintf(out, "\treturn GraphemeBreakClass(p & 0x1F)\n")
	fmt.Fprintf(out, "}\n\n")

	fmt.Fprintf(out, "func (p PackedBreakClass) Word() WordBreakClass {\n")
	fmt.Fprintf(out, "\treturn WordBreakClass((p & 0x3E0) >> 5)\n")
	fmt.Fprintf(out, "}\n\n")

	fmt.Fprintf(out, "func (p PackedBreakClass) Sentence() SentenceBreakClass {\n")
	fmt.Fprintf(out, "\treturn SentenceBreakClass((p & 0x3C00) >> 10)\n")
	fmt.Fprintf(out, "}\n\n")

	// Classification function
	fmt.Fprintf(out, "// classifyRune returns the packed break classes for a rune using binary search\n")
	fmt.Fprintf(out, "func classifyRune(r rune) PackedBreakClass {\n")
	fmt.Fprintf(out, "\tleft, right := 0, len(breakData)-1\n")
	fmt.Fprintf(out, "\tfor left <= right {\n")
	fmt.Fprintf(out, "\t\tmid := (left + right) / 2\n")
	fmt.Fprintf(out, "\t\tentry := breakData[mid]\n")
	fmt.Fprintf(out, "\t\tif r < entry.start {\n")
	fmt.Fprintf(out, "\t\t\tright = mid - 1\n")
	fmt.Fprintf(out, "\t\t} else if r > entry.end {\n")
	fmt.Fprintf(out, "\t\t\tleft = mid + 1\n")
	fmt.Fprintf(out, "\t\t} else {\n")
	fmt.Fprintf(out, "\t\t\treturn entry.class\n")
	fmt.Fprintf(out, "\t\t}\n")
	fmt.Fprintf(out, "\t}\n")
	fmt.Fprintf(out, "\treturn packClasses(GB_Other, WB_Other, SB_Other)\n")
	fmt.Fprintf(out, "}\n")

	return nil
}

// Map Unicode property names to Go constant names
func mapGraphemeClass(prop string) string {
	switch prop {
	case "CR":
		return "CR"
	case "LF":
		return "LF"
	case "Control":
		return "Control"
	case "Extend":
		return "Extend"
	case "ZWJ":
		return "ZWJ"
	case "Regional_Indicator":
		return "Regional_Indicator"
	case "Prepend":
		return "Prepend"
	case "SpacingMark":
		return "SpacingMark"
	case "L":
		return "L"
	case "V":
		return "V"
	case "T":
		return "T"
	case "LV":
		return "LV"
	case "LVT":
		return "LVT"
	default:
		return "Other"
	}
}

func mapWordClass(prop string) string {
	switch prop {
	case "CR":
		return "CR"
	case "LF":
		return "LF"
	case "Newline":
		return "Newline"
	case "Extend":
		return "Extend"
	case "ZWJ":
		return "ZWJ"
	case "Regional_Indicator":
		return "Regional_Indicator"
	case "Format":
		return "Format"
	case "Katakana":
		return "Katakana"
	case "Hebrew_Letter":
		return "Hebrew_Letter"
	case "ALetter":
		return "ALetter"
	case "Single_Quote":
		return "Single_Quote"
	case "Double_Quote":
		return "Double_Quote"
	case "MidNumLet":
		return "MidNumLet"
	case "MidLetter":
		return "MidLetter"
	case "MidNum":
		return "MidNum"
	case "Numeric":
		return "Numeric"
	case "ExtendNumLet":
		return "ExtendNumLet"
	case "WSegSpace":
		return "WSegSpace"
	default:
		return "Other"
	}
}

func mapSentenceClass(prop string) string {
	switch prop {
	case "CR":
		return "CR"
	case "LF":
		return "LF"
	case "Extend":
		return "Extend"
	case "Sep":
		return "Sep"
	case "Format":
		return "Format"
	case "Sp":
		return "Sp"
	case "Lower":
		return "Lower"
	case "Upper":
		return "Upper"
	case "OLetter":
		return "OLetter"
	case "Numeric":
		return "Numeric"
	case "ATerm":
		return "ATerm"
	case "STerm":
		return "STerm"
	case "Close":
		return "Close"
	case "SContinue":
		return "SContinue"
	default:
		return "Other"
	}
}
