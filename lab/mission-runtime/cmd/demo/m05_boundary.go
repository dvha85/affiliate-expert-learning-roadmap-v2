package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

func DecodeM05Evaluation(raw []byte) (EvaluationRecord, string) {
	var v EvaluationRecord
	if contracts.ValidateRaw("evaluation-record.schema.json", raw) != nil || contracts.DecodeStrict(raw, &v) != nil {
		return EvaluationRecord{}, "INVALID_SCHEMA"
	}
	return v, missionValid
}
func DecodeM05Proposal(raw []byte) (ImprovementProposal, string) {
	var v ImprovementProposal
	if contracts.ValidateRaw("improvement-proposal.schema.json", raw) != nil || contracts.DecodeStrict(raw, &v) != nil {
		return ImprovementProposal{}, "INVALID_SCHEMA"
	}
	return v, missionValid
}
func DecodeM05Review(raw []byte) (ReviewRecord, string) {
	var v ReviewRecord
	if contracts.ValidateRaw("review-record.schema.json", raw) != nil || contracts.DecodeStrict(raw, &v) != nil {
		return ReviewRecord{}, "INVALID_SCHEMA"
	}
	return v, missionValid
}

// File collections are nonempty arrays, not canonical artifact envelopes.
func decodeM05Collection[T any](raw []byte, decode func([]byte) (T, string)) ([]T, string) {
	if _, err := contracts.Decode(raw); err != nil {
		return nil, "INVALID_SCHEMA"
	}
	var items []json.RawMessage
	if json.Unmarshal(raw, &items) != nil || len(items) == 0 {
		return nil, "INVALID_SCHEMA"
	}
	values := make([]T, 0, len(items))
	for _, item := range items {
		value, state := decode(item)
		if state != missionValid {
			return nil, state
		}
		values = append(values, value)
	}
	return values, missionValid
}

type M05CheckResult struct {
	Action              HumanActionRecord   `json:"action"`
	Outcomes            []OutcomeRecord     `json:"outcomes"`
	Evaluations         []EvaluationRecord  `json:"evaluations"`
	Proposal            ImprovementProposal `json:"proposal"`
	Review              ReviewRecord        `json:"review"`
	Result              string              `json:"result"`
	ExecutionAuthorized bool                `json:"execution_authorized"`
}

// CheckM05Chain validates all supplied artifacts, even unreferenced ones. It is
// read-only; VALID includes REJECT/REQUEST_CHANGES reviews and grants no authority.
func CheckM05Chain(actionRaw, outcomesRaw, evaluationsRaw, proposalRaw, reviewRaw []byte) (M05CheckResult, string) {
	a, state := DecodeM03Action(actionRaw)
	if state != missionValid {
		return M05CheckResult{}, state
	}
	outcomes, state := decodeM05Collection(outcomesRaw, DecodeM03Outcome)
	if state != missionValid {
		return M05CheckResult{}, state
	}
	evaluations, state := decodeM05Collection(evaluationsRaw, DecodeM05Evaluation)
	if state != missionValid {
		return M05CheckResult{}, state
	}
	p, state := DecodeM05Proposal(proposalRaw)
	if state != missionValid {
		return M05CheckResult{}, state
	}
	r, state := DecodeM05Review(reviewRaw)
	if state != missionValid {
		return M05CheckResult{}, state
	}
	// Do not run semantic checks before every input has crossed its schema boundary.
	if state = ValidateHumanActionRecord(a); state != missionValid {
		return M05CheckResult{}, state
	}
	outcomeTimes := map[string]time.Time{}
	for _, o := range outcomes {
		if _, exists := outcomeTimes[o.OutcomeID]; exists {
			return M05CheckResult{}, "DUPLICATE_ID"
		}
		if state = ValidateActionOutcomeLink(a, o); state != missionValid {
			return M05CheckResult{}, state
		}
		outcomeTimes[o.OutcomeID], _ = time.Parse(time.RFC3339, o.ObservedAt)
	}
	evaluationTimes := map[string]time.Time{}
	for _, e := range evaluations {
		if strings.TrimSpace(e.EvaluationID) == "" {
			return M05CheckResult{}, missionInvalid
		}
		if _, exists := evaluationTimes[e.EvaluationID]; exists {
			return M05CheckResult{}, "DUPLICATE_ID"
		}
		if state = ValidateEvaluationRecord(e, a, outcomes); state != missionValid {
			return M05CheckResult{}, state
		}
		at, _ := time.Parse(time.RFC3339, e.EvaluatedAt)
		for _, id := range e.OutcomeIDs {
			if at.Before(outcomeTimes[id]) {
				return M05CheckResult{}, "EVALUATION_BEFORE_OUTCOME"
			}
		}
		evaluationTimes[e.EvaluationID] = at
	}
	if state = EvaluateImprovementProposal(p); state != "REVIEW_REQUIRED" {
		return M05CheckResult{}, state
	}
	if state = ValidateProposalEvaluationLink(p, evaluations); state != missionValid {
		return M05CheckResult{}, state
	}
	seen := map[string]bool{}
	for _, id := range p.EvaluationIDs {
		if seen[id] {
			return M05CheckResult{}, "DUPLICATE_ID"
		}
		seen[id] = true
	}
	if strings.TrimSpace(r.ReviewID) == "" || strings.TrimSpace(r.Reason) == "" {
		return M05CheckResult{}, missionInvalid
	}
	if state = ValidateReviewRecord(r, p); state != missionValid {
		return M05CheckResult{}, state
	}
	reviewed, _ := time.Parse(time.RFC3339, r.ReviewedAt)
	for _, id := range p.EvaluationIDs {
		if reviewed.Before(evaluationTimes[id]) {
			return M05CheckResult{}, "REVIEW_BEFORE_EVALUATION"
		}
	}
	return M05CheckResult{Action: a, Outcomes: outcomes, Evaluations: evaluations, Proposal: p, Review: r, Result: missionValid, ExecutionAuthorized: false}, missionValid
}

func runM05Check(w io.Writer, args []string) error {
	if len(args) != 5 {
		return fmt.Errorf("usage: demo m05-check ACTION.json OUTCOMES.json EVALUATIONS.json PROPOSAL.json REVIEW.json")
	}
	inputs := make([][]byte, 5)
	for i, path := range args {
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		inputs[i] = raw
	}
	result, state := CheckM05Chain(inputs[0], inputs[1], inputs[2], inputs[3], inputs[4])
	if state != missionValid {
		return fmt.Errorf("M05 chain: %s", state)
	}
	artifacts := []struct {
		schema string
		value  any
	}{
		{"action-record.schema.json", result.Action},
		{"improvement-proposal.schema.json", result.Proposal},
		{"review-record.schema.json", result.Review},
	}
	for _, o := range result.Outcomes {
		artifacts = append(artifacts, struct {
			schema string
			value  any
		}{"outcome-record.schema.json", o})
	}
	for _, e := range result.Evaluations {
		artifacts = append(artifacts, struct {
			schema string
			value  any
		}{"evaluation-record.schema.json", e})
	}
	for _, item := range artifacts {
		raw, err := json.Marshal(item.value)
		if err != nil {
			return err
		}
		if err := contracts.ValidateRaw(item.schema, raw); err != nil {
			return fmt.Errorf("M05 serialized output: %w", err)
		}
	}
	return json.NewEncoder(w).Encode(result)
}
