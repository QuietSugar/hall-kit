package rmempty

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// Main runs the rmempty command with args, the arguments after the
// program name. It is shared by cmd/rmempty and "hall rmempty".
func Main(args []string) {
	i18n.SetLang(i18n.Detect(langFlag(args)))

	dir := ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--lang":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				fmt.Fprintln(os.Stderr, i18n.T("err.unknownflag", a))
				os.Exit(2)
			}
			i++ // value already consumed by langFlag
		case "-h", "--help":
			fmt.Println(i18n.T("usage"))
			return
		default:
			if strings.HasPrefix(a, "--lang=") {
				continue // value already consumed by langFlag
			}
			if strings.HasPrefix(a, "-") {
				fmt.Fprintln(os.Stderr, i18n.T("err.unknownflag", a))
				os.Exit(2)
			}
			if dir != "" {
				fmt.Fprintln(os.Stderr, i18n.T("err.usage"))
				os.Exit(2)
			}
			dir = a
		}
	}
	if dir == "" {
		dir = "."
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, i18n.T("err.abspath", dir))
		os.Exit(2)
	}
	if st, err := os.Stat(abs); err != nil || !st.IsDir() {
		fmt.Fprintln(os.Stderr, i18n.T("err.nodir", dir))
		os.Exit(2)
	}
	fmt.Print(i18n.T("info.target", abs))

	res, err := Clean(abs)
	if err != nil {
		fmt.Fprintln(os.Stderr, i18n.T("err.scan", err))
		os.Exit(1)
	}
	for _, d := range res.Removed {
		fmt.Println(i18n.T("removed", d))
	}
	fmt.Println(i18n.T("summary", len(res.Removed), res.DSStore))
}
