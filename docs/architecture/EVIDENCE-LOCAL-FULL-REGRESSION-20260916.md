# Local full regression record — 2026-09-16

## Scope

This is a maintainer-run, local/offline verification of the checked-in
implementation at main commit `6103eb85034d7807f5ab44d00570a8355a07eb55`.
It does not use ACCESSTRADE credentials, call an affiliate endpoint, invoke a
live executor, or establish business-outcome, clean-machine pilot,
deployment, multi-host, or power-loss evidence.

## Environment

- OS: local macOS workspace
- n8n: `2.38.1` (disposable runtime)
- Node.js: `v24.21.0`
- n8n executable: `/Users/hadinh/.local/n8n-2.38.1/node_modules/.bin/n8n`
- all n8n runs used the explicit Node 24 executable and a disposable SQLite
  runtime; no checked-in workflow or user credential was modified

## Commands and results

| Check | Result |
|---|---|
| `GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...` in `contracts` | PASS |
| same command in `core` | PASS; M00, M03–M11 packages passed |
| same command in `lab/affiliate-bot` | PASS; learner Bot and internal packages passed |
| same command in `lab/mission-runtime` | PASS |
| `python3 -m unittest discover -s scripts/tests -v` | PASS; 104 tests |
| `python3 scripts/validate_n8n_m06.py` | PASS |
| `python3 scripts/validate_n8n_m06_selected_source.py` | PASS |
| `python3 scripts/validate_n8n_m06_cases.py` | PASS |
| `python3 scripts/validate_n8n_m07.py` | PASS |
| `python3 scripts/validate_n8n_m07_adversarial.py` | PASS |
| `python3 scripts/validate_n8n_m07_output_cases.py` | PASS |
| `python3 scripts/smoke_br16a_offline.py` | PASS; shared M00→M11 fixture chain, restore/restart and durable STOP |
| `python3 scripts/smoke_br18b_backup_restore.py` | PASS; runtime-created typed v3 backup/restore graph and STOP checks |
| `python3 scripts/run_n8n_engine_regression.py --n8n-cli ... --n8n-node ...` | PASS; n8n M06/M07 synthetic and sanitized selected-source paths, exact retry and fail-closed rejection |
| `python3 scripts/run_n8n_m06_schedule_regression.py --n8n-cli ... --n8n-node ...` | PASS; Schedule Trigger appended once, exact-deduplicated across restart, adapter outage failed closed |
| `python3 scripts/audit_readiness.py` | `NOT_READY_FOR_PRODUCTION` |
| `git diff --check` | PASS |

## Interpretation

The local implementation and its declared offline regressions are green at
the recorded commit. The readiness conclusion remains
`NOT_READY_FOR_PRODUCTION`: selected-source/provider operation, live business
outcomes, clean-machine self-service pilot, target-host deployment/recovery,
multi-host coordination, and crash/power-loss guarantees are still outside
this record.
