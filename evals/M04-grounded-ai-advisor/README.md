# Eval pack — M04

`cases.json` là executable contract data cho `lab/mission-runtime`.

```bash
cd lab/mission-runtime
go test ./...
```

Eval PASS không tự tạo Reality/Operated PASS.

E07–E12 thêm hồi quy BR-03a. Eval giữ `output` dưới dạng JSON gốc và gọi `DecodeAdvisorOutput` trước `EvaluateAdvisorOutput`, để field thiếu/lạ/null không bị Go zero value che khuất. Tests trong `advisor_boundary_test.go` còn kiểm duplicate key, sai kiểu, trailing JSON, context timestamp/source và output được serialize từ lệnh `advisor-check`.
