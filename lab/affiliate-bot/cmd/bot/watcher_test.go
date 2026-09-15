package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	corem07 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m07"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/lab/affiliate-bot/internal/store"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func watchFixture() watcherFixture {
	return watcherFixture{Version: "br13-offer-fixture/v1", Method: "GET", URL: watcherFixtureURL, ObservedAt: "2026-09-03T00:00:00Z", CorrelationID: "event-1", StatusCode: 200, Body: `{"product_id":"a","product_name":"Fixture A","currency":"USD","price":100,"commission_rate":0.08}`}
}

func accesstradeShopeeCapture() json.RawMessage {
	return json.RawMessage(`{"version":"accesstrade-shopee-campaign-capture/v1","method":"GET","observed_at":"2026-09-13T00:00:00Z","correlation_id":"selected-source-1","status_code":200,"redirected":false,"source_page_sha256":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","campaign_title":"Shopee Việt Nam Smartlink cho tất cả thiết bị","merchant_label":"Shopee","campaign_category":"Thương Mại Điện Tử","campaign_status_label":"Chờ duyệt","campaign_period_label":"05/05/2023 - Nay"}`)
}
func watchRun(t *testing.T, h, input string, f watcherFixture, want string) {
	t.Helper()
	raw, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(input, raw, 0600); err != nil {
		t.Fatal(err)
	}
	var out, diag bytes.Buffer
	code := runWatcher([]string{"fixture-import", h, input}, &out, &diag)
	var env map[string]any
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	success := want == "APPENDED" || want == "EXACT_DUPLICATE"
	if env["status"] != want || (code == 0) != success || env["persisted"] != success || env["execution_permitted"] != false || env["network_fetch_performed"] != false {
		t.Fatal(code, env, diag.String())
	}
	if !success && env["artifact"] != nil {
		t.Fatal("error artifact")
	}
}

func TestM06HTTPAdapterBuildsAndResolvesCanonicalHistory(t *testing.T) {
	dir := t.TempDir()
	history := filepath.Join(dir, "history.jsonl")
	call := func(method string, fixture watcherFixture) *httptest.ResponseRecorder {
		raw, err := json.Marshal(m06AdapterRequest{Fixture: mustRawJSON(t, fixture)})
		if err != nil {
			t.Fatal(err)
		}
		recorder := httptest.NewRecorder()
		m06AdapterHandler(history).ServeHTTP(recorder, httptest.NewRequest(method, "/v1/m06/fixture-import", bytes.NewReader(raw)))
		return recorder
	}
	wrongMethod := call(http.MethodGet, watchFixture())
	if wrongMethod.Code != http.StatusMethodNotAllowed {
		t.Fatal(wrongMethod.Code, wrongMethod.Body.String())
	}
	first := call(http.MethodPost, watchFixture())
	if first.Code != http.StatusOK {
		t.Fatal(first.Code, first.Body.String())
	}
	var response struct {
		Status                    string `json:"status"`
		RecordID                  string `json:"record_id"`
		CanonicalHistoryACK       bool   `json:"canonical_history_ack"`
		CanonicalHistoryPersisted bool   `json:"canonical_history_persisted"`
	}
	if err := json.Unmarshal(first.Body.Bytes(), &response); err != nil || response.Status != appendAdded || response.RecordID == "" || !response.CanonicalHistoryACK || !response.CanonicalHistoryPersisted {
		t.Fatal(err, first.Body.String())
	}
	second := call(http.MethodPost, watchFixture())
	if second.Code != http.StatusOK || !strings.Contains(second.Body.String(), appendDuplicate) {
		t.Fatal(second.Code, second.Body.String())
	}
	reordered := watchFixture()
	reordered.Body = `{"commission_rate":0.08,"price":100,"currency":"USD","product_name":"Fixture A","product_id":"a"}`
	if response := call(http.MethodPost, reordered); response.Code != http.StatusOK || !strings.Contains(response.Body.String(), appendDuplicate) {
		t.Fatal("JSON key reorder was not an exact retry", response.Code, response.Body.String())
	}
	records, err := LoadHistory(history)
	if err != nil || len(records) != 1 || records[0].RecordID != response.RecordID || Replay(records[0]).State != replayMatch {
		t.Fatal(err, records)
	}
	bad := watchFixture()
	bad.URL = "https://other.invalid/offer"
	if result := call(http.MethodPost, bad); result.Code != http.StatusBadRequest {
		t.Fatal("unrecognized fixture source accepted", result.Code, result.Body.String())
	}
	records, err = LoadHistory(history)
	if err != nil || len(records) != 1 {
		t.Fatal("rejected fixture changed canonical history", err, records)
	}
	missing := watchFixture()
	missing.CorrelationID = "event-missing"
	missing.Body = `{"product_id":"missing","product_name":"Missing","currency":"USD","price":null,"commission_rate":null}`
	if response := call(http.MethodPost, missing); response.Code != http.StatusOK || !strings.Contains(response.Body.String(), appendAdded) {
		t.Fatal("missing fields were not persisted canonically", response.Code, response.Body.String())
	}
	records, err = LoadHistory(history)
	if err != nil || len(records) != 2 || Replay(records[1]).State != replayMatch || records[1].Observations[0].Price != nil || records[1].Observations[0].CommissionRate != nil {
		t.Fatal("missing-field record did not replay MATCH", err, records)
	}
}

