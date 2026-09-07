package m05_test

import (
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m05"
	"strings"
	"testing"
)

func TestCanonicalChain(t *testing.T) {
	// Literal expectations independent of the compatibility wrappers in demo.
	base := []string{
		`{"action_id":"a","decision_id":"d","action_type":"synthetic","target":"fixture:none","performed_by":"human","performed_at":"2026-09-01T00:00:00Z","measurement_window_end":"2026-09-02T00:00:00Z","compliance_reviewed":true}`,
		`[{"outcome_id":"o","effect_ref":{"effect_kind":"HUMAN_ACTION","effect_id":"a"},"observed_at":"2026-09-02T00:00:00Z","status":"PENDING","metrics":{},"source_ref":"fixture:test"}]`,
		`[{"evaluation_id":"e","decision_id":"d","effect_ref":{"effect_kind":"HUMAN_ACTION","effect_id":"a"},"outcome_ids":["o"],"evaluated_at":"2026-09-03T00:00:00Z","result":"INCONCLUSIVE","evidence_ids":[],"limitations":["Synthetic pending outcome"]}]`,
		`{"proposal_id":"p","evaluation_ids":["e"],"current_version":"v1","proposed_version":"v2","change_summary":"clarify pending","expected_benefit":"avoid premature conclusions","risks":["unmeasured"],"rollback":"restore v1 manually","auto_apply":false}`,
		`{"review_id":"r","proposal_id":"p","reviewed_by":"human","reviewed_at":"2026-09-04T00:00:00Z","decision":"REQUEST_CHANGES","reason":"collect more data"}`,
	}
	for _, tc := range []struct {
		name           string
		index          int
		old, new, want string
	}{
		{"valid", 0, "synthetic", "synthetic", "VALID"},
		{"approval-not-execution", 4, "REQUEST_CHANGES", "APPROVE_FOR_MANUAL_CHANGE", "VALID"},
		{"wrong-decision", 2, `"decision_id":"d"`, `"decision_id":"orphan"`, "BROKEN_LINK"},
		{"orphan-outcome", 2, `["o"]`, `["missing"]`, "BROKEN_LINK"},
		{"early-evaluation", 2, "2026-09-03", "2026-09-01", "EVALUATION_BEFORE_OUTCOME"},
		{"early-review", 4, "2026-09-04", "2026-09-02", "REVIEW_BEFORE_EVALUATION"},
		{"auto-apply", 3, "false", "true", "INVALID_SCHEMA"},
		{"machine-review", 4, `"human"`, `"agent"`, "INVALID_SCHEMA"},
		{"unknown-field", 2, `"evaluation_id":`, `"extra":true,"evaluation_id":`, "INVALID_SCHEMA"},
		{"duplicate-field", 2, `"evaluation_id":`, `"evaluation_id":"other","evaluation_id":`, "INVALID_SCHEMA"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := append([]string(nil), base...)
			in[tc.index] = strings.Replace(in[tc.index], tc.old, tc.new, 1)
			out, status := m05.CheckM05Chain([]byte(in[0]), []byte(in[1]), []byte(in[2]), []byte(in[3]), []byte(in[4]))
			if status != tc.want || out.ExecutionAuthorized {
				t.Fatal(status, out)
			}
			if status != "VALID" && out.Proposal.ProposalID != "" {
				t.Fatal("rejected artifact escaped")
			}
			if status == "VALID" && (out.Evaluations[0].Result != "INCONCLUSIVE" || out.Proposal.AutoApply) {
				t.Fatal(out)
			}
		})
	}
}
