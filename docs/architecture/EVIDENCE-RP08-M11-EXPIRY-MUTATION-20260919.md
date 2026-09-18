# RP-08 post-merge evidence: M11 exact lease-expiry mutation proof

## Snapshot

- PR #447 (`test: prove M11 gate rejects exact lease expiry`) was squash-merged
  into `main` at `2538968b6d9d94c7c33c2573d2b2b09e55afb000` from implementation
  head `bc32db664bc2c4ebdc71e3affd2b133c0994ea2c`.
- Status remains `NOT_READY_FOR_PRODUCTION`.

## Scope

The merged regression constructs a checksum-valid canonical M11 gate whose
`EvaluatedAt` is exactly the referenced lease `ExpiresAt`. It recomputes the
lease, health, ledger and gate identities so the case exercises the strict
expiry boundary rather than a forged-ID or unrelated cost-bound failure. The
real graph validator must reject the gate.

The disposable mutation runner copies the relevant Go modules to a temporary
directory and replaces the strict `!evaluatedAt.Before(expiresAt)` guard with a
compile-valid always-false expression. It then requires the real regression to
fail with the dedicated `canonical gate at lease expiry was accepted` marker.
An unexpected green test, compile failure, or unrelated failure is not accepted
as mutation evidence.

## Hosted verification

- Curriculum CI run `35382273805` and Mission Agent Path CI run `35382273774`
  passed all 13 required checks on the exact implementation head before merge.
- The required `deterministic-smokes-backup-mutations` job
  `105720996473` passed the exact-expiry mutation proof.
- Windows runtime job `105720996711` and learner race job `105720996581` passed;
  Mission Agent Path runtime, semantics/blueprints and n8n jobs
  `105720993044`, `105720992766` and `105720993103` also passed.

## Local verification and boundary

- `python scripts/audit_readiness.py` and
  `python -m unittest scripts.tests.test_audit_readiness` pass after the
  post-merge synchronization; the tracker remains `NOT_READY_FOR_PRODUCTION`.
- Python syntax and `git diff --check` pass locally.
- `go.exe` is unavailable in the local worktree; hosted Linux, race and
  Windows jobs provide the executable Go acceptance evidence.

This closes one bounded offline M11 mutation-coverage gap only. It does not
establish trusted external clocks, provider/live execution, deployment
recovery, business outcomes, pilot acceptance, distributed or multi-host
safety, power-loss/filesystem-crash durability, atomic multi-file publication,
Windows traversal parity, or production readiness.

Marker: `Baseline sync after PR #447`.
