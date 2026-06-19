package ucase

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
)

const (
	unicodeDataURL      = "https://www.unicode.org/Public/17.0.0/ucd/UnicodeData.txt"
	specialCasingURL    = "https://www.unicode.org/Public/17.0.0/ucd/SpecialCasing.txt"
	caseFoldingURL      = "https://www.unicode.org/Public/17.0.0/ucd/CaseFolding.txt"
	derivedCorePropsURL = "https://www.unicode.org/Public/17.0.0/ucd/DerivedCoreProperties.txt"
)

// ucdData holds the parsed mappings the package is expected to reproduce for
// the default (language-insensitive) algorithm.
type ucdData struct {
	simpleUp, simpleLo, simpleTi    map[rune]rune   // UnicodeData.txt fields 12/13/14
	specialLo, specialTi, specialUp map[rune]string // SpecialCasing.txt, unconditional, lang-neutral
	foldFull                        map[rune]string // CaseFolding.txt status C+F
	foldSimple                      map[rune]rune   // CaseFolding.txt status C+S
}

func hexRune(t *testing.T, s string) rune {
	t.Helper()
	v, err := strconv.ParseInt(strings.TrimSpace(s), 16, 32)
	if err != nil {
		t.Fatalf("bad hex %q: %v", s, err)
	}
	return rune(v)
}

func hexRunes(t *testing.T, s string) []rune {
	t.Helper()
	fields := strings.Fields(s)
	out := make([]rune, 0, len(fields))
	for _, f := range fields {
		out = append(out, hexRune(t, f))
	}
	return out
}

