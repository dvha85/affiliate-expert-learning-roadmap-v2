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

func TestM11PersistenceRaw(t *testing.T) {
	s, c := baseM11()
	for kind, value := range map[string]any{"ledger": s.Ledger, "activation": productionActivationRecord{s.Lease.LeaseID, s.Lease.LeaseVersion, s.Lease.LeaseHash, c.Now}} {
		raw, e := encodeProductionArtifact(kind, value)
		if e != nil {
			t.Fatal(e)
		}
		check := func(b []byte) {
			t.Helper()
			p := filepath.Join(t.TempDir(), "record.json")
			os.WriteFile(p, b, 0600)
			if kind == "ledger" {
				l, e := loadProductionLedger(p, s.Lease)
				if e == nil || !reflect.DeepEqual(l, ProductionLedger{}) {
					t.Fatal("invalid ledger accepted")
				}
			} else if productionActivationMatches(p, s.Lease) {
				t.Fatal("invalid activation accepted")
			}
			got, _ := os.ReadFile(p)
			if !bytes.Equal(got, b) {
				t.Fatal("rewrote file")
			}
		}
		var fields map[string]json.RawMessage
		json.Unmarshal(raw, &fields)
		for key := range fields {
			for _, mutation := range []string{"missing", "null", "type"} {
				t.Run(kind+"/"+key+"/"+mutation, func(t *testing.T) {
					m := map[string]json.RawMessage{}
					for k, v := range fields {
						m[k] = v
					}
					switch mutation {
					case "missing":
						delete(m, key)
					case "null":
						m[key] = json.RawMessage(`null`)
					default:
						m[key] = json.RawMessage(`{}`)
					}
					b, _ := json.Marshal(m)
					check(b)
				})
			}
		}
		for _, bad := range []string{`null`, string(raw) + ` {}`, strings.Replace(string(raw), `{`, `{"extra":true,`, 1), strings.Replace(string(raw), `{`, `{"lease_id":"duplicate",`, 1)} {
			check([]byte(bad))
		}
		for _, key := range []string{"lease_id", "lease_version", "lease_hash"} {
			m := map[string]json.RawMessage{}
			for k, v := range fields {
				m[k] = v
			}
			v := "wrong"
			if key == "lease_hash" {
				v = "sha256:" + strings.Repeat("0", 64)
			}
			m[key], _ = json.Marshal(v)
			b, _ := json.Marshal(m)
			check(b)
		}
	}
}

func TestM11PersistenceWriters(t *testing.T) {
	s, c := baseM11()
	dir := t.TempDir()
	p := filepath.Join(dir, "ledger.json")
	if e := persistProductionLedger(p, s.Ledger); e != nil {
		t.Fatal(e)
	}
	before, _ := os.ReadFile(p)
	bad := s.Ledger
	bad.PendingOutcomes = 1
	if e := persistProductionLedger(p, bad); e == nil {
		t.Fatal("bad ledger written")
	}
	after, _ := os.ReadFile(p)
	if !bytes.Equal(before, after) {
		t.Fatal("overwritten")
	}
	if e := ensureInitialProductionLedger(filepath.Join(dir, "bad.json"), bad); e == nil {
		t.Fatal("bad init")
	}
	if e := writeActivationRecord(filepath.Join(dir, "bad-activation.json"), productionActivationRecord{LeaseID: s.Lease.LeaseID}); e == nil {
		t.Fatal("bad activation")
	}
	s.Ledger = bad
	newDir := filepath.Join(dir, "must-not-exist")
	if got := InitializeProductionLedger(&s, newDir, c.Now); got != "DENY_ACTIVATION" {
		t.Fatal(got)
	}
	if _, e := os.Stat(newDir); !os.IsNotExist(e) {
		t.Fatal("created directory")
	}
	s, _ = baseM11()
	s.Ledger.CostMinorTotal = 9007199254740993
	if e := persistProductionLedger(p, s.Ledger); e != nil {
		t.Fatal(e)
	}
	loaded, e := loadProductionLedger(p, s.Lease)
	if e != nil || loaded.CostMinorTotal != 9007199254740993 {
		t.Fatal(loaded, e)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatal("leaked files")
	}
}

