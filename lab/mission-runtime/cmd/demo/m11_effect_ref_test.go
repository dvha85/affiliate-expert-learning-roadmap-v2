package main

import "testing"

func TestM11CanonicalClosedCycleRequiresMachineEffectRef(t *testing.T) {
	s, c := baseM11()
	dir := t.TempDir()
	initM11(t, &s, dir, c.Now)
	a, g, status := AuthorizeProduction(s, c)
	if status != "AUTHORIZED" {
		t.Fatal(status)
	}
	s.Authorization = &a
	r, status := ExecuteProductionLocalSandbox(&s, a, c, dir)
	if status != "EXECUTED" {
		t.Fatal(status)
	}
	effect := EffectRef{EffectKind: "MACHINE_EXECUTION", EffectID: r.ExecutionID}
	evidenceID := s.Intent.EvidenceIDs[0]
	out := OutcomeRecord{OutcomeID: "out-effect", EffectRef: effect, ObservedAt: "2026-09-03T08:10:00Z", Status: "VALID", Metrics: map[string]float64{"clicks": 1}, SourceRef: "real-analytics"}
	ev := EvaluationRecord{EvaluationID: "eval-effect", DecisionID: s.Intent.DecisionID, EffectRef: effect, OutcomeIDs: []string{out.OutcomeID}, EvaluatedAt: "2026-09-03T08:15:00Z", Result: "SUPPORTED", EvidenceIDs: []string{evidenceID}}
	p := ImprovementProposal{ProposalID: "p-effect", EvaluationIDs: []string{ev.EvaluationID}, CurrentVersion: "v1", ProposedVersion: "v2", ChangeSummary: "narrow retry window", ExpectedBenefit: "reduce uncertainty", Rollback: "restore v1", AutoApply: false}
	review := ReviewRecord{ReviewID: "review-effect", ProposalID: p.ProposalID, ReviewedBy: "human", ReviewedAt: "2026-09-03T08:20:00Z", Decision: "APPROVE_FOR_MANUAL_CHANGE", Reason: "reviewed production evidence"}
	cycle := ProductionCycleRecord{CycleID: "cycle-effect", LeaseID: s.Lease.LeaseID, LeaseVersion: s.Lease.LeaseVersion, LeaseHash: s.Lease.LeaseHash, ObservationIDs: []string{evidenceID}, DecisionID: s.Intent.DecisionID, IntentID: s.Intent.IntentID, IntentHash: s.Intent.IntentHash, GateID: g.GateID, AuthorizationID: a.AuthorizationID, ExecutionID: r.ExecutionID, OutcomeID: out.OutcomeID, EvaluationID: ev.EvaluationID, ImprovementProposalID: p.ProposalID, ReviewID: review.ReviewID, Status: "CLOSED", OpenedAt: "2026-09-03T08:00:00Z", ClosedAt: "2026-09-03T08:20:00Z", CorrelationID: s.Intent.CorrelationID}
	if got := ValidateCanonicalProductionClosedCycle(cycle, s, g, a, r, out, ev, &p, &review); got != missionValid {
		t.Fatalf("canonical machine effect linkage rejected: %s", got)
	}
	bad := out
	bad.EffectRef = EffectRef{EffectKind: "HUMAN_ACTION", EffectID: r.ExecutionID}
	if got := ValidateCanonicalProductionClosedCycle(cycle, s, g, a, r, bad, ev, &p, &review); got != "BROKEN_LINK" {
		t.Fatalf("wrong effect kind must fail closed: %s", got)
	}
	badEval := ev
	badEval.EffectRef = EffectRef{EffectKind: "MACHINE_EXECUTION", EffectID: "other-exec"}
	if got := ValidateCanonicalProductionClosedCycle(cycle, s, g, a, r, out, badEval, &p, &review); got != "BROKEN_LINK" {
		t.Fatalf("outcome/evaluation effect mismatch must fail closed: %s", got)
	}
}
