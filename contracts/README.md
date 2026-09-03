# Canonical contracts

Các schema trong thư mục này mô tả boundary máy đọc của hệ thống. Contract semantics thuộc Deterministic Core, không thuộc Agent/orchestrator/vendor.

## Artifact spine chuẩn

```text
Observation(observation_id, subject_id, provenance)
→ DecisionPacket(evidence_ids)
→ HistoryRecord(reuse canonical Observation + recorded decision linkage)
→ ActionRecord hoặc ActionIntent
→ Execution/Outcome
→ Evaluation
→ ImprovementProposal + ReviewRecord
```

Một artifact có ID nhưng downstream reference không resolve được vẫn là broken lineage (liên kết đứt), không đủ để claim Reality/Operated PASS.

## Activation theo Mission

- M00: canonical `Observation` + human `DecisionPacket`; DecisionPacket phải bind exact `evidence_ids` và `action=null`.
- M01: deterministic decision behavior trên canonical observation/context; fixture synthetic vẫn chỉ là E0.
- M02: `HistoryRecord` reuse canonical `Observation` cho immutable snapshot + version + replay; recorded decision phải bind `decision_id` + exact `evidence_ids`.
- M03: `ActionRecord` human-only + `OutcomeRecord`.
- M04: `AdvisorOutput` grounded, không write.
- M05: `EvaluationRecord` + `ImprovementProposal` + `ReviewRecord`; `auto_apply=false`.
- M06: watcher normalize về canonical `Observation`.
- M07: `ToolRegistry` read-only; tool output untrusted.
- M08: `ActionIntent` + `PolicyDecision` shadow-only; `execution_authorized=false`.
- M09: `ApprovalRecord` + `ExecutionAuthorization(APPROVED_LIVE)` + `ExecutionRecord`.
- M10: `CanaryGrantApproval` + `CanaryGrant` + trusted `CanaryCostBound` + `CanaryLedger` + `CanaryGateDecision`; `RISK2` không delegated.
- M11: `ProductionLeaseApproval` + finite `ProductionLease` + trusted `ProductionHealthSnapshot` + `ProductionLedger` + `ProductionGateDecision` + `ProductionCycleRecord` + trusted `ProductionReconciliationResolution`; execution dùng `GOVERNED_PRODUCTION` và `RISK2` vẫn không delegated.

Contract tồn tại **không tự cấp authority**. `ApprovalRecord`, `CanaryGrant` hay `ProductionLease` đều không phải execution. Gate decision luôn `execution_authorized=false`; executor chỉ chạy khi có authorization riêng, exact binding và deterministic revalidation. Agent/orchestrator không sở hữu truth, trusted cost/health, promotion, lease renewal hoặc authority widening.
