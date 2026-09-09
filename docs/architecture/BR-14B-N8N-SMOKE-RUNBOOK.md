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
2. Giữ fixture synthetic mặc định và không thêm credential hay URL nguồn thật.
   Chạy hai lần cùng fixture: lần đầu `APPENDED`, lần hai `EXACT_DUPLICATE`.
3. Thử fixture sai URL/status/body. Adapter phải reject và workflow không được
   báo persistence. Với hai run thành công, kiểm tra `record_id`, correlation,
   `canonical_history_handoff=ACK`, `canonical_history_persisted=true` và replay
   của record trong canonical history.
4. Dừng adapter để thử sink failure. Kỳ vọng fail closed, không báo
   `persisted=true`; sau khi adapter chạy lại, retry không tạo record mơ hồ.
5. Restart workflow: history canonical phải còn nguyên và replay được. Profile
   này không có watcher cache hoặc trạng thái `NEW/UNCHANGED/CHANGED` n8n.

Kết quả chỉ được ghi PASS khi có execution IDs và history refs. Nếu chưa có
instance hoặc nguồn được phép, ghi `UNVERIFIED` và giữ workflow inactive.
