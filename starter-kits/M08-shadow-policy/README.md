# Bộ khởi đầu — M08

1. Học `curriculum/M08/M08.1 → M08.2 → M08.3`.
2. Đọc `CHECKPOINTS.md` và dùng `M08-OPERATED-EVIDENCE-TEMPLATE.md`.
3. Chạy `cd lab/mission-runtime && go test ./...` và `go run ./cmd/demo M08`.
4. Đọc `evals/M08-shadow-policy/` và dự đoán output trước khi chạy test.
5. Chọn một Decision/Evidence chain thật phù hợp từ Mission trước; tạo **shadow ActionIntent**, không thực thi action.
6. Chạy normal + tampered + expired + missing-link + duplicate/collision cases.
7. Lưu PolicyDecision, intent hash, policy version, expected/observed state và proof `execution_authorized=false` dưới `learner/M08/`.

Không nối write tool/executor. OPA chưa phải yêu cầu M08; Go deterministic baseline đủ cho PASS nếu semantics/audit/replay đạt.
