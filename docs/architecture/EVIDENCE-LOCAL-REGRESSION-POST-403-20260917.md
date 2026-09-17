# Local regression after PR #403 — 2026-09-17

## Baseline and scope

PR #403 was squash-merged into `main` at
`8b44011465ea0c8afcfcc832b76f045da0811bd4`. This maintainer-run regression
used disposable local workspaces, synthetic fixtures, and read-only/loopback
paths only. It did not use provider credentials, call an affiliate endpoint,
invoke a live executor, or establish business-outcome, clean-machine pilot,
deployment, multi-host, or power-loss evidence.

The M07 raw-JSON transport change was also verified by the remote PR checks:
all 12 checks passed on the tested head `7747d6612353b5762f5cecd6735685ed9e9ec5c6`,
including the Node 24 / n8n 2.38.1 engine regression. The current macOS
worktree has no Node or n8n executable, so no local engine PASS is claimed.

## Commands and results

All commands below passed on the merged `main` worktree:

- `GOWORK=off go test -count=1 ./...` and `GOWORK=off go vet ./...` in each
  of `contracts`, `core`, `lab/affiliate-bot`, and `lab/mission-runtime`.
- `python3 -m unittest discover -s scripts/tests -v` — 122 tests, PASS.
- Static mission, repository, artifact-spine, continuity, language, agent,
  semantic-contract, M11, M06 and M07 validators — PASS.
- `python3 scripts/smoke_br12d.py` — M00→M05 continuity and isolated
  fail/pass/rollback, PASS.
- `python3 scripts/smoke_br13b.py` — fixture-to-canonical-history handoff,
  restart/replay, retry and failed-sink behavior, PASS.
- `python3 scripts/smoke_br16a_offline.py` — shared M00→M11 runtime,
  restart/replay, backup/restore and durable STOP, PASS.
- `python3 scripts/smoke_br18b_backup_restore.py` — typed-v3 inventory,
  graph/lineage guards, restart and durable STOP, PASS.
- `python3 -m unittest scripts.tests.test_audit_readiness -q` — 107 tests,
  PASS.
- `python3 scripts/audit_readiness.py` —
  `NOT_READY_FOR_PRODUCTION`, 79 scoped claims resolved.
- `git diff --check` — PASS.

## Interpretation

The merged head preserves the bounded M07 exact-number transport and the
offline learner/runtime regression suite. Overall readiness remains
`NOT_READY_FOR_PRODUCTION`; provider operation, live execution, business
outcome, pilot, deployment, distributed coordination, and stronger filesystem
crash/power-loss claims remain open.
