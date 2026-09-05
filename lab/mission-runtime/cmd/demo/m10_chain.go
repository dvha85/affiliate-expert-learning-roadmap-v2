package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type M10ChainSummary struct {
	Result                  string `json:"result"`
	IntentID                string `json:"intent_id"`
	GrantID                 string `json:"grant_id"`
	ExecutionID             string `json:"execution_id"`
	ProvenanceAuthenticated bool   `json:"provenance_authenticated"`
	ExecutionPermitted      bool   `json:"execution_permitted"`
}

// CheckM10Chain audits a historical ALLOW_CANARY issuance chain with a pre-gate
// ledger snapshot. It never re-evaluates policy, authenticates provenance or executes.
func CheckM10Chain(raw [][]byte) (M10ChainSummary, string) {
	fail := func(s string) (M10ChainSummary, string) { return M10ChainSummary{}, s }
	if len(raw) != 9 {
		return fail("INVALID_INPUT")
	}
	i, status := DecodeM08Intent(raw[0])
	if status != missionValid {
		return fail(status)
	}
	p, status := DecodeM08Policy(raw[1])
	if status != missionValid {
		return fail(status)
	}
	values := make([]any, 7)
	for n, kind := range []string{"grant", "approval", "cost", "ledger", "gate", "authorization", "execution"} {
		v, s := DecodeM10Artifact(kind, raw[n+2])
		if s != missionValid {
			return fail(s)
		}
		values[n] = v
	}
	g := values[0].(*CanaryGrant)
	ap := values[1].(*CanaryGrantApproval)
	c := values[2].(*CanaryCostBound)
	l := values[3].(*CanaryLedger)
	gate := values[4].(*CanaryGateDecision)
	a := values[5].(*CanaryExecutionAuthorization)
	r := values[6].(*CanaryExecutionRecord)
	for _, v := range []string{i.IntentID, i.DecisionID, i.CorrelationID, i.IdempotencyKey, p.PolicyVersion, g.GrantID, g.GrantVersion, g.ApprovalRef, g.ApproverID, c.CostBoundID, c.SourceRef, gate.GateID, a.AuthorizationID, a.ExecutorID, r.ExecutionID} {
		if strings.TrimSpace(v) == "" {
			return fail("INVALID_IDENTITY")
		}
	}
	if i.IntentHash != ComputeShadowIntentHash(i) {
		return fail("TAMPERED_INTENT")
	}
	if p.IntentID != i.IntentID || p.IntentHash != i.IntentHash || g.PolicyVersion != p.PolicyVersion ||
		ap.ApprovalRef != g.ApprovalRef || ap.GrantID != g.GrantID || ap.GrantVersion != g.GrantVersion || ap.GrantHash != g.GrantHash || ap.ApproverID != g.ApproverID ||
		c.IntentID != i.IntentID || c.IntentHash != i.IntentHash || c.CorrelationID != i.CorrelationID || c.Currency != g.Currency ||
		l.GrantID != g.GrantID || l.GrantVersion != g.GrantVersion || l.GrantHash != g.GrantHash {
		return fail("BROKEN_LINK")
	}
	if gate.GrantID != g.GrantID || gate.GrantVersion != g.GrantVersion || gate.GrantHash != g.GrantHash || gate.IntentID != i.IntentID || gate.IntentHash != i.IntentHash || gate.PolicyVersion != p.PolicyVersion || gate.RiskClass != p.RiskClass || gate.CostBoundID != c.CostBoundID || gate.CostBoundHash != c.CostBoundHash || gate.CostBoundMinor != c.MaxCostMinor {
		return fail("BROKEN_LINK")
	}
	if a.IntentID != i.IntentID || a.IntentHash != i.IntentHash || a.PolicyVersion != p.PolicyVersion || a.CanaryGrantID != g.GrantID || a.CanaryGrantVersion != g.GrantVersion || a.CanaryGrantHash != g.GrantHash || a.CanaryGateID != gate.GateID || a.CanaryCostBoundID != c.CostBoundID || a.CanaryCostBoundHash != c.CostBoundHash || a.CanaryCostBoundMinor != c.MaxCostMinor || a.CorrelationID != i.CorrelationID || a.IdempotencyKey != i.IdempotencyKey {
		return fail("BROKEN_LINK")
	}
	if r.AuthorizationID != a.AuthorizationID || r.IntentID != i.IntentID || r.IntentHash != i.IntentHash || r.CanaryGrantID != g.GrantID || r.CanaryGrantVersion != g.GrantVersion || r.CanaryGrantHash != g.GrantHash || r.CanaryGateID != gate.GateID || r.CanaryCostBoundID != c.CostBoundID || r.CanaryCostBoundHash != c.CostBoundHash || r.CanaryCostBoundMinor != c.MaxCostMinor || r.ExecutorID != a.ExecutorID || r.IdempotencyKey != i.IdempotencyKey || r.CorrelationID != i.CorrelationID {
		return fail("BROKEN_LINK")
	}
	if gate.Decision != "ALLOW_CANARY" || gate.Reason != "CANARY_ELIGIBLE" || gate.PerActionApprovalRequired ||
		!(p.RiskClass == "RISK0" && p.Decision == "ALLOW" || p.RiskClass == "RISK1" && p.Decision == "HUMAN_REVIEW" && p.Reason == "RISK1_REQUIRES_REVIEW") {
		return fail("INVALID_GATE_STATE")
	}
	if g.MaxExecutionsPerWindow > g.MaxExecutionsTotal || hasWildcard(g.AllowedActionTypes) || hasWildcard(g.ExecutorIDs) ||
		!containsFold(g.AllowedRiskClasses, p.RiskClass) || !containsFold(g.AllowedActionTypes, i.ActionType) || !allowedHost(i.Target, g.AllowedHosts) || !containsFold(g.ExecutorIDs, a.ExecutorID) {
		return fail("SCOPE_NOT_DELEGATED")
	}
	// Parse all instants explicitly; compare instants rather than timestamp strings.
	times := make([]time.Time, 14)
	for n, v := range []string{i.CreatedAt, i.ExpiresAt, p.PolicyCheckedAt, g.ApprovedAt, ap.ApprovedAt, g.ValidFrom, g.ExpiresAt, c.ObservedAt, c.ExpiresAt, l.WindowStartedAt, l.UpdatedAt, gate.EvaluatedAt, a.AuthorizedAt, a.ExpiresAt} {
		x, e := time.Parse(time.RFC3339, v)
		if e != nil {
			return fail("INVALID_TIME_BINDING")
		}
		times[n] = x
	}
	created, intentEnd, checked, approved, approvalAt, start, grantEnd, observed, costEnd, window, updated, at, authorized, end := times[0], times[1], times[2], times[3], times[4], times[5], times[6], times[7], times[8], times[9], times[10], times[11], times[12], times[13]
	attempted, e := time.Parse(time.RFC3339, r.AttemptedAt)
	if e != nil {
		return fail("INVALID_TIME_BINDING")
	}
	if !intentEnd.After(created) || checked.Before(created) || checked.After(at) || !approved.Equal(approvalAt) || at.Before(start) || !at.Before(grantEnd) || observed.After(at) || !at.Before(costEnd) || !at.Before(intentEnd) || window.Before(start) || updated.After(at) || window.After(at) || !at.Equal(authorized) || end.After(intentEnd) || end.After(grantEnd) || end.After(costEnd) || attempted.Before(authorized) {
		return fail("INVALID_TIME_BINDING")
	}
	if r.SideEffectState == "PERFORMED" && !attempted.Before(end) {
		return fail("EXPIRED_AUTHORIZATION")
	}
	// Unix seconds avoid time.Duration multiplication overflow for large windows.
	elapsed := at.Unix() - window.Unix()
	if at.Nanosecond() < window.Nanosecond() {
		elapsed--
	}
	inWindow := l.ExecutionsInWindow
	if elapsed >= int64(g.WindowSeconds) {
		inWindow = 0
	}
	if gate.ExecutionsTotalBefore != l.ExecutionsTotal || gate.ExecutionsInWindowBefore != inWindow || gate.CostMinorTotalBefore != l.CostMinorTotal || gate.PendingOutcomesBefore != l.PendingOutcomes {
		return fail("LEDGER_SNAPSHOT_MISMATCH")
	}
	if l.ReconciliationRequired || containsFold(l.SuccessfulIdempotencyKeys, i.IdempotencyKey) || containsFold(l.PendingExecutionIDs, r.ExecutionID) {
		return fail("LEDGER_BLOCKED")
	}
	for _, link := range l.OutcomeLinks {
		if link.ExecutionID == r.ExecutionID {
			return fail("LEDGER_BLOCKED")
		}
	}
	// Nonnegative amounts are schema-checked; subtract only after checking total.
	if l.ExecutionsTotal >= g.MaxExecutionsTotal || inWindow >= g.MaxExecutionsPerWindow || l.PendingOutcomes >= g.MaxPendingOutcomes || l.CostMinorTotal > g.MaxCostMinorTotal || c.MaxCostMinor > g.MaxCostMinorTotal-l.CostMinorTotal {
		return fail("BUDGET_EXCEEDED")
	}
	return M10ChainSummary{Result: "CONSISTENT_UNVERIFIED", IntentID: i.IntentID, GrantID: g.GrantID, ExecutionID: r.ExecutionID}, missionValid
}

func runM10ChainCheck(w io.Writer, args []string) error {
	if len(args) != 9 {
		return fmt.Errorf("usage: demo m10-chain-check INTENT POLICY GRANT APPROVAL COST PRE_GATE_LEDGER GATE AUTHORIZATION EXECUTION")
	}
	raw := make([][]byte, 9)
	for n, path := range args {
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		raw[n] = b
	}
	s, status := CheckM10Chain(raw)
	if status != missionValid {
		return fmt.Errorf("M10 chain audit: %s", status)
	}
	return json.NewEncoder(w).Encode(s)
}
