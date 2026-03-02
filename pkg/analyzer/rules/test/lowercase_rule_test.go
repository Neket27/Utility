package test

import (
	"strings"
	"testing"
	"unicode"
	"utility/pkg/analyzer/rules"
)

func TestCheckLowercase(t *testing.T) {
	tests := []struct {
		name        string
		msg         string
		wantValid   bool
		wantSuggest string
	}{
		// === Базовые валидные случаи ===
		{"empty", "", true, ""},
		{"single lowercase", "a", true, ""},
		{"lowercase word", "server started", true, ""},
		{"lowercase with numbers", "error 404", true, ""},
		{"lowercase with punctuation", "connection: ok", true, ""},

		// === Невалидные: uppercase start ===
		{"single uppercase", "A", false, "a"},
		{"uppercase start", "Starting server", false, "starting server"},
		{"uppercase long", "Database connection failed", false, "database connection failed"},
		{"uppercase with numbers", "Error 404 occurred", false, "error 404 occurred"},

		// === Первый символ НЕ буква — невалидно, без suggestion ===
		{"starts with digit", "123 items", false, ""},
		{"starts with space", " server", false, ""},
		{"starts with bracket", "(Error) failed", false, ""},
		{"starts with quote", "\"Error\" msg", false, ""},
		{"starts with emoji", "🚀 launched", false, ""},

		// === Unicode ===
		{"cyrillic lowercase", "сервер", true, ""},
		{"cyrillic uppercase", "Сервер", false, "сервер"},
		{"accented lowercase", "café", true, ""},
		{"accented uppercase", "Café", false, "café"},
		{"greek uppercase", "ΑΒΓ", false, "αΒΓ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, suggest := rules.CheckLowercase(tt.msg)

			if valid != tt.wantValid {
				t.Errorf("CheckLowercase(%q) valid = %v, want %v", tt.msg, valid, tt.wantValid)
			}
			if suggest != tt.wantSuggest {
				t.Errorf("CheckLowercase(%q) suggest = %q, want %q", tt.msg, suggest, tt.wantSuggest)
			}

			if !valid && suggest == "" && tt.msg != "" {
				first := []rune(tt.msg)[0]
				if unicode.IsLetter(first) && unicode.IsUpper(first) {
					t.Errorf("Expected suggestion for uppercase letter %q", first)
				}
			}
		})
	}
}

func TestLowercaseRule_Check(t *testing.T) {
	tests := []struct {
		name          string
		msg           string
		enabled       bool
		wantPassed    bool
		wantSuggested string
	}{
		{"valid lowercase", "server started", true, true, ""},
		{"empty", "", true, true, ""},
		{"uppercase start", "Server started", true, false, "server started"},
		{"unicode uppercase", "Запуск", true, false, "запуск"},
		{"starts with digit", "123 error", true, false, ""}, // нет suggestion
		{"starts with bracket", "[ERROR]", true, false, ""},
		{"disabled + uppercase", "Server", false, true, ""},
		{"suggestion preserves rest", "HTTP Error", true, false, "hTTP Error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := rules.NewLowercaseRule().(*rules.LowercaseRule)
			rule.SetEnabled(tt.enabled)

			ctx := &rules.CheckContext{Msg: tt.msg}
			result := rule.Check(ctx)

			if result.Passed != tt.wantPassed {
				t.Errorf("Check(%q) passed = %v, want %v", tt.msg, result.Passed, tt.wantPassed)
			}

			gotSuggestion := extractSuggestedText(result.SuggestedFix)

			if tt.wantSuggested != "" {
				if gotSuggestion != tt.wantSuggested {
					t.Errorf("SuggestedFix text = %q, want %q", gotSuggestion, tt.wantSuggested)
				}
			} else {
				if gotSuggestion != "" {
					t.Errorf("Expected no suggestion, got %q", gotSuggestion)
				}
			}
		})
	}
}

func TestLowercaseRule_Meta(t *testing.T) {
	rule := rules.NewLowercaseRule()

	t.Run("Name", func(t *testing.T) {
		if got := rule.Name(); got != rules.RuleLowercaseName {
			t.Errorf("Name() = %q, want %q", got, rules.RuleLowercaseName)
		}
	})

	t.Run("Description", func(t *testing.T) {
		desc := rule.Description()
		if desc == "" {
			t.Error("Description() should not be empty")
		}
		if !containsIgnoreCase(desc, "lowercase") {
			t.Errorf("Description() = %q, should mention 'lowercase'", desc)
		}
	})

	t.Run("Enabled by default", func(t *testing.T) {
		if !rule.Enabled() {
			t.Error("Rule should be enabled by default")
		}
	})
}

func TestLowercaseRule_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		msg  string
	}{
		{"long lowercase message", strings.Repeat("a", 1000)},
		{"long uppercase start", "A" + strings.Repeat("b", 1000)},
		{"only punctuation", "!@#$%"},
		{"only digits", "12345"},
		{"only whitespace", " \t\n "},
		{"surrogate pair emoji start", "😀text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := rules.NewLowercaseRule()
			ctx := &rules.CheckContext{Msg: tt.msg}
			result := rule.Check(ctx)

			if result == nil {
				t.Fatal("Check() returned nil")
			}

			if got := extractSuggestedText(result.SuggestedFix); got != "" {
				if len([]rune(got)) != len([]rune(tt.msg)) {
					t.Errorf("Suggestion length mismatch: input=%d runes, got=%d runes",
						len([]rune(tt.msg)), len([]rune(got)))
				}
			}
		})
	}
}
