package main

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func m11Bundle(t *testing.T, profile string) map[string]any {
	t.Helper()
	s, c := baseM11()
	a, g, status := AuthorizeProduction(s, c)
	if status != "AUTHORIZED" {
		t.Fatal(status)
	}
	r := productionExecutionRecord(s, a, c.Now, "SUCCEEDED", "PERFORMED", "", "")
	// Initial snapshot: activation and window began at the lease start.
	activation := productionActivationRecord{s.Lease.LeaseID, s.Lease.LeaseVersion, s.Lease.LeaseHash, s.Ledger.WindowStartedAt}
	post := s.Ledger
	b := map[string]any{"profile": profile, "intent": s.Intent, "policy": s.Policy, "lease": s.Lease, "approval": c.TrustedLeaseApprovals[s.Lease.ApprovalRef], "health": s.Health, "cost": s.CostBound, "pre_ledger": s.Ledger, "gate": g, "authorization": a, "execution": r, "activation": activation}
	if profile == "resolved_stop" {
		r.Status = "RECONCILIATION_REQUIRED"
		r.SideEffectState = "UNKNOWN"
		b["execution"] = r
		stop := s.Ledger
		stop.ControlMode = "STOPPED"
		stop.StopReason = "RECONCILIATION_REQUIRED"
		stop.ReconciliationRequired = true
		stop.UpdatedAt = c.Now
		b["stop_ledger"] = stop
		x := ProductionReconciliationResolution{ResolutionID: "resolution", LeaseID: s.Lease.LeaseID, LeaseVersion: s.Lease.LeaseVersion, LeaseHash: s.Lease.LeaseHash, ExecutionID: r.ExecutionID, ResolvedBy: "human", ResolverID: "synthetic-unverified", ResolvedAt: "2026-09-03T08:10:00Z", EffectState: "NOT_PERFORMED", Reason: "synthetic"}
		b["resolution"] = x
		post = stop
		post.ReconciliationRequired = false
		post.StopReason = "RECOVERY_REVIEW_REQUIRED"
		post.ReconciliationResolutionIDs = []string{x.ResolutionID}
		post.UpdatedAt = x.ResolvedAt
	} else {
		effect := EffectRef{EffectKind: "MACHINE_EXECUTION", EffectID: r.ExecutionID}
		o := OutcomeRecord{OutcomeID: "out", EffectRef: effect, ObservedAt: "2026-09-03T08:10:00Z", Status: "VALID", Metrics: map[string]float64{"clicks": 1}, SourceRef: "synthetic"}
		e := EvaluationRecord{EvaluationID: "eval", DecisionID: s.Intent.DecisionID, EffectRef: effect, OutcomeIDs: []string{o.OutcomeID}, EvaluatedAt: "2026-09-03T08:15:00Z", Result: "SUPPORTED", EvidenceIDs: s.Intent.EvidenceIDs}
		cy := ProductionCycleRecord{CycleID: "cycle", LeaseID: s.Lease.LeaseID, LeaseVersion: s.Lease.LeaseVersion, LeaseHash: s.Lease.LeaseHash, ObservationIDs: s.Intent.EvidenceIDs, DecisionID: s.Intent.DecisionID, IntentID: s.Intent.IntentID, IntentHash: s.Intent.IntentHash, GateID: g.GateID, AuthorizationID: a.AuthorizationID, ExecutionID: r.ExecutionID, OutcomeID: o.OutcomeID, EvaluationID: e.EvaluationID, Status: "CLOSED", OpenedAt: c.Now, ClosedAt: e.EvaluatedAt, CorrelationID: s.Intent.CorrelationID}
		e.Limitations = []string{"synthetic, unverified"}
		b["outcome"], b["evaluation"], b["cycle"] = o, e, cy
		post.ExecutionsTotal = 1
		post.ExecutionsInWindow = 1
		post.CostMinorTotal = s.CostBound.MaxCostMinor
		post.SuccessfulIdempotencyKeys = []string{s.Intent.IdempotencyKey}
		post.OutcomeLinks = []ProductionOutcomeLink{{o.OutcomeID, r.ExecutionID, o.ObservedAt}}
		post.LastExecutionAt = c.Now
		post.LastOutcomeAt = o.ObservedAt
		post.UpdatedAt = o.ObservedAt
	}
	b["post_ledger"] = post
	return b
}

