package m06

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

// OfferFixture is the deliberately narrow synthetic M06 source profile used
// by the learner watcher and the n8n handoff adapter. It is not a generic
// seller-page parser and must not be relabelled as real evidence.
type OfferFixture struct {
	Version       string `json:"version"`
	Method        string `json:"method"`
	URL           string `json:"url"`
	ObservedAt    string `json:"observed_at"`
	CorrelationID string `json:"correlation_id"`
	StatusCode    int    `json:"status_code"`
	Body          string `json:"body"`
}

// OfferFixtureProfile separates the URL embedded in the fixture from the
// provenance URL recorded by a fixed fetch adapter. Both values are supplied
// by code, never by an n8n expression or remote response.
type OfferFixtureProfile struct {
	FixtureURL string
	SourceURL  string
	AllowHost  string
	Access     string
	Role       string
	Limitation string
}

// OfferFixtureBuild is the shared, pre-M02 result. Packet is an exact M00
// input packet; the learner is still responsible for converting it into its
// HistoryRecord and appending that record to the canonical store.
type OfferFixtureBuild struct {
	RecordID   string
	ObservedAt string
	Packet     json.RawMessage
}

func BuildOfferFixture(raw []byte, profile OfferFixtureProfile) (OfferFixtureBuild, error) {
	fail := func(message string) (OfferFixtureBuild, error) {
		return OfferFixtureBuild{}, fmt.Errorf("M06 offer fixture: %s", message)
	}
	if strings.TrimSpace(profile.FixtureURL) == "" || strings.TrimSpace(profile.SourceURL) == "" || strings.TrimSpace(profile.AllowHost) == "" || strings.TrimSpace(profile.Access) == "" || strings.TrimSpace(profile.Role) == "" || strings.TrimSpace(profile.Limitation) == "" {
		return fail("incomplete fixed profile")
	}
	var fixture OfferFixture
	if err := contracts.DecodeStrict(raw, &fixture); err != nil {
		return fail("invalid fixture envelope")
	}
	if fixture.Version != "br13-offer-fixture/v1" || fixture.Method != "GET" || fixture.URL != profile.FixtureURL || fixture.StatusCode != 200 {
		return fail("unsupported fixture profile/source/method")
	}
	if strings.TrimSpace(fixture.CorrelationID) == "" || fixture.CorrelationID != strings.TrimSpace(fixture.CorrelationID) {
		return fail("stable correlation_id required")
	}
	observedAt, err := time.Parse(time.RFC3339, fixture.ObservedAt)
	if err != nil {
		return fail("invalid observed_at")
	}
	fixture.ObservedAt = observedAt.UTC().Format(time.RFC3339Nano)
	request := WatchRequest{Method: fixture.Method, URL: profile.SourceURL, AllowHosts: []string{profile.AllowHost}, ObservedAt: fixture.ObservedAt, CorrelationID: fixture.CorrelationID, Body: fixture.Body}
	normalized, state := NormalizeWatchObservation(request, "pending")
	if state != "NEW" {
		return fail("normalization rejected")
	}
	var offer struct {
		ProductID   string          `json:"product_id"`
		ProductName string          `json:"product_name"`
		Currency    string          `json:"currency"`
		Price       json.RawMessage `json:"price"`
		Commission  json.RawMessage `json:"commission_rate"`
	}
	if err := contracts.DecodeStrict([]byte(fixture.Body), &offer); err != nil {
		return fail("malformed offer body")
	}
	if strings.TrimSpace(offer.ProductID) == "" || strings.TrimSpace(offer.ProductName) == "" || offer.Currency != "USD" {
		return fail("product identity/name and USD required")
	}
	// Recompute after identity validation so the observation ID binds the real
	// subject rather than the temporary validation placeholder above.
	normalized, state = NormalizeWatchObservation(request, offer.ProductID)
	if state != "NEW" {
		return fail("normalization rejected")
	}
	fields := make([]map[string]any, 0, 2)
	for _, item := range []struct {
		name  string
		value json.RawMessage
	}{{"price", offer.Price}, {"commission_rate", offer.Commission}} {
		value := item.value
		fieldState, claim := "observed", "assumption"
		if len(value) == 0 || string(value) == "null" {
			value = json.RawMessage("null")
			fieldState, claim = "missing", "unknown"
		}
		fields = append(fields, map[string]any{
			"observation_id":           normalized.ObservationID + "-" + item.name,
			"subject_id":               offer.ProductID,
			"source_url":               profile.SourceURL,
			"observed_at":              normalized.ObservedAt,
			"access_method":            profile.Access,
			"evidence_kind":            "synthetic",
			"use_context":              "test",
			"field_or_claim":           item.name,
			"claim_kind":               claim,
			"value":                    value,
			"state":                    fieldState,
			"source_authority_or_role": profile.Role,
			"transformation_or_method": "br13-offer-fixture/v1; body_sha256=" + normalized.ContentHash + "; correlation_id=" + fixture.CorrelationID,
			"limitation":               profile.Limitation,
		})
	}
	packet, err := json.Marshal(map[string]any{
		"version":  "m00-input/v1",
		"question": "Synthetic watcher scenario; no business recommendation.",
		"products": []any{map[string]any{
			"observation_id": normalized.ObservationID,
			"subject_id":     offer.ProductID,
			"product_name":   offer.ProductName,
			"currency":       offer.Currency,
			"fields":         fields,
		}},
	})
	if err != nil {
		return fail("projection encoding")
	}
	identity, err := json.Marshal([]string{fixture.Version, profile.SourceURL, fixture.CorrelationID})
	if err != nil {
		return fail("record identity encoding")
	}
	return OfferFixtureBuild{RecordID: "watch-" + ContentHash(string(identity)), ObservedAt: normalized.ObservedAt, Packet: packet}, nil
}
