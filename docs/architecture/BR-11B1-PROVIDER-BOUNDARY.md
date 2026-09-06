# BR-11b.1 — Chuẩn bị ranh giới provider, kiểm thử offline

Trạng thái: DONE trong phạm vi chuẩn bị provider và HTTP fixture offline. [PR #57](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/57) đã review/merge tại `b465fa7`. BR-11 vẫn IN_PROGRESS; chưa có live provider hay bằng chứng vận hành thật.

## Phạm vi

Learner dùng `advisorProvider` để tách việc sinh phản hồi khỏi kiểm chứng. Lệnh `bot advisor mock` không đổi; artifact bổ sung `provenance` gồm provider `mock/v1`, model `deterministic`, prompt version `human-review/v1`. Context giữ version `br11a-context/v1`.

`fixtureHTTPProvider` là giao thức JSON phục vụ kiểm thử, không phải adapter tương thích một hãng API. Không có lệnh CLI bật HTTP, không chọn model trả phí, không gọi dịch vụ bên ngoài. Endpoint chỉ chấp nhận HTTP tới địa chỉ IP loopback literal; không dùng proxy, không theo redirect. Đây không phải cấu hình dùng production.

Request fixture gồm identity, instruction và context; response là AdvisorOutput JSON trực tiếp. Instruction yêu cầu HUMAN_REVIEW hoặc ABSTAIN, coi evidence là dữ liệu không đáng tin và cấm yêu cầu ghi/thực thi. Prompt không là chốt bảo mật: kiểm chứng sau phản hồi mới quyết định chấp nhận. Context stale/future/invalid bị chặn trước khi gọi provider. ADVISE bị từ chối trong phạm vi này; ABSTAIN có dẫn chứng vẫn phải resolve đúng ID.

## Cấu hình và giới hạn

- Model phải có tên; khóa chỉ đọc từ tên biến môi trường khi tạo request, không nằm trong provenance/context.
- Timeout tổng lớn hơn 0, tối đa 30 giây, bao gồm retry và đọc body; hủy từ caller được truyền xuống request.
- Retry 0..2 (tối đa 3 lần gửi), chờ 10/20 ms giữa lần thử trong fixture. Chỉ retry lỗi transport và HTTP 429/502/503/504. Không retry schema, write request, auth, redirect hay body quá lớn. Đây là lịch thử nghiệm, không là chính sách live rate limit; Retry-After/billing/idempotency cần thiết kế theo provider được chọn.
- Request tối đa 1 MiB; response cấu hình 1 byte..1 MiB, đọc thêm tối đa 1 byte để phát hiện vượt ngưỡng.
- Không in URL, khóa, raw lỗi transport hoặc body lỗi của server. Phản hồi chứa nguyên khóa bị từ chối. Bộ lọc này không bảo đảm nhận diện khóa bị biến đổi/mã hóa; live adapter cần review riêng về dữ liệu nhạy cảm trước khi gửi context.
- Chỉ output đạt schema/reference/freshness mới đi qua; lỗi provider trả PROVIDER_ERROR, không trả artifact bị từ chối. Lệnh mock luôn giữ `execution_permitted=false`, không ghi store.

## Bằng chứng kiểm thử

`advisor_provider_test.go` dùng server loopback thật với khóa giả qua `t.Setenv`: request/provenance, thành công, schema sai, unknown ID (cả ABSTAIN), write request, ADVISE, body quá lớn, secret echo, 401 không retry, 429/503 hết lượt, retry rồi phục hồi, redirect không theo, timeout/cancel, lỗi mạng, secret thiếu và giới hạn config. Kiểm tra context stale/future không làm tăng số request.

Chạy trong `lab/affiliate-bot`: `go test ./...` và `go vet ./...`. Chạy `python3 scripts/smoke_br11a.py` từ gốc repo để kiểm tra CLI mock và store không thay đổi. CI hiện có đã chạy toàn bộ test learner nên bao gồm test mới.

Kiểm tra local ngày 2026-09-07: toàn bộ test/vet learner PASS, regression HTTP với race detector PASS, smoke BR-11a PASS, 8 validators và 10 Python regressions PASS. Có regression ngưỡng response bằng đúng giới hạn và request vượt giới hạn bị chặn trước khi gửi. Chưa coi đây là review độc lập hoặc kết quả CI của PR.

Review head `be6fc3f`: không phát hiện lỗi chặn trong phạm vi offline; đã rà soát loopback/proxy/redirect, vòng retry/deadline/body limit, secret/error output, context preflight và schema/reference/state boundary. Chạy lại toàn bộ learner với race detector và vet PASS; smoke BR-11a PASS; CI 4/4 PASS trước merge. Giới hạn live bên dưới vẫn là điều kiện nghiệm thu riêng, không được suy ra từ fixture.

## Chưa nghiệm thu

BR-11b.2 cần chọn provider/model và được phép dùng chi phí trước khi tích hợp giao thức thật, TLS/endpoint allowlist, secret management, chính sách retry/rate limit/cost, lọc dữ liệu gửi đi và bằng chứng live. Chưa có quyền gọi paid API. Mock/fixture không chứng minh chất lượng khuyến nghị hoặc kết quả affiliate; BR-11 chưa DONE.
