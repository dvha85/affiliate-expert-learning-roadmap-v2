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

func m11Raw(t *testing.T) map[string][]byte {
	t.Helper()
	s, c := baseM11()
	a, g, status := AuthorizeProduction(s, c)
	if status != "AUTHORIZED" {
		t.Fatal(status, g)
	}
	r := productionExecutionRecord(s, a, c.Now, "CANCELLED", "NOT_PERFORMED", "", "")
	activation := productionActivationRecord{s.Lease.LeaseID, s.Lease.LeaseVersion, s.Lease.LeaseHash, c.Now}
	resolution := ProductionReconciliationResolution{ResolutionID: "synthetic-resolution", LeaseID: s.Lease.LeaseID, LeaseVersion: s.Lease.LeaseVersion, LeaseHash: s.Lease.LeaseHash, ExecutionID: r.ExecutionID, ResolvedBy: "human", ResolverID: "unverified", ResolvedAt: c.Now, EffectState: "NOT_PERFORMED", Reason: "synthetic claim"}
	cycle := ProductionCycleRecord{CycleID: "synthetic-cycle", LeaseID: s.Lease.LeaseID, LeaseVersion: s.Lease.LeaseVersion, LeaseHash: s.Lease.LeaseHash, ObservationIDs: []string{"e"}, DecisionID: s.Intent.DecisionID, IntentID: s.Intent.IntentID, IntentHash: s.Intent.IntentHash, GateID: g.GateID, AuthorizationID: a.AuthorizationID, ExecutionID: r.ExecutionID, OutcomeID: "unresolved-outcome", EvaluationID: "unresolved-evaluation", Status: "STOPPED", OpenedAt: c.Now, ClosedAt: c.Now, CorrelationID: s.Intent.CorrelationID}
	out := map[string][]byte{}
	for k, v := range map[string]any{"lease": s.Lease, "approval": c.TrustedLeaseApprovals[s.Lease.ApprovalRef], "health": s.Health, "cost": s.CostBound, "ledger": s.Ledger, "gate": g, "authorization": a, "execution": r, "activation": activation, "resolution": resolution, "cycle": cycle} {
		b, e := json.Marshal(v)
		if e != nil {
			t.Fatal(e)
		}
		out[k] = b
	}
	return out
}

func TestM11RawSchema(t *testing.T) {
	for kind, raw := range m11Raw(t) {
		t.Run(kind, func(t *testing.T) {
			if _, s := DecodeM11Artifact(kind, raw); s != missionValid {
				t.Fatal(s, string(raw))
			}
			var fields map[string]json.RawMessage
			if e := json.Unmarshal(raw, &fields); e != nil {
				t.Fatal(e)
			}
			for key := range fields {
				dup := []byte(`{"` + key + `":` + string(fields[key]) + `,` + string(raw[1:]))
				if _, s := DecodeM11Artifact(kind, dup); s != "INVALID_SCHEMA" {
					t.Fatal("duplicate", key, s)
				}
				for _, mutation := range []string{"missing", "null", "type"} {
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
						bad, e := json.Marshal(copy)
						if e != nil {
							t.Fatal(e)
						}
						if v, s := DecodeM11Artifact(kind, bad); s != "INVALID_SCHEMA" || v != nil {
							t.Fatal(key, s)
						}
					})
				}
			}
			for _, bad := range []string{`null`, string(raw) + ` {}`, strings.Replace(string(raw), `{`, `{"extra":true,`, 1)} {
				if _, s := DecodeM11Artifact(kind, []byte(bad)); s != "INVALID_SCHEMA" {
					t.Fatal(s)
				}
			}
		})
	}
}

func TestM11RawMutations(t *testing.T) {
	for _, tc := range []struct {
		kind, key string
		value     any
		want      string
	}{
		{"lease", "max_cost_minor_total", 999, "TAMPERED_LEASE"}, {"lease", "reviewed_by", "agent", "INVALID_SCHEMA"}, {"lease", "kill_switch_required", false, "INVALID_SCHEMA"},
		{"approval", "source_e5_refs", []string{}, "INVALID_SCHEMA"}, {"approval", "validated_risk_classes", []string{"RISK2"}, "INVALID_SCHEMA"},
		{"health", "telemetry_complete", false, "TAMPERED_HEALTH"}, {"gate", "execution_authorized", true, "INVALID_SCHEMA"},
		{"authorization", "execution_mode", "GOVERNED_CANARY", "INVALID_SCHEMA"}, {"authorization", "execution_authorized", false, "INVALID_SCHEMA"}, {"authorization", "expires_at", "2026-09-03T08:00:00Z", "INVALID_TIME_BINDING"},
		{"execution", "status", "SUCCEEDED", "INVALID_SCHEMA"}, {"execution", "canary_gate_id", "other-stage", "INVALID_SCHEMA"},
		{"ledger", "pending_outcomes", 1, "INVALID_LEDGER"}, {"ledger", "executions_in_window", 1, "INVALID_LEDGER"},
		{"ledger", "reconciliation_resolution_ids", []string{"x", "x"}, "INVALID_SCHEMA"}, {"ledger", "last_outcome_at", nil, "INVALID_SCHEMA"},
		{"cycle", "closed_at", "2026-09-03T07:00:00Z", "INVALID_TIME_BINDING"}, {"cycle", "review_id", nil, "INVALID_SCHEMA"},
		{"resolution", "resolved_by", "agent", "INVALID_SCHEMA"}, {"resolution", "effect_state", "UNKNOWN", "INVALID_SCHEMA"},
		{"activation", "activated_at", "invalid", "INVALID_SCHEMA"},
	} {
		t.Run(tc.kind+"/"+tc.key, func(t *testing.T) {
			var fields map[string]json.RawMessage
			json.Unmarshal(m11Raw(t)[tc.kind], &fields)
			fields[tc.key], _ = json.Marshal(tc.value)
			b, _ := json.Marshal(fields)
			if v, s := DecodeM11Artifact(tc.kind, b); s != tc.want || v != nil {
				t.Fatal(s, tc.want)
			}
		})
	}
}

