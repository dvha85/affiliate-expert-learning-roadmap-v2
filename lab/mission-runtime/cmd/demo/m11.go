package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type ProductionLease struct {
	LeaseID                      string   `json:"lease_id"`
	LeaseVersion                 string   `json:"lease_version"`
	PolicyVersion                string   `json:"policy_version"`
	ApprovalRef                  string   `json:"approval_ref"`
	ReviewedBy                   string   `json:"reviewed_by"`
	ReviewerID                   string   `json:"reviewer_id"`
	ReviewedAt                   string   `json:"reviewed_at"`
	PromotionReviewRef           string   `json:"promotion_review_ref"`
	SourceCanaryGrantID          string   `json:"source_canary_grant_id"`
	SourceCanaryGrantVersion     string   `json:"source_canary_grant_version"`
	SourceCanaryGrantHash        string   `json:"source_canary_grant_hash"`
	ValidFrom                    string   `json:"valid_from"`
	ExpiresAt                    string   `json:"expires_at"`
	AllowedRiskClasses           []string `json:"allowed_risk_classes"`
	AllowedActionTypes           []string `json:"allowed_action_types"`
	AllowedHosts                 []string `json:"allowed_hosts"`
	ExecutorIDs                  []string `json:"executor_ids"`
	MaxExecutionsTotal           int      `json:"max_executions_total"`
	MaxExecutionsPerWindow       int      `json:"max_executions_per_window"`
	WindowSeconds                int      `json:"window_seconds"`
	MaxCostMinorTotal            int64    `json:"max_cost_minor_total"`
	Currency                     string   `json:"currency"`
	MaxPendingOutcomes           int      `json:"max_pending_outcomes"`
	MaxConsecutiveFailures       int      `json:"max_consecutive_failures"`
	MaxOutcomeAgeSeconds         int      `json:"max_outcome_age_seconds"`
	MaxHealthSnapshotAgeSeconds  int      `json:"max_health_snapshot_age_seconds"`
	KillSwitchRequired           bool     `json:"kill_switch_required"`
	CorrelationID                string   `json:"correlation_id"`
	HashVersion                  string   `json:"hash_version"`
	LeaseHash                    string   `json:"lease_hash"`
}

type ProductionLeaseApproval struct {
	ApprovalID               string   `json:"approval_id"`
	LeaseID                  string   `json:"lease_id"`
	LeaseVersion             string   `json:"lease_version"`
	LeaseHash                string   `json:"lease_hash"`
	PromotionReviewRef       string   `json:"promotion_review_ref"`
	SourceCanaryGrantID      string   `json:"source_canary_grant_id"`
	SourceCanaryGrantVersion string   `json:"source_canary_grant_version"`
	SourceCanaryGrantHash    string   `json:"source_canary_grant_hash"`
	SourceE5Refs             []string `json:"source_e5_refs"`
	ValidatedRiskClasses     []string `json:"validated_risk_classes"`
	ReviewedBy               string   `json:"reviewed_by"`
	ReviewerID               string   `json:"reviewer_id"`
	ReviewedAt               string   `json:"reviewed_at"`
	Decision                 string   `json:"decision"`
}

type ProductionHealthSnapshot struct {
	SnapshotID                     string   `json:"snapshot_id"`
	LeaseID                        string   `json:"lease_id"`
	LeaseVersion                   string   `json:"lease_version"`
	LeaseHash                      string   `json:"lease_hash"`
	ObservedAt                     string   `json:"observed_at"`
	SourceRefs                     []string `json:"source_refs"`
	DependencyState                string   `json:"dependency_state"`
	TelemetryComplete              bool     `json:"telemetry_complete"`
	ConsecutiveFailures            int      `json:"consecutive_failures"`
	ReconciliationRequired         bool     `json:"reconciliation_required"`
	ComplianceAlertCount           int      `json:"compliance_alert_count"`
	OldestPendingOutcomeAgeSeconds int      `json:"oldest_pending_outcome_age_seconds"`
	HashVersion                    string   `json:"hash_version"`
	SnapshotHash                   string   `json:"snapshot_hash"`
}

type ProductionOutcomeLink struct {
	OutcomeID   string `json:"outcome_id"`
	ExecutionID string `json:"execution_id"`
	ObservedAt  string `json:"observed_at"`
}

type ProductionLedger struct {
	LeaseID                     string                  `json:"lease_id"`
	LeaseVersion                string                  `json:"lease_version"`
	LeaseHash                   string                  `json:"lease_hash"`
	ControlMode                 string                  `json:"control_mode"`
	StopReason                  string                  `json:"stop_reason,omitempty"`
	WindowStartedAt             string                  `json:"window_started_at"`
	ExecutionsTotal             int                     `json:"executions_total"`
	ExecutionsInWindow          int                     `json:"executions_in_window"`
	CostMinorTotal              int64                   `json:"cost_minor_total"`
	PendingOutcomes             int                     `json:"pending_outcomes"`
	PendingExecutionIDs         []string                `json:"pending_execution_ids"`
	SuccessfulIdempotencyKeys   []string                `json:"successful_idempotency_keys"`
	OutcomeLinks                []ProductionOutcomeLink `json:"outcome_links"`
	ConsecutiveFailures         int                     `json:"consecutive_failures"`
	ReconciliationRequired      bool                    `json:"reconciliation_required"`
	ReconciliationResolutionIDs []string                `json:"reconciliation_resolution_ids"`
	LastExecutionAt             string                  `json:"last_execution_at,omitempty"`
	LastOutcomeAt               string                  `json:"last_outcome_at,omitempty"`
	UpdatedAt                   string                  `json:"updated_at"`
}

type ProductionGateDecision struct {
	GateID                   string `json:"gate_id"`
	LeaseID                  string `json:"lease_id"`
	LeaseVersion             string `json:"lease_version"`
	LeaseHash                string `json:"lease_hash"`
	IntentID                 string `json:"intent_id"`
	IntentHash               string `json:"intent_hash"`
	PolicyVersion            string `json:"policy_version"`
	RiskClass                string `json:"risk_class"`
	HealthSnapshotID         string `json:"health_snapshot_id"`
	HealthSnapshotHash       string `json:"health_snapshot_hash"`
	CostBoundID              string `json:"cost_bound_id"`
	CostBoundHash            string `json:"cost_bound_hash"`
	CostBoundMinor           int64  `json:"cost_bound_minor"`
	Decision                 string `json:"decision"`
	Reason                   string `json:"reason"`
	EvaluatedAt              string `json:"evaluated_at"`
	ExecutionsTotalBefore    int    `json:"executions_total_before"`
	ExecutionsInWindowBefore int    `json:"executions_in_window_before"`
	CostMinorTotalBefore     int64  `json:"cost_minor_total_before"`
	PendingOutcomesBefore    int    `json:"pending_outcomes_before"`
	ExecutionAuthorized      bool   `json:"execution_authorized"`
}

type ProductionExecutionAuthorization struct {
	AuthorizationID                 string `json:"authorization_id"`
	IntentID                        string `json:"intent_id"`
	IntentHash                      string `json:"intent_hash"`
	PolicyVersion                   string `json:"policy_version"`
	ProductionLeaseID               string `json:"production_lease_id"`
	ProductionLeaseVersion          string `json:"production_lease_version"`
	ProductionLeaseHash             string `json:"production_lease_hash"`
	ProductionGateID                string `json:"production_gate_id"`
	ProductionHealthSnapshotID      string `json:"production_health_snapshot_id"`
	ProductionHealthSnapshotHash    string `json:"production_health_snapshot_hash"`
	ProductionCostBoundID           string `json:"production_cost_bound_id"`
	ProductionCostBoundHash         string `json:"production_cost_bound_hash"`
	ProductionCostBoundMinor        int64  `json:"production_cost_bound_minor"`
	ExecutorID                      string `json:"executor_id"`
	AuthorizedAt                    string `json:"authorized_at"`
	ExpiresAt                       string `json:"expires_at"`
	IdempotencyKey                  string `json:"idempotency_key"`
	CorrelationID                   string `json:"correlation_id"`
	ExecutionMode                   string `json:"execution_mode"`
	ExecutionAuthorized             bool   `json:"execution_authorized"`
}

