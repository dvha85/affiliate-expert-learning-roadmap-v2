package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
	corem10 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m10"
	corem11 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m11"
)

func TestM11RegistryAfterWriteFailureRecoversAsExactDuplicate(t *testing.T) {
	dir := t.TempDir()
	lease := corem11.ProductionLease{LeaseID: "fault-lease", LeaseVersion: "v1", PolicyVersion: "policy-v1", ApprovalRef: "fault-approval", ReviewedBy: "human", ReviewerID: "reviewer", ReviewedAt: "2026-09-08T00:00:00Z", PromotionReviewRef: "review", SourceCanaryGrantID: "canary", SourceCanaryGrantVersion: "v1", SourceCanaryGrantHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ValidFrom: "2026-09-08T00:00:00Z", ExpiresAt: "2099-09-08T00:00:00Z", AllowedRiskClasses: []string{"RISK0"}, AllowedActionTypes: []string{"DRAFT"}, AllowedHosts: []string{"example.com"}, ExecutorIDs: []string{"fixture_stub"}, MaxExecutionsTotal: 1, MaxExecutionsPerWindow: 1, WindowSeconds: 60, MaxCostMinorTotal: 1, Currency: "USD", MaxPendingOutcomes: 1, MaxConsecutiveFailures: 1, MaxOutcomeAgeSeconds: 60, MaxHealthSnapshotAgeSeconds: 60, KillSwitchRequired: true, CorrelationID: "fault-correlation", HashVersion: "go-json-v1"}
	lease.LeaseHash = corem11.ComputeProductionLeaseHash(lease)
	raw, err := json.Marshal(lease)
	if err != nil {
		t.Fatal(err)
	}
	m11RegistryAppendFault = func(phase string, _ corem11.ArtifactEntry) error {
		if phase == "after_write" {
			return errors.New("injected after-write failure")
		}
		return nil
	}
	t.Cleanup(func() { m11RegistryAppendFault = nil })
	if _, _, err := registerM11Artifact(dir, corem11.ArtifactKindLease, raw); err == nil {
		t.Fatal("after-write failure was not surfaced")
	}
	entries, err := loadM11ArtifactRegistry(dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("written artifact was not recoverable: entries=%d err=%v", len(entries), err)
	}
	m11RegistryAppendFault = nil
	if _, status, err := registerM11Artifact(dir, corem11.ArtifactKindLease, raw); err != nil || status != appendDuplicate {
		t.Fatalf("retry must be exact duplicate: status=%s err=%v", status, err)
	}
}

func TestM11JournalSymlinkFailsClosedBeforeRecoveryOrMutation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions are not portable on Windows")
	}
	dir := t.TempDir()
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("init failed: code=%d response=%+v", code, response)
	}
	external := filepath.Join(t.TempDir(), "untrusted-journal.json")
	if err := os.WriteFile(external, []byte(`{"version":"m11-unknown-stop-journal/v1"}`), 0600); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(external)
	if err != nil {
		t.Fatal(err)
	}
	binary := buildMissionBinary(t)
	for _, journal := range []string{m11OutcomeJournalPath(dir), m11FailedExecutionJournalPath(dir), m11UnknownStopJournalPath(dir)} {
		if err := os.Symlink(external, journal); err != nil {
			t.Fatal(err)
		}
		if err := m11JournalRecoveryRequired(dir); err == nil {
			t.Fatal("symlinked M11 journal was not rejected before recovery")
		}
		if code, response := missionBinaryCall(t, binary, "mission", "status", dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
			t.Fatalf("status did not fail closed for %s: code=%d response=%+v", filepath.Base(journal), code, response)
		}
		if code, response := missionBinaryCall(t, binary, "mission", "m11-register", dir, corem11.ArtifactKindLease, filepath.Join(t.TempDir(), "unread-input.json")); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
			t.Fatalf("writer reached input handling for %s: code=%d response=%+v", filepath.Base(journal), code, response)
		}
		if _, err := os.Stat(m11ArtifactRegistryPath(dir)); !os.IsNotExist(err) {
			t.Fatalf("symlinked journal caused registry mutation: %v", err)
		}
		after, err := os.ReadFile(external)
		if err != nil || string(after) != string(before) {
			t.Fatalf("recovery touched external journal target: err=%v before=%q after=%q", err, before, after)
		}
		if err := os.Remove(journal); err != nil {
			t.Fatal(err)
		}
	}
}

