package m10

import (
	"encoding/json"
	"testing"
	"time"
)

func m10Entry(t *testing.T, kind string, value any) ArtifactEntry {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := NewArtifactEntry(kind, raw)
	if err != nil {
		t.Fatal(err)
	}
	return entry
}

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

func TestArtifactRegistryEntryCanonicalizesAndRejectsTamper(t *testing.T) {
	bound := TrustedCostBound{CostBoundID: "c", IntentID: "i", IntentHash: "sha256:0000000000000000000000000000000000000000000000000000000000000000", MaxCostMinor: 100, Currency: "USD", SourceRef: "fixture:registry", ObservedAt: "2026-09-08T00:00:00Z", ExpiresAt: "2026-09-08T01:00:00Z", CorrelationID: "x", HashVersion: "go-json-v1"}
	bound.CostBoundHash = ComputeTrustedCostBoundHash(bound)
	raw, err := json.Marshal(bound)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := NewArtifactEntry(ArtifactKindTrustedCostBound, raw)
	if err != nil || entry.ArtifactID != bound.CostBoundID || entry.ContentHash == "" {
		t.Fatal(err, entry)
	}
	entryRaw, _ := json.Marshal(entry)
	if _, err := ValidateArtifactEntry(entryRaw); err != nil {
		t.Fatal(err)
	}
	entry.ContentHash = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	entryRaw, _ = json.Marshal(entry)
	if _, err := ValidateArtifactEntry(entryRaw); err == nil {
		t.Fatal("accepted tampered artifact registry entry")
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

func TestCanaryGateRejectsNearInt64CostBoundaryWithoutOverflow(t *testing.T) {
	const maxInt64 = int64(^uint64(0) >> 1)
	grant := CanaryGrant{GrantID: "max-cost-g", GrantVersion: "v1", PolicyVersion: "p1", ApprovalRef: "approval-1", ApprovedBy: "human", ApproverID: "learner", ApprovedAt: "2026-09-08T00:00:00Z", ValidFrom: "2026-09-08T00:00:00Z", ExpiresAt: "2026-09-08T02:00:00Z", AllowedRiskClasses: []string{"RISK0"}, AllowedActionTypes: []string{"DRAFT"}, AllowedHosts: []string{"example.com"}, ExecutorIDs: []string{"local_sandbox"}, MaxExecutionsTotal: 2, MaxExecutionsPerWindow: 2, WindowSeconds: 60, MaxCostMinorTotal: maxInt64, Currency: "USD", MaxPendingOutcomes: 2, KillSwitchRequired: true, CorrelationID: "corr", HashVersion: "go-json-v1"}
	grant.GrantHash = ComputeCanaryGrantHash(grant)
	bound := TrustedCostBound{CostBoundID: "max-cost-bound", IntentID: "intent", IntentHash: "sha256:0000000000000000000000000000000000000000000000000000000000000000", MaxCostMinor: 2, Currency: "USD", SourceRef: "fixture:cost", ObservedAt: "2026-09-08T00:00:00Z", ExpiresAt: "2026-09-08T01:30:00Z", CorrelationID: "corr", HashVersion: "go-json-v1"}
	bound.CostBoundHash = ComputeTrustedCostBoundHash(bound)
	gate := EvaluateCanaryGate(CanaryGateInput{Grant: grant, CostBound: bound, IntentID: "intent", IntentHash: bound.IntentHash, PolicyVersion: "p1", PolicyDecision: "ALLOW", RiskClass: "RISK0", ApprovalID: "approval-1", ApproverID: "learner", CorrelationID: "corr", ActionType: "DRAFT", Target: "https://example.com/draft", Now: "2026-09-08T01:00:00Z", Ledger: CanaryLedgerSnapshot{CostMinorTotal: maxInt64 - 1}})
	if gate.Decision != "REQUIRE_APPROVAL" || gate.Reason != "CANARY_COST_BUDGET_EXHAUSTED" || gate.ExecutionAuthorized {
		t.Fatalf("near-maximum cost boundary was not denied without overflow: %+v", gate)
	}
}

func TestCanaryAuthorizationBindsGateWithoutExecuting(t *testing.T) {
	g := CanaryGrant{GrantID: "g", GrantVersion: "v1", PolicyVersion: "p1", ApprovalRef: "approval-1", ApprovedBy: "human", ApproverID: "learner", ApprovedAt: "2026-09-08T00:00:00Z", ValidFrom: "2026-09-08T00:00:00Z", ExpiresAt: "2026-09-08T02:00:00Z", AllowedRiskClasses: []string{"RISK0"}, AllowedActionTypes: []string{"DRAFT"}, AllowedHosts: []string{"example.com"}, ExecutorIDs: []string{"local_sandbox"}, MaxExecutionsTotal: 1, MaxExecutionsPerWindow: 1, WindowSeconds: 60, MaxCostMinorTotal: 100, Currency: "USD", MaxPendingOutcomes: 1, KillSwitchRequired: true, CorrelationID: "corr", HashVersion: "go-json-v1"}
	g.GrantHash = ComputeCanaryGrantHash(g)
	cost := TrustedCostBound{CostBoundID: "cost", IntentID: "intent", IntentHash: "sha256:0000000000000000000000000000000000000000000000000000000000000000", MaxCostMinor: 100, Currency: "USD", SourceRef: "fixture:cost", ObservedAt: "2026-09-08T00:00:00Z", ExpiresAt: "2026-09-08T01:30:00Z", CorrelationID: "corr", HashVersion: "go-json-v1"}
	cost.CostBoundHash = ComputeTrustedCostBoundHash(cost)
	gateInput := CanaryGateInput{Grant: g, CostBound: cost, IntentID: "intent", IntentHash: cost.IntentHash, PolicyVersion: "p1", PolicyDecision: "ALLOW", RiskClass: "RISK0", ApprovalID: "approval-1", ApproverID: "learner", CorrelationID: "corr", ActionType: "DRAFT", Target: "https://example.com/draft", Now: "2026-09-08T01:00:00Z"}
	gate := EvaluateCanaryGate(gateInput)
	auth, err := AuthorizeCanary(CanaryAuthorizationInput{Gate: gate, Grant: g, CostBound: cost, IntentID: "intent", IntentHash: cost.IntentHash, PolicyVersion: "p1", IdempotencyKey: "key", CorrelationID: "corr", IntentExpiresAt: "2026-09-08T01:45:00Z", ExecutorID: "local_sandbox", AuthorizedAt: "2026-09-08T01:00:00Z"})
	if err != nil || !auth.ExecutionAuthorized || auth.ExecutionMode != "GOVERNED_CANARY" || auth.CanaryGateID != gate.GateID {
		t.Fatal(err, auth)
	}
	raw, _ := json.Marshal(auth)
	if _, err := ValidateExecutionAuthorization(raw); err != nil {
		t.Fatal(err)
	}
	entries := []ArtifactEntry{m10Entry(t, ArtifactKindCanaryGrant, g), m10Entry(t, ArtifactKindTrustedCostBound, cost), m10Entry(t, ArtifactKindCanaryGate, gate), m10Entry(t, ArtifactKindExecutionAuthorization, auth)}
	if err := ValidateArtifactGraph(entries); err != nil {
		t.Fatalf("valid M10 graph rejected: %v", err)
	}
	forgedGate := gate
	forgedGate.GateID = "forged-gate-id"
	forgedGateAuthorization := auth
	forgedGateAuthorization.CanaryGateID = forgedGate.GateID
	forgedGateAuthorization.AuthorizationID = authorizationID(CanaryAuthorizationInput{
		Gate: forgedGate, IntentID: forgedGateAuthorization.IntentID, IntentHash: forgedGateAuthorization.IntentHash,
		ExecutorID: forgedGateAuthorization.ExecutorID, AuthorizedAt: forgedGateAuthorization.AuthorizedAt,
		IdempotencyKey: forgedGateAuthorization.IdempotencyKey,
	})
	forgedGateEntries := []ArtifactEntry{
		m10Entry(t, ArtifactKindCanaryGrant, g),
		m10Entry(t, ArtifactKindTrustedCostBound, cost),
		m10Entry(t, ArtifactKindCanaryGate, forgedGate),
		m10Entry(t, ArtifactKindExecutionAuthorization, forgedGateAuthorization),
	}
	if err := ValidateArtifactGraph(forgedGateEntries); err == nil {
		t.Fatal("graph accepted a forged canary gate ID with otherwise matching links")
	}
	forgedAuthorization := auth
	forgedAuthorization.AuthorizationID = "forged-authorization-id"
	forgedAuthorizationEntries := []ArtifactEntry{
		m10Entry(t, ArtifactKindCanaryGrant, g),
		m10Entry(t, ArtifactKindTrustedCostBound, cost),
		m10Entry(t, ArtifactKindCanaryGate, gate),
		m10Entry(t, ArtifactKindExecutionAuthorization, forgedAuthorization),
	}
	if err := ValidateArtifactGraph(forgedAuthorizationEntries); err == nil {
		t.Fatal("graph accepted a forged authorization ID with otherwise matching links")
	}
	if err := ValidateArtifactGraph(append(entries, entries[0])); err == nil {
		t.Fatal("duplicate immutable grant entry was accepted by canonical M10 graph")
	}
	orphan := auth
	orphan.CanaryGateID = "missing-gate"
	if err := ValidateArtifactGraph(append(entries[:3:3], m10Entry(t, ArtifactKindExecutionAuthorization, orphan))); err == nil {
		t.Fatal("orphan authorization was accepted by canonical M10 graph")
	}
	gate.Decision = "DENY"
	if _, err := AuthorizeCanary(CanaryAuthorizationInput{Gate: gate, Grant: g, CostBound: cost, IntentID: "intent", IntentHash: cost.IntentHash, PolicyVersion: "p1", IdempotencyKey: "key", CorrelationID: "corr", IntentExpiresAt: "2026-09-08T01:45:00Z", ExecutorID: "local_sandbox", AuthorizedAt: "2026-09-08T01:00:00Z"}); err == nil {
		t.Fatal("denied gate authorized execution")
	}
}

func TestCancelledCanaryExecutionRecordHasNoSideEffect(t *testing.T) {
	grant := CanaryGrant{GrantID: "g", GrantVersion: "v1", PolicyVersion: "p1", ApprovalRef: "approval-1", ApprovedBy: "human", ApproverID: "learner", ApprovedAt: "2026-09-08T00:00:00Z", ValidFrom: "2026-09-08T00:00:00Z", ExpiresAt: "2026-09-08T02:00:00Z", AllowedRiskClasses: []string{"RISK0"}, AllowedActionTypes: []string{"DRAFT"}, AllowedHosts: []string{"example.com"}, ExecutorIDs: []string{"local_sandbox"}, MaxExecutionsTotal: 1, MaxExecutionsPerWindow: 1, WindowSeconds: 60, MaxCostMinorTotal: 100, Currency: "USD", MaxPendingOutcomes: 1, KillSwitchRequired: true, CorrelationID: "corr", HashVersion: "go-json-v1"}
	grant.GrantHash = ComputeCanaryGrantHash(grant)
	cost := TrustedCostBound{CostBoundID: "cost", IntentID: "intent", IntentHash: "sha256:0000000000000000000000000000000000000000000000000000000000000000", MaxCostMinor: 100, Currency: "USD", SourceRef: "fixture:cost", ObservedAt: "2026-09-08T00:00:00Z", ExpiresAt: "2026-09-08T01:30:00Z", CorrelationID: "corr", HashVersion: "go-json-v1"}
	cost.CostBoundHash = ComputeTrustedCostBoundHash(cost)
	gate := EvaluateCanaryGate(CanaryGateInput{Grant: grant, CostBound: cost, IntentID: "intent", IntentHash: cost.IntentHash, PolicyVersion: "p1", PolicyDecision: "ALLOW", RiskClass: "RISK0", ApprovalID: "approval-1", ApproverID: "learner", CorrelationID: "corr", ActionType: "DRAFT", Target: "https://example.com/draft", Now: "2026-09-08T01:00:00Z"})
	authorization, err := AuthorizeCanary(CanaryAuthorizationInput{Gate: gate, Grant: grant, CostBound: cost, IntentID: "intent", IntentHash: cost.IntentHash, PolicyVersion: "p1", IdempotencyKey: "key", CorrelationID: "corr", IntentExpiresAt: "2026-09-08T01:45:00Z", ExecutorID: "local_sandbox", AuthorizedAt: "2026-09-08T01:00:00Z"})
	if err != nil {
		t.Fatal(err)
	}
	record, err := CancelCanaryExecution(CancelledExecutionInput{Authorization: authorization, AttemptedAt: "2026-09-08T01:01:00Z", Reason: "learner cancelled before executor"})
	if err != nil || record.Status != "CANCELLED" || record.SideEffectState != "NOT_PERFORMED" || record.ExecutionID == "" {
		t.Fatal(err, record)
	}
	raw, _ := json.Marshal(record)
	if _, err := ValidateExecutionRecord(raw); err != nil {
		t.Fatal(err)
	}
	failed, err := FailCanaryExecutionFixture(FailedExecutionInput{Authorization: authorization, AttemptedAt: "2026-09-08T01:02:00Z", Reason: "fixture executor unavailable before dispatch"})
	if err != nil || failed.Status != "FAILED" || failed.SideEffectState != "NOT_PERFORMED" || failed.ExecutionID == record.ExecutionID {
		t.Fatal(err, failed)
	}
	raw, _ = json.Marshal(failed)
	if _, err := ValidateExecutionRecord(raw); err != nil {
		t.Fatal(err)
	}
	entries := []ArtifactEntry{m10Entry(t, ArtifactKindCanaryGrant, grant), m10Entry(t, ArtifactKindTrustedCostBound, cost), m10Entry(t, ArtifactKindCanaryGate, gate), m10Entry(t, ArtifactKindExecutionAuthorization, authorization), m10Entry(t, ArtifactKindExecutionRecord, record)}
	if err := ValidateArtifactGraph(entries); err != nil {
		t.Fatalf("valid terminal execution graph rejected: %v", err)
	}
	forgedRecord := record
	forgedRecord.ExecutionID = "forged-execution-id"
	forgedRecordEntries := append(entries[:4:4], m10Entry(t, ArtifactKindExecutionRecord, forgedRecord))
	if err := ValidateArtifactGraph(forgedRecordEntries); err == nil {
		t.Fatal("graph accepted a forged execution ID with otherwise matching links")
	}
	secondRecord := record
	secondRecord.ExecutionID = "canary-exec-second-terminal"
	if err := ValidateArtifactGraph(append(entries, m10Entry(t, ArtifactKindExecutionRecord, secondRecord))); err == nil {
		t.Fatal("two terminal execution records for one authorization were accepted")
	}
	if _, err := FailCanaryExecutionFixture(FailedExecutionInput{Authorization: authorization, AttemptedAt: authorization.ExpiresAt, Reason: "fixture attempt after authorization expiry"}); err == nil {
		t.Fatal("accepted fixture execution at authorization expiry")
	}
	record.Status = "SUCCEEDED"
	raw, _ = json.Marshal(record)
	if _, err := ValidateExecutionRecord(raw); err == nil {
		t.Fatal("accepted non-cancelled learner execution record")
	}
}
