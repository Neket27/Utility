package analyzer

import (
	"encoding/json"
	_ "flag"
	"utility/pkg/analyzer/rules"
	"utility/pkg/analyzer/rules/config"
	"utility/pkg/checker"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

const Doc = `loglinter checks log messages for compliance with logging best practices.`

func NewAnalyzer() *analysis.Analyzer {
	var configPath string
	var configData string // Для передачи конфига из golangci-lint

	analyzer := &analysis.Analyzer{
		Name:     "loglinter",
		Doc:      Doc,
		Run:      func(pass *analysis.Pass) (interface{}, error) { return run(pass, configPath, configData) },
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}

	// Флаги для standalone режима
	analyzer.Flags.StringVar(&configPath, "config", "", "Path to config file")
	analyzer.Flags.StringVar(&configData, "config-data", "", "Config data from golangci-lint (base64)")

	return analyzer
}

func run(pass *analysis.Pass, configPath, configData string) (interface{}, error) {
	// 1. Загрузка конфигурации
	cfg, err := loadConfig(configPath, configData)
	if err != nil {
		pass.Reportf(pass.Files[0].Pos(), "loglinter: failed to load config: %v", err)
		cfg = config.DefaultConfig() // Fallback на дефолт
	}

	// 2. Инициализация правил
	rulesList := loadRules(cfg)

	// 3. Создание чекера и запуск
	check := checker.New(rulesList)
	check.Check(pass)

	return nil, nil
}

func loadConfig(path, data string) (*config.Config, error) {
	// Если данные пришли от golangci-lint (через linters-settings)
	if data != "" {
		// Здесь можно распарсить JSON, переданный из YAML golangci-lint
		cfg := config.DefaultConfig()
		if err := json.Unmarshal([]byte(data), cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	}

	// Иначе загружаем из файла
	loader := config.NewLoader(path)
	return loader.Load()
}

func loadRules(cfg *config.Config) []rules.Rule {
	allRules, _ := rules.GetAllRules()
	enabledRules := make([]rules.Rule, 0, len(allRules))

	for _, rule := range allRules {
		var isEnabled bool
		var ruleConfig map[string]any

		// Маппинг конфига на правило
		switch rule.Name() {
		case "lowercase":
			isEnabled = cfg.Rules.Lowercase.IsEnabled()
		case "english_only":
			isEnabled = cfg.Rules.EnglishOnly.IsEnabled()
		case "no_special_chars":
			isEnabled = cfg.Rules.NoSpecialChars.IsEnabled()
		case "sensitive_words":
			if sw := cfg.Rules.SensitiveWords; sw != nil {
				isEnabled = sw.IsEnabled()
				if len(sw.Words) > 0 {
					ruleConfig = map[string]any{"words": sw.Words}
				}
			}
		case "custom_patterns":
			if cp := cfg.Rules.CustomPatterns; cp != nil {
				isEnabled = cp.IsEnabled()
				if len(cp.Patterns) > 0 {
					ruleConfig = map[string]any{"patterns": cp.Patterns}
				}
			}
		default:
			isEnabled = true
		}

		if !isEnabled {
			continue
		}

		if ruleConfig != nil {
			rule.Configure(ruleConfig)
		}

		enabledRules = append(enabledRules, rule)
	}

	return enabledRules
}
