# RP-05b post-merge evidence — PR #429

## Snapshot

- PR #429 (`feat: bind M07 tool trace provenance`) was squash-merged into
  `main` at `4950b2fd1be80f3c0e58d9f270a44e275c3de0d5` from implementation head
  `d079dfc3b2d229805b458ec95c92be0fd15b375f`.
- The implementation makes M07 `request_id` and `content_digest` adapter-derived,
  binds request identity to the canonical record/tool request, and exposes that
  identity on tool evidence. Canonical JSON digesting keeps the content identity
  stable across `json.RawMessage` serialization and restore.
- Forged metadata and legacy registered traces with incomplete provenance fail
  closed. The flow remains read-only and non-authorizing.

## Hosted verification

- Curriculum CI run `35340989562`: all 10 jobs passed, including
  `deterministic-smokes-m06-m07` job `105586601328`, learner race job
  `105586601291`, and Windows runtime job `105586601032`.
- Mission Agent Path CI run `35340989656`: all 3 jobs passed, including
  mission-runtime job `105586601723`, n8n regression job `105586601742`, and
  semantics/blueprints job `105586601529`.
- These hosted checks exercised Go test/vet, Windows runtime, race regression,
  M07 adversarial/output contracts, mission runtime and n8n paths on the exact
  final head before merge.

## Local verification and boundary

- `python scripts/audit_readiness.py` passed with
  `NOT_READY_FOR_PRODUCTION`; the evidence graph resolves 125 scoped claims.
- `python -m unittest scripts.tests.test_audit_readiness` passed: 118 tests.
- `go.exe` is unavailable in the local worktree; hosted CI is the Go/Windows
  runtime acceptance evidence.
- This is bounded offline/fixture/read-only evidence. It does not establish
  provider/model operation, live execution, deployment, business outcomes,
  distributed locking, power-loss durability or multi-file atomicity.

Marker: `Baseline sync after PR #429`.
