// Package uax24 implements Unicode Script Property (UAX #24).
//
// This package provides the Unicode Script property which identifies
// the writing system (script) to which a character belongs. Scripts include
// Latin, Greek, Cyrillic, Han, Arabic, Devanagari, and many others.
//
// Based on: https://www.unicode.org/reports/tr24/
//
// # Script Property
//
// The Script property is a core Unicode property that assigns each character
// to exactly one script, with some characters having the special values:
//   - Common: Characters used across multiple scripts (digits, punctuation, etc.)
//   - Inherited: Combining marks and other characters that take the script of the base character
//   - Unknown: Unassigned code points or characters not yet assigned to a script
//
// # Use Cases
//
// The Script property is fundamental for:
//   - Mixed-script text detection and analysis
//   - Security identifier validation (detecting homograph attacks)
//   - Text rendering and font selection
//   - Input method editor (IME) selection
//   - Search and collation
//   - Language identification
//
// # Script Values
//
// Each character has exactly one script value. The property follows the
// ISO 15924 standard for script codes. Common scripts include:
//   - Latin (Latn): English, French, German, Spanish, etc.
//   - Greek (Grek): Greek
//   - Cyrillic (Cyrl): Russian, Ukrainian, Bulgarian, Serbian, etc.
//   - Han (Hani): Chinese, Japanese, Korean ideographs
//   - Hiragana (Hira): Japanese syllabary
//   - Katakana (Kana): Japanese syllabary
//   - Arabic (Arab): Arabic, Persian, Urdu, etc.
//   - Hebrew (Hebr): Hebrew
//   - Devanagari (Deva): Hindi, Sanskrit, Marathi, Nepali, etc.
//   - Bengali (Beng): Bengali, Assamese
//   - Thai (Thai): Thai
//   - Common (Zyyy): Digits, punctuation, symbols shared across scripts
//   - Inherited (Zinh): Combining marks, format characters
//
// See: https://www.unicode.org/iso15924/iso15924-codes.html
//
// # Conformance
//
// This implementation follows UAX #24 Script Property specification:
//   - https://www.unicode.org/reports/tr24/
//
// The implementation uses Script property assignments from Unicode 17.0.0
// data files (Scripts.txt).
//
// # Usage
//
//	import "github.com/SCKelemen/unicode/uax24"
//
//	// Get the script of a character
//	script := uax24.LookupScript('A')      // Returns ScriptLatin
//	script = uax24.LookupScript('中')      // Returns ScriptHan
//	script = uax24.LookupScript('5')       // Returns ScriptCommon
//
//	// Check if character belongs to a specific script
//	if uax24.IsLatin('A') {
//	    // Character is Latin
//	}
//
//	// Analyze a string for script composition
//	info := uax24.AnalyzeScripts("Hello мир")
//	fmt.Printf("Scripts: %v\n", info.Scripts) // [Latin, Cyrillic]
//
// # References
//
//   - UAX #24: https://www.unicode.org/reports/tr24/
//   - ISO 15924: https://www.unicode.org/iso15924/
//   - Scripts.txt: https://www.unicode.org/Public/17.0.0/ucd/Scripts.txt
package uax24

// Script represents a Unicode script value following ISO 15924.
// Each character in Unicode is assigned exactly one script.
type Script int

