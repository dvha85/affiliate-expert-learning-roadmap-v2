package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"testing"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
	corem10 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m10"
	corem11 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m11"
)

func registerM11TestArtifact(t *testing.T, dir, kind string, value any) {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := registerM11Artifact(dir, kind, raw); err != nil {
		t.Fatalf("register %s: %v", kind, err)
	}
}

func setupM11OutcomeJournalFixture(t *testing.T, dir string) (m11OutcomeJournal, corem11.ArtifactEntry) {
	return setupM11OutcomeJournalFixtureForIntent(t, dir, "journal-intent", "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "journal-correlation")
}

// setupM11OutcomeJournalFixtureForIntent keeps the recovery fixture's M11
// artifact graph intact while allowing command-path tests to bind it to a real
// learner intent and canonical history record.
func setupM11OutcomeJournalFixtureForIntent(t *testing.T, dir, intentID, intentHash, correlationID string) (m11OutcomeJournal, corem11.ArtifactEntry) {
	t.Helper()
	lease := corem11.ProductionLease{LeaseID: "journal-lease", LeaseVersion: "v1", PolicyVersion: "policy-1", ApprovalRef: "journal-approval", ReviewedBy: "human", ReviewerID: "reviewer", ReviewedAt: "2026-09-08T00:00:00Z", PromotionReviewRef: "review", SourceCanaryGrantID: "canary", SourceCanaryGrantVersion: "v1", SourceCanaryGrantHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ValidFrom: "2026-09-08T00:00:00Z", ExpiresAt: "2099-09-08T00:00:00Z", AllowedRiskClasses: []string{"RISK0"}, AllowedActionTypes: []string{"DRAFT"}, AllowedHosts: []string{"example.com"}, ExecutorIDs: []string{"fixture_stub"}, MaxExecutionsTotal: 1, MaxExecutionsPerWindow: 1, WindowSeconds: 60, MaxCostMinorTotal: 10, Currency: "USD", MaxPendingOutcomes: 1, MaxConsecutiveFailures: 1, MaxOutcomeAgeSeconds: 60, MaxHealthSnapshotAgeSeconds: 60, KillSwitchRequired: true, CorrelationID: correlationID, HashVersion: "go-json-v1"}
	lease.LeaseHash = corem11.ComputeProductionLeaseHash(lease)
	approval := corem11.ProductionLeaseApproval{ApprovalID: lease.ApprovalRef, LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, PromotionReviewRef: lease.PromotionReviewRef, SourceCanaryGrantID: lease.SourceCanaryGrantID, SourceCanaryGrantVersion: lease.SourceCanaryGrantVersion, SourceCanaryGrantHash: lease.SourceCanaryGrantHash, SourceE5Refs: []string{"fixture:e5"}, ValidatedRiskClasses: []string{"RISK0"}, ReviewedBy: "human", ReviewerID: lease.ReviewerID, ReviewedAt: lease.ReviewedAt, Decision: "APPROVE_PRODUCTION_LEASE"}
	health := corem11.ProductionHealthSnapshot{SnapshotID: "journal-health", LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ObservedAt: "2026-09-08T00:00:00Z", SourceRefs: []string{"fixture:health"}, DependencyState: "HEALTHY", TelemetryComplete: true, HashVersion: "go-json-v1"}
	health.SnapshotHash = corem11.ComputeProductionHealthHash(health)
	cost := corem10.TrustedCostBound{CostBoundID: "journal-cost", IntentID: intentID, IntentHash: intentHash, MaxCostMinor: 10, Currency: "USD", SourceRef: "fixture:cost", ObservedAt: "2026-09-08T00:00:00Z", ExpiresAt: "2099-09-08T00:00:00Z", CorrelationID: lease.CorrelationID, HashVersion: "go-json-v1"}
	cost.CostBoundHash = corem10.ComputeTrustedCostBoundHash(cost)
	ledger := corem11.ProductionLedger{LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ControlMode: "NORMAL", WindowStartedAt: "2026-09-08T00:00:00Z", ExecutionsTotal: 1, ExecutionsInWindow: 1, CostMinorTotal: 10, PendingOutcomes: 1, PendingExecutionIDs: []string{"journal-exec"}, SuccessfulIdempotencyKeys: []string{}, OutcomeLinks: []corem11.ProductionOutcomeLink{}, ReconciliationResolutionIDs: []string{}, UpdatedAt: "2026-09-08T00:00:01Z"}
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
	gate := corem11.ProductionGateDecision{LedgerArtifactID: gateLedgerEntry.ArtifactID, LedgerContentHash: gateLedgerEntry.ContentHash, LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, IntentID: cost.IntentID, IntentHash: cost.IntentHash, PolicyVersion: lease.PolicyVersion, RiskClass: "RISK0", HealthSnapshotID: health.SnapshotID, HealthSnapshotHash: health.SnapshotHash, CostBoundID: cost.CostBoundID, CostBoundHash: cost.CostBoundHash, CostBoundMinor: cost.MaxCostMinor, Decision: "ALLOW_PRODUCTION", Reason: "fixture", EvaluatedAt: "2026-09-08T00:00:00Z"}
	gate.GateID = corem11.ComputeProductionGateID(lease, gate.IntentID, gate.IntentHash, health, cost, gateLedgerEntry, gate.EvaluatedAt)
	authorization := corem11.ProductionExecutionAuthorization{AuthorizationID: corem11.ComputeProductionAuthorizationID(gate.GateID, "fixture_stub", "2026-09-08T00:00:00Z"), IntentID: cost.IntentID, IntentHash: cost.IntentHash, PolicyVersion: lease.PolicyVersion, ProductionLeaseID: lease.LeaseID, ProductionLeaseVersion: lease.LeaseVersion, ProductionLeaseHash: lease.LeaseHash, ProductionGateID: gate.GateID, ProductionHealthSnapshotID: health.SnapshotID, ProductionHealthSnapshotHash: health.SnapshotHash, ProductionCostBoundID: cost.CostBoundID, ProductionCostBoundHash: cost.CostBoundHash, ProductionCostBoundMinor: cost.MaxCostMinor, ExecutorID: "fixture_stub", AuthorizedAt: "2026-09-08T00:00:00Z", ExpiresAt: "2026-09-08T00:01:00Z", IdempotencyKey: "journal-key", CorrelationID: cost.CorrelationID, ExecutionMode: "GOVERNED_PRODUCTION", ExecutionAuthorized: true}
	execution := corem11.ProductionExecutionRecord{ExecutionID: corem11.ComputeProductionExecutionID(authorization.AuthorizationID), AuthorizationID: authorization.AuthorizationID, ProductionLeaseID: lease.LeaseID, ProductionLeaseVersion: lease.LeaseVersion, ProductionLeaseHash: lease.LeaseHash, ProductionGateID: gate.GateID, ProductionHealthSnapshotID: health.SnapshotID, ProductionHealthSnapshotHash: health.SnapshotHash, ProductionCostBoundID: cost.CostBoundID, ProductionCostBoundHash: cost.CostBoundHash, ProductionCostBoundMinor: cost.MaxCostMinor, IntentID: cost.IntentID, IntentHash: cost.IntentHash, ExecutorID: authorization.ExecutorID, IdempotencyKey: authorization.IdempotencyKey, AttemptedAt: "2026-09-08T00:00:01Z", Status: "FAILED", SideEffectState: "NOT_PERFORMED", CorrelationID: cost.CorrelationID}
	ledger.PendingExecutionIDs = []string{execution.ExecutionID}
	ledgerRaw, err = json.Marshal(ledger)
	if err != nil {
		t.Fatal(err)
	}
	ledgerEntry, err = corem11.NewArtifactEntry(corem11.ArtifactKindLedger, ledgerRaw)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		kind  string
		value any
	}{{corem11.ArtifactKindLease, lease}, {corem11.ArtifactKindLeaseApproval, approval}, {corem11.ArtifactKindHealth, health}, {corem11.ArtifactKindCostBound, cost}, {corem11.ArtifactKindLedger, gateLedger}, {corem11.ArtifactKindActivation, activation}, {corem11.ArtifactKindGate, gate}, {corem11.ArtifactKindAuthorization, authorization}, {corem11.ArtifactKindLedger, ledger}, {corem11.ArtifactKindExecution, execution}} {
		registerM11TestArtifact(t, dir, item.kind, item.value)
	}
	outcome := m03.OutcomeRecord{OutcomeID: "journal-outcome", EffectRef: m03.EffectRef{EffectKind: "MACHINE_EXECUTION", EffectID: execution.ExecutionID}, ObservedAt: "2026-09-08T00:00:02Z", Status: "CANCELLED", Metrics: map[string]float64{}, SourceRef: "fixture:m11-outcome/journal"}
	next, err := nextM11FixtureOutcomeLedger(ledger, outcome, execution)
	if err != nil {
		t.Fatalf("build expected journal ledger: %v", err)
	}
	predecessorEntry, err := resolveM11Artifact(dir, corem11.ArtifactKindLedger, ledgerEntry.ArtifactID, ledgerEntry.ContentHash)
	if err != nil {
		t.Fatalf("resolve journal predecessor: %v", err)
	}
	predecessorValue, status := corem11.DecodeArtifact("ledger", predecessorEntry.Artifact)
	if status != corem11.Valid {
		t.Fatalf("decode journal predecessor: %s", status)
	}
	expected, err := nextM11FixtureOutcomeLedger(*predecessorValue.(*corem11.ProductionLedger), outcome, execution)
	if err != nil || !reflect.DeepEqual(expected, next) {
		t.Fatalf("journal fixture transition does not match persisted predecessor: expected=%+v next=%+v err=%v", expected, next, err)
	}
	journal := m11OutcomeJournal{
		Version:                "m11-outcome-journal/v1",
		PredecessorArtifactID:  ledgerEntry.ArtifactID,
		PredecessorContentHash: ledgerEntry.ContentHash,
		Outcome:                outcome,
		Ledger:                 next,
	}
	return journal, ledgerEntry
}

