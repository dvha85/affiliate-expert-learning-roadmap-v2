# M02 checkpoints

## Capability

- [ ] M01 tests vẫn PASS.
- [ ] `HistoryRecord` lưu `record_id`, `as_of`, `ingested_at`, `formula_version`, `input_hash`, immutable input snapshot và recorded result.
- [ ] Snapshot deep-copy input; caller mutation sau capture không sửa history record.
- [ ] M02 observation có `observation_id` + `observed_at`; `as_of` không sớm hơn observation được dùng.
- [ ] append-only JSONL không overwrite record cũ.
- [ ] exact duplicate idempotent.
- [ ] same `record_id` + different content bị reject.
- [ ] same `observation_id` + different content bị reject.
- [ ] corrupt JSONL, invalid timestamp và input-hash mismatch fail closed.
- [ ] out-of-order record được preserve và query theo parsed `as_of`.
- [ ] canonical input hash không phụ thuộc array order.
- [ ] replay trả `MATCH | DRIFT | UNREPLAYABLE` đúng semantics sau integrity gate.
- [ ] executable M02 eval pack có ít nhất 12 case và `go test ./...` thực sự chạy chúng.

## Reality

- [ ] Có ít nhất một stable `product_id` thật từ M00/M01.
- [ ] Có E1 observation t1 và t2 với `observed_at` khác nhau, hoặc blocker được ghi trung thực.
- [ ] Không copy last-known value rồi gọi là observation mới.
- [ ] `UNCHANGED` được chấp nhận nếu đó là kết quả quan sát thật.

## Operated

- [ ] Append ít nhất hai history records.
- [ ] Restart process và đọc lại record.
- [ ] Chạy duplicate + record conflict + observation identity conflict + out-of-order failure cases.
- [ ] Replay record và rerun để chứng minh deterministic.
- [ ] Explain-back `observed_at != ingested_at != as_of`.
- [ ] Explain-back integrity failure khác `DRIFT` như thế nào.
- [ ] Giải thích vì sao `replay MATCH != business truth`.

```text
M02 PASS = Capability + Reality + Operated
```

Không checkpoint nào tạo Approval hoặc Execution permission.
