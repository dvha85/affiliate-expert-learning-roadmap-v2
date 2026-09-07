package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
)

func writeFixtureJSON(path string, v any) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, err = f.Write(raw)
	if err == nil {
		err = f.Sync()
	}
	closed := f.Close()
	if err != nil {
		return err
	}
	return closed
}

// Build only in a newly-created directory. No user input or network evidence.
func buildBR10AdvisorFixture(dir string) (advisorContext, error) {
	price, rate := 100.0, 0.08
	obs := []Observation{{ObservationID: "br11-observation", SubjectID: "A", SourceRef: "fixture:br11-br10/v1", ObservedAt: "2026-09-01T00:00:00Z", AccessMethod: "local_fixture", EvidenceKind: "synthetic", UseContext: "test", ClaimKind: "assumption", State: "observed", Limitation: "Synthetic; not market truth", ProductID: "A", ProductName: "Synthetic laptop stand", Price: &price, CommissionRate: &rate, Currency: "USD"}}
	r, err := NewHistoryRecord("br11-decision", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z", obs)
	if err != nil {
		return advisorContext{}, err
	}
	h, a, o := filepath.Join(dir, "history.jsonl"), filepath.Join(dir, "actions.jsonl"), filepath.Join(dir, "outcomes.jsonl")
	if _, err = AppendHistory(h, r); err != nil {
		return advisorContext{}, err
	}
	action := m03.HumanActionRecord{ActionID: "br11-action", DecisionID: "br11-decision", ActionType: "synthetic", Target: "fixture:none", PerformedBy: "fixture-human", PerformedAt: "2026-09-04T00:00:00Z", MeasurementWindowEnd: "2026-09-05T00:00:00Z", ComplianceReviewed: true}
	ai := filepath.Join(dir, "action-input.json")
	if err = writeFixtureJSON(ai, action); err != nil {
		return advisorContext{}, err
	}
	if runActionStore([]string{"record", h, a, ai}, io.Discard, io.Discard) != 0 {
		return advisorContext{}, errors.New("fixture action rejected")
	}
	outcome := m03.OutcomeRecord{OutcomeID: "br11-outcome", EffectRef: m03.EffectRef{EffectKind: "HUMAN_ACTION", EffectID: "br11-action"}, ObservedAt: "2026-09-05T00:00:00Z", Status: "PENDING", Metrics: map[string]float64{}, SourceRef: "fixture:br11-br10/v1"}
	oi := filepath.Join(dir, "outcome-input.json")
	if err = writeFixtureJSON(oi, outcome); err != nil {
		return advisorContext{}, err
	}
	if runOutcomeStore([]string{"import", h, a, o, oi}, io.Discard, io.Discard) != 0 {
		return advisorContext{}, errors.New("fixture outcome rejected")
	}
	age := 8760
	return buildAdvisorContext(h, a, o, advisorConfig{DecisionID: "br11-decision", Question: "Nêu bằng chứng còn thiếu trước khi kết luận hiệu quả; không thực thi.", AsOf: "2026-09-06T00:00:00Z", MaxAgeHours: &age})
}

// Offline bundle, separate from paid campaign reservations. No API key read.
func runAdvisorFixture(args []string, stdout, stderr io.Writer) int {
	emit := func(status string, path string, err error, code int) int {
		if err != nil {
			fmt.Fprintln(stderr, err)
		}
		out := map[string]any{"command": "advisor fixture-run", "status": status, "execution_permitted": false, "provider": "mock/v1"}
		if path != "" {
			out["bundle_path"] = path
		}
		if err := json.NewEncoder(stdout).Encode(out); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return code
	}
	if len(args) != 2 {
		return emit("USAGE_ERROR", "", errors.New("usage: bot advisor fixture-run EXISTING_OUTPUT_PARENT"), 2)
	}
	parent, err := filepath.Abs(args[1])
	if err != nil {
		return emit("PATH_ERROR", "", err, 1)
	}
	dir, err := os.MkdirTemp(parent, "br11-offline-")
	if err != nil {
		return emit("IO_ERROR", "", err, 1)
	}
	c, err := buildBR10AdvisorFixture(dir)
	if err != nil {
		return emit("FIXTURE_ERROR", dir, err, 1)
	}
	out, status := evaluateAdvisorProvider(context.Background(), mockAdvisorProvider{}, c)
	if status != "SUPPORTED" && status != "ABSTAIN" {
		return emit(status, dir, errors.New("fixture advisor rejected"), 1)
	}
	if err = writeFixtureJSON(filepath.Join(dir, "advisor.json"), map[string]any{"version": "br11-offline-bundle/v1", "execution_permitted": false, "provenance": (mockAdvisorProvider{}).identity(), "context": c, "advisor_output": out}); err != nil {
		return emit("IO_ERROR", dir, err, 1)
	}
	if err = syncCampaignDir(dir); err == nil {
		err = syncCampaignDir(parent)
	}
	if err != nil {
		return emit("IO_ERROR", dir, err, 1)
	}
	return emit(status, dir, nil, 0)
}
