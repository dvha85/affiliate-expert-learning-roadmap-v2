package m11

import (
	"reflect"
	"strings"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m05"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m08"
	corem10 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m10"
)

// HistoricalChain is the decoded, read-only M11 audit input. It is deliberately
// a value object: validation never consults the wall clock, persists state, or
// grants an execution right.
type HistoricalChain struct {
	Profile       string
	Intent        m08.Intent
	Policy        m08.PolicyDecision
	Lease         ProductionLease
	Approval      ProductionLeaseApproval
	Health        ProductionHealthSnapshot
	Cost          corem10.TrustedCostBound
	PreLedger     ProductionLedger
	PostLedger    ProductionLedger
	Gate          ProductionGateDecision
	Authorization ProductionExecutionAuthorization
	Execution     ProductionExecutionRecord
	Activation    ProductionActivationRecord
	Resolution    *ProductionReconciliationResolution
	StopLedger    *ProductionLedger
	Cycle         *ProductionCycleRecord
	Outcome       *m03.OutcomeRecord
	Evaluation    *m05.EvaluationRecord
	Proposal      *m05.ImprovementProposal
	Review        *m05.ReviewRecord
}

func containsHistorical(values []string, wanted string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(wanted)) {
			return true
		}
	}
	return false
}

func containsHistoricalExact(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func historicalHostAllowed(target string, hosts []string) bool {
	return m08.AllowedHost(target, hosts)
}

func historicalTime(value string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339, value)
	return parsed, err == nil
}

