# RP-08 post-merge evidence: M11 terminal-chain mutation proof

- Snapshot date: 2026-09-18
- Implementation PR: #437
- Implementation head: `156331e4363490ef2e08d29bae6820d4104dc74b`
- Merge commit: `b6010508ed058f8edd828e9bacdcf92fd70e0308`
- Status: `NOT_READY_FOR_PRODUCTION`

## Scope

PR #437 adds a disposable-copy mutation runner for the M11 backup/restore
terminal chain. The runner removes the production guard requiring every
persisted M11 evaluation to resolve to exactly one restored `CLOSED`
`PRODUCTION_CYCLE`, then runs the real BR-18b backup/restore smoke. The
mutated implementation must fail closed on the missing-cycle restore and must
not publish the external restore target; an unrelated failure or an
unexpectedly green smoke fails the mutation job.

This closes one RP-08 mutation-coverage gap for the evaluation-to-closed-cycle
reverse cardinality guard. It does not claim complete mutation breadth,
atomic multi-file durability, power-loss/filesystem-crash recovery,
distributed or multi-host safety, Windows traversal parity, provider/live
execution, deployment readiness, or business outcomes.

## Hosted verification

- Curriculum CI run `35362296661`: all 13 checks passed.
- Mission Agent Path CI run `35362296433`: all 13 checks passed.
- The hosted checks included Windows runtime job `105656245515`, learner race
  job `105656245369`, and the required
  `deterministic-smokes-backup-mutations` job `105656245417`.

## Local verification

- `python -m py_compile scripts/mutate_backup_m11_terminal_chain.py scripts/audit_readiness.py scripts/tests/test_audit_readiness.py`: PASS.
- `python scripts/audit_readiness.py`: PASS; audit remains
  `NOT_READY_FOR_PRODUCTION` and resolves 134 scoped claims.
- `python -m unittest scripts.tests.test_audit_readiness`: PASS, 118 tests.
- `git diff --check`: PASS.
- Local `go.exe` is unavailable; hosted Windows runtime and learner race are
  the Go acceptance gates.

## Boundary

The mutation is intentionally bounded to an offline disposable source copy
and the existing synthetic fixture smoke. It proves that this specific
reverse-cardinality enforcement is exercised by CI; it does not prove that
the unchanged implementation survives unmanaged filesystem races, crash or
power-loss seams, atomic multi-file publication, distributed locking,
provider/live execution, deployment, pilot, or target-host readiness.
