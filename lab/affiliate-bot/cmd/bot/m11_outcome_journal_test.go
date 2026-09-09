package main

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
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

func TestM11OutcomeJournalRecoversLedgerThenOutcomeAppendFailure(t *testing.T) {
	dir := t.TempDir()
	lease := corem11.ProductionLease{LeaseID: "journal-lease", LeaseVersion: "v1", PolicyVersion: "policy-1", ApprovalRef: "journal-approval", ReviewedBy: "human", ReviewerID: "reviewer", ReviewedAt: "2026-09-08T00:00:00Z", PromotionReviewRef: "review", SourceCanaryGrantID: "canary", SourceCanaryGrantVersion: "v1", SourceCanaryGrantHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ValidFrom: "2026-09-08T00:00:00Z", ExpiresAt: "2099-09-08T00:00:00Z", AllowedRiskClasses: []string{"RISK0"}, AllowedActionTypes: []string{"DRAFT"}, AllowedHosts: []string{"example.com"}, ExecutorIDs: []string{"fixture_stub"}, MaxExecutionsTotal: 1, MaxExecutionsPerWindow: 1, WindowSeconds: 60, MaxCostMinorTotal: 10, Currency: "USD", MaxPendingOutcomes: 1, MaxConsecutiveFailures: 1, MaxOutcomeAgeSeconds: 60, MaxHealthSnapshotAgeSeconds: 60, KillSwitchRequired: true, CorrelationID: "journal-correlation", HashVersion: "go-json-v1"}
	lease.LeaseHash = corem11.ComputeProductionLeaseHash(lease)
	approval := corem11.ProductionLeaseApproval{ApprovalID: lease.ApprovalRef, LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, PromotionReviewRef: lease.PromotionReviewRef, SourceCanaryGrantID: lease.SourceCanaryGrantID, SourceCanaryGrantVersion: lease.SourceCanaryGrantVersion, SourceCanaryGrantHash: lease.SourceCanaryGrantHash, SourceE5Refs: []string{"fixture:e5"}, ValidatedRiskClasses: []string{"RISK0"}, ReviewedBy: "human", ReviewerID: lease.ReviewerID, ReviewedAt: lease.ReviewedAt, Decision: "APPROVE_PRODUCTION_LEASE"}
	health := corem11.ProductionHealthSnapshot{SnapshotID: "journal-health", LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ObservedAt: "2026-09-08T00:00:00Z", SourceRefs: []string{"fixture:health"}, DependencyState: "HEALTHY", TelemetryComplete: true, HashVersion: "go-json-v1"}
	health.SnapshotHash = corem11.ComputeProductionHealthHash(health)
	cost := corem10.TrustedCostBound{CostBoundID: "journal-cost", IntentID: "journal-intent", IntentHash: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", MaxCostMinor: 10, Currency: "USD", SourceRef: "fixture:cost", ObservedAt: "2026-09-08T00:00:00Z", ExpiresAt: "2099-09-08T00:00:00Z", CorrelationID: lease.CorrelationID, HashVersion: "go-json-v1"}
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
	activation := corem11.ProductionActivationRecord{LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ActivatedAt: "2026-09-08T00:00:00Z"}
	gate := corem11.ProductionGateDecision{GateID: "journal-gate", LedgerArtifactID: ledgerEntry.ArtifactID, LedgerContentHash: ledgerEntry.ContentHash, LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, IntentID: cost.IntentID, IntentHash: cost.IntentHash, PolicyVersion: lease.PolicyVersion, RiskClass: "RISK0", HealthSnapshotID: health.SnapshotID, HealthSnapshotHash: health.SnapshotHash, CostBoundID: cost.CostBoundID, CostBoundHash: cost.CostBoundHash, CostBoundMinor: cost.MaxCostMinor, Decision: "ALLOW_PRODUCTION", Reason: "fixture", EvaluatedAt: "2026-09-08T00:00:00Z"}
	authorization := corem11.ProductionExecutionAuthorization{AuthorizationID: "journal-auth", IntentID: cost.IntentID, IntentHash: cost.IntentHash, PolicyVersion: lease.PolicyVersion, ProductionLeaseID: lease.LeaseID, ProductionLeaseVersion: lease.LeaseVersion, ProductionLeaseHash: lease.LeaseHash, ProductionGateID: gate.GateID, ProductionHealthSnapshotID: health.SnapshotID, ProductionHealthSnapshotHash: health.SnapshotHash, ProductionCostBoundID: cost.CostBoundID, ProductionCostBoundHash: cost.CostBoundHash, ProductionCostBoundMinor: cost.MaxCostMinor, ExecutorID: "fixture_stub", AuthorizedAt: "2026-09-08T00:00:00Z", ExpiresAt: "2026-09-08T00:01:00Z", IdempotencyKey: "journal-key", CorrelationID: cost.CorrelationID, ExecutionMode: "GOVERNED_PRODUCTION", ExecutionAuthorized: true}
	execution := corem11.ProductionExecutionRecord{ExecutionID: "journal-exec", AuthorizationID: authorization.AuthorizationID, ProductionLeaseID: lease.LeaseID, ProductionLeaseVersion: lease.LeaseVersion, ProductionLeaseHash: lease.LeaseHash, ProductionGateID: gate.GateID, ProductionHealthSnapshotID: health.SnapshotID, ProductionHealthSnapshotHash: health.SnapshotHash, ProductionCostBoundID: cost.CostBoundID, ProductionCostBoundHash: cost.CostBoundHash, ProductionCostBoundMinor: cost.MaxCostMinor, IntentID: cost.IntentID, IntentHash: cost.IntentHash, ExecutorID: authorization.ExecutorID, IdempotencyKey: authorization.IdempotencyKey, AttemptedAt: "2026-09-08T00:00:01Z", Status: "FAILED", SideEffectState: "NOT_PERFORMED", CorrelationID: cost.CorrelationID}
	for _, item := range []struct {
		kind  string
		value any
	}{{corem11.ArtifactKindLease, lease}, {corem11.ArtifactKindLeaseApproval, approval}, {corem11.ArtifactKindHealth, health}, {corem11.ArtifactKindCostBound, cost}, {corem11.ArtifactKindLedger, ledger}, {corem11.ArtifactKindActivation, activation}, {corem11.ArtifactKindGate, gate}, {corem11.ArtifactKindAuthorization, authorization}, {corem11.ArtifactKindExecution, execution}} {
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
	headEntry, _, err := m11LedgerHead(dir, lease.LeaseID)
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
	m11OutcomeAppendFault = nil
	if err := recoverM11OutcomeJournal(dir); err != nil {
		t.Fatalf("journal recovery failed: %v", err)
	}
	if _, err := os.Stat(m11OutcomeJournalPath(dir)); !os.IsNotExist(err) {
		t.Fatalf("journal still exists after recovery: %v", err)
	}
	outcomes, err := loadM11FixtureOutcomes(dir)
	if err != nil || len(outcomes) != 1 || outcomes[0].OutcomeID != outcome.OutcomeID {
		t.Fatalf("outcome was not recovered exactly: outcomes=%+v err=%v", outcomes, err)
	}
	_, head, err := m11LedgerHead(dir, lease.LeaseID)
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
