package uax24_test

import (
	"fmt"

	"github.com/SCKelemen/unicode/uax24"
)

func ExampleLookupScript() {
	// Lookup the script of various characters
	fmt.Println(uax24.LookupScript('A'))    // Latin
	fmt.Println(uax24.LookupScript('中'))   // Han (Chinese)
	fmt.Println(uax24.LookupScript('α'))    // Greek
	fmt.Println(uax24.LookupScript('Д'))    // Cyrillic
	fmt.Println(uax24.LookupScript('5'))    // Common (digits are shared)

	// Output:
	// Latin
	// Han
	// Greek
	// Cyrillic
	// Common
}

func ExampleIsLatin() {
	fmt.Println(uax24.IsLatin('A'))    // true
	fmt.Println(uax24.IsLatin('z'))    // true
	fmt.Println(uax24.IsLatin('中'))   // false - Han script
	fmt.Println(uax24.IsLatin('5'))    // false - Common script

	// Output:
	// true
	// true
	// false
	// false
}

func ExampleIsHan() {
	fmt.Println(uax24.IsHan('中'))   // true - Chinese
	fmt.Println(uax24.IsHan('日'))   // true - Japanese Kanji
	fmt.Println(uax24.IsHan('A'))    // false - Latin
	fmt.Println(uax24.IsHan('あ'))   // false - Hiragana

	// Output:
	// true
	// true
	// false
	// false
}

func ExampleIsCommon() {
	fmt.Println(uax24.IsCommon('5'))    // true - digits
	fmt.Println(uax24.IsCommon(' '))    // true - spaces
	fmt.Println(uax24.IsCommon(','))    // true - punctuation
	fmt.Println(uax24.IsCommon('A'))    // false - Latin

	// Output:
	// true
	// true
	// true
	// false
}

func ExampleAnalyzeScripts() {
	// Analyze a pure Latin string
	info := uax24.AnalyzeScripts("Hello")
	fmt.Printf("Scripts: %d, Mixed: %v\n", len(info.Scripts), info.IsMixedScript)

	// Analyze mixed script string
	info = uax24.AnalyzeScripts("Hello мир")
	fmt.Printf("Scripts: %d, Mixed: %v\n", len(info.Scripts), info.IsMixedScript)
	fmt.Printf("Script names: %v, %v\n", info.Scripts[0], info.Scripts[1])

	// Analyze with Common characters
	info = uax24.AnalyzeScripts("Hello123")
	fmt.Printf("Scripts: %d, HasCommon: %v, Mixed: %v\n",
		len(info.Scripts), info.HasCommon, info.IsMixedScript)

	// Output:
	// Scripts: 1, Mixed: false
	// Scripts: 2, Mixed: true
	// Script names: Latin, Cyrillic
	// Scripts: 1, HasCommon: true, Mixed: false
}

func ExampleIsSingleScript() {
	// Single script strings
	fmt.Println(uax24.IsSingleScript("Hello"))       // true - Latin only
	fmt.Println(uax24.IsSingleScript("Hello123"))    // true - Latin + Common (allowed)
	fmt.Println(uax24.IsSingleScript("中文"))        // true - Han only

	// Mixed script strings
	fmt.Println(uax24.IsSingleScript("Hello мир"))   // false - Latin + Cyrillic
	fmt.Println(uax24.IsSingleScript("Hello世界"))   // false - Latin + Han

	// Output:
	// true
	// true
	// true
	// false
	// false
}

func ExampleAnalyzeScripts_securityValidation() {
	// Example: Detecting homograph attacks by checking for mixed scripts
	identifiers := []string{
		"myVariable",     // Safe: Pure Latin
		"myVariаble",     // Unsafe: Latin + Cyrillic (а is Cyrillic)
		"用户名",          // Safe: Pure Han (Chinese)
		"userНаmе",       // Unsafe: Latin + Cyrillic
	}

	for _, id := range identifiers {
		info := uax24.AnalyzeScripts(id)
		if info.IsMixedScript {
			fmt.Printf("%q: UNSAFE - Mixed scripts: %v\n", id, info.Scripts)
		} else {
			fmt.Printf("%q: Safe - Single script\n", id)
		}
	}

	// Output:
	// "myVariable": Safe - Single script
	// "myVariаble": UNSAFE - Mixed scripts: [Latin Cyrillic]
	// "用户名": Safe - Single script
	// "userНаmе": UNSAFE - Mixed scripts: [Latin Cyrillic]
}

func ExampleAnalyzeScripts_languageDetection() {
	// Example: Using script analysis for language detection hints
	texts := []struct {
		text string
		lang string
	}{
		{"Hello world", "English"},
		{"Привет мир", "Russian"},
		{"你好世界", "Chinese"},
		{"こんにちは", "Japanese (Hiragana)"},
		{"مرحبا", "Arabic"},
		{"שלום", "Hebrew"},
		{"Γειά σου", "Greek"},
	}

	for _, t := range texts {
		script := uax24.LookupScript([]rune(t.text)[0])
		fmt.Printf("%s: %v\n", t.lang, script)
	}

	// Output:
	// English: Latin
	// Russian: Cyrillic
	// Chinese: Han
	// Japanese (Hiragana): Hiragana
	// Arabic: Arabic
	// Hebrew: Hebrew
	// Greek: Greek
}
