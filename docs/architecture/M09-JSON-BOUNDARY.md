# BR-03c.6 — M09: kiểm artifact, không cấp quyền

Trạng thái: IN_REVIEW tại [PR #33](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/33). Commit triển khai `8dcd376`; chưa merge. Xem [CI theo head PR](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/33/checks).

Kiểm local: tests/vet ba module PASS, 8 Python validators và 10 Python regressions PASS, diff check PASS. CLI fixture xuất CONSISTENT_UNVERIFIED/CANCELLED/NOT_PERFORMED với hai flag quyền false. CI phải kiểm riêng theo head PR.

Phạm vi triển khai: raw ApprovalRecord, ExecutionAuthorization và ExecutionRecord của profile M09 APPROVED_LIVE; nối với intent/policy M08 để audit file lịch sử. Không gọi AuthorizeM09 hoặc executor từ CLI, không persist, không consume approval, không tạo trusted approval.

## Chạy bài offline

Từ repo root:

```text
cd lab/mission-runtime
go run ./cmd/demo m09-check testdata/m08-intent.json testdata/m09-policy.json testdata/m09-approval.json testdata/m09-authorization.json testdata/m09-execution.json
go test ./cmd/demo -run TestM09 -count=1 -v
```

Cả năm fixture đều synthetic. Dù schema authorization yêu cầu `execution_mode:APPROVED_LIVE` và `execution_authorized:true`, các giá trị trong fixture **không phải quyền do hệ thống tin cậy cấp**. Không đưa fixture vào executor. Record mẫu là CANCELLED/NOT_PERFORMED; không có side effect được thực hiện.

Output chỉ là summary `result:CONSISTENT_UNVERIFIED`, các ID, trạng thái execution đã khai báo, `approval_authenticated:false`, `execution_permitted:false`. Không echo hoặc phát hành artifact authorization mới. Exit 0 là audit hoàn tất, không phải execution approval hoặc bằng chứng action đã xảy ra. Error input/schema/link/time trả exit 1, không có summary; writer error không bị bỏ qua.

## Thứ tự kiểm

1. JSON gốc của cả năm file → canonical schema → exact typed decode. Required/null/type/duplicate/unknown/enum/const/hash shape bị chặn trước khi thông tin mất vào zero values. Schema conditional của execution authorization/record phân biệt M09/M10/M11. Typed M09 không hỗ trợ các field canary/production, không bỏ ngầm field khác stage.
2. Tính lại hash intent bằng hàm M08, không seal lại hoặc tự sửa quyền/hash.
3. Intent ID/hash, policy version, approval ID, authorization ID, executor ID, correlation và idempotency key phải liên kết đúng. Approval REJECT không thể đi với authorization chain được nhận. Policy phải ALLOW hoặc HUMAN_REVIEW, vẫn NON_AUTHORIZING/false.
4. Kiểm timeline: created ≤ policy checked ≤ approved ≤ authorized ≤ attempted; expiry intent/approval phải chứa authorization window. Authorization có window dương. Effect khai báo PERFORMED phải có attempted_at trước expires_at (đúng bằng end bị từ chối). FAILED/CANCELLED/UNKNOWN có thể ghi nhận attempt sau expiry để audit; không suy nó hợp lệ để execute. So sánh instant, không so chuỗi timezone.

Schema cho `one_time:true` chỉ là khai báo, không xác minh đã consume. Checker không nạp ledger, không nhìn clock hiện tại và không cấp quyền dùng authorization lịch sử. Việc chạy lại cùng file cho cùng summary là audit replay, không phải chứng minh one-time execution. Durable replay/kill-switch/policy-revalidation vẫn được test bởi bộ M09 runtime cũ, không được thay bằng CLI này.

## Bài tập và giới hạn

Copy năm file vào thư mục bài tập; sửa approval hash về hash khác đúng định dạng → BROKEN_LINK. Đổi approved_by thành agent hoặc one_time=false → INVALID_SCHEMA. Đổi approval decision=REJECT → REJECTED_APPROVAL. Đổi execution.authorization_id → BROKEN_LINK. Với record PERFORMED, thử attempted_at bằng expiry → EXPIRED_AUTHORIZATION. Khôi phục từng thay đổi để audit lại; không seal hoặc sửa expected để giấu lỗi.

Một file giả ghi approved_by=human và mọi ID khớp vẫn có thể qua consistency: output luôn **UNVERIFIED**, không xác thực approver_id. Muốn execute phải dùng trusted approval provenance, policy hiện hành, executor profile, kill switch và ledger qua đường vận hành riêng. Checker cũng không resolve decision/evidence payload hoặc re-evaluate policy từ trusted context. Không lấy claim PERFORMED trong file làm live proof.

Không thay LoadM09State/PersistM09State, AuthorizeM09, ExecuteLocalSandbox hay compatibility aliases; các boundary runtime persistence đó vẫn cần tích hợp/audit tiếp. Không xử lý M10/M11 trong profile này. Không có size quota cho public endpoint; chỉ file người vận hành chọn. BR-03 tổng thể và BR-17 còn mở, không tự cập nhật PROGRESS.

Tests mới kiểm required/null/unknown/duplicate, stage/authority, mismatch hash/ID/correlation/key, timeline/expiry/offset, human giả vẫn unverified, đọc lại cùng summary không cấp quyền, file không đổi và writer errors. Test schema còn kiểm artifact serialize từ AuthorizeM09 và executionFailureRecord hiện hữu (không gọi executor trong test mới); regression M09 cũ vẫn giữ đầy đủ.
