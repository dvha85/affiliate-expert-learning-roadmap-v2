# BR-18a — interim macOS local verification decision (2026-09-21)

Trạng thái: **DECISION_RECORDED / UNVERIFIED**. Đây là lựa chọn môi trường
kiểm chứng hiện tại cho lab local; chưa phải operated target-host evidence và
chưa đóng BR-18a.

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

Vì vậy `BR-18a` vẫn `UNVERIFIED` và overall readiness vẫn
`NOT_READY_FOR_PRODUCTION`. Existing BR-18b smoke là fixture/read-only
evidence riêng, không được nâng cấp thành macOS operated evidence chỉ từ quyết
định topology này.

## Run-record gate

Chỉ tạo evidence PASS sau khi có `tested_at_utc`, exact macOS host/OS profile,
Go/Node/n8n version output, bind address, runtime/backup/restore paths, health
result, restart result, replay result, durable-STOP result và reviewer. Khi
chưa có các trường đó, runbook phải ghi `result: UNVERIFIED`.
