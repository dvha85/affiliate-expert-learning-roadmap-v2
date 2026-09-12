# BR-14a — hợp đồng tĩnh cho workflow n8n M06

BR-14a có một cổng kiểm tra tĩnh cho `lab/n8n/M06-readonly-watcher.blueprint.json`.
Validator kiểm tra workflow chỉ đưa fixture synthetic vào shared adapter
`/v1/m06/fixture-import`, không còn node local để GET, parse, hash hay tự dựng
`HistoryRecord`. Adapter phải append, resolve/replay record rồi mới trả ACK.
N8n không có canonical history hoặc watcher cache trong profile này.

Đây chỉ là contract/drift evidence. Nó không tự chứng minh blueprint import hoặc
execution thành công trên một engine n8n cụ thể. Disposable regression hiện chạy
M06 trên n8n `2.38.1`, gồm Schedule Trigger, bằng fixture synthetic/read-only;
release admission trong `lab/n8n/COMPATIBILITY.md` vẫn `UNVERIFIED` vì chưa có
ma trận smoke đầy đủ trên topology/deployment được chọn.

Chạy kiểm tra:

```text
python scripts/validate_n8n_m06.py
```
