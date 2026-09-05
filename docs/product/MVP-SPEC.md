# BR-04 — Đặc tả MVP Affiliate Intelligence Bot

- Trạng thái: IN_REVIEW — đặc tả mục tiêu, không tuyên bố MVP đã được triển khai đầy đủ.
- Người thực hiện: Codex. Reviewer: chủ repo, chờ review.
- Quyết định đầu vào: chủ repo xác nhận **chưa có chương trình affiliate**. Phiên bản này dùng fixture trung lập; chưa chọn chương trình, kênh đăng, tài khoản hay API.
- Case tham chiếu: giá đỡ laptop cho người làm việc tại nhà. Các offer, giá và chỉ số mẫu bên dưới đều giả lập, không phải tư vấn mua hàng hoặc nguồn thu dự kiến.
- Thứ tự học, evidence và authority tuân theo [CURRICULUM](../../CURRICULUM.md); theo dõi triển khai tại [kế hoạch BR](../plans/BEGINNER-READINESS-PLAN.md).

## 1. Ai dùng và bot giúp việc gì?

Người mới làm affiliate muốn so sánh một nhóm offer, biết mình còn thiếu dữ liệu nào và theo dõi kết quả của **hành động do người thực hiện**. Bot giúp tạo báo cáo có nguồn/thời điểm/giới hạn, lưu lịch sử có thể kiểm lại và nối quyết định với action/outcome/evaluation.

Câu hỏi xuyên suốt: “Với dữ liệu đang có, offer nào đáng thu thập thêm bằng chứng hoặc đưa ra cho người review; cần đo thêm gì trước khi quyết định?” Bot không hứa chọn offer bán chạy nhất hoặc tối đa hóa thu nhập.

Đầu ra MVP đầu tiên là báo cáo xếp hạng có giới hạn và các câu hỏi còn thiếu. Vòng tiếp theo ghi nhận hành động thủ công, kết quả đo rồi đề xuất cải thiện cần người review. Không coi một demo in SUPPORTED là đã có sản phẩm này.

## 2. Phạm vi và giới hạn quyền

Trong MVP: nhập JSON/CSV qua adapter có contract; kiểm dữ liệu; xếp hạng deterministic; lưu/replay history; liên kết action/outcome; báo cáo và evaluation. Advisor mock dùng cho Capability; khi có provider mới thêm live proof riêng.

Ngoài MVP ban đầu: tự đăng bài/gửi tin/spend, tạo link affiliate bằng API giả định, tự đăng ký chương trình, crawl nguồn không được phép, tự áp dụng improvement hoặc khẳng định attribution/doanh thu khi chưa có báo cáo. Watcher chỉ đọc thuộc M06; executor thuộc gate M09+, không được dùng để né action của người ở M03.

M00–M05 phải tiến hóa cùng learner Bot tại `lab/affiliate-bot/`; `lab/mission-runtime/` chỉ là oracle/eval. CLI/adapter/store xuyên suốt vẫn là công việc BR-08–BR-12, không phải chức năng đã có chỉ vì đặc tả này được merge.

## 3. Input, nguồn và trạng thái dữ liệu

| Input | Dữ liệu tối thiểu / nơi lấy | Khi chưa có |
|---|---|---|
| Câu hỏi/ngách | Câu hỏi của người, audience, metric và giả định | Không tự suy thành nhu cầu thị trường đã chứng minh |
| Observation offer | observation_id, subject_id, source_url hoặc source_ref, observed_at, access_method, claim_kind, evidence_kind, state, limitation; domain fields như giá/currency/commission | Dùng fixture gắn synthetic; thiếu commission giữ unknown/missing, không điền 0 |
| Decision/history | decision_id, exact evidence_ids, as_of, formula_version, input_hash, recorded_result | Không tạo ID tham chiếu tới evidence không tồn tại |
| Link/campaign | Link chương trình chính thức và mã campaign/sub-ID nếu được chương trình hỗ trợ | Chưa có chương trình: để chưa khả dụng; không dùng link giả như tracking thật |
| Action người thực hiện | action_id, decision_id, performed_by=human, target, performed_at, measurement_window_end, compliance_reviewed | Fixture chỉ mô phỏng; không ghi compliance thật hoặc action thật chưa xảy ra |
| Báo cáo outcome | outcome_id, effect_ref, observed_at, status, metrics, source_ref; kỳ báo cáo/giới hạn nguồn trong evidence | Chưa có nguồn: PENDING/unknown phù hợp; không biến thiếu thành 0 |
| Evaluation/review | evaluation_id, decision_id, cùng effect_ref, outcome_ids, result/limitations; proposal/review IDs | Không tự apply hoặc kết luận thành công từ chỉ số mô phỏng |

Các cột trên là mô tả yêu cầu, không thay đầy đủ required fields trong `contracts/*.schema.json`. Dữ liệu tài khoản/cá nhân giữ ngoài repo; fixture không chứa credential, mã khách hàng, đơn hàng thật hoặc private URL.

