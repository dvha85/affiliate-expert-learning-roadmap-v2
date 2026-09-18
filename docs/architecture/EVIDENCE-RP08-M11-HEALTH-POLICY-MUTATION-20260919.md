# RP-08 post-merge evidence: M11 health-policy mutation proof

- Snapshot date: 2026-09-19
- Implementation PR: #445
- Implementation head: `cce98b4d7ddf0e122da5db5f853ddcd070608066`
- Merge commit: `aa6ead57f68e91d9824f144ba367e4f8b78dd6bd`
- Status: `NOT_READY_FOR_PRODUCTION`

## Scope

PR #445 adds a disposable-copy mutation runner for the M11 production-health
admission policy. The real graph regression constructs a checksum-valid gate
whose health snapshot is `DEGRADED`, recomputes the fixture's canonical gate
identity, and requires production validation to reject it. The mutation
replaces only the `!healthAllowsProduction` admission predicate in a temporary
copy with a compile-preserving always-false expression; the real regression
must then fail with the degraded-health rejection marker. An unexpectedly
green test or an unrelated Go harness failure fails the runner.

This closes one bounded offline mutation-coverage gap. It does not claim
complete M11 mutation breadth, crash or power-loss durability, atomic
multi-file publication, distributed or multi-host locking, Windows traversal
parity, provider/live execution, deployment, business outcomes, pilot
readiness or clean-machine/target-host readiness.

## Hosted verification

- Curriculum CI run `35379166920` and Mission Agent Path CI run `35379166826`
  passed all 13 required hosted checks on the exact implementation head.
- `deterministic-smokes-backup-mutations` job `105710955207` passed,
  including the M11 health-policy mutation.
- Windows runtime job `105710955118` passed.
- Learner race job `105710955837` passed.
- GitHub reported 13/13 checks passed before squash merge.

## Local verification

- `python scripts/audit_readiness.py`: PASS; the tracker remains
  `NOT_READY_FOR_PRODUCTION`.
- `python -m unittest scripts.tests.test_audit_readiness`: PASS (118 tests).
- `python -m py_compile` for the changed Python files: PASS.
- `git diff --check`: PASS.
- Local Go execution remains unavailable because `go.exe` is absent; hosted
  Linux and Windows jobs provide the executable Go acceptance evidence.

## Evidence boundary

The result proves that the tested offline M11 graph regression detects removal
of the production-health admission guard. It does not establish provider or
live execution, deployment recovery, business outcomes, pilot acceptance,
distributed safety, power-loss durability, atomic multi-file publication or
production readiness.
