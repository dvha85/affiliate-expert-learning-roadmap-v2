package main

import (
	"os"
	"reflect"
	"testing"
)

type m10Case struct {
	CaseID           string `json:"case_id"`
	Mutation         string `json:"mutation"`
	ExpectedDecision string `json:"expected_decision"`
	ExpectedReason   string `json:"expected_reason"`
}

func baseM10() (M10State, M10Context) {
	intent := SealShadowActionIntent(ShadowActionIntent{
		IntentID: "i", DecisionID: "d", EvidenceIDs: []string{"e"}, ActionType: "PREPARE_LOCAL_DRAFT",
		Target: "https://example.com/draft", Parameters: map[string]any{"title": "draft"}, ProposedBy: "human",
		CreatedAt: "2026-09-03T07:00:00Z", ExpiresAt: "2026-09-03T11:00:00Z", CorrelationID: "c", IdempotencyKey: "k",
	})
	policyCtx := ShadowPolicyContext{
		PolicyVersion: "v1", Now: "2026-09-03T07:05:00Z", KnownDecisionIDs: []string{"d"}, KnownEvidenceIDs: []string{"e"},
		AllowedHosts: []string{"example.com", "other.example"}, ActionRisk: map[string]string{
			"PREPARE_LOCAL_DRAFT": "RISK0", "UPDATE_DRAFT": "RISK1", "PUBLISH_CONTENT": "RISK2", "UPDATE_PROFILE": "RISK0",
		}, SeenIdempotency: map[string]string{},
	}
	policy := EvaluateShadowPolicy(intent, policyCtx)
	grant := SealCanaryGrant(CanaryGrant{
		GrantID: "g", GrantVersion: "gv1", PolicyVersion: "v1", ApprovalRef: "human-grant-approval",
		ApprovedBy: "human", ApproverID: "u", ApprovedAt: "2026-09-03T07:05:00Z", ValidFrom: "2026-09-03T07:10:00Z", ExpiresAt: "2026-09-03T10:00:00Z",
		AllowedRiskClasses: []string{"RISK0", "RISK1"}, AllowedActionTypes: []string{"PREPARE_LOCAL_DRAFT", "UPDATE_DRAFT"},
		AllowedHosts: []string{"example.com"}, ExecutorIDs: []string{"local_sandbox"}, MaxExecutionsTotal: 3, MaxExecutionsPerWindow: 2,
		WindowSeconds: 3600, MaxCostMinorTotal: 1000, Currency: "USD", MaxPendingOutcomes: 1, KillSwitchRequired: true, CorrelationID: "gc",
	})
	cost := SealCanaryCostBound(CanaryCostBound{
		CostBoundID: "cost-i", IntentID: intent.IntentID, IntentHash: intent.IntentHash, MaxCostMinor: 100, Currency: "USD",
		SourceRef: "deterministic-cost-registry", ObservedAt: "2026-09-03T07:15:00Z", ExpiresAt: "2026-09-03T09:00:00Z", CorrelationID: intent.CorrelationID,
	})
	ledger := CanaryLedger{GrantID: grant.GrantID, GrantVersion: grant.GrantVersion, GrantHash: grant.GrantHash, WindowStartedAt: "2026-09-03T07:10:00Z", UpdatedAt: "2026-09-03T07:10:00Z"}
	state := M10State{Intent: intent, Policy: policy, Grant: grant, CostBound: cost, Ledger: ledger}
	ctx := M10Context{
		Now: "2026-09-03T08:00:00Z", ApprovedGrantHashes: map[string]string{grant.ApprovalRef: grant.GrantHash},
		TrustedCostBounds: map[string]string{cost.CostBoundID: cost.CostBoundHash},
		Executor: ExecutorProfile{ExecutorID: "local_sandbox", AllowedActionTypes: []string{"PREPARE_LOCAL_DRAFT", "UPDATE_DRAFT", "UPDATE_PROFILE", "PUBLISH_CONTENT"}, AllowedHosts: []string{"example.com", "other.example"}},
		AllowedExecutorIDs: []string{"local_sandbox"}, PolicyContext: policyCtx,
	}
	return state, ctx
}

