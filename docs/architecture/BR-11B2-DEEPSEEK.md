# BR-11b.2 — DeepSeek V4 Flash

Người dùng chọn DeepSeek V4 Flash, cho phép tối đa 100 lượt thử với tổng ngân sách $3, chỉ dữ liệu giả lập. Quyền này không cho phép gửi dữ liệu kinh doanh thật hoặc bật thực thi.

## Bước 1: adapter giao thức (đã merge)

PR #58 đã review/merge tại `dbeda41`, head `ec89ac2`; CI 4/4 PASS, learner tests/race/vet chạy lại PASS. Nghiệm thu chỉ adapter nội bộ và tests offline, chưa cho phép CLI live.

`advisor_deepseek.go` triển khai interface provider nội bộ, endpoint HTTPS cố định `https://api.deepseek.com/chat/completions`, model `deepseek-v4-flash`, prompt version `deepseek-human-review/v1`. Không đọc khóa từ file/repo, chỉ `DEEPSEEK_API_KEY`. Không proxy, redirect hoặc retry tự động; timeout 30 giây, context JSON tối đa 8192 byte, output tối đa 1024 token, body phản hồi tối đa 64 KiB. Thinking tắt, JSON mode bật, không gửi tools.

Chỉ nhận một completion của đúng model, role assistant, finish_reason stop, không tool calls và có usage hợp lệ. Nội dung vẫn qua boundary chung: schema, references, freshness, HUMAN_REVIEW/ABSTAIN; không execution. Lỗi không trả raw body/transport/secret. Kiểm tra secret echo cả trên content đã decode. Usage được giữ nội bộ kể cả khi nội dung bị từ chối; chưa là ledger chi phí.

Regression server loopback kiểm tra cấu hình request, phản hồi hợp lệ/rỗng/cắt cụt, sai model, tools, thiếu usage, auth, không retry 503, sai schema, secret echo và giới hạn dữ liệu. Không gọi API thật trong tests hay CI.

## Còn thiếu trước khi chạy live

### Runner fixture BR-10 offline — IN_REVIEW

Từ learner: `go run ./cmd/bot advisor fixture-run EXISTING_OUTPUT_PARENT`. Thư mục cha phải có sẵn; mỗi lần tạo một thư mục con mới private `br11-offline-*`, không ghi đè bundle cũ. Đường dẫn trả trong `bundle_path`. Đây là xuất bundle offline, không phải campaign trả phí; không nhận provider/key hoặc history tùy ý.

Runner tạo observation synthetic giá đỡ laptop, ghi history qua store, ghi human action giả lập và PENDING outcome qua chính lệnh store, rồi đọc context bằng buildAdvisorContext. Mock chạy qua boundary shared, chỉ khi SUPPORTED/ABSTAIN mới lưu `advisor.json` chứa context/payload, output và provider/model/prompt version. History/actions/outcomes cùng input fixture nằm trong bundle để đọc lại và replay. Không sửa dữ liệu học viên hoặc PROGRESS. Partial failure giữ bundle để chẩn đoán, status lỗi không chứng minh output đã lưu hoàn chỉnh.

Expected IDs: br11-decision, br11-observation, br11-action, br11-outcome. Pending không phải zero/doanh thu thật. Đọc reason/unknowns để thấy giới hạn nguồn; exact IDs không chứng minh mọi phát biểu đúng. Bundle mock không là live proof, không dùng ledger trả phí hoặc thay context digest campaign v1. Nối DeepSeek/campaign vào chain này vẫn phải được review riêng trước live.

### Campaign init — đã review/merge, chỉ nghiệm thu offline

PR #64 merged `0f5745d`, head `3a96b11`: review không có lỗi chặn trong scope trusted local directory; CI 4/4, tests/race/vet learner và 8 validators PASS. Không init campaign người dùng hoặc gọi API khi kiểm thử. Các giới hạn TOCTOU, OS support và live proof vẫn giữ nguyên.

`bot advisor campaign-init` tạo một campaign trống tại cùng đường dẫn cố định của report. Không nhận path, không đọc API key hoặc gọi mạng. User config root phải có sẵn, không symlink, đúng owner và không group/other-writable. Thư mục ứng dụng phải private/đúng owner; chỉ tạo bằng 0700 nếu chưa có. Manifest fsync trước thành công. Parent và campaign được kiểm tra, không thay quyền dữ liệu có sẵn.

Chạy lại luôn INIT_ERROR nếu campaign tồn tại (kể cả partial init), không reset reservation; giữ dữ liệu để kiểm tra thủ công. Nếu lỗi sau tạo directory/manifest thì không tự xóa. Thông báo INITIALIZED không cấp execution permission hoặc chứng minh đã gọi provider. Chỉ thử init trong temp directory ở tests, chưa khởi tạo campaign người dùng. Hostile path replacement đồng thời vẫn ngoài scope. OS ngoài Linux/macOS fail closed qua ownership guard.

### Campaign report — IN_REVIEW

