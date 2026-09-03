# Canonical contracts

Các schema trong thư mục này mô tả boundary máy đọc của hệ thống. Contract semantics thuộc Deterministic Core, không thuộc Agent/orchestrator/vendor.

Current minimum set:

- `observation.schema.json`
- `decision-packet.schema.json`
- `history-record.schema.json`
- `action-intent.schema.json`
- `policy-decision.schema.json`

## Activation theo Mission

- M00: Observation + human DecisionPacket.
- M01: deterministic decision behavior; Go runtime là reference implementation.
- M02: HistoryRecord để giữ immutable input snapshot + `formula_version` + `input_hash` + recorded result cho replay.
- `ActionIntent`/`PolicyDecision` được định nghĩa sớm để boundary rõ nhưng không được activate trước Mission tương ứng.

Contract tồn tại **không tự cấp authority** cho Mission trước.
