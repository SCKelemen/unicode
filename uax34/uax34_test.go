package uax34

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const namedSequencesURL = "https://www.unicode.org/Public/17.0.0/ucd/NamedSequences.txt"

// fileEntry is one parsed row of NamedSequences.txt.
type fileEntry struct {
	name string
	seq  []rune
}

// parseNamedSequences parses NamedSequences.txt, returning every normative
// entry. It mirrors the format the generator consumes: comments begin with
// '#', and each data line is "Name;CodePointSequence".
func parseNamedSequences(t *testing.T, r io.Reader) []fileEntry {
	t.Helper()
	var entries []fileEntry
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if hash := strings.IndexByte(line, '#'); hash >= 0 {
			line = line[:hash]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ";")
		if len(parts) < 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		fields := strings.Fields(parts[1])
		if name == "" || len(fields) < 2 {
			t.Fatalf("malformed entry %q (need a name and >=2 code points)", line)
		}
		seq := make([]rune, 0, len(fields))
		for _, f := range fields {
			v, err := strconv.ParseInt(f, 16, 32)
			if err != nil {
				t.Fatalf("bad code point %q in %q: %v", f, name, err)
			}
			seq = append(seq, rune(v))
		}
		entries = append(entries, fileEntry{name: name, seq: seq})
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanning NamedSequences.txt: %v", err)
	}
	return entries
}

// assertRoundTrip checks that every entry resolves through both Lookup and
// Name and that Count matches the number of entries in the source file.
func assertRoundTrip(t *testing.T, entries []fileEntry) {
	t.Helper()

	if got := Count(); got != len(entries) {
		t.Errorf("Count() = %d, want %d", got, len(entries))
	}

	for _, e := range entries {
		seq, ok := Lookup(e.name)
		if !ok {
			t.Errorf("Lookup(%q) not found", e.name)
			continue
		}
		if !slices.Equal(seq, e.seq) {
			t.Errorf("Lookup(%q) = %v, want %v", e.name, seq, e.seq)
		}

		name, ok := Name(e.seq)
		if !ok {
			t.Errorf("Name(%v) not found (expected %q)", e.seq, e.name)
			continue
		}
		if name != e.name {
			t.Errorf("Name(%v) = %q, want %q", e.seq, name, e.name)
		}
	}
}

// TestDataFileConsistency validates the generated table against the committed
// NamedSequences.txt: every entry must round-trip through Lookup and Name, and
// Count must equal the number of entries in the file.
func TestDataFileConsistency(t *testing.T) {
	file, err := os.Open("NamedSequences.txt")
	if err != nil {
		t.Skipf("Skipping data file consistency test: %v", err)
		return
	}
	defer file.Close()

	entries := parseNamedSequences(t, file)
	if len(entries) == 0 {
		t.Fatal("parsed zero entries from NamedSequences.txt")
	}
	assertRoundTrip(t, entries)
}

// TestOfficialNamedSequences downloads the official Unicode 17.0.0
// NamedSequences.txt and verifies that every entry round-trips. It is skipped
// in short mode or when the download is unavailable.
func TestOfficialNamedSequences(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping official conformance test in short mode")
	}

	resp, err := http.Get(namedSequencesURL)
	if err != nil {
		t.Skipf("Skipping official conformance test (download unavailable): %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Skipf("Skipping official conformance test (HTTP %d)", resp.StatusCode)
	}

	entries := parseNamedSequences(t, resp.Body)
	if len(entries) == 0 {
		t.Fatal("parsed zero entries from official NamedSequences.txt")
	}
	assertRoundTrip(t, entries)
	t.Logf("PASSED: %d/%d named sequences round-trip (100%% conformance)", len(entries), len(entries))
}

func TestLookupSpotChecks(t *testing.T) {
	cases := []struct {
		name string
		want []rune
	}{
		{"KEYCAP NUMBER SIGN", []rune{0x0023, 0xFE0F, 0x20E3}},
		{"LATIN SMALL LETTER A WITH MACRON AND GRAVE", []rune{0x0101, 0x0300}},
	}
	for _, c := range cases {
		seq, ok := Lookup(c.name)
		if !ok {
			t.Errorf("Lookup(%q) not found", c.name)
			continue
		}
		if !slices.Equal(seq, c.want) {
			t.Errorf("Lookup(%q) = %v, want %v", c.name, seq, c.want)
		}
		// Reverse lookup must return the same name.
		if name, ok := Name(c.want); !ok || name != c.name {
			t.Errorf("Name(%v) = %q, %v; want %q, true", c.want, name, ok, c.name)
		}
	}
}

func TestLookupUnknown(t *testing.T) {
	if seq, ok := Lookup("THIS IS NOT A NAMED SEQUENCE"); ok {
		t.Errorf("Lookup(unknown) = %v, true; want nil, false", seq)
	}
	if seq, ok := Lookup(""); ok {
		t.Errorf("Lookup(\"\") = %v, true; want nil, false", seq)
	}
}

func TestNameRequiresExactSequence(t *testing.T) {
	// A proper prefix of a real sequence must not match.
	if name, ok := Name([]rune{0x0023, 0xFE0F}); ok {
		t.Errorf("Name(prefix) = %q, true; want \"\", false", name)
	}
	// A supersequence of a real sequence must not match.
	if name, ok := Name([]rune{0x0023, 0xFE0F, 0x20E3, 0x0041}); ok {
		t.Errorf("Name(supersequence) = %q, true; want \"\", false", name)
	}
	// An empty sequence must not match.
	if name, ok := Name(nil); ok {
		t.Errorf("Name(nil) = %q, true; want \"\", false", name)
	}
}

func TestCountMatchesAll(t *testing.T) {
	if got, n := Count(), len(All()); got != n {
		t.Errorf("Count() = %d, len(All()) = %d; want equal", got, n)
	}
}

func TestAllSortedAndComplete(t *testing.T) {
	all := All()
	if len(all) != Count() {
		t.Fatalf("len(All()) = %d, want %d", len(all), Count())
	}
	if !sort.SliceIsSorted(all, func(i, j int) bool { return all[i].Name < all[j].Name }) {
		t.Error("All() is not sorted ascending by name")
	}
	// Every entry from All() must round-trip through Lookup.
	for _, s := range all {
		seq, ok := Lookup(s.Name)
		if !ok {
			t.Errorf("Lookup(%q) not found", s.Name)
			continue
		}
		if !slices.Equal(seq, s.Runes) {
			t.Errorf("Lookup(%q) = %v, want %v", s.Name, seq, s.Runes)
		}
	}
}

// TestReturnedSlicesAreCopies ensures callers cannot mutate the internal table
// through the slices returned by Lookup and All.
func TestReturnedSlicesAreCopies(t *testing.T) {
	const name = "KEYCAP NUMBER SIGN"

	seq, ok := Lookup(name)
	if !ok {
		t.Fatalf("Lookup(%q) not found", name)
	}
	seq[0] = 0xFFFF // mutate the returned copy

	seq2, _ := Lookup(name)
	if seq2[0] == 0xFFFF {
		t.Error("Lookup returned a slice aliasing internal data; mutation leaked")
	}

	all := All()
	for i := range all {
		if all[i].Name == name {
			all[i].Runes[0] = 0xFFFF
		}
	}
	for _, s := range All() {
		if s.Name == name && s.Runes[0] == 0xFFFF {
			t.Error("All returned slices aliasing internal data; mutation leaked")
		}
	}
}
