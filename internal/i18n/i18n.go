// Package i18n provides minimal, zero-dependency bilingual (English/Chinese)
// message lookup shared by hall-kit commands.
//
// Commands call SetLang once at startup (usually with Detect), then fetch
// every user-facing string through T.
package i18n

import (
	"fmt"
	"os"
	"strings"
)

// Lang is a supported language code.
type Lang string

const (
	// EN is English.
	EN Lang = "en"
	// ZH is Simplified Chinese.
	ZH Lang = "zh"
)

// tables holds message templates per language. Packages in this repository
// add their messages via register, so commands stay decoupled.
var tables = map[Lang]map[string]string{
	EN: {},
	ZH: {},
}

var current = EN

// register merges msgs into the global tables.
func register(msgs map[Lang]map[string]string) {
	for lang, entries := range msgs {
		for key, format := range entries {
			tables[lang][key] = format
		}
	}
}

// SetLang sets the active language for T.
func SetLang(l Lang) {
	current = l
}

// Current returns the active language.
func Current() Lang {
	return current
}

// Detect determines the language from an explicit --lang value, then the
// HALL_LANG environment variable, then LC_ALL/LANG. Values are matched by
// prefix ("zh*" -> ZH, "en*" -> EN); empty or unknown values are skipped.
// Falls back to EN.
func Detect(flagValue string) Lang {
	for _, v := range []string{flagValue, os.Getenv("HALL_LANG"), os.Getenv("LC_ALL"), os.Getenv("LANG")} {
		if l := parse(v); l != "" {
			return l
		}
	}
	return EN
}

func parse(s string) Lang {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.HasPrefix(s, "zh"):
		return ZH
	case strings.HasPrefix(s, "en"):
		return EN
	default:
		return ""
	}
}

// T returns the message for key in the active language, formatted with args.
// A missing translation falls back to English, then to the key itself.
func T(key string, args ...any) string {
	format, ok := tables[current][key]
	if !ok {
		format, ok = tables[EN][key]
	}
	if !ok {
		return key
	}
	return fmt.Sprintf(format, args...)
}
