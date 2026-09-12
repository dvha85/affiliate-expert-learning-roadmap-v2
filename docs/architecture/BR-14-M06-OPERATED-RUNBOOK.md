# BR-14 — M06 operated execution runbook

Mục tiêu là xác nhận engine n8n thật đi qua shared M06 adapter, persist canonical
history rồi mới báo ACK. Chỉ dùng fixture synthetic; không dùng campaign, URL
affiliate, dữ liệu khách hàng hay source seller.

## Điều kiện trước khi chạy

- n8n local chỉ bind `127.0.0.1`; workflow M06 được import từ blueprint hiện hành.
- Kiểm graph có đúng đường `Schedule Trigger -> M06 Adapter Input -> Build and
  Append Canonical M06 Adapter -> Require Canonical Store ACK -> Report Canonical
  M06 Result`.
- Adapter learner Bot đang bind loopback với history rỗng hoặc history fixture đã
  replay được. Không để adapter bind địa chỉ ngoài loopback.
- Input giữ `br13-offer-fixture/v1`, method `GET` và source fixture allowlisted.

## Chạy và xác minh

1. Với một bản workflow temporary dùng Execute Workflow Trigger, chạy n8n CLI và
   lưu `--rawOutput` ở ngoài repo. Workflow triển khai thật có thể vẫn dùng Schedule
   Trigger; không sửa trigger production chỉ để lấy capture.
2. Xác nhận report cuối có `result=APPENDED` hoặc `EXACT_DUPLICATE`, cùng record ID
   với adapter ACK, `canonical_history_handoff=ACK`,
   `canonical_history_persisted=true` và `execution_permitted=false`.
3. Dùng verifier repository:

```bash
python3 scripts/validate_n8n_m06_operated_execution.py \
  /path/to/m06-execution.json \
  --history /path/to/history.jsonl \
  --expected-result APPENDED
```

4. Restart n8n và adapter, chạy `bot history replay HISTORY`, rồi chạy lại verifier
   trên capture success. Record ID phải còn resolve được.
5. Chạy một fixture source không allowlisted. Capture phải kết thúc error ở `Build
   and Append Canonical M06 Adapter`, không chạy ACK/report và không làm tăng số
   record history:

```bash
python3 scripts/validate_n8n_m06_operated_execution.py \
  /path/to/m06-rejected-execution.json \
  --history /path/to/history.jsonl \
  --expect-reject --expected-record-count 1
```

`PASS` chỉ chứng minh một M06 synthetic/read-only path trong n8n engine. Nó không
chứng minh nguồn được chọn, seller/business truth, business outcome hay quyền thực
thi.
