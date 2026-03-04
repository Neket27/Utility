package rules

import (
	"strings"
	"testing"
	"unicode"
)

func TestCheckEnglishOnly(t *testing.T) {
	tests := []struct {
		name      string
		msg       string
		wantValid bool
		note      string
	}{
		// === Базовые валидные случаи (только латиница) ===
		{"empty", "", true, "пустое сообщение — валидно"},
		{"single english letter", "a", true, ""},
		{"english word", "server", true, ""},
		{"english sentence", "server started successfully", true, ""},
		{"with numbers", "error 404", true, "цифры разрешены"},
		{"with punctuation", "connection: ok!", true, "знаки препинания разрешены"},
		{"with spaces", "  multiple   spaces  ", true, "пробелы разрешены"},
		{"with tabs newlines", "line1\nline2\tcol", true, "whitespace разрешён"},
		{"uppercase english", "ERROR FAILED", true, ""},
		{"mixed case english", "Server Started", true, ""},

		// === Латиница с диакритикой (считается Latin) ===
		{"latin with acute", "café", true, "é — Latin Extended"},
		{"latin with umlaut", "über", true, "ü — Latin Extended"},
		{"latin with tilde", "señor", true, "ñ — Latin Extended"},
		{"latin with cedilla", "français", true, "ç — Latin Extended"},
		{"latin with grave", "à la carte", true, "à — Latin Extended"},
		{"latin with circumflex", "hôtel", true, "ô — Latin Extended"},
		{"latin with ring", "Ångström", true, "Å — Latin Extended"},
		{"latin with ogonek", "katalog", true, "ę — Latin Extended"},
		{"latin with caron", "český", true, "č — Latin Extended"},
		{"latin with dot", "żółw", true, "ż — Latin Extended"},
		{"latin with stroke", "đàn", true, "đ — Latin Extended"},
		{"latin ligatures", "ﬁle", true, "ﬁ — Latin ligature"},
		{"latin ae", "æther", true, "æ — Latin ligature"},
		{"latin oe", "coeur", true, "œ — Latin ligature"},

		// === Невалидные: другие скрипты ===
		{"cyrillic", "сервер", false, "кириллица — не Latin"},
		{"cyrillic mixed", "server запуск", false, ""},
		{"cyrillic single", "а", false, "кириллическая 'а' (U+0430)"},
		{"chinese simplified", "服务器", false, "китайские иероглифы"},
		{"chinese traditional", "服務器", false, ""},
		{"japanese hiragana", "さーばー", false, "хирагана"},
		{"japanese katakana", "サーバー", false, "катакана"},
		{"japanese kanji", "起動", false, "кандзи"},
		{"korean hangul", "서버", false, "хангыль"},
		{"arabic", "خادم", false, "арабский"},
		{"hebrew", "שרת", false, "иврит"},
		{"greek", "διακομιστής", false, "греческий"},
		{"thai", "เซิร์ฟเวอร์", false, "тайский"},
		{"georgian", "სერვერი", false, "грузинский"},
		{"armenian", "սերվեր", false, "армянский"},
		{"devanagari", "सर्वर", false, "деванагари"},
		{"bengali", "সার্ভার", false, "бенгальский"},
		{"tamil", "சேவையகம்", false, "тамильский"},
		{"telugu", "సర్వర్", false, "телугу"},
		{"kannada", "ಸರ್ವರ್", false, "каннада"},
		{"malayalam", "സർവർ", false, "малаялам"},
		{"sinhala", "සේවකය", false, "сингальский"},
		{"myanmar", "ဆာဗာ", false, "мьянма"},
		{"khmer", "ម៉ាស៊ីន", false, "кхмерский"},
		{"lao", "ເຊີບເວີ", false, "лаосский"},
		{"tibetan", "ཞབས་", false, "тибетский"},
		{"mongolian", "сервер", false, "монгольский"},
		{"ethiopic", "ሰርቨር", false, "эфиопский"},
		{"cherokee", "ᏚᎸᏫᏍᏗ", false, "чероки"},
		{"syriac", "ܫܪܬ", false, "сирийский"},
		{"thaana", "ސާވަރ", false, "таана"},

		// === Смешанные скрипты ===
		{"english + cyrillic", "server сервер", false, ""},
		{"english + chinese", "error 错误", false, ""},
		{"english + arabic", "started بدء", false, ""},
		{"cyrillic + english", "запуск server", false, ""},
		{"multiple scripts", "server サーバー 服务器", false, ""},
		{"latin + cyrillic lookalike", "server аdmin", false, "кириллическая 'а' в середине"},

		// === Эмодзи и символы ===
		{"emoji only", "😀", true, "эмодзи — не буквы, пропускаются"},
		{"emoji with text", "server 😀", true, ""},
		{"emoji multiple", "🚀🔥💥", true, ""},
		{"flag emoji", "🇺🇸🇷🇺", true, ""},
		{"skin tone emoji", "👍🏿", true, ""},
		{"zero width joiner", "👨‍👩‍👧‍👦", true, ""},
		{"symbols", "©®™", true, "символы — не буквы"},
		{"math symbols", "∑∏∫", true, ""},
		{"currency", "$€£¥", true, ""},
		{"arrows", "←→↑↓", true, ""},
		{"box drawing", "┌─┐", true, ""},
		{"braille", "⠓⠑⠇⠇⠕", true, ""},

		// === Граничные случаи Unicode ===
		{"combining acute", "e\u0301", true, "e + combining acute = é (Latin)"},
		{"combining grave", "a\u0300", true, ""},
		{"combining tilde", "n\u0303", true, ""},
		{"combining umlaut", "u\u0308", true, ""},
		{"combining cyrillic", "а\u0483", false, "кириллическая буква + combining"},
		{"zero width space", "test\u200b", true, "ZWSP — не буква"},
		{"zero width non-joiner", "test\u200c", true, ""},
		{"zero width joiner", "test\u200d", true, ""},
		{"left-to-right mark", "\u200etest", true, "LTR — не буква"},
		{"right-to-left mark", "\u200ftest", true, "RTL — не буква"},
		{"byte order mark", "\ufefftest", true, "BOM — не буква"},
		{"non-breaking space", "test\u00a0", true, "NBSP — не буква"},
		{"en dash", "test–end", true, "en-dash — не буква"},
		{"em dash", "test—end", true, "em-dash — не буква"},
		{"ellipsis", "test…", true, "ellipsis — не буква"},
		{"smart quotes", "test'quote'", true, "кавычки — не буквы"},
		{"guillemets", "test«»", true, ""},
		{"micro sign", "10μm", false, "μ — Greek, не разрешён в English Only"},
		// === Позиционные тесты ===
		{"non-latin at start", "сервер started", false, ""},
		{"non-latin at end", "started сервер", false, ""},
		{"non-latin in middle", "start сервер ed", false, ""},
		{"non-latin single char", "aбc", false, ""},
		{"latin at start cyrillic at end", "test тест", false, ""},

		// === Special Latin ranges ===
		{"IPA extensions", "ʃʒ", true, "IPA — Latin Extended"},
		{"phonetic extensions", "ɐɒ", true, ""},
		{"latin-1 supplement", "ñü", true, ""},
		{"latin extended-a", "āē", true, ""},
		{"latin extended-b", "ƀƁ", true, ""},
		{"latin extended-c", "ꞎꞏ", true, ""},
		{"latin extended-d", "Ꝛꝛ", true, ""},
		{"latin extended-e", "ꭣꭤ", true, ""},
		{"latin extended additional", "ḁḃ", true, ""},
		{"latin ligature fi", "ﬁ", true, ""},
		{"latin ligature fl", "ﬂ", true, ""},
		{"latin small capital", "ᴀʙᴄ", true, ""},

		// === Whitespace variations ===
		{"tab", "\t", true, ""},
		{"newline", "\n", true, ""},
		{"carriage return", "\r", true, ""},
		{"form feed", "\f", true, ""},
		{"vertical tab", "\v", true, ""},
		{"non-breaking space", "\u00a0", true, ""},
		{"en space", "\u2002", true, ""},
		{"em space", "\u2003", true, ""},
		{"thin space", "\u2009", true, ""},
		{"hair space", "\u200a", true, ""},
		{"narrow no-break space", "\u202f", true, ""},
		{"medium mathematical space", "\u205f", true, ""},
		{"ideographic space", "\u3000", true, ""},

		// === Длинный текст ===
		{"long english text", strings.Repeat("english text ", 100), true, ""},
		{"long mixed text", strings.Repeat("text текст ", 50), false, ""},
		{"long cyrillic text", strings.Repeat("текст ", 100), false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, _ := CheckEnglishOnly(tt.msg)

			if valid != tt.wantValid {
				t.Errorf("CheckEnglishOnly(%q) valid = %v, want %v (note: %s)",
					tt.msg, valid, tt.wantValid, tt.note)
			}
		})
	}
}