func TestM11RecoveryAdmissionFailsClosedWhileOldRuntimeGateIsHeld(t *testing.T) {
	oldDir, newDir := t.TempDir(), t.TempDir()
	if code, response := missionCall(t, "init", oldDir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("init old runtime: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "init", newDir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("init new runtime: code=%d response=%+v", code, response)
	}
	release, err := acquireRuntimeGate(oldDir)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	_, _, err = admitM11Recovery(newDir, oldDir, filepath.Join(t.TempDir(), "unread-handoff.json"), nil)
	if err == nil || !strings.Contains(err.Error(), "runtime is busy") {
		t.Fatalf("recovery admission read an old runtime while its writer gate was held: %v", err)
	}
	if _, err := os.Stat(m11ArtifactRegistryPath(newDir)); !os.IsNotExist(err) {
		t.Fatalf("blocked recovery admission mutated new runtime: %v", err)
	}
}

func TestM11RecoveryExportFailsClosedWhileRuntimeGateIsHeld(t *testing.T) {
	dir, output := t.TempDir(), filepath.Join(t.TempDir(), "handoff.json")
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("init runtime: code=%d response=%+v", code, response)
	}
	release, err := acquireRuntimeGate(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if code, response := missionCall(t, "m11-recovery-export", dir, "unread-resolution", "unread-ledger", output); code == 0 || response["status"] != "BUSY" {
		t.Fatalf("recovery export read a runtime while its writer gate was held: code=%d response=%+v", code, response)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("blocked recovery export wrote a handoff: %v", err)
	}
}

func TestM11ReadCommandsFailClosedWhileRuntimeGateIsHeld(t *testing.T) {
	dir := t.TempDir()
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("init runtime: code=%d response=%+v", code, response)
	}
	_, ledgerEntry := setupM11OutcomeJournalFixture(t, dir)
	release, err := acquireRuntimeGate(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if code, response := missionCall(t, "m11-resolve", dir, corem11.ArtifactKindLedger, ledgerEntry.ArtifactID); code == 0 || response["status"] != "BUSY" {
		t.Fatalf("M11 resolver read a runtime while its writer gate was held: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "status", dir); code == 0 || response["status"] != "BUSY" {
		t.Fatalf("mission status read a runtime while its writer gate was held: code=%d response=%+v", code, response)
	}
}

// A journal may survive a process exit after it has been synced but before the
// corresponding registry/state transition is complete. A fresh Bot process
// must expose only RECOVERY_REQUIRED until its locked writer prelude replays
// the exact journal; a read-only status call must never perform that replay.
func TestM11InterruptedJournalFreshProcessFailsClosedThenLockedWriterRecovers(t *testing.T) {
	binary := buildMissionBinary(t)
	const attemptedAt = "2026-09-08T00:00:02Z"
	t.Run("unknown_stop", func(t *testing.T) {
		fixture := newM11UnknownStopFixture(t)
		m11RegistryAppendFault = func(phase string, entry corem11.ArtifactEntry) error {
			if entry.ArtifactKind == corem11.ArtifactKindLedger && phase == "before_write" {
				return errors.New("injected interrupted UNKNOWN ledger write")
			}
			return nil
		}
		if _, _, _, err := recordUnknownM11Execution(fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, "fresh-process fixture"); err == nil {
			t.Fatal("interrupted UNKNOWN transition was not surfaced")
		}
		m11RegistryAppendFault = nil
		if code, response := missionBinaryCall(t, binary, "mission", "status", fixture.dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
			t.Fatalf("fresh process exposed UNKNOWN transition: code=%d response=%+v", code, response)
		}
		if code, response := missionBinaryCall(t, binary, "mission", "m11-register", fixture.dir, corem11.ArtifactKindLease, filepath.Join(t.TempDir(), "unread.json")); code == 0 || response["status"] != "STOPPED" {
			t.Fatalf("locked writer did not recover then retain STOP: code=%d response=%+v", code, response)
		}
		assertM11UnknownStopRecovered(t, fixture, attemptedAt, "fresh-process fixture")
	})
	t.Run("failed_execution", func(t *testing.T) {
		fixture := newM11UnknownStopFixture(t)
		m11RegistryAppendFault = func(phase string, entry corem11.ArtifactEntry) error {
			if entry.ArtifactKind == corem11.ArtifactKindLedger && phase == "before_write" {
				return errors.New("injected interrupted FAILED ledger write")
			}
			return nil
		}
		if _, _, _, err := recordFailedM11Execution(fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, "fresh-process fixture"); err == nil {
			t.Fatal("interrupted FAILED transition was not surfaced")
		}
		m11RegistryAppendFault = nil
		if code, response := missionBinaryCall(t, binary, "mission", "status", fixture.dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
			t.Fatalf("fresh process exposed FAILED transition: code=%d response=%+v", code, response)
		}
		if code, response := missionBinaryCall(t, binary, "mission", "m11-register", fixture.dir, corem11.ArtifactKindLease, filepath.Join(t.TempDir(), "unread.json")); code == 0 || response["status"] != "INPUT_ERROR" {
			t.Fatalf("locked writer did not recover FAILED journal before input handling: code=%d response=%+v", code, response)
		}
		if _, err := os.Stat(m11FailedExecutionJournalPath(fixture.dir)); !os.IsNotExist(err) {
			t.Fatalf("fresh-process FAILED recovery left journal: %v", err)
		}
		_, ledger, err := m11LedgerHead(fixture.dir, fixture.lease.LeaseID)
		if err != nil || ledger.ConsecutiveFailures != 1 || ledger.LastExecutionAt != attemptedAt {
			t.Fatalf("fresh-process FAILED recovery did not preserve exact ledger: ledger=%+v err=%v", ledger, err)
		}
	})
	t.Run("fixture_outcome", func(t *testing.T) {
		dir := t.TempDir()
		if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
			t.Fatalf("init failed: code=%d response=%+v", code, response)
		}
		journal, _ := setupM11OutcomeJournalFixture(t, dir)
		if err := writeJSONAtomic(m11OutcomeJournalPath(dir), journal); err != nil {
			t.Fatal(err)
		}
		if code, response := missionBinaryCall(t, binary, "mission", "status", dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
			t.Fatalf("fresh process exposed outcome transition: code=%d response=%+v", code, response)
		}
		if code, response := missionBinaryCall(t, binary, "mission", "m11-register", dir, corem11.ArtifactKindLease, filepath.Join(t.TempDir(), "unread.json")); code == 0 || response["status"] != "INPUT_ERROR" {
			t.Fatalf("locked writer did not recover outcome journal before input handling: code=%d response=%+v", code, response)
		}
		if _, err := os.Stat(m11OutcomeJournalPath(dir)); !os.IsNotExist(err) {
			t.Fatalf("fresh-process outcome recovery left journal: %v", err)
		}
		outcomes, err := loadM11FixtureOutcomes(dir)
		if err != nil || len(outcomes) != 1 || outcomes[0].OutcomeID != journal.Outcome.OutcomeID {
			t.Fatalf("fresh-process outcome recovery was not exact: outcomes=%+v err=%v", outcomes, err)
		}
	})
}

type m11UnknownStopFixture struct {
	dir           string
	lease         corem11.ProductionLease
	authorization corem11.ProductionExecutionAuthorization
	ledgerEntry   corem11.ArtifactEntry
}

func newM11UnknownStopFixture(t *testing.T) m11UnknownStopFixture {
	t.Helper()
	dir := t.TempDir()
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("init failed: code=%d response=%+v", code, response)
	}
	lease := corem11.ProductionLease{LeaseID: "unknown-journal-lease", LeaseVersion: "v1", PolicyVersion: "policy-1", ApprovalRef: "unknown-journal-approval", ReviewedBy: "human", ReviewerID: "reviewer", ReviewedAt: "2026-09-08T00:00:00Z", PromotionReviewRef: "review", SourceCanaryGrantID: "canary", SourceCanaryGrantVersion: "v1", SourceCanaryGrantHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ValidFrom: "2026-09-08T00:00:00Z", ExpiresAt: "2099-09-08T00:00:00Z", AllowedRiskClasses: []string{"RISK0"}, AllowedActionTypes: []string{"DRAFT"}, AllowedHosts: []string{"example.com"}, ExecutorIDs: []string{"fixture_stub"}, MaxExecutionsTotal: 1, MaxExecutionsPerWindow: 1, WindowSeconds: 60, MaxCostMinorTotal: 10, Currency: "USD", MaxPendingOutcomes: 1, MaxConsecutiveFailures: 1, MaxOutcomeAgeSeconds: 60, MaxHealthSnapshotAgeSeconds: 60, KillSwitchRequired: true, CorrelationID: "unknown-journal-correlation", HashVersion: "go-json-v1"}
	lease.LeaseHash = corem11.ComputeProductionLeaseHash(lease)
	approval := corem11.ProductionLeaseApproval{ApprovalID: lease.ApprovalRef, LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, PromotionReviewRef: lease.PromotionReviewRef, SourceCanaryGrantID: lease.SourceCanaryGrantID, SourceCanaryGrantVersion: lease.SourceCanaryGrantVersion, SourceCanaryGrantHash: lease.SourceCanaryGrantHash, SourceE5Refs: []string{"fixture:e5"}, ValidatedRiskClasses: []string{"RISK0"}, ReviewedBy: "human", ReviewerID: lease.ReviewerID, ReviewedAt: lease.ReviewedAt, Decision: "APPROVE_PRODUCTION_LEASE"}
	health := corem11.ProductionHealthSnapshot{SnapshotID: "unknown-journal-health", LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ObservedAt: "2026-09-08T00:00:00Z", SourceRefs: []string{"fixture:health"}, DependencyState: "HEALTHY", TelemetryComplete: true, HashVersion: "go-json-v1"}
	health.SnapshotHash = corem11.ComputeProductionHealthHash(health)
	cost := corem10.TrustedCostBound{CostBoundID: "unknown-journal-cost", IntentID: "unknown-journal-intent", IntentHash: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", MaxCostMinor: 10, Currency: "USD", SourceRef: "fixture:cost", ObservedAt: "2026-09-08T00:00:00Z", ExpiresAt: "2099-09-08T00:00:00Z", CorrelationID: lease.CorrelationID, HashVersion: "go-json-v1"}
	cost.CostBoundHash = corem10.ComputeTrustedCostBoundHash(cost)
	executionID := "prod-exec-unknown-journal-auth"
	ledger := corem11.ProductionLedger{LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ControlMode: "NORMAL", WindowStartedAt: "2026-09-08T00:00:00Z", ExecutionsTotal: 1, ExecutionsInWindow: 1, CostMinorTotal: 10, PendingOutcomes: 1, PendingExecutionIDs: []string{executionID}, SuccessfulIdempotencyKeys: []string{}, OutcomeLinks: []corem11.ProductionOutcomeLink{}, ReconciliationResolutionIDs: []string{}, UpdatedAt: "2026-09-08T00:00:01Z"}
	ledgerRaw, err := json.Marshal(ledger)
	if err != nil {
		t.Fatal(err)
	}
	ledgerEntry, err := corem11.NewArtifactEntry(corem11.ArtifactKindLedger, ledgerRaw)
	if err != nil {
		t.Fatal(err)
	}
	gateLedger := ledger
	gateLedger.ExecutionsTotal, gateLedger.ExecutionsInWindow, gateLedger.CostMinorTotal, gateLedger.PendingOutcomes = 0, 0, 0, 0
	gateLedger.PendingExecutionIDs = []string{}
	gateLedger.UpdatedAt = "2026-09-08T00:00:00Z"
	gateLedgerRaw, err := json.Marshal(gateLedger)
	if err != nil {
		t.Fatal(err)
	}
	gateLedgerEntry, err := corem11.NewArtifactEntry(corem11.ArtifactKindLedger, gateLedgerRaw)
	if err != nil {
		t.Fatal(err)
	}
	activation := corem11.ProductionActivationRecord{LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ActivatedAt: "2026-09-08T00:00:00Z"}
	gate := corem11.ProductionGateDecision{GateID: "unknown-journal-gate", LedgerArtifactID: gateLedgerEntry.ArtifactID, LedgerContentHash: gateLedgerEntry.ContentHash, LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, IntentID: cost.IntentID, IntentHash: cost.IntentHash, PolicyVersion: lease.PolicyVersion, RiskClass: "RISK0", HealthSnapshotID: health.SnapshotID, HealthSnapshotHash: health.SnapshotHash, CostBoundID: cost.CostBoundID, CostBoundHash: cost.CostBoundHash, CostBoundMinor: cost.MaxCostMinor, Decision: "ALLOW_PRODUCTION", Reason: "fixture", EvaluatedAt: "2026-09-08T00:00:00Z"}
	authorization := corem11.ProductionExecutionAuthorization{AuthorizationID: "unknown-journal-auth", IntentID: cost.IntentID, IntentHash: cost.IntentHash, PolicyVersion: lease.PolicyVersion, ProductionLeaseID: lease.LeaseID, ProductionLeaseVersion: lease.LeaseVersion, ProductionLeaseHash: lease.LeaseHash, ProductionGateID: gate.GateID, ProductionHealthSnapshotID: health.SnapshotID, ProductionHealthSnapshotHash: health.SnapshotHash, ProductionCostBoundID: cost.CostBoundID, ProductionCostBoundHash: cost.CostBoundHash, ProductionCostBoundMinor: cost.MaxCostMinor, ExecutorID: "fixture_stub", AuthorizedAt: "2026-09-08T00:00:00Z", ExpiresAt: "2026-09-08T00:01:00Z", IdempotencyKey: "unknown-journal-key", CorrelationID: cost.CorrelationID, ExecutionMode: "GOVERNED_PRODUCTION", ExecutionAuthorized: true}
	for _, item := range []struct {
		kind  string
		value any
	}{{corem11.ArtifactKindLease, lease}, {corem11.ArtifactKindLeaseApproval, approval}, {corem11.ArtifactKindHealth, health}, {corem11.ArtifactKindCostBound, cost}, {corem11.ArtifactKindLedger, gateLedger}, {corem11.ArtifactKindActivation, activation}, {corem11.ArtifactKindGate, gate}, {corem11.ArtifactKindAuthorization, authorization}, {corem11.ArtifactKindLedger, ledger}} {
		registerM11TestArtifact(t, dir, item.kind, item.value)
	}
	return m11UnknownStopFixture{dir: dir, lease: lease, authorization: authorization, ledgerEntry: ledgerEntry}
}

func TestM11UnknownStopJournalRecoversAfterStoppedLedgerWriteFailure(t *testing.T) {
	for _, phase := range []string{"before_write", "after_write", "after_sync"} {
		faultPhase := phase
		t.Run(phase, func(t *testing.T) {
			fixture := newM11UnknownStopFixture(t)
			const attemptedAt, reason = "2026-09-08T00:00:02Z", "fixture timeout"
			m11RegistryAppendFault = func(registryPhase string, entry corem11.ArtifactEntry) error {
				if entry.ArtifactKind == corem11.ArtifactKindLedger && registryPhase == faultPhase {
					return errors.New("injected stopped-ledger write failure")
				}
				return nil
			}
			t.Cleanup(func() { m11RegistryAppendFault = nil })
			if _, _, _, err := recordUnknownM11Execution(fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, reason); err == nil {
				t.Fatal("stopped-ledger fault was not surfaced")
			}
			if _, err := os.Stat(m11UnknownStopJournalPath(fixture.dir)); err != nil {
				t.Fatalf("unknown STOP journal was removed before recovery: %v", err)
			}
			if code, response := missionCall(t, "status", fixture.dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
				t.Fatalf("status did not fail closed on unknown STOP journal: code=%d response=%+v", code, response)
			}
			if code, response := missionCall(t, "m11-resolve", fixture.dir, corem11.ArtifactKindLedger, fixture.ledgerEntry.ArtifactID); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
				t.Fatalf("M11 resolver exposed an UNKNOWN-to-STOP partial transition: code=%d response=%+v", code, response)
			}
			exportPath := filepath.Join(fixture.dir, "partial-recovery-handoff.json")
			if code, response := missionCall(t, "m11-recovery-export", fixture.dir, "missing-resolution", fixture.ledgerEntry.ArtifactID, exportPath); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
				t.Fatalf("M11 recovery export exposed a partial transition: code=%d response=%+v", code, response)
			}
			if _, err := os.Stat(exportPath); !os.IsNotExist(err) {
				t.Fatalf("M11 recovery export wrote during pending journal recovery: %v", err)
			}
			admissionRuntime := t.TempDir()
			if code, response := missionCall(t, "init", admissionRuntime); code != 0 || response["status"] != "INITIALIZED" {
				t.Fatalf("init admission runtime: code=%d response=%+v", code, response)
			}
			if code, response := missionCall(t, "m11-recovery-admit", admissionRuntime, fixture.dir, filepath.Join(t.TempDir(), "missing-handoff.json"), filepath.Join(t.TempDir(), "missing-admission.json")); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
				t.Fatalf("M11 recovery admission read an interrupted old runtime: code=%d response=%+v", code, response)
			}
			if _, err := os.Stat(m11ArtifactRegistryPath(admissionRuntime)); !os.IsNotExist(err) {
				t.Fatalf("M11 recovery admission wrote during pending old journal recovery: %v", err)
			}
			state, err := loadMissionState(fixture.dir)
			if faultPhase == "before_write" && (err != nil || state.Stop) {
				t.Fatalf("pre-write fault marked mission stopped: state=%+v err=%v", state, err)
			}
			if faultPhase == "after_write" && (err == nil || state.Stop) {
				t.Fatalf("post-write fault did not fail closed at the stopped-ledger recovery boundary: state=%+v err=%v", state, err)
			}
			m11RegistryAppendFault = nil
			if err := recoverM11UnknownStopJournal(fixture.dir); err != nil {
				t.Fatalf("unknown STOP journal recovery failed: %v", err)
			}
			assertM11UnknownStopRecovered(t, fixture, attemptedAt, reason)
		})
	}
}