func refreshM10Intent(state *M10State, ctx *M10Context, actionType, target, intentID, idem string, maxCost int64) {
	state.Intent.IntentID = intentID
	state.Intent.ActionType = actionType
	state.Intent.Target = target
	state.Intent.IdempotencyKey = idem
	state.Intent = SealShadowActionIntent(state.Intent)
	pc := ctx.PolicyContext
	pc.Now = "2026-09-03T07:05:00Z"
	state.Policy = EvaluateShadowPolicy(state.Intent, pc)
	state.CostBound = SealCanaryCostBound(CanaryCostBound{
		CostBoundID: "cost-" + intentID, IntentID: state.Intent.IntentID, IntentHash: state.Intent.IntentHash, MaxCostMinor: maxCost, Currency: "USD",
		SourceRef: "deterministic-cost-registry", ObservedAt: "2026-09-03T07:15:00Z", ExpiresAt: "2026-09-03T09:00:00Z", CorrelationID: state.Intent.CorrelationID,
	})
	ctx.TrustedCostBounds[state.CostBound.CostBoundID] = state.CostBound.CostBoundHash
}

func refreshGrantApproval(ctx *M10Context, grant CanaryGrant) {
	if ctx.ApprovedGrantHashes == nil { ctx.ApprovedGrantHashes = map[string]string{} }
	ctx.ApprovedGrantHashes[grant.ApprovalRef] = grant.GrantHash
}

func TestM10EvalPack(t *testing.T) {
	for _, c := range loadCases[m10Case](t, "M10-governed-canary") {
		t.Run(c.CaseID, func(t *testing.T) {
			state, ctx := baseM10()
			switch c.Mutation {
			case "risk1":
				refreshM10Intent(&state, &ctx, "UPDATE_DRAFT", "https://example.com/draft", "i-r1", "k-r1", 100)
			case "risk2":
				refreshM10Intent(&state, &ctx, "PUBLISH_CONTENT", "https://example.com/post", "i-r2", "k-r2", 100)
			case "risk1_not_delegated":
				refreshM10Intent(&state, &ctx, "UPDATE_DRAFT", "https://example.com/draft", "i-r1", "k-r1", 100)
				state.Grant.AllowedRiskClasses = []string{"RISK0"}; state.Grant = SealCanaryGrant(state.Grant); state.Ledger.GrantHash = state.Grant.GrantHash; refreshGrantApproval(&ctx, state.Grant)
			case "scope_action":
				refreshM10Intent(&state, &ctx, "UPDATE_PROFILE", "https://example.com/profile", "i-scope", "k-scope", 100)
			case "scope_host":
				refreshM10Intent(&state, &ctx, "PREPARE_LOCAL_DRAFT", "https://other.example/draft", "i-host", "k-host", 100)
			case "expired_grant":
				state.Grant.ExpiresAt = "2026-09-03T07:30:00Z"; state.Grant = SealCanaryGrant(state.Grant); state.Ledger.GrantHash = state.Grant.GrantHash; refreshGrantApproval(&ctx, state.Grant)
			case "revoked_grant": ctx.RevokedGrantIDs = []string{"g"}
			case "kill_switch": ctx.KillSwitch = true
			case "grant_policy":
				state.Grant.PolicyVersion = "v0"; state.Grant = SealCanaryGrant(state.Grant); state.Ledger.GrantHash = state.Grant.GrantHash; refreshGrantApproval(&ctx, state.Grant)
			case "tampered_grant": state.Grant.MaxExecutionsTotal = 99
			case "missing_grant_provenance": delete(ctx.ApprovedGrantHashes, state.Grant.ApprovalRef)
			case "wrong_executor": ctx.Executor.ExecutorID = "other"; ctx.AllowedExecutorIDs = []string{"local_sandbox", "other"}
			case "total_budget": state.Ledger.ExecutionsTotal = state.Grant.MaxExecutionsTotal
			case "rate_limit": state.Ledger.ExecutionsInWindow = state.Grant.MaxExecutionsPerWindow
			case "cost_budget": state.Ledger.CostMinorTotal = 950
			case "pending_outcome": state.Ledger.PendingOutcomes = state.Grant.MaxPendingOutcomes; state.Ledger.PendingExecutionIDs = []string{"exec-old"}
			case "missing_cost_bound": state.CostBound = CanaryCostBound{}
			case "reconciliation": state.Ledger.ReconciliationRequired = true
			case "duplicate": state.Ledger.SuccessfulIdempotencyKeys = []string{"k"}
			case "stale_policy": state.Policy.Reason = "tampered-policy-result"
			case "wildcard_grant":
				state.Grant.AllowedHosts = []string{"*"}; state.Grant = SealCanaryGrant(state.Grant); state.Ledger.GrantHash = state.Grant.GrantHash; refreshGrantApproval(&ctx, state.Grant)
			case "grant_approval_mismatch": ctx.ApprovedGrantHashes[state.Grant.ApprovalRef] = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
			case "governance_unavailable": ctx.ApprovedGrantHashes = nil
			case "tampered_cost_bound": state.CostBound.MaxCostMinor = 1
			case "expired_cost_bound":
				state.CostBound.ExpiresAt = "2026-09-03T07:30:00Z"; state.CostBound = SealCanaryCostBound(state.CostBound); ctx.TrustedCostBounds[state.CostBound.CostBoundID] = state.CostBound.CostBoundHash
			case "ledger_grant_hash": state.Ledger.GrantHash = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
			}
			got := EvaluateCanaryGate(state, ctx)
			if got.Decision != c.ExpectedDecision || got.Reason != c.ExpectedReason { t.Fatalf("expected %s/%s got %s/%s", c.ExpectedDecision, c.ExpectedReason, got.Decision, got.Reason) }
			if got.ExecutionAuthorized { t.Fatal("CanaryGateDecision must never itself authorize execution") }
		})
	}
}

