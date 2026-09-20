//go:build windows

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func replaceWindowsDirectoryWithJunction(t *testing.T, original, target string) {
	t.Helper()
	moved := original + "-moved"
	if err := os.Rename(original, moved); err != nil {
		t.Fatalf("move preflighted Windows ancestor: %v", err)
	}
	output, err := exec.Command("cmd.exe", "/c", "mklink", "/J", original, target).CombinedOutput()
	if err != nil {
		t.Fatalf("replace Windows ancestor with junction: %v: %s", err, output)
	}
	t.Cleanup(func() {
		_ = os.Remove(original)
		_ = os.Rename(moved, original)
	})
}

func assertWindowsExternalTreeUntouched(t *testing.T, external, sentinel string) {
	t.Helper()
	entries, err := os.ReadDir(external)
	if err != nil {
		t.Fatalf("read external tree: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != filepath.Base(sentinel) {
		t.Fatalf("external tree changed: entries=%v", entries)
	}
	got, err := os.ReadFile(sentinel)
	if err != nil || !bytes.Equal(got, []byte("keep\n")) {
		t.Fatalf("external sentinel changed: %q err=%v", got, err)
	}
}

func TestWindowsBackupRestoreRejectsAncestorJunctionAfterPreflight(t *testing.T) {
	runtimeDir := t.TempDir()
	if _, err := buildBR10AdvisorFixture(runtimeDir); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"action-input.json", "outcome-input.json"} {
		if err := os.Remove(filepath.Join(runtimeDir, name)); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	if code, response := missionCall(t, "init", runtimeDir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("initialize Windows ancestor-race fixture: code=%d response=%+v", code, response)
	}

	root := t.TempDir()
	external := t.TempDir()
	sentinel := filepath.Join(external, "sentinel")
	if err := os.WriteFile(sentinel, []byte("keep\n"), 0600); err != nil {
		t.Fatal(err)
	}
	ancestor := filepath.Join(root, "backup-ancestor")
	if err := os.Mkdir(ancestor, 0700); err != nil {
		t.Fatal(err)
	}
	swapped := false
	outputParentBeforeOpenHook = func() {
		if swapped {
			return
		}
		swapped = true
		replaceWindowsDirectoryWithJunction(t, ancestor, external)
	}
	backupTarget := filepath.Join(ancestor, "missing", "backup")
	code, response := backupCall(t, "create", runtimeDir, backupTarget)
	outputParentBeforeOpenHook = nil
	if code == 0 || response["status"] != "TARGET_ERROR" {
		t.Fatalf("backup followed ancestor junction: code=%d response=%+v", code, response)
	}
	if !swapped {
		t.Fatal("backup did not reach the Windows ancestor-race seam")
	}
	assertWindowsExternalTreeUntouched(t, external, sentinel)

	validBackup := filepath.Join(root, "valid-backup")
	if code, response := backupCall(t, "create", runtimeDir, validBackup); code != 0 || response["status"] != "BACKED_UP" {
		t.Fatalf("prepare valid backup for Windows restore guard: code=%d response=%+v", code, response)
	}
	restoreAncestor := filepath.Join(root, "restore-ancestor")
	if err := os.Mkdir(restoreAncestor, 0700); err != nil {
		t.Fatal(err)
	}
	restoreExternal := t.TempDir()
	restoreSentinel := filepath.Join(restoreExternal, "sentinel")
	if err := os.WriteFile(restoreSentinel, []byte("keep\n"), 0600); err != nil {
		t.Fatal(err)
	}
	restoreSwapped := false
	outputParentBeforeOpenHook = func() {
		if restoreSwapped {
			return
		}
		restoreSwapped = true
		replaceWindowsDirectoryWithJunction(t, restoreAncestor, restoreExternal)
	}
	restoreTarget := filepath.Join(restoreAncestor, "missing", "restored")
	code, response = backupCall(t, "restore", validBackup, restoreTarget)
	outputParentBeforeOpenHook = nil
	if code == 0 || response["status"] != "TARGET_ERROR" {
		t.Fatalf("restore followed ancestor junction: code=%d response=%+v", code, response)
	}
	if !restoreSwapped {
		t.Fatal("restore did not reach the Windows ancestor-race seam")
	}
	assertWindowsExternalTreeUntouched(t, restoreExternal, restoreSentinel)
}

func TestWindowsStableReaderRejectsAncestorJunctionAfterPreflight(t *testing.T) {
	root := t.TempDir()
	ancestor := filepath.Join(root, "reader-ancestor")
	if err := os.Mkdir(ancestor, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(ancestor, "history.jsonl")
	external := t.TempDir()
	externalPath := filepath.Join(external, "history.jsonl")
	original := []byte(`{"id":"canonical"}` + "\n")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(externalPath, original, 0600); err != nil {
		t.Fatal(err)
	}
	stableRegularFileBeforeOpenHook = func(got string) error {
		if filepath.Clean(got) != filepath.Clean(path) {
			t.Fatalf("reader hook path = %q, want %q", got, path)
		}
		replaceWindowsDirectoryWithJunction(t, ancestor, external)
		return nil
	}
	t.Cleanup(func() { stableRegularFileBeforeOpenHook = nil })
	if _, _, err := readStableRegularFile(path); err == nil {
		t.Fatal("stable reader followed an ancestor junction")
	}
	if got, err := os.ReadFile(externalPath); err != nil || !bytes.Equal(got, original) {
		t.Fatalf("external reader target changed: %q err=%v", got, err)
	}
}

func TestWindowsStableAppendRejectsAncestorJunctionAfterPreflight(t *testing.T) {
	root := t.TempDir()
	ancestor := filepath.Join(root, "writer-ancestor")
	if err := os.Mkdir(ancestor, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(ancestor, "history.jsonl")
	external := t.TempDir()
	externalPath := filepath.Join(external, "history.jsonl")
	original := []byte(`{"id":"canonical"}` + "\n")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(externalPath, original, 0600); err != nil {
		t.Fatal(err)
	}
	stableRegularFileAppendHook = func(got string) error {
		if filepath.Clean(got) != filepath.Clean(path) {
			t.Fatalf("append hook path = %q, want %q", got, path)
		}
		replaceWindowsDirectoryWithJunction(t, ancestor, external)
		return nil
	}
	t.Cleanup(func() { stableRegularFileAppendHook = nil })
	file, err := openStableRegularFileForAppendPath(path, false)
	if err == nil {
		_ = file.Close()
		t.Fatal("stable append followed an ancestor junction")
	}
	if got, err := os.ReadFile(externalPath); err != nil || !bytes.Equal(got, original) {
		t.Fatalf("external append target changed: %q err=%v", got, err)
	}
}
