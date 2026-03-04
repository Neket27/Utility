package rules

import (
	"testing"
)

func TestCheckNoSpecialChars(t *testing.T) {
	tests := []struct {
		name      string
		msg       string
		maxDots   int
		wantValid bool
	}{
		// === Базовые валидные случаи (ТЗ ✅) ===
		{"empty", "", 1, true},
		{"whitespace only", "   ", 1, true},
		{"valid simple", "server started", 1, true},
		{"valid with numbers", "error code 404", 1, true},

		// === Разрешённая пунктуация (одиночные знаки) ===
		{"single period", "server started.", 1, true},
		{"single comma", "server started, listening", 1, true},
		{"single colon", "port: 8080", 1, true},
		{"hyphen", "non-fatal error", 1, true},
		{"underscore", "user_name logged", 1, true},
		{"slash", "path/to/file", 1, true},
		{"parentheses", "status (ok)", 1, true},
		{"square brackets", "array[0] accessed", 1, true},
		{"curly braces", "config{env} loaded", 1, true},
		{"equals", "key=value", 1, true},
		{"plus", "cpu+memory", 1, true},
		{"double quote", `msg "quoted"`, 1, true},
		{"single quote", "it's working", 1, true},
		{"mixed allowed punctuation", "path/to/file.txt: ok", 1, true},

		// === Логика с точками (ТЗ: "..." ❌) ===
		{"single dot maxDots=1", "waiting.", 1, true},
		{"two dots maxDots=1", "waiting..", 1, false},
		{"ellipsis maxDots=1", "something went wrong...", 1, false},
		{"two dots maxDots=2", "waiting..", 2, true},
		{"ellipsis maxDots=2", "waiting...", 2, false},
		{"dots reset after space", "a. b", 1, true},
		{"dots reset after letter", "a.b.c", 1, true},

		// === Строгий режим: точки запрещены (maxDots=0) ===
		{"strict: single dot forbidden", "server started.", 0, false},
		{"strict: clean message passes", "server started", 0, true},

		// === Запрещённые спецсимволы (ТЗ ❌) ===
		{"exclamation", "server started!", 1, false},
		{"question mark", "what happened?", 1, false},
		{"at symbol", "user@localhost", 1, false},
		{"hash", "error #404", 1, false},
		{"dollar", "cost $100", 1, false},
		{"percent", "cpu 95%", 1, false},
		{"caret", "2^3=8", 1, false},
		{"ampersand", "A&B test", 1, false},
		{"asterisk", "wildcard *", 1, false},
		{"pipe", "a|b", 1, false},
		{"backtick", "cmd `run`", 1, false},
		{"tilde", "version ~1.0", 1, false},
		{"multiple exclamation", "failed!!", 1, false},

		// === Эмодзи (ТЗ ❌) ===
		{"emoji emoticon", "started 😀", 1, false},
		{"emoji misc", "status ♻️", 1, false},
		{"emoji transport", "car 🚗 moved", 1, false},
		{"emoji flag", "country 🇷🇺", 1, false},

		// === Unicode-буквы (разрешены, проверка языка — отдельное правило) ===
		{"cyrillic", "сервер запущен", 1, true},
		{"chinese", "服务器启动", 1, true},
		{"accented", "café résumé", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, _ := CheckNoSpecialChars(tt.msg, tt.maxDots)
			if valid != tt.wantValid {
				t.Errorf("CheckNoSpecialChars(%q, maxDots=%d) = %v, want %v",
					tt.msg, tt.maxDots, valid, tt.wantValid)
			}
		})
	}
}

