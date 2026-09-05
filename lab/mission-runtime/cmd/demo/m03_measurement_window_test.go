package main

import "testing"

func TestM03MeasurementWindow(t *testing.T) {
	a := HumanActionRecord{ActionID: "a-window", DecisionID: "d-window", ActionType: "PUBLISH_REVIEW", Target: "https://example.com/review", PerformedBy: "human", PerformedAt: "2026-09-03T10:00:00+07:00", MeasurementWindowEnd: "2026-09-10T10:00:00+07:00", ComplianceReviewed: true}
	cases := []struct{ name, observed, status, want string }{
		{"early_zero_regression", "2026-09-03T11:00:00+07:00", "NO_OBSERVED_OUTCOME", "MEASUREMENT_WINDOW_OPEN"},
		{"just_before_end", "2026-09-10T09:59:59.999999999+07:00", "NO_OBSERVED_OUTCOME", "MEASUREMENT_WINDOW_OPEN"},
		{"at_end_zero", "2026-09-10T10:00:00+07:00", "NO_OBSERVED_OUTCOME", missionValid},
		{"after_end_zero", "2026-09-11T10:00:00+07:00", "NO_OBSERVED_OUTCOME", missionValid},
		{"same_instant_utc", "2026-09-10T03:00:00Z", "NO_OBSERVED_OUTCOME", missionValid},
		{"same_instant_negative_offset", "2026-09-09T23:00:00-04:00", "NO_OBSERVED_OUTCOME", missionValid},
		{"early_utc", "2026-09-10T02:59:59Z", "NO_OBSERVED_OUTCOME", "MEASUREMENT_WINDOW_OPEN"},
		{"before_action", "2026-09-03T02:59:59Z", "NO_OBSERVED_OUTCOME", "OUTCOME_BEFORE_ACTION"},
		{"pending_before_action", "2026-09-03T02:59:59Z", "PENDING", "OUTCOME_BEFORE_ACTION"},
		{"pending_at_action", "2026-09-03T03:00:00Z", "PENDING", missionValid},
		{"pending_before_end", "2026-09-03T11:00:00+07:00", "PENDING", missionValid},
		{"pending_after_end", "2026-09-11T10:00:00+07:00", "PENDING", missionValid},
		{"observed_order_before_end", "2026-09-03T11:00:00+07:00", "VALID", missionValid},
		{"cancelled_before_end", "2026-09-03T11:00:00+07:00", "CANCELLED", missionValid},
		{"refunded_before_end", "2026-09-03T11:00:00+07:00", "REFUNDED", missionValid},
		{"paid_before_end", "2026-09-03T11:00:00+07:00", "PAID", missionValid},
		{"late_paid_report", "2026-09-12T10:00:00+07:00", "PAID", missionValid},
		{"invalid_observation_time", "invalid", "PENDING", missionInvalid},
		{"timezone_required", "2026-09-10T10:00:00", "NO_OBSERVED_OUTCOME", missionInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o := OutcomeRecord{OutcomeID: "o-window", EffectRef: EffectRef{EffectKind: "HUMAN_ACTION", EffectID: a.ActionID}, ObservedAt: tc.observed, Status: tc.status, Metrics: map[string]float64{"clicks": 0, "valid_orders": 0}, SourceRef: "synthetic-report"}
			if tc.status == "VALID" || tc.status == "PAID" {
				o.Metrics["valid_orders"] = 1
			}
			if got := ValidateActionOutcomeLink(a, o); got != tc.want {
				t.Fatalf("want %s, got %s", tc.want, got)
			}
		})
	}
}

func TestM05RejectsPrematureNoOutcome(t *testing.T) {
	a := HumanActionRecord{ActionID: "a-window", DecisionID: "d-window", ActionType: "PUBLISH_REVIEW", Target: "https://example.com/review", PerformedBy: "human", PerformedAt: "2026-09-03T10:00:00+07:00", MeasurementWindowEnd: "2026-09-10T10:00:00+07:00", ComplianceReviewed: true}
	o := OutcomeRecord{OutcomeID: "o-window", EffectRef: EffectRef{EffectKind: "HUMAN_ACTION", EffectID: a.ActionID}, ObservedAt: "2026-09-03T11:00:00+07:00", Status: "NO_OBSERVED_OUTCOME", Metrics: map[string]float64{"valid_orders": 0}, SourceRef: "synthetic-report"}
	e := EvaluationRecord{EvaluationID: "e-window", DecisionID: a.DecisionID, EffectRef: o.EffectRef, OutcomeIDs: []string{o.OutcomeID}, EvaluatedAt: "2026-09-11T10:00:00+07:00", Result: "NOT_SUPPORTED"}
	if got := ValidateEvaluationRecord(e, a, []OutcomeRecord{o}); got != "BROKEN_LINK" {
		t.Fatalf("premature absence must not enter evaluation: got %s", got)
	}
	o.ObservedAt = a.MeasurementWindowEnd
	if got := ValidateEvaluationRecord(e, a, []OutcomeRecord{o}); got != missionValid {
		t.Fatalf("measured zero at window end must remain usable: got %s", got)
	}
}
