# BR-15b — contract adversarial M07

Bộ kiểm tra offline cố định các nhóm request nguy hiểm: prompt injection, ID
bịa, thiếu context, host/redirect ngoài allowlist và write request. Nó kiểm tra
instruction coi tool/webpage text là dữ liệu không tin cậy, registry chỉ
read-only `GET` trên host được phép, và grounding boundary luôn trả
`HUMAN_REVIEW` với `write_permission=false`.

Đây không phải model evaluation hay n8n execution evidence. Khi có instance,
maintainer phải chạy cùng các case qua integration path, lưu execution IDs và
history refs theo runbook BR-14b; model output không được tự tạo evidence ID
hoặc nâng quyền.

```text
python scripts/validate_n8n_m07_adversarial.py
```
