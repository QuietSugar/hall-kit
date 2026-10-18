package home

import (
	"runtime"
	"testing"
)

func TestWindowsPath(t *testing.T) {
	ok := map[string]string{
		"/c/Users/x":          `C:\Users\x`,
		"/cygdrive/c/Users/x": `C:\Users\x`,
		"C:/Users/x":          `C:\Users\x`,
		`C:\Users\x`:          `C:\Users\x`,
		"c:/tmp":              `C:\tmp`,
		"//server/share":      `\\server\share`,
	}
	for in, want := range ok {
		got, valid := windowsPath(in)
		if !valid {
			t.Errorf("windowsPath(%q) not convertible, want %q", in, want)
			continue
		}
		if got != want {
			t.Errorf("windowsPath(%q) = %q, want %q", in, got, want)
		}
	}
	for _, in := range []string{"/home/xu", "relative/path", ""} {
		if got, valid := windowsPath(in); valid {
			t.Errorf("windowsPath(%q) = %q, want not convertible", in, got)
		}
	}
}

func TestDirPrefersHome(t *testing.T) {
	t.Setenv("HOME", "/c/Users/example")
	got, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	want := "/c/Users/example"
	if runtime.GOOS == "windows" {
		want = `C:\Users\example`
	}
	if got != want {
		t.Errorf("Dir() = %q, want %q", got, want)
	}
}
