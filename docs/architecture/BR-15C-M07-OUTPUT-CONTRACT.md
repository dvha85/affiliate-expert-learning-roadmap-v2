# BR-15c — contract parse output Agent M07

Contract offline chạy boundary learner sau Agent: chỉ nhận JSON có
`state=HUMAN_REVIEW`, `authority=A2-RO`, `write_permission=false`; mỗi claim còn
phải có `field_or_claim` và `value` khớp exact với một field trong evidence đã
resolve. Output có ID bịa, giá trị bịa, context rỗng, write request hoặc tự nâng
quyền đều chuyển thành `ABSTAIN`.

Đây không phải bằng chứng model đã grounded hoặc n8n đã execute. Khi tích hợp
thật, adapter phải resolve IDs từ canonical store trước để chuyển output tiếp;
không được tự gắn ID từ context vào câu trả lời của model. Validator không coi
câu chữ tự do là bằng chứng; model phải trả cấu trúc claim để adapter kiểm tra
binding dữ liệu.

```text
python scripts/validate_n8n_m07_output_cases.py
```
