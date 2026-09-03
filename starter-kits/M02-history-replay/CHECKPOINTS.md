# M02 checkpoints

## Capability

- [ ] `HistoryRecord` lưu `record_id`, `as_of`, `ingested_at`, `formula_version`, `input_hash`, input snapshot và recorded result.
- [ ] append-only JSONL không overwrite record cũ.
- [ ] exact duplicate idempotent.
- [ ] conflicting duplicate bị reject.
- [ ] corrupt JSONL fail closed.
- [ ] out-of-order record được preserve và query theo `as_of`.
- [ ] replay trả `MATCH | DRIFT | UNREPLAYABLE` đúng semantics.
- [ ] M01 tests vẫn PASS.

## Reality

- [ ] Có ít nhất một stable `product_id` thật từ M00/M01.
- [ ] Có E1 observation t1 và t2 với `observed_at` khác nhau, hoặc blocker được ghi trung thực.
- [ ] Không copy last-known value rồi gọi là observation mới.

## Operated

- [ ] Append history.
- [ ] Restart process và đọc lại record.
- [ ] Chạy duplicate + conflict + out-of-order failure cases.
- [ ] Replay record và rerun để chứng minh deterministic.
- [ ] Explain-back `observed_at != ingested_at != as_of`.

```text
M02 PASS = Capability + Reality + Operated
```
