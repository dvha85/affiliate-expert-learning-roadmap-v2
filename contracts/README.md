# Canonical contracts

Các schema trong thư mục này mô tả boundary máy đọc của hệ thống. Contract semantics thuộc Deterministic Core, không thuộc Agent/orchestrator/vendor.

## Activation theo Mission

- M00: `Observation` + human `DecisionPacket`.
- M01: deterministic decision behavior.
- M02: `HistoryRecord` cho immutable snapshot + version + replay.
- M03: `ActionRecord` human-only + `OutcomeRecord`.
- M04: `AdvisorOutput` grounded, không write.
- M05: `EvaluationRecord` + `ImprovementProposal` + `ReviewRecord`; `auto_apply=false`.
- M06: watcher normalize về canonical `Observation`.
- M07: `ToolRegistry` read-only; tool output untrusted.
- M08: `ActionIntent` + `PolicyDecision` shadow-only; `execution_authorized=false`.
- M09: `ApprovalRecord` + `ExecutionAuthorization(APPROVED_LIVE)` + `ExecutionRecord`.
- M10: `CanaryGrantApproval` + `CanaryGrant` + trusted `CanaryCostBound` + `CanaryLedger` + `CanaryGateDecision`; `RISK2` không delegated.
- M11: `ProductionLeaseApproval` + finite `ProductionLease` + trusted `ProductionHealthSnapshot` + `ProductionLedger` + `ProductionGateDecision` + `ProductionCycleRecord`; execution dùng `GOVERNED_PRODUCTION` và `RISK2` vẫn không delegated.

Contract tồn tại **không tự cấp authority**. `ApprovalRecord`, `CanaryGrant` hay `ProductionLease` đều không phải execution. Gate decision luôn `execution_authorized=false`; executor chỉ chạy khi có authorization riêng, exact binding và deterministic revalidation. Agent/orchestrator không sở hữu truth, trusted cost/health, promotion, lease renewal hoặc authority widening.
