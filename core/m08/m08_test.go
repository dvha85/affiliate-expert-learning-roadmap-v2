package m08

import (
	"encoding/json"
	"testing"
)

func TestDecodeIntentPreservesLargeJSONNumberForHash(t *testing.T) {
	intent := SealIntent(Intent{IntentID: "i", DecisionID: "d", EvidenceIDs: []string{"e2", "e1"}, ActionType: "DRAFT", Target: "https://example.com/draft", Parameters: map[string]any{"id": json.Number("9007199254740993")}, ProposedBy: "human", CreatedAt: "2026-09-03T01:00:00Z", ExpiresAt: "2026-09-03T03:00:00Z", CorrelationID: "c", IdempotencyKey: "k"})
	raw, err := json.Marshal(intent)
	if err != nil {
		t.Fatal(err)
	}
	decoded, status := DecodeIntent(raw)
	if status != "VALID" || ComputeIntentHash(decoded) != intent.IntentHash {
		t.Fatalf("large number changed across decode/hash: %s", status)
	}
	if got, ok := decoded.Parameters["id"].(json.Number); !ok || got.String() != "9007199254740993" {
		t.Fatalf("number was rounded or retyped: %#v", decoded.Parameters["id"])
	}
}

func TestValidatePolicyForIntentRejectsImpossibleDecisionRiskPairs(t *testing.T) {
	intent := SealIntent(Intent{IntentID: "i", DecisionID: "d", EvidenceIDs: []string{"e"}, ActionType: "DRAFT", Target: "https://example.com/draft", Parameters: map[string]any{}, ProposedBy: "human", CreatedAt: "2026-09-03T01:00:00Z", ExpiresAt: "2026-09-03T03:00:00Z", CorrelationID: "c", IdempotencyKey: "k"})
	policy := PolicyDecision{PolicyVersion: "v1", IntentID: intent.IntentID, IntentHash: intent.IntentHash, Decision: "ALLOW", RiskClass: "RISK0", Reason: "allow", PolicyMode: "NON_AUTHORIZING", PolicyCheckedAt: "2026-09-03T02:00:00Z"}
	if status := ValidatePolicyForIntent(intent, policy); status != "VALID" {
		t.Fatalf("valid policy rejected: %s", status)
	}
	for name, mutate := range map[string]func(*PolicyDecision){
		"allow risk2":  func(p *PolicyDecision) { p.RiskClass = "RISK2" },
		"unknown risk": func(p *PolicyDecision) { p.RiskClass = "RISK9" },
		"review risk0": func(p *PolicyDecision) {
			p.Decision, p.RiskClass, p.PolicyReviewRequired = "HUMAN_REVIEW", "RISK0", true
		},
		"policy after intent expiry": func(p *PolicyDecision) { p.PolicyCheckedAt = "2026-09-03T03:00:00Z" },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := policy
			mutate(&candidate)
			if status := ValidatePolicyForIntent(intent, candidate); status == "VALID" {
				t.Fatalf("impossible policy accepted: %+v", candidate)
			}
		})
	}
}

func TestDecodeIntentRejectsNullParametersAndDuplicateKeys(t *testing.T) {
	for _, raw := range [][]byte{
		[]byte(`{"intent_id":"i","decision_id":"d","evidence_ids":["e"],"action_type":"DRAFT","target":"https://example.com","parameters":null,"proposed_by":"human","created_at":"2026-09-03T01:00:00Z","expires_at":"2026-09-03T03:00:00Z","correlation_id":"c","idempotency_key":"k","intent_hash":"sha256:0000000000000000000000000000000000000000000000000000000000000000","intent_mode":"PROPOSAL_ONLY","execution_authorized":false}`),
		[]byte(`{"intent_id":"i","intent_id":"other"}`),
	} {
		if _, status := DecodeIntent(raw); status != "INVALID_SCHEMA" {
			t.Fatalf("invalid input accepted: %s", status)
		}
	}
}
