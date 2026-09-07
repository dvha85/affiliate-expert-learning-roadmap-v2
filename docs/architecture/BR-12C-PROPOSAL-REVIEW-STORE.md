# BR-12c — Proposal và human review trong bot

Baseline `784fdd5`; tiếp nối evaluation store #70/#71. Phạm vi lab, một writer trong workspace tin cậy; không API hoặc execution. Các lệnh dưới chạy tại `lab/affiliate-bot` với các file history/action/outcome/evaluation thực sự đã tạo qua BR-10/BR-12b.

## Lệnh và input

```sh
go run ./cmd/bot proposal import history.jsonl actions.jsonl outcomes.jsonl evaluations.jsonl proposals.jsonl proposal.json
go run ./cmd/bot proposal list history.jsonl actions.jsonl outcomes.jsonl evaluations.jsonl proposals.jsonl
go run ./cmd/bot review import history.jsonl actions.jsonl outcomes.jsonl evaluations.jsonl proposals.jsonl reviews.jsonl review.json
go run ./cmd/bot review list history.jsonl actions.jsonl outcomes.jsonl evaluations.jsonl proposals.jsonl reviews.jsonl
```

Proposal ví dụ: thay evaluation_id bằng ID đã lưu; version là nhãn thay đổi dự kiến, không được xác minh tự động với Git.

```json
{"proposal_id":"proposal-1","evaluation_ids":["br12-evaluation-1"],"current_version":"v1","proposed_version":"v2","change_summary":"Làm rõ thông báo pending","expected_benefit":"Tránh kết luận sớm","risks":["Chưa kiểm chứng lợi ích với người mới"],"rollback":"Khôi phục thông báo v1 sau review thủ công","auto_apply":false}
```

Review ví dụ, chỉ nhập sau khi người review thật sự xem đề xuất. Thay timestamp bằng thời điểm review thực, không dùng timestamp mẫu để giả lập bằng chứng người dùng:

```json
{"review_id":"review-1","proposal_id":"proposal-1","reviewed_by":"human","reviewed_at":"2026-09-07T00:00:00Z","decision":"REQUEST_CHANGES","reason":"Cần regression và cách rollback cụ thể trước khi sửa"}
```

APPENDED là đã ghi record, không phải duyệt đề xuất. EXACT_DUPLICATE không ghi lại; cùng ID khác nội dung trả CONFLICT. List trả VALID và mảng artifact; mọi envelope có auto_apply=false, execution_permitted=false. APPROVE_FOR_MANUAL_CHANGE vẫn không sửa file/code/policy, không kích hoạt executor và không cấp quyền đăng nội dung.

## Boundary và persistence

- Mỗi lệnh load lại toàn bộ history → action → outcome → evaluation bằng BR-12b, sau đó proposal/review; upstream hỏng hoặc mất tham chiếu sẽ chặn cả list lẫn import, không trả artifact lỗi.
- Proposal dùng raw schema và semantic validator M05. Evaluation IDs phải unique và resolve chính xác. Versions khác nhau, không whitespace ở đầu/cuối; change/benefit/rollback không trắng. Importer yêu cầu ít nhất một risk không trắng (chặt hơn schema dùng chung vốn cho risks optional); không thay schema hoặc harness.
- Review dùng raw schema: reviewed_by=human, quyết định chỉ APPROVE_FOR_MANUAL_CHANGE/REJECT/REQUEST_CHANGES, reason không trắng. Proposal phải tồn tại, reviewed_at không trước mọi evaluation được proposal tham chiếu. Proposal không có created_at nên không kiểm được review sau thời điểm tạo proposal; chỉ kiểm mốc evaluation, không tự thêm timestamp.
- `human` là khai báo local, không xác thực danh tính hay chữ ký. Không tự tạo review thay người. Các review khác ID cho cùng proposal được giữ như lịch sử; chưa có luật chọn review có hiệu lực, không tự lấy bản mới nhất để thực thi.
- Store JSONL riêng, record tối đa 1 MiB; chặn duplicate ID, JSON/schema hỏng và thiếu newline cuối trước append. Final store/input phải regular file, không symlink; chặn alias với tất cả input/upstream. Không reset/repair tự động.
- Kế thừa JSONL adapter một writer: không transaction đa file, không khóa liên tiến trình, không cam kết fsync chịu mất điện; local rollback/hostile directory swap ngoài scope. ID linkage không content-pin upstream hợp lệ bị thay nội dung cùng ID. Không dùng trong môi trường writer đồng thời.
- Risks/benefit/rollback/version chỉ là nội dung người nhập, không chứng minh tính đúng/khả thi. Evaluation INCONCLUSIVE không bị đổi thành kết luận có hiệu quả bởi một proposal hay approval.

Lỗi USAGE_ERROR trả exit 2. Các lỗi PATH_ERROR, HISTORY_ERROR, ACTION_STORE_ERROR, OUTCOME_STORE_ERROR, EVALUATION_STORE_ERROR, PROPOSAL_STORE_ERROR, REVIEW_STORE_ERROR, INPUT_ERROR, INVALID_RECORD, CONFLICT, STORE_ERROR trả exit 1 và không artifact. Giữ file lỗi để kiểm tra, không xóa store để né conflict. Proposal/review list trên store chưa tạo trả lỗi; không tự khởi tạo khi chỉ đọc.

## Kiểm thử và phần còn lại

Regression kiểm import/list/reload, exact duplicate/conflict, upstream không đổi, approval không execution, orphan IDs, evaluation IDs lặp, auto_apply, risk rỗng, version/rollback, machine reviewer, thời gian sớm, unknown/duplicate fields, store duplicate/partial/orphan và path alias. Tests dùng fixture/key-free; không chạy review thật thay người.

BR-12d vẫn còn walkthrough M00→M05, diff/review cho thay đổi nhỏ, regression FAIL/PASS và rollback; kết quả test phải tách kết quả đo. BR-12 cha chưa DONE.
