# BR-18a — interim macOS local verification decision (2026-09-21)

Trạng thái: **LOCAL_OPERATED_PASS / REVIEW_PENDING** cho bounded macOS lab run;
`BR-18a` ở phạm vi target-host/provider vẫn **UNVERIFIED**, và overall
readiness vẫn `NOT_READY_FOR_PRODUCTION`. Đây chưa phải operated target-host
evidence.

## Quyết định hiện tại

- Dùng **macOS** của người vận hành để kiểm chứng trước, thay vì chờ máy
  Windows hoặc mua VPS. Profile này chạy trực tiếp trên macOS bằng Terminal;
  Docker không phải prerequisite.
- Dùng dữ liệu synthetic/read-only, không provider credential, không live
  executor và không ghi dữ liệu affiliate thật.
- Chạy Go 1.27, Node 24 và n8n 2.38.1; phải ghi exact version output trong
  run record, không suy ra version từ tài liệu này.
- Mặc định chỉ bind loopback: adapter `127.0.0.1:8787`, n8n
  `127.0.0.1:5678`. Không port-forward hoặc public ingress; private LAN chỉ
  được bật khi có firewall rule cụ thể và được ghi vào run record.
- Giữ runtime, backup và restore tách nhau trên filesystem macOS:

  ```text
  $HOME/affiliate-lab/
  ├── runtime/     # canonical state, managed writers only
  ├── backups/     # backup targets, separate from runtime
  ├── restores/    # empty restore targets, separate from both
  ├── var/log/     # process logs
  └── var/run/     # pid files
  ```

## Operated macOS run — 2026-09-21

Run record: `macos-br18a-20260921T155228Z`, lưu tại
`$HOME/affiliate-lab/runs/macos-br18a-20260921T155228Z`. Reviewer: maintainer
review pending; các kết quả dưới đây là output trực tiếp của automated local
run, không phải human approval hay production acceptance.

| Trường | Giá trị |
|---|---|
| `tested_at_utc` | `2026-09-21T15:52:28Z` |
| host / OS | `Đinh’s MacBook Air` / `macOS 27.0 (26A428)` |
| Go | `go1.27.0 darwin/arm64` |
| Node | `v24.21.0` |
| n8n | `2.38.1` |
| bind | adapter `127.0.0.1:8787`; n8n `127.0.0.1:5678` |
| paths | runtime `$HOME/affiliate-lab/runtime`; backup và restore ở hai thư mục riêng |
| auth | `CANONICAL_ADAPTER_TOKEN` (synthetic local token, không ghi vào log) |

### Results

- n8n disposable health: `PASS`, trả `{"status":"ok"}`; process dừng sạch
  bằng `SIGTERM`.
- Canonical adapter health và authenticated `/v1/history?record_id=demo-1`:
  `PASS`.
- Process restart rồi health lại: `PASS`.
- History replay trước STOP: `PASS`.
- `mission stop` durable marker: `PASS`; restore vẫn trả `stop: true`.
- Backup/restore vào thư mục trống và replay sau restore: `PASS`.
- Sau restore, `m11-activate` và `m11-gate` đều bị từ chối (`exit 1`), không
  tạo `m11-artifacts.jsonl`: `PASS`.
- `scripts/run_n8n_engine_regression.py`: `PASS`; M06 fixture và selected-source
  sanitized metadata append/replay, M07 loopback model-stub grounding và các
  forged-commission/write-boundary rejects đều fail closed.
- `scripts/run_n8n_m06_schedule_regression.py`: `PASS`; Schedule Trigger ghi
  một append, các lần retry là `EXACT_DUPLICATE`, retry sau n8n/adapter restart
  vẫn idempotent, và adapter unavailable bị chặn trước ACK/report.

Run này chỉ dùng fixture/synthetic, loopback và read-only boundary; không có
provider credential, public ingress, live executor, affiliate write hay business
outcome. Các cảnh báo n8n về Python task runner và việc chạy native ngoài Docker
được giữ trong log; chúng không làm thay đổi kết quả bounded JavaScript/M06/M07
run, nhưng cần xử lý riêng trước deployment production.

## Phương án Windows về sau

Khi cần chạy 24/7 hoặc kiểm chứng boundary triển khai khác, có thể chuyển cùng
profile sang máy Windows + Ubuntu 24.04 LTS trong WSL2; Hyper-V + Ubuntu 24.04
LTS vẫn là fallback khi cần VM boundary rõ hơn. Runbook, synthetic/read-only
boundary, version pins và điều kiện evidence được giữ nguyên; chỉ host/guest,
filesystem và network profile thay đổi. Windows profile không bị loại bỏ bởi
quyết định macOS hiện tại.

## Phạm vi có thể kiểm chứng

Sau khi người vận hành thực sự chạy profile và lưu run record, macOS có thể
kiểm chứng bounded lifecycle, health endpoint, backup/restore vào thư mục
trống, replay, process restart và durable STOP trong một filesystem POSIX đơn
host. Kết quả này là local verification, không phải bằng chứng Linux guest,
Windows host, provider hay public deployment.

Profile này **không** đóng các khoảng trống sau:

- target-host/provider/public-network deployment hoặc TLS/Internet exposure;
- Ubuntu-on-WSL2/Hyper-V và Windows-native runtime parity;
- region latency, cloud availability, managed backup và failure của power-loss;
- atomic multi-file durability, distributed/multi-host locking;
- clean-machine beginner pilot, live executor, affiliate outcome hoặc business
  result.

Vì vậy local macOS slice đã có operated run record, nhưng `BR-18a` ở phạm vi
target-host/provider vẫn `UNVERIFIED` và overall readiness vẫn
`NOT_READY_FOR_PRODUCTION`. Existing BR-18b smoke là fixture/read-only
evidence riêng; run này cũng không được nâng cấp thành Linux guest, Windows
host, public deployment hay production evidence.

## Run-record gate

Evidence local chỉ được giữ là `LOCAL_OPERATED_PASS` khi có
`tested_at_utc`, exact macOS host/OS profile, Go/Node/n8n version output, bind
address, runtime/backup/restore paths, health result, restart result, replay
result và durable-STOP result. Human reviewer vẫn phải xác nhận record trước
khi gọi đây là acceptance; khi thiếu run record hoặc reviewer, trạng thái phải
quay về `UNVERIFIED`/`REVIEW_PENDING`.
