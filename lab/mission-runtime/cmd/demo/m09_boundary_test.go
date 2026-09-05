package main

import (
	"bytes"
	"encoding/json"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func m09Artifacts(t *testing.T) []any {
	t.Helper()
	s, c := baseM09()
	a, status := AuthorizeM09(s, c)
	if status != "AUTHORIZED" {
		t.Fatal(status)
	}
	r := executionFailureRecord(s, a, "2026-09-03T08:01:00Z", "CANCELLED", "NOT_PERFORMED", "", "")
	return []any{s.Intent, s.Policy, *s.Approval, a, r}
}
func m09Raw(t *testing.T) [][]byte {
	t.Helper()
	raw := [][]byte{}
	for _, v := range m09Artifacts(t) {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		raw = append(raw, b)
	}
	return raw
}
func checkM09Raw(raw [][]byte) (M09CheckSummary, string) {
	return CheckM09Chain(raw[0], raw[1], raw[2], raw[3], raw[4])
}

func TestM09RawSchema(t *testing.T) {
	raw := m09Raw(t)
	for _, tc := range []struct {
		index  int
		decode func([]byte) string
	}{
		{2, func(b []byte) string { _, s := DecodeM09Approval(b); return s }},
		{3, func(b []byte) string { _, s := DecodeM09Authorization(b); return s }},
		{4, func(b []byte) string { _, s := DecodeM09Execution(b); return s }},
	} {
		if state := tc.decode(raw[tc.index]); state != missionValid {
			t.Fatal(state)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw[tc.index], &fields); err != nil {
			t.Fatal(err)
		}
		for key := range fields {
			for _, mutation := range []string{"missing", "null"} {
				t.Run(key+"/"+mutation, func(t *testing.T) {
					copy := map[string]json.RawMessage{}
					for k, v := range fields {
						copy[k] = v
					}
					if mutation == "missing" {
						delete(copy, key)
					} else {
						copy[key] = json.RawMessage(`null`)
					}
					bad, err := json.Marshal(copy)
					if err != nil {
						t.Fatal(err)
					}
					if tc.decode(bad) != "INVALID_SCHEMA" {
						t.Fatal("required/null accepted", key)
					}
				})
			}
		}
		for _, bad := range []string{`null`, string(raw[tc.index]) + ` {}`, strings.Replace(string(raw[tc.index]), `{`, `{"extra":true,`, 1), strings.Replace(string(raw[tc.index]), `"intent_id":`, `"intent_id":"other","intent_id":`, 1)} {
			if tc.decode([]byte(bad)) != "INVALID_SCHEMA" {
				t.Fatal("ambiguous JSON accepted")
			}
		}
	}
	for _, tc := range []struct {
		index    int
		old, new string
	}{{2, `"human"`, `"agent"`}, {2, `"one_time":true`, `"one_time":false`}, {3, `"execution_authorized":true`, `"execution_authorized":false`}, {3, `APPROVED_LIVE`, `GOVERNED_CANARY`}, {4, `CANCELLED`, `SUCCEEDED`}} {
		copy := m09Raw(t)
		copy[tc.index] = []byte(strings.Replace(string(copy[tc.index]), tc.old, tc.new, 1))
		if _, state := checkM09Raw(copy); state != "INVALID_SCHEMA" {
			t.Fatal(tc, state)
		}
	}
}

func TestM09FileBindings(t *testing.T) {
	for _, tc := range []struct {
		name               string
		index              int
		field, value, want string
	}{
		{"approval_hash", 2, "intent_hash", "sha256:" + strings.Repeat("0", 64), "BROKEN_LINK"},
		{"rejected", 2, "decision", "REJECT", "REJECTED_APPROVAL"},
		{"approval_policy", 2, "policy_version", "other", "BROKEN_LINK"},
		{"approval_correlation", 2, "correlation_id", "other", "BROKEN_LINK"},
		{"wrong_approval", 3, "approval_id", "other", "BROKEN_LINK"},
		{"wrong_intent", 3, "intent_id", "other", "BROKEN_LINK"},
		{"wrong_auth", 4, "authorization_id", "other", "BROKEN_LINK"},
		{"wrong_executor", 4, "executor_id", "other", "BROKEN_LINK"},
		{"wrong_key", 4, "idempotency_key", "other", "BROKEN_LINK"},
		{"tamper", 0, "target", "https://example.com/changed", "TAMPERED_INTENT"},
		{"approval_before_policy", 2, "approved_at", "2026-09-03T07:00:00Z", "INVALID_TIME_BINDING"},
		{"expired_approval_at_authorize", 2, "expires_at", "2026-09-03T07:30:00Z", "INVALID_TIME_BINDING"},
		{"auth_wider_than_intent", 3, "expires_at", "2026-09-03T10:00:00Z", "INVALID_TIME_BINDING"},
		{"execution_before_auth", 4, "attempted_at", "2026-09-03T07:00:00Z", "INVALID_TIME_BINDING"},
		{"blank_approver", 2, "approver_id", " ", missionInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := m09Raw(t)
			var fields map[string]any
			json.Unmarshal(raw[tc.index], &fields)
			fields[tc.field] = tc.value
			raw[tc.index], _ = json.Marshal(fields)
			summary, state := checkM09Raw(raw)
			if state != tc.want || summary.ExecutionPermitted {
				t.Fatal(state, tc.want)
			}
		})
	}
	raw := m09Raw(t)
	s, state := checkM09Raw(raw)
	if state != missionValid || s.Result != "CONSISTENT_UNVERIFIED" || s.ApprovalAuthenticated || s.ExecutionPermitted {
		t.Fatal(s, state)
	}
	// A forged human identity cannot be authenticated from these files. The audit
	// must not claim otherwise, even when all links are structurally consistent.
	raw[2] = []byte(strings.Replace(string(raw[2]), `"approver_id":"u"`, `"approver_id":"unverified-person"`, 1))
	s, state = checkM09Raw(raw)
	if state != missionValid || s.ApprovalAuthenticated || s.ExecutionPermitted {
		t.Fatal(s, state)
	}
	// Re-reading is an audit, not idempotent execution authorization.
	again, state := checkM09Raw(raw)
	if state != missionValid || again != s {
		t.Fatal("audit replay changed authority")
	}
}

