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
→ Observation/Evidence
→ Decision + ActionIntent
→ fresh deterministic PolicyDecision
→ trusted ProductionHealthSnapshot + trusted cost bound
→ ProductionGateDecision
→ ExecutionAuthorization(GOVERNED_PRODUCTION)
→ controlled ExecutionRecord
→ real OutcomeRecord
→ EvaluationRecord
→ ImprovementProposal(auto_apply=false)
→ Human ReviewRecord
↺
```

## Authority ceiling

- `RISK0`: có thể auto trong lease hợp lệ.
- `RISK1`: chỉ auto nếu lease cho phép **và** human promotion approval xác nhận đã có E5 cho RISK1.
- `RISK2`: luôn per-action human approval; không đi qua `GOVERNED_PRODUCTION`.
- `DEGRADE`: read-only/no side effect (chỉ đọc/không có tác động bên ngoài).
- `STOP`: sticky (dính trạng thái); lease cũ không tự resume.
- Agent/orchestrator không tự create/renew/widen grant, policy, budget, credentials hoặc clear STOP/reconciliation.
- Improvement có thể được tự đề xuất nhưng `auto_apply=false`; thay đổi authority cần human review + version mới.

## PASS

Capability (năng lực): eval/runtime/local sandbox PASS.

Reality + Operated: có E6 production thật qua observation window với 3+ closed cycles, exact lease promotion từ E5, durable state/restart, trusted health/cost inputs, kill-switch + auto-stop/degrade drill, recovery bằng reviewed lease/version mới, real outcomes, Evaluation → reviewed ImprovementProposal và audit chứng minh không có self-widening authority. Sandbox/CI không tự tạo E6.
