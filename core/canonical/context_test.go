package canonical

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m00"
)

func canonicalObservationRaw(t *testing.T, price, commission float64) []byte {
	t.Helper()
	packet, err := json.Marshal(m00.Packet{
		Version: "m00-input/v1", Question: "canonical context test",
		Products: []m00.Product{{
			ObservationID: "source-product-1", SubjectID: "product-1", ProductName: "A&B 日本語",
			Currency: "USD", Fields: []m00.Field{
				{ObservationID: "field-price-1", SubjectID: "product-1", SourceRef: "fixture:price", ObservedAt: "2026-09-01T00:00:00Z", AccessMethod: "fixture", EvidenceKind: "synthetic", Field: "price", ClaimKind: "assumption", Value: &price, State: "observed", Role: "fixture", Method: "fixture", Limitation: "synthetic"},
				{ObservationID: "field-commission-1", SubjectID: "product-1", SourceRef: "fixture:commission", ObservedAt: "2026-09-01T00:00:00Z", AccessMethod: "fixture", EvidenceKind: "synthetic", Field: "commission_rate", ClaimKind: "assumption", Value: &commission, State: "observed", Role: "fixture", Method: "fixture", Limitation: "synthetic"},
			},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	converted, err := m00.Convert(packet)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(converted[0])
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestBuildEvidenceContextRetainsAggregateAndOriginalFieldProvenance(t *testing.T) {
	raw := canonicalObservationRaw(t, 12.5, 0.25)
	ctx, err := BuildEvidenceContext("record-1", "decision-1", []string{"aggregate-1"}, []ObservationInput{{
		Raw: raw, ObservationID: "aggregate-1", SubjectID: "product-1", ClaimKind: "assumption",
		SourceAuthorityOrRole: "fixture", Limitation: "synthetic", Value: map[string]any{"product_id": "product-1"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Version != ContextVersion || ctx.Authority != "canonical_history_store" {
		t.Fatalf("unexpected context envelope: %+v", ctx)
	}
	if got, want := strings.Join(ctx.EvidenceIDs, ","), "aggregate-1,field-commission-1,field-price-1"; got != want {
		t.Fatalf("evidence IDs = %s, want %s", got, want)
	}
	if ctx.Evidence[1].Value == nil || ctx.Evidence[2].Value == nil {
		t.Fatalf("original field values were not retained: %+v", ctx.Evidence)
	}
	if ctx.Evidence[1].SourceAuthorityOrRole != "fixture" || ctx.Evidence[1].Limitation != "synthetic" {
		t.Fatalf("field provenance was not retained: %+v", ctx.Evidence[1])
	}
}

func TestBuildEvidenceContextRejectsForgedOrCollidingIDs(t *testing.T) {
	raw := canonicalObservationRaw(t, 12.5, 0.25)
	cases := []struct {
		name string
		ids  []string
		obs  []ObservationInput
	}{
		{name: "forged aggregate", ids: []string{"forged"}, obs: []ObservationInput{{Raw: raw, ObservationID: "aggregate-1", SubjectID: "product-1"}}},
		{name: "duplicate aggregate", ids: []string{"aggregate-1", "aggregate-1"}, obs: []ObservationInput{{Raw: raw, ObservationID: "aggregate-1", SubjectID: "product-1"}, {Raw: raw, ObservationID: "aggregate-1", SubjectID: "product-1"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := BuildEvidenceContext("record-1", "decision-1", tc.ids, tc.obs); err == nil {
				t.Fatal("forged or duplicate evidence IDs were accepted")
			}
		})
	}

	collisionRaw := canonicalObservationRaw(t, 12.5, 0.25)
	if _, err := BuildEvidenceContext("record-1", "decision-1", []string{"field-price-1"}, []ObservationInput{{
		Raw: collisionRaw, ObservationID: "field-price-1", SubjectID: "product-1",
	}}); err == nil {
		t.Fatal("aggregate/field ID collision was accepted")
	}
}

func TestBuildEvidenceContextRejectsMalformedSourceProjection(t *testing.T) {
	_, err := BuildEvidenceContext("record-1", "decision-1", []string{"aggregate-1"}, []ObservationInput{{
		Raw: []byte(`{"observation_id":"aggregate-1","subject_id":"product-1","access_method":"local_packet_conversion","transformation_or_method":"{}"}`), ObservationID: "aggregate-1", SubjectID: "product-1",
	}})
	if err == nil {
		t.Fatal("malformed source projection was accepted")
	}
}
