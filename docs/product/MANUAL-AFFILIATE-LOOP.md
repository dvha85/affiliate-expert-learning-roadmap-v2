# BR-06 — Từ link thủ công đến báo cáo có nguồn

Trạng thái: bản hướng dẫn **trung lập, SYNTHETIC**, chưa có chương trình hoặc kênh được chọn. Đây là bài tham khảo cho M03 sau các gate trong [CURRICULUM](../../CURRICULUM.md), không cấp Mission PASS. Không đăng ký tài khoản, tạo link thật hoặc đăng nội dung khi chạy bài offline.

## 1. Chuẩn bị khi đã có chương trình

Người học tự chọn chương trình mình được phép tham gia. Trước action thật, lưu một bản ghi riêng gồm: tên chương trình, trang hướng dẫn chính thức và ngày đọc, trạng thái tài khoản được chấp thuận, offer/sản phẩm, kênh được phép, cách tạo link, cách đo và độ trễ báo cáo. Nếu thiếu thông tin, ghi UNKNOWN và hỏi hỗ trợ chương trình; không suy điều kiện từ fixture này.

Mẫu ghi chú riêng (không commit thông tin tài khoản):

```text
program / official_terms_url / checked_at:
account_approved / permitted_channel / offer_id:
official_link_tool / account_ownership_check:
campaign_supported / campaign_field / allowed_values:
report_function / report_timezone / reporting_delay:
order_status_definitions / commission_currency / payout_status_definition:
reviewer / unresolved_questions:
```

Chưa có tài khoản: làm tiếp phần offline bên dưới; tất cả ô platform/live giữ chưa xác minh. Không điền tên chương trình giả để hoàn tất checklist.

## 2. Tạo và kiểm link — chỉ khi đã có quyền

Trong giao diện chính thức của chương trình đã chọn, mở offer và dùng chức năng tạo tracking link được tài liệu chương trình chỉ định. Lưu tham chiếu offer và cách kiểm link thuộc đúng tài khoản qua giao diện hoặc công cụ kiểm chính thức. Không tự ghép publisher ID, không sửa query bằng phỏng đoán và không coi URL mở đúng sản phẩm là bằng chứng attribution đúng.

Nếu chương trình hỗ trợ campaign/sub-ID, dùng nhãn không chứa dữ liệu cá nhân cho một action, lưu bảng `campaign → action_id`. Tuân thủ giới hạn ký tự/độ dài từ tài liệu của chương trình. Nếu không hỗ trợ hoặc báo cáo không xuất trường đó, ghi hạn chế attribution; không tự gán số liệu toàn tài khoản cho một action. UTM là nhãn phân tích khác, **không thay bằng chứng affiliate attribution**.

Không tự mua hàng hoặc tạo click/đơn thử để kiểm attribution; cách kiểm được phép phải được xác nhận từ chương trình. Repo không cung cấp tracking link thật. Địa chỉ `https://example.invalid/synthetic-post` trong fixture là placeholder không hoạt động.

## 3. Nội dung và ActionRecord

Ví dụ để người học sửa sau khi có nguồn: “Mình đang so sánh hai giá đỡ laptop; các điểm chưa kiểm chứng được ghi rõ trong báo cáo. Nếu bạn mua qua link affiliate, mình có thể nhận hoa hồng.” Đây là mẫu disclosure, không phải xác nhận đáp ứng quy định pháp lý hoặc chính sách của một kênh; người đăng cần review yêu cầu áp dụng trước khi đăng.

Chỉ ghi ActionRecord sau khi người thực sự thao tác. Giữ `performed_by: human`; `compliance_reviewed` chỉ true nếu đã review, không mặc định true để qua gate. `decision_id` phải trỏ quyết định đã review; quyết định GET_MORE_DATA không tự trở thành đề xuất bán hàng. Case offline dùng action đo thủ công giả lập, không khẳng định quyết định ban đầu cho phép đăng link bán hàng.

Lưu riêng URL/bằng chứng action, người review, tracking/campaign mapping và lý do chọn kỳ đo. Chọn window theo độ trễ nguồn đã xác minh, không coi 7 ngày trong fixture là quy tắc mọi chương trình. Khi chưa có action thật, chỉ dùng `syn-*`, không đổi PROGRESS.

## 4. Lấy báo cáo và giữ provenance

Trong chương trình đã chọn, xác định chức năng báo cáo/xuất dữ liệu từ tài liệu chính thức. Ghi tên chức năng và đường dẫn màn hình thực tế vào ghi chú phần 1; hiện repo **chưa có hướng dẫn màn hình của platform cụ thể**. Chọn kỳ đo, timezone và campaign phù hợp; lưu thời điểm xuất, bộ lọc, phạm vi tài khoản/kênh và trạng thái đơn. Thời điểm xuất không tự là thời điểm dữ liệu đã đầy đủ; cần biết data-as-of/độ trễ của nguồn.

Giữ bản gốc riêng tư và bản đã loại dữ liệu riêng tư cho reviewer được phép. Không commit tên/email/điện thoại/địa chỉ khách, cookie, token, publisher ID, tracking link riêng hoặc mã đơn có thể truy người mua. Dùng source_ref nội bộ để reviewer được phép tìm lại bản gốc; source_ref fixture không được đổi nhãn thành platform export thật.

