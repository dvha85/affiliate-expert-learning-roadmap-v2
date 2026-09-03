# M00 Evidence Packet (gói bằng chứng M00)

## Market question (câu hỏi thị trường)

- Audience/problem (nhóm mục tiêu/vấn đề):
- Offer/category (offer/danh mục):
- Decision cần hỗ trợ:
- Outcome (kết quả) có thể làm đổi quyết định:

## Public observations (các quan sát công khai)

### Observation 1 (quan sát 1)

- source_url: URL nguồn
- observed_at: thời điểm quan sát
- access_method: cách truy cập
- claim/value: phát biểu/giá trị
- claim_kind: `fact | estimate | assumption | unknown` — sự thật được hỗ trợ | ước tính | giả định | chưa biết
- observed_state: `observed | missing | pending | not_yet_observable | inconclusive` — đã quan sát | thiếu | đang chờ | chưa thể quan sát | chưa đủ kết luận
- limitation: giới hạn

### Observation 2 (quan sát 2)

- source_url: URL nguồn
- observed_at: thời điểm quan sát
- access_method: cách truy cập
- claim/value: phát biểu/giá trị
- claim_kind: loại phát biểu
- observed_state: trạng thái quan sát
- limitation: giới hạn

### Observation 3 (quan sát 3)

- source_url: URL nguồn
- observed_at: thời điểm quan sát
- access_method: cách truy cập
- claim/value: phát biểu/giá trị
- claim_kind: loại phát biểu
- observed_state: trạng thái quan sát
- limitation: giới hạn

## Human DecisionPacket (gói quyết định do người lập)

- question: câu hỏi
- supported_facts: các fact được bằng chứng hỗ trợ
- assumptions: các giả định
- unknowns: điều chưa biết
- decision_state: `RANK_SCENARIO | GET_MORE_DATA | HUMAN_REVIEW` — xếp hạng kịch bản | cần thêm dữ liệu | người kiểm tra
- reason: lý do
- missing_evidence: bằng chứng còn thiếu
- next_measurement: phép đo tiếp theo
- action: `null` — không có hành động

## Explain-back (tự giải thích lại)

1. Điều gì trong packet là fact và nguồn nào hỗ trợ?
2. Điều gì vẫn chỉ là assumption/unknown (giả định/chưa biết)?
3. Vì sao real evidence (bằng chứng thật) chưa đủ để gọi output là recommendation (khuyến nghị)?
4. Measurement (phép đo) tiếp theo có khả năng thay đổi decision như thế nào?
