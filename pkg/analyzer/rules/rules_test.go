package rules

import (
	"regexp"
	"testing"
)

func TestCheckLowercase(t *testing.T) {
	tests := []struct {
		name      string
		msg       string
		wantValid bool
		wantFix   string
	}{
		{"valid lowercase", "starting server", true, ""},
		{"invalid uppercase", "Starting server", false, "starting server"},
		{"invalid uppercase long", "Database connection failed", false, "database connection failed"},
		{"empty", "", true, ""},
		{"invalid starts with number", "123 items processed", false, ""},
		{"invalid starts with bracket", "(Starting) server", false, ""},
		{"invalid starts with quote", "\"Starting\" server", false, ""},
		{"valid lowercase after space", " server starting", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, fix := CheckLowercase(tt.msg)
			if valid != tt.wantValid {
				t.Errorf("CheckLowercase(%q) valid = %v, want %v", tt.msg, valid, tt.wantValid)
			}
			if fix != tt.wantFix {
				t.Errorf("CheckLowercase(%q) fix = %q, want %q", tt.msg, fix, tt.wantFix)
			}
		})
	}
}

func TestCheckEnglishOnly(t *testing.T) {
	tests := []struct {
		name      string
		msg       string
		wantValid bool
	}{
		{"valid english", "starting server", true},
		{"valid with numbers", "user 123 logged in", true},
		{"valid with punctuation", "server started on port 8080", true},
		{"valid with emoji", "server started 👋", true},
		{"invalid cyrillic", "запуск сервера", false},
		{"invalid chinese", "启动服务器", false},
		{"invalid arabic", "بدء التشغيل", false},
		{"mixed english and cyrillic", "starting запуск server", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := CheckEnglishOnly(tt.msg)
			if valid != tt.wantValid {
				t.Errorf("CheckEnglishOnly(%q) valid = %v, want %v", tt.msg, valid, tt.wantValid)
			}
		})
	}
}

func TestCheckNoSpecialChars(t *testing.T) {
	tests := []struct {
		name      string
		msg       string
		wantValid bool
	}{
		{"valid simple", "server started", true},
		{"valid with period", "server started.", true},
		{"valid with comma", "server started, listening", true},
		{"valid with colon", "port: 8080", true},
		{"invalid exclamation", "server started!", false},
		{"invalid double exclamation", "connection failed!!", false},
		{"invalid ellipsis", "something went wrong...", false},
		{"invalid at symbol", "user @localhost", false},
		{"invalid hash", "error #404", false},
		{"invalid emoji", "server started 😀", false},
		{"invalid multiple special", "error!!!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, _ := CheckNoSpecialChars(tt.msg)
			if valid != tt.wantValid {
				t.Errorf("CheckNoSpecialChars(%q) valid = %v, want %v", tt.msg, valid, tt.wantValid)
			}
		})
	}
}

