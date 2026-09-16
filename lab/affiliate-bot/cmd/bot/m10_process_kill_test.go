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
// the child at a selected M10 recovery-journal boundary. Production never
// reads these variables.
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
	terminationPhase := os.Getenv("GO_M10_PROCESS_TERMINATION_PHASE")
	if terminationPhase == "" {
		// Preserve the original boundary for the existing cases: canary and
		// cost-bound die before their first canonical artifact, while execution
		// dies after its journal publish.
		terminationPhase = "before_artifact"
		if kind == "execution" {
			terminationPhase = "after_journal_publish"
		}
	}
	terminate := func(phase string) error {
		if phase != terminationPhase {
			return nil
		}
		if err := syscall.Kill(os.Getpid(), syscall.SIGKILL); err != nil {
			return err
		}
		// Keep the child at the injected boundary if signal delivery is deferred
		// until the next scheduling point on the host. SIGKILL is non-catchable;
		// this loop only prevents it from crossing the journal cleanup code.
		for {
			runtime.Gosched()
		}
	}
	switch kind {
	case "canary":
		m10CanaryAppendFault = terminate
	case "cost-bound":
		m10CostBoundAppendFault = terminate
	case "execution":
		m10ExecutionJournalPublishFault = terminate
	default:
		os.Exit(2)
	}
	os.Exit(runMissionCommand(os.Args[separator+1:], os.Stdout, os.Stderr))
}

func TestMissionM10ExecutionProcessKillAfterJournalPublishRequiresLockedReplay(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("SIGKILL process-boundary proof is not portable on Windows")
	}
	runtimeDir, boundPath, gatePath, _ := authorityExpiryFixture(t, "cost")
	root := filepath.Dir(runtimeDir)
	authorizationPath := filepath.Join(root, "process-kill-execution-authorization.json")
	if code, response := missionCall(t, "m10-authorize", runtimeDir, boundPath, gatePath, authorizationPath, "2026-09-08T00:00:00Z", "fixture_stub"); code != 0 || response["status"] != "AUTHORIZED" {
		t.Fatalf("execution authorization setup failed: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m10-reserve-authorization", runtimeDir, authorizationPath, "process-kill-execution-reservation"); code != 0 || response["status"] != "RESERVED" {
		t.Fatalf("execution reservation setup failed: code=%d response=%+v", code, response)
	}
	recordPath := filepath.Join(root, "process-kill-execution-record.json")
	journalPath := m10ExecutionJournalPath(runtimeDir)
	commandArgs := []string{"m10-record-failed", runtimeDir, authorizationPath, recordPath, "2026-09-08T00:00:00Z", "process-kill execution fixture"}
	command := exec.Command(os.Args[0], append([]string{"-test.run=^TestM10ProcessTerminationChild$", "--"}, commandArgs...)...)
	command.Env = append(os.Environ(), "GO_WANT_M10_PROCESS_TERMINATION=1", "GO_M10_PROCESS_TERMINATION_KIND=execution")
	var childStdout, childStderr bytes.Buffer
	command.Stdout, command.Stderr = &childStdout, &childStderr
	err := command.Run()
	if err == nil {
		t.Fatalf("M10 execution child unexpectedly completed after injected SIGKILL: stdout=%q stderr=%q", childStdout.String(), childStderr.String())
	}
	exited, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatal(err)
	}
	status, ok := exited.ProcessState.Sys().(syscall.WaitStatus)
	if !ok || !status.Signaled() || status.Signal() != syscall.SIGKILL {
		t.Fatalf("M10 execution child did not receive SIGKILL: %v stdout=%q stderr=%q", err, childStdout.String(), childStderr.String())
	}
	if _, err := os.Stat(journalPath); err != nil {
		t.Fatalf("process kill did not leave the synced M10 execution journal: %v", err)
	}
	if _, err := os.Stat(recordPath); !os.IsNotExist(err) {
		t.Fatalf("execution journal process kill created portable output before replay: %v", err)
	}
	state, err := loadMissionState(runtimeDir)
	if err != nil || len(state.Reservations) != 1 || state.Reservations[0].ExecutionID != "" {
		t.Fatalf("execution journal process kill changed reservation before replay: state=%+v err=%v", state, err)
	}
	entries, err := loadM10ArtifactRegistry(runtimeDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.ArtifactKind == corem10.ArtifactKindExecutionRecord {
			t.Fatalf("execution registry entry became visible before replay: %+v", entry)
		}
	}
	binary := buildMissionBinary(t)
	if code, response := missionBinaryCall(t, binary, "mission", "status", runtimeDir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
		t.Fatalf("fresh read-only status exposed interrupted M10 execution: code=%d response=%+v", code, response)
	}
	if code, response := missionBinaryCall(t, binary, append([]string{"mission"}, commandArgs...)...); code != 0 || response["status"] != "APPENDED" {
		t.Fatalf("locked retry did not replay M10 execution journal exactly: code=%d response=%+v", code, response)
	}
	if _, err := os.Stat(journalPath); !os.IsNotExist(err) {
		t.Fatalf("M10 execution journal remains after locked replay: %v", err)
	}
	if _, err := os.Stat(recordPath); err != nil {
		t.Fatalf("locked replay did not publish portable execution record: %v", err)
	}
	state, err = loadMissionState(runtimeDir)
	if err != nil || len(state.Reservations) != 1 || state.Reservations[0].ExecutionID == "" {
		t.Fatalf("locked replay did not bind execution to reservation: state=%+v err=%v", state, err)
	}
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

