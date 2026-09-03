# M01 checkpoints

## Capability

- [ ] `go test ./...` PASS.
- [ ] `go vet ./...` PASS.
- [ ] `go run ./cmd/bot` tạo output deterministic.
- [ ] Giải thích được `formula_version` và stable tie-break.
- [ ] Failure cases bảo vệ `0 != missing`, conflict precedence và `real evidence != RECOMMEND`.

## Reality

- [ ] Đã đối chiếu input boundary với M00 Evidence Packet/DecisionPacket.
- [ ] Không fabricate field thật để ép ranking.
- [ ] Nếu thiếu field, ghi đúng `GET_MORE_DATA` + next measurement.

## Operated

- [ ] Đã dự đoán output trước khi chạy ít nhất một case.
- [ ] Đã chạy executable eval pack.
- [ ] Đã ghi một observed failure-case result.
- [ ] Đã explain-back baseline limitation và authority ceiling.

## PASS rule

```text
M01 PASS = Capability + Reality + Operated
```

Không checkbox nào tạo Approval hoặc Execution permission.
