# Eval M11 — Production Closed Loop (vòng production có quản trị)

[BR-03c.8b chain audit](../../docs/architecture/M11-CHAIN-AUDIT.md) bổ sung tests bundle, pre/post ledger, cycle/review và resolved-stop. PASS không thay trusted production authority hoặc E6 evidence.

BR-03c.8 thêm [bài kiểm raw JSON từng artifact](../../docs/architecture/M11-JSON-BOUNDARY.md); PASS không thay kiểm toàn chain, STOP/recovery hoặc live evidence.

Eval kiểm production authority (quyền production) chứ không chỉ happy path (đường chạy thuận lợi).

Bắt buộc chứng minh:

- production chỉ chạy dưới finite `ProductionLease` (lease production hữu hạn) đã được người review và bind exact hash;
- `RISK1` chỉ được promote (nâng lên production) nếu approval record xác nhận risk class đó đã có E5;
- `RISK2` luôn rời auto path sang phê duyệt từng hành động;
- health snapshot (ảnh chụp sức khỏe) phải trusted + fresh;
- `DEGRADE` nghĩa read-only/no side effect (chỉ đọc/không tác động), không phải quyền ghi yếu hơn;
- compliance/reconciliation/failure/outcome-age threshold tạo `STOP`;
- `STOP` là sticky: restart không tự mở lại lease cũ;
- total/rate/cost/outcome budget vẫn fail closed;
- cost hint từ Agent/ActionIntent không sở hữu trusted budget input.

Chạy:

```bash
cd lab/mission-runtime
go test ./...
go run ./cmd/demo M11
```

Offline/local sandbox chỉ chứng minh Capability (năng lực). E6 cần production loop thật qua observation window (khoảng quan sát), recovery (phục hồi) và reviewed improvement (cải tiến đã review).