func TestHistoryHandoffRejectsDuplicateRawKeyBeforePersistence(t *testing.T) {
	dir := t.TempDir()
	history, input := filepath.Join(dir, "history.jsonl"), filepath.Join(dir, "record.json")
	fixtureRaw, err := json.Marshal(watchFixture())
	if err != nil {
		t.Fatal(err)
	}
	record, err := watcherRecord(fixtureRaw)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	needle := []byte(`"record_id":"` + record.RecordID + `"`)
	duplicate := bytes.Replace(valid, needle, append(append([]byte(nil), needle...), append([]byte(`,`), needle...)...), 1)
	if bytes.Equal(valid, duplicate) {
		t.Fatal("could not construct duplicate-key history handoff")
	}
	if err := os.WriteFile(input, duplicate, 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := runWatcherHistoryHandoff([]string{history, input}, &stdout, &stderr); code == 0 || !strings.Contains(stdout.String(), `"status":"INVALID_SCHEMA"`) {
		t.Fatalf("CLI accepted duplicate-key history handoff: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(history); !os.IsNotExist(err) {
		t.Fatalf("CLI duplicate-key handoff created canonical history: %v", err)
	}
	httpResult := httptest.NewRecorder()
	historyHandoffHTTPHandler(history).ServeHTTP(httpResult, httptest.NewRequest(http.MethodPost, "/v1/history/append", bytes.NewReader(duplicate)))
	if httpResult.Code != http.StatusBadRequest || !strings.Contains(httpResult.Body.String(), `"status":"INVALID_SCHEMA"`) {
		t.Fatalf("HTTP accepted duplicate-key history handoff: code=%d body=%s", httpResult.Code, httpResult.Body.String())
	}
	if _, err := os.Stat(history); !os.IsNotExist(err) {
		t.Fatalf("HTTP duplicate-key handoff created canonical history: %v", err)
	}
}

func TestSelectedAccesstradeCampaignUsesSharedM06HistoryAndM07Boundary(t *testing.T) {
	dir := t.TempDir()
	history := filepath.Join(dir, "history.jsonl")
	call := func(path string, body any) *httptest.ResponseRecorder {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		recorder := httptest.NewRecorder()
		m06AdapterHandler(history).ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, path, bytes.NewReader(raw)))
		return recorder
	}
	wrongEndpoint := call("/v1/m06/fixture-import", m06AdapterRequest{Capture: accesstradeShopeeCapture()})
	if wrongEndpoint.Code == http.StatusOK {
		t.Fatal("selected capture entered fixture endpoint")
	}
	first := call("/v1/m06/accesstrade-shopee-campaign", m06AdapterRequest{Capture: accesstradeShopeeCapture()})
	if first.Code != http.StatusOK || !strings.Contains(first.Body.String(), `"canonical_history_ack":true`) || !strings.Contains(first.Body.String(), `"state":"GET_MORE_DATA"`) {
		t.Fatalf("selected campaign was not safely persisted: %d %s", first.Code, first.Body.String())
	}
	second := call("/v1/m06/accesstrade-shopee-campaign", m06AdapterRequest{Capture: accesstradeShopeeCapture()})
	if second.Code != http.StatusOK || !strings.Contains(second.Body.String(), appendDuplicate) {
		t.Fatalf("selected campaign retry was not exact duplicate: %d %s", second.Code, second.Body.String())
	}
	records, err := LoadHistory(history)
	if err != nil || len(records) != 1 || records[0].RecordedResult.State != stateGetMoreData {
		t.Fatal(err, records)
	}
	record := records[0]
	if record.Observations[0].EvidenceKind != "synthetic" || record.Observations[0].Price != nil || record.Observations[0].CommissionRate != nil {
		t.Fatalf("selected campaign was misclassified/ranked: %+v", record.Observations[0])
	}
	ctx, err := m07EvidenceContext(record)
	if err != nil {
		t.Fatal(err)
	}
	var unknown corem07.Evidence
	for _, evidence := range ctx.Evidence {
		if evidence.FieldOrClaim == "commission_rate" {
			unknown = evidence
			break
		}
	}
	if unknown.EvidenceID == "" || unknown.ClaimKind != "unknown" || !strings.Contains(unknown.Limitation, "not independent business truth") {
		t.Fatalf("M07 did not retain selected-source limitation: %+v", unknown)
	}
	claim := corem07.Claim{Text: "commission_rate=0.9 [evidence:" + unknown.EvidenceID + "]", FieldOrClaim: "commission_rate", Value: json.RawMessage("0.9"), EvidenceIDs: []string{unknown.EvidenceID}}
	model := corem07.AgentOutput{State: "HUMAN_REVIEW", Answer: claim.Text, EvidenceIDs: []string{unknown.EvidenceID}, Claims: []corem07.Claim{claim}, ToolCalls: []corem07.ToolRequest{}, Authority: "A2-RO", WritePermission: false}
	if _, err := corem07.ValidateAgentOutput(mustRawJSON(t, model), ctx.Evidence, []corem07.ToolSpec{{Name: "read", ReadOnly: true, AllowedMethods: []string{"GET"}, AllowedHosts: []string{"example.com"}}}); err == nil {
		t.Fatal("M07 grounded an invented commission/earnings claim from selected metadata")
	}
}

func TestSelectedAccesstradeCampaignCLIImportsOnlySanitizedMetadata(t *testing.T) {
	dir := t.TempDir()
	history, capture := filepath.Join(dir, "history.jsonl"), filepath.Join(dir, "sanitized-capture.json")
	if err := os.WriteFile(capture, accesstradeShopeeCapture(), 0600); err != nil {
		t.Fatal(err)
	}
	var out, diag bytes.Buffer
	code := runWatcher([]string{"accesstrade-shopee-campaign-import", history, capture}, &out, &diag)
	var response map[string]any
	if err := json.Unmarshal(out.Bytes(), &response); err != nil || code != 0 || response["status"] != appendAdded || response["execution_permitted"] != false || response["network_fetch_performed"] != false || response["affiliate_link_created"] != false {
		t.Fatalf("selected-source CLI was not read-only canonical import: code=%d response=%v diagnostic=%s err=%v", code, response, diag.String(), err)
	}
	if code := runWatcher([]string{"accesstrade-shopee-campaign-import", history, capture}, &out, &diag); code != 0 || !strings.Contains(out.String(), appendDuplicate) {
		t.Fatalf("selected-source CLI retry was not exact duplicate: %d %s", code, out.String())
	}
}

