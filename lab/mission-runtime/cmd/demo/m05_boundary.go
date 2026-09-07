package main

import (
	"encoding/json"
	"fmt"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m05"
	"io"
	"os"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

func DecodeM05Evaluation(raw []byte) (EvaluationRecord, string)  { return m05.DecodeM05Evaluation(raw) }
func DecodeM05Proposal(raw []byte) (ImprovementProposal, string) { return m05.DecodeM05Proposal(raw) }
func DecodeM05Review(raw []byte) (ReviewRecord, string)          { return m05.DecodeM05Review(raw) }
func decodeM05Collection[T any](raw []byte, decode func([]byte) (T, string)) ([]T, string) {
	return m05.DecodeCollection(raw, decode)
}

type M05CheckResult = m05.M05CheckResult

func CheckM05Chain(a, o, e, p, r []byte) (M05CheckResult, string) {
	return m05.CheckM05Chain(a, o, e, p, r)
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