type ProductionExecutionRecord struct {
	ExecutionID                     string `json:"execution_id"`
	AuthorizationID                 string `json:"authorization_id"`
	ProductionLeaseID               string `json:"production_lease_id"`
	ProductionLeaseVersion          string `json:"production_lease_version"`
	ProductionLeaseHash             string `json:"production_lease_hash"`
	ProductionGateID                string `json:"production_gate_id"`
	ProductionHealthSnapshotID      string `json:"production_health_snapshot_id"`
	ProductionHealthSnapshotHash    string `json:"production_health_snapshot_hash"`
	ProductionCostBoundID           string `json:"production_cost_bound_id"`
	ProductionCostBoundHash         string `json:"production_cost_bound_hash"`
	ProductionCostBoundMinor        int64  `json:"production_cost_bound_minor"`
	IntentID                        string `json:"intent_id"`
	IntentHash                      string `json:"intent_hash"`
	ExecutorID                      string `json:"executor_id"`
	IdempotencyKey                  string `json:"idempotency_key"`
	AttemptedAt                     string `json:"attempted_at"`
	Status                          string `json:"status"`
	SideEffectState                 string `json:"side_effect_state"`
	ExternalRef                     string `json:"external_ref,omitempty"`
	Error                           string `json:"error,omitempty"`
	CorrelationID                   string `json:"correlation_id"`
}

type ProductionReconciliationResolution struct {
	ResolutionID    string `json:"resolution_id"`
	LeaseID         string `json:"lease_id"`
	LeaseVersion    string `json:"lease_version"`
	LeaseHash       string `json:"lease_hash"`
	ExecutionID     string `json:"execution_id"`
	ResolvedBy      string `json:"resolved_by"`
	ResolverID      string `json:"resolver_id"`
	ResolvedAt      string `json:"resolved_at"`
	EffectState     string `json:"effect_state"`
	Reason          string `json:"reason"`
}

type ProductionCycleRecord struct {
	CycleID                string   `json:"cycle_id"`
	LeaseID                string   `json:"lease_id"`
	LeaseVersion           string   `json:"lease_version"`
	LeaseHash              string   `json:"lease_hash"`
	ObservationIDs         []string `json:"observation_ids"`
	DecisionID             string   `json:"decision_id"`
	IntentID               string   `json:"intent_id"`
	IntentHash             string   `json:"intent_hash"`
	GateID                 string   `json:"gate_id"`
	AuthorizationID        string   `json:"authorization_id"`
	ExecutionID            string   `json:"execution_id"`
	OutcomeID              string   `json:"outcome_id"`
	EvaluationID           string   `json:"evaluation_id"`
	ImprovementProposalID  string   `json:"improvement_proposal_id,omitempty"`
	ReviewID               string   `json:"review_id,omitempty"`
	Status                 string   `json:"status"`
	OpenedAt               string   `json:"opened_at"`
	ClosedAt               string   `json:"closed_at"`
	CorrelationID          string   `json:"correlation_id"`
}

type M11State struct {
	Intent        ShadowActionIntent               `json:"intent"`
	Policy        ShadowPolicyDecision             `json:"policy"`
	Lease         ProductionLease                  `json:"lease"`
	Health        ProductionHealthSnapshot         `json:"health"`
	CostBound     CanaryCostBound                  `json:"cost_bound"`
	Ledger        ProductionLedger                 `json:"ledger"`
	Authorization *ProductionExecutionAuthorization `json:"authorization,omitempty"`
	Execution     *ProductionExecutionRecord       `json:"execution,omitempty"`
}

type M11Context struct {
	Now                     string
	KillSwitch              bool
	RevokedLeaseIDs         []string
	TrustedLeaseApprovals   map[string]ProductionLeaseApproval
	KnownCanaryGrantHashes  map[string]string
	KnownE5Refs             []string
	TrustedHealthSnapshots  map[string]string
	TrustedCostBounds       map[string]string
	Executor                ExecutorProfile
	AllowedExecutorIDs      []string
	PolicyContext           ShadowPolicyContext
}

type productionLeaseHashPayload struct {
	LeaseID                     string   `json:"lease_id"`
	LeaseVersion                string   `json:"lease_version"`
	PolicyVersion               string   `json:"policy_version"`
	ApprovalRef                 string   `json:"approval_ref"`
	ReviewedBy                  string   `json:"reviewed_by"`
	ReviewerID                  string   `json:"reviewer_id"`
	ReviewedAt                  string   `json:"reviewed_at"`
	PromotionReviewRef          string   `json:"promotion_review_ref"`
	SourceCanaryGrantID         string   `json:"source_canary_grant_id"`
	SourceCanaryGrantVersion    string   `json:"source_canary_grant_version"`
	SourceCanaryGrantHash       string   `json:"source_canary_grant_hash"`
	ValidFrom                   string   `json:"valid_from"`
	ExpiresAt                   string   `json:"expires_at"`
	AllowedRiskClasses          []string `json:"allowed_risk_classes"`
	AllowedActionTypes          []string `json:"allowed_action_types"`
	AllowedHosts                []string `json:"allowed_hosts"`
	ExecutorIDs                 []string `json:"executor_ids"`
	MaxExecutionsTotal          int      `json:"max_executions_total"`
	MaxExecutionsPerWindow      int      `json:"max_executions_per_window"`
	WindowSeconds               int      `json:"window_seconds"`
	MaxCostMinorTotal           int64    `json:"max_cost_minor_total"`
	Currency                    string   `json:"currency"`
	MaxPendingOutcomes          int      `json:"max_pending_outcomes"`
	MaxConsecutiveFailures      int      `json:"max_consecutive_failures"`
	MaxOutcomeAgeSeconds        int      `json:"max_outcome_age_seconds"`
	MaxHealthSnapshotAgeSeconds int      `json:"max_health_snapshot_age_seconds"`
	KillSwitchRequired          bool     `json:"kill_switch_required"`
	CorrelationID               string   `json:"correlation_id"`
	HashVersion                 string   `json:"hash_version"`
}

type healthHashPayload struct {
	SnapshotID                     string   `json:"snapshot_id"`
	LeaseID                        string   `json:"lease_id"`
	LeaseVersion                   string   `json:"lease_version"`
	LeaseHash                      string   `json:"lease_hash"`
	ObservedAt                     string   `json:"observed_at"`
	SourceRefs                     []string `json:"source_refs"`
	DependencyState                string   `json:"dependency_state"`
	TelemetryComplete              bool     `json:"telemetry_complete"`
	ConsecutiveFailures            int      `json:"consecutive_failures"`
	ReconciliationRequired         bool     `json:"reconciliation_required"`
	ComplianceAlertCount           int      `json:"compliance_alert_count"`
	OldestPendingOutcomeAgeSeconds int      `json:"oldest_pending_outcome_age_seconds"`
	HashVersion                    string   `json:"hash_version"`
}

type productionActivationRecord struct {
	LeaseID      string `json:"lease_id"`
	LeaseVersion string `json:"lease_version"`
	LeaseHash    string `json:"lease_hash"`
	ActivatedAt  string `json:"activated_at"`
}

func productionCanaryKey(id, version string) string { return id + "\x00" + version }

func ComputeProductionLeaseHash(x ProductionLease) string {
	p := productionLeaseHashPayload{
		LeaseID: x.LeaseID, LeaseVersion: x.LeaseVersion, PolicyVersion: x.PolicyVersion,
		ApprovalRef: x.ApprovalRef, ReviewedBy: x.ReviewedBy, ReviewerID: x.ReviewerID, ReviewedAt: x.ReviewedAt,
		PromotionReviewRef: x.PromotionReviewRef, SourceCanaryGrantID: x.SourceCanaryGrantID,
		SourceCanaryGrantVersion: x.SourceCanaryGrantVersion, SourceCanaryGrantHash: x.SourceCanaryGrantHash,
		ValidFrom: x.ValidFrom, ExpiresAt: x.ExpiresAt,
		AllowedRiskClasses: sortedCopy(x.AllowedRiskClasses), AllowedActionTypes: sortedCopy(x.AllowedActionTypes),
		AllowedHosts: sortedCopy(x.AllowedHosts), ExecutorIDs: sortedCopy(x.ExecutorIDs),
		MaxExecutionsTotal: x.MaxExecutionsTotal, MaxExecutionsPerWindow: x.MaxExecutionsPerWindow,
		WindowSeconds: x.WindowSeconds, MaxCostMinorTotal: x.MaxCostMinorTotal, Currency: x.Currency,
		MaxPendingOutcomes: x.MaxPendingOutcomes, MaxConsecutiveFailures: x.MaxConsecutiveFailures,
		MaxOutcomeAgeSeconds: x.MaxOutcomeAgeSeconds, MaxHealthSnapshotAgeSeconds: x.MaxHealthSnapshotAgeSeconds,
		KillSwitchRequired: x.KillSwitchRequired, CorrelationID: x.CorrelationID, HashVersion: x.HashVersion,
	}
	b, err := json.Marshal(p)
	if err != nil { return "" }
	s := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(s[:])
}

