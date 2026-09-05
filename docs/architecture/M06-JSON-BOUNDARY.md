# BR-03c.3 — Observation của M06

Trạng thái IN_REVIEW: [PR #30](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/30), code/tests `f54ea51`; Codex thực hiện, chủ repo chưa review. Chưa merge.

Phạm vi: normalizer của conformance harness và CLI **offline** `m06-check`. Không fetch, persist, watcher scheduler hoặc n8n. BR-13/14 vẫn mở. Observation schema canonical không đổi; `additionalProperties:true` cho phép `content_hash`. Profile typed của normalizer chỉ hỗ trợ các field đang xuất, từ chối extension chưa hỗ trợ thay vì bỏ dữ liệu ngầm.

## Chạy từ repo root

```text
cd lab/mission-runtime
go run ./cmd/demo m06-check testdata/m06-response.json
go test ./cmd/demo -run TestM06 -count=1 -v
```

Input là wrapper riêng của bài kiểm, không phải schema Observation: subject_id, status_code (số nguyên), request. Request bắt buộc method/url/allow_hosts/observed_at/correlation_id/body; previous_hash tùy chọn, nhưng không nhận null. Raw decode từ chối key trùng, field lạ, sai kiểu và null bắt buộc trước semantic. status_code chỉ chấp nhận 2xx; 3xx/4xx/5xx bị REJECT_RESPONSE_STATUS, không theo redirect. Method/URL/host/time/correlation vẫn qua EvaluateWatchRequest trước normalize. Allowlist này do file cung cấp, không phải trusted network policy; không có request mạng nào được gửi.

Output envelope gồm observation, result NEW/UNCHANGED/CHANGED và external_side_effects=false. Observation thực serialize phải qua canonical schema và profile (synthetic/test, content_hash SHA-256 dạng hex 64 ký tự). Schema không áp dụng cho cả envelope. Lỗi schema/request/status/output trả exit 1 và không có success envelope; lỗi đọc/ghi không bị bỏ qua.

## Sửa provenance và identity

Normalizer cũ gắn real/fact cho Body được truyền trực tiếp, dù không có bằng chứng fetch. Nay NormalizeWatchObservation xuất **synthetic + use_context=test** cho đường offline này. Claim fact chỉ là việc quan sát nội dung fixture; không khẳng định nội dung seller đúng. Limitation ghi rõ không fetch và không xác minh business truth. Test cũ mong real được cập nhật có chủ đích; không claim live proof. Adapter nguồn thật phải được thiết kế ở BR-13, không đổi nhãn fixture thành real để đạt Mission.

HEAD hoặc body rỗng/whitespace không chứa quan sát nội dung: state=missing, claim_kind=unknown. Result NEW/UNCHANGED/CHANGED chỉ mô tả hash bytes, không có nghĩa dữ liệu sản phẩm đầy đủ. Body có chuỗi lỗi với status 200 vẫn chưa được parser ngữ nghĩa phát hiện: chưa có parser nguồn/price/commission ở đây. Không tự điền giá hay commission bằng 0.

ID cũ chỉ chứa subject/correlation/body hash nên có thể trùng khi URL/thời điểm/method thay đổi. ID mới là SHA-256 của JSON tuple subject_id, URL nguyên bản, instant UTC RFC3339Nano, method trim+uppercase, correlation_id, content hash. Retry giữ các giá trị này thì cùng ID; offset tương đương cho cùng instant không đổi ID. Thay source/time/method/body tạo ID mới. previous_hash không thuộc identity. Không chuẩn hóa JSON response/key order: hash theo bytes, việc đó còn thuộc integration BR-13/14.

Định dạng ID này thay output fixture, không migrate dữ liệu hoặc rewrite history. Harness chưa sở hữu store; không dùng ID cũ/mới để suy đã persist. Nếu có artifact xuất cũ ngoài repo, giữ nguyên nguồn và version trước đó, không sửa lịch sử để khớp ID mới.

## Bài tập và giới hạn

Copy fixture sang thư mục bài tập, chạy lệnh với path copy: đổi status_code=500 → REJECT_RESPONSE_STATUS; method=POST → REJECT_WRITE_METHOD; bỏ body hoặc body=null → INVALID_SCHEMA. Body="" vẫn hợp lệ nhưng observation missing/unknown. Dùng hash của output làm previous_hash → UNCHANGED; thay body → CHANGED. Giữ bản gốc để so sánh, không cập nhật PROGRESS từ test.

Schema không xác minh nguồn thật, quyền truy cập, freshness theo clock vận hành, hash khớp response bên ngoài, hoặc nội dung có hỗ trợ quyết định không. Profile chỉ kiểm hình dạng hash; normalizer mới tính hash từ Body. URL trong fixture là dữ liệu, không được mở/fetch. Chưa hỗ trợ source_ref-only hoặc mọi domain extension của canonical Observation qua projection này.

Evidence: red/green đã tái hiện nhãn real sai và ID không phân biệt URL/time/method. Tests mới kiểm schema required/null, extension schema cho phép nhưng projection không hỗ trợ, hash/profile, ID/offset/retry, missing/HEAD, lỗi input/status/method, file không đổi và writer failure. M06 eval cũ giữ expected states; thêm lượt chạy qua normalizer rồi kiểm schema output. Không nới schema để lấy PASS.

Kiểm local: tests/vet ba module PASS; 8 Python validators và 10 Python regression tests PASS; git diff --check PASS. CI smoke kiểm fixture qua m06-check, không phải live fetch hoặc n8n compatibility proof.
