package rules

import (
	"fmt"
	"strings"
	"unicode"
	"utility/pkg/translator"
)

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

	hasNonEnglish, nonEnglishPart := CheckEnglishOnly(ctx.Msg)
	if !hasNonEnglish {
		return ResultPass()
	}

	var fix *SuggestedFix
	if r.AutoFixEnabled() {
		translated, err := translator.Translate(ctx.Msg, "ru", "en")
		if err == nil && translated != "" && translated != ctx.Msg {
			fix = &SuggestedFix{
				Message: "Translate to English",
				NewText: translated,
			}
		}
	}

	return &RuleResult{
		Passed:       false,
		Message:      fmt.Sprintf("log message must be in English only (found non-English: %q)", nonEnglishPart),
		SuggestedFix: fix,
	}
}

func CheckEnglishOnly(msg string) (bool, string) {
	var nonEnglish strings.Builder
	for _, ch := range msg {
		if unicode.IsLetter(ch) && !unicode.Is(unicode.Latin, ch) {
			nonEnglish.WriteRune(ch)
		}
	}
	if nonEnglish.Len() > 0 {
		return true, nonEnglish.String()
	}
	return false, ""
}
