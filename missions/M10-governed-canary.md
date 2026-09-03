---
mission_id: M10
title: Governed Canary
status: ready
minimum_evidence: E5
authority: bounded-auto-RISK0-RISK1
external_side_effects: governed-canary-only
runtime: lab/mission-runtime/
eval_pack: evals/M10-governed-canary/
---

# Mission M10 — Governed Canary (canary có quản trị)

## Contract bàn giao

Một canary chain (chuỗi canary) thật:

```text
Decision/Evidence
→ ActionIntent
→ fresh deterministic PolicyDecision
→ human-approved CanaryGrant
→ CanaryGateDecision
→ ExecutionAuthorization(execution_mode=GOVERNED_CANARY)
→ controlled ExecutionRecord
→ real OutcomeRecord
→ reviewed grant decision
```

Grant phải bind exact policy version + trusted human approval provenance + hash + time/scope/executor + total/rate/cost/outcome budgets. `CanaryGateDecision.execution_authorized=false`; chỉ `ExecutionAuthorization` riêng mới cấp quyền cho một execution cụ thể.

## Authority ceiling

- `RISK0`: có thể auto-execute khi nằm trong grant.
- `RISK1`: chỉ có thể auto-execute khi human `CanaryGrant` **explicitly delegates (ủy quyền rõ)** risk này và mọi gate khác PASS; learner nên bắt đầu bằng RISK0.
- `RISK2`: luôn rời canary auto path sang per-action human approval; M10 không được tự chạy RISK2.
- Agent/orchestrator không tự tạo/đổi/renew grant, không tự tăng budget và không tự clear reconciliation state.

## PASS

Capability: offline eval/runtime PASS. Reality + Operated: có E5 external side effect thật trong grant nhỏ, kill-switch test, durable/atomic budget proof, idempotency/reconciliation failure evidence, real outcome observation và review sau canary. Local sandbox, fixture hoặc CI xanh không tự tạo E5.
