package shellinstall

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/QuietSugar/hall-kit/internal/version"
)

func (o *options) status() {
	root := filepath.Join(o.home, ".hall")
	_, _ = fmt.Fprintf(o.out, "version: %s\n", version.String())
	_, _ = fmt.Fprintf(o.out, "home:    %s\n", o.home)
	_, _ = fmt.Fprintf(o.out, "root:    %s\n", root)

	files, err := managedFiles()
	if err != nil {
		o.errorf("cannot list assets: %v", err)
		return
	}
	for _, rel := range files {
		_, _ = fmt.Fprintf(o.out, "  [%s] %s\n", mark(filepath.Join(root, filepath.FromSlash(rel))), rel)
	}
	_, _ = fmt.Fprintf(o.out, "  [%s] env\n", mark(filepath.Join(root, "env")))

	for _, rc := range o.rcTargets() {
		_, _ = fmt.Fprintf(o.out, "  rc %s: %s\n", rc, rcState(rc))
	}
	legacy := legacyRoot(o.home)
	if _, err := os.Stat(legacy); err == nil {
		_, _ = fmt.Fprintf(o.out, "legacy:  %s exists (run \"hall shell migrate\" to carry over config)\n", legacy)
	} else {
		_, _ = fmt.Fprintf(o.out, "legacy:  none\n")
	}
}

func mark(path string) string {
	if _, err := os.Stat(path); err == nil {
		return "ok"
	}
	return "--"
}

func rcState(path string) string {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "missing"
	}
	if err != nil {
		return err.Error()
	}
	content := string(data)
	managed, _ := hasMarked(content, blockStart, blockEnd)
	legacy := hasLegacy(content)
	switch {
	case managed && legacy:
		return "managed + legacy block"
	case managed:
		return "managed block"
	case legacy:
		return "legacy block only"
	default:
		return "no block"
	}
}
