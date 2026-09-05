# Bộ khởi đầu — M05

1. Học `curriculum/M05/`.
2. Đọc `CHECKPOINTS.md` và dùng `M05-OPERATED-EVIDENCE-TEMPLATE.md`.
3. Chạy `cd lab/mission-runtime && go test ./...` và `go run ./cmd/demo M05`.
4. Đọc/chạy `evals/M05-reviewed-improvement/`.
5. Tạo `EvaluationRecord → ImprovementProposal → ReviewRecord` có linkage thật, version/risk/rollback; không auto-apply.
6. Lưu operated evidence dưới `learner/M05/`.

PASS cần Capability + Reality + Operated và proposal/review không được mồ côi provenance.

Kiểm 5 file qua [m05-check](../../docs/architecture/M05-JSON-BOUNDARY.md) trước bàn giao. Tool không lưu store hoặc xác thực người review; APPROVE_FOR_MANUAL_CHANGE không tự áp dụng code, và REQUEST_CHANGES/REJECT vẫn là review record hợp lệ.
