# BR-18a — runbook deployment/recovery 24/7

Runbook này chuẩn bị cho runtime thật nhưng không tự chọn VPS/nhà cung cấp.
Maintainer phải ghi nơi chạy, giới hạn CPU/RAM/disk, secret source và người có
quyền STOP trước activation.

## Profile hiện hành: macOS native local verification (2026-09-21)

Đây là profile được chọn để kiểm chứng ngay trên macOS trong lúc chưa có
target-host/provider. Chạy trực tiếp bằng Go 1.27, Node 24 và n8n 2.38.1;
Docker/Compose không bắt buộc. Chỉ dùng fixture/synthetic và read-only, mặc
định loopback, không provider credential hoặc live executor.

### Phạm vi personal-local

Repo này được maintainer tự vận hành cho mục đích cá nhân. Vì vậy không cần
second operator, user/máy sạch hoặc independent beginner pilot để chấp nhận
bounded macOS lab. Run record `macos-br18a-20260921T155228Z` đã đủ làm
`PERSONAL_LOCAL_SCOPE_ACCEPTED` cho lifecycle local; quyết định chi tiết nằm ở
[personal-local scope evidence](EVIDENCE-PERSONAL-LOCAL-SCOPE-20260921.md).

Maintainer tự tạo run record mới:

- trước mỗi thay đổi lớn về code, schema, toolchain hoặc runtime;
- sau restore hoặc thay đổi runtime/backup/restore path; và
- mỗi tháng khi lab còn được sử dụng.

Mỗi lần kiểm phải dùng backup target mới và restore target trống khác với
runtime, replay history, kiểm tra health và restart, rồi xác nhận `mission
status` vẫn giữ `stop: true`. Khi STOP còn sticky, `m11-activate` và `m11-gate`
phải bị từ chối trước khi tạo artifact mới. Lưu exact versions, host/path,
bind addresses và kết quả vào `$LAB_ROOT/runs/<run-id>`; không ghi token vào
log, URL, fixture hay backup.

Đây là acceptance chỉ cho personal-local lab; nó không đóng target-host,
provider, public network, power-loss, distributed durability, live executor,
business outcome hoặc production readiness. Nếu sau này cần learner-operable
hoặc production claim, phải mở lại các yêu cầu review/pilot tương ứng.

Mọi lệnh chạy trong Terminal macOS. Tạo các đường dẫn tách biệt trên
filesystem macOS trước khi start; không dùng thư mục repo làm canonical store:

```bash
export LAB_ROOT="$HOME/affiliate-lab"
export LAB_RUNTIME="$LAB_ROOT/runtime"
export LAB_BACKUPS="$LAB_ROOT/backups"
export LAB_RESTORES="$LAB_ROOT/restores"
export LAB_LOG="$LAB_ROOT/var/log"
export LAB_RUN="$LAB_ROOT/var/run"
mkdir -p "$LAB_RUNTIME" "$LAB_BACKUPS" "$LAB_RESTORES" "$LAB_LOG" "$LAB_RUN"

go version       # phải là Go 1.27.x
node --version   # phải là v24.x
n8n --version    # phải là 2.38.1
```

Build/start tối thiểu (đổi `REPO_ROOT` theo clone thực tế):

```bash
export REPO_ROOT="$HOME/src/affiliate-expert-learning-roadmap-v2"
export LAB_BIN="$LAB_ROOT/bin/affiliate-bot"
mkdir -p "$(dirname "$LAB_BIN")"
cd "$REPO_ROOT/lab/affiliate-bot"
go build -o "$LAB_BIN" ./cmd/bot

"$LAB_BIN" mission init "$LAB_RUNTIME"
"$LAB_BIN" history capture "$LAB_RUNTIME/history.jsonl" \
  data/m02-sample-observations.json demo-1 \
  2026-09-03T00:00:00Z 2026-09-03T00:00:00Z

export CANONICAL_ADAPTER_TOKEN="$(python3 -c 'import secrets; print(secrets.token_urlsafe(32))')"
nohup "$LAB_BIN" watcher serve "$LAB_RUNTIME/history.jsonl" \
  127.0.0.1:8787 >"$LAB_LOG/watcher.log" 2>&1 &
echo $! >"$LAB_RUN/watcher.pid"
curl --fail http://127.0.0.1:8787/healthz
```

Đặt `N8N_USER_FOLDER="$LAB_ROOT/n8n"`, `N8N_HOST=127.0.0.1` và
`N8N_PORT=5678`, rồi dùng Node 24 trong `PATH`. Giữ log/PID ở `LAB_LOG`/
`LAB_RUN`, không ghi token vào URL, fixture, backup hoặc log. Backup/restore
phải dùng hai thư mục khác nhau:

