# Starter kit — M02 Trustworthy History + Replay

M02 tiếp tục dùng **một runtime duy nhất** tại `lab/affiliate-bot`; starter kit không duplicate implementation.

## Dùng theo thứ tự

1. Học `curriculum/M02/M02.1` → `M02.4`.
2. Chạy M01 regression và M02 eval tại `lab/affiliate-bot`.
3. Dùng `CHECKPOINTS.md` để kiểm Mission gate.
4. Copy `M02-OPERATED-EVIDENCE-TEMPLATE.md` vào `learner/M02/` cho evidence cá nhân.
5. Executable eval pack nằm ở `evals/M02-history-replay/`.

## Commands

```bash
cd lab/affiliate-bot

go test ./...
go vet ./...

# Synthetic smoke fixture có observation_id + observed_at
go run ./cmd/bot history capture data/history.jsonl data/m02-sample-observations.json demo-1 2026-09-01T01:00:00Z 2026-09-03T10:00:00Z

go run ./cmd/bot history list data/history.jsonl
go run ./cmd/bot history replay data/history.jsonl
```

Khi learner dùng evidence thật, observation input phải có stable `product_id`, unique `observation_id` và explicit `observed_at`. `as_of` không được sớm hơn observation mà decision sử dụng.

## Scope

```text
immutable local history
+ stable observation identity
+ observed_at / ingested_at / as_of separation
+ deterministic query
+ input integrity hash
+ versioned replay
+ restart proof
+ NO external action
```

Không thêm database, scheduler, n8n hay Agent chỉ để hoàn thành M02. JSONL là learner implementation dễ audit, không phải production storage mandate.
