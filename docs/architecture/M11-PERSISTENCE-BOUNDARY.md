# BR-03d.3 — Bảo vệ persistence M11

Triển khai trên nhánh `codex/br-03d-m11-persistence`, nối tiếp PR #40 (M10). Chưa merge; BR-03 tổng thể chưa hoàn thành.

## Boundary đọc/ghi

Reader ledger dùng DecodeM11Artifact: canonical raw schema, strict decode và ledger self-consistency trước khi trả state. Lease ID/version/hash phải khớp expected lease; lỗi trả zero ledger. Executor, gate, outcome và reconciliation cùng dùng reader này. Activation reader cũng kiểm raw schema, gồm activated_at bắt buộc/đúng định dạng, rồi đối chiếu lease ID/version/hash.

Writer ledger và activation serialize và kiểm cùng boundary trước I/O. InitializeProductionLedger kiểm cả hai artifact trước khi tạo directory. Writer sai không ghi đè ledger hợp lệ. Collection nil được serialize thành array rỗng qua MarshalJSON hiện hữu; raw null/missing không được tự sửa. Số int64 lớn hơn 2^53 được giữ chính xác.

## Restart và giới hạn

Giữ cơ chế activation hiện hữu: mất ledger không tự tái tạo khi activation còn. File hỏng, sai lease hoặc activation thiếu timestamp làm runtime dừng trước sandbox effect. STOPPED, budget, pending và outcome links được giữ qua restart; ghi outcome để đối soát không tự đưa ledger về ACTIVE.

Không tự migrate ledger legacy chứa null arrays hoặc activation thiếu field. Cần phục hồi/migration có review, không xóa activation để vượt guard. Không chứng minh provenance của expected lease do caller cung cấp; không chống xóa cả ledger/activation hoặc rollback cả snapshot hợp lệ. Temp+rename và locking hiện hữu giữ nguyên; chưa bổ sung transaction nhiều file hoặc chứng minh crash consistency/fsync directory. activated_at hợp schema không đồng nghĩa bằng chứng thời gian đáng tin cậy. Duration arithmetic vẫn ngoài phạm vi.

## Kiểm thử

Từ lab/mission-runtime:

```text
go test ./cmd/demo -run TestM11Persistence -count=1 -v
go test ./...
go vet ./...
```

Regression kiểm required/null/type/duplicate/unknown/trailing; exact lease binding; lỗi trả zero state và file không đổi; writer fail-before-I/O; exact int64; ledger mất/hỏng/sai lease và activation thiếu timestamp không gây effect; restart giữ STOPPED và cho phép ghi nhận outcome mà không resume. Chỉ dùng fixture và sandbox local, không phải bằng chứng affiliate live hoặc Operated.
