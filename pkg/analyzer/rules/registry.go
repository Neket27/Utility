package rules

func GetAllRules() ([]Rule, error) {
	builders := []RuleBuilder{
		NewLowercaseRule,
		NewEnglishOnlyRule,
		NewNoSpecialCharsRule,
		NewSensitiveWordsRule,
		NewCustomPatternsRule,
	}

	rules := make([]Rule, 0, len(builders))
	for _, builder := range builders {
		rules = append(rules, builder())
	}

	return rules, nil
}
