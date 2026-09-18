# RP-08 post-merge evidence: M11 reservation-lineage mutation proof

- Snapshot date: 2026-09-19
- Implementation PR: #441
- Implementation head: `64bf4cb7858ed4670ed40aa61d3b0c9c24a93f96`
- Merge commit: `08319c7998895a54efbcd6a39458c6abdcd28311`
- Status: `NOT_READY_FOR_PRODUCTION`

## Scope

PR #441 adds a disposable-copy mutation runner for the M11 execution-to-
reservation-ledger invariant. The focused Go regression first proves a valid
cancelled execution is accepted, then replaces only the reservation ledger's
pending execution ID with a checksum-valid unrelated ID while preserving the
ledger counters and entry envelope. Validation must reject the orphaned
execution. The mutation removes both duplicate core graph and backup-restore
guards in a disposable copy and runs the focused tests inside their actual Go
module boundaries; an unexpectedly green test or an unrelated harness failure
fails CI.

This closes one bounded offline mutation-coverage gap. It does not claim
complete mutation breadth, crash or power-loss durability, atomic multi-file
publication, distributed or multi-host locking, Windows traversal parity,
provider/live execution, deployment, business outcomes, pilot readiness or
clean-machine/target-host readiness.

## Hosted verification

- Curriculum CI run `35373009084`: all 9 Curriculum jobs passed.
- Mission Agent Path CI run `35373009156`: all 4 Mission checks passed.
- Windows runtime job `105691163624` passed the real Windows test/vet and
  lock/backup/restore coverage.
- Learner race job `105691163542` passed.
- `deterministic-smokes-backup-mutations` job `105691163529` passed, including
  the new reservation-lineage mutation.
- Deterministic quickstart job `105691163755`, foundations job `105691163870`
  and M06/M07 smoke job `105691163838` passed.

## Local verification

- `python scripts/audit_readiness.py`: PASS; the tracker remains
  `NOT_READY_FOR_PRODUCTION` with BR-13, BR-14, BR-15, BR-16a, BR-17, BR-18b
  and BR-19 partial/open.
- `python -m unittest scripts.tests.test_audit_readiness`: PASS (118 tests).
- `python -m py_compile scripts/mutate_m11_reservation_lineage.py`: PASS.
- `git diff --check`: PASS.
- Local Go execution remains unavailable because `go.exe` is absent; hosted
  Linux and Windows jobs are the executable Go acceptance evidence.

## Evidence boundary

The result proves that the tested offline M11 graph and backup/restore paths
fail closed for this checksum-valid orphan reservation and that CI detects
removal of both duplicate guards. It does not establish filesystem crash or
power-loss behavior, atomic multi-file transactions, distributed locking,
provider operation, live execution, deployment recovery, or production
readiness.
