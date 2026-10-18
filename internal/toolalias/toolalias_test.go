package toolalias

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestShimScript(t *testing.T) {
	s := ShimScript("/opt/hall/hall", "gitcheck")
	for _, want := range []string{bashMarker, "gitcheck", "exec", "/opt/hall/hall"} {
		if !strings.Contains(s, want) {
			t.Errorf("shim missing %q:\n%s", want, s)
		}
	}
}

func TestCmdShim(t *testing.T) {
	rel := CmdShim("C:\\Users\\x\\.hall\\bin\\hall.exe", "gitcheck", true)
	if !strings.Contains(rel, cmdMarker) || !strings.Contains(rel, "gitcheck") || !strings.Contains(rel, "%~dp0hall.exe") {
		t.Errorf("relative cmd shim unexpected:\n%s", rel)
	}
	abs := CmdShim("C:\\Users\\x\\.hall\\bin\\hall.exe", "gitcheck", false)
	if !strings.Contains(abs, "C:\\Users\\x\\.hall\\bin\\hall.exe") {
		t.Errorf("absolute cmd shim unexpected:\n%s", abs)
	}
}

func TestCreateIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "hall")
	if err := os.WriteFile(exe, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	created, skipped, err := create(dir, exe, []string{"gitcheck", "gitclone"})
	if err != nil {
		t.Fatal(err)
	}
	if len(created) != 2 || len(skipped) != 0 {
		t.Fatalf("created=%v skipped=%v", created, skipped)
	}

	if runtime.GOOS != "windows" {
		target, err := os.Readlink(filepath.Join(dir, "gitcheck"))
		if err != nil {
			t.Fatalf("expected a symlink: %v", err)
		}
		if target != "hall" {
			t.Errorf("symlink target = %q, want hall", target)
		}
	}

	// second run replaces its own aliases, still nothing skipped
	created2, skipped2, err := create(dir, exe, []string{"gitcheck", "gitclone"})
	if err != nil {
		t.Fatal(err)
	}
	if len(created2) != 2 || len(skipped2) != 0 {
		t.Fatalf("second run created=%v skipped=%v", created2, skipped2)
	}
}

func TestCreateForWindowsMakesBothShims(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "hall.exe")
	if err := os.WriteFile(exe, []byte("pe"), 0o755); err != nil {
		t.Fatal(err)
	}

	created, skipped, err := createFor(dir, exe, []string{"gitcheck"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(created) != 1 || len(skipped) != 0 {
		t.Fatalf("created=%v skipped=%v", created, skipped)
	}

	bash, err := os.ReadFile(filepath.Join(dir, "gitcheck"))
	if err != nil {
		t.Fatalf("posix shim missing: %v", err)
	}
	if !strings.Contains(string(bash), bashMarker) {
		t.Errorf("posix shim missing marker: %s", bash)
	}

	cmd, err := os.ReadFile(filepath.Join(dir, "gitcheck.cmd"))
	if err != nil {
		t.Fatalf("cmd shim missing: %v", err)
	}
	if !strings.Contains(string(cmd), cmdMarker) || !strings.Contains(string(cmd), "%~dp0hall.exe") {
		t.Errorf("cmd shim unexpected: %s", cmd)
	}

	// idempotent
	created2, skipped2, err := createFor(dir, exe, []string{"gitcheck"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(created2) != 1 || len(skipped2) != 0 {
		t.Fatalf("second run created=%v skipped=%v", created2, skipped2)
	}
}

func TestCreateSkipsUserFiles(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "hall")
	if err := os.WriteFile(exe, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	userFile := filepath.Join(dir, "gitcheck")
	if err := os.WriteFile(userFile, []byte("#!/bin/sh\necho mine\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	userCmd := filepath.Join(dir, "gitcheck.cmd")
	if err := os.WriteFile(userCmd, []byte("@echo off\r\necho mine\r\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	created, skipped, err := createFor(dir, exe, []string{"gitcheck"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(created) != 0 || len(skipped) != 1 {
		t.Fatalf("created=%v skipped=%v, want none created and one skipped", created, skipped)
	}
	data, _ := os.ReadFile(userFile)
	if !strings.Contains(string(data), "echo mine") {
		t.Errorf("user file was modified: %s", data)
	}
	data, _ = os.ReadFile(userCmd)
	if !strings.Contains(string(data), "echo mine") {
		t.Errorf("user cmd file was modified: %s", data)
	}
}
