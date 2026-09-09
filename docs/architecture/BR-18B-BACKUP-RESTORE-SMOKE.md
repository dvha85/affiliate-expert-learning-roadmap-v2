# BR-18b — backup/restore smoke offline

Smoke khởi động learner Bot, tạo `history.jsonl` bằng lệnh capture, tạo
`mission-state.json` và STOP bằng runtime, lập SHA-256 manifest, mô phỏng
tamper rồi từ chối checksum sai, khôi phục vào đường dẫn đích chưa tồn tại và dùng process
mới để replay history. Cuối cùng nó thử tạo canary sau restore để chứng minh
STOP còn chặn operation và state budget/approval không bị bootstrap lại.

Đây chỉ là operational lab evidence; không chứng minh storage durability,
backup retention hoặc deployment 24/7 của một host cụ thể.

```text
python scripts/smoke_br18b_backup_restore.py
```
