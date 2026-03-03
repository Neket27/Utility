package test

import (
	"testing"
	"utility/pkg/analyzer/rules"
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
			valid, _ := rules.CheckNoSpecialChars(tt.msg, tt.maxDots)
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
		msg            string
		maxDots        int
		enabled        bool
		wantPassed     bool
		wantSuggestion bool
	}{
		// ТЗ ✅ примеры
		{"ТЗ valid: server started", "server started", 0, true, true, false},

		// ТЗ ❌ примеры
		{"ТЗ invalid: exclamation + emoji", "server started! 🚀", 1, true, false, true},
		{"ТЗ invalid: multiple exclamation", "connection failed!!!", 1, true, false, true},
		{"ТЗ invalid: ellipsis", "something went wrong...", 1, true, false, true},

		// Конфигурируемость
		{"strict mode: dot forbidden", "server started.", 0, true, false, true},
		{"lenient mode: two dots allowed", "waiting..", 2, true, true, false},

		// Включение/выключение правила
		{"disabled rule always passes", "server!", 1, false, true, false},

		// Проверка предложений по исправлению
		{"emoji removed in suggestion", "ok 😀", 1, true, false, true},
		{"special chars cleaned", "test!@#", 1, true, false, true},
		{"allowed punctuation preserved", "path/to/file.txt", 1, true, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules.MaxConsecutiveDots = tt.maxDots

			rule := rules.NewNoSpecialCharsRule()
			rule.SetEnabled(tt.enabled)

			ctx := &rules.CheckContext{Msg: tt.msg}
			result := rule.Check(ctx)

			if result.Passed != tt.wantPassed {
				t.Errorf("Check() passed = %v, want %v", result.Passed, tt.wantPassed)
			}

			hasSuggestion := result.SuggestedFix != nil
			if hasSuggestion != tt.wantSuggestion {
				t.Errorf("Check() has SuggestedFix = %v, want %v", hasSuggestion, tt.wantSuggestion)
			}
		})
	}
}

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rules.CleanSpecialChars(tt.input)
			if got != tt.want {
				t.Errorf("CleanSpecialChars(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
