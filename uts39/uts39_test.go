package uts39

import (
	"testing"

	"github.com/SCKelemen/unicode/uax24"
)

func TestSkeleton(t *testing.T) {
	tests := []struct {
		name string
		s1   string
		s2   string
		want bool // Should have same skeleton
	}{
		{
			name: "Identical strings",
			s1:   "scope",
			s2:   "scope",
			want: true,
		},
		{
			name: "Cyrillic confusables",
			s1:   "scope",
			s2:   "\u0455\u0441\u043E\u0440\u0435", // ѕсоре (Cyrillic lookalikes)
			want: true,
		},
		{
			name: "Latin confusables",
			s1:   "paypal",
			s2:   "p\u0430ypal", // pаypal (Cyrillic 'а')
			want: true,
		},
		{
			name: "Different strings",
			s1:   "hello",
			s2:   "world",
			want: false,
		},
		{
			name: "Case differences",
			s1:   "Hello",
			s2:   "hello",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skel1 := Skeleton(tt.s1)
			skel2 := Skeleton(tt.s2)
			got := skel1 == skel2

			if got != tt.want {
				t.Errorf("Skeleton() comparison failed:\n  s1=%q skeleton=%q\n  s2=%q skeleton=%q\n  got=%v, want=%v",
					tt.s1, skel1, tt.s2, skel2, got, tt.want)
			}
		})
	}
}

func TestAreConfusable(t *testing.T) {
	tests := []struct {
		name string
		s1   string
		s2   string
		want bool
	}{
		{"ASCII identical", "test", "test", true},
		{"ASCII different", "test", "best", false},
		{"Cyrillic a", "paypal", "p\u0430ypal", true}, // Cyrillic а
		{"Mixed case", "Test", "test", true},
		{"Greek rho vs Latin p", "scope", "\u03C1cope", false}, // ρ is not confusable with s
		{"Empty strings", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AreConfusable(tt.s1, tt.s2)
			if got != tt.want {
				t.Errorf("AreConfusable(%q, %q) = %v, want %v", tt.s1, tt.s2, got, tt.want)
			}
		})
	}
}

func TestGetRestrictionLevel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  RestrictionLevel
	}{
		{
			name:  "ASCII only",
			input: "hello_world",
			want:  ASCIIOnly,
		},
		{
			name:  "Single script Latin",
			input: "café",
			want:  SingleScript,
		},
		{
			name:  "Single script Cyrillic",
			input: "привет",
			want:  SingleScript,
		},
		{
			name:  "Latin + Han (minimally restrictive)",
			input: "hello世界",
			want:  MinimallyRestrictive,
		},
		{
			name:  "Mixed Cyrillic and Latin",
			input: "hello мир",
			want:  MinimallyRestrictive,
		},
		{
			name:  "Empty string",
			input: "",
			want:  Unrestricted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRestrictionLevel(tt.input)
			if got != tt.want {
				t.Errorf("GetRestrictionLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsMixedScript(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"ASCII only", "hello", false},
		{"Latin only", "café", false},
		{"Cyrillic only", "привет", false},
		{"Latin + Cyrillic", "hello мир", true},
		{"Latin + Han", "hello世界", true},
		{"With numbers", "hello123", false}, // Numbers are Common
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsMixedScript(tt.input)
			if got != tt.want {
				t.Errorf("IsMixedScript(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetIdentifierScripts(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []uax24.Script
	}{
		{
			name:  "Latin only",
			input: "hello",
			want:  []uax24.Script{uax24.ScriptLatin},
		},
		{
			name:  "Latin with numbers",
			input: "hello123",
			want:  []uax24.Script{uax24.ScriptLatin, uax24.ScriptCommon},
		},
		{
			name:  "Cyrillic",
			input: "привет",
			want:  []uax24.Script{uax24.ScriptCyrillic},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetIdentifierScripts(tt.input)
			if !scriptsEqual(got, tt.want) {
				t.Errorf("GetIdentifierScripts(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidIdentifier(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"Valid ASCII", "myVar", true},
		{"Valid with underscore", "my_var", true},
		{"Valid with digits", "var123", true},
		{"Invalid - starts with digit", "123var", false},
		{"Invalid - hyphen", "my-var", false},
		{"Invalid - space", "my var", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidIdentifier(tt.input)
			if got != tt.want {
				t.Errorf("IsValidIdentifier(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsSafeIdentifier(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"Safe ASCII", "user_name", true},
		{"Safe Latin", "userName", true},
		{"Unsafe - mixed script", "user\u043Dame", false}, // Contains Cyrillic н
		{"Unsafe - zero-width space", "user\u200Bname", false},
		{"Unsafe - invalid start", "123user", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSafeIdentifier(tt.input)
			if got != tt.want {
				t.Errorf("IsSafeIdentifier(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestRestrictionLevelString(t *testing.T) {
	tests := []struct {
		level RestrictionLevel
		want  string
	}{
		{ASCIIOnly, "ASCII-Only"},
		{SingleScript, "Single-Script"},
		{HighlyRestrictive, "Highly-Restrictive"},
		{ModeratelyRestrictive, "Moderately-Restrictive"},
		{MinimallyRestrictive, "Minimally-Restrictive"},
		{Unrestricted, "Unrestricted"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.level.String()
			if got != tt.want {
				t.Errorf("RestrictionLevel.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsInvisible(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{"Zero Width Space", '\u200B', true},
		{"Zero Width Joiner", '\u200D', true},
		{"Regular space", ' ', false},
		{"ASCII letter", 'a', false},
		{"Zero Width No-Break Space", '\uFEFF', true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isInvisible(tt.r)
			if got != tt.want {
				t.Errorf("isInvisible(%U) = %v, want %v", tt.r, got, tt.want)
			}
		})
	}
}

// Helper function to compare script slices
func scriptsEqual(a, b []uax24.Script) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps for comparison
	aMap := make(map[uax24.Script]bool)
	for _, script := range a {
		aMap[script] = true
	}

	for _, script := range b {
		if !aMap[script] {
			return false
		}
	}

	return true
}

// Benchmark tests
func BenchmarkSkeleton(b *testing.B) {
	s := "paypal"
	for i := 0; i < b.N; i++ {
		Skeleton(s)
	}
}

func BenchmarkAreConfusable(b *testing.B) {
	s1 := "paypal"
	s2 := "p\u0430ypal" // Cyrillic а
	for i := 0; i < b.N; i++ {
		AreConfusable(s1, s2)
	}
}

func BenchmarkGetRestrictionLevel(b *testing.B) {
	s := "hello_world"
	for i := 0; i < b.N; i++ {
		GetRestrictionLevel(s)
	}
}

func BenchmarkIsMixedScript(b *testing.B) {
	s := "hello世界"
	for i := 0; i < b.N; i++ {
		IsMixedScript(s)
	}
}