func TestEnglishOnlyRule_Check(t *testing.T) {
	tests := []struct {
		name       string
		msg        string
		enabled    bool
		wantPassed bool
	}{
		// === Валидные сообщения ===
		{"enabled + english", "server started", true, true},
		{"enabled + empty", "", true, true},
		{"enabled + latin extended", "café résumé", true, true},
		{"enabled + with emoji", "server 😀", true, true},
		{"enabled + with numbers", "error 404", true, true},

		// === Невалидные сообщения ===
		{"enabled + cyrillic", "сервер запущен", true, false},
		{"enabled + chinese", "服务器启动", true, false},
		{"enabled + mixed", "server сервер", true, false},
		{"enabled + arabic", "بدء التشغيل", true, false},
		{"enabled + japanese", "サーバー起動", true, false},

		// === Правило отключено ===
		{"disabled + cyrillic", "сервер", false, true},
		{"disabled + chinese", "服务器", false, true},
		{"disabled + mixed", "server сервер", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewEnglishOnlyRule().(*EnglishOnlyRule)
			rule.SetEnabled(tt.enabled)

			ctx := &CheckContext{Msg: tt.msg}
			result := rule.Check(ctx)

			if result.Passed != tt.wantPassed {
				t.Errorf("Check(%q) passed = %v, want %v", tt.msg, result.Passed, tt.wantPassed)
			}

			if !tt.wantPassed && result.Message == "" {
				t.Error("Check() failed but message is empty")
			}

			if !tt.wantPassed && result.Message != "" {
				if !ContainsIgnoreCase(result.Message, "english") {
					t.Errorf("Message = %q, should mention 'english'", result.Message)
				}
			}
		})
	}
}

