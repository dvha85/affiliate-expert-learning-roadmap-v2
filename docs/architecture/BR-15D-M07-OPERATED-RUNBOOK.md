# BR-15d — M07 operated execution runbook

Mục tiêu là kiểm engine n8n thật bằng fixture synthetic, không dùng campaign, URL
affiliate, dữ liệu khách hàng hay quyền thực thi. Cockpit/model credential chỉ tồn tại
trong n8n local; không đưa vào repository, log chia sẻ hoặc execution capture.

## Điều kiện trước khi chạy

- N8n local chỉ bind `127.0.0.1` và workflow M07 đã import từ blueprint hiện hành.
- Learner Bot watcher adapter đang chạy ở loopback với canonical history synthetic.
- Chat model đã được thử thành công trong Cockpit; model phải hỗ trợ JSON output của
  n8n Agent. Model/account không khả dụng phải được coi là provider failure, không đổi
  validator để bỏ qua.
- Input chỉ allowlist một URL fixture công khai đã review, method `GET`, không redirect.

## Chạy và xác minh

1. Chạy workflow M07 với manual trigger và lưu raw output của `n8n execute --rawOutput`
   ở ngoài repo. Không ghi API key vào file output.
2. Xác nhận chain thành công: registered tool ACK, canonical context VALID, Agent,
   grounding VALID, proposal ACK và report cuối ACK. Output luôn phải có
   `execution_permitted=false`.
3. Kiểm bằng parser repository, thay các path bằng runtime local của bạn:

```bash
python3 scripts/validate_n8n_m07_operated_execution.py \
  /path/to/m07-execution.json \
  --proposal-store /path/to/history.jsonl.m07/proposals
```

4. Restart n8n, không thay history/proposal store, rồi chạy lại parser. Proposal ID và
   record ID phải vẫn resolve được từ proposal store.
5. Chạy riêng một negative case qua workflow/adapter (forged evidence ID hoặc write
   request). Case phải dừng trước persistence và không tạo proposal file mới.

`PASS` chỉ chứng minh một M07 synthetic, read-only path đã chạy trong engine. Nó không
chứng minh business outcome, business truth, approval hoặc quyền MACHINE_EXECUTION.