```bash
export N8N_USER_FOLDER="$LAB_ROOT/n8n"
export N8N_HOST=127.0.0.1
export N8N_PORT=5678
export LAB_BACKUP="$LAB_BACKUPS/backup-$(date -u +%Y%m%dT%H%M%SZ)"
export LAB_RESTORE="$LAB_RESTORES/restore-$(date -u +%Y%m%dT%H%M%SZ)"
"$LAB_BIN" backup create "$LAB_RUNTIME" "$LAB_BACKUP"
"$LAB_BIN" backup restore "$LAB_BACKUP" "$LAB_RESTORE"
"$LAB_BIN" history replay "$LAB_RESTORE/history.jsonl"
"$LAB_BIN" mission status "$LAB_RESTORE"
```

Kết quả macOS chỉ được ghi là `PASS` sau khi có health, restart, replay và
durable-STOP output cùng exact version/host/path record. Khi chưa chạy thật,
evidence record phải giữ `result: UNVERIFIED`.

## Profile triển khai về sau: Windows 24/7 + Ubuntu 24.04 LTS (WSL2/Hyper-V)

Đây là phương án always-on về sau, không phải điều kiện để bắt đầu kiểm chứng
trên macOS và không phải target-host production. Dùng **WSL2 + Ubuntu 24.04
LTS** trước; **Hyper-V + Ubuntu 24.04 LTS** là fallback khi cần VM boundary.
Chạy runtime Linux trực tiếp bằng Go 1.27, Node 24 và n8n 2.38.1; Docker/Compose
không bắt buộc.

Mọi lệnh bên dưới chạy **bên trong Ubuntu**, sau khi đã cài đúng toolchain và
clone repo vào filesystem Linux. Không đặt canonical store trong `/mnt/c`.
Tạo các đường dẫn tách biệt trước khi start:

```bash
export LAB_ROOT="$HOME/affiliate-lab"
export LAB_RUNTIME="$LAB_ROOT/runtime"
export LAB_BACKUPS="$LAB_ROOT/backups"
export LAB_RESTORES="$LAB_ROOT/restores"
export LAB_LOG="$LAB_ROOT/var/log"
export LAB_RUN="$LAB_ROOT/var/run"
mkdir -p "$LAB_RUNTIME" "$LAB_BACKUPS" "$LAB_RESTORES" "$LAB_LOG" "$LAB_RUN"

go version       # phải là Go 1.27.x
node --version   # phải là v24.x
n8n --version    # phải là 2.38.1
```

Profile Windows chỉ dùng fixture/synthetic và read-only watcher. Adapter phải bind
`127.0.0.1:8787`; n8n phải bind `127.0.0.1:5678`. Nếu cần kiểm tra từ thiết
bị khác trong cùng LAN, bind private address sau khi có Windows firewall
allowlist cụ thể; không port-forward hoặc mở public ingress. Ghi lại bind
address và firewall rule trong run record.

Build/start tối thiểu (đổi `REPO_ROOT` theo clone thực tế):

```bash
export REPO_ROOT="$HOME/src/affiliate-expert-learning-roadmap-v2"
export LAB_BIN="$LAB_ROOT/bin/affiliate-bot"
mkdir -p "$(dirname "$LAB_BIN")"
cd "$REPO_ROOT/lab/affiliate-bot"
go build -o "$LAB_BIN" ./cmd/bot

"$LAB_BIN" mission init "$LAB_RUNTIME"
"$LAB_BIN" history capture "$LAB_RUNTIME/history.jsonl" \
  data/m02-sample-observations.json demo-1 \
  2026-09-03T00:00:00Z 2026-09-03T00:00:00Z

export CANONICAL_ADAPTER_TOKEN="$(python3 -c 'import secrets; print(secrets.token_urlsafe(32))')"
nohup "$LAB_BIN" watcher serve "$LAB_RUNTIME/history.jsonl" \
  127.0.0.1:8787 >"$LAB_LOG/watcher.log" 2>&1 &
echo $! >"$LAB_RUN/watcher.pid"
curl --fail http://127.0.0.1:8787/healthz
```

Đặt `N8N_USER_FOLDER="$LAB_ROOT/n8n"` để database n8n nằm trong
`$LAB_ROOT/n8n/.n8n`, `N8N_HOST=127.0.0.1`, `N8N_PORT=5678`, giữ log/PID ở
`$LAB_LOG`/`$LAB_RUN`, rồi dùng lệnh n8n local ở phần dưới với Node 24 trong
`PATH`. Không ghi token vào URL, fixture, backup hoặc log.

```bash
export N8N_USER_FOLDER="$LAB_ROOT/n8n"
export N8N_HOST=127.0.0.1
export N8N_PORT=5678
```

