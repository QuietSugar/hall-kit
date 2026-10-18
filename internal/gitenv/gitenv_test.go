package gitenv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupHome points the home directory at a temporary directory and clears
// HALL_GIT_REPO_PATH.
func setupHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv(RepoPathEnv, "")
	return home
}

func writeEnv(t *testing.T, home, content string) {
	t.Helper()
	dir := filepath.Join(home, ".hall-command")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "env"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestExpand(t *testing.T) {
	home := setupHome(t)
	cases := map[string]string{
		"~":              home,
		"~/git-repo":     filepath.Join(home, "git-repo"),
		"${HOME}/a":      filepath.Join(home, "a"),
		"$HOME/a":        filepath.Join(home, "a"),
		"$HOMEDIR/a":     "$HOMEDIR/a",
		"/absolute/path": "/absolute/path",
	}
	for in, want := range cases {
		if got := Expand(in); got != want {
			t.Errorf("Expand(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLookup(t *testing.T) {
	home := setupHome(t)
	writeEnv(t, home, "# comment\n\nHALL_GIT_REPO_PATH=${HOME}/from-file\nOTHER = 1\n")

	if got, want := Lookup(RepoPathEnv), filepath.Join(home, "from-file"); got != want {
		t.Errorf("Lookup(%q) = %q, want %q", RepoPathEnv, got, want)
	}
	if got := Lookup("OTHER"); got != "1" {
		t.Errorf("Lookup(OTHER) = %q, want %q", got, "1")
	}
	if got := Lookup("MISSING"); got != "" {
		t.Errorf("Lookup(MISSING) = %q, want %q", got, "")
	}
}

func TestLookupLastKeyWins(t *testing.T) {
	home := setupHome(t)
	writeEnv(t, home, "HALL_GIT_REPO_PATH=/first\nHALL_GIT_REPO_PATH=/second\n")
	if got := Lookup(RepoPathEnv); got != "/second" {
		t.Errorf("Lookup = %q, want %q", got, "/second")
	}
}

func TestLookupShellSyntax(t *testing.T) {
	home := setupHome(t)
	content := strings.Join([]string{
		"# comment",
		`export HALL_GIT_REPO_PATH="${HOME}/exported"`,
		"HALL_OTHER=plain  # trailing comment",
		"QUOTED='single value'",
		"",
	}, "\n")
	writeEnv(t, home, content)

	if got, want := Lookup(RepoPathEnv), filepath.Join(home, "exported"); got != want {
		t.Errorf("Lookup(%q) = %q, want %q", RepoPathEnv, got, want)
	}
	if got := Lookup("HALL_OTHER"); got != "plain" {
		t.Errorf("Lookup(HALL_OTHER) = %q, want %q", got, "plain")
	}
	if got := Lookup("QUOTED"); got != "single value" {
		t.Errorf("Lookup(QUOTED) = %q, want %q", got, "single value")
	}
}

func TestRepoRootDefault(t *testing.T) {
	home := setupHome(t)
	got, err := RepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, DefaultDir); got != want {
		t.Errorf("RepoRoot = %q, want %q", got, want)
	}
	if st, err := os.Stat(got); err != nil || !st.IsDir() {
		t.Errorf("RepoRoot did not create the directory: %v", err)
	}
}

func TestRepoRootEnvWins(t *testing.T) {
	home := setupHome(t)
	custom := filepath.Join(home, "custom")
	t.Setenv(RepoPathEnv, custom)
	got, err := RepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	if got != custom {
		t.Errorf("RepoRoot = %q, want %q", got, custom)
	}
}

func TestRepoRootFromConfig(t *testing.T) {
	home := setupHome(t)
	writeEnv(t, home, "HALL_GIT_REPO_PATH=$HOME/from-file\n")
	got, err := RepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, "from-file"); got != want {
		t.Errorf("RepoRoot = %q, want %q", got, want)
	}
}

func TestConfigFilePrecedence(t *testing.T) {
	home := setupHome(t)
	writeEnv(t, home, "HALL_GIT_REPO_PATH=/legacy\n")
	if got, want := Lookup(RepoPathEnv), "/legacy"; got != want {
		t.Errorf("legacy fallback: Lookup = %q, want %q", got, want)
	}

	dir := filepath.Join(home, ".hall")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "env"), []byte("HALL_GIT_REPO_PATH=/new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, want := Lookup(RepoPathEnv), "/new"; got != want {
		t.Errorf("new path: Lookup = %q, want %q", got, want)
	}

	alt := filepath.Join(home, "alt.env")
	if err := os.WriteFile(alt, []byte("HALL_GIT_REPO_PATH=/alt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(ConfigEnv, alt)
	if got, want := Lookup(RepoPathEnv), "/alt"; got != want {
		t.Errorf("override: Lookup = %q, want %q", got, want)
	}
}
