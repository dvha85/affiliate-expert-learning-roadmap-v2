# BR-13a — Normalizer M06 dùng chung

Bàn giao qua [PR #74](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/74), implementation `bb88f6b`; trạng thái merge/CI xem tại PR. Codex thực hiện/review theo quyền chủ repo; core/harness/learner tests/race/vet và 8 validators PASS. Không có lỗi chặn trong phạm vi di chuyển code, các giới hạn bên dưới vẫn giữ nguyên.

Baseline `829a0cd`, sau BR-12 lab #73. Di chuyển WatchRequest, CanonicalObservation, EvaluateWatchRequest, NormalizeWatchObservation, ContentHash, M06FileInput và raw decode/profile validation vào `core/m06`. Demo giữ wrappers/aliases; M07 vẫn ở harness. Không thay schema hoặc semantics, không đọc mạng/file/clock trong core.

Core tests độc lập kiểm hash SHA256 literal, NEW/UNCHANGED/CHANGED, identity với thời gian tương đương, HEAD thiếu nội dung, source/method và raw malformed/null/duplicate fields. Regression M06/M07/M11 trong harness giữ nguyên, chạy qua implementation mới. Validator repo đọc core/m06; kiểm field observation_id chịu được khoảng trắng gofmt, không bỏ điều kiện.

## Giới hạn giữ nguyên, không coi là quyền fetch

- Input là response do caller cung cấp, evidence_kind=synthetic, use_context=test. AllowHosts cũng do caller khai báo, không là policy do ứng dụng sở hữu.
- EvaluateWatchRequest so method/host và hash; không kiểm HTTP status. DecodeM06Input kiểm cấu trúc, không bảo đảm response 2xx. CLI demo vẫn kiểm 2xx trước normalize. Không dùng riêng decoder làm boundary fetch đầy đủ.
- URL validation hiện chưa là SSRF defense: không chốt credentials/ports/DNS/private-address/redirect/proxy. Không mở network từ những helper này. Status NEW/UNCHANGED/CHANGED chỉ so body hash, không chứng minh history đã persist hoặc retry thành công.
- Identity gồm subject/URL/time/method/correlation/body hash; thời gian tương đương cùng ID nhưng serialized observed_at giữ cách viết đầu vào. Handoff sau này phải chốt canonical retry/record identity, không mặc định DeepEqual là retry an toàn.
- Normalizer chưa parse giá/commission vào domain input; claim_kind=fact chỉ mô tả fixture response, không chứng minh seller claim đúng. Parser phải giữ nguồn, limitation và missing fields; không tự gán dữ liệu thật.

## Thứ tự tiếp theo

BR-13b: parser fixture cụ thể và handoff CLI vào history BR-09, status chỉ persist sau ghi thành công; test malformed/missing/forbidden method/source, sink failure, idempotency và restart/replay. Profile fixture được kiểm soát bởi ứng dụng, không nhận allowlist tùy ý như quyền truy cập nguồn thật.

BR-13c: nguồn public/được phép cụ thể, cấu hình và proof fetch thực sau khi chốt quyền nguồn, giới hạn HTTP và parser; không suy từ nguồn fixture thành live proof. Chưa có network source được chọn trong BR-13a. BR-13 cha IN_PROGRESS, checklist chính vẫn mở.
