package uax24

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

const scriptsTestURL = "https://www.unicode.org/Public/17.0.0/ucd/Scripts.txt"

func TestOfficialScriptProperty(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping official conformance tests in short mode")
	}

	// Download Scripts.txt
	resp, err := http.Get(scriptsTestURL)
	if err != nil {
		t.Fatalf("Failed to download Scripts.txt: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("HTTP error: %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	lineNum := 0
	testCount := 0
	failCount := 0
	var firstFailures []string

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse line: CODE_POINT(S) ; SCRIPT_NAME # COMMENT
		parts := strings.Split(line, ";")
		if len(parts) < 2 {
			continue
		}

		codePointPart := strings.TrimSpace(parts[0])
		scriptName := strings.TrimSpace(strings.Split(parts[1], "#")[0])

		// Parse code point or range
		var start, end rune
		if strings.Contains(codePointPart, "..") {
			rangeParts := strings.Split(codePointPart, "..")
			startVal, _ := strconv.ParseInt(strings.TrimSpace(rangeParts[0]), 16, 32)
			endVal, _ := strconv.ParseInt(strings.TrimSpace(rangeParts[1]), 16, 32)
			start = rune(startVal)
			end = rune(endVal)
		} else {
			val, _ := strconv.ParseInt(codePointPart, 16, 32)
			start = rune(val)
			end = rune(val)
		}

		// Get expected script constant
		expectedScript := getScriptConstant(scriptName)

		// Test each code point in range
		for r := start; r <= end; r++ {
			testCount++
			actualScript := LookupScript(r)

			if actualScript != expectedScript {
				failCount++
				if len(firstFailures) < 100 {
					firstFailures = append(firstFailures,
						formatFailure(r, scriptName, expectedScript, actualScript, lineNum))
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading Scripts.txt: %v", err)
	}

	// Report results
	if failCount > 0 {
		t.Errorf("FAILED: %d/%d tests failed", failCount, testCount)
		t.Logf("First failures:")
		for i, fail := range firstFailures {
			if i >= 10 {
				t.Logf("  ... and %d more failures", len(firstFailures)-10)
				break
			}
			t.Logf("  %s", fail)
		}
	} else {
		t.Logf("PASSED: %d/%d tests (100%% conformance)", testCount, testCount)
	}
}

func formatFailure(r rune, scriptName string, expected, actual Script, lineNum int) string {
	return fmt.Sprintf("Line %d: U+%04X expected script %s (%s) but got %s",
		lineNum, r, scriptName, expected, actual)
}

func getScriptConstant(scriptName string) Script {
	scriptMap := map[string]Script{
		"Common":                  ScriptCommon,
		"Inherited":               ScriptInherited,
		"Adlam":                   ScriptAdlam,
		"Ahom":                    ScriptAhom,
		"Anatolian_Hieroglyphs":   ScriptAnatolianHieroglyphs,
		"Arabic":                  ScriptArabic,
		"Armenian":                ScriptArmenian,
		"Avestan":                 ScriptAvestan,
		"Balinese":                ScriptBalinese,
		"Bamum":                   ScriptBamum,
		"Bassa_Vah":               ScriptBassaVah,
		"Batak":                   ScriptBatak,
		"Bengali":                 ScriptBengali,
		"Bhaiksuki":               ScriptBhaiksuki,
		"Bopomofo":                ScriptBopomofo,
		"Brahmi":                  ScriptBrahmi,
		"Braille":                 ScriptBraille,
		"Buginese":                ScriptBuginese,
		"Buhid":                   ScriptBuhid,
		"Canadian_Aboriginal":     ScriptCanadianAboriginal,
		"Carian":                  ScriptCarian,
		"Caucasian_Albanian":      ScriptCaucasianAlbanian,
		"Chakma":                  ScriptChakma,
		"Cham":                    ScriptCham,
		"Cherokee":                ScriptCherokee,
		"Chorasmian":              ScriptChorasmian,
		"Coptic":                  ScriptCoptic,
		"Cuneiform":               ScriptCuneiform,
		"Cypriot":                 ScriptCypriot,
		"Cypro_Minoan":            ScriptCyproMinoan,
		"Cyrillic":                ScriptCyrillic,
		"Deseret":                 ScriptDeseret,
		"Devanagari":              ScriptDevanagari,
		"Dives_Akuru":             ScriptDivesAkuru,
		"Dogra":                   ScriptDogra,
		"Duployan":                ScriptDuployan,
		"Egyptian_Hieroglyphs":    ScriptEgyptianHieroglyphs,
		"Elbasan":                 ScriptElbasan,
		"Elymaic":                 ScriptElymaic,
		"Ethiopic":                ScriptEthiopic,
		"Georgian":                ScriptGeorgian,
		"Glagolitic":              ScriptGlagolitic,
		"Gothic":                  ScriptGothic,
		"Grantha":                 ScriptGrantha,
		"Greek":                   ScriptGreek,
		"Gujarati":                ScriptGujarati,
		"Gunjala_Gondi":           ScriptGunjalaGondi,
		"Gurmukhi":                ScriptGurmukhi,
		"Han":                     ScriptHan,
		"Hangul":                  ScriptHangul,
		"Hanifi_Rohingya":         ScriptHanifiRohingya,
		"Hanunoo":                 ScriptHanunoo,
		"Hatran":                  ScriptHatran,
		"Hebrew":                  ScriptHebrew,
		"Hiragana":                ScriptHiragana,
		"Imperial_Aramaic":        ScriptImperialAramaic,
		"Inscriptional_Pahlavi":   ScriptInscriptionalPahlavi,
		"Inscriptional_Parthian":  ScriptInscriptionalParthian,
		"Javanese":                ScriptJavanese,
		"Kaithi":                  ScriptKaithi,
		"Kannada":                 ScriptKannada,
		"Katakana":                ScriptKatakana,
		"Kawi":                    ScriptKawi,
		"Kayah_Li":                ScriptKayahLi,
		"Kharoshthi":              ScriptKharoshthi,
		"Khitan_Small_Script":     ScriptKhitanSmallScript,
		"Khmer":                   ScriptKhmer,
		"Khojki":                  ScriptKhojki,
		"Khudawadi":               ScriptKhudawadi,
		"Lao":                     ScriptLao,
		"Latin":                   ScriptLatin,
		"Lepcha":                  ScriptLepcha,
		"Limbu":                   ScriptLimbu,
		"Linear_A":                ScriptLinearA,
		"Linear_B":                ScriptLinearB,
		"Lisu":                    ScriptLisu,
		"Lycian":                  ScriptLycian,
		"Lydian":                  ScriptLydian,
		"Mahajani":                ScriptMahajani,
		"Makasar":                 ScriptMakasar,
		"Malayalam":               ScriptMalayalam,
		"Mandaic":                 ScriptMandaic,
		"Manichaean":              ScriptManichaean,
		"Marchen":                 ScriptMarchen,
		"Masaram_Gondi":           ScriptMasaramGondi,
		"Medefaidrin":             ScriptMedefaidrin,
		"Meetei_Mayek":            ScriptMeeteiMayek,
		"Mende_Kikakui":           ScriptMendeKikakui,
		"Meroitic_Cursive":        ScriptMeroiticCursive,
		"Meroitic_Hieroglyphs":    ScriptMeroiticHieroglyphs,
		"Miao":                    ScriptMiao,
		"Modi":                    ScriptModi,
		"Mongolian":               ScriptMongolian,
		"Mro":                     ScriptMro,
		"Multani":                 ScriptMultani,
		"Myanmar":                 ScriptMyanmar,
		"Nabataean":               ScriptNabataean,
		"Nag_Mundari":             ScriptNagMundari,
		"Nandinagari":             ScriptNandinagari,
		"New_Tai_Lue":             ScriptNewTaiLue,
		"Newa":                    ScriptNewa,
		"Nko":                     ScriptNko,
		"Nushu":                   ScriptNushu,
		"Nyiakeng_Puachue_Hmong":  ScriptNyiakengPuachueHmong,
		"Ogham":                   ScriptOgham,
		"Ol_Chiki":                ScriptOlChiki,
		"Old_Hungarian":           ScriptOldHungarian,
		"Old_Italic":              ScriptOldItalic,
		"Old_North_Arabian":       ScriptOldNorthArabian,
		"Old_Permic":              ScriptOldPermic,
		"Old_Persian":             ScriptOldPersian,
		"Old_Sogdian":             ScriptOldSogdian,
		"Old_South_Arabian":       ScriptOldSouthArabian,
		"Old_Turkic":              ScriptOldTurkic,
		"Old_Uyghur":              ScriptOldUyghur,
		"Oriya":                   ScriptOriya,
		"Osage":                   ScriptOsage,
		"Osmanya":                 ScriptOsmanya,
		"Pahawh_Hmong":            ScriptPahawhHmong,
		"Palmyrene":               ScriptPalmyrene,
		"Pau_Cin_Hau":             ScriptPauCinHau,
		"Phags_Pa":                ScriptPhagsPa,
		"Phoenician":              ScriptPhoenician,
		"Psalter_Pahlavi":         ScriptPsalterPahlavi,
		"Rejang":                  ScriptRejang,
		"Runic":                   ScriptRunic,
		"Samaritan":               ScriptSamaritan,
		"Saurashtra":              ScriptSaurashtra,
		"Sharada":                 ScriptSharada,
		"Shavian":                 ScriptShavian,
		"Siddham":                 ScriptSiddham,
		"SignWriting":             ScriptSignWriting,
		"Sinhala":                 ScriptSinhala,
		"Sogdian":                 ScriptSogdian,
		"Sora_Sompeng":            ScriptSoraSompeng,
		"Soyombo":                 ScriptSoyombo,
		"Sundanese":               ScriptSundanese,
		"Syloti_Nagri":            ScriptSylotiNagri,
		"Syriac":                  ScriptSyriac,
		"Tagalog":                 ScriptTagalog,
		"Tagbanwa":                ScriptTagbanwa,
		"Tai_Le":                  ScriptTaiLe,
		"Tai_Tham":                ScriptTaiTham,
		"Tai_Viet":                ScriptTaiViet,
		"Takri":                   ScriptTakri,
		"Tamil":                   ScriptTamil,
		"Tangsa":                  ScriptTangsa,
		"Tangut":                  ScriptTangut,
		"Telugu":                  ScriptTelugu,
		"Thaana":                  ScriptThaana,
		"Thai":                    ScriptThai,
		"Tibetan":                 ScriptTibetan,
		"Tifinagh":                ScriptTifinagh,
		"Tirhuta":                 ScriptTirhuta,
		"Toto":                    ScriptToto,
		"Ugaritic":                ScriptUgaritic,
		"Vai":                     ScriptVai,
		"Vithkuqi":                ScriptVithkuqi,
		"Wancho":                  ScriptWancho,
		"Warang_Citi":             ScriptWarangCiti,
		"Yezidi":                  ScriptYezidi,
		"Yi":                      ScriptYi,
		"Zanabazar_Square":        ScriptZanabazarSquare,
		"Garay":                   ScriptGaray,
		"Gurung_Khema":            ScriptGurungKhema,
		"Kirat_Rai":               ScriptKiratRai,
		"Ol_Onal":                 ScriptOlOnal,
		"Sunuwar":                 ScriptSunuwar,
		"Todhri":                  ScriptTodhri,
		"Tulu_Tigalari":           ScriptTuluTigalari,
		"Sidetic":                 ScriptSidetic,
		"Tolong_Siki":             ScriptTolongSiki,
		"Beria_Erfe":              ScriptBeriaErfe,
		"Tai_Yo":                  ScriptTaiYo,
	}

	if script, ok := scriptMap[scriptName]; ok {
		return script
	}
	return ScriptUnknown
}
