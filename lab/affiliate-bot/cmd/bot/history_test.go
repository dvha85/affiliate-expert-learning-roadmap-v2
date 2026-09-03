package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func historyObservation(id, product, name string, price, rate float64, at string) Observation {
	return Observation{ObservationID: id, ProductID: product, ProductName: name, Price: f(price), CommissionRate: f(rate), Currency: "USD", EvidenceKind: "synthetic", ObservedAt: at}
}

func TestHistoryRecordDeepCopiesInput(t *testing.T) {
	price := 100.0
	observations := []Observation{{ObservationID: "o1", ProductID: "p", ProductName: "P", Price: &price, CommissionRate: f(0.1), Currency: "USD", EvidenceKind: "synthetic", ObservedAt: "2026-09-01T00:00:00Z"}}
	record, err := NewHistoryRecord("r1", "2026-09-01T01:00:00Z", "2026-09-01T00:01:00Z", observations)
	if err != nil { t.Fatal(err) }
	price = 999
	observations[0].ProductName = "mutated"
	if *record.Observations[0].Price != 100 || record.Observations[0].ProductName != "P" {
		t.Fatalf("history snapshot changed after caller mutation: %#v", record.Observations[0])
	}
}

func TestHistoryAppendDuplicateConflictRestartAndOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	t2, err := NewHistoryRecord("r2", "2026-09-02T01:00:00Z", "2026-09-03T00:00:00Z", []Observation{historyObservation("o2", "p", "P", 90, 0.1, "2026-09-02T00:00:00Z")})
	if err != nil { t.Fatal(err) }
	t1, err := NewHistoryRecord("r1", "2026-09-01T01:00:00Z", "2026-09-04T00:00:00Z", []Observation{historyObservation("o1", "p", "P", 100, 0.1, "2026-09-01T00:00:00Z")})
	if err != nil { t.Fatal(err) }
	if got, err := AppendHistory(path, t2); err != nil || got != appendAdded { t.Fatalf("append t2: %s %v", got, err) }
	if got, err := AppendHistory(path, t1); err != nil || got != appendAdded { t.Fatalf("append t1: %s %v", got, err) }
	if got, err := AppendHistory(path, t1); err != nil || got != appendDuplicate { t.Fatalf("duplicate: %s %v", got, err) }
	records, err := LoadHistory(path)
	if err != nil { t.Fatal(err) }
	if len(records) != 2 || records[0].RecordID != "r1" || records[1].RecordID != "r2" { t.Fatalf("expected as_of order r1,r2: %#v", records) }
	conflict := t1
	conflict.AsOf = "2026-09-05T00:00:00Z"
	if _, err := AppendHistory(path, conflict); err == nil || !strings.Contains(err.Error(), "CONFLICT") { t.Fatalf("expected conflict, got %v", err) }
	recordsAfterRestart, err := LoadHistory(path)
	if err != nil { t.Fatal(err) }
	if !reflect.DeepEqual(records, recordsAfterRestart) { t.Fatal("restart/read must preserve history") }
}

func TestReplayMatchDriftAndUnreplayable(t *testing.T) {
	record, err := NewHistoryRecord("r1", "2026-09-01T01:00:00Z", "2026-09-01T00:01:00Z", []Observation{historyObservation("o1", "p", "P", 100, 0.1, "2026-09-01T00:00:00Z")})
	if err != nil { t.Fatal(err) }
	if got := Replay(record); got.State != replayMatch { t.Fatalf("expected MATCH: %#v", got) }
	drift := record
	drift.RecordedResult.State = stateHumanReview
	if got := Replay(drift); got.State != replayDrift { t.Fatalf("expected DRIFT: %#v", got) }
	old := record
	old.FormulaVersion = "commission-per-order/v0"
	old.RecordedResult.FormulaVersion = old.FormulaVersion
	if got := Replay(old); got.State != replayUnreplayable { t.Fatalf("expected UNREPLAYABLE: %#v", got) }
}

func TestHistoryHashAndCorruptionFailClosed(t *testing.T) {
	record, err := NewHistoryRecord("r1", "2026-09-01T01:00:00Z", "2026-09-01T00:01:00Z", []Observation{historyObservation("o1", "p", "P", 100, 0.1, "2026-09-01T00:00:00Z")})
	if err != nil { t.Fatal(err) }
	record.Observations[0].ProductName = "tampered"
	if err := validateHistoryRecord(record); err == nil || !strings.Contains(err.Error(), "input_hash mismatch") { t.Fatalf("expected hash failure, got %v", err) }
	path := filepath.Join(t.TempDir(), "bad.jsonl")
	if err := os.WriteFile(path, []byte("{not-json\n"), 0o600); err != nil { t.Fatal(err) }
	if _, err := LoadHistory(path); err == nil || !strings.Contains(err.Error(), "corrupt") { t.Fatalf("expected corrupt history error, got %v", err) }
}

type historyRecordSpec struct {
	RecordID string `json:"record_id"`
	AsOf string `json:"as_of"`
	IngestedAt string `json:"ingested_at"`
	Observations []Observation `json:"observations"`
	FormulaOverride string `json:"formula_version_override"`
	DecisionStateOverride string `json:"decision_state_override"`
	TamperProductName string `json:"tamper_product_name"`
}

