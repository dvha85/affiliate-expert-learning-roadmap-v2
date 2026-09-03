---
mission_id: M00
title: First Real Evidence Packet
status: ready
minimum_evidence: E1
authority: human/read-only
external_side_effects: false
---

# Mission M00 — First Real Evidence Packet (gói bằng chứng thật đầu tiên)

## Mục tiêu bàn giao

```text
3+ public observations E1 (quan sát công khai thật) có observation_id
→ classify fact / estimate / assumption / unknown
  (phân loại sự thật được hỗ trợ / ước tính / giả định / chưa biết)
→ Human DecisionPacket (gói quyết định do người lập)
→ exact evidence_ids + state + reason + missing evidence + next measurement
  (ID bằng chứng chính xác + trạng thái + lý do + bằng chứng thiếu + phép đo tiếp theo)
→ action: null
→ KHÔNG thực thi hành động bên ngoài
```

## Các bài bắt buộc

- `curriculum/M00/M00.1-affiliate-intelligence-objective.md`
- `curriculum/M00/M00.2-evidence-uncertainty.md`
- `curriculum/M00/M00.3-decision-approval-execution.md`

## Gói bằng chứng

Mỗi observation (quan sát) tối thiểu có `observation_id`, `subject_id`, `source_url`, `observed_at`, `access_method`, claim/value (phát biểu/giá trị), `claim_kind`, observed state (trạng thái quan sát) và limitation (giới hạn).

Human DecisionPacket tối thiểu có:

```text
decision_id: ID quyết định
question: câu hỏi cần quyết định
evidence_ids: exact observation IDs đã dùng
supported_facts: các fact được bằng chứng hỗ trợ
assumptions: các giả định
unknowns: điều chưa biết
state: RANK_SCENARIO | GET_MORE_DATA | HUMAN_REVIEW
reason: lý do
missing_evidence: bằng chứng còn thiếu
next_measurement: phép đo tiếp theo
action: null
```

`evidence_ids` phải resolve được tới observation trong packet. `DecisionPacket` không được nhúng action object; Decision và ActionIntent là hai artifact khác nhau.

## Các ca lỗi

- URL placeholder (giữ chỗ) được gọi là E1;
- observation thiếu ID hoặc tái sử dụng cùng ID cho nội dung khác;
- DecisionPacket viện dẫn evidence ID không tồn tại;
- assumption (giả định) ghi thành fact (sự thật được hỗ trợ);
- missing (thiếu) đổi thành observed zero (đã quan sát bằng 0);
- sample/synthetic (mẫu/mô phỏng) được gọi là market reality (thực tế thị trường);
- real provenance (nguồn gốc thật) tự nâng output thành `RECOMMEND`;
- DecisionPacket tạo external action (hành động bên ngoài).

## PASS

### Capability (năng lực)
- tạo evidence packet (gói bằng chứng) có identity/provenance/uncertainty (định danh/nguồn gốc/bất định) đúng;
- tạo Human DecisionPacket có exact evidence linkage + reason/missing evidence/next measurement (liên kết bằng chứng chính xác + lý do/bằng chứng thiếu/phép đo tiếp theo).

### Reality (thực tế)
- có ít nhất 3 public observations E1 thật; nếu không thể, ghi `BLOCKED_EXTERNAL` và chưa PASS Reality.

### Operated (đã tự vận hành chứng minh)
- packet/version (gói/phiên bản) đủ rõ để M01 dùng làm context (ngữ cảnh) cho deterministic baseline (mốc tất định), và mọi `evidence_ids` resolve được.

## Kết quả

`pre-bot` (giai đoạn trước Bot): market truth/context (sự thật/ngữ cảnh thị trường) có trước; Bot v0.1 bắt đầu ở M01.
