# BR-15c — contract parse output Agent M07

Contract offline mô phỏng boundary sau Agent: chỉ nhận JSON có
`state=HUMAN_REVIEW`, `authority=A2-RO`, `write_permission=false` và mọi
`evidence_ids` đều resolve trong context đầu vào. Output có ID bịa, context rỗng,
write request hoặc tự nâng quyền đều chuyển thành `ABSTAIN`.

Đây là parser/drift evidence, không phải bằng chứng model đã grounded hoặc n8n
đã execute. Khi tích hợp thật, adapter phải resolve IDs từ canonical store trước
để chuyển output tiếp; không được tự gắn ID từ context vào câu trả lời của model.

```text
python scripts/validate_n8n_m07_output_cases.py
```