func SealProductionLease(x ProductionLease) ProductionLease {
	if x.HashVersion == "" { x.HashVersion = "go-json-v1" }
	x.Currency = strings.ToUpper(strings.TrimSpace(x.Currency))
	x.LeaseHash = ComputeProductionLeaseHash(x)
	return x
}

func ComputeProductionHealthHash(x ProductionHealthSnapshot) string {
	refs := append([]string(nil), x.SourceRefs...)
	sort.Strings(refs)
	p := healthHashPayload{
		SnapshotID: x.SnapshotID, LeaseID: x.LeaseID, LeaseVersion: x.LeaseVersion, LeaseHash: x.LeaseHash,
		ObservedAt: x.ObservedAt, SourceRefs: refs, DependencyState: x.DependencyState,
		TelemetryComplete: x.TelemetryComplete, ConsecutiveFailures: x.ConsecutiveFailures,
		ReconciliationRequired: x.ReconciliationRequired, ComplianceAlertCount: x.ComplianceAlertCount,
		OldestPendingOutcomeAgeSeconds: x.OldestPendingOutcomeAgeSeconds, HashVersion: x.HashVersion,
	}
	b, err := json.Marshal(p)
	if err != nil { return "" }
	s := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(s[:])
}

func SealProductionHealth(x ProductionHealthSnapshot) ProductionHealthSnapshot {
	if x.HashVersion == "" { x.HashVersion = "go-json-v1" }
	x.SnapshotHash = ComputeProductionHealthHash(x)
	return x
}

func normalizedProductionLedger(l ProductionLedger, lease ProductionLease, now time.Time) (ProductionLedger, bool) {
	if l.LeaseID != lease.LeaseID || l.LeaseVersion != lease.LeaseVersion || l.LeaseHash != lease.LeaseHash ||
		l.ExecutionsTotal < 0 || l.ExecutionsInWindow < 0 || l.CostMinorTotal < 0 || l.PendingOutcomes < 0 ||
		l.PendingOutcomes != len(l.PendingExecutionIDs) || l.ConsecutiveFailures < 0 {
		return l, false
	}
	if l.ControlMode != "NORMAL" && l.ControlMode != "STOPPED" { return l, false }
	started, err := time.Parse(time.RFC3339, l.WindowStartedAt)
	if err != nil || now.Before(started) { return l, false }
	if now.Sub(started) >= time.Duration(lease.WindowSeconds)*time.Second {
		l.WindowStartedAt = now.Format(time.RFC3339)
		l.ExecutionsInWindow = 0
	}
	return l, true
}

func productionGate(state M11State, ctx M11Context) ProductionGateDecision {
	return ProductionGateDecision{
		GateID: "prod-gate-" + state.Lease.LeaseID + "-" + state.Intent.IntentID,
		LeaseID: state.Lease.LeaseID, LeaseVersion: state.Lease.LeaseVersion, LeaseHash: state.Lease.LeaseHash,
		IntentID: state.Intent.IntentID, IntentHash: state.Intent.IntentHash, PolicyVersion: state.Policy.PolicyVersion,
		RiskClass: state.Policy.RiskClass, HealthSnapshotID: state.Health.SnapshotID,
		HealthSnapshotHash: state.Health.SnapshotHash, CostBoundID: state.CostBound.CostBoundID,
		CostBoundHash: state.CostBound.CostBoundHash, CostBoundMinor: state.CostBound.MaxCostMinor,
		Decision: "DENY", Reason: "INVALID_PRODUCTION_STATE", EvaluatedAt: ctx.Now,
		ExecutionsTotalBefore: state.Ledger.ExecutionsTotal, ExecutionsInWindowBefore: state.Ledger.ExecutionsInWindow,
		CostMinorTotalBefore: state.Ledger.CostMinorTotal, PendingOutcomesBefore: state.Ledger.PendingOutcomes,
		ExecutionAuthorized: false,
	}
}

func productionRequireApproval(g *ProductionGateDecision, reason string) ProductionGateDecision {
	g.Decision = "REQUIRE_APPROVAL"; g.Reason = reason; return *g
}
func productionStop(g *ProductionGateDecision, reason string) ProductionGateDecision {
	g.Decision = "STOP"; g.Reason = reason; return *g
}
func productionDegrade(g *ProductionGateDecision, reason string) ProductionGateDecision {
	g.Decision = "DEGRADE"; g.Reason = reason; return *g
}

func validateProductionLeaseApproval(l ProductionLease, ctx M11Context) ([]string, string) {
	approval, ok := ctx.TrustedLeaseApprovals[l.ApprovalRef]
	if !ok { return nil, "LEASE_PROVENANCE_MISSING" }
	if approval.ApprovalID != l.ApprovalRef || approval.LeaseID != l.LeaseID || approval.LeaseVersion != l.LeaseVersion ||
		approval.LeaseHash != l.LeaseHash || approval.PromotionReviewRef != l.PromotionReviewRef ||
		approval.SourceCanaryGrantID != l.SourceCanaryGrantID || approval.SourceCanaryGrantVersion != l.SourceCanaryGrantVersion ||
		approval.SourceCanaryGrantHash != l.SourceCanaryGrantHash || approval.ReviewedBy != "human" ||
		approval.ReviewerID != l.ReviewerID || approval.ReviewedAt != l.ReviewedAt || approval.Decision != "APPROVE_PRODUCTION_LEASE" ||
		len(approval.SourceE5Refs) == 0 || len(approval.ValidatedRiskClasses) == 0 {
		return nil, "LEASE_APPROVAL_MISMATCH"
	}
	knownHash, ok := ctx.KnownCanaryGrantHashes[productionCanaryKey(l.SourceCanaryGrantID, l.SourceCanaryGrantVersion)]
	if !ok || knownHash != l.SourceCanaryGrantHash { return nil, "PROMOTION_SOURCE_MISMATCH" }
	for _, ref := range approval.SourceE5Refs {
		if !containsFold(ctx.KnownE5Refs, ref) { return nil, "PROMOTION_E5_EVIDENCE_MISSING" }
	}
	for _, risk := range approval.ValidatedRiskClasses {
		if risk != "RISK0" && risk != "RISK1" { return nil, "LEASE_APPROVAL_MISMATCH" }
	}
	for _, risk := range l.AllowedRiskClasses {
		if !containsFold(approval.ValidatedRiskClasses, risk) { return nil, "PROMOTION_RISK_NOT_VALIDATED" }
	}
	return approval.ValidatedRiskClasses, ""
}

