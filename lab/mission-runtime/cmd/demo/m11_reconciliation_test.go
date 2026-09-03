package main

import (
	"os"
	"testing"
)

func TestM11ReconciliationRequiresTrustedControlPlaneRecord(t *testing.T) {
	state, ctx := baseM11()
	dir := t.TempDir()
	initM11(t, &state, dir, ctx.Now)
	auth, _, status := AuthorizeProduction(state, ctx)
	if status != "AUTHORIZED" { t.Fatal(status) }
	state.Authorization = &auth
	if err := os.WriteFile(sandboxIdempotencyPath(dir, auth.IdempotencyKey), []byte("unknown"), 0600); err != nil { t.Fatal(err) }
	rec, status := ExecuteProductionLocalSandbox(&state, auth, ctx, dir)
	if status != "STOP_RECONCILIATION" { t.Fatalf("expected reconciliation stop: %s", status) }

	resolution := ProductionReconciliationResolution{
		ResolutionID: "trusted-r1", LeaseID: state.Lease.LeaseID, LeaseVersion: state.Lease.LeaseVersion,
		LeaseHash: state.Lease.LeaseHash, ExecutionID: rec.ExecutionID, ResolvedBy: "human", ResolverID: "learner",
		ResolvedAt: "2026-09-03T08:20:00Z", EffectState: "NOT_PERFORMED", Reason: "provider audit confirmed no side effect",
	}
	if got := ResolveTrustedProductionReconciliation(&state, dir, resolution, ProductionReconciliationContext{}); got != "DENY_RECONCILIATION_PROVENANCE" {
		t.Fatalf("untrusted human-shaped data must not alter reconciliation accounting: %s", got)
	}
	trusted := ProductionReconciliationContext{TrustedResolutions: map[string]ProductionReconciliationResolution{resolution.ResolutionID: resolution}}
	if got := ResolveTrustedProductionReconciliation(&state, dir, resolution, trusted); got != "RECONCILIATION_RESOLVED_STOPPED" {
		t.Fatalf("trusted resolution should correct ledger while staying stopped: %s", got)
	}
	if state.Ledger.ControlMode != "STOPPED" || state.Ledger.ReconciliationRequired {
		t.Fatalf("resolution must clear uncertainty but not STOP: %+v", state.Ledger)
	}
}
