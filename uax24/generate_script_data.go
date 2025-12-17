// +build ignore

// This program generates script_data.go from Unicode data files
// Run with: go run generate_script_data.go

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

const scriptsURL = "https://www.unicode.org/Public/17.0.0/ucd/Scripts.txt"

type scriptRange struct {
	start  rune
	end    rune
	script string
}

// scriptToConst maps Unicode script names to Go constant names
var scriptToConst = map[string]string{
	"Common":                  "ScriptCommon",
	"Inherited":               "ScriptInherited",
	"Adlam":                   "ScriptAdlam",
	"Ahom":                    "ScriptAhom",
	"Anatolian_Hieroglyphs":   "ScriptAnatolianHieroglyphs",
	"Arabic":                  "ScriptArabic",
	"Armenian":                "ScriptArmenian",
	"Avestan":                 "ScriptAvestan",
	"Balinese":                "ScriptBalinese",
	"Bamum":                   "ScriptBamum",
	"Bassa_Vah":               "ScriptBassaVah",
	"Batak":                   "ScriptBatak",
	"Bengali":                 "ScriptBengali",
	"Bhaiksuki":               "ScriptBhaiksuki",
	"Bopomofo":                "ScriptBopomofo",
	"Brahmi":                  "ScriptBrahmi",
	"Braille":                 "ScriptBraille",
	"Buginese":                "ScriptBuginese",
	"Buhid":                   "ScriptBuhid",
	"Canadian_Aboriginal":     "ScriptCanadianAboriginal",
	"Carian":                  "ScriptCarian",
	"Caucasian_Albanian":      "ScriptCaucasianAlbanian",
	"Chakma":                  "ScriptChakma",
	"Cham":                    "ScriptCham",
	"Cherokee":                "ScriptCherokee",
	"Chorasmian":              "ScriptChorasmian",
	"Coptic":                  "ScriptCoptic",
	"Cuneiform":               "ScriptCuneiform",
	"Cypriot":                 "ScriptCypriot",
	"Cyrillic":                "ScriptCyrillic",
	"Deseret":                 "ScriptDeseret",
	"Devanagari":              "ScriptDevanagari",
	"Dives_Akuru":             "ScriptDivesAkuru",
	"Dogra":                   "ScriptDogra",
	"Duployan":                "ScriptDuployan",
	"Egyptian_Hieroglyphs":    "ScriptEgyptianHieroglyphs",
	"Elbasan":                 "ScriptElbasan",
	"Elymaic":                 "ScriptElymaic",
	"Ethiopic":                "ScriptEthiopic",
	"Georgian":                "ScriptGeorgian",
	"Glagolitic":              "ScriptGlagolitic",
	"Gothic":                  "ScriptGothic",
	"Grantha":                 "ScriptGrantha",
	"Greek":                   "ScriptGreek",
	"Gujarati":                "ScriptGujarati",
	"Gunjala_Gondi":           "ScriptGunjalaGondi",
	"Gurmukhi":                "ScriptGurmukhi",
	"Han":                     "ScriptHan",
	"Hangul":                  "ScriptHangul",
	"Hanifi_Rohingya":         "ScriptHanifiRohingya",
	"Hanunoo":                 "ScriptHanunoo",
	"Hatran":                  "ScriptHatran",
	"Hebrew":                  "ScriptHebrew",
	"Hiragana":                "ScriptHiragana",
	"Imperial_Aramaic":        "ScriptImperialAramaic",
	"Inscriptional_Pahlavi":   "ScriptInscriptionalPahlavi",
	"Inscriptional_Parthian":  "ScriptInscriptionalParthian",
	"Javanese":                "ScriptJavanese",
	"Kaithi":                  "ScriptKaithi",
	"Kannada":                 "ScriptKannada",
	"Katakana":                "ScriptKatakana",
	"Kayah_Li":                "ScriptKayahLi",
	"Kharoshthi":              "ScriptKharoshthi",
	"Khmer":                   "ScriptKhmer",
	"Khojki":                  "ScriptKhojki",
	"Khudawadi":               "ScriptKhudawadi",
	"Lao":                     "ScriptLao",
	"Latin":                   "ScriptLatin",
	"Lepcha":                  "ScriptLepcha",
	"Limbu":                   "ScriptLimbu",
	"Linear_A":                "ScriptLinearA",
	"Linear_B":                "ScriptLinearB",
	"Lisu":                    "ScriptLisu",
	"Lycian":                  "ScriptLycian",
	"Lydian":                  "ScriptLydian",
	"Mahajani":                "ScriptMahajani",
	"Makasar":                 "ScriptMakasar",
	"Malayalam":               "ScriptMalayalam",
	"Mandaic":                 "ScriptMandaic",
	"Manichaean":              "ScriptManichaean",
	"Marchen":                 "ScriptMarchen",
	"Masaram_Gondi":           "ScriptMasaramGondi",
	"Medefaidrin":             "ScriptMedefaidrin",
	"Meetei_Mayek":            "ScriptMeeteiMayek",
	"Mende_Kikakui":           "ScriptMendeKikakui",
	"Meroitic_Cursive":        "ScriptMeroiticCursive",
	"Meroitic_Hieroglyphs":    "ScriptMeroiticHieroglyphs",
	"Miao":                    "ScriptMiao",
	"Modi":                    "ScriptModi",
	"Mongolian":               "ScriptMongolian",
	"Mro":                     "ScriptMro",
	"Multani":                 "ScriptMultani",
	"Myanmar":                 "ScriptMyanmar",
	"Nabataean":               "ScriptNabataean",
	"Nandinagari":             "ScriptNandinagari",
	"New_Tai_Lue":             "ScriptNewTaiLue",
	"Newa":                    "ScriptNewa",
	"Nko":                     "ScriptNko",
	"Nushu":                   "ScriptNushu",
	"Nyiakeng_Puachue_Hmong":  "ScriptNyiakengPuachueHmong",
	"Ogham":                   "ScriptOgham",
	"Ol_Chiki":                "ScriptOlChiki",
	"Old_Hungarian":           "ScriptOldHungarian",
	"Old_Italic":              "ScriptOldItalic",
	"Old_North_Arabian":       "ScriptOldNorthArabian",
	"Old_Permic":              "ScriptOldPermic",
	"Old_Persian":             "ScriptOldPersian",
	"Old_Sogdian":             "ScriptOldSogdian",
	"Old_South_Arabian":       "ScriptOldSouthArabian",
	"Old_Turkic":              "ScriptOldTurkic",
	"Oriya":                   "ScriptOriya",
	"Osage":                   "ScriptOsage",
	"Osmanya":                 "ScriptOsmanya",
	"Pahawh_Hmong":            "ScriptPahawhHmong",
	"Palmyrene":               "ScriptPalmyrene",
	"Pau_Cin_Hau":             "ScriptPauCinHau",
	"Phags_Pa":                "ScriptPhagsPa",
	"Phoenician":              "ScriptPhoenician",
	"Psalter_Pahlavi":         "ScriptPsalterPahlavi",
	"Rejang":                  "ScriptRejang",
	"Runic":                   "ScriptRunic",
	"Samaritan":               "ScriptSamaritan",
	"Saurashtra":              "ScriptSaurashtra",
	"Sharada":                 "ScriptSharada",
	"Shavian":                 "ScriptShavian",
	"Siddham":                 "ScriptSiddham",
	"SignWriting":             "ScriptSignWriting",
	"Sinhala":                 "ScriptSinhala",
	"Sogdian":                 "ScriptSogdian",
	"Sora_Sompeng":            "ScriptSoraSompeng",
	"Soyombo":                 "ScriptSoyombo",
	"Sundanese":               "ScriptSundanese",
	"Syloti_Nagri":            "ScriptSylotiNagri",
	"Syriac":                  "ScriptSyriac",
	"Tagalog":                 "ScriptTagalog",
	"Tagbanwa":                "ScriptTagbanwa",
	"Tai_Le":                  "ScriptTaiLe",
	"Tai_Tham":                "ScriptTaiTham",
	"Tai_Viet":                "ScriptTaiViet",
	"Takri":                   "ScriptTakri",
	"Tamil":                   "ScriptTamil",
	"Tangut":                  "ScriptTangut",
	"Telugu":                  "ScriptTelugu",
	"Thaana":                  "ScriptThaana",
	"Thai":                    "ScriptThai",
	"Tibetan":                 "ScriptTibetan",
	"Tifinagh":                "ScriptTifinagh",
	"Tirhuta":                 "ScriptTirhuta",
	"Ugaritic":                "ScriptUgaritic",
	"Vai":                     "ScriptVai",
	"Wancho":                  "ScriptWancho",
	"Warang_Citi":             "ScriptWarangCiti",
	"Yezidi":                  "ScriptYezidi",
	"Yi":                      "ScriptYi",
	"Zanabazar_Square":        "ScriptZanabazarSquare",
	"Garay":                   "ScriptGaray",
	"Gurung_Khema":            "ScriptGurungKhema",
	"Kirat_Rai":               "ScriptKiratRai",
	"Ol_Onal":                 "ScriptOlOnal",
	"Sunuwar":                 "ScriptSunuwar",
	"Todhri":                  "ScriptTodhri",
	"Tulu_Tigalari":           "ScriptTuluTigalari",
	"Vithkuqi":                "ScriptVithkuqi",
	"Sidetic":                 "ScriptSidetic",
	"Old_Uyghur":              "ScriptOldUyghur",
	"Tolong_Siki":             "ScriptTolongSiki",
	"Kawi":                    "ScriptKawi",
	"Cypro_Minoan":            "ScriptCyproMinoan",
	"Tangsa":                  "ScriptTangsa",
	"Beria_Erfe":              "ScriptBeriaErfe",
	"Khitan_Small_Script":     "ScriptKhitanSmallScript",
	"Toto":                    "ScriptToto",
	"Nag_Mundari":             "ScriptNagMundari",
	"Tai_Yo":                  "ScriptTaiYo",
}

