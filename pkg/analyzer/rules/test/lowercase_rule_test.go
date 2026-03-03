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

func TestLowercaseRule_AutoFixEnabled(t *testing.T) {
	tests := []struct {
		name          string
		config        map[string]any
		msg           string
		wantPassed    bool
		wantSuggested string
	}{
		// Автофикс включён (по умолчанию) — предложения есть
		{
			name:          "autoFix=true: uppercase",
			config:        map[string]any{"enabled": true, "auto_fix_enabled": true},
			msg:           "Server started",
			wantPassed:    false,
			wantSuggested: "server started",
		},
		{
			name:          "autoFix=true: unicode",
			config:        map[string]any{"enabled": true, "auto_fix_enabled": true},
			msg:           "Запуск сервера",
			wantPassed:    false,
			wantSuggested: "запуск сервера",
		},
		{
			name:          "autoFix=true: valid msg",
			config:        map[string]any{"enabled": true, "auto_fix_enabled": true},
			msg:           "server started",
			wantPassed:    true,
			wantSuggested: "",
		},

		// Автофикс выключен — предложений нет, но ошибка детектится
		{
			name:          "autoFix=false: uppercase",
			config:        map[string]any{"enabled": true, "auto_fix_enabled": false},
			msg:           "Server started",
			wantPassed:    false,
			wantSuggested: "",
		},
		{
			name:          "autoFix=false: unicode",
			config:        map[string]any{"enabled": true, "auto_fix_enabled": false},
			msg:           "Запуск",
			wantPassed:    false,
			wantSuggested: "",
		},
		{
			name:          "autoFix=false: valid msg",
			config:        map[string]any{"enabled": true, "auto_fix_enabled": false},
			msg:           "server started",
			wantPassed:    true,
			wantSuggested: "",
		},

		// Правило выключено — всегда проходит, без фиксов
		{
			name:          "enabled=false, autoFix=true",
			config:        map[string]any{"enabled": false, "auto_fix_enabled": true},
			msg:           "Server",
			wantPassed:    true,
			wantSuggested: "",
		},
		{
			name:          "enabled=false, autoFix=false",
			config:        map[string]any{"enabled": false, "auto_fix_enabled": false},
			msg:           "Server",
			wantPassed:    true,
			wantSuggested: "",
		},

		// Конфиг по умолчанию (auto_fix_enabled=true)
		{
			name:          "default config",
			config:        map[string]any{},
			msg:           "Starting",
			wantPassed:    false,
			wantSuggested: "starting",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := rules.NewLowercaseRule()

			if err := rule.Configure(tt.config); err != nil {
				t.Fatalf("Configure() error = %v", err)
			}

			ctx := &rules.CheckContext{Msg: tt.msg}
			result := rule.Check(ctx)

			if result.Passed != tt.wantPassed {
				t.Errorf("Check(%q) passed = %v, want %v", tt.msg, result.Passed, tt.wantPassed)
			}

			gotSuggestion := extractSuggestedText(result.SuggestedFix)
			if gotSuggestion != tt.wantSuggested {
				t.Errorf("SuggestedFix = %q, want %q", gotSuggestion, tt.wantSuggested)
			}
		})
	}
}

func TestLowercaseRule_Configure_AutoFix(t *testing.T) {
	tests := []struct {
		name          string
		config        map[string]any
		msg           string
		wantSuggested string
	}{
		{
			name:          "configure auto_fix_enabled=true",
			config:        map[string]any{"auto_fix_enabled": true},
			msg:           "ERROR occurred",
			wantSuggested: "eRROR occurred",
		},
		{
			name:          "configure auto_fix_enabled=false",
			config:        map[string]any{"auto_fix_enabled": false},
			msg:           "ERROR occurred",
			wantSuggested: "",
		},
		{
			name:          "configure with both flags",
			config:        map[string]any{"enabled": true, "auto_fix_enabled": false},
			msg:           "Starting",
			wantSuggested: "",
		},
		{
			name:          "default config (auto_fix enabled)",
			config:        map[string]any{},
			msg:           "Starting",
			wantSuggested: "starting",
		},
		{
			name:          "enabled=false overrides autoFix",
			config:        map[string]any{"enabled": false, "auto_fix_enabled": true},
			msg:           "Starting",
			wantSuggested: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := rules.NewLowercaseRule()

			if err := rule.Configure(tt.config); err != nil {
				t.Fatalf("Configure() error = %v", err)
			}

			ctx := &rules.CheckContext{Msg: tt.msg}
			result := rule.Check(ctx)

			gotSuggestion := extractSuggestedText(result.SuggestedFix)
			if gotSuggestion != tt.wantSuggested {
				t.Errorf("SuggestedFix = %q, want %q", gotSuggestion, tt.wantSuggested)
			}
		})
	}
}

func TestLowercaseRule_AutoFix_Safety(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		config   map[string]any
		expected string
	}{
		{
			name:     "preserve numbers",
			msg:      "Error404",
			config:   map[string]any{"auto_fix_enabled": true},
			expected: "error404",
		},
		{
			name:     "preserve punctuation",
			msg:      "Error: failed!",
			config:   map[string]any{"auto_fix_enabled": true},
			expected: "error: failed!",
		},
		{
			name:     "preserve spaces",
			msg:      "UPPERCASE  ",
			config:   map[string]any{"auto_fix_enabled": true},
			expected: "uPPERCASE  ",
		},
		{
			name:     "preserve unicode rest",
			msg:      "Привет Мир",
			config:   map[string]any{"auto_fix_enabled": true},
			expected: "привет Мир",
		},
		{
			name:     "mixed case",
			msg:      "HTTPServerError",
			config:   map[string]any{"auto_fix_enabled": true},
			expected: "hTTPServerError",
		},
		{
			name:     "autoFix disabled - no suggestion",
			msg:      "Error404",
			config:   map[string]any{"auto_fix_enabled": false},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := rules.NewLowercaseRule()

			if err := rule.Configure(tt.config); err != nil {
				t.Fatalf("Configure() error = %v", err)
			}

			ctx := &rules.CheckContext{Msg: tt.msg}
			result := rule.Check(ctx)

			if tt.expected == "" {
				if result.SuggestedFix != nil {
					t.Errorf("Expected no suggestion, got %q", extractSuggestedText(result.SuggestedFix))
				}
				return
			}

			if result.Passed {
				t.Skip("message is valid, no fix needed")
			}

			got := extractSuggestedText(result.SuggestedFix)
			if got != tt.expected {
				t.Errorf("SuggestedFix = %q, want %q", got, tt.expected)
			}

			if len([]rune(got)) != len([]rune(tt.msg)) {
				t.Errorf("Suggestion rune length mismatch: input=%d, got=%d",
					len([]rune(tt.msg)), len([]rune(got)))
			}
		})
	}
}
