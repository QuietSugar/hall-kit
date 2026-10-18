package shellinstall

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newOpts(t *testing.T, home string) *options {
	t.Helper()
	out := &bytes.Buffer{}
	return &options{home: home, out: out, errOut: out}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestInstallIdempotent(t *testing.T) {
	home := t.TempDir()
	o := newOpts(t, home)
	rc := filepath.Join(home, ".zshrc")
	o.rcFiles = []string{rc}

	if err := o.install(); err != nil {
		t.Fatal(err)
	}
	files, err := managedFiles()
	if err != nil {
		t.Fatal(err)
	}
	for _, rel := range files {
		p := filepath.Join(home, ".hall", filepath.FromSlash(rel))
		if _, err := os.Stat(p); err != nil {
			t.Errorf("missing installed asset %s: %v", p, err)
		}
	}
	if _, err := os.Stat(filepath.Join(home, ".hall", "env")); err != nil {
		t.Errorf("missing env: %v", err)
	}

	first := mustRead(t, rc)
	if n := strings.Count(first, blockStart); n != 1 {
		t.Fatalf("block count = %d, want 1\n%s", n, first)
	}
	if !strings.Contains(first, "$HOME/.hall/shell/source") {
		t.Errorf("block should be relocatable via $HOME:\n%s", first)
	}
	if !strings.Contains(first, "$HOME/.hall/bin") {
		t.Errorf("block should add $HOME/.hall/bin to PATH:\n%s", first)
	}
	if st, err := os.Stat(filepath.Join(home, ".hall", "bin", "tat")); err != nil || st.Mode().Perm()&0o111 == 0 {
		t.Errorf("bin/tat missing or not executable: %v", err)
	}

	if err := o.install(); err != nil {
		t.Fatal(err)
	}
	second := mustRead(t, rc)
	if n := strings.Count(second, blockStart); n != 1 {
		t.Errorf("after reinstall block count = %d, want 1", n)
	}
	if first != second {
		t.Errorf("reinstall changed the rc file")
	}
}

func TestDryRunWritesNothing(t *testing.T) {
	home := t.TempDir()
	o := newOpts(t, home)
	o.dryRun = true
	rc := filepath.Join(home, ".zshrc")
	o.rcFiles = []string{rc}

	if err := o.install(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, ".hall")); !os.IsNotExist(err) {
		t.Errorf("dry-run created ~/.hall")
	}
	if _, err := os.Stat(rc); !os.IsNotExist(err) {
		t.Errorf("dry-run created rc file")
	}
}

func TestInstallNoRC(t *testing.T) {
	home := t.TempDir()
	o := newOpts(t, home)
	o.noRC = true
	if err := o.install(); err != nil {
		t.Fatal(err)
	}
	for _, rc := range []string{".zshrc", ".bashrc", ".profile"} {
		if _, err := os.Stat(filepath.Join(home, rc)); !os.IsNotExist(err) {
			t.Errorf("--no-rc touched %s", rc)
		}
	}
	if _, err := os.Stat(filepath.Join(home, ".hall", "shell", "source", "hall-proxy.sh")); err != nil {
		t.Errorf("assets were not installed: %v", err)
	}
}

func TestUninstall(t *testing.T) {
	home := t.TempDir()
	o := newOpts(t, home)
	rc := filepath.Join(home, ".zshrc")
	o.rcFiles = []string{rc}
	if err := o.install(); err != nil {
		t.Fatal(err)
	}

	if err := o.uninstall(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, ".hall", "shell")); !os.IsNotExist(err) {
		t.Errorf("shell dir still present")
	}
	if _, err := os.Stat(filepath.Join(home, ".hall", "env")); err != nil {
		t.Errorf("env should survive uninstall: %v", err)
	}
	if strings.Contains(mustRead(t, rc), blockStart) {
		t.Errorf("rc block not removed")
	}

	if err := o.install(); err != nil {
		t.Fatal(err)
	}
	o.purge = true
	if err := o.uninstall(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, ".hall")); !os.IsNotExist(err) {
		t.Errorf("--purge left ~/.hall")
	}
}

