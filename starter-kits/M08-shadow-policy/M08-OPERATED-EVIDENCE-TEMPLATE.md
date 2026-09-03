# M08 Bằng chứng đã vận hành

- Decision/Evidence chain (chuỗi quyết định/bằng chứng) đã dùng:
- `intent_id` + `intent_hash`:
- `action_type` + target + parameters (tham số):
- `proposed_by` + `proposal_ref` nếu có:
- `correlation_id` + `idempotency_key`:
- created/expires time (thời điểm tạo/hết hạn):
- policy version (phiên bản policy):
- expected policy state (trạng thái dự kiến):
- observed PolicyDecision (kết quả policy quan sát được):
- risk class (nhóm rủi ro) + reason (lý do):
- tamper/expiry failure case (ca lỗi sửa/hết hạn):
- missing-link case (ca thiếu liên kết):
- duplicate/collision case (ca trùng/xung đột):
- proof không có side effect (bằng chứng không có tác động ngoài):
- limitation (giới hạn):
- next measurement (phép đo tiếp theo):
