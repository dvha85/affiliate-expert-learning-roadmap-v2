# PR #425 post-merge RP-04 evidence — 2026-09-18

## Snapshot

- Implementation PR: [#425](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/425)
- Implementation head before merge: `d3127b99dda310be3620ff54ea0b3add0cd6d44b`
- Squash merge into `main`: `b7cfe1480973254ad106481d22c142651af3732b`
- Scope: shared `core/canonical` evidence context for M07/M08/HTTP, with
  strict aggregate-ID linkage and M00 field/provenance re-derivation.

## Hosted verification

PR #425 passed all 13 required checks at the final head. Curriculum CI run
`35329168897` passed the learner shards, learner race, Windows runtime,
Go vet/structure and deterministic smoke jobs. Mission Agent Path CI run
`35329168855` passed mission semantics/blueprints, mission-runtime and the
n8n engine regression. This is the same exact implementation head that was
reviewed before merge.

The local readiness re-run on the post-merge snapshot passes:

- `python scripts/audit_readiness.py` — `NOT_READY_FOR_PRODUCTION`, 121 claims.
- `python -m unittest scripts.tests.test_audit_readiness` — 118 tests, `OK`.
- Local Go test/vet is unavailable because `go.exe` is not installed in this
  worktree; hosted checks are the Go acceptance evidence.

## Boundary

This record only rebinds bounded offline/fixture/read-only evidence to merged
`main`. It does not claim provider/live execution, deployment, business
outcomes, distributed locking, filesystem crash/power-loss durability or
multi-file atomicity. RP-04 and the overall tracker remain `PARTIAL` and
`NOT_READY_FOR_PRODUCTION` respectively.

Marker: `Baseline sync after PR #425`.
