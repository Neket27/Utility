package config

type Config struct {
	Rules RulesConfig `yaml:"rules"`
}

type RulesConfig struct {
	Lowercase      *RuleConfig     `yaml:"lowercase"`
	EnglishOnly    *RuleConfig     `yaml:"english_only"`
	NoSpecialChars *RuleConfig     `yaml:"no_special_chars"`
	SensitiveWords *SensitiveWords `yaml:"sensitive_words"`
	CustomPatterns *CustomPatterns `yaml:"custom_patterns"`
}

type RuleConfig struct {
	Enabled *bool `yaml:"enabled"`
}

type SensitiveWords struct {
	Enabled     *bool    `yaml:"enabled"`
	Words       []string `yaml:"words"`
	SafePhrases []string `yaml:"safe_phrases"`
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
	return &Config{
		Rules: RulesConfig{
			Lowercase:      &RuleConfig{Enabled: boolPtr(true)},
			EnglishOnly:    &RuleConfig{Enabled: boolPtr(true)},
			NoSpecialChars: &RuleConfig{Enabled: boolPtr(true)},
			SensitiveWords: &SensitiveWords{
				Enabled:     boolPtr(true),
				Words:       make([]string, 0),
				SafePhrases: make([]string, 0),
			},
			CustomPatterns: &CustomPatterns{
				Enabled:  boolPtr(true),
				Patterns: make([]string, 0),
			},
		},
	}
}

func boolPtr(b bool) *bool {
	return &b
}
