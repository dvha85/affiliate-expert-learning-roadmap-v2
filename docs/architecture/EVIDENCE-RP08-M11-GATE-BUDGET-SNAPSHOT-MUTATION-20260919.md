# RP-08 post-merge evidence: M11 gate budget snapshot mutation proof

- Snapshot date: 2026-09-19
- Implementation PR: #443
- Implementation head: `536a9029033020c14e5960aeca641ca304f60392`
- Merge commit: `e95cc321a37e3f8f239a594fa0e23412cbb43571`
- Status: `NOT_READY_FOR_PRODUCTION`

## Scope

PR #443 adds a disposable-copy mutation runner for the M11 production-gate
ledger snapshot invariant. The real core graph regression constructs a valid
gate, then changes only `ExecutionsTotalBefore` while keeping the referenced
ledger and artifact envelopes otherwise valid; validation must reject the
drifted counter. The mutation removes only the execution-counter comparison in
`core/m11/artifact_registry.go` and requires the real regression to fail with
the drifted-counter assertion. An unexpectedly green test or an unrelated Go
harness failure fails the runner.

This closes one bounded offline mutation-coverage gap. It does not claim
complete M11 mutation breadth, crash or power-loss durability, atomic
multi-file publication, distributed or multi-host locking, Windows traversal
parity, provider/live execution, deployment, business outcomes, pilot
readiness or clean-machine/target-host readiness.

## Hosted verification

- Curriculum CI run `35375980090`: all 10 Curriculum jobs passed.
- Mission Agent Path CI run `35375980154`: all 3 Mission checks passed.
- `deterministic-smokes-backup-mutations` job `105700643729` passed,
  including the new gate budget snapshot mutation.
- Windows runtime job `105700644179` passed.
- Learner race job `105700643942` passed.
- GitHub reported 13/13 checks passed on the exact implementation head before
  squash merge.

## Local verification

- `python scripts/audit_readiness.py`: PASS; the tracker remains
  `NOT_READY_FOR_PRODUCTION` with BR-13, BR-14, BR-15, BR-16a, BR-17, BR-18b
  and BR-19 partial/open.
- `python -m unittest scripts.tests.test_audit_readiness`: PASS (118 tests).
- `python -m py_compile` for the changed Python files: PASS.
- `git diff --check`: PASS.
- Local Go execution remains unavailable because `go.exe` is absent; hosted
  Linux and Windows jobs are the executable Go acceptance evidence.

## Evidence boundary

The result proves that the tested offline M11 graph regression detects removal
of the gate execution-counter snapshot guard. It does not establish
filesystem crash or power-loss behavior, atomic multi-file transactions,
distributed locking, provider operation, live execution, deployment recovery,
or production readiness.