// The selected-source capture remains a portable, untrusted input until its
// canonical history append ACK. A same-byte swap after open must therefore be
// rejected before its contents can be interpreted as a capture or persisted.
func TestSelectedAccesstradeCampaignCLIRejectsCaptureSymlinkSwapAfterOpen(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions are not portable on Windows")
	}
	dir := t.TempDir()
	history := filepath.Join(dir, "history.jsonl")
	capture := filepath.Join(dir, "sanitized-capture.json")
	contents := []byte(accesstradeShopeeCapture())
	if err := os.WriteFile(capture, contents, 0600); err != nil {
		t.Fatal(err)
	}
	external := filepath.Join(t.TempDir(), "same-byte-external-capture.json")
	if err := os.WriteFile(external, contents, 0600); err != nil {
		t.Fatal(err)
	}
	swapped := false
	stableRegularFileReadHook = func(openedPath string) error {
		if filepath.Clean(openedPath) != filepath.Clean(capture) || swapped {
			return nil
		}
		swapped = true
		if err := os.Remove(capture); err != nil {
			return err
		}
		return os.Symlink(external, capture)
	}
	t.Cleanup(func() { stableRegularFileReadHook = nil })
	var out, diag bytes.Buffer
	code := runWatcher([]string{"accesstrade-shopee-campaign-import", history, capture}, &out, &diag)
	var response map[string]any
	if err := json.Unmarshal(out.Bytes(), &response); err != nil || code == 0 || response["status"] != "INPUT_ERROR" || response["persisted"] != false {
		t.Fatalf("selected-source import accepted post-open capture swap: code=%d response=%v diagnostic=%s err=%v", code, response, diag.String(), err)
	}
	if !swapped {
		t.Fatal("selected-source import did not use the stable portable-input reader")
	}
	if _, err := os.Lstat(history); !os.IsNotExist(err) {
		t.Fatalf("selected-source import wrote history after capture swap: %v", err)
	}
	if got, err := os.ReadFile(external); err != nil || !bytes.Equal(got, contents) {
		t.Fatalf("capture swap rejection changed external bytes: %q err=%v", got, err)
	}
}

func TestHistoryHTTPReadAndHandoffFailClosedWhileWriterIsActive(t *testing.T) {
	dir := t.TempDir()
	history := filepath.Join(dir, "history.jsonl")
	record, err := watcherRecord(mustRawJSON(t, watchFixture()))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AppendHistory(history, record); err != nil {
		t.Fatal(err)
	}
	release, err := acquireHistoryRuntimeGate(history)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	read := httptest.NewRecorder()
	historyReadHandler(history).ServeHTTP(read, httptest.NewRequest(http.MethodGet, "/v1/history?record_id="+record.RecordID, nil))
	if read.Code != http.StatusConflict || !strings.Contains(read.Body.String(), `"BUSY"`) {
		t.Fatalf("history read did not fail closed: code=%d body=%s", read.Code, read.Body.String())
	}
	if _, _, err := appendResolvedHistory(history, record); err == nil {
		t.Fatal("append/resolution handoff succeeded during writer activity")
	}
}