func main() {
	// Download Scripts.txt
	fmt.Println("Downloading Scripts.txt...")
	resp, err := http.Get(scriptsURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading Scripts.txt: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error: HTTP %d\n", resp.StatusCode)
		os.Exit(1)
	}

	// Parse the file
	ranges := parseScripts(resp.Body)

	// Sort ranges by start position
	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].start < ranges[j].start
	})

	// Generate the output file
	out, err := os.Create("script_data.go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	fmt.Fprintf(out, "// Code generated by generate_script_data.go DO NOT EDIT.\n")
	fmt.Fprintf(out, "// Source: Unicode 17.0.0 Scripts.txt\n")
	fmt.Fprintf(out, "\npackage uax24\n\n")
	fmt.Fprintf(out, "type scriptEntry struct {\n")
	fmt.Fprintf(out, "\tstart  rune\n")
	fmt.Fprintf(out, "\tend    rune\n")
	fmt.Fprintf(out, "\tscript Script\n")
	fmt.Fprintf(out, "}\n\n")

	fmt.Fprintf(out, "// scriptData contains all Unicode Script property assignments\n")
	fmt.Fprintf(out, "// Total ranges: %d\n", len(ranges))
	fmt.Fprintf(out, "var scriptData = []scriptEntry{\n")

	for _, r := range ranges {
		constName := scriptToConst[r.script]
		if constName == "" {
			constName = "ScriptUnknown"
			fmt.Fprintf(os.Stderr, "Warning: Unknown script %s\n", r.script)
		}

		if r.start == r.end {
			fmt.Fprintf(out, "\t{0x%04X, 0x%04X, %s}, // U+%04X %s\n",
				r.start, r.end, constName, r.start, r.script)
		} else {
			fmt.Fprintf(out, "\t{0x%04X, 0x%04X, %s}, // U+%04X..U+%04X %s\n",
				r.start, r.end, constName, r.start, r.end, r.script)
		}
	}

	fmt.Fprintf(out, "}\n")

	fmt.Printf("Generated script_data.go successfully\n")
	fmt.Printf("  Total ranges: %d\n", len(ranges))

	// Count script occurrences
	scriptCounts := make(map[string]int)
	for _, r := range ranges {
		scriptCounts[r.script]++
	}
	fmt.Printf("  Unique scripts: %d\n", len(scriptCounts))
}

func parseScripts(r io.Reader) []scriptRange {
	var ranges []scriptRange
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on semicolon
		parts := strings.Split(line, ";")
		if len(parts) < 2 {
			continue
		}

		codePoints := strings.TrimSpace(parts[0])
		scriptName := strings.TrimSpace(parts[1])

		// Extract just the script name (before any # comment)
		scriptName = strings.Fields(scriptName)[0]

		// Parse code point or range
		var start, end rune
		if strings.Contains(codePoints, "..") {
			// Range
			rangeParts := strings.Split(codePoints, "..")
			startVal, _ := strconv.ParseInt(rangeParts[0], 16, 32)
			endVal, _ := strconv.ParseInt(rangeParts[1], 16, 32)
			start = rune(startVal)
			end = rune(endVal)
		} else {
			// Single code point
			val, _ := strconv.ParseInt(codePoints, 16, 32)
			start = rune(val)
			end = rune(val)
		}

		ranges = append(ranges, scriptRange{
			start:  start,
			end:    end,
			script: scriptName,
		})
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading Scripts.txt: %v\n", err)
		os.Exit(1)
	}

	return ranges
}
