// Package shellinstall installs and manages the shell assets that ship
// embedded in the hall binary, under ~/.hall/.
package shellinstall

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"

	userhome "github.com/QuietSugar/hall-kit/internal/home"
)

const usage = `hall shell - manage hall-kit's shell assets

Usage:
  hall shell install    [--no-rc] [--dry-run] [--home DIR] [--rc-file PATH]...
  hall shell uninstall  [--no-rc] [--dry-run] [--home DIR] [--rc-file PATH]... [--purge]
  hall shell status     [--home DIR] [--rc-file PATH]...
  hall shell migrate    [--dry-run] [--home DIR] [--rc-file PATH]... [--purge]

Flags:
  --no-rc        Do not touch shell rc files
  --dry-run      Print what would change without writing anything
  --home DIR     Use DIR as the home directory (default: the current user's)
  --rc-file PATH Add or use PATH as an rc file (repeatable)
  --purge        install/uninstall: also remove ~/.hall; migrate: back up the legacy ~/.hall-command
`

type options struct {
	home    string
	noRC    bool
	dryRun  bool
	purge   bool
	rcFiles []string
	out     io.Writer
	errOut  io.Writer
}

func (o *options) printf(format string, args ...any) {
	_, _ = fmt.Fprintf(o.out, format+"\n", args...)
}

func (o *options) errorf(format string, args ...any) {
	_, _ = fmt.Fprintf(o.errOut, format+"\n", args...)
}

// rcTargets returns the rc files to manage, honouring an explicit --rc-file.
func (o *options) rcTargets() []string {
	if len(o.rcFiles) > 0 {
		return o.rcFiles
	}
	home := o.home
	if runtime.GOOS == "windows" {
		return []string{filepath.Join(home, ".bash_profile"), filepath.Join(home, ".bashrc")}
	}
	return []string{filepath.Join(home, ".zshrc"), filepath.Join(home, ".bashrc"), filepath.Join(home, ".profile")}
}

// Run executes a "hall shell" invocation and returns a process exit code:
// 0 = success, 1 = runtime failure, 2 = usage error.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		_, _ = fmt.Fprint(stderr, usage)
		if len(args) == 0 {
			return 2
		}
		return 0
	}
	sub := args[0]

	fs := flag.NewFlagSet("hall shell "+sub, flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { _, _ = fmt.Fprint(stderr, usage) }

	var (
		home    = fs.String("home", "", "home directory to use")
		noRC    = fs.Bool("no-rc", false, "do not touch shell rc files")
		dryRun  = fs.Bool("dry-run", false, "print actions without writing")
		purge   = fs.Bool("purge", false, "remove or back up hall state as documented")
		rcFiles stringList
	)
	fs.Var(&rcFiles, "rc-file", "rc file to manage (repeatable)")

	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	if fs.NArg() > 0 {
		_, _ = fmt.Fprintf(stderr, "unexpected argument: %s\n", fs.Arg(0))
		return 2
	}

	h := *home
	if h == "" {
		uh, err := userhome.Dir()
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "cannot determine home directory: %v\n", err)
			return 1
		}
		h = uh
	}
	abs, err := filepath.Abs(expandHome(h))
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "cannot resolve home directory: %v\n", err)
		return 1
	}

	o := &options{home: abs, noRC: *noRC, dryRun: *dryRun, purge: *purge, rcFiles: rcFiles, out: stdout, errOut: stderr}

	switch sub {
	case "install":
		if err := o.install(); err != nil {
			o.errorf("error: %v", err)
			return 1
		}
	case "uninstall":
		if err := o.uninstall(); err != nil {
			o.errorf("error: %v", err)
			return 1
		}
	case "status":
		o.status()
	case "migrate":
		if err := o.migrate(); err != nil {
			o.errorf("error: %v", err)
			return 1
		}
	default:
		_, _ = fmt.Fprintf(stderr, "unknown subcommand: %s\n", sub)
		_, _ = fmt.Fprint(stderr, usage)
		return 2
	}
	return 0
}

// expandHome expands a leading ~ or ~/ in a user-supplied path.
func expandHome(dir string) string {
	if dir == "~" || strings.HasPrefix(dir, "~/") {
		if home, err := userhome.Dir(); err == nil {
			if dir == "~" {
				return home
			}
			return filepath.Join(home, dir[2:])
		}
	}
	return dir
}

type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }

func (s *stringList) Set(v string) error {
	*s = append(*s, v)
	return nil
}
