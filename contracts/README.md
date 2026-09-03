# Canonical contracts

Các schema trong thư mục này mô tả boundary máy đọc của hệ thống. Contract semantics thuộc Deterministic Core, không thuộc Agent/orchestrator/vendor.

Current minimum set:

- `observation.schema.json`
- `decision-packet.schema.json`
- `action-intent.schema.json`
- `policy-decision.schema.json`

M00 chỉ cần Observation + human DecisionPacket. `ActionIntent`/`PolicyDecision` được định nghĩa sớm để boundary rõ nhưng không được activate trước Mission tương ứng.
