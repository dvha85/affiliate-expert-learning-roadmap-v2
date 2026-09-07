package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
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
