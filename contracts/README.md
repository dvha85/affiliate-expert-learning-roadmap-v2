# Canonical contracts

Các schema trong thư mục này mô tả boundary máy đọc của hệ thống. Contract semantics thuộc Deterministic Core, không thuộc Agent/orchestrator/vendor.

## Artifact spine chuẩn

```text
Observation(observation_id, subject_id, provenance)
→ DecisionPacket(evidence_ids)
→ HistoryRecord(reuse canonical Observation + recorded decision linkage)
→ ActionRecord hoặc ActionIntent
→ Execution
→ EffectRef(HUMAN_ACTION | MACHINE_EXECUTION)
→ OutcomeRecord(effect_ref)
→ EvaluationRecord(same effect_ref)
→ ImprovementProposal + ReviewRecord
```

Một artifact có ID nhưng downstream reference không resolve được vẫn là broken lineage (liên kết đứt), không đủ để claim Reality/Operated PASS. `EffectRef` ngăn một `execution_id` bị giả làm `action_id` khi hệ thống chuyển từ M03 human action sang M09+ machine execution.

## Semantics không cấp quyền

```text
ActionIntent.intent_mode = PROPOSAL_ONLY
ActionIntent.execution_authorized = false
PolicyDecision.policy_mode = NON_AUTHORIZING
PolicyDecision.execution_authorized = false
```

`policy_review_required` chỉ là tín hiệu review của policy layer. Execution approval/authority nằm ở `ApprovalRecord`, grant/lease governance, gate và `ExecutionAuthorization` tương ứng; không dùng một boolean generic để trộn các lớp này.

## Activation theo Mission

- M00: canonical `Observation` + human `DecisionPacket`; DecisionPacket phải bind exact `evidence_ids` và `action=null`.
- M01: deterministic decision behavior trên canonical observation/context; fixture synthetic vẫn chỉ là E0.
- M02: `HistoryRecord` reuse canonical `Observation` cho immutable snapshot + version + replay; recorded decision phải bind `decision_id` + exact `evidence_ids`.
- M03: `ActionRecord` human-only + `OutcomeRecord(effect_ref=HUMAN_ACTION)`.
- M04: `AdvisorOutput` grounded, không write.
- M05: `EvaluationRecord` dùng cùng `EffectRef` với Outcome + `ImprovementProposal` + `ReviewRecord`; `auto_apply=false`.
- M06: watcher normalize về canonical `Observation`.
- M07: `ToolRegistry` read-only; tool output untrusted.
- M08: `ActionIntent(PROPOSAL_ONLY)` + `PolicyDecision(NON_AUTHORIZING)`; không artifact nào tự có execution authority.
- M09: `ApprovalRecord` + `ExecutionAuthorization(APPROVED_LIVE)` + `ExecutionRecord`; outcome machine dùng `EffectRef(MACHINE_EXECUTION)`.
- M10: `CanaryGrantApproval` + `CanaryGrant` + trusted stage-neutral `TrustedCostBound` + `CanaryLedger` + `CanaryGateDecision`; `RISK2` không delegated.
- M11: `ProductionLeaseApproval` + finite `ProductionLease` + `ProductionActivationRecord` + trusted `ProductionHealthSnapshot` + `TrustedCostBound` + `ProductionLedger` + `ProductionGateDecision` + `ProductionCycleRecord` + trusted `ProductionReconciliationResolution`; execution dùng `GOVERNED_PRODUCTION` và `RISK2` vẫn không delegated.

`TrustedCostBound` là một canonical contract dùng chung M10/M11; stage-specific authorization field vẫn giữ prefix `canary_`/`production_` để audit provenance. Không tạo hai cost-bound ontology chỉ vì Mission khác nhau.

Contract tồn tại **không tự cấp authority**. `ApprovalRecord`, `CanaryGrant` hay `ProductionLease` đều không phải execution. Gate decision luôn `execution_authorized=false`; executor chỉ chạy khi có authorization riêng, exact binding và deterministic revalidation. Agent/orchestrator không sở hữu truth, trusted cost/health, promotion, lease renewal hoặc authority widening.
