package shellinstall

import (
	"os"
	"path/filepath"
	"time"
)

// legacyRoot is the directory used by the historical hall-command install.
// This whole file is temporary: remove it once the migration window closes.
func legacyRoot(home string) string {
	return filepath.Join(home, ".hall-command")
}

func (o *options) migrate() error {
	root := filepath.Join(o.home, ".hall")
	if err := o.migrateConfig(root); err != nil {
		return err
	}
	if err := o.removeLegacyRCBlocks(); err != nil {
		return err
	}
	if o.purge {
		if err := o.backupLegacy(); err != nil {
			return err
		}
	}
	o.printf("migration done")
	return nil
}

// migrateConfig copies the legacy config to ~/.hall/env when the new file does
// not exist yet. The existing file is never overwritten.
func (o *options) migrateConfig(root string) error {
	legacyEnv := filepath.Join(legacyRoot(o.home), "env")
	newEnv := filepath.Join(root, "env")
	if _, err := os.Stat(newEnv); err == nil {
		return nil
	}
	if _, err := os.Stat(legacyEnv); err != nil {
		return nil
	}
	data, err := os.ReadFile(legacyEnv)
	if err != nil {
		return err
	}
	if o.dryRun {
		o.printf("would migrate config %s -> %s", legacyEnv, newEnv)
		return nil
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(newEnv, data, 0o644); err != nil {
		return err
	}
	o.printf("migrated config %s -> %s", legacyEnv, newEnv)
	return nil
}

// backupLegacy moves ~/.hall-command aside instead of deleting it, so user
// files kept there are never lost. The download cache is safe to remove.
func (o *options) backupLegacy() error {
	legacy := legacyRoot(o.home)
	if _, err := os.Stat(legacy); err != nil {
		return nil
	}
	bak := legacy + ".bak." + time.Now().Format("20060102150405")
	if o.dryRun {
		o.printf("would move %s -> %s", legacy, bak)
		return nil
	}
	if err := os.Rename(legacy, bak); err != nil {
		return err
	}
	o.printf("backed up legacy dir %s -> %s", legacy, bak)

	cache := filepath.Join(o.home, ".cache", "hall-command")
	if _, err := os.Stat(cache); err == nil {
		if err := os.RemoveAll(cache); err != nil {
			return err
		}
		o.printf("removed cache %s", cache)
	}
	return nil
}
