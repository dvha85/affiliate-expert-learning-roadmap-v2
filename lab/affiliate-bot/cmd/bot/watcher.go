package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m00"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m06"
)

const watcherFixtureURL = "https://example.com/br13/offer"

// Deliberately no caller allowlist, endpoint option or network client.
type watcherFixture struct {
	Version       string `json:"version"`
	Method        string `json:"method"`
	URL           string `json:"url"`
	ObservedAt    string `json:"observed_at"`
	CorrelationID string `json:"correlation_id"`
	StatusCode    int    `json:"status_code"`
	Body          string `json:"body"`
}

func watcherRecord(raw []byte) (HistoryRecord, error) {
	fail := func(s string) (HistoryRecord, error) { return HistoryRecord{}, fmt.Errorf("watcher: %s", s) }
	var f watcherFixture
	if err := contracts.DecodeStrict(raw, &f); err != nil {
		return fail("invalid fixture envelope")
	}
	if f.Version != "br13-offer-fixture/v1" || f.Method != "GET" || f.URL != watcherFixtureURL {
		return fail("unsupported fixture profile/source/method")
	}
	if f.StatusCode != 200 {
		return fail("response status must be 200")
	}
	if strings.TrimSpace(f.CorrelationID) == "" || f.CorrelationID != strings.TrimSpace(f.CorrelationID) {
		return fail("stable correlation_id required")
	}
	at, err := time.Parse(time.RFC3339, f.ObservedAt)
	if err != nil {
		return fail("invalid observed_at")
	}
	f.ObservedAt = at.UTC().Format(time.RFC3339Nano)
	var offer struct {
		ProductID   string          `json:"product_id"`
		ProductName string          `json:"product_name"`
		Currency    string          `json:"currency"`
		Price       json.RawMessage `json:"price"`
		Commission  json.RawMessage `json:"commission_rate"`
	}
	if err := contracts.DecodeStrict([]byte(f.Body), &offer); err != nil {
		return fail("malformed offer body")
	}
	if strings.TrimSpace(offer.ProductID) == "" || strings.TrimSpace(offer.ProductName) == "" || offer.Currency != "USD" {
		return fail("product identity/name and USD required")
	}
	request := m06.WatchRequest{Method: f.Method, URL: f.URL, AllowHosts: []string{"example.com"}, ObservedAt: f.ObservedAt, CorrelationID: f.CorrelationID, Body: f.Body}
	normalized, status := m06.NormalizeWatchObservation(request, offer.ProductID)
	if status != "NEW" {
		return fail("normalization rejected")
	}
	// Record identity pins a declared observation event, not body bytes. Changed
	// content/time for the same correlation ID conflicts instead of silently
	// turning a retry into a new history record.
	event, _ := json.Marshal([]string{f.Version, f.URL, f.CorrelationID})
	recordID := "watch-" + m06.ContentHash(string(event))
	fields := []map[string]any{}
	for _, item := range []struct {
		name  string
		value json.RawMessage
	}{{"price", offer.Price}, {"commission_rate", offer.Commission}} {
		value := item.value
		state, claim := "observed", "assumption"
		if len(value) == 0 || string(value) == "null" {
			value = json.RawMessage("null")
			state, claim = "missing", "unknown"
		}
		fields = append(fields, map[string]any{"observation_id": normalized.ObservationID + "-" + item.name, "subject_id": offer.ProductID, "source_url": f.URL, "observed_at": f.ObservedAt, "access_method": "local_fixture", "evidence_kind": "synthetic", "use_context": "test", "field_or_claim": item.name, "claim_kind": claim, "value": value, "state": state, "source_authority_or_role": "synthetic_fixture", "transformation_or_method": "br13-offer-fixture/v1; body_sha256=" + normalized.ContentHash + "; correlation_id=" + f.CorrelationID, "limitation": "Fixture supplied locally; no fetch, seller authority or business truth verified."})
	}
	packet := map[string]any{"version": "m00-input/v1", "question": "Synthetic watcher scenario; no business recommendation.", "products": []any{map[string]any{"observation_id": normalized.ObservationID, "subject_id": offer.ProductID, "product_name": offer.ProductName, "currency": offer.Currency, "fields": fields}}}
	packetRaw, err := json.Marshal(packet)
	if err != nil {
		return fail("projection encoding")
	}
	converted, err := m00.Convert(packetRaw)
	if err != nil {
		return fail("invalid offer fields or numeric precision")
	}
	projection, err := json.Marshal(converted)
	if err != nil {
		return fail("projection encoding")
	}
	var observations []Observation
	if err := json.Unmarshal(projection, &observations); err != nil {
		return fail("projection decoding")
	}
	// In fixture mode both timestamps are declared scenario time, not wall clock.
	return NewHistoryRecord(recordID, f.ObservedAt, f.ObservedAt, observations)
}

func runWatcher(args []string, stdout, stderr io.Writer) int {
	emit := func(status string, artifact any, err error, code int) int {
		if err != nil {
			fmt.Fprintln(stderr, err)
		}
		env := map[string]any{"command": "watcher fixture-import", "status": status, "execution_permitted": false, "network_fetch_performed": false, "persisted": false}
		if artifact != nil {
			env["artifact"] = artifact
			env["persisted"] = true
		}
		if err := json.NewEncoder(stdout).Encode(env); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return code
	}
	if len(args) != 3 || args[0] != "fixture-import" {
		return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot watcher fixture-import HISTORY FIXTURE"), 2)
	}
	if err := distinctActionPaths(args[1:]...); err != nil {
		return emit("PATH_ERROR", nil, err, 1)
	}
	info, err := os.Lstat(args[1])
	if err != nil && !os.IsNotExist(err) {
		return emit("HISTORY_ERROR", nil, err, 1)
	}
	if err == nil && !info.Mode().IsRegular() {
		return emit("PATH_ERROR", nil, fmt.Errorf("history must be regular, not symlink"), 1)
	}
	raw, err := readCampaignFile(args[2], 64<<10)
	if err != nil {
		return emit("INPUT_ERROR", nil, err, 1)
	}
	record, err := watcherRecord(raw)
	if err != nil {
		return emit("FIXTURE_ERROR", nil, err, 1)
	}
	status, err := AppendHistory(args[1], record)
	if err != nil {
		return emit("HANDOFF_ERROR", nil, err, 1)
	}
	return emit(status, map[string]any{"record_id": record.RecordID, "decision_id": record.RecordedResult.DecisionID, "state": record.RecordedResult.State, "observation_ids": record.RecordedResult.EvidenceIDs}, nil, 0)
}
