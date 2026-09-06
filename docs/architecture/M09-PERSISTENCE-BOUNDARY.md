# BR-03d.1 — Bảo vệ persistence M09

IN_REVIEW: [PR #39](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/39), commit triển khai `b04db90`. [CI theo head PR](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/39/checks). Chưa merge.

Evidence local: tests/vet ba module, 8 Python validators, 10 Python regressions và diff check PASS. CI cần xác nhận theo head PR riêng.

#38 đã review/merge `0744f6a`, tests/vet mission-runtime và CI 4/4 PASS trên head `74bec6e`. Phần này sửa trực tiếp LoadM09State/PersistM09State, không chỉ thêm CLI audit.

## Hợp đồng đọc và ghi

DecodeM09State kiểm envelope object/exact key/duplicate/trailing, bắt buộc intent/policy. Mỗi artifact hiện diện phải qua canonical raw schema trước typed decode. Approval/authorization/execution được bỏ hẳn nếu chưa có, không ghi null. Chỉ nhận profile M09, không nhận field quyền legacy hoặc artifact M10/M11.

Sau schema, dùng decoder M08/M09 để giữ json.Number trong parameters, tính lại intent hash và kiểm liên kết ID/hash/policy/correlation/idempotency/executor. Compatibility aliases khôi phục từ canonical mode/authority sau validation, không lấy từ field ngoài schema. Lỗi trả zero state + error, không expose state một phần.

PersistM09State serialize rồi dùng cùng validator trước mọi I/O. State sai không tạo thư mục, file tạm hoặc ghi đè snapshot cũ. Snapshot mới luôn xuất hai marker map dạng object, kể cả rỗng. Giữ temp + rename hiện hữu; chưa đổi locking/fsync hoặc transaction.

## Marker và tương thích

- Marker là set: key không chỉ khoảng trắng, value phải true. Null/array/non-bool/false bị chặn. Executor hiện hữu không tạo false entries; file cũ có dạng này cần review/migrate riêng, không tự loại bỏ.
- Map rỗng bị writer cũ bỏ do omitempty được đọc như empty set; loader không rewrite file. Không tự tạo approval/authorization khi vắng mặt.
- Execution PERFORMED/UNKNOWN cần consumed marker, SUCCEEDED cần success marker. Execution cần authorization, authorization cần approval. Marker lịch sử khác key hiện tại vẫn được giữ nguyên.
- Không phát hiện được việc xóa đồng thời execution và marker, hoặc thay toàn snapshot bằng bản cũ nhất quán. Chưa có authenticated append-only ledger/version/CAS: đây không phải chống rollback toàn lịch sử.

State pending không approval vẫn hợp lệ, AuthorizeM09 trả WAIT_APPROVAL. Schema/link-valid không chứng minh approval thật; provenance, policy hiện hành, expiry, kill switch, executor và idempotency vẫn cần kiểm khi authorize/execute. State expired hoặc approval REJECT không bị sửa để trở thành executable. Typed caller không đi qua Load vẫn cần audit riêng.

## Kiểm thử

Từ lab/mission-runtime:

```text
go test ./cmd/demo -run TestM09Persistence -count=1 -v
go test ./...
go vet ./...
```

Tests thiếu execution_authorized, null/duplicate/unknown/stage/hash/link, zero state on error và file không đổi; số 9007199254740993, decimal/exponent lồng nhau giữ hash qua round-trip; pending/resume; execution local + restart giữ marker cũ/chặn replay; stripped marker bị chặn; writer fail-before-I/O. Regression M09 trước đó giữ nguyên.

M10/M11 persistence, duration overflow, I/O quotas, crash durability/locking, trusted provenance và live evidence còn mở. Không chạy affiliate thật, không đổi PROGRESS học viên, chưa đánh dấu BR-03 hoàn tất.
