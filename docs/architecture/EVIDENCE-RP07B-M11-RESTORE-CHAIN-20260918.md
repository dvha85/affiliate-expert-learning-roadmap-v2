# RP-07b post-merge evidence: restored M11 terminal chain

- Snapshot date: 2026-09-18
- Implementation PR: #435
- Implementation head: `f398c6e5782803eed494fd9a5a609c4aec1320ca`
- Merge commit: `af2eb0d20d10d8f638480b8cdd8141c8b051ec1c`
- Status: `NOT_READY_FOR_PRODUCTION`

## Scope

PR #435 strengthens the real learner Bot BR-18b backup/restore smoke with
explicit terminal-state assertions. After a valid typed-v3 restore, the
positive M11 snapshot must contain the complete production lifecycle, exactly
one `CLOSED` `PRODUCTION_CYCLE`, matching execution/outcome/evaluation/cycle
links, and a `NORMAL` post-ledger with zero pending outcomes. The recovery
restore must also retain the reviewed UNKNOWN-to-STOP reconciliation and its
durable STOP boundary.

These assertions prove the restored fixture reaches the intended RP-07b
terminal chain; they do not create a live executor, provider outcome, or
production recovery authority.

## Hosted verification

- Curriculum CI run `35359232122`: all 13 checks passed.
- Mission Agent Path CI run `35359232042`: all 13 checks passed.
- The hosted checks included Windows runtime, learner race, full Go test/vet,
  backup/restore mutation smoke, and the Mission Agent Path regressions.

## Local verification

- `python -m py_compile scripts/smoke_br18b_backup_restore.py`: PASS.
- `git diff --check`: PASS.
- `python scripts/audit_readiness.py`: PASS after this evidence sync; the
  audit remains `NOT_READY_FOR_PRODUCTION` and resolves 134 scoped claims.
- `python -m unittest scripts.tests.test_audit_readiness`: PASS, 118 tests.
- Local `go.exe` and `gofmt.exe` are unavailable; hosted CI is the Go/Windows
  acceptance gate.

## Boundary

This is bounded offline/fixture/read-only evidence. It does not claim prior
runtime availability or live recovery, power-loss/filesystem-crash durability,
atomic multi-file publication, distributed or multi-host safety,
provider/live execution, deployment readiness, business outcomes, or clean
machine/target-host readiness.
