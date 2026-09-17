package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
)

// TestM11ManualStopProcessTerminationChild is a test-binary-only entrypoint.
// It terminates the child after the journal, mission state, or STOP marker
// rename becomes visible. Production never reads these variables.
func TestM11ManualStopProcessTerminationChild(t *testing.T) {
	if os.Getenv("GO_WANT_M11_MANUAL_STOP_PROCESS_TERMINATION") != "1" {
		return
	}
	mode := os.Getenv("GO_M11_MANUAL_STOP_PROCESS_TERMINATION_MODE")
	targetWrite := map[string]int{
		"after-journal-rename": 1,
		"after-state-rename":   2,
		"after-marker-rename":  3,
	}[mode]
	if targetWrite == 0 {
		os.Exit(2)
	}
	separator := -1
	for index, value := range os.Args {
		if value == "--" {
			separator = index
			break
		}
	}
	if separator == -1 || separator+1 >= len(os.Args) {
		os.Exit(2)
	}
	writes := 0
	missionStateWriteFault = func(phase string) error {
		if phase != "after_rename_before_parent_sync" {
			return nil
		}
		writes++
		if writes != targetWrite {
			return nil
		}
		if err := syscall.Kill(os.Getpid(), syscall.SIGKILL); err != nil {
			os.Exit(98)
		}
		// SIGKILL is non-catchable. Keep the child at the injected boundary if
		// signal delivery is deferred until the next scheduling point.
		for {
			runtime.Gosched()
		}
	}
	os.Exit(runMissionCommand(os.Args[separator+1:], os.Stdout, os.Stderr))
}

// TestM11ManualStopProcessKillRequiresLockedRecovery proves that a real
// process termination at each direct-STOP file boundary leaves the journal as
// the recovery authority. A fresh reader must fail closed; only a locked
// writer may replay the exact STOP transition, once, across all three files.
func TestM11ManualStopProcessKillRequiresLockedRecovery(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("SIGKILL process-boundary proof is not portable on Windows")
	}
	binary := buildMissionBinary(t)
	for _, mode := range []string{"after-journal-rename", "after-state-rename", "after-marker-rename"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
				t.Fatalf("initialize direct STOP fixture: code=%d response=%+v", code, response)
			}

			commandArgs := []string{"m11-stop", dir, "process-kill-direct-stop"}
			command := exec.Command(os.Args[0], append([]string{"-test.run=^TestM11ManualStopProcessTerminationChild$", "--"}, commandArgs...)...)
			command.Env = append(os.Environ(),
				"GO_WANT_M11_MANUAL_STOP_PROCESS_TERMINATION=1",
				"GO_M11_MANUAL_STOP_PROCESS_TERMINATION_MODE="+mode,
			)
			var childStdout, childStderr bytes.Buffer
			command.Stdout, command.Stderr = &childStdout, &childStderr
			err := command.Run()
			if err == nil {
				t.Fatalf("direct STOP child unexpectedly completed after injected SIGKILL: stdout=%q stderr=%q", childStdout.String(), childStderr.String())
			}
			exited, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatal(err)
			}
			status, ok := exited.ProcessState.Sys().(syscall.WaitStatus)
			if !ok || !status.Signaled() || status.Signal() != syscall.SIGKILL {
				t.Fatalf("direct STOP child did not receive SIGKILL: %v stdout=%q stderr=%q", err, childStdout.String(), childStderr.String())
			}

			journalPath := m11ManualStopJournalPath(dir)
			if info, err := os.Stat(journalPath); err != nil || !info.Mode().IsRegular() {
				t.Fatalf("process kill did not leave a regular direct STOP journal: %v", err)
			}
			state, err := loadMissionState(dir)
			if err != nil {
				t.Fatal(err)
			}
			markerActive, err := stopMarkerActive(dir)
			if err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "after-journal-rename":
				if state.Stop || markerActive {
					t.Fatalf("journal-boundary kill unexpectedly committed STOP state: state=%+v marker=%v", state, markerActive)
				}
			case "after-state-rename":
				if !state.Stop || state.StopReason != "process-kill-direct-stop" || markerActive {
					t.Fatalf("state-boundary kill left unexpected STOP state: state=%+v marker=%v", state, markerActive)
				}
			case "after-marker-rename":
				if !state.Stop || state.StopReason != "process-kill-direct-stop" || !markerActive {
					t.Fatalf("marker-boundary kill left unexpected STOP state: state=%+v marker=%v", state, markerActive)
				}
			}

			if code, response := missionBinaryCall(t, binary, "mission", "status", dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
				t.Fatalf("fresh process exposed interrupted direct STOP: code=%d response=%+v", code, response)
			}
			missingInput := filepath.Join(t.TempDir(), "missing-artifact.json")
			if code, response := missionBinaryCall(t, binary, "mission", "m11-register", dir, "PRODUCTION_LEASE", missingInput); code == 0 || response["status"] != "STOPPED" {
				t.Fatalf("locked writer did not recover direct STOP before retaining STOP: code=%d response=%+v", code, response)
			}
			if _, err := os.Stat(journalPath); !os.IsNotExist(err) {
				t.Fatalf("direct STOP journal remains after locked recovery: %v", err)
			}
			state, err = loadMissionState(dir)
			if err != nil || !state.Stop || state.StopReason != "process-kill-direct-stop" {
				t.Fatalf("direct STOP recovery did not retain exact state: state=%+v err=%v", state, err)
			}
			if markerActive, err = stopMarkerActive(dir); err != nil || !markerActive {
				t.Fatalf("direct STOP recovery did not retain marker: active=%v err=%v", markerActive, err)
			}

			stateBefore, err := os.ReadFile(missionStatePath(dir))
			if err != nil {
				t.Fatal(err)
			}
			markerBefore, err := os.ReadFile(filepath.Join(dir, "STOP"))
			if err != nil {
				t.Fatal(err)
			}
			if code, response := missionCall(t, "m11-stop", dir, "different-stop-reason"); code == 0 || response["status"] != "STOPPED" {
				t.Fatalf("exactly recovered STOP accepted a replacement reason: code=%d response=%+v", code, response)
			}
			stateAfter, err := os.ReadFile(missionStatePath(dir))
			if err != nil || !bytes.Equal(stateBefore, stateAfter) {
				t.Fatalf("replacement STOP changed mission state: before=%s after=%s err=%v", stateBefore, stateAfter, err)
			}
			markerAfter, err := os.ReadFile(filepath.Join(dir, "STOP"))
			if err != nil || !bytes.Equal(markerBefore, markerAfter) {
				t.Fatalf("replacement STOP changed marker: before=%s after=%s err=%v", markerBefore, markerAfter, err)
			}
		})
	}
}
