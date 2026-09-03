package main

import "testing"

func TestM11CriticalStopPersistsEvenWhenPolicyIsInvalid(t *testing.T) {
	state, ctx := baseM11()
	dir := t.TempDir()
	initM11(t, &state, dir, ctx.Now)

	// A normal policy evaluation would DENY this tampered state before reaching
	// health checks. The stateful gate must still persist the independently
	// trusted compliance STOP so a later healthy-looking cycle cannot resume it.
	state.Policy.Reason = "tampered-policy-result"
	state.Health.ComplianceAlertCount = 1
	refreshHealthTrust(&state, &ctx)
	gate, status := EnforceProductionGate(&state, ctx, dir)
	if status != "STOP" || gate.Reason != "COMPLIANCE_ALERT" {
		t.Fatalf("critical STOP must outrank ordinary policy failure: %+v %s", gate, status)
	}

	state.Policy = EvaluateShadowPolicy(state.Intent, ctx.PolicyContext)
	state.Health.ComplianceAlertCount = 0
	state.Health = SealProductionHealth(state.Health)
	ctx.TrustedHealthSnapshots[state.Health.SnapshotID] = state.Health.SnapshotHash
	gate, status = EnforceProductionGate(&state, ctx, dir)
	if status != "STOP" || gate.Reason != "STICKY_STOP" {
		t.Fatalf("persisted STOP must survive repaired policy and cleared alert: %+v %s", gate, status)
	}
}
