package shellinstall

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/QuietSugar/hall-kit/internal/version"
	shellassets "github.com/QuietSugar/hall-kit/shell"
)

const stateFile = ".state.json"

type state struct {
	Version     string   `json:"version"`
	InstalledAt string   `json:"installedAt"`
	RCFiles     []string `json:"rcFiles,omitempty"`
	Files       []string `json:"files,omitempty"`
}

// assetDest maps an embedded asset path to its install destination under root.
// A root of "" yields a slash-separated path relative to the install root.
func assetDest(root, asset string) (string, os.FileMode, bool) {
	shellDir := filepath.Join(root, "shell")
	switch {
	case strings.HasPrefix(asset, "lib/"):
		return filepath.Join(shellDir, filepath.FromSlash(asset)), 0o644, true
	case strings.HasPrefix(asset, "source/"):
		return filepath.Join(shellDir, filepath.FromSlash(asset)), 0o755, true
	case strings.HasPrefix(asset, "bin/"):
		return filepath.Join(root, "bin", strings.TrimSuffix(filepath.Base(asset), ".sh")), 0o755, true
	default:
		return "", 0, false
	}
}

// managedFiles lists the files hall shell owns, relative to the install root.
func managedFiles() ([]string, error) {
	var files []string
	err := fs.WalkDir(shellassets.Files, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || path == "example.env" {
			return nil
		}
		if dest, _, ok := assetDest("", path); ok {
			files = append(files, filepath.ToSlash(dest))
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

func (o *options) install() error {
	root := filepath.Join(o.home, ".hall")
	sourceDir := filepath.Join(root, "shell", "source")
	binDir := filepath.Join(root, "bin")

	if !o.dryRun {
		if err := os.MkdirAll(root, 0o755); err != nil {
			return err
		}
	}
	if err := o.extractAssets(root); err != nil {
		return err
	}
	if err := o.reconcile(root); err != nil {
		return err
	}
	if err := o.writeAssetFile(filepath.Join(root, "env.example"), "example.env", 0o644); err != nil {
		return err
	}
	if err := o.migrateConfig(root); err != nil {
		return err
	}
	if err := o.ensureEnv(root); err != nil {
		return err
	}

	if !o.noRC {
		binDisplay := displayDir(o.home, binDir)
		sourceDisplay := displayDir(o.home, sourceDir)
		var firstErr error
		for _, rc := range o.rcTargets() {
			changed, err := updateRC(rc, binDisplay, sourceDisplay, o.dryRun)
			if err != nil {
				o.errorf("rc %s: %v", rc, err)
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			if changed {
				if o.dryRun {
					o.printf("would update %s", rc)
				} else {
					o.printf("updated %s", rc)
				}
			}
		}
		if err := o.removeLegacyRCBlocks(); err != nil && firstErr == nil {
			firstErr = err
		}
		if firstErr != nil {
			return firstErr
		}
	}
	if err := o.writeState(root); err != nil {
		return err
	}
	if o.dryRun {
		o.printf("would install shell assets under %s/shell", root)
		return nil
	}
	o.printf("installed shell assets under %s/shell", root)
	o.printf("open a new shell (or source your rc file) to pick up the shell tools")
	return nil
}

// reconcile removes files installed by an earlier version that are no longer
// shipped. User files were never recorded in the state file, so they are safe.
func (o *options) reconcile(root string) error {
	current, err := managedFiles()
	if err != nil {
		return err
	}
	keep := make(map[string]bool, len(current))
	for _, rel := range current {
		keep[rel] = true
	}
	var stale []string
	for _, rel := range o.readState(root).Files {
		if !keep[rel] {
			stale = append(stale, rel)
		}
	}
	return o.removeFiles(root, stale)
}

func (o *options) uninstall() error {
	root := filepath.Join(o.home, ".hall")

	if !o.noRC {
		for _, rc := range o.rcTargets() {
			changed, err := removeRCBlock(rc, o.dryRun)
			if err != nil {
				o.errorf("rc %s: %v", rc, err)
				continue
			}
			if changed {
				o.printf("removed managed block from %s", rc)
			}
		}
	}
	files := o.readState(root).Files
	if len(files) == 0 {
		var err error
		if files, err = managedFiles(); err != nil {
			return err
		}
	}
	files = append(files, filepath.ToSlash(filepath.Join("shell", stateFile)))
	if err := o.removeFiles(root, files); err != nil {
		return err
	}
	if o.purge {
		if o.dryRun {
			o.printf("would remove %s", root)
		} else if err := os.RemoveAll(root); err != nil {
			return err
		}
	}
	o.printf("uninstalled hall-kit shell assets")
	return nil
}

func (o *options) extractAssets(root string) error {
	return fs.WalkDir(shellassets.Files, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || path == "example.env" {
			return nil
		}
		dest, mode, ok := assetDest(root, path)
		if !ok {
			return nil
		}
		data, err := shellassets.Files.ReadFile(path)
		if err != nil {
			return err
		}
		return o.writeFile(dest, data, mode)
	})
}

// removeFiles deletes the given root-relative files and prunes the directories
// hall shell created once they are empty. Anything not listed (i.e. files
// written by the user) is never touched.
func (o *options) removeFiles(root string, rels []string) error {
	for _, rel := range rels {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if o.dryRun {
			o.printf("would remove %s", p)
			continue
		}
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if o.dryRun {
		return nil
	}
	for _, dir := range []string{
		filepath.Join(root, "shell", "lib"),
		filepath.Join(root, "shell", "source"),
		filepath.Join(root, "shell"),
		filepath.Join(root, "bin"),
	} {
		_ = os.Remove(dir)
	}
	return nil
}

func (o *options) ensureEnv(root string) error {
	env := filepath.Join(root, "env")
	if _, err := os.Stat(env); err == nil {
		return nil
	}
	return o.writeAssetFile(env, "example.env", 0o644)
}

func (o *options) writeAssetFile(dst, asset string, mode os.FileMode) error {
	data, err := shellassets.Files.ReadFile(asset)
	if err != nil {
		return err
	}
	return o.writeFile(dst, data, mode)
}

func (o *options) writeFile(path string, data []byte, mode os.FileMode) error {
	if existing, err := os.ReadFile(path); err == nil && bytes.Equal(existing, data) {
		if mode != 0 {
			if st, err := os.Stat(path); err == nil && st.Mode().Perm() != mode {
				if o.dryRun {
					o.printf("would chmod %s", path)
					return nil
				}
				return os.Chmod(path, mode)
			}
		}
		return nil
	}
	if o.dryRun {
		o.printf("would write %s", path)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, mode)
}

func (o *options) readState(root string) state {
	data, err := os.ReadFile(filepath.Join(root, "shell", stateFile))
	if err != nil {
		return state{}
	}
	var st state
	if err := json.Unmarshal(data, &st); err != nil {
		return state{}
	}
	return st
}

func (o *options) writeState(root string) error {
	files, err := managedFiles()
	if err != nil {
		return err
	}
	installedAt := o.readState(root).InstalledAt
	if installedAt == "" {
		installedAt = time.Now().Format(time.RFC3339)
	}
	st := state{Version: version.String(), InstalledAt: installedAt, Files: files}
	if !o.noRC {
		st.RCFiles = o.rcTargets()
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return o.writeFile(filepath.Join(root, "shell", stateFile), data, 0o644)
}