func auditBundle(t *testing.T, b map[string]any) (M11ChainSummary, string) {
	t.Helper()
	raw, e := json.Marshal(b)
	if e != nil {
		t.Fatal(e)
	}
	return CheckM11Chain(raw)
}
func mutateBundle(t *testing.T, b map[string]any, artifact, field string, value any) {
	t.Helper()
	raw, e := json.Marshal(b[artifact])
	if e != nil {
		t.Fatal(e)
	}
	var m map[string]json.RawMessage
	json.Unmarshal(raw, &m)
	m[field], _ = json.Marshal(value)
	b[artifact] = m
}

func TestM11ChainProfiles(t *testing.T) {
	b0 := m11Bundle(t, "closed_cycle")
	for k, v := range b0 {
		raw, _ := json.Marshal(v)
		switch k {
		case "outcome":
			if _, s := DecodeM03Outcome(raw); s != missionValid {
				t.Fatalf("%s %s %s", k, s, raw)
			}
		case "evaluation":
			if _, s := DecodeM05Evaluation(raw); s != missionValid {
				t.Fatalf("%s %s %s", k, s, raw)
			}
		}
	}
	for _, profile := range []string{"closed_cycle", "resolved_stop"} {
		b := m11Bundle(t, profile)
		s, state := auditBundle(t, b)
		if state != missionValid || s.Result != "CONSISTENT_UNVERIFIED" || s.ExecutionPermitted || s.ResumePermitted || s.ProvenanceAuthenticated {
			t.Fatal(profile, state, s)
		}
	}
	b := m11Bundle(t, "resolved_stop")
	x := b["resolution"].(ProductionReconciliationResolution)
	x.EffectState = "PERFORMED"
	b["resolution"] = x
	p := b["post_ledger"].(ProductionLedger)
	p.ExecutionsTotal++
	p.ExecutionsInWindow++
	p.CostMinorTotal = b["cost"].(CanaryCostBound).MaxCostMinor
	p.PendingOutcomes++
	p.PendingExecutionIDs = []string{b["execution"].(ProductionExecutionRecord).ExecutionID}
	p.SuccessfulIdempotencyKeys = []string{b["intent"].(ShadowActionIntent).IdempotencyKey}
	b["post_ledger"] = p
	if _, state := auditBundle(t, b); state != missionValid {
		t.Fatal(state)
	}
}

func TestM11ChainMutations(t *testing.T) {
	for _, tc := range []struct {
		artifact, field string
		value           any
		want            string
	}{
		{"approval", "source_canary_grant_id", "other", "BROKEN_LINK"}, {"approval", "reviewer_id", "other", "BROKEN_LINK"},
		{"approval", "validated_risk_classes", []string{"RISK1"}, "SCOPE_NOT_DELEGATED"},
		{"activation", "lease_id", "other", "BROKEN_LINK"}, {"activation", "activated_at", "2026-09-03T08:10:00Z", "INVALID_TIME_BINDING"},
		{"gate", "health_snapshot_id", "other", "BROKEN_LINK"}, {"authorization", "production_cost_bound_minor", 101, "BROKEN_LINK"}, {"execution", "idempotency_key", "other", "BROKEN_LINK"},
		{"authorization", "execution_mode", "GOVERNED_CANARY", "INVALID_SCHEMA"},
		{"pre_ledger", "control_mode", "STOPPED", "SAFETY_BLOCKED"}, {"pre_ledger", "reconciliation_required", true, "SAFETY_BLOCKED"},
		{"gate", "executions_total_before", 1, "LEDGER_SNAPSHOT_MISMATCH"}, {"post_ledger", "executions_total", 0, "INVALID_LEDGER"},
		{"post_ledger", "cost_minor_total", 0, "INVALID_LEDGER_TRANSITION"}, {"post_ledger", "control_mode", "STOPPED", "INVALID_LEDGER_TRANSITION"},
		{"cycle", "outcome_id", "other", "BROKEN_LINK"}, {"cycle", "closed_at", "2026-09-03T08:11:00Z", "BROKEN_CYCLE"},
		{"evaluation", "effect_ref", EffectRef{EffectKind: "MACHINE_EXECUTION", EffectID: "wrong"}, "BROKEN_LINK"},
		{"execution", "attempted_at", "2026-09-03T09:00:00Z", "INVALID_TIME_BINDING"},
	} {
		t.Run(tc.artifact+"/"+tc.field, func(t *testing.T) {
			b := m11Bundle(t, "closed_cycle")
			mutateBundle(t, b, tc.artifact, tc.field, tc.value)
			s, state := auditBundle(t, b)
			if state != tc.want || s != (M11ChainSummary{}) {
				t.Fatal(state, tc.want, s)
			}
		})
	}
	for _, profile := range []string{"closed_cycle", "resolved_stop"} {
		for key := range m11Bundle(t, profile) {
			b := m11Bundle(t, profile)
			delete(b, key)
			if _, s := auditBundle(t, b); s == missionValid {
				t.Fatal("missing", key)
			}
			b = m11Bundle(t, profile)
			b[key] = nil
			if _, s := auditBundle(t, b); s == missionValid {
				t.Fatal("null", key)
			}
		}
	}
	b := m11Bundle(t, "resolved_stop")
	mutateBundle(t, b, "post_ledger", "control_mode", "NORMAL")
	if _, s := auditBundle(t, b); s != "INVALID_STOP_TRANSITION" {
		t.Fatal(s)
	}
	b = m11Bundle(t, "resolved_stop")
	mutateBundle(t, b, "stop_ledger", "reconciliation_required", false)
	if _, s := auditBundle(t, b); s != "INVALID_STOP_TRANSITION" {
		t.Fatal(s)
	}
}

