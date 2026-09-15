# Local full regression record — 2026-09-16 — main `c56cf95`

## Scope

This maintainer-run verification exercised the checked-in implementation at
`main` commit `c56cf95469fb9cb95662ac646941bbb7ff27600`. It used disposable
local workspaces and fixture/loopback adapters only. It did not use
ACCESSTRADE credentials, call an affiliate endpoint, invoke a live executor,
or establish business-outcome, clean-machine pilot, deployment, multi-host, or
power-loss evidence.

## Environment

- OS: local macOS workspace
- n8n: `2.38.1` (disposable runtime)
- Node.js: `v24.21.0`
- n8n executable: `/Users/hadinh/.local/n8n-2.38.1/node_modules/.bin/n8n`
- n8n engine runs used the explicit Node 24 executable and disposable SQLite

## Commands and results

| Check | Result |
|---|---|
| `GOWORK=off go test -count=1 ./...` and `GOWORK=off go vet ./...` in each of `contracts`, `core`, `lab/affiliate-bot`, and `lab/mission-runtime` | PASS |
| `python3 -m unittest discover -s scripts/tests -v` | PASS; 104 tests |
| M06/M07 static, adversarial, and output validators | PASS |
| `python3 scripts/smoke_br16a_offline.py --workspace <disposable>` | PASS; one shared M00→M11 chain, restart/replay, backup/restore and durable STOP |
| `python3 scripts/smoke_br18b_backup_restore.py --workspace <disposable>` | PASS; runtime-created typed v3 graph, restore/replay, lease/budget and STOP checks |
| `python3 scripts/run_n8n_engine_regression.py --n8n-cli ... --n8n-node ...` | PASS; M06/M07 persistence, grounding and fail-closed rejection |
| `python3 scripts/run_n8n_m06_schedule_regression.py --n8n-cli ... --n8n-node ...` | PASS; append once, exact duplicate before/after restart, adapter outage rejected |
| `python3 scripts/audit_readiness.py` | `NOT_READY_FOR_PRODUCTION` |
| `git diff --check` | PASS |

## Interpretation

All declared local/offline regressions were green at `c56cf95`. The audit
correctly keeps the repository `NOT_READY_FOR_PRODUCTION`. Provider-operated
capture, live executor and business outcomes, clean-machine self-service pilot,
target-host deployment/recovery, distributed coordination, and crash/power-loss
guarantees remain outside this record.
