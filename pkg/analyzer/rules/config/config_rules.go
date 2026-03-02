package config

// Config — корневая структура конфигурации линтера
type Config struct {
	Rules RulesConfig `json:"rules" yaml:"rules"`
}

// RulesConfig — настройки для каждого правила
type RulesConfig struct {
	Lowercase      *RuleConfig          `json:"lowercase,omitempty" yaml:"lowercase,omitempty"`
	EnglishOnly    *RuleConfig          `json:"english_only,omitempty" yaml:"english_only,omitempty"`
	NoSpecialChars *RuleConfig          `json:"no_special_chars,omitempty" yaml:"no_special_chars,omitempty"`
	SensitiveWords *SensitiveRuleConfig `json:"sensitive_words,omitempty" yaml:"sensitive_words,omitempty"`
	CustomPatterns *CustomRuleConfig    `json:"custom_patterns,omitempty" yaml:"custom_patterns,omitempty"`
}

// RuleConfig — базовая конфигурация правила
type RuleConfig struct {
	Enabled *bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
}

// SensitiveRuleConfig — конфигурация правила чувствительных данных
type SensitiveRuleConfig struct {
	RuleConfig
	Words []string `json:"words,omitempty" yaml:"words,omitempty"`
}

// CustomRuleConfig — конфигурация кастомных паттернов
type CustomRuleConfig struct {
	RuleConfig
	Patterns []string `json:"patterns,omitempty" yaml:"patterns,omitempty"`
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	enabled := true
	return &Config{
		Rules: RulesConfig{
			Lowercase:      &RuleConfig{Enabled: &enabled},
			EnglishOnly:    &RuleConfig{Enabled: &enabled},
			NoSpecialChars: &RuleConfig{Enabled: &enabled},
			SensitiveWords: &SensitiveRuleConfig{
				RuleConfig: RuleConfig{Enabled: &enabled},
				Words:      DefaultSensitiveWords(),
			},
			CustomPatterns: &CustomRuleConfig{
				RuleConfig: RuleConfig{Enabled: &enabled},
				Patterns:   []string{},
			},
		},
	}
}

// DefaultSensitiveWords возвращает слова по умолчанию
func DefaultSensitiveWords() []string {
	return []string{
		"password", "passwd", "pwd",
		"token", "api_key", "apikey",
		"secret", "credential",
		"private_key", "access_token",
	}
}

// IsEnabled проверяет, включено ли правило
func (r *RuleConfig) IsEnabled() bool {
	if r == nil || r.Enabled == nil {
		return true // По умолчанию включено
	}
	return *r.Enabled
}
