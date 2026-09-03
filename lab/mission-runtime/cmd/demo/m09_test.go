package main

import (
	"os"
	"path/filepath"
	"testing"
)

type m09Case struct {
	CaseID   string `json:"case_id"`
	Mutation string `json:"mutation"`
	Expected string `json:"expected"`
}

func baseM09() (M09State, M09Context) {
	i := SealShadowActionIntent(ShadowActionIntent{
		IntentID: "i", DecisionID: "d", EvidenceIDs: []string{"e"}, ActionType: "UPDATE_DRAFT",
		Target: "https://example.com/draft", Parameters: map[string]any{"x": 1}, ProposedBy: "human",
		CreatedAt: "2026-09-03T07:00:00Z", ExpiresAt: "2026-09-03T09:00:00Z", CorrelationID: "c", IdempotencyKey: "k",
	})
	policyCtx := ShadowPolicyContext{
		PolicyVersion: "v1", Now: "2026-09-03T07:05:00Z",
		KnownDecisionIDs: []string{"d"}, KnownEvidenceIDs: []string{"e"}, AllowedHosts: []string{"example.com"},
		ActionRisk: map[string]string{"UPDATE_DRAFT": "RISK1"}, SeenIdempotency: map[string]string{},
	}
	p := EvaluateShadowPolicy(i, policyCtx)
	a := ApprovalRecord{
		ApprovalID: "a", IntentID: i.IntentID, IntentHash: i.IntentHash, PolicyVersion: "v1", Decision: "APPROVE",
		ApprovedBy: "human", ApproverID: "u", ApprovedAt: "2026-09-03T07:10:00Z", ExpiresAt: "2026-09-03T08:30:00Z",
		CorrelationID: "c", OneTime: true,
	}
	return M09State{Intent: i, Policy: p, Approval: &a}, M09Context{
		Now: "2026-09-03T08:00:00Z",
		Executor: ExecutorProfile{ExecutorID: "local_sandbox", AllowedActionTypes: []string{"UPDATE_DRAFT"}, AllowedHosts: []string{"example.com"}},
		AllowedExecutorIDs: []string{"local_sandbox"}, PolicyContext: policyCtx,
	}
}

func TestM09EvalPack(t *testing.T) {
	cases := loadCases[m09Case](t, "M09-approval-execution")
	for _, c := range cases {
		t.Run(c.CaseID, func(t *testing.T) {
			s, ctx := baseM09()
			switch c.Mutation {
			case "no_approval":
				s.Approval = nil
			case "rejected":
				s.Approval.Decision = "REJECT"
			case "machine_approver":
				s.Approval.ApprovedBy = "agent"
			case "approval_hash":
				s.Approval.IntentHash = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
			case "approval_policy":
				s.Approval.PolicyVersion = "v0"
			case "expired_approval":
				s.Approval.ExpiresAt = "2026-09-03T07:30:00Z"
			case "kill_switch":
				ctx.KillSwitch = true
			case "wrong_executor":
				ctx.Executor.ExecutorID = "other"
			case "bad_target":
				s.Intent.Target = "https://evil.example/draft"
				s.Intent.IntentHash = ComputeShadowIntentHash(s.Intent)
				s.Policy.IntentHash = s.Intent.IntentHash
				s.Approval.IntentHash = s.Intent.IntentHash
			case "policy_deny":
				s.Policy.Decision = "DENY"
			case "already_executed":
				s.SucceededIdempotency = map[string]bool{"k": true}
			case "tampered_intent":
				s.Intent.Target = "https://evil.example/draft"
			case "expired_intent":
				ctx.Now = "2026-09-03T09:30:00Z"
			case "approval_before_policy":
				s.Approval.ApprovedAt = "2026-09-03T07:00:00Z"
			}
			_, got := AuthorizeM09(s, ctx)
			if got != c.Expected {
				t.Fatalf("expected %s got %s", c.Expected, got)
			}
		})
	}
}

func TestM09DurableResumeRevalidates(t *testing.T) {
	s, ctx := baseM09()
	p := filepath.Join(t.TempDir(), "state.json")
	if err := PersistM09State(p, s); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadM09State(p)
	if err != nil {
		t.Fatal(err)
	}
	loaded.Policy.PolicyVersion = "v2"
	_, got := AuthorizeM09(loaded, ctx)
	if got != "DENY_APPROVAL_MISMATCH" && got != "DENY_POLICY_REVALIDATION" {
		t.Fatalf("resume must revalidate policy/approval binding, got %s", got)
	}
}

