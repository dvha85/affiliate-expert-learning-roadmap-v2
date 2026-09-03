---
mission_id: M01
title: Smallest Deterministic Bot v0.1
status: ready
minimum_evidence: E0 + E1 support
authority: A0 deterministic
external_side_effects: false
starter: starter-kits/M01-deterministic-bot/
eval_pack: evals/M01-deterministic-bot/
runtime: lab/affiliate-bot/
---

# Mission M01 — Smallest Deterministic Bot v0.1

## Ship target

```text
canonical observations/context + formula_version
→ deterministic formula + stable tie-break
→ HUMAN_REVIEW > GET_MORE_DATA > RANK_SCENARIO
→ reason + missing/invalid evidence
→ NO external action
```

Go là reference implementation hiện hành cho M01, không phải curriculum authority.

## Required behavior

- cùng input + cùng version → cùng result;
- `0 != missing`;
- invalid field không được silently default;
- identity/currency/mixed-origin conflict → `HUMAN_REVIEW`;
- missing/invalid evidence → `GET_MORE_DATA` nếu không có conflict mạnh hơn;
- real evidence không tự nâng weak formula thành `RECOMMEND`;
- output luôn giải thích limitation của formula.

## Learner cards

`M01.1 → M01.2 → M01.3 → M01.4` trong `curriculum/M01/`.

## PASS

### Capability
Runtime, `go vet`, regression tests và executable eval pack đạt behavior contract.

### Reality
Đối chiếu input boundary với M00 evidence/context thật; fixture chỉ là E0 và không được dùng để claim market reality.

### Operated
Learner tự dự đoán, chạy runtime/evals, quan sát failure case và explain-back được giới hạn/authority ceiling.

```text
M01 PASS = Capability + Reality + Operated
```

## Result

Bot v0.1 deterministic, no action authority. M02 mới thêm trustworthy history + replay.
