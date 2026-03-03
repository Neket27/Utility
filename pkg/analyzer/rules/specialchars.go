package rules

import (
	"slices"
	"unicode"
)

const RuleNoSpecialCharsName = "no_special_chars"

// DefaultMaxConsecutiveDots =3 -> For example, allow "waiting..." By TZ=0
const DefaultMaxConsecutiveDots = 0

var allowedPunctuation = []rune{'.', ',', ':', '-', '_', '/', '(', ')', '[', ']', '{', '}', '=', '+', '"', '\''}
var problematicSpecialChar = []rune{'!', '?', '@', '#', '$', '%', '^', '&', '*', '|', '`', '~'}

type NoSpecialCharsRule struct {
	BaseRule
	MaxConsecutiveDots int
}

func NewNoSpecialCharsRule() Rule {
	return &NoSpecialCharsRule{
		BaseRule:           NewBaseRule(RuleNoSpecialCharsName, "Checks that log messages don't contain special characters or emojis"),
		MaxConsecutiveDots: DefaultMaxConsecutiveDots,
	}
}

/*func (r *NoSpecialCharsRule) Configure(config map[string]any) error {
	if err := r.BaseRule.Configure(config); err != nil {
		return err
	}
	if maxDots, ok := config["max_consecutive_dots"].(int); ok {
		r.MaxConsecutiveDots = maxDots
	}

	return nil
}*/

func (r *NoSpecialCharsRule) Check(ctx *CheckContext) *RuleResult {
	if !r.Enabled() {
		return ResultPass()
	}

	valid, _ := CheckNoSpecialChars(ctx.Msg, r.MaxConsecutiveDots)

	if valid {
		return ResultPass()
	}

	var fix *SuggestedFix
	if r.AutoFixEnabled() {
		cleanedMsg := CleanSpecialChars(ctx.Msg)
		cleanedMsg = truncateConsecutiveDots(cleanedMsg, r.MaxConsecutiveDots)

		if cleanedMsg != ctx.Msg {
			fix = &SuggestedFix{
				Message: "Remove special characters and emojis",
				NewText: cleanedMsg,
			}
		}
	}

	return &RuleResult{
		Passed:       false,
		Message:      "log message must not contain special characters or emojis",
		SuggestedFix: fix,
	}
}

func CheckNoSpecialChars(msg string, maxDots int) (bool, rune) {
	runes := []rune(msg)
	dotCount := 0
	prevWasDot := false

	for _, ch := range runes {
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || unicode.IsSpace(ch) {
			prevWasDot = false
			continue
		}

		if isEmoji(ch) {
			return false, ch
		}

		if isProblematicSpecialChar(ch) {
			return false, ch
		}

		if slices.Contains(allowedPunctuation, ch) {
			if ch == '.' {
				if prevWasDot {
					dotCount++
				} else {
					dotCount = 1
				}

				if dotCount > maxDots {
					return false, ch
				}
				prevWasDot = true
			} else {
				prevWasDot = false
			}
			continue
		}

		return false, ch
	}

	return true, 0
}

func isEmoji(ch rune) bool {
	return (ch >= 0x1F600 && ch <= 0x1F64F) || // Emoticon
		(ch >= 0x1F300 && ch <= 0x1F5FF) || // Misc Symbols and Pictographs
		(ch >= 0x1F680 && ch <= 0x1F6FF) || // Transport and Map
		(ch >= 0x1F1E0 && ch <= 0x1F1FF) || // Regional (flags)
		(ch >= 0x2600 && ch <= 0x26FF) || // Misc symbols
		(ch >= 0x2700 && ch <= 0x27BF) || // Dingbats
		(ch >= 0xFE00 && ch <= 0xFE0F) || // Variation Selectors
		(ch >= 0x1F900 && ch <= 0x1F9FF) // Supplemental Symbols and Pictographs
}

func isProblematicSpecialChar(ch rune) bool {
	return slices.Contains(problematicSpecialChar, ch)
}

func CleanSpecialChars(msg string) string {
	var result []rune
	for _, ch := range msg {
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || unicode.IsSpace(ch) {
			result = append(result, ch)
			continue
		}
		if slices.Contains(allowedPunctuation, ch) && !isEmoji(ch) {
			result = append(result, ch)
		}
	}
	return string(result)
}

func truncateConsecutiveDots(msg string, maxDots int) string {
	if maxDots < 0 {
		return msg
	}

	var result []rune
	dotCount := 0
	prevWasDot := false

	for _, ch := range msg {
		if ch == '.' {
			if prevWasDot {
				dotCount++
			} else {
				dotCount = 1
			}

			if dotCount <= maxDots {
				result = append(result, ch)
			}

			prevWasDot = true
		} else {
			result = append(result, ch)
			prevWasDot = false
			dotCount = 0
		}
	}

	return string(result)
}
