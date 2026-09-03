---
mission_id: M01
title: Smallest Deterministic Bot v0.1
status: draft
minimum_evidence: E0 + E1 support
authority: A0 deterministic
external_side_effects: false
---

# Mission M01 — Smallest Deterministic Bot v0.1

## Ship target

```text
canonical observations/context
→ deterministic formula + stable tie-break
→ RANK_SCENARIO | GET_MORE_DATA | HUMAN_REVIEW
→ reason + missing evidence
→ NO external action
```

Go là reference implementation hiện hành cho M01, không phải curriculum authority.

## Required behavior

- cùng input + cùng version → cùng result;
- `missing != 0`;
- identity/currency/semantic conflict → `HUMAN_REVIEW`;
- missing/invalid evidence → `GET_MORE_DATA`;
- real evidence không tự nâng weak formula thành `RECOMMEND`;
- output luôn giải thích limitation của formula.

## PASS

### Capability
Deterministic runtime và tests đạt behavior contract.

### Reality
Dùng M00 evidence/context thật để giải thích input boundary; fixture chỉ là E0.

### Operated
Learner chạy runtime, failure cases và explain-back được giới hạn của baseline.

## Result

Bot v0.1 deterministic, no action authority.
