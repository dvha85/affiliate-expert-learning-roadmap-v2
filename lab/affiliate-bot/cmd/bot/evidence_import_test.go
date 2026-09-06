package main

import (
	"encoding/json"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m00"
	"os"
	"testing"
)

func TestImportedHistoryProjection(t *testing.T) {
	raw, err := os.ReadFile("../../../../examples/m00-import/packet-t1.json")
	if err != nil {
		t.Fatal(err)
	}
	converted, err := m00.Convert(raw)
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(converted)
	var observations []Observation
	if err = json.Unmarshal(encoded, &observations); err != nil {
		t.Fatal(err)
	}
	record, err := NewHistoryRecord("import-test", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z", observations)
	if err != nil || record.RecordedResult.State != stateGetMoreData || Replay(record).State != replayMatch {
		t.Fatal(record, err)
	}
	// Provenance metadata is covered by the existing input hash.
	originalHash := record.InputHash
	observations[0].TransformationOrMethod += " "
	if _, err := NewHistoryRecord("import-test", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z", observations); err == nil {
		t.Fatal("noncanonical provenance accepted")
	}
	observations = record.Observations
	v := 0.0
	observations[0].CommissionRate = &v
	if _, err := NewHistoryRecord("import-test", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z", observations); err == nil {
		t.Fatal("pending converted to zero")
	}
	if originalHash == "" {
		t.Fatal("missing hash")
	}
}
