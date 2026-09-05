package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

// DecodeM03Action checks schema on original bytes, then decodes exact field names.
// VALID here means structurally valid, not reviewed or linked to a stored decision.
func DecodeM03Action(raw []byte) (HumanActionRecord, string) {
	var action HumanActionRecord
	if contracts.ValidateRaw("action-record.schema.json", raw) != nil || contracts.DecodeStrict(raw, &action) != nil {
		return HumanActionRecord{}, "INVALID_SCHEMA"
	}
	return action, missionValid
}

// The M11 internal ActionID alias is never accepted at this canonical boundary.
func DecodeM03Outcome(raw []byte) (OutcomeRecord, string) {
	var outcome OutcomeRecord
	if contracts.ValidateRaw("outcome-record.schema.json", raw) != nil || contracts.DecodeStrict(raw, &outcome) != nil {
		return OutcomeRecord{}, "INVALID_SCHEMA"
	}
	return outcome, missionValid
}

type M03CheckResult struct {
	Action  HumanActionRecord `json:"action"`
	Outcome OutcomeRecord     `json:"outcome"`
	Result  string            `json:"result"`
}

// CheckM03Pair preserves semantic checks after both raw schema checks. It neither
// looks up decision_id in a store nor verifies a report's truth or completeness.
func CheckM03Pair(actionRaw, outcomeRaw []byte) (M03CheckResult, string) {
	a, state := DecodeM03Action(actionRaw)
	if state != missionValid {
		return M03CheckResult{}, state
	}
	o, state := DecodeM03Outcome(outcomeRaw)
	if state != missionValid {
		return M03CheckResult{}, state
	}
	if state = ValidateHumanActionRecord(a); state != missionValid {
		return M03CheckResult{}, state
	}
	if state = ValidateActionOutcomeLink(a, o); state != missionValid {
		return M03CheckResult{}, state
	}
	return M03CheckResult{Action: a, Outcome: o, Result: missionValid}, missionValid
}

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
