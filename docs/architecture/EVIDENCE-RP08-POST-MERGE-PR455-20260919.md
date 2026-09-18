# RP-08 post-merge evidence: M11 reverse-ledger graph mutation proof

- Snapshot date: 2026-09-19
- Implementation PR: #455
- Implementation head: `872f9f67a4993bed9cdeaaf38ede685604f10435`
- Merge commit: `8346d676ee9fbacd364025cb9f28f9cac1c40216`
- Status: `NOT_READY_FOR_PRODUCTION`

## Hosted verification

PR #455 passed all 13 required Curriculum CI and Mission Agent Path CI checks
on the exact implementation head. Curriculum CI run `35395007915` passed
Windows runtime job `105761719057`, learner race job `105761719174` and the
deterministic backup/mutation job `105761719050`, including the new
`scripts/mutate_m11_reverse_ledger_graph.py` step. Mission Agent Path CI run
`35395007875` passed all four hosted checks.

The mutation removes only the core M11 reverse-ledger graph guard in a
disposable copy and requires
`TestArtifactGraphAcceptsAndRejectsExactProductionLifecycleLinks` to fail at
the named orphan reconciliation resolution assertion. An unrelated failure or
an unexpectedly green mutation fails the runner.

## Local verification

- `python scripts/audit_readiness.py`: PASS; 145 scoped claims;
  `NOT_READY_FOR_PRODUCTION` remains unchanged.
- `python -m unittest scripts.tests.test_audit_readiness`: PASS (120 tests).
- Python compilation, readiness JSON parsing and `git diff --check`: PASS.
- Local Go execution remains unavailable because `go.exe` is absent; hosted
  CI is the executable Go and Windows evidence.

## Boundary

This records bounded offline/fixture mutation evidence only. It does not claim
complete M11 mutation breadth, ledger durability, crash or power-loss
recovery, atomic multi-file publication, distributed or multi-host locking,
Windows traversal parity, provider/live execution, deployment, pilot
acceptance, business outcome or production readiness.
