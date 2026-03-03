package rules

func RuleSet() ([]Rule, error) {
	builders := []RuleBuilder{
		NewNoSpecialCharsRule,
		NewLowercaseRule,
		NewEnglishOnlyRule,
		NewSensitiveWordsRule,
		NewCustomPatternsRule,
	}

	rules := make([]Rule, 0, len(builders))

	for _, builder := range builders {
		rules = append(rules, builder())
	}

	return rules, nil
}
