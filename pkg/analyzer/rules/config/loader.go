package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

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

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
