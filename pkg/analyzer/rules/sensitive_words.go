package rules

import (
	"fmt"
	"regexp"
	"strings"
)

const RuleSensitiveWordsName = "sensitive_words"

var DefaultSensitiveWords = []string{
	"password",
	"passwd",
	"pwd",
	"secret",
	"token",
	"api_key",
	"apikey",
	"credential",
	"private_key",
	"access_token",
	"refresh_token",
	"secret_key",
	"encryption_key",
}

type SensitiveWordsRule struct {
	BaseRule
	words []*regexp.Regexp
}

func NewSensitiveWordsRule() Rule {
	rule := &SensitiveWordsRule{
		BaseRule: NewBaseRule(RuleSensitiveWordsName, "Checks that log messages don't contain sensitive data"),
	}
	rule.compileWords(DefaultSensitiveWords)
	return rule
}

func (r *SensitiveWordsRule) Configure(config map[string]any) error {
	if err := r.BaseRule.Configure(config); err != nil {
		return err
	}

	if words, ok := config["words"].([]any); ok && len(words) > 0 {
		customWords := make([]string, 0, len(words))
		for _, w := range words {
			if s, ok := w.(string); ok {
				customWords = append(customWords, s)
			}
		}
		r.compileWords(customWords)
	}

	return nil
}

func (r *SensitiveWordsRule) compileWords(words []string) {
	r.words = make([]*regexp.Regexp, 0, len(words))
	for _, word := range words {
		// Спецсимволы экранированы и проверяются границы слов
		escaped := regexp.QuoteMeta(word)
		pattern := `(?i)\b` + escaped + `\b`
		if re, err := regexp.Compile(pattern); err == nil {
			r.words = append(r.words, re)
		}
	}
}

func (r *SensitiveWordsRule) Check(ctx *CheckContext) *RuleResult {
	if !r.Enabled() {
		return ResultPass()
	}

	if sensitiveWord := r.findSensitiveWord(ctx.Msg); sensitiveWord != "" {
		return ResultFail(fmt.Sprintf("log message contains sensitive data: %s", sensitiveWord))
	}

	return ResultPass()
}

func (r *SensitiveWordsRule) findSensitiveWord(msg string) string {
	msgLower := strings.ToLower(msg)

	for _, re := range r.words {
		if match := re.FindString(msgLower); match != "" {
			return match
		}
	}

	return ""
}