Đường dẫn backup/restore của profile Windows phải khác nhau và nằm trên Linux
filesystem:

```bash
export LAB_BACKUP="$LAB_BACKUPS/backup-$(date -u +%Y%m%dT%H%M%SZ)"
export LAB_RESTORE="$LAB_RESTORES/restore-$(date -u +%Y%m%dT%H%M%SZ)"
"$LAB_BIN" backup create "$LAB_RUNTIME" "$LAB_BACKUP"
"$LAB_BIN" backup restore "$LAB_BACKUP" "$LAB_RESTORE"
"$LAB_BIN" history replay "$LAB_RESTORE/history.jsonl"
"$LAB_BIN" mission status "$LAB_RESTORE"
```

Kết quả profile chỉ được ghi là `PASS` sau khi có health, restart, replay và
durable-STOP output cùng exact version/host/path record. Khi chưa chạy trên
Windows/Ubuntu thật, evidence record phải giữ `result: UNVERIFIED`; quyết định
này không đóng target-host/provider/public-network, region latency,
power-loss/atomic multi-file, distributed durability, clean-machine pilot,
live executor hoặc business outcome.

Quyết định hiện hành được ghi tại [macOS local lab decision](EVIDENCE-BR18A-MACOS-LOCAL-LAB-20260921.md); profile Windows được giữ tại
[Windows always-on lab decision](EVIDENCE-BR18A-WINDOWS-ALWAYS-ON-LAB-20260921.md).

## Cấu hình tham chiếu local

Profile này chỉ chứng minh đường chạy offline của learner Bot. Dùng Go 1.27+
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

# start / status / logs / stop cho canonical adapter local. Token chỉ nằm
# trong môi trường process và header Bearer của n8n/caller; không ghi vào URL,
# fixture hay log.
export CANONICAL_ADAPTER_TOKEN="$(python3 -c 'import secrets; print(secrets.token_urlsafe(32))')"
nohup /tmp/affiliate-bot watcher serve /tmp/affiliate-runtime/history.jsonl \
  127.0.0.1:8787 >/tmp/affiliate-runtime-watcher.log 2>&1 &
echo $! >/tmp/affiliate-runtime-watcher.pid
curl --fail http://127.0.0.1:8787/healthz
# API mutation/read example:
curl --fail -H "Authorization: Bearer $CANONICAL_ADAPTER_TOKEN" \
  -H 'Content-Type: application/json' 'http://127.0.0.1:8787/v1/history?record_id=demo-1'
tail -n 50 /tmp/affiliate-runtime-watcher.log
kill "$(cat /tmp/affiliate-runtime-watcher.pid)"

# status/logs/stop tương đương cho profile local
/tmp/affiliate-bot mission status /tmp/affiliate-runtime
/tmp/affiliate-bot history list /tmp/affiliate-runtime/history.jsonl
/tmp/affiliate-bot mission stop /tmp/affiliate-runtime manual-drill
```

`status` phải trả `stop: true` sau restart; mọi canary mới phải bị từ chối
cho tới khi có state/lease mới do người review tạo. Với n8n, import blueprint
ở `lab/n8n/`, đặt canonical adapter ở `canonical_store_url`, đặt timeout và
redirect policy theo blueprint, rồi lưu execution ID cùng ACK của adapter.

### Khởi động n8n local cùng adapter

Với cài đặt n8n dạng Node, `N8N_USER_FOLDER` là **thư mục cha** chứa `.n8n`,
không phải chính thư mục `.n8n`. n8n sẽ tự đọc database ở
`$N8N_USER_FOLDER/.n8n/database.sqlite`. Đồng thời JS Task Runner tự gọi
`node` từ `PATH`; chỉ gọi n8n bằng đường dẫn tuyệt đối nhưng bỏ Node khỏi
`PATH` sẽ làm runner lặp crash ngay sau khi server mở cổng.

Ví dụ đã kiểm chứng cho n8n `2.38.1` và Node `24` (đổi các đường dẫn theo máy):

```bash
export N8N_USER_FOLDER=/Users/you/.n8n
export PATH=/path/to/node-v24/bin:$PATH

# Kiểm tra trước khi start để không đè một instance đang chạy.
if curl --fail --max-time 2 http://127.0.0.1:5678/healthz; then
  echo "n8n is already listening on 127.0.0.1:5678" >&2
  exit 1
fi

nohup /path/to/node-v24/bin/node /path/to/n8n/bin/n8n start \
  >/tmp/affiliate-n8n.log 2>&1 &
echo $! >/tmp/affiliate-n8n.pid

