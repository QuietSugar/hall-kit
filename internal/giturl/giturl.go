// Package giturl validates git remote URLs and turns them into the relative
// directory layout shared by gitclone and gitarc.
package giturl

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// ErrInvalid reports that a string is not an accepted, safe HTTP(S) or SSH
// git URL.
var ErrInvalid = errors.New("invalid git URL")

// proxyPrefix is stripped before validating or naming, mirroring the shell
// helper make_filename_safe.
const proxyPrefix = "https://gh-proxy.com/"

// httpRE and sshRE mirror the validation used by the original shell helper
// is_git_url_https_ssh: a host (name or IPv4), an optional port, optional
// path segments and an optional .git suffix. [.] is used instead of \\. so
// the pattern stays readable in a raw string literal.
var (
	httpRE = regexp.MustCompile(`^(https?)://([a-zA-Z0-9.-]+|([0-9]{1,3}[.]){3}[0-9]{1,3})(:[0-9]{1,5})?(/[a-zA-Z0-9._/-]+)*([.]git)?$`)
	sshRE  = regexp.MustCompile(`^[a-zA-Z0-9._-]+@([a-zA-Z0-9.-]+|([0-9]{1,3}[.]){3}[0-9]{1,3})(:[0-9]{1,5})?:[a-zA-Z0-9._/-]+([.]git)?$`)
)

// Valid reports whether raw is an accepted git URL that maps safely below the
// archive root. A https://gh-proxy.com/ prefix is tolerated. The host itself
// may be a name or an IPv4 address. URLs containing empty, "." or ".." path
// segments are rejected so they cannot escape the archive root.
func Valid(raw string) bool {
	s := strings.TrimPrefix(strings.TrimSpace(raw), proxyPrefix)
	if !httpRE.MatchString(s) && !sshRE.MatchString(s) {
		return false
	}
	return !escapes(SafeName(raw))
}

// escapes reports whether a slash-separated relative path contains an empty,
// "." or ".." segment.
func escapes(rel string) bool {
	for _, seg := range strings.Split(rel, "/") {
		if seg == "" || seg == "." || seg == ".." {
			return true
		}
	}
	return false
}

// SafeName converts a git URL into a slash-separated relative path, e.g.
//
//	http://1.2.3.4:8080/a/b/c.git -> 1.2.3.4/8080/a/b/c
//	git@1.2.3.4:a/b/c.git         -> 1.2.3.4/a/b/c
func SafeName(raw string) string {
	s := strings.TrimPrefix(strings.TrimSpace(raw), proxyPrefix)
	for _, prefix := range []string{"https://", "http://", "git@"} {
		s = strings.TrimPrefix(s, prefix)
	}
	s = strings.ReplaceAll(s, ":", "/")
	s = strings.TrimSuffix(s, ".git")
	return strings.TrimSuffix(s, "/")
}

// Split returns the relative parent directory and the project name a URL maps
// to. The parent is "." when the URL carries no path component.
func Split(raw string) (parent, name string) {
	rel := SafeName(raw)
	return path.Dir(rel), path.Base(rel)
}

// Target returns the local absolute path a URL maps to under root. It returns
// an error wrapping ErrInvalid when the resulting path would fall outside
// root, as a defence in depth on top of Valid.
func Target(root, raw string) (string, error) {
	parent, name := Split(raw)
	target := filepath.Join(root, name)
	if parent != "." {
		target = filepath.Join(root, filepath.FromSlash(parent), name)
	}
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("%w: %q escapes the archive root", ErrInvalid, raw)
	}
	return target, nil
}
