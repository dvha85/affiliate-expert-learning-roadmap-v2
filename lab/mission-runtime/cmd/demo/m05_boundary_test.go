package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

const m05EvaluationJSON = `{"evaluation_id":"syn-ev","decision_id":"syn-d","effect_ref":{"effect_kind":"HUMAN_ACTION","effect_id":"syn-a"},"outcome_ids":["syn-o"],"evaluated_at":"2026-09-10T11:00:00+07:00","result":"INCONCLUSIVE","evidence_ids":[],"limitations":["synthetic only"]}`
const m05ProposalJSON = `{"proposal_id":"syn-p","evaluation_ids":["syn-ev"],"current_version":"v1","proposed_version":"v2","change_summary":"clarify missing versus zero","expected_benefit":"avoid premature conclusions","rollback":"restore v1 after manual review","auto_apply":false}`
const m05ReviewJSON = `{"review_id":"syn-r","proposal_id":"syn-p","reviewed_by":"human","reviewed_at":"2026-09-10T12:00:00+07:00","decision":"REQUEST_CHANGES","reason":"collect actual evidence first"}`

func TestM05RawSchema(t *testing.T) {
	for _, input := range []struct {
		name, raw string
		decode    func([]byte) string
	}{
		{"evaluation", m05EvaluationJSON, func(raw []byte) string { _, s := DecodeM05Evaluation(raw); return s }},
		{"proposal", m05ProposalJSON, func(raw []byte) string { _, s := DecodeM05Proposal(raw); return s }},
		{"review", m05ReviewJSON, func(raw []byte) string { _, s := DecodeM05Review(raw); return s }},
	} {
		if s := input.decode([]byte(input.raw)); s != missionValid {
			t.Fatal(input.name, s)
		}
		var original map[string]json.RawMessage
		if err := json.Unmarshal([]byte(input.raw), &original); err != nil {
			t.Fatal(err)
		}
		for key := range original {
			for _, mutation := range []string{"missing", "null"} {
				t.Run(input.name+"/"+key+"/"+mutation, func(t *testing.T) {
					fields := map[string]json.RawMessage{}
					for k, v := range original {
						fields[k] = v
					}
					if mutation == "missing" {
						delete(fields, key)
					} else {
						fields[key] = json.RawMessage(`null`)
					}
					raw, err := json.Marshal(fields)
					if err != nil {
						t.Fatal(err)
					}
					if s := input.decode(raw); s != "INVALID_SCHEMA" {
						t.Fatalf("want INVALID_SCHEMA got %s", s)
					}
				})
			}
		}
		for _, raw := range []string{`null`, `[]`, input.raw + ` {}`, strings.Replace(input.raw, `{`, `{"extra":1,`, 1), strings.Replace(input.raw, `{`, `{"extra":1,"extra":2,`, 1)} {
			if s := input.decode([]byte(raw)); s != "INVALID_SCHEMA" {
				t.Errorf("%s accepted %s", input.name, raw)
			}
		}
	}
	for _, tc := range []struct {
		name, raw string
		decode    func([]byte) string
	}{
		{"auto_apply", strings.Replace(m05ProposalJSON, `false`, `true`, 1), func(raw []byte) string { _, s := DecodeM05Proposal(raw); return s }},
		{"machine_review", strings.Replace(m05ReviewJSON, `"human"`, `"agent"`, 1), func(raw []byte) string { _, s := DecodeM05Review(raw); return s }},
		{"duplicate_outcome_ids", strings.Replace(m05EvaluationJSON, `["syn-o"]`, `["syn-o","syn-o"]`, 1), func(raw []byte) string { _, s := DecodeM05Evaluation(raw); return s }},
		{"nested_duplicate", strings.Replace(m05EvaluationJSON, `"effect_id":`, `"effect_id":"other","effect_id":`, 1), func(raw []byte) string { _, s := DecodeM05Evaluation(raw); return s }},
		{"alias", strings.Replace(m05EvaluationJSON, `"evaluation_id":`, `"action_id":"syn-a","evaluation_id":`, 1), func(raw []byte) string { _, s := DecodeM05Evaluation(raw); return s }},
		{"bad_date", strings.Replace(m05ReviewJSON, `2026-09-10T12:00:00+07:00`, `bad-date`, 1), func(raw []byte) string { _, s := DecodeM05Review(raw); return s }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if s := tc.decode([]byte(tc.raw)); s != "INVALID_SCHEMA" {
				t.Fatal(s)
			}
		})
	}
}

