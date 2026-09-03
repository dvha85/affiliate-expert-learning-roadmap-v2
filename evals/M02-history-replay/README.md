# Eval pack — M02 Trustworthy History + Replay

`cases.json` là canonical **executable eval pack** của M02. `lab/affiliate-bot/cmd/bot/history_test.go` đọc payload và thực thi behavior từng case; không chỉ kiểm tên/metadata.

## Coverage

| Case | Invariant |
|---|---|
| M02-E01 | out-of-order append vẫn query theo `as_of` |
| M02-E02 | exact duplicate là idempotent |
| M02-E03 | same `record_id` + khác content → conflict |
| M02-E04 | replay cùng version/input/result → `MATCH` |
| M02-E05 | recorded result khác replay result → `DRIFT` |
| M02-E06 | unknown formula version → `UNREPLAYABLE` |
| M02-E07 | input snapshot bị sửa → integrity failure |
| M02-E08 | corrupt JSONL → fail closed |
| M02-E09 | invalid `as_of` → reject |
| M02-E10 | canonical input hash không phụ thuộc array order |
| M02-E11 | `as_of < observed_at` → reject |
| M02-E12 | reuse `observation_id` với content khác → conflict |

Go unit tests ngoài pack còn bảo vệ deep-copy snapshot và restart/read persistence.

## Chạy

```bash
cd lab/affiliate-bot
go test ./...
go vet ./...
```

Synthetic eval PASS chỉ chứng minh deterministic history/replay behavior. Nó không thay E1 Reality gate của learner M02 PASS.