func TestM09FileWindow(t *testing.T) {
	for _, tc := range []struct{ at, want string }{{"2026-09-03T08:29:59Z", missionValid}, {"2026-09-03T15:29:59+07:00", missionValid}, {"2026-09-03T08:30:00Z", "EXPIRED_AUTHORIZATION"}} {
		raw := m09Raw(t)
		var r ExecutionRecord
		json.Unmarshal(raw[4], &r)
		r.Status = "SUCCEEDED"
		r.SideEffectState = "PERFORMED"
		r.AttemptedAt = tc.at
		raw[4], _ = json.Marshal(r)
		if _, state := checkM09Raw(raw); state != tc.want {
			t.Fatal(state, tc.want)
		}
	}
}

func TestM09RuntimeSerializedArtifacts(t *testing.T) {
	// Exercise existing runtime outputs, without invoking any executor here.
	raw := m09Raw(t)
	for n, schema := range []string{"action-intent.schema.json", "policy-decision.schema.json", "approval-record.schema.json", "execution-authorization.schema.json", "execution-record.schema.json"} {
		if err := contracts.ValidateRaw(schema, raw[n]); err != nil {
			t.Fatal(schema, err)
		}
	}
}

func TestM09AuditCLI(t *testing.T) {
	raw := m09Raw(t)
	dir := t.TempDir()
	paths := []string{}
	for n, name := range []string{"intent", "policy", "approval", "authorization", "execution"} {
		path := filepath.Join(dir, name+".json")
		if err := os.WriteFile(path, raw[n], 0600); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	var out bytes.Buffer
	if err := runM09Check(&out, paths); err != nil {
		t.Fatal(err)
	}
	var summary M09CheckSummary
	if err := contracts.DecodeStrict(out.Bytes(), &summary); err != nil {
		t.Fatal(err)
	}
	if summary.ExecutionPermitted || summary.ApprovalAuthenticated || summary.Result != "CONSISTENT_UNVERIFIED" {
		t.Fatal(summary)
	}
	if bytes.Contains(out.Bytes(), []byte(`"execution_authorized"`)) {
		t.Fatal("authorization artifact emitted")
	}
	if err := runM09Check(m03BrokenWriter{}, paths); err == nil {
		t.Fatal("writer error ignored")
	}
	for n, path := range paths {
		got, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(got, raw[n]) {
			t.Fatal("file changed")
		}
	}
	if err := os.WriteFile(paths[2], []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := runM09Check(&out, paths); err == nil || out.Len() != 0 {
		t.Fatal("bad input emitted summary")
	}
	for _, args := range [][]string{nil, paths[:4], {"missing", paths[1], paths[2], paths[3], paths[4]}} {
		out.Reset()
		if err := runM09Check(&out, args); err == nil || out.Len() != 0 {
			t.Fatal("bad invocation")
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 5 {
		t.Fatal("unexpected persisted state")
	}
}

func TestM09CommittedAuditFixture(t *testing.T) {
	var out bytes.Buffer
	if err := runM09Check(&out, []string{"../../testdata/m08-intent.json", "../../testdata/m09-policy.json", "../../testdata/m09-approval.json", "../../testdata/m09-authorization.json", "../../testdata/m09-execution.json"}); err != nil {
		t.Fatal(err)
	}
	var summary M09CheckSummary
	if err := json.Unmarshal(out.Bytes(), &summary); err != nil {
		t.Fatal(err)
	}
	if summary.Result != "CONSISTENT_UNVERIFIED" || summary.ExecutionPermitted || summary.ApprovalAuthenticated || summary.ExecutionStatus != "CANCELLED" || summary.SideEffectState != "NOT_PERFORMED" {
		t.Fatal(summary)
	}
}
