package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

const m03ActionJSON = `{"action_id":"syn-a","decision_id":"syn-d","action_type":"SYNTHETIC_MEASUREMENT","target":"synthetic-target","performed_by":"human","performed_at":"2026-09-03T10:00:00+07:00","measurement_window_end":"2026-09-10T10:00:00+07:00","compliance_reviewed":true}`
const m03OutcomeJSON = `{"outcome_id":"syn-o","effect_ref":{"effect_kind":"HUMAN_ACTION","effect_id":"syn-a"},"observed_at":"2026-09-10T10:00:00+07:00","status":"NO_OBSERVED_OUTCOME","metrics":{"valid_orders":0},"source_ref":"synthetic-report"}`

func TestM03RawSchemaRegressions(t *testing.T) {
	for _, tc := range []struct {
		name, raw string
		action    bool
	}{
		{"missing_boolean", strings.Replace(m03ActionJSON, `,"compliance_reviewed":true`, "", 1), true},
		{"null_boolean", strings.Replace(m03ActionJSON, `"compliance_reviewed":true`, `"compliance_reviewed":null`, 1), true},
		{"unknown_action", strings.Replace(m03ActionJSON, `"action_id":`, `"extra":1,"action_id":`, 1), true},
		{"duplicate_action", strings.Replace(m03ActionJSON, `"action_id":`, `"action_id":"other","action_id":`, 1), true},
		{"case_alias", strings.Replace(m03ActionJSON, `"action_id"`, `"Action_ID"`, 1), true},
		{"machine_action", strings.Replace(m03ActionJSON, `"human"`, `"agent"`, 1), true},
		{"missing_metrics", strings.Replace(m03OutcomeJSON, `,"metrics":{"valid_orders":0}`, "", 1), false},
		{"null_metrics", strings.Replace(m03OutcomeJSON, `{"valid_orders":0}`, `null`, 1), false},
		{"null_metric", strings.Replace(m03OutcomeJSON, `"valid_orders":0`, `"valid_orders":null`, 1), false},
		{"negative_metric", strings.Replace(m03OutcomeJSON, `"valid_orders":0`, `"valid_orders":-1`, 1), false},
		{"nested_duplicate", strings.Replace(m03OutcomeJSON, `"effect_id":`, `"effect_id":"other","effect_id":`, 1), false},
		{"nested_unknown", strings.Replace(m03OutcomeJSON, `"effect_id":`, `"extra":1,"effect_id":`, 1), false},
		{"internal_alias", strings.Replace(m03OutcomeJSON, `"outcome_id":`, `"action_id":"syn-a","outcome_id":`, 1), false},
		{"bad_time", strings.Replace(m03OutcomeJSON, `2026-09-10T10:00:00+07:00`, `not-a-time`, 1), false},
		{"bad_status", strings.Replace(m03OutcomeJSON, `NO_OBSERVED_OUTCOME`, `SUCCESS`, 1), false},
		{"trailing", m03OutcomeJSON + ` {}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var state string
			if tc.action {
				_, state = DecodeM03Action([]byte(tc.raw))
			} else {
				_, state = DecodeM03Outcome([]byte(tc.raw))
			}
			if state != "INVALID_SCHEMA" {
				t.Fatalf("want INVALID_SCHEMA, got %s", state)
			}
		})
	}
}

func TestM03RequiredFieldsBeforeTypedDecode(t *testing.T) {
	for _, input := range []struct {
		name, raw string
		decode    func([]byte) string
	}{
		{"action", m03ActionJSON, func(raw []byte) string { _, state := DecodeM03Action(raw); return state }},
		{"outcome", m03OutcomeJSON, func(raw []byte) string { _, state := DecodeM03Outcome(raw); return state }},
	} {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal([]byte(input.raw), &fields); err != nil {
			t.Fatal(err)
		}
		keys := make([]string, 0, len(fields))
		for key := range fields {
			keys = append(keys, key)
		}
		for _, key := range keys {
			original := fields[key]
			for _, mutation := range []string{"missing", "null"} {
				t.Run(input.name+"/"+key+"/"+mutation, func(t *testing.T) {
					if mutation == "missing" {
						delete(fields, key)
					} else {
						fields[key] = json.RawMessage(`null`)
					}
					raw, err := json.Marshal(fields)
					if err != nil {
						t.Fatal(err)
					}
					if state := input.decode(raw); state != "INVALID_SCHEMA" {
						t.Fatalf("got %s", state)
					}
					fields[key] = original
				})
			}
		}
	}
}

func TestM03PairSemanticsAfterSchema(t *testing.T) {
	for _, tc := range []struct{ name, action, outcome, want string }{
		{"zero_at_end", m03ActionJSON, m03OutcomeJSON, missionValid},
		{"empty_metrics", m03ActionJSON, strings.Replace(m03OutcomeJSON, `{"valid_orders":0}`, `{}`, 1), missionValid},
		{"review_required", strings.Replace(m03ActionJSON, `true`, `false`, 1), m03OutcomeJSON, "HUMAN_REVIEW"},
		{"early_zero", m03ActionJSON, strings.Replace(m03OutcomeJSON, `2026-09-10T10:00:00+07:00`, `2026-09-03T11:00:00+07:00`, 1), "MEASUREMENT_WINDOW_OPEN"},
		{"pending_before_end", m03ActionJSON, strings.Replace(strings.Replace(m03OutcomeJSON, `NO_OBSERVED_OUTCOME`, `PENDING`, 1), `2026-09-10T10:00:00+07:00`, `2026-09-03T11:00:00+07:00`, 1), missionValid},
		{"before_action", m03ActionJSON, strings.Replace(m03OutcomeJSON, `2026-09-10T10:00:00+07:00`, `2026-09-03T09:00:00+07:00`, 1), "OUTCOME_BEFORE_ACTION"},
		{"equivalent_offset", m03ActionJSON, strings.Replace(m03OutcomeJSON, `2026-09-10T10:00:00+07:00`, `2026-09-10T03:00:00Z`, 1), missionValid},
		{"orphan", m03ActionJSON, strings.Replace(m03OutcomeJSON, `"syn-a"`, `"missing"`, 1), "BROKEN_LINK"},
		{"wrong_kind", m03ActionJSON, strings.Replace(m03OutcomeJSON, `HUMAN_ACTION`, `MACHINE_EXECUTION`, 1), "BROKEN_LINK"},
		{"whitespace_id", strings.Replace(m03ActionJSON, `"syn-d"`, `" "`, 1), m03OutcomeJSON, missionInvalid},
		{"reversed_window", strings.Replace(m03ActionJSON, `2026-09-10T10:00:00+07:00`, `2026-09-02T10:00:00+07:00`, 1), m03OutcomeJSON, missionInvalid},
		{"schema_before_review", strings.Replace(m03ActionJSON, `true`, `false`, 1), strings.Replace(m03OutcomeJSON, `{"valid_orders":0}`, `null`, 1), "INVALID_SCHEMA"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, state := CheckM03Pair([]byte(tc.action), []byte(tc.outcome))
			if state != tc.want {
				t.Fatalf("want %s got %s", tc.want, state)
			}
			if state == missionValid {
				if result.Outcome.ActionID != "" {
					t.Fatal("internal alias populated")
				}
				if tc.name == "empty_metrics" && len(result.Outcome.Metrics) != 0 {
					t.Fatal("missing metrics fabricated")
				}
				if tc.name == "zero_at_end" {
					if v, ok := result.Outcome.Metrics["valid_orders"]; !ok || v != 0 {
						t.Fatal("observed zero lost")
					}
				}
			} else if result.Result != "" {
				t.Fatal("rejected pair has success result")
			}
		})
	}
}

func TestM03RawEvalPack(t *testing.T) {
	type testCase struct {
		ID          string          `json:"case_id"`
		Mode        string          `json:"mode"`
		Expected    string          `json:"expected"`
		RawExpected string          `json:"raw_expected"`
		Action      json.RawMessage `json:"record"`
		Outcome     json.RawMessage `json:"outcome"`
	}
	for _, c := range loadCases[testCase](t, "M03-tracked-human-action") {
		t.Run(c.ID, func(t *testing.T) {
			var state string
			switch c.Mode {
			case "action":
				a, s := DecodeM03Action(c.Action)
				state = s
				if s == missionValid {
					state = ValidateHumanActionRecord(a)
				}
			case "outcome":
				o, s := DecodeM03Outcome(c.Outcome)
				state = s
				if s == missionValid {
					state = ValidateOutcomeRecord(o)
				}
			case "link":
				_, state = CheckM03Pair(c.Action, c.Outcome)
			default:
				t.Fatal("unknown mode")
			}
			want := c.Expected
			if c.RawExpected != "" {
				want = c.RawExpected
			}
			if state != want {
				t.Fatalf("want %s got %s", want, state)
			}
		})
	}
}

type m03BrokenWriter struct{}

func (m03BrokenWriter) Write([]byte) (int, error) { return 0, errors.New("sink unavailable") }

func TestM03CLIReadOnlyAndOutput(t *testing.T) {
	dir := t.TempDir()
	aPath, oPath := filepath.Join(dir, "action.json"), filepath.Join(dir, "outcome.json")
	for _, item := range []struct{ path, raw string }{{aPath, m03ActionJSON}, {oPath, m03OutcomeJSON}} {
		if err := os.WriteFile(item.path, []byte(item.raw), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var output bytes.Buffer
	if err := runM03Check(&output, []string{aPath, oPath}); err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Action  json.RawMessage `json:"action"`
		Outcome json.RawMessage `json:"outcome"`
		Result  string          `json:"result"`
	}
	if err := contracts.DecodeStrict(output.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Result != missionValid {
		t.Fatal(envelope.Result)
	}
	if err := contracts.ValidateRaw("action-record.schema.json", envelope.Action); err != nil {
		t.Fatal(err)
	}
	if err := contracts.ValidateRaw("outcome-record.schema.json", envelope.Outcome); err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(envelope.Outcome, []byte(`"action_id"`)) {
		t.Fatal("internal alias leaked")
	}
	if err := runM03Check(m03BrokenWriter{}, []string{aPath, oPath}); err == nil {
		t.Fatal("writer error swallowed")
	}
	for _, args := range [][]string{nil, {aPath}, {aPath, filepath.Join(dir, "missing")}, {aPath, oPath, "extra"}} {
		output.Reset()
		if err := runM03Check(&output, args); err == nil || output.Len() != 0 {
			t.Fatal("bad args/file accepted")
		}
	}
	for _, invalid := range []string{
		strings.Replace(m03OutcomeJSON, `{"valid_orders":0}`, `null`, 1),
		strings.Replace(m03OutcomeJSON, `2026-09-10T10:00:00+07:00`, `2026-09-03T11:00:00+07:00`, 1),
	} {
		if err := os.WriteFile(oPath, []byte(invalid), 0600); err != nil {
			t.Fatal(err)
		}
		output.Reset()
		if err := runM03Check(&output, []string{aPath, oPath}); err == nil || output.Len() != 0 {
			t.Fatal("rejected input emitted envelope")
		}
		got, err := os.ReadFile(oPath)
		if err != nil || string(got) != invalid {
			t.Fatal("input overwritten")
		}
	}
	got, err := os.ReadFile(aPath)
	if err != nil || string(got) != m03ActionJSON {
		t.Fatal("action overwritten")
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 2 {
		t.Fatal("unexpected store/file created")
	}
}
