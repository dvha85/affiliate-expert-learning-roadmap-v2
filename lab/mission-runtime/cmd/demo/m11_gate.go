package main

import "os"

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
	gate := EvaluateProductionGate(*state, ctx)
	if gate.Decision == "STOP" {
		if err := persistStickyStop(ledgerPath, state, gate.Reason, ctx.Now); err != nil {
			return gate, "STOP_PERSISTENCE_FAILED"
		}
	}
	return gate, gate.Decision
}