func m05Inputs() []string {
	return []string{m03ActionJSON, "[" + m03OutcomeJSON + "]", "[" + m05EvaluationJSON + "]", m05ProposalJSON, m05ReviewJSON}
}

func checkM05Strings(inputs []string) (M05CheckResult, string) {
	return CheckM05Chain([]byte(inputs[0]), []byte(inputs[1]), []byte(inputs[2]), []byte(inputs[3]), []byte(inputs[4]))
}

func TestM05ChainSemantics(t *testing.T) {
	for _, tc := range []struct {
		name           string
		index          int
		old, new, want string
	}{
		{"inconclusive", 2, "INCONCLUSIVE", "INCONCLUSIVE", missionValid},
		{"approval_no_execution", 4, "REQUEST_CHANGES", "APPROVE_FOR_MANUAL_CHANGE", missionValid},
		{"rejection_is_valid_record", 4, "REQUEST_CHANGES", "REJECT", missionValid},
		{"action_review", 0, "true", "false", "HUMAN_REVIEW"},
		{"wrong_decision", 2, `"syn-d"`, `"other"`, "BROKEN_LINK"},
		{"wrong_effect", 2, `"syn-a"`, `"other"`, "BROKEN_LINK"},
		{"missing_outcome", 2, `"syn-o"`, `"missing"`, "BROKEN_LINK"},
		{"early_absence", 1, "2026-09-10T10:00:00+07:00", "2026-09-03T11:00:00+07:00", "MEASUREMENT_WINDOW_OPEN"},
		{"early_evaluation", 2, "2026-09-10T11:00:00+07:00", "2026-09-10T09:00:00+07:00", "EVALUATION_BEFORE_OUTCOME"},
		{"equivalent_evaluation_time", 2, "2026-09-10T11:00:00+07:00", "2026-09-10T03:00:00Z", missionValid},
		{"proposal_orphan", 3, `"syn-ev"`, `"missing"`, "BROKEN_LINK"},
		{"duplicate_proposal_ref", 3, `["syn-ev"]`, `["syn-ev","syn-ev"]`, "DUPLICATE_ID"},
		{"same_version", 3, `"v2"`, `"v1"`, missionInvalid},
		{"empty_rollback", 3, `restore v1 after manual review`, ` `, missionInvalid},
		{"review_orphan", 4, `"syn-p"`, `"missing"`, "BROKEN_LINK"},
		{"early_review", 4, "2026-09-10T12:00:00+07:00", "2026-09-10T10:00:00+07:00", "REVIEW_BEFORE_EVALUATION"},
		{"blank_reason", 4, "collect actual evidence first", " ", missionInvalid},
		{"auto_apply", 3, "false", "true", "INVALID_SCHEMA"},
		{"machine_review", 4, `"human"`, `"agent"`, "INVALID_SCHEMA"},
		{"bad_status", 2, "INCONCLUSIVE", "SUCCESS", "INVALID_SCHEMA"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			inputs := m05Inputs()
			inputs[tc.index] = strings.Replace(inputs[tc.index], tc.old, tc.new, 1)
			result, state := checkM05Strings(inputs)
			if state != tc.want {
				t.Fatalf("want %s got %s", tc.want, state)
			}
			if result.ExecutionAuthorized || result.Proposal.AutoApply {
				t.Fatal("authority widened")
			}
			if state != missionValid && result.Result != "" {
				t.Fatal("rejected chain has success envelope")
			}
		})
	}
	for _, index := range []int{1, 2} {
		for _, bad := range []string{`null`, `[]`, `{}`, `[null]`} {
			inputs := m05Inputs()
			inputs[index] = bad
			if _, state := checkM05Strings(inputs); state != "INVALID_SCHEMA" {
				t.Fatal(index, bad, state)
			}
		}
		inputs := m05Inputs()
		items := strings.TrimSuffix(strings.TrimPrefix(inputs[index], "["), "]")
		inputs[index] = "[" + items + "," + items + "]"
		if _, state := checkM05Strings(inputs); state != "DUPLICATE_ID" {
			t.Fatal(index, state)
		}
	}
	inputs := m05Inputs()
	inputs[0] = strings.Replace(inputs[0], "true", "false", 1)
	inputs[4] = `{}`
	if _, state := checkM05Strings(inputs); state != "INVALID_SCHEMA" {
		t.Fatal("semantic check preceded schema", state)
	}
}