func TestEnglishOnlyRule_Meta(t *testing.T) {
	rule := NewEnglishOnlyRule()

	t.Run("Name", func(t *testing.T) {
		if got := rule.Name(); got != RuleEnglishOnlyName {
			t.Errorf("Name() = %q, want %q", got, RuleEnglishOnlyName)
		}
	})

	t.Run("Description", func(t *testing.T) {
		desc := rule.Description()
		if desc == "" {
			t.Error("Description() should not be empty")
		}
		if !ContainsIgnoreCase(desc, "english") {
			t.Errorf("Description() = %q, should mention 'english'", desc)
		}
	})

	t.Run("Enabled by default", func(t *testing.T) {
		if !rule.Enabled() {
			t.Error("Rule should be enabled by default")
		}
	})
}

func TestEnglishOnlyRule_Enabled(t *testing.T) {
	rule := NewEnglishOnlyRule()

	t.Run("enabled by default", func(t *testing.T) {
		if !rule.Enabled() {
			t.Error("Rule should be enabled by default")
		}
	})

	t.Run("disable rule", func(t *testing.T) {
		rule.SetEnabled(false)
		if rule.Enabled() {
			t.Error("SetEnabled(false) did not disable rule")
		}
	})

	t.Run("re-enable rule", func(t *testing.T) {
		rule.SetEnabled(false)
		rule.SetEnabled(true)
		if !rule.Enabled() {
			t.Error("SetEnabled(true) did not re-enable rule")
		}
	})

	t.Run("disabled rule passes non-english", func(t *testing.T) {
		rule.SetEnabled(false)
		ctx := &CheckContext{Msg: "сервер запущен"}
		result := rule.Check(ctx)
		if !result.Passed {
			t.Error("Disabled rule should pass non-english message")
		}
	})
}

func TestEnglishOnlyRule_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		msg  string
	}{
		{"very long english", strings.Repeat("a", 10000)},
		{"very long cyrillic", strings.Repeat("а", 10000)},
		{"only digits", "123456789"},
		{"only punctuation", "!@#$%^&*()"},
		{"only whitespace", " \t\n\r "},
		{"only emoji", "😀😃😄😁"},
		{"surrogate pairs", "🀄🀅🀆"},
		{"mixed emoji text", "test😀test"},
		{"rtl text", "مرحبا"},
		{"ltr override", "\u202dserver\u202c"},
		{"bidirectional text", "hello مرحبا"},
		{"null byte", "test\x00"},
		{"control chars", "test\x01\x02\x03"},
		{"unicode replacement", "test"},
		{"private use area", "\ue000\uf8ff"},
		{"supplementary planes", "𐀀𐀁"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewEnglishOnlyRule()
			ctx := &CheckContext{Msg: tt.msg}
			result := rule.Check(ctx)

			if result == nil {
				t.Fatal("Check() returned nil")
			}

			// Просто проверяем, что не паникует
			_ = result.Passed
		})
	}
}

