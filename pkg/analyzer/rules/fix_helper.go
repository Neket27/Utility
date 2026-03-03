package rules

//import "unicode"
//
//func ApplyAllFixes(msg string, maxDots int) string {
//	rules, err:= RuleSet()
//
//	if err != nil {
//		return msg
//	}
//	for _, rule := range rules {
//		if(rule.IsAutoFixEnabled()){
//
//		}
//
//	}
//	cleaned := cleanSpecialChars(msg)
//	cleaned = truncateConsecutiveDots(cleaned, maxDots)
//
//	if len(cleaned) > 0 {
//		runes := []rune(cleaned)
//		if unicode.IsLetter(runes[0]) && unicode.IsUpper(runes[0]) {
//			runes[0] = unicode.ToLower(runes[0])
//			cleaned = string(runes)
//		}
//	}
//
//	return cleaned
//}
