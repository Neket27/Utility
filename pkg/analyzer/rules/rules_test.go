package rules

import (
	"testing"
)

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
