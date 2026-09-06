# BR-10c — Audit nghiệm thu chain learner M00 → M03

READY_FOR_REVIEW, chưa tự chốt BR-10 DONE. Baseline #53 `765ac70`, #54 `533cc6f` đã merge. Review #54 head `2e4d276`: không finding chặn trong scope canonical JSON/lab một writer; smoke BR-10b và tests/vet learner chạy lại PASS, CI 4/4 PASS. Các giới hạn không thuộc phạm vi này vẫn mở.

Trạng thái hiện hành: BR-10 lab/schema DONE sau review/merge #55 `6d311a8`; audit chạy lại PASS, CI 4/4 PASS, không finding chặn. Các trạng thái chờ review bên dưới là lịch sử. Importer nền tảng và live proof tách thành BR-10d trong kế hoạch, chưa hoàn thành.

PR audit: [#55](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/55), chưa merge. Clone local độc lập `git clone --no-local` snapshot `f4cefbc`: audit, tests/vet bốn module, 8 validators và 10 Python regressions PASS; git status trống trước/sau. Dùng toolchain/cache sẵn có, không cold install hoặc máy học viên. Sau snapshot chỉ bổ sung bằng chứng/liên kết tài liệu.

## Cách chạy và bằng chứng

Từ root checkout sạch, có Go theo go.mod và Python 3:

```sh
python3 scripts/smoke_br10c.py
```

Script build learner binary trong TemporaryDirectory, GOWORK=off (có thể đặt GO_BIN). Bắt đầu bằng examples/m00-import/packet-t2.json, không dùng artifact cá nhân; mỗi CLI invocation là process mới. Expected IDs/status/metrics là literal độc lập, không lấy output implementation làm oracle cho chính nó. Chỉ mutation bản sao trong workspace tạm. Không ghi PROGRESS, không gọi platform hoặc tự tạo proof thật.

```text
M00 -> decision audit-d -> action audit-a -> outcomes p/z/l: exact IDs, restart/list/replay PASS
Broken upstream, duplicate/corrupt outcome store and hardlink alias: fail closed PASS
Outcome payload limit -1/exact/+1, rejected write isolation: PASS
BR-10c AUDIT PASS: synthetic lab only; no platform/live proof
```

## Ma trận nghiệm thu

| Tiêu chí | Bằng chứng | Phạm vi |
|---|---|---|
| Cùng Bot và workspace | evidence import → history capture → history decision → action record → outcome import | Một binary, ba JSONL riêng, không harness store |
| Exact linkage | DecisionPacket audit-d có evidence_ids=[br09-product-a-t2]; action audit-a trỏ audit-d; p/z/l đều HUMAN_ACTION/audit-a | Aggregate M00 giữ field provenance trong transformation |
| Restart/read/replay | process mới history list/replay, action list, outcome list; exact artifacts; MATCH | Không chỉ kiểm object trong RAM |
| Missing/pending/zero/late | BR-09 smoke thiếu commission → GET_MORE_DATA; BR-10b/c giữ pending metrics={}, zero clicks=0, late PAID ID mới | Không tự cộng các snapshot hoặc diễn giải như doanh thu thật |
| Window/EffectRef | BR-10b tests/smoke before action, window open, exact end khác timezone, machine/orphan | M03 shared semantics, không trusted clock |
| ID bất biến | BR-10a/b duplicate no-write, changed content conflict; BR-10c duplicate store bị reject | Một writer, không concurrent CAS |
| Đứt chain | Empty history → ACTION_STORE_ERROR; empty actions → STORE_ERROR khi đọc outcomes | Không tiếp tục với downstream orphan |
| Corruption/path safety | Corrupt/duplicate outcome store bị chặn; hardlink alias → PATH_ERROR; snapshot bytes upstream không đổi | Không bảo đảm chống race của tiến trình khác |
| Persistence size | Outcome payload 1 MiB-1/đúng/1 MiB+1 qua binary: trong giới hạn list được, vượt giới hạn không tạo file | Dùng cùng store limit; history boundary đã có BR-09 regression |
| Hướng dẫn | [BR-10a](BR-10A-ACTION-STORE.md), [BR-10b](BR-10B-OUTCOME-STORE.md), audit này | Fixture synthetic, không gọi thao tác giả lập là human-operated proof |

## Kết luận và phần còn mở

Đề xuất nghiệm thu **lát cắt lab/schema canonical JSON** sau review/merge BR-10c. Không đổi trạng thái cha sang DONE khi PR audit còn mở. Đây là kiểm chứng kỹ thuật, không chứng minh học viên mới tự chạy được; pilot thuộc BR-16.

BR-10 checklist gốc có importer format báo cáo BR-06. Hiện mới nhận canonical JSON được map thủ công, chưa có adapter chương trình cụ thể. Phần này phải giữ riêng OPEN/BLOCKED theo chương trình được chọn; không dùng audit này để tuyên bố đã nhập báo cáo nền tảng. BR-06b tiếp tục chờ tài khoản/kênh/nguồn thật.

Các giới hạn khác: một writer; không transaction/crash recovery; không immutable upstream digest/context binding; hash không là chữ ký; timestamp/source_ref do caller khai báo; chưa transaction identity/supersedes/reconciliation; không outcome business aggregation; không execution permission hoặc production readiness. Không sửa store để làm nghiệm thu PASS.
