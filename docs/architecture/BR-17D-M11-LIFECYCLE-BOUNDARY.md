# BR-17d — M11 lifecycle/STOP/reconciliation boundary

M11 chỉ cho phép activation trong lease còn hạn, health gate đạt và ledger đã
reconcile. STOP là trạng thái durable, ưu tiên hơn recovery/promotion và phải
được đọc lại sau restart; mọi attempt sau STOP đều bị chặn. Recovery phải có
review, không tự động mở lại quyền.

Closed-cycle chỉ hoàn tất khi action, outcome, usage/cost và reconciliation cùng
liên kết đúng ID; thiếu result hoặc unknown usage giữ cycle mở. Promotion và
self-improvement bị giới hạn bởi policy/lease, không biến test pass thành live
proof.

Các invariant này đã được kiểm tra bởi validator M11 và các test persistence/
time-boundary hiện hành. Đây là lab/schema evidence; chưa có 24/7 deployment,
live executor hay closed cycles bên ngoài.

## Reviewed recovery sang runtime mới

`m11-recovery-export` vẫn chỉ xuất proof đọc từ runtime cũ sau khi UNKNOWN đã
được human reconciliation và ledger ở `RECOVERY_REVIEW_REQUIRED`. Lệnh mới
`m11-recovery-admit NEW_STATE_DIR OLD_STATE_DIR HANDOFF_INPUT ADMISSION_INPUT`
ghi `PRODUCTION_RECOVERY_ADMISSION` **chỉ trong runtime mới**. Admission bắt
buộc runtime mới có mission state riêng, lease + approval khác identity/hash,
activation và NORMAL ledger riêng. Approval mới phải được review sau resolution
cũ, admission review sau approval mới; path runtime cũ/mới (kể cả parent/child)
không được trùng.

Admission luôn có `execution_permitted=false`; nó không xóa STOP cũ, không
copy ledger/budget/approval cũ và không tạo gate hay authorization. Runtime mới
vẫn phải đi đầy đủ health/cost/gate/authorization trước fixture operation. Khi
runtime cũ đã STOP, registry chỉ nhận reconciliation artifact để hoàn tất
handoff; không được append lease/approval mới vào đó.

Smoke BR-16a tạo admission, thử reuse lease và ghi lease vào runtime STOP (đều
reject), rồi backup/restore runtime mới và resolve lại admission. Đây chỉ là
evidence offline với fixture; chưa chứng minh recovery hay executor thật.

UNKNOWN→STOP dùng journal bounded được ghi trước execution/stopped ledger. Khi
restart, writer replay exact predecessor → execution UNKNOWN → ledger STOPPED
→ durable STOP; status và mutation fail closed khi journal còn lại. Đây chỉ che
transition đó, không phải transaction đa-file hay bằng chứng power-loss.

```text
python scripts/validate_m11.py
```
