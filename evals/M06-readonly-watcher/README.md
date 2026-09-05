# Eval pack — M06

`cases.json` là executable contract data cho `lab/mission-runtime`.

BR-03c.3 giữ expected NEW/UNCHANGED/CHANGED và rejection cũ; TestM06NormalizerEvalPack chạy thêm normalizer + schema output. Normalizer offline nay dùng synthetic/test và ID bind source/time/method; xem [mapping/giới hạn](../../docs/architecture/M06-JSON-BOUNDARY.md).

```bash
cd lab/mission-runtime
go test ./...
```

Eval PASS không tự tạo Reality/Operated PASS.
