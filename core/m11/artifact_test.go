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
	ledgerEntry := m11Entry(t, ArtifactKindLedger, ledger)
	activation := ProductionActivationRecord{LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ActivatedAt: "2026-09-08T00:00:00Z"}
	gate := ProductionGateDecision{GateID: "gate-1", LedgerArtifactID: ledgerEntry.ArtifactID, LedgerContentHash: ledgerEntry.ContentHash, LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, IntentID: cost.IntentID, IntentHash: cost.IntentHash, PolicyVersion: lease.PolicyVersion, RiskClass: "RISK0", HealthSnapshotID: health.SnapshotID, HealthSnapshotHash: health.SnapshotHash, CostBoundID: cost.CostBoundID, CostBoundHash: cost.CostBoundHash, CostBoundMinor: cost.MaxCostMinor, Decision: "ALLOW_PRODUCTION", Reason: "fixture", EvaluatedAt: "2026-09-08T00:00:00Z"}
	authorization := ProductionExecutionAuthorization{AuthorizationID: "auth-1", IntentID: cost.IntentID, IntentHash: cost.IntentHash, PolicyVersion: lease.PolicyVersion, ProductionLeaseID: lease.LeaseID, ProductionLeaseVersion: lease.LeaseVersion, ProductionLeaseHash: lease.LeaseHash, ProductionGateID: gate.GateID, ProductionHealthSnapshotID: health.SnapshotID, ProductionHealthSnapshotHash: health.SnapshotHash, ProductionCostBoundID: cost.CostBoundID, ProductionCostBoundHash: cost.CostBoundHash, ProductionCostBoundMinor: cost.MaxCostMinor, ExecutorID: "fixture_stub", AuthorizedAt: "2026-09-08T00:00:00Z", ExpiresAt: "2026-09-08T00:01:00Z", IdempotencyKey: "key-1", CorrelationID: cost.CorrelationID, ExecutionMode: "GOVERNED_PRODUCTION", ExecutionAuthorized: true}
	execution := ProductionExecutionRecord{ExecutionID: "exec-1", AuthorizationID: authorization.AuthorizationID, ProductionLeaseID: lease.LeaseID, ProductionLeaseVersion: lease.LeaseVersion, ProductionLeaseHash: lease.LeaseHash, ProductionGateID: gate.GateID, ProductionHealthSnapshotID: health.SnapshotID, ProductionHealthSnapshotHash: health.SnapshotHash, ProductionCostBoundID: cost.CostBoundID, ProductionCostBoundHash: cost.CostBoundHash, ProductionCostBoundMinor: cost.MaxCostMinor, IntentID: cost.IntentID, IntentHash: cost.IntentHash, ExecutorID: authorization.ExecutorID, IdempotencyKey: authorization.IdempotencyKey, AttemptedAt: "2026-09-08T00:00:01Z", Status: "FAILED", SideEffectState: "NOT_PERFORMED", CorrelationID: cost.CorrelationID}
	reservationLedger := ledger
	reservationLedger.ExecutionsTotal, reservationLedger.ExecutionsInWindow, reservationLedger.CostMinorTotal, reservationLedger.PendingOutcomes = 1, 1, 10, 1
	reservationLedger.PendingExecutionIDs = []string{execution.ExecutionID}
	reservationLedger.UpdatedAt = "2026-09-08T00:00:00.500Z"
	evaluation := ProductionOutcomeEvaluation{EvaluationID: "evaluation-1", LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ExecutionID: execution.ExecutionID, OutcomeID: "outcome-1", EvaluatedAt: "2026-09-08T00:00:02Z", Result: "FIXTURE_NO_SIDE_EFFECT", EvidenceIDs: []string{"outcome-1"}, Limitations: []string{"fixture only"}, SourceProfile: "OFFLINE_FIXTURE"}
	cycle := ProductionCycleRecord{CycleID: "cycle-1", LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ObservationIDs: []string{"observation-1"}, DecisionID: "decision-1", IntentID: cost.IntentID, IntentHash: cost.IntentHash, GateID: gate.GateID, AuthorizationID: authorization.AuthorizationID, ExecutionID: execution.ExecutionID, OutcomeID: evaluation.OutcomeID, EvaluationID: evaluation.EvaluationID, Status: "CLOSED", OpenedAt: execution.AttemptedAt, ClosedAt: "2026-09-08T00:00:03Z", CorrelationID: cost.CorrelationID}
	entries := []ArtifactEntry{m11Entry(t, ArtifactKindLease, lease), m11Entry(t, ArtifactKindLeaseApproval, approval), m11Entry(t, ArtifactKindHealth, health), m11Entry(t, ArtifactKindCostBound, cost), ledgerEntry, m11Entry(t, ArtifactKindActivation, activation), m11Entry(t, ArtifactKindGate, gate), m11Entry(t, ArtifactKindAuthorization, authorization), m11Entry(t, ArtifactKindLedger, reservationLedger), m11Entry(t, ArtifactKindExecution, execution), m11Entry(t, ArtifactKindEvaluation, evaluation), m11Entry(t, ArtifactKindCycle, cycle)}
	if err := ValidateArtifactGraph(entries); err != nil {
		t.Fatal(err)
	}
	unknownExecution := execution
	unknownExecution.ExecutionID = "exec-unknown"
	unknownExecution.Status = "RECONCILIATION_REQUIRED"
	unknownExecution.SideEffectState = "UNKNOWN"
	unknownReservationLedger := reservationLedger
	unknownReservationLedger.PendingExecutionIDs = []string{unknownExecution.ExecutionID}
	resolution := ProductionReconciliationResolution{ResolutionID: "resolution-1", LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ExecutionID: unknownExecution.ExecutionID, ResolvedBy: "human", ResolverID: "reviewer-1", ResolvedAt: "2026-09-08T00:00:02Z", EffectState: "NOT_PERFORMED", Reason: "fixture review"}
	resolutionEntries := append([]ArtifactEntry(nil), entries[:10]...)
	resolutionEntries[8] = m11Entry(t, ArtifactKindLedger, unknownReservationLedger)
	resolutionEntries[9] = m11Entry(t, ArtifactKindExecution, unknownExecution)
	resolutionEntries = append(resolutionEntries, m11Entry(t, ArtifactKindReconciliation, resolution))
	if err := ValidateArtifactGraph(resolutionEntries); err != nil {
		t.Fatalf("valid reconciliation resolution rejected: %v", err)
	}
	duplicateResolution := resolution
	duplicateResolution.ResolutionID = "resolution-2"
	duplicateResolutionEntries := append([]ArtifactEntry(nil), resolutionEntries...)
	duplicateResolutionEntries = append(duplicateResolutionEntries, m11Entry(t, ArtifactKindReconciliation, duplicateResolution))
	if err := ValidateArtifactGraph(duplicateResolutionEntries); err == nil {
		t.Fatal("two reconciliation resolutions for one UNKNOWN execution were accepted")
	}
	wrongExecutionResolution := resolution
	wrongExecutionResolution.ExecutionID = execution.ExecutionID
	wrongExecutionEntries := append([]ArtifactEntry(nil), resolutionEntries...)
	wrongExecutionEntries[len(wrongExecutionEntries)-1] = m11Entry(t, ArtifactKindReconciliation, wrongExecutionResolution)
	if err := ValidateArtifactGraph(wrongExecutionEntries); err == nil {
		t.Fatal("resolution for a non-UNKNOWN execution was accepted")
	}
	performedResolution := resolution
	performedResolution.EffectState = "PERFORMED"
	performedResolutionEntries := append([]ArtifactEntry(nil), resolutionEntries...)
	performedResolutionEntries[len(performedResolutionEntries)-1] = m11Entry(t, ArtifactKindReconciliation, performedResolution)
	if err := ValidateArtifactGraph(performedResolutionEntries); err == nil {
		t.Fatal("PERFORMED reconciliation resolution was accepted")
	}
	preAttemptResolution := resolution
	preAttemptResolution.ResolvedAt = "2026-09-08T00:00:00Z"
	preAttemptResolutionEntries := append([]ArtifactEntry(nil), resolutionEntries...)
	preAttemptResolutionEntries[len(preAttemptResolutionEntries)-1] = m11Entry(t, ArtifactKindReconciliation, preAttemptResolution)
	if err := ValidateArtifactGraph(preAttemptResolutionEntries); err == nil {
		t.Fatal("reconciliation resolution before its UNKNOWN attempt was accepted")
	}
	brokenGate := gate
	brokenGate.LedgerContentHash = "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	brokenGateEntries := append([]ArtifactEntry(nil), entries...)
	brokenGateEntries[6] = m11Entry(t, ArtifactKindGate, brokenGate)
	if err := ValidateArtifactGraph(brokenGateEntries); err == nil {
		t.Fatal("gate with mismatched ledger head was accepted")
	}
	stoppedGateLedger := ledger
	stoppedGateLedger.ControlMode, stoppedGateLedger.StopReason = "STOPPED", "fixture stop"
	stoppedGateLedgerEntry := m11Entry(t, ArtifactKindLedger, stoppedGateLedger)
	stoppedGate := gate
	stoppedGate.LedgerContentHash = stoppedGateLedgerEntry.ContentHash
	stoppedGateEntries := append([]ArtifactEntry(nil), entries...)
	stoppedGateEntries[4] = stoppedGateLedgerEntry
	stoppedGateEntries[6] = m11Entry(t, ArtifactKindGate, stoppedGate)
	if err := ValidateArtifactGraph(stoppedGateEntries); err == nil {
		t.Fatal("allow gate from a stopped ledger was accepted")
	}
	driftedGate := gate
	driftedGate.ExecutionsTotalBefore = 1
	driftedGateEntries := append([]ArtifactEntry(nil), entries...)
	driftedGateEntries[6] = m11Entry(t, ArtifactKindGate, driftedGate)
	if err := ValidateArtifactGraph(driftedGateEntries); err == nil {
		t.Fatal("gate with drifted ledger counters was accepted")
	}
	lateActivation := activation
	lateActivation.ActivatedAt = "2026-09-08T00:00:01Z"
	activationBoundHealth := health
	activationBoundHealth.ObservedAt = lateActivation.ActivatedAt
	activationBoundHealth.SnapshotHash = ComputeProductionHealthHash(activationBoundHealth)
	preActivationGate := gate
	preActivationGate.EvaluatedAt = "2026-09-08T00:00:00.500Z"
	preActivationGate.HealthSnapshotHash = activationBoundHealth.SnapshotHash
	preActivationGateEntries := append([]ArtifactEntry(nil), entries[:7]...)
	preActivationGateEntries[2] = m11Entry(t, ArtifactKindHealth, activationBoundHealth)
	preActivationGateEntries[5] = m11Entry(t, ArtifactKindActivation, lateActivation)
	preActivationGateEntries[6] = m11Entry(t, ArtifactKindGate, preActivationGate)
	if err := ValidateArtifactGraph(preActivationGateEntries); err == nil {
		t.Fatal("gate before activation was accepted")
	}
	postGateHealth := health
	postGateHealth.ObservedAt = "2026-09-08T00:00:01Z"
	postGateHealth.SnapshotHash = ComputeProductionHealthHash(postGateHealth)
	preHealthGate := gate
	preHealthGate.HealthSnapshotHash = postGateHealth.SnapshotHash
	preHealthGateEntries := append([]ArtifactEntry(nil), entries[:7]...)
	preHealthGateEntries[2] = m11Entry(t, ArtifactKindHealth, postGateHealth)
	preHealthGateEntries[6] = m11Entry(t, ArtifactKindGate, preHealthGate)
	if err := ValidateArtifactGraph(preHealthGateEntries); err == nil {
		t.Fatal("gate before its health observation was accepted")
	}
	degradedHealth := health
	degradedHealth.DependencyState = "DEGRADED"
	degradedHealth.SnapshotHash = ComputeProductionHealthHash(degradedHealth)
	degradedHealthGate := gate
	degradedHealthGate.HealthSnapshotHash = degradedHealth.SnapshotHash
	degradedHealthGateEntries := append([]ArtifactEntry(nil), entries[:7]...)
	degradedHealthGateEntries[2] = m11Entry(t, ArtifactKindHealth, degradedHealth)
	degradedHealthGateEntries[6] = m11Entry(t, ArtifactKindGate, degradedHealthGate)
	if err := ValidateArtifactGraph(degradedHealthGateEntries); err == nil {
		t.Fatal("allow gate with degraded health was accepted")
	}
	exhaustedGateLedger := ledger
	exhaustedGateLedger.ExecutionsTotal, exhaustedGateLedger.ExecutionsInWindow, exhaustedGateLedger.CostMinorTotal, exhaustedGateLedger.PendingOutcomes = 1, 1, 10, 1
	exhaustedGateLedger.PendingExecutionIDs = []string{"reserved-1"}
	exhaustedGateLedgerEntry := m11Entry(t, ArtifactKindLedger, exhaustedGateLedger)
	exhaustedBudgetGate := gate
	exhaustedBudgetGate.LedgerArtifactID, exhaustedBudgetGate.LedgerContentHash = exhaustedGateLedgerEntry.ArtifactID, exhaustedGateLedgerEntry.ContentHash
	exhaustedBudgetGate.ExecutionsTotalBefore, exhaustedBudgetGate.ExecutionsInWindowBefore, exhaustedBudgetGate.CostMinorTotalBefore, exhaustedBudgetGate.PendingOutcomesBefore = exhaustedGateLedger.ExecutionsTotal, exhaustedGateLedger.ExecutionsInWindow, exhaustedGateLedger.CostMinorTotal, exhaustedGateLedger.PendingOutcomes
	exhaustedBudgetGateEntries := append([]ArtifactEntry(nil), entries[:7]...)
	exhaustedBudgetGateEntries[4] = exhaustedGateLedgerEntry
	exhaustedBudgetGateEntries[6] = m11Entry(t, ArtifactKindGate, exhaustedBudgetGate)
	if err := ValidateArtifactGraph(exhaustedBudgetGateEntries); err == nil {
		t.Fatal("allow gate with exhausted ledger budget was accepted")
	}
	leaseScopeBypassGate := gate
	leaseScopeBypassGate.RiskClass = "RISK1"
	leaseScopeBypassGateEntries := append([]ArtifactEntry(nil), entries[:7]...)
	leaseScopeBypassGateEntries[6] = m11Entry(t, ArtifactKindGate, leaseScopeBypassGate)
	if err := ValidateArtifactGraph(leaseScopeBypassGateEntries); err == nil {
		t.Fatal("allow gate outside immutable lease risk scope was accepted")
	}
	unvalidatedApproval := approval
	unvalidatedApproval.ValidatedRiskClasses = []string{"RISK1"}
	unvalidatedApprovalEntries := append([]ArtifactEntry(nil), entries[:7]...)
	unvalidatedApprovalEntries[1] = m11Entry(t, ArtifactKindLeaseApproval, unvalidatedApproval)
	if err := ValidateArtifactGraph(unvalidatedApprovalEntries); err == nil {
		t.Fatal("allow gate outside immutable approval risk scope was accepted")
	}
	brokenGateWindow := gate
	brokenGateWindow.EvaluatedAt = lease.ExpiresAt
	brokenGateWindowEntries := append([]ArtifactEntry(nil), entries...)
	brokenGateWindowEntries[6] = m11Entry(t, ArtifactKindGate, brokenGateWindow)
	if err := ValidateArtifactGraph(brokenGateWindowEntries); err == nil {
		t.Fatal("gate at lease expiry was accepted")
	}
	brokenGateCostWindow := gate
	brokenGateCostWindow.EvaluatedAt = cost.ExpiresAt
	brokenGateCostWindowEntries := append([]ArtifactEntry(nil), entries...)
	brokenGateCostWindowEntries[6] = m11Entry(t, ArtifactKindGate, brokenGateCostWindow)
	if err := ValidateArtifactGraph(brokenGateCostWindowEntries); err == nil {
		t.Fatal("gate at cost-bound expiry was accepted")
	}
	brokenAuthorizationWindow := authorization
	brokenAuthorizationWindow.AuthorizedAt = lease.ExpiresAt
	brokenAuthorizationWindow.ExpiresAt = "2099-09-08T00:01:00Z"
	brokenAuthorizationWindowEntries := append([]ArtifactEntry(nil), entries...)
	brokenAuthorizationWindowEntries[7] = m11Entry(t, ArtifactKindAuthorization, brokenAuthorizationWindow)
	if err := ValidateArtifactGraph(brokenAuthorizationWindowEntries); err == nil {
		t.Fatal("authorization at lease expiry was accepted")
	}
	brokenAuthorizationGate := gate
	brokenAuthorizationGate.Decision = "DENY"
	brokenAuthorizationGateEntries := append([]ArtifactEntry(nil), entries...)
	brokenAuthorizationGateEntries[6] = m11Entry(t, ArtifactKindGate, brokenAuthorizationGate)
	if err := ValidateArtifactGraph(brokenAuthorizationGateEntries); err == nil {
		t.Fatal("authorization from a non-allow gate was accepted")
	}
	unauthorizedExecutor := authorization
	unauthorizedExecutor.ExecutorID = "rogue-executor"
	unauthorizedExecutorEntries := append([]ArtifactEntry(nil), entries[:8]...)
	unauthorizedExecutorEntries[7] = m11Entry(t, ArtifactKindAuthorization, unauthorizedExecutor)
	if err := ValidateArtifactGraph(unauthorizedExecutorEntries); err == nil {
		t.Fatal("authorization for an executor outside the lease scope was accepted")
	}
	staleAuthorization := authorization
	staleAuthorization.AuthorizedAt, staleAuthorization.ExpiresAt = "2026-09-08T00:01:00Z", "2026-09-08T00:02:00Z"
	staleAuthorizationEntries := append([]ArtifactEntry(nil), entries...)
	staleAuthorizationEntries[7] = m11Entry(t, ArtifactKindAuthorization, staleAuthorization)
	if err := ValidateArtifactGraph(staleAuthorizationEntries); err == nil {
		t.Fatal("authorization with stale health was accepted")
	}
	shortCostWindow := cost
	shortCostWindow.ExpiresAt = "2026-09-08T00:00:30Z"
	shortCostWindow.CostBoundHash = corem10.ComputeTrustedCostBoundHash(shortCostWindow)
	shortCostGate := gate
	shortCostGate.CostBoundHash = shortCostWindow.CostBoundHash
	shortCostAuthorization := authorization
	shortCostAuthorization.ProductionCostBoundHash = shortCostWindow.CostBoundHash
	shortCostEntries := append([]ArtifactEntry(nil), entries...)
	shortCostEntries[3] = m11Entry(t, ArtifactKindCostBound, shortCostWindow)
	shortCostEntries[6] = m11Entry(t, ArtifactKindGate, shortCostGate)
	shortCostEntries[7] = m11Entry(t, ArtifactKindAuthorization, shortCostAuthorization)
	if err := ValidateArtifactGraph(shortCostEntries); err == nil {
		t.Fatal("authorization outliving its cost bound was accepted")
	}
	lateExecution := execution
	lateExecution.AttemptedAt = authorization.ExpiresAt
	lateExecutionEntries := append([]ArtifactEntry(nil), entries...)
	lateExecutionEntries[9] = m11Entry(t, ArtifactKindExecution, lateExecution)
	if err := ValidateArtifactGraph(lateExecutionEntries); err == nil {
		t.Fatal("execution at authorization expiry was accepted")
	}
	missingReservationEntries := append([]ArtifactEntry(nil), entries[:8]...)
	missingReservationEntries = append(missingReservationEntries, entries[9:]...)
	if err := ValidateArtifactGraph(missingReservationEntries); err == nil {
		t.Fatal("execution without a prior reservation ledger was accepted")
	}
	postReservationAuthorization := authorization
	postReservationAuthorization.AuthorizedAt = "2026-09-08T00:00:00.750Z"
	postReservationAuthorizationEntries := append([]ArtifactEntry(nil), entries...)
	postReservationAuthorizationEntries[7] = m11Entry(t, ArtifactKindAuthorization, postReservationAuthorization)
	if err := ValidateArtifactGraph(postReservationAuthorizationEntries); err == nil {
		t.Fatal("execution with reservation ledger predating authorization was accepted")
	}
	duplicateAuthorizationExecution := execution
	duplicateAuthorizationExecution.ExecutionID = "exec-duplicate-authorization"
	duplicateAuthorizationExecutionEntries := append([]ArtifactEntry(nil), entries...)
	duplicateAuthorizationExecutionEntries = append(duplicateAuthorizationExecutionEntries, m11Entry(t, ArtifactKindExecution, duplicateAuthorizationExecution))
	if err := ValidateArtifactGraph(duplicateAuthorizationExecutionEntries); err == nil {
		t.Fatal("two execution records for one authorization were accepted")
	}
	brokenActivation := activation
	brokenActivation.ActivatedAt = "2026-09-07T23:59:59Z"
	brokenActivationEntries := append([]ArtifactEntry(nil), entries...)
	brokenActivationEntries[5] = m11Entry(t, ArtifactKindActivation, brokenActivation)
	if err := ValidateArtifactGraph(brokenActivationEntries); err == nil {
		t.Fatal("activation outside its lease lifetime was accepted")
	}
	brokenHealth := health
	brokenHealth.ObservedAt = lease.ExpiresAt
	brokenHealth.SnapshotHash = ComputeProductionHealthHash(brokenHealth)
	brokenHealthEntries := append([]ArtifactEntry(nil), entries...)
	brokenHealthEntries[2] = m11Entry(t, ArtifactKindHealth, brokenHealth)
	if err := ValidateArtifactGraph(brokenHealthEntries); err == nil {
		t.Fatal("health at lease expiry was accepted")
	}
	preActivationHealth := health
	preActivationHealth.ObservedAt = "2026-09-07T23:59:59Z"
	preActivationHealth.SnapshotHash = ComputeProductionHealthHash(preActivationHealth)
	preActivationHealthEntries := append([]ArtifactEntry(nil), entries...)
	preActivationHealthEntries[2] = m11Entry(t, ArtifactKindHealth, preActivationHealth)
	if err := ValidateArtifactGraph(preActivationHealthEntries); err == nil {
		t.Fatal("health before activation was accepted")
	}
	brokenLedger := ledger
	brokenLedger.WindowStartedAt = "2026-09-07T23:59:59Z"
	brokenLedgerEntries := append([]ArtifactEntry(nil), entries[:6]...)
	brokenLedgerEntries[4] = m11Entry(t, ArtifactKindLedger, brokenLedger)
	if err := ValidateArtifactGraph(brokenLedgerEntries); err == nil {
		t.Fatal("ledger before its activation was accepted")
	}
	brokenEvaluation := evaluation
	brokenEvaluation.ExecutionID = "orphan-execution"
	brokenEvaluationEntries := append([]ArtifactEntry(nil), entries...)
	brokenEvaluationEntries[len(brokenEvaluationEntries)-2] = m11Entry(t, ArtifactKindEvaluation, brokenEvaluation)
	if err := ValidateArtifactGraph(brokenEvaluationEntries); err == nil {
		t.Fatal("evaluation with orphan execution was accepted")
	}
	preAttemptEvaluation := evaluation
	preAttemptEvaluation.EvaluatedAt = "2026-09-08T00:00:00Z"
	preAttemptEvaluationEntries := append([]ArtifactEntry(nil), entries...)
	preAttemptEvaluationEntries[len(preAttemptEvaluationEntries)-2] = m11Entry(t, ArtifactKindEvaluation, preAttemptEvaluation)
	if err := ValidateArtifactGraph(preAttemptEvaluationEntries); err == nil {
		t.Fatal("evaluation before its execution attempt was accepted")
	}
	duplicateExecutionEvaluation := evaluation
	duplicateExecutionEvaluation.EvaluationID = "evaluation-duplicate-execution"
	duplicateExecutionEvaluationEntries := append([]ArtifactEntry(nil), entries...)
	duplicateExecutionEvaluationEntries = append(duplicateExecutionEvaluationEntries, m11Entry(t, ArtifactKindEvaluation, duplicateExecutionEvaluation))
	if err := ValidateArtifactGraph(duplicateExecutionEvaluationEntries); err == nil {
		t.Fatal("two evaluations for one execution were accepted")
	}
	brokenCycle := cycle
	brokenCycle.OutcomeID = "orphan-outcome"
	brokenCycleEntries := append([]ArtifactEntry(nil), entries...)
	brokenCycleEntries[len(brokenCycleEntries)-1] = m11Entry(t, ArtifactKindCycle, brokenCycle)
	if err := ValidateArtifactGraph(brokenCycleEntries); err == nil {
		t.Fatal("cycle with mismatched evaluation outcome was accepted")
	}
	preEvaluationCycle := cycle
	preEvaluationCycle.ClosedAt = execution.AttemptedAt
	preEvaluationCycleEntries := append([]ArtifactEntry(nil), entries...)
	preEvaluationCycleEntries[len(preEvaluationCycleEntries)-1] = m11Entry(t, ArtifactKindCycle, preEvaluationCycle)
	if err := ValidateArtifactGraph(preEvaluationCycleEntries); err == nil {
		t.Fatal("cycle closed before its evaluation was accepted")
	}
	pendingCycle := cycle
	pendingCycle.Status = "REVIEW_PENDING"
	pendingCycleEntries := append([]ArtifactEntry(nil), entries...)
	pendingCycleEntries[len(pendingCycleEntries)-1] = m11Entry(t, ArtifactKindCycle, pendingCycle)
	if err := ValidateArtifactGraph(pendingCycleEntries); err == nil {
		t.Fatal("non-closed production cycle was accepted")
	}
	duplicateExecutionCycle := cycle
	duplicateExecutionCycle.CycleID = "cycle-duplicate-execution"
	duplicateExecutionCycleEntries := append([]ArtifactEntry(nil), entries...)
	duplicateExecutionCycleEntries = append(duplicateExecutionCycleEntries, m11Entry(t, ArtifactKindCycle, duplicateExecutionCycle))
	if err := ValidateArtifactGraph(duplicateExecutionCycleEntries); err == nil {
		t.Fatal("two closed cycles for one execution were accepted")
	}
	entries[len(entries)-1].Artifact = json.RawMessage(`{"execution_id":"orphan"}`)
	if err := ValidateArtifactGraph(entries); err == nil {
		t.Fatal("invalid cycle entry accepted")
	}
	spentLedger := ledger
	spentLedger.UpdatedAt = "2026-09-08T00:00:04Z"
	spentLedger.ExecutionsTotal, spentLedger.ExecutionsInWindow, spentLedger.CostMinorTotal, spentLedger.PendingOutcomes = 1, 1, 1, 1
	spentLedger.PendingExecutionIDs = []string{"pending-1"}
	resetLedger := ledger
	resetLedger.UpdatedAt = "2026-09-08T00:00:05Z"
	resetEntries := append([]ArtifactEntry(nil), entries[:len(entries)-1]...)
	resetEntries = append(resetEntries, m11Entry(t, ArtifactKindLedger, spentLedger), m11Entry(t, ArtifactKindLedger, resetLedger))
	if err := ValidateArtifactGraph(resetEntries); err == nil {
		t.Fatal("later ledger reset was accepted")
	}
	windowResetLedger := reservationLedger
	windowResetLedger.WindowStartedAt = "2026-09-08T00:00:00.500Z"
	windowResetLedger.UpdatedAt = "2026-09-08T00:00:00.750Z"
	windowResetLedger.ExecutionsInWindow = 0
	windowResetEntries := append([]ArtifactEntry(nil), entries...)
	windowResetEntries = append(windowResetEntries, m11Entry(t, ArtifactKindLedger, windowResetLedger))
	if err := ValidateArtifactGraph(windowResetEntries); err == nil {
		t.Fatal("later ledger window reset was accepted")
	}
	pendingSwapLedger := reservationLedger
	pendingSwapLedger.UpdatedAt = "2026-09-08T00:00:00.750Z"
	pendingSwapLedger.PendingExecutionIDs = []string{"exec-without-reservation"}
	pendingSwapEntries := append([]ArtifactEntry(nil), entries...)
	pendingSwapEntries = append(pendingSwapEntries, m11Entry(t, ArtifactKindLedger, pendingSwapLedger))
	if err := ValidateArtifactGraph(pendingSwapEntries); err == nil {
		t.Fatal("ledger pending execution swap without a new reservation was accepted")
	}
	completedLedger := reservationLedger
	completedLedger.PendingOutcomes = 0
	completedLedger.PendingExecutionIDs = []string{}
	completedLedger.OutcomeLinks = []ProductionOutcomeLink{{OutcomeID: evaluation.OutcomeID, ExecutionID: execution.ExecutionID, ObservedAt: "2026-09-08T00:00:02Z"}}
	completedLedger.LastOutcomeAt = "2026-09-08T00:00:02Z"
	completedLedger.UpdatedAt = "2026-09-08T00:00:02Z"
	erasedOutcomeLedger := completedLedger
	erasedOutcomeLedger.OutcomeLinks = []ProductionOutcomeLink{}
	erasedOutcomeLedger.UpdatedAt = "2026-09-08T00:00:03Z"
	erasedOutcomeEntries := append([]ArtifactEntry(nil), entries...)
	erasedOutcomeEntries = append(erasedOutcomeEntries, m11Entry(t, ArtifactKindLedger, completedLedger), m11Entry(t, ArtifactKindLedger, erasedOutcomeLedger))
	if err := ValidateArtifactGraph(erasedOutcomeEntries); err == nil {
		t.Fatal("later ledger erased an immutable outcome link")
	}
}

