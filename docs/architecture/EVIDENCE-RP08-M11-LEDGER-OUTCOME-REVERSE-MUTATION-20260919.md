# RP-08 evidence: M11 ledger-outcome reverse mutation proof

- Snapshot date: 2026-09-19
- Implementation branch: `codex/rp08-m11-ledger-outcome-mutation`
- Implementation PR: pending hosted review
- Status: `NOT_READY_FOR_PRODUCTION`

## Scope

The BR-18b backup/restore smoke already creates a checksum-valid copy whose
`PRODUCTION_LEDGER` has no `outcome_links`. The restore path must reject that
snapshot before publishing the target because every persisted M11 outcome must
remain represented in exactly one restored ledger.

This change adds a disposable-copy mutation runner that removes only the
reverse `ledgerOutcomeLinks` cardinality guard in `backup_command.go`. The real
BR-18b smoke must then fail at the named `orphan-ledger-outcome-restored`
assertion; an unexpectedly green smoke or an unrelated failure fails the
mutation runner.

This is bounded offline fixture mutation evidence. It does not prove complete
M11 mutation breadth, crash or power-loss durability, atomic multi-file
publication, distributed or multi-host locking, Windows traversal parity,
provider/live execution, deployment, pilot acceptance, business outcome or
production readiness.

## Verification boundary

- Local Python syntax and readiness tests are required on the implementation
  head.
- The BR-18b smoke and mutation runner require Go and are therefore accepted
  through the exact-head hosted Curriculum CI run when local `go.exe` is absent.
- Hosted run IDs and the merged-main baseline are recorded in a separate
  post-merge evidence synchronization change.
