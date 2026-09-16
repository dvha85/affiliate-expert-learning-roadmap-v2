# Local regression after PR #377 (2026-09-16)

Trạng thái: **local automated verification / fixture-only / read-only**.

## Run record

- `run_id`: `local-regression-post-377-20260916`
- `repo_commit`: `c11880b19be8b979f90f6dcc1d21cc8bd032aaa5` (`main` sau PR #377)
- `host`: local Darwin arm64
- `n8n`: 2.38.1 disposable engine
- `node`: 24.21.0
- `result`: tất cả lệnh trong record PASS

## Commands and results

Bốn Go module đều PASS với test/vet không dùng workspace:

```text
(cd contracts && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...)          PASS
(cd core && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...)               PASS
(cd lab/affiliate-bot && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...)  PASS
(cd lab/mission-runtime && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...) PASS
```

Các regression Python và static contract đều PASS:

```text
python3 -m unittest discover -s scripts/tests -v                  PASS (113 tests)
scripts/validate_missions.py                                      PASS
scripts/validate_repo.py                                          PASS
scripts/validate_artifact_spine.py                                PASS
scripts/validate_continuity.py                                    PASS
scripts/validate_language_policy.py                               PASS
scripts/validate_agent_semantics.py                               PASS
scripts/validate_semantic_contracts.py                             PASS
scripts/validate_m11.py                                           PASS
scripts/validate_n8n_m06.py                                       PASS
scripts/validate_n8n_m06_selected_source.py                       PASS
scripts/validate_n8n_m06_cases.py                                 PASS
scripts/validate_n8n_m07.py                                       PASS
scripts/validate_n8n_m07_adversarial.py                           PASS
scripts/validate_n8n_m07_output_cases.py                          PASS
python3 scripts/smoke_br16a_offline.py                            PASS
python3 scripts/smoke_br18b_backup_restore.py                     PASS
python3 scripts/audit_readiness.py                                PASS (NOT_READY_FOR_PRODUCTION)
```

Disposable n8n cũng PASS:

```text
python3 scripts/run_n8n_engine_regression.py \
  --n8n-cli /Users/hadinh/.local/n8n-2.38.1/node_modules/n8n/bin/n8n \
  --n8n-node /Users/hadinh/.local/node-v24.21.0-darwin-arm64/bin/node       PASS
python3 scripts/run_n8n_m06_schedule_regression.py \
  --n8n-cli /Users/hadinh/.local/n8n-2.38.1/node_modules/n8n/bin/n8n \
  --n8n-node /Users/hadinh/.local/node-v24.21.0-darwin-arm64/bin/node       PASS
```

Engine regression chứng minh đường M06/M07 fixture và selected-source metadata
đã sanitized qua n8n, exact retry và fail-closed policy/grounding/tool-result
reject. Schedule Trigger append đúng một lần, retry idempotent trước/sau n8n và
adapter restart, rồi fail-closed khi loopback adapter unavailable.

## Scope and limits

Kết quả chỉ là evidence local fixture/read-only từ implementation thật. Không có
request tới ACCESSTRADE/provider và không có business outcome, payout hay live
executor. Record này không đóng provider/Cockpit operated proof, clean-machine
pilot, target-host deployment/recovery, distributed locking hoặc power-loss/
atomic multi-file durability. Readiness vẫn là `NOT_READY_FOR_PRODUCTION`.
