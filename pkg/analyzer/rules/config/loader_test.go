package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoader_Load(t *testing.T) {
	t.Run("empty path returns default config", func(t *testing.T) {
		loader := NewLoader("")
		cfg, err := loader.Load()

		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg == nil {
			t.Fatal("Load() returned nil config")
		}

		if !cfg.Rules.Lowercase.IsEnabled() {
			t.Error("Lowercase should be enabled by default")
		}
		if !cfg.Rules.EnglishOnly.IsEnabled() {
			t.Error("EnglishOnly should be enabled by default")
		}
		if !cfg.Rules.NoSpecialChars.IsEnabled() {
			t.Error("NoSpecialChars should be enabled by default")
		}
		if !cfg.Rules.SensitiveWords.IsEnabled() {
			t.Error("SensitiveWords should be enabled by default")
		}
		if !cfg.Rules.CustomPatterns.IsEnabled() {
			t.Error("CustomPatterns should be enabled by default")
		}
	})

	t.Run("valid yaml file loads successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		yamlContent := `
rules:
  lowercase:
    enabled: true
  english_only:
    enabled: false
  no_special_chars:
    enabled: true
`
		if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		loader := NewLoader(configPath)
		cfg, err := loader.Load()

		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg == nil {
			t.Fatal("Load() returned nil config")
		}

		if !cfg.Rules.Lowercase.IsEnabled() {
			t.Error("Lowercase should be enabled")
		}
		if cfg.Rules.EnglishOnly.IsEnabled() {
			t.Error("EnglishOnly should be disabled")
		}
		if !cfg.Rules.NoSpecialChars.IsEnabled() {
			t.Error("NoSpecialChars should be enabled")
		}
	})

	t.Run("nonexistent file returns error", func(t *testing.T) {
		loader := NewLoader("/nonexistent/path/config.yaml")
		_, err := loader.Load()

		if err == nil {
			t.Error("Load() should return error for nonexistent file")
		}
		if !os.IsNotExist(err) {
			t.Errorf("Expected os.IsNotExist error, got %v", err)
		}
	})

	t.Run("invalid yaml returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "invalid.yaml")

		invalidYaml := `
rules:
  lowercase:
    enabled: true
  english_only: [invalid yaml structure
`
		if err := os.WriteFile(configPath, []byte(invalidYaml), 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		loader := NewLoader(configPath)
		_, err := loader.Load()

		if err == nil {
			t.Error("Load() should return error for invalid YAML")
		}
	})

	t.Run("partial config loads with defaults for missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "partial.yaml")

		yamlContent := `
rules:
  lowercase:
    enabled: false
`
		if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		loader := NewLoader(configPath)
		cfg, err := loader.Load()

		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if cfg.Rules.Lowercase.IsEnabled() {
			t.Error("Lowercase should be disabled (explicitly set)")
		}

		if cfg.Rules.EnglishOnly != nil {
			t.Error("EnglishOnly should be nil when not specified")
		}
		if cfg.Rules.NoSpecialChars != nil {
			t.Error("NoSpecialChars should be nil when not specified")
		}
	})

	t.Run("empty yaml file loads with nil rules", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "empty.yaml")

		if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		loader := NewLoader(configPath)
		cfg, err := loader.Load()

		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg == nil {
			t.Fatal("Load() returned nil config")
		}

		if cfg.Rules.Lowercase != nil {
			t.Error("Lowercase should be nil for empty YAML")
		}
	})
}

func TestRuleConfig_IsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		config   *RuleConfig
		wantBool bool
	}{
		{"nil config", nil, false},
		{"enabled nil", &RuleConfig{Enabled: nil}, false},
		{"enabled true", &RuleConfig{Enabled: boolPtr(true)}, true},
		{"enabled false", &RuleConfig{Enabled: boolPtr(false)}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.IsEnabled()
			if got != tt.wantBool {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.wantBool)
			}
		})
	}
}

func TestSensitiveWords_IsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		config   *SensitiveWords
		wantBool bool
	}{
		{"nil config", nil, false},
		{"enabled nil", &SensitiveWords{Enabled: nil}, false},
		{"enabled true", &SensitiveWords{Enabled: boolPtr(true)}, true},
		{"enabled false", &SensitiveWords{Enabled: boolPtr(false)}, false},
		{"with words but enabled nil", &SensitiveWords{Words: []string{"password"}}, false},
		{"with words and enabled true", &SensitiveWords{Enabled: boolPtr(true), Words: []string{"password"}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.IsEnabled()
			if got != tt.wantBool {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.wantBool)
			}
		})
	}
}

func TestCustomPatterns_IsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		config   *CustomPatterns
		wantBool bool
	}{
		{"nil config", nil, false},
		{"enabled nil", &CustomPatterns{Enabled: nil}, false},
		{"enabled true", &CustomPatterns{Enabled: boolPtr(true)}, true},
		{"enabled false", &CustomPatterns{Enabled: boolPtr(false)}, false},
		{"with patterns but enabled nil", &CustomPatterns{Patterns: []string{`\d+`}}, false},
		{"with patterns and enabled true", &CustomPatterns{Enabled: boolPtr(true), Patterns: []string{`\d+`}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.IsEnabled()
			if got != tt.wantBool {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.wantBool)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	t.Run("all rules present", func(t *testing.T) {
		if cfg.Rules.Lowercase == nil {
			t.Error("Lowercase should not be nil")
		}
		if cfg.Rules.EnglishOnly == nil {
			t.Error("EnglishOnly should not be nil")
		}
		if cfg.Rules.NoSpecialChars == nil {
			t.Error("NoSpecialChars should not be nil")
		}
		if cfg.Rules.SensitiveWords == nil {
			t.Error("SensitiveWords should not be nil")
		}
		if cfg.Rules.CustomPatterns == nil {
			t.Error("CustomPatterns should not be nil")
		}
	})

	t.Run("all rules enabled by default", func(t *testing.T) {
		if !cfg.Rules.Lowercase.IsEnabled() {
			t.Error("Lowercase should be enabled by default")
		}
		if !cfg.Rules.EnglishOnly.IsEnabled() {
			t.Error("EnglishOnly should be enabled by default")
		}
		if !cfg.Rules.NoSpecialChars.IsEnabled() {
			t.Error("NoSpecialChars should be enabled by default")
		}
		if !cfg.Rules.SensitiveWords.IsEnabled() {
			t.Error("SensitiveWords should be enabled by default")
		}
		if !cfg.Rules.CustomPatterns.IsEnabled() {
			t.Error("CustomPatterns should be enabled by default")
		}
	})

	t.Run("sensitive words has default empty slices", func(t *testing.T) {
		sw := cfg.Rules.SensitiveWords
		if sw.Words == nil {
			t.Error("SensitiveWords.Words should not be nil (should be empty slice)")
		}
		if sw.SafePhrases == nil {
			t.Error("SensitiveWords.SafePhrases should not be nil (should be empty slice)")
		}
	})

	t.Run("custom patterns has default empty slice", func(t *testing.T) {
		cp := cfg.Rules.CustomPatterns
		if cp.Patterns == nil {
			t.Error("CustomPatterns.Patterns should not be nil (should be empty slice)")
		}
	})
}

func TestConfig_YAMLParsing(t *testing.T) {
	t.Run("full config with all options", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "full.yaml")

		yamlContent := `
rules:
  lowercase:
    enabled: true
  english_only:
    enabled: false
  no_special_chars:
    enabled: true
  sensitive_words:
    enabled: true
    words:
      - password
      - token
      - secret
    safe_phrases:
      - validated
      - expired
      - rotated
  custom_patterns:
    enabled: true
    patterns:
      - "\\d{3}-\\d{3}-\\d{4}"
      - "[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}"
`
		if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		loader := NewLoader(configPath)
		cfg, err := loader.Load()

		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if !cfg.Rules.Lowercase.IsEnabled() {
			t.Error("Lowercase should be enabled")
		}
		if cfg.Rules.EnglishOnly.IsEnabled() {
			t.Error("EnglishOnly should be disabled")
		}
		if !cfg.Rules.NoSpecialChars.IsEnabled() {
			t.Error("NoSpecialChars should be enabled")
		}

		sw := cfg.Rules.SensitiveWords
		if !sw.IsEnabled() {
			t.Error("SensitiveWords should be enabled")
		}
		if len(sw.Words) != 3 {
			t.Errorf("Expected 3 sensitive words, got %d", len(sw.Words))
		}
		if sw.Words[0] != "password" {
			t.Errorf("First word = %q, want %q", sw.Words[0], "password")
		}
		if len(sw.SafePhrases) != 3 {
			t.Errorf("Expected 3 safe phrases, got %d", len(sw.SafePhrases))
		}

		cp := cfg.Rules.CustomPatterns
		if !cp.IsEnabled() {
			t.Error("CustomPatterns should be enabled")
		}
		if len(cp.Patterns) != 2 {
			t.Errorf("Expected 2 patterns, got %d", len(cp.Patterns))
		}
	})

	t.Run("config with disabled sensitive words", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "disabled_sw.yaml")

		yamlContent := `
rules:
  sensitive_words:
    enabled: false
    words:
      - password
`
		if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		loader := NewLoader(configPath)
		cfg, err := loader.Load()

		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if cfg.Rules.SensitiveWords.IsEnabled() {
			t.Error("SensitiveWords should be disabled")
		}

		if len(cfg.Rules.SensitiveWords.Words) != 1 {
			t.Errorf("Expected 1 word, got %d", len(cfg.Rules.SensitiveWords.Words))
		}
	})

	t.Run("config with empty arrays", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "empty_arrays.yaml")

		yamlContent := `
rules:
  sensitive_words:
    enabled: true
    words: []
    safe_phrases: []
  custom_patterns:
    enabled: true
    patterns: []
`
		if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		loader := NewLoader(configPath)
		cfg, err := loader.Load()

		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		sw := cfg.Rules.SensitiveWords
		if sw.Words == nil {
			t.Error("Words should be empty slice, not nil")
		}
		if len(sw.Words) != 0 {
			t.Errorf("Expected 0 words, got %d", len(sw.Words))
		}

		cp := cfg.Rules.CustomPatterns
		if cp.Patterns == nil {
			t.Error("Patterns should be empty slice, not nil")
		}
		if len(cp.Patterns) != 0 {
			t.Errorf("Expected 0 patterns, got %d", len(cp.Patterns))
		}
	})

	t.Run("config with only enabled field", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "enabled_only.yaml")

		yamlContent := `
rules:
  lowercase:
    enabled: false
  english_only:
    enabled: true
`
		if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		loader := NewLoader(configPath)
		cfg, err := loader.Load()

		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if cfg.Rules.Lowercase.IsEnabled() {
			t.Error("Lowercase should be disabled")
		}
		if !cfg.Rules.EnglishOnly.IsEnabled() {
			t.Error("EnglishOnly should be enabled")
		}
	})

	t.Run("config with special characters in patterns", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "special_chars.yaml")

		yamlContent := `
rules:
  custom_patterns:
    enabled: true
    patterns:
      - "^[A-Z]+$"
      - "\\s+"
      - "[^\\w\\s]+"
`
		if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		loader := NewLoader(configPath)
		cfg, err := loader.Load()

		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		cp := cfg.Rules.CustomPatterns
		if len(cp.Patterns) != 3 {
			t.Errorf("Expected 3 patterns, got %d", len(cp.Patterns))
		}
	})
}

func TestNewLoader(t *testing.T) {
	t.Run("creates loader with path", func(t *testing.T) {
		loader := NewLoader("/path/to/config.yaml")
		if loader == nil {
			t.Fatal("NewLoader() returned nil")
		}
		if loader.path != "/path/to/config.yaml" {
			t.Errorf("path = %q, want %q", loader.path, "/path/to/config.yaml")
		}
	})

	t.Run("creates loader with empty path", func(t *testing.T) {
		loader := NewLoader("")
		if loader == nil {
			t.Fatal("NewLoader() returned nil")
		}
		if loader.path != "" {
			t.Errorf("path = %q, want empty string", loader.path)
		}
	})
}

func TestBoolPtr(t *testing.T) {
	t.Run("true pointer", func(t *testing.T) {
		ptr := boolPtr(true)
		if ptr == nil {
			t.Fatal("boolPtr(true) returned nil")
		}
		if *ptr != true {
			t.Errorf("*ptr = %v, want true", *ptr)
		}
	})

	t.Run("false pointer", func(t *testing.T) {
		ptr := boolPtr(false)
		if ptr == nil {
			t.Fatal("boolPtr(false) returned nil")
		}
		if *ptr != false {
			t.Errorf("*ptr = %v, want false", *ptr)
		}
	})

	t.Run("different pointers for different values", func(t *testing.T) {
		ptr1 := boolPtr(true)
		ptr2 := boolPtr(false)
		if ptr1 == ptr2 {
			t.Error("boolPtr should return different pointers for different values")
		}
	})
}

func TestLoader_Integration(t *testing.T) {
	t.Run("load default then override with file", func(t *testing.T) {
		// Load default
		defaultLoader := NewLoader("")
		defaultCfg, err := defaultLoader.Load()
		if err != nil {
			t.Fatalf("Load() default error = %v", err)
		}

		// Create config file with overrides
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "override.yaml")

		yamlContent := `
rules:
  lowercase:
    enabled: false
  sensitive_words:
    enabled: false
`
		if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		fileLoader := NewLoader(configPath)
		fileCfg, err := fileLoader.Load()
		if err != nil {
			t.Fatalf("Load() file error = %v", err)
		}

		if defaultCfg.Rules.Lowercase.IsEnabled() == fileCfg.Rules.Lowercase.IsEnabled() {
			t.Error("File config should override default")
		}
		if defaultCfg.Rules.SensitiveWords.IsEnabled() == fileCfg.Rules.SensitiveWords.IsEnabled() {
			t.Error("File config should override default")
		}
	})

	t.Run("multiple loads same file consistent", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "consistent.yaml")

		yamlContent := `
rules:
  lowercase:
    enabled: true
  english_only:
    enabled: false
`
		if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		loader := NewLoader(configPath)

		cfg1, err := loader.Load()
		if err != nil {
			t.Fatalf("Load() first error = %v", err)
		}

		cfg2, err := loader.Load()
		if err != nil {
			t.Fatalf("Load() second error = %v", err)
		}

		if cfg1.Rules.Lowercase.IsEnabled() != cfg2.Rules.Lowercase.IsEnabled() {
			t.Error("Multiple loads should return consistent results")
		}
		if cfg1.Rules.EnglishOnly.IsEnabled() != cfg2.Rules.EnglishOnly.IsEnabled() {
			t.Error("Multiple loads should return consistent results")
		}
	})
}

func TestConfig_EdgeCases(t *testing.T) {
	t.Run("yaml with comments only", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "comments.yaml")

		yamlContent := `
# This is a comment
# Another comment
`
		if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		loader := NewLoader(configPath)
		cfg, err := loader.Load()

		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg == nil {
			t.Fatal("Load() returned nil")
		}
	})

	t.Run("yaml with null values", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "null.yaml")

		yamlContent := `
rules:
  lowercase: null
  english_only: ~
`
		if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		loader := NewLoader(configPath)
		cfg, err := loader.Load()

		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if cfg.Rules.Lowercase != nil {
			t.Error("Lowercase should be nil for null value")
		}
		if cfg.Rules.EnglishOnly != nil {
			t.Error("EnglishOnly should be nil for null value")
		}
	})

	t.Run("yaml with wrong types", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "wrong_types.yaml")

		yamlContent := `
rules:
  lowercase:
    enabled: "yes"  # Should be boolean, not string
`
		if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		loader := NewLoader(configPath)
		_, err := loader.Load()

		if err == nil {
			t.Log("YAML unmarshal succeeded with wrong type (may be handled gracefully)")
		}
	})

	t.Run("very long config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "long.yaml")

		yamlContent := "rules:\n  custom_patterns:\n    enabled: true\n    patterns:\n"
		for i := 0; i < 100; i++ {
			yamlContent += `      - "pattern` + string(rune('0'+i%10)) + `"` + "\n"
		}

		if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		loader := NewLoader(configPath)
		cfg, err := loader.Load()

		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if len(cfg.Rules.CustomPatterns.Patterns) != 100 {
			t.Errorf("Expected 100 patterns, got %d", len(cfg.Rules.CustomPatterns.Patterns))
		}
	})

	t.Run("file permission denied", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("Skipping test when running as root")
		}

		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "noperm.yaml")

		if err := os.WriteFile(configPath, []byte("rules: {}"), 0000); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		loader := NewLoader(configPath)
		_, err := loader.Load()

		if err == nil {
			t.Error("Load() should return error for permission denied")
		}
	})
}