func TestM11UnknownStopJournalRecoversAfterDurableStopWriteFailure(t *testing.T) {
	for _, tc := range []struct {
		name           string
		failureAt      int
		stateCommitted bool
	}{
		{name: "mission_state", failureAt: 2},
		{name: "stop_marker", failureAt: 3, stateCommitted: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newM11UnknownStopFixture(t)
			const attemptedAt, reason = "2026-09-08T00:00:02Z", "fixture timeout"
			atomicWrites := 0
			missionStateWriteFault = func(phase string) error {
				if phase != "before_rename" {
					return nil
				}
				atomicWrites++ // journal, mission-state after stopped ledger, then STOP marker
				if atomicWrites == tc.failureAt {
					return errors.New("injected durable STOP write failure")
				}
				return nil
			}
			t.Cleanup(func() { missionStateWriteFault = nil })
			if _, _, _, err := recordUnknownM11Execution(fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, reason); err == nil {
				t.Fatal("durable STOP write fault was not surfaced")
			}
			if atomicWrites != tc.failureAt {
				t.Fatalf("fault did not reach expected durable STOP write: atomic writes=%d", atomicWrites)
			}
			if _, err := os.Stat(m11UnknownStopJournalPath(fixture.dir)); err != nil {
				t.Fatalf("unknown STOP journal was removed before recovery: %v", err)
			}
			state, err := loadMissionState(fixture.dir)
			if tc.stateCommitted && (err != nil || !state.Stop) {
				t.Fatalf("STOP marker fault did not retain stopped mission state: state=%+v err=%v", state, err)
			}
			if !tc.stateCommitted && err == nil {
				t.Fatalf("mission-state fault did not fail closed before STOP: state=%+v", state)
			}
			if code, response := missionCall(t, "status", fixture.dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
				t.Fatalf("status did not fail closed on durable STOP write fault: code=%d response=%+v", code, response)
			}
			missionStateWriteFault = nil
			if err := recoverM11UnknownStopJournal(fixture.dir); err != nil {
				t.Fatalf("unknown STOP journal did not recover durable STOP write failure: %v", err)
			}
			assertM11UnknownStopRecovered(t, fixture, attemptedAt, reason)
		})
	}
}

