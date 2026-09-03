---
mission_id: M09
title: Durable Approval + Controlled Executor
status: ready
minimum_evidence: E4
readiness_target: E5
authority: approval-gated execution
external_side_effects: bounded-and-approved
runtime: lab/mission-runtime/
eval_pack: evals/M09-approval-execution/
---

# Mission M09 — Phê duyệt bền vững + bộ thực thi có kiểm soát

## Contract bàn giao

Một chuỗi `ActionIntent → PolicyDecision → ApprovalRecord → ExecutionAuthorization → ExecutionRecord` có linkage/hash/version/correlation đầy đủ; approval do người tạo, one-time, có expiry; executor bị allowlist; resume/recovery revalidate và không duplicate side effect.

`ActionIntent` vẫn là `PROPOSAL_ONLY` và không tự có quyền thực thi. `PolicyDecision.policy_review_required` chỉ là tín hiệu review của policy layer; **M09 execution approval là `ApprovalRecord` riêng và bắt buộc cho mọi machine execution**, kể cả khi policy decision là `ALLOW`.

## Authority ceiling

Mọi machine execution ở M09 đều cần human approval hợp lệ. Agent/orchestrator không tự approve. `ALLOW` của policy không tự cấp execution. Không mở bounded auto-action; đó là M10.

## Evidence semantics (ngữ nghĩa bằng chứng)

```text
minimum_evidence: E4
readiness_target: E5
```

M09 cần E4 context thật để chứng minh approval/execution chain. `readiness_target: E5` chỉ có nghĩa M09 phải tạo nền tảng đủ an toàn để bước sang governed canary; nó **không** tự tạo E5.

## PASS

Capability có thể chứng minh bằng local sandbox executor. Reality + Operated cần E4 context thật phù hợp, persisted approval/restart test, kill switch + idempotency failure cases và bằng chứng executor không vượt profile. E5 chỉ là readiness cho canary M10, không tự được claim từ fixture.
