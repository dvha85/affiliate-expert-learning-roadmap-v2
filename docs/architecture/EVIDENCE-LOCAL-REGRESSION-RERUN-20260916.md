# Local regression re-verification (2026-09-16)

Trạng thái: **local automated verification / fixture-only / read-only**.

## Run record

- `run_id`: `local-regression-rerun-20260916`
- `repo_commit`: `3895cff` (`main` trước khi ghi evidence)
- `host`: local Darwin arm64
- `n8n`: 2.38.1 disposable engine
- `node`: 24.21.0
- `result`: all executed commands PASS

## Commands and results

Go modules:

```text
core: env GOWORK=off go test -race ./...                         PASS
contracts: env GOWORK=off go test ./...                          PASS
lab/mission-runtime: env GOWORK=off go test ./...                PASS
lab/affiliate-bot: env GOWORK=off go test -race ./...            PASS
```

Repository and semantic validators:

```text
python3 scripts/validate_repo.py
python3 scripts/validate_continuity.py
python3 scripts/validate_artifact_spine.py
python3 scripts/validate_language_policy.py
python3 scripts/validate_missions.py
python3 scripts/validate_agent_semantics.py
python3 scripts/validate_semantic_contracts.py
python3 scripts/validate_m11.py
python3 scripts/validate_n8n_m06.py
python3 scripts/validate_n8n_m06_cases.py
python3 scripts/validate_n8n_m06_selected_source.py
python3 scripts/validate_n8n_m07.py
python3 scripts/validate_n8n_m07_adversarial.py
python3 scripts/validate_n8n_m07_output_cases.py
python3 scripts/smoke_br16a_offline.py
python3 scripts/smoke_br18b_backup_restore.py
```

Kết quả: tất cả PASS; `BR-16a` giữ chung artifact lineage, restart/replay và
durable STOP; `BR-18b` giữ typed backup/restore, budget/lease-window và STOP.

Disposable n8n engine:

```text
python3 scripts/run_n8n_engine_regression.py \
  --n8n-cli /Users/hadinh/.local/n8n-2.38.1/node_modules/n8n/bin/n8n \
  --n8n-node /Users/hadinh/.local/node-v24.21.0-darwin-arm64/bin/node
```

Kết quả: `N8N ENGINE REGRESSION PASS`. M06 synthetic và sanitized selected
source append/replay đúng, reject fail-closed; M07 policy và forged
commission reject; Agent chỉ persist grounded canonical output qua model stub
loopback. Không có provider request hay business outcome.

Schedule Trigger:

```text
python3 scripts/run_n8n_m06_schedule_regression.py \
  --n8n-cli /Users/hadinh/.local/n8n-2.38.1/node_modules/n8n/bin/n8n \
  --n8n-node /Users/hadinh/.local/node-v24.21.0-darwin-arm64/bin/node
```

Kết quả: một `APPENDED`, các retry là `EXACT_DUPLICATE` sau n8n/adapter
restart, và adapter unavailable bị chặn trước ACK/report. Đây là loopback
fixture evidence; không phải deployment hoặc provider proof.

## Giới hạn

Các script `validate_*_operated_execution.py` cần execution artifact do n8n
engine tạo nên không chạy như validator độc lập; đường end-to-end đã được
runner engine gọi và PASS. Không có ACCESSTRADE request, Cockpit/provider
credential, live executor, payout, business outcome, target-host deployment,
clean-machine pilot hay distributed/power-loss guarantee. Readiness vẫn là
`NOT_READY_FOR_PRODUCTION`.