func TestM11ChainReview(t *testing.T) {
	b := m11Bundle(t, "closed_cycle")
	p := ImprovementProposal{ProposalID: "p", EvaluationIDs: []string{"eval"}, CurrentVersion: "v1", ProposedVersion: "v2", ChangeSummary: "synthetic", ExpectedBenefit: "synthetic", Rollback: "v1", AutoApply: false}
	b["proposal"] = p
	cy := b["cycle"].(ProductionCycleRecord)
	cy.ImprovementProposalID = p.ProposalID
	cy.Status = "REVIEW_PENDING"
	b["cycle"] = cy
	if _, s := auditBundle(t, b); s != missionValid {
		t.Fatal(s)
	}
	r := ReviewRecord{ReviewID: "review", ProposalID: p.ProposalID, ReviewedBy: "human", ReviewedAt: "2026-09-03T08:20:00Z", Decision: "APPROVE_FOR_MANUAL_CHANGE", Reason: "synthetic"}
	b["review"] = r
	cy.ReviewID = r.ReviewID
	cy.Status = "CLOSED"
	cy.ClosedAt = r.ReviewedAt
	b["cycle"] = cy
	if _, s := auditBundle(t, b); s != missionValid {
		t.Fatal(s)
	}
	p.AutoApply = true
	b["proposal"] = p
	if _, s := auditBundle(t, b); s == missionValid {
		t.Fatal("auto apply")
	}
}

func TestM11ChainCLI(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bundle.json")
	raw, _ := json.Marshal(m11Bundle(t, "closed_cycle"))
	if e := os.WriteFile(p, raw, 0600); e != nil {
		t.Fatal(e)
	}
	var out bytes.Buffer
	for n := 0; n < 2; n++ {
		out.Reset()
		if e := runM11ChainCheck(&out, []string{p}); e != nil {
			t.Fatal(e)
		}
		var s M11ChainSummary
		json.Unmarshal(out.Bytes(), &s)
		if s.Result != "CONSISTENT_UNVERIFIED" || s.ExecutionPermitted || s.ResumePermitted || s.ProvenanceAuthenticated {
			t.Fatal(s)
		}
	}
	if e := runM11ChainCheck(m03BrokenWriter{}, []string{p}); e == nil {
		t.Fatal("writer")
	}
	got, e := os.ReadFile(p)
	if e != nil || !bytes.Equal(got, raw) {
		t.Fatal("changed input")
	}
	for _, args := range [][]string{nil, {p, p}, {"missing"}} {
		out.Reset()
		if e := runM11ChainCheck(&out, args); e == nil || out.Len() != 0 {
			t.Fatal("args")
		}
	}
	for _, bad := range [][]byte{[]byte(`null`), append(append([]byte{}, raw...), []byte(` {}`)...), append([]byte(`{"profile":"closed_cycle",`), raw[1:]...)} {
		if _, s := CheckM11Chain(bad); s == missionValid {
			t.Fatal("ambiguous bundle")
		}
	}
	entries, e := os.ReadDir(dir)
	if e != nil || len(entries) != 1 {
		t.Fatal("persisted state")
	}
}

