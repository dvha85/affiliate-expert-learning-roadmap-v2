# BR-11a — Mock advisor M04 offline

Triển khai chờ review; BR-11 chưa DONE. BR-10 lab đã audit/merge #55 `6d311a8`. Mock này đọc history/action/outcome canonical JSON, tạo context có version và trả HUMAN_REVIEW; không gọi provider, không ghi store, không thực thi.

## Lệnh

```text
bot advisor mock HISTORY.jsonl ACTIONS.jsonl OUTCOMES.jsonl CONFIG.json
```

CONFIG yêu cầu `decision_id`, `question`, RFC3339 `as_of`, `max_age_hours` trong 0..8760. Context resolve decision replay MATCH → action → outcome và evidence IDs exact; duplicate cross-kind ID bị từ chối. Mỗi evidence có observed_at/source_ref. `as_of` và max age được truyền rõ, không dùng clock ngầm.

Mock output dùng shared M04 advisor boundary/schema, `provider=mock/v1`, `state=HUMAN_REVIEW`, reason/unknowns rõ và `write_tool_requested=false`. `SUPPORTED` chỉ nghĩa context/evidence references hợp lệ theo freshness, không là recommendation, approval, truth hay execution permission. Không có recommendation business; mock không thay live proof.

Stale evidence trả `ABSTAIN_STALE`; evidence tương lai trả `ABSTAIN_FUTURE`; decision/action/outcome thiếu trả CONTEXT_ERROR hoặc upstream error. Output malformed trả INVALID_SCHEMA, write request trả REJECT_WRITE_REQUEST; unknown ID bị shared validator từ chối. Input/upstream bytes không đổi.

## Chạy và giới hạn

```sh
python3 scripts/smoke_br11a.py
```

Smoke tạo M02 decision, action và PENDING outcome synthetic trong TemporaryDirectory, chạy mock qua process mới, kiểm IDs `d`, `obs-a-1`, `obs-b-1`, `a`, `o`; sau đó kiểm stale/future/orphan và no-write. Kết quả: `BR-11a PASS: mock grounded context; stale/future/orphan boundaries; no writes`.

Không có retry/timeout/provider SDK vì chưa có live adapter; không secret/model/prompt version live. Context vẫn caller-declared, source_ref không xác thực. Không dùng outcome snapshots để cộng commission; không ghi EvaluationRecord/Proposal (BR-12), không gọi action executor. Live provider và proof riêng chỉ làm sau khi có cấu hình/nguồn được cấp quyền.
