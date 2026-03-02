package test

import (
	"strings"
	"testing"
	"utility/pkg/analyzer/rules"
)

func TestCustomPatternsRule_Configure(t *testing.T) {
	tests := []struct {
		name         string
		config       map[string]any
		wantErr      bool
		wantPatterns int
		note         string
	}{
		// === Валидная конфигурация ===
		{"empty config", map[string]any{}, false, 0, "паттерны не обязательны"},
		{"nil config", nil, false, 0, ""},
		{"patterns as []string", map[string]any{"patterns": []string{`password`, `token`}}, false, 2, ""},
		{"patterns as []any", map[string]any{"patterns": []any{`secret`, `api_key`}}, false, 2, ""},
		{"single pattern", map[string]any{"patterns": []string{`\d{3}-\d{3}-\d{4}`}}, false, 1, "phone pattern"},
		{"complex regex", map[string]any{"patterns": []string{`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`}}, false, 1, "email pattern"},
		{"multiple complex", map[string]any{"patterns": []string{`\d+`, `[A-Z]+`, `\s+`}}, false, 3, ""},
		{"anchored pattern", map[string]any{"patterns": []string{`^ERROR`, `FAILED$`}}, false, 2, ""},
		{"case insensitive", map[string]any{"patterns": []string{`(?i)password`}}, false, 1, ""},
		{"word boundary", map[string]any{"patterns": []string{`\bsecret\b`}}, false, 1, ""},
		{"optional group", map[string]any{"patterns": []string{`https?://`}}, false, 1, ""},
		{"character class", map[string]any{"patterns": []string{`[aeiou]+`}}, false, 1, ""},
		{"quantifiers", map[string]any{"patterns": []string{`a{2,4}`, `b*`, `c+`}}, false, 3, ""},
		{"non-capturing group", map[string]any{"patterns": []string{`(?:foo|bar)`}}, false, 1, ""},

		// === Невалидная конфигурация ===
		{"invalid regex unclosed bracket", map[string]any{"patterns": []string{`[abc`}}, true, 0, "syntax error"},
		{"invalid regex unclosed paren", map[string]any{"patterns": []string{`(abc`}}, true, 0, ""},
		{"invalid regex invalid escape", map[string]any{"patterns": []string{`\xZZ`}}, true, 0, ""},
		{"invalid regex nested quantifier", map[string]any{"patterns": []string{`a**`}}, true, 0, ""},
		{"mixed valid invalid", map[string]any{"patterns": []string{`valid`, `[invalid`}}, true, 1, "first invalid fails all"},
		{"unicode pattern", map[string]any{"patterns": []string{`[\u0400-\u04FF]+`}}, true, 0, "cyrillic range"},
		{"lookahead", map[string]any{"patterns": []string{`foo(?=bar)`}}, true, 0, ""},
		{"lookbehind", map[string]any{"patterns": []string{`(?<=foo)bar`}}, true, 0, ""},

		// === Граничные типы данных ===
		{"patterns as string", map[string]any{"patterns": `not-a-slice`}, false, 0, "игнорируется"},
		{"patterns as int", map[string]any{"patterns": 123}, false, 0, ""},
		{"patterns as map", map[string]any{"patterns": map[string]any{}}, false, 0, ""},
		{"patterns as nil", map[string]any{"patterns": nil}, false, 0, ""},
		{"empty patterns slice", map[string]any{"patterns": []string{}}, false, 0, ""},
		{"empty patterns any", map[string]any{"patterns": []any{}}, false, 0, ""},
		{"mixed types in []any", map[string]any{"patterns": []any{`valid`, 123, `also_valid`, nil}}, false, 2, "строки извлекаются, остальное игнорируется"},
		{"patterns with spaces", map[string]any{"patterns": []string{`  spaced  `}}, false, 1, "пробелы в паттерне"},

		// === Другие поля конфигурации ===
		{"enabled field", map[string]any{"enabled": false}, false, 0, "BaseRule обрабатывает"},
		{"unknown field", map[string]any{"unknown": "value"}, false, 0, "игнорируется"},
		{"all fields together", map[string]any{"patterns": []string{`test`}, "enabled": true}, false, 1, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := rules.NewCustomPatternsRule().(*rules.CustomPatternsRule)
			err := rule.Configure(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("Configure() error = %v, wantErr = %v", err, tt.wantErr)
			}

			if len(rule.CompiledRegex) != tt.wantPatterns {
				t.Errorf("compiledRegex count = %d, want %d", len(rule.CompiledRegex), tt.wantPatterns)
			}

			if tt.wantErr && err != nil {
				if !containsIgnoreCase(err.Error(), "compile") {
					t.Errorf("Error message = %q, should mention compile error", err.Error())
				}
			}
		})
	}
}

