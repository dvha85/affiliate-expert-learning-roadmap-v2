# BR-10a — Ghi ActionRecord thủ công gắn với decision

Trạng thái: triển khai chờ review, BR-10 chưa DONE. Baseline BR-09 lab đã merge #52. Không có chương trình/kênh thật; mọi ví dụ dưới đây là synthetic, không là proof đã đăng bài.

PR: [#53](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/53) đã review/merge `765ac70`. Review scope lab không finding chặn; tests/vet learner và smoke BR-10a chạy lại PASS, CI 4/4 PASS. Các ghi chú chờ review ở đầu là lịch sử, BR-10 tổng thể vẫn mở.

## Quyết định store và authority

Giữ history M02 nguyên format/hash/formula. File `actions.jsonl` riêng chỉ chứa canonical ActionRecord, mỗi dòng một record; không nhét action vào history hoặc dùng state harness M09–M11. Learner application trong cmd/bot chịu trách nhiệm resolve/replay/duplicate/conflict; `internal/store.JSONL` chỉ I/O, dùng giới hạn chung 1 MiB/record. M03 schema và semantics dùng shared core, không copy validator.

`decision_id` phải khớp đúng một recorded decision trong history được chỉ định và replay MATCH. Đây không phải resolve một DecisionPacket do caller tự khai báo, không chứng minh human context, nguồn đáng tin hay approval. GET_MORE_DATA/HUMAN_REVIEW vẫn có thể được tham chiếu khi ghi nhận điều con người đã làm; record không cấp phép làm điều đó. `performed_by=human`, compliance_reviewed=true, window không trước performed_at là điều kiện M03 hiện hữu, không phải máy xác thực compliance.

CLI không gọi executor, HTTP, affiliate API hoặc đăng bài. Không suy diễn rằng thời điểm được khai báo là trusted clock hoặc action thực sự đã xảy ra. Không tự tạo outcome (BR-10b), không migration/reset/reseal file cũ.

## Lệnh và output

```text
bot action record HISTORY.jsonl ACTIONS.jsonl ACTION.json
bot action list HISTORY.jsonl ACTIONS.jsonl
```

Record kiểm cả history và toàn action store trước append. Thiếu action store được tạo khi record hợp lệ; list trên file chưa tồn tại trả STORE_ERROR. Store hỏng/duplicate action_id/orphan decision trong store bị từ chối, không tự bỏ qua hoặc sửa. Các path phải khác nhau, gồm alias symlink/hardlink của file tồn tại. Không redirect stdout vào bất kỳ input/store nào.

Stdout là JSON envelope command/status/execution_permitted=false. APPENDED hoặc EXACT_DUPLICATE có artifact là action; list VALID có artifact là array. Duplicate là cùng toàn bộ typed fields (timestamp spelling khác được xem là nội dung khác); cùng action_id khác nội dung trả CONFLICT. Lỗi không artifact, stderr có diagnostic; exit 0 thành công, 1 lỗi, 2 usage. `action validate ACTION OUTCOME` cũ không đổi.

## Bài thực hành synthetic trong workspace riêng

Từ root repo, Go theo go.mod và Python 3:

```sh
python3 scripts/smoke_br10a.py
```

Script build bot trong thư mục tạm, capture `br10a-decision` từ fixture M02, ghi action dưới đây, gọi lại để kiểm EXACT_DUPLICATE, khởi động process mới list, đổi decision thành ID không tồn tại để kiểm DECISION_ERROR, và kiểm history replay MATCH. Tất cả file được tạo trong workspace tạm riêng, không sửa evidence cá nhân.

```json
{"action_id":"br10a-action","decision_id":"br10a-decision","action_type":"synthetic_manual_post","target":"fixture:br10a-no-publication","performed_by":"human","performed_at":"2026-09-04T00:00:00Z","measurement_window_end":"2026-09-05T00:00:00Z","compliance_reviewed":true}
```

Output smoke:

```text
BR-10a PASS: synthetic record/duplicate/restart/list/orphan; no execution
```

Khi thực hành thủ công, copy JSON vào ACTION.json của workspace lab riêng, dùng decision_id tồn tại từ `history list`, chạy record rồi list với cùng hai path. Không sửa ID trong history để làm test pass. Sửa target nhưng giữ action_id để quan sát CONFLICT; muốn ghi hành động mới cần action_id mới. Giữ lại history vì list phải resolve lại decision; di chuyển cả hai file cùng nhau. Không dùng fixture để cập nhật E1/PROGRESS.

## Regression và giới hạn nghiệm thu

Tests: APPENDED/EXACT_DUPLICATE, read lại, conflict, orphan không tạo store, machine bị schema reject, compliance false, window ngược, path alias, store hỏng không overwrite, usage. Smoke chạy từng lệnh trong process riêng, kiểm input/history bytes không đổi và action list exact artifact. CI chạy smoke cùng tests/vet module learner.

Một writer theo quy trình lab; chưa có locking/transaction/CAS, crash recovery hoặc bảo đảm rollback nếu write/close lỗi. Đây là I/O seam đã có, không tuyên bố fail-atomic trước lỗi thiết bị. Hash M02 không là chữ ký; sửa toàn bộ history có thể thay ý nghĩa decision cùng ID. Store action chưa bind immutable decision digest, chưa archive context/pilot thật. Không chạy nhiều writer; không coi store là audit log chống giả mạo. BR-10b và nghiệm thu BR-10 tổng thể còn mở.