func EvaluateProductionGate(state M11State, ctx M11Context) ProductionGateDecision {
	g := productionGate(state, ctx)
	i, p, l, h, c := state.Intent, state.Policy, state.Lease, state.Health, state.CostBound
	if i.IntentHash == "" || i.IntentHash != ComputeShadowIntentHash(i) || !i.ShadowOnly || !i.DryRun {
		g.Reason = "INVALID_INTENT_BINDING"; return g
	}
	if l.HashVersion != "go-json-v1" || l.LeaseHash == "" || l.LeaseHash != ComputeProductionLeaseHash(l) {
		g.Reason = "TAMPERED_LEASE"; return g
	}
	if l.ReviewedBy != "human" || strings.TrimSpace(l.ReviewerID) == "" || strings.TrimSpace(l.ApprovalRef) == "" ||
		strings.TrimSpace(l.PromotionReviewRef) == "" || strings.TrimSpace(l.SourceCanaryGrantID) == "" ||
		strings.TrimSpace(l.SourceCanaryGrantVersion) == "" || strings.TrimSpace(l.SourceCanaryGrantHash) == "" ||
		!l.KillSwitchRequired || l.MaxExecutionsTotal < 1 || l.MaxExecutionsPerWindow < 1 ||
		l.MaxExecutionsPerWindow > l.MaxExecutionsTotal || l.WindowSeconds < 60 || l.MaxCostMinorTotal < 0 ||
		l.MaxPendingOutcomes < 1 || l.MaxConsecutiveFailures < 1 || l.MaxOutcomeAgeSeconds < 60 ||
		l.MaxHealthSnapshotAgeSeconds < 60 || len(l.AllowedRiskClasses) == 0 || len(l.AllowedActionTypes) == 0 ||
		len(l.AllowedHosts) == 0 || len(l.ExecutorIDs) == 0 || hasWildcard(l.AllowedActionTypes) ||
		hasWildcard(l.AllowedHosts) || hasWildcard(l.ExecutorIDs) || len(l.Currency) != 3 || l.Currency != strings.ToUpper(l.Currency) {
		g.Reason = "INVALID_LEASE"; return g
	}
	for _, risk := range l.AllowedRiskClasses {
		if risk != "RISK0" && risk != "RISK1" { g.Reason = "INVALID_LEASE"; return g }
	}
	if ctx.TrustedLeaseApprovals == nil || ctx.KnownCanaryGrantHashes == nil || ctx.TrustedHealthSnapshots == nil || ctx.TrustedCostBounds == nil {
		g.Reason = "GOVERNANCE_INPUTS_UNAVAILABLE"; return g
	}
	validatedRisks, approvalReason := validateProductionLeaseApproval(l, ctx)
	if approvalReason != "" { g.Reason = approvalReason; return g }
	if containsFold(ctx.RevokedLeaseIDs, l.LeaseID) { return productionStop(&g, "LEASE_REVOKED") }
	if ctx.KillSwitch { return productionStop(&g, "KILL_SWITCH_ACTIVE") }

	now, errNow := time.Parse(time.RFC3339, ctx.Now)
	reviewedAt, errReviewed := time.Parse(time.RFC3339, l.ReviewedAt)
	validFrom, errValid := time.Parse(time.RFC3339, l.ValidFrom)
	leaseExpires, errLease := time.Parse(time.RFC3339, l.ExpiresAt)
	if errNow != nil || errReviewed != nil || errValid != nil || errLease != nil || reviewedAt.After(validFrom) || !leaseExpires.After(validFrom) {
		g.Reason = "INVALID_LEASE_TIME"; return g
	}
	if now.Before(validFrom) { g.Decision = "WAIT"; g.Reason = "LEASE_NOT_YET_ACTIVE"; return g }
	if !leaseExpires.After(now) { return productionRequireApproval(&g, "LEASE_EXPIRED") }

	ledger, ok := normalizedProductionLedger(state.Ledger, l, now)
	if !ok { g.Reason = "LEDGER_MISMATCH"; return g }
	g.ExecutionsTotalBefore = ledger.ExecutionsTotal
	g.ExecutionsInWindowBefore = ledger.ExecutionsInWindow
	g.CostMinorTotalBefore = ledger.CostMinorTotal
	g.PendingOutcomesBefore = ledger.PendingOutcomes
	if ledger.ControlMode == "STOPPED" { return productionStop(&g, "STICKY_STOP") }
	if ledger.ReconciliationRequired { return productionStop(&g, "RECONCILIATION_REQUIRED") }
	if ledger.ConsecutiveFailures >= l.MaxConsecutiveFailures { return productionStop(&g, "FAILURE_THRESHOLD") }

	policyChecked, errPolicy := time.Parse(time.RFC3339, p.PolicyCheckedAt)
	if errPolicy != nil || policyChecked.After(now) || p.PolicyVersion == "" || p.IntentID != i.IntentID || p.IntentHash != i.IntentHash || !p.ShadowOnly || p.ExecutionAuthorized {
		g.Reason = "INVALID_POLICY_STATE"; return g
	}
	freshCtx := ctx.PolicyContext
	freshCtx.Now = ctx.Now
	fresh := EvaluateShadowPolicy(i, freshCtx)
	if fresh.Reason == "POLICY_UNAVAILABLE" || !policyEquivalentForExecution(p, fresh) {
		g.Reason = "POLICY_REVALIDATION_FAILED"; return g
	}
	g.RiskClass = fresh.RiskClass
	g.PolicyVersion = fresh.PolicyVersion
	if l.PolicyVersion != fresh.PolicyVersion { g.Reason = "LEASE_POLICY_MISMATCH"; return g }
	if fresh.RiskClass == "RISK2" { return productionRequireApproval(&g, "RISK2_PER_ACTION_APPROVAL_REQUIRED") }
	if fresh.RiskClass != "RISK0" && fresh.RiskClass != "RISK1" { g.Reason = "UNKNOWN_RISK_CLASS"; return g }
	if fresh.RiskClass == "RISK0" && fresh.Decision != "ALLOW" { g.Reason = "POLICY_STATE_NOT_PRODUCTION_ELIGIBLE"; return g }
	if fresh.RiskClass == "RISK1" && (fresh.Decision != "HUMAN_REVIEW" || fresh.Reason != "RISK1_REQUIRES_REVIEW") {
		g.Reason = "POLICY_STATE_NOT_PRODUCTION_ELIGIBLE"; return g
	}
	if !containsFold(validatedRisks, fresh.RiskClass) { g.Reason = "PROMOTION_RISK_NOT_VALIDATED"; return g }
	if !containsFold(l.AllowedRiskClasses, fresh.RiskClass) { return productionRequireApproval(&g, "RISK_NOT_DELEGATED") }
	if !containsFold(l.AllowedActionTypes, i.ActionType) || !allowedHost(i.Target, l.AllowedHosts) {
		return productionRequireApproval(&g, "SCOPE_NOT_DELEGATED")
	}
	if !containsFold(l.ExecutorIDs, ctx.Executor.ExecutorID) || !containsFold(ctx.AllowedExecutorIDs, ctx.Executor.ExecutorID) || !executorAllows(ctx.Executor, i) {
		g.Reason = "EXECUTOR_NOT_ALLOWED"; return g
	}

	if h.HashVersion != "go-json-v1" || h.SnapshotHash == "" || h.SnapshotHash != ComputeProductionHealthHash(h) {
		g.Reason = "TAMPERED_HEALTH"; return g
	}
	if h.LeaseID != l.LeaseID || h.LeaseVersion != l.LeaseVersion || h.LeaseHash != l.LeaseHash || len(h.SourceRefs) == 0 {
		g.Reason = "HEALTH_MISMATCH"; return g
	}
	trustedHealthHash, ok := ctx.TrustedHealthSnapshots[h.SnapshotID]
	if !ok || trustedHealthHash != h.SnapshotHash { g.Reason = "HEALTH_UNTRUSTED"; return g }
	healthObservedAt, errHealth := time.Parse(time.RFC3339, h.ObservedAt)
	if errHealth != nil || healthObservedAt.After(now) { g.Reason = "HEALTH_TIME_INVALID"; return g }
	// Severe known signals must outrank lower-severity DEGRADE signals.
	if h.ComplianceAlertCount > 0 { return productionStop(&g, "COMPLIANCE_ALERT") }
	if h.ReconciliationRequired { return productionStop(&g, "RECONCILIATION_REQUIRED") }
	if h.ConsecutiveFailures >= l.MaxConsecutiveFailures { return productionStop(&g, "FAILURE_THRESHOLD") }
	if h.OldestPendingOutcomeAgeSeconds > l.MaxOutcomeAgeSeconds { return productionStop(&g, "OUTCOME_STALE") }
	if now.Sub(healthObservedAt) > time.Duration(l.MaxHealthSnapshotAgeSeconds)*time.Second { return productionDegrade(&g, "HEALTH_STALE") }
	if !h.TelemetryComplete { return productionDegrade(&g, "TELEMETRY_INCOMPLETE") }
	if h.DependencyState != "HEALTHY" { return productionDegrade(&g, "DEPENDENCY_DEGRADED") }

	if strings.TrimSpace(c.CostBoundID) == "" { g.Reason = "COST_BOUND_MISSING"; return g }
	if c.HashVersion != "go-json-v1" || c.CostBoundHash == "" || c.CostBoundHash != ComputeCanaryCostBoundHash(c) {
		g.Reason = "TAMPERED_COST_BOUND"; return g
	}
	if c.IntentID != i.IntentID || c.IntentHash != i.IntentHash || c.CorrelationID != i.CorrelationID ||
		c.MaxCostMinor < 0 || c.Currency != l.Currency || strings.TrimSpace(c.SourceRef) == "" {
		g.Reason = "COST_BOUND_MISMATCH"; return g
	}
	trustedCostHash, ok := ctx.TrustedCostBounds[c.CostBoundID]
	if !ok || trustedCostHash != c.CostBoundHash { g.Reason = "COST_BOUND_UNTRUSTED"; return g }
	costObservedAt, errCostObserved := time.Parse(time.RFC3339, c.ObservedAt)
	costExpires, errCostExpires := time.Parse(time.RFC3339, c.ExpiresAt)
	if errCostObserved != nil || errCostExpires != nil || costObservedAt.After(now) || !costExpires.After(costObservedAt) || !costExpires.After(now) {
		g.Reason = "COST_BOUND_EXPIRED"; return g
	}

	if containsFold(ledger.SuccessfulIdempotencyKeys, i.IdempotencyKey) { g.Decision = "WAIT"; g.Reason = "DUPLICATE_EXECUTION"; return g }
	if ledger.ExecutionsTotal >= l.MaxExecutionsTotal { return productionRequireApproval(&g, "PRODUCTION_TOTAL_BUDGET_EXHAUSTED") }
	if ledger.ExecutionsInWindow >= l.MaxExecutionsPerWindow { g.Decision = "WAIT"; g.Reason = "RATE_LIMIT_REACHED"; return g }
	if ledger.CostMinorTotal > l.MaxCostMinorTotal || c.MaxCostMinor > l.MaxCostMinorTotal-ledger.CostMinorTotal {
		return productionRequireApproval(&g, "PRODUCTION_COST_BUDGET_EXHAUSTED")
	}
	if ledger.PendingOutcomes >= l.MaxPendingOutcomes { g.Decision = "WAIT"; g.Reason = "OUTCOME_BACKPRESSURE"; return g }
	g.CostBoundID = c.CostBoundID
	g.CostBoundHash = c.CostBoundHash
	g.CostBoundMinor = c.MaxCostMinor
	g.Decision = "ALLOW_PRODUCTION"
	g.Reason = "PRODUCTION_ELIGIBLE"
	return g
}

