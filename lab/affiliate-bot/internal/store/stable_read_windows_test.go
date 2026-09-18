//go:build windows

package store

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func replaceStoreWindowsDirectoryWithJunction(t *testing.T, original, target string) {
	t.Helper()
	moved := original + "-moved"
	if err := os.Rename(original, moved); err != nil {
		t.Fatalf("move preflighted store ancestor: %v", err)
	}
	output, err := exec.Command("cmd.exe", "/c", "mklink", "/J", original, target).CombinedOutput()
	if err != nil {
		t.Fatalf("replace store ancestor with junction: %v: %s", err, output)
	}
	t.Cleanup(func() {
		_ = os.Remove(original)
		_ = os.Rename(moved, original)
	})
}

func TestWindowsJSONLReaderRejectsAncestorJunctionAfterPreflight(t *testing.T) {
	root := t.TempDir()
	ancestor := filepath.Join(root, "history-ancestor")
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
	openPathHook = func(got string) error {
		if filepath.Clean(got) != filepath.Clean(path) {
			t.Fatalf("store reader hook path = %q, want %q", got, path)
		}
		replaceStoreWindowsDirectoryWithJunction(t, ancestor, external)
		return nil
	}
	t.Cleanup(func() { openPathHook = nil })
	reader, err := (JSONL{}).Open(path)
	if err == nil {
		_ = reader.Close()
		t.Fatal("store reader followed an ancestor junction")
	}
	if got, err := os.ReadFile(externalPath); err != nil || !bytes.Equal(got, original) {
		t.Fatalf("external store reader target changed: %q err=%v", got, err)
	}
}
