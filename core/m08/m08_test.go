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
