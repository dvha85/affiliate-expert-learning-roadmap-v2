# BR-14a — hợp đồng tĩnh cho workflow n8n M06

BR-14a bổ sung một cổng kiểm tra tĩnh cho `lab/n8n/M06-readonly-watcher.blueprint.json`.
Validator kiểm tra đủ node và các bất biến an toàn: HTTP chỉ `GET`, URL mẫu là
HTTPS, canonical JSON theo thứ tự key, ba trạng thái `NEW/UNCHANGED/CHANGED`, giới
hạn kích thước cache, và handoff rõ ràng sang Deterministic Core. Static data của
n8n được đánh dấu là watcher cache, không phải canonical history.

Đây chỉ là contract/drift evidence. Nó không chứng minh blueprint import hoặc
execution thành công trên một engine n8n cụ thể. `lab/n8n/COMPATIBILITY.md` tiếp
tục giữ `tested_n8n_version: UNVERIFIED` cho tới khi có instance, ngày thử,
import result, execution IDs và history refs.

Chạy kiểm tra:

```text
python scripts/validate_n8n_m06.py
```