func AuthorizeProduction(state M11State, ctx M11Context) (ProductionExecutionAuthorization, ProductionGateDecision, string) {
	g := EvaluateProductionGate(state, ctx)
	if g.Decision != "ALLOW_PRODUCTION" { return ProductionExecutionAuthorization{}, g, g.Decision }
	now, errNow := time.Parse(time.RFC3339, ctx.Now)
	intentExpires, errIntent := time.Parse(time.RFC3339, state.Intent.ExpiresAt)
	leaseExpires, errLease := time.Parse(time.RFC3339, state.Lease.ExpiresAt)
	costExpires, errCost := time.Parse(time.RFC3339, state.CostBound.ExpiresAt)
	healthObservedAt, errHealth := time.Parse(time.RFC3339, state.Health.ObservedAt)
	if errNow != nil || errIntent != nil || errLease != nil || errCost != nil || errHealth != nil {
		g.Decision = "DENY"; g.Reason = "AUTHORIZATION_TIME_INVALID"; return ProductionExecutionAuthorization{}, g, g.Decision
	}
	healthExpires := healthObservedAt.Add(time.Duration(state.Lease.MaxHealthSnapshotAgeSeconds) * time.Second)
	expires := minTime(minTime(intentExpires, leaseExpires), minTime(costExpires, healthExpires))
	if !expires.After(now) {
		g.Decision = "DEGRADE"; g.Reason = "HEALTH_STALE"; return ProductionExecutionAuthorization{}, g, g.Decision
	}
	a := ProductionExecutionAuthorization{
		AuthorizationID: "prod-auth-" + state.Lease.LeaseID + "-" + state.Intent.IntentID,
		IntentID: state.Intent.IntentID, IntentHash: state.Intent.IntentHash, PolicyVersion: g.PolicyVersion,
		ProductionLeaseID: state.Lease.LeaseID, ProductionLeaseVersion: state.Lease.LeaseVersion,
		ProductionLeaseHash: state.Lease.LeaseHash, ProductionGateID: g.GateID,
		ProductionHealthSnapshotID: state.Health.SnapshotID, ProductionHealthSnapshotHash: state.Health.SnapshotHash,
		ProductionCostBoundID: state.CostBound.CostBoundID, ProductionCostBoundHash: state.CostBound.CostBoundHash,
		ProductionCostBoundMinor: state.CostBound.MaxCostMinor, ExecutorID: ctx.Executor.ExecutorID,
		AuthorizedAt: ctx.Now, ExpiresAt: expires.Format(time.RFC3339), IdempotencyKey: state.Intent.IdempotencyKey,
		CorrelationID: state.Intent.CorrelationID, ExecutionMode: "GOVERNED_PRODUCTION", ExecutionAuthorized: true,
	}
	return a, g, "AUTHORIZED"
}

func productionLedgerPath(dir, id, version string) string {
	s := sha256.Sum256([]byte("production\x00" + id + "\x00" + version))
	return filepath.Join(dir, "production-ledger-"+hex.EncodeToString(s[:])+".json")
}
func productionLockPath(dir, id, version string) string {
	s := sha256.Sum256([]byte("production-lock\x00" + id + "\x00" + version))
	return filepath.Join(dir, "production-lock-"+hex.EncodeToString(s[:]))
}
func productionActivationPath(dir, id, version string) string {
	s := sha256.Sum256([]byte("production-activation\x00" + id + "\x00" + version))
	return filepath.Join(dir, "production-activation-"+hex.EncodeToString(s[:])+".json")
}
func persistProductionLedger(path string, l ProductionLedger) error {
	b, err := json.MarshalIndent(l, "", "  ")
	if err != nil { return err }
	tmp := path + ".tmp"
	if err = os.WriteFile(tmp, b, 0600); err != nil { return err }
	return os.Rename(tmp, path)
}
func loadProductionLedger(path string) (ProductionLedger, error) {
	var l ProductionLedger
	b, err := os.ReadFile(path)
	if err != nil { return l, err }
	err = json.Unmarshal(b, &l)
	return l, err
}
func ensureInitialProductionLedger(path string, l ProductionLedger) error {
	b, err := json.MarshalIndent(l, "", "  ")
	if err != nil { return err }
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil { return err }
	if _, err = f.Write(b); err != nil { _ = f.Close(); return err }
	if err = f.Sync(); err != nil { _ = f.Close(); return err }
	return f.Close()
}
func writeActivationRecord(path string, r productionActivationRecord) error {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil { return err }
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil { return err }
	if _, err = f.Write(b); err != nil { _ = f.Close(); return err }
	if err = f.Sync(); err != nil { _ = f.Close(); return err }
	return f.Close()
}
func productionActivationMatches(path string, lease ProductionLease) bool {
	var r productionActivationRecord
	b, err := os.ReadFile(path)
	if err != nil || json.Unmarshal(b, &r) != nil { return false }
	return r.LeaseID == lease.LeaseID && r.LeaseVersion == lease.LeaseVersion && r.LeaseHash == lease.LeaseHash
}

// InitializeProductionLedger is an explicit activation step. The executor never creates a missing production ledger.
func InitializeProductionLedger(state *M11State, dir, activatedAt string) string {
	if state == nil || state.Lease.LeaseHash == "" || state.Lease.LeaseHash != ComputeProductionLeaseHash(state.Lease) {
		return "DENY_ACTIVATION"
	}
	if state.Ledger.LeaseID != state.Lease.LeaseID || state.Ledger.LeaseVersion != state.Lease.LeaseVersion ||
		state.Ledger.LeaseHash != state.Lease.LeaseHash || state.Ledger.ControlMode != "NORMAL" ||
		state.Ledger.ExecutionsTotal != 0 || state.Ledger.ExecutionsInWindow != 0 || state.Ledger.CostMinorTotal != 0 ||
		state.Ledger.PendingOutcomes != 0 || len(state.Ledger.PendingExecutionIDs) != 0 ||
		len(state.Ledger.SuccessfulIdempotencyKeys) != 0 || len(state.Ledger.OutcomeLinks) != 0 ||
		state.Ledger.ReconciliationRequired || len(state.Ledger.ReconciliationResolutionIDs) != 0 {
		return "DENY_ACTIVATION"
	}
	if _, err := time.Parse(time.RFC3339, activatedAt); err != nil { return "DENY_ACTIVATION" }
	if err := os.MkdirAll(dir, 0755); err != nil { return "ACTIVATION_FAILED" }
	lockPath := productionLockPath(dir, state.Lease.LeaseID, state.Lease.LeaseVersion)
	lock, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil { return "WAIT_PRODUCTION_LOCK" }
	_ = lock.Close(); defer os.Remove(lockPath)
	activationPath := productionActivationPath(dir, state.Lease.LeaseID, state.Lease.LeaseVersion)
	if _, err = os.Stat(activationPath); err == nil { return "ALREADY_INITIALIZED" }
	if !os.IsNotExist(err) { return "ACTIVATION_FAILED" }
	ledgerPath := productionLedgerPath(dir, state.Lease.LeaseID, state.Lease.LeaseVersion)
	if _, err = os.Stat(ledgerPath); err == nil { return "DENY_ACTIVATION" }
	if !os.IsNotExist(err) { return "ACTIVATION_FAILED" }
	activation := productionActivationRecord{LeaseID: state.Lease.LeaseID, LeaseVersion: state.Lease.LeaseVersion, LeaseHash: state.Lease.LeaseHash, ActivatedAt: activatedAt}
	if err = writeActivationRecord(activationPath, activation); err != nil { return "ACTIVATION_FAILED" }
	if err = ensureInitialProductionLedger(ledgerPath, state.Ledger); err != nil {
		_ = os.Remove(activationPath)
		return "ACTIVATION_FAILED"
	}
	return "INITIALIZED"
}

