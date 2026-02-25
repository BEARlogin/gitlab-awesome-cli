package keymap

// ruToEn maps Russian keyboard layout characters to their Latin equivalents.
var ruToEn = map[string]string{
	"й": "q", "ц": "w", "у": "e", "к": "r", "е": "t",
	"н": "y", "г": "u", "ш": "i", "щ": "o", "з": "p",
	"х": "[", "ъ": "]",
	"ф": "a", "ы": "s", "в": "d", "а": "f", "п": "g",
	"р": "h", "о": "j", "л": "k", "д": "l", "ж": ";",
	"э": "'",
	"я": "z", "ч": "x", "с": "c", "м": "v", "и": "b",
	"т": "n", "ь": "m", "б": ",", "ю": ".",
	"Й": "Q", "Ц": "W", "У": "E", "К": "R", "Е": "T",
	"Н": "Y", "Г": "U", "Ш": "I", "Щ": "O", "З": "P",
	"Х": "{", "Ъ": "}",
	"Ф": "A", "Ы": "S", "В": "D", "А": "F", "П": "G",
	"Р": "H", "О": "J", "Л": "K", "Д": "L", "Ж": ":",
	"Э": "\"",
	"Я": "Z", "Ч": "X", "С": "C", "М": "V", "И": "B",
	"Т": "N", "Ь": "M",
}

// Normalize converts a key string from Russian layout to English equivalent.
// If the key is already Latin or a special key, returns it unchanged.
func Normalize(key string) string {
	if en, ok := ruToEn[key]; ok {
		return en
	}
	return key
}
