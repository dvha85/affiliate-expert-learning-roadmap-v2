# RP-08 post-merge evidence: M11 ledger activation mutation proof

## Snapshot

- PR [#451](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/451)
  was squash-merged into `main` at
  `f34d8c838bc7a2154e0555c5e12da7c2e06b89ec` from implementation head
  `810a4bafc5fb58774c7ca1bfa505c75421bfc590`.
- Scope: bounded RP-08 offline mutation coverage and readiness bookkeeping;
  no production runtime source was changed.

## Mutation coverage

The disposable runner `scripts/mutate_m11_ledger_activation.py` copies `core`
and `contracts` to a temporary directory, removes only the M11
ledger-to-activation timing predicate, and runs the real
`TestArtifactGraphAcceptsAndRejectsExactProductionLifecycleLinks` regression.
It requires the pre-activation ledger assertion to fail and rejects compile
failures, an unexpectedly green test, or an unrelated failure. The final
mutation keeps the unused parsed timestamp referenced while making only the
target predicate always false in the disposable copy.

## Hosted verification

- Curriculum CI run `35388186329` and Mission Agent Path CI run `35388186527`
  passed all 13 required checks on the exact implementation head before merge.
- The required `deterministic-smokes-backup-mutations` job
  `105739978015` passed the M11 ledger activation mutation proof.
- Windows runtime job `105739977805` and learner race job `105739978038` passed.
- The final exact-head checks included the fixed mutation harness after two
  CI-observed defects were corrected before merge.

## Local verification and boundary

- `python scripts/audit_readiness.py` passes after this synchronization with
  `NOT_READY_FOR_PRODUCTION` and 143 scoped claims.
- `python -m unittest scripts.tests.test_audit_readiness` passes with 118
  tests; Python syntax and `git diff --check` also pass.
- `go.exe` is unavailable in the local worktree; hosted Linux, race and
  Windows jobs provide the executable Go acceptance evidence.

This closes one bounded offline M11 mutation-coverage gap only. It does not
establish complete mutation breadth, trusted external clocks, provider/live
execution, deployment recovery, business outcomes, pilot acceptance,
distributed or multi-host safety, power-loss/filesystem-crash durability,
atomic multi-file publication, Windows traversal parity, or production
readiness.

Marker: `Baseline sync after PR #451`.