func TestM11FailedExecutionJournalRecoversAfterLedgerWriteFailure(t *testing.T) {
	for _, phase := range []string{"before_write", "after_write", "after_sync"} {
		faultPhase := phase
		t.Run(phase, func(t *testing.T) {
			fixture := newM11UnknownStopFixture(t)
			const attemptedAt, reason = "2026-09-08T00:00:02Z", "fixture execution failure"
			m11RegistryAppendFault = func(registryPhase string, entry corem11.ArtifactEntry) error {
				if entry.ArtifactKind == corem11.ArtifactKindLedger && registryPhase == faultPhase {
					return errors.New("injected failed-execution ledger write failure")
				}
				return nil
			}
			t.Cleanup(func() { m11RegistryAppendFault = nil })
			if _, _, _, err := recordFailedM11Execution(fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, reason); err == nil {
				t.Fatal("failed-execution ledger fault was not surfaced")
			}
			if _, err := os.Stat(m11FailedExecutionJournalPath(fixture.dir)); err != nil {
				t.Fatalf("failed execution journal was removed before recovery: %v", err)
			}
			if code, response := missionCall(t, "status", fixture.dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
				t.Fatalf("status did not fail closed on failed-execution journal: code=%d response=%+v", code, response)
			}
			if code, response := missionCall(t, "m11-resolve", fixture.dir, corem11.ArtifactKindLedger, fixture.ledgerEntry.ArtifactID); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
				t.Fatalf("M11 resolver exposed a FAILED partial transition: code=%d response=%+v", code, response)
			}
			m11RegistryAppendFault = nil
			if err := recoverM11FailedExecutionJournal(fixture.dir); err != nil {
				t.Fatalf("failed-execution journal recovery failed: %v", err)
			}
			if _, err := os.Stat(m11FailedExecutionJournalPath(fixture.dir)); !os.IsNotExist(err) {
				t.Fatalf("failed execution journal remains after recovery: %v", err)
			}
			_, head, err := m11LedgerHead(fixture.dir, fixture.lease.LeaseID)
			if err != nil || head.ConsecutiveFailures != 1 || head.LastExecutionAt != attemptedAt || head.UpdatedAt != attemptedAt {
				t.Fatalf("failed-execution ledger transition was not recovered: ledger=%+v err=%v", head, err)
			}
			record, ledger, status, err := recordFailedM11Execution(fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, reason)
			if err != nil || status != appendDuplicate || record.Status != "FAILED" || ledger.UpdatedAt != attemptedAt || ledger.ConsecutiveFailures != 1 {
				t.Fatalf("failed-execution retry was not exact after recovery: record=%+v ledger=%+v status=%s err=%v", record, ledger, status, err)
			}
		})
	}
}

