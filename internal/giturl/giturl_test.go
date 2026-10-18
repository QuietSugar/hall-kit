package giturl

import (
	"path/filepath"
	"testing"
)

func TestValid(t *testing.T) {
	valid := []string{
		"https://github.com/QuietSugar/hall-kit.git",
		"https://github.com/QuietSugar/hall-kit",
		"http://1.2.3.4:8080/a/b/c.git",
		"https://gh-proxy.com/https://github.com/user/repo.git",
		"git@github.com:user/repo.git",
		"git@1.2.3.4:a/b/c.git",
		"ssh-user@example.com:2222/path/repo.git",
		"https://example.com",
		"https://example.com/.git",
	}
	for _, u := range valid {
		if !Valid(u) {
			t.Errorf("Valid(%q) = false, want true", u)
		}
	}

	invalid := []string{
		"",
		"github.com/user/repo",
		"ftp://example.com/x.git",
		"https://example.com/",
		"git@github.com",
		"https://ex ample.com/x",
		"https://host/../../../etc/foo.git",
		"git@host:../../../foo.git",
		"https://..",
		"https://host/a/./b",
		"https://host//a",
	}
	for _, u := range invalid {
		if Valid(u) {
			t.Errorf("Valid(%q) = true, want false", u)
		}
	}
}

func TestSafeName(t *testing.T) {
	cases := map[string]string{
		"https://github.com/QuietSugar/hall-kit.git":            "github.com/QuietSugar/hall-kit",
		"http://1.2.3.4:8080/a/b/c.git":                         "1.2.3.4/8080/a/b/c",
		"git@1.2.3.4:a/b/c.git":                                 "1.2.3.4/a/b/c",
		"git@github.com:user/repo.git":                          "github.com/user/repo",
		"https://gh-proxy.com/https://github.com/user/repo.git": "github.com/user/repo",
		"  https://github.com/user/repo.git  ":                  "github.com/user/repo",
		"https://host/.git":                                     "host",
	}
	for in, want := range cases {
		if got := SafeName(in); got != want {
			t.Errorf("SafeName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSplit(t *testing.T) {
	parent, name := Split("git@1.2.3.4:a/b/c.git")
	if parent != "1.2.3.4/a/b" || name != "c" {
		t.Errorf("Split = (%q, %q), want (%q, %q)", parent, name, "1.2.3.4/a/b", "c")
	}

	parent, name = Split("https://github.com")
	if parent != "." || name != "github.com" {
		t.Errorf("Split = (%q, %q), want (%q, %q)", parent, name, ".", "github.com")
	}
}

func TestTarget(t *testing.T) {
	root := filepath.Join("root", "archive")

	got, err := Target(root, "git@1.2.3.4:a/b/c.git")
	if err != nil {
		t.Fatalf("Target returned error: %v", err)
	}
	if want := filepath.Join(root, "1.2.3.4", "a", "b", "c"); got != want {
		t.Errorf("Target = %q, want %q", got, want)
	}

	got, err = Target(root, "https://github.com")
	if err != nil {
		t.Fatalf("Target returned error: %v", err)
	}
	if want := filepath.Join(root, "github.com"); got != want {
		t.Errorf("Target = %q, want %q", got, want)
	}
}

func TestTargetRejectsEscape(t *testing.T) {
	root := filepath.Join("root", "archive")
	for _, raw := range []string{"https://../x", "git@host:../../x.git"} {
		if got, err := Target(root, raw); err == nil {
			t.Errorf("Target(%q) = %q, want error", raw, got)
		}
	}
}
