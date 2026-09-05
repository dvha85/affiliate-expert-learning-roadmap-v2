package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

func adapterContext() DecisionContext {
	return DecisionContext{Question: "What evidence is missing?", SupportedFacts: []string{}, Assumptions: []string{"Synthetic ranking only"}, Unknowns: []string{"Real conversion"}, NextMeasurement: "Collect permitted evidence"}
}

func adapterRecord(t *testing.T) HistoryRecord {
	t.Helper()
	o := historyObservation("obs-1", "offer-1", "Fixture", 100, 0.1, "2026-09-01T00:00:00Z")
	r, err := NewHistoryRecord("d-1", "2026-09-01T01:00:00Z", "2026-09-01T01:01:00Z", []Observation{o})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestDecisionPacketAdapterPreservesProjectionAndLineage(t *testing.T) {
	r := adapterRecord(t)
	before, _ := json.Marshal(r)
	context := adapterContext()
	packet, err := HistoryDecisionPacket(r, context)
	if err != nil {
		t.Fatal(err)
	}
	if packet.DecisionID != r.RecordID || !reflect.DeepEqual(packet.EvidenceIDs, r.RecordedResult.EvidenceIDs) || packet.State != r.RecordedResult.State || packet.Reason != strings.Join(r.RecordedResult.Reasons, "\n") || packet.Action != nil {
		t.Fatal("mapping changed source decision")
	}
	raw, err := json.Marshal(packet)
	if err != nil {
		t.Fatal(err)
	}
	if err := contracts.ValidateRaw("decision-packet.schema.json", raw); err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte(`"ranked"`)) || bytes.Contains(raw, []byte(`"formula_version"`)) {
		t.Fatal("projection fields leaked into packet")
	}
	packet.EvidenceIDs[0] = "mutated"
	packet.Assumptions[0] = "mutated"
	after, _ := json.Marshal(r)
	if !bytes.Equal(before, after) || context.Assumptions[0] == "mutated" {
		t.Fatal("adapter aliases source arrays")
	}
	if Replay(r).State != replayMatch {
		t.Fatal("adapter changed replay")
	}
}

func TestDecisionPacketRefusesInvalidContextAndHistory(t *testing.T) {
	for _, mode := range []string{"question", "nil_array", "hash", "evidence_id", "drift", "formula"} {
		t.Run(mode, func(t *testing.T) {
			r := adapterRecord(t)
			c := adapterContext()
			switch mode {
			case "question":
				c.Question = " "
			case "nil_array":
				c.Unknowns = nil
			case "hash":
				r.InputHash = strings.Repeat("0", 64)
			case "evidence_id":
				r.RecordedResult.EvidenceIDs = []string{"ghost"}
			case "drift":
				r.RecordedResult.State = stateHumanReview
			case "formula":
				r.FormulaVersion = "unsupported"
				r.RecordedResult.FormulaVersion = "unsupported"
			}
			if _, err := HistoryDecisionPacket(r, c); err == nil {
				t.Fatal("invalid export accepted")
			}
		})
	}
}

