# Eval pack — M05

`cases.json` là executable contract data cho `lab/mission-runtime`.

BR-03c.2 thêm TestM05RawEvalPack dùng schema trước typed semantic. `raw-expectations.json` chỉ ghi khác biệt E02 auto_apply=true và E06 reviewed_by=agent → INVALID_SCHEMA; expected typed cũ không đổi. TestM05ChainSemantics kiểm thêm cả chuỗi/ID trùng/timeline qua boundary mới; xem [M05 JSON boundary](../../docs/architecture/M05-JSON-BOUNDARY.md).

```bash
cd lab/mission-runtime
go test ./...
```

Eval PASS không tự tạo Reality/Operated PASS.
