# BR-14a — hợp đồng tĩnh cho workflow n8n M06

BR-14a có một cổng kiểm tra tĩnh cho `lab/n8n/M06-readonly-watcher.blueprint.json`.
Validator kiểm tra workflow chỉ đưa fixture synthetic vào shared adapter
`/v1/m06/fixture-import`, không còn node local để GET, parse, hash hay tự dựng
`HistoryRecord`. Adapter phải append, resolve/replay record rồi mới trả ACK.
N8n không có canonical history hoặc watcher cache trong profile này.

Đây chỉ là contract/drift evidence. Nó không chứng minh blueprint import hoặc
execution thành công trên một engine n8n cụ thể. `lab/n8n/COMPATIBILITY.md` tiếp
tục giữ `tested_n8n_version: UNVERIFIED` cho tới khi có instance, ngày thử,
import result, execution IDs và history refs.

Chạy kiểm tra:

```text
python scripts/validate_n8n_m06.py
```