func TestCanaryGrantHashStableAcrossOrdering(t *testing.T) {
	state, _ := baseM10(); a := state.Grant; b := state.Grant
	b.AllowedRiskClasses = []string{"RISK1", "RISK0"}; b.AllowedActionTypes = []string{"UPDATE_DRAFT", "PREPARE_LOCAL_DRAFT"}
	if ComputeCanaryGrantHash(a) != ComputeCanaryGrantHash(b) { t.Fatal("grant hash must be stable across set-like ordering") }
	if !reflect.DeepEqual(a.AllowedActionTypes, []string{"PREPARE_LOCAL_DRAFT", "UPDATE_DRAFT"}) { t.Fatal("hashing must not mutate grant slices") }
}

func TestM10AgentCostHintDoesNotOwnBudget(t *testing.T) {
	state, ctx := baseM10()
	state.Intent.Parameters["estimated_cost_minor"] = 0
	state.Intent.Parameters["currency"] = "USD"
	state.Intent = SealShadowActionIntent(state.Intent)
	pc := ctx.PolicyContext; pc.Now = "2026-09-03T07:05:00Z"; state.Policy = EvaluateShadowPolicy(state.Intent, pc)
	state.CostBound = SealCanaryCostBound(CanaryCostBound{CostBoundID:"cost-new",IntentID:state.Intent.IntentID,IntentHash:state.Intent.IntentHash,MaxCostMinor:100,Currency:"USD",SourceRef:"deterministic-cost-registry",ObservedAt:"2026-09-03T07:15:00Z",ExpiresAt:"2026-09-03T09:00:00Z",CorrelationID:state.Intent.CorrelationID})
	ctx.TrustedCostBounds[state.CostBound.CostBoundID] = state.CostBound.CostBoundHash
	state.Ledger.CostMinorTotal = 950
	gate := EvaluateCanaryGate(state, ctx)
	if gate.Reason != "CANARY_COST_BUDGET_EXHAUSTED" { t.Fatalf("untrusted intent cost hint must not bypass budget: %+v", gate) }
}

func TestM10CanaryExecutionUpdatesLedgerAndOutcomeBinding(t *testing.T) {
	state, ctx := baseM10(); auth, _, status := AuthorizeCanary(state, ctx); if status != "AUTHORIZED" { t.Fatal(status) }
	state.Authorization = &auth; dir := t.TempDir(); rec, status := ExecuteCanaryLocalSandbox(&state, auth, ctx, dir)
	if status != "EXECUTED" || rec.Status != "SUCCEEDED" || rec.SideEffectState != "PERFORMED" { t.Fatalf("unexpected canary execution %+v %s", rec, status) }
	if state.Ledger.ExecutionsTotal != 1 || state.Ledger.CostMinorTotal != 100 || state.Ledger.PendingOutcomes != 1 || len(state.Ledger.PendingExecutionIDs) != 1 { t.Fatalf("ledger not updated: %+v", state.Ledger) }
	if got := RecordCanaryOutcome(&state, dir, "outcome-wrong", "exec-other", "2026-09-03T08:10:00Z"); got != "DENY_OUTCOME_LINK" { t.Fatalf("wrong execution link must not release backpressure: %s", got) }
	if got := RecordCanaryOutcome(&state, dir, "outcome-1", rec.ExecutionID, "2026-09-03T08:10:00Z"); got != "OUTCOME_RECORDED" { t.Fatal(got) }
	if state.Ledger.PendingOutcomes != 0 || len(state.Ledger.OutcomeLinks) != 1 || state.Ledger.OutcomeLinks[0].ExecutionID != rec.ExecutionID { t.Fatalf("outcome linkage not durable: %+v", state.Ledger) }
}

