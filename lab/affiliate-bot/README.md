# Affiliate Bot — deterministic learner runtime

Workspace duy nhất cho **M01 — Smallest Deterministic Bot v0.1** và **M02 — Trustworthy History + Replay v0.2**.

## M01 baseline

```bash
cd lab/affiliate-bot
go run ./cmd/bot
go test ./...
```

Fixture mặc định là `synthetic`; nó chỉ chứng minh behavior kỹ thuật.

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

Capture một immutable decision snapshot:

```bash
go run ./cmd/bot history capture data/history.jsonl data/sample-observations.json demo-1 2026-09-01T10:00:00Z 2026-09-03T10:00:00Z
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

`input_hash` bảo vệ integrity của canonical input snapshot; nó không chứng minh input là market truth.

```text
replay MATCH != RECOMMEND
history exists != execution permission
```
