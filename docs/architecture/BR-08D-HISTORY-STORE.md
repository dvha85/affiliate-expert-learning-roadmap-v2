# BR-08d — Store seam M02 và quyền sở hữu dữ liệu

IN_REVIEW: [PR #50](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/50), commit triển khai `79556ca`. Tests/vet bốn module GOWORK=off, 8 validators và 10 Python regressions local PASS. Chưa merge.

BR-08c đã review/merge #49 `b8fa83b`, CI 4/4 PASS. BR-08d chỉ tách filesystem I/O, không thay database, hash, schema, command hoặc dữ liệu đã lưu.

## Quyền sở hữu

Learner application sở hữu canonical history và đường dẫn tường minh do người dùng truyền. Trong lát cắt này, composition root và logic M02 hiện hữu vẫn ở cmd/bot/history.go: LoadHistory/AppendHistory dựng adapter `store.JSONL{}` rồi gọi loadHistoryWith/appendHistoryWith. Không có global store mutable. `internal/store.History` chỉ có Open và AppendLine, nhận path và bytes; store không biết DecisionPacket, grant hoặc executor.

Validation/schema, record/observation ID conflict, duplicate, sort và replay vẫn ở ứng dụng M02. Core/m03 không import store. Việc chuyển tiếp toàn bộ logic M02 khỏi package main không thuộc PR này; không gọi seam I/O nhỏ là đã refactor toàn application layer. Adapter khác chỉ được nối sau khi có compatibility tests và quyết định storage riêng.

## Tương thích và lỗi

- Giữ JSONL hiện hữu: json.Marshal(record) + một newline, append-only; không truncate, rewrite, migrate hoặc reseal. ID/hash/formula version, legacy ranked:null, sorting và replay giữ nguyên.
- Load đọc qua adapter rồi dùng Scanner/validation cũ; lỗi đọc/corruption trả lỗi và không append. Chỉ os.ErrNotExist cho phép append khởi tạo như trước; không coi mọi I/O error là file rỗng.
- Append validate và kiểm conflicts trước serialize/I/O. EXACT_DUPLICATE không ghi thêm. Adapter chặn frame rỗng hoặc có CR/LF trước OpenFile, không sửa buffer caller, không tự tạo thư mục cha.
- File mode mới 0600, O_CREATE|O_APPEND|O_WRONLY; giữ bytes đã có. Lỗi Write, short write hoặc Close không trả APPENDED. Đây là cải thiện báo lỗi, không transaction mới.
- **Không bảo đảm append atomic/crash-safe**: lỗi sau partial write có thể để lại dòng cuối chưa hoàn chỉnh. Không truncate để “phục hồi”; giữ prefix, báo lỗi, dừng và đối soát/restore từ backup có review. Test injected failure trước write không chứng minh mọi lỗi disk sau partial write. H-02/H-03 vẫn mở, không có locking/CAS đa writer.

## Cách mở rộng cùng workspace

1. Giữ file history đang dùng và backup riêng trước đổi phiên bản; không chạy capture vào fixture repo.
2. Dùng lại `bot history list PATH` và `bot history replay PATH` để kiểm history cũ. Lệnh/đối số không đổi, không có command migrate.
3. Khi thêm capability, gọi ứng dụng validate trước và dùng seam I/O; không cho core, n8n cache hoặc Agent trực tiếp ghi canonical history.
4. Action validate của BR-08c vẫn read-only. Action/outcome persistence chỉ thiết kế ở BR-10; không thêm dòng loại mới vào history M02.
5. Khi read/append lỗi, không xóa file hoặc reset state để retry. Kiểm đường dẫn/quyền/file hỏng; lỗi partial-write cần recovery có review. Revert code không thay hoặc xóa dữ liệu người học.

## Bằng chứng

history_store_test.go inject read/corrupt/append failure; no write trên read/corrupt, status không APPENDED khi write lỗi, prefix không đổi trong fake, duplicate/conflict không ghi, replay MATCH. internal/store/history_test.go kiểm byte framing, caller buffer không đổi, append giữ prefix, path lỗi không tạo parent/truncate blocker. Toàn bộ history_test/decision_packet_test và binary CLI tests hiện hữu tiếp tục chạy qua default adapter, gồm legacy và raw schema regression.

Nghiệm thu BR-08d là seam/ownership/compatibility trong lab, không chứng minh durability production. BR-08e review tích hợp vẫn còn mở.