func parseUCD(t *testing.T, unicodeData, specialCasing, caseFolding io.Reader) *ucdData {
	t.Helper()
	d := &ucdData{
		simpleUp:   map[rune]rune{},
		simpleLo:   map[rune]rune{},
		simpleTi:   map[rune]rune{},
		specialLo:  map[rune]string{},
		specialTi:  map[rune]string{},
		specialUp:  map[rune]string{},
		foldFull:   map[rune]string{},
		foldSimple: map[rune]rune{},
	}

	// UnicodeData.txt: simple mappings (fields 12/13/14).
	sc := bufio.NewScanner(unicodeData)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Split(line, ";")
		if len(f) < 15 {
			continue
		}
		cp := hexRune(t, f[0])
		if s := strings.TrimSpace(f[12]); s != "" {
			d.simpleUp[cp] = hexRune(t, s)
		}
		if s := strings.TrimSpace(f[13]); s != "" {
			d.simpleLo[cp] = hexRune(t, s)
		}
		if s := strings.TrimSpace(f[14]); s != "" {
			d.simpleTi[cp] = hexRune(t, s)
		} else if up, ok := d.simpleUp[cp]; ok {
			d.simpleTi[cp] = up // UAX #44 default
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan UnicodeData: %v", err)
	}

	// SpecialCasing.txt: unconditional, language-neutral full mappings only.
	sc = bufio.NewScanner(specialCasing)
	for sc.Scan() {
		line := sc.Text()
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		f := strings.Split(line, ";")
		if len(f) < 4 {
			continue
		}
		if len(f) >= 5 && strings.TrimSpace(f[4]) != "" {
			// Skip every conditional and language-sensitive line: for an
			// isolated character no context condition fires, so the expected
			// mapping is the unconditional/simple one.
			continue
		}
		cp := hexRune(t, f[0])
		d.specialLo[cp] = string(hexRunes(t, f[1]))
		d.specialTi[cp] = string(hexRunes(t, f[2]))
		d.specialUp[cp] = string(hexRunes(t, f[3]))
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan SpecialCasing: %v", err)
	}

	// CaseFolding.txt: full (C+F) and simple (C+S).
	sc = bufio.NewScanner(caseFolding)
	for sc.Scan() {
		line := sc.Text()
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		f := strings.Split(line, ";")
		if len(f) < 3 {
			continue
		}
		cp := hexRune(t, f[0])
		status := strings.TrimSpace(f[1])
		target := hexRunes(t, f[2])
		if len(target) == 0 {
			continue
		}
		switch status {
		case "C", "F":
			d.foldFull[cp] = string(target)
		}
		switch status {
		case "C", "S":
			d.foldSimple[cp] = target[0]
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan CaseFolding: %v", err)
	}
	return d
}

// verify checks that the package reproduces every parsed mapping exactly for
// the default, language-insensitive algorithm.
func verify(t *testing.T, d *ucdData) {
	t.Helper()

	// Simple, context-free rune mappings.
	for cp, want := range d.simpleLo {
		if got := SimpleLower(cp); got != want {
			t.Errorf("SimpleLower(U+%04X) = U+%04X, want U+%04X", cp, got, want)
		}
	}
	for cp, want := range d.simpleUp {
		if got := SimpleUpper(cp); got != want {
			t.Errorf("SimpleUpper(U+%04X) = U+%04X, want U+%04X", cp, got, want)
		}
	}
	for cp, want := range d.simpleTi {
		if got := SimpleTitle(cp); got != want {
			t.Errorf("SimpleTitle(U+%04X) = U+%04X, want U+%04X", cp, got, want)
		}
	}
	for cp, want := range d.foldSimple {
		if got := SimpleFold(cp); got != want {
			t.Errorf("SimpleFold(U+%04X) = U+%04X, want U+%04X", cp, got, want)
		}
	}

	// Full case folding.
	for cp, want := range d.foldFull {
		if got := ToFold(string(cp)); got != want {
			t.Errorf("ToFold(U+%04X) = %q, want %q", cp, got, want)
		}
	}

	// Full lower/upper of an isolated character: special unconditional mapping
	// wins over the simple mapping; otherwise the simple mapping (as a string)
	// applies; otherwise identity.
	expectFull := func(cp rune, special map[rune]string, simple map[rune]rune) string {
		if s, ok := special[cp]; ok {
			return s
		}
		if r, ok := simple[cp]; ok {
			return string(r)
		}
		return string(cp)
	}
	seen := map[rune]bool{}
	check := func(cp rune) {
		if seen[cp] {
			return
		}
		seen[cp] = true
		if got, want := ToLower(string(cp)), expectFull(cp, d.specialLo, d.simpleLo); got != want {
			t.Errorf("ToLower(U+%04X) = %q, want %q", cp, got, want)
		}
		if got, want := ToUpper(string(cp)), expectFull(cp, d.specialUp, d.simpleUp); got != want {
			t.Errorf("ToUpper(U+%04X) = %q, want %q", cp, got, want)
		}
	}
	for cp := range d.simpleLo {
		check(cp)
	}
	for cp := range d.simpleUp {
		check(cp)
	}
	for cp := range d.specialLo {
		check(cp)
	}
	for cp := range d.specialUp {
		check(cp)
	}
}

// parseProperty parses the inclusive ranges of a binary property (e.g. "Cased")
// from DerivedCoreProperties.txt.
func parseProperty(t *testing.T, r io.Reader, name string) [][2]rune {
	t.Helper()
	var out [][2]rune
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		f := strings.Split(line, ";")
		if len(f) < 2 || strings.TrimSpace(f[1]) != name {
			continue
		}
		span := strings.TrimSpace(f[0])
		if i := strings.Index(span, ".."); i >= 0 {
			out = append(out, [2]rune{hexRune(t, span[:i]), hexRune(t, span[i+2:])})
		} else {
			v := hexRune(t, span)
			out = append(out, [2]rune{v, v})
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan DerivedCoreProperties: %v", err)
	}
	return out
}

// verifyProperty checks pred against the full Unicode code space using the
// expected ranges, catching both missing entries and false positives.
func verifyProperty(t *testing.T, name string, ranges [][2]rune, pred func(rune) bool) {
	t.Helper()
	want := make(map[rune]bool)
	for _, rg := range ranges {
		for r := rg[0]; r <= rg[1]; r++ {
			want[r] = true
		}
	}
	mismatches := 0
	for r := rune(0); r <= 0x10FFFF; r++ {
		if pred(r) != want[r] {
			if mismatches < 10 {
				t.Errorf("%s(U+%04X) = %v, want %v", name, r, pred(r), want[r])
			}
			mismatches++
		}
	}
	if mismatches > 0 {
		t.Errorf("%s: %d total mismatches", name, mismatches)
	}
}

func openLocal(t *testing.T, name string) (*os.File, bool) {
	t.Helper()
	f, err := os.Open(name)
	if err != nil {
		t.Skipf("skipping (missing %s): %v", name, err)
		return nil, false
	}
	return f, true
}

// TestDataFileConsistency validates the generated tables against the committed
// UCD files. It is skipped gracefully if any file is missing.
func TestDataFileConsistency(t *testing.T) {
	ud, ok := openLocal(t, "UnicodeData.txt")
	if !ok {
		return
	}
	defer ud.Close()
	scf, ok := openLocal(t, "SpecialCasing.txt")
	if !ok {
		return
	}
	defer scf.Close()
	cf, ok := openLocal(t, "CaseFolding.txt")
	if !ok {
		return
	}
	defer cf.Close()
	dcp, ok := openLocal(t, "DerivedCoreProperties.txt")
	if !ok {
		return
	}
	defer dcp.Close()

	verify(t, parseUCD(t, ud, scf, cf))
	verifyProperty(t, "IsCased", parseProperty(t, dcp, "Cased"), IsCased)
	// Re-open for the second property scan.
	dcp2, ok := openLocal(t, "DerivedCoreProperties.txt")
	if !ok {
		return
	}
	defer dcp2.Close()
	verifyProperty(t, "IsCaseIgnorable", parseProperty(t, dcp2, "Case_Ignorable"), IsCaseIgnorable)
}

// TestOfficialConformance downloads the official Unicode 17.0.0 UCD files and
// verifies the package reproduces every default mapping. Skipped in short mode
// or when offline.
func TestOfficialConformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping official conformance test in short mode")
	}
	get := func(url string) io.ReadCloser {
		resp, err := http.Get(url)
		if err != nil {
			t.Skipf("skipping official conformance (download unavailable): %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Skipf("skipping official conformance (HTTP %d for %s)", resp.StatusCode, url)
		}
		return resp.Body
	}
	ud := get(unicodeDataURL)
	defer ud.Close()
	scf := get(specialCasingURL)
	defer scf.Close()
	cf := get(caseFoldingURL)
	defer cf.Close()

	verify(t, parseUCD(t, ud, scf, cf))

	dcp := get(derivedCorePropsURL)
	verifyProperty(t, "IsCased", parseProperty(t, dcp, "Cased"), IsCased)
	dcp.Close()
	dcp2 := get(derivedCorePropsURL)
	verifyProperty(t, "IsCaseIgnorable", parseProperty(t, dcp2, "Case_Ignorable"), IsCaseIgnorable)
	dcp2.Close()

	t.Logf("PASSED: default case mappings and properties reproduce official Unicode %s data", UnicodeVersion)
}