func TestM05OptionalFields(t *testing.T) {
	for _, notes := range []string{`""`, `"keep this note"`} {
		raw := strings.Replace(m05EvaluationJSON, `"evaluation_id":`, `"notes":`+notes+`,"evaluation_id":`, 1)
		e, state := DecodeM05Evaluation([]byte(raw))
		if state != missionValid {
			t.Fatal(state)
		}
		encoded, err := json.Marshal(e)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(encoded, []byte(`"notes":`+notes)) {
			t.Fatal("notes lost", string(encoded))
		}
	}
	for _, risks := range []string{"", `,"risks":[]`, `,"risks":["overfit"]`} {
		raw := strings.TrimSuffix(m05ProposalJSON, "}") + risks + "}"
		p, state := DecodeM05Proposal([]byte(raw))
		if state != missionValid {
			t.Fatal(state)
		}
		encoded, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		if err := contracts.ValidateRaw("improvement-proposal.schema.json", encoded); err != nil {
			t.Fatal(err)
		}
		if risks == `,"risks":["overfit"]` && !bytes.Contains(encoded, []byte("overfit")) {
			t.Fatal("risk lost")
		}
	}
	for _, tc := range []struct {
		raw  string
		eval bool
	}{
		{strings.Replace(m05EvaluationJSON, `{`, `{"notes":null,`, 1), true},
		{strings.Replace(m05ProposalJSON, `{`, `{"risks":null,`, 1), false},
		{strings.Replace(m05EvaluationJSON, `"evaluation_id"`, `"Evaluation_ID"`, 1), true},
	} {
		var state string
		if tc.eval {
			_, state = DecodeM05Evaluation([]byte(tc.raw))
		} else {
			_, state = DecodeM05Proposal([]byte(tc.raw))
		}
		if state != "INVALID_SCHEMA" {
			t.Fatal(state)
		}
	}
}

func TestM05ValidatesWholeCollection(t *testing.T) {
	inputs := m05Inputs()
	second := strings.Replace(m05EvaluationJSON, `"syn-ev"`, `"syn-ev-2"`, 1)
	inputs[2] = "[" + m05EvaluationJSON + "," + second + "]"
	inputs[3] = strings.Replace(inputs[3], `["syn-ev"]`, `["syn-ev","syn-ev-2"]`, 1)
	if _, state := checkM05Strings(inputs); state != missionValid {
		t.Fatal(state)
	}
	inputs = m05Inputs()
	bad := strings.Replace(strings.Replace(m03OutcomeJSON, `"syn-o"`, `"unused"`, 1), `"syn-a"`, `"other-action"`, 1)
	inputs[1] = "[" + m03OutcomeJSON + "," + bad + "]"
	if _, state := checkM05Strings(inputs); state != "BROKEN_LINK" {
		t.Fatal("unused invalid outcome ignored", state)
	}
	inputs = m05Inputs()
	bad = strings.Replace(second, `"syn-d"`, `"other-decision"`, 1)
	inputs[2] = "[" + m05EvaluationJSON + "," + bad + "]"
	if _, state := checkM05Strings(inputs); state != "BROKEN_LINK" {
		t.Fatal("unused invalid evaluation ignored", state)
	}
}