func persistStickyStop(path string, state *M11State, reason, now string) error {
	state.Ledger.ControlMode = "STOPPED"
	state.Ledger.StopReason = reason
	state.Ledger.UpdatedAt = now
	return persistProductionLedger(path, state.Ledger)
}

func productionExecutionRecord(state M11State, a ProductionExecutionAuthorization, now, status, effect, ref, message string) ProductionExecutionRecord {
	return ProductionExecutionRecord{
		ExecutionID: "exec-" + a.AuthorizationID, AuthorizationID: a.AuthorizationID,
		ProductionLeaseID: a.ProductionLeaseID, ProductionLeaseVersion: a.ProductionLeaseVersion,
		ProductionLeaseHash: a.ProductionLeaseHash, ProductionGateID: a.ProductionGateID,
		ProductionHealthSnapshotID: a.ProductionHealthSnapshotID, ProductionHealthSnapshotHash: a.ProductionHealthSnapshotHash,
		ProductionCostBoundID: a.ProductionCostBoundID, ProductionCostBoundHash: a.ProductionCostBoundHash,
		ProductionCostBoundMinor: a.ProductionCostBoundMinor, IntentID: a.IntentID, IntentHash: a.IntentHash,
		ExecutorID: a.ExecutorID, IdempotencyKey: a.IdempotencyKey, AttemptedAt: now, Status: status,
		SideEffectState: effect, ExternalRef: ref, Error: message, CorrelationID: state.Intent.CorrelationID,
	}
}

