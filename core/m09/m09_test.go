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
		"denied policy":          func(_ *m08.Intent, p *m08.PolicyDecision, _ *ApprovalRecord) { p.Decision = "DENY" },
		"authorizing policy":     func(_ *m08.Intent, p *m08.PolicyDecision, _ *ApprovalRecord) { p.ExecutionAuthorized = true },
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