func TestM11OutcomeJournalRecoversLedgerThenOutcomeAppendFailure(t *testing.T) {
	dir := t.TempDir()
	journal, ledgerEntry := setupM11OutcomeJournalFixture(t, dir)
	tamperedJournal := journal
	tamperedJournal.Ledger.PendingOutcomes = 1
	if err := writeJSONAtomic(m11OutcomeJournalPath(dir), tamperedJournal); err != nil {
		t.Fatal(err)
	}
	if err := recoverM11OutcomeJournal(dir); err == nil {
		t.Fatalf("tampered journal was not rejected before recovery: %v", err)
	}
	if _, err := os.Stat(m11OutcomeJournalPath(dir)); err != nil {
		t.Fatalf("tampered journal was unexpectedly removed: %v", err)
	}
	headEntry, _, err := m11LedgerHead(dir, journal.Ledger.LeaseID)
	if err != nil || headEntry.ArtifactID != ledgerEntry.ArtifactID || headEntry.ContentHash != ledgerEntry.ContentHash {
		t.Fatalf("tampered journal changed the active ledger: entry=%+v err=%v", headEntry, err)
	}
	if err := writeJSONAtomic(m11OutcomeJournalPath(dir), journal); err != nil {
		t.Fatal(err)
	}
	m11OutcomeAppendFault = func(phase string) error {
		if phase == "before_append" {
			return errors.New("injected outcome append failure")
		}
		return nil
	}
	t.Cleanup(func() { m11OutcomeAppendFault = nil })
	if err := recoverM11OutcomeJournal(dir); err == nil {
		t.Fatal("outcome append failure was not surfaced")
	} else if !strings.Contains(err.Error(), "injected outcome append failure") {
		t.Fatalf("journal did not reach the outcome append fault: %v", err)
	}
	if _, err := os.Stat(m11OutcomeJournalPath(dir)); err != nil {
		t.Fatalf("journal was removed before recovery completed: %v", err)
	}
	if code, response := missionCall(t, "m11-resolve", dir, corem11.ArtifactKindLedger, ledgerEntry.ArtifactID); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
		t.Fatalf("M11 resolver exposed an outcome partial transition: code=%d response=%+v", code, response)
	}
	m11OutcomeAppendFault = nil
	if err := recoverM11OutcomeJournal(dir); err != nil {
		t.Fatalf("journal recovery failed: %v", err)
	}
	if _, err := os.Stat(m11OutcomeJournalPath(dir)); !os.IsNotExist(err) {
		t.Fatalf("journal still exists after recovery: %v", err)
	}
	outcomes, err := loadM11FixtureOutcomes(dir)
	if err != nil || len(outcomes) != 1 || outcomes[0].OutcomeID != journal.Outcome.OutcomeID {
		t.Fatalf("outcome was not recovered exactly: outcomes=%+v err=%v", outcomes, err)
	}
	_, head, err := m11LedgerHead(dir, journal.Ledger.LeaseID)
	if err != nil || head.PendingOutcomes != 0 || len(head.OutcomeLinks) != 1 {
		t.Fatalf("ledger recovery mismatch: ledger=%+v err=%v", head, err)
	}
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("initialize stopped-ledger recovery fixture: code=%d response=%+v", code, response)
	}
	stopped := head
	stopped.ControlMode, stopped.StopReason, stopped.ReconciliationRequired = "STOPPED", "RECOVERY_REQUIRED", true
	stopped.UpdatedAt = "2026-09-08T00:00:03Z"
	registerM11TestArtifact(t, dir, corem11.ArtifactKindLedger, stopped)
	if err := recoverM11DurableStop(dir); err != nil {
		t.Fatalf("stopped ledger did not repair mission STOP: %v", err)
	}
	state, err := loadMissionState(dir)
	if err != nil || !state.Stop || state.StopReason != "RECOVERY_REQUIRED" {
		t.Fatalf("mission STOP was not recovered: state=%+v err=%v", state, err)
	}
}

