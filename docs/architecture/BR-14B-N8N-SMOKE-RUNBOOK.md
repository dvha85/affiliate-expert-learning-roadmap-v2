# BR-14b — runbook smoke n8n M06

Đây là runbook để maintainer ghi integration evidence trên **một n8n engine cụ
thể**. Repo hiện chưa có instance được cấp quyền, nên `tested_n8n_version` vẫn
`UNVERIFIED`; không dùng JSON parse hoặc contract test offline để thay thế.

## Bản ghi bắt buộc

```text
tested_n8n_version: <ví dụ 1.x.y>
tested_at_utc: <RFC3339>
blueprint_sha256: <sha256 của lab/n8n/M06-readonly-watcher.blueprint.json>
import_execution_id: <id hoặc UNAVAILABLE>
smoke_execution_ids: [<id>, ...]
history_record_refs: [<record-id>, ...]
credential_scope: read-only hoặc mock
result: PASS | FAIL | UNVERIFIED
```

## Trình tự

1. Dùng instance sạch, import blueprint và ghi lại engine/node versions cùng
   import result; không activate schedule mặc định.
2. Thay placeholder bằng fixture/public source được phép, giới hạn GET và
   credential read-only. Chạy hai lần cùng nội dung (`NEW`, rồi `UNCHANGED`),
   đổi nội dung (`CHANGED`), rồi đổi thứ tự key JSON (vẫn `UNCHANGED`).
3. Gửi output qua adapter BR-13; kiểm tra `observation_id`, correlation,
   `canonical_history_handoff=ACK`, `canonical_history_persisted=true` và record ID trong canonical history.
4. Chạy response quá 200000 ký tự và sink failure. Kỳ vọng fail closed, không
   báo `persisted=true`; retry sau đó không tạo record mơ hồ.
5. Restart workflow/mất watcher cache: lần sau có thể `NEW`, nhưng history
   canonical phải còn nguyên và replay được.

Kết quả chỉ được ghi PASS khi có execution IDs và history refs. Nếu chưa có
instance hoặc nguồn được phép, ghi `UNVERIFIED` và giữ workflow inactive.
