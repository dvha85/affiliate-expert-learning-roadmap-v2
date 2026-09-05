# Starter M11 — Production Closed Loop (vòng production có quản trị)

Làm [bài m11-check từng artifact](../../docs/architecture/M11-JSON-BOUNDARY.md) trước; PASS chưa xác thực chain hoặc cấp quyền production.

Mục tiêu: luyện production lease/health/stop/recovery/closed-loop semantics (ngữ nghĩa lease/sức khỏe/dừng/phục hồi/vòng kín) trước khi bật adapter thật.

## Baseline

```bash
cd lab/mission-runtime
go test ./...
go run ./cmd/demo M11
```

Dùng `CHECKPOINTS.md` và `M11-OPERATED-EVIDENCE-TEMPLATE.md` cho E6. Local sandbox chỉ chứng minh Capability.

## Production activation gate

Chỉ activate production khi:

- M10 E5 đã có review và exact production promotion approval;
- lease hữu hạn, scope/budget/risk nhỏ và RISK1 chỉ khi đã canary E5 đúng risk;
- durable ledger + idempotency + reconciliation đã chứng minh;
- health sources, kill switch và stop path độc lập với Agent;
- telemetry outage không làm mất mandatory audit;
- recovery runbook (sổ tay phục hồi) đã thử;
- ImprovementProposal vẫn `auto_apply=false`;
- không có đường code cho Bot tự renew/widen authority.