func TestM09ControlledExecutorConsumesApprovalAndIdempotency(t *testing.T) {
	s, ctx := baseM09()
	auth, status := AuthorizeM09(s, ctx)
	if status != "AUTHORIZED" {
		t.Fatal(status)
	}
	s.Authorization = &auth
	dir := t.TempDir()
	r, status := ExecuteLocalSandbox(&s, auth, ctx, dir)
	if status != "EXECUTED" || r.SideEffectState != "PERFORMED" || r.Status != "SUCCEEDED" {
		t.Fatalf("unexpected execution %+v %s", r, status)
	}
	if !s.ConsumedApprovalIDs[auth.ApprovalID] || !s.SucceededIdempotency[auth.IdempotencyKey] {
		t.Fatal("successful execution must durably consume approval and idempotency key in canonical state")
	}
	if _, status = AuthorizeM09(s, ctx); status != "WAIT_ALREADY_EXECUTED" {
		t.Fatalf("duplicate authorization must not be issued: %s", status)
	}
}

func TestM09DurableSuccessBlocksAfterRestart(t *testing.T) {
	s, ctx := baseM09()
	auth, status := AuthorizeM09(s, ctx)
	if status != "AUTHORIZED" {
		t.Fatal(status)
	}
	s.Authorization = &auth
	if _, status = ExecuteLocalSandbox(&s, auth, ctx, t.TempDir()); status != "EXECUTED" {
		t.Fatal(status)
	}
	p := filepath.Join(t.TempDir(), "state.json")
	if err := PersistM09State(p, s); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadM09State(p)
	if err != nil {
		t.Fatal(err)
	}
	if _, status = AuthorizeM09(loaded, ctx); status != "WAIT_ALREADY_EXECUTED" {
		t.Fatalf("restart must preserve successful idempotency guard: %s", status)
	}
}

func TestM09ExecutorRevalidatesAuthorizationBinding(t *testing.T) {
	s, ctx := baseM09()
	auth, status := AuthorizeM09(s, ctx)
	if status != "AUTHORIZED" {
		t.Fatal(status)
	}
	s.Authorization = &auth
	tampered := auth
	tampered.PolicyVersion = "other-policy"
	if _, status = ExecuteLocalSandbox(&s, tampered, ctx, t.TempDir()); status != "DENY_AUTHORIZATION" {
		t.Fatalf("executor must reject tampered authorization: %s", status)
	}
}

func TestM09KillSwitchRecheckedAtExecution(t *testing.T) {
	s, ctx := baseM09()
	auth, status := AuthorizeM09(s, ctx)
	if status != "AUTHORIZED" {
		t.Fatal(status)
	}
	s.Authorization = &auth
	ctx.KillSwitch = true
	if _, status = ExecuteLocalSandbox(&s, auth, ctx, t.TempDir()); status != "DENY_KILL_SWITCH" {
		t.Fatalf("kill switch must stop execution: %s", status)
	}
}

func TestM09ExecutorProfileCannotSelfAuthorize(t *testing.T) {
	s, ctx := baseM09()
	ctx.Executor.ExecutorID = "other"
	ctx.Executor.AllowedActionTypes = []string{"UPDATE_DRAFT"}
	ctx.Executor.AllowedHosts = []string{"example.com"}
	if _, status := AuthorizeM09(s, ctx); status != "DENY_EXECUTOR" {
		t.Fatalf("unregistered executor must be denied: %s", status)
	}
}

func TestM09UnknownExistingSideEffectRequiresReconciliation(t *testing.T) {
	s, ctx := baseM09()
	auth, status := AuthorizeM09(s, ctx)
	if status != "AUTHORIZED" {
		t.Fatal(status)
	}
	s.Authorization = &auth
	dir := t.TempDir()
	marker := sandboxIdempotencyPath(dir, auth.IdempotencyKey)
	if err := os.WriteFile(marker, []byte("pre-existing external effect marker"), 0600); err != nil {
		t.Fatal(err)
	}
	r, status := ExecuteLocalSandbox(&s, auth, ctx, dir)
	if status != "WAIT_RECONCILIATION" || r.Status != "RECONCILIATION_REQUIRED" || r.SideEffectState != "UNKNOWN" {
		t.Fatalf("uncertain prior effect must stop for reconciliation: %+v %s", r, status)
	}
	if _, status = AuthorizeM09(s, ctx); status != "WAIT_RECONCILIATION" {
		t.Fatalf("reconciliation state must block reauthorization: %s", status)
	}
}
