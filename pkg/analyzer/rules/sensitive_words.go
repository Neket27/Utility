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

var DefaultSafePhrases = []string{
	"validated",
	"verified",
	"expired",
	"refreshed",
	"rotated",
	"changed",
	"updated",
	"deleted",
	"created",
	"generated",
	"revoked",
	"invalid",
	"missing",
	"required",
	"optional",
	"configured",
	"initialized",
	"loaded",
	"saved",
	"cleared",
	"reset",
}

var dangerousPatterns = []string{
	`^\s*[:=+]\s*\S+`,
	`^\s*is\s+\S+`,
	`^\s*has\s+\S+`,
	`^\s*value\s*[:=]`,
}

type SensitiveWordsRule struct {
	BaseRule
	words       []*regexp.Regexp
	safePattern *regexp.Regexp
}

func NewSensitiveWordsRule() Rule {
	rule := &SensitiveWordsRule{
		BaseRule: NewBaseRule(RuleSensitiveWordsName, "Checks that log messages don't contain sensitive data"),
	}
	rule.compileWords(DefaultSensitiveWords)
	rule.compileSafePhrases(DefaultSafePhrases)
	return rule
}

func (r *SensitiveWordsRule) Configure(config map[string]any) error {
	if err := r.BaseRule.Configure(config); err != nil {
		return err
	}

	if words, ok := config["words"].([]string); ok && len(words) > 0 {
		r.compileWords(words)
	}

	if safePhrases, ok := config["safe_phrases"].([]string); ok && len(safePhrases) > 0 {
		r.compileSafePhrases(safePhrases)
	}

	return nil
}

func (r *SensitiveWordsRule) compileWords(words []string) {
	r.words = make([]*regexp.Regexp, 0, len(words))
	for _, word := range words {
		escaped := regexp.QuoteMeta(word)
		patterns := []string{
			`(?i)\b` + escaped + `\b`,        // точное совпадение
			`(?i)\b` + escaped + `\d+\b`,     // password123
			`(?i)\b` + escaped + `_[a-z]+\b`, // password_prod
		}
		for _, pattern := range patterns {
			if re, err := regexp.Compile(pattern); err == nil {
				r.words = append(r.words, re)
			}
		}
	}
}

func (r *SensitiveWordsRule) compileSafePhrases(phrases []string) {
	if len(phrases) == 0 {
		return
	}
	pattern := `(?i)\b(` + strings.Join(phrases, "|") + `)\b`
	r.safePattern = regexp.MustCompile(pattern)
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
	for _, re := range r.words {
		match := re.FindString(msg)
		if match == "" {
			continue
		}

		if r.hasSafePhrase(msg) {
			continue
		}

		return match
	}

	return ""
}

func (r *SensitiveWordsRule) hasDangerousContext(msg, word string) bool {
	wordIdx := strings.Index(strings.ToLower(msg), strings.ToLower(word))
	if wordIdx == -1 {
		return false
	}

	afterWord := msg[wordIdx+len(word):]

	for _, pattern := range dangerousPatterns {
		if matched, _ := regexp.MatchString(pattern, afterWord); matched {
			return true
		}
	}

	return false
}

func (r *SensitiveWordsRule) hasSafePhrase(msg string) bool {
	if r.safePattern == nil {
		return false
	}
	return r.safePattern.MatchString(msg)
}