// Script constants following ISO 15924 codes.
// These are the most commonly used scripts in Unicode 17.0.0.
const (
	ScriptUnknown Script = iota // Unknown or unassigned
	ScriptCommon                 // Zyyy - Characters used across multiple scripts
	ScriptInherited              // Zinh - Combining marks that inherit script
	ScriptAdlam
	ScriptAhom
	ScriptAnatolianHieroglyphs
	ScriptArabic
	ScriptArmenian
	ScriptAvestan
	ScriptBalinese
	ScriptBamum
	ScriptBassaVah
	ScriptBatak
	ScriptBengali
	ScriptBhaiksuki
	ScriptBopomofo
	ScriptBrahmi
	ScriptBraille
	ScriptBuginese
	ScriptBuhid
	ScriptCanadianAboriginal
	ScriptCarian
	ScriptCaucasianAlbanian
	ScriptChakma
	ScriptCham
	ScriptCherokee
	ScriptChorasmian
	ScriptCoptic
	ScriptCuneiform
	ScriptCypriot
	ScriptCyrillic
	ScriptDeseret
	ScriptDevanagari
	ScriptDivesAkuru
	ScriptDogra
	ScriptDuployan
	ScriptEgyptianHieroglyphs
	ScriptElbasan
	ScriptElymaic
	ScriptEthiopic
	ScriptGeorgian
	ScriptGlagolitic
	ScriptGothic
	ScriptGrantha
	ScriptGreek
	ScriptGujarati
	ScriptGunjalaGondi
	ScriptGurmukhi
	ScriptHan
	ScriptHangul
	ScriptHanifiRohingya
	ScriptHanunoo
	ScriptHatran
	ScriptHebrew
	ScriptHiragana
	ScriptImperialAramaic
	ScriptInscriptionalPahlavi
	ScriptInscriptionalParthian
	ScriptJavanese
	ScriptKaithi
	ScriptKannada
	ScriptKatakana
	ScriptKayahLi
	ScriptKharoshthi
	ScriptKhmer
	ScriptKhojki
	ScriptKhudawadi
	ScriptLao
	ScriptLatin
	ScriptLepcha
	ScriptLimbu
	ScriptLinearA
	ScriptLinearB
	ScriptLisu
	ScriptLycian
	ScriptLydian
	ScriptMahajani
	ScriptMakasar
	ScriptMalayalam
	ScriptMandaic
	ScriptManichaean
	ScriptMarchen
	ScriptMasaramGondi
	ScriptMedefaidrin
	ScriptMeeteiMayek
	ScriptMendeKikakui
	ScriptMeroiticCursive
	ScriptMeroiticHieroglyphs
	ScriptMiao
	ScriptModi
	ScriptMongolian
	ScriptMro
	ScriptMultani
	ScriptMyanmar
	ScriptNabataean
	ScriptNandinagari
	ScriptNewTaiLue
	ScriptNewa
	ScriptNko
	ScriptNushu
	ScriptNyiakengPuachueHmong
	ScriptOgham
	ScriptOlChiki
	ScriptOldHungarian
	ScriptOldItalic
	ScriptOldNorthArabian
	ScriptOldPermic
	ScriptOldPersian
	ScriptOldSogdian
	ScriptOldSouthArabian
	ScriptOldTurkic
	ScriptOriya
	ScriptOsage
	ScriptOsmanya
	ScriptPahawhHmong
	ScriptPalmyrene
	ScriptPauCinHau
	ScriptPhagsPa
	ScriptPhoenician
	ScriptPsalterPahlavi
	ScriptRejang
	ScriptRunic
	ScriptSamaritan
	ScriptSaurashtra
	ScriptSharada
	ScriptShavian
	ScriptSiddham
	ScriptSignWriting
	ScriptSinhala
	ScriptSogdian
	ScriptSoraSompeng
	ScriptSoyombo
	ScriptSundanese
	ScriptSylotiNagri
	ScriptSyriac
	ScriptTagalog
	ScriptTagbanwa
	ScriptTaiLe
	ScriptTaiTham
	ScriptTaiViet
	ScriptTakri
	ScriptTamil
	ScriptTangut
	ScriptTelugu
	ScriptThaana
	ScriptThai
	ScriptTibetan
	ScriptTifinagh
	ScriptTirhuta
	ScriptUgaritic
	ScriptVai
	ScriptWancho
	ScriptWarangCiti
	ScriptYezidi
	ScriptYi
	ScriptZanabazarSquare

	// Additional Unicode 17.0.0 scripts
	ScriptGaray
	ScriptGurungKhema
	ScriptKiratRai
	ScriptOlOnal
	ScriptSunuwar
	ScriptTodhri
	ScriptTuluTigalari
	ScriptVithkuqi
	ScriptSidetic
	ScriptOldUyghur
	ScriptTolongSiki
	ScriptKawi
	ScriptCyproMinoan
	ScriptTangsa
	ScriptBeriaErfe
	ScriptKhitanSmallScript
	ScriptToto
	ScriptNagMundari
	ScriptTaiYo
)

// String returns the ISO 15924 4-letter code for the script.
func (s Script) String() string {
	if s < 0 || int(s) >= len(scriptNames) {
		return "Unknown"
	}
	return scriptNames[s]
}

