package m08

import "testing"

func TestPolicyConformanceTable(t *testing.T) {
	for _, scenario := range PolicyConformanceCases() {
		t.Run(scenario.Name, func(t *testing.T) {
			got := EvaluatePolicy(scenario.Intent, scenario.Context)
			if got.Decision != scenario.Decision || got.RiskClass != scenario.RiskClass || got.Reason != scenario.Reason || got.PolicyReviewRequired != scenario.PolicyReview || got.ExecutionAuthorized != scenario.ExecutionAuth {
				t.Fatalf("unexpected policy result: got=%+v want decision=%s risk=%s reason=%s review=%v authority=%v", got, scenario.Decision, scenario.RiskClass, scenario.Reason, scenario.PolicyReview, scenario.ExecutionAuth)
			}
			if got.PolicyVersion != scenario.Context.PolicyVersion || got.PolicyMode != "NON_AUTHORIZING" || got.PolicyCheckedAt != scenario.Context.Now {
				t.Fatalf("policy metadata drifted: got=%+v", got)
			}
		})
	}
}
