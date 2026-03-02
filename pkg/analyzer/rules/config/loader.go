package config

import (
	"encoding/json"
	"fmt"
	"os"
	_ "path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Loader загружает конфигурацию из различных источников
type Loader struct {
	configPath string
}

// NewLoader создаёт новый загрузчик
func NewLoader(configPath string) *Loader {
	return &Loader{configPath: configPath}
}

// Load загружает конфигурацию
// Приоритет: 1. Явно указанный файл, 2. Файлы в текущей директории, 3. Default
func (l *Loader) Load() (*Config, error) {
	cfg := DefaultConfig()

	path := l.findConfigFile()
	if path == "" {
		return cfg, nil // Возвращаем дефолт, если файл не найден
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	// Определяем формат по расширению
	if strings.HasSuffix(path, ".json") {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse JSON config: %w", err)
		}
	} else {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse YAML config: %w", err)
		}
	}

	return cfg, nil
}

// findConfigFile ищет файл конфигурации
func (l *Loader) findConfigFile() string {
	// 1. Явно указанный путь
	if l.configPath != "" {
		if _, err := os.Stat(l.configPath); err == nil {
			return l.configPath
		}
	}

	// 2. Поиск в текущей директории
	candidates := []string{
		".loglinter.yml",
		".loglinter.yaml",
		".loglinter.json",
		"loglinter.yml",
		"loglinter.yaml",
		"loglinter.json",
	}

	for _, name := range candidates {
		if _, err := os.Stat(name); err == nil {
			return name
		}
	}

	return ""
}
