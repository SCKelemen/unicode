// +build ignore

// This program generates normalization_data.go from Unicode data files
// Run with: go run generate_normalization_data.go

package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	unicodeDataURL       = "https://www.unicode.org/Public/17.0.0/ucd/UnicodeData.txt"
	compositionExclURL   = "https://www.unicode.org/Public/17.0.0/ucd/CompositionExclusions.txt"
)

type decompositionData struct {
	canonical    map[rune][]rune
	compatibility map[rune][]rune
	combiningClass map[rune]int
	compositions   map[[2]rune]rune
	exclusions     map[rune]bool
}

func main() {
	data := &decompositionData{
		canonical:      make(map[rune][]rune),
		compatibility:  make(map[rune][]rune),
		combiningClass: make(map[rune]int),
		compositions:   make(map[[2]rune]rune),
		exclusions:     make(map[rune]bool),
	}

	// Download and parse UnicodeData.txt
	fmt.Println("Downloading UnicodeData.txt...")
	if err := downloadAndParseUnicodeData(unicodeDataURL, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Download and parse CompositionExclusions.txt
	fmt.Println("Downloading CompositionExclusions.txt...")
	if err := downloadAndParseExclusions(compositionExclURL, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Generate composition pairs (reverse of decomposition)
	generateCompositions(data)

	// Generate the output file
	fmt.Println("Generating normalization_data.go...")
	if err := generateFile(data); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated normalization_data.go successfully\n")
	fmt.Printf("  Canonical decompositions: %d\n", len(data.canonical))
	fmt.Printf("  Compatibility decompositions: %d\n", len(data.compatibility))
	fmt.Printf("  Combining classes: %d\n", len(data.combiningClass))
	fmt.Printf("  Composition pairs: %d\n", len(data.compositions))
	fmt.Printf("  Composition exclusions: %d\n", len(data.exclusions))
}

func downloadAndParseUnicodeData(url string, data *decompositionData) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return parseUnicodeData(resp.Body, data)
}

func parseUnicodeData(r io.Reader, data *decompositionData) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Split(line, ";")
		if len(fields) < 6 {
			continue
		}

		// Parse code point
		cpVal, err := strconv.ParseInt(fields[0], 16, 32)
		if err != nil {
			continue
		}
		cp := rune(cpVal)

		// Parse combining class (field 3)
		if fields[3] != "" && fields[3] != "0" {
			class, err := strconv.Atoi(fields[3])
			if err == nil && class > 0 {
				data.combiningClass[cp] = class
			}
		}

		// Parse decomposition mapping (field 5)
		decompField := strings.TrimSpace(fields[5])
		if decompField == "" {
			continue
		}

		// Check if it's a compatibility decomposition
		isCompatibility := strings.HasPrefix(decompField, "<")
		if isCompatibility {
			// Remove compatibility tag
			endTag := strings.Index(decompField, ">")
			if endTag > 0 {
				decompField = strings.TrimSpace(decompField[endTag+1:])
			}
		}

		// Parse the decomposition sequence
		parts := strings.Fields(decompField)
		if len(parts) == 0 {
			continue
		}

		decomp := make([]rune, 0, len(parts))
		for _, part := range parts {
			val, err := strconv.ParseInt(part, 16, 32)
			if err != nil {
				continue
			}
			decomp = append(decomp, rune(val))
		}

		if len(decomp) > 0 {
			if isCompatibility {
				data.compatibility[cp] = decomp
			} else {
				data.canonical[cp] = decomp
			}
		}
	}

	return scanner.Err()
}

func downloadAndParseExclusions(url string, data *decompositionData) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse code point (before any comment)
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cpVal, err := strconv.ParseInt(parts[0], 16, 32)
		if err != nil {
			continue
		}

		data.exclusions[rune(cpVal)] = true
	}

	return scanner.Err()
}

func generateCompositions(data *decompositionData) {
	// Generate composition pairs from canonical decompositions
	for cp, decomp := range data.canonical {
		// Only two-character decompositions can compose
		if len(decomp) != 2 {
			continue
		}

		// Skip if in exclusion list
		if data.exclusions[cp] {
			continue
		}

		// Skip if the first character is a combining mark
		if data.combiningClass[decomp[0]] != 0 {
			continue
		}

		// Add to composition map
		data.compositions[[2]rune{decomp[0], decomp[1]}] = cp
	}
}

func generateFile(data *decompositionData) error {
	out, err := os.Create("normalization_data.go")
	if err != nil {
		return err
	}
	defer out.Close()

	fmt.Fprintf(out, "// Code generated by generate_normalization_data.go DO NOT EDIT.\n")
	fmt.Fprintf(out, "// Source: Unicode 17.0.0\n")
	fmt.Fprintf(out, "\npackage uts15\n\n")

	// Canonical decomposition map
	fmt.Fprintf(out, "// canonicalDecompositionMap maps runes to their canonical decomposition\n")
	fmt.Fprintf(out, "var canonicalDecompositionMap = map[rune][]rune{\n")
	for cp, decomp := range data.canonical {
		fmt.Fprintf(out, "\t0x%04X: {", cp)
		for i, r := range decomp {
			if i > 0 {
				fmt.Fprintf(out, ", ")
			}
			fmt.Fprintf(out, "0x%04X", r)
		}
		fmt.Fprintf(out, "},\n")
	}
	fmt.Fprintf(out, "}\n\n")

	// Compatibility decomposition map
	fmt.Fprintf(out, "// compatibilityDecompositionMap maps runes to their compatibility decomposition\n")
	fmt.Fprintf(out, "var compatibilityDecompositionMap = map[rune][]rune{\n")
	for cp, decomp := range data.compatibility {
		fmt.Fprintf(out, "\t0x%04X: {", cp)
		for i, r := range decomp {
			if i > 0 {
				fmt.Fprintf(out, ", ")
			}
			fmt.Fprintf(out, "0x%04X", r)
		}
		fmt.Fprintf(out, "},\n")
	}
	fmt.Fprintf(out, "}\n\n")

	// Combining class map
	fmt.Fprintf(out, "// combiningClassMap maps runes to their canonical combining class\n")
	fmt.Fprintf(out, "var combiningClassMap = map[rune]int{\n")
	for cp, class := range data.combiningClass {
		fmt.Fprintf(out, "\t0x%04X: %d,\n", cp, class)
	}
	fmt.Fprintf(out, "}\n\n")

	// Composition map
	fmt.Fprintf(out, "// compositionMap maps pairs of runes to their composed form\n")
	fmt.Fprintf(out, "var compositionMap = map[[2]rune]rune{\n")
	for pair, cp := range data.compositions {
		fmt.Fprintf(out, "\t{0x%04X, 0x%04X}: 0x%04X,\n", pair[0], pair[1], cp)
	}
	fmt.Fprintf(out, "}\n")

	return nil
}
