package winpath

import "testing"

func TestPOSIX(t *testing.T) {
	cases := []struct {
		in, shell, want string
	}{
		{`C:\Users\x`, "git-bash", "/c/Users/x"},
		{`C:\Users\x`, "cygwin", "/cygdrive/c/Users/x"},
		{`c:\tmp\a`, "git-bash", "/c/tmp/a"},
		{`C:\Users\x`, "", ""},
		{"/home/x", "git-bash", ""},
		{"relative", "git-bash", ""},
	}
	for _, c := range cases {
		if got := POSIX(c.in, c.shell); got != c.want {
			t.Errorf("POSIX(%q, %q) = %q, want %q", c.in, c.shell, got, c.want)
		}
	}
}

func TestShellDetection(t *testing.T) {
	t.Setenv("MSYSTEM", "")
	t.Setenv("OSTYPE", "")
	t.Setenv("CYGWIN", "")
	if got := Shell(); got != "" {
		t.Errorf("Shell() = %q, want empty on this host", got)
	}
}
