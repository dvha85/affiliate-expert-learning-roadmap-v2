# BR-16b — walkthrough và pilot người mới

## Đường walkthrough

Chạy từ repo root. Tất cả lệnh dưới đây dùng cùng một thư mục tạm; ID ở bước
trước được đưa vào input bước sau. Output JSON là artifact kỹ thuật, không phải
affiliate truth.

```bash
python3 scripts/smoke_br16a_offline.py

cd lab/affiliate-bot
go run ./cmd/bot evidence import ../../examples/m00-import/packet-t2.json \
  | python3 -c 'import json,sys; print(json.dumps(json.load(sys.stdin)["artifact"]))' > /tmp/m00-observations.json
go run ./cmd/bot history capture /tmp/history.jsonl /tmp/m00-observations.json demo-d \
  2026-09-03T00:00:00Z 2026-09-03T00:00:00Z
go run ./cmd/bot m07 context /tmp/history.jsonl demo-d
go run ./cmd/bot watcher fixture-import /tmp/history.jsonl \
  ../../examples/watcher/offer-valid.json
go run ./cmd/bot history replay /tmp/history.jsonl
```

Expected: `APPENDED` on first writes, `EXACT_DUPLICATE` on retry, an M07
context containing resolvable `evidence_ids`, watcher `RANK_SCENARIO`, and
`replay=MATCH`. A forged ID, a POST tool call, or corrupt history must exit
non-zero and must not write a success artifact.

| Chặng | Người học tự làm | Fixture/automation đã có | Cần ghi nhận |
|---|---|---|---|
| Setup | clone, chọn working directory, chạy quickstart | smoke_quickstart | lỗi môi trường |
| History | import observation, list và replay | BR-13b smoke | record ID, retry/conflict |
| Action/outcome | validate link, thử pending/0 | BR-10a/b/c smoke | EffectRef, orphan ID |
| Advisor/review | đọc abstain và evidence limits | BR-11/12 smoke | unknowns, human review |
| Watcher | chạy fixture fetch/import, đổi nội dung | BR-13b/c smoke | NEW/duplicate, provenance |
| Agent | tạo output JSON có `claims[].field_or_claim`, `claims[].value` và `claims[].evidence_ids`, chạy validator | `bot m07 context/validate` | forged ID, forged value, prompt injection, write request |
| M08–M11 | tạo intent → policy → approval → bounded canary → STOP | `bot mission ...`, BR-16a | hash/link, human approval, durable STOP |

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
