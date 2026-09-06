package main

import (
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
	"sort"
	"strings"
	"time"
)

const (
	missionValid   = "VALID"
	missionInvalid = "INVALID"
)

type SyntheticArtifact struct {
	Kind      string `json:"kind"`
	ID        string `json:"id"`
	ParentID  string `json:"parent_id,omitempty"`
	State     string `json:"state"`
	Synthetic bool   `json:"synthetic"`
}
type SyntheticWalkthroughReport struct {
	Artifacts           []SyntheticArtifact `json:"artifacts"`
	ExternalSideEffects bool                `json:"external_side_effects"`
	FinalState          string              `json:"final_state"`
}

func RunSyntheticWalkthrough() SyntheticWalkthroughReport {
	artifacts := []SyntheticArtifact{
		{Kind: "Observation", ID: "obs-s1", State: "OBSERVED", Synthetic: true},
		{Kind: "DecisionPacket", ID: "dec-s1", ParentID: "obs-s1", State: "GET_MORE_DATA", Synthetic: true},
		{Kind: "AdvisorOutput", ID: "adv-s1", ParentID: "dec-s1", State: "ABSTAIN", Synthetic: true},
		{Kind: "ActionIntent", ID: "intent-s1", ParentID: "dec-s1", State: "PROPOSAL_ONLY", Synthetic: true},
		{Kind: "PolicyDecision", ID: "policy-s1", ParentID: "intent-s1", State: "HUMAN_REVIEW", Synthetic: true},
		{Kind: "ApprovalReview", ID: "review-s1", ParentID: "policy-s1", State: "NOT_AUTHORIZED_FOR_LIVE_EXECUTION", Synthetic: true},
		{Kind: "ExecutionRecord", ID: "exec-s1", ParentID: "intent-s1", State: "DRY_RUN_ONLY", Synthetic: true},
		{Kind: "OutcomeRecord", ID: "out-s1", ParentID: "exec-s1", State: "SYNTHETIC_OUTCOME", Synthetic: true},
		{Kind: "EvaluationRecord", ID: "eval-s1", ParentID: "out-s1", State: "INCONCLUSIVE", Synthetic: true},
	}
	return SyntheticWalkthroughReport{Artifacts: artifacts, ExternalSideEffects: false, FinalState: "DRY_RUN_ONLY"}
}

func ValidateSyntheticWalkthrough(r SyntheticWalkthroughReport) string {
	if r.ExternalSideEffects || r.FinalState != "DRY_RUN_ONLY" || len(r.Artifacts) < 9 {
		return missionInvalid
	}
	byID := map[string]SyntheticArtifact{}
	for _, a := range r.Artifacts {
		if a.ID == "" || !a.Synthetic {
			return missionInvalid
		}
		if a.ParentID != "" {
			if _, ok := byID[a.ParentID]; !ok {
				return "BROKEN_LINK"
			}
		}
		byID[a.ID] = a
	}
	if byID["exec-s1"].State != "DRY_RUN_ONLY" {
		return missionInvalid
	}
	return missionValid
}

type HumanActionRecord = m03.HumanActionRecord
type EffectRef = m03.EffectRef
type OutcomeRecord = m03.OutcomeRecord

func ValidateEffectRef(ref EffectRef) string { return m03.ValidateEffectRef(ref) }
func ValidateHumanActionRecord(r HumanActionRecord) string { return m03.ValidateHumanActionRecord(r) }
func ValidateOutcomeRecord(r OutcomeRecord) string { return m03.ValidateOutcomeRecord(r) }
func ValidateActionOutcomeLink(a HumanActionRecord, o OutcomeRecord) string { return m03.ValidateActionOutcomeLink(a, o) }

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
	if state := validateAdvisorFields(o); state != missionValid {
		return state
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
		if strings.TrimSpace(x.EvidenceID) == "" || strings.TrimSpace(x.SourceRef) == "" {
			return missionInvalid
		}
		if _, exists := idx[x.EvidenceID]; exists {
			return missionInvalid
		}
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
		if at.After(now) {
			return "ABSTAIN_FUTURE"
		}
		if maxAgeHours >= 0 && now.Sub(at) > time.Duration(maxAgeHours)*time.Hour {
			return "ABSTAIN_STALE"
		}
	}
	return "SUPPORTED"
}

