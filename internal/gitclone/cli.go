// Package gitclone implements the gitclone command, shared by the
// cmd/gitclone shim and by "hall gitclone".
package gitclone

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/QuietSugar/hall-kit/internal/gitenv"
	"github.com/QuietSugar/hall-kit/internal/giturl"
	"github.com/QuietSugar/hall-kit/internal/i18n"
	"github.com/QuietSugar/hall-kit/internal/winpath"
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

// Main runs the gitclone command with args, the arguments after the
// program name. It is shared by cmd/gitclone and "hall gitclone".
func Main(args []string) {
	i18n.SetLang(i18n.Detect(langFlag(args)))

	var repoURL string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--lang":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				fmt.Fprintln(os.Stderr, i18n.T("gitclone.err.unknownflag", a))
				os.Exit(2)
			}
			i++ // value already consumed by langFlag
		case "-h", "--help":
			fmt.Println(i18n.T("gitclone.usage"))
			return
		default:
			if strings.HasPrefix(a, "--lang=") {
				continue // value already consumed by langFlag
			}
			if strings.HasPrefix(a, "-") {
				fmt.Fprintln(os.Stderr, i18n.T("gitclone.err.unknownflag", a))
				os.Exit(2)
			}
			if repoURL != "" {
				fmt.Fprintln(os.Stderr, i18n.T("gitclone.err.usage"))
				os.Exit(2)
			}
			repoURL = a
		}
	}

	if repoURL == "" {
		fmt.Fprintln(os.Stderr, i18n.T("gitclone.usage"))
		os.Exit(2)
	}
	if !giturl.Valid(repoURL) {
		fmt.Fprintln(os.Stderr, i18n.T("gitclone.err.invalid", repoURL))
		os.Exit(2)
	}
	if _, err := exec.LookPath("git"); err != nil {
		fmt.Fprintln(os.Stderr, i18n.T("gitclone.err.nogit"))
		os.Exit(2)
	}

	root, err := gitenv.RepoRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, i18n.T("gitclone.err.root", err))
		os.Exit(2)
	}
	target, err := giturl.Target(root, repoURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, i18n.T("gitclone.err.invalid", repoURL))
		os.Exit(2)
	}
	// Keep the output short: the remote and the final target. On Windows the
	// target is also shown in the POSIX form of the shell hall runs under
	// (git-bash /c/..., Cygwin /cygdrive/c/...).
	fmt.Println(i18n.T("gitclone.info.url", repoURL))
	fmt.Println(i18n.T("gitclone.info.target", target))
	if shell := winpath.Shell(); shell != "" {
		if posix := winpath.POSIX(target, shell); posix != "" {
			fmt.Println(i18n.T("gitclone.info.posix", shell, posix))
		}
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, i18n.T("gitclone.err.mkdir", filepath.Dir(target), err))
		os.Exit(1)
	}
	if _, err := os.Stat(target); err == nil {
		fmt.Fprintln(os.Stderr, i18n.T("gitclone.warn.exists", target))
		return
	}

	cmd := exec.Command("git", "clone", repoURL, target)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, i18n.T("gitclone.err.clone", err))
		os.Exit(1)
	}
}
