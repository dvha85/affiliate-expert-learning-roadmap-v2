# Bộ khởi đầu — M04

## Thứ tự

1. Học các bài trong `curriculum/M04/`.
2. Chạy `cd lab/mission-runtime && go test ./...`.
3. Chạy `go run ./cmd/demo M04`.
4. Đọc/chạy eval pack `evals/M04-grounded-ai-advisor/`.
5. Lưu operated evidence cá nhân dưới `learner/M04/` (không commit dữ liệu nhạy cảm).

## Checklist + operated evidence

Chạy grounded advisor với evidence thật, thử hallucinated evidence + stale evidence + abstain, ghi model/provider/version nếu dùng LLM thật.

Ghi tối thiểu: predicted result, observed result, một failure case, evidence/reality boundary, limitation, next measurement.
