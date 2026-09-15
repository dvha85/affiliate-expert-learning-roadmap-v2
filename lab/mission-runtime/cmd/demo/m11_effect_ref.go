package main

import (
	"encoding/json"

	corem03 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
	corem05 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m05"
	corem08 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m08"
	corem11 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m11"
)

func missionCoreValue(in any, out any) bool {
	raw, err := json.Marshal(in)
	return err == nil && json.Unmarshal(raw, out) == nil
}

// ValidateCanonicalProductionClosedCycle enforces the stage-neutral EffectRef
// contract before delegating to the existing closed-cycle linkage validator.
func ValidateCanonicalProductionClosedCycle(cycle ProductionCycleRecord, state M11State, gate ProductionGateDecision, auth ProductionExecutionAuthorization, exec ProductionExecutionRecord, outcome OutcomeRecord, evaluation EvaluationRecord, proposal *ImprovementProposal, review *ReviewRecord) string {
	if outcome.EffectRef.EffectKind != "MACHINE_EXECUTION" || outcome.EffectRef.EffectID != exec.ExecutionID {
		return "BROKEN_LINK"
	}
	if evaluation.EffectRef != outcome.EffectRef {
		return "BROKEN_LINK"
	}
	// Populate internal compatibility aliases only after canonical linkage passes.
	outcome.ActionID = outcome.EffectRef.EffectID
	evaluation.ActionID = evaluation.EffectRef.EffectID
	var ci corem08.Intent
	var cl corem11.ProductionLease
	var cg corem11.ProductionGateDecision
	var ca corem11.ProductionExecutionAuthorization
	var ce corem11.ProductionExecutionRecord
	var ccycle corem11.ProductionCycleRecord
	var co corem03.OutcomeRecord
	var cev corem05.EvaluationRecord
	if !missionCoreValue(state.Intent, &ci) || !missionCoreValue(state.Lease, &cl) || !missionCoreValue(gate, &cg) || !missionCoreValue(auth, &ca) || !missionCoreValue(exec, &ce) || !missionCoreValue(cycle, &ccycle) || !missionCoreValue(outcome, &co) || !missionCoreValue(evaluation, &cev) {
		return "INVALID_SCHEMA"
	}
	var cp *corem05.ImprovementProposal
	if proposal != nil {
		var value corem05.ImprovementProposal
		if !missionCoreValue(*proposal, &value) {
			return "INVALID_SCHEMA"
		}
		cp = &value
	}
	var cr *corem05.ReviewRecord
	if review != nil {
		var value corem05.ReviewRecord
		if !missionCoreValue(*review, &value) {
			return "INVALID_SCHEMA"
		}
		cr = &value
	}
	return corem11.ValidateClosedCycle(ccycle, ci, cl, cg, ca, ce, co, cev, cp, cr)
}