func TestMigrateLegacyConfigAndBlocks(t *testing.T) {
	home := t.TempDir()
	legacyEnv := "HALL_GIT_REPO_PATH=/legacy\n"
	mustWrite(t, filepath.Join(home, ".hall-command", "env"), legacyEnv)

	rc := filepath.Join(home, ".zshrc")
	mustWrite(t, rc, strings.Join([]string{
		"# HALL_COMMAND_PATH_START",
		"# 由 install.sh 自动生成，请勿手动修改",
		"export PATH=\"$HOME/.hall-command/command:$PATH\"",
		"# HALL_COMMAND_PATH_END",
		"",
		"alias ll='ls -l'",
		"",
	}, "\n"))

	o := newOpts(t, home)
	o.rcFiles = []string{rc}
	if err := o.install(); err != nil {
		t.Fatal(err)
	}

	if got := mustRead(t, filepath.Join(home, ".hall", "env")); got != legacyEnv {
		t.Errorf("migrated env = %q, want %q", got, legacyEnv)
	}
	content := mustRead(t, rc)
	if strings.Contains(content, legacyPathStart) {
		t.Errorf("legacy block not cleaned:\n%s", content)
	}
	if strings.Count(content, blockStart) != 1 {
		t.Errorf("managed block missing")
	}
	if !strings.Contains(content, "alias ll='ls -l'") {
		t.Errorf("unrelated rc content was lost:\n%s", content)
	}
}

func TestMarkerMismatchRefuses(t *testing.T) {
	home := t.TempDir()
	rc := filepath.Join(home, ".zshrc")
	broken := "# kept\n" + blockStart + "\nif true; then :; fi\n"
	mustWrite(t, rc, broken)

	o := newOpts(t, home)
	o.rcFiles = []string{rc}
	if err := o.install(); err == nil {
		t.Fatal("expected an error for an unmatched marker")
	}
	if got := mustRead(t, rc); got != broken {
		t.Errorf("rc file was modified despite the error:\n%s", got)
	}
}

func TestMigratePurgeBacksUpLegacy(t *testing.T) {
	home := t.TempDir()
	mustWrite(t, filepath.Join(home, ".hall-command", "source", "mine.sh"), "# mine\n")

	o := newOpts(t, home)
	o.purge = true
	if err := o.migrate(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, ".hall-command")); !os.IsNotExist(err) {
		t.Errorf("legacy dir should have been moved away")
	}
	matches, err := filepath.Glob(filepath.Join(home, ".hall-command.bak.*"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("backup not found: %v %v", matches, err)
	}
	if _, err := os.Stat(filepath.Join(matches[0], "source", "mine.sh")); err != nil {
		t.Errorf("user file lost in backup: %v", err)
	}
}

func TestStatusReportsLegacy(t *testing.T) {
	home := t.TempDir()
	mustWrite(t, filepath.Join(home, ".hall-command", "env"), "PROXY_ADDR=\n")
	o := newOpts(t, home)
	o.rcFiles = []string{filepath.Join(home, ".zshrc")}
	o.status()
	if !strings.Contains(o.out.(*bytes.Buffer).String(), "legacy:") {
		t.Errorf("status did not report legacy state")
	}
}

func TestUserFilesSurvive(t *testing.T) {
	home := t.TempDir()
	o := newOpts(t, home)
	rc := filepath.Join(home, ".zshrc")
	o.rcFiles = []string{rc}

	userFile := filepath.Join(home, ".hall", "shell", "source", "mine.sh")
	mustWrite(t, userFile, "# user snippet\n")

	if err := o.install(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(userFile); err != nil {
		t.Fatalf("install touched the user file: %v", err)
	}
	if err := o.uninstall(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(userFile); err != nil {
		t.Errorf("uninstall removed the user file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".hall", "shell", "source", "hall-proxy.sh")); !os.IsNotExist(err) {
		t.Errorf("managed file was not removed")
	}
}

func TestDryRunAfterInstallIsQuiet(t *testing.T) {
	home := t.TempDir()
	o := newOpts(t, home)
	o.rcFiles = []string{filepath.Join(home, ".zshrc")}
	if err := o.install(); err != nil {
		t.Fatal(err)
	}
	o.dryRun = true
	o.out.(*bytes.Buffer).Reset()
	if err := o.install(); err != nil {
		t.Fatal(err)
	}
	if out := o.out.(*bytes.Buffer).String(); strings.Contains(out, "would write") {
		t.Errorf("unchanged dry-run should not report writes:\n%s", out)
	}
}