func TestCustomPatternsRule_Check(t *testing.T) {
	tests := []struct {
		name         string
		patterns     []string
		msg          string
		enabled      bool
		wantPassed   bool
		wantRedacted string
		wantPattern  string
	}{
		// === Валидные сообщения (нет совпадений) ===
		{"no match simple", []string{`password`}, "server started", true, true, "", ""},
		{"no match case sensitive", []string{`Password`}, "password leaked", true, true, "", ""},
		{"no match partial", []string{`\bpass\b`}, "password", true, true, "", "word boundary"},
		{"no match anchored start", []string{`^error`}, "no error here", true, true, "", ""},
		{"no match anchored end", []string{`failed$`}, "failed earlier", true, true, "", ""},
		{"empty message", []string{`.*`}, "", true, true, "[REDACTED]", ""}, // .* матчит пустую строку
		{"whitespace only", []string{`\S+`}, "   ", true, true, "", "non-whitespace pattern"},

		// === Невалидные сообщения (есть совпадения) ===
		{"match simple", []string{`password`}, "user password: 123", true, false, "user [REDACTED]: 123", `password`},
		{"match case insensitive", []string{`(?i)password`}, "USER PASSWORD: 123", true, false, "USER [REDACTED]: 123", `(?i)password`},
		{"match digit pattern", []string{`\d{3}-\d{3}-\d{4}`}, "phone: 123-456-7890", true, false, "phone: [REDACTED]", `\d{3}-\d{3}-\d{4}`},
		{"match email", []string{`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`}, "email: test@example.com", true, false, "email: [REDACTED]", ""},
		{"match url", []string{`https?://\S+`}, "url: https://api.example.com/v1", true, false, "url: [REDACTED]", ""},
		{"match ip address", []string{`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`}, "ip: 192.168.1.1", true, false, "ip: [REDACTED]", ""},
		{"match credit card", []string{`\d{4}-\d{4}-\d{4}-\d{4}`}, "card: 1234-5678-9012-3456", true, false, "card: [REDACTED]", ""},
		{"match uuid", []string{`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`}, "id: 550e8400-e29b-41d4-a716-446655440000", true, false, "id: [REDACTED]", ""},
		{"match multiple occurrences", []string{`secret`}, "secret and secret again", true, false, "[REDACTED] and [REDACTED] again", ""},
		{"match at start", []string{`^ERROR`}, "ERROR: disk full", true, false, "[REDACTED]: disk full", ""},
		{"match at end", []string{`FAILED$`}, "operation FAILED", true, false, "operation [REDACTED]", ""},
		{"match in middle", []string{`token`}, "auth token expired", true, false, "auth [REDACTED] expired", ""},

		// === Multiple patterns ===
		{"multiple patterns first match", []string{`password`, `token`}, "password here", true, false, "[REDACTED] here", `password`},
		{"multiple patterns second match", []string{`password`, `token`}, "token here", true, false, "[REDACTED] here", `token`},

		// === Правило отключено ===
		{"disabled with match", []string{`password`}, "password leaked", false, true, "", ""},
		{"disabled no patterns", []string{}, "any message", false, true, "", ""},

		// === Нет паттернов ===
		{"no patterns configured", []string{}, "password", true, true, "", ""},
		{"nil patterns", nil, "password", true, true, "", ""},

		// === Unicode и специальные символы ===
		{"special chars in msg", []string{`\$`}, "price: $100", true, false, "price: [REDACTED]100", ""},
		{"newline in msg", []string{`\n`}, "line1\nline2", true, false, "line1[REDACTED]line2", ""},
		{"tab in msg", []string{`\t`}, "col1\tcol2", true, false, "col1[REDACTED]col2", ""},

		// === Граничные случаи regex ===
		{"capture group replaced", []string{`(foo)(bar)`}, "foobar", true, false, "[REDACTED]", ""},
		{"alternation", []string{`foo|bar|baz`}, "test bar test", true, false, "test [REDACTED] test", ""},
		{"character range", []string{`[0-9]+`}, "version 1.2.3", true, false, "version [REDACTED].[REDACTED].[REDACTED]", ""},
		{"negated class", []string{`[^a-z]+`}, "ABC123", true, false, "[REDACTED]", ""},
		{"optional match", []string{`colou?r`}, "color and colour", true, false, "[REDACTED] and [REDACTED]", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := rules.NewCustomPatternsRule().(*rules.CustomPatternsRule)
			rule.SetEnabled(tt.enabled)

			if len(tt.patterns) > 0 {
				config := map[string]any{"patterns": tt.patterns}
				if err := rule.Configure(config); err != nil {
					t.Fatalf("Configure() error = %v", err)
				}
			}

			ctx := &rules.CheckContext{Msg: tt.msg}
			result := rule.Check(ctx)

			if result.Passed != tt.wantPassed {
				t.Errorf("Check(%q) passed = %v, want %v", tt.msg, result.Passed, tt.wantPassed)
			}

			if !tt.wantPassed {
				if result.Message == "" {
					t.Error("Check() failed but message is empty")
				}
				if tt.wantPattern != "" && !containsIgnoreCase(result.Message, tt.wantPattern) {
					t.Errorf("Message = %q, should contain pattern %q", result.Message, tt.wantPattern)
				}
				if result.SuggestedFix == nil {
					t.Error("Expected SuggestedFix for failed check")
				} else {
					gotRedacted := extractSuggestedText(result.SuggestedFix)
					if gotRedacted != tt.wantRedacted {
						t.Errorf("SuggestedFix.NewText = %q, want %q", gotRedacted, tt.wantRedacted)
					}
				}
			} else {
				if result.SuggestedFix != nil {
					t.Errorf("Expected nil SuggestedFix for passed check, got %v", result.SuggestedFix)
				}
			}
		})
	}
}

