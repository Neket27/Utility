package rules

import "unicode"

const RuleEnglishOnlyName = "english_only"

type EnglishOnlyRule struct {
	BaseRule
}

func NewEnglishOnlyRule() Rule {
	return &EnglishOnlyRule{
		BaseRule: NewBaseRule(RuleEnglishOnlyName, "Checks that log messages contain only English characters"),
	}
}

func (r *EnglishOnlyRule) Check(ctx *CheckContext) *RuleResult {
	if !r.Enabled() {
		return ResultPass()
	}

	if !CheckEnglishOnly(ctx.Msg) {
		return ResultFail("log message must be in English only")
	}

	return ResultPass()
}

func CheckEnglishOnly(msg string) bool {
	for _, ch := range msg {
		if unicode.IsLetter(ch) && !unicode.Is(unicode.Latin, ch) {
			return false
		}
	}
	return true
}
