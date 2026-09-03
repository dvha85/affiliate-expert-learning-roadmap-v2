---
mission_id: M09
title: Durable Approval + Controlled Executor
status: ready
minimum_evidence: E4/E5-ready
authority: approval-gated execution
external_side_effects: bounded-and-approved
runtime: lab/mission-runtime/
eval_pack: evals/M09-approval-execution/
---

# Mission M09 — Phê duyệt bền vững + bộ thực thi có kiểm soát

## Contract bàn giao

Một chuỗi `ActionIntent → PolicyDecision → ApprovalRecord → ExecutionAuthorization → ExecutionRecord` có linkage/hash/version/correlation đầy đủ; approval do người tạo, one-time, có expiry; executor bị allowlist; resume/recovery revalidate và không duplicate side effect.

## Authority ceiling

Mọi machine execution ở M09 đều cần human approval hợp lệ. Agent/orchestrator không tự approve. `ALLOW` của policy không tự cấp execution. Không mở bounded auto-action; đó là M10.

## PASS

Capability có thể chứng minh bằng local sandbox executor. Reality + Operated cần E4 context thật phù hợp, persisted approval/restart test, kill switch + idempotency failure cases và bằng chứng executor không vượt profile. E5 chỉ là readiness cho canary M10, không tự được claim từ fixture.
