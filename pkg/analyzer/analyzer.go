package analyzer

import (
	"fmt"
	"os"
	"utility/pkg/analyzer/rules"
	"utility/pkg/analyzer/rules/config"
	"utility/pkg/checker"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

const Doc = `loglinter checks log messages for compliance with logging best practices.`

func NewAnalyzer() *analysis.Analyzer {
	var configPath string

	analyzer := &analysis.Analyzer{
		Name: "loglinter",
		Doc:  Doc,
		Run: func(pass *analysis.Pass) (interface{}, error) {
			return run(pass, configPath)
		},
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}

	analyzer.Flags.StringVar(&configPath, "config", "", "Path to config file")
	analyzer.Flags.StringVar(&configPath, "loglinter.yml", "", "Path to config file")

	return analyzer
}

func run(pass *analysis.Pass, configPath string) (interface{}, error) {
	// 1. Загрузка конфигурации
	cfg, err := loadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loglinter: warning: failed to load config: %v\n", err)
		cfg = config.DefaultConfig()
	}

	// 2. Отладочный вывод (удалите после отладки)
	fmt.Fprintf(os.Stderr, "=== LOGLINTER CONFIG ===\n")
	fmt.Fprintf(os.Stderr, "Config path: %s\n", configPath)
	fmt.Fprintf(os.Stderr, "Lowercase enabled: %v\n", cfg.Rules.Lowercase.IsEnabled())
	fmt.Fprintf(os.Stderr, "EnglishOnly enabled: %v\n", cfg.Rules.EnglishOnly.IsEnabled())
	fmt.Fprintf(os.Stderr, "NoSpecialChars enabled: %v\n", cfg.Rules.NoSpecialChars.IsEnabled())
	fmt.Fprintf(os.Stderr, "SensitiveWords enabled: %v\n", cfg.Rules.SensitiveWords.IsEnabled())
	fmt.Fprintf(os.Stderr, "CustomPatterns enabled: %v\n", cfg.Rules.CustomPatterns.IsEnabled())
	fmt.Fprintf(os.Stderr, "========================\n")

	// 3. Инициализация правил
	rulesList := loadRules(cfg)

	// 4. Создание чекера и запуск
	check := checker.New(rulesList)
	check.Check(pass)

	return nil, nil
}

func loadConfig(path string) (*config.Config, error) {
	loader := config.NewLoader(path)
	return loader.Load()
}

func loadRules(cfg *config.Config) []rules.Rule {
	allRules, _ := rules.GetAllRules()
	enabledRules := make([]rules.Rule, 0, len(allRules))

	for _, rule := range allRules {
		var isEnabled bool
		var ruleConfig map[string]any

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
				fmt.Fprintf(os.Stderr, "CustomPatterns enabled: %v\n", isEnabled)
				fmt.Fprintf(os.Stderr, "CustomPatterns patterns: %v\n", cp.Patterns)
				if len(cp.Patterns) > 0 {
					ruleConfig = map[string]any{"patterns": cp.Patterns}
				}
			}
		default:
			isEnabled = true
		}

		if !isEnabled {
			fmt.Fprintf(os.Stderr, "Disabled rule: %s\n", rule.Name())
			continue
		}

		if ruleConfig != nil {
			fmt.Fprintf(os.Stderr, "Configuring rule: %s = %v\n", rule.Name(), ruleConfig)
			if err := rule.Configure(ruleConfig); err != nil {
				fmt.Fprintf(os.Stderr, "loglinter: failed to configure rule %s: %v\n", rule.Name(), err)
			}
		}

		fmt.Fprintf(os.Stderr, "Enabled rule: %s\n", rule.Name())
		enabledRules = append(enabledRules, rule)
	}

	return enabledRules
}