type m02EvalCase struct {
	CaseID string `json:"case_id"`
	Mode string `json:"mode"`
	Records []historyRecordSpec `json:"records"`
	Expected string `json:"expected"`
	ExpectedOrder []string `json:"expected_order"`
	RawLine string `json:"raw_line"`
}

func recordFromSpec(spec historyRecordSpec) (HistoryRecord, error) {
	record, err := NewHistoryRecord(spec.RecordID, spec.AsOf, spec.IngestedAt, spec.Observations)
	if err != nil { return HistoryRecord{}, err }
	if spec.FormulaOverride != "" {
		record.FormulaVersion = spec.FormulaOverride
		record.RecordedResult.FormulaVersion = spec.FormulaOverride
	}
	if spec.DecisionStateOverride != "" { record.RecordedResult.State = spec.DecisionStateOverride }
	if spec.TamperProductName != "" && len(record.Observations) > 0 { record.Observations[0].ProductName = spec.TamperProductName }
	return record, nil
}

func TestM02EvalPack(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok { t.Fatal("cannot resolve test file path") }
	path := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "..", "evals", "M02-history-replay", "cases.json"))
	raw, err := os.ReadFile(path)
	if err != nil { t.Fatalf("read M02 eval pack: %v", err) }
	var cases []m02EvalCase
	if err := json.Unmarshal(raw, &cases); err != nil { t.Fatalf("decode M02 eval pack: %v", err) }
	if len(cases) < 12 { t.Fatalf("expected at least 12 M02 eval cases, got %d", len(cases)) }

	for _, tc := range cases {
		t.Run(tc.CaseID, func(t *testing.T) {
			switch tc.Mode {
			case "append_order":
				historyPath := filepath.Join(t.TempDir(), "history.jsonl")
				for _, spec := range tc.Records {
					record, err := recordFromSpec(spec); if err != nil { t.Fatal(err) }
					if _, err := AppendHistory(historyPath, record); err != nil { t.Fatal(err) }
				}
				records, err := LoadHistory(historyPath); if err != nil { t.Fatal(err) }
				ids := make([]string, 0, len(records)); for _, record := range records { ids = append(ids, record.RecordID) }
				if !reflect.DeepEqual(ids, tc.ExpectedOrder) { t.Fatalf("expected order %v, got %v", tc.ExpectedOrder, ids) }
			case "duplicate":
				historyPath := filepath.Join(t.TempDir(), "history.jsonl")
				record, err := recordFromSpec(tc.Records[0]); if err != nil { t.Fatal(err) }
				if _, err := AppendHistory(historyPath, record); err != nil { t.Fatal(err) }
				status, err := AppendHistory(historyPath, record)
				if err != nil || status != tc.Expected { t.Fatalf("expected %s, got %s, err=%v", tc.Expected, status, err) }
			case "conflict", "observation_conflict":
				historyPath := filepath.Join(t.TempDir(), "history.jsonl")
				first, err := recordFromSpec(tc.Records[0]); if err != nil { t.Fatal(err) }
				second, err := recordFromSpec(tc.Records[1]); if err != nil { t.Fatal(err) }
				if _, err := AppendHistory(historyPath, first); err != nil { t.Fatal(err) }
				if _, err := AppendHistory(historyPath, second); err == nil || !strings.Contains(err.Error(), tc.Expected) { t.Fatalf("expected conflict containing %q, got %v", tc.Expected, err) }
			case "replay":
				record, err := recordFromSpec(tc.Records[0]); if err != nil { t.Fatal(err) }
				if got := Replay(record); got.State != tc.Expected { t.Fatalf("expected %s, got %#v", tc.Expected, got) }
			case "hash_tamper":
				record, err := recordFromSpec(tc.Records[0]); if err != nil { t.Fatal(err) }
				if err := validateHistoryRecord(record); err == nil || !strings.Contains(err.Error(), tc.Expected) { t.Fatalf("expected validation error containing %q, got %v", tc.Expected, err) }
			case "build_error":
				if _, err := recordFromSpec(tc.Records[0]); err == nil || !strings.Contains(err.Error(), tc.Expected) { t.Fatalf("expected build error containing %q, got %v", tc.Expected, err) }
			case "corrupt":
				historyPath := filepath.Join(t.TempDir(), "history.jsonl")
				if err := os.WriteFile(historyPath, []byte(tc.RawLine+"\n"), 0o600); err != nil { t.Fatal(err) }
				if _, err := LoadHistory(historyPath); err == nil || !strings.Contains(err.Error(), tc.Expected) { t.Fatalf("expected corrupt error containing %q, got %v", tc.Expected, err) }
			case "hash_order_invariant":
				first, err := recordFromSpec(tc.Records[0]); if err != nil { t.Fatal(err) }
				second, err := recordFromSpec(tc.Records[1]); if err != nil { t.Fatal(err) }
				if first.InputHash != second.InputHash { t.Fatalf("canonical input hash changed with input order: %s != %s", first.InputHash, second.InputHash) }
			default:
				t.Fatalf("unknown M02 eval mode %q", tc.Mode)
			}
		})
	}
}
