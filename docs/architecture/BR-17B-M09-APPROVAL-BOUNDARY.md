# BR-17b — M09 approval/persistence boundary

M09 chỉ chuyển sang `AUTHORIZED` khi approval còn hạn, đúng intent/policy,
đúng approver hợp lệ và kill switch không active. Approval phải được persist
qua canonical store, đọc lại sau restart rồi revalidate trước executor; thiếu,
tamper, mismatch hoặc orphan link đều `DENY`/`WAIT`, không tạo external effect.

Các case M09 hiện hành bao phủ authorized, chờ approval, approver không hợp lệ,
mismatch, approval trước policy, expired, kill switch, executor/policy state,
duplicate và tampered intent. Đây là lab/schema evidence; fixture approval
không phải human authorization thật.

```text
python scripts/validate_agent_semantics.py
```
