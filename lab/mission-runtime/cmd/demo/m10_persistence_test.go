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

func TestM10PersistenceRawValidation(t *testing.T) {
	s, _ := baseM10()
	raw, e := encodeCanaryLedger(s.Ledger)
	if e != nil {
		t.Fatal(e)
	}
	var fields map[string]json.RawMessage
	json.Unmarshal(raw, &fields)
	for key := range fields {
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
				bad, _ := json.Marshal(copy)
				p := filepath.Join(t.TempDir(), "ledger.json")
				os.WriteFile(p, bad, 0600)
				l, e := loadExistingCanaryLedger(p, s.Grant)
				if e == nil || !reflect.DeepEqual(l, CanaryLedger{}) {
					t.Fatal("invalid input returned ledger", e)
				}
				got, _ := os.ReadFile(p)
				if !bytes.Equal(got, bad) {
					t.Fatal("rewrote malformed file")
				}
			})
		}
	}
	for _, bad := range []string{`null`, string(raw) + ` {}`, strings.Replace(string(raw), `{`, `{"extra":true,`, 1), strings.Replace(string(raw), `{`, `{"grant_id":"duplicate",`, 1)} {
		if _, e := decodeCanaryLedger([]byte(bad)); e == nil {
			t.Fatal("ambiguous JSON")
		}
	}
	for _, field := range []string{"grant_id", "grant_version", "grant_hash"} {
		copy := map[string]json.RawMessage{}
		for k, v := range fields {
			copy[k] = v
		}
		value := "wrong"
		if field == "grant_hash" {
			value = "sha256:" + strings.Repeat("0", 64)
		}
		copy[field], _ = json.Marshal(value)
		b, _ := json.Marshal(copy)
		p := filepath.Join(t.TempDir(), "ledger.json")
		os.WriteFile(p, b, 0600)
		if l, e := loadExistingCanaryLedger(p, s.Grant); e == nil || !reflect.DeepEqual(l, CanaryLedger{}) {
			t.Fatal(field, e)
		}
	}
}

func TestM10PersistenceWriterFailsBeforeIO(t *testing.T) {
	s, _ := baseM10()
	dir := t.TempDir()
	p := canaryLedgerPath(dir, s.Grant.GrantID, s.Grant.GrantVersion)
	if e := persistCanaryLedger(p, s.Ledger); e != nil {
		t.Fatal(e)
	}
	before, _ := os.ReadFile(p)
	bad := s.Ledger
	bad.PendingOutcomes = 1
	if e := persistCanaryLedger(p, bad); e == nil {
		t.Fatal("invalid write accepted")
	}
	after, _ := os.ReadFile(p)
	if !bytes.Equal(before, after) {
		t.Fatal("overwrote valid file")
	}
	newPath := filepath.Join(dir, "invalid.json")
	if e := ensureInitialCanaryLedger(newPath, bad); e == nil {
		t.Fatal("invalid initialization")
	}
	if _, e := os.Stat(canaryInitializationPath(newPath)); !os.IsNotExist(e) {
		t.Fatal("invalid state left marker")
	}
	// int64 is decoded directly, not via float64.
	s.Ledger.CostMinorTotal = 9007199254740993
	if e := persistCanaryLedger(p, s.Ledger); e != nil {
		t.Fatal(e)
	}
	l, e := loadExistingCanaryLedger(p, s.Grant)
	if e != nil || l.CostMinorTotal != 9007199254740993 {
		t.Fatal(l, e)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatal("unexpected temp files")
	}
}

