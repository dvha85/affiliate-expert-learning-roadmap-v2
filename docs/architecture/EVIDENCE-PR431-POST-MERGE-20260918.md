# RP-06 post-merge evidence — PR #431

## Snapshot

- PR #431 (`fix: validate historical M10 backup registries`) was squash-merged
  into `main` at `ad5347bf1aae4c0770e0471b668c94bf8621b141` from implementation
  head `321eb74b91e991ab319c784205d3f587ef0b01bc`.
- Backup and restore now include every existing `m10-artifacts.jsonl` registry
  in the required inventory and validate its semantic graph even when mutable
  mission state has no active canary.
- A valid historical M10 registry remains restorable; a checksum-valid orphaned
  registry is rejected before a restore target is published.

## Hosted verification

- Curriculum CI run `35344493952`: all 10 jobs passed, including backup/restore
  mutation job `105597788342`, learner race job `105597788461`, and Windows
  runtime job `105597788500`.
- Mission Agent Path CI run `35344493969`: all 3 jobs passed, including
  mission-runtime job `105597787824`, n8n regression job `105597787737`, and
  semantics/blueprints job `105597787518`.
- The exact implementation head passed all 13 hosted checks before merge.

## Local verification and boundary

- `python scripts/audit_readiness.py` passed with
  `NOT_READY_FOR_PRODUCTION`; the evidence graph resolves 129 scoped claims.
- `python -m unittest scripts.tests.test_audit_readiness` passed: 118 tests.
- `go.exe` and `gofmt.exe` are unavailable in the local worktree; hosted Go and
  Windows CI is the runtime acceptance evidence.
- This is bounded offline/fixture/read-only evidence. It does not establish
  power-loss or filesystem-crash durability, atomic multi-file transactions,
  distributed/multi-host locking, provider/model operation, live execution,
  deployment or business outcomes.

Marker: `Baseline sync after PR #431`.