func ExecuteProductionLocalSandbox(state *M11State, a ProductionExecutionAuthorization, ctx M11Context, dir string) (ProductionExecutionRecord, string) {
	if state == nil || state.Authorization == nil || *state.Authorization != a || !a.ExecutionAuthorized || a.ExecutionMode != "GOVERNED_PRODUCTION" {
		return ProductionExecutionRecord{}, "DENY_AUTHORIZATION"
	}
	if err := os.MkdirAll(dir, 0755); err != nil { return ProductionExecutionRecord{}, "EXECUTION_FAILED" }
	lockPath := productionLockPath(dir, state.Lease.LeaseID, state.Lease.LeaseVersion)
	lock, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		if os.IsExist(err) { return ProductionExecutionRecord{}, "WAIT_PRODUCTION_LOCK" }
		return ProductionExecutionRecord{}, "EXECUTION_FAILED"
	}
	_ = lock.Close(); defer os.Remove(lockPath)
	activationPath := productionActivationPath(dir, state.Lease.LeaseID, state.Lease.LeaseVersion)
	if !productionActivationMatches(activationPath, state.Lease) { return ProductionExecutionRecord{}, "STOP_ACTIVATION_STATE_MISSING" }
	ledgerPath := productionLedgerPath(dir, state.Lease.LeaseID, state.Lease.LeaseVersion)
	ledger, err := loadProductionLedger(ledgerPath)
	if os.IsNotExist(err) { return ProductionExecutionRecord{}, "STOP_LEDGER_MISSING" }
	if err != nil { return ProductionExecutionRecord{}, "STOP_LEDGER_UNREADABLE" }
	state.Ledger = ledger

	expected, gate, status := AuthorizeProduction(*state, ctx)
	if status != "AUTHORIZED" {
		if gate.Decision == "STOP" {
			if err := persistStickyStop(ledgerPath, state, gate.Reason, ctx.Now); err != nil { return ProductionExecutionRecord{}, "STOP_PERSISTENCE_FAILED" }
		}
		return ProductionExecutionRecord{}, status
	}
	if expected != a { return ProductionExecutionRecord{}, "DENY_AUTHORIZATION" }
	now, errNow := time.Parse(time.RFC3339, ctx.Now)
	authorizedAt, errAuthorized := time.Parse(time.RFC3339, a.AuthorizedAt)
	expiresAt, errExpires := time.Parse(time.RFC3339, a.ExpiresAt)
	if errNow != nil || errAuthorized != nil || errExpires != nil || authorizedAt.After(now) || !expiresAt.After(now) {
		return ProductionExecutionRecord{}, "DENY_EXPIRED_AUTHORIZATION"
	}

	marker := sandboxIdempotencyPath(dir, a.IdempotencyKey)
	if _, err = os.Stat(marker); err == nil {
		if containsFold(state.Ledger.SuccessfulIdempotencyKeys, a.IdempotencyKey) { return ProductionExecutionRecord{}, "WAIT_ALREADY_EXECUTED" }
		state.Ledger.ReconciliationRequired = true
		if err := persistStickyStop(ledgerPath, state, "RECONCILIATION_REQUIRED", ctx.Now); err != nil { return ProductionExecutionRecord{}, "STOP_PERSISTENCE_FAILED" }
		rec := productionExecutionRecord(*state, a, ctx.Now, "RECONCILIATION_REQUIRED", "UNKNOWN", marker, "idempotency marker exists without durable successful ledger entry")
		state.Execution = &rec
		return rec, "STOP_RECONCILIATION"
	} else if !os.IsNotExist(err) {
		return ProductionExecutionRecord{}, "STOP_LEDGER_UNREADABLE"
	}

	payload := map[string]any{
		"intent_id": state.Intent.IntentID, "intent_hash": state.Intent.IntentHash, "action_type": state.Intent.ActionType,
		"target": state.Intent.Target, "parameters": state.Intent.Parameters, "production_lease_id": a.ProductionLeaseID,
		"production_lease_hash": a.ProductionLeaseHash, "production_gate_id": a.ProductionGateID,
		"health_snapshot_hash": a.ProductionHealthSnapshotHash, "cost_bound_hash": a.ProductionCostBoundHash,
		"cost_bound_minor": a.ProductionCostBoundMinor, "execution_mode": a.ExecutionMode, "idempotency_key": a.IdempotencyKey,
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil { return ProductionExecutionRecord{}, "EXECUTION_FAILED" }
	f, err := os.OpenFile(marker, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil { return ProductionExecutionRecord{}, "STOP_RECONCILIATION" }
	uncertain := func(message string) (ProductionExecutionRecord, string) {
		_ = f.Close()
		state.Ledger.ReconciliationRequired = true
		if err := persistStickyStop(ledgerPath, state, "RECONCILIATION_REQUIRED", ctx.Now); err != nil { return ProductionExecutionRecord{}, "STOP_PERSISTENCE_FAILED" }
		rec := productionExecutionRecord(*state, a, ctx.Now, "RECONCILIATION_REQUIRED", "UNKNOWN", marker, message)
		state.Execution = &rec
		return rec, "STOP_RECONCILIATION"
	}
	if _, err = f.Write(b); err != nil { return uncertain(err.Error()) }
	if err = f.Sync(); err != nil { return uncertain(err.Error()) }
	if err = f.Close(); err != nil {
		state.Ledger.ReconciliationRequired = true
		if pErr := persistStickyStop(ledgerPath, state, "RECONCILIATION_REQUIRED", ctx.Now); pErr != nil { return ProductionExecutionRecord{}, "STOP_PERSISTENCE_FAILED" }
		rec := productionExecutionRecord(*state, a, ctx.Now, "RECONCILIATION_REQUIRED", "UNKNOWN", marker, err.Error())
		state.Execution = &rec
		return rec, "STOP_RECONCILIATION"
	}
	normalized, ok := normalizedProductionLedger(state.Ledger, state.Lease, now)
	if !ok { return uncertain("ledger invalid before durable update") }
	normalized.ExecutionsTotal++
	normalized.ExecutionsInWindow++
	normalized.CostMinorTotal += a.ProductionCostBoundMinor
	normalized.PendingOutcomes++
	normalized.PendingExecutionIDs = append(normalized.PendingExecutionIDs, "exec-"+a.AuthorizationID)
	normalized.SuccessfulIdempotencyKeys = append(normalized.SuccessfulIdempotencyKeys, a.IdempotencyKey)
	normalized.ConsecutiveFailures = 0
	normalized.LastExecutionAt = ctx.Now
	normalized.UpdatedAt = ctx.Now
	if err = persistProductionLedger(ledgerPath, normalized); err != nil {
		state.Ledger.ReconciliationRequired = true
		_ = persistStickyStop(ledgerPath, state, "RECONCILIATION_REQUIRED", ctx.Now)
		rec := productionExecutionRecord(*state, a, ctx.Now, "RECONCILIATION_REQUIRED", "UNKNOWN", marker, "side effect written but ledger persistence failed")
		state.Execution = &rec
		return rec, "STOP_RECONCILIATION"
	}
	state.Ledger = normalized
	rec := productionExecutionRecord(*state, a, ctx.Now, "SUCCEEDED", "PERFORMED", marker, "")
	state.Execution = &rec
	return rec, "EXECUTED"
}

// ResolveProductionReconciliation can establish what happened while preserving STOPPED authority.
func ResolveProductionReconciliation(state *M11State, dir string, r ProductionReconciliationResolution) string {
	if state == nil || state.Execution == nil || strings.TrimSpace(r.ResolutionID) == "" || r.ResolvedBy != "human" ||
		strings.TrimSpace(r.ResolverID) == "" || strings.TrimSpace(r.Reason) == "" ||
		r.LeaseID != state.Lease.LeaseID || r.LeaseVersion != state.Lease.LeaseVersion || r.LeaseHash != state.Lease.LeaseHash ||
		r.ExecutionID != state.Execution.ExecutionID || (r.EffectState != "PERFORMED" && r.EffectState != "NOT_PERFORMED") {
		return "DENY_RECONCILIATION_RESOLUTION"
	}
	if _, err := time.Parse(time.RFC3339, r.ResolvedAt); err != nil { return "DENY_RECONCILIATION_RESOLUTION" }
	lockPath := productionLockPath(dir, state.Lease.LeaseID, state.Lease.LeaseVersion)
	lock, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil { return "WAIT_PRODUCTION_LOCK" }
	_ = lock.Close(); defer os.Remove(lockPath)
	ledgerPath := productionLedgerPath(dir, state.Lease.LeaseID, state.Lease.LeaseVersion)
	ledger, err := loadProductionLedger(ledgerPath)
	if err != nil || ledger.ControlMode != "STOPPED" || !ledger.ReconciliationRequired { return "DENY_RECONCILIATION_RESOLUTION" }
	if containsFold(ledger.ReconciliationResolutionIDs, r.ResolutionID) { return "DUPLICATE_RECONCILIATION_RESOLUTION" }
	if r.EffectState == "PERFORMED" && !containsFold(ledger.SuccessfulIdempotencyKeys, state.Execution.IdempotencyKey) {
		if state.Execution.ProductionCostBoundMinor < 0 || ledger.CostMinorTotal > int64(^uint64(0)>>1)-state.Execution.ProductionCostBoundMinor {
			return "DENY_RECONCILIATION_RESOLUTION"
		}
		ledger.ExecutionsTotal++
		ledger.ExecutionsInWindow++
		ledger.CostMinorTotal += state.Execution.ProductionCostBoundMinor
		ledger.PendingOutcomes++
		ledger.PendingExecutionIDs = append(ledger.PendingExecutionIDs, state.Execution.ExecutionID)
		ledger.SuccessfulIdempotencyKeys = append(ledger.SuccessfulIdempotencyKeys, state.Execution.IdempotencyKey)
	}
	ledger.ReconciliationRequired = false
	ledger.ControlMode = "STOPPED"
	ledger.StopReason = "RECOVERY_REVIEW_REQUIRED"
	ledger.ReconciliationResolutionIDs = append(ledger.ReconciliationResolutionIDs, r.ResolutionID)
	ledger.UpdatedAt = r.ResolvedAt
	if err = persistProductionLedger(ledgerPath, ledger); err != nil { return "WAIT_RECONCILIATION" }
	state.Ledger = ledger
	return "RECONCILIATION_RESOLVED_STOPPED"
}

// Outcomes may still arrive after STOP. Recording them never resumes the stopped lease.
func RecordProductionOutcome(state *M11State, dir, outcomeID, executionID, observedAt string) string {
	if state == nil || strings.TrimSpace(outcomeID) == "" || strings.TrimSpace(executionID) == "" { return "DENY_OUTCOME" }
	lockPath := productionLockPath(dir, state.Lease.LeaseID, state.Lease.LeaseVersion)
	lock, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil { return "WAIT_PRODUCTION_LOCK" }
	_ = lock.Close(); defer os.Remove(lockPath)
	ledgerPath := productionLedgerPath(dir, state.Lease.LeaseID, state.Lease.LeaseVersion)
	ledger, err := loadProductionLedger(ledgerPath)
	if err != nil || ledger.ReconciliationRequired || ledger.PendingOutcomes < 1 { return "WAIT_RECONCILIATION" }
	if ledger.LeaseID != state.Lease.LeaseID || ledger.LeaseVersion != state.Lease.LeaseVersion || ledger.LeaseHash != state.Lease.LeaseHash {
		return "WAIT_RECONCILIATION"
	}
	for _, link := range ledger.OutcomeLinks { if link.OutcomeID == outcomeID { return "DUPLICATE_OUTCOME" } }
	idx := -1
	for i, id := range ledger.PendingExecutionIDs { if id == executionID { idx = i; break } }
	if idx < 0 { return "DENY_OUTCOME_LINK" }
	observed, err := time.Parse(time.RFC3339, observedAt)
	if err != nil { return "DENY_OUTCOME" }
	if state.Execution != nil && state.Execution.ExecutionID == executionID {
		attempted, err := time.Parse(time.RFC3339, state.Execution.AttemptedAt)
		if err != nil || observed.Before(attempted) { return "DENY_OUTCOME" }
	}
	ledger.PendingExecutionIDs = append(ledger.PendingExecutionIDs[:idx], ledger.PendingExecutionIDs[idx+1:]...)
	ledger.PendingOutcomes--
	ledger.OutcomeLinks = append(ledger.OutcomeLinks, ProductionOutcomeLink{OutcomeID: outcomeID, ExecutionID: executionID, ObservedAt: observedAt})
	ledger.ConsecutiveFailures = 0
	ledger.LastOutcomeAt = observedAt
	ledger.UpdatedAt = observedAt
	if err = persistProductionLedger(ledgerPath, ledger); err != nil { return "WAIT_RECONCILIATION" }
	state.Ledger = ledger
	return "OUTCOME_RECORDED"
}

func ValidateProductionClosedCycle(cycle ProductionCycleRecord, state M11State, gate ProductionGateDecision, auth ProductionExecutionAuthorization, exec ProductionExecutionRecord, outcome OutcomeRecord, evaluation EvaluationRecord, proposal *ImprovementProposal, review *ReviewRecord) string {
	if strings.TrimSpace(cycle.CycleID) == "" || cycle.LeaseID != state.Lease.LeaseID || cycle.LeaseVersion != state.Lease.LeaseVersion ||
		cycle.LeaseHash != state.Lease.LeaseHash || len(cycle.ObservationIDs) == 0 || cycle.DecisionID != state.Intent.DecisionID ||
		cycle.IntentID != state.Intent.IntentID || cycle.IntentHash != state.Intent.IntentHash || cycle.GateID != gate.GateID ||
		cycle.AuthorizationID != auth.AuthorizationID || cycle.ExecutionID != exec.ExecutionID || cycle.OutcomeID != outcome.OutcomeID ||
		cycle.EvaluationID != evaluation.EvaluationID || cycle.CorrelationID != state.Intent.CorrelationID {
		return "BROKEN_LINK"
	}
	for _, id := range cycle.ObservationIDs { if !containsFold(state.Intent.EvidenceIDs, id) { return "BROKEN_LINK" } }
	openedAt, errOpen := time.Parse(time.RFC3339, cycle.OpenedAt)
	closedAt, errClose := time.Parse(time.RFC3339, cycle.ClosedAt)
	if errOpen != nil || errClose != nil || closedAt.Before(openedAt) { return missionInvalid }
	if gate.Decision != "ALLOW_PRODUCTION" || gate.ExecutionAuthorized || !auth.ExecutionAuthorized || auth.ExecutionMode != "GOVERNED_PRODUCTION" ||
		auth.IntentID != state.Intent.IntentID || auth.IntentHash != state.Intent.IntentHash || auth.ProductionLeaseHash != state.Lease.LeaseHash ||
		exec.AuthorizationID != auth.AuthorizationID || exec.IntentHash != auth.IntentHash || exec.Status != "SUCCEEDED" || exec.SideEffectState != "PERFORMED" {
		return "BROKEN_LINK"
	}
	if ValidateOutcomeRecord(outcome) != missionValid || outcome.ActionID != exec.ExecutionID { return "BROKEN_LINK" }
	outcomeAt, _ := time.Parse(time.RFC3339, outcome.ObservedAt)
	execAt, errExec := time.Parse(time.RFC3339, exec.AttemptedAt)
	if errExec != nil || outcomeAt.Before(execAt) { return "OUTCOME_BEFORE_ACTION" }
	if evaluation.DecisionID != state.Intent.DecisionID || evaluation.ActionID != exec.ExecutionID || len(evaluation.OutcomeIDs) == 0 ||
		!containsFold(evaluation.OutcomeIDs, outcome.OutcomeID) || evaluation.EvaluationID != cycle.EvaluationID {
		return "BROKEN_LINK"
	}
	validResult := map[string]bool{"SUPPORTED": true, "NOT_SUPPORTED": true, "INCONCLUSIVE": true, "NEEDS_MORE_DATA": true}
	if !validResult[evaluation.Result] { return missionInvalid }
	evaluatedAt, errEval := time.Parse(time.RFC3339, evaluation.EvaluatedAt)
	if errEval != nil || evaluatedAt.Before(outcomeAt) { return missionInvalid }
	if proposal == nil {
		if cycle.ImprovementProposalID != "" || cycle.ReviewID != "" || cycle.Status != "CLOSED" { return "BROKEN_LINK" }
		return missionValid
	}
	if EvaluateImprovementProposal(*proposal) == "REJECT_AUTO_APPLY" { return "REJECT_AUTO_APPLY" }
	if ValidateProposalEvaluationLink(*proposal, []EvaluationRecord{evaluation}) != missionValid || cycle.ImprovementProposalID != proposal.ProposalID {
		return "BROKEN_LINK"
	}
	if review == nil {
		if cycle.Status != "REVIEW_PENDING" || cycle.ReviewID != "" { return "BROKEN_LINK" }
		return "REVIEW_REQUIRED"
	}
	if ValidateReviewRecord(*review, *proposal) != missionValid || cycle.ReviewID != review.ReviewID || cycle.Status != "CLOSED" {
		return "BROKEN_LINK"
	}
	return missionValid
}

func baseDemoM11() (M11State, M11Context) {
	intent := SealShadowActionIntent(ShadowActionIntent{IntentID: "i11", DecisionID: "d11", EvidenceIDs: []string{"e11"}, ActionType: "PREPARE_LOCAL_DRAFT", Target: "https://example.com/prod", Parameters: map[string]any{"title": "production"}, ProposedBy: "human", CreatedAt: "2026-09-03T07:00:00Z", ExpiresAt: "2026-09-03T11:00:00Z", CorrelationID: "c11", IdempotencyKey: "k11"})
	policyCtx := ShadowPolicyContext{PolicyVersion: "m11-v1", Now: "2026-09-03T07:05:00Z", KnownDecisionIDs: []string{"d11"}, KnownEvidenceIDs: []string{"e11"}, AllowedHosts: []string{"example.com"}, ActionRisk: map[string]string{"PREPARE_LOCAL_DRAFT": "RISK0", "UPDATE_DRAFT": "RISK1", "PUBLISH_CONTENT": "RISK2"}, SeenIdempotency: map[string]string{}}
	policy := EvaluateShadowPolicy(intent, policyCtx)
	lease := SealProductionLease(ProductionLease{LeaseID: "lease11", LeaseVersion: "v1", PolicyVersion: "m11-v1", ApprovalRef: "prod-approval-11", ReviewedBy: "human", ReviewerID: "learner", ReviewedAt: "2026-09-03T07:10:00Z", PromotionReviewRef: "promotion-11", SourceCanaryGrantID: "g10", SourceCanaryGrantVersion: "v1", SourceCanaryGrantHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ValidFrom: "2026-09-03T07:15:00Z", ExpiresAt: "2026-09-03T10:00:00Z", AllowedRiskClasses: []string{"RISK0"}, AllowedActionTypes: []string{"PREPARE_LOCAL_DRAFT"}, AllowedHosts: []string{"example.com"}, ExecutorIDs: []string{"local_sandbox"}, MaxExecutionsTotal: 10, MaxExecutionsPerWindow: 2, WindowSeconds: 3600, MaxCostMinorTotal: 1000, Currency: "USD", MaxPendingOutcomes: 1, MaxConsecutiveFailures: 2, MaxOutcomeAgeSeconds: 3600, MaxHealthSnapshotAgeSeconds: 300, KillSwitchRequired: true, CorrelationID: "lease-c11"})
	approval := ProductionLeaseApproval{ApprovalID: lease.ApprovalRef, LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, PromotionReviewRef: lease.PromotionReviewRef, SourceCanaryGrantID: lease.SourceCanaryGrantID, SourceCanaryGrantVersion: lease.SourceCanaryGrantVersion, SourceCanaryGrantHash: lease.SourceCanaryGrantHash, SourceE5Refs: []string{"e5-canary-10"}, ValidatedRiskClasses: []string{"RISK0"}, ReviewedBy: "human", ReviewerID: lease.ReviewerID, ReviewedAt: lease.ReviewedAt, Decision: "APPROVE_PRODUCTION_LEASE"}
	health := SealProductionHealth(ProductionHealthSnapshot{SnapshotID: "health11", LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ObservedAt: "2026-09-03T07:59:00Z", SourceRefs: []string{"otel:service", "provider:status"}, DependencyState: "HEALTHY", TelemetryComplete: true, HashVersion: "go-json-v1"})
	cost := SealCanaryCostBound(CanaryCostBound{CostBoundID: "cost11", IntentID: intent.IntentID, IntentHash: intent.IntentHash, MaxCostMinor: 10, Currency: "USD", SourceRef: "deterministic-cost-registry", ObservedAt: "2026-09-03T07:55:00Z", ExpiresAt: "2026-09-03T09:00:00Z", CorrelationID: intent.CorrelationID})
	ledger := ProductionLedger{LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ControlMode: "NORMAL", WindowStartedAt: "2026-09-03T07:15:00Z", UpdatedAt: "2026-09-03T07:15:00Z"}
	state := M11State{Intent: intent, Policy: policy, Lease: lease, Health: health, CostBound: cost, Ledger: ledger}
	ctx := M11Context{Now: "2026-09-03T08:00:00Z", TrustedLeaseApprovals: map[string]ProductionLeaseApproval{approval.ApprovalID: approval}, KnownCanaryGrantHashes: map[string]string{productionCanaryKey(lease.SourceCanaryGrantID, lease.SourceCanaryGrantVersion): lease.SourceCanaryGrantHash}, KnownE5Refs: []string{"e5-canary-10"}, TrustedHealthSnapshots: map[string]string{health.SnapshotID: health.SnapshotHash}, TrustedCostBounds: map[string]string{cost.CostBoundID: cost.CostBoundHash}, Executor: ExecutorProfile{ExecutorID: "local_sandbox", AllowedActionTypes: []string{"PREPARE_LOCAL_DRAFT"}, AllowedHosts: []string{"example.com"}}, AllowedExecutorIDs: []string{"local_sandbox"}, PolicyContext: policyCtx}
	return state, ctx
}

func demoM11() (map[string]any, error) {
	state, ctx := baseDemoM11()
	dir, err := os.MkdirTemp("", "m11-production-")
	if err != nil { return nil, err }
	if status := InitializeProductionLedger(&state, dir, ctx.Now); status != "INITIALIZED" { return nil, errors.New(status) }
	auth, gate, status := AuthorizeProduction(state, ctx)
	if status != "AUTHORIZED" { return nil, errors.New(status) }
	state.Authorization = &auth
	rec, status := ExecuteProductionLocalSandbox(&state, auth, ctx, dir)
	if status != "EXECUTED" { return nil, errors.New(status) }
	return map[string]any{"gate": gate, "authorization": auth, "execution": rec, "ledger": state.Ledger, "authority": "GOVERNED_PRODUCTION"}, nil
}
