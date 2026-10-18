// Package winpath renders Windows paths in the form the POSIX shell hall is
// running under can use (git-bash /c/... or Cygwin /cygdrive/c/...).
package winpath

import (
	"os"
	"runtime"
	"strings"
)

// Shell reports which POSIX-on-Windows shell hall is running under:
// "git-bash", "cygwin", or "" when neither is detected (native cmd/PowerShell
// or a non-Windows host).
func Shell() string {
	if runtime.GOOS != "windows" {
		return ""
	}
	if os.Getenv("MSYSTEM") != "" {
		return "git-bash"
	}
	if strings.Contains(strings.ToLower(os.Getenv("OSTYPE")), "cygwin") || os.Getenv("CYGWIN") != "" {
		return "cygwin"
	}
	return ""
}

// POSIX converts a Windows path with a drive letter into the POSIX form used
// by shell. It returns "" when shell is unknown or p has no drive letter.
func POSIX(p, shell string) string {
	q := strings.ReplaceAll(p, "\\", "/")
	if len(q) < 2 || q[1] != ':' || !isAlpha(q[0]) {
		return ""
	}
	drive := strings.ToLower(q[:1])
	tail := q[2:]
	switch shell {
	case "git-bash":
		return "/" + drive + tail
	case "cygwin":
		return "/cygdrive/" + drive + tail
	}
	return ""
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
