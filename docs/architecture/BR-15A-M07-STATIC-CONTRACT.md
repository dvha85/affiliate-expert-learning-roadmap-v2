# BR-15a — hợp đồng tĩnh M07 evidence-agent

Blueprint M07 gọi adapter loopback cố định; registry cũng là policy JSON cố
định trong blueprint, không nhận host/method từ event. Adapter phải ACK trace
tool đã đăng ký, context canonical, output grounding và proposal đã lưu theo
đúng thứ tự. Agent chỉ nhận context/evidence sau hai ACK đầu; proposal chỉ đi
tiếp khi output là `HUMAN_REVIEW`, có `proposed_action`,
`execution_permitted=false` và adapter trả ACK persistence. Validator tĩnh
kiểm tra các binding này, registry GET/read-only, cảnh báo tool/webpage là dữ
liệu không tin cậy và trần `write_permission=false`.

Đây là drift/contract evidence, không phải bằng chứng model hoặc n8n engine đã
chạy. Prompt injection, ID bịa, host/redirect lạ và write request có regression
qua learner adapter; vẫn cần chạy workflow thật theo BR-15.

```text
python scripts/validate_n8n_m07.py
```
