# BR-17c — M10 ledger/cost/outcome boundary

M10 chỉ được coi canary hợp lệ khi grant, trusted cost và reservation cùng
quyền sở hữu ledger đã được kiểm tra. Mọi kết quả phải liên kết tới
`EffectRef` của action đã được cấp quyền; thiếu/đổi effect, duplicate, unknown
usage hoặc reconciliation chưa đủ đều giữ `WAIT/REQUIRE_APPROVAL`, không tự
chuyển thành success.

Các invariant lab gồm budget tổng/per-action, atomic reservation, sticky stop,
grant expiry/revoke, duplicate idempotency, missing result và outcome
backpressure. Đây là lab/schema evidence; chưa có external side effect hay
live adapter được cấp quyền.

```text
python scripts/validate_agent_semantics.py
python scripts/validate_m11.py
```