func TestRecoveryAdmissionIsNonAuthorizingAndCannotReusePriorIdentity(t *testing.T) {
	valid := ProductionRecoveryAdmission{
		RecoveryAdmissionID: "recovery-1", PriorRuntimeDir: "/runtime/old", PriorLeaseID: "lease-old", PriorLeaseVersion: "v1",
		PriorLeaseHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", PriorApprovalID: "approval-old",
		ResolutionID: "resolution-1", NewRuntimeID: "runtime-new", NewRuntimeDir: "/runtime/new", NewLeaseID: "lease-new",
		NewLeaseVersion: "v1", NewLeaseHash: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		NewApprovalID: "approval-new", ReviewedBy: "human", ReviewerID: "reviewer-new", ReviewedAt: "2026-09-08T00:00:07Z", ExecutionPermitted: false,
	}
	raw, err := json.Marshal(valid)
	if err != nil {
		t.Fatal(err)
	}
	if _, status := DecodeArtifact("recovery_admission", raw); status != Valid {
		t.Fatalf("valid recovery admission rejected: %s", status)
	}
	entry, err := NewArtifactEntry(ArtifactKindRecoveryAdmission, raw)
	if err != nil || entry.ArtifactID != valid.RecoveryAdmissionID {
		t.Fatalf("recovery admission was not registered canonically: entry=%+v err=%v", entry, err)
	}
	badExecution := valid
	badExecution.ExecutionPermitted = true
	raw, _ = json.Marshal(badExecution)
	if _, status := DecodeArtifact("recovery_admission", raw); status == Valid {
		t.Fatal("execution-permitted recovery admission was accepted")
	}
	reusedLease := valid
	reusedLease.NewLeaseID, reusedLease.NewLeaseHash = reusedLease.PriorLeaseID, reusedLease.PriorLeaseHash
	raw, _ = json.Marshal(reusedLease)
	if _, status := DecodeArtifact("recovery_admission", raw); status == Valid {
		t.Fatal("prior lease reuse was accepted")
	}
	reusedApproval := valid
	reusedApproval.NewApprovalID = reusedApproval.PriorApprovalID
	raw, _ = json.Marshal(reusedApproval)
	if _, status := DecodeArtifact("recovery_admission", raw); status == Valid {
		t.Fatal("prior approval reuse was accepted")
	}
}
