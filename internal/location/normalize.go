package location

import (
	"strings"
	"unicode"
)

// Normalize prepares user input for matching: lowercase, trim, Cyrillic→Latin, apostrophe normalization.
func Normalize(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = strings.ToLower(s)
	s = cyrillicToLatin(s)
	s = normalizeApostrophe(s)
	return s
}

func normalizeApostrophe(s string) string {
	// Unicode apostrophe and backtick → ASCII single quote for consistent matching
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\'', '`', 0x2018, 0x2019, 0x02BC: // ASCII, backtick, left/right single quote, modifier letter apostrophe
			b.WriteByte('\'')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// cyrillicToLatin maps Uzbek Cyrillic to Latin (common equivalents for city names).
var cyrillicToLatin = func() func(string) string {
	m := map[rune]string{
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "yo", 'ж': "j", 'з': "z",
		'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n", 'о': "o", 'п': "p", 'р': "r",
		'с': "s", 'т': "t", 'у': "u", 'ф': "f", 'х': "x", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "sh",
		'ъ': "", 'ы': "i", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
		'А': "a", 'Б': "b", 'В': "v", 'Г': "g", 'Д': "d", 'Е': "e", 'Ё': "yo", 'Ж': "j", 'З': "z",
		'И': "i", 'Й': "y", 'К': "k", 'Л': "l", 'М': "m", 'Н': "n", 'О': "o", 'П': "p", 'Р': "r",
		'С': "s", 'Т': "t", 'У': "u", 'Ф': "f", 'Х': "x", 'Ц': "ts", 'Ч': "ch", 'Ш': "sh", 'Щ': "sh",
		'Ъ': "", 'Ы': "i", 'Ь': "", 'Э': "e", 'Ю': "yu", 'Я': "ya",
		// Uzbek-specific
		'\u04E3': "g'", '\u04E9': "o'", '\u049B': "q", '\u04B3': "h", '\u04AF': "u",
		'\u04E2': "g'", '\u04E8': "o'", '\u049A': "q", '\u04B2': "h", '\u04AE': "u",
	}
	return func(s string) string {
		var b strings.Builder
		for _, r := range s {
			if repl, ok := m[r]; ok {
				b.WriteString(repl)
			} else {
				b.WriteRune(unicode.ToLower(r))
			}
		}
		return b.String()
	}
}()
