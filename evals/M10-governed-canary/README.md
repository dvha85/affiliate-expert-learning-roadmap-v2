# Eval M10 — Governed Canary (canary có quản trị)

Chạy `go test ./cmd/demo -run TestM10Chain -count=1` từ mission-runtime để kiểm [chain audit](../../docs/architecture/M10-CHAIN-AUDIT.md), gồm trường hợp từng file hợp lệ nhưng chain sai. Không thay trusted runtime hoặc E5 evidence.

BR-03c.7 thêm [raw artifact tests và giới hạn](../../docs/architecture/M10-JSON-BOUNDARY.md). PASS từng file không thay eval gate, trusted ledger, hoặc live evidence.

Eval này kiểm deterministic gate (cổng tất định) trước bounded auto-action (hành động tự động bị giới hạn).

Các nhóm ca bắt buộc:

- `RISK0` và `RISK1` chỉ được `ALLOW_CANARY` khi grant hợp lệ và đúng scope (phạm vi);
- `RISK2` luôn rời auto path (đường tự động) sang `REQUIRE_APPROVAL`;
- grant hết hạn/revoke (thu hồi), hash bị đổi hoặc provenance (nguồn gốc phê duyệt) không resolve phải fail closed;
- budget/rate/cost/outcome backpressure (ngân sách/tần suất/chi phí/áp lực chờ kết quả) phải chặn tăng exposure (mức phơi nhiễm);
- executor, idempotency (chống lặp), reconciliation (đối soát) và policy revalidation (kiểm lại chính sách) phải giữ authority boundary.

Chạy:

```bash
cd lab/mission-runtime
go test ./...
go run ./cmd/demo M10
```

Eval offline chỉ chứng minh Capability (năng lực). E5 cần governed canary thật với external side effect (tác động bên ngoài) bị giới hạn, audit, kill switch và outcome observation (quan sát kết quả).
