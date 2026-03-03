package config

import (
	"fmt"
	"strings"
)

type Config struct {
	Rules RuleConfigs `yaml:"rules"`
}

type RuleConfigs struct {
	Lowercase      *RuleConfig           `yaml:"lowercase"`
	EnglishOnly    *RuleConfig           `yaml:"english_only"`
	NoSpecialChars *NoSpecialCharsConfig `yaml:"no_special_chars"`
	SensitiveWords *SensitiveWordsConfig `yaml:"sensitive_words"`
	CustomPatterns *CustomPatternsConfig `yaml:"custom_patterns"`
}

type RuleConfig struct {
	Enabled        *bool `yaml:"enabled"`
	AutoFixEnabled *bool `yaml:"auto_fix_enabled"`
}

type NoSpecialCharsConfig struct {
	RuleConfig         `yaml:",inline"`
	MaxConsecutiveDots *int `yaml:"max_consecutive_dots"`
}

type SensitiveWordsConfig struct {
	RuleConfig  `yaml:",inline"`
	Words       []string `yaml:"words"`
	SafePhrases []string `yaml:"safe_phrases"`
}

type SensitiveWords struct {
	Enabled     *bool    `yaml:"enabled"`
	Words       []string `yaml:"words"`
	SafePhrases []string `yaml:"safe_phrases"`
}

type CustomPatternsConfig struct {
	RuleConfig `yaml:",inline"`
	Patterns   []string `yaml:"patterns"`
}

type CustomPatterns struct {
	Enabled  *bool    `yaml:"enabled"`
	Patterns []string `yaml:"patterns"`
}

func (r *RuleConfig) IsEnabled() bool {
	if r == nil || r.Enabled == nil {
		return false
	}
	return *r.Enabled
}

func (s *SensitiveWords) IsEnabled() bool {
	if s == nil || s.Enabled == nil {
		return false
	}
	return *s.Enabled
}

func (c *CustomPatterns) IsEnabled() bool {
	if c == nil || c.Enabled == nil {
		return false
	}
	return *c.Enabled
}

func DefaultConfig() *Config {
	enabled := true
	autoFix := true
	defaultDots := 0

	return &Config{
		Rules: RuleConfigs{
			Lowercase: &RuleConfig{
				Enabled:        &enabled,
				AutoFixEnabled: &autoFix,
			},
			EnglishOnly: &RuleConfig{
				Enabled: &enabled,
			},
			NoSpecialChars: &NoSpecialCharsConfig{
				RuleConfig: RuleConfig{
					Enabled:        &enabled,
					AutoFixEnabled: &autoFix,
				},
				MaxConsecutiveDots: &defaultDots,
			},
			SensitiveWords: &SensitiveWordsConfig{
				RuleConfig: RuleConfig{
					Enabled:        &enabled,
					AutoFixEnabled: &autoFix,
				},
				Words:       make([]string, 0),
				SafePhrases: make([]string, 0),
			},
			CustomPatterns: &CustomPatternsConfig{
				RuleConfig: RuleConfig{
					Enabled: &enabled,
				},
			},
		},
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func (r *RuleConfig) IsAutoFixEnabled() bool {
	if r == nil || r.AutoFixEnabled == nil {
		return true
	}
	return *r.AutoFixEnabled
}

func (c *Config) Print() {
	if c == nil {
		fmt.Println("Config is nil")
		return
	}

	fmt.Println("========== Loaded Config ==========")

	printRule := func(name string, r *RuleConfig) {
		if r == nil {
			fmt.Printf("%s: <nil>\n", name)
			return
		}

		fmt.Printf("%s:\n", name)
		fmt.Printf("  enabled: %v\n", valueOrNilBool(r.Enabled))
		fmt.Printf("  auto_fix_enabled: %v\n", valueOrNilBool(r.AutoFixEnabled))
	}

	printRule("lowercase", c.Rules.Lowercase)
	printRule("english_only", c.Rules.EnglishOnly)

	// NoSpecialChars
	if n := c.Rules.NoSpecialChars; n != nil {
		fmt.Println("no_special_chars:")
		fmt.Printf("  enabled: %v\n", valueOrNilBool(n.Enabled))
		fmt.Printf("  auto_fix_enabled: %v\n", valueOrNilBool(n.AutoFixEnabled))
		fmt.Printf("  max_consecutive_dots: %v\n", valueOrNilInt(n.MaxConsecutiveDots))
	} else {
		fmt.Println("no_special_chars: <nil>")
	}

	// SensitiveWords
	if s := c.Rules.SensitiveWords; s != nil {
		fmt.Println("sensitive_words:")
		fmt.Printf("  enabled: %v\n", valueOrNilBool(s.Enabled))
		fmt.Printf("  auto_fix_enabled: %v\n", valueOrNilBool(s.AutoFixEnabled))
		fmt.Printf("  words: [%s]\n", strings.Join(s.Words, ", "))
		fmt.Printf("  safe_phrases: [%s]\n", strings.Join(s.SafePhrases, ", "))
	} else {
		fmt.Println("sensitive_words: <nil>")
	}

	// CustomPatterns
	if p := c.Rules.CustomPatterns; p != nil {
		fmt.Println("custom_patterns:")
		fmt.Printf("  enabled: %v\n", valueOrNilBool(p.Enabled))
		fmt.Printf("  auto_fix_enabled: %v\n", valueOrNilBool(p.AutoFixEnabled))
		fmt.Printf("  patterns: [%s]\n", strings.Join(p.Patterns, ", "))
	} else {
		fmt.Println("custom_patterns: <nil>")
	}

	fmt.Println("===================================")
}

func valueOrNilBool(b *bool) interface{} {
	if b == nil {
		return nil
	}
	return *b
}

func valueOrNilInt(i *int) interface{} {
	if i == nil {
		return nil
	}
	return *i
}
