# BR-18a — interim Windows always-on lab decision (2026-09-21)

Trạng thái: **DECISION_RECORDED / UNVERIFIED**. Đây là quyết định topology
trung gian cho lab always-on; chưa phải operated target-host evidence và chưa
đóng BR-18a.

## Quyết định

- Tạm hoãn VPS trả phí. Máy Windows của chủ repo có thể chạy 24/7 cho lab
  synthetic/read-only; các trao đổi giá Contabo/VNPT/DigitalOcean chỉ là tham
  khảo, không phải bằng chứng đã mua hoặc đã triển khai cloud resource.
- Ưu tiên **Ubuntu 24.04 LTS trong WSL2**. Dùng **Hyper-V với Ubuntu 24.04
  LTS** khi cần VM boundary rõ hơn hoặc WSL2 không phù hợp.
- Chạy runtime Linux trực tiếp trước: Go binary, Node.js và n8n; Docker/Compose
  là lựa chọn sau, không phải prerequisite của profile này.
- Chỉ dùng dữ liệu synthetic/read-only. Mặc định adapter và n8n bind loopback;
  nếu cần thử từ máy khác trong LAN thì chỉ bind private address được firewall
  allowlist, không port-forward/public ingress.
- Giữ canonical runtime và backup trên filesystem Linux, không dùng `/mnt/c`
  làm store chuẩn. Layout tham chiếu:

  ```text
  /home/<linux-user>/affiliate-lab/
  ├── runtime/     # canonical state, managed writers only
  ├── backups/     # backup targets, separate from runtime
  ├── restores/    # empty restore targets, separate from both
  ├── var/log/     # process logs
  └── var/run/     # pid files
  ```

- Supported lab pins: **Go 1.27**, **Node 24**, **n8n 2.38.1**. Pin/version
  output must be recorded in a later run record rather than inferred from this
  decision.

## Phạm vi có thể kiểm chứng

Sau khi người vận hành thực sự chạy profile và lưu run record, lab này có thể
kiểm chứng bounded lifecycle, health endpoint, backup/restore vào thư mục
trống, replay, process/guest restart và durable STOP. Có thể ghi thêm CPU/RAM/
disk limits của lab như deployment configuration, nhưng không suy ra capacity
production.

Profile này **không** đóng các khoảng trống sau:

- target-host/provider/public-network deployment hoặc TLS/Internet exposure;
- region latency, cloud availability, managed backup và failure của power-loss;
- atomic multi-file durability, distributed/multi-host locking;
- clean-machine beginner pilot, live executor, affiliate outcome hoặc business
  result.

Vì vậy `BR-18a` vẫn `UNVERIFIED` và overall readiness vẫn
`NOT_READY_FOR_PRODUCTION`. Existing local BR-18b smoke là fixture/read-only
evidence riêng, không được nâng cấp thành Windows operated evidence chỉ từ
quyết định topology này.

## Run-record gate

Chỉ tạo evidence PASS sau khi có `tested_at_utc`, exact host/guest profile,
Go/Node/n8n version output, bind address, runtime/backup/restore paths, health
result, restart result, replay result, durable-STOP result và reviewer. Khi
chưa có các trường đó, runbook phải ghi `result: UNVERIFIED`.
