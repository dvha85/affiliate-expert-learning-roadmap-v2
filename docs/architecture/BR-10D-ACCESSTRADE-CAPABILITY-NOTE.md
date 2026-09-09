# BR-10d — ACCESSTRADE Publisher capability note

Trạng thái: **PARTIAL / read-only observation**. Ghi chú này mô tả những gì
chủ repo đã cho xem trong dashboard Publisher ngày 08/09/2026. Nó không chứa
credential, publisher ID, tracking link, mã đơn thật hoặc ảnh chụp có dữ liệu
cá nhân; không phải chứng cứ conversion, payout, campaign approval hay quyền
automation.

## Capability đã quan sát

| Khu vực | Capability quan sát | Cách dùng trong repo | Không được suy ra |
|---|---|---|---|
| Công cụ | Product Link, Convert Link, Deep Link, Script Direct Link | Chỉ ghi nhận rằng platform có các loại link này | Link đã thuộc campaign cụ thể hoặc attribution hoạt động |
| Báo cáo tổng quan | click, CVR, EPC, conversion, giá trị conversion, hoa hồng; lọc thời gian/campaign | Vocabulary cho docs và future operator review | Số 0 khi dashboard chưa có dữ liệu là `NO_OBSERVED_OUTCOME` |
| Báo cáo đơn hàng | thời gian mua, mã đơn, trạng thái, website, danh mục, giá trị đơn, hoa hồng, advertiser, ngày duyệt dự kiến, lý do hủy, UTM, thiết bị/platform; có nút tải file | CSV adapter nhận bốn cột: `Mã đơn`, `Trạng thái`, `Giá trị đơn hàng`, `Hoa hồng` | Đã tải hoặc đã review một export thật |
| Báo cáo UTM | lọc/source/medium/content/campaign; tổng click, CVR, chuyển đổi, giá trị conversion, hoa hồng | Phân tích kênh sau này | UTM là proof affiliate attribution hay được phép map tự động tới action |

## Boundary của adapter

[`accesstrade-import`](../../lab/affiliate-bot/cmd/bot/accesstrade_import.go)
nhận CSV UTF-8 đã khử dữ liệu riêng tư cùng manifest tường minh. Manifest map
`order_id` đã ẩn danh tới một `action_id` đã tồn tại; không đọc UTM để tạo
EffectRef.

| Giá trị report | Mapping hiện tại | Lý do |
|---|---|---|
| `Chờ xử lý`, `Tạm duyệt` | `PENDING` | Chưa là kết luận chốt hoặc payout |
| `Được duyệt` | `VALID` | Được ghi nhận, nhưng không đồng nghĩa payment |
| `Từ chối` | `CANCELLED` | Giữ kết quả âm, không sửa thành zero/missing |
| Hoa hồng report | `reported_commission_vnd` | Không gọi là `commission_paid_vnd` nếu chưa có evidence payment |
| Status khác | reject | Không tự đoán `REFUNDED` hoặc `PAID` |

Snapshot được validate hết trước khi ghi atomically vào outcome store. Retry
chỉ được chấp nhận khi **toàn bộ** snapshot là exact duplicate; một batch vừa
cũ vừa mới bị từ chối để không che việc thay đổi dữ liệu nguồn.

Mỗi snapshot mới còn có một receipt bất biến trong
`accesstrade-receipts.jsonl`: receipt bind `snapshot_id`, `source_ref`, thời
điểm, tiền tệ, hash SHA-256 của CSV/manifest và các canonical outcome/action
IDs. Nó không lưu order ID, customer data hay tracking URL. CLI
`outcome accesstrade-receipts` resolve lại history/action/outcome và fail nếu
receipt mồ côi, thiếu outcome, khác store hoặc có outcome ACCESSTRADE thiếu
receipt. Một journal pending được giữ fail-closed khi commit hai store bị ngắt.
Backup/restore runtime tiêu chuẩn cũng derive receipt requirement từ
`outcomes.jsonl`, nên không thể sửa manifest backup để bỏ receipt mà vẫn
restore thành công. Đây là kiểm integrity/replay của fixture, **không** là
chứng minh dữ liệu report là thật.

## Điều còn thiếu trước operated evidence

1. Campaign và kênh được ACCESSTRADE duyệt.
2. Export thật đã khử dữ liệu riêng tư, timezone/filter/export time và status
   dictionary được reviewer kiểm.
3. Cách platform hiển thị `PAID`, `REFUNDED`, correction/late update; adapter
   hiện fail-closed với các status này.
4. Mapping có review từ hành động người thật đến report scope. UTM không thay
   mapping đó.

Cho đến lúc ấy, fixture và smoke chỉ chứng minh behavior của learner Bot. Chúng
không thay business outcome hoặc thay đổi `PROGRESS.md`.