func TestM05CLI(t *testing.T) {
	dir := t.TempDir()
	inputs := m05Inputs()
	paths := make([]string, len(inputs))
	for i, name := range []string{"action", "outcomes", "evaluations", "proposal", "review"} {
		paths[i] = filepath.Join(dir, name+".json")
		if err := os.WriteFile(paths[i], []byte(inputs[i]), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var output bytes.Buffer
	if err := runM05Check(&output, paths); err != nil {
		t.Fatal(err)
	}
	var result M05CheckResult
	if err := contracts.DecodeStrict(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Result != missionValid || result.ExecutionAuthorized || result.Review.Decision != "REQUEST_CHANGES" {
		t.Fatal("invalid output")
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(output.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	for name, schema := range map[string]string{"action": "action-record.schema.json", "proposal": "improvement-proposal.schema.json", "review": "review-record.schema.json"} {
		if err := contracts.ValidateRaw(schema, raw[name]); err != nil {
			t.Fatal(err)
		}
	}
	for name, schema := range map[string]string{"outcomes": "outcome-record.schema.json", "evaluations": "evaluation-record.schema.json"} {
		var items []json.RawMessage
		if err := json.Unmarshal(raw[name], &items); err != nil {
			t.Fatal(err)
		}
		for _, item := range items {
			if err := contracts.ValidateRaw(schema, item); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := runM05Check(m03BrokenWriter{}, paths); err == nil {
		t.Fatal("writer error swallowed")
	}
	for _, args := range [][]string{nil, paths[:4], append(append([]string{}, paths...), "extra"), {"missing", paths[1], paths[2], paths[3], paths[4]}} {
		output.Reset()
		if err := runM05Check(&output, args); err == nil || output.Len() != 0 {
			t.Fatal("invalid invocation emitted output")
		}
	}
	for _, bad := range []string{`{}`, strings.Replace(inputs[4], `"syn-p"`, `"orphan"`, 1)} {
		if err := os.WriteFile(paths[4], []byte(bad), 0600); err != nil {
			t.Fatal(err)
		}
		output.Reset()
		if err := runM05Check(&output, paths); err == nil || output.Len() != 0 {
			t.Fatal("bad chain emitted output")
		}
		got, err := os.ReadFile(paths[4])
		if err != nil || string(got) != bad {
			t.Fatal("review overwritten")
		}
	}
	for i := 0; i < 4; i++ {
		got, err := os.ReadFile(paths[i])
		if err != nil || string(got) != inputs[i] {
			t.Fatal("input overwritten")
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 5 {
		t.Fatal("unexpected persisted state")
	}
}

func TestM05RawEvalPack(t *testing.T) {
	path := filepath.Join(filepath.Dir(evalPath(t, "M05-reviewed-improvement")), "raw-expectations.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var overrides map[string]string
	if err := json.Unmarshal(raw, &overrides); err != nil {
		t.Fatal(err)
	}
	type testCase struct {
		ID          string          `json:"case_id"`
		Mode        string          `json:"mode"`
		Expected    string          `json:"expected"`
		Proposal    json.RawMessage `json:"proposal"`
		Evaluation  json.RawMessage `json:"evaluation"`
		Evaluations json.RawMessage `json:"evaluations"`
		Action      json.RawMessage `json:"action"`
		Outcomes    json.RawMessage `json:"outcomes"`
		Review      json.RawMessage `json:"review"`
	}
	for _, c := range loadCases[testCase](t, "M05-reviewed-improvement") {
		t.Run(c.ID, func(t *testing.T) {
			state := ""
			switch c.Mode {
			case "proposal":
				p, s := DecodeM05Proposal(c.Proposal)
				state = s
				if s == missionValid {
					state = EvaluateImprovementProposal(p)
				}
			case "evaluation":
				e, s := DecodeM05Evaluation(c.Evaluation)
				state = s
				if s != missionValid {
					break
				}
				a, s := DecodeM03Action(c.Action)
				state = s
				if s != missionValid {
					break
				}
				out, s := decodeM05Collection(c.Outcomes, DecodeM03Outcome)
				state = s
				if s == missionValid {
					state = ValidateEvaluationRecord(e, a, out)
				}
			case "proposal-link":
				p, s := DecodeM05Proposal(c.Proposal)
				state = s
				if s != missionValid {
					break
				}
				ev, s := decodeM05Collection(c.Evaluations, DecodeM05Evaluation)
				state = s
				if s == missionValid {
					state = ValidateProposalEvaluationLink(p, ev)
				}
			case "review":
				p, s := DecodeM05Proposal(c.Proposal)
				state = s
				if s != missionValid {
					break
				}
				r, s := DecodeM05Review(c.Review)
				state = s
				if s == missionValid {
					state = ValidateReviewRecord(r, p)
				}
			default:
				t.Fatal("unknown mode")
			}
			want := c.Expected
			if override, ok := overrides[c.ID]; ok {
				want = override
				delete(overrides, c.ID)
			}
			if state != want {
				t.Fatalf("want %s got %s", want, state)
			}
		})
	}
	if len(overrides) != 0 {
		t.Fatal("orphan raw expectations", overrides)
	}
}

func TestM05CommittedFixture(t *testing.T) {
	var output bytes.Buffer
	if err := runM05Check(&output, []string{"../../testdata/m03-action.json", "../../testdata/m05-outcomes.json", "../../testdata/m05-evaluations.json", "../../testdata/m05-proposal.json", "../../testdata/m05-review.json"}); err != nil {
		t.Fatal(err)
	}
	var result M05CheckResult
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Result != missionValid || result.ExecutionAuthorized || result.Evaluations[0].Result != "INCONCLUSIVE" || result.Review.Decision != "REQUEST_CHANGES" {
		t.Fatal("fixture authority/conclusion changed")
	}
}