func TestCustomPatternsRule_Meta(t *testing.T) {
	rule := rules.NewCustomPatternsRule()

	t.Run("Name", func(t *testing.T) {
		if got := rule.Name(); got != rules.RuleCustomPatternsName {
			t.Errorf("Name() = %q, want %q", got, rules.RuleCustomPatternsName)
		}
	})

	t.Run("Description", func(t *testing.T) {
		desc := rule.Description()
		if desc == "" {
			t.Error("Description() should not be empty")
		}
		if !containsIgnoreCase(desc, "custom") || !containsIgnoreCase(desc, "pattern") {
			t.Errorf("Description() = %q, should mention custom patterns", desc)
		}
	})

	t.Run("Enabled by default", func(t *testing.T) {
		if !rule.Enabled() {
			t.Error("Rule should be enabled by default")
		}
	})

	t.Run("Empty patterns by default", func(t *testing.T) {
		customRule := rule.(*rules.CustomPatternsRule)
		if len(customRule.CompiledRegex) != 0 {
			t.Errorf("compiledRegex should be empty by default, got %d", len(customRule.CompiledRegex))
		}
	})
}

func TestCustomPatternsRule_Enabled(t *testing.T) {
	rule := rules.NewCustomPatternsRule().(*rules.CustomPatternsRule)

	t.Run("enable disable toggle", func(t *testing.T) {
		config := map[string]any{"patterns": []string{`test`}}
		if err := rule.Configure(config); err != nil {
			t.Fatalf("Configure() error = %v", err)
		}

		rule.SetEnabled(true)
		ctx := &rules.CheckContext{Msg: "test message"}
		if result := rule.Check(ctx); result.Passed {
			t.Error("Enabled rule should fail on matching pattern")
		}

		rule.SetEnabled(false)
		if result := rule.Check(ctx); !result.Passed {
			t.Error("Disabled rule should pass regardless of pattern")
		}

		rule.SetEnabled(true)
		if result := rule.Check(ctx); result.Passed {
			t.Error("Re-enabled rule should fail again")
		}
	})

	t.Run("configure with enabled field", func(t *testing.T) {
		config := map[string]any{
			"patterns": []string{`secret`},
			"enabled":  false,
		}
		if err := rule.Configure(config); err != nil {
			t.Fatalf("Configure() error = %v", err)
		}

		if rule.Enabled() {
			t.Error("Rule should be disabled after configure with enabled=false")
		}
	})
}