func TestCheckEnglishOnly_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	longEnglish := strings.Repeat("english text ", 1000)
	longMixed := strings.Repeat("text текст ", 500)

	t.Run("long english", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			CheckEnglishOnly(longEnglish)
		}
	})

	t.Run("long mixed", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			CheckEnglishOnly(longMixed)
		}
	})
}

func TestCheckEnglishOnly_UnicodeCategories(t *testing.T) {
	tests := []struct {
		name      string
		runes     []rune
		wantValid bool
	}{
		{"Lu (uppercase)", []rune{'A', 'B', 'Z'}, true},
		{"Ll (lowercase)", []rune{'a', 'b', 'z'}, true},
		{"Lt (titlecase)", []rune{'ǅ', 'ǈ', 'ǋ'}, true},
		{"Lm (modifier)", []rune{'ʰ', 'ʲ', 'ʷ'}, true},
		{"Lo (other latin)", []rune{'å', 'ø', 'ł'}, true},
		{"Cyrl (cyrillic)", []rune{'А', 'Б', 'Я'}, false},
		{"Han (chinese)", []rune{'中', '文', '字'}, false},
		{"Hira (hiragana)", []rune{'あ', 'い', 'う'}, false},
		{"Kata (katakana)", []rune{'ア', 'イ', 'ウ'}, false},
		{"Hang (hangul)", []rune{'가', '나', '다'}, false},
		{"Arab (arabic)", []rune{'ا', 'ب', 'ت'}, false},
		{"Hebr (hebrew)", []rune{'א', 'ב', 'ג'}, false},
		{"Grek (greek)", []rune{'Α', 'Β', 'Γ'}, false},
		{"Thai (thai)", []rune{'ก', 'ข', 'ค'}, false},
		{"Deva (devanagari)", []rune{'अ', 'आ', 'इ'}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, ch := range tt.runes {
				msg := string(ch)
				valid, _ := CheckEnglishOnly(msg)
				isLatin := unicode.Is(unicode.Latin, ch)

				if valid != tt.wantValid {
					t.Errorf("CheckEnglishOnly(%q U+%04X) valid = %v, want %v (IsLatin = %v)",
						msg, ch, valid, tt.wantValid, isLatin)
				}
			}
		})
	}
}

func TestEnglishOnlyRule_LookalikeCharacters(t *testing.T) {
	tests := []struct {
		name      string
		msg       string
		wantValid bool
		note      string
	}{
		{"latin a", "a", true, "U+0061"},
		{"cyrillic a", "а", false, "U+0430 — похожа на 'a'"},
		{"latin c", "c", true, "U+0063"},
		{"cyrillic c", "с", false, "U+0441 — похожа на 'c'"},
		{"latin e", "e", true, "U+0065"},
		{"cyrillic e", "е", false, "U+0435 — похожа на 'e'"},
		{"latin o", "o", true, "U+006F"},
		{"cyrillic o", "о", false, "U+043E — похожа на 'o'"},
		{"latin p", "p", true, "U+0070"},
		{"cyrillic p", "р", false, "U+0440 — похожа на 'p'"},
		{"latin x", "x", true, "U+0078"},
		{"cyrillic x", "х", false, "U+0445 — похожа на 'x'"},
		{"latin y", "y", true, "U+0079"},
		{"cyrillic y", "у", false, "U+0443 — похожа на 'y'"},
		{"mixed lookalikes", "server", true, "все латинские"},
		{"mixed lookalikes cyrillic", "ѕerver", false, "первая 'ѕ' — кириллическая"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, _ := CheckEnglishOnly(tt.msg)
			if valid != tt.wantValid {
				t.Errorf("CheckEnglishOnly(%q) valid = %v, want %v (note: %s)",
					tt.msg, valid, tt.wantValid, tt.note)
			}
		})
	}
}
