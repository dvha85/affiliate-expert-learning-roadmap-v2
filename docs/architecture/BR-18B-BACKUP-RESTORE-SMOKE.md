# BR-18b — backup/restore smoke offline

Smoke tạo history/ledger append-only và STOP marker trong thư mục tạm, lập
SHA-256 manifest, mô phỏng tamper rồi từ chối checksum sai, khôi phục byte
giống hệt và xác nhận STOP vẫn active sau “restart”.

Đây chỉ là operational lab evidence; không chứng minh storage durability,
backup retention hoặc deployment 24/7 của một host cụ thể.

```text
python scripts/smoke_br18b_backup_restore.py
```
