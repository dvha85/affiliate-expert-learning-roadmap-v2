# BR-10b — Nhập OutcomeRecord và resolve ActionRecord

Triển khai chờ review; BR-10 chưa DONE. Baseline #53 đã review/merge `765ac70`: tests/vet learner, smoke BR-10a và CI 4/4 PASS, không finding chặn trong scope lab một writer. Không live proof hoặc account affiliate.

PR [#54](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/54) đã review/merge `533cc6f`; CI 4/4, smoke BR-10b và tests/vet learner chạy lại PASS, không finding chặn scope lab. [BR-10c audit](BR-10C-ACCEPTANCE.md) kiểm toàn chain; BR-10 chưa tự DONE.

## Profile và ownership

```text
bot outcome import HISTORY.jsonl ACTIONS.jsonl OUTCOMES.jsonl INPUT.json
bot outcome list HISTORY.jsonl ACTIONS.jsonl OUTCOMES.jsonl
```

INPUT là canonical OutcomeRecord theo contracts/outcome-record.schema.json, không phải CSV export của nền tảng. Người học map báo cáo theo BR-06 trước, không tự điền commission hoặc đổi pending thành paid. Store outcomes riêng, không trộn vào history/actions. Application cmd/bot resolve history → action → outcome; core M03 giữ semantics, store JSONL chỉ I/O. Không execution/network/đăng bài. Không sửa schema/hash/formula hoặc ledger cũ.

Mỗi lần import/list đọc và kiểm upstream history, action store; action phải resolve đúng một decision replay MATCH. EffectRef phải HUMAN_ACTION và effect_id trỏ action tồn tại. MACHINE_EXECUTION bị từ chối, không ép thành human. Outcome trước performed_at bị OUTCOME_BEFORE_ACTION. NO_OBSERVED_OUTCOME trước measurement_window_end bị MEASUREMENT_WINDOW_OPEN; tại đúng thời điểm (kể cả timezone khác) được nhận. PENDING/VALID/PAID/CANCELLED/REFUNDED là snapshot giao dịch, không tự kết luận cả cửa sổ; có thể xuất hiện trước window end nếu không trước action.

Metrics giữ nguyên tên và giá trị không âm; `{}` nghĩa chưa khai báo metric, khác `{ "clicks": 0 }`. Null không hợp schema nên bị từ chối, không đổi thành 0. PENDING giữ PENDING kể cả metric có 0. Không suy luận paid từ valid_orders, không cộng các snapshot để tính doanh thu. Numeric literal tối đa 128 ký tự, exponent [-400,400]; mất nghĩa thập phân qua float64 hoặc underflow bị METRIC_PRECISION_ERROR.

Outcome ID bất biến: cùng ID/cùng typed content → EXACT_DUPLICATE không ghi; cùng ID/nội dung khác → CONFLICT. Late update dùng ID mới, cùng EffectRef nếu cùng action; không ghi đè observation cũ. Chưa có supersedes/transaction identity/reconciliation nên list chỉ là các snapshot, không chọn kết quả mới nhất hay tổng hợp business metrics.

## Output và lỗi

JSON stdout gồm command/status/execution_permitted=false; APPENDED/EXACT_DUPLICATE có artifact outcome, list VALID có array. Lỗi không có artifact; diagnostic stderr. Exit 0 thành công, 1 lỗi, 2 usage. Store outcome chưa tồn tại chỉ được tạo khi import hợp lệ; list báo STORE_ERROR. Store hỏng/duplicate IDs/orphan link bị chặn, không sửa hoặc bỏ qua. Paths phải khác nhau kể cả alias file tồn tại; không redirect vào input/upstream/store. Payload tối đa 1 MiB không tính LF/CRLF, dùng cùng giới hạn read/write của BR-09.

## Bài thực hành synthetic

Từ root, chạy `python3 scripts/smoke_br10b.py` (Go theo go.mod, Python 3). Script build binary và tạo mọi file trong thư mục tạm riêng, process mới cho mỗi lệnh. Nó capture decision d, record action a, import ba snapshot dưới đây và đọc lại exact array:

| outcome_id | status | observed_at | metrics | EffectRef |
|---|---|---|---|---|
| pending | PENDING | 2026-09-04T01:00:00Z | {} | HUMAN_ACTION/a |
| zero | NO_OBSERVED_OUTCOME | 2026-09-05T07:00:00+07:00 | clicks=0 | HUMAN_ACTION/a |
| late | PAID | 2026-09-06T00:00:00Z | commission=8 | HUMAN_ACTION/a |

Action performed_at=2026-09-04T00:00:00Z, window_end=2026-09-05T00:00:00Z. source_ref=fixture:synthetic-report ở cả ba, không có giao dịch thật. Mẫu đầy đủ để copy vào INPUT.json trong workspace lab riêng:

```json
{"outcome_id":"pending","effect_ref":{"effect_kind":"HUMAN_ACTION","effect_id":"a"},"observed_at":"2026-09-04T01:00:00Z","status":"PENDING","metrics":{},"source_ref":"fixture:synthetic-report"}
```

Dùng action_id thật của workspace lab khi thực hành thủ công; không sửa history để bypass orphan. Đổi effect_id thành absent, thử thời điểm trước action, rồi thử NO_OBSERVED_OUTCOME trước/sau window. Muốn thay metric dưới ID cũ phải thấy CONFLICT; tạo ID mới cho snapshot mới. Source/time là khai báo, không phải trusted evidence.

Output smoke: `BR-10b PASS: linked outcomes; pending/zero/late; duplicate/conflict/orphan/window; restart; synthetic only`.

## Giới hạn

Tests unit và smoke bao phủ expected độc lập, no-write trên semantic rejection, upstream bytes không đổi và restart. Vẫn lab một writer, không transaction/lock/crash-safe hoặc trusted clock, không bind digest upstream, không xác thực nguồn, không importer nền tảng/live report. Failure sau write có thể đã append; không giả định rollback khi stdout/disk lỗi. BR-06b vẫn chờ chương trình/kênh; BR-10 cần review nghiệm thu toàn chain trước DONE.
