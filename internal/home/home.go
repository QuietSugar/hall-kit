// Package home resolves the user's home directory.
//
// On Windows os.UserHomeDir only consults USERPROFILE, while Cygwin and
// git-bash keep the per-shell home in HOME and hand native programs a
// POSIX-ish path (/c/Users/x, /cygdrive/c/Users/x or, under Cygwin, a mount
// path such as /home/xu). Honouring HOME keeps hall's state next to the rc
// files those shells actually source; when HOME has no Windows equivalent the
// native HOMEDRIVE/HOMEPATH pair is used instead.
package home

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Dir returns the user's home directory.
func Dir() (string, error) {
	if h := strings.TrimSpace(os.Getenv("HOME")); h != "" {
		if runtime.GOOS != "windows" {
			return h, nil
		}
		if p, ok := windowsPath(h); ok {
			return p, nil
		}
	}
	if runtime.GOOS == "windows" {
		if d, p := os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"); d != "" && p != "" {
			return d + p, nil
		}
	}
	return os.UserHomeDir()
}

// windowsPath converts the POSIX-ish paths Cygwin and git-bash export into a
// Windows path. ok is false when p has no Windows equivalent (for example a
// Cygwin mount path like /home/xu).
func windowsPath(p string) (string, bool) {
	p = filepath.ToSlash(p)
	switch {
	case strings.HasPrefix(p, "//"):
		return toBackslash(p), true
	case strings.HasPrefix(p, "/cygdrive/"):
		return drivePath(strings.TrimPrefix(p, "/cygdrive/")), true
	case len(p) >= 3 && p[0] == '/' && p[2] == '/' && isAlpha(p[1]):
		return drivePath(p[1:]), true
	case len(p) >= 2 && p[1] == ':' && isAlpha(p[0]):
		return drivePath(p), true
	}
	return "", false
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// toBackslash is filepath.FromSlash with Windows semantics, independent of the
// host the tests run on.
func toBackslash(p string) string {
	return strings.ReplaceAll(p, "/", "\\")
}

// drivePath converts "c/Users/x" or "c:/Users/x" into "C:\\Users\\x".
func drivePath(rest string) string {
	rest = strings.TrimPrefix(rest, "/")
	if rest == "" {
		return ""
	}
	drive := strings.ToUpper(rest[:1])
	tail := strings.TrimPrefix(rest[1:], ":")
	return drive + ":" + toBackslash(tail)
}
