# Eval pack — M02 History + Replay

`cases.json` mô tả failure/behavior cases canonical của M02. Go tests đọc trực tiếp pack này.

Coverage:

- first append;
- exact duplicate idempotency;
- conflicting duplicate rejection;
- out-of-order query ordering;
- replay MATCH;
- replay DRIFT;
- unknown formula → UNREPLAYABLE;
- input-hash mismatch → integrity failure;
- corrupt JSONL → fail closed.

Chạy:

```bash
cd lab/affiliate-bot
go test ./...
```

Synthetic replay PASS không thay E1 Reality gate.
