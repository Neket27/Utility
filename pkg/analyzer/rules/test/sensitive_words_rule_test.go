package test

import (
	"strings"
	"testing"
	"utility/pkg/analyzer/rules"
)

func TestSensitiveWordsRule_DefaultWords(t *testing.T) {
	tests := []struct {
		name       string
		msg        string
		wantPassed bool
		wantWord   string
	}{
		// === Базовые чувствительные слова ===
		{"password lowercase", "user password: 123", false, "password"},
		{"password uppercase", "PASSWORD leaked", false, "PASSWORD"},
		{"password mixed case", "PaSsWoRd field", false, "PaSsWoRd"},
		{"passwd variant", "passwd file", false, "passwd"},
		{"pwd abbreviation", "pwd=secret", false, "pwd"},
		{"token generic", "bearer token", false, "token"},
		{"api_key underscore", "api_key=abc", false, "api_key"},
		{"apikey no underscore", "apikey xyz", false, "apikey"},
		{"credential plural", "user credential", false, "credential"},
		{"private_key", "private_key loaded", true, "private_key"},
		{"access_token", "access_token expired", true, "access_token"},
		{"refresh_token", "refresh_token used", false, "refresh_token"},
		{"secret_key", "secret_key value", false, "secret_key"},
		{"encryption_key", "encryption_key rotated", true, "encryption_key"},
		{"secret standalone", "secret value", false, "secret"},

		// === Word boundary: частичные совпадения НЕ должны срабатывать ===
		{"auth inside authenticated", "user authenticated", true, ""},
		{"auth inside authorization", "authorization header", true, ""},
		{"password inside passwords", "passwords policy", true, ""},
		{"token inside tokenizer", "tokenizer initialized", true, ""},
		{"secret inside secrecy", "secrecy mode", true, ""},
		{"key inside keyboard", "keyboard input", true, ""},

		// === Safe phrases: чувствительные слова в безопасном контексте ===
		{"password validated", "password validated", true, ""},
		{"token expired", "access_token expired", true, ""},
		{"secret rotated", "secret_key rotated", true, ""},
		{"credential refreshed", "credential refreshed", true, ""},
		{"api_key changed", "api_key changed successfully", true, ""},
		{"password updated", "password updated for user", true, ""},
		{"token revoked", "token revoked by admin", true, ""},
		{"secret deleted", "secret deleted from cache", true, ""},
		{"password required", "password field is required", true, ""},
		{"token optional", "token parameter is optional", true, ""},
		{"credential missing", "credential missing in request", true, ""},
		{"password configured", "password configured via env", true, ""},
		{"token initialized", "token initialized properly", true, ""},
		{"secret loaded", "secret loaded from vault", true, ""},
		{"password cleared", "password cleared from memory", true, ""},
		{"token reset", "token reset after logout", true, ""},
		{"password created", "password created for new user", true, ""},
		{"token generated", "token generated successfully", true, ""},
		{"secret saved", "secret saved to keystore", true, ""},
		{"password invalid", "password invalid format", true, ""},

		// === Safe phrase position variations ===
		{"safe phrase before", "validated password field", true, ""},
		{"safe phrase after", "password was validated", true, ""},
		{"safe phrase far", "the password for this user has been successfully validated", true, ""},

		// === Чувствительное слово + unsafe context (должно failing) ===
		{"password with value:", "password: secret123", false, "password"},
		{"token with equals", "token=abc123", false, "token"},
		{"api_key with colon", "api_key: xyz", false, "api_key"},
		{"secret is leaked", "secret is exposed", false, "secret"},
		{"credential has value", "credential has sensitive data", false, "credential"},

		// === Пустые и граничные сообщения ===
		{"empty message", "", true, ""},
		{"whitespace only", "   ", true, ""},
		{"no sensitive words", "server started successfully", true, ""},
		{"numbers and punctuation", "error 404: not found", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := rules.NewSensitiveWordsRule()
			ctx := &rules.CheckContext{Msg: tt.msg}
			result := rule.Check(ctx)

			if result.Passed != tt.wantPassed {
				t.Errorf("Check(%q) passed = %v, want %v", tt.msg, result.Passed, tt.wantPassed)
			}
			if !tt.wantPassed && result.Message == "" {
				t.Error("Check() failed but message is empty")
			}
			// Проверяем, что в сообщении ошибки упоминается найденное слово
			if !tt.wantPassed && tt.wantWord != "" && result.Message != "" {
				if !containsIgnoreCase(result.Message, tt.wantWord) {
					t.Errorf("Check() message = %q, want to contain %q", result.Message, tt.wantWord)
				}
			}
		})
	}
}

