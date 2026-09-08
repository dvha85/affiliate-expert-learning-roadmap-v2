package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	return watcherRecordSource(raw, "")
}

// Remote source is supplied only by the fixed, hash-verified fetch adapter.
func watcherRecordSource(raw []byte, remote string) (HistoryRecord, error) {
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
	access, limitation := "local_fixture", "Fixture supplied locally; no fetch, seller authority or business truth verified."
	host := "example.com"
	if remote != "" {
		if remote != watcherPinnedURL {
			return fail("unsupported remote source")
		}
		f.URL, host = remote, "raw.githubusercontent.com"
		access, limitation = "GET", "Fetched pinned synthetic fixture over HTTPS; scenario observed_at is not fetch time; no seller authority or business truth verified."
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
	request := m06.WatchRequest{Method: f.Method, URL: f.URL, AllowHosts: []string{host}, ObservedAt: f.ObservedAt, CorrelationID: f.CorrelationID, Body: f.Body}
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
		fields = append(fields, map[string]any{"observation_id": normalized.ObservationID + "-" + item.name, "subject_id": offer.ProductID, "source_url": f.URL, "observed_at": f.ObservedAt, "access_method": access, "evidence_kind": "synthetic", "use_context": "test", "field_or_claim": item.name, "claim_kind": claim, "value": value, "state": state, "source_authority_or_role": "synthetic_fixture", "transformation_or_method": "br13-offer-fixture/v1; body_sha256=" + normalized.ContentHash + "; correlation_id=" + f.CorrelationID, "limitation": limitation})
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
	if len(args) > 0 && args[0] == "serve" {
		return runWatcherServer(args[1:], stdout, stderr)
	}
	if len(args) > 0 && (args[0] == "history-handoff" || args[0] == "handoff") {
		return runWatcherHistoryHandoff(args[1:], stdout, stderr)
	}
	if len(args) > 0 && args[0] == "fetch-fixture" {
		return runWatcherFetch(args, stdout, stderr)
	}
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

// runWatcherServer exposes the local BR-13 canonical adapter used by the n8n
// blueprint. It accepts only complete HistoryRecord JSON on a loopback-bound
// POST endpoint; all validation and append semantics remain in the learner
// store implementation.
func runWatcherServer(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 || len(args) > 2 {
		fmt.Fprintln(stderr, "usage: bot watcher serve HISTORY [127.0.0.1:8787]")
		return 2
	}
	historyPath := args[0]
	address := "127.0.0.1:8787"
	if len(args) == 2 {
		address = args[1]
	}
	if !strings.HasPrefix(address, "127.0.0.1:") {
		fmt.Fprintln(stderr, "canonical adapter must bind to loopback 127.0.0.1")
		return 1
	}
	if err := distinctActionPaths(historyPath); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if info, err := os.Lstat(historyPath); err == nil && !info.Mode().IsRegular() {
		fmt.Fprintln(stderr, "history must be regular, not symlink")
		return 1
	}
	if _, err := os.Stat(historyPath); err == nil {
		if _, err := LoadHistory(historyPath); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":"OK","execution_permitted":false}`+"\n")
	})
	mux.HandleFunc("/v1/history/append", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = io.WriteString(w, `{"status":"REJECT_METHOD","canonical_history_ack":false}`+"\n")
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"status":"INPUT_ERROR","canonical_history_ack":false}`+"\n")
			return
		}
		var record HistoryRecord
		if err := json.Unmarshal(body, &record); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"status":"INVALID_SCHEMA","canonical_history_ack":false}`+"\n")
			return
		}
		if err := validateHistoryRecord(record); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"status":"INVALID_HISTORY","canonical_history_ack":false}`+"\n")
			return
		}
		status, err := AppendHistory(historyPath, record)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			_, _ = io.WriteString(w, `{"status":"HANDOFF_ERROR","canonical_history_ack":false}`+"\n")
			return
		}
		resolved, err := resolveCanonicalRecord(historyPath, record.RecordID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "RESOLUTION_ERROR", "canonical_history_ack": false, "execution_permitted": false})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": status, "record_id": resolved.RecordID, "canonical_history_ack": true, "canonical_history_persisted": true, "execution_permitted": false, "artifact": resolved})
	})
	mux.HandleFunc("/v1/history", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = io.WriteString(w, `{"status":"REJECT_METHOD"}`+"\n")
			return
		}
		recordID := strings.TrimSpace(r.URL.Query().Get("record_id"))
		if recordID == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"status":"RECORD_ID_REQUIRED"}`+"\n")
			return
		}
		found, err := resolveCanonicalRecord(historyPath, recordID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(w, `{"status":"NOT_FOUND"}`+"\n")
			return
		}
		_ = json.NewEncoder(w).Encode(found)
	})
	server := &http.Server{Addr: address, Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 30 * time.Second}
	fmt.Fprintf(stdout, "watcher canonical adapter listening on http://%s\n", address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

// runWatcherHistoryHandoff is the local BR-13 adapter used by the n8n
// blueprint. It accepts a complete HistoryRecord, validates it with the same
// loader used by list/replay, and reports persistence only after AppendHistory
// returns APPENDED or EXACT_DUPLICATE.
func runWatcherHistoryHandoff(args []string, stdout, stderr io.Writer) int {
	emit := func(status string, artifact any, err error, code int) int {
		if err != nil {
			fmt.Fprintln(stderr, err)
		}
		out := map[string]any{"command": "watcher history-handoff", "status": status, "execution_permitted": false, "canonical_history_ack": false}
		if status == appendAdded || status == appendDuplicate || status == "ACK" {
			out["canonical_history_ack"] = true
		}
		if artifact != nil {
			out["artifact"] = artifact
		}
		if err := json.NewEncoder(stdout).Encode(out); err != nil {
			return 1
		}
		return code
	}
	if len(args) != 2 {
		return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot watcher history-handoff HISTORY RECORD.json"), 2)
	}
	if err := distinctActionPaths(args...); err != nil {
		return emit("PATH_ERROR", nil, err, 1)
	}
	raw, err := os.ReadFile(args[1])
	if err != nil {
		return emit("INPUT_ERROR", nil, err, 1)
	}
	var record HistoryRecord
	if err := json.Unmarshal(raw, &record); err != nil {
		return emit("INVALID_SCHEMA", nil, err, 1)
	}
	if err := validateHistoryRecord(record); err != nil {
		return emit("INVALID_HISTORY", nil, err, 1)
	}
	status, err := AppendHistory(args[0], record)
	if err != nil {
		return emit("HANDOFF_ERROR", nil, err, 1)
	}
	resolved, err := resolveCanonicalRecord(args[0], record.RecordID)
	if err != nil {
		return emit("RESOLUTION_ERROR", nil, err, 1)
	}
	return emit(status, map[string]any{"record_id": resolved.RecordID, "decision_id": resolved.RecordedResult.DecisionID, "state": resolved.RecordedResult.State, "evidence_ids": resolved.RecordedResult.EvidenceIDs, "record": resolved}, nil, 0)
}
