package m11

import (
	"encoding/json"
	"testing"

	corem10 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m10"
)

func m11Entry(t *testing.T, kind string, value any) ArtifactEntry {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := NewArtifactEntry(kind, raw)
	if err != nil {
		t.Fatalf("%s: %v", kind, err)
	}
	return entry
}

func TestArtifactGraphAcceptsExactProductionLifecycleLinks(t *testing.T) {
	lease := ProductionLease{LeaseID: "lease-1", LeaseVersion: "v1", PolicyVersion: "policy-1", ApprovalRef: "approval-1", ReviewedBy: "human", ReviewerID: "reviewer-1", ReviewedAt: "2026-09-08T00:00:00Z", PromotionReviewRef: "review-1", SourceCanaryGrantID: "grant-1", SourceCanaryGrantVersion: "v1", SourceCanaryGrantHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ValidFrom: "2026-09-08T00:00:00Z", ExpiresAt: "2099-09-08T00:00:00Z", AllowedRiskClasses: []string{"RISK0"}, AllowedActionTypes: []string{"DRAFT"}, AllowedHosts: []string{"example.com"}, ExecutorIDs: []string{"fixture_stub"}, MaxExecutionsTotal: 1, MaxExecutionsPerWindow: 1, WindowSeconds: 60, MaxCostMinorTotal: 10, Currency: "USD", MaxPendingOutcomes: 1, MaxConsecutiveFailures: 1, MaxOutcomeAgeSeconds: 60, MaxHealthSnapshotAgeSeconds: 60, KillSwitchRequired: true, CorrelationID: "corr-1", HashVersion: "go-json-v1"}
	lease.LeaseHash = ComputeProductionLeaseHash(lease)
	approval := ProductionLeaseApproval{ApprovalID: "approval-1", LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, PromotionReviewRef: lease.PromotionReviewRef, SourceCanaryGrantID: lease.SourceCanaryGrantID, SourceCanaryGrantVersion: lease.SourceCanaryGrantVersion, SourceCanaryGrantHash: lease.SourceCanaryGrantHash, SourceE5Refs: []string{"e5-1"}, ValidatedRiskClasses: []string{"RISK0"}, ReviewedBy: "human", ReviewerID: lease.ReviewerID, ReviewedAt: lease.ReviewedAt, Decision: "APPROVE_PRODUCTION_LEASE"}
	health := ProductionHealthSnapshot{SnapshotID: "health-1", LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ObservedAt: "2026-09-08T00:00:00Z", SourceRefs: []string{"fixture:health"}, DependencyState: "HEALTHY", TelemetryComplete: true, HashVersion: "go-json-v1"}
	health.SnapshotHash = ComputeProductionHealthHash(health)
	cost := corem10.TrustedCostBound{CostBoundID: "cost-1", IntentID: "intent-1", IntentHash: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", MaxCostMinor: 10, Currency: "USD", SourceRef: "fixture:cost", ObservedAt: "2026-09-08T00:00:00Z", ExpiresAt: "2099-09-08T00:00:00Z", CorrelationID: lease.CorrelationID, HashVersion: "go-json-v1"}
	cost.CostBoundHash = corem10.ComputeTrustedCostBoundHash(cost)
	ledger := ProductionLedger{LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ControlMode: "NORMAL", WindowStartedAt: "2026-09-08T00:00:00Z", PendingExecutionIDs: []string{}, SuccessfulIdempotencyKeys: []string{}, OutcomeLinks: []ProductionOutcomeLink{}, ReconciliationResolutionIDs: []string{}, UpdatedAt: "2026-09-08T00:00:00Z"}
	activation := ProductionActivationRecord{LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ActivatedAt: "2026-09-08T00:00:00Z"}
	gate := ProductionGateDecision{GateID: "gate-1", LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, IntentID: cost.IntentID, IntentHash: cost.IntentHash, PolicyVersion: lease.PolicyVersion, RiskClass: "RISK0", HealthSnapshotID: health.SnapshotID, HealthSnapshotHash: health.SnapshotHash, CostBoundID: cost.CostBoundID, CostBoundHash: cost.CostBoundHash, CostBoundMinor: cost.MaxCostMinor, Decision: "ALLOW_PRODUCTION", Reason: "fixture", EvaluatedAt: "2026-09-08T00:00:00Z"}
	authorization := ProductionExecutionAuthorization{AuthorizationID: "auth-1", IntentID: cost.IntentID, IntentHash: cost.IntentHash, PolicyVersion: lease.PolicyVersion, ProductionLeaseID: lease.LeaseID, ProductionLeaseVersion: lease.LeaseVersion, ProductionLeaseHash: lease.LeaseHash, ProductionGateID: gate.GateID, ProductionHealthSnapshotID: health.SnapshotID, ProductionHealthSnapshotHash: health.SnapshotHash, ProductionCostBoundID: cost.CostBoundID, ProductionCostBoundHash: cost.CostBoundHash, ProductionCostBoundMinor: cost.MaxCostMinor, ExecutorID: "fixture_stub", AuthorizedAt: "2026-09-08T00:00:00Z", ExpiresAt: "2026-09-08T00:01:00Z", IdempotencyKey: "key-1", CorrelationID: cost.CorrelationID, ExecutionMode: "GOVERNED_PRODUCTION", ExecutionAuthorized: true}
	execution := ProductionExecutionRecord{ExecutionID: "exec-1", AuthorizationID: authorization.AuthorizationID, ProductionLeaseID: lease.LeaseID, ProductionLeaseVersion: lease.LeaseVersion, ProductionLeaseHash: lease.LeaseHash, ProductionGateID: gate.GateID, ProductionHealthSnapshotID: health.SnapshotID, ProductionHealthSnapshotHash: health.SnapshotHash, ProductionCostBoundID: cost.CostBoundID, ProductionCostBoundHash: cost.CostBoundHash, ProductionCostBoundMinor: cost.MaxCostMinor, IntentID: cost.IntentID, IntentHash: cost.IntentHash, ExecutorID: authorization.ExecutorID, IdempotencyKey: authorization.IdempotencyKey, AttemptedAt: "2026-09-08T00:00:01Z", Status: "FAILED", SideEffectState: "NOT_PERFORMED", CorrelationID: cost.CorrelationID}
	evaluation := ProductionOutcomeEvaluation{EvaluationID: "evaluation-1", LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ExecutionID: execution.ExecutionID, OutcomeID: "outcome-1", EvaluatedAt: "2026-09-08T00:00:02Z", Result: "FIXTURE_NO_SIDE_EFFECT", EvidenceIDs: []string{"outcome-1"}, Limitations: []string{"fixture only"}, SourceProfile: "OFFLINE_FIXTURE"}
	cycle := ProductionCycleRecord{CycleID: "cycle-1", LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ObservationIDs: []string{"observation-1"}, DecisionID: "decision-1", IntentID: cost.IntentID, IntentHash: cost.IntentHash, GateID: gate.GateID, AuthorizationID: authorization.AuthorizationID, ExecutionID: execution.ExecutionID, OutcomeID: evaluation.OutcomeID, EvaluationID: evaluation.EvaluationID, Status: "CLOSED", OpenedAt: execution.AttemptedAt, ClosedAt: "2026-09-08T00:00:03Z", CorrelationID: cost.CorrelationID}
	entries := []ArtifactEntry{m11Entry(t, ArtifactKindLease, lease), m11Entry(t, ArtifactKindLeaseApproval, approval), m11Entry(t, ArtifactKindHealth, health), m11Entry(t, ArtifactKindCostBound, cost), m11Entry(t, ArtifactKindLedger, ledger), m11Entry(t, ArtifactKindActivation, activation), m11Entry(t, ArtifactKindGate, gate), m11Entry(t, ArtifactKindAuthorization, authorization), m11Entry(t, ArtifactKindExecution, execution), m11Entry(t, ArtifactKindEvaluation, evaluation), m11Entry(t, ArtifactKindCycle, cycle)}
	if err := ValidateArtifactGraph(entries); err != nil {
		t.Fatal(err)
	}
	brokenEvaluation := evaluation
	brokenEvaluation.ExecutionID = "orphan-execution"
	brokenEvaluationEntries := append([]ArtifactEntry(nil), entries...)
	brokenEvaluationEntries[len(brokenEvaluationEntries)-2] = m11Entry(t, ArtifactKindEvaluation, brokenEvaluation)
	if err := ValidateArtifactGraph(brokenEvaluationEntries); err == nil {
		t.Fatal("evaluation with orphan execution was accepted")
	}
	brokenCycle := cycle
	brokenCycle.OutcomeID = "orphan-outcome"
	brokenCycleEntries := append([]ArtifactEntry(nil), entries...)
	brokenCycleEntries[len(brokenCycleEntries)-1] = m11Entry(t, ArtifactKindCycle, brokenCycle)
	if err := ValidateArtifactGraph(brokenCycleEntries); err == nil {
		t.Fatal("cycle with mismatched evaluation outcome was accepted")
	}
	entries[len(entries)-1].Artifact = json.RawMessage(`{"execution_id":"orphan"}`)
	if err := ValidateArtifactGraph(entries); err == nil {
		t.Fatal("invalid cycle entry accepted")
	}
}
