package config

import (
	"testing"
	"utility/pkg/analyzer/rules"
)

func TestNoSpecialCharsRule_Configure_MaxDots(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]any
		wantMaxDots int
		wantAutoFix bool
		wantEnabled bool
	}{
		{
			name:        "default values",
			config:      map[string]any{},
			wantMaxDots: 0,
			wantAutoFix: true,
			wantEnabled: true,
		},
		{
			name:        "configure max_consecutive_dots",
			config:      map[string]any{"max_consecutive_dots": 2},
			wantMaxDots: 2,
			wantAutoFix: true,
			wantEnabled: true,
		},
		{
			name: "configure all flags",
			config: map[string]any{
				"enabled":              false,
				"auto_fix_enabled":     false,
				"max_consecutive_dots": 3,
			},
			wantMaxDots: 3,
			wantAutoFix: false,
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := rules.NewNoSpecialCharsRule().(*rules.NoSpecialCharsRule)

			if err := rule.Configure(tt.config); err != nil {
				t.Fatalf("Configure() error = %v", err)
			}

			if rule.MaxConsecutiveDots != tt.wantMaxDots {
				t.Errorf("maxConsecutiveDots = %d, want %d", rule.MaxConsecutiveDots, tt.wantMaxDots)
			}
			if rule.AutoFixEnabled() != tt.wantAutoFix {
				t.Errorf("AutoFixEnabled = %v, want %v", rule.AutoFixEnabled(), tt.wantAutoFix)
			}
			if rule.Enabled() != tt.wantEnabled {
				t.Errorf("Enabled = %v, want %v", rule.Enabled(), tt.wantEnabled)
			}
		})
	}
}
