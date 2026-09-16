# Local regression after PR #398 — 2026-09-16

## Scope

This maintainer-run verification exercised the merged `main` head
`1c071ca539de7e77050f43804d91f36884921b58` after PR #398. It used disposable
local workspaces, synthetic fixtures and read-only/loopback paths only. It did
not use provider credentials, call an affiliate endpoint, invoke a live
executor, or establish business-outcome, clean-machine pilot, deployment,
multi-host or power-loss evidence.

## Results

All checks below passed:

- `GOWORK=off go test -count=1 ./...` and `GOWORK=off go vet ./...` in each of
  `contracts`, `core`, `lab/affiliate-bot` and `lab/mission-runtime`.
- `python3 -m unittest discover -s scripts/tests -v` — 119 tests, PASS.
- `python3 scripts/smoke_br16a_offline.py` — BR-16a PASS: shared M00→M11
  fixture, restart/replay, backup/restore and durable STOP.
- `python3 scripts/smoke_br18b_backup_restore.py` — BR-18b PASS: typed v3
  graph, M11 lineage/restore checks and durable STOP.
- `GOWORK=off go test -race -count=1 ./cmd/bot -run
  '^TestM11ProcessKillAfterLedgerAppendLeavesJournalForFreshRecovery$'` —
  PASS for FAILED and UNKNOWN→STOP.
- `python3 scripts/audit_readiness.py` — `NOT_READY_FOR_PRODUCTION`, 76
  scoped evidence claims resolved.

The post-ledger regression killed a separately executed learner Bot child with
POSIX `SIGKILL` after the M11 execution/STOP ledger transition was appended and
directory-synced. A fresh reader returned `RECOVERY_REQUIRED`; a locked writer
recovered the exact transition without duplicating execution or ledger
artifacts. The detailed boundary record is
`docs/architecture/EVIDENCE-M11-PROCESS-KILL-AFTER-LEDGER-20260916.md`.

## Interpretation

The result confirms the declared local fixture/process-termination boundary on
the merged head. It does not upgrade any readiness package or prove filesystem
power-loss durability, atomic multi-file commit, Windows native locking,
distributed coordination, provider operation, business outcomes, pilot or
deployment readiness. Overall readiness remains
`NOT_READY_FOR_PRODUCTION`.