func TestCustomPatternsRule_ErrorMessages(t *testing.T) {
	tests := []struct {
		name      string
		pattern   string
		msg       string
		wantInMsg string
	}{
		{"simple pattern", `password`, "password leaked", "password"},
		{"complex pattern", `\d{3}-\d{3}-\d{4}`, "123-456-7890", `\d{3}-\d{3}-\d{4}`},
		{"case insensitive", `(?i)SECRET`, "secret value", `(?i)SECRET`},
		{"anchored", `^ERROR`, "ERROR occurred", `^ERROR`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := rules.NewCustomPatternsRule().(*rules.CustomPatternsRule)
			config := map[string]any{"patterns": []string{tt.pattern}}
			if err := rule.Configure(config); err != nil {
				t.Fatalf("Configure() error = %v", err)
			}

			ctx := &rules.CheckContext{Msg: tt.msg}
			result := rule.Check(ctx)

			if result.Passed {
				t.Fatal("Check() should fail")
			}

			if !containsIgnoreCase(result.Message, "forbidden pattern") {
				t.Errorf("Message = %q, should mention 'forbidden pattern'", result.Message)
			}

			if !containsIgnoreCase(result.Message, tt.wantInMsg) {
				t.Errorf("Message = %q, should contain pattern %q", result.Message, tt.wantInMsg)
			}
		})
	}
}

func TestCustomPatternsRule_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		msg      string
	}{
		{"very long message", []string{`secret`}, strings.Repeat("a", 10000) + "secret" + strings.Repeat("b", 10000)},
		{"very long pattern match", []string{`a+`}, strings.Repeat("a", 10000)},
		{"many patterns", []string{`a`, `b`, `c`, `d`, `e`, `f`, `g`, `h`, `i`, `j`}, "abcdefghij"},
		{"overlapping patterns", []string{`test`, `testing`, `tester`}, "testing"},
		{"empty pattern", []string{``}, "any message"}, // пустой паттерн матчит всё
		{"only anchors", []string{`^$`}, ""},           // матчит пустую строку
		{"named group", []string{`(?P<word>\w+)`}, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := rules.NewCustomPatternsRule().(*rules.CustomPatternsRule)
			if len(tt.patterns) > 0 {
				config := map[string]any{"patterns": tt.patterns}

				if err := rule.Configure(config); err != nil {
					if tt.name == "empty pattern" {
						return
					}
					t.Fatalf("Configure() error = %v", err)
				}
			}

			ctx := &rules.CheckContext{Msg: tt.msg}
			result := rule.Check(ctx)

			if result == nil {
				t.Fatal("Check() returned nil")
			}

			// Просто проверяем, что не паникует
			_ = result.Passed
		})
	}
}