func TestM11FixtureOutcomeLoaderRejectsDuplicateExecutionBeforeRuntimeUse(t *testing.T) {
	dir := t.TempDir()
	journal, _ := setupM11OutcomeJournalFixture(t, dir)
	duplicate := journal.Outcome
	duplicate.OutcomeID = "journal-outcome-duplicate"
	first, err := json.Marshal(journal.Outcome)
	if err != nil {
		t.Fatal(err)
	}
	second, err := json.Marshal(duplicate)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(m11OutcomeStorePath(dir), append(append(first, '\n'), append(second, '\n')...), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadM11FixtureOutcomes(dir); err == nil || !strings.Contains(err.Error(), "invalid M11 fixture outcome store") {
		t.Fatalf("two outcomes for one M11 execution reached runtime loader: %v", err)
	}
}

func TestM11OutcomeJournalRecoversAfterOutcomeAppendAckFailure(t *testing.T) {
	dir := t.TempDir()
	journal, _ := setupM11OutcomeJournalFixture(t, dir)
	if err := writeJSONAtomic(m11OutcomeJournalPath(dir), journal); err != nil {
		t.Fatal(err)
	}
	m11OutcomeAppendFault = func(phase string) error {
		if phase == "after_append" {
			return errors.New("injected outcome append acknowledgement failure")
		}
		return nil
	}
	t.Cleanup(func() { m11OutcomeAppendFault = nil })
	if err := recoverM11OutcomeJournal(dir); err == nil {
		t.Fatal("outcome append acknowledgement failure was not surfaced")
	} else if !strings.Contains(err.Error(), "injected outcome append acknowledgement failure") {
		t.Fatalf("journal did not reach the post-append fault: %v", err)
	}
	if _, err := os.Stat(m11OutcomeJournalPath(dir)); err != nil {
		t.Fatalf("journal was removed after the outcome append ACK failure: %v", err)
	}
	outcomes, err := loadM11FixtureOutcomes(dir)
	if err != nil || len(outcomes) != 1 || !reflect.DeepEqual(outcomes[0], journal.Outcome) {
		t.Fatalf("outcome append was not persisted exactly once before replay: outcomes=%+v err=%v", outcomes, err)
	}
	_, head, err := m11LedgerHead(dir, journal.Ledger.LeaseID)
	if err != nil || !reflect.DeepEqual(head, journal.Ledger) {
		t.Fatalf("ledger transition was not persisted before replay: ledger=%+v err=%v", head, err)
	}
	m11OutcomeAppendFault = nil
	if err := recoverM11OutcomeJournal(dir); err != nil {
		t.Fatalf("post-append journal replay failed: %v", err)
	}
	if _, err := os.Stat(m11OutcomeJournalPath(dir)); !os.IsNotExist(err) {
		t.Fatalf("journal still exists after post-append replay: %v", err)
	}
	outcomes, err = loadM11FixtureOutcomes(dir)
	if err != nil || len(outcomes) != 1 || !reflect.DeepEqual(outcomes[0], journal.Outcome) {
		t.Fatalf("replay duplicated or changed the outcome: outcomes=%+v err=%v", outcomes, err)
	}
	_, head, err = m11LedgerHead(dir, journal.Ledger.LeaseID)
	if err != nil || !reflect.DeepEqual(head, journal.Ledger) {
		t.Fatalf("replay duplicated or changed the ledger transition: ledger=%+v err=%v", head, err)
	}
}

func TestMissionM11OutcomeDisclosesVisibleAppendAcknowledgementUncertainty(t *testing.T) {
	dir := t.TempDir()
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("initialize outcome fixture: code=%d response=%+v", code, response)
	}
	journal, predecessor := setupM11OutcomeJournalFixture(t, dir)
	input := filepath.Join(dir, "visible-m11-outcome-append.json")
	writeMissionTestJSON(t, input, journal.Outcome)
	m11OutcomeAppendFault = func(phase string) error {
		if phase == "after_append" {
			return errors.New("injected outcome append acknowledgement failure")
		}
		return nil
	}
	t.Cleanup(func() { m11OutcomeAppendFault = nil })
	if code, response := missionCall(t, "m11-outcome", dir, input, predecessor.ArtifactID); code == 0 || response["status"] != "PUBLISHED_RECOVERY_REQUIRED" {
		t.Fatalf("visible M11 outcome append did not disclose recovery: code=%d response=%+v", code, response)
	} else if artifact, ok := response["artifact"].(map[string]any); !ok || artifact["outcome"] == nil || artifact["post_ledger"] == nil {
		t.Fatalf("visible M11 outcome append omitted deterministic transition: %+v", response)
	}
	if _, err := os.Stat(m11OutcomeJournalPath(dir)); err != nil {
		t.Fatalf("visible outcome append did not retain recovery journal: %v", err)
	}
	outcomes, err := loadM11FixtureOutcomes(dir)
	if err != nil || len(outcomes) != 1 || !reflect.DeepEqual(outcomes[0], journal.Outcome) {
		t.Fatalf("visible outcome append did not persist exactly one record: outcomes=%+v err=%v", outcomes, err)
	}
	m11OutcomeAppendFault = nil
	binary := buildMissionBinary(t)
	if code, response := missionBinaryCall(t, binary, "mission", "status", dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
		t.Fatalf("fresh status exposed visible append uncertainty: code=%d response=%+v", code, response)
	}
	if code, response := missionBinaryCall(t, binary, "mission", "m11-outcome", dir, input, predecessor.ArtifactID); code != 0 || response["status"] != "EXACT_DUPLICATE" {
		t.Fatalf("locked exact retry did not close visible append journal: code=%d response=%+v", code, response)
	}
	if _, err := os.Stat(m11OutcomeJournalPath(dir)); !os.IsNotExist(err) {
		t.Fatalf("visible append journal remains after exact retry: %v", err)
	}
	outcomes, err = loadM11FixtureOutcomes(dir)
	if err != nil || len(outcomes) != 1 || !reflect.DeepEqual(outcomes[0], journal.Outcome) {
		t.Fatalf("exact retry duplicated or changed outcome: outcomes=%+v err=%v", outcomes, err)
	}
}

