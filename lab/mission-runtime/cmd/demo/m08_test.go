package main

import "testing"

type m08EvalCase struct {
	CaseID              string             `json:"case_id"`
	Intent              ShadowActionIntent `json:"intent"`
	Now                 string             `json:"now"`
	PolicyVersion       string             `json:"policy_version"`
	KnownDecisionIDs    []string           `json:"known_decision_ids"`
	KnownEvidenceIDs    []string           `json:"known_evidence_ids"`
	AllowedHosts        []string           `json:"allowed_hosts"`
	Mutation            string             `json:"mutation"`
	SeenMode            string             `json:"seen_mode"`
	ExpectedDecision    string             `json:"expected_decision"`
	ExpectedRisk        string             `json:"expected_risk"`
	ExpectedReason      string             `json:"expected_reason"`
}

func TestM08EvalPack(t *testing.T) {
	cases := loadCases[m08EvalCase](t, "M08-shadow-policy")
	for _, c := range cases {
		t.Run(c.CaseID, func(t *testing.T) {
			intent := SealShadowActionIntent(c.Intent)
			switch c.Mutation {
			case "target_after_seal":
				intent.Target = "https://evil.example/publish"
			case "shadow_off_after_seal":
				intent.ShadowOnly = false
			case "dry_run_off_after_seal":
				intent.DryRun = false
			}
			seen := map[string]string{}
			switch c.SeenMode {
			case "same_hash":
				seen[intent.IdempotencyKey] = intent.IntentHash
			case "collision":
				seen[intent.IdempotencyKey] = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
			}
			ctx := ShadowPolicyContext{
				PolicyVersion: c.PolicyVersion, Now: c.Now, KnownDecisionIDs: c.KnownDecisionIDs,
				KnownEvidenceIDs: c.KnownEvidenceIDs, AllowedHosts: c.AllowedHosts,
				ActionRisk: map[string]string{"PREPARE_LOCAL_DRAFT": "RISK0", "UPDATE_DRAFT": "RISK1", "PUBLISH_CONTENT": "RISK2"},
				SeenIdempotency: seen,
			}
			got := EvaluateShadowPolicy(intent, ctx)
			if got.Decision != c.ExpectedDecision || got.RiskClass != c.ExpectedRisk || got.Reason != c.ExpectedReason {
				t.Fatalf("expected %s/%s/%s got %s/%s/%s", c.ExpectedDecision, c.ExpectedRisk, c.ExpectedReason, got.Decision, got.RiskClass, got.Reason)
			}
			if !got.ShadowOnly || got.ExecutionAuthorized {
				t.Fatalf("M08 must never authorize execution: %+v", got)
			}
		})
	}
}

func TestM08HashStableAcrossEvidenceOrder(t *testing.T) {
	base := ShadowActionIntent{IntentID: "i", DecisionID: "d", EvidenceIDs: []string{"e2", "e1"}, ActionType: "UPDATE_DRAFT", Target: "https://example.com/draft", Parameters: map[string]any{"b": 2, "a": 1}, ProposedBy: "human", CreatedAt: "2026-09-03T01:00:00Z", ExpiresAt: "2026-09-03T03:00:00Z", CorrelationID: "c", IdempotencyKey: "k", ShadowOnly: true, DryRun: true}
	a := ComputeShadowIntentHash(base)
	base.EvidenceIDs = []string{"e1", "e2"}
	b := ComputeShadowIntentHash(base)
	if a != b {
		t.Fatalf("canonical hash must ignore evidence ordering: %s != %s", a, b)
	}
}
