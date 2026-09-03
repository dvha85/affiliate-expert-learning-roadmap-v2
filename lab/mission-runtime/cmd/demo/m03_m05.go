package main

import (
	"sort"
	"strings"
	"time"
)

const (
	missionValid   = "VALID"
	missionInvalid = "INVALID"
)

type SyntheticWalkthroughReport struct {
	Stages              []string `json:"stages"`
	ExternalSideEffects bool     `json:"external_side_effects"`
	FinalState          string   `json:"final_state"`
}

func RunSyntheticWalkthrough() SyntheticWalkthroughReport {
	return SyntheticWalkthroughReport{[]string{"synthetic_evidence", "deterministic_decision", "grounded_ai_advisor", "shadow_action_intent", "deterministic_policy", "human_review", "dry_run_executor", "synthetic_outcome", "evaluation"}, false, "DRY_RUN_ONLY"}
}

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
type OutcomeRecord struct {
	OutcomeID  string             `json:"outcome_id"`
	ActionID   string             `json:"action_id"`
	ObservedAt string             `json:"observed_at"`
	Status     string             `json:"status"`
	Metrics    map[string]float64 `json:"metrics"`
	SourceRef  string             `json:"source_ref"`
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
	if strings.TrimSpace(r.OutcomeID) == "" || strings.TrimSpace(r.ActionID) == "" || strings.TrimSpace(r.SourceRef) == "" {
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

type AdvisorEvidence struct {
	EvidenceID string `json:"evidence_id"`
	ObservedAt string `json:"observed_at"`
	SourceRef  string `json:"source_ref"`
}
type AdvisorOutput struct {
	State              string   `json:"state"`
	Recommendation     string   `json:"recommendation"`
	Reason             string   `json:"reason"`
	EvidenceIDs        []string `json:"evidence_ids"`
	Unknowns           []string `json:"unknowns"`
	WriteToolRequested bool     `json:"write_tool_requested"`
}

func EvaluateAdvisorOutput(o AdvisorOutput, ev []AdvisorEvidence, asOf string, maxAgeHours int) string {
	if o.WriteToolRequested {
		return "REJECT_WRITE_REQUEST"
	}
	if o.State == "ABSTAIN" {
		return "ABSTAIN"
	}
	if o.State != "ADVISE" && o.State != "HUMAN_REVIEW" {
		return missionInvalid
	}
	if len(o.EvidenceIDs) == 0 {
		return "REJECT_UNGROUNDED"
	}
	now, e := time.Parse(time.RFC3339, asOf)
	if e != nil {
		return missionInvalid
	}
	idx := map[string]AdvisorEvidence{}
	for _, x := range ev {
		idx[x.EvidenceID] = x
	}
	for _, id := range o.EvidenceIDs {
		x, ok := idx[id]
		if !ok {
			return "REJECT_UNGROUNDED"
		}
		at, e := time.Parse(time.RFC3339, x.ObservedAt)
		if e != nil {
			return missionInvalid
		}
		if maxAgeHours >= 0 && now.Sub(at) > time.Duration(maxAgeHours)*time.Hour {
			return "ABSTAIN_STALE"
		}
	}
	return "SUPPORTED"
}

type ImprovementProposal struct {
	ProposalID      string   `json:"proposal_id"`
	EvaluationIDs   []string `json:"evaluation_ids"`
	CurrentVersion  string   `json:"current_version"`
	ProposedVersion string   `json:"proposed_version"`
	ChangeSummary   string   `json:"change_summary"`
	ExpectedBenefit string   `json:"expected_benefit"`
	Risks           []string `json:"risks"`
	Rollback        string   `json:"rollback"`
	AutoApply       bool     `json:"auto_apply"`
}

func EvaluateImprovementProposal(p ImprovementProposal) string {
	if p.AutoApply {
		return "REJECT_AUTO_APPLY"
	}
	if strings.TrimSpace(p.ProposalID) == "" || len(p.EvaluationIDs) == 0 || strings.TrimSpace(p.ChangeSummary) == "" || strings.TrimSpace(p.ExpectedBenefit) == "" || strings.TrimSpace(p.Rollback) == "" {
		return missionInvalid
	}
	if strings.TrimSpace(p.CurrentVersion) == "" || strings.TrimSpace(p.ProposedVersion) == "" || p.CurrentVersion == p.ProposedVersion {
		return missionInvalid
	}
	return "REVIEW_REQUIRED"
}

func sortedStrings(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}
