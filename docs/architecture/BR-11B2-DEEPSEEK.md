# BR-11b.2 — DeepSeek V4 Flash

Người dùng chọn DeepSeek V4 Flash, cho phép tối đa 100 lượt thử với tổng ngân sách $3, chỉ dữ liệu giả lập. Quyền này không cho phép gửi dữ liệu kinh doanh thật hoặc bật thực thi.

## Bước 1: adapter giao thức (đã merge)

PR #58 đã review/merge tại `dbeda41`, head `ec89ac2`; CI 4/4 PASS, learner tests/race/vet chạy lại PASS. Nghiệm thu chỉ adapter nội bộ và tests offline, chưa cho phép CLI live.

`advisor_deepseek.go` triển khai interface provider nội bộ, endpoint HTTPS cố định `https://api.deepseek.com/chat/completions`, model `deepseek-v4-flash`, prompt version `deepseek-human-review/v1`. Không đọc khóa từ file/repo, chỉ `DEEPSEEK_API_KEY`. Không proxy, redirect hoặc retry tự động; timeout 30 giây, context JSON tối đa 8192 byte, output tối đa 1024 token, body phản hồi tối đa 64 KiB. Thinking tắt, JSON mode bật, không gửi tools.

Chỉ nhận một completion của đúng model, role assistant, finish_reason stop, không tool calls và có usage hợp lệ. Nội dung vẫn qua boundary chung: schema, references, freshness, HUMAN_REVIEW/ABSTAIN; không execution. Lỗi không trả raw body/transport/secret. Kiểm tra secret echo cả trên content đã decode. Usage được giữ nội bộ kể cả khi nội dung bị từ chối; chưa là ledger chi phí.

Regression server loopback kiểm tra cấu hình request, phản hồi hợp lệ/rỗng/cắt cụt, sai model, tools, thiếu usage, auth, không retry 503, sai schema, secret echo và giới hạn dữ liệu. Không gọi API thật trong tests hay CI.

## Còn thiếu trước khi chạy live

Bước 2a đang triển khai: `advisor_budget.go` có ledger reservation nội bộ và runner fixture cố định. Khởi tạo phải gọi riêng, không tự tạo lại khi thiếu; mkdir lock loại trừ writer đồng thời, reservation file O_EXCL + fsync file/directory trước khi provider được gọi. Restart đọc lại các file liên tục và manifest exact. File hỏng/thiếu giữa chuỗi hoặc lock còn lại sau crash đều dừng; không tự sửa/reset. Không hoàn tiền dự phòng và không retry.

Review #59 phát hiện đọc manifest/reservation chưa giới hạn kích thước và có thể treo với FIFO. Bản sửa kiểm regular file bằng Lstat trước khi mở, giới hạn size theo độ dài canonical và đọc tối đa limit+1. Regression kiểm manifest và reservation quá lớn, FIFO trên Linux/macOS. Không bảo vệ khỏi tác nhân local thay thế path đồng thời; giới hạn trusted directory vẫn giữ nguyên. #59 chưa được nghiệm thu chỉ dựa trên CI của head cũ.

Reservation tạm thời là 500.000 microUSD ($0,50)/lượt; cap $3 nên chỉ cho tối đa 6 lượt, dù quyền người dùng là tối đa 100. Đây là chính sách bảo thủ để test ledger, **không phải bằng chứng về trần phí thực tế của provider**. Chưa có đối soát usage/giá, chưa có báo cáo kết quả lưu bền và chưa có CLI live. Không được bật live dựa riêng vào ledger này.

Giới hạn ledger: dành cho thư mục local tin cậy, không chống người dùng xóa/rollback toàn bộ dữ liệu hoặc tạo campaign mới ở đường dẫn khác. Chưa bảo vệ đường dẫn khỏi tác nhân local độc hại; trước live phải chốt một campaign path do ứng dụng sở hữu và kiểm tra quyền/identity. Runner nội bộ không nhận context tùy ý, chỉ fixture synthetic. Tests kiểm restart, budget exhaustion, corruption, stale lock, contention và failure vẫn tiêu thụ reservation.

- Ledger lưu bền, khóa đồng thời, giữ chỗ ngân sách/lượt trước khi gửi; lỗi không rõ kết quả vẫn tiêu thụ lượt và giữ dự phòng chi phí. Không reset ledger khi khởi động lại.
- Runner chỉ dùng fixture do repo tạo; không nhận history/context tùy ý để gửi ra ngoài.
- Ghi usage, phiên bản model trả về, giá theo thời điểm và kết quả kiểm chứng; đối soát khoản đã dùng. Mức giá/cách đếm phải được kiểm tra trước khi bật live, không tuyên bố trần $3 chỉ dựa vào giới hạn token.
- Kiểm tra khả năng truy cập khóa mà không in giá trị, chạy canary trước, dừng nếu lỗi; không bắt buộc dùng hết 100 lượt.

Chưa mở lệnh CLI DeepSeek, chưa gọi live, chưa nghiệm thu BR-11b.2. Không dùng constructor nội bộ để bỏ qua các điều kiện trên.

Nguồn giao thức: [Chat Completions](https://api-docs.deepseek.com/api/create-chat-completion/), [JSON Output](https://api-docs.deepseek.com/guides/json_mode/), [giá](https://api-docs.deepseek.com/quick_start/pricing/). Model alias có thể được cập nhật bởi nhà cung cấp; cần chạy lại eval khi thay đổi.
