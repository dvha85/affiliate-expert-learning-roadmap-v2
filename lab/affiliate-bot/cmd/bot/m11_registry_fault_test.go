package main

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

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
	activation := corem11.ProductionActivationRecord{LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ActivatedAt: "2026-09-08T00:00:00Z"}
	gate := corem11.ProductionGateDecision{GateID: "unknown-journal-gate", LedgerArtifactID: ledgerEntry.ArtifactID, LedgerContentHash: ledgerEntry.ContentHash, LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, IntentID: cost.IntentID, IntentHash: cost.IntentHash, PolicyVersion: lease.PolicyVersion, RiskClass: "RISK0", HealthSnapshotID: health.SnapshotID, HealthSnapshotHash: health.SnapshotHash, CostBoundID: cost.CostBoundID, CostBoundHash: cost.CostBoundHash, CostBoundMinor: cost.MaxCostMinor, Decision: "ALLOW_PRODUCTION", Reason: "fixture", EvaluatedAt: "2026-09-08T00:00:00Z"}
	authorization := corem11.ProductionExecutionAuthorization{AuthorizationID: "unknown-journal-auth", IntentID: cost.IntentID, IntentHash: cost.IntentHash, PolicyVersion: lease.PolicyVersion, ProductionLeaseID: lease.LeaseID, ProductionLeaseVersion: lease.LeaseVersion, ProductionLeaseHash: lease.LeaseHash, ProductionGateID: gate.GateID, ProductionHealthSnapshotID: health.SnapshotID, ProductionHealthSnapshotHash: health.SnapshotHash, ProductionCostBoundID: cost.CostBoundID, ProductionCostBoundHash: cost.CostBoundHash, ProductionCostBoundMinor: cost.MaxCostMinor, ExecutorID: "fixture_stub", AuthorizedAt: "2026-09-08T00:00:00Z", ExpiresAt: "2026-09-08T00:01:00Z", IdempotencyKey: "unknown-journal-key", CorrelationID: cost.CorrelationID, ExecutionMode: "GOVERNED_PRODUCTION", ExecutionAuthorized: true}
	for _, item := range []struct {
		kind  string
		value any
	}{{corem11.ArtifactKindLease, lease}, {corem11.ArtifactKindLeaseApproval, approval}, {corem11.ArtifactKindHealth, health}, {corem11.ArtifactKindCostBound, cost}, {corem11.ArtifactKindLedger, ledger}, {corem11.ArtifactKindActivation, activation}, {corem11.ArtifactKindGate, gate}, {corem11.ArtifactKindAuthorization, authorization}} {
		registerM11TestArtifact(t, dir, item.kind, item.value)
	}
	return m11UnknownStopFixture{dir: dir, lease: lease, authorization: authorization, ledgerEntry: ledgerEntry}
}

func TestM11UnknownStopJournalRecoversAfterStoppedLedgerWriteFailure(t *testing.T) {
	for _, phase := range []string{"before_write", "after_write"} {
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

func TestM11UnknownStopJournalRecoversAfterMissionStateWriteFailure(t *testing.T) {
	fixture := newM11UnknownStopFixture(t)
	const attemptedAt, reason = "2026-09-08T00:00:02Z", "fixture timeout"
	atomicWrites := 0
	missionStateWriteFault = func(phase string) error {
		if phase != "before_rename" {
			return nil
		}
		atomicWrites++ // journal first, then mission-state after stopped ledger
		if atomicWrites == 2 {
			return errors.New("injected mission-state write failure")
		}
		return nil
	}
	t.Cleanup(func() { missionStateWriteFault = nil })
	if _, _, _, err := recordUnknownM11Execution(fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, reason); err == nil {
		t.Fatal("mission-state write fault was not surfaced")
	}
	if atomicWrites != 2 {
		t.Fatalf("fault did not reach mission-state write: atomic writes=%d", atomicWrites)
	}
	if _, err := os.Stat(m11UnknownStopJournalPath(fixture.dir)); err != nil {
		t.Fatalf("unknown STOP journal was removed before recovery: %v", err)
	}
	if code, response := missionCall(t, "status", fixture.dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
		t.Fatalf("status did not fail closed on mission-state write fault: code=%d response=%+v", code, response)
	}
	missionStateWriteFault = nil
	if err := recoverM11UnknownStopJournal(fixture.dir); err != nil {
		t.Fatalf("unknown STOP journal did not recover mission-state write failure: %v", err)
	}
	assertM11UnknownStopRecovered(t, fixture, attemptedAt, reason)
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