// TestM11OutcomeProcessKillAfterJournalBeforeLedgerAppend exercises the real
// two-file command boundary in a separate process. The journal must remain the
// only recovery authority after SIGKILL; a fresh writer replays the exact
// ledger/outcome transition once and an exact retry must not duplicate either
// side. This is process-termination evidence, not a kernel power-loss claim.
func TestM11OutcomeProcessKillAfterJournalBeforeLedgerAppend(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("SIGKILL process-boundary proof is not portable on Windows")
	}
	dir := t.TempDir()
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("initialize outcome process-kill fixture: code=%d response=%+v", code, response)
	}
	journal, predecessor := setupM11OutcomeJournalFixture(t, dir)
	input := filepath.Join(dir, "process-kill-m11-outcome.json")
	writeMissionTestJSON(t, input, journal.Outcome)
	childArgs := []string{"-test.run=^TestM11ProcessTerminationChild$", "--", "m11-outcome", dir, input, predecessor.ArtifactID}
	command := exec.Command(os.Args[0], childArgs...)
	command.Env = append(os.Environ(), "GO_WANT_M11_PROCESS_TERMINATION=1", "GO_M11_PROCESS_TERMINATION_MODE=kill")
	var childStdout, childStderr strings.Builder
	command.Stdout, command.Stderr = &childStdout, &childStderr
	if err := command.Run(); err == nil {
		t.Fatal("M11 outcome child unexpectedly completed after injected SIGKILL")
	} else {
		exited, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}
		status, ok := exited.ProcessState.Sys().(syscall.WaitStatus)
		if !ok || !status.Signaled() || status.Signal() != syscall.SIGKILL {
			t.Fatalf("M11 outcome child did not receive SIGKILL: %v stdout=%q stderr=%q", err, childStdout.String(), childStderr.String())
		}
	}
	if _, err := os.Stat(m11OutcomeJournalPath(dir)); err != nil {
		t.Fatalf("M11 outcome journal was not left for recovery: %v", err)
	}
	if code, response := missionCall(t, "status", dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
		t.Fatalf("fresh read-only process exposed interrupted M11 outcome: code=%d response=%+v", code, response)
	}
	binary := buildMissionBinary(t)
	if code, response := missionBinaryCall(t, binary, "mission", "m11-outcome", dir, input, predecessor.ArtifactID); code != 0 || response["status"] != "EXACT_DUPLICATE" {
		t.Fatalf("locked exact retry did not replay M11 outcome exactly once: code=%d response=%+v", code, response)
	}
	if _, err := os.Stat(m11OutcomeJournalPath(dir)); !os.IsNotExist(err) {
		t.Fatalf("M11 outcome journal remains after exact recovery: %v", err)
	}
	outcomes, err := loadM11FixtureOutcomes(dir)
	if err != nil || len(outcomes) != 1 || !reflect.DeepEqual(outcomes[0], journal.Outcome) {
		t.Fatalf("process-kill recovery duplicated or changed outcome: outcomes=%+v err=%v", outcomes, err)
	}
	_, head, err := m11LedgerHead(dir, journal.Ledger.LeaseID)
	if err != nil || !reflect.DeepEqual(head, journal.Ledger) {
		t.Fatalf("process-kill recovery changed post-outcome ledger: ledger=%+v err=%v", head, err)
	}
}

