package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func m10ArtifactRaw(t *testing.T) map[string][]byte {
	t.Helper()
	s, c := baseM10()
	a, g, state := AuthorizeCanary(s, c)
	if state != "AUTHORIZED" {
		t.Fatal(state)
	}
	r := canaryExecutionRecord(s, a, c.Now, "CANCELLED", "NOT_PERFORMED", "", "")
	p := CanaryGrantApproval{s.Grant.ApprovalRef, s.Grant.GrantID, s.Grant.GrantVersion, s.Grant.GrantHash, "APPROVE", "human", s.Grant.ApproverID, s.Grant.ApprovedAt}
	out := map[string][]byte{}
	for k, v := range map[string]any{"grant": s.Grant, "approval": p, "cost": s.CostBound, "ledger": s.Ledger, "gate": g, "authorization": a, "execution": r} {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		out[k] = b
	}
	return out
}

func TestM10RawArtifacts(t *testing.T) {
	for kind, raw := range m10ArtifactRaw(t) {
		t.Run(kind, func(t *testing.T) {
			if _, state := DecodeM10Artifact(kind, raw); state != missionValid {
				t.Fatal(state, string(raw))
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(raw, &fields); err != nil {
				t.Fatal(err)
			}
			for key := range fields {
				duplicate := []byte(`{"` + key + `":` + string(fields[key]) + `,` + string(raw[1:]))
				if _, state := DecodeM10Artifact(kind, duplicate); state != "INVALID_SCHEMA" {
					t.Fatal("duplicate accepted", key)
				}
				for _, mutation := range []string{"missing", "null", "wrong_type"} {
					t.Run(key+"/"+mutation, func(t *testing.T) {
						copy := map[string]json.RawMessage{}
						for k, v := range fields {
							copy[k] = v
						}
						switch mutation {
						case "missing":
							delete(copy, key)
						case "null":
							copy[key] = json.RawMessage(`null`)
						default:
							copy[key] = json.RawMessage(`{}`)
						}
						bad, err := json.Marshal(copy)
						if err != nil {
							t.Fatal(err)
						}
						if v, state := DecodeM10Artifact(kind, bad); state != "INVALID_SCHEMA" || v != nil {
							t.Fatal(state, key)
						}
					})
				}
			}
			for _, bad := range []string{`null`, string(raw) + ` {}`, strings.Replace(string(raw), `{`, `{"unknown":true,`, 1), strings.Replace(string(raw), `{`, `{"unknown":true,"unknown":false,`, 1)} {
				if _, state := DecodeM10Artifact(kind, []byte(bad)); state != "INVALID_SCHEMA" {
					t.Fatal(state)
				}
			}
		})
	}
}

func TestM10RawMutations(t *testing.T) {
	for _, tc := range []struct {
		kind, key string
		value     any
		want      string
	}{
		{"grant", "max_executions_total", 99, "TAMPERED_GRANT"},
		{"grant", "approved_by", "agent", "INVALID_SCHEMA"},
		{"grant", "kill_switch_required", false, "INVALID_SCHEMA"},
		{"grant", "allowed_risk_classes", []string{"RISK2"}, "INVALID_SCHEMA"},
		{"cost", "max_cost_minor", 101, "TAMPERED_COST_BOUND"},
		{"cost", "max_cost_minor", -1, "INVALID_SCHEMA"},
		{"cost", "currency", "usd", "INVALID_SCHEMA"},
		{"approval", "decision", "REJECT", "INVALID_SCHEMA"},
		{"gate", "execution_authorized", true, "INVALID_SCHEMA"},
		{"gate", "per_action_approval_required", true, "INVALID_SCHEMA"},
		{"authorization", "execution_mode", "APPROVED_LIVE", "INVALID_SCHEMA"},
		{"authorization", "execution_authorized", false, "INVALID_SCHEMA"},
		{"authorization", "expires_at", "2026-09-03T08:00:00Z", "INVALID_TIME_BINDING"},
		{"execution", "status", "SUCCEEDED", "INVALID_SCHEMA"},
		{"execution", "approval_id", "other-stage", "INVALID_SCHEMA"},
		{"ledger", "pending_outcomes", 1, "INVALID_LEDGER"},
		{"ledger", "executions_in_window", 1, "INVALID_LEDGER"},
		{"ledger", "updated_at", "2026-09-03T07:00:00Z", "INVALID_TIME_BINDING"},
		{"ledger", "pending_execution_ids", []string{"x", "x"}, "INVALID_SCHEMA"},
	} {
		t.Run(tc.kind+"/"+tc.key, func(t *testing.T) {
			var fields map[string]any
			json.Unmarshal(m10ArtifactRaw(t)[tc.kind], &fields)
			fields[tc.key] = tc.value
			bad, _ := json.Marshal(fields)
			if v, state := DecodeM10Artifact(tc.kind, bad); state != tc.want || v != nil {
				t.Fatal(state, tc.want)
			}
		})
	}
}

func TestM10RawHashesAndNumbers(t *testing.T) {
	s, _ := baseM10()
	s.CostBound.MaxCostMinor = 9007199254740993
	s.CostBound = SealCanaryCostBound(s.CostBound)
	raw, _ := json.Marshal(s.CostBound)
	v, state := DecodeM10Artifact("cost", raw)
	if state != missionValid || v.(*CanaryCostBound).MaxCostMinor != 9007199254740993 {
		t.Fatal(v, state)
	}
	for _, value := range []string{"9223372036854775808", "1.5"} {
		bad := bytes.Replace(raw, []byte("9007199254740993"), []byte(value), 1)
		if _, state := DecodeM10Artifact("cost", bad); state != "INVALID_SCHEMA" {
			t.Fatal(state)
		}
	}
	// A recomputed hash does not excuse a reversed time window.
	s.CostBound.ExpiresAt = s.CostBound.ObservedAt
	s.CostBound = SealCanaryCostBound(s.CostBound)
	raw, _ = json.Marshal(s.CostBound)
	if _, state := DecodeM10Artifact("cost", raw); state != "INVALID_TIME_BINDING" {
		t.Fatal(state)
	}
	s.Grant.ValidFrom = "2026-09-03T14:10:00+07:00"
	s.Grant = SealCanaryGrant(s.Grant)
	raw, _ = json.Marshal(s.Grant)
	if _, state := DecodeM10Artifact("grant", raw); state != missionValid {
		t.Fatal(state)
	}
	s.Grant.ValidFrom = s.Grant.ExpiresAt
	s.Grant = SealCanaryGrant(s.Grant)
	raw, _ = json.Marshal(s.Grant)
	if _, state := DecodeM10Artifact("grant", raw); state != "INVALID_TIME_BINDING" {
		t.Fatal(state)
	}
}

func TestM10LedgerWireShape(t *testing.T) {
	s, _ := baseM10()
	before := s.Ledger
	raw, err := json.Marshal(s.Ledger)
	if err != nil || !reflect.DeepEqual(before, s.Ledger) {
		t.Fatal("marshal mutated ledger", err)
	}
	if bytes.Contains(raw, []byte(`:null`)) {
		t.Fatal(string(raw))
	}
	var l CanaryLedger
	json.Unmarshal(raw, &l)
	l.ExecutionsTotal = 1
	l.OutcomeLinks = []CanaryOutcomeLink{{"o", "e", l.UpdatedAt}, {"o", "e", l.UpdatedAt}}
	raw, _ = json.Marshal(l)
	if _, state := DecodeM10Artifact("ledger", raw); state != "INVALID_LEDGER" {
		t.Fatal(state)
	}
}

func TestM10AuditCLI(t *testing.T) {
	var fixture bytes.Buffer
	if err := runM10Check(&fixture, []string{"approval", "../../testdata/m10-approval.json"}); err != nil {
		t.Fatal(err)
	}
	var summary M10CheckSummary
	if err := json.Unmarshal(fixture.Bytes(), &summary); err != nil {
		t.Fatal(err)
	}
	if summary.Result != "ARTIFACT_VALID_UNVERIFIED" || summary.ProvenanceAuthenticated || summary.ChainValidated || summary.ExecutionPermitted {
		t.Fatal(summary)
	}
	dir := t.TempDir()
	for kind, raw := range m10ArtifactRaw(t) {
		path := filepath.Join(dir, kind+".json")
		if err := os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
		var out bytes.Buffer
		for n := 0; n < 2; n++ {
			out.Reset()
			if err := runM10Check(&out, []string{kind, path}); err != nil {
				t.Fatal(err)
			}
			var summary M10CheckSummary
			if err := json.Unmarshal(out.Bytes(), &summary); err != nil {
				t.Fatal(err)
			}
			if summary.Result != "ARTIFACT_VALID_UNVERIFIED" || summary.ArtifactType != kind || summary.ExecutionPermitted || summary.ProvenanceAuthenticated || summary.ChainValidated {
				t.Fatal(summary)
			}
			if bytes.Contains(out.Bytes(), []byte(`"execution_authorized"`)) {
				t.Fatal("emitted authorization")
			}
		}
		if err := runM10Check(m03BrokenWriter{}, []string{kind, path}); err == nil {
			t.Fatal("writer")
		}
		got, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(got, raw) {
			t.Fatal("changed input")
		}
		for _, args := range [][]string{nil, {kind}, {"unknown", path}, {kind, "missing"}} {
			out.Reset()
			if err := runM10Check(&out, args); err == nil || out.Len() != 0 {
				t.Fatal("bad args")
			}
		}
		os.WriteFile(path, []byte(`{}`), 0600)
		out.Reset()
		if err := runM10Check(&out, []string{kind, path}); err == nil || out.Len() != 0 {
			t.Fatal("bad input emitted output")
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 7 {
		t.Fatal("unexpected state")
	}
}
