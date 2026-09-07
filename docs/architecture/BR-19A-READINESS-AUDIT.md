# BR-19a — audit readiness tổng thể

`scripts/audit_readiness.py` kiểm tra kế hoạch luôn nêu rõ các khoảng trống bằng
chứng cứ trước khi công bố readiness: affiliate source được phép, n8n/model
execution, pilot người mới, live adapter và runtime/backup 24/7. Audit chỉ báo
cáo; không tự chuyển TODO/IN_PROGRESS thành DONE.

Kết quả hiện tại phải là `NOT_READY_FOR_PRODUCTION`. Fixture/schema/unit test,
CI xanh hoặc một lần gọi model không đủ thay thế operated/live evidence.

```text
python scripts/audit_readiness.py
```