func TestM11ReconcileRejectsSecondResolutionForSameUnknownExecution(t *testing.T) {
	fixture := newM11UnknownStopFixture(t)
	const attemptedAt = "2026-09-08T00:00:02Z"
	execution, _, status, err := recordUnknownM11Execution(fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, "fixture timeout")
	if err != nil || status != appendAdded {
		t.Fatalf("create stopped UNKNOWN fixture: execution=%+v status=%s err=%v", execution, status, err)
	}
	stoppedEntry, _, err := m11LedgerHead(fixture.dir, fixture.lease.LeaseID)
	if err != nil {
		t.Fatalf("resolve stopped ledger: %v", err)
	}
	first := corem11.ProductionReconciliationResolution{ResolutionID: "resolution-first", LeaseID: fixture.lease.LeaseID, LeaseVersion: fixture.lease.LeaseVersion, LeaseHash: fixture.lease.LeaseHash, ExecutionID: execution.ExecutionID, ResolvedBy: "human", ResolverID: "reviewer-1", ResolvedAt: "2026-09-08T00:00:03Z", EffectState: "NOT_PERFORMED", Reason: "human reviewed fixture timeout"}
	registerM11TestArtifact(t, fixture.dir, corem11.ArtifactKindReconciliation, first)
	if resolution, ledger, status, err := reconcileM11Execution(fixture.dir, first.ResolutionID, stoppedEntry.ArtifactID); err != nil || status != appendAdded || len(ledger.ReconciliationResolutionIDs) != 1 || ledger.ReconciliationResolutionIDs[0] != first.ResolutionID || resolution.ResolutionID != first.ResolutionID {
		t.Fatalf("first reconciliation did not advance the stopped ledger: resolution=%+v ledger=%+v status=%s err=%v", resolution, ledger, status, err)
	}
	if _, ledger, status, err := reconcileM11Execution(fixture.dir, first.ResolutionID, stoppedEntry.ArtifactID); err != nil || status != appendDuplicate || len(ledger.ReconciliationResolutionIDs) != 1 || ledger.ReconciliationResolutionIDs[0] != first.ResolutionID {
		t.Fatalf("exact reconciliation retry did not resolve from the new head: ledger=%+v status=%s err=%v", ledger, status, err)
	}
	second := corem11.ProductionReconciliationResolution{ResolutionID: "resolution-stale", LeaseID: fixture.lease.LeaseID, LeaseVersion: fixture.lease.LeaseVersion, LeaseHash: fixture.lease.LeaseHash, ExecutionID: execution.ExecutionID, ResolvedBy: "human", ResolverID: "reviewer-2", ResolvedAt: "2026-09-08T00:00:04Z", EffectState: "NOT_PERFORMED", Reason: "attempt stale fork"}
	secondRaw, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := registerM11Artifact(fixture.dir, corem11.ArtifactKindReconciliation, secondRaw); err == nil {
		t.Fatal("second reconciliation resolution for one UNKNOWN execution was registered")
	}
	_, head, err := m11LedgerHead(fixture.dir, fixture.lease.LeaseID)
	if err != nil || len(head.ReconciliationResolutionIDs) != 1 || head.ReconciliationResolutionIDs[0] != first.ResolutionID || head.UpdatedAt != first.ResolvedAt {
		t.Fatalf("rejected reconciliation changed the current ledger head: ledger=%+v err=%v", head, err)
	}
}

