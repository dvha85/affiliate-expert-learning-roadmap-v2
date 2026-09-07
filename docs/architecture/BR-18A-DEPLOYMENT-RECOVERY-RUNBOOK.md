# BR-18a — runbook deployment/recovery 24/7

Runbook này chuẩn bị cho runtime thật nhưng không tự chọn VPS/nhà cung cấp.
Maintainer phải ghi nơi chạy, giới hạn CPU/RAM/disk, secret source và người có
quyền STOP trước activation.

## Gates bắt buộc

- health check fail hoặc lease hết hạn → không activate;
- STOP durable được ghi trước, đọc lại sau restart và chặn mọi attempt;
- backup append-only history + ledger, kiểm tra checksum và restore vào thư mục
  cô lập;
- restore xong phải replay khớp, không reset reservation/STOP;
- recovery cần human review, không tự mở lại executor.

## Evidence record

```text
runtime: <host/service/version>
tested_at_utc: <RFC3339>
resource_limits: <cpu/ram/disk/timeouts>
backup_ref: <immutable ref>
restore_ref: <run/id>
stop_drill_ref: <run/id>
replay_result: MATCH | FAIL
reviewer: <human>
result: PASS | FAIL | UNVERIFIED
```

Chưa có runtime/backup target thật thì ghi `UNVERIFIED`; tài liệu này không phải
production evidence.
