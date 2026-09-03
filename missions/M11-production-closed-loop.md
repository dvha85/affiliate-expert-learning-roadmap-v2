---
mission_id: M11
title: Production Closed Loop
status: ready
minimum_evidence: E6
authority: governed-production-RISK0-RISK1
external_side_effects: production-lease-only
runtime: lab/mission-runtime/
eval_pack: evals/M11-production-closed-loop/
---

# Mission M11 — Production Closed Loop (vòng production có quản trị)

## Contract bàn giao

```text
E5 reviewed promotion
→ ProductionLeaseApproval
→ finite ProductionLease
→ ProductionActivationRecord
→ Observation/Evidence
→ Decision + ActionIntent(PROPOSAL_ONLY)
→ fresh deterministic PolicyDecision(NON_AUTHORIZING)
→ trusted ProductionHealthSnapshot + TrustedCostBound
→ ProductionGateDecision
→ ExecutionAuthorization(GOVERNED_PRODUCTION)
→ controlled ExecutionRecord
→ real OutcomeRecord(effect_ref=MACHINE_EXECUTION/exact execution_id)
→ EvaluationRecord(same effect_ref)
→ ImprovementProposal(auto_apply=false)
→ Human ReviewRecord
↺
```

`ProductionActivationRecord` exact-binds lease ID/version/hash và được tạo trước khi executor được mở. Nó là canonical durable artifact, không chỉ là implementation detail (chi tiết triển khai); mất activation/ledger không được suy diễn thành fresh state.

## Authority ceiling

- `RISK0`: có thể auto trong lease hợp lệ.
- `RISK1`: chỉ auto nếu lease cho phép **và** human promotion approval xác nhận đã có E5 cho RISK1.
- `RISK2`: luôn per-action human approval; không đi qua `GOVERNED_PRODUCTION`.
- `DEGRADE`: read-only/no side effect (chỉ đọc/không có tác động bên ngoài).
- `STOP`: sticky (dính trạng thái); lease cũ không tự resume.
- Agent/orchestrator không tự create/renew/widen grant, policy, budget, credentials hoặc clear STOP/reconciliation.
- Improvement có thể được tự đề xuất nhưng `auto_apply=false`; thay đổi authority cần human review + version mới.
- `ActionIntent` và `PolicyDecision` luôn non-authorizing (không cấp quyền); authority chỉ xuất hiện ở exact `ExecutionAuthorization` sau production gate.

## PASS

Capability (năng lực): eval/runtime/local sandbox PASS.

Reality + Operated: có E6 production thật qua observation window với 3+ closed cycles, exact lease promotion từ E5, durable activation + ledger qua restart, trusted health/cost inputs, kill-switch + auto-stop/degrade drill, recovery bằng reviewed lease/version mới, real outcomes với exact machine `EffectRef`, Evaluation → reviewed ImprovementProposal và audit chứng minh không có self-widening authority. Sandbox/CI không tự tạo E6.