func TestM10RISK2NeverGetsCanaryAuthorization(t *testing.T) {
	state, ctx := baseM10(); refreshM10Intent(&state, &ctx, "PUBLISH_CONTENT", "https://example.com/post", "i-r2", "k-r2", 100)
	_, gate, status := AuthorizeCanary(state, ctx); if status != "REQUIRE_APPROVAL" || gate.Reason != "RISK2_PER_ACTION_APPROVAL_REQUIRED" { t.Fatalf("RISK2 must leave auto path: %+v %s", gate, status) }
}

func TestM10DurableLedgerStopsStaleSecondWorker(t *testing.T) {
	first, ctx := baseM10(); first.Grant.MaxExecutionsTotal = 1; first.Grant.MaxExecutionsPerWindow = 1; first.Grant = SealCanaryGrant(first.Grant); first.Ledger.GrantHash = first.Grant.GrantHash; refreshGrantApproval(&ctx, first.Grant)
	auth1, _, status := AuthorizeCanary(first, ctx); if status != "AUTHORIZED" { t.Fatal(status) }; first.Authorization = &auth1
	dir := t.TempDir(); if _, status = ExecuteCanaryLocalSandbox(&first, auth1, ctx, dir); status != "EXECUTED" { t.Fatal(status) }
	second, _ := baseM10(); second.Grant = first.Grant; second.Ledger.GrantHash = first.Grant.GrantHash; refreshM10Intent(&second, &ctx, "PREPARE_LOCAL_DRAFT", "https://example.com/draft", "i2", "k2", 100)
	auth2, _, status := AuthorizeCanary(second, ctx); if status != "AUTHORIZED" { t.Fatalf("stale worker may have stale authorization before reload; got %s", status) }; second.Authorization = &auth2
	if _, status = ExecuteCanaryLocalSandbox(&second, auth2, ctx, dir); status != "REQUIRE_APPROVAL" { t.Fatalf("executor must reload durable budget before side effect: %s", status) }
}

func TestM10MissingDurableLedgerCannotResetNonZeroBudget(t *testing.T) {
	state, ctx := baseM10(); state.Ledger.ExecutionsTotal = 1; state.Ledger.CostMinorTotal = 100
	auth, _, status := AuthorizeCanary(state, ctx); if status != "AUTHORIZED" { t.Fatal(status) }; state.Authorization = &auth
	if _, status = ExecuteCanaryLocalSandbox(&state, auth, ctx, t.TempDir()); status != "WAIT_LEDGER_MISSING" { t.Fatalf("missing durable ledger must fail closed, got %s", status) }
}

func TestM10UnknownEffectRequiresReconciliation(t *testing.T) {
	state, ctx := baseM10(); auth, _, status := AuthorizeCanary(state, ctx); if status != "AUTHORIZED" { t.Fatal(status) }; state.Authorization = &auth
	dir := t.TempDir(); ledgerPath := canaryLedgerPath(dir, state.Grant.GrantID, state.Grant.GrantVersion); if err := ensureInitialCanaryLedger(ledgerPath, state.Ledger); err != nil { t.Fatal(err) }
	marker := sandboxIdempotencyPath(dir, auth.IdempotencyKey); if err := os.WriteFile(marker, []byte("pre-existing effect marker"), 0600); err != nil { t.Fatal(err) }
	rec, status := ExecuteCanaryLocalSandbox(&state, auth, ctx, dir); if status != "WAIT_RECONCILIATION" || rec.Status != "RECONCILIATION_REQUIRED" || rec.SideEffectState != "UNKNOWN" { t.Fatalf("uncertain effect must stop canary: %+v %s", rec, status) }
	if gate := EvaluateCanaryGate(state, ctx); gate.Decision != "WAIT" || gate.Reason != "RECONCILIATION_REQUIRED" { t.Fatalf("reconciliation must block more canary actions: %+v", gate) }
}
