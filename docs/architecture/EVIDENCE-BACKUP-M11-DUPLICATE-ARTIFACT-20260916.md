# Backup/restore M11 duplicate-artifact guard (2026-09-16)

Trạng thái: **local automated verification / synthetic fixture / read-only**.

## Run record

- `run_id`: `backup-m11-duplicate-artifact-20260916`
- `implementation_commit`: `9dcd2ef6ab5bd811c15ad1ceb5319be296b1b60a`
- `merge_commit`: `6edd244a164e4e7e561812852f47bc3e40b56b0f` (PR #366 squash merge)
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

## Verification

```text
python3 scripts/smoke_br18b_backup_restore.py  PASS
```

## Limits

Đây là bounded local graph/inventory validation. Nó không chứng minh atomic
multi-file crash/power-loss, Windows native locking, distributed/multi-host,
provider, live executor, business outcome, pilot hoặc deployment readiness.
Overall readiness remains `NOT_READY_FOR_PRODUCTION`.