func TestMissionM11OutcomeDisclosesPublishedJournalCleanupUncertainty(t *testing.T) {
	dir := t.TempDir()
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("initialize outcome fixture: code=%d response=%+v", code, response)
	}
	journal, predecessor := setupM11OutcomeJournalFixture(t, dir)
	input := filepath.Join(dir, "published-m11-outcome-journal.json")
	writeMissionTestJSON(t, input, journal.Outcome)
	m11OutcomeAppendFault = func(phase string) error {
		if phase == "before_remove" {
			return errors.New("injected outcome journal cleanup failure")
		}
		return nil
	}
	t.Cleanup(func() { m11OutcomeAppendFault = nil })
	if code, response := missionCall(t, "m11-outcome", dir, input, predecessor.ArtifactID); code == 0 || response["status"] != "PUBLISHED_RECOVERY_REQUIRED" {
		t.Fatalf("published outcome journal cleanup was not disclosed: code=%d response=%+v", code, response)
	} else if artifact, ok := response["artifact"].(map[string]any); !ok || artifact["outcome"] == nil || artifact["post_ledger"] == nil {
		t.Fatalf("published journal cleanup omitted deterministic transition: %+v", response)
	}
	if _, err := os.Stat(m11OutcomeJournalPath(dir)); err != nil {
		t.Fatalf("published journal cleanup did not retain recovery journal: %v", err)
	}
	outcomes, err := loadM11FixtureOutcomes(dir)
	if err != nil || len(outcomes) != 1 || !reflect.DeepEqual(outcomes[0], journal.Outcome) {
		t.Fatalf("published journal cleanup did not persist exactly one record: outcomes=%+v err=%v", outcomes, err)
	}
	_, head, err := m11LedgerHead(dir, journal.Ledger.LeaseID)
	if err != nil || !reflect.DeepEqual(head, journal.Ledger) {
		t.Fatalf("published journal cleanup did not preserve post-outcome ledger: ledger=%+v err=%v", head, err)
	}
	m11OutcomeAppendFault = nil
	binary := buildMissionBinary(t)
	if code, response := missionBinaryCall(t, binary, "mission", "status", dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
		t.Fatalf("fresh status exposed published journal cleanup uncertainty: code=%d response=%+v", code, response)
	}
	if code, response := missionBinaryCall(t, binary, "mission", "m11-outcome", dir, input, predecessor.ArtifactID); code != 0 || response["status"] != "EXACT_DUPLICATE" {
		t.Fatalf("locked exact retry did not close published cleanup journal: code=%d response=%+v", code, response)
	}
	if _, err := os.Stat(m11OutcomeJournalPath(dir)); !os.IsNotExist(err) {
		t.Fatalf("published cleanup journal remains after exact retry: %v", err)
	}
	outcomes, err = loadM11FixtureOutcomes(dir)
	if err != nil || len(outcomes) != 1 || !reflect.DeepEqual(outcomes[0], journal.Outcome) {
		t.Fatalf("exact retry duplicated or changed published cleanup outcome: outcomes=%+v err=%v", outcomes, err)
	}
}

