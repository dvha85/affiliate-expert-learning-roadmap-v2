# RP-08 post-merge evidence: M11 failed-outcome mutation proof

- Snapshot date: 2026-09-18
- Implementation PR: #439
- Implementation head: `b2762358405135db2b2dc7c07e412d823f0d1cae`
- Merge commit: `4032bb238bb121a90d451628a333fab8ff2f623d`
- Status: `NOT_READY_FOR_PRODUCTION`

## Scope

PR #439 adds a disposable-copy mutation runner and a real BR-18b
backup/restore case. The fixture keeps one valid FAILED/outcome pair, then
appends a checksum-valid second FAILED M11 execution without its fixture
outcome. Restore must return `GRAPH_FAILED` and must not publish the target.
The mutation removes only the semantic guard; the real smoke must fail on that
mutated source or the CI job fails. This closes one bounded offline mutation
gap only. It does not claim complete mutation breadth,
power-loss/filesystem-crash recovery, atomic multi-file publication,
distributed/multi-host safety, Windows traversal parity, provider/live
execution, deployment, business outcomes, or clean-machine/target-host
readiness.

## Hosted verification

- Curriculum CI run `35367101749`: all 13 checks passed.
- Mission Agent Path CI run `35367101788`: all 13 checks passed.
- Hosted checks included Windows runtime job `105672103306`, learner race job
  `105672102842`, and the required `deterministic-smokes-backup-mutations`
  job `105672102565`.

## Local verification

- `python -m py_compile scripts/smoke_br18b_backup_restore.py scripts/mutate_backup_m11_failed_outcome.py`: PASS.
- `python scripts/audit_readiness.py`: PASS; the tracker remains
  `NOT_READY_FOR_PRODUCTION`.
- `python -m unittest scripts.tests.test_audit_readiness`: PASS (118 tests);
  local Go execution remains unavailable because `go.exe` is absent.
- `git diff --check`: PASS.

## Boundary

This is bounded offline fixture mutation evidence for the M11
failed-execution-to-outcome graph guard. It does not establish crash or
power-loss durability, atomic multi-file publication, distributed/multi-host
locking, provider/live execution, deployment or pilot readiness, business
outcomes, or clean-machine/target-host readiness.
