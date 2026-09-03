# Eval pack — M01 deterministic Bot

`cases.json` là canonical executable eval pack của M01. `lab/affiliate-bot/cmd/bot/main_test.go` đọc trực tiếp file này; không duy trì một bản cases thứ hai trong code.

## Coverage tối thiểu

| Case | Invariant |
|---|---|
| M01-E01 | valid synthetic input → deterministic ranking |
| M01-E02 | real evidence → `RANK_SCENARIO`, không auto recommend |
| M01-E03 | missing price → `GET_MORE_DATA` |
| M01-E04 | `0 != missing` |
| M01-E05 | mixed real/synthetic → `HUMAN_REVIEW` |
| M01-E06 | currency conflict → `HUMAN_REVIEW` |
| M01-E07 | conflict precedence > missing |
| M01-E08 | invalid evidence origin → `GET_MORE_DATA` |

## Chạy

```bash
cd lab/affiliate-bot
go test ./...
```

Eval pack kiểm behavior contract, không chứng minh business formula tốt hay Affiliate outcome thật.
