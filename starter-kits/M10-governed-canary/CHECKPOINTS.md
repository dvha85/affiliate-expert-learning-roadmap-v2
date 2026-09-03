# M10 checkpoints — điểm kiểm canary có quản trị

- [ ] `CanaryGrant` có trusted `approval_ref`, `approver_id`, exact `policy_version`, expiry và valid `grant_hash`.
- [ ] Grant không chứa wildcard và không cho `RISK2`.
- [ ] `RISK1` chỉ auto khi grant explicitly delegates (ủy quyền rõ); baseline learner bắt đầu RISK0.
- [ ] Mỗi intent có `estimated_cost_minor` + `currency`; unknown cost không bị coi là 0.
- [ ] Scope/action/host/executor được kiểm lại ngay trước side effect.
- [ ] Total/rate/cost budget và outcome backpressure đều có failure test.
- [ ] Kill switch + revoked grant chặn action mới.
- [ ] Executor reload durable ledger; stale authorization không tự đảm bảo execution.
- [ ] Atomic reservation/idempotency semantics được chứng minh cho live adapter.
- [ ] Unknown side effect → `RECONCILIATION_REQUIRED`; không retry mù.
- [ ] `CanaryGateDecision.execution_authorized=false`.
- [ ] Chỉ `ExecutionAuthorization(execution_mode=GOVERNED_CANARY)` mới mở action cụ thể.
- [ ] Không có RISK2 auto-execution.
- [ ] Có real `OutcomeRecord` trước khi tăng exposure hoặc tạo grant rộng hơn.
- [ ] E5 evidence lưu trong template; sandbox/CI không được ghi thành Reality PASS.
