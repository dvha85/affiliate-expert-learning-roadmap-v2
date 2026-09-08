package main

import (
	"bytes"
	"encoding/json"
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
	tool := corem07.ToolResult{RecordID: record.RecordID, ToolCall: corem07.ToolRequest{ToolName: "public_http", Method: "GET", Target: "https://example.com/a"}, StatusCode: 200, ReceivedAt: "2026-09-03T00:01:00Z", Body: json.RawMessage(`{"price":100}`)}
	registered := call("/v1/m07/register-tool-result", m07AdapterRequest{RecordID: record.RecordID, Registry: registry, ToolResult: mustRawJSON(t, tool)})
	if registered.Code != http.StatusOK {
		t.Fatal(registered.Code, registered.Body.String())
	}
	var registration struct {
		ArtifactID string           `json:"artifact_id"`
		Evidence   corem07.Evidence `json:"evidence"`
	}
	if err := json.Unmarshal(registered.Body.Bytes(), &registration); err != nil {
		t.Fatal(err)
	}
	claim := corem07.Claim{FieldOrClaim: registration.Evidence.FieldOrClaim, Value: json.RawMessage(`{"price":100}`), EvidenceIDs: []string{registration.Evidence.EvidenceID}}
	claim.Text = corem07.RenderGroundedAnswer([]corem07.Claim{claim})
	model := corem07.AgentOutput{State: "HUMAN_REVIEW", Answer: corem07.RenderGroundedAnswer([]corem07.Claim{claim}), Claims: []corem07.Claim{claim}, EvidenceIDs: []string{registration.Evidence.EvidenceID}, ToolCalls: []corem07.ToolRequest{}, Authority: "A2-RO", WritePermission: false, ProposedAction: &corem07.ProposedAction{ActionType: "DRAFT", Target: "https://example.com/draft", Parameters: json.RawMessage(`{}`)}}
	valid := call("/v1/m07/validate", m07AdapterRequest{RecordID: record.RecordID, Registry: registry, ModelOutput: mustRawJSON(t, model), ToolResultID: registration.ArtifactID})
	if valid.Code != http.StatusOK {
		t.Fatal(valid.Code, valid.Body.String())
	}
	proposal := call("/v1/m07/register-proposal", m07AdapterRequest{RecordID: record.RecordID, Registry: registry, ModelOutput: mustRawJSON(t, model), ToolResultID: registration.ArtifactID})
	if proposal.Code != http.StatusOK {
		t.Fatal(proposal.Code, proposal.Body.String())
	}
	forged := call("/v1/m07/validate", m07AdapterRequest{RecordID: record.RecordID, Registry: registry, ModelOutput: mustRawJSON(t, model), ToolResultID: "sha256:" + strings.Repeat("0", 64)})
	if forged.Code == http.StatusOK {
		t.Fatal("forged tool artifact id accepted")
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