// TestSensitiveWordsRule_Configure — тесты конфигурации правила
func TestSensitiveWordsRule_Configure(t *testing.T) {
	t.Run("add custom sensitive words", func(t *testing.T) {
		rule := rules.NewSensitiveWordsRule().(*rules.SensitiveWordsRule)
		config := map[string]any{
			"words": []string{"my_secret", "internal_token"},
		}
		if err := rule.Configure(config); err != nil {
			t.Fatalf("Configure() error = %v", err)
		}

		ctx := &rules.CheckContext{Msg: "my_secret value"}
		if result := rule.Check(ctx); result.Passed {
			t.Error("Check() should fail for custom sensitive word")
		}
	})

	t.Run("add custom safe phrases", func(t *testing.T) {
		rule := rules.NewSensitiveWordsRule().(*rules.SensitiveWordsRule)
		config := map[string]any{
			"safe_phrases": []string{"sanitized", "masked"},
		}
		if err := rule.Configure(config); err != nil {
			t.Fatalf("Configure() error = %v", err)
		}

		ctx := &rules.CheckContext{Msg: "password sanitized"}
		if result := rule.Check(ctx); !result.Passed {
			t.Errorf("Check() should pass with custom safe phrase, got: %v", result.Message)
		}
	})

	t.Run("custom words replace defaults", func(t *testing.T) {
		rule := rules.NewSensitiveWordsRule().(*rules.SensitiveWordsRule)
		config := map[string]any{
			"words": []string{"custom_only"},
		}
		if err := rule.Configure(config); err != nil {
			t.Fatalf("Configure() error = %v", err)
		}

		// Default word should NOT trigger anymore
		ctx := &rules.CheckContext{Msg: "password leaked"}
		if result := rule.Check(ctx); !result.Passed {
			t.Error("Check() should pass - default words replaced by custom")
		}

		// Custom word should trigger
		ctx = &rules.CheckContext{Msg: "custom_only here"}
		if result := rule.Check(ctx); result.Passed {
			t.Error("Check() should fail for custom word")
		}
	})

	t.Run("empty config does not break", func(t *testing.T) {
		rule := rules.NewSensitiveWordsRule()
		config := map[string]any{}
		if err := rule.Configure(config); err != nil {
			t.Errorf("Configure() with empty config should not error, got %v", err)
		}
	})

	t.Run("invalid config types ignored", func(t *testing.T) {
		rule := rules.NewSensitiveWordsRule().(*rules.SensitiveWordsRule)
		config := map[string]any{
			"words":        "not-a-slice", // wrong type
			"safe_phrases": 123,           // wrong type
		}
		if err := rule.Configure(config); err != nil {
			t.Errorf("Configure() should ignore invalid types, got %v", err)
		}
		// Defaults should still work
		ctx := &rules.CheckContext{Msg: "secret value"}
		if result := rule.Check(ctx); result.Passed {
			t.Error("Check() should still detect default sensitive words")
		}
	})
}

// TestSensitiveWordsRule_Enabled — тесты включения/выключения правила
func TestSensitiveWordsRule_Enabled(t *testing.T) {
	rule := rules.NewSensitiveWordsRule()

	t.Run("enabled by default", func(t *testing.T) {
		if !rule.Enabled() {
			t.Error("Rule should be enabled by default")
		}
	})

	t.Run("disabled rule always passes", func(t *testing.T) {
		rule.SetEnabled(false)
		ctx := &rules.CheckContext{Msg: "password: 123"}
		if result := rule.Check(ctx); !result.Passed {
			t.Error("Disabled rule should always pass")
		}
	})

	t.Run("re-enabled rule works again", func(t *testing.T) {
		rule.SetEnabled(true)
		ctx := &rules.CheckContext{Msg: "secret leaked"}
		if result := rule.Check(ctx); result.Passed {
			t.Error("Re-enabled rule should detect sensitive words")
		}
	})
}

// Граничные случаи и Unicode
func TestSensitiveWordsRule_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		msg        string
		wantPassed bool
	}{
		// === Word boundary edge cases ===
		{"word at start", "password is secret", false},
		{"word at end", "leaked password", false},
		{"word surrounded by punctuation", "error:password;failed", false},
		{"word with quotes", `"token" value`, false},
		{"word in parentheses", "(secret) data", false},

		// === Unicode and international ===
		{"cyrillic text no sensitive", "пароль не найден", true},
		{"mixed latin-cyrillic", "password: пароль", false},
		{"emoji with sensitive", "token 🔑 leaked", false},
		{"accented letters", "café secret", false},

		// === Multiple sensitive words ===
		{"two sensitive words", "password and token", false},
		{"first match returned", "secret then password", false},

		// === Case variations ===
		{"all caps", "PASSWORD TOKEN SECRET", false},
		{"camelCase", "myPassword field", true},
		{"snake_case", "user_password", true},

		// === Safe phrase with multiple words ===
		{"safe phrase complex", "the access_token has been successfully rotated", true},
		{"safe phrase negation still safe", "password not validated", true}, // "validated" present

		// === Dangerous patterns context (indirect via hasDangerousContext) ===
		// Note: hasDangerousContext is not used in Check() directly in current impl,
		// but we test that sensitive words with := or = still fail
		{"assignment colon", "password := secret", false},
		{"assignment equals", "token=abc123", false},
		{"is pattern", "secret is exposed", false},
		{"has pattern", "credential has value", false},
		{"value: pattern", "value: password123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := rules.NewSensitiveWordsRule()
			ctx := &rules.CheckContext{Msg: tt.msg}
			result := rule.Check(ctx)

			if result.Passed != tt.wantPassed {
				t.Errorf("Check(%q) passed = %v, want %v (message: %s)",
					tt.msg, result.Passed, tt.wantPassed, result.Message)
			}
		})
	}
}

// TestSensitiveWordsRule_NameAndDescription — мета-тесты правила
func TestSensitiveWordsRule_NameAndDescription(t *testing.T) {
	rule := rules.NewSensitiveWordsRule()

	if rule.Name() != rules.RuleSensitiveWordsName {
		t.Errorf("Name() = %q, want %q", rule.Name(), rules.RuleSensitiveWordsName)
	}

	desc := rule.Description()
	if desc == "" {
		t.Error("Description() should not be empty")
	}
	if !containsIgnoreCase(desc, "sensitive") {
		t.Errorf("Description() = %q, should mention sensitive data", desc)
	}
}

func indexIgnoreCase(s, substr string) int {
	sLower := strings.ToLower(s)
	subLower := strings.ToLower(substr)
	return strings.Index(sLower, subLower)
}
