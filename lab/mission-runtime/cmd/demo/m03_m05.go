package main

import (
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m04"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m05"
	"sort"
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

func ValidateEffectRef(ref EffectRef) string               { return m03.ValidateEffectRef(ref) }
func ValidateHumanActionRecord(r HumanActionRecord) string { return m03.ValidateHumanActionRecord(r) }
func ValidateOutcomeRecord(r OutcomeRecord) string         { return m03.ValidateOutcomeRecord(r) }
func ValidateActionOutcomeLink(a HumanActionRecord, o OutcomeRecord) string {
	return m03.ValidateActionOutcomeLink(a, o)
}

type AdvisorEvidence = m04.AdvisorEvidence
type AdvisorOutput = m04.AdvisorOutput

func EvaluateAdvisorOutput(o AdvisorOutput, ev []AdvisorEvidence, asOf string, maxAgeHours int) string {
	return m04.EvaluateAdvisorOutput(o, ev, asOf, maxAgeHours)
}

type EvaluationRecord = m05.EvaluationRecord
type ImprovementProposal = m05.ImprovementProposal
type ReviewRecord = m05.ReviewRecord

func ValidateEvaluationRecord(e EvaluationRecord, a HumanActionRecord, outcomes []OutcomeRecord) string {
	return m05.ValidateEvaluationRecord(e, a, outcomes)
}
func EvaluateImprovementProposal(p ImprovementProposal) string {
	return m05.EvaluateImprovementProposal(p)
}
func ValidateProposalEvaluationLink(p ImprovementProposal, e []EvaluationRecord) string {
	return m05.ValidateProposalEvaluationLink(p, e)
}
func ValidateReviewRecord(r ReviewRecord, p ImprovementProposal) string {
	return m05.ValidateReviewRecord(r, p)
}

func sortedStrings(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}