func TestBackupCreateRecoversPendingM11UnknownStopJournal(t *testing.T) {
	fixture := newM11UnknownStopFixture(t)
	const attemptedAt, reason = "2026-09-08T00:00:02Z", "fixture timeout"
	m11RegistryAppendFault = func(phase string, entry corem11.ArtifactEntry) error {
		if entry.ArtifactKind == corem11.ArtifactKindLedger && phase == "before_write" {
			return errors.New("injected stopped-ledger write failure")
		}
		return nil
	}
	if _, _, _, err := recordUnknownM11Execution(fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, reason); err == nil {
		t.Fatal("stopped-ledger fault was not surfaced")
	}
	m11RegistryAppendFault = nil
	t.Cleanup(func() { m11RegistryAppendFault = nil })
	if _, err := buildBR10AdvisorFixture(fixture.dir); err != nil {
		t.Fatalf("build backup history fixture: %v", err)
	}
	for _, name := range []string{"action-input.json", "outcome-input.json"} {
		if err := os.Remove(filepath.Join(fixture.dir, name)); err != nil {
			t.Fatalf("remove fixture-only backup input %s: %v", name, err)
		}
	}
	backup, restored := filepath.Join(t.TempDir(), "backup"), filepath.Join(t.TempDir(), "restored")
	if code, response := backupCall(t, "create", fixture.dir, backup); code != 0 || response["status"] != "BACKED_UP" {
		t.Fatalf("backup create did not recover pending UNKNOWN STOP journal: code=%d response=%+v", code, response)
	}
	assertM11UnknownStopRecovered(t, fixture, attemptedAt, reason)
	if code, response := backupCall(t, "restore", backup, restored); code != 0 || response["status"] != "RESTORED" {
		t.Fatalf("recovered UNKNOWN STOP backup did not restore: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "status", restored); code != 0 || response["artifact"].(map[string]any)["stop"] != true {
		t.Fatalf("restored UNKNOWN STOP runtime was not durably stopped: code=%d response=%+v", code, response)
	}
}

func TestBackupCreateRecoversPendingM11FailedExecutionJournal(t *testing.T) {
	fixture := newM11UnknownStopFixture(t)
	if _, err := buildBR10AdvisorFixture(fixture.dir); err != nil {
		t.Fatalf("build backup history fixture: %v", err)
	}
	const attemptedAt, reason = "2026-09-08T00:00:02Z", "fixture execution failure"
	m11RegistryAppendFault = func(phase string, entry corem11.ArtifactEntry) error {
		if entry.ArtifactKind == corem11.ArtifactKindLedger && phase == "before_write" {
			return errors.New("injected failed-execution ledger write failure")
		}
		return nil
	}
	if _, _, _, err := recordFailedM11Execution(fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, reason); err == nil {
		t.Fatal("failed-execution ledger fault was not surfaced")
	}
	m11RegistryAppendFault = nil
	t.Cleanup(func() { m11RegistryAppendFault = nil })
	backup, restored := filepath.Join(t.TempDir(), "backup"), filepath.Join(t.TempDir(), "restored")
	if code, response := backupCall(t, "create", fixture.dir, backup); code == 0 || response["status"] != "INPUT_ERROR" {
		t.Fatalf("backup accepted a recovered FAILED execution without an outcome: code=%d response=%+v", code, response)
	}
	if _, err := os.Stat(m11FailedExecutionJournalPath(fixture.dir)); !os.IsNotExist(err) {
		t.Fatalf("backup did not recover failed-execution journal: %v", err)
	}
	postEntry, _, err := m11LedgerHead(fixture.dir, fixture.lease.LeaseID)
	if err != nil {
		t.Fatalf("resolve recovered post-execution ledger: %v", err)
	}
	outcome := m03.OutcomeRecord{OutcomeID: "backup-recovered-failed-outcome", EffectRef: m03.EffectRef{EffectKind: "MACHINE_EXECUTION", EffectID: "prod-exec-" + fixture.authorization.AuthorizationID}, ObservedAt: "2026-09-08T00:00:03Z", Status: "CANCELLED", Metrics: map[string]float64{}, SourceRef: "fixture:m11-outcome/backup-recovery"}
	outcomeInput := filepath.Join(fixture.dir, "m11-outcome-input.json")
	if err := writeFixtureJSON(outcomeInput, outcome); err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "m11-outcome", fixture.dir, outcomeInput, postEntry.ArtifactID); code != 0 || response["status"] != "APPENDED" {
		t.Fatalf("record recovered FAILED outcome: code=%d response=%+v", code, response)
	}
	for _, name := range []string{"action-input.json", "outcome-input.json", "m11-outcome-input.json"} {
		if err := os.Remove(filepath.Join(fixture.dir, name)); err != nil {
			t.Fatalf("remove fixture-only backup input %s: %v", name, err)
		}
	}
	if code, response := backupCall(t, "create", fixture.dir, backup); code != 0 || response["status"] != "BACKED_UP" {
		t.Fatalf("backup did not persist recovered FAILED execution: code=%d response=%+v", code, response)
	}
	if code, response := backupCall(t, "restore", backup, restored); code != 0 || response["status"] != "RESTORED" {
		t.Fatalf("recovered FAILED execution backup did not restore: code=%d response=%+v", code, response)
	}
}

func assertM11UnknownStopRecovered(t *testing.T, fixture m11UnknownStopFixture, attemptedAt, reason string) {
	t.Helper()
	if _, err := os.Stat(m11UnknownStopJournalPath(fixture.dir)); !os.IsNotExist(err) {
		t.Fatalf("unknown STOP journal remains after recovery: %v", err)
	}
	state, err := loadMissionState(fixture.dir)
	if err != nil || !state.Stop || state.StopReason != "RECONCILIATION_REQUIRED" {
		t.Fatalf("recovered mission is not durably stopped: state=%+v err=%v", state, err)
	}
	_, head, err := m11LedgerHead(fixture.dir, fixture.lease.LeaseID)
	if err != nil || head.ControlMode != "STOPPED" || !head.ReconciliationRequired {
		t.Fatalf("recovered stopped ledger is invalid: ledger=%+v err=%v", head, err)
	}
	if _, _, status, err := recordUnknownM11Execution(fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, reason); err != nil || status != appendDuplicate {
		t.Fatalf("unknown retry was not exact duplicate: status=%s err=%v", status, err)
	}
}
