package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

// Config - основная структура конфигурации
type Config struct {
	Rules RulesConfig `yaml:"rules"`
}

// RulesConfig - конфигурация правил
type RulesConfig struct {
	Lowercase      *RuleConfig     `yaml:"lowercase"`
	EnglishOnly    *RuleConfig     `yaml:"english_only"`
	NoSpecialChars *RuleConfig     `yaml:"no_special_chars"`
	SensitiveWords *SensitiveWords `yaml:"sensitive_words"`
	CustomPatterns *CustomPatterns `yaml:"custom_patterns"`
}

// RuleConfig - базовая конфигурация правила
type RuleConfig struct {
	Enabled *bool `yaml:"enabled"`
}

// SensitiveWords - конфигурация правила чувствительных слов
type SensitiveWords struct {
	Enabled *bool    `yaml:"enabled"`
	Words   []string `yaml:"words"`
}

// CustomPatterns - конфигурация кастомных паттернов
type CustomPatterns struct {
	Enabled  *bool    `yaml:"enabled"`
	Patterns []string `yaml:"patterns"`
}

// IsEnabled - проверка включено ли правило
func (r *RuleConfig) IsEnabled() bool {
	if r == nil || r.Enabled == nil {
		return false // ✅ По умолчанию ВЫКЛЮЧЕНО
	}
	return *r.Enabled
}

// IsEnabled для SensitiveWords
func (s *SensitiveWords) IsEnabled() bool {
	if s == nil || s.Enabled == nil {
		return false
	}
	return *s.Enabled
}

// IsEnabled для CustomPatterns
func (c *CustomPatterns) IsEnabled() bool {
	if c == nil || c.Enabled == nil {
		return false
	}
	return *c.Enabled
}

// DefaultConfig - конфигурация по умолчанию
func DefaultConfig() *Config {
	return &Config{
		Rules: RulesConfig{
			Lowercase:      &RuleConfig{Enabled: boolPtr(true)},
			EnglishOnly:    &RuleConfig{Enabled: boolPtr(true)},
			NoSpecialChars: &RuleConfig{Enabled: boolPtr(true)},
			SensitiveWords: &SensitiveWords{Enabled: boolPtr(true)},
			CustomPatterns: &CustomPatterns{Enabled: boolPtr(true)},
		},
	}
}

func boolPtr(b bool) *bool {
	return &b
}

// Loader - загрузчик конфигурации
type Loader struct {
	path string
}

func NewLoader(path string) *Loader {
	return &Loader{path: path}
}

func (l *Loader) Load() (*Config, error) {
	if l.path == "" {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(l.path)
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
