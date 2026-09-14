package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/lab/affiliate-bot/internal/store"
)

func appendThenLoseAcknowledgement(path string, encoded []byte) error {
	if err := (store.JSONL{}).AppendLine(path, encoded); err != nil {
		return err
	}
	return errors.New("injected acknowledgement loss after append")
}

func commandStatus(t *testing.T, run func(*bytes.Buffer, *bytes.Buffer) int) (int, map[string]any) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := run(&stdout, &stderr)
	var envelope map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("decode command envelope: %v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
	}
	return code, envelope
}

func TestActionOutcomeEvaluationAndImprovementDiscloseVisibleAppendUncertainty(t *testing.T) {
	t.Run("action", func(t *testing.T) {
		dir := t.TempDir()
		history := filepath.Join(dir, "history.jsonl")
		actions := filepath.Join(dir, "actions.jsonl")
		input := filepath.Join(dir, "action.json")
		obs, err := loadHistoryObservations("../../data/m02-sample-observations.json")
		if err != nil {
			t.Fatal(err)
		}
		record, err := NewHistoryRecord("append-recovery-decision", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z", obs)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := AppendHistory(history, record); err != nil {
			t.Fatal(err)
		}
		action := m03.HumanActionRecord{ActionID: "append-recovery-action", DecisionID: record.RecordedResult.DecisionID, ActionType: "synthetic_manual_post", Target: "fixture:append-recovery", PerformedBy: "human", PerformedAt: "2026-09-04T00:00:00Z", MeasurementWindowEnd: "2026-09-05T00:00:00Z", ComplianceReviewed: true}
		raw, _ := json.Marshal(action)
		if err := os.WriteFile(input, raw, 0600); err != nil {
			t.Fatal(err)
		}
		actionAppend = appendThenLoseAcknowledgement
		t.Cleanup(func() {
			actionAppend = func(path string, record []byte) error { return (store.JSONL{}).AppendLine(path, record) }
		})
		args := []string{"record", history, actions, input}
		if code, response := commandStatus(t, func(out, errOut *bytes.Buffer) int { return runActionStore(args, out, errOut) }); code == 0 || response["status"] != "PUBLISHED_RECOVERY_REQUIRED" || response["artifact"] == nil {
			t.Fatalf("visible action append uncertainty hidden: code=%d response=%+v", code, response)
		}
		loaded, err := loadActions(actions, []HistoryRecord{record})
		if err != nil || len(loaded) != 1 || loaded[0] != action {
			t.Fatalf("action was not canonically replayable: actions=%+v err=%v", loaded, err)
		}
		actionAppend = func(path string, record []byte) error { return (store.JSONL{}).AppendLine(path, record) }
		if code, response := commandStatus(t, func(out, errOut *bytes.Buffer) int { return runActionStore(args, out, errOut) }); code != 0 || response["status"] != "EXACT_DUPLICATE" {
			t.Fatalf("action exact retry failed: code=%d response=%+v", code, response)
		}
	})

	t.Run("outcome", func(t *testing.T) {
		dir := t.TempDir()
		history := filepath.Join(dir, "history.jsonl")
		actions := filepath.Join(dir, "actions.jsonl")
		outcomes := filepath.Join(dir, "outcomes.jsonl")
		input := filepath.Join(dir, "outcome.json")
		obs, err := loadHistoryObservations("../../data/m02-sample-observations.json")
		if err != nil {
			t.Fatal(err)
		}
		record, err := NewHistoryRecord("append-recovery-outcome-decision", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z", obs)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := AppendHistory(history, record); err != nil {
			t.Fatal(err)
		}
		action := m03.HumanActionRecord{ActionID: "append-recovery-outcome-action", DecisionID: record.RecordedResult.DecisionID, ActionType: "synthetic_manual_post", Target: "fixture:append-recovery", PerformedBy: "human", PerformedAt: "2026-09-04T00:00:00Z", MeasurementWindowEnd: "2026-09-05T00:00:00Z", ComplianceReviewed: true}
		actionRaw, _ := json.Marshal(action)
		if err := os.WriteFile(input, actionRaw, 0600); err != nil {
			t.Fatal(err)
		}
		if code, response := commandStatus(t, func(out, errOut *bytes.Buffer) int {
			return runActionStore([]string{"record", history, actions, input}, out, errOut)
		}); code != 0 || response["status"] != "APPENDED" {
			t.Fatalf("action setup failed: code=%d response=%+v", code, response)
		}
		outcome := m03.OutcomeRecord{OutcomeID: "append-recovery-outcome", EffectRef: m03.EffectRef{EffectKind: "HUMAN_ACTION", EffectID: action.ActionID}, ObservedAt: "2026-09-04T01:00:00Z", Status: "PENDING", Metrics: map[string]float64{}, SourceRef: "fixture:append-recovery"}
		outcomeRaw, _ := json.Marshal(outcome)
		if err := os.WriteFile(input, outcomeRaw, 0600); err != nil {
			t.Fatal(err)
		}
		outcomeAppend = appendThenLoseAcknowledgement
		t.Cleanup(func() {
			outcomeAppend = func(path string, record []byte) error { return (store.JSONL{}).AppendLine(path, record) }
		})
		args := []string{"import", history, actions, outcomes, input}
		if code, response := commandStatus(t, func(out, errOut *bytes.Buffer) int { return runOutcomeStore(args, out, errOut) }); code == 0 || response["status"] != "PUBLISHED_RECOVERY_REQUIRED" || response["artifact"] == nil {
			t.Fatalf("visible outcome append uncertainty hidden: code=%d response=%+v", code, response)
		}
		loadedActions, err := loadActions(actions, []HistoryRecord{record})
		loaded, loadErr := loadOutcomes(outcomes, loadedActions)
		if err != nil || loadErr != nil || len(loaded) != 1 || !sameOutcome(loaded[0], outcome) {
			t.Fatalf("outcome was not canonically replayable: outcomes=%+v actionErr=%v outcomeErr=%v", loaded, err, loadErr)
		}
		outcomeAppend = func(path string, record []byte) error { return (store.JSONL{}).AppendLine(path, record) }
		if code, response := commandStatus(t, func(out, errOut *bytes.Buffer) int { return runOutcomeStore(args, out, errOut) }); code != 0 || response["status"] != "EXACT_DUPLICATE" {
			t.Fatalf("outcome exact retry failed: code=%d response=%+v", code, response)
		}
	})

	t.Run("evaluation", func(t *testing.T) {
		args, _ := evaluationFixture(t)
		evaluationAppend = appendThenLoseAcknowledgement
		t.Cleanup(func() {
			evaluationAppend = func(path string, record []byte) error { return (store.JSONL{}).AppendLine(path, record) }
		})
		if code, response := commandStatus(t, func(out, errOut *bytes.Buffer) int { return runEvaluationStore(args, out, errOut) }); code == 0 || response["status"] != "PUBLISHED_RECOVERY_REQUIRED" || response["artifact"] == nil {
			t.Fatalf("visible evaluation append uncertainty hidden: code=%d response=%+v", code, response)
		}
		evaluationAppend = func(path string, record []byte) error { return (store.JSONL{}).AppendLine(path, record) }
		if code, response := commandStatus(t, func(out, errOut *bytes.Buffer) int { return runEvaluationStore(args, out, errOut) }); code != 0 || response["status"] != "EXACT_DUPLICATE" {
			t.Fatalf("evaluation exact retry failed: code=%d response=%+v", code, response)
		}
	})

	t.Run("proposal and review", func(t *testing.T) {
		proposalArgs, reviewArgs := improvementFixture(t)
		improvementAppend = appendThenLoseAcknowledgement
		t.Cleanup(func() {
			improvementAppend = func(path string, record []byte) error { return (store.JSONL{}).AppendLine(path, record) }
		})
		if code, response := commandStatus(t, func(out, errOut *bytes.Buffer) int { return runImprovementStore("proposal", proposalArgs, out, errOut) }); code == 0 || response["status"] != "PUBLISHED_RECOVERY_REQUIRED" || response["artifact"] == nil {
			t.Fatalf("visible proposal append uncertainty hidden: code=%d response=%+v", code, response)
		}
		improvementAppend = func(path string, record []byte) error { return (store.JSONL{}).AppendLine(path, record) }
		if code, response := commandStatus(t, func(out, errOut *bytes.Buffer) int { return runImprovementStore("proposal", proposalArgs, out, errOut) }); code != 0 || response["status"] != "EXACT_DUPLICATE" {
			t.Fatalf("proposal exact retry failed: code=%d response=%+v", code, response)
		}
		improvementAppend = appendThenLoseAcknowledgement
		if code, response := commandStatus(t, func(out, errOut *bytes.Buffer) int { return runImprovementStore("review", reviewArgs, out, errOut) }); code == 0 || response["status"] != "PUBLISHED_RECOVERY_REQUIRED" || response["artifact"] == nil {
			t.Fatalf("visible review append uncertainty hidden: code=%d response=%+v", code, response)
		}
		improvementAppend = func(path string, record []byte) error { return (store.JSONL{}).AppendLine(path, record) }
		if code, response := commandStatus(t, func(out, errOut *bytes.Buffer) int { return runImprovementStore("review", reviewArgs, out, errOut) }); code != 0 || response["status"] != "EXACT_DUPLICATE" {
			t.Fatalf("review exact retry failed: code=%d response=%+v", code, response)
		}
	})
}
