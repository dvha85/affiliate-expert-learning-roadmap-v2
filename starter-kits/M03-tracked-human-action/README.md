# Bộ khởi đầu — M03

1. Học `curriculum/M03/`.
2. Đọc `CHECKPOINTS.md` và dùng `M03-OPERATED-EVIDENCE-TEMPLATE.md`.
3. Chạy `cd lab/mission-runtime && go test ./...` và `go run ./cmd/demo M03`.
4. Đọc/chạy `evals/M03-tracked-human-action/`.
5. Thực hiện một action nhỏ do **người** làm, ghi ActionRecord rồi OutcomeRecord; kiểm `decision_id/action_id` linkage.
6. Lưu operated evidence dưới `learner/M03/`, không commit dữ liệu nhạy cảm.

PASS cần Capability + Reality + Operated; fixture/CI chỉ chứng minh capability.