Cập nhật sau #62 merge `9645e6d`: path guard đang review. CLI kiểm mọi ancestor hiện hữu bằng Lstat, chặn symlink/non-directory, yêu cầu thư mục ứng dụng và campaign không cho group/other truy cập, owner là effective UID trên Linux/macOS. OS khác fail closed. Không tự chmod/chown/tạo đường dẫn. Missing path nay trả PATH_ERROR trước khi đọc report. Không bảo vệ khỏi hostile concurrent replacement hoặc thay OS-user/config root; đây chưa phải capability filesystem chống TOCTOU. Lệnh init/run live vẫn chưa bật.

Lệnh `bot advisor campaign-report` không nhận path và không đọc API key. Nó dùng thư mục cấu hình của tài khoản máy (`os.UserConfigDir`) cộng `affiliate-expert-learning-roadmap-v2/deepseek-br11-v1`, không phụ thuộc repo/cwd. Trên macOS thông thường là `~/Library/Application Support/affiliate-expert-learning-roadmap-v2/deepseek-br11-v1`. Không tự tạo thư mục, không có init/reset hoặc run live; chưa có campaign thì trả REPORT_ERROR, không artifact.

Report giữ lock ngắn trong lúc đọc, kiểm manifest/results/reservation canonical và sequence, không thay bytes ledger hay refund. Có attempts, reserved_microusd, remaining_reservation_microusd, estimated_known_microusd, missing_result_attempts, unknown_usage_attempts. Tổng estimate chỉ cộng usage đã có, không đại diện tổng hóa đơn khi còn unknown/missing. `invoice_reconciled=false`, `execution_permitted=false` luôn giữ nguyên. Số còn lại là ngân sách reservation, không phải số dư tài khoản provider.

Đây mới là vị trí cố định của lệnh report. Trusted local user config root vẫn là giả định; thay OS profile/config environment, sửa/xóa/rollback dữ liệu hoặc symlink ancestor chưa được chống. Runner live phải kiểm ownership/path và dùng cùng campaign trước khi được bật. Không hướng dẫn người học tạo ledger bằng tay hoặc khởi tạo campaign mới để vượt cap.

Tests kiểm report sau reservation chưa có result và mock result không có usage, ledger bytes không đổi, lock được nhả, file bất thường bị reject, missing không auto-init và CLI từ chối path argument. Tests/vet/race learner được chạy offline; chưa có live proof.

### Result ledger — đã review/merge trong phạm vi metadata offline

[#61](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/61) merged `75f8f7b`, head sửa `92eb078`. Review không còn lỗi chặn trong phạm vi nội bộ sau hai bản sửa dưới đây; tests/race/vet, smoke BR-11a, 8 validators và CI 4/4 PASS. Không nghiệm thu live, hóa đơn hoặc nội dung khuyến nghị. Không phát sinh API request trong quá trình review.

Review #61 bổ sung hai chốt: context digest phải khớp fixture campaign hiện hành, không chỉ là hex 64 ký tự; reader/result writer phải xác minh manifest và thư mục campaign. Regression thay digest bằng hash khác vẫn đúng định dạng hoặc thay manifest, kiểm read/reserve/persist đều từ chối. Khi thay fixture phải version campaign rõ, không dùng fixture mới để diễn giải ledger cũ.

Sau #60 merged `94c32c8` (chain conformance offline), `advisor_results.go` bổ sung result metadata immutable trong cùng thư mục campaign. Result liên kết attempt đã reserve, lưu provider/model/prompt version, SHA256 context, status, usage nullable, snapshot giá và estimate microUSD. Không lưu raw provider body, reasoning, secret hay output bị reject. Đây chưa là artifact nội dung phục vụ review chất lượng model.

Writer dùng lock cùng reservation, O_EXCL và fsync file/directory. Duplicate/overwrite/orphan bị từ chối; reader strict/canonical/bounded kiểm chi phí tính lại. Reservation kế tiếp kiểm mọi result đã có; result hỏng hoặc orphan chặn request. Nếu crash trước khi có result, reservation vẫn giữ nguyên và lần sau không tái dùng attempt đó. Partial result sau write failure chặn lượt mới, không tự xóa hoặc refund. RESULT_ERROR không có nghĩa request chưa bị tính phí.

Estimate dùng snapshot [giá chính thức](https://api-docs.deepseek.com/quick_start/pricing/) kiểm tra ngày 2026-09-07: Flash peak cache-miss input $0.44/MTok, output $1.32/MTok, tính nguyên microUSD và làm tròn lên. Bỏ qua giảm giá cache/off-peak để ước tính bảo thủ; không phải invoice, không là giá đảm bảo tương lai. Usage không có là null, không phải zero. Không giải phóng reservation theo estimate. Mock không được có usage tính phí. Runner ghi metadata hiện chỉ dùng fixture ledger; chưa nối thành lệnh live nhận artifact BR-10.

Tests: restart đọc lại results, immutable/orphan, kết quả thiếu vẫn giữ reservation, cap không refund, negative/inconsistent usage, integer rounding và unknown != zero. Fixed campaign identity/path, báo cáo tổng hợp/đối soát nhà cung cấp và lưu artifact nội dung có kiểm chứng còn mở trước live. Không dùng ledger hoặc result này để tự đánh DONE BR-11.

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
