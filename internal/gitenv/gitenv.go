// Package gitenv resolves the archive root shared by gitclone and gitarc.
//
// It stays compatible with the shell implementation: HALL_GIT_REPO_PATH wins,
// then HALL_GIT_REPO_PATH from ~/.hall/env (or the legacy
// ~/.hall-command/env), then ~/git-repo. The selected directory is created
// when missing.
package gitenv

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	userhome "github.com/QuietSugar/hall-kit/internal/home"
)

// ConfigEnv overrides the configuration file path.
const ConfigEnv = "HALL_CONFIG_FILE"

// RepoPathEnv is the environment variable that overrides the archive root.
const RepoPathEnv = "HALL_GIT_REPO_PATH"

// DefaultDir is the archive root used when nothing else is configured.
const DefaultDir = "git-repo"

// homeVarRE matches a bare $HOME that is not followed by an identifier
// character, so $HOMEDIR is left alone.
var homeVarRE = regexp.MustCompile(`\$HOME\b`)

// ConfigFile returns the config file to read. HALL_CONFIG_FILE overrides,
// then ~/.hall/env, then the legacy ~/.hall-command/env while it still
// exists (temporary migration fallback).
func ConfigFile() string {
	if p := strings.TrimSpace(os.Getenv(ConfigEnv)); p != "" {
		return Expand(p)
	}
	home, err := userhome.Dir()
	if err != nil {
		return ""
	}
	current := filepath.Join(home, ".hall", "env")
	if _, err := os.Stat(current); err == nil {
		return current
	}
	legacy := filepath.Join(home, ".hall-command", "env")
	if _, err := os.Stat(legacy); err == nil {
		return legacy
	}
	return current
}

// Lookup returns the value of key from ~/.hall-command/env. The parser accepts
// the shell assignment syntax the file historically used: an optional leading
// "export", optional surrounding single or double quotes and a trailing
// comment. Values may reference ~, $HOME and ${HOME}. When the key appears
// more than once the last assignment wins, as in the shell. A missing file or
// key yields "".
func Lookup(key string) string {
	file := ConfigFile()
	if file == "" {
		return ""
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	var value string
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if rest, ok := strings.CutPrefix(line, "export"); ok && (rest == "" || rest[0] == ' ' || rest[0] == '\t') {
			line = strings.TrimSpace(rest)
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(k) != key {
			continue
		}
		value = Expand(unquote(strings.TrimSpace(v)))
	}
	return value
}

// unquote removes one pair of surrounding quotes and strips an unquoted
// trailing comment.
func unquote(v string) string {
	if len(v) >= 2 {
		first, last := v[0], v[len(v)-1]
		// 0x22 is the double quote, 0x27 the single quote.
		if (first == 0x22 && last == 0x22) || (first == 0x27 && last == 0x27) {
			return v[1 : len(v)-1]
		}
	}
	if i := strings.Index(v, " #"); i >= 0 {
		v = strings.TrimSpace(v[:i])
	}
	return v
}

// RepoRoot returns the archive root, creating it when needed.
func RepoRoot() (string, error) {
	dir := strings.TrimSpace(os.Getenv(RepoPathEnv))
	if dir == "" {
		dir = Lookup(RepoPathEnv)
	}
	if dir == "" {
		home, err := userhome.Dir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, DefaultDir)
	}
	abs, err := filepath.Abs(Expand(dir))
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return "", err
	}
	return abs, nil
}

// Expand replaces a leading ~ and ${HOME} / $HOME references with the user's
// home directory. A bare $HOME is only expanded when it is not the prefix of a
// longer identifier such as $HOMEDIR.
func Expand(s string) string {
	home, err := userhome.Dir()
	if err != nil {
		return s
	}
	if s == "~" {
		return home
	}
	if strings.HasPrefix(s, "~/") {
		s = filepath.Join(home, s[2:])
	}
	s = strings.ReplaceAll(s, "${HOME}", home)
	s = homeVarRE.ReplaceAllLiteralString(s, home)
	return s
}
