package main

import (
	"encoding/json"
	"fmt"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
	"io"
	"os"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

func DecodeM03Action(raw []byte) (HumanActionRecord, string) { return m03.DecodeM03Action(raw) }
func DecodeM03Outcome(raw []byte) (OutcomeRecord, string)    { return m03.DecodeM03Outcome(raw) }

type M03CheckResult = m03.M03CheckResult

func CheckM03Pair(a, o []byte) (M03CheckResult, string) { return m03.CheckM03Pair(a, o) }

// runM03Check is read-only and emits no success envelope on any rejected pair.
func runM03Check(w io.Writer, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: demo m03-check ACTION.json OUTCOME.json")
	}
	a, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	o, err := os.ReadFile(args[1])
	if err != nil {
		return err
	}
	result, state := CheckM03Pair(a, o)
	if state != missionValid {
		return fmt.Errorf("M03 pair: %s", state)
	}
	// Assert the actual serialized artifacts too, not just the input objects.
	for _, item := range []struct {
		schema string
		value  any
	}{
		{"action-record.schema.json", result.Action}, {"outcome-record.schema.json", result.Outcome},
	} {
		raw, err := json.Marshal(item.value)
		if err != nil {
			return err
		}
		if err := contracts.ValidateRaw(item.schema, raw); err != nil {
			return fmt.Errorf("M03 serialized output: %w", err)
		}
	}
	return json.NewEncoder(w).Encode(result)
}
