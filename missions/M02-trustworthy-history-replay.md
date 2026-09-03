---
mission_id: M02
title: Trustworthy History + Replay v0.2
status: planned
requires_missions: [M01]
minimum_evidence: E1 + replayable local history
authority: A0 deterministic
external_side_effects: false
runtime: lab/affiliate-bot/
starter: starter-kits/M02-history-replay/
eval_pack: evals/M02-history-replay/
---

# Mission M02 — Trustworthy History + Replay v0.2

## Vì sao Mission này tồn tại

M01 có thể trả một deterministic decision ở hiện tại. Nhưng nếu Bot không giữ được **input nào**, **version công thức nào** và **decision nào** đã tạo ra output trước đây, ta không thể audit, replay hay phân biệt market change với code drift.

M02 biến một lần chạy thành lịch sử có thể kiểm chứng:

```text
M01 canonical input
+ explicit as_of
+ explicit ingested_at
+ formula_version
→ immutable HistoryRecord
→ append-only JSONL
→ deterministic query
→ replay bằng đúng formula version
→ MATCH | DRIFT | UNREPLAYABLE
→ NO external action
```

## Contract cốt lõi

### Identity

```text
product_id     = stable subject identity trong learner runtime
observation_id = một observation cụ thể của subject đó
record_id      = một decision/history snapshot cụ thể
```

Không dùng product name làm stable identity và không tái sử dụng một `observation_id` cho nội dung khác.

### Time

```text
observed_at = khi source/world được quan sát
ingested_at = khi Bot nhận/lưu record
as_of       = thời điểm decision context được đánh giá
```

Ba timestamp này có vai trò khác nhau. Arrival order không được giả làm world-time order.

### Append-only history

History learner canonical là JSONL append-only. Record cũ không được silently overwrite.

```text
same record_id + same canonical content
→ EXACT_DUPLICATE
→ idempotent, không append lần hai

same record_id + different canonical content
→ CONFLICT
→ reject / human review
```

Valid late/out-of-order record vẫn phải được preserve; query/report sort theo `as_of`, không theo thứ tự dòng trong file.

### Replay

Replay không có nghĩa “chạy code hiện tại lên dữ liệu cũ” một cách mơ hồ. Mỗi record phải lưu:

```text
formula_version
canonical input snapshot
input_hash
recorded decision
```

Replay dùng version registry của deterministic core:

```text
known version + same result → MATCH
known version + different result → DRIFT
unknown/retired version → UNREPLAYABLE
```

`UNREPLAYABLE` phải được báo rõ; không được tự động chạy formula mới rồi gọi đó là replay thành công.

## Ship target

M02 hoàn chỉnh khi repo có:

- lesson cards M02.1–M02.4;
- `HistoryRecord` machine contract;
- append/read/query/replay path trong `lab/affiliate-bot`;
- executable eval pack cho duplicate/conflict/out-of-order/corruption/replay drift;
- starter/checkpoint + private operated-evidence template;
- restart proof;
- CI bảo vệ M01 regression + M02 replay semantics.

## Reality boundary

Minimum reality của learner PASS:

```text
ít nhất 1 product_id thật
+ hai E1 observations ở observed_at khác nhau
+ cùng stable identity
+ history vẫn đọc được sau process restart
```

`UNCHANGED` là outcome hợp lệ. Không cần thị trường phải thay đổi để PASS.

Nếu t2 không quan sát được, ghi missing/access limitation trung thực; không copy last-known value và gọi đó là observation mới.

Synthetic fixtures chỉ chứng minh failure/replay behavior, không thay E1 reality.

## Safety / authority ceiling

M02 vẫn là local deterministic processing:

- không tự scrape/login;
- không publish/message/spend;
- không n8n/Agent tự thu thập dữ liệu;
- không external side effect;
- history/replay không tạo Approval hoặc Execution permission.

```text
replay MATCH != business truth
history exists != permission to act
Decision != Approval != Execution
```

## PASS

### Capability
- append-only history không overwrite evidence;
- exact duplicate idempotent, conflict fail rõ;
- corrupt/truncated JSONL fail closed;
- out-of-order record được preserve và query theo `as_of`;
- replay phân biệt `MATCH`, `DRIFT`, `UNREPLAYABLE`;
- M01 regression vẫn PASS.

### Reality
- có ít nhất hai E1 observations của cùng stable subject ở hai thời điểm quan sát khác nhau, hoặc blocker được ghi trung thực.

### Operated
- learner đã append, restart, query, replay và rerun cùng history/version để chứng minh output deterministic;
- learner explain-back được `observed_at != ingested_at != as_of` và giới hạn của replay.

## Result

Bot v0.2 có trustworthy local history + deterministic replay. Authority vẫn A0, không external action.
