package dockerage

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/QuietSugar/hall-kit/internal/i18n"
)

// langFlag pre-scans --lang (accepting both "--lang zh" and "--lang=zh") so
// that usage and error messages are printed in the requested language.
func langFlag(args []string) string {
	for i, a := range args {
		if a == "--lang" && i+1 < len(args) {
			return args[i+1]
		}
		if v, ok := strings.CutPrefix(a, "--lang="); ok {
			return v
		}
	}
	return ""
}

// Main runs the dockerage command with args, the arguments after the
// program name. It is shared by cmd/dockerage and "hall dockerage".
func Main(args []string) {
	i18n.SetLang(i18n.Detect(langFlag(args)))

	var input string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--lang":
			i++ // value already consumed by langFlag
		case "-h", "--help":
			fmt.Print(i18n.T("usage"))
			return
		default:
			if strings.HasPrefix(a, "--lang=") {
				continue // value already consumed by langFlag
			}
			if strings.HasPrefix(a, "-") {
				fmt.Fprintln(os.Stderr, i18n.T("error.unknownflag", a))
				os.Exit(2)
			}
			if input != "" {
				fmt.Fprintln(os.Stderr, i18n.T("error.toomany", a))
				os.Exit(2)
			}
			input = a
		}
	}

	if input == "" {
		fmt.Fprint(os.Stderr, i18n.T("usage"))
		os.Exit(2)
	}

	t, err := Parse(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, i18n.T("error.badtime", err))
		os.Exit(2)
	}

	now := time.Now()
	seconds, err := SecondsSince(t, now)
	if err != nil {
		fmt.Fprintln(os.Stderr, i18n.T("error.future", t.Format(time.RFC3339), now.Format(time.RFC3339)))
		os.Exit(2)
	}
	fmt.Println(seconds)
}
