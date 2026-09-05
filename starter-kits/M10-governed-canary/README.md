# Starter M10 — Governed Canary (canary có quản trị)

Sau kiểm từng artifact, làm [bài audit chain M10](../../docs/architecture/M10-CHAIN-AUDIT.md): chín file với pre-gate ledger; PASS vẫn không cấp quyền execute.

Bài kiểm file offline: [BR-03c.7 M10 JSON boundary](../../docs/architecture/M10-JSON-BOUNDARY.md). PASS từng artifact không xác thực grant, không kiểm toàn chain và không cho phép execute.

Mục tiêu: luyện grant/gate/ledger/executor semantics (ngữ nghĩa grant/cổng/sổ theo dõi/bộ thực thi) trước khi chọn một live adapter (adapter thật) có impact thấp.

## Chạy baseline

```bash
cd lab/mission-runtime
go test ./...
go run ./cmd/demo M10
```

Sau baseline, dùng `CHECKPOINTS.md` và `M10-OPERATED-EVIDENCE-TEMPLATE.md` để chuẩn bị E5. Không dùng sandbox output để claim E5.

## Live adapter gate

Chỉ cấu hình adapter thật khi:

- target/account/action đã được người review;
- credential least privilege (quyền tối thiểu);
- provider terms/compliance (điều khoản/tuân thủ) đã xem;
- atomic idempotency/budget reservation chứng minh được;
- kill switch chạy được ngoài Agent;
- first grant rất nhỏ, ưu tiên RISK0/cost 0;
- outcome source và reconciliation procedure (quy trình đối soát) đã rõ.