// TestMissionM10ProcessKillAfterCanonicalAppendRequiresLockedReplay exercises
// the later sides of the M10 two-store journal transitions. The child dies
// after the immutable artifact, or after the mutable state/index, is visible
// but before journal cleanup. A fresh reader must fail closed and a locked
// retry must converge without duplicating either store.
func TestMissionM10ProcessKillAfterCanonicalAppendRequiresLockedReplay(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("SIGKILL process-boundary proof is not portable on Windows")
	}
	base := time.Now().UTC().Truncate(time.Second)
	for _, scenario := range []struct {
		name           string
		kind           string
		phase          string
		setupCanary    bool
		artifactKind   string
		expectedStatus string
	}{
		{name: "canary-after-artifact", kind: "canary", phase: "after_artifact", artifactKind: corem10.ArtifactKindCanaryGrant, expectedStatus: "ACK"},
		{name: "canary-after-state", kind: "canary", phase: "after_state", artifactKind: corem10.ArtifactKindCanaryGrant, expectedStatus: "ACK"},
		{name: "cost-bound-after-artifact", kind: "cost-bound", phase: "after_artifact", setupCanary: true, artifactKind: corem10.ArtifactKindTrustedCostBound, expectedStatus: "EXACT_DUPLICATE"},
		{name: "cost-bound-after-index", kind: "cost-bound", phase: "after_index", setupCanary: true, artifactKind: corem10.ArtifactKindTrustedCostBound, expectedStatus: "EXACT_DUPLICATE"},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			runtimeDir, inputPath, _, _ := authorityFixtureAt(t, base, "none", scenario.setupCanary, 1)
			var expectedArtifactID string
			if scenario.kind == "canary" {
				var grant corem10.CanaryGrant
				if err := readJSON(inputPath, &grant); err != nil {
					t.Fatal(err)
				}
				expectedArtifactID = grant.GrantID
			}
			if scenario.kind == "cost-bound" {
				var bound corem10.TrustedCostBound
				if err := readJSON(inputPath, &bound); err != nil {
					t.Fatal(err)
				}
				bound.CostBoundID = "process-kill-" + scenario.phase
				bound.CostBoundHash = corem10.ComputeTrustedCostBoundHash(bound)
				expectedArtifactID = bound.CostBoundID
				newInput := filepath.Join(filepath.Dir(inputPath), "process-kill-"+scenario.phase+".json")
				writeMissionTestJSON(t, newInput, bound)
				inputPath = newInput
			}

			commandArgs := []string{"m10-canary", runtimeDir, inputPath}
			journalPath := m10CanaryJournalPath(runtimeDir)
			if scenario.kind == "cost-bound" {
				commandArgs = []string{"m10-cost-register", runtimeDir, inputPath}
				journalPath = m10CostBoundJournalPath(runtimeDir)
			}
			command := exec.Command(os.Args[0], append([]string{"-test.run=^TestM10ProcessTerminationChild$", "--"}, commandArgs...)...)
			command.Env = append(os.Environ(),
				"GO_WANT_M10_PROCESS_TERMINATION=1",
				"GO_M10_PROCESS_TERMINATION_KIND="+scenario.kind,
				"GO_M10_PROCESS_TERMINATION_PHASE="+scenario.phase,
			)
			var childStdout, childStderr bytes.Buffer
			command.Stdout, command.Stderr = &childStdout, &childStderr
			err := command.Run()
			if err == nil {
				t.Fatalf("M10 child unexpectedly completed after injected %s SIGKILL", scenario.phase)
			}
			exited, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatal(err)
			}
			status, ok := exited.ProcessState.Sys().(syscall.WaitStatus)
			if !ok || !status.Signaled() || status.Signal() != syscall.SIGKILL {
				t.Fatalf("M10 child did not receive SIGKILL at %s: %v stdout=%q stderr=%q", scenario.phase, err, childStdout.String(), childStderr.String())
			}
			if _, err := os.Stat(journalPath); err != nil {
				t.Fatalf("post-%s process kill did not leave the recovery journal: %v stdout=%q stderr=%q", scenario.phase, err, childStdout.String(), childStderr.String())
			}

			binary := buildMissionBinary(t)
			if code, response := missionBinaryCall(t, binary, "mission", "status", runtimeDir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
				t.Fatalf("fresh read-only status exposed interrupted %s transition: code=%d response=%+v", scenario.name, code, response)
			}

			entries, err := loadM10ArtifactRegistry(runtimeDir)
			if err != nil {
				t.Fatal(err)
			}
			artifactCount := 0
			for _, entry := range entries {
				if entry.ArtifactKind == scenario.artifactKind && entry.ArtifactID == expectedArtifactID {
					artifactCount++
				}
			}
			if scenario.phase == "after_artifact" && artifactCount == 0 {
				t.Fatalf("post-artifact process kill hid the canonical %s artifact", scenario.kind)
			}

			if scenario.kind == "canary" {
				state, err := loadMissionState(runtimeDir)
				if err != nil {
					t.Fatal(err)
				}
				if scenario.phase == "after_artifact" && state.Canary != nil {
					t.Fatal("post-artifact process kill unexpectedly published the canary state binding")
				}
				if scenario.phase == "after_state" && state.Canary == nil {
					t.Fatal("post-state process kill lost the visible canary state binding")
				}
			}
			if scenario.kind == "cost-bound" && scenario.phase == "after_artifact" {
				bounds, err := loadTrustedCostBounds(runtimeDir)
				if err != nil {
					t.Fatal(err)
				}
				for _, bound := range bounds {
					if bound.CostBoundID == expectedArtifactID {
						t.Fatal("post-artifact process kill unexpectedly published the compact cost-bound index")
					}
				}
			}

			if code, response := missionBinaryCall(t, binary, append([]string{"mission"}, commandArgs...)...); code != 0 || response["status"] != scenario.expectedStatus {
				t.Fatalf("locked retry did not converge %s transition: code=%d response=%+v", scenario.name, code, response)
			}
			if _, err := os.Stat(journalPath); !os.IsNotExist(err) {
				t.Fatalf("%s recovery journal remains after locked retry: %v", scenario.name, err)
			}
			entries, err = loadM10ArtifactRegistry(runtimeDir)
			if err != nil {
				t.Fatal(err)
			}
			finalArtifactCount := 0
			for _, entry := range entries {
				if entry.ArtifactKind == scenario.artifactKind && entry.ArtifactID == expectedArtifactID {
					finalArtifactCount++
				}
			}
			if finalArtifactCount != 1 {
				t.Fatalf("%s recovery left %d canonical %s artifacts, want 1", scenario.name, finalArtifactCount, scenario.artifactKind)
			}
			if scenario.kind == "cost-bound" {
				bounds, err := loadTrustedCostBounds(runtimeDir)
				if err != nil {
					t.Fatal(err)
				}
				count := 0
				for _, bound := range bounds {
					if bound.CostBoundID == expectedArtifactID {
						count++
					}
				}
				if count != 1 {
					t.Fatalf("%s recovery left %d compact cost-bound index entries, want 1", scenario.name, count)
				}
			}
		})
	}
}

