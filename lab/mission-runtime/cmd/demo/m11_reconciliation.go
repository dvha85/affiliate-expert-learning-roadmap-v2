package main

// ProductionReconciliationContext represents the trusted control-plane view of
// human reconciliation decisions. Agent text/tool output must never populate it.
type ProductionReconciliationContext struct {
	TrustedResolutions map[string]ProductionReconciliationResolution
}

// ResolveTrustedProductionReconciliation is the canonical M11 reconciliation
// entrypoint. It requires an exact trusted control-plane record before the
// deterministic ledger can be corrected. The underlying ledger correction still
// leaves the old lease STOPPED; reconciliation is not recovery authorization.
func ResolveTrustedProductionReconciliation(state *M11State, dir string, resolution ProductionReconciliationResolution, ctx ProductionReconciliationContext) string {
	if ctx.TrustedResolutions == nil {
		return "DENY_RECONCILIATION_PROVENANCE"
	}
	trusted, ok := ctx.TrustedResolutions[resolution.ResolutionID]
	if !ok || trusted != resolution {
		return "DENY_RECONCILIATION_PROVENANCE"
	}
	return ResolveProductionReconciliation(state, dir, resolution)
}
