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
→ ActionIntent(PROPOSAL_ONLY, execution_authorized=false)
→ fresh deterministic PolicyDecision(NON_AUTHORIZING)
→ human CanaryGrantApproval(exact grant_hash)
→ CanaryGrant
→ trusted TrustedCostBound(exact intent_hash)
→ CanaryLedger(exact grant_hash)
→ CanaryGateDecision
→ ExecutionAuthorization(execution_mode=GOVERNED_CANARY)
→ controlled ExecutionRecord
→ real OutcomeRecord(effect_ref=MACHINE_EXECUTION/exact execution_id)
→ reviewed grant decision
```

Grant phải bind exact policy version + exact human approval + hash + time/scope/executor + total/rate/cost/outcome budgets. Cost budget dùng trusted `TrustedCostBound`, không tin estimate do Agent/ActionIntent tự khai báo. `CanaryGateDecision.execution_authorized=false`; chỉ `ExecutionAuthorization` riêng mới cấp quyền cho một execution cụ thể.

`PolicyDecision.policy_review_required` không thay per-action approval semantics. Canary gate mới quyết định `per_action_approval_required`; `ApprovalRecord`/grant governance vẫn là authority artifact riêng.

## Authority ceiling

- `RISK0`: có thể auto-execute khi nằm trong grant.
- `RISK1`: chỉ có thể auto-execute khi human `CanaryGrant` **explicitly delegates (ủy quyền rõ)** risk này và mọi gate khác PASS; learner nên bắt đầu bằng RISK0.
- `RISK2`: luôn rời canary auto path sang per-action human approval; M10 không được tự chạy RISK2.
- Agent/orchestrator không tự tạo/đổi/renew grant, không tự tăng budget, không tự tạo trusted cost bound và không tự clear reconciliation state.
- Mất durable ledger sau khi đã có execution/spend phải fail closed; không reset budget từ RAM.
- Outcome chỉ release backpressure khi `effect_ref.effect_kind=MACHINE_EXECUTION` và bind đúng pending `execution_id`.

## PASS

Capability: offline eval/runtime PASS. Reality + Operated: có E5 external side effect thật trong grant nhỏ, exact grant-approval proof, trusted cost-bound proof, kill-switch test, durable/atomic budget proof, idempotency/reconciliation failure evidence, real OutcomeRecord resolve đúng execution và review sau canary. Local sandbox, fixture hoặc CI xanh không tự tạo E5.
