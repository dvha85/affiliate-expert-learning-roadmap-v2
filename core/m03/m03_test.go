package m03_test

import (
	"encoding/json"
	"testing"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
)

const action = `{"action_id":"a","decision_id":"d","action_type":"PUBLISH_REVIEW","target":"https://example.com","performed_by":"human","performed_at":"2026-09-03T08:00:00Z","measurement_window_end":"2026-09-04T08:00:00Z","compliance_reviewed":true}`
const outcome = `{"outcome_id":"o","effect_ref":{"effect_kind":"HUMAN_ACTION","effect_id":"a"},"observed_at":"2026-09-04T08:00:00Z","status":"NO_OBSERVED_OUTCOME","metrics":{},"source_ref":"fixture"}`

func TestPairIndependentExpectations(t *testing.T) {
	for _, tc := range []struct {
		name, key string
		value     any
		want      string
	}{
		{"valid", "status", "NO_OBSERVED_OUTCOME", "VALID"},
		{"open", "observed_at", "2026-09-03T08:00:00Z", "MEASUREMENT_WINDOW_OPEN"},
		{"before", "observed_at", "2026-09-03T07:59:59Z", "OUTCOME_BEFORE_ACTION"},
		{"link", "effect_ref", map[string]string{"effect_kind": "HUMAN_ACTION", "effect_id": "other"}, "BROKEN_LINK"},
		{"alias", "action_id", "a", "INVALID_SCHEMA"},
		{"null", "metrics", nil, "INVALID_SCHEMA"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var fields map[string]any
			if e := json.Unmarshal([]byte(outcome), &fields); e != nil {
				t.Fatal(e)
			}
			fields[tc.key] = tc.value
			b, e := json.Marshal(fields)
			if e != nil {
				t.Fatal(e)
			}
			r, s := m03.CheckM03Pair([]byte(action), b)
			if s != tc.want {
				t.Fatal(s, tc.want)
			}
			if s == "VALID" {
				a, _ := json.Marshal(r.Action)
				o, _ := json.Marshal(r.Outcome)
				if _, s = m03.DecodeM03Action(a); s != "VALID" {
					t.Fatal(s)
				}
				if _, s = m03.DecodeM03Outcome(o); s != "VALID" {
					t.Fatal(s)
				}
			}
		})
	}
}

func TestStrictRawAndCompatibility(t *testing.T) {
	for _, b := range []string{`null`, action + ` {}`, `{"action_id":"a","action_id":"b"}`} {
		if _, s := m03.DecodeM03Action([]byte(b)); s != "INVALID_SCHEMA" {
			t.Fatal(s)
		}
	}
	o := m03.OutcomeRecord{OutcomeID: "o", ActionID: "legacy", ObservedAt: "2026-09-03T08:00:00Z", Status: "VALID", SourceRef: "fixture"}
	if m03.ValidateOutcomeRecord(o) != "VALID" {
		t.Fatal("typed compatibility changed")
	}
	b, _ := json.Marshal(o)
	var fields map[string]json.RawMessage
	json.Unmarshal(b, &fields)
	if _, ok := fields["action_id"]; ok {
		t.Fatal("alias leaked")
	}
	if _, s := m03.DecodeM03Outcome(b); s != "INVALID_SCHEMA" {
		t.Fatal("legacy typed record became canonical", s)
	}
}