func TestMissionM11OutcomeDisclosesPostRemoveCleanupUncertainty(t *testing.T) {
	dir := t.TempDir()
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("initialize outcome fixture: code=%d response=%+v", code, response)
	}
	journal, predecessor := setupM11OutcomeJournalFixture(t, dir)
	input := filepath.Join(dir, "post-remove-m11-outcome-journal.json")
	writeMissionTestJSON(t, input, journal.Outcome)
	m11OutcomeAppendFault = func(phase string) error {
		if phase == "after_remove_before_parent_sync" {
			return errors.New("injected post-remove outcome journal cleanup failure")
		}
		return nil
	}
	t.Cleanup(func() { m11OutcomeAppendFault = nil })
	if code, response := missionCall(t, "m11-outcome", dir, input, predecessor.ArtifactID); code == 0 || response["status"] != "PUBLISHED_RECOVERY_REQUIRED" {
		t.Fatalf("post-remove outcome cleanup was not disclosed: code=%d response=%+v", code, response)
	} else if artifact, ok := response["artifact"].(map[string]any); !ok || artifact["outcome"] == nil || artifact["post_ledger"] == nil {
		t.Fatalf("post-remove cleanup omitted deterministic transition: %+v", response)
	}
	if _, err := os.Stat(m11OutcomeJournalPath(dir)); !os.IsNotExist(err) {
		t.Fatalf("post-remove outcome cleanup did not remove journal name: %v", err)
	}
	m11OutcomeAppendFault = nil
	if code, response := missionCall(t, "m11-outcome", dir, input, predecessor.ArtifactID); code != 0 || response["status"] != "EXACT_DUPLICATE" {
		t.Fatalf("post-remove exact retry did not preserve outcome: code=%d response=%+v", code, response)
	}
	outcomes, err := loadM11FixtureOutcomes(dir)
	if err != nil || len(outcomes) != 1 || !reflect.DeepEqual(outcomes[0], journal.Outcome) {
		t.Fatalf("post-remove exact retry duplicated or changed outcome: outcomes=%+v err=%v", outcomes, err)
	}
}