func TestM11PersistenceRuntimeRejectsCorruption(t *testing.T) {
	for _, mode := range []string{"missing_ledger", "corrupt_ledger", "wrong_lease", "missing_activation_time"} {
		t.Run(mode, func(t *testing.T) {
			s, c := baseM11()
			dir := t.TempDir()
			initM11(t, &s, dir, c.Now)
			a, _, status := AuthorizeProduction(s, c)
			if status != "AUTHORIZED" {
				t.Fatal(status)
			}
			s.Authorization = &a
			p := productionLedgerPath(dir, s.Lease.LeaseID, s.Lease.LeaseVersion)
			ap := productionActivationPath(dir, s.Lease.LeaseID, s.Lease.LeaseVersion)
			switch mode {
			case "missing_ledger":
				os.Remove(p)
			case "corrupt_ledger":
				os.WriteFile(p, []byte(`{}`), 0600)
			case "wrong_lease":
				l := s.Ledger
				l.LeaseVersion = "other"
				b, _ := json.Marshal(l)
				os.WriteFile(p, b, 0600)
			case "missing_activation_time":
				b, _ := json.Marshal(map[string]string{"lease_id": s.Lease.LeaseID, "lease_version": s.Lease.LeaseVersion, "lease_hash": s.Lease.LeaseHash})
				os.WriteFile(ap, b, 0600)
			}
			before, _ := os.ReadFile(p)
			r, status := ExecuteProductionLocalSandbox(&s, a, c, dir)
			if r != (ProductionExecutionRecord{}) || !strings.HasPrefix(status, "STOP_") {
				t.Fatal(status, r)
			}
			_, status = EnforceProductionGate(&s, c, dir)
			if !strings.HasPrefix(status, "STOP_") {
				t.Fatal(status)
			}
			if _, e := os.Stat(sandboxIdempotencyPath(dir, a.IdempotencyKey)); !os.IsNotExist(e) {
				t.Fatal("side effect")
			}
			after, e := os.ReadFile(p)
			if mode == "missing_ledger" {
				if !os.IsNotExist(e) {
					t.Fatal("recreated ledger")
				}
			} else if !bytes.Equal(before, after) {
				t.Fatal("changed corrupt ledger")
			}
			if mode != "missing_activation_time" {
				if got := RecordProductionOutcome(&s, dir, "out", "exec-x", c.Now); got == "OUTCOME_RECORDED" {
					t.Fatal(got)
				}
			}
		})
	}
}

func TestM11PersistenceStoppedRestart(t *testing.T) {
	s, c := baseM11()
	dir := t.TempDir()
	initM11(t, &s, dir, c.Now)
	a, _, status := AuthorizeProduction(s, c)
	if status != "AUTHORIZED" {
		t.Fatal(status)
	}
	s.Authorization = &a
	r, status := ExecuteProductionLocalSandbox(&s, a, c, dir)
	if status != "EXECUTED" {
		t.Fatal(status)
	}
	s.Health.ComplianceAlertCount = 1
	refreshHealthTrust(&s, &c)
	if _, status := EnforceProductionGate(&s, c, dir); status != "STOP" {
		t.Fatal(status)
	}
	p := productionLedgerPath(dir, s.Lease.LeaseID, s.Lease.LeaseVersion)
	loaded, e := loadProductionLedger(p, s.Lease)
	if e != nil || loaded.ControlMode != "STOPPED" || loaded.CostMinorTotal != 100 || loaded.PendingOutcomes != 1 {
		t.Fatal(loaded, e)
	}
	fresh, fc := baseM11()
	fresh.Authorization = &a
	if _, status := EnforceProductionGate(&fresh, fc, dir); status != "STOP" {
		t.Fatal("resumed", status)
	}
	if got := RecordProductionOutcome(&fresh, dir, "out", r.ExecutionID, "2026-09-03T08:10:00Z"); got != "OUTCOME_RECORDED" {
		t.Fatal(got)
	}
	loaded, e = loadProductionLedger(p, s.Lease)
	if e != nil || loaded.ControlMode != "STOPPED" || loaded.PendingOutcomes != 0 || loaded.ExecutionsTotal != 1 || len(loaded.SuccessfulIdempotencyKeys) != 1 || len(loaded.OutcomeLinks) != 1 {
		t.Fatal(loaded, e)
	}
}