curl --fail --retry 10 --retry-connrefused http://127.0.0.1:5678/healthz
tail -n 50 /tmp/affiliate-n8n.log
```

Log và PID của process phải nằm ngoài `/tmp/affiliate-runtime`: runtime là
canonical artifact store có inventory nghiêm ngặt, nên backup sẽ từ chối file
vận hành không thuộc graph như `watcher.log` hoặc `n8n.log`.

Expected log có `n8n ready`, `n8n Task Broker ready` và `Registered runner
"JS Task Runner"`. Lỗi Python internal runner chỉ liên quan Code node Python;
không phải lý do bỏ qua M06/M07 JavaScript workflow. Tuy vậy trước deployment
thật phải cấu hình external runner phù hợp nếu workflow dùng Python. Lệnh trên
chỉ là local synthetic/read-only profile, không phải cấu hình production.

## Backup, restore và kiểm tra sau restart

```bash
/tmp/affiliate-bot backup create /tmp/affiliate-runtime /tmp/affiliate-backup
/tmp/affiliate-bot backup restore /tmp/affiliate-backup /tmp/affiliate-restored
/tmp/affiliate-bot history replay /tmp/affiliate-restored/history.jsonl
/tmp/affiliate-bot mission status /tmp/affiliate-restored

# Stop phải còn chặn một lifecycle operation sau restore (exit non-zero).
if /tmp/affiliate-bot mission m11-activate /tmp/affiliate-restored fixture-lease \
  2026-09-03T00:00:00Z; then
  echo "unexpected activation after durable STOP" >&2
  exit 1
fi

# Gate cũng bị chặn trước khi resolve input hoặc ghi registry artifact.
test ! -e /tmp/affiliate-restored/m11-artifacts.jsonl
if /tmp/affiliate-bot mission m11-gate /tmp/affiliate-restored missing-lease \
  missing-health missing-cost missing-ledger 2026-09-03T00:00:00Z; then
  echo "unexpected gate after durable STOP" >&2
  exit 1
fi
test ! -e /tmp/affiliate-restored/m11-artifacts.jsonl
```

Expected: `BACKED_UP`, `RESTORED`, `replay=MATCH`, rồi `stop: true`. Manifest
SHA-256 phải được kiểm trước khi chép file; target restore phải **chưa tồn tại**
(không dùng lại cả thư mục rỗng). Restore chép vào staging cùng parent, chạy các
loader/graph gate rồi mới publish atomically. Chạy
`python3 scripts/smoke_br18b_backup_restore.py` để kiểm cả tamper, process mới,
budget/canary link và STOP durable trong môi trường tạm.

Lệnh `m11-activate` cuối phải in JSON `status: "STOPPED"` và exit non-zero.
`fixture-lease` không được resolve/activate; nó chỉ chứng minh STOP được đọc
trước mọi điều kiện lease, nên không thể biến drill này thành activation thật.
`m11-gate` cũng phải in `STOPPED`, không resolve các ID `missing-*` và không
tạo `m11-artifacts.jsonl`; sau STOP, reconciliation là post-STOP mutation duy
nhất được phép.

Target của `backup create` cũng phải trống và không được là runtime hoặc thư
mục con của runtime; tạo backup mới vào một thư mục khác thay vì ghi đè snapshot
cũ.

## Gates bắt buộc

- health check fail hoặc lease hết hạn → không activate;
- STOP durable được ghi trước, đọc lại sau restart và chặn mọi attempt;
- `backup create` và mọi writer managed dùng cùng runtime gate. Nhận `BUSY`
  nghĩa là giữ nguyên state, chờ writer/snapshot kết thúc hoặc thực hiện recovery
  rõ ràng cho gate stale; không xóa lock tự động;
- backup history + ledger và, nếu M07 adapter đã persist, toàn bộ sidecar
  `history.jsonl.m07/`; manifest v3 chỉ nhận relative path chuẩn hóa trong
  layout artifact đã biết, và ghi kind, kích thước cùng SHA-256. Restore đòi
  inventory thực tế trùng manifest trước khi chép sang thư mục cô lập;
- backup `v2` không tương thích với verifier `v3`; khi nâng cấp, tạo một
  snapshot mới từ runtime gốc trước, không sửa tay manifest cũ;
- gate M11 phải có `ledger_artifact_id` và `ledger_content_hash`. Gate đời cũ
  thiếu hai binding này không được restore để cấp authorization; tạo gate mới
  từ ledger hiện hành qua luồng review, không sửa lại artifact bất biến;
- restore xong phải replay khớp, không reset reservation/STOP;
- recovery cần human review, không tự mở lại executor.
- restore không được tự tạo lại ledger, reservation, approval hoặc lease đã
  hết hạn; artifact đã hết hạn vẫn phải replay/resolve được từ store để audit,
  nhưng gate/authorization/reservation mới phải bị chặn theo thời gian hiện tại.

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
