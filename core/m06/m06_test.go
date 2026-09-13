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

func TestCanonicalContentHashIgnoresJSONKeyOrderOnly(t *testing.T) {
	first := `{"product_id":"offer-1","price":100,"title":"A&B"}`
	reordered := `{"title":"A&B","price":100,"product_id":"offer-1"}`
	if m06.CanonicalContentHash(first) != m06.CanonicalContentHash(reordered) {
		t.Fatal("JSON key order changed the M06 content fingerprint")
	}
	if m06.ContentHash(first) == m06.ContentHash(reordered) {
		t.Fatal("test requires distinct transport bytes")
	}
	if m06.CanonicalContentHash("plain text") != m06.ContentHash("plain text") {
		t.Fatal("non-JSON body was normalized without a parser profile")
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
	reordered := []byte(`{"version":"br13-offer-fixture/v1","method":"GET","url":"https://example.com/br13/offer","observed_at":"2026-09-03T07:00:00+07:00","correlation_id":"event-1","status_code":200,"body":"{\"commission_rate\":0.08,\"price\":100,\"currency\":\"USD\",\"product_name\":\"A&B Bé\",\"product_id\":\"a\"}"}`)
	unicode := []byte(`{"version":"br13-offer-fixture/v1","method":"GET","url":"https://example.com/br13/offer","observed_at":"2026-09-03T07:00:00+07:00","correlation_id":"event-1","status_code":200,"body":"{\"product_id\":\"a\",\"product_name\":\"A&B Bé\",\"currency\":\"USD\",\"price\":100,\"commission_rate\":0.08}"}`)
	first, err := m06.BuildOfferFixture(unicode, profile)
	if err != nil {
		t.Fatal(err)
	}
	second, err := m06.BuildOfferFixture(reordered, profile)
	if err != nil || first.RecordID != second.RecordID || string(first.Packet) != string(second.Packet) {
		t.Fatalf("equivalent Unicode fixture was not canonical: first=%+v second=%+v err=%v", first, second, err)
	}
}

func TestOfferFixtureMarksMissingPriceAndCommissionWithoutInventingValues(t *testing.T) {
	profile := m06.OfferFixtureProfile{FixtureURL: "https://example.com/br13/offer", SourceURL: "https://example.com/br13/offer", AllowHost: "example.com", Access: "local_fixture", Role: "synthetic_fixture", Limitation: "offline only"}
	raw := []byte(`{"version":"br13-offer-fixture/v1","method":"GET","url":"https://example.com/br13/offer","observed_at":"2026-09-03T07:00:00+07:00","correlation_id":"event-missing","status_code":200,"body":"{\"product_id\":\"a\",\"product_name\":\"Fixture A\",\"currency\":\"USD\",\"price\":null}"}`)
	built, err := m06.BuildOfferFixture(raw, profile)
	if err != nil {
		t.Fatal(err)
	}
	var packet struct {
		Products []struct {
			Fields []struct {
				Field string `json:"field_or_claim"`
				Value any    `json:"value"`
				State string `json:"state"`
				Claim string `json:"claim_kind"`
			} `json:"fields"`
		} `json:"products"`
	}
	if err := json.Unmarshal(built.Packet, &packet); err != nil || len(packet.Products) != 1 || len(packet.Products[0].Fields) != 2 {
		t.Fatalf("decode missing-value packet: %v %s", err, built.Packet)
	}
	for _, field := range packet.Products[0].Fields {
		if field.Field != "price" && field.Field != "commission_rate" {
			t.Fatalf("unexpected projected field: %+v", field)
		}
		if field.Value != nil || field.State != "missing" || field.Claim != "unknown" {
			t.Fatalf("missing source value was invented or misclassified: %+v", field)
		}
	}
}

func TestAccesstradeShopeeCampaignIsASeparateReadOnlyMetadataProfile(t *testing.T) {
	profile := m06.AccesstradeShopeeCampaignProfile{SourceURL: m06.AccesstradeShopeeSmartlinkURL}
	raw := []byte(`{"version":"accesstrade-shopee-campaign-capture/v1","method":"GET","observed_at":"2026-09-13T07:00:00+07:00","correlation_id":"campaign-metadata-1","status_code":200,"redirected":false,"source_page_sha256":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","campaign_title":"Shopee Việt Nam Smartlink cho tất cả thiết bị","merchant_label":"Shopee","campaign_category":"Thương Mại Điện Tử","campaign_status_label":"Chờ duyệt","campaign_period_label":"05/05/2023 - Nay"}`)
	built, err := m06.BuildAccesstradeShopeeCampaign(raw, profile)
	if err != nil || !strings.HasPrefix(built.RecordID, "watch-") || built.ObservedAt != "2026-09-13T00:00:00Z" {
		t.Fatalf("selected source build failed: %+v %v", built, err)
	}
	var packet struct {
		Question string `json:"question"`
		Products []struct {
			ProductName string `json:"product_name"`
			Fields      []struct {
				Field      string `json:"field_or_claim"`
				Value      any    `json:"value"`
				Kind       string `json:"evidence_kind"`
				Claim      string `json:"claim_kind"`
				State      string `json:"state"`
				Limitation string `json:"limitation"`
			} `json:"fields"`
		} `json:"products"`
	}
	if err := json.Unmarshal(built.Packet, &packet); err != nil || len(packet.Products) != 1 || packet.Products[0].ProductName != "Shopee Việt Nam Smartlink cho tất cả thiết bị" || !strings.Contains(packet.Question, "no ranking") {
		t.Fatalf("invalid selected packet: %v %s", err, built.Packet)
	}
	for _, field := range packet.Products[0].Fields {
		if field.Value != nil || field.Kind != "real" || field.Claim != "unknown" || field.State != "missing" || !strings.Contains(field.Limitation, "not independent business truth") {
			t.Fatalf("selected campaign invented a business field: %+v", field)
		}
	}
	// Strict decoding and the reviewed profile reject source substitution,
	// write/redirect/status paths, mutable quantitative claims and raw extras.
	for _, bad := range []string{
		strings.Replace(string(raw), `"method":"GET"`, `"method":"POST"`, 1),
		strings.Replace(string(raw), `"status_code":200`, `"status_code":201`, 1),
		strings.Replace(string(raw), `"redirected":false`, `"redirected":true`, 1),
		strings.Replace(string(raw), `"campaign_title":"Shopee Việt Nam Smartlink cho tất cả thiết bị"`, `"campaign_title":"Other"`, 1),
		strings.Replace(string(raw), `}`, `,"commission_rate":0.9}`, 1),
	} {
		if _, err := m06.BuildAccesstradeShopeeCampaign([]byte(bad), profile); err == nil {
			t.Fatalf("unsafe selected capture accepted: %s", bad)
		}
	}
	wrongProfile := profile
	wrongProfile.SourceURL = "https://pub2.accesstrade.vn/campaign/5087153089503673507"
	if _, err := m06.BuildAccesstradeShopeeCampaign(raw, wrongProfile); err == nil {
		t.Fatal("alternate selected source URL accepted")
	}
}

func TestAccesstradeShopeeCampaignCanonicalizesEquivalentCapture(t *testing.T) {
	profile := m06.AccesstradeShopeeCampaignProfile{SourceURL: m06.AccesstradeShopeeSmartlinkURL}
	first := []byte(`{"version":"accesstrade-shopee-campaign-capture/v1","method":"GET","observed_at":"2026-09-13T00:00:00Z","correlation_id":"campaign-metadata-2","status_code":200,"redirected":false,"source_page_sha256":"abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd","campaign_title":"Shopee Việt Nam Smartlink cho tất cả thiết bị","merchant_label":"Shopee","campaign_category":"Thương Mại Điện Tử","campaign_status_label":"Chờ duyệt","campaign_period_label":"05/05/2023 - Nay"}`)
	second := []byte(`{"campaign_period_label":"05/05/2023 - Nay","campaign_status_label":"Chờ duyệt","campaign_category":"Thương Mại Điện Tử","merchant_label":"Shopee","campaign_title":"Shopee Việt Nam Smartlink cho tất cả thiết bị","source_page_sha256":"abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd","redirected":false,"status_code":200,"correlation_id":"campaign-metadata-2","observed_at":"2026-09-13T07:00:00+07:00","method":"GET","version":"accesstrade-shopee-campaign-capture/v1"}`)
	a, err := m06.BuildAccesstradeShopeeCampaign(first, profile)
	if err != nil {
		t.Fatal(err)
	}
	b, err := m06.BuildAccesstradeShopeeCampaign(second, profile)
	if err != nil || a.RecordID != b.RecordID || string(a.Packet) != string(b.Packet) {
		t.Fatalf("equivalent capture did not deduplicate: a=%+v b=%+v err=%v", a, b, err)
	}
	changed := strings.Replace(string(second), `"campaign_status_label":"Chờ duyệt"`, `"campaign_status_label":"Đã duyệt"`, 1)
	c, err := m06.BuildAccesstradeShopeeCampaign([]byte(changed), profile)
	if err != nil || c.RecordID == a.RecordID {
		t.Fatalf("changed campaign metadata was not a new observation: %+v %v", c, err)
	}
	changedPage := strings.Replace(string(second), `"source_page_sha256":"a`, `"source_page_sha256":"b`, 1)
	d, err := m06.BuildAccesstradeShopeeCampaign([]byte(changedPage), profile)
	if err != nil || d.RecordID == a.RecordID {
		t.Fatalf("changed source-page fingerprint was not a new observation: %+v %v", d, err)
	}
}
