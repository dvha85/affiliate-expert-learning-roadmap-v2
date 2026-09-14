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

func TestMissionStopJournalRecoversEveryPostRenameBoundary(t *testing.T) {
	for _, scenario := range []struct {
		name, write string
		faultWrite  int
		stateAlive  bool
		markerAlive bool
	}{
		{"journal", "m11-manual-stop-journal.json", 1, false, false},
		{"state", "mission-state.json", 2, true, false},
		{"marker", "STOP", 3, true, true},
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
				t.Fatalf("post-rename %s fault was not disclosed as recoverable: code=%d response=%+v", scenario.write, code, response)
			}
			missionStateWriteFault = nil
			state, err := loadMissionState(dir)
			if err != nil || state.Stop != scenario.stateAlive || (state.Stop && state.StopReason != "post-rename-stop") {
				t.Fatalf("unexpected mission state after post-rename %s fault: state=%+v err=%v", scenario.write, state, err)
			}
			if active, err := stopMarkerActive(dir); err != nil || active != scenario.markerAlive {
				t.Fatalf("unexpected STOP marker state after %s fault: active=%v err=%v", scenario.write, active, err)
			}
			binary := buildMissionBinary(t)
			if code, response := missionBinaryCall(t, binary, "mission", "status", dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
				t.Fatalf("fresh Bot exposed partial direct STOP after %s fault: code=%d response=%+v", scenario.write, code, response)
			}
			if code, response := missionBinaryCall(t, binary, "mission", "m11-register", dir, "PRODUCTION_LEASE", filepath.Join(t.TempDir(), "unread.json")); code == 0 || response["status"] != "STOPPED" {
				t.Fatalf("locked fresh Bot did not recover then retain STOP after %s fault: code=%d response=%+v", scenario.write, code, response)
			}
			if _, err := os.Stat(m11ManualStopJournalPath(dir)); !os.IsNotExist(err) {
				t.Fatalf("direct STOP journal remains after locked recovery: %v", err)
			}
			state, err = loadMissionState(dir)
			if err != nil || !state.Stop || state.StopReason != "post-rename-stop" {
				t.Fatalf("direct STOP recovery did not retain state: state=%+v err=%v", state, err)
			}
			if active, err := stopMarkerActive(dir); err != nil || !active {
				t.Fatalf("direct STOP recovery did not retain marker: active=%v err=%v", active, err)
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

func TestBackupCreateRecoversPendingDirectStopJournal(t *testing.T) {
	dir := t.TempDir()
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("initialize stop fixture: code=%d response=%+v", code, response)
	}
	record, err := NewHistoryRecord("backup-stop-recovery", "2026-09-08T00:00:01Z", "2026-09-08T00:00:01Z", []Observation{historyObservation("backup-stop-observation", "backup-stop-product", "Backup Stop Product", 1, 0, "2026-09-08T00:00:00Z")})
	if err != nil {
		t.Fatal(err)
	}
	if status, err := AppendHistory(filepath.Join(dir, "history.jsonl"), record); err != nil || status != appendAdded {
		t.Fatalf("append backup history fixture: status=%s err=%v", status, err)
	}
	journal := m11ManualStopJournal{Version: "m11-manual-stop-journal/v1", Reason: "backup-stop-recovery"}
	if err := writeJSONAtomic(m11ManualStopJournalPath(dir), journal); err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "status", dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
		t.Fatalf("pending direct STOP journal was visible to status: code=%d response=%+v", code, response)
	}
	backupDir := filepath.Join(t.TempDir(), "backup")
	if code, response := backupCall(t, "create", dir, backupDir); code != 0 || response["status"] != "BACKED_UP" {
		t.Fatalf("backup create did not replay direct STOP journal: code=%d response=%+v", code, response)
	}
	if _, err := os.Stat(m11ManualStopJournalPath(dir)); !os.IsNotExist(err) {
		t.Fatalf("direct STOP journal remains after backup recovery: %v", err)
	}
	state, err := loadMissionState(dir)
	if err != nil || !state.Stop || state.StopReason != journal.Reason {
		t.Fatalf("backup recovery did not retain direct STOP state: state=%+v err=%v", state, err)
	}
	if active, err := stopMarkerActive(dir); err != nil || !active {
		t.Fatalf("backup recovery did not retain STOP marker: active=%v err=%v", active, err)
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
