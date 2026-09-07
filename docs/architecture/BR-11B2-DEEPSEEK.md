# BR-11b.2 — DeepSeek V4 Flash

Người dùng chọn DeepSeek V4 Flash, cho phép tối đa 100 lượt thử với tổng ngân sách $3, chỉ dữ liệu giả lập. Quyền này không cho phép gửi dữ liệu kinh doanh thật hoặc bật thực thi.

## Bước 1: adapter giao thức (chờ review)

`advisor_deepseek.go` triển khai interface provider nội bộ, endpoint HTTPS cố định `https://api.deepseek.com/chat/completions`, model `deepseek-v4-flash`, prompt version `deepseek-human-review/v1`. Không đọc khóa từ file/repo, chỉ `DEEPSEEK_API_KEY`. Không proxy, redirect hoặc retry tự động; timeout 30 giây, context JSON tối đa 8192 byte, output tối đa 1024 token, body phản hồi tối đa 64 KiB. Thinking tắt, JSON mode bật, không gửi tools.

Chỉ nhận một completion của đúng model, role assistant, finish_reason stop, không tool calls và có usage hợp lệ. Nội dung vẫn qua boundary chung: schema, references, freshness, HUMAN_REVIEW/ABSTAIN; không execution. Lỗi không trả raw body/transport/secret. Kiểm tra secret echo cả trên content đã decode. Usage được giữ nội bộ kể cả khi nội dung bị từ chối; chưa là ledger chi phí.

Regression server loopback kiểm tra cấu hình request, phản hồi hợp lệ/rỗng/cắt cụt, sai model, tools, thiếu usage, auth, không retry 503, sai schema, secret echo và giới hạn dữ liệu. Không gọi API thật trong tests hay CI.

## Còn thiếu trước khi chạy live

- Ledger lưu bền, khóa đồng thời, giữ chỗ ngân sách/lượt trước khi gửi; lỗi không rõ kết quả vẫn tiêu thụ lượt và giữ dự phòng chi phí. Không reset ledger khi khởi động lại.
- Runner chỉ dùng fixture do repo tạo; không nhận history/context tùy ý để gửi ra ngoài.
- Ghi usage, phiên bản model trả về, giá theo thời điểm và kết quả kiểm chứng; đối soát khoản đã dùng. Mức giá/cách đếm phải được kiểm tra trước khi bật live, không tuyên bố trần $3 chỉ dựa vào giới hạn token.
- Kiểm tra khả năng truy cập khóa mà không in giá trị, chạy canary trước, dừng nếu lỗi; không bắt buộc dùng hết 100 lượt.

Chưa mở lệnh CLI DeepSeek, chưa gọi live, chưa nghiệm thu BR-11b.2. Không dùng constructor nội bộ để bỏ qua các điều kiện trên.

Nguồn giao thức: [Chat Completions](https://api-docs.deepseek.com/api/create-chat-completion/), [JSON Output](https://api-docs.deepseek.com/guides/json_mode/), [giá](https://api-docs.deepseek.com/quick_start/pricing/). Model alias có thể được cập nhật bởi nhà cung cấp; cần chạy lại eval khi thay đổi.
