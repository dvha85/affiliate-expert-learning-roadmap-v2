# M00 Evidence Packet (gói bằng chứng M00)

## Market question (câu hỏi thị trường)

- Audience/problem (nhóm mục tiêu/vấn đề):
- Offer/category (offer/danh mục):
- Decision cần hỗ trợ:
- Outcome (kết quả) có thể làm đổi quyết định:

## Public observations (các quan sát công khai)

Mỗi observation phải có `observation_id` riêng. ID này là khóa provenance (nguồn gốc) để DecisionPacket và các Mission sau tham chiếu đúng observation, không dùng số thứ tự hiển thị như `Observation 1` làm identity.

### Observation 1 (quan sát 1)

- observation_id: ID ổn định của lần quan sát, ví dụ `obs-offer-a-20260903-01`
- subject_id: ID ổn định của đối tượng đang quan sát
- source_url: URL nguồn thật
- observed_at: thời điểm quan sát
- access_method: cách truy cập
- claim/value: phát biểu/giá trị
- claim_kind: `fact | estimate | assumption | unknown` — sự thật được hỗ trợ | ước tính | giả định | chưa biết
- observed_state: `observed | missing | pending | not_yet_observable | inconclusive` — đã quan sát | thiếu | đang chờ | chưa thể quan sát | chưa đủ kết luận
- source_authority_or_role: vai trò của nguồn, nếu biết
- transformation_or_method: cách biến đổi/tính toán, nếu có
- limitation: giới hạn

### Observation 2 (quan sát 2)

- observation_id:
- subject_id:
- source_url:
- observed_at:
- access_method:
- claim/value:
- claim_kind:
- observed_state:
- limitation:

### Observation 3 (quan sát 3)

- observation_id:
- subject_id:
- source_url:
- observed_at:
- access_method:
- claim/value:
- claim_kind:
- observed_state:
- limitation:

## Human DecisionPacket (gói quyết định do người lập)

- decision_id: ID của quyết định
- question: câu hỏi
- evidence_ids: danh sách chính xác các `observation_id` được dùng
- supported_facts: các fact được bằng chứng hỗ trợ
- assumptions: các giả định
- unknowns: điều chưa biết
- state: `RANK_SCENARIO | GET_MORE_DATA | HUMAN_REVIEW` — xếp hạng kịch bản | cần thêm dữ liệu | người kiểm tra
- reason: lý do
- missing_evidence: bằng chứng còn thiếu
- next_measurement: phép đo tiếp theo
- action: `null` — không có hành động

Không ghi một fact vào `supported_facts` nếu `evidence_ids` không resolve được tới observation hỗ trợ nó. `DecisionPacket` không chứa action object; ActionIntent là artifact riêng ở Mission sau.

## Explain-back (tự giải thích lại)

1. Điều gì trong packet là fact và `observation_id` nào hỗ trợ?
2. Điều gì vẫn chỉ là assumption/unknown (giả định/chưa biết)?
3. Vì sao real evidence (bằng chứng thật) chưa đủ để gọi output là recommendation (khuyến nghị)?
4. Measurement (phép đo) tiếp theo có khả năng thay đổi decision như thế nào?