func TestNoSpecialCharsRule_Check(t *testing.T) {
	tests := []struct {
		name           string
		config         map[string]any
		msg            string
		wantPassed     bool
		wantSuggestion bool
	}{
		{
			name:           "ТЗ valid: server started",
			config:         map[string]any{"enabled": true, "auto_fix_enabled": true, "max_consecutive_dots": 0},
			msg:            "server started",
			wantPassed:     true,
			wantSuggestion: false,
		},

		{
			name:           "ТЗ invalid: exclamation + emoji",
			config:         map[string]any{"enabled": true, "auto_fix_enabled": true, "max_consecutive_dots": 1},
			msg:            "server started! 🚀",
			wantPassed:     false,
			wantSuggestion: true,
		},

		{
			name:           "ТЗ invalid: multiple exclamation",
			config:         map[string]any{"enabled": true, "auto_fix_enabled": true, "max_consecutive_dots": 1},
			msg:            "connection failed!!!",
			wantPassed:     false,
			wantSuggestion: true,
		},

		{
			name:           "ТЗ invalid: ellipsis with maxDots=0",
			config:         map[string]any{"enabled": true, "auto_fix_enabled": true, "max_consecutive_dots": 0},
			msg:            "something went wrong...",
			wantPassed:     false,
			wantSuggestion: true,
		},

		{
			name:           "strict mode: dot forbidden (maxDots=0)",
			config:         map[string]any{"enabled": true, "auto_fix_enabled": true, "max_consecutive_dots": 0},
			msg:            "server started.",
			wantPassed:     false,
			wantSuggestion: true,
		},

		{
			name:           "lenient mode: one dot allowed (maxDots=1)",
			config:         map[string]any{"enabled": true, "auto_fix_enabled": true, "max_consecutive_dots": 1},
			msg:            "server started.",
			wantPassed:     true,
			wantSuggestion: false,
		},

		{
			name:           "lenient mode: two dots allowed (maxDots=2)",
			config:         map[string]any{"enabled": true, "auto_fix_enabled": true, "max_consecutive_dots": 2},
			msg:            "waiting..",
			wantPassed:     true,
			wantSuggestion: false,
		},

		{
			name:           "ellipsis forbidden even with maxDots=2",
			config:         map[string]any{"enabled": true, "auto_fix_enabled": true, "max_consecutive_dots": 2},
			msg:            "waiting...",
			wantPassed:     false,
			wantSuggestion: true,
		},

		{
			name:           "autoFix enabled: suggestion provided",
			config:         map[string]any{"enabled": true, "auto_fix_enabled": true, "max_consecutive_dots": 0},
			msg:            "test!@#",
			wantPassed:     false,
			wantSuggestion: true,
		},

		{
			name:           "autoFix disabled: no suggestion",
			config:         map[string]any{"enabled": true, "auto_fix_enabled": false, "max_consecutive_dots": 0},
			msg:            "test!@#",
			wantPassed:     false,
			wantSuggestion: false,
		},

		{
			name:           "disabled rule always passes",
			config:         map[string]any{"enabled": false, "auto_fix_enabled": true},
			msg:            "server! 🚀",
			wantPassed:     true,
			wantSuggestion: false,
		},

		{
			name:           "allowed punctuation preserved in suggestion",
			config:         map[string]any{"enabled": true, "auto_fix_enabled": true, "max_consecutive_dots": 1},
			msg:            "path/to/file.txt",
			wantPassed:     true,
			wantSuggestion: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewNoSpecialCharsRule()

			if err := rule.Configure(tt.config); err != nil {
				t.Fatalf("Configure() error = %v", err)
			}

			ctx := &CheckContext{Msg: tt.msg}
			result := rule.Check(ctx)

			if result.Passed != tt.wantPassed {
				t.Errorf("Check(%q) passed = %v, want %v", tt.msg, result.Passed, tt.wantPassed)
			}

			hasSuggestion := result.SuggestedFix != nil
			if hasSuggestion != tt.wantSuggestion {
				t.Errorf("Check(%q) has SuggestedFix = %v, want %v",
					tt.msg, hasSuggestion, tt.wantSuggestion)
			}
		})
	}
}

// === Тесты для CleanSpecialChars ===

