// Package gitarc implements the gitarc command, shared by the
// cmd/gitarc shim and by "hall gitarc".
package gitarc

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/QuietSugar/hall-kit/internal/gitenv"
	"github.com/QuietSugar/hall-kit/internal/giturl"
	"github.com/QuietSugar/hall-kit/internal/i18n"
)

// langFlag pre-scans --lang (accepting both "--lang zh" and "--lang=zh") so
// that usage and error messages are printed in the requested language.
func langFlag(args []string) string {
	for i, a := range args {
		if a == "--lang" && i+1 < len(args) {
			return args[i+1]
		}
		if v, ok := strings.CutPrefix(a, "--lang="); ok {
			return v
		}
	}
	return ""
}

// findRepos returns the outermost git repositories under root. Once a
// repository is found its subtree is not searched further, matching the
// historical shell behaviour. A .git directory or file (worktree) both count.
func findRepos(root string) ([]string, error) {
	var repos []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		// Like the shell glob "$dir"/*, hidden subdirectories are skipped so
		// repositories inside .cache/.venv are not moved by accident.
		if path != root && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}
		if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
			repos = append(repos, path)
			return filepath.SkipDir
		}
		return nil
	})
	return repos, err
}

// originURL returns the remote.origin.url of a repository, or an error when
// the repository has no such remote.
func originURL(repo string) (string, error) {
	out, err := exec.Command("git", "-C", repo, "config", "--get", "remote.origin.url").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// moveDir moves src to dst, falling back to a recursive copy when the two
// paths live on different filesystems (os.Rename then fails with EXDEV).
func moveDir(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}
	if !errors.Is(err, syscall.EXDEV) {
		return err
	}
	if err := copyTree(src, dst); err != nil {
		_ = os.RemoveAll(dst)
		return err
	}
	return os.RemoveAll(src)
}

// copyTree recursively copies src to dst, preserving symlinks and modes.
func copyTree(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(link, dst)
	}
	if !info.IsDir() {
		return copyFile(src, dst, info.Mode().Perm())
	}
	if err := os.MkdirAll(dst, info.Mode().Perm()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := copyTree(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

// copyFile copies one regular file, creating dst with the given permissions.
func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

// Main runs the gitarc command with args, the arguments after the
// program name. It is shared by cmd/gitarc and "hall gitarc".
func Main(args []string) {
	i18n.SetLang(i18n.Detect(langFlag(args)))

	dir := ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--lang":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				fmt.Fprintln(os.Stderr, i18n.T("gitarc.err.unknownflag", a))
				os.Exit(2)
			}
			i++ // value already consumed by langFlag
		case "-h", "--help":
			fmt.Println(i18n.T("gitarc.usage"))
			return
		default:
			if strings.HasPrefix(a, "--lang=") {
				continue // value already consumed by langFlag
			}
			if strings.HasPrefix(a, "-") {
				fmt.Fprintln(os.Stderr, i18n.T("gitarc.err.unknownflag", a))
				os.Exit(2)
			}
			if dir != "" {
				fmt.Fprintln(os.Stderr, i18n.T("gitarc.usage"))
				os.Exit(2)
			}
			dir = a
		}
	}
	if dir == "" {
		dir = "."
	}
	if _, err := exec.LookPath("git"); err != nil {
		fmt.Fprintln(os.Stderr, i18n.T("gitarc.err.nogit"))
		os.Exit(2)
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, i18n.T("gitarc.err.abspath", dir))
		os.Exit(2)
	}
	if st, err := os.Stat(abs); err != nil || !st.IsDir() {
		fmt.Fprintln(os.Stderr, i18n.T("gitarc.err.nodir", dir))
		os.Exit(2)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}

	root, err := gitenv.RepoRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, i18n.T("gitarc.err.root", err))
		os.Exit(2)
	}
	fmt.Print(i18n.T("gitarc.info.scan", abs, root))

	repos, err := findRepos(abs)
	if err != nil {
		fmt.Fprintln(os.Stderr, i18n.T("gitarc.err.scan", err))
		os.Exit(1)
	}
	if len(repos) == 0 {
		fmt.Println(i18n.T("gitarc.scan.norepos"))
		return
	}

	moved, skipped, failed := 0, 0, 0
	for _, repo := range repos {
		url, err := originURL(repo)
		if err != nil {
			fmt.Fprintln(os.Stderr, i18n.T("gitarc.warn.noremote", repo))
			skipped++
			continue
		}
		if !giturl.Valid(url) {
			fmt.Fprintln(os.Stderr, i18n.T("gitarc.warn.invalid", url, repo))
			skipped++
			continue
		}
		target, err := giturl.Target(root, url)
		if err != nil {
			fmt.Fprintln(os.Stderr, i18n.T("gitarc.warn.invalid", url, repo))
			skipped++
			continue
		}
		parent := filepath.Dir(target)
		if err := os.MkdirAll(parent, 0o755); err != nil {
			fmt.Fprintln(os.Stderr, i18n.T("gitarc.err.mkdir", parent, err))
			failed++
			continue
		}
		if _, err := os.Stat(target); err == nil {
			fmt.Fprintln(os.Stderr, i18n.T("gitarc.warn.exists", target))
			skipped++
			continue
		}
		if err := moveDir(repo, target); err != nil {
			fmt.Fprintln(os.Stderr, i18n.T("gitarc.err.move", repo, target, err))
			failed++
			continue
		}
		fmt.Println(i18n.T("gitarc.info.move", repo, target))
		moved++
	}

	fmt.Println(i18n.T("gitarc.scan.summary", moved, skipped, failed, len(repos)))
	if failed > 0 {
		os.Exit(1)
	}
}
