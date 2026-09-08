package m10

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTrustedCostBoundDecodeAndBinding(t *testing.T) {
	c := TrustedCostBound{CostBoundID: "c", IntentID: "i", IntentHash: "sha256:0000000000000000000000000000000000000000000000000000000000000000", MaxCostMinor: 9007199254740993, Currency: "USD", SourceRef: "fixture:registry", ObservedAt: "2026-09-08T00:00:00Z", ExpiresAt: "2026-09-08T01:00:00Z", CorrelationID: "x", HashVersion: "go-json-v1"}
	c.CostBoundHash = ComputeTrustedCostBoundHash(c)
	raw, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	decoded, status := DecodeTrustedCostBound(raw)
	if status != "VALID" || decoded.MaxCostMinor != c.MaxCostMinor {
		t.Fatalf("decode: %s %+v", status, decoded)
	}
	if ValidFor(decoded, "i", c.IntentHash, "x", "USD", time.Date(2026, 9, 8, 0, 30, 0, 0, time.UTC)) != "VALID" {
		t.Fatal("valid bound rejected")
	}
	c.MaxCostMinor = 1
	raw, _ = json.Marshal(c)
	if _, status := DecodeTrustedCostBound(raw); status != "TAMPERED_COST_BOUND" {
		t.Fatal(status)
	}
}

func TestCanaryGrantDecodeAndBinding(t *testing.T) {
	g := CanaryGrant{
		GrantID: "g", GrantVersion: "v1", PolicyVersion: "p1", ApprovalRef: "approval-1", ApprovedBy: "human", ApproverID: "learner",
		ApprovedAt: "2026-09-08T00:00:00Z", ValidFrom: "2026-09-08T00:00:00Z", ExpiresAt: "2026-09-08T02:00:00Z",
		AllowedRiskClasses: []string{"RISK0"}, AllowedActionTypes: []string{"DRAFT"}, AllowedHosts: []string{"example.com"}, ExecutorIDs: []string{"local_sandbox"},
		MaxExecutionsTotal: 1, MaxExecutionsPerWindow: 1, WindowSeconds: 60, MaxCostMinorTotal: 100, Currency: "USD", MaxPendingOutcomes: 1,
		KillSwitchRequired: true, CorrelationID: "corr", HashVersion: "go-json-v1",
	}
	g.GrantHash = ComputeCanaryGrantHash(g)
	raw, err := json.Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	decoded, status := DecodeCanaryGrant(raw)
	if status != "VALID" || decoded.GrantHash != g.GrantHash {
		t.Fatalf("decode: %s %+v", status, decoded)
	}
	now := time.Date(2026, 9, 8, 1, 0, 0, 0, time.UTC)
	if status := ValidCanaryGrantFor(decoded, "intent", "sha256:0000000000000000000000000000000000000000000000000000000000000000", "p1", "approval-1", "learner", "corr", "RISK0", "DRAFT", "https://example.com/draft", now); status != "VALID" {
		t.Fatal(status)
	}
	g.AllowedHosts = []string{"other.invalid"}
	raw, _ = json.Marshal(g)
	if _, status := DecodeCanaryGrant(raw); status != "TAMPERED_GRANT" {
		t.Fatal(status)
	}
}

func TestCanaryGateIsNonAuthorizingAndBounded(t *testing.T) {
	g := CanaryGrant{GrantID: "g", GrantVersion: "v1", PolicyVersion: "p1", ApprovalRef: "approval-1", ApprovedBy: "human", ApproverID: "learner", ApprovedAt: "2026-09-08T00:00:00Z", ValidFrom: "2026-09-08T00:00:00Z", ExpiresAt: "2026-09-08T02:00:00Z", AllowedRiskClasses: []string{"RISK0"}, AllowedActionTypes: []string{"DRAFT"}, AllowedHosts: []string{"example.com"}, ExecutorIDs: []string{"local_sandbox"}, MaxExecutionsTotal: 1, MaxExecutionsPerWindow: 1, WindowSeconds: 60, MaxCostMinorTotal: 100, Currency: "USD", MaxPendingOutcomes: 1, KillSwitchRequired: true, CorrelationID: "corr", HashVersion: "go-json-v1"}
	g.GrantHash = ComputeCanaryGrantHash(g)
	cost := TrustedCostBound{CostBoundID: "cost", IntentID: "intent", IntentHash: "sha256:0000000000000000000000000000000000000000000000000000000000000000", MaxCostMinor: 100, Currency: "USD", SourceRef: "fixture:cost", ObservedAt: "2026-09-08T00:00:00Z", ExpiresAt: "2026-09-08T01:30:00Z", CorrelationID: "corr", HashVersion: "go-json-v1"}
	cost.CostBoundHash = ComputeTrustedCostBoundHash(cost)
	in := CanaryGateInput{Grant: g, CostBound: cost, IntentID: "intent", IntentHash: cost.IntentHash, PolicyVersion: "p1", PolicyDecision: "ALLOW", RiskClass: "RISK0", ApprovalID: "approval-1", ApproverID: "learner", CorrelationID: "corr", ActionType: "DRAFT", Target: "https://example.com/draft", Now: "2026-09-08T01:00:00Z"}
	gate := EvaluateCanaryGate(in)
	if gate.Decision != "ALLOW_CANARY" || gate.ExecutionAuthorized || gate.PerActionApprovalRequired {
		t.Fatal(gate)
	}
	raw, _ := json.Marshal(gate)
	if _, err := ValidateCanaryGateDecision(raw); err != nil {
		t.Fatal(err)
	}
	in.Ledger.ExecutionsTotal = 1
	if denied := EvaluateCanaryGate(in); denied.Decision != "REQUIRE_APPROVAL" || denied.Reason != "CANARY_TOTAL_BUDGET_EXHAUSTED" || denied.ExecutionAuthorized {
		t.Fatal(denied)
	}
}