func TestCleanSpecialChars(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no special chars", "hello world", "hello world"},
		{"remove exclamation", "hello!", "hello"},
		{"remove emoji", "hi 😀", "hi "},
		{"keep allowed punct", "path/to/file.txt", "path/to/file.txt"},
		{"mixed cleanup", "test!@# ok.", "test ok."},
		{"empty", "", ""},
		{"only special", "!@#", ""},
		{"preserve unicode", "сервер! ok", "сервер ok"},
		{"ellipsis kept by cleaner", "waiting...", "waiting..."},
		{"preserve quotes", `msg "quoted"`, `msg "quoted"`},
		{"preserve parentheses", "status (ok)", "status (ok)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanSpecialChars(tt.input)
			if got != tt.want {
				t.Errorf("CleanSpecialChars(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNoSpecialCharsRule_AutoFix(t *testing.T) {
	tests := []struct {
		name          string
		config        map[string]any
		msg           string
		wantPassed    bool
		wantSuggested string
	}{

		{
			name:          "autoFix=true: remove emoji",
			config:        map[string]any{"enabled": true, "auto_fix_enabled": true, "max_consecutive_dots": 0},
			msg:           "started 😀",
			wantPassed:    false,
			wantSuggested: "started ",
		},
		{
			name:          "autoFix=true: remove special chars",
			config:        map[string]any{"enabled": true, "auto_fix_enabled": true, "max_consecutive_dots": 0},
			msg:           "test!@#",
			wantPassed:    false,
			wantSuggested: "test",
		},
		{
			name:          "autoFix=true: preserve allowed punct",
			config:        map[string]any{"enabled": true, "auto_fix_enabled": true, "max_consecutive_dots": 1},
			msg:           "path/to/file.txt",
			wantPassed:    true,
			wantSuggested: "",
		},

		{
			name:          "autoFix=false: no suggestion for emoji",
			config:        map[string]any{"enabled": true, "auto_fix_enabled": false, "max_consecutive_dots": 0},
			msg:           "started 😀",
			wantPassed:    false,
			wantSuggested: "",
		},
		{
			name:          "autoFix=false: no suggestion for special chars",
			config:        map[string]any{"enabled": true, "auto_fix_enabled": false, "max_consecutive_dots": 0},
			msg:           "test!@#",
			wantPassed:    false,
			wantSuggested: "",
		},

		{
			name:          "enabled=false: no check, no suggestion",
			config:        map[string]any{"enabled": false, "auto_fix_enabled": true},
			msg:           "test!@# 😀",
			wantPassed:    true,
			wantSuggested: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewNoSpecialCharsRule()

			if err := rule.Configure(tt.config); err != nil {
				t.Fatalf("Configure() error = %v", err)
			}

			ctx := &CheckContext{Msg: tt.msg}
			result := rule.Check(ctx)

			if result.Passed != tt.wantPassed {
				t.Errorf("Check(%q) passed = %v, want %v", tt.msg, result.Passed, tt.wantPassed)
			}

			gotSuggestion := ExtractSuggestedText(result.SuggestedFix)
			if gotSuggestion != tt.wantSuggested {
				t.Errorf("SuggestedFix = %q, want %q", gotSuggestion, tt.wantSuggested)
			}
		})
	}
}

func TestNoSpecialCharsRule_Configure(t *testing.T) {
	tests := []struct {
		name           string
		config         map[string]any
		msg            string
		wantPassed     bool
		wantSuggestion bool
	}{
		{
			name:           "default config (maxDots=0, autoFix=true)",
			config:         map[string]any{},
			msg:            "server started.",
			wantPassed:     false,
			wantSuggestion: true,
		},
		{
			name:           "configure max_consecutive_dots=1",
			config:         map[string]any{"max_consecutive_dots": 1},
			msg:            "server started.",
			wantPassed:     true,
			wantSuggestion: false,
		},
		{
			name:           "configure auto_fix_enabled=false",
			config:         map[string]any{"auto_fix_enabled": false},
			msg:            "test!",
			wantPassed:     false,
			wantSuggestion: false,
		},
		{
			name:           "configure enabled=false",
			config:         map[string]any{"enabled": false},
			msg:            "test! 🚀",
			wantPassed:     true,
			wantSuggestion: false,
		},
		{
			name:           "full config",
			config:         map[string]any{"enabled": true, "auto_fix_enabled": true, "max_consecutive_dots": 2},
			msg:            "waiting..",
			wantPassed:     true,
			wantSuggestion: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewNoSpecialCharsRule()

			if err := rule.Configure(tt.config); err != nil {
				t.Fatalf("Configure() error = %v", err)
			}

			ctx := &CheckContext{Msg: tt.msg}
			result := rule.Check(ctx)

			if result.Passed != tt.wantPassed {
				t.Errorf("Check(%q) passed = %v, want %v", tt.msg, result.Passed, tt.wantPassed)
			}

			hasSuggestion := result.SuggestedFix != nil
			if hasSuggestion != tt.wantSuggestion {
				t.Errorf("Check(%q) has SuggestedFix = %v, want %v",
					tt.msg, hasSuggestion, tt.wantSuggestion)
			}
		})
	}
}
