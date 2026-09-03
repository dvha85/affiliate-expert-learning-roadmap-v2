# Canonical contracts

Các schema trong thư mục này mô tả boundary máy đọc của hệ thống. Contract semantics thuộc Deterministic Core, không thuộc Agent/orchestrator/vendor.

## Activation theo Mission

- M00: `Observation` + human `DecisionPacket`.
- M01: deterministic decision behavior.
- M02: `HistoryRecord` cho immutable snapshot + version + replay.
- M03: `ActionRecord` human-only + `OutcomeRecord`, với linkage phải resolve tới decision/action phù hợp.
- M04: `AdvisorOutput` với grounding/abstention và `write_tool_requested=false` bắt buộc.
- M05: `EvaluationRecord` + `ImprovementProposal` + `ReviewRecord`; proposal versioned, reviewable, rollbackable, `auto_apply=false`.
- M06: watcher phải normalize về canonical `Observation`; read-only/change-detection semantics được bảo vệ bằng runtime/eval.
- M07: `ToolRegistry` chỉ cho read-only tools/methods/hosts; tool output luôn là untrusted data.
- M08: `ActionIntent` + `PolicyDecision` được activate ở **shadow only**; exact intent binding/hash/expiry/idempotency/linkage phải được kiểm và `execution_authorized=false` luôn bắt buộc.
- M09: `ApprovalRecord` + `ExecutionAuthorization` + `ExecutionRecord`; approval phải do người tạo, bind đúng `intent_hash` + `policy_version`, one-time + expiry; authorization chỉ được tạo sau revalidation và executor/kill-switch/idempotency gate.

Contract tồn tại **không tự cấp authority** cho Mission trước hoặc sau. `ALLOW` trong M08 không phải execution permission. `ApprovalRecord` cũng không tự là execution: M09 phải deterministic revalidate rồi mới tạo `ExecutionAuthorization`. Schema-valid reference nhưng không resolve được provenance không đủ cho Reality/Operated PASS.
