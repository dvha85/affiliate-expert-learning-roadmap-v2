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