## 4. Output người mới cần nhìn thấy

Một báo cáo nên có:

1. Câu hỏi, thời điểm `as_of`, nguồn đầu vào và nhãn synthetic/real rõ ràng.
2. Các offer được so sánh, lý do/thứ tự xếp hạng theo công thức đã version; chưa đủ dữ liệu thì nêu trường còn thiếu.
3. Mỗi kết luận có evidence IDs và source/time/limitation truy ngược được; phân biệt fact, estimate, assumption, unknown.
4. State có giới hạn như GET_MORE_DATA/HUMAN_REVIEW/RANK_SCENARIO; action không tự được cấp quyền.
5. Việc người cần review và phép đo tiếp theo; không có lời hứa doanh thu.
6. Khi đã có action/outcome: dòng thời gian ID, trạng thái tạm/chốt, phiên bản báo cáo và evaluation dựa trên những outcome IDs nào.

Xem [case giả lập xuyên suốt](REFERENCE-CASE.md). Đây là mẫu nội dung mong muốn, **không phải ảnh chụp output CLI hiện tại**. BR-08–BR-12 sẽ quyết định lệnh và format lưu trữ, không tự phát minh lệnh chạy chưa tồn tại.

## 5. ID và liên kết xuyên suốt

| Liên kết | Quy tắc phải giữ |
|---|---|
| subject → observations | Một subject ổn định có nhiều observation_id cho các lần quan sát khác thời điểm |
| observations → decision/history | evidence_ids là các observation ID thực được dùng, không trỏ tới tên sản phẩm chung |
| decision → human action | ActionRecord.decision_id trỏ exact quyết định được người xem xét |
| action → outcomes | effect_ref.effect_kind=HUMAN_ACTION; effect_id=ActionRecord.action_id |
| outcomes → evaluation | Cùng effect_ref; outcome_ids chọn snapshot/kỳ phù hợp, không cộng trùng bản cập nhật |
| evaluation → proposal → review | ID truy ngược được; improvement không auto_apply, review do người thực hiện |

M02 `recorded_result` hiện là projection của ranking, không được gọi là canonical DecisionPacket nếu chưa có adapter hợp lệ (BR-03). Không đổi lịch sử hoặc ID để làm bằng chứng trông hợp lệ.

## 6. Tiêu chí nghiệm thu MVP (mục tiêu triển khai)

Đây là acceptance criteria cho các BR triển khai, chưa đánh dấu PASS trong BR-04:

- Người mới theo quickstart nhập fixture từ file, chạy bot và tạo báo cáo mà không sửa hard-code trong demo.
- Thay input tạo báo cáo khác có giải thích; input sai schema/ID/freshness bị chặn hoặc trả trạng thái giới hạn, không âm thầm SUPPORTED.
- Tắt/chạy lại vẫn đọc được history; replay giữ formula/version/input hash và phân biệt MATCH với business truth.
- Ghi action do người thực hiện, nối outcome đúng effect; NO_OBSERVED_OUTCOME quá sớm bị chặn, số 0 thực đo tại/sau window được giữ.
- Báo cáo cập nhật thêm record mới, giữ bản cũ và không cộng trùng snapshot; evaluation chỉ rõ record đã dùng.
- Không có secret/dữ liệu cá nhân trong fixture/log được commit; không có external write tự động.
- Có tests tích hợp, failure cases và pilot người mới (BR-16), không chỉ unit test của oracle.

## 7. Đo kỹ thuật khác đo kinh doanh

Kỹ thuật: tỷ lệ input hợp lệ xử lý thành công, lỗi bị phát hiện đúng, truy xuất evidence/ID, replay ổn định và người mới hoàn tất luồng theo hướng dẫn. Fixture dùng để đo các tiêu chí này.

Kinh doanh: click, đơn hợp lệ/hủy/hoàn, commission được xác nhận/chi trả, kỳ đo và giới hạn attribution từ nguồn. Chưa có chương trình thì các KPI này **chưa đo được**; số liệu synthetic không được dùng như doanh thu, ROI hoặc conversion proof. Không đặt mục tiêu thu nhập làm điều kiện code PASS.

## 8. Khi có chương trình affiliate

Chủ repo chọn chương trình và kênh có quyền sử dụng; kiểm điều kiện tham gia, cách lấy link/báo cáo, disclosure, quyền riêng tư và quy định hiện hành từ nguồn chính thức (BR-06). Chỉ sau khi xác minh mới viết hướng dẫn platform-specific hoặc adapter.

Chưa chọn không chặn học bằng fixture hoặc viết code trung lập; nhưng chặn tuyên bố link thật, attribution thật và Reality/Operated proof phụ thuộc tài khoản. Không mua dịch vụ, tạo tài khoản hay đăng nội dung chỉ dựa vào việc phê duyệt đặc tả này.