// TestMissionM10ExecutionProcessKillAfterCanonicalAppendRequiresLockedReplay
// covers the two canonical sides of a governed execution transition. The
// child dies after the execution registry append or reservation binding is
// visible, but before journal cleanup. A fresh reader must remain fail-closed;
// a locked retry must preserve one execution record and one reservation link.
func TestMissionM10ExecutionProcessKillAfterCanonicalAppendRequiresLockedReplay(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("SIGKILL process-boundary proof is not portable on Windows")
	}
	for _, phase := range []string{"after_artifact", "after_state"} {
		phase := phase
		t.Run(phase, func(t *testing.T) {
			runtimeDir, boundPath, gatePath, _ := authorityExpiryFixture(t, "cost")
			root := filepath.Dir(runtimeDir)
			authorizationPath := filepath.Join(root, "process-kill-execution-"+phase+"-authorization.json")
			if code, response := missionCall(t, "m10-authorize", runtimeDir, boundPath, gatePath, authorizationPath, "2026-09-08T00:00:00Z", "fixture_stub"); code != 0 || response["status"] != "AUTHORIZED" {
				t.Fatalf("execution authorization setup failed: code=%d response=%+v", code, response)
			}
			if code, response := missionCall(t, "m10-reserve-authorization", runtimeDir, authorizationPath, "process-kill-execution-"+phase+"-reservation"); code != 0 || response["status"] != "RESERVED" {
				t.Fatalf("execution reservation setup failed: code=%d response=%+v", code, response)
			}
			recordPath := filepath.Join(root, "process-kill-execution-"+phase+"-record.json")
			journalPath := m10ExecutionJournalPath(runtimeDir)
			commandArgs := []string{"m10-record-failed", runtimeDir, authorizationPath, recordPath, "2026-09-08T00:00:00Z", "process-kill canonical append fixture"}
			command := exec.Command(os.Args[0], append([]string{"-test.run=^TestM10ProcessTerminationChild$", "--"}, commandArgs...)...)
			command.Env = append(os.Environ(),
				"GO_WANT_M10_PROCESS_TERMINATION=1",
				"GO_M10_PROCESS_TERMINATION_KIND=execution",
				"GO_M10_PROCESS_TERMINATION_PHASE="+phase,
			)
			var childStdout, childStderr bytes.Buffer
			command.Stdout, command.Stderr = &childStdout, &childStderr
			err := command.Run()
			if err == nil {
				t.Fatalf("M10 execution child unexpectedly completed after injected %s SIGKILL", phase)
			}
			exited, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatal(err)
			}
			status, ok := exited.ProcessState.Sys().(syscall.WaitStatus)
			if !ok || !status.Signaled() || status.Signal() != syscall.SIGKILL {
				t.Fatalf("M10 execution child did not receive SIGKILL at %s: %v stdout=%q stderr=%q", phase, err, childStdout.String(), childStderr.String())
			}
			journalRaw, err := os.ReadFile(journalPath)
			if err != nil {
				t.Fatalf("%s process kill did not leave the recovery journal: %v", phase, err)
			}
			var journal m10ExecutionJournal
			if err := json.Unmarshal(journalRaw, &journal); err != nil {
				t.Fatalf("decode execution recovery journal: %v", err)
			}
			if journal.Record.ExecutionID == "" {
				t.Fatal("execution recovery journal has no canonical execution ID")
			}
			if _, err := os.Stat(recordPath); !os.IsNotExist(err) {
				t.Fatalf("execution journal process kill created portable output before replay: %v", err)
			}

			entries, err := loadM10ArtifactRegistry(runtimeDir)
			if err != nil {
				t.Fatal(err)
			}
			executionCount := 0
			for _, entry := range entries {
				if entry.ArtifactKind == corem10.ArtifactKindExecutionRecord {
					executionCount++
					if entry.ArtifactID != journal.Record.ExecutionID {
						t.Fatalf("execution registry contains the wrong canonical ID: entry=%+v journal=%s", entry, journal.Record.ExecutionID)
					}
				}
			}
			if executionCount != 1 {
				t.Fatalf("%s process kill left %d canonical execution records, want 1", phase, executionCount)
			}
			state, err := loadMissionState(runtimeDir)
			if err != nil {
				t.Fatal(err)
			}
			if len(state.Reservations) != 1 {
				t.Fatalf("unexpected reservations after %s process kill: %+v", phase, state.Reservations)
			}
			if phase == "after_artifact" && state.Reservations[0].ExecutionID != "" {
				t.Fatalf("execution reservation binding became visible too early: %+v", state.Reservations[0])
			}
			if phase == "after_state" && state.Reservations[0].ExecutionID != journal.Record.ExecutionID {
				t.Fatalf("execution reservation binding was not visible at %s: %+v", phase, state.Reservations[0])
			}

			binary := buildMissionBinary(t)
			if code, response := missionBinaryCall(t, binary, "mission", "status", runtimeDir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
				t.Fatalf("fresh read-only status exposed interrupted %s execution: code=%d response=%+v", phase, code, response)
			}
			if code, response := missionBinaryCall(t, binary, append([]string{"mission"}, commandArgs...)...); code != 0 || response["status"] != "APPENDED" {
				t.Fatalf("locked retry did not replay %s execution journal: code=%d response=%+v", phase, code, response)
			}
			if _, err := os.Stat(journalPath); !os.IsNotExist(err) {
				t.Fatalf("%s execution journal remains after locked retry: %v", phase, err)
			}
			if _, err := os.Stat(recordPath); err != nil {
				t.Fatalf("locked retry did not publish portable execution record: %v", err)
			}

			entries, err = loadM10ArtifactRegistry(runtimeDir)
			if err != nil {
				t.Fatal(err)
			}
			executionCount = 0
			for _, entry := range entries {
				if entry.ArtifactKind == corem10.ArtifactKindExecutionRecord && entry.ArtifactID == journal.Record.ExecutionID {
					executionCount++
				}
			}
			if executionCount != 1 {
				t.Fatalf("%s locked replay left %d canonical execution records, want 1", phase, executionCount)
			}
			state, err = loadMissionState(runtimeDir)
			if err != nil || len(state.Reservations) != 1 || state.Reservations[0].ExecutionID != journal.Record.ExecutionID {
				t.Fatalf("%s locked replay did not retain one reservation-to-execution link: state=%+v err=%v", phase, state, err)
			}
		})
	}
}