func TestM11RawHashTimeNumbers(t *testing.T) {
	s, _ := baseM11()
	s.Lease.MaxCostMinorTotal = 9007199254740993
	s.Lease = SealProductionLease(s.Lease)
	b, _ := json.Marshal(s.Lease)
	v, status := DecodeM11Artifact("lease", b)
	if status != missionValid || v.(*ProductionLease).MaxCostMinorTotal != 9007199254740993 {
		t.Fatal(status, v)
	}
	for _, n := range []string{"9223372036854775808", "1.5"} {
		bad := bytes.Replace(b, []byte("9007199254740993"), []byte(n), 1)
		if _, status := DecodeM11Artifact("lease", bad); status != "INVALID_SCHEMA" {
			t.Fatal(status)
		}
	}
	s.Lease.ValidFrom = s.Lease.ExpiresAt
	s.Lease = SealProductionLease(s.Lease)
	b, _ = json.Marshal(s.Lease)
	if _, status := DecodeM11Artifact("lease", b); status != "INVALID_TIME_BINDING" {
		t.Fatal(status)
	}
	// Health can be degraded and schema/hash-valid without being safe to execute.
	s.Health.DependencyState = "DEGRADED"
	s.Health = SealProductionHealth(s.Health)
	b, _ = json.Marshal(s.Health)
	if _, status := DecodeM11Artifact("health", b); status != missionValid {
		t.Fatal(status)
	}
	before := s.Ledger
	b, _ = json.Marshal(s.Ledger)
	if !reflect.DeepEqual(before, s.Ledger) || bytes.Contains(b, []byte(`:null`)) {
		t.Fatal("ledger wire shape/mutation")
	}
	var a ProductionExecutionAuthorization
	json.Unmarshal(m11Raw(t)["authorization"], &a)
	a.AuthorizedAt = "2026-09-03T15:00:00+07:00"
	b, _ = json.Marshal(a)
	if _, status := DecodeM11Artifact("authorization", b); status != missionValid {
		t.Fatal(status)
	}
}

func TestM11AuditCLI(t *testing.T) {
	var fixture bytes.Buffer
	if e := runM11Check(&fixture, []string{"activation", "../../testdata/m11-activation.json"}); e != nil {
		t.Fatal(e)
	}
	var summary M10CheckSummary
	if e := json.Unmarshal(fixture.Bytes(), &summary); e != nil {
		t.Fatal(e)
	}
	if summary.Result != "ARTIFACT_VALID_UNVERIFIED" || summary.ProvenanceAuthenticated || summary.ChainValidated || summary.ExecutionPermitted {
		t.Fatal(summary)
	}
	dir := t.TempDir()
	for kind, raw := range m11Raw(t) {
		p := filepath.Join(dir, kind+".json")
		if e := os.WriteFile(p, raw, 0600); e != nil {
			t.Fatal(e)
		}
		var out bytes.Buffer
		for n := 0; n < 2; n++ {
			out.Reset()
			if e := runM11Check(&out, []string{kind, p}); e != nil {
				t.Fatal(e)
			}
			var s M10CheckSummary
			if e := json.Unmarshal(out.Bytes(), &s); e != nil {
				t.Fatal(e)
			}
			if s.Result != "ARTIFACT_VALID_UNVERIFIED" || s.ArtifactType != kind || s.ChainValidated || s.ProvenanceAuthenticated || s.ExecutionPermitted {
				t.Fatal(s)
			}
		}
		if e := runM11Check(m03BrokenWriter{}, []string{kind, p}); e == nil {
			t.Fatal("writer")
		}
		got, e := os.ReadFile(p)
		if e != nil || !bytes.Equal(got, raw) {
			t.Fatal("input changed")
		}
		for _, args := range [][]string{nil, {kind}, {"unknown", p}, {kind, "missing"}} {
			out.Reset()
			if e := runM11Check(&out, args); e == nil || out.Len() != 0 {
				t.Fatal("args")
			}
		}
		os.WriteFile(p, []byte(`{}`), 0600)
		out.Reset()
		if e := runM11Check(&out, []string{kind, p}); e == nil || out.Len() != 0 {
			t.Fatal("bad input")
		}
	}
	entries, e := os.ReadDir(dir)
	if e != nil || len(entries) != 11 {
		t.Fatal("persisted state")
	}
}
