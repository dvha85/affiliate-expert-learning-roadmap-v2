# BR-03d.2 — Bảo vệ persistence M10

#39 đã review/merge `d009d1f`, CI 4/4 PASS trên head `ca708a5`; tests/vet mission-runtime PASS. Phần này nối validation vào reader/writer ledger M10, không chỉ CLI audit.

## Đọc và ghi

loadExistingCanaryLedger nhận expected grant, kiểm canonical raw schema + strict typed decode và ledger self-consistency qua DecodeM10Artifact trước khi trả state. Grant ID/version/hash phải khớp exact. Lỗi trả zero ledger + error. Executor và RecordCanaryOutcome cùng dùng reader này; file hỏng không được coi như file chưa tồn tại.

persistCanaryLedger và ensureInitialCanaryLedger serialize rồi kiểm cùng boundary trước mọi I/O. Collection nil được xuất array rỗng theo MarshalJSON hiện hữu; raw null/missing vẫn bị chặn. Số cost int64 giữ chính xác. Initial ledger phải rỗng, không reconciliation hoặc last_execution_at. Writer sai không ghi đè file hợp lệ hoặc để lại temp/initialization marker.

## Dấu khởi tạo và mất file

Khởi tạo mới tạo file `LEDGER_PATH.initialized` bằng O_EXCL và sync trước khi tạo ledger. Đây là dấu tồn tại local, không chứa quyền hoặc provenance. Nếu ledger mất nhưng dấu còn, executor trả WAIT_LEDGER_MISSING, kể cả caller dùng state rỗng cũ. Không xóa dấu khi bước tạo ledger sau đó lỗi: ưu tiên chặn và yêu cầu đối soát thay vì retry có thể reset ngân sách. Không tự xóa dấu để phục hồi.

Ledger cũ hợp lệ nhưng chưa có dấu vẫn được đọc để tương thích; reader không mutate file hoặc tạo dấu ngầm. Vì vậy không claim chống reset sau mất file đối với ledger legacy không có dấu. Cần quy trình migration có review riêng. Ledger legacy có array null bị chặn, không tự normalize/rewrite dữ liệu lưu.

Không chống được xóa cả ledger+dấu, thay cả snapshot bằng bản cũ, đổi directory hoặc sửa file từ bên ngoài. Chưa có authenticated append-only history/CAS; không xác thực grant từ expected object do caller cung cấp. Temp+rename/locking hiện hữu giữ nguyên, chưa chứng minh crash consistency hoặc fsync directory. Duration arithmetic và persistence M11 vẫn còn mở.

## Kiểm thử

Từ lab/mission-runtime:

```text
go test ./cmd/demo -run TestM10Persistence -count=1 -v
go test ./...
go vet ./...
```

Tests required/null/type/duplicate/unknown/trailing, exact grant binding, zero-state errors và file không đổi; writer fail-before-I/O; int64 lớn hơn 2^53; restart giữ budget/pending/idempotency và outcome links; ledger hỏng/sai grant/mất file không được sửa hoặc tái tạo; stale empty state với dấu còn không phát sinh effect mới. Chỉ dùng thư mục sandbox tạm, không affiliate live. BR-03 tổng thể chưa hoàn thành.
