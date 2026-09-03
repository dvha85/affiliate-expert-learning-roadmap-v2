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
- M09: `ApprovalRecord` + `ExecutionAuthorization(APPROVED_LIVE)` + `ExecutionRecord`; approval phải do người tạo, bind đúng `intent_hash` + `policy_version`, one-time + expiry; authorization chỉ được tạo sau revalidation và executor/kill-switch/idempotency gate.
- M10: `CanaryGrantApproval` bind human approval vào **exact `grant_hash`**; `CanaryGrant` mô tả authority envelope; `CanaryCostBound` bind trusted cost ceiling vào exact `intent_hash`; `CanaryLedger` bind exact `grant_hash` và execution/outcome linkage; `CanaryGateDecision.execution_authorized=false`; chỉ `ExecutionAuthorization(GOVERNED_CANARY)` riêng mới cấp quyền cho một execution cụ thể. `RISK2` không được delegated.

Contract tồn tại **không tự cấp authority** cho Mission trước hoặc sau. `ALLOW` trong M08 không phải execution permission. `ApprovalRecord` ở M09 cũng không tự là execution. `CanaryGrant` ở M10 không phải blanket approval; `approval_ref` chỉ có nghĩa khi resolve về `CanaryGrantApproval` có cùng `grant_id/version/hash`. Giá trị cost do ActionIntent/Agent tự khai báo không phải trusted budget input; M10 dùng `CanaryCostBound` từ deterministic/control-plane source. Schema-valid reference nhưng không resolve được provenance không đủ cho Reality/Operated PASS.
