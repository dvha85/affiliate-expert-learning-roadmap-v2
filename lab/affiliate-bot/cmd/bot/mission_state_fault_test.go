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

func TestMissionStopPostRenameSyncFaultKeepsDurableStop(t *testing.T) {
	for _, scenario := range []struct {
		name, write string
		faultWrite  int
		markerAlive bool
	}{
		{"state", "mission-state.json", 1, false},
		{"marker", "STOP", 2, true},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			dir := t.TempDir()
			if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
				t.Fatalf("initialize stop fixture: code=%d response=%+v", code, response)
			}
			writes := 0
			missionStateWriteFault = func(phase string) error {
				if phase == "after_rename_before_parent_sync" {
					writes++
					if writes == scenario.faultWrite {
						return errors.New("injected STOP parent-directory sync fault")
					}
				}
				return nil
			}
			t.Cleanup(func() { missionStateWriteFault = nil })
			if code, response := missionCall(t, "m11-stop", dir, "post-rename-stop"); code == 0 || response["status"] != "PUBLISHED_RECOVERY_REQUIRED" {
				t.Fatalf("post-rename %s fault hid a visible durable stop: code=%d response=%+v", scenario.write, code, response)
			}
			missionStateWriteFault = nil
			state, err := loadMissionState(dir)
			if err != nil || !state.Stop || state.StopReason != "post-rename-stop" {
				t.Fatalf("post-rename STOP was not visible in mission state: state=%+v err=%v", state, err)
			}
			if active, err := stopMarkerActive(dir); err != nil || active != scenario.markerAlive {
				t.Fatalf("unexpected STOP marker state after %s fault: active=%v err=%v", scenario.write, active, err)
			}
			binary := buildMissionBinary(t)
			if code, response := missionBinaryCall(t, binary, "mission", "m11-register", dir, "PRODUCTION_LEASE", filepath.Join(t.TempDir(), "unread.json")); code == 0 || response["status"] != "STOPPED" {
				t.Fatalf("fresh Bot did not retain durable STOP after post-rename fault: code=%d response=%+v", code, response)
			}
		})
	}
}

func TestMissionStopCannotOverwriteDurableReason(t *testing.T) {
	dir := t.TempDir()
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("initialize stop fixture: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m11-stop", dir, "original-stop"); code != 0 || response["status"] != "STOPPED" {
		t.Fatalf("initial stop failed: code=%d response=%+v", code, response)
	}
	stateBefore, err := os.ReadFile(missionStatePath(dir))
	if err != nil {
		t.Fatal(err)
	}
	markerBefore, err := os.ReadFile(filepath.Join(dir, "STOP"))
	if err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "m11-stop", dir, "replacement-stop"); code == 0 || response["status"] != "STOPPED" {
		t.Fatalf("second stop did not fail closed: code=%d response=%+v", code, response)
	}
	stateAfter, err := os.ReadFile(missionStatePath(dir))
	if err != nil || !bytes.Equal(stateBefore, stateAfter) {
		t.Fatalf("second stop rewrote canonical state: before=%s after=%s err=%v", stateBefore, stateAfter, err)
	}
	markerAfter, err := os.ReadFile(filepath.Join(dir, "STOP"))
	if err != nil || !bytes.Equal(markerBefore, markerAfter) {
		t.Fatalf("second stop rewrote durable marker: before=%s after=%s err=%v", markerBefore, markerAfter, err)
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
