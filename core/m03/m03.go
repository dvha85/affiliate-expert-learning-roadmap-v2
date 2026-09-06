package m03

import (
	"strings"
	"time"
)

const (
	missionValid   = "VALID"
	missionInvalid = "INVALID"
)

type HumanActionRecord struct {
	ActionID             string `json:"action_id"`
	DecisionID           string `json:"decision_id"`
	ActionType           string `json:"action_type"`
	Target               string `json:"target"`
	PerformedBy          string `json:"performed_by"`
	PerformedAt          string `json:"performed_at"`
	MeasurementWindowEnd string `json:"measurement_window_end"`
	ComplianceReviewed   bool   `json:"compliance_reviewed"`
}

type EffectRef struct {
	EffectKind string `json:"effect_kind"`
	EffectID   string `json:"effect_id"`
}

func ValidateEffectRef(ref EffectRef) string {
	if strings.TrimSpace(ref.EffectID) == "" {
		return missionInvalid
	}
	if ref.EffectKind != "HUMAN_ACTION" && ref.EffectKind != "MACHINE_EXECUTION" {
		return missionInvalid
	}
	return missionValid
}

type OutcomeRecord struct {
	OutcomeID  string             `json:"outcome_id"`
	EffectRef  EffectRef          `json:"effect_ref"`
	ObservedAt string             `json:"observed_at"`
	Status     string             `json:"status"`
	Metrics    map[string]float64 `json:"metrics"`
	SourceRef  string             `json:"source_ref"`
	// Internal alias used by pre-cleanup M11 reference code; never serialized as canonical JSON.
	ActionID string `json:"-"`
}

func ValidateHumanActionRecord(r HumanActionRecord) string {
	if strings.TrimSpace(r.ActionID) == "" || strings.TrimSpace(r.DecisionID) == "" || strings.TrimSpace(r.ActionType) == "" || strings.TrimSpace(r.Target) == "" {
		return missionInvalid
	}
	if r.PerformedBy != "human" {
		return "REJECT_MACHINE_EXECUTION"
	}
	a, e := time.Parse(time.RFC3339, r.PerformedAt)
	if e != nil {
		return missionInvalid
	}
	w, e := time.Parse(time.RFC3339, r.MeasurementWindowEnd)
	if e != nil || w.Before(a) {
		return missionInvalid
	}
	if !r.ComplianceReviewed {
		return "HUMAN_REVIEW"
	}
	return missionValid
}

func ValidateOutcomeRecord(r OutcomeRecord) string {
	if strings.TrimSpace(r.OutcomeID) == "" || strings.TrimSpace(r.SourceRef) == "" {
		return missionInvalid
	}
	if ValidateEffectRef(r.EffectRef) != missionValid && strings.TrimSpace(r.ActionID) == "" {
		return missionInvalid
	}
	if _, e := time.Parse(time.RFC3339, r.ObservedAt); e != nil {
		return missionInvalid
	}
	ok := map[string]bool{"PENDING": true, "VALID": true, "CANCELLED": true, "REFUNDED": true, "PAID": true, "NO_OBSERVED_OUTCOME": true}
	if !ok[r.Status] {
		return missionInvalid
	}
	for _, v := range r.Metrics {
		if v < 0 {
			return missionInvalid
		}
	}
	return missionValid
}

func ValidateActionOutcomeLink(a HumanActionRecord, o OutcomeRecord) string {
	if ValidateHumanActionRecord(a) != missionValid || ValidateOutcomeRecord(o) != missionValid {
		return missionInvalid
	}
	if o.EffectRef.EffectKind != "HUMAN_ACTION" || o.EffectRef.EffectID != a.ActionID {
		return "BROKEN_LINK"
	}
	performed, _ := time.Parse(time.RFC3339, a.PerformedAt)
	observed, _ := time.Parse(time.RFC3339, o.ObservedAt)
	if observed.Before(performed) {
		return "OUTCOME_BEFORE_ACTION"
	}
	// Absence is a window-level conclusion; interim transaction observations are not.
	windowEnd, _ := time.Parse(time.RFC3339, a.MeasurementWindowEnd)
	if o.Status == "NO_OBSERVED_OUTCOME" && observed.Before(windowEnd) {
		return "MEASUREMENT_WINDOW_OPEN"
	}
	return missionValid
}