func TestM06AdapterAndM08ResolveTheSameCanonicalFieldIDs(t *testing.T) {
	dir := t.TempDir()
	history := filepath.Join(dir, "history.jsonl")
	fixture := watchFixture()
	fixture.Body = `{"commission_rate":0.08,"price":100,"currency":"USD","product_name":"A&B Bé","product_id":"a"}`
	raw, err := json.Marshal(m06AdapterRequest{Fixture: mustRawJSON(t, fixture)})
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	m06AdapterHandler(history).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/m06/fixture-import", bytes.NewReader(raw)))
	if response.Code != http.StatusOK {
		t.Fatal(response.Code, response.Body.String())
	}
	records, err := LoadHistory(history)
	if err != nil || len(records) != 1 {
		t.Fatal(err, records)
	}
	record := records[0]
	context, err := m07EvidenceContext(record)
	if err != nil {
		t.Fatal(err)
	}
	var priceFieldID string
	for _, evidence := range context.Evidence {
		if evidence.FieldOrClaim == "price" {
			priceFieldID = evidence.EvidenceID
		}
	}
	if priceFieldID == "" {
		t.Fatal("M06 record did not expose its canonical price field ID")
	}
	requestPath, intentPath := filepath.Join(dir, "intent-request.json"), filepath.Join(dir, "intent.json")
	request := learnerIntentRequest{IntentID: "m06-intent", DecisionID: record.RecordID, EvidenceIDs: []string{priceFieldID}, ActionType: "DRAFT", Target: "https://example.com/draft", Parameters: map[string]any{}, ProposedBy: "human", CreatedAt: "2026-09-03T00:01:00Z", ExpiresAt: "2026-09-03T00:05:00Z", CorrelationID: "m06-correlation", IdempotencyKey: "m06-key"}
	writeRequest := func() {
		raw, err := json.Marshal(request)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(requestPath, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	writeRequest()
	if code, envelope := missionCall(t, "m08-intent", history, requestPath, intentPath); code != 0 || envelope["status"] != appendAdded {
		t.Fatalf("M08 did not accept resolved M06 field IDs: code=%d envelope=%+v", code, envelope)
	}
	request.IntentID, request.IdempotencyKey = "m06-forged", "m06-forged-key"
	request.EvidenceIDs = []string{record.RecordID + "#price"}
	writeRequest()
	if code, envelope := missionCall(t, "m08-intent", history, requestPath, filepath.Join(dir, "forged-intent.json")); code == 0 || envelope["status"] != "REJECTED" {
		t.Fatalf("M08 accepted a forged M06 field ID: code=%d envelope=%+v", code, envelope)
	}

	// A syntactically valid but replay-DRIFT record must be unusable by M08.
	record.RecordedResult.Reasons = []string{"tampered replay result"}
	tampered, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(history, append(tampered, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	request.IntentID, request.IdempotencyKey = "m06-drift", "m06-drift-key"
	request.EvidenceIDs = []string{priceFieldID}
	writeRequest()
	if code, envelope := missionCall(t, "m08-intent", history, requestPath, filepath.Join(dir, "drift-intent.json")); code == 0 || envelope["status"] != "REJECTED" {
		t.Fatalf("M08 accepted a replay-DRIFT M06 record: code=%d envelope=%+v", code, envelope)
	}
}

// The workflow-facing M06 adapter must not turn a post-append acknowledgement
// loss into a generic retryable handoff error. Its caller needs the same
// PUBLISHED_RECOVERY_REQUIRED boundary as the direct history-handoff command
// so an exact retry resolves the record without a second append.
func TestM06AdapterDisclosesVisibleAppendUncertainty(t *testing.T) {
	dir := t.TempDir()
	history := filepath.Join(dir, "history.jsonl")
	fixtureRaw, err := json.Marshal(m06AdapterRequest{Fixture: mustRawJSON(t, watchFixture())})
	if err != nil {
		t.Fatal(err)
	}
	canonicalHistoryAppend = func(_ store.History, path string, candidate HistoryRecord) (string, error) {
		if status, err := appendHistoryWith(store.JSONL{}, path, candidate); err != nil || status != appendAdded {
			return status, err
		}
		return appendPublished, &publishedAppendUncertainty{cause: errors.New("injected adapter acknowledgement loss after history append")}
	}
	t.Cleanup(func() { canonicalHistoryAppend = appendHistoryWith })
	call := func() *httptest.ResponseRecorder {
		response := httptest.NewRecorder()
		m06AdapterHandler(history).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/m06/fixture-import", bytes.NewReader(fixtureRaw)))
		return response
	}
	first := call()
	if first.Code != http.StatusInternalServerError {
		t.Fatalf("visible append uncertainty returned %d: %s", first.Code, first.Body.String())
	}
	var envelope map[string]any
	if err := json.Unmarshal(first.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope["status"] != appendPublished || envelope["canonical_history_ack"] != false || envelope["canonical_history_persisted"] != true || envelope["record"] == nil {
		t.Fatalf("M06 adapter hid visible append uncertainty: %+v", envelope)
	}
	if records, err := LoadHistory(history); err != nil || len(records) != 1 {
		t.Fatalf("M06 adapter visible record was not canonical: records=%+v err=%v", records, err)
	}
	canonicalHistoryAppend = appendHistoryWith
	second := call()
	if second.Code != http.StatusOK {
		t.Fatalf("exact retry returned %d: %s", second.Code, second.Body.String())
	}
	if err := json.Unmarshal(second.Body.Bytes(), &envelope); err != nil || envelope["status"] != appendDuplicate || envelope["canonical_history_ack"] != true || envelope["canonical_history_persisted"] != true {
		t.Fatalf("M06 adapter exact retry did not acknowledge canonical record: %+v err=%v", envelope, err)
	}
}

func TestM06FixtureImportDisclosesVisibleAppendUncertainty(t *testing.T) {
	dir := t.TempDir()
	history, input := filepath.Join(dir, "history.jsonl"), filepath.Join(dir, "fixture.json")
	fixtureRaw, err := json.Marshal(watchFixture())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(input, fixtureRaw, 0600); err != nil {
		t.Fatal(err)
	}
	canonicalHistoryAppend = func(_ store.History, path string, candidate HistoryRecord) (string, error) {
		if status, err := appendHistoryWith(store.JSONL{}, path, candidate); err != nil || status != appendAdded {
			return status, err
		}
		return appendPublished, &publishedAppendUncertainty{cause: errors.New("injected fixture-import acknowledgement loss after history append")}
	}
	t.Cleanup(func() { canonicalHistoryAppend = appendHistoryWith })
	var stdout, stderr bytes.Buffer
	if code := runWatcher([]string{"fixture-import", history, input}, &stdout, &stderr); code == 0 {
		t.Fatalf("visible append uncertainty returned success: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	var envelope map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope["status"] != appendPublished || envelope["persisted"] != true || envelope["artifact"] == nil {
		t.Fatalf("fixture import hid visible append uncertainty: %+v", envelope)
	}
	if records, err := LoadHistory(history); err != nil || len(records) != 1 {
		t.Fatalf("fixture import visible record was not canonical: records=%+v err=%v", records, err)
	}
	canonicalHistoryAppend = appendHistoryWith
	stdout.Reset()
	stderr.Reset()
	if code := runWatcher([]string{"fixture-import", history, input}, &stdout, &stderr); code != 0 {
		t.Fatalf("exact retry failed: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil || envelope["status"] != appendDuplicate || envelope["persisted"] != true {
		t.Fatalf("fixture import exact retry did not acknowledge canonical record: %+v err=%v", envelope, err)
	}
}

func TestM06PinnedFetchDisclosesVisibleAppendUncertainty(t *testing.T) {
	dir := t.TempDir()
	history := filepath.Join(dir, "history.jsonl")
	fixtureRaw, err := json.Marshal(watchFixture())
	if err != nil {
		t.Fatal(err)
	}
	watcherPinnedFetch = func(_ context.Context, _ *http.Client) ([]byte, error) {
		return fixtureRaw, nil
	}
	t.Cleanup(func() { watcherPinnedFetch = fetchPinnedWatcher })
	canonicalHistoryAppend = func(_ store.History, path string, candidate HistoryRecord) (string, error) {
		if status, err := appendHistoryWith(store.JSONL{}, path, candidate); err != nil || status != appendAdded {
			return status, err
		}
		return appendPublished, &publishedAppendUncertainty{cause: errors.New("injected pinned-fetch acknowledgement loss after history append")}
	}
	t.Cleanup(func() { canonicalHistoryAppend = appendHistoryWith })
	var stdout, stderr bytes.Buffer
	if code := runWatcherFetch([]string{"fetch-fixture", history}, &stdout, &stderr); code == 0 {
		t.Fatalf("visible append uncertainty returned success: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	var envelope map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope["status"] != appendPublished || envelope["persisted"] != true || envelope["artifact"] == nil || envelope["network_fetch_attempted"] != true {
		t.Fatalf("pinned fetch hid visible append uncertainty: %+v", envelope)
	}
	if records, err := LoadHistory(history); err != nil || len(records) != 1 {
		t.Fatalf("pinned fetch visible record was not canonical: records=%+v err=%v", records, err)
	}
	canonicalHistoryAppend = appendHistoryWith
	stdout.Reset()
	stderr.Reset()
	if code := runWatcherFetch([]string{"fetch-fixture", history}, &stdout, &stderr); code != 0 {
		t.Fatalf("exact retry failed: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil || envelope["status"] != appendDuplicate || envelope["persisted"] != true {
		t.Fatalf("pinned fetch exact retry did not acknowledge canonical record: %+v err=%v", envelope, err)
	}
}

func TestM07HTTPAdapterRegistersThenResolvesToolEvidence(t *testing.T) {
	dir := t.TempDir()
	history, input := filepath.Join(dir, "history.jsonl"), filepath.Join(dir, "fixture.json")
	watchRun(t, history, input, watchFixture(), "APPENDED")
	records, err := LoadHistory(history)
	if err != nil || len(records) != 1 {
		t.Fatal(err, records)
	}
	record := records[0]
	registry := []corem07.ToolSpec{{Name: "public_http", ReadOnly: true, AllowedMethods: []string{"GET"}, AllowedHosts: []string{"example.com"}, TimeoutMS: 1000, FollowRedirects: false}}
	wrongMethod := httptest.NewRecorder()
	m07AdapterHandler(history).ServeHTTP(wrongMethod, httptest.NewRequest(http.MethodGet, "/v1/m07/fetch-and-register", nil))
	if wrongMethod.Code != http.StatusMethodNotAllowed {
		t.Fatal("M07 adapter accepted a non-POST request", wrongMethod.Code)
	}
	call := func(path string, payload any) *httptest.ResponseRecorder {
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		recorder := httptest.NewRecorder()
		m07AdapterHandler(history).ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, path, bytes.NewReader(raw)))
		return recorder
	}
	preflight := call("/v1/m07/preflight", m07AdapterRequest{RecordID: record.RecordID, Registry: registry, ToolRequest: &corem07.ToolRequest{ToolName: "public_http", Method: "GET", Target: "https://example.com/a"}})
	if preflight.Code != http.StatusOK {
		t.Fatal(preflight.Code, preflight.Body.String())
	}
	writeAttempt := call("/v1/m07/preflight", m07AdapterRequest{RecordID: record.RecordID, Registry: registry, ToolRequest: &corem07.ToolRequest{ToolName: "public_http", Method: "POST", Target: "https://example.com/a"}})
	if writeAttempt.Code == http.StatusOK {
		t.Fatal("write request passed adapter preflight")
	}
	tool := corem07.ToolResult{RecordID: record.RecordID, ToolCall: corem07.ToolRequest{ToolName: "public_http", Method: "GET", Target: "https://example.com/a"}, StatusCode: 200, ReceivedAt: "2026-09-03T00:01:00Z", Body: json.RawMessage(`{"amount":9007199254740993}`)}
	registered := call("/v1/m07/register-tool-result", m07AdapterRequest{RecordID: record.RecordID, Registry: registry, ToolResult: mustRawJSON(t, tool)})
	if registered.Code != http.StatusOK {
		t.Fatal(registered.Code, registered.Body.String())
	}
	var registration struct {
		ArtifactID      string           `json:"artifact_id"`
		Evidence        corem07.Evidence `json:"evidence"`
		EvidenceRawJSON string           `json:"evidence_raw_json"`
	}
	if err := json.Unmarshal(registered.Body.Bytes(), &registration); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(registration.EvidenceRawJSON, "9007199254740993") {
		t.Fatalf("adapter rounded evidence before workflow handoff: %s", registration.EvidenceRawJSON)
	}
	claim := corem07.Claim{FieldOrClaim: registration.Evidence.FieldOrClaim, Value: json.RawMessage(`{"amount":9007199254740993}`), EvidenceIDs: []string{registration.Evidence.EvidenceID}}
	claim.Text = corem07.RenderGroundedAnswer([]corem07.Claim{claim})
	model := corem07.AgentOutput{State: "HUMAN_REVIEW", Answer: corem07.RenderGroundedAnswer([]corem07.Claim{claim}), Claims: []corem07.Claim{claim}, EvidenceIDs: []string{registration.Evidence.EvidenceID}, ToolCalls: []corem07.ToolRequest{}, Authority: "A2-RO", WritePermission: false, ProposedAction: &corem07.ProposedAction{ActionType: "DRAFT", Target: "https://example.com/draft", Parameters: json.RawMessage(`{}`)}}
	valid := call("/v1/m07/validate", m07AdapterRequest{RecordID: record.RecordID, Registry: registry, ModelOutputText: string(mustRawJSON(t, model)), ToolResultID: registration.ArtifactID})
	if valid.Code != http.StatusOK {
		t.Fatal(valid.Code, valid.Body.String())
	}
	proposal := call("/v1/m07/register-proposal", m07AdapterRequest{RecordID: record.RecordID, Registry: registry, ModelOutputText: string(mustRawJSON(t, model)), ToolResultID: registration.ArtifactID})
	if proposal.Code != http.StatusOK {
		t.Fatal(proposal.Code, proposal.Body.String())
	}
	var persisted struct {
		Status     string                          `json:"status"`
		ArtifactID string                          `json:"artifact_id"`
		Artifact   corem07.RegisteredAgentProposal `json:"artifact"`
	}
	if err := json.Unmarshal(proposal.Body.Bytes(), &persisted); err != nil || persisted.Status != "ACK" || persisted.ArtifactID == "" || persisted.Artifact.RecordID != record.RecordID {
		t.Fatalf("proposal did not receive a durable ACK: %s err=%v", proposal.Body.String(), err)
	}
	proposalPath, err := m07ArtifactPath(history, "proposals", persisted.ArtifactID)
	if err != nil {
		t.Fatal(err)
	}
	beforeProposal, err := os.ReadFile(proposalPath)
	if err != nil {
		t.Fatal(err)
	}
	model.Answer = "guaranteed profit"
	rejectedProposal := call("/v1/m07/register-proposal", m07AdapterRequest{RecordID: record.RecordID, Registry: registry, ModelOutput: mustRawJSON(t, model), ToolResultID: registration.ArtifactID})
	if rejectedProposal.Code == http.StatusOK {
		t.Fatal("ungrounded output was persisted as an M07 proposal")
	}
	afterProposal, err := os.ReadFile(proposalPath)
	if err != nil || !bytes.Equal(beforeProposal, afterProposal) {
		t.Fatal("rejected M07 proposal changed the persisted artifact", err)
	}
	forged := call("/v1/m07/validate", m07AdapterRequest{RecordID: record.RecordID, Registry: registry, ModelOutput: mustRawJSON(t, model), ToolResultID: "sha256:" + strings.Repeat("0", 64)})
	if forged.Code == http.StatusOK {
		t.Fatal("forged tool artifact id accepted")
	}
}

// The HTTP adapter must not describe a tool trace as merely unpersisted when
// the immutable sidecar is already visible but its directory-sync completion
// is uncertain. A retry for the exact trace is safe and produces the normal
// durable ACK without a second artifact.
func TestM07HTTPAdapterReportsUnconfirmedVisibleToolArtifact(t *testing.T) {
	dir := t.TempDir()
	history, input := filepath.Join(dir, "history.jsonl"), filepath.Join(dir, "fixture.json")
	watchRun(t, history, input, watchFixture(), appendAdded)
	records, err := LoadHistory(history)
	if err != nil || len(records) != 1 {
		t.Fatal(err, records)
	}
	registry := []corem07.ToolSpec{{Name: "public_http", ReadOnly: true, AllowedMethods: []string{"GET"}, AllowedHosts: []string{"example.com"}, TimeoutMS: 1000, FollowRedirects: false}}
	tool := corem07.ToolResult{RecordID: records[0].RecordID, ToolCall: corem07.ToolRequest{ToolName: "public_http", Method: "GET", Target: "https://example.com/a"}, StatusCode: 200, ReceivedAt: "2026-09-03T00:01:00Z", Body: json.RawMessage(`{"amount":100}`)}
	payload, err := json.Marshal(m07AdapterRequest{RecordID: records[0].RecordID, Registry: registry, ToolResult: mustRawJSON(t, tool)})
	if err != nil {
		t.Fatal(err)
	}
	artifactWriteFault = func(phase string) error {
		if phase == "after_publish_before_parent_sync" {
			return errors.New("injected tool artifact parent sync failure")
		}
		return nil
	}
	t.Cleanup(func() { artifactWriteFault = nil })
	first := httptest.NewRecorder()
	m07AdapterHandler(history).ServeHTTP(first, httptest.NewRequest(http.MethodPost, "/v1/m07/register-tool-result", bytes.NewReader(payload)))
	if first.Code != http.StatusInternalServerError || !strings.Contains(first.Body.String(), `"status":"PUBLISHED_RECOVERY_REQUIRED"`) {
		t.Fatalf("adapter did not disclose unconfirmed visible artifact: status=%d body=%s", first.Code, first.Body.String())
	}
	registered, err := corem07.RegisterToolResult(mustRawJSON(t, tool), registry)
	if err != nil {
		t.Fatal(err)
	}
	path, err := m07ArtifactPath(history, "tool-results", registered.TraceID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("tool artifact was not visible after unconfirmed publish: %v", err)
	}
	if !strings.Contains(first.Body.String(), `"artifact_id":"`+registered.TraceID+`"`) || !strings.Contains(first.Body.String(), `"artifact":`) {
		t.Fatalf("adapter did not disclose deterministic recovery artifact: %s", first.Body.String())
	}
	artifactWriteFault = nil
	retry := httptest.NewRecorder()
	m07AdapterHandler(history).ServeHTTP(retry, httptest.NewRequest(http.MethodPost, "/v1/m07/register-tool-result", bytes.NewReader(payload)))
	if retry.Code != http.StatusOK || !strings.Contains(retry.Body.String(), `"status":"ACK"`) {
		t.Fatalf("exact tool retry did not acknowledge visible artifact: status=%d body=%s", retry.Code, retry.Body.String())
	}
}

func TestM07HTTPAdapterReportsUnconfirmedVisibleProposal(t *testing.T) {
	dir := t.TempDir()
	history, input := filepath.Join(dir, "history.jsonl"), filepath.Join(dir, "fixture.json")
	watchRun(t, history, input, watchFixture(), appendAdded)
	records, err := LoadHistory(history)
	if err != nil || len(records) != 1 {
		t.Fatal(err, records)
	}
	ctx, err := m07EvidenceContext(records[0])
	if err != nil || len(ctx.Evidence) == 0 {
		t.Fatalf("missing canonical M07 evidence: %+v err=%v", ctx, err)
	}
	value, err := json.Marshal(ctx.Evidence[0].Value)
	if err != nil {
		t.Fatal(err)
	}
	claim := corem07.Claim{FieldOrClaim: ctx.Evidence[0].FieldOrClaim, Value: value, EvidenceIDs: []string{ctx.Evidence[0].EvidenceID}}
	claim.Text = corem07.RenderGroundedAnswer([]corem07.Claim{claim})
	registry := []corem07.ToolSpec{{Name: "public_http", ReadOnly: true, AllowedMethods: []string{"GET"}, AllowedHosts: []string{"example.com"}, TimeoutMS: 1000, FollowRedirects: false}}
	model := corem07.AgentOutput{State: "HUMAN_REVIEW", Answer: claim.Text, EvidenceIDs: []string{ctx.Evidence[0].EvidenceID}, Claims: []corem07.Claim{claim}, ToolCalls: []corem07.ToolRequest{}, Authority: "A2-RO", WritePermission: false, ProposedAction: &corem07.ProposedAction{ActionType: "DRAFT", Target: "https://example.com/draft", Parameters: json.RawMessage(`{}`)}}
	payload, err := json.Marshal(m07AdapterRequest{RecordID: records[0].RecordID, Registry: registry, ModelOutput: mustRawJSON(t, model)})
	if err != nil {
		t.Fatal(err)
	}
	artifactWriteFault = func(phase string) error {
		if phase == "after_publish_before_parent_sync" {
			return errors.New("injected proposal artifact parent sync failure")
		}
		return nil
	}
	t.Cleanup(func() { artifactWriteFault = nil })
	first := httptest.NewRecorder()
	m07AdapterHandler(history).ServeHTTP(first, httptest.NewRequest(http.MethodPost, "/v1/m07/register-proposal", bytes.NewReader(payload)))
	if first.Code != http.StatusInternalServerError || !strings.Contains(first.Body.String(), `"status":"PUBLISHED_RECOVERY_REQUIRED"`) {
		t.Fatalf("adapter did not disclose unconfirmed visible proposal: status=%d body=%s", first.Code, first.Body.String())
	}
	proposal, err := corem07.RegisterAgentProposal(mustRawJSON(t, model), ctx.Evidence, registry, records[0].RecordID)
	if err != nil {
		t.Fatal(err)
	}
	path, err := m07ArtifactPath(history, "proposals", proposal.ProposalID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("proposal artifact was not visible after unconfirmed publish: %v", err)
	}
	if !strings.Contains(first.Body.String(), `"artifact_id":"`+proposal.ProposalID+`"`) || !strings.Contains(first.Body.String(), `"artifact":`) {
		t.Fatalf("adapter did not disclose deterministic proposal recovery artifact: %s", first.Body.String())
	}
	artifactWriteFault = nil
	retry := httptest.NewRecorder()
	m07AdapterHandler(history).ServeHTTP(retry, httptest.NewRequest(http.MethodPost, "/v1/m07/register-proposal", bytes.NewReader(payload)))
	if retry.Code != http.StatusOK || !strings.Contains(retry.Body.String(), `"status":"ACK"`) {
		t.Fatalf("exact proposal retry did not acknowledge visible artifact: status=%d body=%s", retry.Code, retry.Body.String())
	}
}

// Tool-result IDs resolve to adapter-owned sidecar paths, not caller-supplied
// portable input. Reject a same-byte external symlink replacement after open
// before it can become evidence for grounding or proposal persistence.
func TestM07ToolArtifactRejectsSymlinkSwapAfterOpen(t *testing.T) {
	dir := t.TempDir()
	history, input := filepath.Join(dir, "history.jsonl"), filepath.Join(dir, "fixture.json")
	watchRun(t, history, input, watchFixture(), "APPENDED")
	records, err := LoadHistory(history)
	if err != nil || len(records) != 1 {
		t.Fatal(err, records)
	}
	record := records[0]
	registry := []corem07.ToolSpec{{Name: "public_http", ReadOnly: true, AllowedMethods: []string{"GET"}, AllowedHosts: []string{"example.com"}, TimeoutMS: 1000, FollowRedirects: false}}
	result := corem07.ToolResult{RecordID: record.RecordID, ToolCall: corem07.ToolRequest{ToolName: "public_http", Method: "GET", Target: "https://example.com/a"}, StatusCode: 200, ReceivedAt: "2026-09-03T00:01:00Z", Body: json.RawMessage(`{"amount":100}`)}
	registered, err := corem07.RegisterToolResult(mustRawJSON(t, result), registry)
	if err != nil {
		t.Fatal(err)
	}
	path, err := m07ArtifactPath(history, "tool-results", registered.TraceID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writeNewJSON(path, registered); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	external := filepath.Join(dir, "same-byte-external-tool-result.json")
	if err := os.WriteFile(external, contents, 0600); err != nil {
		t.Fatal(err)
	}
	swapped := false
	stableRegularFileReadHook = func(openedPath string) error {
		if filepath.Clean(openedPath) != filepath.Clean(path) || swapped {
			return nil
		}
		swapped = true
		if err := os.Remove(path); err != nil {
			return err
		}
		return os.Symlink(external, path)
	}
	t.Cleanup(func() { stableRegularFileReadHook = nil })

	if _, err := loadM07ToolArtifact(history, registered.TraceID, registry, record.RecordID); err == nil || !strings.Contains(err.Error(), "changed while reading stable regular file") {
		t.Fatalf("M07 tool loader accepted a same-byte symlink swap: %v", err)
	}
	if !swapped {
		t.Fatal("M07 tool loader did not reach stable-reader swap seam")
	}
}

func TestM07TransportRejectsPrivateDNSAndOversizedResponse(t *testing.T) {
	for _, raw := range []string{"127.0.0.1", "10.0.0.1", "100.64.0.1", "169.254.1.1", "::1", "fe80::1"} {
		if m07PublicAddress(net.ParseIP(raw)) {
			t.Fatalf("non-public address accepted: %s", raw)
		}
	}
	lookup := func(_ context.Context, _ string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("127.0.0.1")}}, nil
	}
	if _, err := m07ResolvePublicHost(context.Background(), "example.com", lookup); err == nil {
		t.Fatal("private resolver result accepted")
	}
	mixedLookup := func(_ context.Context, _ string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}, {IP: net.ParseIP("127.0.0.1")}}, nil
	}
	if _, err := m07ResolvePublicHost(context.Background(), "example.com", mixedLookup); err == nil {
		t.Fatal("mixed public/private resolver result accepted")
	}
	publicLookup := func(_ context.Context, _ string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}, nil
	}
	if _, err := m07ResolvePublicHost(context.Background(), "example.com", publicLookup); err != nil {
		t.Fatal(err)
	}
	if _, err := m07ToolResponseBody(bytes.NewReader(bytes.Repeat([]byte("x"), m07MaxToolResponseBytes+1))); err == nil {
		t.Fatal("oversized response accepted")
	}
	body, err := m07ToolResponseBody(bytes.NewBufferString("untrusted html"))
	if err != nil || string(body) != `"untrusted html"` {
		t.Fatalf("untrusted text was not retained as JSON data: %s %v", body, err)
	}
}

func TestM07AdapterRejectsReceivedRedirectWithoutFollowingOrPersisting(t *testing.T) {
	dir := t.TempDir()
	history, input := filepath.Join(dir, "history.jsonl"), filepath.Join(dir, "fixture.json")
	watchRun(t, history, input, watchFixture(), "APPENDED")
	records, err := LoadHistory(history)
	if err != nil || len(records) != 1 {
		t.Fatal(err, records)
	}
	registry := []corem07.ToolSpec{{Name: "public_http", ReadOnly: true, AllowedMethods: []string{"GET"}, AllowedHosts: []string{"example.com"}, TimeoutMS: 1000, FollowRedirects: false}}
	request := corem07.ToolRequest{ToolName: "public_http", Method: "GET", Target: "https://example.com/redirect"}
	var calls []string
	redirectTransport := roundTripperFunc(func(httpRequest *http.Request) (*http.Response, error) {
		calls = append(calls, httpRequest.URL.String())
		return &http.Response{StatusCode: http.StatusFound, Header: http.Header{"Location": []string{"https://evil.invalid/redirected"}}, Body: io.NopCloser(strings.NewReader("redirect")), Request: httpRequest}, nil
	})
	publicLookup := func(_ context.Context, _ string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}, nil
	}
	fetcher := func(ctx context.Context, recordID string, toolRequest corem07.ToolRequest, tools []corem07.ToolSpec, _ m07LookupIPAddr) (corem07.ToolResult, error) {
		return fetchM07ToolResultWithTransport(ctx, recordID, toolRequest, tools, publicLookup, redirectTransport)
	}
	raw, err := json.Marshal(m07AdapterRequest{RecordID: records[0].RecordID, Registry: registry, ToolRequest: &request})
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	m07AdapterHandlerWithFetcher(history, fetcher).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/m07/fetch-and-register", bytes.NewReader(raw)))
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "TOOL_TRANSPORT_REJECTED") {
		t.Fatalf("received redirect was not rejected by the M07 adapter: %d %s", response.Code, response.Body.String())
	}
	if len(calls) != 1 || calls[0] != request.Target {
		t.Fatalf("M07 followed a redirect or changed target: %v", calls)
	}
	toolStore := history + ".m07/tool-results"
	entries, err := os.ReadDir(toolStore)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("received redirect persisted tool artifacts: %v", entries)
	}
}

func mustRawJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
func TestWatcherRetryAndRestart(t *testing.T) {
	dir := t.TempDir()
	h, input := filepath.Join(dir, "history"), filepath.Join(dir, "fixture")
	f := watchFixture()
	watchRun(t, h, input, f, "APPENDED")
	before, _ := os.ReadFile(h)
	watchRun(t, h, input, f, "EXACT_DUPLICATE")
	f.ObservedAt = "2026-09-03T07:00:00+07:00"
	watchRun(t, h, input, f, "EXACT_DUPLICATE")
	now, _ := os.ReadFile(h)
	if !bytes.Equal(before, now) {
		t.Fatal("retry changed bytes")
	}
	f.Body = `{"product_id":"a","product_name":"Fixture A","currency":"USD","price":120,"commission_rate":0.08}`
	watchRun(t, h, input, f, "HANDOFF_ERROR")
	now, _ = os.ReadFile(h)
	if !bytes.Equal(before, now) {
		t.Fatal("conflict changed bytes")
	}
	f = watchFixture()
	f.ObservedAt = "2026-09-04T00:00:00Z"
	watchRun(t, h, input, f, "HANDOFF_ERROR")
	f.CorrelationID = "event-2"
	watchRun(t, h, input, f, "APPENDED")
	records, err := LoadHistory(h)
	if err != nil || len(records) != 2 {
		t.Fatal(err, records)
	}
	for _, r := range records {
		if Replay(r).State != "MATCH" || r.RecordedResult.State != "RANK_SCENARIO" {
			t.Fatal(r)
		}
		o := r.Observations[0]
		if o.EvidenceKind != "synthetic" || o.Price == nil || *o.Price != 100 {
			t.Fatal(o)
		}
	}
}
func TestWatcherRejectAndMissing(t *testing.T) {
	for _, kind := range []string{"method", "source", "status", "malformed", "unknown", "precision", "negative", "missing", "null"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			h, input := filepath.Join(dir, "history"), filepath.Join(dir, "fixture")
			f := watchFixture()
			want := "FIXTURE_ERROR"
			switch kind {
			case "method":
				f.Method = "POST"
			case "source":
				f.URL = "https://other.invalid"
			case "status":
				f.StatusCode = 503
			case "malformed":
				f.Body = "<html>error"
			case "unknown":
				f.Body = `{"extra":true}`
			case "precision":
				f.Body = `{"product_id":"a","product_name":"A","currency":"USD","price":9007199254740993,"commission_rate":0.08}`
			case "negative":
				f.Body = `{"product_id":"a","product_name":"A","currency":"USD","price":-1,"commission_rate":0.08}`
			case "missing":
				f.Body = `{"product_id":"a","product_name":"A","currency":"USD","price":100}`
				want = "APPENDED"
			case "null":
				f.Body = `{"product_id":"a","product_name":"A","currency":"USD","price":100,"commission_rate":null}`
				want = "APPENDED"
			}
			watchRun(t, h, input, f, want)
			if want == "APPENDED" {
				records, err := LoadHistory(h)
				if err != nil || records[0].RecordedResult.State != "GET_MORE_DATA" || records[0].Observations[0].CommissionRate != nil {
					t.Fatal(records, err)
				}
			} else {
				if _, err := os.Lstat(h); !os.IsNotExist(err) {
					t.Fatal("rejected created history", err)
				}
			}
		})
	}
}
func TestWatcherSinkAndAlias(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "fixture")
	f := watchFixture()
	watchRun(t, filepath.Join(dir, "missing-parent", "history"), input, f, "HANDOFF_ERROR")
	h := filepath.Join(dir, "history")
	if err := os.WriteFile(h, []byte("corrupt\n"), 0600); err != nil {
		t.Fatal(err)
	}
	watchRun(t, h, input, f, "HANDOFF_ERROR")
	b, _ := os.ReadFile(h)
	if string(b) != "corrupt\n" {
		t.Fatal("corruption overwritten")
	}
	watchRun(t, input, input, f, "PATH_ERROR")
}

func TestWatcherFetchPreflightDoesNotAttemptNetwork(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		setup func(t *testing.T, dir string) string
		want  string
	}{
		{name: "missing-parent", args: []string{"fetch-fixture"}, setup: func(t *testing.T, dir string) string {
			return filepath.Join(dir, "missing", "history")
		}, want: "PATH_ERROR"},
		{name: "corrupt-history", args: []string{"fetch-fixture"}, setup: func(t *testing.T, dir string) string {
			path := filepath.Join(dir, "history")
			if err := os.WriteFile(path, []byte("not-json\n"), 0600); err != nil {
				t.Fatal(err)
			}
			return path
		}, want: "HISTORY_ERROR"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			history := tc.setup(t, dir)
			var out, diag bytes.Buffer
			code := runWatcher(append(tc.args, history), &out, &diag)
			var env map[string]any
			if err := json.Unmarshal(out.Bytes(), &env); err != nil {
				t.Fatal(err, out.String(), diag.String())
			}
			if code == 0 || env["status"] != tc.want || env["network_fetch_attempted"] != false || env["persisted"] != false {
				t.Fatalf("code=%d env=%v diag=%s", code, env, diag.String())
			}
		})
	}
}
