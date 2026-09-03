# Bộ khởi đầu — M02 Trustworthy History + Replay (lịch sử đáng tin + phát lại)

M02 tiếp tục dùng **một runtime duy nhất** tại `lab/affiliate-bot`; starter kit (bộ khởi đầu) không duplicate implementation (tạo triển khai trùng lặp).

## Dùng theo thứ tự

1. Học `curriculum/M02/M02.1` → `M02.4`.
2. Chạy M01 regression (kiểm thử hồi quy) và M02 eval (đánh giá) tại `lab/affiliate-bot`.
3. Dùng `CHECKPOINTS.md` để kiểm Mission gate (cổng Mission).
4. Copy `M02-OPERATED-EVIDENCE-TEMPLATE.md` vào `learner/M02/` cho evidence (bằng chứng) cá nhân.
5. Executable eval pack (bộ ca đánh giá có thể chạy) nằm ở `evals/M02-history-replay/`.

## Các lệnh

```bash
cd lab/affiliate-bot

go test ./...
go vet ./...

# Synthetic smoke fixture có observation_id + observed_at
go run ./cmd/bot history capture data/history.jsonl data/m02-sample-observations.json demo-1 2026-09-01T01:00:00Z 2026-09-03T10:00:00Z

go run ./cmd/bot history list data/history.jsonl
go run ./cmd/bot history replay data/history.jsonl
```

Khi người học dùng evidence thật, observation input (đầu vào quan sát) phải có stable `product_id` (định danh sản phẩm ổn định), unique `observation_id` (ID quan sát duy nhất) và explicit `observed_at` (mốc quan sát được ghi rõ). `as_of` không được sớm hơn observation mà decision sử dụng.

## Phạm vi

```text
immutable local history (lịch sử cục bộ bất biến)
+ stable observation identity (định danh quan sát ổn định)
+ observed_at / ingested_at / as_of separation (tách ba mốc thời gian)
+ deterministic query (truy vấn tất định)
+ input integrity hash (mã băm toàn vẹn đầu vào)
+ versioned replay (phát lại theo phiên bản)
+ restart proof (bằng chứng qua khởi động lại)
+ KHÔNG có hành động bên ngoài
```

Không thêm database (cơ sở dữ liệu), scheduler (bộ lập lịch), n8n hay Agent chỉ để hoàn thành M02. JSONL là learner implementation (triển khai cho người học) dễ audit (kiểm toán), không phải production storage mandate (quy định bắt buộc về lưu trữ production).
