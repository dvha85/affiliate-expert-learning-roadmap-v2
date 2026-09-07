# BR-11c — Nghiệm thu adapter và canary trên fixture BR-10

## Phạm vi và nguồn bằng chứng

Baseline review: `2e84d7f`; runtime canary từ PR #67, head `0871d93`, merge `4068979`. Người vận hành là chủ repo; Codex review báo cáo và ảnh do người dùng cung cấp trong cuộc hội thoại. Codex không tự chạy request live hoặc đọc key. Không có commit/hash binary từ Terminal người vận hành để chứng minh độc lập bản chạy; đánh giá dựa trên báo cáo được cung cấp và đối chiếu code đã merge.

Tài liệu này là bản chép lại có chọn lọc, không phải export hóa đơn hoặc raw HTTP response. Không lưu ảnh tài khoản, key, đường dẫn cá nhân hoặc thay ledger. Ảnh dashboard người dùng gửi sau reload hiển thị Time=Today, API Key=All; không có timestamp/request ID để nối độc lập tới từng giao dịch. Tổng usage khớp canary nhưng không là chứng minh mật mã.

## Kết quả người vận hành cung cấp

- `campaign-init`: INITIALIZED.
- Report trước chạy: OK, attempts=0, reserved=0, không có result/missing/unknown.
- `campaign-canary`: attempt=1, ABSTAIN, execution_permitted=false.
- Report sau chạy: OK, attempts=1; missing_result_attempts=[]; unknown_usage_attempts=[].
- Provider `deepseek-chat-completions/v1`; model `deepseek-v4-flash`; prompt `deepseek-human-review/v1`.
- Context `br11a-context/v1`, SHA256 `9260065e0ccfe2789ed0b5b4309513444fa2fdcd9777448b7a24745838674fda`; result `br11-result/v2`.
- Usage: prompt=992, completion=158, total=1150.
- Reservation: 500000 microUSD ($0.50); còn 2500000 microUSD là hạn mức reservation, không phải số dư provider.
- Estimate: 646 microUSD ($0.000646), theo snapshot `deepseek-flash-peak-cache-miss-2026-09-07`; invoice_reconciled=false.
- Dashboard sau reload: 1 API request, 1,150 tokens, cost < $0.01 USD. Ảnh trước reload từng hiển thị zero; không xác định được nguyên nhân hoặc độ trễ từ ảnh.

## Review nội dung

Output ABSTAIN, recommendation “Cần người kiểm tra”, reason “Bằng chứng hiện có là giả lập và thiếu dữ liệu thực về kết quả, hiệu quả; không đủ cơ sở để kết luận”. Hai unknowns: chưa có dữ liệu thực tế về doanh thu/lợi nhuận sau hành động; kết quả còn PENDING, chưa có số liệu đo lường.

Expected evidence: một decision RANK_SCENARIO từ observation synthetic; human action có target `fixture:none`; outcome PENDING có metrics rỗng. Output tham chiếu đúng `br11-decision`, `br11-observation`, `br11-action`, `br11-outcome`; không biến pending thành zero, không coi score=8 là doanh thu thật. Không đề xuất thực thi, write_tool_requested=false và execution_permitted=false. ABSTAIN phù hợp câu hỏi về bằng chứng thiếu trước khi kết luận hiệu quả.

Giới hạn: prompt/context đã chủ động giới hạn HUMAN_REVIEW/ABSTAIN; context còn câu “Mock only”. Canary do đó chứng minh transport/provider thật trả output qua boundary trên fixture, không chứng minh model tự phân biệt dữ liệu thật/giả hoặc chất lượng tư vấn trên các tình huống khác. Không thay context đã đóng băng hoặc chạy thêm để làm đẹp kết quả.

## Ma trận nghiệm thu

| Tiêu chí BR-11 | Bằng chứng | Kết luận |
|---|---|---|
| Mock offline, không cần key | #56/#57 và smoke BR-11a | Đạt |
| Adapter thực, cấu hình không secret, provenance | #58/#67; báo cáo canary trên | Đạt trong fixture |
| Context từ history/action/outcome BR-10 | #60/#65/#66; đúng digest và bốn IDs | Đạt |
| Schema/grounding/freshness và cấm write | Regression #60; accepted-output kiểm lại #66 | Đạt offline cho các case lỗi, không gọi lỗi live |
| Timeout hữu hạn, không retry mơ hồ, lỗi cấu hình | #58/#67; preflight, lock, first-attempt-only tests | Đạt |
| Nội dung được review, exact ID không đồng nghĩa đúng | Review ABSTAIN ở trên | Đạt cho một câu hỏi, không quality benchmark |
| Usage phía provider | Dashboard 1 request/1150 tokens khớp report | Đã đối chiếu tổng usage |
| Phí | Dashboard < $0.01, estimate $0.000646 | Chỉ xác nhận mức dưới một cent; chưa phí chính xác |

## Quyết định và phần giữ mở

Đã chốt BR-11 DONE **chỉ phạm vi lab/schema + một live canary trên fixture synthetic**, sau #68 merged `395c728`, head `720fbd9`; CI 4/4 và 8 validators PASS, Codex review theo quyền chủ repo. Không yêu cầu thêm lượt chỉ để hoàn tất checklist. Cho phép bắt đầu BR-12 ở cùng phạm vi lab; không mở quyền thực thi hoặc mở rộng ngân sách.

Giữ invoice_reconciled=false trong ledger: ảnh không đủ precision để xác nhận estimate chính xác. Nếu cần quyết toán chính xác, cần export theo key/khoảng thời gian; không coi đó là blocker của nghiệm thu tích hợp lab. Không sửa metadata canonical để giả lập đối soát.

BR-06b/BR-10d vẫn cần chương trình/kênh/export thật; hiệu quả kinh doanh, chất lượng model tổng quát, triển khai production và pilot người mới chưa nghiệm thu. Không reset campaign; lượt tiếp theo vẫn bị REVIEW_REQUIRED, cần review riêng.
