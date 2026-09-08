package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteJSONAtomicFailurePreservesPriorStateAndRetry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mission-state.json")
	if err := os.WriteFile(path, []byte(`{"state":"old"}`), 0600); err != nil {
		t.Fatal(err)
	}
	missionStateWriteFault = func(phase string) error {
		if phase == "before_rename" {
			return errors.New("injected before-rename failure")
		}
		return nil
	}
	t.Cleanup(func() { missionStateWriteFault = nil })
	if err := writeJSONAtomic(path, map[string]string{"state": "new"}); err == nil {
		t.Fatal("injected failure was not surfaced")
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != `{"state":"old"}` {
		t.Fatalf("prior state changed after failed atomic write: %q err=%v", got, err)
	}
	matches, err := filepath.Glob(filepath.Join(dir, ".mission-state-*"))
	if err != nil || len(matches) != 0 {
		t.Fatalf("temporary state file leaked: %v err=%v", matches, err)
	}
	missionStateWriteFault = nil
	if err := writeJSONAtomic(path, map[string]string{"state": "new"}); err != nil {
		t.Fatalf("retry failed: %v", err)
	}
	got, err = os.ReadFile(path)
	var committed map[string]string
	if err == nil {
		err = json.Unmarshal(got, &committed)
	}
	if err != nil || committed["state"] != "new" {
		t.Fatalf("retry did not commit new state: %q err=%v", got, err)
	}
}