var scriptNames = []string{
	"Unknown",
	"Common",
	"Inherited",
	"Adlam",
	"Ahom",
	"Anatolian_Hieroglyphs",
	"Arabic",
	"Armenian",
	"Avestan",
	"Balinese",
	"Bamum",
	"Bassa_Vah",
	"Batak",
	"Bengali",
	"Bhaiksuki",
	"Bopomofo",
	"Brahmi",
	"Braille",
	"Buginese",
	"Buhid",
	"Canadian_Aboriginal",
	"Carian",
	"Caucasian_Albanian",
	"Chakma",
	"Cham",
	"Cherokee",
	"Chorasmian",
	"Coptic",
	"Cuneiform",
	"Cypriot",
	"Cyrillic",
	"Deseret",
	"Devanagari",
	"Dives_Akuru",
	"Dogra",
	"Duployan",
	"Egyptian_Hieroglyphs",
	"Elbasan",
	"Elymaic",
	"Ethiopic",
	"Georgian",
	"Glagolitic",
	"Gothic",
	"Grantha",
	"Greek",
	"Gujarati",
	"Gunjala_Gondi",
	"Gurmukhi",
	"Han",
	"Hangul",
	"Hanifi_Rohingya",
	"Hanunoo",
	"Hatran",
	"Hebrew",
	"Hiragana",
	"Imperial_Aramaic",
	"Inscriptional_Pahlavi",
	"Inscriptional_Parthian",
	"Javanese",
	"Kaithi",
	"Kannada",
	"Katakana",
	"Kayah_Li",
	"Kharoshthi",
	"Khmer",
	"Khojki",
	"Khudawadi",
	"Lao",
	"Latin",
	"Lepcha",
	"Limbu",
	"Linear_A",
	"Linear_B",
	"Lisu",
	"Lycian",
	"Lydian",
	"Mahajani",
	"Makasar",
	"Malayalam",
	"Mandaic",
	"Manichaean",
	"Marchen",
	"Masaram_Gondi",
	"Medefaidrin",
	"Meetei_Mayek",
	"Mende_Kikakui",
	"Meroitic_Cursive",
	"Meroitic_Hieroglyphs",
	"Miao",
	"Modi",
	"Mongolian",
	"Mro",
	"Multani",
	"Myanmar",
	"Nabataean",
	"Nandinagari",
	"New_Tai_Lue",
	"Newa",
	"Nko",
	"Nushu",
	"Nyiakeng_Puachue_Hmong",
	"Ogham",
	"Ol_Chiki",
	"Old_Hungarian",
	"Old_Italic",
	"Old_North_Arabian",
	"Old_Permic",
	"Old_Persian",
	"Old_Sogdian",
	"Old_South_Arabian",
	"Old_Turkic",
	"Oriya",
	"Osage",
	"Osmanya",
	"Pahawh_Hmong",
	"Palmyrene",
	"Pau_Cin_Hau",
	"Phags_Pa",
	"Phoenician",
	"Psalter_Pahlavi",
	"Rejang",
	"Runic",
	"Samaritan",
	"Saurashtra",
	"Sharada",
	"Shavian",
	"Siddham",
	"SignWriting",
	"Sinhala",
	"Sogdian",
	"Sora_Sompeng",
	"Soyombo",
	"Sundanese",
	"Syloti_Nagri",
	"Syriac",
	"Tagalog",
	"Tagbanwa",
	"Tai_Le",
	"Tai_Tham",
	"Tai_Viet",
	"Takri",
	"Tamil",
	"Tangut",
	"Telugu",
	"Thaana",
	"Thai",
	"Tibetan",
	"Tifinagh",
	"Tirhuta",
	"Ugaritic",
	"Vai",
	"Wancho",
	"Warang_Citi",
	"Yezidi",
	"Yi",
	"Zanabazar_Square",
	"Garay",
	"Gurung_Khema",
	"Kirat_Rai",
	"Ol_Onal",
	"Sunuwar",
	"Todhri",
	"Tulu_Tigalari",
	"Vithkuqi",
	"Sidetic",
	"Old_Uyghur",
	"Tolong_Siki",
	"Kawi",
	"Cypro_Minoan",
	"Tangsa",
	"Beria_Erfe",
	"Khitan_Small_Script",
	"Toto",
	"Nag_Mundari",
	"Tai_Yo",
}

