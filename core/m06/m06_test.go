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
