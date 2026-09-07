# BR-15a — hợp đồng tĩnh M07 evidence-agent

Blueprint M07 không còn gán cứng placeholder `e1`. `evidence_ids` phải đi vào từ
input context; khi không có context, agent không có bằng chứng để viện dẫn và
phải giữ trần `HUMAN_REVIEW`. Validator tĩnh kiểm tra binding này cùng registry
read-only, HTTP GET-only, cảnh báo tool/webpage là dữ liệu không tin cậy và
`write_permission=false`.

Đây là drift/contract evidence, không phải bằng chứng model hoặc n8n engine đã
chạy. Prompt injection, ID bịa, host/redirect lạ và write request vẫn cần smoke
trên integration path thật theo BR-15.

```text
python scripts/validate_n8n_m07.py
```