type EvaluationRecord struct {
	EvaluationID string    `json:"evaluation_id"`
	DecisionID   string    `json:"decision_id"`
	EffectRef    EffectRef `json:"effect_ref"`
	OutcomeIDs   []string  `json:"outcome_ids"`
	EvaluatedAt  string    `json:"evaluated_at"`
	Result       string    `json:"result"`
	EvidenceIDs  []string  `json:"evidence_ids"`
	Limitations  []string  `json:"limitations"`
	Notes        *string   `json:"notes,omitempty"`
	// Internal alias used by pre-cleanup M11 reference code; never serialized as canonical JSON.
	ActionID string `json:"-"`
}
type ImprovementProposal struct {
	ProposalID      string   `json:"proposal_id"`
	EvaluationIDs   []string `json:"evaluation_ids"`
	CurrentVersion  string   `json:"current_version"`
	ProposedVersion string   `json:"proposed_version"`
	ChangeSummary   string   `json:"change_summary"`
	ExpectedBenefit string   `json:"expected_benefit"`
	Risks           []string `json:"risks,omitempty"`
	Rollback        string   `json:"rollback"`
	AutoApply       bool     `json:"auto_apply"`
}
type ReviewRecord struct {
	ReviewID   string `json:"review_id"`
	ProposalID string `json:"proposal_id"`
	ReviewedBy string `json:"reviewed_by"`
	ReviewedAt string `json:"reviewed_at"`
	Decision   string `json:"decision"`
	Reason     string `json:"reason"`
}

func ValidateEvaluationRecord(e EvaluationRecord, a HumanActionRecord, outcomes []OutcomeRecord) string {
	if e.EvaluationID == "" || e.DecisionID == "" || ValidateEffectRef(e.EffectRef) != missionValid || len(e.OutcomeIDs) == 0 {
		return missionInvalid
	}
	if e.EffectRef.EffectKind != "HUMAN_ACTION" || e.EffectRef.EffectID != a.ActionID || e.DecisionID != a.DecisionID {
		return "BROKEN_LINK"
	}
	if _, err := time.Parse(time.RFC3339, e.EvaluatedAt); err != nil {
		return missionInvalid
	}
	validResult := map[string]bool{"SUPPORTED": true, "NOT_SUPPORTED": true, "INCONCLUSIVE": true, "NEEDS_MORE_DATA": true}
	if !validResult[e.Result] {
		return missionInvalid
	}
	idx := map[string]bool{}
	for _, o := range outcomes {
		if ValidateActionOutcomeLink(a, o) == missionValid {
			idx[o.OutcomeID] = true
		}
	}
	for _, id := range e.OutcomeIDs {
		if !idx[id] {
			return "BROKEN_LINK"
		}
	}
	return missionValid
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

func ValidateProposalEvaluationLink(p ImprovementProposal, evaluations []EvaluationRecord) string {
	if EvaluateImprovementProposal(p) != "REVIEW_REQUIRED" {
		return missionInvalid
	}
	idx := map[string]bool{}
	for _, e := range evaluations {
		idx[e.EvaluationID] = true
	}
	for _, id := range p.EvaluationIDs {
		if !idx[id] {
			return "BROKEN_LINK"
		}
	}
	return missionValid
}

func ValidateReviewRecord(r ReviewRecord, p ImprovementProposal) string {
	if r.ReviewID == "" || r.ProposalID != p.ProposalID || r.ReviewedBy != "human" || r.Reason == "" {
		return "BROKEN_LINK"
	}
	if _, err := time.Parse(time.RFC3339, r.ReviewedAt); err != nil {
		return missionInvalid
	}
	ok := map[string]bool{"APPROVE_FOR_MANUAL_CHANGE": true, "REJECT": true, "REQUEST_CHANGES": true}
	if !ok[r.Decision] {
		return missionInvalid
	}
	return missionValid
}

func sortedStrings(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}
