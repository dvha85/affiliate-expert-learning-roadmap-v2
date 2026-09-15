package m10

import (
	"net/url"
	"strings"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m08"
)

// CanaryGrantApproval is the immutable approval envelope used by the
// mission-harness canary profile. It is deliberately non-authorizing; the
// later gate and authorization artifacts remain separate capabilities.
type CanaryGrantApproval struct {
	ApprovalRef  string `json:"approval_ref"`
	GrantID      string `json:"grant_id"`
	GrantVersion string `json:"grant_version"`
	GrantHash    string `json:"grant_hash"`
	Decision     string `json:"decision"`
	ApprovedBy   string `json:"approved_by"`
	ApproverID   string `json:"approver_id"`
	ApprovedAt   string `json:"approved_at"`
}

// CanaryOutcomeLink is the immutable link from a pending execution to its
// offline fixture outcome. Mutable ledger ownership remains with the runtime.
type CanaryOutcomeLink struct {
	OutcomeID   string `json:"outcome_id"`
	ExecutionID string `json:"execution_id"`
	ObservedAt  string `json:"observed_at"`
}

// CanaryLedger is the historical pre-gate ledger snapshot consumed by the
// canary chain validator. It is not an authorization or a writer API.
type CanaryLedger struct {
	GrantID                   string              `json:"grant_id"`
	GrantVersion              string              `json:"grant_version"`
	GrantHash                 string              `json:"grant_hash"`
	WindowStartedAt           string              `json:"window_started_at"`
	ExecutionsTotal           int                 `json:"executions_total"`
	ExecutionsInWindow        int                 `json:"executions_in_window"`
	CostMinorTotal            int64               `json:"cost_minor_total"`
	PendingOutcomes           int                 `json:"pending_outcomes"`
	PendingExecutionIDs       []string            `json:"pending_execution_ids"`
	SuccessfulIdempotencyKeys []string            `json:"successful_idempotency_keys"`
	OutcomeLinks              []CanaryOutcomeLink `json:"outcome_links"`
	ReconciliationRequired    bool                `json:"reconciliation_required"`
	LastExecutionAt           string              `json:"last_execution_at,omitempty"`
	UpdatedAt                 string              `json:"updated_at"`
}

func containsCanaryValue(values []string, want string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(want)) {
			return true
		}
	}
	return false
}

func hasCanaryWildcard(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "*" {
			return true
		}
	}
	return false
}

func canaryTargetAllowed(target string, hosts []string) bool {
	parsed, err := url.Parse(target)
	return err == nil && parsed.Scheme == "https" && m08.AllowedHost(target, hosts)
}

func parseChainTimes(values ...string) ([]time.Time, bool) {
	out := make([]time.Time, len(values))
	for index, value := range values {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return nil, false
		}
		out[index] = parsed
	}
	return out, true
}

