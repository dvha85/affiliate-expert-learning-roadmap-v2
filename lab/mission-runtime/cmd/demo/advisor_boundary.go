package main

import (
	"encoding/json"
	"fmt"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m04"
	"io"
	"os"
	"strconv"
)

func DecodeAdvisorOutput(raw []byte) (AdvisorOutput, string) { return m04.DecodeAdvisorOutput(raw) }

func validateAdvisorFields(o AdvisorOutput) string { return m04.ValidateAdvisorFields(o) }

func runAdvisorCheck(w io.Writer, args []string) error {
	if len(args) != 4 {
		return fmt.Errorf("usage: demo advisor-check OUTPUT.json EVIDENCE.json AS_OF MAX_AGE_HOURS")
	}
	raw, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	o, state := DecodeAdvisorOutput(raw)
	if state != missionValid {
		return fmt.Errorf("advisor output: %s", state)
	}
	evidenceRaw, err := os.ReadFile(args[1])
	if err != nil {
		return err
	}
	var evidence []AdvisorEvidence
	if err := json.Unmarshal(evidenceRaw, &evidence); err != nil {
		return fmt.Errorf("invalid evidence JSON: %w", err)
	}
	maxAge, err := strconv.Atoi(args[3])
	if err != nil || maxAge < 0 {
		return fmt.Errorf("MAX_AGE_HOURS must be a non-negative integer")
	}
	state = EvaluateAdvisorOutput(o, evidence, args[2], maxAge)
	return json.NewEncoder(w).Encode(struct {
		Output AdvisorOutput `json:"advisor_output"`
		Result string        `json:"result"`
	}{o, state})
}
