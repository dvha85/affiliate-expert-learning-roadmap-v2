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

```text
python scripts/validate_m11.py
```