func TestSensitiveWordsRule(t *testing.T) {
	rule := NewSensitiveWordsRule().(*SensitiveWordsRule)

	t.Run("default words loaded", func(t *testing.T) {
		if len(rule.words) == 0 {
			t.Error("Expected default words to be loaded")
		}
	})

	t.Run("check with sensitive word in message", func(t *testing.T) {
		rule.words = []*regexp.Regexp{
			regexp.MustCompile(`(?i)\bpassword\b`),
		}
		ctx := &CheckContext{
			Msg: "user password: 123", // ✅ Проверяем содержимое строки
		}
		result := rule.Check(ctx)
		if result.Passed {
			t.Error("Check() should fail for sensitive word in message")
		}
	})

	t.Run("check with sensitive word in binary op message", func(t *testing.T) {
		rule.words = []*regexp.Regexp{
			regexp.MustCompile(`(?i)\bapi_key\b`),
		}
		ctx := &CheckContext{
			Msg: "api_key=secret", // ✅ Содержимое строки после конкатенации
		}
		result := rule.Check(ctx)
		if result.Passed {
			t.Error("Check() should fail for sensitive word in message")
		}
	})

	t.Run("check with sensitive word from field message", func(t *testing.T) {
		rule.words = []*regexp.Regexp{
			regexp.MustCompile(`(?i)\btoken\b`),
		}
		ctx := &CheckContext{
			Msg: "token: abc123", // ✅ Содержимое строки
		}
		result := rule.Check(ctx)
		if result.Passed {
			t.Error("Check() should fail for sensitive word in message")
		}
	})

	t.Run("configure with custom words appends to defaults", func(t *testing.T) {
		config := map[string]any{
			"words": []string{"custom_secret", "my_token"},
		}

		err := rule.Configure(config)
		if err != nil {
			t.Fatalf("Configure() error = %v", err)
		}

		if len(rule.words) != 2 {
			t.Errorf("expected %d patterns, got %d", 2, len(rule.words))
		}
	})

	t.Run("non-sensitive message passes", func(t *testing.T) {
		ctx := &CheckContext{
			Msg: "user authenticated successfully",
		}
		result := rule.Check(ctx)
		if !result.Passed {
			t.Error("Check() should pass for non-sensitive message")
		}
	})

	t.Run("word boundary check - authenticated passes", func(t *testing.T) {
		ctx := &CheckContext{
			Msg: "user authenticated successfully", // "auth" внутри слова
		}
		result := rule.Check(ctx)
		if !result.Passed {
			t.Error("Check() should pass - 'auth' inside 'authenticated' is OK")
		}
	})

	t.Run("word boundary check - authorization passes", func(t *testing.T) {
		ctx := &CheckContext{
			Msg: "authorization header set",
		}
		result := rule.Check(ctx)
		if !result.Passed {
			t.Error("Check() should pass - 'auth' inside 'authorization' is OK")
		}
	})

	t.Run("word boundary check - passwords policy passes", func(t *testing.T) {
		ctx := &CheckContext{
			Msg: "passwords policy updated", // "password" внутри слова
		}
		result := rule.Check(ctx)
		if !result.Passed {
			t.Error("Check() should pass - 'password' inside 'passwords' is OK")
		}
	})

	t.Run("case insensitive check", func(t *testing.T) {
		rule.words = []*regexp.Regexp{
			regexp.MustCompile(`(?i)\bpassword\b`),
		}
		ctx := &CheckContext{
			Msg: "user PASSWORD: 123", // Верхний регистр
		}
		result := rule.Check(ctx)
		if result.Passed {
			t.Error("Check() should fail for sensitive word (case insensitive)")
		}
	})

	t.Run("empty message passes", func(t *testing.T) {
		ctx := &CheckContext{
			Msg: "",
		}
		result := rule.Check(ctx)
		if !result.Passed {
			t.Error("Check() should pass for empty message")
		}
	})

	t.Run("secret word detection", func(t *testing.T) {
		ctx := &CheckContext{
			Msg: "secret value here",
		}
		result := rule.Check(ctx)
		if !result.Passed {
			t.Error("Check() should fail for 'secret' word")
		}
	})

	t.Run("credential word detection", func(t *testing.T) {
		ctx := &CheckContext{
			Msg: "user credential leaked",
		}
		result := rule.Check(ctx)
		if !result.Passed {
			t.Error("Check() should fail for 'credential' word")
		}
	})

	t.Run("access_token detection", func(t *testing.T) {
		ctx := &CheckContext{
			Msg: "access_token expired",
		}
		result := rule.Check(ctx)
		if !result.Passed {
			t.Error("Check() should fail for 'access_token' word")
		}
	})

	t.Run("private_key detection", func(t *testing.T) {
		ctx := &CheckContext{
			Msg: "private_key loaded",
		}
		result := rule.Check(ctx)
		if !result.Passed {
			t.Error("Check() should fail for 'private_key' word")
		}
	})
}

func TestBaseRule(t *testing.T) {
	rule := NewBaseRule("test_rule", "Test description")

	t.Run("Name", func(t *testing.T) {
		if rule.Name() != "test_rule" {
			t.Errorf("Name() = %q, want %q", rule.Name(), "test_rule")
		}
	})

	t.Run("Description", func(t *testing.T) {
		if rule.Description() != "Test description" {
			t.Errorf("Description() = %q, want %q", rule.Description(), "Test description")
		}
	})

	t.Run("Enabled", func(t *testing.T) {
		if !rule.Enabled() {
			t.Error("Enabled() returned false")
		}
	})

	t.Run("SetEnabled", func(t *testing.T) {
		rule.SetEnabled(false)
		if rule.Enabled() {
			t.Error("SetEnabled() did not disable rule")
		}
	})
}

func TestResultHelpers(t *testing.T) {
	t.Run("ResultPass", func(t *testing.T) {
		result := ResultPass()
		if !result.Passed {
			t.Error("ResultPass() returned failing result")
		}
	})

	t.Run("ResultFail", func(t *testing.T) {
		result := ResultFail("test message")
		if result.Passed {
			t.Error("ResultFail() returned passing result")
		}
	})

	t.Run("ResultFailWithSuggestion", func(t *testing.T) {
		result := ResultFailWithSuggestion("msg", "fix msg", "new text")
		if result.Passed {
			t.Error("ResultFailWithSuggestion() returned passing result")
		}
		if result.SuggestedFix == nil {
			t.Error("ResultFailWithSuggestion() returned nil SuggestedFix")
		}
	})
}
