# Bộ khởi đầu — M09

[m09-check offline](../../docs/architecture/M09-JSON-BOUNDARY.md) kiểm năm file, không authenticate approval/execute. Giữ bằng chứng consistency riêng khỏi trusted human approval và vận hành thật.

1. Học `curriculum/M09/M09.1 → M09.2 → M09.3`.
2. Đọc `CHECKPOINTS.md` và dùng `M09-OPERATED-EVIDENCE-TEMPLATE.md`.
3. Chạy `cd lab/mission-runtime && go test ./...`.
4. Chạy `go run ./cmd/demo M09`; demo chỉ dùng `local_sandbox` bị giới hạn.
5. Tạo ApprovalRecord do **người** quyết định; không để Agent/orchestrator tự approve.
6. Persist state, restart/load lại và revalidate trước execute.
7. Chạy ca mismatch hash/policy, expired approval, kill switch, wrong executor/target và duplicate.
8. Chỉ khi bạn chủ động cấu hình adapter live: dùng credential least-privilege, target allowlist, compliance review và rollback/recovery phù hợp.

Không cần Temporal/OPA để PASS M09 nếu baseline deterministic + persisted state đã giải quyết use case.
