# Local regression re-verification (2026-09-16)

Trạng thái: **local automated verification / fixture-only / read-only**.

## Run record

- `run_id`: `local-regression-rerun-20260916`
- `repo_commit`: `3ebe32e3aa5157c1c89d34d4cd3b86a94b2f66f5` (`main` sau PR #370)
- `host`: local Darwin arm64
- `n8n`: 2.38.1 disposable engine
- `node`: 24.21.0
- `result`: all corrected commands PASS

## Commands and results

Go modules:

```text
contracts: GOWORK=off go test -count=1 ./... && go vet ./...     PASS
core: GOWORK=off go test -count=1 ./... && go vet ./...          PASS
lab/mission-runtime: GOWORK=off go test -count=1 ./... && go vet ./... PASS
lab/affiliate-bot: GOWORK=off go test -count=1 ./... && go vet ./... PASS
```

`python3 -m unittest discover -s scripts/tests -v`: **113 tests PASS**.

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

Kết quả: tất cả static validator và smoke PASS; BR-12d, BR-13b, BR-16a và
BR-18b giữ continuity, restart/replay, typed backup/restore, budget/lease-window
và durable STOP.

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

## BR-18A runbook command smoke

- `tested_at_utc`: `2026-09-16T04:50Z–04:52Z`
- `runtime_binary`: learner Bot build từ main `65dc4e3`
- `workspace`: disposable directories dưới `/tmp/affiliate-runbook-*.XXXXXX`
- `result`: **PASS** cho chuỗi lệnh runbook local

Chuỗi thực tế đã chạy gồm `mission init`, `history capture`, `history list`,
`history replay`, watcher serve trên loopback, `/healthz`, log tail, process
stop và `mission status`. Kết quả là `APPENDED`, `replay=MATCH`,
`{"status":"OK","execution_permitted":false}`, rồi `stop: true` và
`status: VALID` sau STOP.

Backup/restore standalone cũng chạy trên workspace mới: `backup create` trả
`BACKED_UP` với manifest `affiliate-bot-backup/v3`, `backup restore` trả
`RESTORED`, replay trả `MATCH`, và status sau restore giữ
`stop_reason: "backup-drill"`. Hai lệnh `m11-activate` và `m11-gate` bị chặn
trước khi resolve input, không tạo `m11-artifacts.jsonl`; đây là STOP drill
đúng chủ đích và đã PASS.

Lần chạy đầu tiên phát hiện runbook cũ ghi `watcher.log` trong canonical
runtime, khiến strict backup inventory trả `INPUT_ERROR` vì file không thuộc
artifact graph. Runbook đã được sửa để log/PID của watcher và n8n nằm ngoài
runtime; lần chạy tươi sau sửa đã pass. Đây là lỗi tài liệu đã được khắc phục,
không phải nới verifier hoặc bỏ qua artifact.

## BR-18b M00-M05 backup graph guards

BR-18b smoke tiếp tục tạo các store upstream thật trước khi snapshot: một
`actions.jsonl`, `outcomes.jsonl`, `evaluations.jsonl`, `proposals.jsonl` và
`reviews.jsonl`, nối lần lượt từ canonical history đến synthetic review. Sau
đó smoke tạo năm bản sao backup có checksum manifest hợp lệ nhưng làm orphan
từng liên kết: action→decision, outcome→action, evaluation→outcome,
proposal→evaluation và review→proposal.

Kết quả: cả năm `backup restore` đều trả `VERIFY_FAILED`, target restore không
được publish, còn backup gốc vẫn restore được trong cùng lượt chạy. Đây là
coverage runtime thật cho `m00ToM05BackupFiles` và các loader liên kết; không
được diễn giải thành atomic multi-file/power-loss, distributed lock, provider,
deployment, pilot hoặc business-outcome proof.

## Giới hạn

Các script `validate_*_operated_execution.py` cần execution artifact do n8n
engine tạo nên không chạy như validator độc lập; đường end-to-end đã được
runner engine gọi và PASS. Không có ACCESSTRADE request, Cockpit/provider
credential, live executor, payout, business outcome, target-host deployment,
clean-machine pilot hay distributed/power-loss guarantee. Readiness vẫn là
`NOT_READY_FOR_PRODUCTION`.
