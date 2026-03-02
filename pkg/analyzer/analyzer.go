package analyzer

import (
	"utility/pkg/analyzer/rules"
	"utility/pkg/analyzer/rules/config"
	"utility/pkg/checker"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

const Doc = `loglinter checks log messages for compliance with logging best practices.`

var configPath string

func NewAnalyzer() *analysis.Analyzer {
	analyzer := &analysis.Analyzer{
		Name:     "loglinter",
		Doc:      Doc,
		Run:      run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}

	analyzer.Flags.StringVar(&configPath, "config", "", "Path to config file")

	return analyzer
}

func run(pass *analysis.Pass) (interface{}, error) {
	cfg, err := loadConfig(configPath)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	rulesList := loadRules(cfg)

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
				ruleConfig = make(map[string]any)
				if len(sw.Words) > 0 {
					ruleConfig["words"] = sw.Words
				}
				if len(sw.SafePhrases) > 0 {
					ruleConfig["safe_phrases"] = sw.SafePhrases
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
			if err := rule.Configure(ruleConfig); err != nil {
				return nil
			}
		}

		enabledRules = append(enabledRules, rule)
	}

	return enabledRules
}