func TestSpotChecks(t *testing.T) {
	t.Run("ToUpper", func(t *testing.T) {
		cases := map[string]string{
			"straße": "STRASSE", // ß -> SS
			"ﬀ":      "FF",      // U+FB00 ligature ff
			"ﬁ":      "FI",      // U+FB01 ligature fi
			"İ":      "İ",       // U+0130 stays (no uppercase change)
			"hello":  "HELLO",
			"ΟΔΟΣ":   "ΟΔΟΣ",
			"weiß":   "WEISS",
		}
		for in, want := range cases {
			if got := ToUpper(in); got != want {
				t.Errorf("ToUpper(%q) = %q, want %q", in, got, want)
			}
		}
	})

	t.Run("ToLower", func(t *testing.T) {
		cases := map[string]string{
			"HELLO":   "hello",
			"I":       "i",       // default I -> i
			"İ":       "i\u0307", // U+0130 -> i + combining dot above (default)
			"ΟΔΟΣ":    "οδος",    // medial+final sigma: final sigma at word end
			"ΣΊΣΥΦΟΣ": "σίσυφος", // leading & internal sigma -> σ, final -> ς
		}
		for in, want := range cases {
			if got := ToLower(in); got != want {
				t.Errorf("ToLower(%q) = %q, want %q", in, got, want)
			}
		}
	})

	t.Run("FinalSigma", func(t *testing.T) {
		// A lone capital sigma (no preceding cased char) is not final.
		if got := ToLower("Σ"); got != "σ" {
			t.Errorf("ToLower(%q) = %q, want %q", "Σ", got, "σ")
		}
		// Sigma followed by a cased letter is not final.
		if got := ToLower("ΣΑ"); got != "σα" {
			t.Errorf("ToLower(%q) = %q, want %q", "ΣΑ", got, "σα")
		}
	})

	t.Run("ToFold", func(t *testing.T) {
		cases := map[string]string{
			"ﬀ":        "ff",
			" STRASSE": " strasse",
			"ß":        "ss",
			"Σ":        "σ",
			"ς":        "σ", // final sigma folds to plain sigma
			"İ":        "i\u0307",
		}
		for in, want := range cases {
			if got := ToFold(in); got != want {
				t.Errorf("ToFold(%q) = %q, want %q", in, got, want)
			}
		}
	})

	t.Run("ToTitle", func(t *testing.T) {
		cases := map[string]string{
			"hello world": "Hello World",
			"HELLO WORLD": "Hello World",
			"don't stop":  "Don't Stop",
			"ﬂood":        "Flood", // first-letter ligature titlecases to "Fl"
		}
		for in, want := range cases {
			if got := ToTitle(in); got != want {
				t.Errorf("ToTitle(%q) = %q, want %q", in, got, want)
			}
		}
	})
}

