# M09 Checkpoints

- [ ] ApprovalRecord do người tạo và bind đúng `intent_hash` + `policy_version`.
- [ ] approval có expiry, one-time và correlation rõ.
- [ ] Agent/orchestrator không thể tự approve.
- [ ] resume/restart luôn revalidate intent/policy/approval.
- [ ] kill switch được kiểm ngay trước side effect.
- [ ] executor/action/target đều nằm trong allowlist/profile.
- [ ] idempotency key không thể tạo successful side effect lần hai.
- [ ] mismatch policy/approval phải fail closed (đóng an toàn).
- [ ] uncertain outcome không auto-retry mù.
- [ ] ExecutionRecord resolve tới authorization/approval/intent.
- [ ] bounded auto-action chưa được mở; M10 mới có authority đó.
- [ ] Reality + Operated evidence đã lưu.
