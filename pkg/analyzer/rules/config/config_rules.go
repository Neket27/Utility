package config

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

func (r *RuleConfig) IsAutoFixEnabled() bool {
	return *r.AutoFixEnabled
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
				Patterns: make([]string, 0),
			},
		},
	}
}

func boolPtr(b bool) *bool {
	return &b
}
