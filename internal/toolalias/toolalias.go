// Package toolalias creates the short command names (dockerage, gitcheck,
// gitclone, gitarc, rmempty) that dispatch into the multi-call hall binary.
package toolalias

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Names are the short command names provided by hall.
var Names = []string{"dockerage", "gitcheck", "gitclone", "gitarc", "rmempty"}

const (
	// bashMarker marks a POSIX shim as created by hall.
	bashMarker = "# hall-kit alias"
	// cmdMarker marks a cmd.exe/PowerShell shim as created by hall.
	cmdMarker = "rem hall-kit alias"
)

// ShimScript returns the git-bash/Cygwin shim that runs "hall <name>".
func ShimScript(exe, name string) string {
	return "#!/bin/sh\n" + bashMarker + "\nexec \"" + filepath.ToSlash(exe) + "\" " + name + " \"$@\"\n"
}

// CmdShim returns the cmd.exe/PowerShell shim that runs "hall <name>". When
// relative is true it calls "%~dp0hall.exe", so the pair can be moved
// together.
func CmdShim(exe, name string, relative bool) string {
	target := strings.ReplaceAll(exe, "/", "\\")
	if relative {
		target = "%~dp0" + winBase(exe)
	}
	return "@echo off\r\n" + cmdMarker + "\r\n\"" + target + "\" " + name + " %*\r\n"
}

// winBase is filepath.Base for Windows paths, independent of the host the
// tests run on.
func winBase(p string) string {
	if i := strings.LastIndexAny(p, "/\\"); i >= 0 {
		return p[i+1:]
	}
	return p
}

// Create creates one short name per entry in names inside dir, pointing at the
// running hall executable. On Unix it uses symlinks; on Windows it writes a
// POSIX shim (for git-bash/Cygwin) plus a .cmd shim (for cmd.exe and
// PowerShell), because symlinks need extra privileges there. It returns the
// names created and the names skipped because a file hall does not own already
// exists there.
func Create(dir string, names []string) (created, skipped []string, err error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, nil, err
	}
	if resolved, rerr := filepath.EvalSymlinks(exe); rerr == nil {
		exe = resolved
	}
	return create(dir, exe, names)
}

// create is the runtime-flavoured wrapper used by Create.
func create(dir, exe string, names []string) (created, skipped []string, err error) {
	return createFor(dir, exe, names, runtime.GOOS == "windows")
}

// createFor is the testable core; windows selects the shim flavour.
func createFor(dir, exe string, names []string, windows bool) (created, skipped []string, err error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, nil, err
	}
	for _, name := range names {
		ok := true
		steps := []func() error{
			func() error { return replace(dir, exe, name, windows) },
		}
		if windows {
			steps = append(steps, func() error { return replaceCmd(dir, exe, name) })
		}
		for _, step := range steps {
			if err := step(); err != nil {
				if err == errNotOwned {
					ok = false
					continue
				}
				return created, skipped, err
			}
		}
		if ok {
			created = append(created, name)
		} else {
			skipped = append(skipped, name)
		}
	}
	return created, skipped, nil
}

var errNotOwned = os.ErrPermission

// replace writes the main shim (an executable shim on Windows, a symlink
// elsewhere) unless a file hall does not own is in the way.
func replace(dir, exe, name string, windows bool) error {
	dest := filepath.Join(dir, name)
	owned, err := ownedByHall(dest)
	if err != nil {
		return err
	}
	if !owned {
		return errNotOwned
	}
	if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
		return err
	}
	if windows {
		return os.WriteFile(dest, []byte(ShimScript(exe, name)), 0o755)
	}
	target := exe
	if filepath.Dir(exe) == filepath.Dir(dest) {
		target = filepath.Base(exe)
	}
	return os.Symlink(target, dest)
}

// replaceCmd writes the cmd.exe/PowerShell .cmd shim unless a foreign file is
// in the way.
func replaceCmd(dir, exe, name string) error {
	dest := filepath.Join(dir, name+".cmd")
	owned, err := ownedByHall(dest)
	if err != nil {
		return err
	}
	if !owned {
		return errNotOwned
	}
	if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.WriteFile(dest, []byte(CmdShim(exe, name, filepath.Dir(exe) == dir)), 0o755)
}

// ownedByHall reports whether dest may be replaced: it is missing, a symlink,
// or a shim script written by hall.
func ownedByHall(dest string) (bool, error) {
	fi, err := os.Lstat(dest)
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return true, nil
	}
	if !fi.Mode().IsRegular() {
		return false, nil
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		return false, err
	}
	s := string(data)
	return strings.Contains(s, bashMarker) || strings.Contains(s, cmdMarker), nil
}
