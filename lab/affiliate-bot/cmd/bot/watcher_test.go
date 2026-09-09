package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	corem07 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m07"
)

func watchFixture() watcherFixture {
	return watcherFixture{Version: "br13-offer-fixture/v1", Method: "GET", URL: watcherFixtureURL, ObservedAt: "2026-09-03T00:00:00Z", CorrelationID: "event-1", StatusCode: 200, Body: `{"product_id":"a","product_name":"Fixture A","currency":"USD","price":100,"commission_rate":0.08}`}
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
