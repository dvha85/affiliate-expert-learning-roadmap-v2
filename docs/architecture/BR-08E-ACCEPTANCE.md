# BR-08e — Review tích hợp và nghiệm thu

Baseline sau #50 `46a8e45`. BR-08a/#47 `8ecd60a`, b/#48 `56a1ccf`, c/#49 `b8fa83b`, d/#50 `46a8e45` đã review/merge. BR-08e chờ review PR, không tự đóng item cha.

## Tiêu chí và bằng chứng

| Tiêu chí BR-08 | Evidence | Kết luận kỹ thuật |
|---|---|---|
| Shared capability không copy | core/m03 sở hữu types/decode/validation, harness wrapper; import core chỉ contracts/strings/time | Đạt lát cắt M03; chưa extract toàn M04–M11 |
| Expected độc lập | Core tests literal; harness eval fixtures không đổi; scripts/smoke_br08.py so từng CLI với literal status, không lấy output CLI này làm expected của CLI kia | Đạt |
| File người học qua learner entrypoint | bot action validate nhận hai path, binary integration tests và smoke valid/window/link/schema | Đạt local validation, chưa resolve decision store |
| Phân biệt validate/demo/persist/execute | action validate chỉ đọc; history capture ghi; demo M09–M11 sandbox; chưa thêm live execute | Đạt phạm vi ADR |
| Store owner rõ | internal/store.History/JSONL chỉ I/O; learner application giữ validation/ID/hash/conflict/replay | Đạt seam; orchestration M02 vẫn ở cmd/bot, không claim đã tách toàn app |
| Compatibility | History/decision adapter/legacy tests; smoke capture→duplicate→list→replay MATCH không rewrite | Đạt lab một writer |
| Hướng dẫn | BR-08C-ACTION-CLI và BR-08D-HISTORY-STORE, README learner | Có đường mở rộng cùng Bot |

## Chạy từ checkout sạch

Sau clone repo, chạy từ root:

```sh
python3 scripts/smoke_br08.py
```

Script build hai binary vào TemporaryDirectory, GOWORK=off, copy fixture synthetic vào đó rồi kiểm status/exit/stdout và file không đổi. Expected: VALID, MEASUREMENT_WINDOW_OPEN, BROKEN_LINK, INVALID_SCHEMA; usage exit 2; M02 APPENDED/EXACT_DUPLICATE/MATCH. Không ghi fixture repo hoặc history người học. Có thể đặt GO_BIN để chọn executable Go; không đổi go.mod/toolchain.

Full regression: tests/vet bốn module core, contracts, lab/affiliate-bot, lab/mission-runtime với GOWORK=off; 8 validators và 10 Python regressions. CI bổ sung smoke hai binary trong deterministic-runtime; mission-runtime vẫn chạy core với toolchain riêng và demo cũ.

### Kết quả kiểm chứng ngày 2026-09-06

Clone local độc lập bằng `git clone --no-local --branch codex/br-08e-acceptance` tại snapshot `f5730c0`: smoke PASS cả bốn case M03, usage và M02 capture/duplicate/list/replay; tests/vet cả bốn module PASS; 8 validators và 10 Python regressions PASS. `git status --porcelain` trống trước và sau kiểm thử. Sau snapshot này, bước smoke CI được chuyển xuống sau setup-python để dùng Python 3.12 đã khai báo.

Đây là clone sạch từ repository local, dùng Go/toolchain/cache đã có trên máy; không phải bằng chứng cài đặt cold-cache, clone remote trên máy người mới hoặc vận hành affiliate thật.

## Kết luận đề xuất

READY_FOR_REVIEW cho BR-08 trong phạm vi shared M03 + learner CLI read-only + store seam M02. Chỉ đổi BR-08 DONE sau review/merge PR nghiệm thu và chấp thuận scope. Không claim pilot máy người mới/Windows, toàn bộ Bot liên tục M00–M11 hoặc production-ready.

BR-09 mới nối M00→M01/M02; BR-10 mới persist action/outcome và resolve decision; BR-06b vẫn chờ chương trình/kênh. Store một writer, partial-write/crash recovery, provenance/trusted clock vẫn thuộc backlog riêng. Không đổi JSONL/hash hoặc reset ledger để nghiệm thu.
