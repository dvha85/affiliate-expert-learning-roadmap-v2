package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const actionJSON = `{"action_id":"a","decision_id":"d","action_type":"PUBLISH_REVIEW","target":"https://example.com","performed_by":"human","performed_at":"2026-09-03T08:00:00Z","measurement_window_end":"2026-09-04T08:00:00Z","compliance_reviewed":true}`
const outcomeJSON = `{"outcome_id":"o","effect_ref":{"effect_kind":"HUMAN_ACTION","effect_id":"a"},"observed_at":"2026-09-04T08:00:00Z","status":"NO_OBSERVED_OUTCOME","metrics":{},"source_ref":"synthetic-fixture"}`

func TestActionValidateReadOnly(t *testing.T) {
	for _, tc := range []struct {
		name, raw, status string
		code              int
	}{
		{"valid", outcomeJSON, "VALID", 0},
		{"schema", `{}`, "INVALID_SCHEMA", 1},
		{"early", string(bytes.ReplaceAll([]byte(outcomeJSON), []byte("2026-09-04T08:00:00Z"), []byte("2026-09-03T09:00:00Z"))), "MEASUREMENT_WINDOW_OPEN", 1},
		{"duplicate", `{"outcome_id":"x","outcome_id":"y"}`, "INVALID_SCHEMA", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			a := filepath.Join(dir, "a.json")
			o := filepath.Join(dir, "o.json")
			if e := os.WriteFile(a, []byte(actionJSON), 0600); e != nil {
				t.Fatal(e)
			}
			if e := os.WriteFile(o, []byte(tc.raw), 0600); e != nil {
				t.Fatal(e)
			}
			var stdout, stderr bytes.Buffer
			code := RunAction([]string{"validate", a, o}, &stdout, &stderr)
			var r actionEnvelope
			if e := json.Unmarshal(stdout.Bytes(), &r); e != nil {
				t.Fatal(e)
			}
			if code != tc.code || r.Status != tc.status || r.ExecutionPermitted || (r.Artifact != nil) != (code == 0) {
				t.Fatal(code, r)
			}
			if (stderr.Len() == 0) != (code == 0) {
				t.Fatal(stderr.String())
			}
			b, _ := os.ReadFile(a)
			c, _ := os.ReadFile(o)
			entries, _ := os.ReadDir(dir)
			if string(b) != actionJSON || string(c) != tc.raw || len(entries) != 2 {
				t.Fatal("mutated input/store")
			}
		})
	}
}

func TestActionErrors(t *testing.T) {
	for _, tc := range []struct {
		args   []string
		code   int
		status string
	}{{nil, 2, "USAGE_ERROR"}, {[]string{"record"}, 2, "USAGE_ERROR"}, {[]string{"validate", "/does-not-exist/action", "/does-not-exist/outcome"}, 1, "IO_ERROR"}} {
		var out, err bytes.Buffer
		if code := RunAction(tc.args, &out, &err); code != tc.code {
			t.Fatal(code)
		}
		var r actionEnvelope
		if e := json.Unmarshal(out.Bytes(), &r); e != nil {
			t.Fatal(e)
		}
		if r.Status != tc.status || r.Artifact != nil || err.Len() == 0 {
			t.Fatal(r)
		}
	}
}
