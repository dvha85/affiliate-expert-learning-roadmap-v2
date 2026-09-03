package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func sampleHistoryRecord(t *testing.T, id, asOf string) HistoryRecord {
	t.Helper()
	record, err := NewHistoryRecord(id, asOf, "2026-09-03T10:00:00Z", []Observation{{ProductID: "a", ProductName: "A", Price: f(100), CommissionRate: f(0.10), Currency: "USD", EvidenceKind: "synthetic"}})
	if err != nil { t.Fatal(err) }
	return record
}

func TestHistoryAppendDuplicateConflictAndRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	record := sampleHistoryRecord(t, "r1", "2026-09-01T10:00:00Z")
	if state, err := AppendHistory(path, record); err != nil || state != "APPENDED" { t.Fatalf("append: %s %v", state, err) }
	if state, err := AppendHistory(path, record); err != nil || state != "EXACT_DUPLICATE" { t.Fatalf("duplicate: %s %v", state, err) }
	conflict := record
	conflict.IngestedAt = "2026-09-04T10:00:00Z"
	if state, err := AppendHistory(path, conflict); err == nil || state != "CONFLICT" { t.Fatalf("conflict should reject: %s %v", state, err) }
	loaded, err := LoadHistory(path)
	if err != nil || len(loaded) != 1 { t.Fatalf("restart load: %d %v", len(loaded), err) }
}

func TestHistoryQueryUsesAsOfNotAppendOrder(t *testing.T) {
	late := sampleHistoryRecord(t, "r2", "2026-09-02T10:00:00Z")
	early := sampleHistoryRecord(t, "r1", "2026-09-01T10:00:00Z")
	ordered, err := QueryHistory([]HistoryRecord{late, early})
	if err != nil { t.Fatal(err) }
	if ordered[0].RecordID != "r1" || ordered[1].RecordID != "r2" { t.Fatalf("unexpected order: %#v", ordered) }
}

func TestReplayStates(t *testing.T) {
	record := sampleHistoryRecord(t, "r1", "2026-09-01T10:00:00Z")
	if got := Replay(record); got.State != replayMatch { t.Fatalf("expected MATCH, got %#v", got) }
	drift := record
	drift.RecordedResult.State = stateHumanReview
	if got := Replay(drift); got.State != replayDrift { t.Fatalf("expected DRIFT, got %#v", got) }
	unknown := record
	unknown.FormulaVersion = "unknown/v9"
	if got := Replay(unknown); got.State != replayUnreplayable { t.Fatalf("expected UNREPLAYABLE, got %#v", got) }
	tampered := record
	tampered.InputHash = "bad"
	if got := Replay(tampered); got.State != "INTEGRITY_ERROR" { t.Fatalf("expected INTEGRITY_ERROR, got %#v", got) }
}

func TestCorruptHistoryFailsClosed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	if err := os.WriteFile(path, []byte("{not-json}\n"), 0o644); err != nil { t.Fatal(err) }
	if _, err := LoadHistory(path); err == nil { t.Fatal("corrupt history must fail closed") }
}

type m02EvalCase struct {
	CaseID   string `json:"case_id"`
	Kind     string `json:"kind"`
	Expected string `json:"expected"`
}

func TestM02EvalPackCoverage(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok { t.Fatal("cannot resolve test path") }
	path := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "..", "evals", "M02-history-replay", "cases.json"))
	raw, err := os.ReadFile(path)
	if err != nil { t.Fatal(err) }
	var cases []m02EvalCase
	if err := json.Unmarshal(raw, &cases); err != nil { t.Fatal(err) }
	if len(cases) < 8 { t.Fatalf("expected at least 8 cases, got %d", len(cases)) }
	seen := map[string]bool{}
	for _, tc := range cases {
		if tc.CaseID == "" || tc.Kind == "" || tc.Expected == "" { t.Fatalf("incomplete eval case: %#v", tc) }
		if seen[tc.CaseID] { t.Fatalf("duplicate case id %s", tc.CaseID) }
		seen[tc.CaseID] = true
	}
	if len(seen) != len(cases) { t.Fatal("eval case coverage mismatch") }
}
