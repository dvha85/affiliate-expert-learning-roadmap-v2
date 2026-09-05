# M10 checkpoints — điểm kiểm canary có quản trị

- [ ] Chạy [m10-check](../../docs/architecture/M10-JSON-BOUNDARY.md) và giải thích vì sao `ARTIFACT_VALID_UNVERIFIED` không xác thực chain, provenance hoặc quyền execute.

- [ ] `ActionIntent.intent_mode=PROPOSAL_ONLY` và `execution_authorized=false`; intent không tự biến thành live authority.
- [ ] `PolicyDecision.policy_mode=NON_AUTHORIZING`; `policy_review_required` không bị dùng thay execution approval.
- [ ] `CanaryGrantApproval` do người tạo bind đúng `grant_id/version/hash`; không chỉ kiểm `approval_ref` có tồn tại.
- [ ] `CanaryGrant` có exact `policy_version`, expiry và valid `grant_hash`.
- [ ] Grant không chứa wildcard và không cho `RISK2`.
- [ ] `RISK1` chỉ auto khi grant explicitly delegates (ủy quyền rõ); baseline learner bắt đầu RISK0.
- [ ] Cost budget dùng trusted `TrustedCostBound` bind exact `intent_hash`; estimate trong ActionIntent/Agent không tự là budget input.
- [ ] Scope/action/host/executor được kiểm lại ngay trước side effect.
- [ ] `CanaryLedger` bind exact `grant_hash`; đổi grant content với cùng ID/version không được reuse ledger mơ hồ.
- [ ] Total/rate/cost budget và outcome backpressure đều có failure test.
- [ ] Kill switch + revoked grant chặn action mới.
- [ ] Executor reload durable ledger; stale authorization không tự đảm bảo execution.
- [ ] Mất durable ledger sau khi đã có spend/execution phải fail closed (`WAIT_LEDGER_MISSING`), không reset từ RAM.
- [ ] Atomic reservation/idempotency semantics được chứng minh cho live adapter.
- [ ] Unknown side effect → `RECONCILIATION_REQUIRED`; không retry mù.
- [ ] `CanaryGateDecision.execution_authorized=false`.
- [ ] Chỉ `ExecutionAuthorization(execution_mode=GOVERNED_CANARY)` bind grant hash + trusted cost bound mới mở action cụ thể.
- [ ] Không có RISK2 auto-execution.
- [ ] `OutcomeRecord.effect_ref.effect_kind=MACHINE_EXECUTION` và `effect_id` resolve đúng pending `execution_id` trước khi giảm outcome backpressure.
- [ ] Có real `OutcomeRecord` trước khi tăng exposure hoặc tạo grant rộng hơn.
- [ ] E5 evidence lưu trong template; sandbox/CI không được ghi thành Reality PASS.