func TestCaselessMatch(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"hello", "HELLO", true},
		{"straße", "STRASSE", true},
		{"ﬀ", "ff", true},
		{"Σίσυφος", "ΣΊΣΥΦΟΣ", true}, // same word, differs only by case
		{"café", "cafe", false},      // differ by an accent, not case
		{"abc", "abd", false},
		{"", "", true},
	}
	for _, c := range cases {
		if got := CaselessMatch(c.a, c.b); got != c.want {
			t.Errorf("CaselessMatch(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestProperties(t *testing.T) {
	cased := []rune{'A', 'z', 0x00DF /* ß */, 0x01C5 /* Lt Dž */, 0x03A3 /* Σ */}
	for _, r := range cased {
		if !IsCased(r) {
			t.Errorf("IsCased(U+%04X) = false, want true", r)
		}
	}
	notCased := []rune{'0', ' ', '!', 0x0307 /* combining dot above */}
	for _, r := range notCased {
		if IsCased(r) {
			t.Errorf("IsCased(U+%04X) = true, want false", r)
		}
	}
	ignorable := []rune{0x0027 /* ' */, 0x002E /* . */, 0x0307 /* combining dot above */}
	for _, r := range ignorable {
		if !IsCaseIgnorable(r) {
			t.Errorf("IsCaseIgnorable(U+%04X) = false, want true", r)
		}
	}
	if IsCaseIgnorable('A') {
		t.Error("IsCaseIgnorable('A') = true, want false")
	}
}

func TestIdempotentFoldAndIdentity(t *testing.T) {
	// Folding twice equals folding once for a range of inputs.
	inputs := []string{"Hello", "ΟΔΟΣ", "straße", "ﬀ", "İ", "ASCII 123!"}
	for _, in := range inputs {
		once := ToFold(in)
		if twice := ToFold(once); twice != once {
			t.Errorf("ToFold not idempotent for %q: %q != %q", in, twice, once)
		}
	}
	// ASCII digits and punctuation are unchanged by every operation.
	const stable = "0123456789 !?."
	if ToLower(stable) != stable || ToUpper(stable) != stable || ToFold(stable) != stable {
		t.Errorf("stable string changed by a case operation")
	}
}
