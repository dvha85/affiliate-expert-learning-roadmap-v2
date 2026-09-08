package m10

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

// CanaryLedgerSnapshot is the immutable view of mutable usage used when a
// gate is evaluated. The gate never changes this snapshot or grants execution
// permission by itself.
type CanaryLedgerSnapshot struct {
	ExecutionsTotal    int   `json:"executions_total"`
	ExecutionsInWindow int   `json:"executions_in_window"`
	CostMinorTotal     int64 `json:"cost_minor_total"`
	PendingOutcomes    int   `json:"pending_outcomes"`
}

type CanaryGateInput struct {
	Grant          CanaryGrant
	CostBound      TrustedCostBound
	IntentID       string
	IntentHash     string
	PolicyVersion  string
	PolicyDecision string
	RiskClass      string
	ApprovalID     string
	ApproverID     string
	CorrelationID  string
	ActionType     string
	Target         string
	Now            string
	Ledger         CanaryLedgerSnapshot
}

// CanaryGateDecision is the non-authorizing decision artifact consumed by a
// later authorization boundary. Its JSON is compatible with the contract
// schema and has execution_authorized literally false.
type CanaryGateDecision struct {
	GateID                    string `json:"gate_id"`
	GrantID                   string `json:"grant_id"`
	GrantVersion              string `json:"grant_version"`
	GrantHash                 string `json:"grant_hash"`
	IntentID                  string `json:"intent_id"`
	IntentHash                string `json:"intent_hash"`
	PolicyVersion             string `json:"policy_version"`
	RiskClass                 string `json:"risk_class"`
	CostBoundID               string `json:"cost_bound_id"`
	CostBoundHash             string `json:"cost_bound_hash"`
	CostBoundMinor            int64  `json:"cost_bound_minor"`
	Decision                  string `json:"decision"`
	Reason                    string `json:"reason"`
	EvaluatedAt               string `json:"evaluated_at"`
	ExecutionsTotalBefore     int    `json:"executions_total_before"`
	ExecutionsInWindowBefore  int    `json:"executions_in_window_before"`
	CostMinorTotalBefore      int64  `json:"cost_minor_total_before"`
	PendingOutcomesBefore     int    `json:"pending_outcomes_before"`
	PerActionApprovalRequired bool   `json:"per_action_approval_required"`
	ExecutionAuthorized       bool   `json:"execution_authorized"`
}

func canaryGateID(in CanaryGateInput) string {
	raw, _ := json.Marshal([]any{in.Grant.GrantID, in.Grant.GrantVersion, in.Grant.GrantHash, in.CostBound.CostBoundID, in.CostBound.CostBoundHash, in.IntentID, in.IntentHash, in.PolicyVersion, in.Ledger})
	sum := sha256.Sum256(raw)
	return "gate-" + hex.EncodeToString(sum[:])
}

func newCanaryGate(in CanaryGateInput) CanaryGateDecision {
	return CanaryGateDecision{
		GateID: canaryGateID(in), GrantID: in.Grant.GrantID, GrantVersion: in.Grant.GrantVersion, GrantHash: in.Grant.GrantHash,
		IntentID: in.IntentID, IntentHash: in.IntentHash, PolicyVersion: in.PolicyVersion, RiskClass: in.RiskClass,
		CostBoundID: in.CostBound.CostBoundID, CostBoundHash: in.CostBound.CostBoundHash, CostBoundMinor: in.CostBound.MaxCostMinor,
		Decision: "DENY", Reason: "INVALID_CANARY_INPUT", EvaluatedAt: in.Now,
		ExecutionsTotalBefore: in.Ledger.ExecutionsTotal, ExecutionsInWindowBefore: in.Ledger.ExecutionsInWindow,
		CostMinorTotalBefore: in.Ledger.CostMinorTotal, PendingOutcomesBefore: in.Ledger.PendingOutcomes,
		ExecutionAuthorized: false,
	}
}

func gateRequiresApproval(g *CanaryGateDecision, reason string) CanaryGateDecision {
	g.Decision = "REQUIRE_APPROVAL"
	g.Reason = reason
	g.PerActionApprovalRequired = true
	return *g
}

// EvaluateCanaryGate is intentionally a compact shared learner profile. It
// evaluates a fully sealed grant, trusted bound and a snapshot of ledger use;
// it is not an executor and cannot create a side effect.
func EvaluateCanaryGate(in CanaryGateInput) CanaryGateDecision {
	g := newCanaryGate(in)
	now, err := time.Parse(time.RFC3339, in.Now)
	if err != nil || in.Ledger.ExecutionsTotal < 0 || in.Ledger.ExecutionsInWindow < 0 || in.Ledger.CostMinorTotal < 0 || in.Ledger.PendingOutcomes < 0 {
		g.Reason = "INVALID_GATE_TIME_OR_LEDGER"
		return g
	}
	grantRaw, err := json.Marshal(in.Grant)
	if err != nil {
		g.Reason = "INVALID_GRANT"
		return g
	}
	if _, status := DecodeCanaryGrant(grantRaw); status != "VALID" {
		g.Reason = status
		return g
	}
	if status := ValidCanaryGrantFor(in.Grant, in.IntentID, in.IntentHash, in.PolicyVersion, in.ApprovalID, in.ApproverID, in.CorrelationID, in.RiskClass, in.ActionType, in.Target, now); status != "VALID" {
		g.Reason = status
		return g
	}
	boundRaw, err := json.Marshal(in.CostBound)
	if err != nil {
		g.Reason = "INVALID_COST_BOUND"
		return g
	}
	if _, status := DecodeTrustedCostBound(boundRaw); status != "VALID" {
		g.Reason = status
		return g
	}
	if status := ValidFor(in.CostBound, in.IntentID, in.IntentHash, in.CorrelationID, in.Grant.Currency, now); status != "VALID" {
		g.Reason = status
		return g
	}
	if in.PolicyDecision != "ALLOW" || in.RiskClass != "RISK0" {
		return gateRequiresApproval(&g, "POLICY_NOT_CANARY_ELIGIBLE")
	}
	if in.Ledger.ExecutionsTotal >= in.Grant.MaxExecutionsTotal {
		return gateRequiresApproval(&g, "CANARY_TOTAL_BUDGET_EXHAUSTED")
	}
	if in.Ledger.ExecutionsInWindow >= in.Grant.MaxExecutionsPerWindow {
		g.Decision, g.Reason = "WAIT", "CANARY_WINDOW_BUDGET_EXHAUSTED"
		return g
	}
	if in.Ledger.CostMinorTotal > in.Grant.MaxCostMinorTotal || in.CostBound.MaxCostMinor > in.Grant.MaxCostMinorTotal-in.Ledger.CostMinorTotal {
		return gateRequiresApproval(&g, "CANARY_COST_BUDGET_EXHAUSTED")
	}
	if in.Ledger.PendingOutcomes >= in.Grant.MaxPendingOutcomes {
		g.Decision, g.Reason = "WAIT", "OUTCOME_BACKPRESSURE"
		return g
	}
	g.Decision, g.Reason = "ALLOW_CANARY", "CANARY_ELIGIBLE"
	return g
}

func ValidateCanaryGateDecision(raw []byte) (CanaryGateDecision, error) {
	var gate CanaryGateDecision
	if err := contracts.ValidateRaw("canary-gate-decision.schema.json", raw); err != nil {
		return gate, err
	}
	if err := contracts.DecodeStrict(raw, &gate); err != nil || gate.ExecutionAuthorized {
		return CanaryGateDecision{}, fmt.Errorf("invalid canary gate decision")
	}
	return gate, nil
}