func historicalClosedCycle(c HistoricalChain) string {
	cycle, outcome, evaluation := c.Cycle, c.Outcome, c.Evaluation
	if cycle == nil || outcome == nil || evaluation == nil {
		return "BROKEN_CYCLE"
	}
	if strings.TrimSpace(cycle.CycleID) == "" || cycle.LeaseID != c.Lease.LeaseID || cycle.LeaseVersion != c.Lease.LeaseVersion ||
		cycle.LeaseHash != c.Lease.LeaseHash || len(cycle.ObservationIDs) == 0 || cycle.DecisionID != c.Intent.DecisionID ||
		cycle.IntentID != c.Intent.IntentID || cycle.IntentHash != c.Intent.IntentHash || cycle.GateID != c.Gate.GateID ||
		cycle.AuthorizationID != c.Authorization.AuthorizationID || cycle.ExecutionID != c.Execution.ExecutionID ||
		cycle.OutcomeID != outcome.OutcomeID || cycle.EvaluationID != evaluation.EvaluationID || cycle.CorrelationID != c.Intent.CorrelationID {
		return "BROKEN_LINK"
	}
	for _, id := range cycle.ObservationIDs {
		if !containsHistorical(c.Intent.EvidenceIDs, id) {
			return "BROKEN_LINK"
		}
	}
	opened, openOK := historicalTime(cycle.OpenedAt)
	closed, closeOK := historicalTime(cycle.ClosedAt)
	if !openOK || !closeOK || closed.Before(opened) {
		return "INVALID"
	}
	gateAt, gateOK := historicalTime(c.Gate.EvaluatedAt)
	if !gateOK || opened.After(gateAt) || closed.Before(gateAt) {
		return "BROKEN_CYCLE"
	}
	if c.Gate.Decision != "ALLOW_PRODUCTION" || c.Gate.ExecutionAuthorized || !c.Authorization.ExecutionAuthorized ||
		c.Authorization.ExecutionMode != "GOVERNED_PRODUCTION" || c.Authorization.IntentID != c.Intent.IntentID ||
		c.Authorization.IntentHash != c.Intent.IntentHash || c.Authorization.ProductionLeaseHash != c.Lease.LeaseHash ||
		c.Execution.AuthorizationID != c.Authorization.AuthorizationID || c.Execution.IntentHash != c.Authorization.IntentHash ||
		c.Execution.Status != "SUCCEEDED" || c.Execution.SideEffectState != "PERFORMED" {
		return "BROKEN_LINK"
	}
	if m03.ValidateOutcomeRecord(*outcome) != "VALID" || outcome.EffectRef.EffectKind != "MACHINE_EXECUTION" || outcome.EffectRef.EffectID != c.Execution.ExecutionID {
		return "BROKEN_LINK"
	}
	outcomeAt, outcomeOK := historicalTime(outcome.ObservedAt)
	execAt, execOK := historicalTime(c.Execution.AttemptedAt)
	if !outcomeOK || !execOK || outcomeAt.Before(execAt) {
		return "OUTCOME_BEFORE_ACTION"
	}
	if evaluation.DecisionID != c.Intent.DecisionID || evaluation.EffectRef.EffectKind != "MACHINE_EXECUTION" ||
		evaluation.EffectRef.EffectID != c.Execution.ExecutionID || len(evaluation.OutcomeIDs) == 0 ||
		!containsHistorical(evaluation.OutcomeIDs, outcome.OutcomeID) || evaluation.EvaluationID != cycle.EvaluationID {
		return "BROKEN_LINK"
	}
	for _, id := range evaluation.EvidenceIDs {
		if !containsHistoricalExact(c.Intent.EvidenceIDs, id) {
			return "BROKEN_CYCLE"
		}
	}
	if !map[string]bool{"SUPPORTED": true, "NOT_SUPPORTED": true, "INCONCLUSIVE": true, "NEEDS_MORE_DATA": true}[evaluation.Result] {
		return "INVALID"
	}
	evaluated, evaluatedOK := historicalTime(evaluation.EvaluatedAt)
	if !evaluatedOK || evaluated.Before(outcomeAt) {
		return "INVALID"
	}
	if c.Proposal == nil {
		if cycle.ImprovementProposalID != "" || cycle.ReviewID != "" || cycle.Status != "CLOSED" {
			return "BROKEN_LINK"
		}
		return Valid
	}
	proposal := c.Proposal
	if proposal.AutoApply {
		return "REJECT_AUTO_APPLY"
	}
	if strings.TrimSpace(proposal.ProposalID) == "" || len(proposal.EvaluationIDs) == 0 ||
		strings.TrimSpace(proposal.ChangeSummary) == "" || strings.TrimSpace(proposal.ExpectedBenefit) == "" ||
		strings.TrimSpace(proposal.Rollback) == "" || strings.TrimSpace(proposal.CurrentVersion) == "" ||
		strings.TrimSpace(proposal.ProposedVersion) == "" || proposal.CurrentVersion == proposal.ProposedVersion ||
		len(proposal.EvaluationIDs) != 1 || proposal.EvaluationIDs[0] != evaluation.EvaluationID || cycle.ImprovementProposalID != proposal.ProposalID {
		return "BROKEN_LINK"
	}
	if c.Review == nil {
		if cycle.Status != "REVIEW_PENDING" || cycle.ReviewID != "" {
			return "BROKEN_LINK"
		}
		return "REVIEW_REQUIRED"
	}
	review := c.Review
	reviewed, reviewedOK := historicalTime(review.ReviewedAt)
	if review.ReviewID == "" || review.ProposalID != proposal.ProposalID || review.ReviewedBy != "human" ||
		strings.TrimSpace(review.Reason) == "" || !reviewedOK ||
		!map[string]bool{"APPROVE_FOR_MANUAL_CHANGE": true, "REJECT": true, "REQUEST_CHANGES": true}[review.Decision] ||
		cycle.ReviewID != review.ReviewID || cycle.Status != "CLOSED" {
		return "BROKEN_LINK"
	}
	if reviewed.Before(evaluated) || reviewed.After(closed) {
		return "INVALID_TIME_BINDING"
	}
	return Valid
}