func TestHistorySchemaAndLegacyNullRanking(t *testing.T) {
	r := adapterRecord(t)
	r.Observations[0].CommissionRate = nil
	current, err := NewHistoryRecord(r.RecordID, r.AsOf, r.IngestedAt, r.Observations)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(current)
	if err := contracts.ValidateRaw("history-record.schema.json", raw); err != nil {
		t.Fatal(err)
	}
	packet, err := HistoryDecisionPacket(current, adapterContext())
	if err != nil {
		t.Fatal(err)
	}
	if packet.State != stateGetMoreData || len(packet.MissingEvidence) == 0 {
		t.Fatal("missing commission became zero/success")
	}
	legacy := bytes.Replace(raw, []byte(`"ranked":[]`), []byte(`"ranked":null`), 1)
	if bytes.Equal(raw, legacy) {
		t.Fatal("test did not mutate legacy representation")
	}
	path := filepath.Join(t.TempDir(), "legacy.jsonl")
	if err := os.WriteFile(path, append(legacy, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadHistory(path)
	if err != nil {
		t.Fatal(err)
	}
	if Replay(loaded[0]).State != replayMatch {
		t.Fatal("legacy replay lost")
	}
	if _, err := HistoryDecisionPacket(loaded[0], adapterContext()); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, append(legacy, '\n')) {
		t.Fatal("legacy history overwritten")
	}
	if loaded[0].InputHash != current.InputHash {
		t.Fatal("input hash changed")
	}
}

func TestHistoryRawSchemaMutations(t *testing.T) {
	r := adapterRecord(t)
	raw, _ := json.Marshal(r)
	for _, mode := range []string{"missing_ranked", "missing_price", "unknown", "wrong_case", "nested_unknown", "null_observations", "duplicate_ids", "bad_timestamp", "negative_price", "bad_state"} {
		t.Run(mode, func(t *testing.T) {
			var value map[string]any
			if err := json.Unmarshal(raw, &value); err != nil {
				t.Fatal(err)
			}
			result := value["recorded_result"].(map[string]any)
			obs := value["observations"].([]any)[0].(map[string]any)
			switch mode {
			case "missing_ranked":
				delete(result, "ranked")
			case "missing_price":
				delete(obs, "price")
			case "unknown":
				value["extra"] = true
			case "wrong_case":
				obs["Price"] = 999
			case "nested_unknown":
				result["confidence"] = 1
			case "null_observations":
				value["observations"] = nil
			case "duplicate_ids":
				result["evidence_ids"] = []string{"obs-1", "obs-1"}
			case "bad_timestamp":
				obs["observed_at"] = "yesterday"
			case "negative_price":
				obs["price"] = -1
			case "bad_state":
				result["state"] = "RECOMMEND"
			}
			mutant, _ := json.Marshal(value)
			if err := validateHistoryJSON(mutant); err == nil {
				t.Fatal("invalid history accepted")
			}
		})
	}
	duplicate := bytes.Replace(raw, []byte(`"record_id":`), []byte(`"record_id":"other","record_id":`), 1)
	if validateHistoryJSON(duplicate) == nil {
		t.Fatal("duplicate key accepted")
	}
}

func TestDecisionExportCLIHandler(t *testing.T) {
	dir := t.TempDir()
	history := filepath.Join(dir, "history.jsonl")
	contextPath := filepath.Join(dir, "context.json")
	r := adapterRecord(t)
	if _, err := AppendHistory(history, r); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(adapterContext())
	if err := os.WriteFile(contextPath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(history)
	var output bytes.Buffer
	if err := exportHistoryDecision(&output, []string{history, r.RecordID, contextPath}); err != nil {
		t.Fatal(err)
	}
	if err := contracts.ValidateRaw("decision-packet.schema.json", output.Bytes()); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(history)
	if !bytes.Equal(before, after) {
		t.Fatal("read-only export wrote history")
	}
	output.Reset()
	if err := exportHistoryDecision(&output, []string{history, "missing", contextPath}); err == nil || output.Len() != 0 {
		t.Fatal("missing record produced packet")
	}
}

func TestDecisionContextRawBoundary(t *testing.T) {
	raw, _ := json.Marshal(adapterContext())
	for _, mode := range []string{"missing", "null", "null_item", "unknown", "duplicate", "trailing", "wrong_type"} {
		t.Run(mode, func(t *testing.T) {
			var value map[string]any
			_ = json.Unmarshal(raw, &value)
			switch mode {
			case "missing":
				delete(value, "unknowns")
			case "null":
				value["unknowns"] = nil
			case "null_item":
				value["unknowns"] = []any{nil}
			case "unknown":
				value["action"] = nil
			case "wrong_type":
				value["question"] = 3
			}
			mutant, _ := json.Marshal(value)
			if mode == "duplicate" {
				mutant = bytes.Replace(mutant, []byte(`"question":`), []byte(`"question":"other","question":`), 1)
			}
			if mode == "trailing" {
				mutant = append(mutant, []byte(`{}`)...)
			}
			if _, err := DecodeDecisionContext(mutant); err == nil {
				t.Fatal("invalid context accepted")
			}
		})
	}
}
