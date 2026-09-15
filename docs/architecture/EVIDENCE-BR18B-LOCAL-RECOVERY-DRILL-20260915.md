# BR-18b — local deployment/recovery drill (2026-09-15)

Trạng thái: **local runtime evidence / fixture-only / chưa phải target-host
deployment evidence**.

## Run record

- `drill_id`: `br18b-local-recovery-20260915-7be0ccc`
- `repo_commit`: `7be0ccc59aa05d4068ade3193ace0d3c26295aa5`
- `tested_at_utc`: `2026-09-15`
- `runtime`: learner Bot thật, build từ `lab/affiliate-bot/go.mod`
- `host`: local Darwin arm64
- `go`: `go1.27.0`
- `python`: `3.9.6`
- `result`: `PASS`

## Command and result

Command chạy từ repo root:

```text
python3 scripts/smoke_br18b_backup_restore.py
```

Output:

```text
BR-18b PASS: runtime-created M10 graph, M11 fixture evaluation/cycle, and UNKNOWN-to-human-reconciliation chain use a typed v3 manifest; checksum, exact inventory, lease-window activation, activation-bound health, broken evaluation/cycle links, reversed cycle time, restart, and durable STOP are verified
```

Smoke tạo artifact bằng Bot đã biên dịch trong thư mục tạm, sau đó thực hiện
backup/restore bằng runtime thật và đọc lại bằng process mới. Các kiểm tra bao
gồm inventory/hash/kind/path, graph M10/M11, budget/lease-window, restart,
replay và STOP durable; các mutation/fault case phải reject mà không mở lại
execution.

## Giới hạn

Đây là local fixture evidence, không phải deployment recovery drill trên target
host. Không có VPS/service production, provider request, ACCESSTRADE/Cockpit,
affiliate link, live executor hay business outcome. Nó cũng không chứng minh
atomic multi-file recovery sau power-loss, distributed locking hoặc
24/7 availability. BR-18b/RP-10 vẫn `PARTIAL`/`OPEN` cho các bằng chứng
external tương ứng; readiness vẫn `NOT_READY_FOR_PRODUCTION`.