// ValidateClosedCycle validates the canonical machine-execution outcome
// linkage used by the closed-cycle profile. It is exported so adapters that
// still expose a legacy local state type can reuse the same implementation.
func ValidateClosedCycle(
	cycle ProductionCycleRecord,
	intent m08.Intent,
	lease ProductionLease,
	gate ProductionGateDecision,
	authorization ProductionExecutionAuthorization,
	execution ProductionExecutionRecord,
	outcome m03.OutcomeRecord,
	evaluation m05.EvaluationRecord,
	proposal *m05.ImprovementProposal,
	review *m05.ReviewRecord,
) string {
	return historicalClosedCycle(HistoricalChain{
		Intent: intent, Lease: lease, Gate: gate, Authorization: authorization,
		Execution: execution, Cycle: &cycle, Outcome: &outcome, Evaluation: &evaluation,
		Proposal: proposal, Review: review,
	})
}

// ValidateHistoricalChain validates both M11 audit profiles. A VALID result
// means the supplied historical records are internally consistent only; it is
// not proof of provenance, provider success, or permission to execute.
func ValidateHistoricalChain(c HistoricalChain) string {
	if c.Profile != "closed_cycle" && c.Profile != "resolved_stop" {
		return "INVALID_PROFILE"
	}
	for _, value := range []string{
		c.Intent.IntentID, c.Intent.DecisionID, c.Intent.CorrelationID, c.Intent.IdempotencyKey,
		c.Policy.PolicyVersion, c.Lease.LeaseID, c.Lease.LeaseVersion, c.Lease.LeaseHash,
		c.Approval.ApprovalID, c.Health.SnapshotID, c.Cost.CostBoundID, c.PreLedger.LeaseID,
		c.PostLedger.LeaseID, c.Gate.GateID, c.Authorization.AuthorizationID, c.Execution.ExecutionID,
		c.Activation.LeaseID,
	} {
		if strings.TrimSpace(value) == "" {
			return "INVALID_IDENTITY"
		}
	}
	if c.Intent.IntentHash != m08.ComputeIntentHash(c.Intent) {
		return "TAMPERED_INTENT"
	}
	if c.Lease.LeaseHash != ComputeProductionLeaseHash(c.Lease) || c.Health.SnapshotHash != ComputeProductionHealthHash(c.Health) {
		return "BROKEN_LINK"
	}
	if c.Lease.LeaseID != c.Approval.LeaseID || c.Lease.LeaseVersion != c.Approval.LeaseVersion || c.Lease.LeaseHash != c.Approval.LeaseHash ||
		c.Lease.PromotionReviewRef != c.Approval.PromotionReviewRef || c.Lease.SourceCanaryGrantID != c.Approval.SourceCanaryGrantID ||
		c.Lease.SourceCanaryGrantVersion != c.Approval.SourceCanaryGrantVersion || c.Lease.SourceCanaryGrantHash != c.Approval.SourceCanaryGrantHash ||
		c.Lease.ReviewerID != c.Approval.ReviewerID || c.Lease.ReviewedAt != c.Approval.ReviewedAt || c.Lease.ApprovalRef != c.Approval.ApprovalID {
		return "BROKEN_LINK"
	}
	if c.Intent.IntentID != c.Policy.IntentID || c.Intent.IntentHash != c.Policy.IntentHash ||
		c.Intent.IntentID != c.Gate.IntentID || c.Intent.IntentHash != c.Gate.IntentHash ||
		c.Intent.IntentID != c.Authorization.IntentID ||
		c.Intent.CorrelationID != c.Cost.CorrelationID || c.Intent.CorrelationID != c.Authorization.CorrelationID ||
		c.Intent.IdempotencyKey != c.Authorization.IdempotencyKey || c.Intent.IdempotencyKey != c.Execution.IdempotencyKey ||
		c.Intent.IntentID != c.Execution.IntentID || c.Intent.IntentHash != c.Execution.IntentHash {
		return "BROKEN_LINK"
	}
	if c.Policy.PolicyVersion != c.Lease.PolicyVersion || c.Policy.PolicyVersion != c.Gate.PolicyVersion || c.Policy.PolicyVersion != c.Authorization.PolicyVersion ||
		c.Lease.LeaseID != c.Health.LeaseID || c.Lease.LeaseVersion != c.Health.LeaseVersion || c.Lease.LeaseHash != c.Health.LeaseHash ||
		c.Lease.LeaseID != c.PreLedger.LeaseID || c.Lease.LeaseVersion != c.PreLedger.LeaseVersion || c.Lease.LeaseHash != c.PreLedger.LeaseHash ||
		c.Lease.LeaseID != c.PostLedger.LeaseID || c.Lease.LeaseVersion != c.PostLedger.LeaseVersion || c.Lease.LeaseHash != c.PostLedger.LeaseHash {
		return "BROKEN_LINK"
	}
	if c.Gate.LeaseID != c.Lease.LeaseID || c.Gate.LeaseVersion != c.Lease.LeaseVersion || c.Gate.LeaseHash != c.Lease.LeaseHash ||
		c.Authorization.ProductionLeaseID != c.Lease.LeaseID || c.Authorization.ProductionLeaseVersion != c.Lease.LeaseVersion || c.Authorization.ProductionLeaseHash != c.Lease.LeaseHash ||
		c.Execution.ProductionLeaseID != c.Lease.LeaseID || c.Execution.ProductionLeaseVersion != c.Lease.LeaseVersion || c.Execution.ProductionLeaseHash != c.Lease.LeaseHash ||
		c.Authorization.AuthorizationID != c.Execution.AuthorizationID || c.Authorization.ExecutorID != c.Execution.ExecutorID ||
		c.Gate.GateID != c.Authorization.ProductionGateID || c.Gate.GateID != c.Execution.ProductionGateID ||
		c.Health.SnapshotID != c.Gate.HealthSnapshotID || c.Health.SnapshotHash != c.Gate.HealthSnapshotHash ||
		c.Health.SnapshotID != c.Authorization.ProductionHealthSnapshotID || c.Health.SnapshotHash != c.Authorization.ProductionHealthSnapshotHash ||
		c.Health.SnapshotID != c.Execution.ProductionHealthSnapshotID || c.Health.SnapshotHash != c.Execution.ProductionHealthSnapshotHash ||
		c.Cost.CostBoundID != c.Gate.CostBoundID || c.Cost.CostBoundHash != c.Gate.CostBoundHash || c.Cost.MaxCostMinor != c.Gate.CostBoundMinor ||
		c.Cost.CostBoundID != c.Authorization.ProductionCostBoundID || c.Cost.CostBoundHash != c.Authorization.ProductionCostBoundHash || c.Cost.MaxCostMinor != c.Authorization.ProductionCostBoundMinor ||
		c.Cost.CostBoundID != c.Execution.ProductionCostBoundID || c.Cost.CostBoundHash != c.Execution.ProductionCostBoundHash || c.Cost.MaxCostMinor != c.Execution.ProductionCostBoundMinor {
		return "BROKEN_LINK"
	}
	if c.Activation.LeaseID != c.Lease.LeaseID || c.Activation.LeaseVersion != c.Lease.LeaseVersion || c.Activation.LeaseHash != c.Lease.LeaseHash {
		return "BROKEN_LINK"
	}
	if c.Gate.RiskClass != c.Policy.RiskClass || c.Cost.Currency != c.Lease.Currency || c.Cost.CostBoundHash != corem10.ComputeTrustedCostBoundHash(c.Cost) || (c.Policy.Decision != "ALLOW" && c.Policy.Decision != "HUMAN_REVIEW") {
		return "BROKEN_LINK"
	}
	if c.Gate.Decision != "ALLOW_PRODUCTION" || c.Gate.Reason != "PRODUCTION_ELIGIBLE" ||
		!((c.Policy.RiskClass == "RISK0" && c.Policy.Decision == "ALLOW") || (c.Policy.RiskClass == "RISK1" && c.Policy.Decision == "HUMAN_REVIEW" && c.Policy.Reason == "RISK1_REQUIRES_REVIEW")) {
		return "INVALID_GATE_STATE"
	}
	if !containsHistorical(c.Lease.AllowedRiskClasses, c.Policy.RiskClass) || !containsHistorical(c.Approval.ValidatedRiskClasses, c.Policy.RiskClass) ||
		!containsHistorical(c.Lease.AllowedActionTypes, c.Intent.ActionType) || !historicalHostAllowed(c.Intent.Target, c.Lease.AllowedHosts) ||
		!containsHistorical(c.Lease.ExecutorIDs, c.Authorization.ExecutorID) {
		return "SCOPE_NOT_DELEGATED"
	}
	if c.PreLedger.ControlMode != "NORMAL" || c.PreLedger.StopReason != "" || c.PreLedger.ReconciliationRequired || len(c.PreLedger.ReconciliationResolutionIDs) != 0 ||
		c.PreLedger.ConsecutiveFailures >= c.Lease.MaxConsecutiveFailures || c.Health.ReconciliationRequired || c.Health.ComplianceAlertCount > 0 ||
		c.Health.ConsecutiveFailures >= c.Lease.MaxConsecutiveFailures || c.Health.OldestPendingOutcomeAgeSeconds > c.Lease.MaxOutcomeAgeSeconds ||
		c.Health.DependencyState != "HEALTHY" || !c.Health.TelemetryComplete {
		return "SAFETY_BLOCKED"
	}
	now, nowOK := historicalTime(c.Gate.EvaluatedAt)
	start, startOK := historicalTime(c.Lease.ValidFrom)
	leaseEnd, leaseEndOK := historicalTime(c.Lease.ExpiresAt)
	if !nowOK || !startOK || !leaseEndOK {
		return "INVALID_TIME_BINDING"
	}
	intentEnd, intentEndOK := historicalTime(c.Intent.ExpiresAt)
	created, createdOK := historicalTime(c.Intent.CreatedAt)
	policyChecked, policyOK := historicalTime(c.Policy.PolicyCheckedAt)
	approvalAt, approvalOK := historicalTime(c.Approval.ReviewedAt)
	leaseReviewedAt, leaseReviewedOK := historicalTime(c.Lease.ReviewedAt)
	costObserved, costObservedOK := historicalTime(c.Cost.ObservedAt)
	costEnd, costEndOK := historicalTime(c.Cost.ExpiresAt)
	healthObserved, healthObservedOK := historicalTime(c.Health.ObservedAt)
	authorizedAt, authorizedOK := historicalTime(c.Authorization.AuthorizedAt)
	authEnd, authEndOK := historicalTime(c.Authorization.ExpiresAt)
	attemptedAt, attemptedOK := historicalTime(c.Execution.AttemptedAt)
	activatedAt, activatedOK := historicalTime(c.Activation.ActivatedAt)
	windowStarted, windowOK := historicalTime(c.PreLedger.WindowStartedAt)
	preUpdated, preUpdatedOK := historicalTime(c.PreLedger.UpdatedAt)
	postUpdated, postUpdatedOK := historicalTime(c.PostLedger.UpdatedAt)
	if !intentEndOK || !createdOK || !policyOK || !approvalOK || !leaseReviewedOK || !costObservedOK || !costEndOK || !healthObservedOK || !authorizedOK || !authEndOK || !attemptedOK || !activatedOK || !windowOK || !preUpdatedOK || !postUpdatedOK {
		return "INVALID_TIME_BINDING"
	}
	if !approvalAt.Equal(leaseReviewedAt) || now.Before(start) || !now.Before(leaseEnd) || !now.Before(intentEnd) || !intentEnd.After(created) || policyChecked.Before(created) || policyChecked.After(now) ||
		costObserved.After(now) || !now.Before(costEnd) || healthObserved.After(now) || !authorizedAt.Equal(now) || authEnd.After(leaseEnd) || authEnd.After(intentEnd) || authEnd.After(costEnd) ||
		attemptedAt.Before(now) || !attemptedAt.Before(authEnd) || activatedAt.Before(start) || activatedAt.After(now) || windowStarted.Before(start) || preUpdated.After(now) || windowStarted.After(now) || postUpdated.Before(attemptedAt) {
		return "INVALID_TIME_BINDING"
	}
	if now.Unix()-healthObserved.Unix() > int64(c.Lease.MaxHealthSnapshotAgeSeconds) || now.Unix()-healthObserved.Unix() == int64(c.Lease.MaxHealthSnapshotAgeSeconds) && now.Nanosecond() > healthObserved.Nanosecond() {
		return "STALE_HEALTH"
	}
	elapsed := now.Unix() - windowStarted.Unix()
	if now.Nanosecond() < windowStarted.Nanosecond() {
		elapsed--
	}
	inWindow := c.PreLedger.ExecutionsInWindow
	window := windowStarted
	if elapsed >= int64(c.Lease.WindowSeconds) {
		inWindow = 0
		window = now
	}
	if c.Gate.ExecutionsTotalBefore != c.PreLedger.ExecutionsTotal || c.Gate.ExecutionsInWindowBefore != inWindow || c.Gate.CostMinorTotalBefore != c.PreLedger.CostMinorTotal || c.Gate.PendingOutcomesBefore != c.PreLedger.PendingOutcomes {
		return "LEDGER_SNAPSHOT_MISMATCH"
	}
	if containsHistorical(c.PreLedger.SuccessfulIdempotencyKeys, c.Intent.IdempotencyKey) || containsHistorical(c.PreLedger.PendingExecutionIDs, c.Execution.ExecutionID) {
		return "REPLAYED_EXECUTION"
	}
	for _, link := range c.PreLedger.OutcomeLinks {
		if link.ExecutionID == c.Execution.ExecutionID {
			return "REPLAYED_EXECUTION"
		}
	}
	if c.PreLedger.ExecutionsTotal >= c.Lease.MaxExecutionsTotal || inWindow >= c.Lease.MaxExecutionsPerWindow || c.PreLedger.PendingOutcomes >= c.Lease.MaxPendingOutcomes || c.PreLedger.CostMinorTotal > c.Lease.MaxCostMinorTotal || c.Cost.MaxCostMinor > c.Lease.MaxCostMinorTotal-c.PreLedger.CostMinorTotal {
		return "BUDGET_EXCEEDED"
	}
	if c.Profile == "closed_cycle" {
		if c.PostLedger.ExecutionsTotal != c.PreLedger.ExecutionsTotal+1 || c.PostLedger.ExecutionsInWindow != inWindow+1 || c.PostLedger.CostMinorTotal != c.PreLedger.CostMinorTotal+c.Cost.MaxCostMinor || !historicalTimeEqual(c.PostLedger.WindowStartedAt, window) || !historicalTimesEqual(c.PostLedger.LastExecutionAt, c.Execution.AttemptedAt) {
			return "INVALID_LEDGER_TRANSITION"
		}
		status := historicalClosedCycle(c)
		if status != Valid && status != "REVIEW_REQUIRED" {
			return status
		}
		cycle := c.Cycle
		outcome := c.Outcome
		evaluation := c.Evaluation
		closed, _ := historicalTime(cycle.ClosedAt)
		evaluated, _ := historicalTime(evaluation.EvaluatedAt)
		outcomeObserved, _ := historicalTime(outcome.ObservedAt)
		if !reflect.DeepEqual(cycle.ObservationIDs, c.Intent.EvidenceIDs) || !reflect.DeepEqual(evaluation.OutcomeIDs, []string{outcome.OutcomeID}) || outcome.Status != "VALID" || closed.Before(evaluated) || !historicalTimeEqual(c.PostLedger.UpdatedAt, outcomeObserved) {
			return "BROKEN_CYCLE"
		}
		links := append([]ProductionOutcomeLink{}, c.PreLedger.OutcomeLinks...)
		links = append(links, ProductionOutcomeLink{OutcomeID: outcome.OutcomeID, ExecutionID: c.Execution.ExecutionID, ObservedAt: outcome.ObservedAt})
		keys := append([]string{}, c.PreLedger.SuccessfulIdempotencyKeys...)
		keys = append(keys, c.Intent.IdempotencyKey)
		if c.PostLedger.ControlMode != "NORMAL" || c.PostLedger.StopReason != "" || c.PostLedger.ReconciliationRequired || c.PostLedger.ConsecutiveFailures != 0 || c.PostLedger.PendingOutcomes != c.PreLedger.PendingOutcomes || !reflect.DeepEqual(c.PostLedger.PendingExecutionIDs, c.PreLedger.PendingExecutionIDs) || !reflect.DeepEqual(c.PostLedger.OutcomeLinks, links) || !reflect.DeepEqual(c.PostLedger.SuccessfulIdempotencyKeys, keys) || !reflect.DeepEqual(c.PostLedger.ReconciliationResolutionIDs, c.PreLedger.ReconciliationResolutionIDs) || !historicalTimesEqual(c.PostLedger.LastOutcomeAt, outcome.ObservedAt) {
			return "INVALID_LEDGER_TRANSITION"
		}
		return Valid
	}
	if c.Resolution == nil || c.StopLedger == nil {
		return "INVALID_STOP_TRANSITION"
	}
	resolution := c.Resolution
	if resolution.ResolutionID == "" || resolution.LeaseID != c.Lease.LeaseID || resolution.LeaseVersion != c.Lease.LeaseVersion || resolution.LeaseHash != c.Lease.LeaseHash ||
		c.StopLedger.LeaseID != c.Lease.LeaseID || c.StopLedger.LeaseVersion != c.Lease.LeaseVersion || c.StopLedger.LeaseHash != c.Lease.LeaseHash {
		return "BROKEN_LINK"
	}
	expectedStop := c.PreLedger
	expectedStop.ControlMode = "STOPPED"
	expectedStop.StopReason = "RECONCILIATION_REQUIRED"
	expectedStop.ReconciliationRequired = true
	expectedStop.UpdatedAt = c.Execution.AttemptedAt
	if !reflect.DeepEqual(*c.StopLedger, expectedStop) {
		return "INVALID_STOP_TRANSITION"
	}
	expectedPost := *c.StopLedger
	expectedPost.ReconciliationRequired = false
	expectedPost.StopReason = "RECOVERY_REVIEW_REQUIRED"
	expectedPost.UpdatedAt = resolution.ResolvedAt
	expectedPost.ReconciliationResolutionIDs = []string{resolution.ResolutionID}
	if resolution.EffectState == "PERFORMED" {
		expectedPost.ExecutionsTotal++
		expectedPost.ExecutionsInWindow++
		expectedPost.CostMinorTotal += c.Cost.MaxCostMinor
		expectedPost.PendingOutcomes++
		expectedPost.PendingExecutionIDs = append(append([]string{}, c.PreLedger.PendingExecutionIDs...), c.Execution.ExecutionID)
		expectedPost.SuccessfulIdempotencyKeys = append(append([]string{}, c.PreLedger.SuccessfulIdempotencyKeys...), c.Intent.IdempotencyKey)
	}
	resolvedAt, resolvedOK := historicalTime(resolution.ResolvedAt)
	if resolution.ExecutionID != c.Execution.ExecutionID || c.Execution.Status != "RECONCILIATION_REQUIRED" || c.Execution.SideEffectState != "UNKNOWN" || !resolvedOK || resolvedAt.Before(attemptedAt) || !reflect.DeepEqual(c.PostLedger, expectedPost) {
		return "INVALID_STOP_TRANSITION"
	}
	return Valid
}

func historicalTimeEqual(value string, expected time.Time) bool {
	parsed, ok := historicalTime(value)
	return ok && parsed.Equal(expected)
}

func historicalTimesEqual(left, right string) bool {
	a, aOK := historicalTime(left)
	b, bOK := historicalTime(right)
	return aOK && bOK && a.Equal(b)
}