// LookupScript returns the Script property for the given rune.
// If the rune is unassigned or invalid, it returns ScriptUnknown.
func LookupScript(r rune) Script {
	// Binary search through the script ranges
	lo, hi := 0, len(scriptData)
	for lo < hi {
		mid := lo + (hi-lo)/2
		entry := scriptData[mid]
		if r < entry.start {
			hi = mid
		} else if r > entry.end {
			lo = mid + 1
		} else {
			return entry.script
		}
	}
	return ScriptUnknown
}

// IsCommon reports whether the rune has the Common script property.
// Common characters are used across multiple scripts (digits, punctuation, etc.).
func IsCommon(r rune) bool {
	return LookupScript(r) == ScriptCommon
}

// IsInherited reports whether the rune has the Inherited script property.
// Inherited characters (combining marks) take the script of their base character.
func IsInherited(r rune) bool {
	return LookupScript(r) == ScriptInherited
}

// IsLatin reports whether the rune belongs to the Latin script.
func IsLatin(r rune) bool {
	return LookupScript(r) == ScriptLatin
}

// IsGreek reports whether the rune belongs to the Greek script.
func IsGreek(r rune) bool {
	return LookupScript(r) == ScriptGreek
}

// IsCyrillic reports whether the rune belongs to the Cyrillic script.
func IsCyrillic(r rune) bool {
	return LookupScript(r) == ScriptCyrillic
}

// IsHan reports whether the rune belongs to the Han script (CJK ideographs).
func IsHan(r rune) bool {
	return LookupScript(r) == ScriptHan
}

// IsHiragana reports whether the rune belongs to the Hiragana script.
func IsHiragana(r rune) bool {
	return LookupScript(r) == ScriptHiragana
}

// IsKatakana reports whether the rune belongs to the Katakana script.
func IsKatakana(r rune) bool {
	return LookupScript(r) == ScriptKatakana
}

// IsArabic reports whether the rune belongs to the Arabic script.
func IsArabic(r rune) bool {
	return LookupScript(r) == ScriptArabic
}

// IsHebrew reports whether the rune belongs to the Hebrew script.
func IsHebrew(r rune) bool {
	return LookupScript(r) == ScriptHebrew
}

// IsDevanagari reports whether the rune belongs to the Devanagari script.
func IsDevanagari(r rune) bool {
	return LookupScript(r) == ScriptDevanagari
}

// IsBengali reports whether the rune belongs to the Bengali script.
func IsBengali(r rune) bool {
	return LookupScript(r) == ScriptBengali
}

// IsThai reports whether the rune belongs to the Thai script.
func IsThai(r rune) bool {
	return LookupScript(r) == ScriptThai
}

// ScriptInfo contains information about the scripts used in a string.
type ScriptInfo struct {
	// Scripts is the list of scripts found (excluding Common and Inherited)
	Scripts []Script
	// HasCommon indicates if Common script characters are present
	HasCommon bool
	// HasInherited indicates if Inherited script characters are present
	HasInherited bool
	// IsMixedScript indicates if multiple scripts (excluding Common/Inherited) are present
	IsMixedScript bool
}

// AnalyzeScripts analyzes a string and returns information about the scripts used.
// This is useful for:
//   - Detecting mixed-script identifiers (potential homograph attacks)
//   - Selecting appropriate fonts for rendering
//   - Language identification
//   - Input method selection
func AnalyzeScripts(s string) ScriptInfo {
	info := ScriptInfo{}
	seen := make(map[Script]bool)

	for _, r := range s {
		script := LookupScript(r)

		switch script {
		case ScriptCommon:
			info.HasCommon = true
		case ScriptInherited:
			info.HasInherited = true
		default:
			if !seen[script] {
				seen[script] = true
				info.Scripts = append(info.Scripts, script)
			}
		}
	}

	info.IsMixedScript = len(info.Scripts) > 1

	return info
}

// IsSingleScript reports whether the string contains characters from only
// a single script (excluding Common and Inherited).
// This is useful for security validation of identifiers.
func IsSingleScript(s string) bool {
	info := AnalyzeScripts(s)
	return len(info.Scripts) <= 1
}
