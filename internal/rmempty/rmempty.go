// Package rmempty removes .DS_Store files and empty directories under a
// tree, mirroring the historical rm_empty_dir.sh helper.
package rmempty

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Result reports what Clean removed.
type Result struct {
	// Removed lists the directories that were removed, in removal order.
	Removed []string
	// DSStore is the number of .DS_Store files removed.
	DSStore int
}

// Clean removes every .DS_Store file under root and then removes all
// directories that are left empty, deepest first. The .git directory and its
// contents are left untouched, and root itself is never removed.
func Clean(root string) (Result, error) {
	var res Result
	if err := removeDSStore(root, &res); err != nil {
		return res, err
	}
	dirs, err := collectDirs(root)
	if err != nil {
		return res, err
	}
	for _, dir := range dirs {
		empty, err := isEmpty(dir)
		if err != nil || !empty {
			continue
		}
		if err := os.Remove(dir); err != nil {
			continue
		}
		res.Removed = append(res.Removed, dir)
	}
	return res, nil
}

func removeDSStore(root string, res *Result) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == ".DS_Store" {
			if err := os.Remove(path); err != nil {
				return err
			}
			res.DSStore++
		}
		return nil
	})
}

// collectDirs returns every directory under root except root itself and the
// .git subtrees, ordered deepest first so children are removed before their
// parents.
func collectDirs(root string) ([]string, error) {
	root = filepath.Clean(root)
	var dirs []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if path != root && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if path != root {
			dirs = append(dirs, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(dirs, func(i, j int) bool {
		return depth(dirs[i]) > depth(dirs[j])
	})
	return dirs, nil
}

func depth(path string) int {
	return strings.Count(path, string(filepath.Separator))
}

func isEmpty(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }()

	_, err = f.Readdirnames(1)
	switch {
	case err == io.EOF:
		return true, nil
	case err != nil:
		return false, err
	default:
		return false, nil
	}
}
