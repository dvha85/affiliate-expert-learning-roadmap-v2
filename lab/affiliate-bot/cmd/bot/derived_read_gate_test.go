package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestDerivedStoreListsFailClosedWhileHistoryWriterIsActive(t *testing.T) {
	dir := t.TempDir()
	history := filepath.Join(dir, "history.jsonl")
	actions := filepath.Join(dir, "actions.jsonl")
	outcomes := filepath.Join(dir, "outcomes.jsonl")
	evaluations := filepath.Join(dir, "evaluations.jsonl")
	proposals := filepath.Join(dir, "proposals.jsonl")
	reviews := filepath.Join(dir, "reviews.jsonl")
	release, err := acquireHistoryRuntimeGate(history)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	cases := []struct {
		name string
		run  func(*bytes.Buffer, *bytes.Buffer) int
	}{
		{"action", func(out, diag *bytes.Buffer) int {
			return runActionStore([]string{"list", history, actions}, out, diag)
		}},
		{"outcome", func(out, diag *bytes.Buffer) int {
			return runOutcomeStore([]string{"list", history, actions, outcomes}, out, diag)
		}},
		{"evaluation", func(out, diag *bytes.Buffer) int {
			return runEvaluationStore([]string{"list", history, actions, outcomes, evaluations}, out, diag)
		}},
		{"proposal", func(out, diag *bytes.Buffer) int {
			return runImprovementStore("proposal", []string{"list", history, actions, outcomes, evaluations, proposals}, out, diag)
		}},
		{"review", func(out, diag *bytes.Buffer) int {
			return runImprovementStore("review", []string{"list", history, actions, outcomes, evaluations, proposals, reviews}, out, diag)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out, diag bytes.Buffer
			if code := tc.run(&out, &diag); code == 0 {
				t.Fatalf("derived list succeeded during history write: response=%s", out.String())
			}
			var envelope map[string]any
			if err := json.Unmarshal(out.Bytes(), &envelope); err != nil || envelope["status"] != "BUSY" {
				t.Fatalf("derived list did not fail closed: response=%s err=%v stderr=%s", out.String(), err, diag.String())
			}
		})
	}
}
