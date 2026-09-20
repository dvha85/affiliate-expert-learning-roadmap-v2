# PR #427 post-merge RP-05a evidence — 2026-09-18

## Snapshot

- Implementation PR: [#427](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/427)
- Implementation head before merge: `0f9f3aaee9b0e41afd31eaa448df0c59bd7b0504`
- Squash merge into `main`: `dcda7894ad72227c588de0f7077dc18cdab92fea`
- Scope: persist and revalidate M07 proposal validation metadata, A2-RO
  authority ceiling, and canonical context/record/decision/evidence provenance.

## Hosted verification

PR #427 passed all 13 required checks at the final implementation head.
Curriculum CI run `35338091532` passed the learner shards, learner race,
Windows runtime, Go vet and deterministic smoke jobs. Mission Agent Path CI
run `35338091503` passed mission semantics/blueprints, mission-runtime and the
n8n engine regression. This is the same exact implementation head reviewed
before merge; the initial missing test import was fixed before this final run.

The post-merge readiness re-run remains bounded:

- `python scripts/audit_readiness.py` — `NOT_READY_FOR_PRODUCTION`, 123 scoped
  claims resolved.
- `python -m unittest scripts.tests.test_audit_readiness` — 118 tests, `OK`.
- Local Go test/vet is unavailable because `go.exe` is not installed in this
  worktree; hosted checks are the Go acceptance evidence.

## Boundary

This record rebinds bounded offline/fixture/read-only RP-05a evidence to the
merged `main`. It does not claim provider/model operation, n8n business-path
execution, deployment, business outcomes, distributed locking, filesystem
crash/power-loss durability or multi-file atomicity. RP-05 and the overall
tracker remain `PARTIAL` and `NOT_READY_FOR_PRODUCTION` respectively.

Marker: `Baseline sync after PR #427`.
