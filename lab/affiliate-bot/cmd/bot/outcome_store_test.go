package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
)

func TestOutcomeStoreLifecycle(t *testing.T) {
	dir := t.TempDir()
	h := filepath.Join(dir, "history")
	aPath := filepath.Join(dir, "actions")
	outPath := filepath.Join(dir, "outcomes")
	input := filepath.Join(dir, "input")
	obs, err := loadHistoryObservations("../../data/m02-sample-observations.json")
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewHistoryRecord("d", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z", obs)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = AppendHistory(h, r); err != nil {
		t.Fatal(err)
	}
	a := m03.HumanActionRecord{ActionID: "a", DecisionID: "d", ActionType: "synthetic", Target: "fixture:test", PerformedBy: "human", PerformedAt: "2026-09-04T00:00:00Z", MeasurementWindowEnd: "2026-09-05T00:00:00Z", ComplianceReviewed: true}
	raw, _ := json.Marshal(a)
	if err = os.WriteFile(input, raw, 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if runActionStore([]string{"record", h, aPath, input}, &stdout, &stderr) != 0 {
		t.Fatal(stderr.String())
	}
	beforeH, _ := os.ReadFile(h)
	beforeA, _ := os.ReadFile(aPath)
	o := m03.OutcomeRecord{OutcomeID: "o", EffectRef: m03.EffectRef{EffectKind: "HUMAN_ACTION", EffectID: "a"}, ObservedAt: "2026-09-04T01:00:00Z", Status: "PENDING", Metrics: map[string]float64{}, SourceRef: "fixture:test"}
	call := func(o m03.OutcomeRecord, status string, code int) {
		t.Helper()
		raw, _ := json.Marshal(o)
		if err := os.WriteFile(input, raw, 0600); err != nil {
			t.Fatal(err)
		}
		stdout.Reset()
		stderr.Reset()
		if got := runOutcomeStore([]string{"import", h, aPath, outPath, input}, &stdout, &stderr); got != code {
			t.Fatal(got, stderr.String())
		}
		var e map[string]any
		if err := json.Unmarshal(stdout.Bytes(), &e); err != nil {
			t.Fatal(err)
		}
		if e["status"] != status || e["execution_permitted"] != false {
			t.Fatal(e)
		}
		if code != 0 && e["artifact"] != nil {
			t.Fatal(e)
		}
		after, _ := os.ReadFile(input)
		if !bytes.Equal(raw, after) {
			t.Fatal("input rewritten")
		}
	}
	orphan := o
	orphan.EffectRef.EffectID = "absent"
	call(orphan, "ORPHAN_ACTION", 1)
	if _, err := os.Stat(outPath); !os.IsNotExist(err) {
		t.Fatal("orphan created store")
	}
	call(o, "APPENDED", 0)
	call(o, "EXACT_DUPLICATE", 0)
	saved, _ := os.ReadFile(outPath)
	for _, tc := range []struct {
		change func(*m03.OutcomeRecord)
		status string
	}{
		{func(o *m03.OutcomeRecord) { o.Metrics = map[string]float64{"clicks": 0} }, "CONFLICT"},
		{func(o *m03.OutcomeRecord) { o.EffectRef.EffectKind = "MACHINE_EXECUTION" }, "REJECT_MACHINE_EXECUTION"},
		{func(o *m03.OutcomeRecord) { o.ObservedAt = "2026-09-03T23:59:59Z" }, "OUTCOME_BEFORE_ACTION"},
		{func(o *m03.OutcomeRecord) { o.Status = "NO_OBSERVED_OUTCOME" }, "MEASUREMENT_WINDOW_OPEN"},
		{func(o *m03.OutcomeRecord) { o.Metrics = map[string]float64{"clicks": -1} }, "INVALID_SCHEMA"},
	} {
		candidate := o
		tc.change(&candidate)
		call(candidate, tc.status, 1)
	}
	after, _ := os.ReadFile(outPath)
	if !bytes.Equal(saved, after) {
		t.Fatal("rejection modified store")
	}
	o.OutcomeID = "zero"
	o.Status = "NO_OBSERVED_OUTCOME"
	o.ObservedAt = "2026-09-05T07:00:00+07:00"
	o.Metrics = map[string]float64{"clicks": 0}
	call(o, "APPENDED", 0)
	o.OutcomeID = "late"
	o.Status = "PAID"
	o.ObservedAt = "2026-09-06T00:00:00Z"
	o.Metrics = map[string]float64{"commission": 8}
	call(o, "APPENDED", 0)
	loaded, err := loadOutcomes(outPath, []m03.HumanActionRecord{a})
	if err != nil || len(loaded) != 3 {
		t.Fatal(err)
	}
	if len(loaded[0].Metrics) != 0 || loaded[0].Status != "PENDING" || loaded[1].Metrics["clicks"] != 0 {
		t.Fatal("pending/zero changed")
	}
	afterH, _ := os.ReadFile(h)
	afterA, _ := os.ReadFile(aPath)
	if !bytes.Equal(beforeH, afterH) || !bytes.Equal(beforeA, afterA) {
		t.Fatal("upstream changed")
	}
	stdout.Reset()
	if runOutcomeStore([]string{"list", h, aPath, outPath}, &stdout, &stderr) != 0 {
		t.Fatal(stderr.String())
	}
	if runOutcomeStore([]string{"import", h, aPath, aPath, input}, &stdout, &stderr) != 1 {
		t.Fatal("alias accepted")
	}
	if err = os.WriteFile(outPath, []byte("broken\n"), 0600); err != nil {
		t.Fatal(err)
	}
	call(o, "STORE_ERROR", 1)
	after, _ = os.ReadFile(outPath)
	if string(after) != "broken\n" {
		t.Fatal("corruption overwritten")
	}
}

func TestOutcomeMetricPrecision(t *testing.T) {
	a := m03.HumanActionRecord{ActionID: "a", DecisionID: "d", ActionType: "test", Target: "fixture:t", PerformedBy: "human", PerformedAt: "2026-09-04T00:00:00Z", MeasurementWindowEnd: "2026-09-05T00:00:00Z", ComplianceReviewed: true}
	raw := `{"outcome_id":"o","effect_ref":{"effect_kind":"HUMAN_ACTION","effect_id":"a"},"observed_at":"2026-09-04T00:00:00Z","status":"PENDING","metrics":{"clicks":VALUE},"source_ref":"fixture:t"}`
	for _, value := range []string{"1e-1000", "9007199254740993", "0.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"} {
		if _, status := linkedOutcome([]byte(strings.Replace(raw, "VALUE", value, 1)), []m03.HumanActionRecord{a}); status != "METRIC_PRECISION_ERROR" {
			t.Fatal(value, status)
		}
	}
}
