# Canonical contracts

Các schema trong thư mục này mô tả boundary máy đọc của hệ thống. Contract semantics thuộc Deterministic Core, không thuộc Agent/orchestrator/vendor.

## Activation theo Mission

- M00: `Observation` + human `DecisionPacket`.
- M01: deterministic decision behavior.
- M02: `HistoryRecord` cho immutable snapshot + version + replay.
- M03: `ActionRecord` human-only + `OutcomeRecord`.
- M04: `AdvisorOutput` với grounding/abstention và không write tool.
- M05: `ImprovementProposal` versioned, reviewable, rollbackable, `auto_apply=false`.
- M06: watcher read-only semantics được bảo vệ bằng runtime/eval.
- M07: `ToolRegistry` chỉ cho read-only tools/methods/hosts.
- `ActionIntent`/`PolicyDecision` được định nghĩa sớm nhưng chỉ activate từ M08+ theo curriculum.

Contract tồn tại **không tự cấp authority** cho Mission trước.
