# Backup/restore M11 graph guards (2026-09-16)

Trạng thái: **local automated verification / synthetic fixture / read-only**.

## Run record

- `run_id`: `backup-m11-duplicate-artifact-20260916`
- `implementation_commit`: `90a729689e9201472a3e72a64b775b9d25cc45d2`
- `merge_commit`: `edab3a2db756245ceda57a227366f004c5236771` (PR #368 squash merge)
- `host`: local Darwin arm64
- `scope`: learner Bot backup/restore v3 and M11 canonical artifact registry

## Scenario and result

`smoke_br18b_backup_restore.py` tạo backup từ runtime thật, sao chép nguyên
line `PRODUCTION_EXECUTION_RECORD` vào `m11-artifacts.jsonl`, cập nhật checksum
và kích thước manifest để mutation vẫn hợp lệ ở lớp integrity. Restore bằng
learner Bot thật trả `VERIFY_FAILED` trước khi publish target; thư mục restore
không xuất hiện.

Ca này kiểm tra canonical registry/cardinality guard trên đường backup/restore,
không chỉ gọi trực tiếp core validator.

## Required-artifact absence

Cùng smoke run tạo các bản sao backup checksum-valid nhưng lần lượt bỏ
`PRODUCTION_LEASE`, `PRODUCTION_LEASE_APPROVAL`, `PRODUCTION_HEALTH_SNAPSHOT`,
`TRUSTED_COST_BOUND`, `PRODUCTION_GATE`,
`PRODUCTION_EXECUTION_AUTHORIZATION`, `PRODUCTION_EXECUTION_RECORD` hoặc
`PRODUCTION_OUTCOME_EVALUATION`. Mỗi bản sao bị learner Bot reject với
`VERIFY_FAILED` hoặc `GRAPH_FAILED` trước khi publish target. `PRODUCTION_CYCLE`
không nằm trong nhóm bắt buộc của snapshot lịch sử hiện tại; test không tuyên
bố nó là required artifact.

## Verification

```text
python3 scripts/smoke_br18b_backup_restore.py  PASS
```

## Limits

Đây là bounded local graph/inventory validation. Nó không chứng minh atomic
multi-file crash/power-loss, Windows native locking, distributed/multi-host,
provider, live executor, business outcome, pilot hoặc deployment readiness.
Overall readiness remains `NOT_READY_FOR_PRODUCTION`.
