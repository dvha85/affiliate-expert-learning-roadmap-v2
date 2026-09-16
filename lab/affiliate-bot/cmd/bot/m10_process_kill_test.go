package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"

	corem10 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m10"
)

// TestM10ProcessTerminationChild is a test-binary-only entrypoint. It kills
// the child after the M10 recovery journal is synced and visible, but before
// the first canonical side effect. Production never reads these variables.
func TestM10ProcessTerminationChild(t *testing.T) {
	if os.Getenv("GO_WANT_M10_PROCESS_TERMINATION") != "1" {
		return
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
	kind := os.Getenv("GO_M10_PROCESS_TERMINATION_KIND")
	terminate := func(phase string) error {
		if phase != "before_artifact" {
			return nil
		}
		if err := syscall.Kill(os.Getpid(), syscall.SIGKILL); err != nil {
			return err
		}
		return nil
	}
	switch kind {
	case "canary":
		m10CanaryAppendFault = terminate
	case "cost-bound":
		m10CostBoundAppendFault = terminate
	default:
		os.Exit(2)
	}
	os.Exit(runMissionCommand(os.Args[separator+1:], os.Stdout, os.Stderr))
}

func TestMissionM10ProcessKillAfterJournalPublishRequiresLockedReplay(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("SIGKILL process-boundary proof is not portable on Windows")
	}
	base := time.Now().UTC().Truncate(time.Second)
	for _, scenario := range []struct {
		name string
		kind string
	}{
		{name: "canary", kind: "canary"},
		{name: "cost-bound", kind: "cost-bound"},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			var runtimeDir, inputPath, journalPath string
			if scenario.kind == "canary" {
				runtimeDir, inputPath, _, _ = authorityFixtureAt(t, base, "none", false, 1)
				journalPath = m10CanaryJournalPath(runtimeDir)
			} else {
				runtimeDir, inputPath, _, _ = authorityFixtureAt(t, base, "none", true, 1)
				var bound corem10.TrustedCostBound
				if err := readJSON(inputPath, &bound); err != nil {
					t.Fatal(err)
				}
				bound.CostBoundID = "process-kill-cost-bound"
				bound.CostBoundHash = corem10.ComputeTrustedCostBoundHash(bound)
				newInput := filepath.Join(filepath.Dir(inputPath), "process-kill-cost-bound.json")
				writeMissionTestJSON(t, newInput, bound)
				inputPath = newInput
				journalPath = m10CostBoundJournalPath(runtimeDir)
			}

			commandArgs := []string{"m10-canary", runtimeDir, inputPath}
			if scenario.kind == "cost-bound" {
				commandArgs = []string{"m10-cost-register", runtimeDir, inputPath}
			}
			command := exec.Command(os.Args[0], append([]string{"-test.run=^TestM10ProcessTerminationChild$", "--"}, commandArgs...)...)
			command.Env = append(os.Environ(), "GO_WANT_M10_PROCESS_TERMINATION=1", "GO_M10_PROCESS_TERMINATION_KIND="+scenario.kind)
			var childStdout, childStderr bytes.Buffer
			command.Stdout, command.Stderr = &childStdout, &childStderr
			err := command.Run()
			if err == nil {
				t.Fatal("M10 child unexpectedly completed after injected SIGKILL")
			}
			exited, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatal(err)
			}
			status, ok := exited.ProcessState.Sys().(syscall.WaitStatus)
			if !ok || !status.Signaled() || status.Signal() != syscall.SIGKILL {
				t.Fatalf("M10 child did not receive SIGKILL: %v stdout=%q stderr=%q", err, childStdout.String(), childStderr.String())
			}
			if _, err := os.Stat(journalPath); err != nil {
				t.Fatalf("process kill did not leave the synced M10 journal: %v", err)
			}
			if code, response := missionCall(t, "status", runtimeDir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
				t.Fatalf("read-only status exposed interrupted M10 transition: code=%d response=%+v", code, response)
			}
			if scenario.kind == "canary" {
				if code, response := missionCall(t, "m10-canary", runtimeDir, inputPath); code != 0 || response["status"] != "ACK" {
					t.Fatalf("locked canary retry did not replay exact journal: code=%d response=%+v", code, response)
				}
				if _, err := os.Stat(journalPath); !os.IsNotExist(err) {
					t.Fatalf("canary journal remains after exact replay: %v", err)
				}
			} else {
				if code, response := missionCall(t, "m10-cost-register", runtimeDir, inputPath); code != 0 || response["status"] != "EXACT_DUPLICATE" {
					t.Fatalf("locked cost-bound retry did not replay exact journal: code=%d response=%+v", code, response)
				}
				if _, err := os.Stat(journalPath); !os.IsNotExist(err) {
					t.Fatalf("cost-bound journal remains after exact replay: %v", err)
				}
				var bound corem10.TrustedCostBound
				if err := readJSON(inputPath, &bound); err != nil {
					t.Fatal(err)
				}
				raw, err := json.Marshal(bound)
				if err != nil || !resolveM10Artifact(runtimeDir, corem10.ArtifactKindTrustedCostBound, raw) {
					t.Fatalf("replayed cost bound is not in canonical registry: err=%v", err)
				}
			}
		})
	}
}
