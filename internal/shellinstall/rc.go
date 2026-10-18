package shellinstall

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Markers delimiting the block hall shell owns in rc files.
const (
	blockStart = "# >>> hall-kit shell start >>>"
	blockEnd   = "# <<< hall-kit shell end <<<"

	legacyPathStart = "# HALL_COMMAND_PATH_START"
	legacyPathEnd   = "# HALL_COMMAND_PATH_END"
	legacySrcStart  = "# HALL_COMMAND_SOURCE_START"
	legacySrcEnd    = "# HALL_COMMAND_SOURCE_END"
)

// snippet builds the managed rc block that sources every *.sh under dir.
// dir may contain the literal "$HOME" to stay relocatable.
// snippet builds the managed rc block: it puts binDir on PATH and sources
// every *.sh under sourceDir. Both may contain the literal "$HOME" to stay
// relocatable.
func snippet(binDir, sourceDir string) string {
	return strings.Join([]string{
		blockStart,
		"if [ -d \"" + binDir + "\" ]; then",
		"  case \":$PATH:\" in",
		"    *\":\"" + binDir + "\":\"*) ;;",
		"    *) PATH=\"" + binDir + ":$PATH\"; export PATH ;;",
		"  esac",
		"fi",
		"if [ -d \"" + sourceDir + "\" ]; then",
		"  for __hall_sh in \"" + sourceDir + "\"/*.sh; do",
		"    [ -f \"$__hall_sh\" ] || continue",
		"    . \"$__hall_sh\" || echo \"[WARN] Failed to source: $__hall_sh\" >&2",
		"  done",
		"  unset __hall_sh",
		"fi",
		blockEnd,
	}, "\n")
}

// displayDir renders dir as "$HOME/..." when it lives under home, so the
// written block keeps working if the home directory moves.
func displayDir(home, dir string) string {
	rel, err := filepath.Rel(home, dir)
	if err == nil && rel != "." && rel != ".." && !filepath.IsAbs(rel) && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "$HOME/" + filepath.ToSlash(rel)
	}
	return filepath.ToSlash(dir)
}

// hasMarked reports whether content contains a complete start..end block.
func hasMarked(content, start, end string) (bool, error) {
	lines := strings.Split(content, "\n")
	started := false
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if !started && t == start {
			started = true
			continue
		}
		if started && t == end {
			return true, nil
		}
	}
	if started {
		return false, fmt.Errorf("marker %q has no matching %q", start, end)
	}
	return false, nil
}

// removeMarked removes a start..end block, returning the new content and
// whether anything was removed.
func removeMarked(content, start, end string) (string, bool, error) {
	lines := strings.Split(content, "\n")
	si := -1
	for i, l := range lines {
		if strings.TrimSpace(l) == start {
			si = i
			break
		}
	}
	if si == -1 {
		return content, false, nil
	}
	ei := -1
	for i := si + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == end {
			ei = i
			break
		}
	}
	if ei == -1 {
		return content, false, fmt.Errorf("marker %q has no matching %q", start, end)
	}
	out := append(append([]string{}, lines[:si]...), lines[ei+1:]...)
	return strings.Join(out, "\n"), true, nil
}

// ensureBlock rewrites content so it contains exactly one managed block and
// no legacy installer blocks.
func ensureBlock(content, binDir, sourceDir string) (string, error) {
	c, _, err := removeMarked(content, blockStart, blockEnd)
	if err != nil {
		return content, err
	}
	for _, m := range [][2]string{{legacyPathStart, legacyPathEnd}, {legacySrcStart, legacySrcEnd}} {
		c, _, err = removeMarked(c, m[0], m[1])
		if err != nil {
			return content, err
		}
	}
	c = strings.TrimRight(c, "\n")
	if c != "" {
		c += "\n\n"
	}
	return c + snippet(binDir, sourceDir) + "\n", nil
}

func rcPerm(path string) os.FileMode {
	if st, err := os.Stat(path); err == nil {
		return st.Mode().Perm()
	}
	return 0o644
}

// updateRC writes the managed block into path. It returns whether the file
// changed.
func updateRC(path, binDir, sourceDir string, dryRun bool) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	content := string(data)
	next, err := ensureBlock(content, binDir, sourceDir)
	if err != nil {
		return false, err
	}
	if next == content {
		return false, nil
	}
	if dryRun {
		return true, nil
	}
	if d := filepath.Dir(path); d != "" {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return false, err
		}
	}
	return true, os.WriteFile(path, []byte(next), rcPerm(path))
}

// removeRCBlock removes the managed block from path.
func removeRCBlock(path string, dryRun bool) (bool, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	content := string(data)
	next, removed, err := removeMarked(content, blockStart, blockEnd)
	if err != nil {
		return false, err
	}
	if !removed {
		return false, nil
	}
	if dryRun {
		return true, nil
	}
	next = strings.TrimRight(next, "\n") + "\n"
	return true, os.WriteFile(path, []byte(next), rcPerm(path))
}

// removeLegacyRCBlocks strips the marker blocks written by the historical
// hall-command installer. It is part of the temporary migration path.
func (o *options) removeLegacyRCBlocks() error {
	var firstErr error
	for _, rc := range o.rcTargets() {
		data, err := os.ReadFile(rc)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			o.errorf("rc %s: %v", rc, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		content := string(data)
		changed := false
		for _, m := range [][2]string{{legacyPathStart, legacyPathEnd}, {legacySrcStart, legacySrcEnd}} {
			next, removed, err := removeMarked(content, m[0], m[1])
			if err != nil {
				o.errorf("rc %s: %v", rc, err)
				if firstErr == nil {
					firstErr = err
				}
				break
			}
			if removed {
				content = next
				changed = true
			}
		}
		if !changed {
			continue
		}
		if o.dryRun {
			o.printf("would clean legacy block from %s", rc)
			continue
		}
		if err := os.WriteFile(rc, []byte(strings.TrimRight(content, "\n")+"\n"), rcPerm(rc)); err != nil {
			o.errorf("rc %s: %v", rc, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		o.printf("cleaned legacy block from %s", rc)
	}
	return firstErr
}

func hasLegacy(content string) bool {
	for _, m := range [][2]string{{legacyPathStart, legacyPathEnd}, {legacySrcStart, legacySrcEnd}} {
		if ok, _ := hasMarked(content, m[0], m[1]); ok {
			return true
		}
	}
	return false
}
