# M00 Evidence Packet — Learner Working File (file làm việc bằng chứng M00)

> File này lưu **evidence (bằng chứng) thực tế của learner** cho Mission M00. Không dùng file template để ghi dữ liệu thật.

## Market question (câu hỏi thị trường)

- Audience/problem (nhóm mục tiêu/vấn đề): Freelancer / remote worker tại Việt Nam làm việc bằng laptop/MacBook trên bàn nhỏ, cần giải pháp gọn và rẻ để nâng laptop khi làm việc.
- Offer/category (offer/danh mục): Giá đỡ laptop nhôm gấp gọn, điều chỉnh độ cao, phân khúc giá rẻ.
- Decision cần hỗ trợ: Trong các giá đỡ laptop nhôm gấp gọn giá rẻ có thể quan sát công khai, offer nào đáng để tiếp tục nghiên cứu như một cơ hội Affiliate cho freelancer/remote worker Việt Nam?
- Metric trung gian dễ gây hiểu nhầm: giá rẻ nhất, commission rate (tỷ lệ hoa hồng) cao nhất, số đã bán cao nhất, rating (điểm đánh giá) cao nhất.
- Outcome (kết quả) có thể làm đổi quyết định: một offer khác tuy metric trung gian thấp hơn nhưng có product–audience fit (độ phù hợp sản phẩm–nhóm mục tiêu), conversion potential (khả năng chuyển đổi), valid-order probability (xác suất đơn hợp lệ) và expected affiliate value (giá trị Affiliate kỳ vọng) tốt hơn.
- Điều chưa biết chính: demand (nhu cầu) thực tế của ngách, conversion rate, valid-order rate, refund/cancel rate, commission thực nhận, seller quality (chất lượng người bán), competition (cạnh tranh), tracking reliability (độ tin cậy ghi nhận Affiliate).

## Cách điền observation (quan sát)

Dùng 1 block cho **mỗi lần quan sát cụ thể**. Nếu bạn mở lại cùng một offer vào thời điểm khác, giữ `subject_id` nhưng tạo `observation_id` mới.

### Observation 1 — điền sau khi tự mở source bằng trình duyệt

- `observation_id`: `obs-laptop-stand-20260904-01`
- `subject_id`: 
- `source_url`: 
- `observed_at`: 
- `access_method`: `manual_browser`
- `field_or_claim`: 
- `claim_kind`: `fact | estimate | assumption | unknown`
- `observed_value_or_state`: 
- `observed_state`: `observed | missing | pending | not_yet_observable | inconclusive`
- `source_authority_or_role`: 
- `transformation_or_method`: `none`
- `limitation`: 

### Observation 2 — điền sau khi tự mở source bằng trình duyệt

- `observation_id`: `obs-laptop-stand-20260904-02`
- `subject_id`: 
- `source_url`: 
- `observed_at`: 
- `access_method`: `manual_browser`
- `field_or_claim`: 
- `claim_kind`: `fact | estimate | assumption | unknown`
- `observed_value_or_state`: 
- `observed_state`: `observed | missing | pending | not_yet_observable | inconclusive`
- `source_authority_or_role`: 
- `transformation_or_method`: `none`
- `limitation`: 

### Observation 3 — điền sau khi tự mở source bằng trình duyệt

- `observation_id`: `obs-laptop-stand-20260904-03`
- `subject_id`: 
- `source_url`: 
- `observed_at`: 
- `access_method`: `manual_browser`
- `field_or_claim`: 
- `claim_kind`: `fact | estimate | assumption | unknown`
- `observed_value_or_state`: 
- `observed_state`: `observed | missing | pending | not_yet_observable | inconclusive`
- `source_authority_or_role`: 
- `transformation_or_method`: `none`
- `limitation`: 

## Unknowns / missing evidence (điều chưa biết / bằng chứng còn thiếu)

- conversion rate (tỷ lệ chuyển đổi): `unknown`
- valid-order rate (tỷ lệ đơn hợp lệ): `unknown`
- refund/cancel rate (tỷ lệ hoàn/hủy): `unknown`
- Affiliate commission thực nhận: `unknown`
- tracking reliability (độ tin cậy ghi nhận Affiliate): `unknown`
- freelancer-specific demand (nhu cầu riêng của freelancer): `unknown`
- product–audience fit thực tế: `unknown`

## Human DecisionPacket (gói quyết định do người lập)

> Chưa điền ở M00.2. Hoàn thiện phần này trong M00.3 sau khi có ít nhất 3 E1 public observations (quan sát công khai thật cấp E1).

- `decision_id`: 
- `question`: 
- `evidence_ids`: 
- `supported_facts`: 
- `assumptions`: 
- `unknowns`: 
- `state`: `RANK_SCENARIO | GET_MORE_DATA | HUMAN_REVIEW`
- `reason`: 
- `missing_evidence`: 
- `next_measurement`: 
- `action`: `null`

## Checklist M00.2 trước khi gửi review (kiểm tra)

- [ ] Tôi đã tự mở ít nhất 3 source công khai thật.
- [ ] Mỗi observation có `observation_id` riêng.
- [ ] `subject_id` nhận diện đúng offer/đối tượng.
- [ ] Có `source_url` thật, không phải placeholder (giữ chỗ).
- [ ] Có `observed_at` gồm timezone.
- [ ] `access_method = manual_browser` nếu tôi trực tiếp quan sát bằng trình duyệt.
- [ ] Chỉ ghi `fact` cho điều source hỗ trợ trực tiếp.
- [ ] Không đổi `missing` thành `0`.
- [ ] Có `source_authority_or_role`.
- [ ] Có `limitation` cho từng observation.
- [ ] Không commit credential/account/personal/customer data.

## Gợi ý limitation (giới hạn)

Có thể dùng và chỉnh lại cho đúng từng trường hợp:

```text
Giá có thể thay đổi theo thời điểm, voucher, tài khoản hoặc chương trình khuyến mãi.
```

```text
Số đã bán/rating là chỉ số nền tảng đang hiển thị; không chứng minh conversion rate hoặc Affiliate performance (hiệu quả Affiliate) của traffic của tôi.
```

```text
Thông tin vật liệu/tính năng là seller/listing claim (tuyên bố của người bán/trang sản phẩm), chưa được kiểm nghiệm độc lập.
```
