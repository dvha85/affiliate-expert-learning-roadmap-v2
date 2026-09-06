package main

import (
	"encoding/json"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestAdvisorConfigStrictAndMockReadOnly(t *testing.T) {
	dir := t.TempDir()
	history := filepath.Join(dir, "h")
	actions := filepath.Join(dir, "a")
	outcomes := filepath.Join(dir, "o")
	obs, e := loadHistoryObservations("../../data/m02-sample-observations.json")
	if e != nil {
		t.Fatal(e)
	}
	r, e := NewHistoryRecord("d", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z", obs)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = AppendHistory(history, r); e != nil {
		t.Fatal(e)
	}
	action := m03.HumanActionRecord{ActionID: "a", DecisionID: "d", ActionType: "synthetic", Target: "fixture", PerformedBy: "human", PerformedAt: "2026-09-04T00:00:00Z", MeasurementWindowEnd: "2026-09-05T00:00:00Z", ComplianceReviewed: true}
	raw, _ := json.Marshal(action)
	ap := filepath.Join(dir, "action")
	if e = os.WriteFile(ap, raw, 0600); e != nil {
		t.Fatal(e)
	}
	if runActionStore([]string{"record", history, actions, ap}, io.Discard, io.Discard) != 0 {
		t.Fatal("action")
	}
	op := filepath.Join(dir, "outcome")
	o := m03.OutcomeRecord{OutcomeID: "o", EffectRef: m03.EffectRef{EffectKind: "HUMAN_ACTION", EffectID: "a"}, ObservedAt: "2026-09-05T00:00:00Z", Status: "PENDING", Metrics: map[string]float64{}, SourceRef: "fixture"}
	raw, _ = json.Marshal(o)
	if e = os.WriteFile(op, raw, 0600); e != nil {
		t.Fatal(e)
	}
	if runOutcomeStore([]string{"import", history, actions, outcomes, op}, io.Discard, io.Discard) != 0 {
		t.Fatal("outcome")
	}
	cfg := advisorConfig{DecisionID: "d", Question: "q", AsOf: "2026-09-06T00:00:00Z", MaxAgeHours: ptr(8760)}
	ctx, e := buildAdvisorContext(history, actions, outcomes, cfg)
	if e != nil {
		t.Fatal(e)
	}
	if status := func() string { _, s := checkAdvisorResponse(mockAdvisor(ctx), ctx); return s }(); status != "SUPPORTED" {
		t.Fatal(status)
	}
	var bad advisorConfig
	if e := contracts.DecodeStrict([]byte(`{"decision_id":"d","question":"q","as_of":"x","max_age_hours":1,"extra":true}`), &bad); e == nil {
		t.Fatal("unknown config accepted")
	}
}
func ptr(v int) *int { return &v }
