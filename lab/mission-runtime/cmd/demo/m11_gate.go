package main

import (
	"os"
	"time"
)

// trustedCriticalProductionStop inspects only already-bound durable/trusted safety
// inputs. It deliberately runs before ordinary policy/action eligibility checks so
// a simultaneous policy error cannot hide a compliance/reconciliation/failure STOP.
func trustedCriticalProductionStop(state M11State, ctx M11Context) string {
	if state.Ledger.ControlMode == "STOPPED" { return "STICKY_STOP" }
	if state.Ledger.ReconciliationRequired { return "RECONCILIATION_REQUIRED" }
	if state.Ledger.ConsecutiveFailures >= state.Lease.MaxConsecutiveFailures && state.Lease.MaxConsecutiveFailures > 0 { return "FAILURE_THRESHOLD" }
	if containsFold(ctx.RevokedLeaseIDs, state.Lease.LeaseID) { return "LEASE_REVOKED" }
	if ctx.KillSwitch { return "KILL_SWITCH_ACTIVE" }

	h := state.Health
	if h.HashVersion != "go-json-v1" || h.SnapshotHash == "" || h.SnapshotHash != ComputeProductionHealthHash(h) { return "" }
	if h.LeaseID != state.Lease.LeaseID || h.LeaseVersion != state.Lease.LeaseVersion || h.LeaseHash != state.Lease.LeaseHash || len(h.SourceRefs) == 0 { return "" }
	trustedHash, ok := ctx.TrustedHealthSnapshots[h.SnapshotID]
	if !ok || trustedHash != h.SnapshotHash { return "" }
	observedAt, errObserved := time.Parse(time.RFC3339, h.ObservedAt)
	now, errNow := time.Parse(time.RFC3339, ctx.Now)
	if errObserved != nil || errNow != nil || observedAt.After(now) { return "" }
	if h.ComplianceAlertCount > 0 { return "COMPLIANCE_ALERT" }
	if h.ReconciliationRequired { return "RECONCILIATION_REQUIRED" }
	if h.ConsecutiveFailures >= state.Lease.MaxConsecutiveFailures && state.Lease.MaxConsecutiveFailures > 0 { return "FAILURE_THRESHOLD" }
	if h.OldestPendingOutcomeAgeSeconds > state.Lease.MaxOutcomeAgeSeconds && state.Lease.MaxOutcomeAgeSeconds > 0 { return "OUTCOME_STALE" }
	return ""
}

// EnforceProductionGate evaluates the current production state against durable
// ledger state and persists every STOP decision before the next cycle can run.
// DEGRADE/WAIT/DENY remain non-mutating because they do not revoke the lease.
func EnforceProductionGate(state *M11State, ctx M11Context, dir string) (ProductionGateDecision, string) {
	if state == nil {
		return ProductionGateDecision{Decision: "DENY", Reason: "INVALID_PRODUCTION_STATE"}, "DENY"
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ProductionGateDecision{Decision: "STOP", Reason: "DURABLE_STATE_UNAVAILABLE"}, "STOP_PERSISTENCE_FAILED"
	}
	lockPath := productionLockPath(dir, state.Lease.LeaseID, state.Lease.LeaseVersion)
	lock, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		if os.IsExist(err) { return ProductionGateDecision{Decision:"WAIT",Reason:"PRODUCTION_LOCKED"}, "WAIT_PRODUCTION_LOCK" }
		return ProductionGateDecision{Decision:"STOP",Reason:"DURABLE_STATE_UNAVAILABLE"}, "STOP_PERSISTENCE_FAILED"
	}
	_ = lock.Close()
	defer os.Remove(lockPath)

	activationPath := productionActivationPath(dir, state.Lease.LeaseID, state.Lease.LeaseVersion)
	if !productionActivationMatches(activationPath, state.Lease) {
		return ProductionGateDecision{Decision:"STOP",Reason:"ACTIVATION_STATE_MISSING"}, "STOP_ACTIVATION_STATE_MISSING"
	}
	ledgerPath := productionLedgerPath(dir, state.Lease.LeaseID, state.Lease.LeaseVersion)
	ledger, err := loadProductionLedger(ledgerPath)
	if os.IsNotExist(err) {
		return ProductionGateDecision{Decision:"STOP",Reason:"LEDGER_MISSING"}, "STOP_LEDGER_MISSING"
	}
	if err != nil {
		return ProductionGateDecision{Decision:"STOP",Reason:"LEDGER_UNREADABLE"}, "STOP_LEDGER_UNREADABLE"
	}
	state.Ledger = ledger

	if reason := trustedCriticalProductionStop(*state, ctx); reason != "" {
		gate := productionGate(*state, ctx)
		gate.Decision = "STOP"
		gate.Reason = reason
		if err := persistStickyStop(ledgerPath, state, reason, ctx.Now); err != nil {
			return gate, "STOP_PERSISTENCE_FAILED"
		}
		return gate, "STOP"
	}

	gate := EvaluateProductionGate(*state, ctx)
	if gate.Decision == "STOP" {
		if err := persistStickyStop(ledgerPath, state, gate.Reason, ctx.Now); err != nil {
			return gate, "STOP_PERSISTENCE_FAILED"
		}
	}
	return gate, gate.Decision
}