func TestBackupCreateRecoversPendingM11OutcomeJournal(t *testing.T) {
	dir := t.TempDir()
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("init backup recovery fixture: code=%d response=%+v", code, response)
	}
	journal, _ := setupM11OutcomeJournalFixture(t, dir)
	if _, err := buildBR10AdvisorFixture(dir); err != nil {
		t.Fatalf("build backup history fixture: %v", err)
	}
	for _, name := range []string{"action-input.json", "outcome-input.json"} {
		if err := os.Remove(filepath.Join(dir, name)); err != nil {
			t.Fatalf("remove fixture-only backup input %s: %v", name, err)
		}
	}
	if err := writeJSONAtomic(m11OutcomeJournalPath(dir), journal); err != nil {
		t.Fatal(err)
	}
	backup, restored := filepath.Join(t.TempDir(), "backup"), filepath.Join(t.TempDir(), "restored")
	if code, response := backupCall(t, "create", dir, backup); code != 0 || response["status"] != "BACKED_UP" {
		t.Fatalf("backup create did not recover pending M11 outcome journal: code=%d response=%+v", code, response)
	}
	if _, err := os.Stat(m11OutcomeJournalPath(dir)); !os.IsNotExist(err) {
		t.Fatalf("outcome journal remains after backup recovery: %v", err)
	}
	if code, response := backupCall(t, "restore", backup, restored); code != 0 || response["status"] != "RESTORED" {
		t.Fatalf("outcome-journal backup did not restore: code=%d response=%+v", code, response)
	}
	outcomes, err := loadM11FixtureOutcomes(restored)
	if err != nil || len(outcomes) != 1 || !reflect.DeepEqual(outcomes[0], journal.Outcome) {
		t.Fatalf("restored outcome differs after journal recovery: outcomes=%+v err=%v", outcomes, err)
	}
	_, head, err := m11LedgerHead(restored, journal.Ledger.LeaseID)
	if err != nil || !reflect.DeepEqual(head, journal.Ledger) {
		t.Fatalf("restored ledger differs after journal recovery: ledger=%+v err=%v", head, err)
	}
}
