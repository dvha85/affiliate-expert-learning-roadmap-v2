# Eval pack — M07

`cases.json` là executable contract data cho `lab/mission-runtime`.

BR-03c.4 thêm TestM07RawEval: E02/E04 registry rỗng, E03 registry ghi → INVALID_SCHEMA theo canonical registry. Expected typed cũ giữ nguyên. Các tests raw riêng vẫn kiểm hallucinated ID với registry hợp lệ và ABSTAIN write. [Phạm vi](../../docs/architecture/M07-JSON-BOUNDARY.md).

```bash
cd lab/mission-runtime
go test ./...
```

Eval PASS không tự tạo Reality/Operated PASS.
