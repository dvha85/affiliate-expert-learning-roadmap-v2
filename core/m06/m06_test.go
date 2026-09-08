package m06_test

import (
	"encoding/json"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m06"
	"strings"
	"testing"
)

func TestFixtureNormalization(t *testing.T) {
	r := m06.WatchRequest{Method: "GET", URL: "https://example.com/offer", AllowHosts: []string{"example.com"}, ObservedAt: "2026-09-03T01:00:00Z", CorrelationID: "fixture-1", Body: "abc"}
	base, status := m06.NormalizeWatchObservation(r, "offer-1")
	if status != "NEW" || base.ContentHash != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" || base.EvidenceKind != "synthetic" || base.UseContext != "test" {
		t.Fatal(status, base)
	}
	raw, err := json.Marshal(base)
	if err != nil || m06.ValidateM06Observation(raw) != "VALID" {
		t.Fatal(err, string(raw))
	}
	r.PreviousHash = base.ContentHash
	same, status := m06.NormalizeWatchObservation(r, "offer-1")
	if status != "UNCHANGED" || same.ObservationID != base.ObservationID {
		t.Fatal(status)
	}
	r.ObservedAt = "2026-09-03T08:00:00+07:00"
	same, _ = m06.NormalizeWatchObservation(r, "offer-1")
	if same.ObservationID != base.ObservationID {
		t.Fatal("equivalent time changed ID")
	}
	r.Body = "abcd"
	changed, status := m06.NormalizeWatchObservation(r, "offer-1")
	if status != "CHANGED" || changed.ObservationID == base.ObservationID {
		t.Fatal(status)
	}
	r.Method = "HEAD"
	missing, _ := m06.NormalizeWatchObservation(r, "offer-1")
	if missing.State != "missing" || missing.ClaimKind != "unknown" {
		t.Fatal(missing)
	}
	r.Method = "POST"
	rejected, status := m06.NormalizeWatchObservation(r, "offer-1")
	if status != "REJECT_WRITE_METHOD" || rejected.ObservationID != "" {
		t.Fatal(status)
	}
	r.Method = "GET"
	r.URL = "https://other.invalid/offer"
	if m06.EvaluateWatchRequest(r) != "REJECT_SOURCE" {
		t.Fatal("source accepted")
	}
}

func TestRawFixtureBoundary(t *testing.T) {
	raw := `{"subject_id":"offer-1","status_code":200,"request":{"method":"GET","url":"https://example.com/offer","allow_hosts":["example.com"],"observed_at":"2026-09-03T01:00:00Z","correlation_id":"fixture-1","body":"abc"}}`
	if _, status := m06.DecodeM06Input([]byte(raw)); status != "VALID" {
		t.Fatal(status)
	}
	for _, bad := range []string{"null", "[]", raw + " {}", strings.Replace(raw, `"body":"abc"`, `"body":null`, 1), strings.Replace(raw, `"subject_id":`, `"extra":1,"subject_id":`, 1), strings.Replace(raw, `"status_code":`, `"status_code":500,"status_code":`, 1)} {
		if _, status := m06.DecodeM06Input([]byte(bad)); status != "INVALID_SCHEMA" {
			t.Fatal(status, bad)
		}
	}
}

func TestOfferFixtureBuildUsesOneCanonicalPacket(t *testing.T) {
	profile := m06.OfferFixtureProfile{
		FixtureURL: "https://example.com/br13/offer",
		SourceURL:  "https://example.com/br13/offer",
		AllowHost:  "example.com",
		Access:     "local_fixture",
		Role:       "synthetic_fixture",
		Limitation: "offline only",
	}
	raw := []byte(`{"version":"br13-offer-fixture/v1","method":"GET","url":"https://example.com/br13/offer","observed_at":"2026-09-03T07:00:00+07:00","correlation_id":"event-1","status_code":200,"body":"{\"product_id\":\"a\",\"product_name\":\"Fixture A\",\"currency\":\"USD\",\"price\":100,\"commission_rate\":0.08}"}`)
	built, err := m06.BuildOfferFixture(raw, profile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(built.RecordID, "watch-") || built.ObservedAt != "2026-09-03T00:00:00Z" {
		t.Fatal(built)
	}
	var packet struct {
		Products []struct {
			Fields []struct {
				Field string `json:"field_or_claim"`
				Value any    `json:"value"`
			} `json:"fields"`
		} `json:"products"`
	}
	if err := json.Unmarshal(built.Packet, &packet); err != nil || len(packet.Products) != 1 || len(packet.Products[0].Fields) != 2 || packet.Products[0].Fields[0].Field != "price" {
		t.Fatal(err, string(built.Packet))
	}
	if _, err := m06.BuildOfferFixture([]byte(strings.Replace(string(raw), `"status_code":200`, `"status_code":201`, 1)), profile); err == nil {
		t.Fatal("non-200 fixture accepted")
	}
	wrongProfile := profile
	wrongProfile.SourceURL = "https://other.invalid/offer"
	wrongProfile.AllowHost = "example.com"
	if _, err := m06.BuildOfferFixture(raw, wrongProfile); err == nil {
		t.Fatal("source/profile mismatch accepted")
	}
}
