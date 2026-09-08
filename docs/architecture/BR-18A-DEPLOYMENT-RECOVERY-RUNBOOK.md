# BR-18a — runbook deployment/recovery 24/7

Runbook này chuẩn bị cho runtime thật nhưng không tự chọn VPS/nhà cung cấp.
Maintainer phải ghi nơi chạy, giới hạn CPU/RAM/disk, secret source và người có
quyền STOP trước activation.

## Cấu hình tham chiếu local

Profile này chỉ chứng minh đường chạy offline của learner Bot. Dùng Go 1.23+
và một thư mục dữ liệu riêng, không dùng `lab/affiliate-bot/data` làm store
runtime.

```bash
cd lab/affiliate-bot
go build -o /tmp/affiliate-bot ./cmd/bot
mkdir -p /tmp/affiliate-runtime

# khởi tạo state, sau đó chạy một lần capture vào store append-only
/tmp/affiliate-bot mission init /tmp/affiliate-runtime
/tmp/affiliate-bot history capture /tmp/affiliate-runtime/history.jsonl \
  data/m02-sample-observations.json demo-1 \
  2026-09-03T00:00:00Z 2026-09-03T00:00:00Z

# start / status / logs / stop cho canonical adapter local
nohup /tmp/affiliate-bot watcher serve /tmp/affiliate-runtime/history.jsonl \
  127.0.0.1:8787 >/tmp/affiliate-runtime/watcher.log 2>&1 &
echo $! >/tmp/affiliate-runtime/watcher.pid
curl --fail http://127.0.0.1:8787/healthz
tail -n 50 /tmp/affiliate-runtime/watcher.log
kill "$(cat /tmp/affiliate-runtime/watcher.pid)"

# status/logs/stop tương đương cho profile local
/tmp/affiliate-bot mission status /tmp/affiliate-runtime
/tmp/affiliate-bot history list /tmp/affiliate-runtime/history.jsonl
/tmp/affiliate-bot mission stop /tmp/affiliate-runtime manual-drill
```

`status` phải trả `stop: true` sau restart; mọi canary mới phải bị từ chối
cho tới khi có state/lease mới do người review tạo. Với n8n, import blueprint
ở `lab/n8n/`, đặt canonical adapter ở `canonical_store_url`, đặt timeout và
redirect policy theo blueprint, rồi lưu execution ID cùng ACK của adapter.

## Backup, restore và kiểm tra sau restart

```bash
/tmp/affiliate-bot backup create /tmp/affiliate-runtime /tmp/affiliate-backup
/tmp/affiliate-bot backup restore /tmp/affiliate-backup /tmp/affiliate-restored
/tmp/affiliate-bot history replay /tmp/affiliate-restored/history.jsonl
/tmp/affiliate-bot mission status /tmp/affiliate-restored
```

Expected: `BACKED_UP`, `RESTORED`, `replay=MATCH`, rồi `stop: true`. Manifest
SHA-256 phải được kiểm trước khi chép file; target restore phải trống. Chạy
`python3 scripts/smoke_br18b_backup_restore.py` để kiểm cả tamper, process mới,
budget/canary link và STOP durable trong môi trường tạm.

Target của `backup create` cũng phải trống và không được là runtime hoặc thư
mục con của runtime; tạo backup mới vào một thư mục khác thay vì ghi đè snapshot
cũ.

## Gates bắt buộc

- health check fail hoặc lease hết hạn → không activate;
- STOP durable được ghi trước, đọc lại sau restart và chặn mọi attempt;
- backup history + ledger và, nếu M07 adapter đã persist, toàn bộ sidecar
  `history.jsonl.m07/`; manifest chỉ nhận relative path chuẩn hóa, checksum và
  restore vào thư mục cô lập;
- restore xong phải replay khớp, không reset reservation/STOP;
- recovery cần human review, không tự mở lại executor.
- restore không được tự tạo lại ledger, reservation, approval hoặc lease đã
  hết hạn; các artifact phải replay/resolve được từ store.

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
