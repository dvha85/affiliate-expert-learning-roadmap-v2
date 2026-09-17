# BR-19a — audit readiness tổng thể

`scripts/audit_readiness.py` kiểm tra kế hoạch luôn nêu rõ các khoảng trống bằng
chứng cứ trước khi công bố readiness: affiliate source được phép, n8n/model
execution, pilot người mới, live adapter và runtime/backup 24/7. Audit chỉ báo
cáo; không tự chuyển TODO/IN_PROGRESS thành DONE.

Kết quả hiện tại phải là `NOT_READY_FOR_PRODUCTION`. Fixture/schema/unit test,
CI xanh hoặc một lần gọi model không đủ thay thế operated/live evidence.

Snapshot ngày 17/09/2026 trên `main` `8b440114` vẫn trả
`NOT_READY_FOR_PRODUCTION` và resolve 79 scoped claims. Regression sau PR #403
được ghi tại
[EVIDENCE-LOCAL-REGRESSION-POST-403-20260917.md](EVIDENCE-LOCAL-REGRESSION-POST-403-20260917.md);
phạm vi vẫn synthetic/read-only/loopback. M07 engine proof là 12 PR checks trên
head `7747d661`, không phải provider, live-executor, pilot hay deployment proof.

```text
python scripts/audit_readiness.py
```