func TestM11ChainHealthAndBudget(t *testing.T) {
	for _, change := range []string{"stale", "degraded", "alert"} {
		b := m11Bundle(t, "closed_cycle")
		h := b["health"].(ProductionHealthSnapshot)
		switch change {
		case "stale":
			h.ObservedAt = "2026-09-03T07:00:00Z"
		case "degraded":
			h.DependencyState = "DEGRADED"
		case "alert":
			h.ComplianceAlertCount = 1
		}
		h = SealProductionHealth(h)
		b["health"] = h
		mutateBundle(t, b, "gate", "health_snapshot_hash", h.SnapshotHash)
		for _, k := range []string{"authorization", "execution"} {
			mutateBundle(t, b, k, "production_health_snapshot_hash", h.SnapshotHash)
		}
		_, s := auditBundle(t, b)
		if change == "stale" && s != "STALE_HEALTH" || change != "stale" && s != "SAFETY_BLOCKED" {
			t.Fatal(change, s)
		}
	}
	b := m11Bundle(t, "closed_cycle")
	mutateBundle(t, b, "pre_ledger", "cost_minor_total", math.MaxInt64)
	mutateBundle(t, b, "gate", "cost_minor_total_before", math.MaxInt64)
	if _, s := auditBundle(t, b); s != "BUDGET_EXCEEDED" {
		t.Fatal(s)
	}
	b = m11Bundle(t, "closed_cycle")
	mutateBundle(t, b, "pre_ledger", "cost_minor_total", 901)
	mutateBundle(t, b, "gate", "cost_minor_total_before", 901)
	if _, s := auditBundle(t, b); s != "BUDGET_EXCEEDED" {
		t.Fatal(s)
	}
	// An expired rate window resets its counter without resetting total.
	b = m11Bundle(t, "closed_cycle")
	mutateBundle(t, b, "lease", "window_seconds", 60)
	raw, _ := json.Marshal(b["lease"])
	var l ProductionLease
	json.Unmarshal(raw, &l)
	l = SealProductionLease(l)
	b["lease"] = l
	for _, k := range []string{"approval", "health", "pre_ledger", "post_ledger", "gate", "activation", "cycle"} {
		mutateBundle(t, b, k, "lease_hash", l.LeaseHash)
	}
	for _, k := range []string{"authorization", "execution"} {
		mutateBundle(t, b, k, "production_lease_hash", l.LeaseHash)
	}
	raw, _ = json.Marshal(b["health"])
	var h ProductionHealthSnapshot
	json.Unmarshal(raw, &h)
	h = SealProductionHealth(h)
	b["health"] = h
	mutateBundle(t, b, "gate", "health_snapshot_hash", h.SnapshotHash)
	for _, k := range []string{"authorization", "execution"} {
		mutateBundle(t, b, k, "production_health_snapshot_hash", h.SnapshotHash)
	}
	mutateBundle(t, b, "post_ledger", "window_started_at", "2026-09-03T08:00:00Z")
	if _, s := auditBundle(t, b); s != missionValid {
		t.Fatal(s)
	}
}

// Verify audit profiles against actual local sandbox transitions, not only
// hand-assembled ledgers. No external adapter or network is used.
func TestM11ChainRuntimeTransitions(t *testing.T) {
	for _, profile := range []string{"closed_cycle", "resolved_stop"} {
		t.Run(profile, func(t *testing.T) {
			b := m11Bundle(t, profile)
			s, c := baseM11()
			dir := t.TempDir()
			initM11(t, &s, dir, c.Now)
			b["pre_ledger"] = s.Ledger
			b["activation"] = productionActivationRecord{s.Lease.LeaseID, s.Lease.LeaseVersion, s.Lease.LeaseHash, c.Now}
			a, g, status := AuthorizeProduction(s, c)
			if status != "AUTHORIZED" {
				t.Fatal(status)
			}
			b["gate"], b["authorization"] = g, a
			s.Authorization = &a
			if profile == "resolved_stop" {
				if e := os.WriteFile(sandboxIdempotencyPath(dir, a.IdempotencyKey), []byte("uncertain"), 0600); e != nil {
					t.Fatal(e)
				}
			}
			r, status := ExecuteProductionLocalSandbox(&s, a, c, dir)
			b["execution"] = r
			if profile == "closed_cycle" {
				if status != "EXECUTED" {
					t.Fatal(status)
				}
				if status = RecordProductionOutcome(&s, dir, "out", r.ExecutionID, "2026-09-03T08:10:00Z"); status != "OUTCOME_RECORDED" {
					t.Fatal(status)
				}
			} else {
				if status != "STOP_RECONCILIATION" {
					t.Fatal(status)
				}
				b["stop_ledger"] = s.Ledger
				x := b["resolution"].(ProductionReconciliationResolution)
				if status = ResolveTrustedProductionReconciliation(&s, dir, x, ProductionReconciliationContext{TrustedResolutions: map[string]ProductionReconciliationResolution{x.ResolutionID: x}}); status != "RECONCILIATION_RESOLVED_STOPPED" {
					t.Fatal(status)
				}
			}
			b["post_ledger"] = s.Ledger
			if _, status = auditBundle(t, b); status != missionValid {
				t.Fatal(status)
			}
		})
	}
}
