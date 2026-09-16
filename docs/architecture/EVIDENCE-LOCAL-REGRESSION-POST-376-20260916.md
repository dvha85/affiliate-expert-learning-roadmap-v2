# Local regression after PR #376 (2026-09-16)

Trạng thái: **local automated verification / fixture-only / read-only**.

## Run record

- `run_id`: `local-regression-post-376-20260916`
- `repo_commit`: `5bca64f88c1050779ed37882267a1d65d6f5223a` (`main` sau PR #376)
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
```

BR-18b lần này bao gồm ca checksum/manifest hợp lệ nhưng thiếu
`PRODUCTION_CYCLE` trong khi evaluation vẫn tồn tại; restore trả
`GRAPH_FAILED` và không publish target. Snapshot recovery-only không có
evaluation vẫn restore được.

Disposable n8n cũng PASS:

```text
python3 scripts/run_n8n_engine_regression.py                       PASS
python3 scripts/run_n8n_m06_schedule_regression.py                 PASS
```

Engine chứng minh M06/M07 fixture append/replay, exact retry và fail-closed
grounding/tool policy qua loopback; Schedule Trigger append một lần, retry
idempotent sau restart và chặn khi adapter unavailable. Không có request tới
ACCESSTRADE/provider và không có business outcome.

Readiness audit:

```text
python3 scripts/audit_readiness.py
READINESS AUDIT: NOT_READY_FOR_PRODUCTION
```

## Scope and limits

Record này cập nhật baseline kiểm chứng local sau PR #376, không đóng các
evidence ngoài repo: provider/Cockpit operated proof, live executor, business
outcome/payout, clean-machine pilot, target-host deployment/recovery,
distributed locking hoặc power-loss/atomic multi-file durability.
