# M00 Evidence Packet (gói bằng chứng M00)

Template (mẫu) chuẩn dùng lại cho Mission M00. File này là **mẫu**, không phải evidence (bằng chứng) thật. Learner copy cấu trúc này sang `evidence/M00/M00-EVIDENCE-PACKET.md` và điền dữ liệu quan sát thật.

## Market question (câu hỏi thị trường)

- Audience/problem (nhóm mục tiêu/vấn đề):
- Offer/category (offer/danh mục):
- Decision cần hỗ trợ:
- Metric trung gian dễ gây hiểu nhầm:
- Outcome (kết quả) có thể làm đổi quyết định:
- Điều chưa biết chính:

## Quy tắc identity/provenance (định danh/nguồn gốc)

- Mỗi observation (quan sát) phải có `observation_id` duy nhất cho **một lần quan sát cụ thể**.
- `subject_id` nhận diện đối tượng ổn định đang được quan sát; cùng một offer có thể có nhiều `observation_id` theo thời gian.
- E1 public observation (quan sát công khai thật cấp E1) phải có `source_url` thật có thể kiểm tra.
- `observed_at` dùng ISO 8601 có timezone (múi giờ), ví dụ `2026-09-04T15:37:09+07:00`.
- `access_method` phải nói rõ cách quan sát, ví dụ `manual_browser`, `public_web_fetch`, `public_web_search_index`.
- Không biến `missing` thành `0`; không biến assumption (giả định) thành fact (sự thật được nguồn hỗ trợ trực tiếp).
- Không commit credential (thông tin xác thực), account data (dữ liệu tài khoản) hoặc personal/customer data (dữ liệu cá nhân/khách hàng).

## Public observations (các quan sát công khai)

> Lặp lại block (khối) dưới đây cho mỗi observation. M00 cần ít nhất 3 observation E1 thật.

### Observation — COPY THIS BLOCK (sao chép khối này)

- `observation_id`: ví dụ `obs-offer-a-20260904-01`
- `subject_id`: ví dụ `offer-a`
- `source_url`: URL nguồn công khai thật
- `observed_at`: thời điểm quan sát ISO 8601 + timezone
- `access_method`: `manual_browser | public_web_fetch | public_web_search_index | other:<mô tả>`
- `field_or_claim`: trường/phát biểu đang ghi nhận
- `claim_kind`: `fact | estimate | assumption | unknown`
  - `fact` — nguồn hiện có hỗ trợ trực tiếp
  - `estimate` — giá trị suy tính/xấp xỉ, phải có method (phương pháp)
  - `assumption` — giả định để thử, chưa được đo
  - `unknown` — chưa đủ evidence để kết luận
- `observed_value_or_state`: giá trị quan sát được hoặc trạng thái đặc biệt
- `observed_state`: `observed | missing | pending | not_yet_observable | inconclusive`
  - `observed` — đã quan sát được giá trị, kể cả giá trị bằng `0`
  - `missing` — field (trường) không có/không lấy được
  - `pending` — measurement window (khoảng thời gian đo) chưa hoàn tất
  - `not_yet_observable` — theo thiết kế hiện chưa thể quan sát
  - `inconclusive` — có dữ liệu nhưng chưa đủ phân biệt hypothesis (giả thuyết)
- `source_authority_or_role`: `platform | seller | publisher/reviewer | third-party-report | other:<mô tả>`
- `transformation_or_method`: `none` nếu chép trực tiếp; nếu estimate phải ghi cách tính
- `limitation`: giới hạn về freshness (độ mới), seller claim (tuyên bố người bán), voucher, account-specific price (giá theo tài khoản), sampling (lấy mẫu), parse (đọc/trích xuất), thiếu context, v.v.

## Unknowns / missing evidence (điều chưa biết / bằng chứng còn thiếu)

Liệt kê riêng các điều chưa biết, không tự điền `0` hoặc suy đoán:

- 
- 
- 

## Human DecisionPacket (gói quyết định do người lập)

- `decision_id`: ID quyết định
- `question`: câu hỏi cần quyết định
- `evidence_ids`: danh sách **chính xác** các `observation_id` được dùng
- `supported_facts`: các fact được evidence hỗ trợ
- `assumptions`: các giả định
- `unknowns`: điều chưa biết
- `state`: `RANK_SCENARIO | GET_MORE_DATA | HUMAN_REVIEW`
  - `RANK_SCENARIO` — xếp hạng kịch bản
  - `GET_MORE_DATA` — cần thêm dữ liệu
  - `HUMAN_REVIEW` — cần người xem xét
- `reason`: lý do
- `missing_evidence`: bằng chứng còn thiếu
- `next_measurement`: phép đo/quan sát tiếp theo có khả năng thay đổi decision
- `action`: `null`

Không ghi một fact vào `supported_facts` nếu `evidence_ids` không resolve được tới observation hỗ trợ nó. `DecisionPacket` không chứa action object (đối tượng hành động); `ActionIntent` là artifact (đối tượng dữ liệu) riêng ở Mission sau.

## Checklist E1 trước khi gọi packet là real evidence (bằng chứng thật)

- [ ] Có ít nhất 3 public observations (quan sát công khai) thật.
- [ ] Mỗi observation có `observation_id` duy nhất.
- [ ] Mỗi observation có `subject_id` phù hợp.
- [ ] Mỗi E1 observation có `source_url` thật.
- [ ] Có `observed_at` với timezone.
- [ ] Có `access_method`.
- [ ] `claim_kind` được phân loại đúng.
- [ ] `0`, `missing`, `pending`, `not_yet_observable`, `inconclusive` không bị trộn.
- [ ] Có `source_authority_or_role`.
- [ ] Có `transformation_or_method`.
- [ ] Có `limitation`.
- [ ] Không có secret/personal/customer data (bí mật/dữ liệu cá nhân/khách hàng).
- [ ] `DecisionPacket.evidence_ids` resolve được tới observation tồn tại.
- [ ] `action: null`.

## Explain-back (tự giải thích lại)

1. Điều gì trong packet là `fact` và `observation_id` nào hỗ trợ?
2. Điều gì vẫn chỉ là `assumption` hoặc `unknown`?
3. Vì sao `real evidence (bằng chứng thật) != reliable/current/authoritative/complete (đáng tin/hiện hành/có thẩm quyền/đầy đủ)`?
4. Vì sao real evidence chưa tự tạo recommendation (khuyến nghị)?
5. `next_measurement` nào có khả năng thay đổi decision và vì sao?
