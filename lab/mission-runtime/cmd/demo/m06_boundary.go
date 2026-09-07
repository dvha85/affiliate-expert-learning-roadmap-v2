package main

import (
	"encoding/json"
	"fmt"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m06"
	"io"
	"os"
)

type M06FileInput = m06.M06FileInput

func DecodeM06Input(raw []byte) (M06FileInput, string) { return m06.DecodeM06Input(raw) }
func ValidateM06Observation(raw []byte) string         { return m06.ValidateM06Observation(raw) }

func runM06Check(w io.Writer, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: demo m06-check RESPONSE-FIXTURE.json")
	}
	raw, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	input, state := DecodeM06Input(raw)
	if state != missionValid {
		return fmt.Errorf("M06 input: %s", state)
	}
	if state = EvaluateWatchRequest(input.Request); state != "NEW" && state != "UNCHANGED" && state != "CHANGED" {
		return fmt.Errorf("M06 request: %s", state)
	}
	if input.StatusCode < 200 || input.StatusCode >= 300 {
		return fmt.Errorf("M06 response: REJECT_RESPONSE_STATUS")
	}
	observation, state := NormalizeWatchObservation(input.Request, input.SubjectID)
	if state != "NEW" && state != "UNCHANGED" && state != "CHANGED" {
		return fmt.Errorf("M06 observation: %s", state)
	}
	output, err := json.Marshal(observation)
	if err != nil {
		return err
	}
	if stateSchema := ValidateM06Observation(output); stateSchema != missionValid {
		return fmt.Errorf("M06 output: %s", stateSchema)
	}
	return json.NewEncoder(w).Encode(struct {
		Observation         CanonicalObservation `json:"observation"`
		Result              string               `json:"result"`
		ExternalSideEffects bool                 `json:"external_side_effects"`
	}{observation, state, false})
}
