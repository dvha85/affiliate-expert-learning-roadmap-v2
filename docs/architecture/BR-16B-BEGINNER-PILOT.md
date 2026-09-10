# BR-16b — walkthrough và pilot người mới

## Phạm vi và chuẩn bị

Chạy từ repo root với Go theo `lab/affiliate-bot/go.mod` và Python 3.  Đường
này tạo **fixture offline** trong một thư mục trống do người học chọn; nó không
gọi ACCESSTRADE, không gọi provider/model, không thực thi affiliate action và
không tạo business outcome. JSON tạo ra là artifact kỹ thuật để kiểm lineage,
không phải affiliate truth.

```bash
WORKSPACE="$(mktemp -d /tmp/br16a-walkthrough.XXXXXX)"
python3 scripts/smoke_br16a_offline.py --workspace "$WORKSPACE"
```

Không dùng thư mục có sẵn dữ liệu: `--workspace` từ chối thư mục không trống và
không tự xóa workspace do người học cung cấp. Lệnh thành công in `BR-16a PASS`
và giữ toàn bộ artifact tại `$WORKSPACE`; thất bại phải được coi là `INCOMPLETE`,
không bỏ qua assertion hoặc sửa tay artifact để đi tiếp.

## Kiểm M00–M07 và canonical history

```bash
python3 -m json.tool "$WORKSPACE/walkthrough-result.json"
"$WORKSPACE/bot" history replay "$WORKSPACE/restored/history.jsonl"
"$WORKSPACE/bot" m07 context "$WORKSPACE/restored/history.jsonl" br16-d
```

Expected: report có `m00_to_m11_shared_lineage: "PASS"` và
`history_replay_after_restore: "MATCH"`; replay in `replay=MATCH`; M07 context
có `evidence_ids` resolve được. Input M00/M06/M07 và outputs hiện hữu nằm trong
`$WORKSPACE/runtime/` và `$WORKSPACE/restored/`; không ghi đè chúng.

Để xem đường M00 độc lập (không thay artifact của walkthrough), dùng một thư
mục khác:

```bash
cd lab/affiliate-bot
go run ./cmd/bot evidence import ../../examples/m00-import/packet-t2.json
```

Expected: JSON `status: "APPENDED"` cùng artifact observation. Một forged ID,
POST tool call, hoặc history bị hỏng phải exit non-zero và không được có output
success.

## Chuỗi M08–M11 cùng artifact/store

Smoke đã tạo, bind và dùng cùng M07 proposal → M08 intent/policy → M09 approval
→ M10 grant/cost/gate/authorization/reservation → M11 lease/outcome/evaluation/
cycle. Các đường dẫn chính được ghi trong report, nên không cần tự đoán ID hoặc
tạo JSON/hash bằng tay:

```bash
python3 - <<'PY'
import json, os
report = json.load(open(os.environ["WORKSPACE"] + "/walkthrough-result.json"))
for name, path in report["paths"].items():
    print(f"{name}: {path}")
PY

"$WORKSPACE/bot" mission status "$WORKSPACE/restored"
"$WORKSPACE/bot" mission m11-resolve "$WORKSPACE/restored" \
  PRODUCTION_OUTCOME_EVALUATION br16-production-e
"$WORKSPACE/bot" mission m11-resolve "$WORKSPACE/restored" \
  PRODUCTION_CYCLE br16-production-cycle
```

Expected: `status` trả `stop: true` với
`stop_reason: "RECONCILIATION_REQUIRED"`; hai lệnh resolve trả `RESOLVED`.
Đó là bằng chứng rằng output từ M08–M11 đã được persist rồi load lại sau
backup/restore, không phải sáu smoke độc lập.

## Failure/fix và restart drill

Smoke đã chạy hai failure cases bắt buộc trước khi PASS: request M08 đổi target
so với M07 proposal bị `REJECTED`, và M10 output path collision trả `CONFLICT`
nhưng canonical reservation vẫn được bind. Xem state sau restart và xác nhận
STOP vẫn chặn canary:

```bash
"$WORKSPACE/bot" mission status "$WORKSPACE/runtime"
"$WORKSPACE/bot" mission m10-canary "$WORKSPACE/runtime" \
  "$WORKSPACE/grant.json"; test "$?" -ne 0
```

Expected: status có `stop: true`; lệnh canary in JSON `status: "STOPPED"` và
exit non-zero. `test` chỉ xác nhận đây là failure có chủ đích. Không sửa hay xóa
STOP để làm lệnh thành công; recovery cần runtime/lease/approval mới đã được
review theo M11.

## Backup/restore drill

`$WORKSPACE/backup/` và `$WORKSPACE/restored/` là backup artifact tạo bởi runtime
thật. Để lặp lại standalone với một runtime mới, xem
`BR-18A-DEPLOYMENT-RECOVERY-RUNBOOK.md` và chạy:

```bash
python3 scripts/smoke_br18b_backup_restore.py
```

Expected: `BR-18b PASS`; smoke kiểm manifest tamper, loader/graph, restart,
budget/canary link và durable STOP. Nó vẫn là proof offline; không suy ra
power-loss atomicity, 24/7 availability hay provider/business result.

| Chặng | Người học tự làm | Fixture/automation đã có | Cần ghi nhận |
|---|---|---|---|
| Setup | clone, chọn working directory, chạy quickstart | smoke_quickstart | lỗi môi trường |
| History | import observation, list và replay | BR-13b smoke | record ID, retry/conflict |
| Action/outcome | validate link, thử pending/0 | BR-10a/b/c smoke | EffectRef, orphan ID |
| Advisor/review | đọc abstain và evidence limits | BR-11/12 smoke | unknowns, human review |
| Watcher | chạy fixture fetch/import, đổi nội dung | BR-13b/c smoke | NEW/duplicate, provenance |
| Agent | tạo output JSON có `claims[].field_or_claim`, `claims[].value` và `claims[].evidence_ids`, chạy validator | `bot m07 context/validate` | forged ID, forged value, prompt injection, write request |
| M08–M11 | chạy workspace BR-16a, resolve evaluation/cycle sau restore, xác nhận STOP reject canary | `bot mission ...`, BR-16a | path trong report, `RESOLVED`, `STOPPED` |

## Mẫu bản ghi pilot

```text
pilot_id: <ẩn danh>
started_at_utc: <RFC3339>
completed_at_utc: <RFC3339>
repo_commit: <sha>
completed_steps: [..]
minutes_total: <number>
questions: [..]
assistance_given: [..]
failure_explanations: [..]
residual_gaps: [..]
learner_can_explain_input_output_limit: true|false
result: PASS | NEEDS_HELP | INCOMPLETE
```

Một pilot chỉ PASS khi người học tự hoàn thành đường walkthrough, giải thích
được input/output, một failure case và giới hạn authority. Hỗ trợ thiết kế/code
phải ghi rõ, không tính là self-service PASS. Chưa có pilot thật thì giữ
`INCOMPLETE`.
