package rmempty

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClean(t *testing.T) {
	root := t.TempDir()

	mustMkdir(t, filepath.Join(root, "a", "b"))
	mustMkdir(t, filepath.Join(root, "a", "c", "d"))
	mustMkdir(t, filepath.Join(root, "a", "keep"))
	mustMkdir(t, filepath.Join(root, "a", ".git", "empty"))
	mustWrite(t, filepath.Join(root, "a", "keep", "file.txt"))
	mustWrite(t, filepath.Join(root, "a", ".DS_Store"))
	mustWrite(t, filepath.Join(root, "a", ".git", "empty", "x"))

	res, err := Clean(root)
	if err != nil {
		t.Fatal(err)
	}

	if res.DSStore != 1 {
		t.Errorf("DSStore = %d, want 1", res.DSStore)
	}
	removed := map[string]bool{}
	for _, d := range res.Removed {
		removed[filepath.Base(d)] = true
	}
	for _, want := range []string{"d", "c", "b"} {
		if !removed[want] {
			t.Errorf("directory %q was not removed; removed=%v", want, res.Removed)
		}
	}

	for _, mustExist := range []string{
		filepath.Join(root, "a"),
		filepath.Join(root, "a", "keep", "file.txt"),
		filepath.Join(root, "a", ".git", "empty", "x"),
	} {
		if _, err := os.Stat(mustExist); err != nil {
			t.Errorf("expected %s to exist: %v", mustExist, err)
		}
	}
	for _, mustGone := range []string{
		filepath.Join(root, "a", "b"),
		filepath.Join(root, "a", "c"),
		filepath.Join(root, "a", ".DS_Store"),
	} {
		if _, err := os.Stat(mustGone); !os.IsNotExist(err) {
			t.Errorf("expected %s to be gone, stat err = %v", mustGone, err)
		}
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}
