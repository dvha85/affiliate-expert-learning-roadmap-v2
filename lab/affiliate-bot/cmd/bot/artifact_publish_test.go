package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteNewJSONFailureLeavesNoPartialArtifactAndRetryPublishes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "artifact.json")
	artifactWriteFault = func(phase string) error {
		if phase == "before_publish" {
			return errors.New("injected publish interruption")
		}
		return nil
	}
	t.Cleanup(func() { artifactWriteFault = nil })
	if _, err := writeNewJSON(path, map[string]string{"state": "new"}); err == nil {
		t.Fatal("injected publish interruption was not surfaced")
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("partial artifact remained after failed publish: %v", err)
	}
	leftovers, err := filepath.Glob(filepath.Join(dir, ".artifact-*"))
	if err != nil || len(leftovers) != 0 {
		t.Fatalf("temporary artifact leaked: %v err=%v", leftovers, err)
	}
	artifactWriteFault = nil
	if status, err := writeNewJSON(path, map[string]string{"state": "new"}); err != nil || status != appendAdded {
		t.Fatalf("retry did not publish: status=%s err=%v", status, err)
	}
	first, err := os.ReadFile(path)
	if err != nil || !bytes.Contains(first, []byte(`"state": "new"`)) {
		t.Fatalf("published artifact invalid: %s err=%v", first, err)
	}
	if status, err := writeNewJSON(path, map[string]string{"state": "new"}); err != nil || status != appendDuplicate {
		t.Fatalf("exact retry did not preserve immutable artifact: status=%s err=%v", status, err)
	}
	if current, err := os.ReadFile(path); err != nil || !bytes.Equal(first, current) {
		t.Fatalf("artifact changed on exact retry: %s err=%v", current, err)
	}
}

func TestWriteNewJSONRejectsExternalSymlinkAndPostOpenSwap(t *testing.T) {
	dir := t.TempDir()
	external := filepath.Join(t.TempDir(), "external.json")
	expected, err := marshalJSON(map[string]string{"state": "new"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(external, expected, 0600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "artifact.json")
	if err := os.Symlink(external, path); err != nil {
		t.Fatal(err)
	}
	if _, err := writeNewJSON(path, map[string]string{"state": "new"}); err == nil {
		t.Fatal("symlink artifact output was accepted as an exact duplicate")
	}
	if got, err := os.ReadFile(external); err != nil || !bytes.Equal(got, expected) {
		t.Fatalf("external target changed through rejected output: %q err=%v", got, err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, expected, 0600); err != nil {
		t.Fatal(err)
	}
	stableRegularFileReadHook = func(openedPath string) error {
		if openedPath != path {
			return nil
		}
		stableRegularFileReadHook = nil
		if err := os.Remove(path); err != nil {
			return err
		}
		return os.Symlink(external, path)
	}
	t.Cleanup(func() { stableRegularFileReadHook = nil })
	if _, err := writeNewJSON(path, map[string]string{"state": "new"}); err == nil {
		t.Fatal("post-open symlink replacement was accepted as an exact duplicate")
	}
	if got, err := os.ReadFile(external); err != nil || !bytes.Equal(got, expected) {
		t.Fatalf("external target changed after post-open replacement: %q err=%v", got, err)
	}
}

func TestWriteNewJSONRejectsSymlinkParent(t *testing.T) {
	outside := t.TempDir()
	dir := t.TempDir()
	parent := filepath.Join(dir, "artifact-parent")
	if err := os.Symlink(outside, parent); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(parent, "artifact.json")
	if _, err := writeNewJSON(path, map[string]string{"state": "new"}); err == nil {
		t.Fatal("artifact publisher accepted a symlink parent")
	}
	if _, err := os.Lstat(filepath.Join(outside, "artifact.json")); !os.IsNotExist(err) {
		t.Fatalf("artifact was published outside the owned parent: %v", err)
	}
}
