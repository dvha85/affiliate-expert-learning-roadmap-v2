package main

import (
	"bytes"
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

func TestMissionInitRejectsSymlinkRuntimeDirectoryWithoutExternalMutation(t *testing.T) {
	outside := t.TempDir()
	alias := filepath.Join(t.TempDir(), "runtime-alias")
	if err := os.Symlink(outside, alias); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := runMissionCommand([]string{"init", alias}, &stdout, &stderr); code == 0 {
		t.Fatalf("mission init accepted symlink runtime directory: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	if _, err := os.Lstat(filepath.Join(outside, "mission-state.json")); !os.IsNotExist(err) {
		t.Fatalf("mission init created external state through symlink runtime: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(outside, ".mission.lock")); !os.IsNotExist(err) {
		t.Fatalf("mission init created external lock through symlink runtime: %v", err)
	}
}

func TestMissionStatusRejectsSymlinkRuntimeDirectoryBeforeRead(t *testing.T) {
	outside := t.TempDir()
	if code, response := missionCall(t, "init", outside); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("initialize external fixture: code=%d response=%+v", code, response)
	}
	alias := filepath.Join(t.TempDir(), "runtime-alias")
	if err := os.Symlink(outside, alias); err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "status", alias); code == 0 || response["status"] != "STATE_ERROR" {
		t.Fatalf("status accepted symlink runtime directory: code=%d response=%+v", code, response)
	}
}
