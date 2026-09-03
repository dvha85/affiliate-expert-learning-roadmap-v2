# Affiliate Bot — deterministic learner baseline và continuity anchor

Workspace này là reference baseline (mốc tham chiếu) cho **M01 — Smallest Deterministic Bot v0.1** và **M02 — Trustworthy History + Replay v0.2**.

Sau M02, learner **không bỏ Bot này để chuyển sang một demo khác**. M03–M11 được tích hợp dần vào cùng learner Bot/workspace; `lab/mission-runtime/` chỉ là conformance oracle/harness để đối chiếu behavior/failure semantics, không phải Bot thứ hai.

Continuity contract: `docs/architecture/LEARNER-BOT-CONTINUITY.md`  
Continuity checklist: `starter-kits/CONTINUITY-CHECKPOINT.md`

## M01 baseline

```bash
cd lab/affiliate-bot
go run ./cmd/bot
go test ./...
```

Fixture mặc định `data/sample-observations.json` là canonical-shaped synthetic fixture (mẫu mô phỏng có shape chuẩn); nó chỉ chứng minh behavior kỹ thuật M01.

```text
known observations
→ deterministic formula + stable tie-break
→ RANK_SCENARIO | GET_MORE_DATA | HUMAN_REVIEW
→ NO external action
```

Baseline cố ý yếu: `price × commission_rate`.

```text
real evidence != RECOMMEND
RANK_SCENARIO != Approval != Execution
```

## M02 history + replay

M02 capture yêu cầu canonical identity/time: mỗi observation có `observation_id`, `subject_id`, provenance và `observed_at`. Fixture synthetic riêng là `data/m02-sample-observations.json`.

Capture một immutable decision snapshot:

```bash
go run ./cmd/bot history capture data/history.jsonl data/m02-sample-observations.json demo-1 2026-09-01T01:00:00Z 2026-09-03T10:00:00Z
```

Dừng process, chạy lại rồi query:

```bash
go run ./cmd/bot history list data/history.jsonl
```

Replay bằng `formula_version` đã record:

```bash
go run ./cmd/bot history replay data/history.jsonl
```

Replay states:

```text
MATCH        — same version + same input integrity + same result
DRIFT        — same supported version nhưng result khác record
UNREPLAYABLE — formula version không được registry hiện tại hỗ trợ
```

Nếu `input_hash` sai, evidence linkage sai, timestamp/identity invalid hoặc JSONL corrupt, record fail closed với integrity error **trước khi** được coi là replay hợp lệ.

`input_hash` bảo vệ integrity của canonical input snapshot; nó không chứng minh input là market truth.

```text
replay MATCH != business truth
replay MATCH != RECOMMEND
history exists != execution permission
```

M02 vẫn A0 deterministic, local-only: không scrape/login/publish/message/spend và không tạo Approval/Execution permission.

## Từ M03 trở đi — mở rộng cùng learner Bot

Repo không copy reference implementation M03–M11 vào workspace này vì việc đó sẽ tạo hai implementation phải giữ parity. Thay vào đó, với mỗi Mission learner phải:

1. đọc contract + lesson + starter của Mission;
2. chạy `lab/mission-runtime` như test oracle;
3. thêm capability/adapter vào **cùng learner Bot/workspace**;
4. bind artifact mới tới exact artifact IDs từ Mission trước;
5. lưu Integration Evidence + Reality/Operated evidence.

Ví dụ M03 bắt đầu từ `DecisionPacket.decision_id` hiện có rồi thêm human-only `ActionRecord`/`OutcomeRecord`; M06 thêm watcher adapter nhưng canonical Observation/History vẫn thuộc Deterministic Core; M11 chỉ mở production sau E5 promotion + finite ProductionLease.