func TestM10PersistenceMissingAndCorruptRestart(t *testing.T) {
	for _, mode := range []string{"missing", "corrupt", "wrong_grant"} {
		t.Run(mode, func(t *testing.T) {
			s, c := baseM10()
			a, _, status := AuthorizeCanary(s, c)
			if status != "AUTHORIZED" {
				t.Fatal(status)
			}
			s.Authorization = &a
			dir := t.TempDir()
			if _, status := ExecuteCanaryLocalSandbox(&s, a, c, dir); status != "EXECUTED" {
				t.Fatal(status)
			}
			p := canaryLedgerPath(dir, s.Grant.GrantID, s.Grant.GrantVersion)
			switch mode {
			case "missing":
				if e := os.Remove(p); e != nil {
					t.Fatal(e)
				}
			case "corrupt":
				os.WriteFile(p, []byte(`{}`), 0600)
			case "wrong_grant":
				l := s.Ledger
				l.GrantID = "other"
				b, _ := json.Marshal(l)
				os.WriteFile(p, b, 0600)
			}
			before, _ := os.ReadFile(p)
			// Stale empty in-memory snapshot must not recreate a lost managed ledger.
			fresh, fc := baseM10()
			refreshM10Intent(&fresh, &fc, "PREPARE_LOCAL_DRAFT", "https://example.com/draft", "i-next", "k-next", 100)
			fa, _, status := AuthorizeCanary(fresh, fc)
			if status != "AUTHORIZED" {
				t.Fatal(status)
			}
			fresh.Authorization = &fa
			r, status := ExecuteCanaryLocalSandbox(&fresh, fa, fc, dir)
			want := "WAIT_RECONCILIATION"
			if mode == "missing" {
				want = "WAIT_LEDGER_MISSING"
			}
			if status != want || r != (CanaryExecutionRecord{}) {
				t.Fatal(status, want)
			}
			after, e := os.ReadFile(p)
			if mode == "missing" {
				if !os.IsNotExist(e) {
					t.Fatal("recreated ledger")
				}
			} else if !bytes.Equal(before, after) {
				t.Fatal("rewrote bad ledger")
			}
			if _, e := os.Stat(sandboxIdempotencyPath(dir, fa.IdempotencyKey)); !os.IsNotExist(e) {
				t.Fatal("new side effect")
			}
			if got := RecordCanaryOutcome(&fresh, dir, "out", s.Execution.ExecutionID, "2026-09-03T08:10:00Z"); got != "WAIT_RECONCILIATION" {
				t.Fatal(got)
			}
		})
	}
}

func TestM10PersistenceHistoryRoundTrip(t *testing.T) {
	s, c := baseM10()
	a, _, status := AuthorizeCanary(s, c)
	if status != "AUTHORIZED" {
		t.Fatal(status)
	}
	s.Authorization = &a
	dir := t.TempDir()
	r, status := ExecuteCanaryLocalSandbox(&s, a, c, dir)
	if status != "EXECUTED" {
		t.Fatal(status)
	}
	p := canaryLedgerPath(dir, s.Grant.GrantID, s.Grant.GrantVersion)
	loaded, e := loadExistingCanaryLedger(p, s.Grant)
	if e != nil || loaded.ExecutionsTotal != 1 || loaded.CostMinorTotal != 100 || loaded.PendingOutcomes != 1 || !reflect.DeepEqual(loaded.SuccessfulIdempotencyKeys, s.Ledger.SuccessfulIdempotencyKeys) {
		t.Fatal(loaded, e)
	}
	fresh, fc := baseM10()
	fresh.Authorization = &a
	if _, status := ExecuteCanaryLocalSandbox(&fresh, a, fc, dir); status != "WAIT" {
		t.Fatal("replay", status)
	}
	if status := RecordCanaryOutcome(&fresh, dir, "out", r.ExecutionID, "2026-09-03T08:10:00Z"); status != "OUTCOME_RECORDED" {
		t.Fatal(status)
	}
	loaded, e = loadExistingCanaryLedger(p, s.Grant)
	if e != nil || loaded.PendingOutcomes != 0 || len(loaded.OutcomeLinks) != 1 || loaded.CostMinorTotal != 100 || loaded.ExecutionsTotal != 1 || len(loaded.SuccessfulIdempotencyKeys) != 1 {
		t.Fatal(loaded, e)
	}
	before, _ := os.ReadFile(p)
	if e := ensureInitialCanaryLedger(p, baseLedgerM10()); e == nil {
		t.Fatal("reinitialized existing ledger")
	}
	after, _ := os.ReadFile(p)
	if !bytes.Equal(before, after) {
		t.Fatal("reset budget")
	}
}

func baseLedgerM10() CanaryLedger { s, _ := baseM10(); return s.Ledger }
