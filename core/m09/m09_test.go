package m09

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m08"
)

func approvalFixture(t *testing.T) (m08.Intent, m08.PolicyDecision, ApprovalRecord, time.Time) {
	t.Helper()
	intent := m08.SealIntent(m08.Intent{
		IntentID: "i", DecisionID: "d", EvidenceIDs: []string{"e"}, ActionType: "DRAFT", Target: "https://example.com/draft",
		Parameters: map[string]any{}, ProposedBy: "human", CreatedAt: "2026-09-07T00:00:00Z", ExpiresAt: "2026-09-07T03:00:00Z",
		CorrelationID: "c", IdempotencyKey: "k",
	})
	policy := m08.PolicyDecision{PolicyVersion: "v1", IntentID: intent.IntentID, IntentHash: intent.IntentHash, Decision: "HUMAN_REVIEW", RiskClass: "RISK1", Reason: "review", PolicyReviewRequired: true, PolicyMode: "NON_AUTHORIZING", PolicyCheckedAt: "2026-09-07T01:00:00Z"}
	approval := ApprovalRecord{ApprovalID: "a", IntentID: intent.IntentID, IntentHash: intent.IntentHash, PolicyVersion: policy.PolicyVersion, Decision: "APPROVE", ApprovedBy: "human", ApproverID: "reviewer", ApprovedAt: "2026-09-07T01:30:00Z", ExpiresAt: "2026-09-07T02:30:00Z", CorrelationID: intent.CorrelationID, OneTime: true}
	now, err := time.Parse(time.RFC3339, "2026-09-07T02:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	return intent, policy, approval, now
}

func TestValidateApproval(t *testing.T) {
	intent, policy, approval, now := approvalFixture(t)
	if status := ValidateApproval(intent, policy, approval, now); status != Valid {
		t.Fatalf("valid approval: %s", status)
	}
	for name, mutate := range map[string]func(*m08.Intent, *m08.PolicyDecision, *ApprovalRecord){
		"denied policy":      func(_ *m08.Intent, p *m08.PolicyDecision, _ *ApprovalRecord) { p.Decision = "DENY" },
		"authorizing policy": func(_ *m08.Intent, p *m08.PolicyDecision, _ *ApprovalRecord) { p.ExecutionAuthorized = true },
		"allow risk2 policy": func(_ *m08.Intent, p *m08.PolicyDecision, _ *ApprovalRecord) {
			p.Decision, p.RiskClass, p.PolicyReviewRequired = "ALLOW", "RISK2", false
		},
		"approval before policy": func(_ *m08.Intent, _ *m08.PolicyDecision, a *ApprovalRecord) { a.ApprovedAt = "2026-09-07T00:30:00Z" },
		"approval mismatch":      func(_ *m08.Intent, _ *m08.PolicyDecision, a *ApprovalRecord) { a.CorrelationID = "other" },
		"rejected":               func(_ *m08.Intent, _ *m08.PolicyDecision, a *ApprovalRecord) { a.Decision = "REJECT" },
	} {
		t.Run(name, func(t *testing.T) {
			i, p, a := intent, policy, approval
			mutate(&i, &p, &a)
			if status := ValidateApproval(i, p, a, now); status == Valid {
				t.Fatal("invalid approval was accepted")
			}
		})
	}
}

func TestDecodeApprovalRejectsDuplicateAndUnknownFields(t *testing.T) {
	_, _, approval, _ := approvalFixture(t)
	raw, err := json.Marshal(approval)
	if err != nil {
		t.Fatal(err)
	}
	if _, status := DecodeApproval(raw); status != Valid {
		t.Fatalf("valid raw approval: %s", status)
	}
	duplicated := strings.Replace(string(raw), "\"approval_id\":\"a\"", "\"approval_id\":\"a\",\"approval_id\":\"other\"", 1)
	if _, status := DecodeApproval([]byte(duplicated)); status != InvalidSchema {
		t.Fatalf("duplicate approval_id: %s", status)
	}
	unknown := strings.Replace(string(raw), "{", "{\"unexpected\":true,", 1)
	if _, status := DecodeApproval([]byte(unknown)); status != InvalidSchema {
		t.Fatalf("unknown approval field: %s", status)
	}
}

func TestDecodeAuthorizationUsesM09Profile(t *testing.T) {
	raw, err := json.Marshal(ExecutionAuthorization{
		AuthorizationID: "auth-1", IntentID: "intent-1", IntentHash: "sha256:" + strings.Repeat("a", 64),
		PolicyVersion: "policy-1", ApprovalID: "approval-1", ExecutorID: "fixture-executor",
		AuthorizedAt: "2026-09-07T01:45:00Z", ExpiresAt: "2026-09-07T02:30:00Z",
		IdempotencyKey: "idem-1", CorrelationID: "corr-1", ExecutionMode: "APPROVED_LIVE", ExecutionAuthorized: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, status := DecodeAuthorization(raw); status != Valid {
		t.Fatalf("valid M09 authorization: %s", status)
	}

	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	fields["execution_mode"] = "GOVERNED_CANARY"
	canaryRaw, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	if _, status := DecodeAuthorization(canaryRaw); status != InvalidSchema {
		t.Fatalf("incomplete canary authorization: %s", status)
	}
	fields["execution_mode"] = "APPROVED_LIVE"
	fields["unexpected"] = true
	unknownRaw, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	if _, status := DecodeAuthorization(unknownRaw); status != InvalidSchema {
		t.Fatalf("unknown authorization field: %s", status)
	}
}

func TestDecodeExecutionRejectsSchemaAmbiguity(t *testing.T) {
	raw, err := json.Marshal(ExecutionRecord{
		ExecutionID: "exec-1", AuthorizationID: "auth-1", ApprovalID: "approval-1", IntentID: "intent-1",
		IntentHash: "sha256:" + strings.Repeat("a", 64), ExecutorID: "fixture-executor", IdempotencyKey: "idem-1",
		AttemptedAt: "2026-09-07T01:50:00Z", Status: "CANCELLED", SideEffectState: "NOT_PERFORMED",
		Error: "fixture cancellation", CorrelationID: "corr-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, status := DecodeExecution(raw); status != Valid {
		t.Fatalf("valid M09 execution: %s", status)
	}
	duplicated := strings.Replace(string(raw), `"execution_id":"exec-1"`, `"execution_id":"exec-1","execution_id":"other"`, 1)
	if _, status := DecodeExecution([]byte(duplicated)); status != InvalidSchema {
		t.Fatalf("duplicate execution_id: %s", status)
	}
	badSucceeded := strings.Replace(string(raw), `"status":"CANCELLED"`, `"status":"SUCCEEDED"`, 1)
	if _, status := DecodeExecution([]byte(badSucceeded)); status != InvalidSchema {
		t.Fatalf("succeeded without performed side effect: %s", status)
	}
}

func TestValidateHistoricalChainBindsEveryM09Artifact(t *testing.T) {
	intent, policy, approval, _ := approvalFixture(t)
	authorization := ExecutionAuthorization{
		AuthorizationID: "auth-1", IntentID: intent.IntentID, IntentHash: intent.IntentHash,
		PolicyVersion: policy.PolicyVersion, ApprovalID: approval.ApprovalID, ExecutorID: "fixture-executor",
		AuthorizedAt: "2026-09-07T01:40:00Z", ExpiresAt: "2026-09-07T02:30:00Z",
		IdempotencyKey: intent.IdempotencyKey, CorrelationID: intent.CorrelationID,
		ExecutionMode: "APPROVED_LIVE", ExecutionAuthorized: true,
	}
	execution := ExecutionRecord{
		ExecutionID: "exec-1", AuthorizationID: authorization.AuthorizationID, ApprovalID: approval.ApprovalID,
		IntentID: intent.IntentID, IntentHash: intent.IntentHash, ExecutorID: authorization.ExecutorID,
		IdempotencyKey: intent.IdempotencyKey, AttemptedAt: "2026-09-07T02:00:00Z",
		Status: "CANCELLED", SideEffectState: "NOT_PERFORMED", CorrelationID: intent.CorrelationID,
	}
	if status := ValidateHistoricalChain(intent, policy, approval, authorization, execution); status != Valid {
		t.Fatalf("valid historical chain: %s", status)
	}
	wrongExecutor := execution
	wrongExecutor.ExecutorID = "other-executor"
	if status := ValidateHistoricalChain(intent, policy, approval, authorization, wrongExecutor); status != "BROKEN_LINK" {
		t.Fatalf("broken execution link: %s", status)
	}
	performedAfterExpiry := execution
	performedAfterExpiry.AttemptedAt = authorization.ExpiresAt
	performedAfterExpiry.SideEffectState = "PERFORMED"
	if status := ValidateHistoricalChain(intent, policy, approval, authorization, performedAfterExpiry); status != "EXPIRED_AUTHORIZATION" {
		t.Fatalf("performed effect after expiry: %s", status)
	}
}
