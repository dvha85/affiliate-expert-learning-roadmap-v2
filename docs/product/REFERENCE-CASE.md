# Case BR-04 — Giá đỡ laptop (toàn bộ giả lập)

**SYNTHETIC / KHÔNG PHẢI BẰNG CHỨNG THẬT.** Không có merchant, chương trình affiliate, tracking link hoặc giao dịch thật. Bảng dưới là ví dụ nội dung để thiết kế bot; không phải JSON contract đầy đủ hay output runtime đã chạy.

## Câu hỏi

Người làm việc tại nhà cần so sánh hai loại giá đỡ laptop. Ta biết giá giả lập nhưng chưa có commission đã xác nhận, tỷ lệ chuyển đổi hoặc bằng chứng về nhu cầu. Bot nên nêu phần còn thiếu, không tự khuyên mua hay hứa khả năng sinh lời.

| Observation ID | Subject ID | Giá trị giả lập | Thời điểm | Nguồn / giới hạn |
|---|---|---|---|---|
| syn-obs-a-01 | syn-offer-a | Giá 250.000 VND | 03/09/2026 10:00 +07:00 | fixture-a-v1; không phải giá bán thực tế |
| syn-obs-b-01 | syn-offer-b | Giá 350.000 VND | 03/09/2026 10:00 +07:00 | fixture-b-v1; không phải giá bán thực tế |
| syn-obs-a-02 | syn-offer-a | Commission chưa biết | 03/09/2026 10:05 +07:00 | fixture-a-terms-v1; không có chương trình nào được xác nhận |

Commission thiếu không được ghi 0. Hai lần quan sát offer A dùng cùng subject_id nhưng khác observation_id. Đây chỉ là source_ref fixture, không được thay bằng URL giả để claim E1.

## Báo cáo mục tiêu

- decision_id: syn-dec-01; evidence_ids: syn-obs-a-01, syn-obs-b-01, syn-obs-a-02.
- state: GET_MORE_DATA; chưa đủ cơ sở xếp hạng hiệu quả affiliate.
- Điều biết trong fixture: hai mức giá giả lập khác nhau; không suy chất lượng hoặc nhu cầu từ giá.
- Điều chưa biết: điều kiện chương trình, commission, click/conversion, chi phí hợp lệ và kỳ xác nhận.
- Next measurement: khi có chương trình và nguồn hợp lệ, ghi điều kiện offer và cách xuất báo cáo; không tự crawl hoặc tạo tracking URL.
- Action: null trong DecisionPacket; hành động thực tế vẫn cần người review và thực hiện theo Mission.

## Vòng action/outcome/evaluation giả lập

Giả sử bài tập offline có `syn-act-01` tham chiếu syn-dec-01, performed_at 03/09/2026 11:00 +07:00, window end 10/09/2026 11:00 +07:00. Đây **không phải action đã thực hiện** và không đủ Operated PASS.

| Outcome ID | Thời điểm | Status/metrics giả lập | Cách hiểu |
|---|---|---|---|
| syn-out-01 | 03/09 12:00 +07:00 | PENDING; clicks=0 đã mô phỏng đo | Không chốt kết quả cả window; thiếu số đơn thì không điền 0 |
| syn-out-02 | 10/09 11:00 +07:00 | NO_OBSERVED_OUTCOME; clicks=0, valid_orders=0 | Ví dụ số 0 tại end; không biến thành missing, không claim số liệu thật |
| syn-out-03 | 12/09 11:00 +07:00 | PAID; valid_orders=1 trong báo cáo cập nhật giả lập | Record mới bổ sung/đính chính syn-out-02; không overwrite hoặc cộng trùng kỳ |

Cả ba record dùng `effect_ref={effect_kind:HUMAN_ACTION,effect_id:syn-act-01}`. Nếu NO_OBSERVED_OUTCOME của syn-out-02 bị đổi timestamp về 03/09 12:00 thì validator phải trả MEASUREMENT_WINDOW_OPEN. Timestamp khác offset nhưng cùng instant không được làm lệch boundary.

`syn-eval-01` tham chiếu đúng syn-dec-01, syn-act-01 và các outcome IDs dùng cho kỳ đánh giá. Kết luận hợp lý của bài tập là INCONCLUSIVE về hiệu quả kinh doanh vì dữ liệu hoàn toàn giả lập. `syn-proposal-01` có thể đề xuất làm rõ nguồn báo cáo; phải có review riêng, không auto_apply.

## Người review cần trả lời được

1. Bot đang hỗ trợ câu hỏi nào và phần dữ liệu nào còn thiếu?
2. Từ một kết luận có truy lại được observation/source/time/limitation không?
3. Vì sao PENDING/số 0/báo cáo đến muộn không được gộp thành cùng một trạng thái?
4. Vì sao fixture PASS vẫn chưa có bằng chứng doanh thu, platform compliance hoặc action thật?

Chuyển case này sang fixture JSON chạy được xuyên suốt là việc BR-09–BR-12; không claim đã có pipeline đó trong BR-04.