## 5. Fixture và mapping đề xuất cho BR-10

[manual-loop.json](../../examples/affiliate-manual/manual-loop.json) chứa một action, ba snapshot báo cáo và ba OutcomeRecord kỳ vọng. Đây là **format trung gian do repo định nghĩa**, không phải export của nhà cung cấp, không phải importer đã triển khai. `null` ở report là chưa biết; OutcomeRecord không nhận metric null nên bỏ metric đó, tuyệt đối không thay bằng 0.

| Trường report trung gian | Đích / cách kiểm |
|---|---|
| `snapshot_id` | Khóa snapshot nguồn; retry cùng snapshot không được tạo sự kiện kinh doanh mới |
| `source_ref` | Outcome `source_ref`; chỉ resolve trong bộ fixture này |
| `action_id` | `effect_ref={effect_kind:HUMAN_ACTION,effect_id:action_id}`; phải resolve action |
| `observed_at` | Outcome `observed_at`, instant RFC3339 có offset; không bỏ timezone |
| `period_start`, `period_end` | Metadata nguồn; đối chiếu kỳ đo, không thêm field lạ vào canonical OutcomeRecord |
| `status` | Enum trung gian đã giải thích; status platform thật cần bảng mapping riêng được review |
| `clicks`, `valid_orders` | Metric cùng tên chỉ khi có số; không suy valid_orders từ tổng đơn phát sinh |
| `commission_paid_vnd` | Chỉ số tiền đã thanh toán, VND; không map commission dự kiến thành đã trả |
| `supersedes` | Quan hệ snapshot cập nhật trong metadata; không overwrite OutcomeRecord cũ |

Mỗi snapshot có Outcome ID riêng trong fixture. Metadata provenance/kỳ đo không được bỏ mất ở importer tương lai dù không nằm trong schema OutcomeRecord. Schema hợp lệ không xác minh liên kết ID, dữ liệu đầy đủ hoặc quyền sử dụng nguồn; BR-10 phải kiểm semantic trước khi persist.

## 6. Đọc đúng trạng thái

- Báo cáo tải lỗi, chưa có dữ liệu hoặc rỗng không rõ độ đầy đủ: chưa kết luận số 0. Dừng import hoặc ghi quan sát PENDING với metrics chưa biết và limitation ở provenance.
- PENDING trước window end: quan sát tạm, không đóng kỳ. Fixture có clicks=0 nhưng valid_orders chưa biết.
- NO_OBSERVED_OUTCOME: chỉ dùng sau/tại window end và có nguồn xác nhận không quan sát kết quả trong phạm vi đo; không suy từ file rỗng. Fixture số 0 có đủ metadata giả lập để minh họa, không phải nguồn thật.
- VALID: giao dịch đã được nguồn xác nhận hợp lệ, không đồng nghĩa PAID. CANCELLED/REFUNDED phải giữ trạng thái tương ứng, không biến thành valid_orders hoặc commission âm.
- PAID: chỉ khi nguồn xác nhận đã thanh toán; tách khỏi commission dự kiến/chờ duyệt. Không cộng nhiều loại tiền vào cùng metric.
- Báo cáo đến muộn: append snapshot/outcome mới, giữ liên kết cập nhật. Chọn snapshot phù hợp kỳ đánh giá; không cộng ba snapshot tích lũy của cùng kỳ để tính tổng.

## 7. Bài chạy offline và lỗi có chủ đích

Từ repo root (đã cài công cụ theo [quickstart](../../curriculum/BOOT/QUICKSTART.md)):

```text
cd contracts
go test . -run TestManualAffiliateFixture -count=1 -v
```

Kỳ vọng PASS: action và ba outcome đúng canonical schema, mapping số 0/missing giữ nguyên, IDs/kỳ đo trong fixture khớp. Test cũng thử đổi observed_at của kết luận zero về trước window end và thay null thành zero để chắc verifier fixture phát hiện. Đây là **verifier bài mẫu trong test**, không phải runtime importer/store hoặc bằng chứng affiliate thật.

Trong bản sao bài tập, đổi `outcomes[0].metrics` từ `{"clicks":0}` sang `{"clicks":0,"valid_orders":0}` rồi chạy lại: phải FAIL vì report chưa có valid_orders. Khôi phục riêng sửa đổi này và chạy về PASS. Không sửa expected để che lỗi.

Reviewer cần giải thích được: vì sao null khác 0; vì sao window đủ chưa chắc báo cáo đã đầy đủ; vì sao PAID không là commission dự kiến; vì sao không cộng ba snapshot. Lưu output test và câu trả lời, không tự đánh dấu Mission PASS.

## 8. Phần còn mở

BR-06a: tài liệu trung lập + fixture/mapping có test, chờ review. BR-06b: hướng dẫn một chương trình/kênh cụ thể và case link → báo cáo thật, BLOCKED do chưa có lựa chọn/quyền truy cập. BR-10 chưa có importer; fixture này chỉ là đầu vào thiết kế, không chứng minh pipeline đã chạy.
