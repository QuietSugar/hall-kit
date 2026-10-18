// Command hall is the multi-call entry point of hall-kit.
//
// It can be used as:
//
//	hall <tool> [args...]    run a tool, e.g. hall gitcheck ~/code
//	hall shell <command>     install and manage the shell assets under ~/.hall
//	hall aliases [--dir DIR] create the short command names next to this binary
//	hall -v | --version
//	hall -h | --help
//
// When the binary is reached through one of the short names (a symlink on
// Unix, a shim script in git-bash) it runs that tool directly, so the short
// form and "hall <tool>" behave identically.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/QuietSugar/hall-kit/internal/dockerage"
	"github.com/QuietSugar/hall-kit/internal/gitarc"
	"github.com/QuietSugar/hall-kit/internal/gitcheck"
	"github.com/QuietSugar/hall-kit/internal/gitclone"
	"github.com/QuietSugar/hall-kit/internal/rmempty"
	"github.com/QuietSugar/hall-kit/internal/shellinstall"
	"github.com/QuietSugar/hall-kit/internal/toolalias"
	"github.com/QuietSugar/hall-kit/internal/version"
)

// toolTable maps a tool name to its CLI entry.
var toolTable = map[string]func([]string){
	"dockerage": dockerage.Main,
	"gitcheck":  gitcheck.Main,
	"gitclone":  gitclone.Main,
	"gitarc":    gitarc.Main,
	"rmempty":   rmempty.Main,
}

const usageText = `hall - hall-kit multi-call entry point

Usage:
  hall <tool> [args...]     Run a tool
  hall shell <command>      Manage the shell assets under ~/.hall
  hall aliases [--dir DIR]  Create the short tool names next to this binary
  hall -v, --version        Print version information
  hall -h, --help           Print this help

Tools:
  dockerage   Print the age in seconds of a docker image creation timestamp
  gitcheck    Check whether git repositories can be safely deleted
  gitclone    Clone a repository into HALL_GIT_REPO_PATH by its URL layout
  gitarc      Move repository trees into HALL_GIT_REPO_PATH by remote URL
  rmempty     Remove empty directories (and .DS_Store files)

The tools are also installed under their own names (symlinks on Unix, shim
scripts in git-bash), so "gitcheck ..." and "hall gitcheck ..." behave the same.
`

func main() {
	base := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
	if fn, ok := toolTable[base]; ok {
		fn(os.Args[1:])
		return
	}

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("hall-kit - a personal collection of Go CLI tools")
		fmt.Println(version.String())
		fmt.Println()
		fmt.Println("Run \"hall --help\" for usage.")
		return
	}

	switch args[0] {
	case "-v", "--version":
		fmt.Println(version.String())
		return
	case "-h", "--help":
		fmt.Print(usageText)
		return
	case "shell":
		os.Exit(shellinstall.Run(args[1:], os.Stdout, os.Stderr))
	case "aliases":
		os.Exit(runAliases(args[1:]))
	}
	if fn, ok := toolTable[args[0]]; ok {
		fn(args[1:])
		return
	}
	fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
	fmt.Fprint(os.Stderr, usageText)
	os.Exit(2)
}

// runAliases implements "hall aliases".
func runAliases(args []string) int {
	fs := flag.NewFlagSet("hall aliases", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	dir := fs.String("dir", "", "directory for the short names (default: this binary's directory)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	target := *dir
	if target == "" {
		exe, err := os.Executable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot locate this binary: %v\n", err)
			return 1
		}
		target = filepath.Dir(exe)
	}
	created, skipped, err := toolalias.Create(target, toolalias.Names)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	for _, name := range created {
		fmt.Printf("created %s\n", filepath.Join(target, name))
	}
	for _, name := range skipped {
		fmt.Fprintf(os.Stderr, "skipped %s: a file hall does not own already exists\n", filepath.Join(target, name))
	}
	return 0
}
