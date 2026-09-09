// Package m11 owns the canonical, non-authorizing M11 artifact boundary.
// It validates persisted production lifecycle records, but never grants an
// execution right or invokes an executor.
package m11

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
	corem10 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m10"
)

const Valid = "VALID"

type ProductionLease struct {
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
	LeaseHash                   string   `json:"lease_hash"`
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

func (l ProductionLedger) MarshalJSON() ([]byte, error) {
	type wire ProductionLedger
	v := wire(l)
	if v.PendingExecutionIDs == nil {
		v.PendingExecutionIDs = []string{}
	}
	if v.SuccessfulIdempotencyKeys == nil {
		v.SuccessfulIdempotencyKeys = []string{}
	}
	if v.OutcomeLinks == nil {
		v.OutcomeLinks = []ProductionOutcomeLink{}
	}
	if v.ReconciliationResolutionIDs == nil {
		v.ReconciliationResolutionIDs = []string{}
	}
	return json.Marshal(v)
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
	AuthorizationID              string `json:"authorization_id"`
	IntentID                     string `json:"intent_id"`
	IntentHash                   string `json:"intent_hash"`
	PolicyVersion                string `json:"policy_version"`
	ProductionLeaseID            string `json:"production_lease_id"`
	ProductionLeaseVersion       string `json:"production_lease_version"`
	ProductionLeaseHash          string `json:"production_lease_hash"`
	ProductionGateID             string `json:"production_gate_id"`
	ProductionHealthSnapshotID   string `json:"production_health_snapshot_id"`
	ProductionHealthSnapshotHash string `json:"production_health_snapshot_hash"`
	ProductionCostBoundID        string `json:"production_cost_bound_id"`
	ProductionCostBoundHash      string `json:"production_cost_bound_hash"`
	ProductionCostBoundMinor     int64  `json:"production_cost_bound_minor"`
	ExecutorID                   string `json:"executor_id"`
	AuthorizedAt                 string `json:"authorized_at"`
	ExpiresAt                    string `json:"expires_at"`
	IdempotencyKey               string `json:"idempotency_key"`
	CorrelationID                string `json:"correlation_id"`
	ExecutionMode                string `json:"execution_mode"`
	ExecutionAuthorized          bool   `json:"execution_authorized"`
}

type ProductionExecutionRecord struct {
	ExecutionID                  string `json:"execution_id"`
	AuthorizationID              string `json:"authorization_id"`
	ProductionLeaseID            string `json:"production_lease_id"`
	ProductionLeaseVersion       string `json:"production_lease_version"`
	ProductionLeaseHash          string `json:"production_lease_hash"`
	ProductionGateID             string `json:"production_gate_id"`
	ProductionHealthSnapshotID   string `json:"production_health_snapshot_id"`
	ProductionHealthSnapshotHash string `json:"production_health_snapshot_hash"`
	ProductionCostBoundID        string `json:"production_cost_bound_id"`
	ProductionCostBoundHash      string `json:"production_cost_bound_hash"`
	ProductionCostBoundMinor     int64  `json:"production_cost_bound_minor"`
	IntentID                     string `json:"intent_id"`
	IntentHash                   string `json:"intent_hash"`
	ExecutorID                   string `json:"executor_id"`
	IdempotencyKey               string `json:"idempotency_key"`
	AttemptedAt                  string `json:"attempted_at"`
	Status                       string `json:"status"`
	SideEffectState              string `json:"side_effect_state"`
	ExternalRef                  string `json:"external_ref,omitempty"`
	Error                        string `json:"error,omitempty"`
	CorrelationID                string `json:"correlation_id"`
}

type ProductionActivationRecord struct {
	LeaseID      string `json:"lease_id"`
	LeaseVersion string `json:"lease_version"`
	LeaseHash    string `json:"lease_hash"`
	ActivatedAt  string `json:"activated_at"`
}
type ProductionReconciliationResolution struct {
	ResolutionID string `json:"resolution_id"`
	LeaseID      string `json:"lease_id"`
	LeaseVersion string `json:"lease_version"`
	LeaseHash    string `json:"lease_hash"`
	ExecutionID  string `json:"execution_id"`
	ResolvedBy   string `json:"resolved_by"`
	ResolverID   string `json:"resolver_id"`
	ResolvedAt   string `json:"resolved_at"`
	EffectState  string `json:"effect_state"`
	Reason       string `json:"reason"`
}
type ProductionOutcomeEvaluation struct {
	EvaluationID  string   `json:"evaluation_id"`
	LeaseID       string   `json:"lease_id"`
	LeaseVersion  string   `json:"lease_version"`
	LeaseHash     string   `json:"lease_hash"`
	ExecutionID   string   `json:"execution_id"`
	OutcomeID     string   `json:"outcome_id"`
	EvaluatedAt   string   `json:"evaluated_at"`
	Result        string   `json:"result"`
	EvidenceIDs   []string `json:"evidence_ids"`
	Limitations   []string `json:"limitations"`
	SourceProfile string   `json:"source_profile"`
}
type ProductionCycleRecord struct {
	CycleID               string   `json:"cycle_id"`
	LeaseID               string   `json:"lease_id"`
	LeaseVersion          string   `json:"lease_version"`
	LeaseHash             string   `json:"lease_hash"`
	ObservationIDs        []string `json:"observation_ids"`
	DecisionID            string   `json:"decision_id"`
	IntentID              string   `json:"intent_id"`
	IntentHash            string   `json:"intent_hash"`
	GateID                string   `json:"gate_id"`
	AuthorizationID       string   `json:"authorization_id"`
	ExecutionID           string   `json:"execution_id"`
	OutcomeID             string   `json:"outcome_id"`
	EvaluationID          string   `json:"evaluation_id"`
	ImprovementProposalID string   `json:"improvement_proposal_id,omitempty"`
	ReviewID              string   `json:"review_id,omitempty"`
	Status                string   `json:"status"`
	OpenedAt              string   `json:"opened_at"`
	ClosedAt              string   `json:"closed_at"`
	CorrelationID         string   `json:"correlation_id"`
}

type leaseHashPayload struct {
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

func digest(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func sorted(values []string) []string {
	copied := append([]string(nil), values...)
	sort.Strings(copied)
	return copied
}
func hasWildcard(values []string) bool {
	for _, value := range values {
		if value == "*" {
			return true
		}
	}
	return false
}

func ComputeProductionLeaseHash(x ProductionLease) string {
	return digest(leaseHashPayload{LeaseID: x.LeaseID, LeaseVersion: x.LeaseVersion, PolicyVersion: x.PolicyVersion, ApprovalRef: x.ApprovalRef, ReviewedBy: x.ReviewedBy, ReviewerID: x.ReviewerID, ReviewedAt: x.ReviewedAt, PromotionReviewRef: x.PromotionReviewRef, SourceCanaryGrantID: x.SourceCanaryGrantID, SourceCanaryGrantVersion: x.SourceCanaryGrantVersion, SourceCanaryGrantHash: x.SourceCanaryGrantHash, ValidFrom: x.ValidFrom, ExpiresAt: x.ExpiresAt, AllowedRiskClasses: sorted(x.AllowedRiskClasses), AllowedActionTypes: sorted(x.AllowedActionTypes), AllowedHosts: sorted(x.AllowedHosts), ExecutorIDs: sorted(x.ExecutorIDs), MaxExecutionsTotal: x.MaxExecutionsTotal, MaxExecutionsPerWindow: x.MaxExecutionsPerWindow, WindowSeconds: x.WindowSeconds, MaxCostMinorTotal: x.MaxCostMinorTotal, Currency: x.Currency, MaxPendingOutcomes: x.MaxPendingOutcomes, MaxConsecutiveFailures: x.MaxConsecutiveFailures, MaxOutcomeAgeSeconds: x.MaxOutcomeAgeSeconds, MaxHealthSnapshotAgeSeconds: x.MaxHealthSnapshotAgeSeconds, KillSwitchRequired: x.KillSwitchRequired, CorrelationID: x.CorrelationID, HashVersion: x.HashVersion})
}
func ComputeProductionHealthHash(x ProductionHealthSnapshot) string {
	return digest(healthHashPayload{SnapshotID: x.SnapshotID, LeaseID: x.LeaseID, LeaseVersion: x.LeaseVersion, LeaseHash: x.LeaseHash, ObservedAt: x.ObservedAt, SourceRefs: sorted(x.SourceRefs), DependencyState: x.DependencyState, TelemetryComplete: x.TelemetryComplete, ConsecutiveFailures: x.ConsecutiveFailures, ReconciliationRequired: x.ReconciliationRequired, ComplianceAlertCount: x.ComplianceAlertCount, OldestPendingOutcomeAgeSeconds: x.OldestPendingOutcomeAgeSeconds, HashVersion: x.HashVersion})
}

func before(left, right string) bool {
	a, errA := time.Parse(time.RFC3339, left)
	b, errB := time.Parse(time.RFC3339, right)
	return errA == nil && errB == nil && a.Before(b)
}

// DecodeArtifact validates exactly one untrusted M11 artifact. VALID means
// schema and local integrity checks passed; it never means execution is allowed.
func DecodeArtifact(kind string, raw []byte) (any, string) {
	if kind == "cost" {
		return corem10.DecodeTrustedCostBound(raw)
	}
	var schema string
	var value any
	switch kind {
	case "lease":
		schema, value = "production-lease.schema.json", &ProductionLease{}
	case "approval":
		schema, value = "production-lease-approval.schema.json", &ProductionLeaseApproval{}
	case "health":
		schema, value = "production-health-snapshot.schema.json", &ProductionHealthSnapshot{}
	case "ledger":
		schema, value = "production-ledger.schema.json", &ProductionLedger{}
	case "gate":
		schema, value = "production-gate-decision.schema.json", &ProductionGateDecision{}
	case "authorization":
		schema, value = "execution-authorization.schema.json", &ProductionExecutionAuthorization{}
	case "execution":
		schema, value = "execution-record.schema.json", &ProductionExecutionRecord{}
	case "activation":
		schema, value = "production-activation-record.schema.json", &ProductionActivationRecord{}
	case "resolution":
		schema, value = "production-reconciliation-resolution.schema.json", &ProductionReconciliationResolution{}
	case "evaluation":
		schema, value = "production-outcome-evaluation.schema.json", &ProductionOutcomeEvaluation{}
	case "cycle":
		schema, value = "production-cycle-record.schema.json", &ProductionCycleRecord{}
	default:
		return nil, "INVALID_PROFILE"
	}
	if contracts.ValidateRaw(schema, raw) != nil || contracts.DecodeStrict(raw, value) != nil {
		return nil, "INVALID_SCHEMA"
	}
	switch x := value.(type) {
	case *ProductionLease:
		if x.LeaseHash != ComputeProductionLeaseHash(*x) {
			return nil, "TAMPERED_LEASE"
		}
		if before(x.ValidFrom, x.ReviewedAt) || !before(x.ValidFrom, x.ExpiresAt) {
			return nil, "INVALID_TIME_BINDING"
		}
		if x.MaxExecutionsPerWindow > x.MaxExecutionsTotal || hasWildcard(x.AllowedActionTypes) || hasWildcard(x.ExecutorIDs) {
			return nil, "INVALID_LEASE"
		}
	case *ProductionHealthSnapshot:
		if x.SnapshotHash != ComputeProductionHealthHash(*x) {
			return nil, "TAMPERED_HEALTH"
		}
	case *ProductionExecutionAuthorization:
		if x.ExecutionMode != "GOVERNED_PRODUCTION" {
			return nil, "INVALID_PROFILE"
		}
		if !before(x.AuthorizedAt, x.ExpiresAt) {
			return nil, "INVALID_TIME_BINDING"
		}
	case *ProductionCycleRecord:
		if before(x.ClosedAt, x.OpenedAt) {
			return nil, "INVALID_TIME_BINDING"
		}
	case *ProductionOutcomeEvaluation:
		if x.SourceProfile != "OFFLINE_FIXTURE" || x.Result != "FIXTURE_NO_SIDE_EFFECT" || len(x.EvidenceIDs) == 0 || len(x.Limitations) == 0 {
			return nil, "INVALID_EVALUATION"
		}
	case *ProductionLedger:
		if x.ExecutionsInWindow > x.ExecutionsTotal || x.PendingOutcomes != len(x.PendingExecutionIDs) || x.PendingOutcomes > x.ExecutionsTotal || len(x.SuccessfulIdempotencyKeys) > x.ExecutionsTotal {
			return nil, "INVALID_LEDGER"
		}
		if before(x.UpdatedAt, x.WindowStartedAt) || x.LastExecutionAt != "" && before(x.UpdatedAt, x.LastExecutionAt) || x.LastOutcomeAt != "" && before(x.UpdatedAt, x.LastOutcomeAt) {
			return nil, "INVALID_TIME_BINDING"
		}
		outcomes, executions := map[string]bool{}, map[string]bool{}
		for _, link := range x.OutcomeLinks {
			if outcomes[link.OutcomeID] || executions[link.ExecutionID] || before(x.UpdatedAt, link.ObservedAt) {
				return nil, "INVALID_LEDGER"
			}
			outcomes[link.OutcomeID], executions[link.ExecutionID] = true, true
		}
		for _, id := range x.PendingExecutionIDs {
			if executions[id] {
				return nil, "INVALID_LEDGER"
			}
		}
	}
	return value, Valid
}

func CanonicalArtifact(kind string, raw []byte) (string, []byte, string) {
	value, status := DecodeArtifact(kind, raw)
	if status != Valid {
		return "", nil, status
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return "", nil, "INVALID_SCHEMA"
	}
	var id string
	switch x := value.(type) {
	case *ProductionLease:
		id = x.LeaseID
	case *ProductionLeaseApproval:
		id = x.ApprovalID
	case *ProductionHealthSnapshot:
		id = x.SnapshotID
	case *ProductionLedger:
		id = x.LeaseID + "/" + x.UpdatedAt
	case *ProductionGateDecision:
		id = x.GateID
	case *ProductionExecutionAuthorization:
		id = x.AuthorizationID
	case *ProductionExecutionRecord:
		id = x.ExecutionID
	case *ProductionActivationRecord:
		id = x.LeaseID + "/" + x.LeaseVersion
	case *ProductionReconciliationResolution:
		id = x.ResolutionID
	case *ProductionOutcomeEvaluation:
		id = x.EvaluationID
	case *ProductionCycleRecord:
		id = x.CycleID
	case corem10.TrustedCostBound:
		id = x.CostBoundID
	}
	if strings.TrimSpace(id) == "" {
		return "", nil, "INVALID_SCHEMA"
	}
	return id, canonical, Valid
}
