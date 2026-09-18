# RP-08 evidence: M11 ledger-outcome reverse mutation proof

- Snapshot date: 2026-09-19
- Implementation PR: #453
- Implementation head: `eb36c98a4c4c0bc55ee06857db338e3286dbd0bd`
- Merge commit: `f1e29560561da1f8757f725f00d39ee50536b763`
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

## Hosted verification

- Curriculum CI run `35391489583` and Mission Agent Path CI run `35391489747`
  passed all 13 required hosted checks on the exact implementation head.
- Deterministic backup/mutation job `105750642048` passed the BR-18b smoke and
  the new ledger-outcome reverse mutation proof.
- Windows runtime job `105750642188` and learner race job `105750641947` passed.

## Local verification

- `python scripts/audit_readiness.py`: PASS; 144 claims; remains
  `NOT_READY_FOR_PRODUCTION`.
- `python -m unittest scripts.tests.test_audit_readiness`: PASS (119 tests).
- Python compilation, readiness JSON parsing and `git diff --check`: PASS.
- Local Go execution remains unavailable because `go.exe` is absent; hosted
  Linux and Windows jobs provide the executable Go acceptance evidence.
