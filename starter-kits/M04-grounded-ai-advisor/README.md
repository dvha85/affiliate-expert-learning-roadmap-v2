# Bộ khởi đầu — M04

1. Học `curriculum/M04/`.
2. Đọc `CHECKPOINTS.md` và dùng `M04-OPERATED-EVIDENCE-TEMPLATE.md`.
3. Chạy `cd lab/mission-runtime && go test ./...` và `go run ./cmd/demo M04`.
4. Đọc/chạy `evals/M04-grounded-ai-advisor/`.
5. Chạy advisor với evidence thật; thử hallucinated, stale, future evidence, abstain và write request. Ghi model/provider/version nếu dùng LLM thật.
6. Lưu operated evidence dưới `learner/M04/`.

PASS cần Capability + Reality + Operated; model/fixture success không tự tạo business truth.