func TestCustomPatternsRule_Reconfigure(t *testing.T) {
	rule := rules.NewCustomPatternsRule().(*rules.CustomPatternsRule)

	t.Run("replace patterns", func(t *testing.T) {
		// First config
		config1 := map[string]any{"patterns": []string{`password`}}
		if err := rule.Configure(config1); err != nil {
			t.Fatalf("Configure() error = %v", err)
		}

		ctx := &rules.CheckContext{Msg: "password leaked"}
		if result := rule.Check(ctx); result.Passed {
			t.Error("Should match password")
		}

		// Second config - replace patterns
		config2 := map[string]any{"patterns": []string{`token`}}
		if err := rule.Configure(config2); err != nil {
			t.Fatalf("Configure() error = %v", err)
		}

		// Password should not match anymore
		if result := rule.Check(ctx); !result.Passed {
			t.Error("Password should not match after reconfigure")
		}

		// Token should match
		ctx = &rules.CheckContext{Msg: "token leaked"}
		if result := rule.Check(ctx); result.Passed {
			t.Error("Token should match")
		}
	})

	t.Run("add more patterns", func(t *testing.T) {
		config := map[string]any{"patterns": []string{`secret`, `password`, `token`}}
		if err := rule.Configure(config); err != nil {
			t.Fatalf("Configure() error = %v", err)
		}

		if len(rule.CompiledRegex) != 3 {
			t.Errorf("Expected 3 patterns, got %d", len(rule.CompiledRegex))
		}
	})

	t.Run("clear patterns", func(t *testing.T) {
		config1 := map[string]any{"patterns": []string{`test`}}
		rule.Configure(config1)

		config2 := map[string]any{"patterns": []string{}}
		rule.Configure(config2)

		if len(rule.CompiledRegex) != 0 {
			t.Errorf("Expected 0 patterns after clear, got %d", len(rule.CompiledRegex))
		}
	})
}

func TestCustomPatternsRule_SuggestedFix(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		msg     string
		wantFix string
		wantMsg string
	}{
		{"basic redaction", `password`, "user password: 123", "user [REDACTED]: 123", "forbidden pattern"},
		{"multiple redactions", `secret`, "secret and secret", "[REDACTED] and [REDACTED]", ""},
		{"partial redaction", `\d+`, "version 1.2.3", "version [REDACTED].[REDACTED].[REDACTED]", ""},
		{"full redaction", `.*`, "everything", "[REDACTED]", ""},
		{"no change needed", `nomatch`, "test message", "", ""}, // не должно быть suggestion
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := rules.NewCustomPatternsRule().(*rules.CustomPatternsRule)
			config := map[string]any{"patterns": []string{tt.pattern}}
			if err := rule.Configure(config); err != nil {
				t.Fatalf("Configure() error = %v", err)
			}

			ctx := &rules.CheckContext{Msg: tt.msg}
			result := rule.Check(ctx)

			if tt.wantFix != "" {
				if result.SuggestedFix == nil {
					t.Fatal("Expected SuggestedFix")
				}
				gotFix := extractSuggestedText(result.SuggestedFix)
				if gotFix != tt.wantFix {
					t.Errorf("SuggestedFix.NewText = %q, want %q", gotFix, tt.wantFix)
				}
			} else {
				if result.SuggestedFix != nil {
					t.Errorf("Expected nil SuggestedFix, got %v", result.SuggestedFix)
				}
			}

			if tt.wantMsg != "" && !containsIgnoreCase(result.Message, tt.wantMsg) {
				t.Errorf("Message = %q, should contain %q", result.Message, tt.wantMsg)
			}
		})
	}
}

func TestCustomPatternsRule_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	rule := rules.NewCustomPatternsRule().(*rules.CustomPatternsRule)
	config := map[string]any{
		"patterns": []string{
			`\d+`,
			`[A-Za-z]+`,
			`\s+`,
			`[^\w\s]+`,
		},
	}
	if err := rule.Configure(config); err != nil {
		t.Fatalf("Configure() error = %v", err)
	}

	longMsg := strings.Repeat("test123 message ", 1000)

	t.Run("many patterns long message", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			ctx := &rules.CheckContext{Msg: longMsg}
			_ = rule.Check(ctx)
		}
	})
}
