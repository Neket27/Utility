package test

import (
	"testing"
	"utility/pkg/analyzer/rules"
)

func TestCheckNoSpecialChars(t *testing.T) {
	tests := []struct {
		name      string
		msg       string
		wantValid bool
	}{
		// === Базовые валидные случаи ===
		{"empty", "", true},
		{"whitespace only", "   ", true},
		{"valid simple", "server started", true},
		{"valid with numbers", "error code 404", true},
		{"valid with period", "server started.", true},
		{"valid with comma", "server started, listening", true},
		{"valid with colon", "port: 8080", true},

		// === Все разрешённые знаки препинания ===
		{"allowed: hyphen", "non-fatal error", true},
		{"allowed: underscore", "user_name logged", true},
		{"allowed: slash", "path/to/file", true},
		{"allowed: parentheses", "status (ok)", true},
		{"allowed: square brackets", "array[0] accessed", true},
		{"allowed: curly braces", "config{env} loaded", true},
		{"allowed: equals", "key=value", true},
		{"allowed: plus", "cpu+memory", true},
		{"allowed: double quote", `msg "quoted"`, true},
		{"allowed: single quote", "it's working", true},
		{"allowed: mixed punctuation", "path/to/file.txt: ok", true},

		// === Логика с точками ===
		{"three dots ellipsis", "waiting.", true},
		{"four dots", "really..", false},
		{"dots reset after space", "a. b", true},
		{"dots reset after letter", "a.b.c", true},

		// === Запрещённые специальные символы ===
		{"invalid: exclamation", "server started!", false},
		{"invalid: question", "what happened?", false},
		{"invalid: at symbol", "user@localhost", false},
		{"invalid: hash", "error #404", false},
		{"invalid: dollar", "cost $100", false},
		{"invalid: percent", "cpu 95%", false},
		{"invalid: caret", "2^3=8", false},
		{"invalid: ampersand", "A&B test", false},
		{"invalid: asterisk", "wildcard *", false},
		{"invalid: pipe", "a|b", false},
		{"invalid: backtick", "cmd `run`", false},
		{"invalid: tilde", "version ~1.0", false},
		{"invalid: double exclamation", "failed!!", false},

		// === Эмодзи ===
		{"emoji: emoticon", "started 😀", false},
		{"emoji: misc symbols", "status ♻️", false},
		{"emoji: transport", "car 🚗 moved", false},
		{"emoji: flag", "country 🇷🇺", false},

		// === Unicode буквы (должны быть валидны) ===
		{"cyrillic letters", "сервер запущен", true},
		{"chinese letters", "服务器启动", true},
		{"accented letters", "café résumé", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, _ := rules.CheckNoSpecialChars(tt.msg)
			if valid != tt.wantValid {
				t.Errorf("CheckNoSpecialChars(%q) valid = %v, want %v", tt.msg, valid, tt.wantValid)
			}
		})
	}
}

func TestNoSpecialCharsRule_Check(t *testing.T) {
	tests := []struct {
		name           string
		msg            string
		enabled        bool
		wantPassed     bool
		wantSuggestion bool
	}{
		{"enabled valid", "server started", true, true, false},
		{"enabled invalid", "server started!", true, false, true},
		{"disabled always passes", "server!", false, true, false},
		{"emoji removed in suggestion", "ok 😀", true, false, true},
		{"multiple special cleaned", "test!@#", true, false, true},
		{"allowed punctuation preserved", "path/to/file.txt", true, true, false},
		{"ellipsis triggers suggestion", "waiting...", true, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
		{"preserve unicode letters", "сервер! ok", "сервер ok"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rules.CleanSpecialChars(tt.input)
			if got != tt.want {
				t.Errorf("cleanSpecialChars(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
