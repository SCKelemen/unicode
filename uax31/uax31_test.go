package uax31

import (
	"testing"
)

func TestIsXIDStart(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		// ASCII letters
		{"ASCII uppercase", 'A', true},
		{"ASCII lowercase", 'z', true},

		// ASCII non-letters
		{"ASCII digit", '5', false},
		{"ASCII underscore", '_', false},
		{"ASCII space", ' ', false},
		{"ASCII plus", '+', false},

		// Unicode letters
		{"Greek alpha", 'α', true},
		{"Cyrillic A", 'А', true},
		{"Hebrew alef", 'א', true},
		{"Arabic ain", 'ع', true},
		{"Han ideograph", '中', true},
		{"Hiragana", 'あ', true},

		// Punctuation and symbols
		{"Exclamation", '!', false},
		{"Asterisk", '*', false},
		{"Parenthesis", '(', false},

		// Whitespace
		{"Tab", '\t', false},
		{"Newline", '\n', false},

		// Edge cases
		{"NULL", '\x00', false},
		{"Replacement character", '\uFFFD', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsXIDStart(tt.r); got != tt.want {
				t.Errorf("IsXIDStart(%U) = %v, want %v", tt.r, got, tt.want)
			}
		})
	}
}

func TestIsXIDContinue(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		// XID_Start characters are also XID_Continue
		{"ASCII uppercase", 'A', true},
		{"ASCII lowercase", 'z', true},

		// Additional XID_Continue characters
		{"ASCII digit", '5', true},
		{"ASCII underscore", '_', true},
		{"Combining acute", '\u0301', true},
		{"Zero-width joiner", '\u200D', true},

		// Not XID_Continue
		{"ASCII space", ' ', false},
		{"ASCII plus", '+', false},
		{"ASCII hyphen", '-', false},

		// Edge cases
		{"NULL", '\x00', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsXIDContinue(tt.r); got != tt.want {
				t.Errorf("IsXIDContinue(%U) = %v, want %v", tt.r, got, tt.want)
			}
		})
	}
}

func TestIsPatternSyntax(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		// Pattern syntax characters
		{"Exclamation", '!', true},
		{"Number sign", '#', true},
		{"Dollar", '$', true},
		{"Percent", '%', true},
		{"Asterisk", '*', true},
		{"Plus", '+', true},
		{"Comma", ',', true},
		{"Hyphen", '-', true},
		{"Period", '.', true},
		{"Slash", '/', true},
		{"Colon", ':', true},
		{"Semicolon", ';', true},
		{"Less than", '<', true},
		{"Equals", '=', true},
		{"Greater than", '>', true},
		{"Question mark", '?', true},
		{"At sign", '@', true},
		{"Left bracket", '[', true},
		{"Backslash", '\\', true},
		{"Right bracket", ']', true},
		{"Caret", '^', true},
		{"Backtick", '`', true},
		{"Left brace", '{', true},
		{"Pipe", '|', true},
		{"Right brace", '}', true},
		{"Tilde", '~', true},

		// Not pattern syntax
		{"ASCII letter", 'A', false},
		{"ASCII digit", '5', false},
		{"ASCII underscore", '_', false},
		{"Space", ' ', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPatternSyntax(tt.r); got != tt.want {
				t.Errorf("IsPatternSyntax(%U) = %v, want %v", tt.r, got, tt.want)
			}
		})
	}
}

func TestIsPatternWhiteSpace(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		// Pattern whitespace
		{"Space", ' ', true},
		{"Tab", '\t', true},
		{"Newline", '\n', true},
		{"Carriage return", '\r', true},
		{"Form feed", '\f', true},
		{"Vertical tab", '\v', true},

		// Not pattern whitespace
		{"ASCII letter", 'A', false},
		{"ASCII digit", '5', false},
		{"Non-breaking space", '\u00A0', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPatternWhiteSpace(tt.r); got != tt.want {
				t.Errorf("IsPatternWhiteSpace(%U) = %v, want %v", tt.r, got, tt.want)
			}
		})
	}
}

func TestIsValidIdentifier(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		// Valid identifiers
		{"Simple ASCII", "myVar", true},
		{"With underscore", "my_var", true},
		{"With digits", "myVar123", true},
		{"Unicode Chinese", "变量", true},
		{"Unicode Russian", "переменная", true},
		{"Unicode Greek", "μετβλητή", true},
		{"Unicode Japanese", "変数", true},
		{"Mixed scripts", "myВар", true},

		// Invalid identifiers
		{"Starts with digit", "123var", false},
		{"Starts with underscore", "_private", false}, // Underscore is XID_Continue but not XID_Start
		{"Only underscore", "_", false},               // Underscore is XID_Continue but not XID_Start
		{"Contains hyphen", "my-var", false},
		{"Contains space", "my var", false},
		{"Contains plus", "my+var", false},
		{"Empty string", "", false},
		{"Contains pattern syntax", "my*var", false},
		{"Starts with pattern syntax", "*var", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidIdentifier(tt.s); got != tt.want {
				t.Errorf("IsValidIdentifier(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestXIDStartContinueRelationship(t *testing.T) {
	// All XID_Start characters should also be XID_Continue
	// Test a sample of XID_Start characters
	testRunes := []rune{'A', 'z', 'α', '中', 'א', 'ع'}

	for _, r := range testRunes {
		if IsXIDStart(r) && !IsXIDContinue(r) {
			t.Errorf("Rune %U is XID_Start but not XID_Continue", r)
		}
	}
}

// Benchmark tests
func BenchmarkIsXIDStart(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsXIDStart('A')
	}
}

func BenchmarkIsXIDContinue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsXIDContinue('5')
	}
}

func BenchmarkIsPatternSyntax(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsPatternSyntax('*')
	}
}

func BenchmarkIsValidIdentifier(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsValidIdentifier("myVar123")
	}
}

func BenchmarkIsValidIdentifier_Unicode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsValidIdentifier("переменная")
	}
}
