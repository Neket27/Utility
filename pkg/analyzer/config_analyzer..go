package analyzer

type rulesConfig struct {
	Rules map[string]ruleConfig
}

type ruleConfig struct {
	Enabled *bool
	Data    map[string]any
}

func parseConfig(cfg any) rulesConfig {
	result := rulesConfig{
		Rules: make(map[string]ruleConfig),
	}

	if cfg == nil {
		return result
	}

	cfgMap, ok := cfg.(map[string]any)
	if !ok {
		return result
	}

	if rulesCfg, ok := cfgMap["rules"].(map[string]any); ok {
		for ruleName, ruleCfg := range rulesCfg {
			if ruleData, ok := ruleCfg.(map[string]any); ok {
				rc := ruleConfig{Data: make(map[string]any)}
				if enabled, ok := ruleData["enabled"].(bool); ok {
					rc.Enabled = &enabled
				}
				for k, v := range ruleData {
					if k != "enabled" {
						rc.Data[k] = v
					}
				}
				result.Rules[ruleName] = rc
			}
		}
	}

	return result
}
