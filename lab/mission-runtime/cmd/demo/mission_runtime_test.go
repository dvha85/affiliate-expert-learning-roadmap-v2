package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func evalPath(t *testing.T, dir string) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "..", "evals", dir, "cases.json"))
}

func loadCases[T any](t *testing.T, dir string) []T {
	t.Helper()
	raw, err := os.ReadFile(evalPath(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	var cases []T
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	return cases
}

func TestO00EvalPack(t *testing.T) {
	type tc struct {
		CaseID                      string `json:"case_id"`
		ExpectedFinalState          string `json:"expected_final_state"`
		ExpectedExternalSideEffects *bool  `json:"expected_external_side_effects"`
		RequiredStage               string `json:"required_stage"`
	}
	cases := loadCases[tc](t, "O00-safe-walkthrough")
	report := RunSyntheticWalkthrough()
	for _, c := range cases {
		t.Run(c.CaseID, func(t *testing.T) {
			if c.ExpectedFinalState != "" && report.FinalState != c.ExpectedFinalState {
				t.Fatalf("expected %s got %s", c.ExpectedFinalState, report.FinalState)
			}
			if c.ExpectedExternalSideEffects != nil && report.ExternalSideEffects != *c.ExpectedExternalSideEffects {
				t.Fatalf("side effect mismatch")
			}
			if c.RequiredStage != "" {
				found := false
				for _, stage := range report.Stages {
					if stage == c.RequiredStage {
						found = true
					}
				}
				if !found {
					t.Fatalf("missing stage %s in %v", c.RequiredStage, report.Stages)
				}
			}
		})
	}
}

func TestM03EvalPack(t *testing.T) {
	type tc struct {
		CaseID, Mode, Expected string
		Record                 HumanActionRecord `json:"record"`
		Outcome                OutcomeRecord     `json:"outcome"`
	}
	for _, c := range loadCases[tc](t, "M03-tracked-human-action") {
		t.Run(c.CaseID, func(t *testing.T) {
			var got string
			if c.Mode == "action" {
				got = ValidateHumanActionRecord(c.Record)
			} else if c.Mode == "outcome" {
				got = ValidateOutcomeRecord(c.Outcome)
			} else {
				t.Fatalf("unknown mode %s", c.Mode)
			}
			if got != c.Expected {
				t.Fatalf("expected %s got %s", c.Expected, got)
			}
		})
	}
}

func TestM04EvalPack(t *testing.T) {
	type tc struct {
		CaseID      string            `json:"case_id"`
		Output      AdvisorOutput     `json:"output"`
		Evidence    []AdvisorEvidence `json:"evidence"`
		AsOf        string            `json:"as_of"`
		MaxAgeHours int               `json:"max_age_hours"`
		Expected    string            `json:"expected"`
	}
	for _, c := range loadCases[tc](t, "M04-grounded-ai-advisor") {
		t.Run(c.CaseID, func(t *testing.T) {
			if got := EvaluateAdvisorOutput(c.Output, c.Evidence, c.AsOf, c.MaxAgeHours); got != c.Expected {
				t.Fatalf("expected %s got %s", c.Expected, got)
			}
		})
	}
}

func TestM05EvalPack(t *testing.T) {
	type tc struct {
		CaseID   string              `json:"case_id"`
		Proposal ImprovementProposal `json:"proposal"`
		Expected string              `json:"expected"`
	}
	for _, c := range loadCases[tc](t, "M05-reviewed-improvement") {
		t.Run(c.CaseID, func(t *testing.T) {
			if got := EvaluateImprovementProposal(c.Proposal); got != c.Expected {
				t.Fatalf("expected %s got %s", c.Expected, got)
			}
		})
	}
}

func TestM06EvalPack(t *testing.T) {
	type tc struct {
		CaseID   string       `json:"case_id"`
		Request  WatchRequest `json:"request"`
		Expected string       `json:"expected"`
	}
	for _, c := range loadCases[tc](t, "M06-readonly-watcher") {
		t.Run(c.CaseID, func(t *testing.T) {
			if got := EvaluateWatchRequest(c.Request); got != c.Expected {
				t.Fatalf("expected %s got %s", c.Expected, got)
			}
		})
	}
}

func TestM07EvalPack(t *testing.T) {
	type tc struct {
		CaseID               string        `json:"case_id"`
		Proposal             AgentProposal `json:"proposal"`
		Registry             []ToolSpec    `json:"registry"`
		AvailableEvidenceIDs []string      `json:"available_evidence_ids"`
		Expected             string        `json:"expected"`
	}
	for _, c := range loadCases[tc](t, "M07-readonly-evidence-agent") {
		t.Run(c.CaseID, func(t *testing.T) {
			if got := EvaluateAgentProposal(c.Proposal, c.Registry, c.AvailableEvidenceIDs); got != c.Expected {
				t.Fatalf("expected %s got %s", c.Expected, got)
			}
		})
	}
}

func TestSortedStringsDoesNotMutateInput(t *testing.T) {
	in := []string{"b", "a"}
	got := sortedStrings(in)
	if !reflect.DeepEqual(got, []string{"a", "b"}) || !reflect.DeepEqual(in, []string{"b", "a"}) {
		t.Fatal("sort helper must be deterministic and non-mutating")
	}
}