// ValidateHistoricalChain audits the complete M10 canary issuance chain.
// It has no wall-clock side effects, does not reserve budget and never calls
// an executor. The caller supplies already decoded files; this function owns
// all cross-artifact links and historical ledger/time checks.
func ValidateHistoricalChain(
	intent m08.Intent,
	policy m08.PolicyDecision,
	grant CanaryGrant,
	approval CanaryGrantApproval,
	cost TrustedCostBound,
	ledger CanaryLedger,
	gate CanaryGateDecision,
	authorization ExecutionAuthorization,
	execution ExecutionRecord,
) string {
	for _, value := range []string{
		intent.IntentID, intent.DecisionID, intent.CorrelationID, intent.IdempotencyKey,
		policy.PolicyVersion, grant.GrantID, grant.GrantVersion, grant.ApprovalRef,
		grant.ApproverID, cost.CostBoundID, cost.SourceRef, gate.GateID,
		authorization.AuthorizationID, authorization.ExecutorID, execution.ExecutionID,
	} {
		if strings.TrimSpace(value) == "" {
			return "INVALID_IDENTITY"
		}
	}
	if intent.IntentHash != m08.ComputeIntentHash(intent) {
		return "TAMPERED_INTENT"
	}
	if policy.IntentID != intent.IntentID || policy.IntentHash != intent.IntentHash ||
		grant.PolicyVersion != policy.PolicyVersion ||
		approval.ApprovalRef != grant.ApprovalRef || approval.GrantID != grant.GrantID ||
		approval.GrantVersion != grant.GrantVersion || approval.GrantHash != grant.GrantHash ||
		approval.ApproverID != grant.ApproverID || cost.IntentID != intent.IntentID ||
		cost.IntentHash != intent.IntentHash || cost.CorrelationID != intent.CorrelationID ||
		cost.Currency != grant.Currency || ledger.GrantID != grant.GrantID ||
		ledger.GrantVersion != grant.GrantVersion || ledger.GrantHash != grant.GrantHash {
		return "BROKEN_LINK"
	}
	if gate.GrantID != grant.GrantID || gate.GrantVersion != grant.GrantVersion ||
		gate.GrantHash != grant.GrantHash || gate.IntentID != intent.IntentID ||
		gate.IntentHash != intent.IntentHash || gate.PolicyVersion != policy.PolicyVersion ||
		gate.RiskClass != policy.RiskClass || gate.CostBoundID != cost.CostBoundID ||
		gate.CostBoundHash != cost.CostBoundHash || gate.CostBoundMinor != cost.MaxCostMinor {
		return "BROKEN_LINK"
	}
	if authorization.IntentID != intent.IntentID || authorization.IntentHash != intent.IntentHash ||
		authorization.PolicyVersion != policy.PolicyVersion || authorization.CanaryGrantID != grant.GrantID ||
		authorization.CanaryGrantVersion != grant.GrantVersion || authorization.CanaryGrantHash != grant.GrantHash ||
		authorization.CanaryGateID != gate.GateID || authorization.CanaryCostBoundID != cost.CostBoundID ||
		authorization.CanaryCostBoundHash != cost.CostBoundHash || authorization.CanaryCostBoundMinor != cost.MaxCostMinor ||
		authorization.CorrelationID != intent.CorrelationID || authorization.IdempotencyKey != intent.IdempotencyKey {
		return "BROKEN_LINK"
	}
	if execution.AuthorizationID != authorization.AuthorizationID || execution.IntentID != intent.IntentID ||
		execution.IntentHash != intent.IntentHash || execution.CanaryGrantID != grant.GrantID ||
		execution.CanaryGrantVersion != grant.GrantVersion || execution.CanaryGrantHash != grant.GrantHash ||
		execution.CanaryGateID != gate.GateID || execution.CanaryCostBoundID != cost.CostBoundID ||
		execution.CanaryCostBoundHash != cost.CostBoundHash || execution.CanaryCostBoundMinor != cost.MaxCostMinor ||
		execution.ExecutorID != authorization.ExecutorID || execution.IdempotencyKey != intent.IdempotencyKey ||
		execution.CorrelationID != intent.CorrelationID {
		return "BROKEN_LINK"
	}
	if gate.Decision != "ALLOW_CANARY" || gate.Reason != "CANARY_ELIGIBLE" || gate.PerActionApprovalRequired ||
		!((policy.RiskClass == "RISK0" && policy.Decision == "ALLOW") ||
			(policy.RiskClass == "RISK1" && policy.Decision == "HUMAN_REVIEW" && policy.Reason == "RISK1_REQUIRES_REVIEW")) {
		return "INVALID_GATE_STATE"
	}
	if grant.MaxExecutionsPerWindow > grant.MaxExecutionsTotal || hasCanaryWildcard(grant.AllowedActionTypes) ||
		hasCanaryWildcard(grant.ExecutorIDs) || !containsCanaryValue(grant.AllowedRiskClasses, policy.RiskClass) ||
		!containsCanaryValue(grant.AllowedActionTypes, intent.ActionType) ||
		!canaryTargetAllowed(intent.Target, grant.AllowedHosts) || !containsCanaryValue(grant.ExecutorIDs, authorization.ExecutorID) {
		return "SCOPE_NOT_DELEGATED"
	}
	times, ok := parseChainTimes(
		intent.CreatedAt, intent.ExpiresAt, policy.PolicyCheckedAt, grant.ApprovedAt,
		approval.ApprovedAt, grant.ValidFrom, grant.ExpiresAt, cost.ObservedAt,
		cost.ExpiresAt, ledger.WindowStartedAt, ledger.UpdatedAt, gate.EvaluatedAt,
		authorization.AuthorizedAt, authorization.ExpiresAt,
	)
	if !ok {
		return "INVALID_TIME_BINDING"
	}
	created, intentEnd, checked, approved, approvalAt, start, grantEnd := times[0], times[1], times[2], times[3], times[4], times[5], times[6]
	observed, costEnd, window, updated, evaluated, authorized, end := times[7], times[8], times[9], times[10], times[11], times[12], times[13]
	attempted, err := time.Parse(time.RFC3339, execution.AttemptedAt)
	if err != nil {
		return "INVALID_TIME_BINDING"
	}
	if !intentEnd.After(created) || checked.Before(created) || checked.After(evaluated) || !approved.Equal(approvalAt) ||
		evaluated.Before(start) || !evaluated.Before(grantEnd) || observed.After(evaluated) || !evaluated.Before(costEnd) ||
		!evaluated.Before(intentEnd) || window.Before(start) || updated.After(evaluated) || window.After(evaluated) ||
		!evaluated.Equal(authorized) || end.After(intentEnd) || end.After(grantEnd) || end.After(costEnd) ||
		attempted.Before(authorized) {
		return "INVALID_TIME_BINDING"
	}
	if execution.SideEffectState == "PERFORMED" && !attempted.Before(end) {
		return "EXPIRED_AUTHORIZATION"
	}
	elapsed := evaluated.Unix() - window.Unix()
	if evaluated.Nanosecond() < window.Nanosecond() {
		elapsed--
	}
	inWindow := ledger.ExecutionsInWindow
	if elapsed >= int64(grant.WindowSeconds) {
		inWindow = 0
	}
	if gate.ExecutionsTotalBefore != ledger.ExecutionsTotal || gate.ExecutionsInWindowBefore != inWindow ||
		gate.CostMinorTotalBefore != ledger.CostMinorTotal || gate.PendingOutcomesBefore != ledger.PendingOutcomes {
		return "LEDGER_SNAPSHOT_MISMATCH"
	}
	if ledger.ReconciliationRequired || containsCanaryValue(ledger.SuccessfulIdempotencyKeys, intent.IdempotencyKey) ||
		containsCanaryValue(ledger.PendingExecutionIDs, execution.ExecutionID) {
		return "LEDGER_BLOCKED"
	}
	for _, link := range ledger.OutcomeLinks {
		if link.ExecutionID == execution.ExecutionID {
			return "LEDGER_BLOCKED"
		}
	}
	if ledger.ExecutionsTotal >= grant.MaxExecutionsTotal || inWindow >= grant.MaxExecutionsPerWindow ||
		ledger.PendingOutcomes >= grant.MaxPendingOutcomes || ledger.CostMinorTotal > grant.MaxCostMinorTotal ||
		cost.MaxCostMinor > grant.MaxCostMinorTotal-ledger.CostMinorTotal {
		return "BUDGET_EXCEEDED"
	}
	return "VALID"
}
