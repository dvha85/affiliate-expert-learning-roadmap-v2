# RP-08 evidence: M11 reverse-ledger graph mutation proof

- Snapshot date: 2026-09-19
- Implementation PR: pending hosted review
- Status: `PARTIAL`

## Scope

The implementation branch adds a disposable-copy mutation that removes only
the core M11 reverse-ledger graph guard. The guard requires every
checksum-valid `ReconciliationResolutionIDs` link in a stopped ledger to
resolve to the same reconciliation execution and lease lineage. The mutation
then runs the real `core/m11` graph regression and requires the named orphan
resolution assertion to fail. An unrelated failure or an unexpectedly green
mutation is itself a mutation-runner failure.

This is bounded offline/read-only mutation evidence. It does not prove ledger
durability, crash or power-loss recovery, atomic multi-file publication,
distributed or multi-host locking, Windows traversal parity, provider/live
execution, deployment, pilot acceptance, business outcome, or production
readiness.

## Verification

- The focused core graph regression is `TestArtifactGraphAcceptsAndRejectsExactProductionLifecycleLinks`.
- Hosted exact-head verification is required before this evidence can be marked
  `VERIFIED_OFFLINE`.
- Local `go.exe` is unavailable in the current worktree; local Python audit
  validation remains separate from hosted Go execution.

## Boundary

The readiness state remains `NOT_READY_FOR_PRODUCTION`. This change proves only
that the named reverse-ledger graph control is connected to its real core
regression; it does not broaden the production, operated, provider, or
deployment evidence boundary.
