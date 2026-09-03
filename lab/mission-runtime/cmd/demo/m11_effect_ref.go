package main

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
	return ValidateProductionClosedCycle(cycle, state, gate, auth, exec, outcome, evaluation, proposal, review)
}
