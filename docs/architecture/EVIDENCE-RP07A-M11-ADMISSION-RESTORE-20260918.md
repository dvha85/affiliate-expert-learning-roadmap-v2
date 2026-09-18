# RP-07a post-merge evidence: M11 recovery admission backup/restore

- Snapshot date: 2026-09-18
- Implementation PR: #433
- Implementation head: `e406e405b43c954353c6fe58c21f096f28a5ea1f`
- Merge commit: `c8fe14218ef2a22f6faf0cbd1c9a335c1073dffb`
- Status: `NOT_READY_FOR_PRODUCTION`

## Scope

PR #433 adds a real learner Bot backup/restore regression for the M11
`ProductionRecoveryAdmission` graph. A valid typed-v3 backup preserves the
admission and its new lease, immutable approval, activation and NORMAL ledger
links. A checksum-valid mutation that removes the immutable new-runtime
approval is rejected before a restore target is published. Depending on the
verification stage, the rejected restore may return `VERIFY_FAILED` before
`GRAPH_FAILED`; the contract is fail closed and no target publication occurs.

The fixture seeds a canonical history record so the required inventory is
complete. `PriorRuntimeDir` remains an audit/handoff reference only; this
evidence does not establish prior-runtime availability or live recovery.

## Hosted verification

- Curriculum CI run `35350741433`: all 13 checks passed.
- Mission Agent Path CI run `35350741436`: all 13 checks passed.
- Explicit acceptance jobs included Windows runtime `105618004234`, learner
  race `105618003984`, and backup/restore shard-1 `105618003825`.
- The hosted checks exercised `go test ./...`, `go vet ./...`,
  `go test -race ./...`, backup/restore smoke, mission runtime and n8n
  regressions.

## Local verification

- `python scripts/audit_readiness.py`: PASS, remains
  `NOT_READY_FOR_PRODUCTION`, and resolves 132 scoped claims after this
  evidence graph update.
- `python -m unittest scripts.tests.test_audit_readiness`: PASS, 118 tests.
- Local `go.exe` and `gofmt.exe` are unavailable; hosted CI is the Go/Windows
  acceptance gate.

## Boundary

This is bounded offline/fixture/read-only evidence. It does not claim
prior-runtime availability or live recovery, power-loss/filesystem-crash
durability, atomic multi-file publication, distributed or multi-host safety,
provider/live execution, deployment readiness or business outcomes.
