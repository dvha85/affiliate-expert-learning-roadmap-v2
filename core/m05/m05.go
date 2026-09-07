package m05

import (
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
	"strings"
	"time"
)

const (
	missionValid   = "VALID"
	missionInvalid = "INVALID"
)

type HumanActionRecord = m03.HumanActionRecord
type EffectRef = m03.EffectRef
type OutcomeRecord = m03.OutcomeRecord

func ValidateEffectRef(r EffectRef) string                 { return m03.ValidateEffectRef(r) }
func ValidateHumanActionRecord(a HumanActionRecord) string { return m03.ValidateHumanActionRecord(a) }
func ValidateActionOutcomeLink(a HumanActionRecord, o OutcomeRecord) string {
	return m03.ValidateActionOutcomeLink(a, o)
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
